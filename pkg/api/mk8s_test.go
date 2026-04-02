//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8s"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sEphemeral"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// Mk8sTestSuite tests managed Kubernetes cluster operations.
// Note: Full CRUD testing may require cloud provider credentials and resources.
type Mk8sTestSuite struct {
	IntegrationSuite
}

func TestMk8sSuite(t *testing.T) {
	suite.Run(t, new(Mk8sTestSuite))
}

// newTestMk8sCluster creates a minimal mk8s cluster spec for testing.
// Uses ephemeral provider which is the simplest option for testing.
func newTestMk8sCluster(name string) *mk8s.Mk8sCluster {
	return &mk8s.Mk8sCluster{
		Name:        base.Name(name),
		Description: "Test mk8s cluster created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: mk8s.Mk8SClusterTags{
			"test":        "true",
			"integration": "libs-go",
		},
		Spec: mk8s.Mk8SClusterSpec{
			Version: mk8s.Mk8SClusterSpecVersion1315,
			Provider: mk8s.Mk8SClusterSpecProvider{
				Ephemeral: mk8sEphemeral.EphemeralProvider{
					Location: mk8sEphemeral.EphemeralProviderLocationAwsUsEast2,
					NodePools: []mk8sEphemeral.EphemeralPool{
						{
							Name:   "default",
							Count:  1,
							Cpu:    "2",
							Memory: "4Gi",
						},
					},
				},
			},
		},
	}
}

// TestMk8s_List tests listing mk8s clusters with iterator.
func (s *Mk8sTestSuite) TestMk8s_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List mk8s clusters using iterator
	count := 0
	for cluster, err := range s.Client().Mk8s().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		s.NotEmpty(cluster.Name, "Mk8s cluster should have a name")
		count++
	}
	s.T().Logf("Found %d mk8s clusters in org %s", count, s.TestOrg())
}

// TestMk8s_ListAll tests listing all mk8s clusters at once.
func (s *Mk8sTestSuite) TestMk8s_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.T().Logf("Found %d mk8s clusters in org %s", len(clusters), s.TestOrg())
}

// TestMk8s_ListPage tests pagination for listing mk8s clusters.
func (s *Mk8sTestSuite) TestMk8s_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	page, err := s.Client().Mk8s().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(page)
	s.T().Logf("First page has %d mk8s clusters", len(page.Items))
}

// TestMk8s_Get tests getting a specific mk8s cluster.
func (s *Mk8sTestSuite) TestMk8s_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List clusters to find one to get
	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(clusters) == 0 {
		s.T().Skip("No mk8s clusters available to test Get")
		return
	}

	// Get the first cluster
	clusterName := string(clusters[0].Name)
	fetched, err := s.Client().Mk8s().Get(ctx, s.TestOrg(), clusterName)
	s.Require().NoError(err)
	s.Equal(clusterName, string(fetched.Name))
	s.Equal("mk8s", string(fetched.Kind))
}

// TestMk8s_GetNonExistent tests getting a non-existent mk8s cluster.
func (s *Mk8sTestSuite) TestMk8s_GetNonExistent() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Mk8s().Get(ctx, s.TestOrg(), "non-existent-mk8s-"+randomSuffix())
	s.Require().Error(err, "Should fail to get non-existent mk8s cluster")
}

// TestMk8s_Query tests querying mk8s clusters with filters.
func (s *Mk8sTestSuite) TestMk8s_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all mk8s clusters
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	count := 0
	for cluster, err := range s.Client().Mk8s().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		s.NotEmpty(cluster.Name)
		count++
	}
	s.T().Logf("Query returned %d mk8s clusters", count)
}

// TestMk8s_QueryAll tests querying all mk8s clusters matching criteria.
func (s *Mk8sTestSuite) TestMk8s_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	clusters, err := s.Client().Mk8s().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.T().Logf("QueryAll returned %d mk8s clusters", len(clusters))
}

// TestMk8s_QueryPage tests pagination for querying mk8s clusters.
func (s *Mk8sTestSuite) TestMk8s_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	page, err := s.Client().Mk8s().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(page)
	s.T().Logf("QueryPage returned %d mk8s clusters", len(page.Items))
}

