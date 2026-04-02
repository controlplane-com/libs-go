package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/quota"
)

// QuotaService handles operations on quotas.
type QuotaService struct {
	client *Client
}

// List returns an iterator over all quotas in the organization.
func (s *QuotaService) List(ctx context.Context, org string) iter.Seq2[quota.Quota, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota", org)
	return listIterator[quota.Quota](ctx, s.client, path)
}

// ListAll returns all quotas in the organization.
func (s *QuotaService) ListAll(ctx context.Context, org string) ([]quota.Quota, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota", org)
	return listAll[quota.Quota](ctx, s.client, path)
}

// ListPage returns a single page of quotas.
func (s *QuotaService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[quota.Quota], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPageWithClient[quota.Quota](ctx, s.client, path)
}

// Get returns a quota by name.
func (s *QuotaService) Get(ctx context.Context, org, name string) (*quota.Quota, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota/%s", org, name)
	var result quota.Quota
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new quota.
func (s *QuotaService) Create(ctx context.Context, org string, q *quota.Quota) (*quota.Quota, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota", org)
	var result quota.Quota
	if err := s.client.post(ctx, path, q, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing quota.
func (s *QuotaService) Update(ctx context.Context, org, name string, q *quota.Quota) (*quota.Quota, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota/%s", org, name)
	var result quota.Quota
	if err := s.client.patch(ctx, path, q, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a quota by name.
func (s *QuotaService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over quotas matching the query.
func (s *QuotaService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[quota.Quota, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota/-query", org)
	return queryIterator[quota.Quota](ctx, s.client, path, q)
}

// QueryAll returns all quotas matching the query.
func (s *QuotaService) QueryAll(ctx context.Context, org string, q *query.Query) ([]quota.Quota, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota/-query", org)
	return queryAll[quota.Quota](ctx, s.client, path, q)
}

// QueryPage returns a single page of quotas matching the query.
func (s *QuotaService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[quota.Quota], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/quota/-query", org)
	return queryPageWithClient[quota.Quota](ctx, s.client, path, q)
}
