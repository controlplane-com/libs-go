package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/query"
	"github.com/controlplane-com/types-go/pkg/serviceaccount"
)

// ServiceAccountService handles operations on service accounts.
type ServiceAccountService struct {
	client *Client
}

// List returns an iterator over all service accounts in the organization.
func (s *ServiceAccountService) List(ctx context.Context, org string) iter.Seq2[serviceaccount.ServiceAccount, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount", org)
	return listIterator[serviceaccount.ServiceAccount](ctx, s.client, path)
}

// ListAll returns all service accounts in the organization.
func (s *ServiceAccountService) ListAll(ctx context.Context, org string) ([]serviceaccount.ServiceAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount", org)
	return listAll[serviceaccount.ServiceAccount](ctx, s.client, path)
}

// ListPage returns a single page of service accounts.
func (s *ServiceAccountService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[serviceaccount.ServiceAccount], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[serviceaccount.ServiceAccount](ctx, s.client, path)
}

// Get returns a service account by name.
func (s *ServiceAccountService) Get(ctx context.Context, org, name string) (*serviceaccount.ServiceAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount/%s", org, name)
	var result serviceaccount.ServiceAccount
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new service account.
func (s *ServiceAccountService) Create(ctx context.Context, org string, sa *serviceaccount.ServiceAccount) (*serviceaccount.ServiceAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount", org)
	var result serviceaccount.ServiceAccount
	if err := s.client.post(ctx, path, sa, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing service account.
func (s *ServiceAccountService) Update(ctx context.Context, org, name string, sa *serviceaccount.ServiceAccount) (*serviceaccount.ServiceAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount/%s", org, name)
	var result serviceaccount.ServiceAccount
	if err := s.client.patch(ctx, path, sa, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a service account by name.
func (s *ServiceAccountService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount/%s", org, name)
	return s.client.delete(ctx, path)
}

// AddKeyRequest is the request body for adding a key to a service account.
type AddKeyRequest struct {
	Description string `json:"description,omitempty"`
}

// AddKeyResponse is the response from adding a key to a service account.
type AddKeyResponse struct {
	Key string `json:"key"`
}

// AddKey adds a new key to a service account.
func (s *ServiceAccountService) AddKey(ctx context.Context, org, name string, req *AddKeyRequest) (*AddKeyResponse, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount/%s/-addKey", org, name)
	var result AddKeyResponse
	if err := s.client.post(ctx, path, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Query returns an iterator over service accounts matching the query.
func (s *ServiceAccountService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[serviceaccount.ServiceAccount, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount/-query", org)
	return queryIterator[serviceaccount.ServiceAccount](ctx, s.client, path, q)
}

// QueryAll returns all service accounts matching the query.
func (s *ServiceAccountService) QueryAll(ctx context.Context, org string, q *query.Query) ([]serviceaccount.ServiceAccount, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount/-query", org)
	return queryAll[serviceaccount.ServiceAccount](ctx, s.client, path, q)
}

// QueryPage returns a single page of service accounts matching the query.
func (s *ServiceAccountService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[serviceaccount.ServiceAccount], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/serviceaccount/-query", org)
	return queryPage[serviceaccount.ServiceAccount](ctx, s.client, path, q)
}
