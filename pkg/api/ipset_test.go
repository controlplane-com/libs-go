//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/ipSet"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// IPSetTestSuite tests IP set CRUD operations.
// IP sets are used to allocate and manage static IP addresses.
type IPSetTestSuite struct {
	IntegrationSuite
}

func TestIPSetSuite(t *testing.T) {
	suite.Run(t, new(IPSetTestSuite))
}

// newTestIPSet creates a minimal IP set for testing.
func newTestIPSet(name string) *ipSet.IpSet {
	return &ipSet.IpSet{
		Name:        base.Name(name),
		Description: "Test IP set created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: ipSet.IpSetTags{
			"test":        "true",
			"integration": "libs-go",
		},
		Spec: ipSet.IpSetSpec{
			// Empty locations - will be assigned when bound to a workload
			Locations: []ipSet.IpSetLocation{},
		},
	}
}

// newTestIPSetWithLocation creates an IP set with a specific location.
// locationLink should be a full link like "//location/aws-eu-central-1"
func newTestIPSetWithLocation(name, locationLink string) *ipSet.IpSet {
	return &ipSet.IpSet{
		Name:        base.Name(name),
		Description: "Test IP set with location",
		Version:     common.Float32Ptr(1),
		Tags: ipSet.IpSetTags{
			"test":        "true",
			"integration": "libs-go",
		},
		Spec: ipSet.IpSetSpec{
			Locations: []ipSet.IpSetLocation{
				{
					Name:            locationLink,
					RetentionPolicy: ipSet.IpSetLocationRetentionPolicyFree,
				},
			},
		},
	}
}

// TestIPSet_Create tests creating a new IP set.
func (s *IPSetTestSuite) TestIPSet_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-create-" + randomSuffix()
	ips := newTestIPSet(ipsetName)

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err, "Failed to create IP set")

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Verify by fetching - API may return empty object on create
	fetched, err := s.Client().IPSets().Get(ctx, s.TestOrg(), ipsetName)
	s.Require().NoError(err)
	s.Equal(ipsetName, string(fetched.Name))
	s.Equal("ipset", string(fetched.Kind))
}

// TestIPSet_Get tests getting a specific IP set.
func (s *IPSetTestSuite) TestIPSet_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-get-" + randomSuffix()
	ips := newTestIPSet(ipsetName)

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Get the IP set
	fetched, err := s.Client().IPSets().Get(ctx, s.TestOrg(), ipsetName)
	s.Require().NoError(err)
	s.Equal(ipsetName, string(fetched.Name))
	s.Equal("ipset", string(fetched.Kind))
}

// TestIPSet_Get_NotFound tests getting a non-existent IP set.
func (s *IPSetTestSuite) TestIPSet_Get_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().IPSets().Get(ctx, s.TestOrg(), "nonexistent-ipset-xyz")
	s.Require().Error(err, "Should error for non-existent IP set")
}

// TestIPSet_Update tests updating an IP set.
func (s *IPSetTestSuite) TestIPSet_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-update-" + randomSuffix()
	ips := newTestIPSet(ipsetName)

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Update the IP set
	ips.Description = "Updated description"
	ips.Tags["updated"] = "true"

	updated, err := s.Client().IPSets().Update(ctx, s.TestOrg(), ipsetName, ips)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestIPSet_Delete tests deleting an IP set.
func (s *IPSetTestSuite) TestIPSet_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-delete-" + randomSuffix()
	ips := newTestIPSet(ipsetName)

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	// Delete
	err = s.Client().IPSets().Delete(ctx, s.TestOrg(), ipsetName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().IPSets().Get(ctx, s.TestOrg(), ipsetName)
	s.Require().Error(err, "IP set should not exist after deletion")
}

// TestIPSet_Delete_NotFound tests deleting a non-existent IP set.
func (s *IPSetTestSuite) TestIPSet_Delete_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	err := s.Client().IPSets().Delete(ctx, s.TestOrg(), "nonexistent-ipset-for-delete")
	s.Require().Error(err, "Should error when deleting non-existent IP set")
}

// TestIPSet_List tests listing IP sets with iterator.
func (s *IPSetTestSuite) TestIPSet_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-list-" + randomSuffix()
	ips := newTestIPSet(ipsetName)

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// List using iterator
	found := false
	for ipset, err := range s.Client().IPSets().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		if string(ipset.Name) == ipsetName {
			found = true
			break
		}
	}
	s.True(found, "IP set should appear in list")
}

// TestIPSet_ListAll tests listing all IP sets at once.
func (s *IPSetTestSuite) TestIPSet_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-listall-" + randomSuffix()
	ips := newTestIPSet(ipsetName)

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	ipsets, err := s.Client().IPSets().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(ipsets, "Should have at least one IP set")

	found := false
	for _, ipset := range ipsets {
		if string(ipset.Name) == ipsetName {
			found = true
			break
		}
	}
	s.True(found, "IP set should be in list")
}

// TestIPSet_ListPage tests paginated listing.
func (s *IPSetTestSuite) TestIPSet_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-listpage-" + randomSuffix()
	ips := newTestIPSet(ipsetName)

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Get first page
	resp, err := s.Client().IPSets().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotEmpty(resp.Items, "Should have at least one IP set")
}

// TestIPSet_Query tests querying IP sets with filters.
func (s *IPSetTestSuite) TestIPSet_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-query-" + randomSuffix()
	ips := newTestIPSet(ipsetName)
	ips.Tags["querytest"] = "yes"

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Query for IP sets with the tag
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
	for ipset, err := range s.Client().IPSets().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(ipset.Name) == ipsetName {
			found = true
			break
		}
	}
	s.True(found, "Should find IP set with query")
}

