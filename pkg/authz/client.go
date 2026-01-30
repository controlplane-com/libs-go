package authz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/controlplane-com/libs-go/pkg/errors"
)

// Client is an HTTP client for the authz service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ClientOption configures the client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new authz client
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Authenticate authenticates a token using the specified profile
func (c *Client) Authenticate(ctx context.Context, req *AuthenticateRequest) (*AuthenticateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/authenticate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var authResp AuthenticateResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &authResp, nil
}

// Authorize authenticates and authorizes a token using the specified profile
func (c *Client) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/authorize", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var authResp AuthorizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &authResp, nil
}

// AuthorizeAnyValidUser is a convenience method that authorizes any valid user
func (c *Client) AuthorizeAnyValidUser(ctx context.Context, token string, actions ...string) (*User, error) {
	resp, err := c.Authorize(ctx, &AuthorizeRequest{
		Token:       token,
		Profile:     string(ProfileAnyValidUser),
		Permissions: actions,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Authorized {
		if resp.Error != "" {
			return nil, fmt.Errorf("authorization failed: %s", resp.Error)
		}
		return nil, errors.UnauthorizedError
	}
	return resp.User, nil
}

// AuthorizeRootUser is a convenience method that authorizes root users only
func (c *Client) AuthorizeRootUser(ctx context.Context, token string) (*User, error) {
	resp, err := c.Authorize(ctx, &AuthorizeRequest{
		Token:   token,
		Profile: string(ProfileRootUser),
	})
	if err != nil {
		return nil, err
	}
	if !resp.Authorized {
		if resp.Error != "" {
			return nil, fmt.Errorf("authorization failed: %s", resp.Error)
		}
		return nil, errors.UnauthorizedError
	}
	return resp.User, nil
}

// AuthorizeAccountUser is a convenience method that authorizes account users
func (c *Client) AuthorizeAccountUser(ctx context.Context, token, accountId string, roles ...string) (*User, error) {
	// Format permissions as "accountId/role"
	permissions := make([]string, len(roles))
	for i, role := range roles {
		permissions[i] = fmt.Sprintf("%s/%s", accountId, role)
	}

	resp, err := c.Authorize(ctx, &AuthorizeRequest{
		Token:       token,
		Profile:     string(ProfileAccountUser),
		Permissions: permissions,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Authorized {
		if resp.Error != "" {
			return nil, fmt.Errorf("authorization failed: %s", resp.Error)
		}
		return nil, errors.UnauthorizedError
	}
	return resp.User, nil
}

// AuthorizeOrgUser is a convenience method that authorizes users for an org with given permissions
func (c *Client) AuthorizeOrgUser(ctx context.Context, token, org string, permissions []string) (*User, error) {
	resp, err := c.Authorize(ctx, &AuthorizeRequest{
		Token:       token,
		Profile:     string(ProfileDataService),
		Scope:       org,
		Permissions: permissions,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Authorized {
		if resp.Error != "" {
			return nil, fmt.Errorf("authorization failed: %s", resp.Error)
		}
		return nil, errors.UnauthorizedError
	}
	return resp.User, nil
}

// AuthorizeMetering is a convenience method for metering service authorization
func (c *Client) AuthorizeMetering(ctx context.Context, token, scope string, actions ...string) (*User, error) {
	resp, err := c.Authorize(ctx, &AuthorizeRequest{
		Token:       token,
		Profile:     string(ProfileMetering),
		Scope:       scope,
		Permissions: actions,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Authorized {
		if resp.Error != "" {
			return nil, fmt.Errorf("authorization failed: %s", resp.Error)
		}
		return nil, errors.UnauthorizedError
	}
	return resp.User, nil
}
