package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/org"
	"github.com/controlplane-com/types-go/pkg/query"
)

// OrgService handles operations on organizations.
type OrgService struct {
	client *Client
}

// List returns an iterator over all organizations the user has access to.
func (s *OrgService) List(ctx context.Context) iter.Seq2[org.Org, error] {
	path := "/org"
	return listIterator[org.Org](ctx, s.client, path)
}

// ListAll returns all organizations the user has access to.
func (s *OrgService) ListAll(ctx context.Context) ([]org.Org, error) {
	path := "/org"
	return listAll[org.Org](ctx, s.client, path)
}

// ListPage returns a single page of organizations.
func (s *OrgService) ListPage(ctx context.Context, cursor string) (*ListResponse[org.Org], error) {
	path := "/org"
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[org.Org](ctx, s.client, path)
}

// Get returns an organization by name.
func (s *OrgService) Get(ctx context.Context, name string) (*org.Org, error) {
	path := fmt.Sprintf("/org/%s", name)
	var result org.Org
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing organization.
func (s *OrgService) Update(ctx context.Context, name string, o *org.Org) (*org.Org, error) {
	path := fmt.Sprintf("/org/%s", name)
	var result org.Org
	if err := s.client.patch(ctx, path, o, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Query returns an iterator over organizations matching the query.
func (s *OrgService) Query(ctx context.Context, q *query.Query) iter.Seq2[org.Org, error] {
	path := "/org/-query"
	return queryIterator[org.Org](ctx, s.client, path, q)
}

// QueryAll returns all organizations matching the query.
func (s *OrgService) QueryAll(ctx context.Context, q *query.Query) ([]org.Org, error) {
	path := "/org/-query"
	return queryAll[org.Org](ctx, s.client, path, q)
}

// QueryPage returns a single page of organizations matching the query.
func (s *OrgService) QueryPage(ctx context.Context, q *query.Query) (*QueryResponse[org.Org], error) {
	path := "/org/-query"
	return queryPage[org.Org](ctx, s.client, path, q)
}
