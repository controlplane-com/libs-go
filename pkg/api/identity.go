package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/identity"
	"github.com/controlplane-com/types-go/pkg/query"
)

// IdentityService handles operations on identities.
type IdentityService struct {
	client *Client
}

// List returns an iterator over all identities in the GVC.
func (s *IdentityService) List(ctx context.Context, org, gvc string) iter.Seq2[identity.Identity, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity", org, gvc)
	return listIterator[identity.Identity](ctx, s.client, path)
}

// ListAll returns all identities in the GVC.
func (s *IdentityService) ListAll(ctx context.Context, org, gvc string) ([]identity.Identity, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity", org, gvc)
	return listAll[identity.Identity](ctx, s.client, path)
}

// ListPage returns a single page of identities.
func (s *IdentityService) ListPage(ctx context.Context, org, gvc, cursor string) (*ListResponse[identity.Identity], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity", org, gvc)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[identity.Identity](ctx, s.client, path)
}

// Get returns an identity by name.
func (s *IdentityService) Get(ctx context.Context, org, gvc, name string) (*identity.Identity, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity/%s", org, gvc, name)
	var result identity.Identity
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new identity.
func (s *IdentityService) Create(ctx context.Context, org, gvc string, i *identity.Identity) (*identity.Identity, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity", org, gvc)
	var result identity.Identity
	if err := s.client.post(ctx, path, i, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing identity.
func (s *IdentityService) Update(ctx context.Context, org, gvc, name string, i *identity.Identity) (*identity.Identity, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity/%s", org, gvc, name)
	var result identity.Identity
	if err := s.client.patch(ctx, path, i, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an identity by name.
func (s *IdentityService) Delete(ctx context.Context, org, gvc, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity/%s", org, gvc, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over identities matching the query.
func (s *IdentityService) Query(ctx context.Context, org, gvc string, q *query.Query) iter.Seq2[identity.Identity, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity/-query", org, gvc)
	return queryIterator[identity.Identity](ctx, s.client, path, q)
}

// QueryAll returns all identities matching the query.
func (s *IdentityService) QueryAll(ctx context.Context, org, gvc string, q *query.Query) ([]identity.Identity, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity/-query", org, gvc)
	return queryAll[identity.Identity](ctx, s.client, path, q)
}

// QueryPage returns a single page of identities matching the query.
func (s *IdentityService) QueryPage(ctx context.Context, org, gvc string, q *query.Query) (*QueryResponse[identity.Identity], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/identity/-query", org, gvc)
	return queryPage[identity.Identity](ctx, s.client, path, q)
}
