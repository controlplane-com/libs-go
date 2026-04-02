//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/api"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/serviceaccount"
)

// ServiceAccountTestSuite tests service account CRUD operations.
type ServiceAccountTestSuite struct {
	IntegrationSuite
}

func TestServiceAccountSuite(t *testing.T) {
	suite.Run(t, new(ServiceAccountTestSuite))
}

// newTestServiceAccount creates a minimal service account for testing.
func newTestServiceAccount(name string) *serviceaccount.ServiceAccount {
	return &serviceaccount.ServiceAccount{
		Name:        base.Name(name),
		Description: "Test service account created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: serviceaccount.ServiceAccountTags{
			"test":        "true",
			"integration": "libs-go",
		},
	}
}

// TestServiceAccount_Create tests creating a new service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-create-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err, "Failed to create service account")

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Verify it exists by fetching it
	fetched, err := s.Client().ServiceAccounts().Get(ctx, s.TestOrg(), saName)
	s.Require().NoError(err)
	s.Equal(saName, string(fetched.Name))
	s.Equal("serviceaccount", string(fetched.Kind))
}

// TestServiceAccount_Get tests getting a specific service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-get-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Get the service account
	fetched, err := s.Client().ServiceAccounts().Get(ctx, s.TestOrg(), saName)
	s.Require().NoError(err)
	s.Equal(saName, string(fetched.Name))
	s.Equal("serviceaccount", string(fetched.Kind))
	s.Equal("Test service account created by integration test", fetched.Description)
}

// TestServiceAccount_Update tests updating a service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-update-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Update the service account
	sa.Description = "Updated description"
	sa.Tags["updated"] = "true"

	updated, err := s.Client().ServiceAccounts().Update(ctx, s.TestOrg(), saName, sa)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestServiceAccount_Delete tests deleting a service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-delete-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	// Delete
	err = s.Client().ServiceAccounts().Delete(ctx, s.TestOrg(), saName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().ServiceAccounts().Get(ctx, s.TestOrg(), saName)
	s.Require().Error(err, "Service account should not exist after deletion")
}

// TestServiceAccount_List tests listing service accounts with iterator.
func (s *ServiceAccountTestSuite) TestServiceAccount_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-list-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// List using iterator
	found := false
	for account, err := range s.Client().ServiceAccounts().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		if string(account.Name) == saName {
			found = true
			break
		}
	}
	s.True(found, "Service account should appear in list")
}

// TestServiceAccount_ListAll tests listing all service accounts at once.
func (s *ServiceAccountTestSuite) TestServiceAccount_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-listall-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	accounts, err := s.Client().ServiceAccounts().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(accounts, "Should have at least one service account")

	found := false
	for _, account := range accounts {
		if string(account.Name) == saName {
			found = true
			break
		}
	}
	s.True(found, "Service account should be in list")
}

// TestServiceAccount_Query tests querying service accounts with filters.
func (s *ServiceAccountTestSuite) TestServiceAccount_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-query-" + randomSuffix()
	sa := newTestServiceAccount(saName)
	sa.Tags["querytest"] = "yes"

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Query for service accounts with the tag
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
	for account, err := range s.Client().ServiceAccounts().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(account.Name) == saName {
			found = true
			break
		}
	}
	s.True(found, "Should find service account with query")
}

// TestServiceAccount_QueryAll tests querying all service accounts with filters.
func (s *ServiceAccountTestSuite) TestServiceAccount_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-queryall-" + randomSuffix()
	sa := newTestServiceAccount(saName)
	sa.Tags["queryalltest"] = "yes"

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Query for service accounts with the tag
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

	accounts, err := s.Client().ServiceAccounts().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)

	found := false
	for _, account := range accounts {
		if string(account.Name) == saName {
			found = true
			break
		}
	}
	s.True(found, "Should find service account with QueryAll")
}

// TestServiceAccount_ListPage tests paginated listing of service accounts.
func (s *ServiceAccountTestSuite) TestServiceAccount_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-listpage-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Get first page
	page, err := s.Client().ServiceAccounts().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(page, "Should return a page")
	s.NotEmpty(page.Items, "Should have at least one service account")
}

// TestServiceAccount_GetNotFound tests getting a non-existent service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_GetNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().ServiceAccounts().Get(ctx, s.TestOrg(), "nonexistent-sa-"+randomSuffix())
	s.Require().Error(err, "Should return error for non-existent service account")
}

// TestServiceAccount_DeleteNotFound tests deleting a non-existent service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_DeleteNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	err := s.Client().ServiceAccounts().Delete(ctx, s.TestOrg(), "nonexistent-sa-"+randomSuffix())
	s.Require().Error(err, "Should return error for deleting non-existent service account")
}

// TestServiceAccount_CreateDuplicate tests creating a duplicate service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_CreateDuplicate() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-dup-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Try to create again with the same name
	_, err = s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().Error(err, "Should return error for duplicate service account")
}

// TestServiceAccount_AddKey tests adding a key to a service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_AddKey() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-addkey-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Add a key
	req := &api.AddKeyRequest{
		Description: "Test key for integration test",
	}
	resp, err := s.Client().ServiceAccounts().AddKey(ctx, s.TestOrg(), saName, req)
	s.Require().NoError(err)
	s.NotEmpty(resp.Key, "Should return a key")

	// Verify the service account now has a key
	fetched, err := s.Client().ServiceAccounts().Get(ctx, s.TestOrg(), saName)
	s.Require().NoError(err)
	s.NotEmpty(fetched.Keys, "Service account should have at least one key")

	// Verify the key description
	foundKey := false
	for _, key := range fetched.Keys {
		if key.Description == "Test key for integration test" {
			foundKey = true
			break
		}
	}
	s.True(foundKey, "Should find the key with the expected description")
}

// TestServiceAccount_AddMultipleKeys tests adding multiple keys to a service account.
func (s *ServiceAccountTestSuite) TestServiceAccount_AddMultipleKeys() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-multikey-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Add first key
	req1 := &api.AddKeyRequest{
		Description: "First test key",
	}
	resp1, err := s.Client().ServiceAccounts().AddKey(ctx, s.TestOrg(), saName, req1)
	s.Require().NoError(err)
	s.NotEmpty(resp1.Key, "Should return first key")

	// Add second key
	req2 := &api.AddKeyRequest{
		Description: "Second test key",
	}
	resp2, err := s.Client().ServiceAccounts().AddKey(ctx, s.TestOrg(), saName, req2)
	s.Require().NoError(err)
	s.NotEmpty(resp2.Key, "Should return second key")

	// Verify the keys are different
	s.NotEqual(resp1.Key, resp2.Key, "Keys should be different")

	// Verify the service account now has two keys
	fetched, err := s.Client().ServiceAccounts().Get(ctx, s.TestOrg(), saName)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(fetched.Keys), 2, "Service account should have at least two keys")
}

// TestServiceAccount_UpdateTags tests updating service account tags.
func (s *ServiceAccountTestSuite) TestServiceAccount_UpdateTags() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	saName := "test-sa-tags-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Update with new tags
	sa.Tags = serviceaccount.ServiceAccountTags{
		"environment": "production",
		"team":        "platform",
		"managed-by":  "automation",
	}

	updated, err := s.Client().ServiceAccounts().Update(ctx, s.TestOrg(), saName, sa)
	s.Require().NoError(err)
	s.Equal("production", updated.Tags["environment"])
	s.Equal("platform", updated.Tags["team"])
	s.Equal("automation", updated.Tags["managed-by"])
}
