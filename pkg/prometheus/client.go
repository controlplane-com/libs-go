// Package prometheus builds authenticated clients for the Prometheus HTTP API (Mimir,
// metrics-query, ...) and normalizes the values they answer with.
package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// BearerTransport adds a bearer token to every request it forwards. An empty token sends the
// request through untouched.
type BearerTransport struct {
	Token string
	Next  http.RoundTripper
}

func (t *BearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.Token)
	}
	next := t.Next
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(req)
}

// NewApi returns a Prometheus v1 API client for endpoint that authenticates with bearerToken.
func NewApi(endpoint, bearerToken string) (v1.API, error) {
	client, err := api.NewClient(api.Config{
		Address:      endpoint,
		RoundTripper: &BearerTransport{Token: bearerToken, Next: http.DefaultTransport},
	})
	if err != nil {
		return nil, err
	}
	return v1.NewAPI(client), nil
}

// TenantHeader is the header Mimir reads the tenant from.
const TenantHeader = "X-Scope-OrgID"

// HeaderTransport sets fixed headers on every request it forwards.
type HeaderTransport struct {
	Headers map[string]string
	Next    http.RoundTripper
}

func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}
	next := t.Next
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(req)
}

// NewApiWithHeaders returns a Prometheus v1 API client for endpoint that sends headers on every
// request.
func NewApiWithHeaders(endpoint string, headers map[string]string) (v1.API, error) {
	client, err := api.NewClient(api.Config{
		Address:      endpoint,
		RoundTripper: &HeaderTransport{Headers: headers, Next: http.DefaultTransport},
	})
	if err != nil {
		return nil, err
	}
	return v1.NewAPI(client), nil
}

// NewTenantApi returns a client for a Mimir endpoint that reads from one tenant. It is meant
// for in-cluster endpoints that do not require authentication (the query-frontend's direct
// port); a public front door such as metrics-query derives the tenant from the query instead.
func NewTenantApi(endpoint, tenant string) (v1.API, error) {
	return NewApiWithHeaders(endpoint, map[string]string{TenantHeader: tenant})
}
