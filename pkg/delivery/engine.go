package delivery

import (
	"context"
	"encoding/json"
	"time"

	leaderElection "github.com/controlplane-com/libs-go/pkg/leader-election"
	"github.com/controlplane-com/libs-go/pkg/logging"
	"github.com/controlplane-com/libs-go/pkg/negotiate"
	"github.com/controlplane-com/libs-go/pkg/pubsub"
)

// Sender performs the actual delivery for a record. Errors are returned verbatim
// so the engine can classify them for retry (include HTTP status text where
// possible, e.g. "sendgrid returned 429: ...").
type Sender[D Delivery] interface {
	Send(ctx context.Context, d D) error
}

// TaskMessage is the Pub/Sub payload — just the delivery id; the engine re-loads
// the record. One message shape across all pipelines (each uses its own topic).
type TaskMessage struct {
	DeliveryID  string    `json:"deliveryId"`
	PublishedAt time.Time `json:"publishedAt"`
}

func (m TaskMessage) Marshal() ([]byte, error) { return json.Marshal(m) }

func unmarshalTaskMessage(data []byte) (*TaskMessage, error) {
	var m TaskMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func taskAttributes(m *TaskMessage) map[string]string {
	return map[string]string{"deliveryId": m.DeliveryID, "version": "1"}
}

// QueueClient is the per-pipeline Pub/Sub client (its own topic/subscription/DLQ).
type QueueClient = pubsub.Client[TaskMessage]

// NewQueueClient creates a delivery queue client (auto-creates topic/DLQ/sub).
func NewQueueClient(ctx context.Context, cfg pubsub.ClientConfig) (QueueClient, error) {
	return pubsub.NewClient(ctx, cfg, taskAttributes, unmarshalTaskMessage)
}

// Config configures an Engine.
type Config struct {
	// Name uniquely identifies the pipeline (leader-election class + job/service
	// name + log label).
	Name string
	// SyncInterval is how often the discovery loop scans for due records.
	SyncInterval time.Duration
	// MaxRetries caps retry attempts (0 = strategy default).
	MaxRetries int
	// Elector gates the discovery loop to a single leader.
	Elector leaderElection.Elector
	// Queue, when non-nil, distributes processing via Pub/Sub. Nil = poll-only
	// (the leader processes inline).
	Queue QueueClient
	// OnPermanentFailure, when set, is called after a record exhausts retries
	// (e.g. to raise an alert). Optional.
	OnPermanentFailure func(d Delivery, err error)
}

// Engine drives the shared delivery lifecycle. It satisfies the Name/Start/Stop
// background-job shape so callers can register it like any other service.
type Engine[D Delivery] struct {
	*negotiate.SyncJob[D]
	store      Store[D]
	sender     Sender[D]
	strategy   *RetryStrategy
	queue      QueueClient
	onPermFail func(Delivery, error)
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewEngine wires the discovery loop and (optional) consumer around a Store and Sender.
func NewEngine[D Delivery](store Store[D], sender Sender[D], cfg Config) *Engine[D] {
	ctx, cancel := context.WithCancel(context.Background())
	strategy := DefaultRetryStrategy()
	if cfg.MaxRetries > 0 {
		strategy.MaxRetries = cfg.MaxRetries
	}
	e := &Engine[D]{
		store:      store,
		sender:     sender,
		strategy:   strategy,
		queue:      cfg.Queue,
		onPermFail: cfg.OnPermanentFailure,
		ctx:        ctx,
		cancel:     cancel,
		SyncJob: negotiate.NewSyncJob(negotiate.SyncJobOptions[D]{
			NumWorkers:   1,
			BufferSize:   1,
			SyncInterval: cfg.SyncInterval,
			Name:         cfg.Name,
			Elector:      cfg.Elector,
		}).WithMetrics(nil),
	}
	e.InputDelegate = e.discover
	e.OutputDelegate = e.enqueueOrProcess
	return e
}

// discover is the SyncJob InputDelegate: all records currently due.
func (e *Engine[D]) discover() ([]D, error) {
	return e.store.ListDue(e.ctx, time.Now().UTC())
}

// enqueueOrProcess is the SyncJob OutputDelegate — it runs only on the leader.
// It submits each due record into the delivery system exactly once: publish to
// the queue (so the consumer processes it — distributed) or, in poll-only mode,
// process inline. Either way the record is stamped pushed_at so the discovery
// loop won't keep re-submitting an in-flight row.
func (e *Engine[D]) enqueueOrProcess(d D) error {
	if e.queue != nil {
		return e.pushToQueue(d.GetID())
	}
	// Poll-only: the leader is the delivery system. Stamp pushed_at for
	// observability/consistency (best-effort — processing is what matters here).
	if err := e.store.MarkPushed(e.ctx, d.GetID()); err != nil {
		logging.Logger().Sugar().Warnf("delivery[%s]: failed to mark %s pushed: %v", e.Name(), d.GetID(), err)
	}
	return e.Process(e.ctx, d.GetID())
}

// pushToQueue publishes an id to the queue and, on success, stamps pushed_at.
// Ordering is publish-then-mark: if the mark fails after a successful publish,
// pushed_at stays null and the next discovery cycle re-publishes (idempotent).
func (e *Engine[D]) pushToQueue(id string) error {
	if err := e.publish(id); err != nil {
		return err
	}
	return e.store.MarkPushed(e.ctx, id)
}

// Enqueue is a best-effort, low-latency nudge that submits already-committed
// outbox rows into the delivery system immediately, rather than waiting for the
// leader's discovery loop to find them. It is NOT the durability guarantee — the
// leader's discovery loop re-pushes any row still lacking pushed_at — so callers
// may ignore its error. No-op when the queue is disabled, since processing on a
// non-leader pod could duplicate sends (the leader poll picks the rows up).
func (e *Engine[D]) Enqueue(ids ...string) error {
	if e.queue == nil {
		logging.Logger().Sugar().Debugf("delivery[%s]: enqueue signal for %d delivery(ies) (poll path)", e.Name(), len(ids))
		return nil
	}
	for _, id := range ids {
		if err := e.pushToQueue(id); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine[D]) publish(id string) error {
	if err := e.queue.Publisher().Publish(e.ctx, &TaskMessage{DeliveryID: id, PublishedAt: time.Now().UTC()}); err != nil {
		logging.Logger().Sugar().Errorf("delivery[%s]: failed to publish %s: %v", e.Name(), id, err)
		return err
	}
	return nil
}

// Start runs the Pub/Sub consumer (when a queue is configured) and the
// leader-elected discovery loop.
func (e *Engine[D]) Start() <-chan error {
	errChan := make(chan error, 2)
	go func() {
		if e.queue == nil {
			return
		}
		logging.Logger().Sugar().Infof("delivery[%s]: starting Pub/Sub consumer", e.Name())
		if err := e.queue.Consumer().Start(e.ctx, e.handleMessage); err != nil && e.ctx.Err() == nil {
			logging.Logger().Sugar().Errorf("delivery[%s]: consumer error: %v", e.Name(), err)
			select {
			case errChan <- err:
			default:
			}
		}
	}()
	syncErr := e.SyncJob.Start()
	go func() {
		for err := range syncErr {
			select {
			case errChan <- err:
			default:
			}
		}
	}()
	return errChan
}

func (e *Engine[D]) Stop() {
	e.cancel()
	e.SyncJob.Stop()
	if e.queue != nil {
		e.queue.Consumer().Stop()
	}
}

func (e *Engine[D]) handleMessage(ctx context.Context, msg *TaskMessage) error {
	return e.Process(ctx, msg.DeliveryID)
}

// Process drives a single delivery through claim -> send -> finalize. Idempotent:
// a record already delivered or permanently failed is skipped, and the claim
// (status=in_progress) guards against concurrent processing.
func (e *Engine[D]) Process(ctx context.Context, id string) error {
	d, err := e.store.GetByID(ctx, id)
	if err != nil {
		logging.LoggerWithContext(ctx).Errorf("delivery[%s]: failed to load %s: %v", e.Name(), id, err)
		return err
	}
	st := d.DeliveryState()
	if st.Status == StatusDelivered || st.Status == StatusPermanentlyFailed {
		return nil
	}

	now := time.Now().UTC()
	fallback := now.Add(10 * time.Minute) // stuck-recovery: re-picked-up if a crash interrupts the send
	st.Status = StatusInProgress
	st.AttemptCount++
	st.LastAttemptAt = &now
	st.NextRetryAt = &fallback
	if err := e.store.Save(ctx, d); err != nil {
		return err
	}

	if sendErr := e.sender.Send(ctx, d); sendErr != nil {
		retried := e.strategy.ApplyRetry(st, sendErr)
		if err := e.store.Save(ctx, d); err != nil {
			return err
		}
		if !retried && e.onPermFail != nil {
			e.onPermFail(d, sendErr)
		}
		logging.LoggerWithContext(ctx).Warnf("delivery[%s]: %s attempt %d failed (status=%s): %v", e.Name(), id, st.AttemptCount, st.Status, sendErr)
		return sendErr
	}

	delivered := time.Now().UTC()
	st.Status = StatusDelivered
	st.DeliveredAt = &delivered
	st.NextRetryAt = nil
	st.LastErrorType = nil
	st.ErrorMessages = nil
	if err := e.store.Save(ctx, d); err != nil {
		return err
	}
	logging.LoggerWithContext(ctx).Infof("delivery[%s]: delivered %s", e.Name(), id)
	return nil
}
