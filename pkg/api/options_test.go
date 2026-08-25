package api

import (
	"net/http"
	"testing"
	"time"
)

func TestWithOrg(t *testing.T) {
	c := &Client{}
	WithOrg("test-org")(c)
	if c.org != "test-org" {
		t.Errorf("org = %q, want %q", c.org, "test-org")
	}
}

func TestWithOrg_Empty(t *testing.T) {
	c := &Client{org: "existing-org"}
	WithOrg("")(c)
	if c.org != "" {
		t.Errorf("org = %q, want empty string", c.org)
	}
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := &Client{}
	WithHTTPClient(custom)(c)
	if c.httpClient != custom {
		t.Error("HTTP client not set correctly")
	}
}

func TestWithHTTPClient_ReplacesExisting(t *testing.T) {
	original := &http.Client{Timeout: 10 * time.Second}
	replacement := &http.Client{Timeout: 20 * time.Second}
	c := &Client{httpClient: original}
	WithHTTPClient(replacement)(c)
	if c.httpClient != replacement {
		t.Error("HTTP client not replaced correctly")
	}
	if c.httpClient.Timeout != 20*time.Second {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, 20*time.Second)
	}
}

func TestWithTimeout(t *testing.T) {
	c := &Client{httpClient: &http.Client{}}
	WithTimeout(30 * time.Second)(c)
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, 30*time.Second)
	}
}

func TestWithTimeout_UpdatesExisting(t *testing.T) {
	c := &Client{httpClient: &http.Client{Timeout: 10 * time.Second}}
	WithTimeout(60 * time.Second)(c)
	if c.httpClient.Timeout != 60*time.Second {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, 60*time.Second)
	}
}

func TestWithUserAgent(t *testing.T) {
	c := &Client{}
	WithUserAgent("test-agent/1.0")(c)
	if c.userAgent != "test-agent/1.0" {
		t.Errorf("userAgent = %q, want %q", c.userAgent, "test-agent/1.0")
	}
}

func TestWithUserAgent_Empty(t *testing.T) {
	c := &Client{userAgent: "existing-agent"}
	WithUserAgent("")(c)
	if c.userAgent != "" {
		t.Errorf("userAgent = %q, want empty string", c.userAgent)
	}
}

func TestWithHeader(t *testing.T) {
	c := &Client{}
	WithHeader("X-Custom", "value")(c)
	if c.headers == nil {
		t.Fatal("headers map should be initialized")
	}
	if c.headers.Get("X-Custom") != "value" {
		t.Errorf("header = %q, want %q", c.headers.Get("X-Custom"), "value")
	}
}

func TestWithHeader_MultipleHeaders(t *testing.T) {
	c := &Client{}
	WithHeader("X-One", "1")(c)
	WithHeader("X-Two", "2")(c)
	if c.headers.Get("X-One") != "1" {
		t.Errorf("X-One = %q, want %q", c.headers.Get("X-One"), "1")
	}
	if c.headers.Get("X-Two") != "2" {
		t.Errorf("X-Two = %q, want %q", c.headers.Get("X-Two"), "2")
	}
}

func TestWithHeader_Overwrite(t *testing.T) {
	c := &Client{}
	WithHeader("X-Custom", "original")(c)
	WithHeader("X-Custom", "updated")(c)
	if c.headers.Get("X-Custom") != "updated" {
		t.Errorf("header = %q, want %q", c.headers.Get("X-Custom"), "updated")
	}
}

func TestWithHeaders(t *testing.T) {
	c := &Client{}
	h := http.Header{"X-One": []string{"1"}, "X-Two": []string{"2"}}
	WithHeaders(h)(c)
	if c.headers.Get("X-One") != "1" {
		t.Errorf("X-One = %q, want %q", c.headers.Get("X-One"), "1")
	}
	if c.headers.Get("X-Two") != "2" {
		t.Errorf("X-Two = %q, want %q", c.headers.Get("X-Two"), "2")
	}
}

