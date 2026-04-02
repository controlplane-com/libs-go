//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/gvc"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// GVCTestSuite tests GVC CRUD operations.
type GVCTestSuite struct {
	IntegrationSuite
}

func TestGVCSuite(t *testing.T) {
	suite.Run(t, new(GVCTestSuite))
}

// TestGVC_Create tests creating a new GVC.
func (s *GVCTestSuite) TestGVC_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	gvcName := "test-gvc-create-" + randomSuffix()

	g := &gvc.Gvc{
		Name:        base.Name(gvcName),
		Description: "Test GVC created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: gvc.GvcTags{
			"test":        "true",
			"integration": "libs-go",
		},
	}

	_, err := s.Client().GVCs().Create(ctx, s.TestOrg(), g)
	s.Require().NoError(err, "Failed to create GVC")

	// Cleanup
	defer func() {
		_ = s.Client().GVCs().Delete(context.Background(), s.TestOrg(), gvcName)
	}()

	// Verify it exists by fetching it
	fetched, err := s.Client().GVCs().Get(ctx, s.TestOrg(), gvcName)
	s.Require().NoError(err)
	s.Equal(gvcName, string(fetched.Name))
	s.Equal("Test GVC created by integration test", fetched.Description)
}

// TestGVC_Get tests getting a specific GVC.
func (s *GVCTestSuite) TestGVC_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get the test GVC created by the suite
	g, err := s.Client().GVCs().Get(ctx, s.TestOrg(), s.TestGVC())
	s.Require().NoError(err)
	s.Equal(s.TestGVC(), string(g.Name))
	s.Equal("gvc", string(g.Kind))
}

// TestGVC_Update tests updating a GVC.
func (s *GVCTestSuite) TestGVC_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	gvcName := "test-gvc-update-" + randomSuffix()

	// Create GVC
	g := &gvc.Gvc{
		Name:        base.Name(gvcName),
		Description: "Original description",
		Version:     common.Float32Ptr(1),
	}

	_, err := s.Client().GVCs().Create(ctx, s.TestOrg(), g)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().GVCs().Delete(context.Background(), s.TestOrg(), gvcName)
	}()

	// Update GVC
	g.Description = "Updated description"
	g.Tags = gvc.GvcTags{"updated": "true"}

	updated, err := s.Client().GVCs().Update(ctx, s.TestOrg(), gvcName, g)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestGVC_Delete tests deleting a GVC.
func (s *GVCTestSuite) TestGVC_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	gvcName := "test-gvc-delete-" + randomSuffix()

	// Create GVC
	g := &gvc.Gvc{
		Name:        base.Name(gvcName),
		Description: "GVC to be deleted",
		Version:     common.Float32Ptr(1),
	}

	_, err := s.Client().GVCs().Create(ctx, s.TestOrg(), g)
	s.Require().NoError(err)

	// Delete GVC
	err = s.Client().GVCs().Delete(ctx, s.TestOrg(), gvcName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().GVCs().Get(ctx, s.TestOrg(), gvcName)
	s.Require().Error(err, "GVC should not exist after deletion")
}

// TestGVC_List tests listing GVCs with iterator.
func (s *GVCTestSuite) TestGVC_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List using iterator
	found := false
	for g, err := range s.Client().GVCs().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		if string(g.Name) == s.TestGVC() {
			found = true
			break
		}
	}
	s.True(found, "Test GVC should appear in list")
}

// TestGVC_ListAll tests listing all GVCs at once.
func (s *GVCTestSuite) TestGVC_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	gvcs, err := s.Client().GVCs().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(gvcs, "Should have at least one GVC")

	// Find our test GVC
	found := false
	for _, g := range gvcs {
		if string(g.Name) == s.TestGVC() {
			found = true
			break
		}
	}
	s.True(found, "Test GVC should be in list")
}

// TestGVC_ListPage tests paginated listing.
func (s *GVCTestSuite) TestGVC_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	resp, err := s.Client().GVCs().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotEmpty(resp.Items, "Should have at least one GVC")
}

// TestGVC_Query tests querying GVCs with filters.
func (s *GVCTestSuite) TestGVC_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Create a GVC with specific tags for querying
	gvcName := "test-gvc-query-" + randomSuffix()
	g := &gvc.Gvc{
		Name:        base.Name(gvcName),
		Description: "GVC for query test",
		Version:     common.Float32Ptr(1),
		Tags: gvc.GvcTags{
			"querytest": "yes",
		},
	}

	_, err := s.Client().GVCs().Create(ctx, s.TestOrg(), g)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().GVCs().Delete(context.Background(), s.TestOrg(), gvcName)
	}()

	// Query for GVCs with the tag
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
	for gvcResult, err := range s.Client().GVCs().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(gvcResult.Name) == gvcName {
			found = true
			break
		}
	}
	s.True(found, "Should find GVC with query")
}

// TestGVC_QueryAll tests querying all GVCs matching criteria.
func (s *GVCTestSuite) TestGVC_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all GVCs (empty query matches all)
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	gvcs, err := s.Client().GVCs().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotEmpty(gvcs, "Should have at least one GVC")
}

// TestGVC_ListPage_NextPage tests pagination using the NextPage method.
func (s *GVCTestSuite) TestGVC_ListPage_NextPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	page, err := s.Client().GVCs().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(page)

	// Test NextPage returns nil when no more pages (or returns next page if exists)
	if page.HasNext() {
		nextPage, err := page.NextPage(ctx)
		s.Require().NoError(err)
		s.NotNil(nextPage, "NextPage should return a valid page when HasNext is true")
	} else {
		nextPage, err := page.NextPage(ctx)
		s.Require().NoError(err)
		s.Nil(nextPage, "NextPage should return nil when no more pages")
	}
}

// TestGVC_QueryPage_NextPage tests query pagination using the NextPage method.
func (s *GVCTestSuite) TestGVC_QueryPage_NextPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	page, err := s.Client().GVCs().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(page)

	// Test NextPage
	if page.HasNext() {
		nextPage, err := page.NextPage(ctx)
		s.Require().NoError(err)
		s.NotNil(nextPage, "NextPage should return a valid page when HasNext is true")
	} else {
		nextPage, err := page.NextPage(ctx)
		s.Require().NoError(err)
		s.Nil(nextPage, "NextPage should return nil when no more pages")
	}
}
