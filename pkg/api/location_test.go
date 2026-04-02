//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// LocationTestSuite tests location read operations.
// Locations are read-only resources managed by the platform.
type LocationTestSuite struct {
	IntegrationSuite
}

func TestLocationSuite(t *testing.T) {
	suite.Run(t, new(LocationTestSuite))
}

// TestLocation_List tests listing locations with iterator.
func (s *LocationTestSuite) TestLocation_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List using iterator
	count := 0
	for loc, err := range s.Client().Locations().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		s.NotEmpty(loc.Name, "Location should have a name")
		count++
		if count >= 3 {
			break // Just verify we can iterate
		}
	}
	s.Greater(count, 0, "Should have at least one location")
}

// TestLocation_ListAll tests listing all locations at once.
func (s *LocationTestSuite) TestLocation_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	locations, err := s.Client().Locations().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(locations, "Should have at least one location")

	// Verify location structure
	for _, loc := range locations {
		s.NotEmpty(loc.Name, "Location should have a name")
		s.Equal("location", string(loc.Kind), "Kind should be 'location'")
	}
}

// TestLocation_ListPage tests paginated listing.
func (s *LocationTestSuite) TestLocation_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	resp, err := s.Client().Locations().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotEmpty(resp.Items, "Should have at least one location")
}

// TestLocation_Get tests getting a specific location.
func (s *LocationTestSuite) TestLocation_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First, list locations to get a valid name
	locations, err := s.Client().Locations().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.Require().NotEmpty(locations, "Need at least one location to test Get")

	locationName := string(locations[0].Name)

	// Get the specific location
	loc, err := s.Client().Locations().Get(ctx, s.TestOrg(), locationName)
	s.Require().NoError(err)
	s.Equal(locationName, string(loc.Name))
	s.Equal("location", string(loc.Kind))
	s.NotEmpty(loc.Provider, "Location should have a provider")
}

// TestLocation_Get_NotFound tests getting a non-existent location.
func (s *LocationTestSuite) TestLocation_Get_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Locations().Get(ctx, s.TestOrg(), "nonexistent-location-xyz")
	s.Require().Error(err, "Should error for non-existent location")
}

// TestLocation_Query tests querying locations with filters.
func (s *LocationTestSuite) TestLocation_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all locations (empty query matches all)
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	count := 0
	for loc, err := range s.Client().Locations().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		s.NotEmpty(loc.Name, "Location should have a name")
		count++
		if count >= 3 {
			break
		}
	}
	s.Greater(count, 0, "Should find at least one location")
}

// TestLocation_QueryAll tests querying all locations matching criteria.
func (s *LocationTestSuite) TestLocation_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all locations
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	locations, err := s.Client().Locations().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotEmpty(locations, "Should have at least one location")
}

// TestLocation_QueryPage tests paginated querying.
func (s *LocationTestSuite) TestLocation_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	resp, err := s.Client().Locations().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotEmpty(resp.Items, "Should have at least one location")
}

// TestLocation_VerifyProviders tests that locations have valid providers.
func (s *LocationTestSuite) TestLocation_VerifyProviders() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	locations, err := s.Client().Locations().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Valid providers based on schema
	validProviders := map[string]bool{
		"aws":     true,
		"gcp":     true,
		"azure":   true,
		"byok":    true,
		"linode":  true,
		"vultr":   true,
		"equinix": true,
		"oci":     true,
	}

	for _, loc := range locations {
		s.True(validProviders[string(loc.Provider)], "Location %s should have a valid provider, got: %s", loc.Name, loc.Provider)
	}
}

// TestLocation_VerifyOrigins tests that locations have valid origins.
func (s *LocationTestSuite) TestLocation_VerifyOrigins() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	locations, err := s.Client().Locations().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Valid origins based on schema
	validOrigins := map[string]bool{
		"builtin": true,
		"default": true,
		"custom":  true,
	}

	for _, loc := range locations {
		if loc.Origin != "" {
			s.True(validOrigins[string(loc.Origin)], "Location %s should have a valid origin, got: %s", loc.Name, loc.Origin)
		}
	}
}
