package bucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	dataService "github.com/controlplane-com/libs-go/pkg/data-service"
)

// newTestResolver builds a resolver pointed at server with negligible backoff.
func newTestResolver(t *testing.T, server *httptest.Server, maxAttempts int) OrgResolver {
	t.Helper()
	resolver := NewDataServiceOrgResolver(dataService.NewClient(server.URL, "token", "test"), maxAttempts).(*dataServiceOrgResolver)
	resolver.backoff.InitialBackoff = time.Millisecond
	resolver.backoff.MaxBackoff = time.Millisecond
	return resolver
}

func TestResolveOrgFound(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = w.Write([]byte(`{"name":"my-org"}`))
	}))
	defer server.Close()

	exists, err := newTestResolver(t, server, 3).ResolveOrg(context.Background(), "my-org")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, []string{"/org/my-org"}, paths)
}

func TestResolveOrgNotFoundIsNotRetried(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	exists, err := newTestResolver(t, server, 3).ResolveOrg(context.Background(), "missing-org")
	require.NoError(t, err)
	require.False(t, exists)
	require.Equal(t, 1, calls)
}

func TestResolveOrgRetriesTransientFailures(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"name":"my-org"}`))
	}))
	defer server.Close()

	exists, err := newTestResolver(t, server, 3).ResolveOrg(context.Background(), "my-org")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, 3, calls)
}

func TestResolveOrgFailsAfterMaxAttempts(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	exists, err := newTestResolver(t, server, 2).ResolveOrg(context.Background(), "my-org")
	require.Error(t, err)
	require.False(t, exists)
	require.Equal(t, 2, calls)
	require.ErrorContains(t, err, "after 2 attempts")
}

func TestResolveOrgUnauthorizedIsRetriedNotTreatedAsMissing(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	exists, err := newTestResolver(t, server, 2).ResolveOrg(context.Background(), "my-org")
	require.Error(t, err)
	require.False(t, exists)
	require.Equal(t, 2, calls)
}

func TestResolveOrgUnreachableService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close()

	exists, err := newTestResolver(t, server, 2).ResolveOrg(context.Background(), "my-org")
	require.Error(t, err)
	require.False(t, exists)
}
