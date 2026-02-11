package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/policy"
	"github.com/controlplane-com/types-go/pkg/query"
)

// PolicyService handles operations on policies.
type PolicyService struct {
	client *Client
}

// List returns an iterator over all policies in the organization.
func (s *PolicyService) List(ctx context.Context, org string) iter.Seq2[policy.Policy, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy", org)
	return listIterator[policy.Policy](ctx, s.client, path)
}

// ListAll returns all policies in the organization.
func (s *PolicyService) ListAll(ctx context.Context, org string) ([]policy.Policy, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy", org)
	return listAll[policy.Policy](ctx, s.client, path)
}

// ListPage returns a single page of policies.
func (s *PolicyService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[policy.Policy], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[policy.Policy](ctx, s.client, path)
}

// Get returns a policy by name.
func (s *PolicyService) Get(ctx context.Context, org, name string) (*policy.Policy, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy/%s", org, name)
	var result policy.Policy
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new policy.
func (s *PolicyService) Create(ctx context.Context, org string, p *policy.Policy) (*policy.Policy, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy", org)
	var result policy.Policy
	if err := s.client.post(ctx, path, p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing policy.
func (s *PolicyService) Update(ctx context.Context, org, name string, p *policy.Policy) (*policy.Policy, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy/%s", org, name)
	var result policy.Policy
	if err := s.client.patch(ctx, path, p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a policy by name.
func (s *PolicyService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over policies matching the query.
func (s *PolicyService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[policy.Policy, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy/-query", org)
	return queryIterator[policy.Policy](ctx, s.client, path, q)
}

// QueryAll returns all policies matching the query.
func (s *PolicyService) QueryAll(ctx context.Context, org string, q *query.Query) ([]policy.Policy, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy/-query", org)
	return queryAll[policy.Policy](ctx, s.client, path, q)
}

// QueryPage returns a single page of policies matching the query.
func (s *PolicyService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[policy.Policy], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/policy/-query", org)
	return queryPage[policy.Policy](ctx, s.client, path, q)
}
