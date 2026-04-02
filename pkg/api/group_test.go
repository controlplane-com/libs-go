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
	"github.com/controlplane-com/libs-go/pkg/schema/group"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// GroupTestSuite tests group CRUD operations.
type GroupTestSuite struct {
	IntegrationSuite
}

func TestGroupSuite(t *testing.T) {
	suite.Run(t, new(GroupTestSuite))
}

// newTestGroup creates a minimal group for testing.
func newTestGroup(name string) *group.Group {
	return &group.Group{
		Name:        base.Name(name),
		Description: "Test group created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: group.GroupTags{
			"test":        "true",
			"integration": "libs-go",
		},
		IdentityMatcher: group.GroupIdentityMatcher{
			Expression: "false", // A valid expression that matches nothing
			Language:   group.GroupIdentityMatcherLanguageJavascript,
		},
	}
}

// TestGroup_Create tests creating a new group.
func (s *GroupTestSuite) TestGroup_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-create-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err, "Failed to create group")

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Verify it exists by fetching it
	fetched, err := s.Client().Groups().Get(ctx, s.TestOrg(), groupName)
	s.Require().NoError(err)
	s.Equal(groupName, string(fetched.Name))
	s.Equal("group", string(fetched.Kind))
}

// TestGroup_Get tests getting a specific group.
func (s *GroupTestSuite) TestGroup_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-get-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Get the group
	fetched, err := s.Client().Groups().Get(ctx, s.TestOrg(), groupName)
	s.Require().NoError(err)
	s.Equal(groupName, string(fetched.Name))
	s.Equal("group", string(fetched.Kind))
	s.Equal("Test group created by integration test", fetched.Description)
}

// TestGroup_Update tests updating a group.
func (s *GroupTestSuite) TestGroup_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-update-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Update the group
	grp.Description = "Updated description"
	grp.Tags["updated"] = "true"

	updated, err := s.Client().Groups().Update(ctx, s.TestOrg(), groupName, grp)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestGroup_Delete tests deleting a group.
func (s *GroupTestSuite) TestGroup_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-delete-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	// Delete
	err = s.Client().Groups().Delete(ctx, s.TestOrg(), groupName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().Groups().Get(ctx, s.TestOrg(), groupName)
	s.Require().Error(err, "Group should not exist after deletion")
}

// TestGroup_List tests listing groups with iterator.
func (s *GroupTestSuite) TestGroup_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-list-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// List using iterator
	found := false
	for g, err := range s.Client().Groups().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		if string(g.Name) == groupName {
			found = true
			break
		}
	}
	s.True(found, "Group should appear in list")
}

// TestGroup_ListAll tests listing all groups at once.
func (s *GroupTestSuite) TestGroup_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-listall-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	groups, err := s.Client().Groups().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(groups, "Should have at least one group")

	found := false
	for _, g := range groups {
		if string(g.Name) == groupName {
			found = true
			break
		}
	}
	s.True(found, "Group should be in list")
}

// TestGroup_Query tests querying groups with filters.
func (s *GroupTestSuite) TestGroup_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-query-" + randomSuffix()
	grp := newTestGroup(groupName)
	grp.Tags["querytest"] = "yes"

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Query for groups with the tag
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
	for g, err := range s.Client().Groups().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(g.Name) == groupName {
			found = true
			break
		}
	}
	s.True(found, "Should find group with query")
}

// TestGroup_QueryAll tests querying all groups with filters.
func (s *GroupTestSuite) TestGroup_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-queryall-" + randomSuffix()
	grp := newTestGroup(groupName)
	grp.Tags["queryalltest"] = "yes"

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Query for groups with the tag
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

	groups, err := s.Client().Groups().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)

	found := false
	for _, g := range groups {
		if string(g.Name) == groupName {
			found = true
			break
		}
	}
	s.True(found, "Should find group with QueryAll")
}

// TestGroup_ListPage tests paginated listing of groups.
func (s *GroupTestSuite) TestGroup_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-listpage-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Get first page
	page, err := s.Client().Groups().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(page, "Should return a page")
	s.NotEmpty(page.Items, "Should have at least one group")
}

// TestGroup_GetNotFound tests getting a non-existent group.
func (s *GroupTestSuite) TestGroup_GetNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Groups().Get(ctx, s.TestOrg(), "nonexistent-group-"+randomSuffix())
	s.Require().Error(err, "Should return error for non-existent group")
}

// TestGroup_DeleteNotFound tests deleting a non-existent group.
func (s *GroupTestSuite) TestGroup_DeleteNotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	err := s.Client().Groups().Delete(ctx, s.TestOrg(), "nonexistent-group-"+randomSuffix())
	s.Require().Error(err, "Should return error for deleting non-existent group")
}

// TestGroup_CreateDuplicate tests creating a duplicate group.
func (s *GroupTestSuite) TestGroup_CreateDuplicate() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-dup-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Try to create again with the same name
	_, err = s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().Error(err, "Should return error for duplicate group")
}

