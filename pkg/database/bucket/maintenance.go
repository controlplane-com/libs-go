package bucket

import (
	"time"

	leaderElection "github.com/controlplane-com/libs-go/pkg/leader-election"
	"github.com/controlplane-com/libs-go/pkg/logging"
	timeUtils "github.com/controlplane-com/libs-go/pkg/time-utils"
)

// BucketMaintenanceOptions configures the bucket maintenance loop
type BucketMaintenanceOptions struct {
	Repository                    PartitionedRepository
	LeaderElectorClass            string
	MaintenanceFrequency          time.Duration
	PartitionPreparationThreshold time.Duration
}

// RunBucketMaintenanceLoop continuously ensures bucket partitions are prepared ahead of time.
// This function runs indefinitely and should be called in a goroutine.
// Only the leader replica will perform maintenance work; followers will sleep and check periodically.
func RunBucketMaintenanceLoop(opts BucketMaintenanceOptions) {
	logger := logging.Logger().Sugar()
	elector := leaderElection.LeaderElectorFromConfig(opts.LeaderElectorClass)

	for {
		if role := elector.Role(); role != leaderElection.TypeLeader {
			// Check once per minute to see whether this replica has become the leader
			time.Sleep(time.Minute)
			continue
		}

		ensureBucketPartitions(opts.Repository, opts.PartitionPreparationThreshold, logger)
		time.Sleep(opts.MaintenanceFrequency)
	}
}

func ensureBucketPartitions(repository PartitionedRepository, threshold time.Duration, logger interface{ Errorf(string, ...interface{}) }) {
	buckets, err := repository.ListBuckets()
	if err != nil {
		logger.Errorf("Error listing the buckets: %v", err)
		return
	}

	for _, b := range buckets {
		if timeUtils.IsZero(&b.PartitionsEnding) {
			b.PartitionsEnding = time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		} else if b.PartitionsEnding.Sub(time.Now()) > threshold {
			continue
		}
		if err := repository.EnsureBucketPartitions(b, b.PartitionsEnding, 1); err != nil {
			logger.Errorf("Error ensuring bucket partitions: %v", err)
			continue
		}
	}
}
