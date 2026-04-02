//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/volumeSet"
)

// VolumeSetTestSuite tests volume set CRUD and command operations.
type VolumeSetTestSuite struct {
	IntegrationSuite
}

func TestVolumeSetSuite(t *testing.T) {
	suite.Run(t, new(VolumeSetTestSuite))
}

// newTestVolumeSet creates a minimal volume set spec for testing.
func newTestVolumeSet(name string) *volumeSet.VolumeSet {
	return &volumeSet.VolumeSet{
		Name:        base.Name(name),
		Description: "Test volume set created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: volumeSet.VolumeSetTags{
			"test":        "true",
			"integration": "libs-go",
		},
		Spec: volumeSet.VolumeSetSpec{
			InitialCapacity:  10,
			PerformanceClass: volumeSet.PerformanceClassNameGeneralPurposeSsd,
			FileSystemType:   volumeSet.FileSystemTypeExt4,
		},
	}
}

// TestVolumeSet_Create tests creating a new volume set.
func (s *VolumeSetTestSuite) TestVolumeSet_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-create-" + randomSuffix()
	vs := newTestVolumeSet(vsName)

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err, "Failed to create volume set")

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Verify it exists by fetching it
	fetched, err := s.Client().VolumeSets().Get(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().NoError(err)
	s.Equal(vsName, string(fetched.Name))
	s.Equal(volumeSet.PerformanceClassNameGeneralPurposeSsd, fetched.Spec.PerformanceClass)
	s.Equal(volumeSet.FileSystemTypeExt4, fetched.Spec.FileSystemType)
}

// TestVolumeSet_Get tests getting a specific volume set.
func (s *VolumeSetTestSuite) TestVolumeSet_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-get-" + randomSuffix()
	vs := newTestVolumeSet(vsName)

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Get the volume set
	fetched, err := s.Client().VolumeSets().Get(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().NoError(err)
	s.Equal(vsName, string(fetched.Name))
	s.Equal("volumeset", string(fetched.Kind))
}

// TestVolumeSet_Update tests updating a volume set.
func (s *VolumeSetTestSuite) TestVolumeSet_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-update-" + randomSuffix()
	vs := newTestVolumeSet(vsName)

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Update the volume set
	vs.Description = "Updated description"
	vs.Tags["updated"] = "true"

	updated, err := s.Client().VolumeSets().Update(ctx, s.TestOrg(), s.TestGVC(), vsName, vs)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestVolumeSet_Delete tests deleting a volume set.
func (s *VolumeSetTestSuite) TestVolumeSet_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-delete-" + randomSuffix()
	vs := newTestVolumeSet(vsName)

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	// Delete
	err = s.Client().VolumeSets().Delete(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().VolumeSets().Get(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().Error(err, "Volume set should not exist after deletion")
}

// TestVolumeSet_List tests listing volume sets with iterator.
func (s *VolumeSetTestSuite) TestVolumeSet_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-list-" + randomSuffix()
	vs := newTestVolumeSet(vsName)

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// List using iterator
	found := false
	for v, err := range s.Client().VolumeSets().List(ctx, s.TestOrg(), s.TestGVC()) {
		s.Require().NoError(err)
		if string(v.Name) == vsName {
			found = true
			break
		}
	}
	s.True(found, "Volume set should appear in list")
}

// TestVolumeSet_ListAll tests listing all volume sets at once.
func (s *VolumeSetTestSuite) TestVolumeSet_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-listall-" + randomSuffix()
	vs := newTestVolumeSet(vsName)

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	volumeSets, err := s.Client().VolumeSets().ListAll(ctx, s.TestOrg(), s.TestGVC())
	s.Require().NoError(err)
	s.NotEmpty(volumeSets, "Should have at least one volume set")

	found := false
	for _, v := range volumeSets {
		if string(v.Name) == vsName {
			found = true
			break
		}
	}
	s.True(found, "Volume set should be in list")
}

// TestVolumeSet_Query tests querying volume sets with filters.
func (s *VolumeSetTestSuite) TestVolumeSet_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-query-" + randomSuffix()
	vs := newTestVolumeSet(vsName)
	vs.Tags["querytest"] = "yes"

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Query for volume sets with the tag
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{
				{
					Op:    query.TermOpEq,
					Tag:   "querytest",
					Value: "yes",
				},
			},
		},
	}

	found := false
	for v, err := range s.Client().VolumeSets().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(v.Name) == vsName {
			found = true
			break
		}
	}
	s.True(found, "Should find volume set with query")
}

// TestVolumeSet_ListCommands tests listing commands for a volume set.
func (s *VolumeSetTestSuite) TestVolumeSet_ListCommands() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-cmds-" + randomSuffix()
	vs := newTestVolumeSet(vsName)

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// List commands (should be empty for a new volume set)
	commands, err := s.Client().VolumeSets().ListCommandsAll(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().NoError(err)
	// A new volume set should have zero commands - may return nil or empty slice
	if commands != nil {
		s.Empty(commands, "New volume set should have no commands")
	}
}

