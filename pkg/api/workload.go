package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/types-go/pkg/command"
	"github.com/controlplane-com/types-go/pkg/query"
	"github.com/controlplane-com/types-go/pkg/workload"
)

// WorkloadService handles operations on workloads.
type WorkloadService struct {
	client *Client
}

// List returns an iterator over all workloads in the GVC.
func (s *WorkloadService) List(ctx context.Context, org, gvc string) iter.Seq2[workload.Workload, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload", org, gvc)
	return listIterator[workload.Workload](ctx, s.client, path)
}

// ListAll returns all workloads in the GVC.
func (s *WorkloadService) ListAll(ctx context.Context, org, gvc string) ([]workload.Workload, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload", org, gvc)
	return listAll[workload.Workload](ctx, s.client, path)
}

// ListPage returns a single page of workloads.
func (s *WorkloadService) ListPage(ctx context.Context, org, gvc, cursor string) (*ListResponse[workload.Workload], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload", org, gvc)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPage[workload.Workload](ctx, s.client, path)
}

// Get returns a workload by name.
func (s *WorkloadService) Get(ctx context.Context, org, gvc, name string) (*workload.Workload, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s", org, gvc, name)
	var result workload.Workload
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new workload.
func (s *WorkloadService) Create(ctx context.Context, org, gvc string, w *workload.Workload) (*workload.Workload, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload", org, gvc)
	var result workload.Workload
	if err := s.client.post(ctx, path, w, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing workload.
func (s *WorkloadService) Update(ctx context.Context, org, gvc, name string, w *workload.Workload) (*workload.Workload, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s", org, gvc, name)
	var result workload.Workload
	if err := s.client.patch(ctx, path, w, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a workload by name.
func (s *WorkloadService) Delete(ctx context.Context, org, gvc, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s", org, gvc, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over workloads matching the query.
func (s *WorkloadService) Query(ctx context.Context, org, gvc string, q *query.Query) iter.Seq2[workload.Workload, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/-query", org, gvc)
	return queryIterator[workload.Workload](ctx, s.client, path, q)
}

// QueryAll returns all workloads matching the query.
func (s *WorkloadService) QueryAll(ctx context.Context, org, gvc string, q *query.Query) ([]workload.Workload, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/-query", org, gvc)
	return queryAll[workload.Workload](ctx, s.client, path, q)
}

// QueryPage returns a single page of workloads matching the query.
func (s *WorkloadService) QueryPage(ctx context.Context, org, gvc string, q *query.Query) (*QueryResponse[workload.Workload], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/-query", org, gvc)
	return queryPage[workload.Workload](ctx, s.client, path, q)
}

// ForceRedeployRequest is the request body for forcing a workload redeployment.
type ForceRedeployRequest struct{}

// ForceRedeploy forces a workload to redeploy.
func (s *WorkloadService) ForceRedeploy(ctx context.Context, org, gvc, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/-forceRedeploy", org, gvc, name)
	return s.client.post(ctx, path, &ForceRedeployRequest{}, nil)
}

// CronStartRequest is the request body for starting a cron workload job.
type CronStartRequest struct {
	Location string `json:"location,omitempty"`
}

// CronStart manually triggers a cron workload job.
func (s *WorkloadService) CronStart(ctx context.Context, org, gvc, name string, req *CronStartRequest) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/-cronStart", org, gvc, name)
	return s.client.post(ctx, path, req, nil)
}

// Command operations

// ListCommands returns an iterator over all commands for a workload.
func (s *WorkloadService) ListCommands(ctx context.Context, org, gvc, workloadName string) iter.Seq2[command.Command, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/command", org, gvc, workloadName)
	return listIterator[command.Command](ctx, s.client, path)
}

// ListCommandsAll returns all commands for a workload.
func (s *WorkloadService) ListCommandsAll(ctx context.Context, org, gvc, workloadName string) ([]command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/command", org, gvc, workloadName)
	return listAll[command.Command](ctx, s.client, path)
}

// GetCommand returns a command by ID.
func (s *WorkloadService) GetCommand(ctx context.Context, org, gvc, workloadName, commandID string) (*command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/command/%s", org, gvc, workloadName, commandID)
	var result command.Command
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCommand creates a new command for a workload.
func (s *WorkloadService) CreateCommand(ctx context.Context, org, gvc, workloadName string, cmd *command.Command) (*command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/command", org, gvc, workloadName)
	var result command.Command
	if err := s.client.post(ctx, path, cmd, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateCommand updates an existing command.
func (s *WorkloadService) UpdateCommand(ctx context.Context, org, gvc, workloadName, commandID string, cmd *command.Command) (*command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/command/%s", org, gvc, workloadName, commandID)
	var result command.Command
	if err := s.client.patch(ctx, path, cmd, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteCommand deletes a command by ID.
func (s *WorkloadService) DeleteCommand(ctx context.Context, org, gvc, workloadName, commandID string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/command/%s", org, gvc, workloadName, commandID)
	return s.client.delete(ctx, path)
}

// QueryCommands returns an iterator over commands for a workload matching the query.
func (s *WorkloadService) QueryCommands(ctx context.Context, org, gvc, workloadName string, q *query.Query) iter.Seq2[command.Command, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/-command/-query", org, gvc, workloadName)
	return queryIterator[command.Command](ctx, s.client, path, q)
}

// QueryCommandsAll returns all commands for a workload matching the query.
func (s *WorkloadService) QueryCommandsAll(ctx context.Context, org, gvc, workloadName string, q *query.Query) ([]command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/-command/-query", org, gvc, workloadName)
	return queryAll[command.Command](ctx, s.client, path, q)
}

// QueryCommandsPage returns a single page of commands for a workload matching the query.
func (s *WorkloadService) QueryCommandsPage(ctx context.Context, org, gvc, workloadName string, q *query.Query) (*QueryResponse[command.Command], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/workload/%s/-command/-query", org, gvc, workloadName)
	return queryPage[command.Command](ctx, s.client, path, q)
}

// RunCronWorkload creates a runCronWorkload command to manually trigger a cron workload.
func (s *WorkloadService) RunCronWorkload(ctx context.Context, org, gvc, workloadName string, spec *command.RunCronWorkloadSpec) (*command.Command, error) {
	cmd := &command.Command{
		Type: "runCronWorkload",
		Spec: command.CommandSpec{
			"location":           spec.Location,
			"containerOverrides": spec.ContainerOverrides,
		},
	}
	return s.CreateCommand(ctx, org, gvc, workloadName, cmd)
}

// StopReplica creates a stopReplica command to stop a specific replica.
func (s *WorkloadService) StopReplica(ctx context.Context, org, gvc, workloadName string, spec *command.StopReplicaSpec) (*command.Command, error) {
	cmd := &command.Command{
		Type: "stopReplica",
		Spec: command.CommandSpec{
			"location": spec.Location,
			"replica":  spec.Replica,
		},
	}
	return s.CreateCommand(ctx, org, gvc, workloadName, cmd)
}
