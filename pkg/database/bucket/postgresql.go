package bucket

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/controlplane-com/libs-go/pkg/checkpoints"
	"github.com/controlplane-com/libs-go/pkg/database"
	cplnErrors "github.com/controlplane-com/libs-go/pkg/errors"
	timeUtils "github.com/controlplane-com/libs-go/pkg/time-utils"
	"gorm.io/gorm"
)

// SchemaHandler defines the interface for handling schema-specific operations
// Implementations of this interface handle creating the database schema structure,
// copying and scrubbing org data, and managing partitions.
type SchemaHandler interface {
	// EnsureSchema creates all necessary tables, indexes, and schema structures for a bucket
	EnsureSchema(db *gorm.DB, bucket *Bucket) error

	// EnsurePartitions creates time-based partitions for a bucket's tables
	EnsurePartitions(db *gorm.DB, bucket *Bucket, startTime time.Time, years int) (nextTime time.Time, err error)

	// CopyOrgData copies all data for an organization from one bucket to another within a time range
	CopyOrgData(db *gorm.DB, org *Org, startTime time.Time, endTime time.Time, oldBucket *Bucket, newBucket *Bucket) error

	// ScrubOrgData deletes all data for an organization from a bucket within a time range
	ScrubOrgData(db *gorm.DB, org *Org, startTime time.Time, endTime time.Time, bucket *Bucket) error

	// AuditBuckets executes a query to audit differences between two buckets and returns data-model-specific results
	AuditBuckets(db *gorm.DB, oldBucket *Bucket, newBucket *Bucket, startTime time.Time, endTime time.Time, orgNames []string) ([]any, error)
}

// PostgresqlPartitionedRepository is a generic repository for managing bucket-partitioned data
type PostgresqlPartitionedRepository struct {
	checkpoints.CheckpointRepository
	connection          database.Connection
	schemaHandler       SchemaHandler
	config              Config
	bucketIndexSelector BucketIndexSelector
}

// NewPostgresqlPartitionedRepository creates a new repository with a schema handler
func NewPostgresqlPartitionedRepository(config Config, schemaHandler SchemaHandler, indexSelector BucketIndexSelector) PartitionedRepository {
	checkpointRepository := checkpoints.NewPostgresqlCheckpointRepository()
	if indexSelector == nil {
		indexSelector = func(org *Org, bucketCount int) int { return ModulatedHash(org.Name, bucketCount) }
	}
	return &PostgresqlPartitionedRepository{
		CheckpointRepository: checkpointRepository,
		schemaHandler:        schemaHandler,
		config:               config,
		bucketIndexSelector:  indexSelector,
	}
}

func (r *PostgresqlPartitionedRepository) Initialize(connection database.Connection) error {
	if err := r.CheckpointRepository.Initialize(connection); err != nil {
		return err
	}
	r.connection = connection

	// Ensure base tables exist
	if err := r.ensureBaseTables(); err != nil {
		return err
	}

	return r.ensurePartitionFunction()
}

func (r *PostgresqlPartitionedRepository) ensurePartitionFunction() error {
	// Helper function to prepare partitions (shared, idempotent)
	createFunc := `
CREATE OR REPLACE FUNCTION public.prepare_partitions(tableName varchar, partitionInterval interval, startDate timestamptz,
                                              iterations int)
    RETURNS int
AS
$$
DECLARE
    stmt    text;
    suf     text;
    xnow    timestamptz;
    xfrom   timestamptz;
    xto     timestamptz;
    created integer;
BEGIN
    xnow := startDate;
    created := 0;
    xfrom := xnow;
    FOR _ IN 1..iterations
        LOOP
            xto := xfrom + partitionInterval;
            -- suffix by calendar date component of the provided timestamp (caller controls zone)
            suf := REPLACE((xfrom::date)::text, '-', '_');
            stmt := FORMAT(
                'CREATE TABLE %s%s PARTITION OF %s FOR VALUES FROM (''%s'') TO (''%s'');',
                tableName, suf, tableName, xfrom, xto
            );
            BEGIN
                EXECUTE stmt;
                created := created + 1;
            EXCEPTION
                WHEN OTHERS THEN
                --
            END;
            xfrom := xto;
        END LOOP;
    RETURN created;
END;

$$
    LANGUAGE plpgsql;`
	return r.connection.Db().Exec(createFunc).Error
}

