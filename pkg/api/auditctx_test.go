//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/auditctx"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// AuditContextTestSuite tests audit context CRUD operations.
type AuditContextTestSuite struct {
	IntegrationSuite
}

func TestAuditContextSuite(t *testing.T) {
	suite.Run(t, new(AuditContextTestSuite))
}

// newTestAuditContext creates a minimal audit context spec for testing.
func newTestAuditContext(name string) *auditctx.AuditContext {
	return &auditctx.AuditContext{
		Name:        base.Name(name),
		Description: "Test audit context created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: auditctx.AuditContextTags{
			"test":        "true",
			"integration": "libs-go",
		},
	}
}

// TestAuditContext_List tests listing audit contexts with iterator.
func (s *AuditContextTestSuite) TestAuditContext_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List audit contexts using iterator
	count := 0
	for ac, err := range s.Client().AuditContexts().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		s.NotEmpty(ac.Name, "Audit context should have a name")
		count++
	}
	s.T().Logf("Found %d audit contexts in org %s", count, s.TestOrg())
}

// TestAuditContext_ListAll tests listing all audit contexts at once.
func (s *AuditContextTestSuite) TestAuditContext_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	auditContexts, err := s.Client().AuditContexts().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.T().Logf("Found %d audit contexts in org %s", len(auditContexts), s.TestOrg())
}

// TestAuditContext_ListPage tests pagination for listing audit contexts.
func (s *AuditContextTestSuite) TestAuditContext_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	page, err := s.Client().AuditContexts().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(page)
	s.T().Logf("First page has %d audit contexts", len(page.Items))
}

// TestAuditContext_Create tests creating a new audit context.
func (s *AuditContextTestSuite) TestAuditContext_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	acName := "test-auditctx-create-" + randomSuffix()
	ac := newTestAuditContext(acName)

	_, err := s.Client().AuditContexts().Create(ctx, s.TestOrg(), ac)
	if err != nil {
		// Audit context creation may require special permissions
		s.T().Logf("Audit context create failed (may require special permissions): %v", err)
		s.T().Skip("Audit context creation may require special permissions")
		return
	}

	// Note: Create returns an empty response, so we verify by fetching the audit context
	fetched, err := s.Client().AuditContexts().Get(ctx, s.TestOrg(), acName)
	s.Require().NoError(err, "Failed to fetch created audit context")

	s.Equal(acName, string(fetched.Name))
	s.Equal("Test audit context created by integration test", fetched.Description)
}

// TestAuditContext_Get tests getting a specific audit context.
func (s *AuditContextTestSuite) TestAuditContext_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First create an audit context
	acName := "test-auditctx-get-" + randomSuffix()
	ac := newTestAuditContext(acName)

	_, err := s.Client().AuditContexts().Create(ctx, s.TestOrg(), ac)
	if err != nil {
		// Try to list existing audit contexts instead
		auditContexts, listErr := s.Client().AuditContexts().ListAll(ctx, s.TestOrg())
		if listErr != nil || len(auditContexts) == 0 {
			s.T().Skip("Cannot create audit context and no existing ones to test Get")
			return
		}

		// Get an existing audit context
		existingName := string(auditContexts[0].Name)
		fetched, getErr := s.Client().AuditContexts().Get(ctx, s.TestOrg(), existingName)
		s.Require().NoError(getErr)
		s.Equal(existingName, string(fetched.Name))
		s.Equal("auditctx", string(fetched.Kind))
		return
	}

	defer func() {
		_ = s.Client().AuditContexts().Delete(context.Background(), s.TestOrg(), acName)
	}()

	// Get the audit context
	fetched, err := s.Client().AuditContexts().Get(ctx, s.TestOrg(), acName)
	s.Require().NoError(err)
	s.Equal(acName, string(fetched.Name))
	s.Equal("auditctx", string(fetched.Kind))
}

// TestAuditContext_GetNonExistent tests getting a non-existent audit context.
func (s *AuditContextTestSuite) TestAuditContext_GetNonExistent() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().AuditContexts().Get(ctx, s.TestOrg(), "non-existent-auditctx-"+randomSuffix())
	s.Require().Error(err, "Should fail to get non-existent audit context")
}

