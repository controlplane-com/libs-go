package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/controlplane-com/libs-go/pkg/logging"
)

// Client emits audit events to the audit service. It is fire-and-forget: Write
// never blocks the caller and never returns an error — a failed post is logged
// and swallowed (matching the TS AuditService.write .catch()). A disabled client
// is a no-op. Zero-value is not usable; construct with New.
type Client struct {
	baseURL  string
	disabled bool
	token    string
	http     *http.Client
}

// New builds a Client. baseURL is the audit service origin (e.g.
// "http://localhost:43000"); Write posts to
// {baseURL}/audit/org/{org}/auditctx/cpln?async=true — the same path the TS
// service and terminal-server use. When disabled is true every Write is a no-op.
// token, when non-empty, is sent as a Bearer Authorization header (the audit
// service's public endpoint requires it; the data-service sidecar does not, so
// it passes "").
func New(baseURL string, disabled bool, token string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		disabled: disabled,
		token:    token,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Write marshals the event now, so a caller may mutate what it references once
// Write returns, and posts it on a goroutine with a detached context, so the
// round-trip never adds latency to or fails the caller's request. Errors are logged
// at info and swallowed. No-op when the client is disabled or the event is nil.
func (c *Client) Write(_ context.Context, ev *AuditEvent) {
	if c == nil || c.disabled || ev == nil {
		return
	}
	org := ""
	if ev.Context != nil {
		org = ev.Context.Org
	}
	body, err := json.Marshal(ev)
	if err != nil {
		logging.Logger().Sugar().Infow("Did not persist audit event", "err", err)
		return
	}
	go c.post(org, body)
}

// post performs the synchronous POST of a marshalled event.
func (c *Client) post(org string, body []byte) {
	url := fmt.Sprintf("%s/audit/org/%s/auditctx/cpln?async=true", c.baseURL, org)

	// Detached, bounded context: the post must outlive the request that triggered
	// it, but must not hang forever.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		logging.Logger().Sugar().Infow("Did not persist audit event", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logging.Logger().Sugar().Infow("Did not persist audit event", "err", err)
		return
	}
	_ = resp.Body.Close()
}
