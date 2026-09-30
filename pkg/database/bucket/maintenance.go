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

// RunBucketMaintenanceLoop continuously ensures bucket partitions are prepared ahead of time and drops expired ones.
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

		maintainBuckets(opts.Repository, opts.PartitionPreparationThreshold, time.Now(), logger)
		time.Sleep(opts.MaintenanceFrequency)
	}
}

func maintainBuckets(repository PartitionedRepository, threshold time.Duration, now time.Time, logger interface{ Errorf(string, ...interface{}) }) {
	buckets, err := repository.ListBuckets()
	if err != nil {
		logger.Errorf("Error listing the buckets: %v", err)
		return
	}

	for _, b := range buckets {
		if err := ensureBucketPartitions(repository, b, threshold, now); err != nil {
			logger.Errorf("Error ensuring bucket partitions: %v", err)
		}
		if err := repository.DropExpiredPartitions(b, now); err != nil {
			logger.Errorf("Error dropping expired partitions of bucket %d: %v", b.Id, err)
		}
	}
}

func ensureBucketPartitions(repository PartitionedRepository, b *Bucket, threshold time.Duration, now time.Time) error {
	if timeUtils.IsZero(&b.PartitionsEnding) {
		b.PartitionsEnding = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	} else if b.PartitionsEnding.Sub(now) > threshold {
		return nil
	}
	return repository.EnsureBucketPartitions(b, b.PartitionsEnding, 1)
}
