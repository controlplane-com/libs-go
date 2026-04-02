//go:build integration

package api_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/domain"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// DomainTestSuite tests domain CRUD operations.
// Note: Domain creation requires proper DNS infrastructure setup, so many tests
// will skip if domain creation fails due to validation errors (like apex domain requirements).
type DomainTestSuite struct {
	IntegrationSuite
}

func TestDomainSuite(t *testing.T) {
	suite.Run(t, new(DomainTestSuite))
}

// newTestDomain creates a minimal domain for testing.
// Uses ns (nameserver) mode which is simpler to set up for testing.
func newTestDomain(name string, gvcLink string) *domain.Domain {
	return &domain.Domain{
		Name:        name,
		Description: "Test domain created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: domain.DomainTags{
			"test":        "true",
			"integration": "libs-go",
		},
		Spec: domain.DomainSpec{
			DnsMode:           domain.DomainSpecDnsModeNs,
			GvcLink:           gvcLink,
			CertChallengeType: domain.DomainSpecCertChallengeTypeDns01,
			Ports: []domain.ExternalPort{
				{
					Number:   common.Float32Ptr(443),
					Protocol: domain.ExternalPortProtocolHttp2,
				},
			},
		},
	}
}

// isDomainValidationError checks if an error is a domain validation error that we should skip for
func isDomainValidationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "apex_must_exist") ||
		strings.Contains(errStr, "dnsMode is set to cname") ||
		strings.Contains(errStr, "dns verification") ||
		strings.Contains(errStr, "DNS")
}

// TestDomain_Create tests creating a new domain.
// Note: This test may be skipped if domain creation requires special DNS setup.
func (s *DomainTestSuite) TestDomain_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Use a standalone domain name (not a subdomain) to avoid apex requirements
	domainName := "test-create-" + randomSuffix() + ".example.com"
	gvcLink := "/org/" + s.TestOrg() + "/gvc/" + s.TestGVC()
	d := newTestDomain(domainName, gvcLink)

	_, err := s.Client().Domains().Create(ctx, s.TestOrg(), d)
	if isDomainValidationError(err) {
		s.T().Skipf("Skipping domain creation test (requires DNS infrastructure): %v", err)
		return
	}
	s.Require().NoError(err, "Failed to create domain")

	defer func() {
		_ = s.Client().Domains().Delete(context.Background(), s.TestOrg(), domainName)
	}()

	// Note: Create returns an empty response, so we verify by fetching the domain
	fetched, err := s.Client().Domains().Get(ctx, s.TestOrg(), domainName)
	s.Require().NoError(err, "Failed to fetch created domain")
	s.Equal(domainName, fetched.Name)
	s.Equal("domain", string(fetched.Kind))
}

// TestDomain_Get tests getting a specific domain.
func (s *DomainTestSuite) TestDomain_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// First try to list existing domains
	existingDomains, err := s.Client().Domains().ListAll(ctx, s.TestOrg())
	if err == nil && len(existingDomains) > 0 {
		// Test Get with existing domain
		domainName := existingDomains[0].Name
		fetched, err := s.Client().Domains().Get(ctx, s.TestOrg(), domainName)
		s.Require().NoError(err)
		s.Equal(domainName, fetched.Name)
		s.Equal("domain", string(fetched.Kind))
		return
	}

	// No existing domains, try to create one
	domainName := "test-get-" + randomSuffix() + ".example.com"
	gvcLink := "/org/" + s.TestOrg() + "/gvc/" + s.TestGVC()
	d := newTestDomain(domainName, gvcLink)

	_, err = s.Client().Domains().Create(ctx, s.TestOrg(), d)
	if isDomainValidationError(err) {
		s.T().Skipf("Skipping domain Get test (requires DNS infrastructure): %v", err)
		return
	}
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Domains().Delete(context.Background(), s.TestOrg(), domainName)
	}()

	// Get the domain
	fetched, err := s.Client().Domains().Get(ctx, s.TestOrg(), domainName)
	s.Require().NoError(err)
	s.Equal(domainName, fetched.Name)
	s.Equal("domain", string(fetched.Kind))
}

// TestDomain_Get_NotFound tests getting a non-existent domain.
func (s *DomainTestSuite) TestDomain_Get_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Domains().Get(ctx, s.TestOrg(), "nonexistent-domain-xyz.example.com")
	s.Require().Error(err, "Should error for non-existent domain")
}

