package audit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// received captures what the fake audit service saw for one request.
type received struct {
	method string
	path   string
	query  string
	auth   string
	body   *AuditEvent
}

// newFakeServer returns an httptest server plus a channel that receives one
// `received` per inbound request. The channel is buffered so the handler never
// blocks even if the test is slow to read.
func newFakeServer(t *testing.T) (*httptest.Server, <-chan received) {
	t.Helper()
	ch := make(chan received, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var ev AuditEvent
		_ = json.Unmarshal(raw, &ev)
		ch <- received{
			method: r.Method,
			path:   r.URL.Path,
			query:  r.URL.RawQuery,
			auth:   r.Header.Get("Authorization"),
			body:   &ev,
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, ch
}

func sampleEvent() *AuditEvent {
	return &AuditEvent{
		EventTime: "2026-07-08T00:00:00.000Z",
		RequestID: "req-1",
		Context:   &Context{RemoteIP: "1.2.3.4", EventSource: "api.cpln.io", Org: "acme"},
		Subject:   &Subject{Name: "pesho@google.com", Data: map[string]any{"type": "user"}},
		Resource:  &Resource{ID: "id-1", Type: "ipset", Name: "myset"},
		Action:    &Action{Type: "create"},
		Result:    &Result{Status: "OK", Message: "Created ipset myset"},
	}
}

func TestWritePostsCorrectBodyAndPath(t *testing.T) {
	srv, ch := newFakeServer(t)
	c := New(srv.URL, false, "")

	c.Write(context.Background(), sampleEvent())

	select {
	case got := <-ch:
		if got.method != http.MethodPost {
			t.Errorf("method = %q, want POST", got.method)
		}
		if got.path != "/audit/org/acme/auditctx/cpln" {
			t.Errorf("path = %q, want /audit/org/acme/auditctx/cpln", got.path)
		}
		if got.query != "async=true" {
			t.Errorf("query = %q, want async=true", got.query)
		}
		if got.auth != "" {
			t.Errorf("auth = %q, want empty (no token)", got.auth)
		}
		if got.body.RequestID != "req-1" {
			t.Errorf("requestId = %q, want req-1", got.body.RequestID)
		}
		if got.body.Context == nil || got.body.Context.Org != "acme" {
			t.Errorf("context.org not round-tripped: %+v", got.body.Context)
		}
		if got.body.Resource == nil || got.body.Resource.Type != "ipset" || got.body.Resource.Name != "myset" {
			t.Errorf("resource not round-tripped: %+v", got.body.Resource)
		}
		if got.body.Action == nil || got.body.Action.Type != "create" {
			t.Errorf("action.type not round-tripped: %+v", got.body.Action)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for audit POST")
	}
}

func TestWriteSendsBearerTokenWhenSet(t *testing.T) {
	srv, ch := newFakeServer(t)
	c := New(srv.URL, false, "sekret")

	c.Write(context.Background(), sampleEvent())

	select {
	case got := <-ch:
		if got.auth != "Bearer sekret" {
			t.Errorf("auth = %q, want Bearer sekret", got.auth)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for audit POST")
	}
}

func TestDisabledClientDoesNotPost(t *testing.T) {
	srv, ch := newFakeServer(t)
	c := New(srv.URL, true, "")

	c.Write(context.Background(), sampleEvent())

	select {
	case got := <-ch:
		t.Fatalf("disabled client posted to %q, want no request", got.path)
	case <-time.After(300 * time.Millisecond):
		// no request within the window — correct
	}
}

func TestNilEventDoesNotPost(t *testing.T) {
	srv, ch := newFakeServer(t)
	c := New(srv.URL, false, "")

	c.Write(context.Background(), nil)

	select {
	case got := <-ch:
		t.Fatalf("nil event posted to %q, want no request", got.path)
	case <-time.After(300 * time.Millisecond):
	}
}

// TestBaseURLTrailingSlashTrimmed guards the path join against a double slash
// when the endpoint is configured with a trailing slash.
func TestBaseURLTrailingSlashTrimmed(t *testing.T) {
	srv, ch := newFakeServer(t)
	c := New(srv.URL+"/", false, "")

	c.Write(context.Background(), sampleEvent())

	select {
	case got := <-ch:
		if got.path != "/audit/org/acme/auditctx/cpln" {
			t.Errorf("path = %q, want /audit/org/acme/auditctx/cpln", got.path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for audit POST")
	}
}
