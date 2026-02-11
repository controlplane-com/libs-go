package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/agent"
	"github.com/controlplane-com/types-go/pkg/query"
)

// AgentService handles operations on agents.
type AgentService struct {
	client *Client
}

// List returns an iterator over all agents in the organization.
func (s *AgentService) List(ctx context.Context, org string) iter.Seq2[agent.Agent, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent", org)
	return listIterator[agent.Agent](ctx, s.client, path)
}

// ListAll returns all agents in the organization.
func (s *AgentService) ListAll(ctx context.Context, org string) ([]agent.Agent, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent", org)
	return listAll[agent.Agent](ctx, s.client, path)
}

// ListPage returns a single page of agents.
func (s *AgentService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[agent.Agent], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[agent.Agent](ctx, s.client, path)
}

// Get returns an agent by name.
func (s *AgentService) Get(ctx context.Context, org, name string) (*agent.Agent, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent/%s", org, name)
	var result agent.Agent
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new agent.
func (s *AgentService) Create(ctx context.Context, org string, a *agent.Agent) (*agent.Agent, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent", org)
	var result agent.Agent
	if err := s.client.post(ctx, path, a, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing agent.
func (s *AgentService) Update(ctx context.Context, org, name string, a *agent.Agent) (*agent.Agent, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent/%s", org, name)
	var result agent.Agent
	if err := s.client.patch(ctx, path, a, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an agent by name.
func (s *AgentService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over agents matching the query.
func (s *AgentService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[agent.Agent, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent/-query", org)
	return queryIterator[agent.Agent](ctx, s.client, path, q)
}

// QueryAll returns all agents matching the query.
func (s *AgentService) QueryAll(ctx context.Context, org string, q *query.Query) ([]agent.Agent, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent/-query", org)
	return queryAll[agent.Agent](ctx, s.client, path, q)
}

// QueryPage returns a single page of agents matching the query.
func (s *AgentService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[agent.Agent], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/agent/-query", org)
	return queryPage[agent.Agent](ctx, s.client, path, q)
}