func TestWithHeaders_Clone(t *testing.T) {
	c := &Client{}
	h := http.Header{"X-Custom": []string{"original"}}
	WithHeaders(h)(c)

	// Modify the original header
	h.Set("X-Custom", "modified")

	// Client should have the original value (cloned)
	if c.headers.Get("X-Custom") != "original" {
		t.Errorf("header should be cloned, got %q, want %q", c.headers.Get("X-Custom"), "original")
	}
}

func TestWithHeaders_ReplacesExisting(t *testing.T) {
	c := &Client{headers: http.Header{"X-Old": []string{"old"}}}
	h := http.Header{"X-New": []string{"new"}}
	WithHeaders(h)(c)
	if c.headers.Get("X-Old") != "" {
		t.Error("old headers should be replaced")
	}
	if c.headers.Get("X-New") != "new" {
		t.Errorf("X-New = %q, want %q", c.headers.Get("X-New"), "new")
	}
}

func TestNewClient_DefaultValues(t *testing.T) {
	c := New("test-token")
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.token != "test-token" {
		t.Errorf("token = %q, want %q", c.token, "test-token")
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if c.httpClient.Timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, DefaultTimeout)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	c := New("test-token",
		WithOrg("my-org"),
		WithTimeout(60*time.Second),
		WithUserAgent("my-agent/2.0"),
	)
	if c.org != "my-org" {
		t.Errorf("org = %q, want %q", c.org, "my-org")
	}
	if c.httpClient.Timeout != 60*time.Second {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, 60*time.Second)
	}
	if c.userAgent != "my-agent/2.0" {
		t.Errorf("userAgent = %q, want %q", c.userAgent, "my-agent/2.0")
	}
}

func TestNewWithBaseURL(t *testing.T) {
	c := NewWithBaseURL("https://custom.api.io", "test-token")
	if c.baseURL != "https://custom.api.io" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://custom.api.io")
	}
	if c.token != "test-token" {
		t.Errorf("token = %q, want %q", c.token, "test-token")
	}
}

func TestNewWithBaseURL_WithOptions(t *testing.T) {
	c := NewWithBaseURL("https://custom.api.io", "test-token",
		WithOrg("custom-org"),
	)
	if c.baseURL != "https://custom.api.io" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://custom.api.io")
	}
	if c.org != "custom-org" {
		t.Errorf("org = %q, want %q", c.org, "custom-org")
	}
}

func TestClient_SetToken(t *testing.T) {
	c := New("original-token")
	c.SetToken("new-token")
	if c.token != "new-token" {
		t.Errorf("token = %q, want %q", c.token, "new-token")
	}
}

func TestClient_SetOrg(t *testing.T) {
	c := New("token")
	c.SetOrg("my-org")
	if c.org != "my-org" {
		t.Errorf("org = %q, want %q", c.org, "my-org")
	}
}

func TestClient_Org(t *testing.T) {
	c := New("token", WithOrg("test-org"))
	if c.Org() != "test-org" {
		t.Errorf("Org() = %q, want %q", c.Org(), "test-org")
	}
}

func TestClient_resolveOrg(t *testing.T) {
	c := New("token", WithOrg("default-org"))

	// When org is provided, use it
	if got := c.resolveOrg("provided-org"); got != "provided-org" {
		t.Errorf("resolveOrg(provided) = %q, want %q", got, "provided-org")
	}

	// When org is empty, use default
	if got := c.resolveOrg(""); got != "default-org" {
		t.Errorf("resolveOrg(empty) = %q, want %q", got, "default-org")
	}
}

func TestWithRatelimitBypass(t *testing.T) {
	c := &Client{}
	WithRatelimitBypass("bypass-token")(c)
	if c.headers.Get(RatelimitBypassHeader) != "bypass-token" {
		t.Errorf("header = %q, want %q", c.headers.Get(RatelimitBypassHeader), "bypass-token")
	}
}

func TestWithRatelimitBypass_EmptyTokenIsNoOp(t *testing.T) {
	c := &Client{}
	WithRatelimitBypass("")(c)
	if c.headers.Get(RatelimitBypassHeader) != "" {
		t.Errorf("header = %q, want empty", c.headers.Get(RatelimitBypassHeader))
	}
}
