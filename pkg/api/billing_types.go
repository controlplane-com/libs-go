package api

import "time"

// BillingAccount represents a billing account.
type BillingAccount struct {
	ID          string            `json:"id,omitempty"`
	AccountName string            `json:"fullName,omitempty"`
	Description string            `json:"accountName,omitempty"`
	Email       string            `json:"email,omitempty"`
	Enabled     bool              `json:"enabled"`
	MainOrg     string            `json:"mainOrg,omitempty"`
	OrgQuota    int               `json:"orgQuota,omitempty"`
	Created     *time.Time        `json:"created,omitempty"`
	ExtraInfo   map[string]string `json:"extraInfo,omitempty"`
}

// BillingOrg represents an organization in the billing system.
type BillingOrg struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	AccountID   string            `json:"accountId,omitempty"`
	Created     time.Time         `json:"created,omitempty"`
}

// CreateOrgRequest is the request body for creating an org via billing-ng.
type CreateOrgRequest struct {
	Org      *BillingOrg `json:"org"`
	Invitees []string    `json:"invitees,omitempty"`
}

// CreateOrgResponse is the response from creating an org.
// Note: The API embeds the Org type, so "created" is actually a timestamp, not a boolean.
// The boolean "Created" field from billing-ng is serialized with capital C.
type CreateOrgResponse struct {
	BillingOrg
	WasCreated bool `json:"Created,omitempty"` // Whether the org was newly created (vs already existed)
}

// ChargeQueryRequest is the request for querying charges.
type ChargeQueryRequest struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	TimeStep  string    `json:"timeStep,omitempty"` // hour, day, week, month
	Orgs      []string  `json:"orgs,omitempty"`
	Detailed  bool      `json:"detailed,omitempty"`
}

// ChargeQueryResponse is the response from a charge query.
type ChargeQueryResponse struct {
	Periods         []ChargePeriod `json:"periods,omitempty"`
	DetailedResults []any          `json:"detailedResults,omitempty"`
}

// ChargePeriod represents charges for a time period.
type ChargePeriod struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Groups    []any     `json:"groups,omitempty"`
}

// Invoice represents a billing invoice.
type Invoice struct {
	ID        string     `json:"id,omitempty"`
	AccountID string     `json:"accountId,omitempty"`
	Status    string     `json:"status,omitempty"`
	Total     float64    `json:"total,omitempty"`
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
	Created   *time.Time `json:"created,omitempty"`
}

// BillingAddress represents a billing address.
type BillingAddress struct {
	Name       string `json:"name,omitempty"`
	Address1   string `json:"address1,omitempty"`
	Address2   string `json:"address2,omitempty"`
	City       string `json:"city,omitempty"`
	Country    string `json:"country,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	State      string `json:"state,omitempty"`
}

// CreateAccountRequest is the request body for creating a billing account.
type CreateAccountRequest struct {
	Account *BillingAccountInput `json:"account"`
	Type    string               `json:"type"`
}

// BillingAccountInput is the input for creating/updating a billing account.
type BillingAccountInput struct {
	AccountName string            `json:"fullName,omitempty"`
	Description string            `json:"accountName,omitempty"`
	Email       string            `json:"email,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	Address     *BillingAddress   `json:"address,omitempty"`
	ExtraInfo   map[string]string `json:"extraInfo,omitempty"`
}

// CreateAccountResponse is the response from creating an account.
type CreateAccountResponse struct {
	BillingAccount
}

// AcceptTOURequest is the request body for accepting Terms of Use.
type AcceptTOURequest struct {
	TOU *TOUInput `json:"tou"`
}

// TOUInput is the input for accepting Terms of Use.
type TOUInput struct {
	CustomerName string `json:"customerName"`
	Accepted     bool   `json:"accepted"`
}