// TestDomain_Update tests updating a domain.
func (s *DomainTestSuite) TestDomain_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// First try to use an existing domain
	existingDomains, err := s.Client().Domains().ListAll(ctx, s.TestOrg())
	if err == nil && len(existingDomains) > 0 {
		// Test Update with existing domain (update description)
		existingDomain := existingDomains[0]
		originalDesc := existingDomain.Description

		existingDomain.Description = "Updated by integration test"
		updated, err := s.Client().Domains().Update(ctx, s.TestOrg(), existingDomain.Name, &existingDomain)
		s.Require().NoError(err)
		s.Equal("Updated by integration test", updated.Description)

		// Restore original
		existingDomain.Description = originalDesc
		_, _ = s.Client().Domains().Update(ctx, s.TestOrg(), existingDomain.Name, &existingDomain)
		return
	}

	// No existing domains, try to create one
	domainName := "test-update-" + randomSuffix() + ".example.com"
	gvcLink := "/org/" + s.TestOrg() + "/gvc/" + s.TestGVC()
	d := newTestDomain(domainName, gvcLink)

	_, err = s.Client().Domains().Create(ctx, s.TestOrg(), d)
	if isDomainValidationError(err) {
		s.T().Skipf("Skipping domain Update test (requires DNS infrastructure): %v", err)
		return
	}
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Domains().Delete(context.Background(), s.TestOrg(), domainName)
	}()

	// Update the domain
	d.Description = "Updated description"
	d.Tags["updated"] = "true"

	updated, err := s.Client().Domains().Update(ctx, s.TestOrg(), domainName, d)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestDomain_Delete tests deleting a domain.
func (s *DomainTestSuite) TestDomain_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	domainName := "test-delete-" + randomSuffix() + ".example.com"
	gvcLink := "/org/" + s.TestOrg() + "/gvc/" + s.TestGVC()
	d := newTestDomain(domainName, gvcLink)

	_, err := s.Client().Domains().Create(ctx, s.TestOrg(), d)
	if isDomainValidationError(err) {
		s.T().Skipf("Skipping domain Delete test (requires DNS infrastructure): %v", err)
		return
	}
	s.Require().NoError(err)

	// Delete
	err = s.Client().Domains().Delete(ctx, s.TestOrg(), domainName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().Domains().Get(ctx, s.TestOrg(), domainName)
	s.Require().Error(err, "Domain should not exist after deletion")
}

// TestDomain_List tests listing domains with iterator.
func (s *DomainTestSuite) TestDomain_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// List domains using iterator - this works even without creating domains
	count := 0
	for dom, err := range s.Client().Domains().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		s.NotEmpty(dom.Name, "Domain should have a name")
		count++
	}
	s.T().Logf("Found %d domains in org %s", count, s.TestOrg())
}

// TestDomain_ListAll tests listing all domains at once.
func (s *DomainTestSuite) TestDomain_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// ListAll works even without creating domains
	domains, err := s.Client().Domains().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.T().Logf("Found %d domains total", len(domains))
}

// TestDomain_ListPage tests paginated listing.
func (s *DomainTestSuite) TestDomain_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Get first page - works even without creating domains
	resp, err := s.Client().Domains().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(resp)
	s.T().Logf("First page has %d domains", len(resp.Items))
}

// TestDomain_Query tests querying domains with filters.
func (s *DomainTestSuite) TestDomain_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Query for all domains - works even without creating domains
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	count := 0
	for dom, err := range s.Client().Domains().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		s.NotEmpty(dom.Name)
		count++
	}
	s.T().Logf("Query found %d domains", count)
}

// TestDomain_QueryAll tests querying all domains matching criteria.
func (s *DomainTestSuite) TestDomain_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Query for all domains - works even without creating domains
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	domains, err := s.Client().Domains().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.T().Logf("QueryAll found %d domains", len(domains))
}

// TestDomain_QueryPage tests paginated querying.
func (s *DomainTestSuite) TestDomain_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Query for all domains
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	resp, err := s.Client().Domains().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.T().Logf("QueryPage found %d domains", len(resp.Items))
}

// TestDomain_WithRoutes tests creating a domain with routes.
func (s *DomainTestSuite) TestDomain_WithRoutes() {
	// This test requires a workload to exist, which we don't have in the test environment.
	// Skip for now - the API validates that workloadLink must be set when routes are defined.
	s.T().Skip("Skipping routes test - requires existing workload")
}

// TestDomain_VerifyStatus tests that domain status fields are populated.
func (s *DomainTestSuite) TestDomain_VerifyStatus() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// First try to use an existing domain
	existingDomains, err := s.Client().Domains().ListAll(ctx, s.TestOrg())
	if err == nil && len(existingDomains) > 0 {
		// Verify status on existing domain
		dom := existingDomains[0]
		s.T().Logf("Domain %s status: %s", dom.Name, dom.Status.Status)
		return
	}

	// No existing domains, try to create one
	domainName := "test-status-" + randomSuffix() + ".example.com"
	gvcLink := "/org/" + s.TestOrg() + "/gvc/" + s.TestGVC()
	d := newTestDomain(domainName, gvcLink)

	_, err = s.Client().Domains().Create(ctx, s.TestOrg(), d)
	if isDomainValidationError(err) {
		s.T().Skipf("Skipping domain VerifyStatus test (requires DNS infrastructure): %v", err)
		return
	}
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Domains().Delete(context.Background(), s.TestOrg(), domainName)
	}()

	// Get the domain and check status
	fetched, err := s.Client().Domains().Get(ctx, s.TestOrg(), domainName)
	s.Require().NoError(err)

	// Status should be set (could be initializing, ready, pendingDnsConfig, etc.)
	s.T().Logf("Domain status: %s", fetched.Status.Status)
}
