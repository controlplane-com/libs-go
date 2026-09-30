package bucket

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeMaintenanceRepo struct {
	PartitionedRepository
	buckets   []*Bucket
	calls     []string
	ensureErr error
}

func (f *fakeMaintenanceRepo) ListBuckets() ([]*Bucket, error) {
	return f.buckets, nil
}

func (f *fakeMaintenanceRepo) EnsureBucketPartitions(b *Bucket, _ time.Time, _ int) error {
	f.calls = append(f.calls, fmt.Sprintf("ensure:%d", b.Id))
	return f.ensureErr
}

func (f *fakeMaintenanceRepo) DropExpiredPartitions(b *Bucket, _ time.Time) error {
	f.calls = append(f.calls, fmt.Sprintf("drop:%d", b.Id))
	return nil
}

type recordingLogger struct{ errors []string }

func (l *recordingLogger) Errorf(format string, args ...interface{}) {
	l.errors = append(l.errors, fmt.Sprintf(format, args...))
}

func TestMaintainBucketsDropsExpiredPartitionsAfterEnsuring(t *testing.T) {
	now := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	repo := &fakeMaintenanceRepo{buckets: []*Bucket{
		{Id: 1, PartitionsEnding: now.Add(time.Hour)},
		{Id: 2, PartitionsEnding: now.AddDate(1, 0, 0)},
	}}

	maintainBuckets(repo, 168*time.Hour, now, &recordingLogger{})

	require.Equal(t, []string{"ensure:1", "drop:1", "drop:2"}, repo.calls)
}

func TestMaintainBucketsStillDropsWhenEnsuringFails(t *testing.T) {
	now := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	repo := &fakeMaintenanceRepo{buckets: []*Bucket{{Id: 1, PartitionsEnding: now}}, ensureErr: errors.New("boom")}
	logger := &recordingLogger{}

	maintainBuckets(repo, 168*time.Hour, now, logger)

	require.Equal(t, []string{"ensure:1", "drop:1"}, repo.calls)
	require.Len(t, logger.errors, 1)
}

type plainSchemaHandler struct{ SchemaHandler }

type retainingSchemaHandler struct {
	SchemaHandler
	dropped []int
}

func (h *retainingSchemaHandler) DropExpiredPartitions(_ *gorm.DB, b *Bucket, _ time.Time) error {
	h.dropped = append(h.dropped, b.Id)
	return nil
}

type nilConnection struct{}

func (nilConnection) Db() *gorm.DB      { return nil }
func (nilConnection) DbRo() *gorm.DB    { return nil }
func (nilConnection) Initialize() error { return nil }

func TestDropExpiredPartitionsIsNoOpForHandlersWithoutRetention(t *testing.T) {
	repo := &PostgresqlPartitionedRepository{schemaHandler: plainSchemaHandler{}}
	require.NoError(t, repo.DropExpiredPartitions(&Bucket{Id: 1}, time.Now()))
}

func TestDropExpiredPartitionsDelegatesToRetentionHandler(t *testing.T) {
	handler := &retainingSchemaHandler{}
	repo := &PostgresqlPartitionedRepository{schemaHandler: handler, connection: nilConnection{}}
	require.NoError(t, repo.DropExpiredPartitions(&Bucket{Id: 3}, time.Now()))
	require.Equal(t, []int{3}, handler.dropped)
}
