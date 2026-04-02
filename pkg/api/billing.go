package api

import (
	"context"
	"fmt"
	"time"
)

// BillingService handles billing operations via billing-ng.
type BillingService struct {
	client *Client
}

// doBilling makes a request to the billing service.
func (s *BillingService) doBilling(ctx context.Context, method, path string, body, result any) error {
	if s.client.billingURL == "" {
		return fmt.Errorf("billing URL not configured; use WithBillingURL option")
	}
	return s.client.doWithBaseURL(ctx, method, s.client.billingURL, path, body, result)
}

// doBillingRaw makes a request to the billing service and returns raw bytes.
// Used for non-JSON responses like PDF downloads.
func (s *BillingService) doBillingRaw(ctx context.Context, method, path string, body any, accept string) ([]byte, error) {
	if s.client.billingURL == "" {
		return nil, fmt.Errorf("billing URL not configured; use WithBillingURL option")
	}
	return s.client.doWithBaseURLRaw(ctx, method, s.client.billingURL, path, body, accept)
}

// --- Org Operations ---

// CreateOrg creates a new organization in the specified account.
func (s *BillingService) CreateOrg(ctx context.Context, accountID string, req *CreateOrgRequest) (*CreateOrgResponse, error) {
	path := fmt.Sprintf("/account/%s/org", accountID)
	var result CreateOrgResponse
	if err := s.doBilling(ctx, "POST", path, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOrg returns organization details by name.
func (s *BillingService) GetOrg(ctx context.Context, orgName string) (*BillingOrg, error) {
	path := fmt.Sprintf("/org/%s", orgName)
	var result BillingOrg
	if err := s.doBilling(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOrgAccount returns the billing account for an organization.
func (s *BillingService) GetOrgAccount(ctx context.Context, orgName string) (*BillingAccount, error) {
	path := fmt.Sprintf("/org/%s/account", orgName)
	var result BillingAccount
	if err := s.doBilling(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Charge Operations ---

// QueryChargesByAccount queries charges for an account.
func (s *BillingService) QueryChargesByAccount(ctx context.Context, accountID string, req *ChargeQueryRequest) (*ChargeQueryResponse, error) {
	path := fmt.Sprintf("/account/%s/charges/-query", accountID)
	var result ChargeQueryResponse
	if err := s.doBilling(ctx, "POST", path, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryChargesByOrg queries charges for an organization.
func (s *BillingService) QueryChargesByOrg(ctx context.Context, orgName string, req *ChargeQueryRequest) (*ChargeQueryResponse, error) {
	path := fmt.Sprintf("/org/%s/charges/-query", orgName)
	var result ChargeQueryResponse
	if err := s.doBilling(ctx, "POST", path, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Invoice Operations ---

// ListInvoices returns invoices for an account.
func (s *BillingService) ListInvoices(ctx context.Context, accountID string, start, end *time.Time) ([]Invoice, error) {
	path := fmt.Sprintf("/account/%s/charges/invoices", accountID)
	params := map[string]string{}
	if start != nil {
		params["start"] = start.Format(time.RFC3339)
	}
	if end != nil {
		params["end"] = end.Format(time.RFC3339)
	}
	if len(params) > 0 {
		path = buildPath(path, params)
	}

	var result struct {
		Items []Invoice `json:"items"`
	}
	if err := s.doBilling(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// GetInvoiceJSON returns a specific invoice as JSON.
func (s *BillingService) GetInvoiceJSON(ctx context.Context, accountID, invoiceID string) (*Invoice, error) {
	path := fmt.Sprintf("/account/%s/charges/invoices/%s", accountID, invoiceID)
	var result Invoice
	if err := s.doBilling(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetInvoicePDF returns a specific invoice as PDF bytes.
func (s *BillingService) GetInvoicePDF(ctx context.Context, accountID, invoiceID string) ([]byte, error) {
	path := fmt.Sprintf("/account/%s/charges/invoices/%s", accountID, invoiceID)
	return s.doBillingRaw(ctx, "GET", path, nil, "application/pdf")
}

// --- Account Operations ---

// CreateAccount creates a new billing account.
func (s *BillingService) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*CreateAccountResponse, error) {
	var result CreateAccountResponse
	if err := s.doBilling(ctx, "POST", "/account", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAccount returns account details by ID.
func (s *BillingService) GetAccount(ctx context.Context, accountID string) (*BillingAccount, error) {
	path := fmt.Sprintf("/account/%s", accountID)
	var result BillingAccount
	if err := s.doBilling(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AcceptTOU accepts the Terms of Use for an account.
func (s *BillingService) AcceptTOU(ctx context.Context, accountID string, req *AcceptTOURequest) error {
	path := fmt.Sprintf("/account/%s/tou", accountID)
	return s.doBilling(ctx, "POST", path, req, nil)
}