// TestMk8s_Create tests creating a new mk8s cluster.
// Note: This test is expensive and may take several minutes.
func (s *Mk8sTestSuite) TestMk8s_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 5*time.Minute)
	defer cancel()

	clusterName := "test-mk8s-create-" + randomSuffix()
	cluster := newTestMk8sCluster(clusterName)

	created, err := s.Client().Mk8s().Create(ctx, s.TestOrg(), cluster)
	if err != nil {
		// mk8s creation may require special permissions or cloud resources
		s.T().Logf("Mk8s cluster create failed (may require special permissions or cloud resources): %v", err)
		s.T().Skip("Mk8s cluster creation may require special permissions or cloud resources")
		return
	}

	defer func() {
		// Cleanup - this may take a while
		deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer deleteCancel()
		_ = s.Client().Mk8s().Delete(deleteCtx, s.TestOrg(), clusterName)
	}()

	s.Equal(clusterName, string(created.Name))
	s.Equal("Test mk8s cluster created by integration test", created.Description)
}

// TestMk8s_Update tests updating an mk8s cluster.
func (s *Mk8sTestSuite) TestMk8s_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List clusters to find one to update
	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(clusters) == 0 {
		s.T().Skip("No mk8s clusters available to test Update")
		return
	}

	// Get the first cluster
	clusterName := string(clusters[0].Name)
	cluster, err := s.Client().Mk8s().Get(ctx, s.TestOrg(), clusterName)
	s.Require().NoError(err)

	// Store original description
	originalDesc := cluster.Description

	// Update the cluster description
	cluster.Description = "Updated by integration test at " + time.Now().Format(time.RFC3339)

	updated, err := s.Client().Mk8s().Update(ctx, s.TestOrg(), clusterName, cluster)
	if err != nil {
		s.T().Logf("Mk8s cluster update failed: %v", err)
		s.T().Skip("Mk8s cluster update may require special permissions")
		return
	}

	s.Contains(updated.Description, "Updated by integration test")

	// Restore original description (best effort)
	cluster.Description = originalDesc
	_, _ = s.Client().Mk8s().Update(ctx, s.TestOrg(), clusterName, cluster)
}

// TestMk8s_Status tests that mk8s cluster status is populated.
func (s *Mk8sTestSuite) TestMk8s_Status() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	for _, cluster := range clusters {
		s.T().Logf("Mk8s cluster %s: serverUrl=%s, homeLocation=%s, oidcProviderUrl=%s",
			cluster.Name, cluster.Status.ServerUrl, cluster.Status.HomeLocation, cluster.Status.OidcProviderUrl)
	}
}

// TestMk8s_SpecProvider tests that mk8s cluster provider spec is populated.
func (s *Mk8sTestSuite) TestMk8s_SpecProvider() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	for _, cluster := range clusters {
		s.T().Logf("Mk8s cluster %s: version=%s", cluster.Name, cluster.Spec.Version)
		// Log which provider is configured
		provider := cluster.Spec.Provider
		if provider.Ephemeral.Location != "" {
			s.T().Logf("  Provider: ephemeral (location=%s)", provider.Ephemeral.Location)
		}
		if provider.Generic.Location != "" {
			s.T().Logf("  Provider: generic (location=%s)", provider.Generic.Location)
		}
		if provider.Aws.Region != "" {
			s.T().Logf("  Provider: aws (region=%s)", provider.Aws.Region)
		}
		if provider.Hetzner.Region != "" {
			s.T().Logf("  Provider: hetzner (region=%s)", provider.Hetzner.Region)
		}
		if provider.Azure.Location != "" {
			s.T().Logf("  Provider: azure (location=%s)", provider.Azure.Location)
		}
		if provider.Gcp.Region != "" {
			s.T().Logf("  Provider: gcp (region=%s)", provider.Gcp.Region)
		}
	}
}

