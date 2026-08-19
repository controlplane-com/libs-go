package nats

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/controlplane-com/libs-go/pkg/threading"
	"go.uber.org/zap"
)

type EventDisposition string

const (
	EventDispositionAccepted EventDisposition = "accepted"
	EventDispositionErrored  EventDisposition = "errored"
	EventDispositionIgnored  EventDisposition = "ignored"
)

type Msg = nats.Msg

type MessageHandler func(msg *Msg) (EventDisposition, error)

type Connection struct {
	ConnectionOptions
	conn                    *nats.Conn
	natsSubscription        *nats.Subscription
	subscriptionChannel     chan *nats.Msg
	eventDispositions       *prometheus.CounterVec
	eventHistogram          *prometheus.HistogramVec
	versionConflictsCounter prometheus.Counter
	queueDepth              *prometheus.GaugeVec
	connected               prometheus.Gauge
	closing                 atomic.Bool
	m                       *sync.Mutex
}

type ConnectionOptions struct {
	Endpoint            string
	Name                string
	Logger              *zap.SugaredLogger
	CredentialsFilePath string
	Registerer          prometheus.Registerer
}

type SubscriptionOptions struct {
	Name        string
	Topic       string
	Group       string
	Handler     MessageHandler
	BufferSize  int
	WorkerCount int
}

func NewConnection(options ConnectionOptions) (*Connection, error) {
	c := &Connection{
		ConnectionOptions: options,
		m:                 &sync.Mutex{},
	}
	if c.Registerer == nil {
		c.Registerer = prometheus.DefaultRegisterer
	}
	if err := c.initializeMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics")
	}
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}
	return c, nil
}

func (c *Connection) initializeMetrics() error {
	c.eventDispositions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: c.Name + "_events",
		Help: "Count of event by disposition",
	}, []string{"disposition", "topic"})

	c.eventHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    c.Name + "_event_processing_seconds",
		Help:    "Distribution of seconds spent processing events of varying dispositions",
		Buckets: []float64{0.0001, 0.001, 0.01, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.75, 1, 1.5, 2, 4, 8, 10},
	}, []string{"disposition", "topic"})

	c.queueDepth = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: c.Name + "_queue_depth",
		Help: "Number of pending events received on each topic",
	}, []string{"topic"})

	c.connected = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: c.Name + "_nats_connected",
		Help: "1 while the NATS connection is established, 0 while disconnected",
	})
	if err := c.Registerer.Register(c.eventDispositions); err != nil {
		return err
	}
	if err := c.Registerer.Register(c.eventHistogram); err != nil {
		return err
	}
	if err := c.Registerer.Register(c.queueDepth); err != nil {
		return err
	}
	if err := c.Registerer.Register(c.connected); err != nil {
		return err
	}
	return nil
}

func (c *Connection) IsConnected() bool {
	if c.conn == nil {
		return false
	}
	return c.conn.IsConnected()
}

func (c *Connection) connect() error {
	if c.IsConnected() {
		return nil
	}

	// Default reconnect (~2 min) leaves ChanSubscription loops blocked forever (nats.go
	// does not close the message channel on disconnect). MaxReconnects(-1) retries and
	// resubscribes indefinitely.
	opts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.ReconnectWait(5 * time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			c.connected.Set(0)
			c.Logger.Warnw("NATS disconnected", "endpoint", c.Endpoint, "err", err)
		}),
		// Fires on every failed reconnect attempt so an extended outage stays visible
		// in the logs instead of going silent after the single disconnect warning.
		nats.ReconnectErrHandler(func(_ *nats.Conn, err error) {
			c.Logger.Warnw("NATS reconnect attempt failed", "endpoint", c.Endpoint, "err", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			c.connected.Set(1)
			c.Logger.Infow("NATS reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ErrorHandler(func(_ *nats.Conn, sub *nats.Subscription, err error) {
			subj := ""
			if sub != nil {
				subj = sub.Subject
			}
			c.Logger.Warnw("NATS async error", "subject", subj, "err", err)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			c.connected.Set(0)
			if c.closing.Load() {
				c.Logger.Infow("NATS connection closed", "endpoint", c.Endpoint)
				return
			}
			// Fatalw exits; with MaxReconnects(-1) this only runs on terminal close.
			// Let Kubernetes restart the pod to recover.
			c.Logger.Fatalw("NATS connection closed terminally",
				"endpoint", c.Endpoint, "lastErr", nc.LastError())
		}),
	}
	if c.CredentialsFilePath != "" {
		opts = append(opts, nats.UserCredentials(c.CredentialsFilePath))
	}
	nc, err := nats.Connect(c.Endpoint, opts...)
	if err != nil {
		return err
	}
	c.Logger.Infof("Connected to NATS on %s", c.Endpoint)
	c.conn = nc
	c.connected.Set(1)
	return nil
}

func (c *Connection) Subscribe(ctx context.Context, options SubscriptionOptions) error {
	if c.conn == nil {
		err := c.connect()
		if err != nil {
			return err
		}
	}
	subscriptionChannel := make(chan *nats.Msg, options.BufferSize)
	natsSubscription, err := c.conn.ChanQueueSubscribe(options.Topic, options.Group, subscriptionChannel)
	if err != nil {
		return err
	}

	if options.WorkerCount <= 0 {
		options.WorkerCount = 1
	}

	_ = threading.WaitGroup(options.WorkerCount, func() error {
		c.rangeOverChannel(ctx, subscriptionChannel, options)
		return nil
	})

	if natsSubscription.IsValid() {
		err = natsSubscription.Unsubscribe()
		close(subscriptionChannel)
	}
	return err
}

func (c *Connection) Close() {
	if c != nil && c.conn != nil {
		// Mark the close as intentional so the ClosedHandler doesn't treat a
		// graceful shutdown (Drain/SIGTERM) as a terminal failure and Fatal-exit.
		c.closing.Store(true)
		c.conn.Close()
	}
}

func (c *Connection) rangeOverChannel(ctx context.Context, subscriptionChannel chan *nats.Msg, options SubscriptionOptions) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-subscriptionChannel:
			c.setQueueDepth(subscriptionChannel, options)
			if m == nil {
				return
			}
			start := time.Now()
			d, err := options.Handler(m)
			if err != nil {
				c.Logger.Errorf("Error while handling event: %v", err)
			}
			stop := time.Now()
			l := prometheus.Labels{"disposition": string(d), "topic": options.Topic}
			c.eventDispositions.With(l).Inc()
			c.eventHistogram.With(l).Observe(stop.Sub(start).Seconds())
		}
	}
}

func (c *Connection) metricName(prefixes ...string) string {
	return strings.ReplaceAll(strings.Join(prefixes, "_"), "-", "_")
}

func (c *Connection) setQueueDepth(channel chan *nats.Msg, options SubscriptionOptions) {
	c.m.Lock()
	defer c.m.Unlock()
	c.queueDepth.With(prometheus.Labels{"topic": options.Topic}).Set(float64(len(channel)))
}
