package pubsub

import (
	"context"
)

// Message is the constraint for message types that can be published/consumed.
// Types must implement Marshal for serialization.
type Message interface {
	Marshal() ([]byte, error)
}

// MessageHandler is a function that processes a message of type T.
// Returns nil on success, or an error if processing failed (will trigger nack/retry).
type MessageHandler[T any] func(ctx context.Context, msg *T) error

// Publisher defines the interface for publishing messages to a Pub/Sub topic.
type Publisher[T Message] interface {
	// Publish sends a message to the Pub/Sub topic.
	// Blocks until the publish is confirmed or fails.
	Publish(ctx context.Context, msg *T) error

	// Close stops the publisher and flushes any pending messages.
	Close() error
}

// Consumer defines the interface for consuming messages from a Pub/Sub subscription.
type Consumer[T any] interface {
	// Start begins receiving messages and processing them with the provided handler.
	// Blocks until the context is cancelled or an unrecoverable error occurs.
	Start(ctx context.Context, handler MessageHandler[T]) error

	// Stop gracefully shuts down the consumer.
	Stop()
}

// Client defines the interface for a Pub/Sub client that provides both
// publishing and consuming capabilities.
type Client[T Message] interface {
	// Publisher returns the publisher for sending messages.
	Publisher() Publisher[T]

	// Consumer returns the consumer for receiving messages.
	Consumer() Consumer[T]

	// Close closes the client and all underlying connections.
	Close() error
}

// AttributeExtractor is a function that extracts message attributes for Pub/Sub.
// These attributes are stored alongside the message and can be used for filtering.
type AttributeExtractor[T any] func(msg *T) map[string]string

// Unmarshaler is a function that deserializes message data into a typed message.
type Unmarshaler[T any] func(data []byte) (*T, error)
