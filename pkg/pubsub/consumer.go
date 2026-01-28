package pubsub

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"go.uber.org/zap"
)

// consumerImpl implements the Consumer interface for GCP Pub/Sub.
type consumerImpl[T any] struct {
	sub                *pubsub.Subscriber
	minAckExtension    time.Duration
	maxOutstandingMsgs int
	numWorkers         int
	logger             *zap.SugaredLogger
	unmarshaler        Unmarshaler[T]
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
}

// Start begins receiving messages and processing them with the provided handler.
// It blocks until the context is cancelled or an unrecoverable error occurs.
func (c *consumerImpl[T]) Start(ctx context.Context, handler MessageHandler[T]) error {
	ctx, c.cancel = context.WithCancel(ctx)

	// Configure receive settings for exactly-once delivery
	c.sub.ReceiveSettings = pubsub.ReceiveSettings{
		// High extension period prevents premature ack expiry during processing
		// This is critical for exactly-once delivery to avoid unintentional ack expirations
		MinDurationPerAckExtension: c.minAckExtension,
		MaxDurationPerAckExtension: c.minAckExtension,
		// Control concurrency
		MaxOutstandingMessages: c.maxOutstandingMsgs,
		NumGoroutines:          c.numWorkers,
	}

	c.logger.Infow("Starting Pub/Sub consumer",
		"minAckExtension", c.minAckExtension,
		"maxOutstandingMsgs", c.maxOutstandingMsgs,
		"numWorkers", c.numWorkers,
	)

	c.wg.Add(1)
	defer c.wg.Done()

	return c.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		c.processMessage(ctx, msg, handler)
	})
}

// Stop gracefully shuts down the consumer.
func (c *consumerImpl[T]) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	c.logger.Info("Pub/Sub consumer stopped")
}

// processMessage handles a single message with exactly-once acknowledgment.
func (c *consumerImpl[T]) processMessage(ctx context.Context, msg *pubsub.Message, handler MessageHandler[T]) {
	// Parse the message using the provided unmarshaler
	typedMsg, err := c.unmarshaler(msg.Data)
	if err != nil {
		c.logger.Errorw("Failed to unmarshal message, acking to remove invalid message",
			"msgId", msg.ID,
			"error", err,
		)
		c.ackMessage(ctx, msg)
		return
	}

	logger := c.logger.With("msgId", msg.ID)
	logger.Debug("Processing message")

	// Process the message
	if err := handler(ctx, typedMsg); err != nil {
		logger.Warnw("Message processing failed, nacking for redelivery",
			"error", err,
		)
		msg.Nack()
		return
	}

	// Success - ack with result verification (required for exactly-once)
	c.ackMessage(ctx, msg)
	logger.Debug("Message processed and acked")
}

// ackMessage acknowledges a message with exactly-once result verification.
func (c *consumerImpl[T]) ackMessage(ctx context.Context, msg *pubsub.Message) {
	result := msg.AckWithResult()
	ackStatus, err := result.Get(ctx)

	switch ackStatus {
	case pubsub.AcknowledgeStatusSuccess:
		c.logger.Debugw("Message acked successfully", "msgId", msg.ID)

	case pubsub.AcknowledgeStatusInvalidAckID:
		// Ack ID expired - with exactly-once enabled, if the message was already
		// processed and acked by this or another consumer, it won't be redelivered.
		// If not, it will be redelivered, which is fine since processing is idempotent.
		c.logger.Warnw("Ack ID expired (message may be redelivered if not already processed)",
			"msgId", msg.ID,
		)

	case pubsub.AcknowledgeStatusFailedPrecondition:
		// With exactly-once delivery, this typically means another consumer
		// already acked this message. This is the deduplication at work.
		c.logger.Debugw("Already acked by another consumer (exactly-once dedup)",
			"msgId", msg.ID,
		)

	case pubsub.AcknowledgeStatusPermissionDenied:
		c.logger.Errorw("Permission denied for ack",
			"msgId", msg.ID,
			"error", err,
		)

	default:
		c.logger.Errorw("Ack failed with unknown status",
			"msgId", msg.ID,
			"status", ackStatus,
			"error", err,
		)
	}
}
