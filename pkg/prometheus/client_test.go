package prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBearerTransportSetsAuthorization(t *testing.T) {
	var seen string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("Authorization")
	}))
	defer server.Close()

	client := &http.Client{Transport: &BearerTransport{Token: "secret"}}
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, "Bearer secret", seen)

	client = &http.Client{Transport: &BearerTransport{}}
	resp, err = client.Get(server.URL)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, "", seen, "an empty token leaves the request untouched")
}

func TestNewApi(t *testing.T) {
	api, err := NewApi("http://metrics-query:4047", "token")
	require.NoError(t, err)
	assert.NotNil(t, api)

	_, err = NewApi("://not a url", "token")
	assert.Error(t, err)
}

func TestHeaderTransportSetsHeaders(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer srv.Close()

	api, err := NewTenantApi(srv.URL, "controlplane")
	if err != nil {
		t.Fatalf("NewTenantApi: %v", err)
	}
	if _, _, err := api.Query(context.Background(), "up", time.Now()); err != nil {
		t.Fatalf("Query: %v", err)
	}
	if got.Get(TenantHeader) != "controlplane" {
		t.Fatalf("%s = %q, want controlplane", TenantHeader, got.Get(TenantHeader))
	}
}
