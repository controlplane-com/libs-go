package bucket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	dataService "github.com/controlplane-com/libs-go/pkg/data-service"
	"github.com/controlplane-com/libs-go/pkg/logging"
	"github.com/controlplane-com/libs-go/pkg/retry"
)

const defaultOrgResolveMaxAttempts = 3

// dataServiceOrgResolver resolves orgs by asking the data-service whether they exist.
type dataServiceOrgResolver struct {
	client      *dataService.DataServiceClient
	maxAttempts int
	backoff     retry.Config
}

// NewDataServiceOrgResolver returns an OrgResolver backed by the data-service. A 404
// answers "no such org" immediately; every other failure (unreachable service, 5xx,
// auth errors) is retried up to maxAttempts times before being reported to the caller.
func NewDataServiceOrgResolver(client *dataService.DataServiceClient, maxAttempts int) OrgResolver {
	if maxAttempts < 1 {
		maxAttempts = defaultOrgResolveMaxAttempts
	}
	return &dataServiceOrgResolver{
		client:      client,
		maxAttempts: maxAttempts,
		backoff: retry.Config{
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     2 * time.Second,
			Multiplier:     2.0,
		},
	}
}

func (r *dataServiceOrgResolver) ResolveOrg(ctx context.Context, name string) (bool, error) {
	var exists bool
	var lookupErr error
	attempt := 0
	err := retry.WithExponentialBackoff(ctx, r.backoff, func() (bool, error) {
		attempt++
		exists, lookupErr = r.lookup(ctx, name)
		if lookupErr == nil {
			return true, nil
		}
		if attempt >= r.maxAttempts {
			return true, fmt.Errorf("data-service lookup of org %s failed after %d attempts: %w", name, attempt, lookupErr)
		}
		logging.LoggerWithContext(ctx).Warnf("data-service lookup of org %s failed (attempt %d/%d): %v", name, attempt, r.maxAttempts, lookupErr)
		return false, lookupErr
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}

// lookup reports existence from a single GET. Only a 404 is a definitive "does not exist".
func (r *dataServiceOrgResolver) lookup(ctx context.Context, name string) (bool, error) {
	response, err := r.client.GetWithContext(ctx, fmt.Sprintf("/org/%s", url.PathEscape(name)), nil)
	if response != nil && response.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
