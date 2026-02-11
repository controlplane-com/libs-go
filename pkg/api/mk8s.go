package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/mk8s"
	"github.com/controlplane-com/types-go/pkg/query"
)

// Mk8sService handles operations on managed Kubernetes clusters.
type Mk8sService struct {
	client *Client
}

// List returns an iterator over all managed Kubernetes clusters in the organization.
func (s *Mk8sService) List(ctx context.Context, org string) iter.Seq2[mk8s.Mk8sCluster, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s", org)
	return listIterator[mk8s.Mk8sCluster](ctx, s.client, path)
}

// ListAll returns all managed Kubernetes clusters in the organization.
func (s *Mk8sService) ListAll(ctx context.Context, org string) ([]mk8s.Mk8sCluster, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s", org)
	return listAll[mk8s.Mk8sCluster](ctx, s.client, path)
}

// ListPage returns a single page of managed Kubernetes clusters.
func (s *Mk8sService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[mk8s.Mk8sCluster], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[mk8s.Mk8sCluster](ctx, s.client, path)
}

// Get returns a managed Kubernetes cluster by name.
func (s *Mk8sService) Get(ctx context.Context, org, name string) (*mk8s.Mk8sCluster, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s/%s", org, name)
	var result mk8s.Mk8sCluster
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new managed Kubernetes cluster.
func (s *Mk8sService) Create(ctx context.Context, org string, m *mk8s.Mk8sCluster) (*mk8s.Mk8sCluster, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s", org)
	var result mk8s.Mk8sCluster
	if err := s.client.post(ctx, path, m, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing managed Kubernetes cluster.
func (s *Mk8sService) Update(ctx context.Context, org, name string, m *mk8s.Mk8sCluster) (*mk8s.Mk8sCluster, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s/%s", org, name)
	var result mk8s.Mk8sCluster
	if err := s.client.patch(ctx, path, m, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a managed Kubernetes cluster by name.
func (s *Mk8sService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over managed Kubernetes clusters matching the query.
func (s *Mk8sService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[mk8s.Mk8sCluster, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s/-query", org)
	return queryIterator[mk8s.Mk8sCluster](ctx, s.client, path, q)
}

// QueryAll returns all managed Kubernetes clusters matching the query.
func (s *Mk8sService) QueryAll(ctx context.Context, org string, q *query.Query) ([]mk8s.Mk8sCluster, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s/-query", org)
	return queryAll[mk8s.Mk8sCluster](ctx, s.client, path, q)
}

// QueryPage returns a single page of managed Kubernetes clusters matching the query.
func (s *Mk8sService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[mk8s.Mk8sCluster], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/mk8s/-query", org)
	return queryPage[mk8s.Mk8sCluster](ctx, s.client, path, q)
}
