//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ClientTestSuite tests client initialization and basic operations.
type ClientTestSuite struct {
	IntegrationSuite
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

// TestClient_OrgResolution tests that org resolution works correctly.
func (s *ClientTestSuite) TestClient_OrgResolution() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Set default org on client
	s.Client().SetOrg(s.TestOrg())
	s.Equal(s.TestOrg(), s.Client().Org(), "Default org should be set")

	// List GVCs without specifying org - should use default
	found := false
	for gvc, err := range s.Client().GVCs().List(ctx, "") {
		s.Require().NoError(err)
		if string(gvc.Name) == s.TestGVC() {
			found = true
			break
		}
	}
	s.True(found, "Should find test GVC using default org")
}

// TestClient_ExplicitOrgOverridesDefault tests that explicit org overrides default.
func (s *ClientTestSuite) TestClient_ExplicitOrgOverridesDefault() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Set a bogus default org
	s.Client().SetOrg("non-existent-org-that-should-not-be-used")

	// But specify the real org explicitly - should work
	gvcs, err := s.Client().GVCs().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotNil(gvcs, "Should be able to list GVCs with explicit org")
}

// TestClient_GetNonExistentResource tests error handling for 404s.
func (s *ClientTestSuite) TestClient_GetNonExistentResource() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().GVCs().Get(ctx, s.TestOrg(), "non-existent-gvc-12345")
	s.Require().Error(err, "Should error for non-existent resource")
}

// TestClient_ListOrgs tests listing organizations.
func (s *ClientTestSuite) TestClient_ListOrgs() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List all orgs - should find our test org
	found := false
	for org, err := range s.Client().Orgs().List(ctx) {
		s.Require().NoError(err)
		if string(org.Name) == s.TestOrg() {
			found = true
			break
		}
	}
	s.True(found, "Should find test org in org list")
}

// TestClient_GetOrg tests getting a specific organization.
func (s *ClientTestSuite) TestClient_GetOrg() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	org, err := s.Client().Orgs().Get(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.Equal(s.TestOrg(), string(org.Name))
}
