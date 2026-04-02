//go:build integration

package api_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/api"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/gvc"
)

// Package-level variables for shared test account (created once across all suites)
var (
	sharedAccountOnce sync.Once
	sharedAccountID   string
	sharedAccountErr  error
)

// IntegrationSuite provides the base test suite with a pre-configured API client.
// All integration test suites should embed this struct.
type IntegrationSuite struct {
	suite.Suite
	client        *api.Client
	billingClient *api.Client
	testOrg       string
	testGVC       string
	testAccountID string
	ctx           context.Context
	cancel        context.CancelFunc
}

// SetupSuite initializes the API client and creates test resources.
func (s *IntegrationSuite) SetupSuite() {
	dsURL := os.Getenv("DATA_SERVICE_URL")
	if dsURL == "" {
		s.T().Fatal("DATA_SERVICE_URL environment variable is required")
	}

	token := os.Getenv("CONTROLLER_TOKEN")
	if token == "" {
		s.T().Fatal("CONTROLLER_TOKEN environment variable is required")
	}

	s.client = api.NewWithBaseURL(dsURL, token)
	s.testOrg = getEnvOrDefault("TEST_ORG", "libs-go-test-org")
	s.testGVC = "libs-go-test-gvc-" + randomSuffix()

	// Create a context with timeout for the entire test suite
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	// Initialize billing client if BILLING_URL is set
	if billingURL := os.Getenv("BILLING_URL"); billingURL != "" {
		s.billingClient = api.NewWithBaseURL(dsURL, token, api.WithBillingURL(billingURL))
		// Ensure test account exists (creates account and org via billing-ng)
		s.ensureTestAccount()
	} else {
		// Fall back to just checking org exists in data-service
		s.ensureTestOrg()
	}

	// Create a test GVC for workload tests
	s.ensureTestGVC()
}

// TearDownSuite cleans up test resources.
func (s *IntegrationSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}

	// Clean up test GVC (best effort)
	if s.testGVC != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = s.client.GVCs().Delete(ctx, s.testOrg, s.testGVC)
	}
}

// Client returns the API client.
func (s *IntegrationSuite) Client() *api.Client {
	return s.client
}

// TestOrg returns the test organization name.
func (s *IntegrationSuite) TestOrg() string {
	return s.testOrg
}

// TestGVC returns the test GVC name.
func (s *IntegrationSuite) TestGVC() string {
	return s.testGVC
}

// TestAccountID returns the test account ID.
func (s *IntegrationSuite) TestAccountID() string {
	return s.testAccountID
}

// BillingClient returns the billing API client (may be nil if BILLING_URL not set).
func (s *IntegrationSuite) BillingClient() *api.Client {
	return s.billingClient
}

// Ctx returns the test context.
func (s *IntegrationSuite) Ctx() context.Context {
	return s.ctx
}

// ensureTestAccount creates or reuses the shared test account and org via billing-ng.
func (s *IntegrationSuite) ensureTestAccount() {
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	defer cancel()

	// First check if org already exists
	org, err := s.billingClient.Billing().GetOrg(ctx, s.testOrg)
	if err == nil && org != nil {
		s.testAccountID = org.AccountID
		s.T().Logf("Test org %q already exists in account %s", s.testOrg, s.testAccountID)
		return
	}

	// Use sync.Once to ensure account is only created once across all test suites
	sharedAccountOnce.Do(func() {
		sharedAccountID, sharedAccountErr = s.createTestAccount(ctx)
	})

	if sharedAccountErr != nil {
		s.Require().NoError(sharedAccountErr, "Failed to create shared test account")
	}
	s.testAccountID = sharedAccountID

	// Create the test org under this account (each suite may have different org name)
	s.createTestOrg(ctx)
}

