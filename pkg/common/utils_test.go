package common

import (
	"errors"
	"testing"
)

func TestExecuteWithRetries(t *testing.T) {
	type executeWithRetriesTest struct {
		failureCount int
		maxRetries   int
	}
	tests := map[string]executeWithRetriesTest{
		"success_after_one_retry_a": {
			failureCount: 1,
			maxRetries:   2,
		},
		"success_after_one_retry_b": {
			failureCount: 1,
			maxRetries:   1,
		},
		"failure_after_two_retries": {
			failureCount: 2,
			maxRetries:   1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			executions := 0
			retries, err := ExecuteWithRetries(func() error {
				executions++
				if executions <= test.failureCount {
					return errors.New("")
				}
				return nil
			}, test.maxRetries, 1, 1)
			if err != nil && test.maxRetries >= test.failureCount {
				t.Errorf("Expected ExecuteWithRetries not to return an error")
				t.FailNow()
				return
			}
			if err == nil && test.maxRetries < test.failureCount {
				t.Errorf("Expected ExecuteWithRetries to return an error")
				t.FailNow()
				return
			}
			if retries != test.failureCount {
				t.Errorf("Expected ExecuteWithRetries to retry %d times. Got %d", test.failureCount, retries)
			}
		})
	}
}
