package delivery

import (
	"errors"
	"testing"
)

func TestApplyRetry_TransientSchedulesRetry(t *testing.T) {
	s := DefaultRetryStrategy()
	s.MaxRetries = 3
	st := &State{AttemptCount: 1}

	if !s.ApplyRetry(st, errors.New("connection timeout while sending")) {
		t.Fatal("expected a retry to be scheduled")
	}
	if st.Status != StatusFailed {
		t.Fatalf("expected status %q, got %q", StatusFailed, st.Status)
	}
	if st.NextRetryAt == nil {
		t.Fatal("expected next_retry_at scheduled")
	}
	if st.LastErrorType == nil || *st.LastErrorType != string(ErrorTypeTransient) {
		t.Fatalf("expected lastErrorType=transient, got %v", st.LastErrorType)
	}
	if len(st.ErrorMessages) != 1 {
		t.Fatalf("expected 1 error history entry, got %d", len(st.ErrorMessages))
	}
}

func TestApplyRetry_ExhaustedIsPermanent(t *testing.T) {
	s := DefaultRetryStrategy()
	s.MaxRetries = 3
	st := &State{AttemptCount: 3} // no attempts left

	if s.ApplyRetry(st, errors.New("503 service unavailable")) {
		t.Fatal("expected no retry once attempts are exhausted")
	}
	if st.Status != StatusPermanentlyFailed {
		t.Fatalf("expected status %q, got %q", StatusPermanentlyFailed, st.Status)
	}
	if st.NextRetryAt != nil {
		t.Fatal("expected next_retry_at cleared")
	}
}

func TestApplyRetry_AuthIsPermanent(t *testing.T) {
	s := DefaultRetryStrategy()
	s.MaxRetries = 3
	st := &State{AttemptCount: 1}

	if s.ApplyRetry(st, errors.New("slack returned 401: unauthorized")) {
		t.Fatal("auth errors must not be retried")
	}
	if st.Status != StatusPermanentlyFailed {
		t.Fatalf("expected status %q, got %q", StatusPermanentlyFailed, st.Status)
	}
	if st.LastErrorType == nil || *st.LastErrorType != string(ErrorTypeAuth) {
		t.Fatalf("expected lastErrorType=auth, got %v", st.LastErrorType)
	}
}
