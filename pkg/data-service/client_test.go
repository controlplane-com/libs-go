package data_service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	errors2 "github.com/controlplane-com/libs-go/pkg/errors"
)

func TestDoRetriesRateLimitAndResendsBody(t *testing.T) {
	var attempts int
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if attempts < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "test")
	var out map[string]bool
	response, err := client.Post("/thing", `{"a":1}`, &out)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	for i, b := range bodies {
		if b != `{"a":1}` {
			t.Fatalf("attempt %d got body %q, expected request body to be resent", i+1, b)
		}
	}
	if !out["ok"] {
		t.Fatalf("response body not unmarshalled: %v", out)
	}
}

func TestDoGivesUpAfterMaxRateLimitAttempts(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "test")
	_, err := client.Get("/org", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if attempts != rateLimitMaxAttempts {
		t.Fatalf("expected %d attempts, got %d", rateLimitMaxAttempts, attempts)
	}
	var ec errors2.ErrorCode
	ok := false
	if ec, ok = err.(errors2.ErrorCode); !ok {
		t.Fatalf("expected ErrorCode, got %T: %v", err, err)
	}
	if ec.Code() != http.StatusTooManyRequests {
		t.Fatalf("expected code 429, got %d", ec.Code())
	}
	if !strings.Contains(err.Error(), "GET "+server.URL+"/org") {
		t.Fatalf("error should identify the failing request, got: %v", err)
	}
}

func TestDoDoesNotRetryOtherErrorsAndIdentifiesRequest(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", "test")
	_, err := client.Get("/org", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if attempts != 1 {
		t.Fatalf("expected no retries for 500, got %d attempts", attempts)
	}
	if !strings.Contains(err.Error(), "GET "+server.URL+"/org") {
		t.Fatalf("error should identify the failing request, got: %v", err)
	}
}

func TestRateLimitBackoff(t *testing.T) {
	if got := rateLimitBackoff(1, "7"); got != 7*time.Second {
		t.Fatalf("Retry-After should win, got %s", got)
	}
	if got := rateLimitBackoff(1, "9999"); got != rateLimitMaxBackoff {
		t.Fatalf("Retry-After should be capped, got %s", got)
	}
	if got := rateLimitBackoff(1, "not-a-number"); got != rateLimitInitialBackoff {
		t.Fatalf("bad Retry-After should fall back to exponential, got %s", got)
	}
	if got := rateLimitBackoff(3, ""); got != 4*time.Second {
		t.Fatalf("expected 4s on attempt 3, got %s", got)
	}
	if got := rateLimitBackoff(20, ""); got != rateLimitMaxBackoff {
		t.Fatalf("exponential backoff should be capped, got %s", got)
	}
}
