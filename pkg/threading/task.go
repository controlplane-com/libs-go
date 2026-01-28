package threading

import (
	"errors"
	"fmt"
	cplnErrors "github.com/controlplane-com/libs-go/pkg/errors"
)

type PanicError struct {
	inner error
	stack string
}

func (p *PanicError) Stack() string {
	return p.stack
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("(recovered): %v", p.inner)
}

type TaskResult[T any] struct {
	Value T
	Error error
}

type Task[T any] struct {
	worker  func() (T, error)
	c       chan *TaskResult[T]
	r       *TaskResult[T]
	running bool
}

func (t *Task[T]) Await() (T, error) {
	if t.r != nil {
		return t.r.Value, t.r.Error
	}

	if !t.running {
		t.Run()
	}
	t.r = <-t.c
	t.running = false
	return t.r.Value, t.r.Error
}

func NewTask[T any](worker func() (T, error)) *Task[T] {
	return &Task[T]{c: make(chan *TaskResult[T]), worker: worker}
}

func (t *Task[T]) Run() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					t.c <- &TaskResult[T]{Error: &PanicError{inner: err}}
				} else {
					t.c <- &TaskResult[T]{Error: &PanicError{inner: errors.New(fmt.Sprintf("Panic during task execution: %v", r))}}
				}
			}
		}()

		r, err := t.worker()
		t.c <- &TaskResult[T]{Value: r, Error: err}
	}()
	t.running = true
}

func RunTasks[T any](tasks ...*Task[T]) ([]T, error) {
	var errs []error
	var results []T
	for _, t := range tasks {
		t.Run()
	}
	for _, t := range tasks {
		r, err := t.Await()
		if err != nil {
			errs = append(errs, err)
		}
		results = append(results, r)
	}
	if len(errs) > 0 {
		return results, cplnErrors.NewCompoundError("error(s) while running tasks", errs...)
	}
	return results, nil
}