// ensureBaseTables ensures the buckets and orgs tables exist in the public schema
func (r *PostgresqlPartitionedRepository) ensureBaseTables() error {
	sql := `
		CREATE TABLE IF NOT EXISTS buckets
		(
			id                bigint PRIMARY KEY,
			created           timestamptz DEFAULT now(),
			partitions_ending timestamptz
		);

		CREATE TABLE IF NOT EXISTS orgs
		(
			name      varchar PRIMARY KEY,
			created   timestamptz default now(),
			bucket_id bigint,
			FOREIGN KEY(bucket_id) REFERENCES buckets(id)
		);

		CREATE INDEX IF NOT EXISTS idx_orgs_01 ON orgs (bucket_id);
		CREATE INDEX IF NOT EXISTS idx_buckets_01 ON buckets (partitions_ending);
	`

	return r.connection.Db().Exec(sql).Error
}

// CreateOrg creates a new organization and assigns it to a bucket
func (r *PostgresqlPartitionedRepository) CreateOrg(orgName string) error {
	_, err := r.createOrgInternal(orgName)
	return err
}

func (r *PostgresqlPartitionedRepository) createOrgInternal(orgName string) (*Org, error) {
	org := &Org{
		Name:     orgName,
		BucketId: -1,
	}
	err := r.connection.Db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("orgs").Preload("Bucket").First(org).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if org.BucketId > -1 {
			return nil // Already assigned to a bucket
		}

		// Choose bucket using injected strategy (1..OrgBucketCount)
		chosen := r.bucketIndexSelector(org, r.config.OrgBucketCount)
		bucket, err := r.ensureBucket(tx, chosen)
		if err != nil {
			return err
		}

		if err := r.saveOrg(tx, org, bucket); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return org, nil
}

// GetOrg retrieves an organization by name
func (r *PostgresqlPartitionedRepository) GetOrg(name string) (*Org, error) {
	org := &Org{Name: name}
	if err := r.connection.DbRo().Preload("Bucket").First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cplnErrors.NewErrorCode(fmt.Sprintf("invalid org: %s", name), 404)
		}
		return nil, err
	}
	return org, nil
}

// ListOrgs lists all organizations matching a filter pattern
func (r *PostgresqlPartitionedRepository) ListOrgs(pattern string) ([]*Org, error) {
	orgs, err := r.listAllOrgs()
	if err != nil {
		return nil, err
	}
	return r.filterOrgList(orgs, pattern)
}

// FilterOrgs filters a list of organizations by a pattern
func (r *PostgresqlPartitionedRepository) FilterOrgs(orgs []*Org, filter string) ([]*Org, error) {
	return r.filterOrgList(orgs, filter)
}

// ListBucketOrgs lists all organizations in a specific bucket
func (r *PostgresqlPartitionedRepository) ListBucketOrgs(bucket *Bucket) ([]*Org, error) {
	if bucket == nil {
		return nil, nil
	}
	var orgs []*Org
	if err := r.connection.DbRo().Where(map[string]any{"bucket_id": bucket.Id}).Find(&orgs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return orgs, nil
}

// ListOrgsByName lists organizations by their names
func (r *PostgresqlPartitionedRepository) ListOrgsByName(names []string) ([]*Org, error) {
	return r.ListOrgs(fmt.Sprintf("/^(%s)$/", stringJoin(names, "|")))
}

// GetBucket retrieves a bucket by ID
func (r *PostgresqlPartitionedRepository) GetBucket(id int) (*Bucket, error) {
	if id == 0 {
		return &Bucket{Id: 0}, nil
	}
	if err := r.validateBucketId(id); err != nil {
		return nil, err
	}
	return r.ensureBucket(r.connection.Db(), id)
}

// validateBucketId checks that a bucket ID is within the valid range [1, OrgBucketCount]
func (r *PostgresqlPartitionedRepository) validateBucketId(id int) error {
	if id < 1 || id > r.config.OrgBucketCount {
		return cplnErrors.NewErrorCode(
			fmt.Sprintf("bucket id out of range. Valid bucket ids: [1,%d], got: %d", r.config.OrgBucketCount, id), 400)
	}
	return nil
}

// ListBuckets lists all buckets
func (r *PostgresqlPartitionedRepository) ListBuckets() ([]*Bucket, error) {
	var buckets []*Bucket
	if err := r.connection.Db().Preload("Orgs").Preload("Orgs.Bucket").Find(&buckets).Error; err != nil {
		return nil, err
	}
	return buckets, nil
}

// EnsureBucketPartitions ensures that partitions exist for a bucket
func (r *PostgresqlPartitionedRepository) EnsureBucketPartitions(b *Bucket, startTime time.Time, years int) error {
	return r.ensureBucketPartitions(r.connection.Db(), b, startTime, years)
}

// MoveOrgToBucket moves an organization to a different bucket
func (r *PostgresqlPartitionedRepository) MoveOrgToBucket(org *Org, bucket *Bucket) error {
	if err := r.validateBucketId(bucket.Id); err != nil {
		return err
	}
	return r.connection.Db().Transaction(func(tx *gorm.DB) error {
		if _, err := r.ensureBucket(tx, bucket.Id); err != nil {
			return err
		}
		org.BucketId = bucket.Id
		org.Bucket = bucket
		return tx.Save(org).Error
	})
}

// CopyOrgData delegates to SchemaHandler to copy org data between buckets
func (r *PostgresqlPartitionedRepository) CopyOrgData(org *Org, startTime time.Time, endTime time.Time, oldBucket *Bucket, newBucket *Bucket) error {
	if err := r.validateBucketId(oldBucket.Id); err != nil {
		return err
	}
	if err := r.validateBucketId(newBucket.Id); err != nil {
		return err
	}
	return r.connection.Db().Transaction(func(tx *gorm.DB) error {
		return r.schemaHandler.CopyOrgData(tx, org, startTime, endTime, oldBucket, newBucket)
	})
}

// ScrubOrgData delegates to SchemaHandler to scrub org data from a bucket
func (r *PostgresqlPartitionedRepository) ScrubOrgData(org *Org, startTime time.Time, endTime time.Time, oldBucket *Bucket) error {
	if err := r.validateBucketId(oldBucket.Id); err != nil {
		return err
	}
	return r.connection.Db().Transaction(func(tx *gorm.DB) error {
		return r.schemaHandler.ScrubOrgData(tx, org, startTime, endTime, oldBucket)
	})
}

// AuditBucket audits differences between two buckets
func (r *PostgresqlPartitionedRepository) AuditBucket(startTime time.Time, endTime time.Time, oldBucket *Bucket, newBucket *Bucket, orgFilter string) ([]any, error) {
	orgs, err := r.ListBucketOrgs(newBucket)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, nil
	}

	orgs, err = r.filterOrgList(orgs, orgFilter)
	if err != nil {
		return nil, err
	}

	var orgNames []string
	for _, o := range orgs {
		orgNames = append(orgNames, o.Name)
	}

	// Delegate to SchemaHandler for execution and parsing
	return r.schemaHandler.AuditBuckets(r.connection.DbRo(), oldBucket, newBucket, startTime, endTime, orgNames)
}

