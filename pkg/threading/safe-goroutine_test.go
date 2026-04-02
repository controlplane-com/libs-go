package threading

import (
	"errors"
	"strings"
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

func TestGoSafely_PanicStackTraceContainsPanicLocation(t *testing.T) {
	e := <-GoSafely(func() error {
		causePanicInNestedFunction()
		return nil
	})
	var panicError *PanicError
	if !errors.As(e, &panicError) {
		t.Fatal("expected a PanicError")
	}
	stack := panicError.Stack()
	if stack == "" {
		t.Error("expected non-empty stack trace")
	}
	// Stack should contain the function that caused the panic
	if !strings.Contains(stack, "causePanicInNestedFunction") {
		t.Errorf("stack trace should contain panic origin function, got:\n%s", stack)
	}
}

func causePanicInNestedFunction() {
	var s []int
	_ = s[0] // index out of range panic
}

func TestGoSafely_NestedChannelPreservesPanicError(t *testing.T) {
	// Simulate the metrics-agent pattern: GoSafely wrapping a select on channels
	innerChan := GoSafely(func() error {
		var s []int
		_ = s[0] // panic
		return nil
	})

	outerChan := GoSafely(func() error {
		// This is how the error propagates in metrics-agent
		return <-innerChan
	})

	err := <-outerChan
	var panicError *PanicError
	if !errors.As(err, &panicError) {
		t.Fatal("expected PanicError to be preserved through channel, got regular error")
	}
	stack := panicError.Stack()
	if !strings.Contains(stack, "TestGoSafely_NestedChannelPreservesPanicError") {
		t.Errorf("stack trace should contain test function, got:\n%s", stack)
	}
}

func TestGoSafely_IndexOutOfRangePanicCapturesStack(t *testing.T) {
	// Recreate the exact conditions from the bug report
	e := <-GoSafely(func() error {
		emptySlice := []byte{}
		_ = emptySlice[0] // "index out of range [0] with length 0"
		return nil
	})
	var panicError *PanicError
	if !errors.As(e, &panicError) {
		t.Fatal("expected a PanicError")
	}

	// Verify error message
	if !strings.Contains(panicError.Error(), "index out of range") {
		t.Errorf("expected 'index out of range' in error, got: %s", panicError.Error())
	}

	// Verify stack trace is meaningful (not just recover location)
	stack := panicError.Stack()
	lines := strings.Split(stack, "\n")
	if len(lines) < 4 {
		t.Errorf("stack trace too short, expected multiple frames, got:\n%s", stack)
	}
}
