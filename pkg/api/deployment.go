package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/deployment"
)

// DeploymentService handles operations on deployments.
// Deployments are read-only resources that represent the deployment state of workloads.
type DeploymentService struct {
	client *Client
}

// List returns an iterator over all deployments for a workload.
func (s *DeploymentService) List(ctx context.Context, org, gvc, workloadName string) iter.Seq2[deployment.Deployment, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/deployment", org, gvc, workloadName)
	return listIterator[deployment.Deployment](ctx, s.client, path)
}

// ListAll returns all deployments for a workload.
func (s *DeploymentService) ListAll(ctx context.Context, org, gvc, workloadName string) ([]deployment.Deployment, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/deployment", org, gvc, workloadName)
	return listAll[deployment.Deployment](ctx, s.client, path)
}

// ListPage returns a single page of deployments.
func (s *DeploymentService) ListPage(ctx context.Context, org, gvc, workloadName, cursor string) (*ListResponse[deployment.Deployment], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/deployment", org, gvc, workloadName)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[deployment.Deployment](ctx, s.client, path)
}

// Get returns a deployment by name.
func (s *DeploymentService) Get(ctx context.Context, org, gvc, workloadName, name string) (*deployment.Deployment, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/deployment/%s", org, gvc, workloadName, name)
	var result deployment.Deployment
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
