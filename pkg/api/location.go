package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/location"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// LocationService handles operations on locations.
type LocationService struct {
	client *Client
}

// List returns an iterator over all locations in the organization.
func (s *LocationService) List(ctx context.Context, org string) iter.Seq2[location.Location, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/location", org)
	return listIterator[location.Location](ctx, s.client, path)
}

// ListAll returns all locations in the organization.
func (s *LocationService) ListAll(ctx context.Context, org string) ([]location.Location, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/location", org)
	return listAll[location.Location](ctx, s.client, path)
}

// ListPage returns a single page of locations.
func (s *LocationService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[location.Location], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/location", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPageWithClient[location.Location](ctx, s.client, path)
}

// Get returns a location by name.
func (s *LocationService) Get(ctx context.Context, org, name string) (*location.Location, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/location/%s", org, name)
	var result location.Location
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Query returns an iterator over locations matching the query.
func (s *LocationService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[location.Location, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/location/-query", org)
	return queryIterator[location.Location](ctx, s.client, path, q)
}

// QueryAll returns all locations matching the query.
func (s *LocationService) QueryAll(ctx context.Context, org string, q *query.Query) ([]location.Location, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/location/-query", org)
	return queryAll[location.Location](ctx, s.client, path, q)
}

// QueryPage returns a single page of locations matching the query.
func (s *LocationService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[location.Location], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/location/-query", org)
	return queryPageWithClient[location.Location](ctx, s.client, path, q)
}