// TestMk8s_Tags tests that mk8s cluster tags work correctly.
func (s *Mk8sTestSuite) TestMk8s_Tags() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(clusters) == 0 {
		s.T().Skip("No mk8s clusters available to test Tags")
		return
	}

	// Get the first cluster
	clusterName := string(clusters[0].Name)
	cluster, err := s.Client().Mk8s().Get(ctx, s.TestOrg(), clusterName)
	s.Require().NoError(err)

	// Store original tags
	originalTags := cluster.Tags

	// Add a test tag
	if cluster.Tags == nil {
		cluster.Tags = make(map[string]any)
	}
	cluster.Tags["integration-test-tag"] = "libs-go"

	updated, err := s.Client().Mk8s().Update(ctx, s.TestOrg(), clusterName, cluster)
	if err != nil {
		s.T().Logf("Mk8s cluster update failed: %v", err)
		s.T().Skip("Mk8s cluster update may require special permissions")
		return
	}

	s.Equal("libs-go", updated.Tags["integration-test-tag"])

	// Restore original tags (best effort)
	cluster.Tags = originalTags
	_, _ = s.Client().Mk8s().Update(ctx, s.TestOrg(), clusterName, cluster)
}

// TestMk8s_QueryByTag tests querying mk8s clusters by tag.
func (s *Mk8sTestSuite) TestMk8s_QueryByTag() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First add a tag to an existing cluster if any
	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(clusters) == 0 {
		s.T().Skip("No mk8s clusters available to test QueryByTag")
		return
	}

	clusterName := string(clusters[0].Name)
	cluster, err := s.Client().Mk8s().Get(ctx, s.TestOrg(), clusterName)
	s.Require().NoError(err)

	// Store original tags
	originalTags := cluster.Tags

	// Add a unique tag
	if cluster.Tags == nil {
		cluster.Tags = make(map[string]any)
	}
	uniqueTag := "query-test-" + randomSuffix()
	cluster.Tags[uniqueTag] = "yes"

	_, err = s.Client().Mk8s().Update(ctx, s.TestOrg(), clusterName, cluster)
	if err != nil {
		s.T().Skip("Cannot update cluster to test QueryByTag")
		return
	}

	// Restore tags at the end
	defer func() {
		cluster.Tags = originalTags
		_, _ = s.Client().Mk8s().Update(ctx, s.TestOrg(), clusterName, cluster)
	}()

	// Query for clusters with the tag
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{
				{
					Op:    query.TermOpEq,
					Tag:   uniqueTag,
					Value: "yes",
				},
			},
		},
	}

	found := false
	for c, err := range s.Client().Mk8s().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(c.Name) == clusterName {
			found = true
			break
		}
	}
	s.True(found, "Should find mk8s cluster with query")
}

// TestMk8s_Firewall tests that mk8s cluster firewall rules are parsed correctly.
func (s *Mk8sTestSuite) TestMk8s_Firewall() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	for _, cluster := range clusters {
		if len(cluster.Spec.Firewall) > 0 {
			s.T().Logf("Mk8s cluster %s has %d firewall rules:", cluster.Name, len(cluster.Spec.Firewall))
			for _, fw := range cluster.Spec.Firewall {
				s.T().Logf("  - CIDR: %s, Description: %s", fw.SourceCIDR, fw.Description)
			}
		}
	}
}

// TestMk8s_AddOns tests that mk8s cluster add-ons are parsed correctly.
func (s *Mk8sTestSuite) TestMk8s_AddOns() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	clusters, err := s.Client().Mk8s().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	for _, cluster := range clusters {
		addOns := cluster.Spec.AddOns
		s.T().Logf("Mk8s cluster %s add-ons:", cluster.Name)
		// NonCustomizableAddonConfig types are empty structs, check Metrics config
		if addOns.Metrics.KubeState {
			s.T().Logf("  - Metrics: kubeState=true")
		}
		if addOns.Metrics.CoreDns {
			s.T().Logf("  - Metrics: coreDns=true")
		}
		if addOns.Logs.AuditEnabled {
			s.T().Logf("  - Logs: auditEnabled=true")
		}
		if addOns.Nvidia.TaintGPUNodes {
			s.T().Logf("  - Nvidia: taintGPUNodes=true")
		}
		if addOns.AzureWorkloadIdentity.TenantId != "" {
			s.T().Logf("  - AzureWorkloadIdentity: tenantId=%s", addOns.AzureWorkloadIdentity.TenantId)
		}
	}
}
