package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/auditctx"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// AuditContextService handles operations on audit contexts.
type AuditContextService struct {
	client *Client
}

// List returns an iterator over all audit contexts in the organization.
func (s *AuditContextService) List(ctx context.Context, org string) iter.Seq2[auditctx.AuditContext, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx", org)
	return listIterator[auditctx.AuditContext](ctx, s.client, path)
}

// ListAll returns all audit contexts in the organization.
func (s *AuditContextService) ListAll(ctx context.Context, org string) ([]auditctx.AuditContext, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx", org)
	return listAll[auditctx.AuditContext](ctx, s.client, path)
}

// ListPage returns a single page of audit contexts.
func (s *AuditContextService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[auditctx.AuditContext], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPageWithClient[auditctx.AuditContext](ctx, s.client, path)
}

// Get returns an audit context by name.
func (s *AuditContextService) Get(ctx context.Context, org, name string) (*auditctx.AuditContext, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx/%s", org, name)
	var result auditctx.AuditContext
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new audit context.
func (s *AuditContextService) Create(ctx context.Context, org string, a *auditctx.AuditContext) (*auditctx.AuditContext, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx", org)
	var result auditctx.AuditContext
	if err := s.client.post(ctx, path, a, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing audit context.
func (s *AuditContextService) Update(ctx context.Context, org, name string, a *auditctx.AuditContext) (*auditctx.AuditContext, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx/%s", org, name)
	var result auditctx.AuditContext
	if err := s.client.patch(ctx, path, a, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an audit context by name.
func (s *AuditContextService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over audit contexts matching the query.
func (s *AuditContextService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[auditctx.AuditContext, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx/-query", org)
	return queryIterator[auditctx.AuditContext](ctx, s.client, path, q)
}

// QueryAll returns all audit contexts matching the query.
func (s *AuditContextService) QueryAll(ctx context.Context, org string, q *query.Query) ([]auditctx.AuditContext, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx/-query", org)
	return queryAll[auditctx.AuditContext](ctx, s.client, path, q)
}

// QueryPage returns a single page of audit contexts matching the query.
func (s *AuditContextService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[auditctx.AuditContext], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/auditctx/-query", org)
	return queryPageWithClient[auditctx.AuditContext](ctx, s.client, path, q)
}
