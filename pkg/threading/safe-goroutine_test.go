package threading

import (
	"errors"
	"testing"
)

func TestGoSafely_PanicError(t *testing.T) {
	e := <-GoSafely(func() error {
		var m map[int]int
		m[0] = 1
		return nil
	})
	var panicError *PanicError
	if !errors.As(e, &panicError) {
		t.Error("expected a PanicError")
	}
}

func TestGoSafely_PanicWithNonErrorValue(t *testing.T) {
	e := <-GoSafely(func() error {
		panic(1)
	})
	var panicError *PanicError
	if !errors.As(e, &panicError) {
		t.Error("expected a PanicError")
	}
}

func TestGoSafely_GenericError(t *testing.T) {
	e := <-GoSafely(func() error {
		return errors.New("generic error")
	})
	var panicError *PanicError
	if errors.As(e, &panicError) {
		t.Error("expected a generic error. Got a PanicError")
	}
}

func TestGoSafely_NoError(t *testing.T) {
	e := <-GoSafely(func() error {
		return nil
	})
	if e != nil {
		t.Error("expected nil error")
	}
}
