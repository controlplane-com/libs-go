package checkpoints

import (
	"github.com/controlplane-com/libs-go/pkg/database"
	"time"
)

type CheckpointRepository interface {
	database.Repository
	SetCheckpoint(checkpoint *Checkpoint) error
	ListCheckpoints() ([]*Checkpoint, error)
	QueryCheckpoints(query *CheckpointQuery) ([]*Checkpoint, error)
	GetCheckpoint(query *CheckpointQuery) (*Checkpoint, error)

	ResetAllCheckpoints(newCheckpointTime time.Time) error
	ResetCheckpoints(request *ResetCheckpointsRequest) error
	EnforceMinimumCheckpointTime(minimum time.Time) error
}