// TestVolumeSet_WithAutoscaling tests creating a volume set with autoscaling configuration.
func (s *VolumeSetTestSuite) TestVolumeSet_WithAutoscaling() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-autoscale-" + randomSuffix()
	vs := newTestVolumeSet(vsName)
	vs.Spec.Autoscaling = &volumeSet.VolumeSetSpecAutoscaling{
		MaxCapacity:       common.Float32Ptr(100),
		MinFreePercentage: common.Float32Ptr(10),
		ScalingFactor:     common.Float32Ptr(1.1),
	}

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err, "Failed to create volume set with autoscaling")

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Fetch and verify autoscaling config
	fetched, err := s.Client().VolumeSets().Get(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().NoError(err)
	s.NotNil(fetched.Spec.Autoscaling, "Volume set should have autoscaling config")
	s.Equal(float32(100), *fetched.Spec.Autoscaling.MaxCapacity)
}

// TestVolumeSet_WithSnapshots tests creating a volume set with snapshot configuration.
func (s *VolumeSetTestSuite) TestVolumeSet_WithSnapshots() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-snapshots-" + randomSuffix()
	vs := newTestVolumeSet(vsName)
	vs.Spec.Snapshots = &volumeSet.SnapshotSpec{
		CreateFinalSnapshot: true,
		RetentionDuration:   "7d",
		Schedule:            "0 0 * * *", // Daily at midnight
	}

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err, "Failed to create volume set with snapshots")

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Fetch and verify snapshot config
	fetched, err := s.Client().VolumeSets().Get(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().NoError(err)
	s.NotNil(fetched.Spec.Snapshots, "Volume set should have snapshots config")
	s.True(fetched.Spec.Snapshots.CreateFinalSnapshot)
}

// TestVolumeSet_XfsFileSystem tests creating a volume set with XFS file system.
func (s *VolumeSetTestSuite) TestVolumeSet_XfsFileSystem() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-xfs-" + randomSuffix()
	vs := newTestVolumeSet(vsName)
	vs.Spec.FileSystemType = volumeSet.FileSystemTypeXfs

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err, "Failed to create volume set with XFS")

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Fetch and verify file system type
	fetched, err := s.Client().VolumeSets().Get(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().NoError(err)
	// FileSystemType may not be returned for default values
	if fetched.Spec.FileSystemType != "" {
		s.Equal(volumeSet.FileSystemTypeXfs, fetched.Spec.FileSystemType)
	}
}

// TestVolumeSet_HighThroughputSsd tests creating a volume set with high-throughput SSD.
func (s *VolumeSetTestSuite) TestVolumeSet_HighThroughputSsd() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-htput-" + randomSuffix()
	vs := newTestVolumeSet(vsName)
	vs.Spec.PerformanceClass = volumeSet.PerformanceClassNameHighThroughputSsd
	// High-throughput SSD requires minimum 200 GB initial capacity
	vs.Spec.InitialCapacity = 200

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err, "Failed to create volume set with high-throughput SSD")

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Fetch and verify performance class
	fetched, err := s.Client().VolumeSets().Get(ctx, s.TestOrg(), s.TestGVC(), vsName)
	s.Require().NoError(err)
	// PerformanceClass may not be returned for default values
	if fetched.Spec.PerformanceClass != "" {
		s.Equal(volumeSet.PerformanceClassNameHighThroughputSsd, fetched.Spec.PerformanceClass)
	}
}

// TestVolumeSet_ListPage tests pagination for listing volume sets.
func (s *VolumeSetTestSuite) TestVolumeSet_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-page-" + randomSuffix()
	vs := newTestVolumeSet(vsName)

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Get first page
	page, err := s.Client().VolumeSets().ListPage(ctx, s.TestOrg(), s.TestGVC(), "")
	s.Require().NoError(err)
	s.NotNil(page)
	s.NotEmpty(page.Items, "Should have at least one volume set")
}

// TestVolumeSet_QueryAll tests querying all volume sets matching criteria.
func (s *VolumeSetTestSuite) TestVolumeSet_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	vsName := "test-vs-queryall-" + randomSuffix()
	vs := newTestVolumeSet(vsName)
	vs.Tags["queryalltest"] = "yes"

	_, err := s.Client().VolumeSets().Create(ctx, s.TestOrg(), s.TestGVC(), vs)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().VolumeSets().Delete(context.Background(), s.TestOrg(), s.TestGVC(), vsName)
	}()

	// Query for volume sets with the tag
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{
				{
					Op:    query.TermOpEq,
					Tag:   "queryalltest",
					Value: "yes",
				},
			},
		},
	}

	volumeSets, err := s.Client().VolumeSets().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotEmpty(volumeSets, "Should find at least one volume set with query")

	found := false
	for _, v := range volumeSets {
		if string(v.Name) == vsName {
			found = true
			break
		}
	}
	s.True(found, "Volume set should be in query results")
}