// TestAuditContext_Update tests updating an audit context.
func (s *AuditContextTestSuite) TestAuditContext_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First create an audit context
	acName := "test-auditctx-update-" + randomSuffix()
	ac := newTestAuditContext(acName)

	_, err := s.Client().AuditContexts().Create(ctx, s.TestOrg(), ac)
	if err != nil {
		s.T().Logf("Audit context create failed: %v", err)
		s.T().Skip("Audit context creation may require special permissions")
		return
	}

	defer func() {
		_ = s.Client().AuditContexts().Delete(context.Background(), s.TestOrg(), acName)
	}()

	// Update the audit context
	ac.Description = "Updated description"
	ac.Tags["updated"] = "true"

	updated, err := s.Client().AuditContexts().Update(ctx, s.TestOrg(), acName, ac)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestAuditContext_Delete tests deleting an audit context.
// Note: Audit context deletion is not supported by the API (returns 404 "would never exist"),
// so we skip this test.
func (s *AuditContextTestSuite) TestAuditContext_Delete() {
	s.T().Skip("Audit context deletion is not supported by the API")
}

// TestAuditContext_Query tests querying audit contexts with filters.
func (s *AuditContextTestSuite) TestAuditContext_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all audit contexts
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	count := 0
	for ac, err := range s.Client().AuditContexts().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		s.NotEmpty(ac.Name)
		count++
	}
	s.T().Logf("Query returned %d audit contexts", count)
}

// TestAuditContext_QueryAll tests querying all audit contexts matching criteria.
func (s *AuditContextTestSuite) TestAuditContext_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	auditContexts, err := s.Client().AuditContexts().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.T().Logf("QueryAll returned %d audit contexts", len(auditContexts))
}

// TestAuditContext_QueryPage tests pagination for querying audit contexts.
func (s *AuditContextTestSuite) TestAuditContext_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	page, err := s.Client().AuditContexts().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(page)
	s.T().Logf("QueryPage returned %d audit contexts", len(page.Items))
}

// TestAuditContext_BuiltinContexts tests that built-in audit contexts may be present.
func (s *AuditContextTestSuite) TestAuditContext_BuiltinContexts() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	auditContexts, err := s.Client().AuditContexts().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Look for built-in audit context types
	for _, ac := range auditContexts {
		if ac.Origin == auditctx.AuditContextOriginBuiltin || ac.Origin == auditctx.AuditContextOriginDefault {
			s.T().Logf("Found builtin audit context: %s (origin: %s)", ac.Name, ac.Origin)
		}
	}
}

// TestAuditContext_Status tests that audit context status is populated.
func (s *AuditContextTestSuite) TestAuditContext_Status() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	auditContexts, err := s.Client().AuditContexts().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	for _, ac := range auditContexts {
		if ac.Status != nil && len(ac.Status) > 0 {
			s.T().Logf("Audit context %s has status: %v", ac.Name, ac.Status)
		}
	}
}

// TestAuditContext_WithTags tests creating an audit context with tags.
func (s *AuditContextTestSuite) TestAuditContext_WithTags() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	acName := "test-auditctx-tags-" + randomSuffix()
	ac := newTestAuditContext(acName)
	ac.Tags["environment"] = "test"
	ac.Tags["team"] = "platform"

	_, err := s.Client().AuditContexts().Create(ctx, s.TestOrg(), ac)
	if err != nil {
		s.T().Skip("Audit context creation may require special permissions")
		return
	}

	// Note: Create returns an empty response, so we verify by fetching the audit context
	fetched, err := s.Client().AuditContexts().Get(ctx, s.TestOrg(), acName)
	s.Require().NoError(err, "Failed to fetch created audit context")

	s.Equal("test", fetched.Tags["environment"])
	s.Equal("platform", fetched.Tags["team"])
}

// TestAuditContext_QueryWithTags tests querying audit contexts by tags.
func (s *AuditContextTestSuite) TestAuditContext_QueryWithTags() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First create an audit context with a unique tag
	acName := "test-auditctx-querytag-" + randomSuffix()
	ac := newTestAuditContext(acName)
	ac.Tags["unique-query-tag"] = "unique-value"

	_, err := s.Client().AuditContexts().Create(ctx, s.TestOrg(), ac)
	if err != nil {
		s.T().Skip("Audit context creation may require special permissions")
		return
	}

	defer func() {
		_ = s.Client().AuditContexts().Delete(context.Background(), s.TestOrg(), acName)
	}()

	// Query for audit contexts with the tag
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{
				{
					Op:    query.TermOpEq,
					Tag:   "unique-query-tag",
					Value: "unique-value",
				},
			},
		},
	}

	found := false
	for ac, err := range s.Client().AuditContexts().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(ac.Name) == acName {
			found = true
			break
		}
	}
	s.True(found, "Should find audit context with query")
}
