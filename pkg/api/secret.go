package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/query"
	"github.com/controlplane-com/types-go/pkg/secret"
)

// SecretService handles operations on secrets.
type SecretService struct {
	client *Client
}

// List returns an iterator over all secrets in the organization.
func (s *SecretService) List(ctx context.Context, org string) iter.Seq2[secret.Secret, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret", org)
	return listIterator[secret.Secret](ctx, s.client, path)
}

// ListAll returns all secrets in the organization.
func (s *SecretService) ListAll(ctx context.Context, org string) ([]secret.Secret, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret", org)
	return listAll[secret.Secret](ctx, s.client, path)
}

// ListPage returns a single page of secrets.
func (s *SecretService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[secret.Secret], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[secret.Secret](ctx, s.client, path)
}

// Get returns a secret by name.
func (s *SecretService) Get(ctx context.Context, org, name string) (*secret.Secret, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret/%s", org, name)
	var result secret.Secret
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new secret.
func (s *SecretService) Create(ctx context.Context, org string, sec *secret.Secret) (*secret.Secret, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret", org)
	var result secret.Secret
	if err := s.client.post(ctx, path, sec, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing secret.
func (s *SecretService) Update(ctx context.Context, org, name string, sec *secret.Secret) (*secret.Secret, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret/%s", org, name)
	var result secret.Secret
	if err := s.client.patch(ctx, path, sec, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a secret by name.
func (s *SecretService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret/%s", org, name)
	return s.client.delete(ctx, path)
}

// Reveal returns the secret with its data revealed.
// This requires the 'reveal' permission on the secret.
func (s *SecretService) Reveal(ctx context.Context, org, name string) (*secret.Secret, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret/%s/-reveal", org, name)
	var result secret.Secret
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Query returns an iterator over secrets matching the query.
func (s *SecretService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[secret.Secret, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret/-query", org)
	return queryIterator[secret.Secret](ctx, s.client, path, q)
}

// QueryAll returns all secrets matching the query.
func (s *SecretService) QueryAll(ctx context.Context, org string, q *query.Query) ([]secret.Secret, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret/-query", org)
	return queryAll[secret.Secret](ctx, s.client, path, q)
}

// QueryPage returns a single page of secrets matching the query.
func (s *SecretService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[secret.Secret], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/secret/-query", org)
	return queryPage[secret.Secret](ctx, s.client, path, q)
}
