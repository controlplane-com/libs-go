# Bucket Partitioned Repository

This package provides a generalized, schema-agnostic interface for managing partitioned data across organization buckets in PostgreSQL.

## Architecture

The architecture separates bucket/org management from schema-specific operations through dependency injection:

- **`PartitionedRepository`**: Interface for org/bucket management and data operations
- **`SchemaHandler`**: Interface for schema-specific operations (table creation, data copying, scrubbing)
- **`PostgresqlPartitionedRepository`**: Generic implementation that delegates schema operations to a `SchemaHandler`
- **`CachingRepository`**: Optional wrapper that adds org collection caching to any `PartitionedRepository`

## Usage

### Basic Usage Example

```go
import (
    "time"
    "github.com/controlplane-com/libs-go/pkg/database/bucket"
)

// Create the config (only repository-related settings)
config := bucket.Config{
    OrgBucketCount:                256,
    PartitionPreparationThreshold: 90 * 24 * time.Hour,
}

// Create your schema handler (implementation specific)
schemaHandler := myschema.NewMyCustomSchemaHandler()

// Create the repository (pass nil to use default ModulatedHash selector)
repo := bucket.NewPostgresqlPartitionedRepository(config, schemaHandler, nil)

// Optionally, provide a custom bucket index selector
// chooser := func(org *bucket.Org, bucketCount int) int { return 1 /* always bucket 1 */ }
// repo = bucket.NewPostgresqlPartitionedRepository(config, schemaHandler, chooser)

// Initialize with database connection
err := repo.Initialize(dbConnection)
```

#### Enable caching (optional)

Wrap the repository with a CachingRepository and provide a CollectionCache implementation (for example, Redis-backed):

```go
import (
    "time"
    "github.com/controlplane-com/libs-go/pkg/cache"
    "github.com/controlplane-com/libs-go/pkg/database/bucket"
)

redisCfg := cache.RedisClientConfig{ Hosts: []string{"localhost:6379"}, Mode: cache.RedisModeStandalone }

// nameOf extracts the cache key (org name) from an Org
nameOf := func(o *bucket.Org) (string, error) { return o.Name, nil }
orgCache := cache.NewRedisCollectionCache[bucket.Org](
    redisCfg,
    100,            // batch size
    "orgs",        // key prefix
    "orgs:index",  // index key for validity TTL
    nameOf,
)

// Wrap with caching (24h validity for the index)
repo = bucket.NewCachingRepository(repo, orgCache, 24*time.Hour)
```

### Creating a Custom Schema Handler

To use this architecture with a different schema structure:

```go
package myschema

import (
    "time"
    "gorm.io/gorm"
    "github.com/controlplane-com/libs-go/pkg/database/bucket"
)

type MyCustomSchemaHandler struct {
    // Add any configuration or dependencies needed
}

func NewMyCustomSchemaHandler() *MyCustomSchemaHandler {
    return &MyCustomSchemaHandler{}
}

func (h *MyCustomSchemaHandler) EnsureSchema(db *gorm.DB, b *bucket.Bucket) error {
    schema := b.SchemaName()
    // Create your custom tables, indexes, etc.
    return db.Exec(`
        CREATE SCHEMA IF NOT EXISTS ` + schema + `;
        CREATE TABLE IF NOT EXISTS ` + schema + `.my_data (
            id SERIAL PRIMARY KEY,
            org_id VARCHAR,
            data JSONB,
            created_at TIMESTAMPTZ
        );
    `).Error
}

func (h *MyCustomSchemaHandler) EnsurePartitions(db *gorm.DB, b *bucket.Bucket, startTime time.Time, years int) (time.Time, error) {
    // Implement partition creation for your schema
    // Return the next partition time
    nextTime := startTime.AddDate(years, 0, 0)
    return nextTime, nil
}

func (h *MyCustomSchemaHandler) CopyOrgData(db *gorm.DB, org *bucket.Org, startTime time.Time, endTime time.Time, oldBucket *bucket.Bucket, newBucket *bucket.Bucket) error {
    // Implement data copying logic for your schema
    oldSchema := oldBucket.SchemaName()
    newSchema := newBucket.SchemaName()

    query := `
        INSERT INTO ` + newSchema + `.my_data (org_id, data, created_at)
        SELECT org_id, data, created_at
        FROM ` + oldSchema + `.my_data
        WHERE org_id = ? AND created_at >= ? AND created_at < ?
        ON CONFLICT DO NOTHING;
    `
    return db.Exec(query, org.Name, startTime, endTime).Error
}

func (h *MyCustomSchemaHandler) ScrubOrgData(db *gorm.DB, org *bucket.Org, startTime time.Time, endTime time.Time, b *bucket.Bucket) error {
    // Implement data deletion logic for your schema
    schema := b.SchemaName()
    query := `
        DELETE FROM ` + schema + `.my_data
        WHERE org_id = ? AND created_at >= ? AND created_at < ?;
    `
    return db.Exec(query, org.Name, startTime, endTime).Error
}

func (h *MyCustomSchemaHandler) BuildAuditQuery(oldBucket *bucket.Bucket, newBucket *bucket.Bucket, startTime time.Time, endTime time.Time) string {
    // Build a query to audit differences between buckets
    oldSchema := oldBucket.SchemaName()
    newSchema := newBucket.SchemaName()

    return `
        SELECT o.org_id, o.data as old_data, n.data as new_data
        FROM ` + oldSchema + `.my_data o
        FULL OUTER JOIN ` + newSchema + `.my_data n
            ON o.id = n.id
        WHERE o.data IS NULL OR n.data IS NULL OR o.data != n.data;
    `
}
```

