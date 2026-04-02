package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/gvc"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// GVCService handles operations on GVCs.
type GVCService struct {
	client *Client
}

// List returns an iterator over all GVCs in the organization.
func (s *GVCService) List(ctx context.Context, org string) iter.Seq2[gvc.Gvc, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc", org)
	return listIterator[gvc.Gvc](ctx, s.client, path)
}

// ListAll returns all GVCs in the organization.
func (s *GVCService) ListAll(ctx context.Context, org string) ([]gvc.Gvc, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc", org)
	return listAll[gvc.Gvc](ctx, s.client, path)
}

// ListPage returns a single page of GVCs.
func (s *GVCService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[gvc.Gvc], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPageWithClient[gvc.Gvc](ctx, s.client, path)
}

// Get returns a GVC by name.
func (s *GVCService) Get(ctx context.Context, org, name string) (*gvc.Gvc, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s", org, name)
	var result gvc.Gvc
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new GVC.
func (s *GVCService) Create(ctx context.Context, org string, g *gvc.Gvc) (*gvc.Gvc, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc", org)
	var result gvc.Gvc
	if err := s.client.post(ctx, path, g, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing GVC.
func (s *GVCService) Update(ctx context.Context, org, name string, g *gvc.Gvc) (*gvc.Gvc, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s", org, name)
	var result gvc.Gvc
	if err := s.client.patch(ctx, path, g, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a GVC by name.
func (s *GVCService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over GVCs matching the query.
func (s *GVCService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[gvc.Gvc, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/-query", org)
	return queryIterator[gvc.Gvc](ctx, s.client, path, q)
}

// QueryAll returns all GVCs matching the query.
func (s *GVCService) QueryAll(ctx context.Context, org string, q *query.Query) ([]gvc.Gvc, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/-query", org)
	return queryAll[gvc.Gvc](ctx, s.client, path, q)
}

// QueryPage returns a single page of GVCs matching the query.
func (s *GVCService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[gvc.Gvc], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/-query", org)
	return queryPageWithClient[gvc.Gvc](ctx, s.client, path, q)
}
