package pubsub

import (
	"context"
	"sync"
)

// MockClient implements the Client interface for testing.
type MockClient[T Message] struct {
	MockPub *MockPublisher[T]
	MockCon *MockConsumer[T]
	Closed  bool
}

// NewMockClient creates a new mock client for testing.
func NewMockClient[T Message]() *MockClient[T] {
	return &MockClient[T]{
		MockPub: &MockPublisher[T]{},
		MockCon: &MockConsumer[T]{},
	}
}

func (c *MockClient[T]) Publisher() Publisher[T] {
	return c.MockPub
}

func (c *MockClient[T]) Consumer() Consumer[T] {
	return c.MockCon
}

func (c *MockClient[T]) Close() error {
	c.Closed = true
	return nil
}

// MockPublisher implements the Publisher interface for testing.
type MockPublisher[T Message] struct {
	mu            sync.Mutex
	PublishedMsgs []*T
	PublishError  error
	CloseCalled   bool
}

func (p *MockPublisher[T]) Publish(ctx context.Context, msg *T) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.PublishError != nil {
		return p.PublishError
	}
	p.PublishedMsgs = append(p.PublishedMsgs, msg)
	return nil
}

func (p *MockPublisher[T]) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CloseCalled = true
	return nil
}

// MockConsumer implements the Consumer interface for testing.
type MockConsumer[T any] struct {
	mu           sync.Mutex
	StartCalled  bool
	StopCalled   bool
	StartError   error
	Handler      MessageHandler[T]
	StartContext context.Context
}

func (c *MockConsumer[T]) Start(ctx context.Context, handler MessageHandler[T]) error {
	c.mu.Lock()
	c.StartCalled = true
	c.Handler = handler
	c.StartContext = ctx
	c.mu.Unlock()

	if c.StartError != nil {
		return c.StartError
	}

	// Block until context is cancelled (simulates real consumer)
	<-ctx.Done()
	return ctx.Err()
}

func (c *MockConsumer[T]) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.StopCalled = true
}

// SimulateMessage allows tests to simulate receiving a message by invoking the handler.
// This must be called after Start() has been invoked.
func (c *MockConsumer[T]) SimulateMessage(ctx context.Context, msg *T) error {
	c.mu.Lock()
	handler := c.Handler
	c.mu.Unlock()

	if handler == nil {
		return nil
	}
	return handler(ctx, msg)
}