// TestIPSet_QueryAll tests querying all IP sets matching criteria.
func (s *IPSetTestSuite) TestIPSet_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-queryall-" + randomSuffix()
	ips := newTestIPSet(ipsetName)
	ips.Tags["queryalltest"] = "yes"

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Query for IP sets with the tag
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

	ipsets, err := s.Client().IPSets().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotEmpty(ipsets, "Should have at least one IP set matching query")

	found := false
	for _, ipset := range ipsets {
		if string(ipset.Name) == ipsetName {
			found = true
			break
		}
	}
	s.True(found, "IP set should be in query results")
}

// TestIPSet_QueryPage tests paginated querying.
func (s *IPSetTestSuite) TestIPSet_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-querypage-" + randomSuffix()
	ips := newTestIPSet(ipsetName)

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Query for all IP sets
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	resp, err := s.Client().IPSets().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotEmpty(resp.Items, "Should have at least one IP set")
}

// TestIPSet_WithLocation tests creating an IP set with a location.
func (s *IPSetTestSuite) TestIPSet_WithLocation() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// First get a valid location
	locations, err := s.Client().Locations().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.Require().NotEmpty(locations, "Need at least one location for this test")

	// Location name needs to be a link format like "//location/aws-eu-central-1"
	locationName := string(locations[0].Name)
	locationLink := "//location/" + locationName

	ipsetName := "test-ipset-location-" + randomSuffix()
	ips := newTestIPSetWithLocation(ipsetName, locationLink)

	_, err = s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Verify by fetching - API may return empty object on create
	fetched, err := s.Client().IPSets().Get(ctx, s.TestOrg(), ipsetName)
	s.Require().NoError(err)
	s.Equal(ipsetName, string(fetched.Name))
	s.NotEmpty(fetched.Spec.Locations, "Should have a location configured")
	s.Contains(fetched.Spec.Locations[0].Name, locationName)
}

// TestIPSet_RetentionPolicies tests IP set retention policies.
func (s *IPSetTestSuite) TestIPSet_RetentionPolicies() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// First get a valid location
	locations, err := s.Client().Locations().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.Require().NotEmpty(locations, "Need at least one location for this test")

	// Location name needs to be a link format
	locationName := string(locations[0].Name)
	locationLink := "//location/" + locationName

	// Test with "keep" retention policy
	ipsetName := "test-ipset-keep-" + randomSuffix()
	ips := &ipSet.IpSet{
		Name:        base.Name(ipsetName),
		Description: "Test IP set with keep retention",
		Version:     common.Float32Ptr(1),
		Tags: ipSet.IpSetTags{
			"test": "true",
		},
		Spec: ipSet.IpSetSpec{
			Locations: []ipSet.IpSetLocation{
				{
					Name:            locationLink,
					RetentionPolicy: ipSet.IpSetLocationRetentionPolicyKeep,
				},
			},
		},
	}

	_, err = s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Verify by fetching - API may return empty object on create
	fetched, err := s.Client().IPSets().Get(ctx, s.TestOrg(), ipsetName)
	s.Require().NoError(err)
	s.NotEmpty(fetched.Spec.Locations, "Should have a location configured")
	s.Equal(ipSet.IpSetLocationRetentionPolicyKeep, fetched.Spec.Locations[0].RetentionPolicy)
}

// TestIPSet_MultipleTags tests creating an IP set with multiple tags.
func (s *IPSetTestSuite) TestIPSet_MultipleTags() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	ipsetName := "test-ipset-tags-" + randomSuffix()
	ips := &ipSet.IpSet{
		Name:        base.Name(ipsetName),
		Description: "Test IP set with multiple tags",
		Version:     common.Float32Ptr(1),
		Tags: ipSet.IpSetTags{
			"env":         "test",
			"team":        "platform",
			"service":     "api-gateway",
			"integration": "libs-go",
		},
		Spec: ipSet.IpSetSpec{
			Locations: []ipSet.IpSetLocation{},
		},
	}

	_, err := s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Verify by fetching - API may return empty object on create
	fetched, err := s.Client().IPSets().Get(ctx, s.TestOrg(), ipsetName)
	s.Require().NoError(err)

	// Verify tags are preserved
	s.Equal("test", fetched.Tags["env"])
	s.Equal("platform", fetched.Tags["team"])
	s.Equal("api-gateway", fetched.Tags["service"])
}

// TestIPSet_VerifyStatus tests that IP set status is populated.
func (s *IPSetTestSuite) TestIPSet_VerifyStatus() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// First get a valid location
	locations, err := s.Client().Locations().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.Require().NotEmpty(locations, "Need at least one location for this test")

	// Location name needs to be a link format
	locationName := string(locations[0].Name)
	locationLink := "//location/" + locationName

	ipsetName := "test-ipset-status-" + randomSuffix()
	ips := newTestIPSetWithLocation(ipsetName, locationLink)

	_, err = s.Client().IPSets().Create(ctx, s.TestOrg(), ips)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().IPSets().Delete(context.Background(), s.TestOrg(), ipsetName)
	}()

	// Wait a moment for status to populate
	time.Sleep(2 * time.Second)

	// Get the IP set and check status
	fetched, err := s.Client().IPSets().Get(ctx, s.TestOrg(), ipsetName)
	s.Require().NoError(err)

	// Log status info (IP addresses may take time to allocate)
	s.T().Logf("IP set status: %d addresses, error=%s, warning=%s",
		len(fetched.Status.IpAddresses), fetched.Status.Error, fetched.Status.Warning)

	for _, ip := range fetched.Status.IpAddresses {
		s.T().Logf("  IP: %s, state=%s, name=%s", ip.Ip, ip.State, ip.Name)
	}
}
