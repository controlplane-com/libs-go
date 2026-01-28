package threading

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// GoSafely executes the given function in a goroutine that will never panic. If the given function panics, GoSafely
// will recover and return a special error (PanicError)
func GoSafely(f func() error) <-chan error {
	c := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					stack := string(debug.Stack())
					c <- &PanicError{inner: err, stack: stack}
				} else {
					c <- &PanicError{inner: errors.New(fmt.Sprintf("Panic during goroutine execution: %v", r)), stack: string(debug.Stack())}
				}
			}
			close(c)
		}()

		err := f()
		if err != nil {
			c <- err
		}
	}()

	return c
}
