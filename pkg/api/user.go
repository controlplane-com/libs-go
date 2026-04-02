package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/user"
)

// UserService handles operations on users.
type UserService struct {
	client *Client
}

// List returns an iterator over all users in the organization.
func (s *UserService) List(ctx context.Context, org string) iter.Seq2[user.User, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user", org)
	return listIterator[user.User](ctx, s.client, path)
}

// ListAll returns all users in the organization.
func (s *UserService) ListAll(ctx context.Context, org string) ([]user.User, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user", org)
	return listAll[user.User](ctx, s.client, path)
}

// ListPage returns a single page of users.
func (s *UserService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[user.User], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPageWithClient[user.User](ctx, s.client, path)
}

// Get returns a user by name.
func (s *UserService) Get(ctx context.Context, org, name string) (*user.User, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user/%s", org, name)
	var result user.User
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a user from the organization.
func (s *UserService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user/%s", org, name)
	return s.client.delete(ctx, path)
}

// InviteRequest is the request body for inviting a user.
type InviteRequest struct {
	Email string `json:"email"`
}

// Invite invites a user to the organization.
func (s *UserService) Invite(ctx context.Context, org string, req *InviteRequest) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user/-invite", org)
	return s.client.post(ctx, path, req, nil)
}

// Query returns an iterator over users matching the query.
func (s *UserService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[user.User, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user/-query", org)
	return queryIterator[user.User](ctx, s.client, path, q)
}

// QueryAll returns all users matching the query.
func (s *UserService) QueryAll(ctx context.Context, org string, q *query.Query) ([]user.User, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user/-query", org)
	return queryAll[user.User](ctx, s.client, path, q)
}

// QueryPage returns a single page of users matching the query.
func (s *UserService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[user.User], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/user/-query", org)
	return queryPageWithClient[user.User](ctx, s.client, path, q)
}
