package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/task"
)

// TaskService handles operations on tasks.
type TaskService struct {
	client *Client
}

// List returns an iterator over all tasks in the organization.
func (s *TaskService) List(ctx context.Context, org string) iter.Seq2[task.Task, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/task", org)
	return listIterator[task.Task](ctx, s.client, path)
}

// ListAll returns all tasks in the organization.
func (s *TaskService) ListAll(ctx context.Context, org string) ([]task.Task, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/task", org)
	return listAll[task.Task](ctx, s.client, path)
}

// ListPage returns a single page of tasks.
func (s *TaskService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[task.Task], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/task", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[task.Task](ctx, s.client, path)
}

// Get returns a task by name.
func (s *TaskService) Get(ctx context.Context, org, name string) (*task.Task, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/task/%s", org, name)
	var result task.Task
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Query returns an iterator over tasks matching the query.
func (s *TaskService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[task.Task, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/task/-query", org)
	return queryIterator[task.Task](ctx, s.client, path, q)
}

// QueryAll returns all tasks matching the query.
func (s *TaskService) QueryAll(ctx context.Context, org string, q *query.Query) ([]task.Task, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/task/-query", org)
	return queryAll[task.Task](ctx, s.client, path, q)
}

// QueryPage returns a single page of tasks matching the query.
func (s *TaskService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[task.Task], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/task/-query", org)
	return queryPage[task.Task](ctx, s.client, path, q)
}
