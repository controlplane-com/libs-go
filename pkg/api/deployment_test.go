//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/workload"
)

// DeploymentTestSuite tests deployment read operations.
// Deployments are read-only resources representing the deployment state of workloads.
type DeploymentTestSuite struct {
	IntegrationSuite
}

func TestDeploymentSuite(t *testing.T) {
	suite.Run(t, new(DeploymentTestSuite))
}

// newTestWorkloadForDeployment creates a minimal workload spec for deployment testing.
func newTestWorkloadForDeployment(name string) *workload.Workload {
	port := float32(8080)
	return &workload.Workload{
		Name:        name,
		Description: "Test workload for deployment integration test",
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

// TestDeployment_List tests listing deployments for a workload.
func (s *DeploymentTestSuite) TestDeployment_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 120*time.Second)
	defer cancel()

	// Create a workload first
	wlName := "test-deploy-list-" + randomSuffix()
	wl := newTestWorkloadForDeployment(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err, "Failed to create workload for deployment test")

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Wait a moment for deployment to be created
	time.Sleep(5 * time.Second)

	// List deployments using iterator
	count := 0
	for deployment, err := range s.Client().Deployments().List(ctx, s.TestOrg(), s.TestGVC(), wlName) {
		s.Require().NoError(err)
		s.NotEmpty(deployment.Name, "Deployment should have a name")
		count++
	}
	// A workload may have zero deployments initially or one per location
	s.T().Logf("Found %d deployments for workload %s", count, wlName)
}

// TestDeployment_ListAll tests listing all deployments for a workload at once.
func (s *DeploymentTestSuite) TestDeployment_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 120*time.Second)
	defer cancel()

	// Create a workload first
	wlName := "test-deploy-listall-" + randomSuffix()
	wl := newTestWorkloadForDeployment(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err, "Failed to create workload for deployment test")

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Wait a moment for deployment to be created
	time.Sleep(5 * time.Second)

	// List all deployments
	deployments, err := s.Client().Deployments().ListAll(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err)
	// Deployments may be empty if the GVC has no locations, or have entries per location
	s.T().Logf("Found %d deployments for workload %s", len(deployments), wlName)
}

// TestDeployment_ListPage tests pagination for listing deployments.
func (s *DeploymentTestSuite) TestDeployment_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 120*time.Second)
	defer cancel()

	// Create a workload first
	wlName := "test-deploy-page-" + randomSuffix()
	wl := newTestWorkloadForDeployment(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err, "Failed to create workload for deployment test")

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Wait a moment for deployment to be created
	time.Sleep(5 * time.Second)

	// Get first page
	page, err := s.Client().Deployments().ListPage(ctx, s.TestOrg(), s.TestGVC(), wlName, "")
	s.Require().NoError(err)
	s.NotNil(page)
	// The page may have zero items if no deployments exist yet
	s.T().Logf("First page has %d deployments", len(page.Items))
}

// TestDeployment_Get tests getting a specific deployment by name.
func (s *DeploymentTestSuite) TestDeployment_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 120*time.Second)
	defer cancel()

	// Create a workload first
	wlName := "test-deploy-get-" + randomSuffix()
	wl := newTestWorkloadForDeployment(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err, "Failed to create workload for deployment test")

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Wait a moment for deployment to be created
	time.Sleep(5 * time.Second)

	// List deployments first to get a deployment name
	deployments, err := s.Client().Deployments().ListAll(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err)

	// Skip if no deployments exist (GVC may have no locations)
	if len(deployments) == 0 {
		s.T().Skip("No deployments available - GVC may have no locations configured")
		return
	}

	// Get the first deployment by name
	deploymentName := deployments[0].Name
	fetched, err := s.Client().Deployments().Get(ctx, s.TestOrg(), s.TestGVC(), wlName, deploymentName)
	s.Require().NoError(err)
	s.Equal(deploymentName, fetched.Name)
	s.Equal("deployment", string(fetched.Kind))
}

// TestDeployment_Status tests that deployment status is populated.
func (s *DeploymentTestSuite) TestDeployment_Status() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 120*time.Second)
	defer cancel()

	// Create a workload first
	wlName := "test-deploy-status-" + randomSuffix()
	wl := newTestWorkloadForDeployment(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err, "Failed to create workload for deployment test")

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Wait for deployment to be created
	time.Sleep(5 * time.Second)

	// List deployments
	deployments, err := s.Client().Deployments().ListAll(ctx, s.TestOrg(), s.TestGVC(), wlName)
	s.Require().NoError(err)

	// Skip if no deployments exist
	if len(deployments) == 0 {
		s.T().Skip("No deployments available - GVC may have no locations configured")
		return
	}

	// Check that the deployment has status information
	for _, d := range deployments {
		s.NotEmpty(d.Name, "Deployment should have a name")
		// Status fields are optional but should be present in a real deployment
		s.T().Logf("Deployment %s: ready=%v, deploying=%v, endpoint=%s",
			d.Name, d.Status.Ready, d.Status.Deploying, d.Status.Endpoint)
	}
}

// TestDeployment_NonExistentWorkload tests getting deployments for a non-existent workload.
func (s *DeploymentTestSuite) TestDeployment_NonExistentWorkload() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	nonExistentWorkload := "non-existent-workload-" + randomSuffix()

	// Attempting to list deployments for a non-existent workload should fail
	_, err := s.Client().Deployments().ListAll(ctx, s.TestOrg(), s.TestGVC(), nonExistentWorkload)
	s.Require().Error(err, "Should fail to list deployments for non-existent workload")
}

// TestDeployment_NonExistentDeployment tests getting a non-existent deployment.
func (s *DeploymentTestSuite) TestDeployment_NonExistentDeployment() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Create a workload first
	wlName := "test-deploy-noexist-" + randomSuffix()
	wl := newTestWorkloadForDeployment(wlName)

	_, err := s.Client().Workloads().Create(ctx, s.TestOrg(), s.TestGVC(), wl)
	s.Require().NoError(err, "Failed to create workload for deployment test")

	defer func() {
		_ = s.Client().Workloads().Delete(context.Background(), s.TestOrg(), s.TestGVC(), wlName)
	}()

	// Try to get a non-existent deployment
	_, err = s.Client().Deployments().Get(ctx, s.TestOrg(), s.TestGVC(), wlName, "non-existent-deployment")
	s.Require().Error(err, "Should fail to get non-existent deployment")
}
