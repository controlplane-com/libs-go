package delivery

import (
	"context"
	"errors"
	"fmt"
	"testing"

	cplnErrors "github.com/controlplane-com/libs-go/pkg/errors"
)

// dsClientError reproduces the error go-libs/data-service returns for a non-2xx
// response, so classification is tested against the shape senders actually emit.
func dsClientError(status int, body string) error {
	return cplnErrors.NewErrorCode(
		fmt.Sprintf("invalid HTTP response code (%d). Details: %s", status, body),
		status,
	)
}

func TestClassifyError(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantType  ErrorType
		wantRetry bool
	}{
		// Structured status wins over anything the message text might suggest. The
		// data-service client's own wording ("invalid HTTP response code") used to
		// match the "invalid" => permanent rule and strand every 5xx unretried.
		{"ds 500 is transient despite 'invalid' in message", dsClientError(500, "upstream exploded"), ErrorTypeTransient, true},
		{"ds 503 is transient", dsClientError(503, "unavailable"), ErrorTypeTransient, true},
		{"ds 502 is transient", dsClientError(502, "bad gateway"), ErrorTypeTransient, true},
		{"ds 400 is permanent", dsClientError(400, "malformed"), ErrorTypePermanent, false},
		{"ds 404 is permanent", dsClientError(404, "no such org"), ErrorTypePermanent, false},
		{"ds 429 is rate limited", dsClientError(429, "slow down"), ErrorTypeRateLimit, true},
		{"ds 408 is transient", dsClientError(408, "request timeout"), ErrorTypeTransient, true},

		// A 5xx body echoing "403" must not be read as an auth failure.
		{"5xx body containing 403 stays transient", dsClientError(500, `{"trace":"403 upstream"}`), ErrorTypeTransient, true},

		// The sentinel the data-service client returns for 401/403.
		{"UnauthorizedError sentinel", cplnErrors.UnauthorizedError, ErrorTypeAuth, false},

		// Bare numbers in prose are not status codes.
		// Not auth is the point: unrecognized wording falls to unknown, which still retries.
		{"id containing 403 is not auth", errors.New("delivery 1403 failed: could not reach host"), ErrorTypeUnknown, true},
		{"byte count containing 401 is not auth", errors.New("wrote 8401 bytes then connection reset by peer"), ErrorTypeTransient, true},
		{"duration containing 500 is not a server error", errors.New("slack rejected the payload: invalid_blocks after 500 ms"), ErrorTypeUnknown, true},

		// Labelled statuses in unstructured text are still honored.
		{"aws sdk StatusCode: 403", errors.New("https response error StatusCode: 403, RequestID: abc, api error AccessDenied: not authorized"), ErrorTypeAuth, false},
		{"aws sdk StatusCode: 500", errors.New("https response error StatusCode: 500, RequestID: abc"), ErrorTypeTransient, true},

		// Text signals for errors with no status at all.
		{"unauthorized text", errors.New("slack returned 401: unauthorized"), ErrorTypeAuth, false},
		{"rate limit text", errors.New("rate limit exceeded, retry later"), ErrorTypeRateLimit, true},
		{"timeout text", errors.New("connection timeout while sending"), ErrorTypeTransient, true},
		{"unknown stays retryable", errors.New("something nobody anticipated"), ErrorTypeUnknown, true},

		// Domain errors classify from their kind.
		{"domain unauthorized", &cplnErrors.DomainError{Kind: cplnErrors.ErrKindUnauthorized, Message: "nope"}, ErrorTypeAuth, false},
		{"domain unavailable", &cplnErrors.DomainError{Kind: cplnErrors.ErrKindUnavailable, Message: "down"}, ErrorTypeTransient, true},
		{"domain not found", &cplnErrors.DomainError{Kind: cplnErrors.ErrKindNotFound, Message: "gone"}, ErrorTypePermanent, false},

		// A context timeout means the outcome is unknown, so retry.
		{"context deadline", fmt.Errorf("sending: %w", context.DeadlineExceeded), ErrorTypeTransient, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyError(tc.err, "")
			if got.ErrorType != tc.wantType {
				t.Errorf("type = %q, want %q (err: %v)", got.ErrorType, tc.wantType, tc.err)
			}
			if got.IsRetryable != tc.wantRetry {
				t.Errorf("retryable = %v, want %v (err: %v)", got.IsRetryable, tc.wantRetry, tc.err)
			}
			if got.OriginalError == nil {
				t.Error("OriginalError must be preserved")
			}
		})
	}
}

func TestClassifyError_NilIsNil(t *testing.T) {
	if ClassifyError(nil, "") != nil {
		t.Fatal("expected nil for a nil error")
	}
}

// The original error must survive into the persisted history: a generic summary
// like "Authentication or authorization failed" is not diagnosable on its own.
func TestApplyRetry_PreservesOriginalError(t *testing.T) {
	s := DefaultRetryStrategy()
	s.MaxRetries = 3
	st := &State{AttemptCount: 1}

	s.ApplyRetry(st, cplnErrors.UnauthorizedError)

	if len(st.ErrorMessages) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(st.ErrorMessages))
	}
	if got := st.ErrorMessages[0].Detail; got != "unauthorized" {
		t.Errorf("Detail = %q, want the sender's original error %q", got, "unauthorized")
	}
}

func TestApplyRetry_PreservesOriginalErrorOnRetry(t *testing.T) {
	s := DefaultRetryStrategy()
	s.MaxRetries = 3
	st := &State{AttemptCount: 1}

	if !s.ApplyRetry(st, dsClientError(503, "upstream down")) {
		t.Fatal("a 503 from data-service must be retried")
	}
	if got := st.ErrorMessages[0].Detail; got == "" {
		t.Fatal("Detail must carry the original error on retryable attempts too")
	}
}
