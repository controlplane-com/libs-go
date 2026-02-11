package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/cloudaccount"
	"github.com/controlplane-com/types-go/pkg/query"
)

// CloudAccountService handles operations on cloud accounts.
type CloudAccountService struct {
	client *Client
}

// List returns an iterator over all cloud accounts in the organization.
func (s *CloudAccountService) List(ctx context.Context, org string) iter.Seq2[cloudaccount.CloudAccount, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount", org)
	return listIterator[cloudaccount.CloudAccount](ctx, s.client, path)
}

// ListAll returns all cloud accounts in the organization.
func (s *CloudAccountService) ListAll(ctx context.Context, org string) ([]cloudaccount.CloudAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount", org)
	return listAll[cloudaccount.CloudAccount](ctx, s.client, path)
}

// ListPage returns a single page of cloud accounts.
func (s *CloudAccountService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[cloudaccount.CloudAccount], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[cloudaccount.CloudAccount](ctx, s.client, path)
}

// Get returns a cloud account by name.
func (s *CloudAccountService) Get(ctx context.Context, org, name string) (*cloudaccount.CloudAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount/%s", org, name)
	var result cloudaccount.CloudAccount
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new cloud account.
func (s *CloudAccountService) Create(ctx context.Context, org string, ca *cloudaccount.CloudAccount) (*cloudaccount.CloudAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount", org)
	var result cloudaccount.CloudAccount
	if err := s.client.post(ctx, path, ca, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing cloud account.
func (s *CloudAccountService) Update(ctx context.Context, org, name string, ca *cloudaccount.CloudAccount) (*cloudaccount.CloudAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount/%s", org, name)
	var result cloudaccount.CloudAccount
	if err := s.client.patch(ctx, path, ca, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a cloud account by name.
func (s *CloudAccountService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over cloud accounts matching the query.
func (s *CloudAccountService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[cloudaccount.CloudAccount, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount/-query", org)
	return queryIterator[cloudaccount.CloudAccount](ctx, s.client, path, q)
}

// QueryAll returns all cloud accounts matching the query.
func (s *CloudAccountService) QueryAll(ctx context.Context, org string, q *query.Query) ([]cloudaccount.CloudAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount/-query", org)
	return queryAll[cloudaccount.CloudAccount](ctx, s.client, path, q)
}

// QueryPage returns a single page of cloud accounts matching the query.
func (s *CloudAccountService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[cloudaccount.CloudAccount], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/cloudaccount/-query", org)
	return queryPage[cloudaccount.CloudAccount](ctx, s.client, path, q)
}
