package delivery

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	cplnErrors "github.com/controlplane-com/libs-go/pkg/errors"
)

// ErrorType classifies a delivery error to determine retry behavior.
type ErrorType string

const (
	ErrorTypeTransient ErrorType = "transient"  // network issues, timeouts, 5xx
	ErrorTypeRateLimit ErrorType = "rate_limit" // 429
	ErrorTypeAuth      ErrorType = "auth"       // 401/403 — credentials invalid
	ErrorTypePermanent ErrorType = "permanent"  // 400/404 — won't succeed on retry
	ErrorTypeUnknown   ErrorType = "unknown"    // unclassified (retried conservatively)
)

// ClassifiedError wraps an error with its classification.
type ClassifiedError struct {
	OriginalError error
	ErrorType     ErrorType
	RetryAfter    *time.Duration // for rate-limit errors
	IsRetryable   bool
	Message       string
}

func (e *ClassifiedError) Error() string { return e.Message }

// statusCoder is implemented by errors carrying an HTTP status code, notably
// go-libs/errors.ErrorCode.
type statusCoder interface{ Code() int }

// messageStatusPattern extracts an HTTP status code from an error message, but
// only where it is explicitly labelled as one ("StatusCode: 403", "invalid HTTP
// response code (500)"). A bare number is never read as a status: an id, byte
// count or duration that happens to contain "403" is not an auth failure.
var messageStatusPattern = regexp.MustCompile(`(?i)(?:status(?:\s*code)?|response\s*code)\s*[:=]?\s*\(?([1-5]\d{2})\b`)

func classified(err error, t ErrorType, retryable bool, message string) *ClassifiedError {
	return &ClassifiedError{OriginalError: err, ErrorType: t, IsRetryable: retryable, Message: message}
}

// ClassifyError classifies an error to determine retry behavior. deliveryContext
// is an optional label (e.g. "email", "slack") used only for diagnostics.
//
// Structured signals win: an HTTP status carried by the error, a DomainError
// kind, or a known sentinel is authoritative. Message text is only consulted for
// errors that carry no structure, since matching prose is inherently ambiguous.
// Anything unrecognized stays retryable — dropping a delivery is worse than
// sending it twice.
func ClassifyError(err error, deliveryContext string) *ClassifiedError {
	if err == nil {
		return nil
	}
	if c := classifyStructured(err); c != nil {
		return c
	}
	return classifyMessage(err)
}

// classifyStructured classifies from the error's type. Returns nil if the error
// carries no structure to classify from.
func classifyStructured(err error) *ClassifiedError {
	var domainErr *cplnErrors.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Kind {
		case cplnErrors.ErrKindUnauthorized, cplnErrors.ErrKindForbidden:
			return classified(err, ErrorTypeAuth, false, "Authentication or authorization failed")
		case cplnErrors.ErrKindNotFound, cplnErrors.ErrKindValidation, cplnErrors.ErrKindConflict:
			return classified(err, ErrorTypePermanent, false, "Permanent error - will not retry")
		case cplnErrors.ErrKindUnavailable, cplnErrors.ErrKindInternal:
			return classified(err, ErrorTypeTransient, true, "Transient error - will retry")
		}
	}

	var coder statusCoder
	if errors.As(err, &coder) {
		if c := classifyStatus(err, coder.Code()); c != nil {
			return c
		}
	}

	// A cancelled or timed-out context means we never learned the outcome; retry.
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return classified(err, ErrorTypeTransient, true, "Transient error - will retry")
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return classified(err, ErrorTypeTransient, true, "Transient error - will retry")
	}
	return nil
}

// classifyStatus classifies an HTTP status code. Returns nil for codes outside
// the range it understands, so callers fall through to other signals.
func classifyStatus(err error, status int) *ClassifiedError {
	switch {
	case status == http.StatusTooManyRequests:
		return classified(err, ErrorTypeRateLimit, true, "Rate limit exceeded")
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return classified(err, ErrorTypeAuth, false, "Authentication or authorization failed")
	// 408 is the server giving up on a slow request — the retryable exception to 4xx.
	case status == http.StatusRequestTimeout:
		return classified(err, ErrorTypeTransient, true, "Transient error - will retry")
	case status >= 500 && status <= 599:
		return classified(err, ErrorTypeTransient, true, "Transient error - will retry")
	case status >= 400 && status <= 499:
		return classified(err, ErrorTypePermanent, false, "Permanent error - will not retry")
	}
	return nil
}

