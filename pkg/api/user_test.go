//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// UserTestSuite tests user list and query operations.
// Note: Users cannot be created via the API, so we only test read operations.
// The Invite operation is tested separately as it doesn't create users directly.
type UserTestSuite struct {
	IntegrationSuite
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

// TestUser_List tests listing users with iterator.
func (s *UserTestSuite) TestUser_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List using iterator - test org may not have users (service account auth doesn't add users)
	count := 0
	for _, err := range s.Client().Users().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		count++
	}
	s.T().Logf("Found %d users in org %s", count, s.TestOrg())
	// Note: Service account authenticated orgs may have zero users
}

// TestUser_ListAll tests listing all users at once.
func (s *UserTestSuite) TestUser_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	users, err := s.Client().Users().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.T().Logf("Found %d users in org %s", len(users), s.TestOrg())

	// Verify user structure if any users exist
	for _, u := range users {
		s.NotEmpty(u.Name, "User should have a name")
		// Note: User.Kind is UserKind which is different from base.Kind
	}
}

// TestUser_ListPage tests paginated listing of users.
func (s *UserTestSuite) TestUser_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	page, err := s.Client().Users().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(page, "Should return a page")
	s.T().Logf("ListPage returned %d users", len(page.Items))

	// Verify structure of returned users if any exist
	for _, u := range page.Items {
		s.NotEmpty(u.Name, "User should have a name")
	}
}

// TestUser_Get tests getting a specific user.
// This test relies on having at least one user in the organization.
func (s *UserTestSuite) TestUser_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First list users to get a valid user name
	users, err := s.Client().Users().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(users) == 0 {
		s.T().Skip("No users in organization - skipping Get test")
		return
	}

	// Get the first user
	userName := users[0].Name
	fetched, err := s.Client().Users().Get(ctx, s.TestOrg(), userName)
	s.Require().NoError(err)
	s.Equal(userName, fetched.Name)
}

// TestUser_GetNotFound tests getting a non-existent user.
func (s *UserTestSuite) TestUser_GetNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Users().Get(ctx, s.TestOrg(), "nonexistent-user-"+randomSuffix())
	s.Require().Error(err, "Should return error for non-existent user")
}

// TestUser_Query tests querying users.
// Note: This test may not find results if no users match the query.
func (s *UserTestSuite) TestUser_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query with a broad filter - we just want to verify the query works
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	count := 0
	for _, err := range s.Client().Users().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		count++
	}
	s.T().Logf("Query returned %d users", count)
	// Note: Service account authenticated orgs may have zero users
}

// TestUser_QueryAll tests querying all users with filters.
func (s *UserTestSuite) TestUser_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query with no filter to get all users
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	users, err := s.Client().Users().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.T().Logf("QueryAll returned %d users", len(users))
	// Note: Service account authenticated orgs may have zero users
}

// TestUser_QueryPage tests paginated querying of users.
func (s *UserTestSuite) TestUser_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query with no filter
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	page, err := s.Client().Users().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(page, "Should return a page")
	s.T().Logf("QueryPage returned %d users", len(page.Items))
	// Note: Service account authenticated orgs may have zero users
}

// TestUser_DeleteNotFound tests deleting a non-existent user.
// Note: Delete removes a user from the organization, not from the system.
func (s *UserTestSuite) TestUser_DeleteNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	err := s.Client().Users().Delete(ctx, s.TestOrg(), "nonexistent-user-"+randomSuffix())
	s.Require().Error(err, "Should return error for deleting non-existent user")
}

// TestUser_UserStructure tests the structure of returned user objects.
func (s *UserTestSuite) TestUser_UserStructure() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	users, err := s.Client().Users().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(users) == 0 {
		s.T().Skip("No users in organization - skipping UserStructure test")
		return
	}

	user := users[0]

	// Verify user has expected fields
	s.NotEmpty(user.Name, "User should have a name")
	s.NotEmpty(user.Id, "User should have an id")
	s.NotEmpty(user.Links, "User should have links")

	// Optional fields may or may not be present
	// email and idp may be populated
}

// TestUser_ListConsistency tests that List and ListAll return consistent results.
func (s *UserTestSuite) TestUser_ListConsistency() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Get all users via ListAll
	allUsers, err := s.Client().Users().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Count users via List iterator
	listCount := 0
	listNames := make(map[string]bool)
	for u, err := range s.Client().Users().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		listNames[u.Name] = true
		listCount++
	}

	// Verify counts match
	s.Equal(len(allUsers), listCount, "List and ListAll should return same number of users")

	// Verify all users from ListAll are in List
	for _, u := range allUsers {
		s.True(listNames[u.Name], "User %s from ListAll should be in List", u.Name)
	}
}

// TestUser_Invite tests the invite functionality.
// Note: This doesn't actually create a user - it sends an invitation.
// The invited user must accept the invitation to become a member.
// We skip the actual invitation since it requires a valid email and would send real emails.
func (s *UserTestSuite) TestUser_Invite_Skipped() {
	// We don't actually test successful invite as it would send real emails
	// and create real invitations that need to be cleaned up.
	// The invite endpoint is tested by verifying it exists and handles errors appropriately.
	s.T().Skip("Skipping actual invite test to avoid sending real emails")
}
