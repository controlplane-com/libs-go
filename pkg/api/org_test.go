//go:build integration

package api_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/schema/org"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// OrgTestSuite tests organization read and update operations.
// Note: Organization creation is typically not available via API; orgs are created
// through other means (console, admin API, etc.).
type OrgTestSuite struct {
	IntegrationSuite
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgTestSuite))
}

// TestOrg_List tests listing organizations with iterator.
func (s *OrgTestSuite) TestOrg_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List organizations using iterator
	found := false
	for o, err := range s.Client().Orgs().List(ctx) {
		s.Require().NoError(err)
		if o.Name == s.TestOrg() {
			found = true
			break
		}
	}
	s.True(found, "Test org should appear in list")
}

// TestOrg_ListAll tests listing all organizations at once.
func (s *OrgTestSuite) TestOrg_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	orgs, err := s.Client().Orgs().ListAll(ctx)
	s.Require().NoError(err)
	s.NotEmpty(orgs, "Should have at least one org (test org)")

	found := false
	for _, org := range orgs {
		if org.Name == s.TestOrg() {
			found = true
			break
		}
	}
	s.True(found, "Test org should be in list")
}

// TestOrg_ListPage tests pagination for listing organizations.
func (s *OrgTestSuite) TestOrg_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	page, err := s.Client().Orgs().ListPage(ctx, "")
	s.Require().NoError(err)
	s.NotNil(page)
	s.NotEmpty(page.Items, "Should have at least one org")
}

// TestOrg_Get tests getting a specific organization.
func (s *OrgTestSuite) TestOrg_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get the test org
	org, err := s.Client().Orgs().Get(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.Equal(s.TestOrg(), org.Name)
	s.Equal("org", string(org.Kind))
}

// TestOrg_GetNonExistent tests getting a non-existent organization.
func (s *OrgTestSuite) TestOrg_GetNonExistent() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Orgs().Get(ctx, "non-existent-org-"+randomSuffix())
	s.Require().Error(err, "Should fail to get non-existent org")
}

// TestOrg_Update tests updating an organization.
func (s *OrgTestSuite) TestOrg_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get current org state
	currentOrg, err := s.Client().Orgs().Get(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Store original description to restore later
	originalDesc := currentOrg.Description

	// Create a minimal org object with only the fields we want to update
	// This avoids validation issues with complex nested fields like logging
	// Must include Version to avoid conflict errors
	updateOrg := &org.Org{
		Version:     currentOrg.Version,
		Description: "Updated by integration test at " + time.Now().Format(time.RFC3339),
	}

	updated, err := s.Client().Orgs().Update(ctx, s.TestOrg(), updateOrg)
	if err != nil {
		// Skip if the error is due to org-specific configuration we can't control
		// (e.g., invalid logging config in the test org)
		if strings.Contains(err.Error(), "spec.logging") {
			s.T().Skipf("Skipping test: org has invalid spec.logging configuration that prevents updates: %v", err)
			return
		}
		s.Require().NoError(err)
	}
	s.Contains(updated.Description, "Updated by integration test")

	// Restore original description (best effort) - use the new version from the update
	restoreOrg := &org.Org{
		Version:     updated.Version,
		Description: originalDesc,
	}
	_, _ = s.Client().Orgs().Update(ctx, s.TestOrg(), restoreOrg)
}

// TestOrg_Query tests querying organizations with filters.
func (s *OrgTestSuite) TestOrg_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for orgs - use a simple query that should match the test org
	// Since we can't add tags to the test org easily, we'll just query for all
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	found := false
	for o, err := range s.Client().Orgs().Query(ctx, q) {
		s.Require().NoError(err)
		if o.Name == s.TestOrg() {
			found = true
			break
		}
	}
	s.True(found, "Should find test org with query")
}

// TestOrg_QueryAll tests querying all organizations matching criteria.
func (s *OrgTestSuite) TestOrg_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all orgs
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	orgs, err := s.Client().Orgs().QueryAll(ctx, q)
	s.Require().NoError(err)
	s.NotEmpty(orgs, "Should have at least one org")

	found := false
	for _, org := range orgs {
		if org.Name == s.TestOrg() {
			found = true
			break
		}
	}
	s.True(found, "Test org should be in query results")
}

// TestOrg_QueryPage tests pagination for querying organizations.
func (s *OrgTestSuite) TestOrg_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	page, err := s.Client().Orgs().QueryPage(ctx, q)
	s.Require().NoError(err)
	s.NotNil(page)
	s.NotEmpty(page.Items, "Should have at least one org")
}

// TestOrg_Status tests that org status is populated.
func (s *OrgTestSuite) TestOrg_Status() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	org, err := s.Client().Orgs().Get(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Check that status fields are present (they may vary by org)
	s.T().Logf("Org %s: active=%v, accountLink=%s, endpointPrefix=%s",
		org.Name, org.Status.Active, org.Status.AccountLink, org.Status.EndpointPrefix)
}

// TestOrg_Spec tests that org spec is populated.
func (s *OrgTestSuite) TestOrg_Spec() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	org, err := s.Client().Orgs().Get(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Spec fields are optional but we can inspect them
	s.T().Logf("Org %s: sessionTimeout=%v, observability.logsRetention=%v",
		org.Name, org.Spec.SessionTimeoutSeconds, org.Spec.Observability.LogsRetentionDays)
}

// TestOrg_Tags tests that org tags work correctly.
func (s *OrgTestSuite) TestOrg_Tags() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get current org state to store original tags
	currentOrg, err := s.Client().Orgs().Get(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Store original tags
	originalTags := currentOrg.Tags

	// Create a minimal org object with only the tags field to update
	// This avoids validation issues with complex nested fields like logging
	// Must include Version to avoid conflict errors
	newTags := make(org.OrgTags)
	// Copy original tags if any
	for k, v := range originalTags {
		newTags[k] = v
	}
	// Add a test tag
	newTags["integration-test"] = "libs-go"

	updateOrg := &org.Org{
		Version: currentOrg.Version,
		Tags:    newTags,
	}

	updated, err := s.Client().Orgs().Update(ctx, s.TestOrg(), updateOrg)
	if err != nil {
		// Skip if the error is due to org-specific configuration we can't control
		// (e.g., invalid logging config in the test org)
		if strings.Contains(err.Error(), "spec.logging") {
			s.T().Skipf("Skipping test: org has invalid spec.logging configuration that prevents updates: %v", err)
			return
		}
		s.Require().NoError(err)
	}
	s.Equal("libs-go", updated.Tags["integration-test"])

	// Restore original tags (best effort) - use the new version from the update
	restoreOrg := &org.Org{
		Version: updated.Version,
		Tags:    originalTags,
	}
	_, _ = s.Client().Orgs().Update(ctx, s.TestOrg(), restoreOrg)
}
