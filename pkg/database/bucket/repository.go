package bucket

import (
	"time"

	"github.com/controlplane-com/libs-go/pkg/checkpoints"
	"github.com/controlplane-com/libs-go/pkg/database"
)

// BucketIndexSelector selects the target bucket ID (1..bucketCount) for a given org when creating/assigning it.
// Implementations must return a value in the inclusive range [1, bucketCount].
type BucketIndexSelector func(org *Org, bucketCount int) int

// PartitionedRepository is a generalized interface for managing partitioned data across organization buckets.
// It provides methods for org/bucket management, and schema-agnostic operations.
type PartitionedRepository interface {
	checkpoints.CheckpointRepository
	database.Repository

	// Org Management
	CreateOrg(org string) error
	GetOrg(name string) (*Org, error)
	ListOrgs(filter string) ([]*Org, error)
	FilterOrgs(orgs []*Org, filter string) ([]*Org, error)
	ListBucketOrgs(bucket *Bucket) ([]*Org, error)
	ListOrgsByName(names []string) ([]*Org, error)

	// Bucket Management
	GetBucket(id int) (*Bucket, error)
	ListBuckets() ([]*Bucket, error)
	EnsureBucketPartitions(b *Bucket, startTime time.Time, years int) error
	MoveOrgToBucket(org *Org, bucket *Bucket) error

	// Data Operations
	CopyOrgData(org *Org, startTime time.Time, endTime time.Time, oldBucket *Bucket, newBucket *Bucket) error
	ScrubOrgData(org *Org, startTime time.Time, endTime time.Time, oldBucket *Bucket) error
	AuditBucket(startTime time.Time, endTime time.Time, oldBucket *Bucket, newBucket *Bucket, orgFilter string) ([]any, error)
}

// Config holds configuration for PartitionedRepository
type Config struct {
	OrgBucketCount                int
	PartitionPreparationThreshold time.Duration
}
