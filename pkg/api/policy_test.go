//go:build integration

package api_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/policy"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// PolicyTestSuite tests policy CRUD operations.
type PolicyTestSuite struct {
	IntegrationSuite
}

func TestPolicySuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

// newTestPolicy creates a minimal policy for testing.
func newTestPolicy(name, org string) *policy.Policy {
	return &policy.Policy{
		Name:        base.Name(name),
		Description: "Test policy created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: policy.PolicyTags{
			"test":        "true",
			"integration": "libs-go",
		},
		TargetKind: "secret",
		Target:     policy.PolicyTargetAll,
		Bindings: []policy.Binding{
			{
				Permissions:    []string{"view"},
				PrincipalLinks: []string{fmt.Sprintf("/org/%s/group/superusers", org)},
			},
		},
	}
}

// TestPolicy_Create tests creating a new policy.
func (s *PolicyTestSuite) TestPolicy_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-create-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err, "Failed to create policy")

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Verify it exists by fetching it
	fetched, err := s.Client().Policies().Get(ctx, s.TestOrg(), policyName)
	s.Require().NoError(err)
	s.Equal(policyName, string(fetched.Name))
	s.Equal("policy", string(fetched.Kind))
}

// TestPolicy_Get tests getting a specific policy.
func (s *PolicyTestSuite) TestPolicy_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-get-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Get the policy
	fetched, err := s.Client().Policies().Get(ctx, s.TestOrg(), policyName)
	s.Require().NoError(err)
	s.Equal(policyName, string(fetched.Name))
	s.Equal("policy", string(fetched.Kind))
	s.Equal("Test policy created by integration test", fetched.Description)
	s.Equal(base.Kind("secret"), fetched.TargetKind)
}

// TestPolicy_Update tests updating a policy.
func (s *PolicyTestSuite) TestPolicy_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-update-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Update the policy
	pol.Description = "Updated description"
	pol.Tags["updated"] = "true"

	updated, err := s.Client().Policies().Update(ctx, s.TestOrg(), policyName, pol)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestPolicy_Delete tests deleting a policy.
func (s *PolicyTestSuite) TestPolicy_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-delete-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	// Delete
	err = s.Client().Policies().Delete(ctx, s.TestOrg(), policyName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().Policies().Get(ctx, s.TestOrg(), policyName)
	s.Require().Error(err, "Policy should not exist after deletion")
}

// TestPolicy_List tests listing policies with iterator.
func (s *PolicyTestSuite) TestPolicy_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-list-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// List using iterator
	found := false
	for p, err := range s.Client().Policies().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		if string(p.Name) == policyName {
			found = true
			break
		}
	}
	s.True(found, "Policy should appear in list")
}

// TestPolicy_ListAll tests listing all policies at once.
func (s *PolicyTestSuite) TestPolicy_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-listall-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	policies, err := s.Client().Policies().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(policies, "Should have at least one policy")

	found := false
	for _, p := range policies {
		if string(p.Name) == policyName {
			found = true
			break
		}
	}
	s.True(found, "Policy should be in list")
}

// TestPolicy_Query tests querying policies with filters.
func (s *PolicyTestSuite) TestPolicy_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-query-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())
	pol.Tags["querytest"] = "yes"

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Query for policies with the tag
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
	for p, err := range s.Client().Policies().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(p.Name) == policyName {
			found = true
			break
		}
	}
	s.True(found, "Should find policy with query")
}

// TestPolicy_QueryAll tests querying all policies with filters.
func (s *PolicyTestSuite) TestPolicy_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-queryall-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())
	pol.Tags["queryalltest"] = "yes"

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Query for policies with the tag
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

	policies, err := s.Client().Policies().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)

	found := false
	for _, p := range policies {
		if string(p.Name) == policyName {
			found = true
			break
		}
	}
	s.True(found, "Should find policy with QueryAll")
}

// TestPolicy_ListPage tests paginated listing of policies.
func (s *PolicyTestSuite) TestPolicy_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-listpage-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Get first page
	page, err := s.Client().Policies().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(page, "Should return a page")
	s.NotEmpty(page.Items, "Should have at least one policy")
}

// TestPolicy_GetNotFound tests getting a non-existent policy.
func (s *PolicyTestSuite) TestPolicy_GetNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Policies().Get(ctx, s.TestOrg(), "nonexistent-policy-"+randomSuffix())
	s.Require().Error(err, "Should return error for non-existent policy")
}

// TestPolicy_DeleteNotFound tests deleting a non-existent policy.
func (s *PolicyTestSuite) TestPolicy_DeleteNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	err := s.Client().Policies().Delete(ctx, s.TestOrg(), "nonexistent-policy-"+randomSuffix())
	s.Require().Error(err, "Should return error for deleting non-existent policy")
}

// TestPolicy_CreateDuplicate tests creating a duplicate policy.
func (s *PolicyTestSuite) TestPolicy_CreateDuplicate() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-dup-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Try to create again with the same name
	_, err = s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().Error(err, "Should return error for duplicate policy")
}

// TestPolicy_WithTargetLinks tests creating a policy with specific target links.
func (s *PolicyTestSuite) TestPolicy_WithTargetLinks() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// First create a secret to target
	secretName := "test-secret-for-policy-" + randomSuffix()
	sec := newTestSecret(secretName)

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	// Create a policy targeting the specific secret
	policyName := "test-policy-targets-" + randomSuffix()
	pol := &policy.Policy{
		Name:        base.Name(policyName),
		Description: "Test policy with target links",
		Version:     common.Float32Ptr(1),
		Tags: policy.PolicyTags{
			"test": "true",
		},
		TargetKind:  "secret",
		TargetLinks: []string{fmt.Sprintf("/org/%s/secret/%s", s.TestOrg(), secretName)},
		Bindings: []policy.Binding{
			{
				Permissions:    []string{"view", "reveal"},
				PrincipalLinks: []string{fmt.Sprintf("/org/%s/group/superusers", s.TestOrg())},
			},
		},
	}

	_, err = s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Verify the policy
	fetched, err := s.Client().Policies().Get(ctx, s.TestOrg(), policyName)
	s.Require().NoError(err)
	s.NotEmpty(fetched.TargetLinks, "Policy should have target links")
	s.Contains(fetched.TargetLinks[0], secretName, "Target link should reference the secret")
}

// TestPolicy_UpdateBindings tests updating policy bindings.
func (s *PolicyTestSuite) TestPolicy_UpdateBindings() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	policyName := "test-policy-bindings-" + randomSuffix()
	pol := newTestPolicy(policyName, s.TestOrg())

	_, err := s.Client().Policies().Create(ctx, s.TestOrg(), pol)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Policies().Delete(context.Background(), s.TestOrg(), policyName)
	}()

	// Update bindings to add more permissions
	pol.Bindings = []policy.Binding{
		{
			Permissions:    []string{"view", "edit", "manage"},
			PrincipalLinks: []string{fmt.Sprintf("/org/%s/group/superusers", s.TestOrg())},
		},
	}

	updated, err := s.Client().Policies().Update(ctx, s.TestOrg(), policyName, pol)
	s.Require().NoError(err)
	s.Len(updated.Bindings, 1, "Should have one binding")
	s.Contains(updated.Bindings[0].Permissions, "edit", "Should have edit permission")
	s.Contains(updated.Bindings[0].Permissions, "manage", "Should have manage permission")
}
