package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/domain"
	"github.com/controlplane-com/types-go/pkg/query"
)

// DomainService handles operations on domains.
type DomainService struct {
	client *Client
}

// List returns an iterator over all domains in the organization.
func (s *DomainService) List(ctx context.Context, org string) iter.Seq2[domain.Domain, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain", org)
	return listIterator[domain.Domain](ctx, s.client, path)
}

// ListAll returns all domains in the organization.
func (s *DomainService) ListAll(ctx context.Context, org string) ([]domain.Domain, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain", org)
	return listAll[domain.Domain](ctx, s.client, path)
}

// ListPage returns a single page of domains.
func (s *DomainService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[domain.Domain], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[domain.Domain](ctx, s.client, path)
}

// Get returns a domain by name.
func (s *DomainService) Get(ctx context.Context, org, name string) (*domain.Domain, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain/%s", org, name)
	var result domain.Domain
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new domain.
func (s *DomainService) Create(ctx context.Context, org string, d *domain.Domain) (*domain.Domain, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain", org)
	var result domain.Domain
	if err := s.client.post(ctx, path, d, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing domain.
func (s *DomainService) Update(ctx context.Context, org, name string, d *domain.Domain) (*domain.Domain, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain/%s", org, name)
	var result domain.Domain
	if err := s.client.patch(ctx, path, d, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a domain by name.
func (s *DomainService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over domains matching the query.
func (s *DomainService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[domain.Domain, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain/-query", org)
	return queryIterator[domain.Domain](ctx, s.client, path, q)
}

// QueryAll returns all domains matching the query.
func (s *DomainService) QueryAll(ctx context.Context, org string, q *query.Query) ([]domain.Domain, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain/-query", org)
	return queryAll[domain.Domain](ctx, s.client, path, q)
}

// QueryPage returns a single page of domains matching the query.
func (s *DomainService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[domain.Domain], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/domain/-query", org)
	return queryPage[domain.Domain](ctx, s.client, path, q)
}
