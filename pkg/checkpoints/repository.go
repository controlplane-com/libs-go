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
	// EnforceMinimumCheckpointTime advances lingering checkpoints to the minimum, except the rows
	// named in excludeNames — rows a caller manages itself, where a checkpoint sitting in the past
	// is a deliberate instruction to re-meter rather than a leftover.
	EnforceMinimumCheckpointTime(minimum time.Time, excludeNames []string) error
}