// classifyMessage is the fallback for errors that carry no structured signal.
func classifyMessage(err error) *ClassifiedError {
	msg := strings.ToLower(err.Error())

	if m := messageStatusPattern.FindStringSubmatch(msg); m != nil {
		if status, convErr := strconv.Atoi(m[1]); convErr == nil {
			if c := classifyStatus(err, status); c != nil {
				return c
			}
		}
	}

	containsAny := func(needles ...string) bool {
		for _, n := range needles {
			if strings.Contains(msg, n) {
				return true
			}
		}
		return false
	}

	switch {
	case containsAny("rate limit", "too many requests", "throttl"):
		return classified(err, ErrorTypeRateLimit, true, "Rate limit exceeded")
	case containsAny("unauthorized", "forbidden", "access denied", "accessdenied",
		"permission denied", "invalid token", "expired token", "invalid credentials",
		"authentication failed", "authorization failed"):
		return classified(err, ErrorTypeAuth, false, "Authentication or authorization failed")
	case containsAny("bad request", "not found", "invalid request", "does not exist"):
		return classified(err, ErrorTypePermanent, false, "Permanent error - will not retry")
	case containsAny("timeout", "timed out", "deadline exceeded", "connection", "network",
		"unavailable", "no such host", "unexpected eof", "reset by peer", "broken pipe"):
		return classified(err, ErrorTypeTransient, true, "Transient error - will retry")
	}
	return classified(err, ErrorTypeUnknown, true, fmt.Sprintf("Unknown error: %s", err.Error()))
}

// RetryStrategy determines when to retry based on error type and attempt count.
type RetryStrategy struct {
	MaxRetries     int
	BaseDelay      time.Duration
	MaxDelay       time.Duration
	Multiplier     float64
	JitterFraction float64 // 0.0..1.0
}

// DefaultRetryStrategy returns a sensible default (exponential backoff + jitter).
func DefaultRetryStrategy() *RetryStrategy {
	return &RetryStrategy{MaxRetries: 5, BaseDelay: time.Second, MaxDelay: 5 * time.Minute, Multiplier: 2.0, JitterFraction: 0.1}
}

// CalculateNextRetry returns when to retry next and whether to retry at all.
func (s *RetryStrategy) CalculateNextRetry(classifiedErr *ClassifiedError, attemptCount int) (time.Time, bool) {
	if attemptCount >= s.MaxRetries || classifiedErr == nil || !classifiedErr.IsRetryable {
		return time.Time{}, false
	}
	var delay time.Duration
	switch classifiedErr.ErrorType {
	case ErrorTypeRateLimit:
		if classifiedErr.RetryAfter != nil {
			delay = *classifiedErr.RetryAfter
		} else {
			delay = s.calculateExponentialBackoff(attemptCount, 30*time.Second)
		}
	case ErrorTypeTransient:
		delay = s.calculateExponentialBackoff(attemptCount, s.BaseDelay)
	case ErrorTypeUnknown:
		delay = s.calculateExponentialBackoff(attemptCount, s.BaseDelay*2)
	default:
		return time.Time{}, false
	}
	if delay > s.MaxDelay {
		delay = s.MaxDelay
	}
	return time.Now().UTC().Add(delay), true
}

func (s *RetryStrategy) calculateExponentialBackoff(attemptCount int, baseDelay time.Duration) time.Duration {
	exponential := float64(baseDelay) * math.Pow(s.Multiplier, float64(attemptCount))
	jitter := exponential * s.JitterFraction * (rand.Float64()*2 - 1)
	return time.Duration(exponential + jitter)
}

// ApplyRetry classifies a send error and updates the delivery State for the next
// attempt: a retryable error sets status=failed with a scheduled NextRetryAt;
// otherwise status=permanently_failed. Either way the error is appended to the
// audit history. Returns true if a retry was scheduled.
func (s *RetryStrategy) ApplyRetry(state *State, err error) bool {
	c := ClassifyError(err, "")
	nextRetry, shouldRetry := s.CalculateNextRetry(c, state.AttemptCount)
	errType := string(c.ErrorType)
	state.LastErrorType = &errType
	// Record the sender's error verbatim alongside the classification summary, so a
	// failed delivery can be diagnosed from its history without also having to find
	// the logs from the attempt.
	detail := ""
	if c.OriginalError != nil {
		detail = c.OriginalError.Error()
	}
	if shouldRetry {
		state.Status = StatusFailed
		state.NextRetryAt = &nextRetry
		state.appendError(errType, fmt.Sprintf("[Attempt %d] %s: %s", state.AttemptCount, c.ErrorType, c.Message), detail)
		return true
	}
	state.Status = StatusPermanentlyFailed
	state.NextRetryAt = nil
	state.appendError(errType, fmt.Sprintf("[Permanently failed] %s: %s - exhausted retries", c.ErrorType, c.Message), detail)
	return false
}
