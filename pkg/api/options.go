package api

import (
	"net/http"
	"strings"
	"time"
)

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithOrg sets the default organization for all API calls.
func WithOrg(org string) Option {
	return func(c *Client) {
		c.org = org
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithHeader adds a custom header to all requests.
func WithHeader(key, value string) Option {
	return func(c *Client) {
		if c.headers == nil {
			c.headers = make(http.Header)
		}
		c.headers.Set(key, value)
	}
}

// WithHeaders sets multiple custom headers for all requests.
func WithHeaders(headers http.Header) Option {
	return func(c *Client) {
		c.headers = headers.Clone()
	}
}

// WithBillingURL overrides the default billing service URL (billing-ng).
// This is typically only needed for local development or testing.
func WithBillingURL(url string) Option {
	return func(c *Client) {
		c.billingURL = strings.TrimSuffix(url, "/")
	}
}

// RatelimitBypassHeader carries the token that exempts a request from
// data-service rate limiting.
const RatelimitBypassHeader = "x-ratelimit-bypass"

// WithRatelimitBypass presents the bypass token so the client's calls skip
// data-service rate limiting. An empty token is a no-op.
func WithRatelimitBypass(token string) Option {
	return func(c *Client) {
		if token == "" {
			return
		}
		WithHeader(RatelimitBypassHeader, token)(c)
	}
}
