//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/workload"
)

// WorkloadTestSuite tests workload CRUD and command operations.
type WorkloadTestSuite struct {
	IntegrationSuite
}

func TestWorkloadSuite(t *testing.T) {
	suite.Run(t, new(WorkloadTestSuite))
}

// newTestWorkload creates a minimal workload spec for testing.
func newTestWorkload(name string) *workload.Workload {
	port := float32(8080)
	return &workload.Workload{
		Name:        name,
		Description: "Test workload created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: workload.WorkloadTags{
			"test":        "true",
			"integration": "libs-go",
		},
		Spec: workload.WorkloadSpec{
			Type: workload.WorkloadTypeServerless,
			Containers: []workload.ContainerSpec{
				{
					Name:   "main",
					Image:  "nginx:latest",
					Port:   &port,
					Memory: "128Mi",
					Cpu:    "50m",
				},
			},
		},
	}
}

// TestWorkload_Create tests creating a new workload.
func (s *WorkloadTestSuite) TestWorkload_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-create-" + randomSuffix()
	wl := newTestWorkload(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err, "Failed to create workload")

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Verify it exists by fetching it
	fetched, err := s.Client().Workloads().Get(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err)
	s.Equal(wlName, string(fetched.Name))
	s.Equal(workload.WorkloadTypeServerless, fetched.Spec.Type)
}

// TestWorkload_Get tests getting a specific workload.
func (s *WorkloadTestSuite) TestWorkload_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-get-" + randomSuffix()
	wl := newTestWorkload(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Get the workload
	fetched, err := s.Client().Workloads().Get(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err)
	s.Equal(wlName, string(fetched.Name))
	s.Equal("workload", string(fetched.Kind))
}

// TestWorkload_Update tests updating a workload.
func (s *WorkloadTestSuite) TestWorkload_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-update-" + randomSuffix()
	wl := newTestWorkload(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Update the workload
	wl.Description = "Updated description"
	wl.Tags["updated"] = "true"

	updated, err := s.Client().Workloads().Update(ctx, s.TestOrg(), s.TestGVC(), wlName, wl)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestWorkload_Delete tests deleting a workload.
func (s *WorkloadTestSuite) TestWorkload_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-delete-" + randomSuffix()
	wl := newTestWorkload(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err)

	// Delete
	err = s.Client().Workloads().Delete(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().Workloads().Get(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().Error(err, "Workload should not exist after deletion")
}

// TestWorkload_List tests listing workloads with iterator.
func (s *WorkloadTestSuite) TestWorkload_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-list-" + randomSuffix()
	wl := newTestWorkload(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// List using iterator
	found := false
	for w, err := range s.Client().Workloads().List(ctx, s.TestOrg(), s.TestGVC()) {
		s.Require().NoError(err)
		if string(w.Name) == wlName {
			found = true
			break
		}
	}
	s.True(found, "Workload should appear in list")
}

// TestWorkload_ListAll tests listing all workloads at once.
func (s *WorkloadTestSuite) TestWorkload_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-listall-" + randomSuffix()
	wl := newTestWorkload(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	workloads, err := s.Client().Workloads().ListAll(ctx, s.TestOrg(), s.TestGVC())
	s.Require().NoError(err)
	s.NotEmpty(workloads, "Should have at least one workload")

	found := false
	for _, w := range workloads {
		if string(w.Name) == wlName {
			found = true
			break
		}
	}
	s.True(found, "Workload should be in list")
}

// TestWorkload_Query tests querying workloads with filters.
func (s *WorkloadTestSuite) TestWorkload_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-query-" + randomSuffix()
	wl := newTestWorkload(wlName)
	wl.Tags["querytest"] = "yes"

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Query for workloads with the tag (org-level query)
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
	for w, err := range s.Client().Workloads().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if w.Name == wlName {
			found = true
			break
		}
	}
	s.True(found, "Should find workload with query")
}

// TestWorkload_ForceRedeploy tests force redeploying a workload.
func (s *WorkloadTestSuite) TestWorkload_ForceRedeploy() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-redeploy-" + randomSuffix()
	wl := newTestWorkload(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Force redeploy (sets cpln/deployTimestamp tag)
	err = s.Client().Workloads().ForceRedeploy(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err, "ForceRedeploy should succeed")

	// Verify the timestamp tag was set
	fetched, err := s.Client().Workloads().Get(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err)
	_, hasTag := fetched.Tags["cpln/deployTimestamp"]
	s.True(hasTag, "Workload should have cpln/deployTimestamp tag after ForceRedeploy")
}

// TestWorkload_ListCommands tests listing commands for a workload.
func (s *WorkloadTestSuite) TestWorkload_ListCommands() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	wlName := "test-wl-cmds-" + randomSuffix()
	wl := newTestWorkload(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// List commands (should be empty for a new workload)
	commands, err := s.Client().Workloads().ListCommandsAll(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err)
	// A new workload should have zero commands - may return nil or empty slice
	if commands != nil {
		s.Empty(commands, "New workload should have no commands")
	}
}
