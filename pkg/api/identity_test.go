//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/identity"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// IdentityTestSuite tests identity CRUD operations.
type IdentityTestSuite struct {
	IntegrationSuite
}

func TestIdentitySuite(t *testing.T) {
	suite.Run(t, new(IdentityTestSuite))
}

// newTestIdentity creates a minimal identity for testing.
func newTestIdentity(name string) *identity.Identity {
	return &identity.Identity{
		Name:        base.Name(name),
		Description: "Test identity created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: identity.IdentityTags{
			"test":        "true",
			"integration": "libs-go",
		},
	}
}

// TestIdentity_Create tests creating a new identity.
func (s *IdentityTestSuite) TestIdentity_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-create-" + randomSuffix()
	ident := newTestIdentity(identityName)

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err, "Failed to create identity")

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	// Verify it exists by fetching it
	fetched, err := s.Client().Identities().Get(ctx, s.TestOrg(), s.TestGVC(), identityName)
	s.Require().NoError(err)
	s.Equal(identityName, string(fetched.Name))
	s.Equal("identity", string(fetched.Kind))
}

// TestIdentity_Get tests getting a specific identity.
func (s *IdentityTestSuite) TestIdentity_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-get-" + randomSuffix()
	ident := newTestIdentity(identityName)

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	// Get the identity
	fetched, err := s.Client().Identities().Get(ctx, s.TestOrg(), s.TestGVC(), identityName)
	s.Require().NoError(err)
	s.Equal(identityName, string(fetched.Name))
	s.Equal("identity", string(fetched.Kind))
	s.Equal("Test identity created by integration test", fetched.Description)
}

// TestIdentity_Update tests updating an identity.
func (s *IdentityTestSuite) TestIdentity_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-update-" + randomSuffix()
	ident := newTestIdentity(identityName)

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	// Update the identity
	ident.Description = "Updated description"
	ident.Tags["updated"] = "true"

	updated, err := s.Client().Identities().Update(ctx, s.TestOrg(), s.TestGVC(), identityName, ident)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestIdentity_Delete tests deleting an identity.
func (s *IdentityTestSuite) TestIdentity_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-delete-" + randomSuffix()
	ident := newTestIdentity(identityName)

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	// Delete
	err = s.Client().Identities().Delete(ctx, s.TestOrg(), s.TestGVC(), identityName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().Identities().Get(ctx, s.TestOrg(), s.TestGVC(), identityName)
	s.Require().Error(err, "Identity should not exist after deletion")
}

// TestIdentity_List tests listing identities with iterator.
func (s *IdentityTestSuite) TestIdentity_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-list-" + randomSuffix()
	ident := newTestIdentity(identityName)

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	// List using iterator
	found := false
	for i, err := range s.Client().Identities().List(ctx, s.TestOrg(), s.TestGVC()) {
		s.Require().NoError(err)
		if string(i.Name) == identityName {
			found = true
			break
		}
	}
	s.True(found, "Identity should appear in list")
}

// TestIdentity_ListAll tests listing all identities at once.
func (s *IdentityTestSuite) TestIdentity_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-listall-" + randomSuffix()
	ident := newTestIdentity(identityName)

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	identities, err := s.Client().Identities().ListAll(ctx, s.TestOrg(), s.TestGVC())
	s.Require().NoError(err)
	s.NotEmpty(identities, "Should have at least one identity")

	found := false
	for _, i := range identities {
		if string(i.Name) == identityName {
			found = true
			break
		}
	}
	s.True(found, "Identity should be in list")
}

// TestIdentity_Query tests querying identities with filters.
func (s *IdentityTestSuite) TestIdentity_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-query-" + randomSuffix()
	ident := newTestIdentity(identityName)
	ident.Tags["querytest"] = "yes"

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	// Query for identities with the tag
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
	for i, err := range s.Client().Identities().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(i.Name) == identityName {
			found = true
			break
		}
	}
	s.True(found, "Should find identity with query")
}

// TestIdentity_QueryAll tests querying all identities with filters.
func (s *IdentityTestSuite) TestIdentity_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-queryall-" + randomSuffix()
	ident := newTestIdentity(identityName)
	ident.Tags["queryalltest"] = "yes"

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	// Query for identities with the tag
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

	identities, err := s.Client().Identities().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)

	found := false
	for _, i := range identities {
		if string(i.Name) == identityName {
			found = true
			break
		}
	}
	s.True(found, "Should find identity with QueryAll")
}

// TestIdentity_ListPage tests paginated listing of identities.
func (s *IdentityTestSuite) TestIdentity_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-listpage-" + randomSuffix()
	ident := newTestIdentity(identityName)

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	// Get first page
	page, err := s.Client().Identities().ListPage(ctx, s.TestOrg(), s.TestGVC(), "")
	s.Require().NoError(err)
	s.NotNil(page, "Should return a page")
	s.NotEmpty(page.Items, "Should have at least one identity")
}

// TestIdentity_GetNotFound tests getting a non-existent identity.
func (s *IdentityTestSuite) TestIdentity_GetNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Identities().Get(ctx, s.TestOrg(), s.TestGVC(), "nonexistent-identity-"+randomSuffix())
	s.Require().Error(err, "Should return error for non-existent identity")
}

// TestIdentity_DeleteNotFound tests deleting a non-existent identity.
func (s *IdentityTestSuite) TestIdentity_DeleteNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	err := s.Client().Identities().Delete(ctx, s.TestOrg(), s.TestGVC(), "nonexistent-identity-"+randomSuffix())
	s.Require().Error(err, "Should return error for deleting non-existent identity")
}

// TestIdentity_CreateDuplicate tests creating a duplicate identity.
func (s *IdentityTestSuite) TestIdentity_CreateDuplicate() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	identityName := "test-identity-dup-" + randomSuffix()
	ident := newTestIdentity(identityName)

	_, err := s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Identities().Delete(context.Background(), s.TestOrg(), s.TestGVC(), identityName)
	}()

	// Try to create again with the same name
	_, err = s.Client().Identities().Create(ctx, s.TestOrg(), s.TestGVC(), ident)
	s.Require().Error(err, "Should return error for duplicate identity")
}
