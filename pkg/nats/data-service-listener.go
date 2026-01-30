package nats

import (
	"context"
	"encoding/json"
	"fmt"

	gonats "github.com/nats-io/nats.go"

	"github.com/controlplane-com/libs-go/pkg/events"
	"github.com/controlplane-com/libs-go/pkg/web/services"
)

type DataServiceListenerOptions[T any] struct {
	Name        string
	Kind        string
	Handler     func(t events.DataServiceEvent[T]) error
	BufferSize  int
	WorkerCount int
	Connection  *Connection
}

func (t DataServiceListenerOptions[T]) Validate() error {
	if t.Connection == nil {
		return fmt.Errorf("connection is nil")
	}
	if t.WorkerCount <= 0 {
		t.WorkerCount = 1
	}
	if t.BufferSize < 0 {
		t.BufferSize = 0
	}
	if t.Name == "" {
		return fmt.Errorf("name is empty")
	}
	if t.Kind == "" {
		return fmt.Errorf("kind is empty")
	}
	if t.Handler == nil {
		return fmt.Errorf("handler is empty")
	}
	return nil
}

type DataServiceListener[T any] struct {
	ctx          context.Context
	cancelFunc   context.CancelFunc
	subscription gonats.Subscription
	startChan    chan error
	running      bool
	options      DataServiceListenerOptions[T]
}

func NewDataServiceListener[T any](options DataServiceListenerOptions[T]) (services.Service, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	return &DataServiceListener[T]{
		options: options,
	}, nil
}

func (d *DataServiceListener[T]) Start() <-chan error {
	if d.running {
		return nil
	}
	d.ctx, d.cancelFunc = context.WithCancel(context.Background())
	d.startChan = make(chan error, d.options.BufferSize)
	d.running = true
	go func() {
		d.startChan <- d.subscribe()
		close(d.startChan)
	}()
	return d.startChan
}

func (d *DataServiceListener[T]) Stop() {
	d.cancelFunc()
	<-d.startChan
	d.running = false
}

func (d *DataServiceListener[T]) Name() string {
	return fmt.Sprintf("dsl-%s", d.options.Name)
}

func (d *DataServiceListener[T]) subscribe() error {
	return d.options.Connection.Subscribe(d.ctx, SubscriptionOptions{
		Name:        d.options.Name,
		WorkerCount: d.options.WorkerCount,
		Topic:       "controlplane.data.*." + d.options.Kind + ".*",
		Group:       d.Name(),
		BufferSize:  d.options.BufferSize,
		Handler: func(m *Msg) (EventDisposition, error) {
			var event events.DataServiceEvent[T]
			if err := json.Unmarshal(m.Data, &event); err != nil {
				return EventDispositionErrored, err
			}
			if err := d.options.Handler(event); err != nil {
				return EventDispositionErrored, err
			}
			return EventDispositionAccepted, nil
		},
	})
}