// Private helper methods

func (r *PostgresqlPartitionedRepository) ensureBucket(db *gorm.DB, id int) (*Bucket, error) {
	bucket := &Bucket{Id: id}
	if err := db.First(bucket).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		bucket.PartitionsEnding = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		if err := db.Save(bucket).Error; err != nil {
			return nil, err
		}
	}

	if err := r.schemaHandler.EnsureSchema(db, bucket); err != nil {
		return nil, err
	}

	if bucket.PartitionsEnding.Sub(time.Now()) > r.config.PartitionPreparationThreshold {
		return bucket, nil
	}
	if err := r.ensureBucketPartitions(db, bucket, bucket.PartitionsEnding, 1); err != nil {
		return nil, err
	}
	return bucket, nil
}

func (r *PostgresqlPartitionedRepository) ensureBucketPartitions(tx *gorm.DB, bucket *Bucket, startTime time.Time, years int) error {
	if timeUtils.IsZero(&bucket.PartitionsEnding) {
		return fmt.Errorf("bucket.PartitionsEnding has an invalid value: %s", bucket.PartitionsEnding.Format(time.RFC3339))
	}

	nextTime, err := r.schemaHandler.EnsurePartitions(tx, bucket, startTime, years)
	if err != nil {
		return err
	}

	if nextTime.After(bucket.PartitionsEnding) {
		bucket.PartitionsEnding = nextTime
		return tx.Save(&bucket).Error
	}
	return nil
}

func (r *PostgresqlPartitionedRepository) saveOrg(db *gorm.DB, org *Org, bucket *Bucket) error {
	org.BucketId = bucket.Id
	org.Bucket = bucket
	return db.Save(org).Error
}

func (r *PostgresqlPartitionedRepository) listAllOrgs() ([]*Org, error) {
	var orgs []*Org
	err := r.connection.DbRo().Preload("Bucket").Order("created").Find(&orgs).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return orgs, nil
}

func (r *PostgresqlPartitionedRepository) filterOrgList(orgs []*Org, pattern string) ([]*Org, error) {
	if pattern == "" {
		return orgs, nil
	}

	var filteredOrgs []*Org
	lenPattern := len(pattern)
	if lenPattern >= 2 && pattern[0] == '/' && pattern[lenPattern-1] == '/' {
		regex, err := regexp.CompilePOSIX(pattern[1 : lenPattern-1])
		if err != nil {
			return nil, err
		}
		for _, org := range orgs {
			if regex.MatchString(org.Name) {
				filteredOrgs = append(filteredOrgs, org)
			}
		}
		return filteredOrgs, nil
	}

	for _, org := range orgs {
		if org.Name == pattern {
			return []*Org{org}, nil
		}
	}

	return nil, nil
}

// Utility function
func stringJoin(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