The example above shows how to integrate your custom schema handler with the repository. See the "Basic Usage Example" section for repository setup, and the "Enable caching (optional)" section to add caching.

## Key Interfaces

### PartitionedRepository

Main interface for interacting with bucket-partitioned data:

- **Org Management**: `CreateOrg`, `GetOrg`, `ListOrgs`, `FilterOrgs`, `ListBucketOrgs`, `ListOrgsByName`
- **Bucket Management**: `GetBucket`, `ListBuckets`, `EnsureBucketPartitions`, `MoveOrgToBucket`
- **Data Operations**: `CopyOrgData`, `ScrubOrgData`, `AuditBucket`

### SchemaHandler

Interface for schema-specific operations that must be implemented for each data structure:

- **`EnsureSchema`**: Create all necessary tables, indexes, and schema structures
- **`EnsurePartitions`**: Create time-based partitions for tables
- **`CopyOrgData`**: Copy organization data between buckets
- **`ScrubOrgData`**: Delete organization data from a bucket
- **`BuildAuditQuery`**: Build query to audit differences between buckets

## Configuration

### Config

Configuration struct for the partitioned repository:

```go
type Config struct {
    OrgBucketCount                int           // Number of buckets for org partitioning
    PartitionPreparationThreshold time.Duration // Time before partition end to create next partition
}
```

Note: Caching configuration is not part of this Config. To add org caching, wrap your repository with NewCachingRepository and provide a cache.CollectionCache implementation (see "Enable caching (optional)").

## Benefits

1. **Schema Flexibility**: Easily support different data structures by implementing a new `SchemaHandler`
2. **Separation of Concerns**: Bucket/org management is separated from schema-specific operations
3. **Dependency Injection**: Schema behavior can be injected at runtime
4. **Testability**: Schema handlers can be mocked for testing repository logic
5. **No External Dependencies**: Only depends on `go-libs` packages (cache, checkpoints, database, etc.)

## Types

### Org

```go
type Org struct {
    Name     string
    Created  time.Time
    BucketId int
    Bucket   *Bucket
}
```

### Bucket

```go
type Bucket struct {
    Id               int
    PartitionsEnding time.Time
    Created          time.Time
    Orgs             []*Org
}
```

### AuditResult

```go
type AuditResult struct {
    Org              string
    Tags             interface{}
    OriginalValue    *float64
    DestinationValue *float64
    StartTime        time.Time
}
```

## Helper Functions

The package also provides several helper functions in `helpers.go`:

- **`ModulatedHash(input string, space int) int`**: Converts a string to a number in the range (0, space] for consistent bucket assignment
- **`IsPostgresIdentifier(identifier string) bool`**: Validates PostgreSQL identifier syntax
- **`SchemaName(schema string) string`**: Sanitizes schema names by replacing hyphens with underscores
- **`TableName(schema string, tableName string) string`**: Returns a fully qualified table name
- **`IndexName(schema string, tableName string, suffix string) string`**: Generates an index name
