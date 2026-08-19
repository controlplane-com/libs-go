package nats

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func startServer(t *testing.T, port int) *server.Server {
	t.Helper()
	opts := &server.Options{Host: "127.0.0.1", Port: port, NoLog: true, NoSigs: true}
	s, err := server.NewServer(opts)
	require.NoError(t, err)
	go s.Start()
	require.True(t, s.ReadyForConnections(10*time.Second), "embedded NATS server failed to start")
	return s
}

func waitFor(t *testing.T, timeout time.Duration, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out after %s waiting for %s", timeout, what)
}

// TestConnectionSurvivesServerOutage proves the failure mode from the 2026-07-16
// image-resolver incident cannot recur: during a full NATS outage the client must
// log the disconnect, keep logging failed reconnect attempts, expose a connected=0
// gauge, and — once the server returns — reconnect, resubscribe, and resume
// delivering messages without a process restart.
func TestConnectionSurvivesServerOutage(t *testing.T) {
	s := startServer(t, -1)
	port := s.Addr().(*net.TCPAddr).Port
	url := s.ClientURL()

	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core).Sugar()
	hasLog := func(substr string) func() bool {
		return func() bool {
			for _, e := range logs.All() {
				if strings.Contains(e.Message, substr) {
					return true
				}
			}
			return false
		}
	}

	c, err := NewConnection(ConnectionOptions{
		Endpoint:   url,
		Name:       "test_conn",
		Logger:     logger,
		Registerer: prometheus.NewRegistry(),
	})
	require.NoError(t, err)
	require.True(t, c.IsConnected())
	require.Equal(t, float64(1), testutil.ToFloat64(c.connected))

	received := make(chan string, 16)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = c.Subscribe(ctx, SubscriptionOptions{
			Name:  "test",
			Topic: "test.subject",
			Group: "test-group",
			Handler: func(msg *Msg) (EventDisposition, error) {
				received <- string(msg.Data)
				return EventDispositionAccepted, nil
			},
			BufferSize:  16,
			WorkerCount: 1,
		})
	}()

	publish := func(pub *nats.Conn, payload string) {
		t.Helper()
		// Retry until the queue subscription is registered server-side and the
		// message round-trips to the handler.
		waitFor(t, 10*time.Second, fmt.Sprintf("handler to receive %q", payload), func() bool {
			require.NoError(t, pub.Publish("test.subject", []byte(payload)))
			require.NoError(t, pub.Flush())
			select {
			case got := <-received:
				require.Equal(t, payload, got)
				return true
			case <-time.After(200 * time.Millisecond):
				return false
			}
		})
	}

	pub, err := nats.Connect(url)
	require.NoError(t, err)
	publish(pub, "before-outage")
	pub.Close()

	// Full outage: the server goes away entirely.
	s.Shutdown()
	s.WaitForShutdown()

	waitFor(t, 10*time.Second, "disconnect warning", hasLog("NATS disconnected"))
	waitFor(t, 10*time.Second, "reconnect-attempt failure warning", hasLog("NATS reconnect attempt failed"))
	require.Equal(t, float64(0), testutil.ToFloat64(c.connected))

	// Server comes back on the same port; the client must reconnect and resubscribe
	// on its own (ReconnectWait is 5s, so allow a generous window).
	s2 := startServer(t, port)
	defer s2.Shutdown()

	waitFor(t, 30*time.Second, "reconnection", hasLog("NATS reconnected"))
	waitFor(t, 10*time.Second, "connected gauge to recover", func() bool {
		return testutil.ToFloat64(c.connected) == 1
	})

	pub2, err := nats.Connect(url)
	require.NoError(t, err)
	defer pub2.Close()
	publish(pub2, "after-outage")

	// Graceful shutdown must not take the Fatal path (which would os.Exit the
	// test binary) — it logs an informational close instead.
	c.Close()
	waitFor(t, 10*time.Second, "graceful close log", hasLog("NATS connection closed"))
	require.False(t, hasLog("NATS connection closed terminally")())
	require.Equal(t, float64(0), testutil.ToFloat64(c.connected))
}

// TestNewConnectionFailsFastWhenUnreachable pins the initial-connect contract:
// NewConnection returns an error (callers exit and the pod restarts) rather than
// hanging when NATS is down at startup.
func TestNewConnectionFailsFastWhenUnreachable(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	_, err := NewConnection(ConnectionOptions{
		Endpoint:   "nats://127.0.0.1:1", // nothing listens here
		Name:       "test_conn_unreachable",
		Logger:     zap.New(core).Sugar(),
		Registerer: prometheus.NewRegistry(),
	})
	require.Error(t, err)
}