// TestGroup_WithMemberLinks tests creating a group with member links.
func (s *GroupTestSuite) TestGroup_WithMemberLinks() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// First create a service account to add as a member
	saName := "test-sa-for-group-" + randomSuffix()
	sa := newTestServiceAccount(saName)

	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName)
	}()

	// Create a group with the service account as a member
	groupName := "test-group-members-" + randomSuffix()
	grp := &group.Group{
		Name:        base.Name(groupName),
		Description: "Test group with member links",
		Version:     common.Float32Ptr(1),
		Tags: group.GroupTags{
			"test": "true",
		},
		MemberLinks: []string{fmt.Sprintf("/org/%s/serviceaccount/%s", s.TestOrg(), saName)},
		IdentityMatcher: group.GroupIdentityMatcher{
			Expression: "false",
			Language:   group.GroupIdentityMatcherLanguageJavascript,
		},
	}

	_, err = s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Verify the group has the member
	fetched, err := s.Client().Groups().Get(ctx, s.TestOrg(), groupName)
	s.Require().NoError(err)
	s.NotEmpty(fetched.MemberLinks, "Group should have member links")
	s.Contains(fetched.MemberLinks[0], saName, "Member link should reference the service account")
}

// TestGroup_UpdateMemberLinks tests updating group member links.
func (s *GroupTestSuite) TestGroup_UpdateMemberLinks() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Create two service accounts
	saName1 := "test-sa-grp1-" + randomSuffix()
	sa1 := newTestServiceAccount(saName1)
	_, err := s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa1)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName1)
	}()

	saName2 := "test-sa-grp2-" + randomSuffix()
	sa2 := newTestServiceAccount(saName2)
	_, err = s.Client().ServiceAccounts().Create(ctx, s.TestOrg(), sa2)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().ServiceAccounts().Delete(context.Background(), s.TestOrg(), saName2)
	}()

	// Create a group with the first service account
	groupName := "test-group-updmembers-" + randomSuffix()
	grp := &group.Group{
		Name:        base.Name(groupName),
		Description: "Test group for updating members",
		Version:     common.Float32Ptr(1),
		Tags: group.GroupTags{
			"test": "true",
		},
		MemberLinks: []string{fmt.Sprintf("/org/%s/serviceaccount/%s", s.TestOrg(), saName1)},
		IdentityMatcher: group.GroupIdentityMatcher{
			Expression: "false",
			Language:   group.GroupIdentityMatcherLanguageJavascript,
		},
	}

	_, err = s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Update the group to include both service accounts
	grp.MemberLinks = []string{
		fmt.Sprintf("/org/%s/serviceaccount/%s", s.TestOrg(), saName1),
		fmt.Sprintf("/org/%s/serviceaccount/%s", s.TestOrg(), saName2),
	}

	updated, err := s.Client().Groups().Update(ctx, s.TestOrg(), groupName, grp)
	s.Require().NoError(err)
	s.Len(updated.MemberLinks, 2, "Group should have two member links")
}

// TestGroup_WithIdentityMatcher tests creating a group with an identity matcher.
func (s *GroupTestSuite) TestGroup_WithIdentityMatcher() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-matcher-" + randomSuffix()
	grp := &group.Group{
		Name:        base.Name(groupName),
		Description: "Test group with identity matcher",
		Version:     common.Float32Ptr(1),
		Tags: group.GroupTags{
			"test": "true",
		},
		IdentityMatcher: group.GroupIdentityMatcher{
			Expression: "email && email.endsWith('@example.com')",
			Language:   group.GroupIdentityMatcherLanguageJavascript,
		},
	}

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Verify the group has the identity matcher
	fetched, err := s.Client().Groups().Get(ctx, s.TestOrg(), groupName)
	s.Require().NoError(err)
	s.NotEmpty(fetched.IdentityMatcher.Expression, "Group should have identity matcher expression")
	s.Equal(group.GroupIdentityMatcherLanguageJavascript, fetched.IdentityMatcher.Language)
}

// TestGroup_UpdateTags tests updating group tags.
func (s *GroupTestSuite) TestGroup_UpdateTags() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	groupName := "test-group-tags-" + randomSuffix()
	grp := newTestGroup(groupName)

	_, err := s.Client().Groups().Create(ctx, s.TestOrg(), grp)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Groups().Delete(context.Background(), s.TestOrg(), groupName)
	}()

	// Update with new tags
	grp.Tags = group.GroupTags{
		"environment": "production",
		"team":        "platform",
		"managed-by":  "automation",
	}

	updated, err := s.Client().Groups().Update(ctx, s.TestOrg(), groupName, grp)
	s.Require().NoError(err)
	s.Equal("production", updated.Tags["environment"])
	s.Equal("platform", updated.Tags["team"])
	s.Equal("automation", updated.Tags["managed-by"])
}
