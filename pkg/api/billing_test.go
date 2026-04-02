//go:build integration

package api_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/api"
)

// BillingTestSuite tests billing service operations.
type BillingTestSuite struct {
	IntegrationSuite
	billingClient *api.Client
}

// TestBillingSuite runs the billing test suite.
func TestBillingSuite(t *testing.T) {
	suite.Run(t, new(BillingTestSuite))
}

// SetupSuite initializes the billing client.
func (s *BillingTestSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()

	// Use BILLING_URL if set (for local dev), otherwise use the default
	opts := []api.Option{}
	if billingURL := os.Getenv("BILLING_URL"); billingURL != "" {
		opts = append(opts, api.WithBillingURL(billingURL))
	}

	s.billingClient = api.NewWithBaseURL(
		os.Getenv("DATA_SERVICE_URL"),
		os.Getenv("CONTROLLER_TOKEN"),
		opts...,
	)
}

// TestBilling_GetOrg tests fetching org details from billing service.
func (s *BillingTestSuite) TestBilling_GetOrg() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	org, err := s.billingClient.Billing().GetOrg(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.Equal(s.TestOrg(), org.Name)
	s.NotEmpty(org.AccountID)
}

// TestBilling_GetOrgAccount tests fetching account for an org.
func (s *BillingTestSuite) TestBilling_GetOrgAccount() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	account, err := s.billingClient.Billing().GetOrgAccount(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(account.ID)
}

// TestBilling_QueryChargesByOrg tests querying charges for an org.
func (s *BillingTestSuite) TestBilling_QueryChargesByOrg() {
	s.T().Skip("Skipped: requires metering service to be configured")
}

// TestBilling_CreateOrg tests creating an org via billing service.
func (s *BillingTestSuite) TestBilling_CreateOrg() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First need an account ID - get from existing org
	org, err := s.billingClient.Billing().GetOrg(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(org.AccountID)

	// Create a new org
	orgName := "test-billing-org-" + randomSuffix()
	req := &api.CreateOrgRequest{
		Org: &api.BillingOrg{
			Name:        orgName,
			Description: "Test org for billing integration",
		},
	}

	resp, err := s.billingClient.Billing().CreateOrg(ctx, org.AccountID, req)
	s.Require().NoError(err)
	s.Equal(orgName, resp.Name)
	// Note: WasCreated field isn't properly serialized by billing-ng (uses capital C)

	// Note: Org cleanup would be done via billing-ng DELETE /org/{name}
	// which is not yet implemented in this SDK
}

// TestBilling_ListInvoices tests listing invoices for an account.
func (s *BillingTestSuite) TestBilling_ListInvoices() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get account ID from org
	account, err := s.billingClient.Billing().GetOrgAccount(ctx, s.TestOrg())
	s.Require().NoError(err)

	// List invoices for the last 30 days (API requires start/end aligned to hour boundaries)
	now := time.Now().UTC().Truncate(time.Hour)
	start := now.Add(-30 * 24 * time.Hour)
	invoices, err := s.billingClient.Billing().ListInvoices(ctx, account.ID, &start, &now)
	s.Require().NoError(err)
	// invoices may be nil or empty slice for test account with no billing history
	s.T().Logf("Found %d invoices", len(invoices))
}
