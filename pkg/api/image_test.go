//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// ImageTestSuite tests image read operations.
// Images are managed by the registry and are primarily read-only via this API.
// Note: Image creation happens through docker push to the registry, not via API.
type ImageTestSuite struct {
	IntegrationSuite
}

func TestImageSuite(t *testing.T) {
	suite.Run(t, new(ImageTestSuite))
}

// TestImage_List tests listing images with iterator.
func (s *ImageTestSuite) TestImage_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// List using iterator - may be empty if no images pushed
	count := 0
	for img, err := range s.Client().Images().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		s.NotEmpty(img.Name, "Image should have a name")
		count++
		if count >= 5 {
			break // Just verify we can iterate
		}
	}
	// Note: count may be 0 if no images have been pushed to this org
	s.T().Logf("Found %d images in org", count)
}

// TestImage_ListAll tests listing all images at once.
func (s *ImageTestSuite) TestImage_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	images, err := s.Client().Images().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	// Note: images slice may be empty if no images have been pushed

	// If we have images, verify structure
	for _, img := range images {
		s.NotEmpty(img.Name, "Image should have a name")
		s.Equal("image", string(img.Kind), "Kind should be 'image'")
	}
	s.T().Logf("Found %d images total", len(images))
}

// TestImage_ListPage tests paginated listing.
func (s *ImageTestSuite) TestImage_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	resp, err := s.Client().Images().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(resp)
	// Note: Items may be empty if no images
	s.T().Logf("First page has %d images", len(resp.Items))
}

// TestImage_Get tests getting a specific image.
func (s *ImageTestSuite) TestImage_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First, list images to get a valid name
	images, err := s.Client().Images().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(images) == 0 {
		s.T().Skip("No images available to test Get operation")
		return
	}

	imageName := images[0].Name

	// Get the specific image
	img, err := s.Client().Images().Get(ctx, s.TestOrg(), imageName)
	s.Require().NoError(err)
	s.Equal(imageName, img.Name)
	s.Equal("image", string(img.Kind))
}

// TestImage_Get_NotFound tests getting a non-existent image.
func (s *ImageTestSuite) TestImage_Get_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Images().Get(ctx, s.TestOrg(), "nonexistent-image-xyz:v999")
	s.Require().Error(err, "Should error for non-existent image")
}

// TestImage_Query tests querying images with filters.
func (s *ImageTestSuite) TestImage_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Query for all images (empty query matches all)
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	count := 0
	for img, err := range s.Client().Images().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		s.NotEmpty(img.Name, "Image should have a name")
		count++
		if count >= 5 {
			break
		}
	}
	s.T().Logf("Query found %d images", count)
}

// TestImage_QueryAll tests querying all images matching criteria.
func (s *ImageTestSuite) TestImage_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 60*time.Second)
	defer cancel()

	// Query for all images
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	images, err := s.Client().Images().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	// Note: images slice may be empty
	s.T().Logf("QueryAll found %d images", len(images))
}

// TestImage_QueryPage tests paginated querying.
func (s *ImageTestSuite) TestImage_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	resp, err := s.Client().Images().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.T().Logf("QueryPage found %d images", len(resp.Items))
}

// TestImage_VerifyStructure tests that images have expected fields when available.
func (s *ImageTestSuite) TestImage_VerifyStructure() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	images, err := s.Client().Images().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(images) == 0 {
		s.T().Skip("No images available to verify structure")
		return
	}

	for _, img := range images {
		s.NotEmpty(img.Id, "Image should have an ID")
		s.NotEmpty(img.Name, "Image should have a name")
		s.Equal("image", string(img.Kind), "Kind should be 'image'")
		// Repository and Tag may or may not be present depending on how image is named
		s.T().Logf("Image: %s, Repository: %s, Tag: %s, Digest: %s",
			img.Name, img.Repository, img.Tag, img.Digest)
	}
}

// TestImage_Delete tests that delete operation works (on a non-existent image should error).
// Note: We don't want to actually delete images that might be in use,
// so we test the API call with a known non-existent image.
func (s *ImageTestSuite) TestImage_Delete_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Attempt to delete a non-existent image
	err := s.Client().Images().Delete(ctx, s.TestOrg(), "nonexistent-image-for-delete-test:v999")
	s.Require().Error(err, "Should error when deleting non-existent image")
}
