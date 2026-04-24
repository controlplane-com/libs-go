package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/command"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/volumeSet"
)

// VolumeSetService handles operations on volume sets.
type VolumeSetService struct {
	client *Client
}

// List returns an iterator over all volume sets in the GVC.
func (s *VolumeSetService) List(ctx context.Context, org, gvc string) iter.Seq2[volumeSet.VolumeSet, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset", org, gvc)
	return listIterator[volumeSet.VolumeSet](ctx, s.client, path)
}

// ListAll returns all volume sets in the GVC.
func (s *VolumeSetService) ListAll(ctx context.Context, org, gvc string) ([]volumeSet.VolumeSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset", org, gvc)
	return listAll[volumeSet.VolumeSet](ctx, s.client, path)
}

// ListPage returns a single page of volume sets.
func (s *VolumeSetService) ListPage(ctx context.Context, org, gvc, cursor string) (*ListResponse[volumeSet.VolumeSet], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset", org, gvc)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPageWithClient[volumeSet.VolumeSet](ctx, s.client, path)
}

// Get returns a volume set by name.
func (s *VolumeSetService) Get(ctx context.Context, org, gvc, name string) (*volumeSet.VolumeSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s", org, gvc, name)
	var result volumeSet.VolumeSet
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new volume set.
func (s *VolumeSetService) Create(ctx context.Context, org, gvc string, v *volumeSet.VolumeSet) (*volumeSet.VolumeSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset", org, gvc)
	var result volumeSet.VolumeSet
	if err := s.client.post(ctx, path, v, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing volume set.
func (s *VolumeSetService) Update(ctx context.Context, org, gvc, name string, v *volumeSet.VolumeSet) (*volumeSet.VolumeSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s", org, gvc, name)
	var result volumeSet.VolumeSet
	if err := s.client.patch(ctx, path, v, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a volume set by name.
func (s *VolumeSetService) Delete(ctx context.Context, org, gvc, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s", org, gvc, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over volume sets matching the query.
// Note: VolumeSet queries are org-level, not GVC-scoped.
func (s *VolumeSetService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[volumeSet.VolumeSet, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/volumeset/-query", org)
	return queryIterator[volumeSet.VolumeSet](ctx, s.client, path, q)
}

// QueryAll returns all volume sets matching the query.
// Note: VolumeSet queries are org-level, not GVC-scoped.
func (s *VolumeSetService) QueryAll(ctx context.Context, org string, q *query.Query) ([]volumeSet.VolumeSet, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/volumeset/-query", org)
	return queryAll[volumeSet.VolumeSet](ctx, s.client, path, q)
}

// QueryPage returns a single page of volume sets matching the query.
// Note: VolumeSet queries are org-level, not GVC-scoped.
func (s *VolumeSetService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[volumeSet.VolumeSet], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/volumeset/-query", org)
	return queryPageWithClient[volumeSet.VolumeSet](ctx, s.client, path, q)
}

// Command operations

// ListCommands returns an iterator over all commands for a volume set.
func (s *VolumeSetService) ListCommands(ctx context.Context, org, gvc, volumeSetName string) iter.Seq2[command.Command, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command", org, gvc, volumeSetName)
	return listIterator[command.Command](ctx, s.client, path)
}

// ListCommandsAll returns all commands for a volume set.
func (s *VolumeSetService) ListCommandsAll(ctx context.Context, org, gvc, volumeSetName string) ([]command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command", org, gvc, volumeSetName)
	return listAll[command.Command](ctx, s.client, path)
}

// GetCommand returns a command by ID.
func (s *VolumeSetService) GetCommand(ctx context.Context, org, gvc, volumeSetName, commandID string) (*command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command/%s", org, gvc, volumeSetName, commandID)
	var result command.Command
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCommand creates a new command for a volume set.
func (s *VolumeSetService) CreateCommand(ctx context.Context, org, gvc, volumeSetName string, cmd *command.Command) (*command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command", org, gvc, volumeSetName)
	var result command.Command
	if err := s.client.postAndFollow(ctx, path, cmd, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateCommand updates an existing command.
func (s *VolumeSetService) UpdateCommand(ctx context.Context, org, gvc, volumeSetName, commandID string, cmd *command.Command) (*command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command/%s", org, gvc, volumeSetName, commandID)
	var result command.Command
	if err := s.client.patch(ctx, path, cmd, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteCommand deletes a command by ID.
func (s *VolumeSetService) DeleteCommand(ctx context.Context, org, gvc, volumeSetName, commandID string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command/%s", org, gvc, volumeSetName, commandID)
	return s.client.delete(ctx, path)
}

// QueryCommands returns an iterator over commands for a volume set matching the query.
func (s *VolumeSetService) QueryCommands(ctx context.Context, org, gvc, volumeSetName string, q *query.Query) iter.Seq2[command.Command, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command/-query", org, gvc, volumeSetName)
	return queryIterator[command.Command](ctx, s.client, path, q)
}

// QueryCommandsAll returns all commands for a volume set matching the query.
func (s *VolumeSetService) QueryCommandsAll(ctx context.Context, org, gvc, volumeSetName string, q *query.Query) ([]command.Command, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command/-query", org, gvc, volumeSetName)
	return queryAll[command.Command](ctx, s.client, path, q)
}

// QueryCommandsPage returns a single page of commands for a volume set matching the query.
func (s *VolumeSetService) QueryCommandsPage(ctx context.Context, org, gvc, volumeSetName string, q *query.Query) (*QueryResponse[command.Command], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/gvc/%s/volumeset/%s/-command/-query", org, gvc, volumeSetName)
	return queryPageWithClient[command.Command](ctx, s.client, path, q)
}

// CreateVolumeSnapshotRequest is the request for creating a volume snapshot.
type CreateVolumeSnapshotRequest struct {
	Location               string              `json:"location"`
	VolumeIndex            float32             `json:"volumeIndex"`
	SnapshotName           string              `json:"snapshotName"`
	SnapshotExpirationDate string              `json:"snapshotExpirationDate,omitempty"`
	SnapshotTags           []map[string]string `json:"snapshotTags,omitempty"`
}

// CreateVolumeSnapshot creates a createVolumeSnapshot command.
func (s *VolumeSetService) CreateVolumeSnapshot(ctx context.Context, org, gvc, volumeSetName string, req *CreateVolumeSnapshotRequest) (*command.Command, error) {
	cmd := &command.Command{
		Type: "createVolumeSnapshot",
		Spec: command.CommandSpec{
			"location":               req.Location,
			"volumeIndex":            req.VolumeIndex,
			"snapshotName":           req.SnapshotName,
			"snapshotExpirationDate": req.SnapshotExpirationDate,
			"snapshotTags":           req.SnapshotTags,
		},
	}
	return s.CreateCommand(ctx, org, gvc, volumeSetName, cmd)
}

// RestoreVolumeRequest is the request for restoring a volume from a snapshot.
type RestoreVolumeRequest struct {
	Location     string  `json:"location"`
	VolumeIndex  float32 `json:"volumeIndex"`
	SnapshotName string  `json:"snapshotName"`
	Zone         string  `json:"zone,omitempty"`
}

// RestoreVolume creates a restoreVolume command to restore a volume from a snapshot.
func (s *VolumeSetService) RestoreVolume(ctx context.Context, org, gvc, volumeSetName string, req *RestoreVolumeRequest) (*command.Command, error) {
	cmd := &command.Command{
		Type: "restoreVolume",
		Spec: command.CommandSpec{
			"location":     req.Location,
			"volumeIndex":  req.VolumeIndex,
			"snapshotName": req.SnapshotName,
			"zone":         req.Zone,
		},
	}
	return s.CreateCommand(ctx, org, gvc, volumeSetName, cmd)
}

// ExpandVolumeRequest is the request for expanding a volume.
type ExpandVolumeRequest struct {
	Location           string  `json:"location"`
	VolumeIndex        float32 `json:"volumeIndex"`
	NewStorageCapacity float32 `json:"newStorageCapacity"`
	TimeoutSeconds     float32 `json:"timeoutSeconds,omitempty"`
}

// ExpandVolume creates an expandVolume command to increase volume capacity.
func (s *VolumeSetService) ExpandVolume(ctx context.Context, org, gvc, volumeSetName string, req *ExpandVolumeRequest) (*command.Command, error) {
	cmd := &command.Command{
		Type: "expandVolume",
		Spec: command.CommandSpec{
			"location":           req.Location,
			"volumeIndex":        req.VolumeIndex,
			"newStorageCapacity": req.NewStorageCapacity,
			"timeoutSeconds":     req.TimeoutSeconds,
		},
	}
	return s.CreateCommand(ctx, org, gvc, volumeSetName, cmd)
}

// DeleteVolumeRequest is the request for deleting a volume.
type DeleteVolumeRequest struct {
	Location    string  `json:"location"`
	VolumeIndex float32 `json:"volumeIndex"`
}

// DeleteVolume creates a deleteVolume command to delete a volume.
func (s *VolumeSetService) DeleteVolume(ctx context.Context, org, gvc, volumeSetName string, req *DeleteVolumeRequest) (*command.Command, error) {
	cmd := &command.Command{
		Type: "deleteVolume",
		Spec: command.CommandSpec{
			"location":    req.Location,
			"volumeIndex": req.VolumeIndex,
		},
	}
	return s.CreateCommand(ctx, org, gvc, volumeSetName, cmd)
}

// DeleteVolumeSnapshotRequest is the request for deleting a volume snapshot.
type DeleteVolumeSnapshotRequest struct {
	Location     string  `json:"location"`
	VolumeIndex  float32 `json:"volumeIndex"`
	SnapshotName string  `json:"snapshotName"`
}

// DeleteVolumeSnapshot creates a deleteVolumeSnapshot command to delete a volume snapshot.
func (s *VolumeSetService) DeleteVolumeSnapshot(ctx context.Context, org, gvc, volumeSetName string, req *DeleteVolumeSnapshotRequest) (*command.Command, error) {
	cmd := &command.Command{
		Type: "deleteVolumeSnapshot",
		Spec: command.CommandSpec{
			"location":     req.Location,
			"volumeIndex":  req.VolumeIndex,
			"snapshotName": req.SnapshotName,
		},
	}
	return s.CreateCommand(ctx, org, gvc, volumeSetName, cmd)
}
