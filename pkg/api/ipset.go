package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/ipSet"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// IPSetService handles operations on IP sets.
type IPSetService struct {
	client *Client
}

// List returns an iterator over all IP sets in the organization.
func (s *IPSetService) List(ctx context.Context, org string) iter.Seq2[ipSet.IpSet, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset", org)
	return listIterator[ipSet.IpSet](ctx, s.client, path)
}

// ListAll returns all IP sets in the organization.
func (s *IPSetService) ListAll(ctx context.Context, org string) ([]ipSet.IpSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset", org)
	return listAll[ipSet.IpSet](ctx, s.client, path)
}

// ListPage returns a single page of IP sets.
func (s *IPSetService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[ipSet.IpSet], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPageWithClient[ipSet.IpSet](ctx, s.client, path)
}

// Get returns an IP set by name.
func (s *IPSetService) Get(ctx context.Context, org, name string) (*ipSet.IpSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset/%s", org, name)
	var result ipSet.IpSet
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new IP set.
func (s *IPSetService) Create(ctx context.Context, org string, i *ipSet.IpSet) (*ipSet.IpSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset", org)
	var result ipSet.IpSet
	if err := s.client.post(ctx, path, i, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing IP set.
func (s *IPSetService) Update(ctx context.Context, org, name string, i *ipSet.IpSet) (*ipSet.IpSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset/%s", org, name)
	var result ipSet.IpSet
	if err := s.client.patch(ctx, path, i, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an IP set by name.
func (s *IPSetService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over IP sets matching the query.
func (s *IPSetService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[ipSet.IpSet, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset/-query", org)
	return queryIterator[ipSet.IpSet](ctx, s.client, path, q)
}

// QueryAll returns all IP sets matching the query.
func (s *IPSetService) QueryAll(ctx context.Context, org string, q *query.Query) ([]ipSet.IpSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset/-query", org)
	return queryAll[ipSet.IpSet](ctx, s.client, path, q)
}

// QueryPage returns a single page of IP sets matching the query.
func (s *IPSetService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[ipSet.IpSet], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/ipset/-query", org)
	return queryPageWithClient[ipSet.IpSet](ctx, s.client, path, q)
}
