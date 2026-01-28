package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	"go.uber.org/zap"
)

// publisherImpl implements the Publisher interface for GCP Pub/Sub.
type publisherImpl[T Message] struct {
	pub                *pubsub.Publisher
	logger             *zap.SugaredLogger
	attributeExtractor AttributeExtractor[T]
}

// Publish sends a message to the Pub/Sub topic.
// It blocks until the publish is confirmed or fails.
func (p *publisherImpl[T]) Publish(ctx context.Context, msg *T) error {
	data, err := (*msg).Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Extract attributes using the provided extractor
	var attributes map[string]string
	if p.attributeExtractor != nil {
		attributes = p.attributeExtractor(msg)
	}

	result := p.pub.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: attributes,
	})

	// Wait for publish confirmation
	msgID, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Debugw("Published message",
		"msgId", msgID,
	)

	return nil
}

// Close stops the publisher and flushes any pending messages.
func (p *publisherImpl[T]) Close() error {
	p.pub.Stop()
	return nil
}
