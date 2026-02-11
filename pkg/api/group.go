package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/group"
	"github.com/controlplane-com/types-go/pkg/query"
)

// GroupService handles operations on groups.
type GroupService struct {
	client *Client
}

// List returns an iterator over all groups in the organization.
func (s *GroupService) List(ctx context.Context, org string) iter.Seq2[group.Group, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group", org)
	return listIterator[group.Group](ctx, s.client, path)
}

// ListAll returns all groups in the organization.
func (s *GroupService) ListAll(ctx context.Context, org string) ([]group.Group, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group", org)
	return listAll[group.Group](ctx, s.client, path)
}

// ListPage returns a single page of groups.
func (s *GroupService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[group.Group], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[group.Group](ctx, s.client, path)
}

// Get returns a group by name.
func (s *GroupService) Get(ctx context.Context, org, name string) (*group.Group, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group/%s", org, name)
	var result group.Group
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new group.
func (s *GroupService) Create(ctx context.Context, org string, g *group.Group) (*group.Group, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group", org)
	var result group.Group
	if err := s.client.post(ctx, path, g, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing group.
func (s *GroupService) Update(ctx context.Context, org, name string, g *group.Group) (*group.Group, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group/%s", org, name)
	var result group.Group
	if err := s.client.patch(ctx, path, g, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a group by name.
func (s *GroupService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over groups matching the query.
func (s *GroupService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[group.Group, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group/-query", org)
	return queryIterator[group.Group](ctx, s.client, path, q)
}

// QueryAll returns all groups matching the query.
func (s *GroupService) QueryAll(ctx context.Context, org string, q *query.Query) ([]group.Group, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group/-query", org)
	return queryAll[group.Group](ctx, s.client, path, q)
}

// QueryPage returns a single page of groups matching the query.
func (s *GroupService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[group.Group], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/group/-query", org)
	return queryPage[group.Group](ctx, s.client, path, q)
}