// createTestAccount creates a new billing account (called once via sync.Once).
func (s *IntegrationSuite) createTestAccount(ctx context.Context) (string, error) {
	testEmail := fmt.Sprintf("%s@example.com", s.testOrg)
	s.T().Logf("Creating shared test account with email %s", testEmail)

	accountReq := &api.CreateAccountRequest{
		Account: &api.BillingAccountInput{
			AccountName: "Integration Test Account",
			Description: "libs-go-integration-test",
			Email:       testEmail,
			Phone:       "555-1234",
			Address: &api.BillingAddress{
				Address1:   "123 Test Street",
				City:       "Test City",
				State:      "CA",
				Country:    "US",
				PostalCode: "12345",
			},
			ExtraInfo: map[string]string{
				"company": "Integration Test Company",
			},
		},
		Type: "trial",
	}

	accountResp, err := s.billingClient.Billing().CreateAccount(ctx, accountReq)
	if err != nil {
		return "", fmt.Errorf("create account: %w", err)
	}
	s.T().Logf("Created test account %s", accountResp.ID)

	// Accept Terms of Use
	touReq := &api.AcceptTOURequest{
		TOU: &api.TOUInput{
			CustomerName: "Integration Test",
			Accepted:     true,
		},
	}
	if err := s.billingClient.Billing().AcceptTOU(ctx, accountResp.ID, touReq); err != nil {
		return "", fmt.Errorf("accept TOU: %w", err)
	}
	s.T().Log("Accepted Terms of Use")

	// Increase org quota using raw HTTP request (not adding to SDK)
	if err := s.updateOrgQuota(ctx, accountResp.ID, 10); err != nil {
		return "", fmt.Errorf("update org quota: %w", err)
	}
	s.T().Log("Updated org quota to 10")

	return accountResp.ID, nil
}

// updateOrgQuota updates the org quota for an account using a raw HTTP request.
func (s *IntegrationSuite) updateOrgQuota(ctx context.Context, accountID string, quota int) error {
	billingURL := os.Getenv("BILLING_URL")
	if billingURL == "" {
		return fmt.Errorf("BILLING_URL not set")
	}

	url := fmt.Sprintf("%s/account/%s", billingURL, accountID)
	// PUT requires all required account fields
	body := map[string]any{
		"account": map[string]any{
			"fullName":    "Integration Test Account",
			"accountName": "libs-go-integration-test",
			"email":       fmt.Sprintf("%s@example.com", s.testOrg),
			"phone":       "555-1234",
			"address": map[string]string{
				"address1":   "123 Test Street",
				"city":       "Test City",
				"state":      "CA",
				"country":    "US",
				"postalCode": "12345",
			},
			"extraInfo": map[string]string{
				"company": "Integration Test Company",
			},
			"orgQuota": quota,
		},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CONTROLLER_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to update org quota: status %d", resp.StatusCode)
	}
	return nil
}

// createTestOrg creates the test org under the shared account.
func (s *IntegrationSuite) createTestOrg(ctx context.Context) {
	orgReq := &api.CreateOrgRequest{
		Org: &api.BillingOrg{
			Name:        s.testOrg,
			Description: "Integration test org for libs-go",
		},
		Invitees: []string{fmt.Sprintf("%s@example.com", s.testOrg)},
	}

	orgResp, err := s.billingClient.Billing().CreateOrg(ctx, s.testAccountID, orgReq)
	s.Require().NoError(err, "Failed to create test org")
	s.T().Logf("Created test org %s (wasCreated=%v)", orgResp.Name, orgResp.WasCreated)
}

// ensureTestOrg verifies the test organization exists.
// The org must be created before tests run (e.g., by the CI pipeline).
func (s *IntegrationSuite) ensureTestOrg() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	_, err := s.client.Orgs().Get(ctx, s.testOrg)
	if err != nil {
		s.T().Fatalf("Test org %q does not exist. It must be created before running integration tests: %v", s.testOrg, err)
	}
}

// ensureTestGVC creates the test GVC.
func (s *IntegrationSuite) ensureTestGVC() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	g := &gvc.Gvc{
		Name:        base.Name(s.testGVC),
		Description: "Integration test GVC for libs-go",
		Version:     common.Float32Ptr(1),
	}

	_, err := s.client.GVCs().Create(ctx, s.testOrg, g)
	s.Require().NoError(err, "Failed to create test GVC")
}

// randomSuffix generates a random 6-character hex suffix.
func randomSuffix() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "000000"
	}
	return hex.EncodeToString(bytes)
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestIntegrationSuite is required to run the base suite (mostly for validation).
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

// TestSuite_ClientInitialized verifies the client is properly initialized.
func (s *IntegrationSuite) TestSuite_ClientInitialized() {
	s.NotNil(s.client, "Client should be initialized")
	s.NotEmpty(s.testOrg, "Test org should be set")
	s.NotEmpty(s.testGVC, "Test GVC should be set")
}
