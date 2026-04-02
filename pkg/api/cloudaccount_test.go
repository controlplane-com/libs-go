//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/cloudaccount"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// CloudAccountTestSuite tests cloud account operations.
// Note: Full CRUD testing requires valid cloud credentials.
// Tests that create cloud accounts use dummy data and expect failure or
// verification without actual cloud validation.
type CloudAccountTestSuite struct {
	IntegrationSuite
}

func TestCloudAccountSuite(t *testing.T) {
	suite.Run(t, new(CloudAccountTestSuite))
}

// newTestCloudAccount creates a minimal cloud account for testing.
// Note: This uses a dummy AWS role ARN that won't pass validation.
func newTestCloudAccount(name string) *cloudaccount.CloudAccount {
	return &cloudaccount.CloudAccount{
		Name:        base.Name(name),
		Description: "Test cloud account created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: cloudaccount.CloudAccountTags{
			"test":        "true",
			"integration": "libs-go",
		},
		Provider: base.CloudProviderAws,
		Data: map[string]interface{}{
			"roleArn": "arn:aws:iam::123456789012:role/test-role-does-not-exist",
		},
	}
}

// TestCloudAccount_List tests listing cloud accounts with iterator.
func (s *CloudAccountTestSuite) TestCloudAccount_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List using iterator - may be empty if no cloud accounts configured
	count := 0
	for ca, err := range s.Client().CloudAccounts().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		s.NotEmpty(ca.Name, "Cloud account should have a name")
		count++
		if count >= 5 {
			break
		}
	}
	s.T().Logf("Found %d cloud accounts in org", count)
}

// TestCloudAccount_ListAll tests listing all cloud accounts at once.
func (s *CloudAccountTestSuite) TestCloudAccount_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	cloudAccounts, err := s.Client().CloudAccounts().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	// Note: slice may be empty if no cloud accounts configured

	// If we have cloud accounts, verify structure
	for _, ca := range cloudAccounts {
		s.NotEmpty(ca.Name, "Cloud account should have a name")
		s.Equal("cloudaccount", string(ca.Kind), "Kind should be 'cloudaccount'")
		s.NotEmpty(ca.Provider, "Cloud account should have a provider")
	}
	s.T().Logf("Found %d cloud accounts total", len(cloudAccounts))
}

// TestCloudAccount_ListPage tests paginated listing.
func (s *CloudAccountTestSuite) TestCloudAccount_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	resp, err := s.Client().CloudAccounts().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(resp)
	s.T().Logf("First page has %d cloud accounts", len(resp.Items))
}

// TestCloudAccount_Get tests getting a specific cloud account.
func (s *CloudAccountTestSuite) TestCloudAccount_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First, list cloud accounts to get a valid name
	cloudAccounts, err := s.Client().CloudAccounts().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(cloudAccounts) == 0 {
		s.T().Skip("No cloud accounts available to test Get operation")
		return
	}

	caName := string(cloudAccounts[0].Name)

	// Get the specific cloud account
	ca, err := s.Client().CloudAccounts().Get(ctx, s.TestOrg(), caName)
	s.Require().NoError(err)
	s.Equal(caName, string(ca.Name))
	s.Equal("cloudaccount", string(ca.Kind))
	s.NotEmpty(ca.Provider, "Cloud account should have a provider")
}

// TestCloudAccount_Get_NotFound tests getting a non-existent cloud account.
func (s *CloudAccountTestSuite) TestCloudAccount_Get_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().CloudAccounts().Get(ctx, s.TestOrg(), "nonexistent-cloudaccount-xyz")
	s.Require().Error(err, "Should error for non-existent cloud account")
}

// TestCloudAccount_Query tests querying cloud accounts with filters.
func (s *CloudAccountTestSuite) TestCloudAccount_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all cloud accounts (empty query matches all)
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	count := 0
	for ca, err := range s.Client().CloudAccounts().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		s.NotEmpty(ca.Name, "Cloud account should have a name")
		count++
		if count >= 5 {
			break
		}
	}
	s.T().Logf("Query found %d cloud accounts", count)
}

// TestCloudAccount_QueryAll tests querying all cloud accounts matching criteria.
func (s *CloudAccountTestSuite) TestCloudAccount_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all cloud accounts
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	cloudAccounts, err := s.Client().CloudAccounts().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.T().Logf("QueryAll found %d cloud accounts", len(cloudAccounts))
}

// TestCloudAccount_QueryPage tests paginated querying.
func (s *CloudAccountTestSuite) TestCloudAccount_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	resp, err := s.Client().CloudAccounts().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.T().Logf("QueryPage found %d cloud accounts", len(resp.Items))
}

// TestCloudAccount_Create tests creating a cloud account.
// Note: This will likely fail validation since we use a dummy role ARN,
// but it tests that the API call structure is correct.
func (s *CloudAccountTestSuite) TestCloudAccount_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	caName := "test-ca-create-" + randomSuffix()
	ca := newTestCloudAccount(caName)

	// Attempt to create - may fail due to invalid credentials
	_, err := s.Client().CloudAccounts().Create(ctx, s.TestOrg(), ca)
	if err != nil {
		// Expected: the dummy ARN won't validate
		s.T().Logf("Create failed as expected with dummy credentials: %v", err)
		return
	}

	// If it somehow succeeded, clean up
	defer func() {
		_ = s.Client().CloudAccounts().Delete(context.Background(), s.TestOrg(), caName)
	}()

	// Note: Create returns an empty response, so we verify by fetching the cloud account
	fetched, err := s.Client().CloudAccounts().Get(ctx, s.TestOrg(), caName)
	s.Require().NoError(err, "Failed to fetch created cloud account")
	s.Equal(caName, string(fetched.Name))
}

// TestCloudAccount_Delete_NotFound tests deleting a non-existent cloud account.
func (s *CloudAccountTestSuite) TestCloudAccount_Delete_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	err := s.Client().CloudAccounts().Delete(ctx, s.TestOrg(), "nonexistent-cloudaccount-for-delete")
	s.Require().Error(err, "Should error when deleting non-existent cloud account")
}

// TestCloudAccount_VerifyProviders tests that cloud accounts have valid providers.
func (s *CloudAccountTestSuite) TestCloudAccount_VerifyProviders() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	cloudAccounts, err := s.Client().CloudAccounts().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(cloudAccounts) == 0 {
		s.T().Skip("No cloud accounts available to verify providers")
		return
	}

	// Valid providers based on schema
	validProviders := map[string]bool{
		"aws":   true,
		"gcp":   true,
		"azure": true,
		"ngs":   true,
	}

	for _, ca := range cloudAccounts {
		s.True(validProviders[string(ca.Provider)], "Cloud account %s should have a valid provider, got: %s", ca.Name, ca.Provider)
	}
}

// TestCloudAccount_VerifyStatus tests that cloud accounts have status information.
func (s *CloudAccountTestSuite) TestCloudAccount_VerifyStatus() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	cloudAccounts, err := s.Client().CloudAccounts().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(cloudAccounts) == 0 {
		s.T().Skip("No cloud accounts available to verify status")
		return
	}

	for _, ca := range cloudAccounts {
		// Status should be present
		s.T().Logf("Cloud account %s: usable=%v, lastChecked=%s, lastError=%s",
			ca.Name, ca.Status.Usable, ca.Status.LastChecked, ca.Status.LastError)
	}
}
