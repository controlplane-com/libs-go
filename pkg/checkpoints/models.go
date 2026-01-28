package checkpoints

import (
	"time"
)

const DefaultCheckpointName = "default"

type QueryCheckpointsResult struct {
	Checkpoints []*Checkpoint `json:"checkpoints"`
}

type ResetCheckpointsRequest struct {
	NewCheckpointTime time.Time `json:"newCheckpointTime"`
	CheckpointQuery
}

type CheckpointQuery struct {
	Name string `json:"name,omitempty"`
	Job  string `json:"job,omitempty"`
}

type ListCheckpointsResult struct {
	Checkpoints []*Checkpoint `json:"checkpoints"`
}

type QueryCheckpointsRequest struct {
	CheckpointQuery
}

type Checkpoint struct {
	Name              string    `json:"name" gorm:"primaryKey;autoIncrement:false"`
	Job               string    `json:"job" gorm:"primaryKey;autoIncrement:false"`
	LastCompletedTime time.Time `json:"lastCompletedTime" gorm:"type:timestamptz"`
}
