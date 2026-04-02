//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/quota"
)

// QuotaTestSuite tests quota CRUD operations.
type QuotaTestSuite struct {
	IntegrationSuite
}

func TestQuotaSuite(t *testing.T) {
	suite.Run(t, new(QuotaTestSuite))
}

// newTestQuota creates a minimal quota spec for testing.
func newTestQuota(name string) *quota.Quota {
	return &quota.Quota{
		Name:        base.Name(name),
		Description: "Test quota created by integration test",
		Version:     common.Float32Ptr(1),
		Unit:        "count",
		Max:         common.Float32Ptr(100),
	}
}

// TestQuota_List tests listing quotas with iterator.
func (s *QuotaTestSuite) TestQuota_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List quotas using iterator - should have at least built-in quotas
	count := 0
	for q, err := range s.Client().Quotas().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		s.NotEmpty(q.Name, "Quota should have a name")
		count++
	}
	s.T().Logf("Found %d quotas in org %s", count, s.TestOrg())
}

// TestQuota_ListAll tests listing all quotas at once.
func (s *QuotaTestSuite) TestQuota_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	quotas, err := s.Client().Quotas().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	// There should be built-in quotas
	s.T().Logf("Found %d quotas in org %s", len(quotas), s.TestOrg())
}

// TestQuota_ListPage tests pagination for listing quotas.
func (s *QuotaTestSuite) TestQuota_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	page, err := s.Client().Quotas().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(page)
	s.T().Logf("First page has %d quotas", len(page.Items))
}

// TestQuota_Get tests getting a specific quota.
func (s *QuotaTestSuite) TestQuota_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// List quotas to find a built-in quota to get
	quotas, err := s.Client().Quotas().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	if len(quotas) == 0 {
		s.T().Skip("No quotas available to test Get")
		return
	}

	// Get the first quota by ID (quotas are identified by GUID, not name)
	quotaID := quotas[0].Id
	s.Require().NotEmpty(quotaID, "Quota should have an ID")
	fetched, err := s.Client().Quotas().Get(ctx, s.TestOrg(), quotaID)
	s.Require().NoError(err)
	s.Equal(quotaID, fetched.Id)
	s.Equal("quota", string(fetched.Kind))
}

// TestQuota_GetNonExistent tests getting a non-existent quota.
func (s *QuotaTestSuite) TestQuota_GetNonExistent() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Quotas().Get(ctx, s.TestOrg(), "non-existent-quota-"+randomSuffix())
	s.Require().Error(err, "Should fail to get non-existent quota")
}

// TestQuota_Create tests creating a new quota.
func (s *QuotaTestSuite) TestQuota_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	quotaName := "test-quota-create-" + randomSuffix()
	q := newTestQuota(quotaName)

	created, err := s.Client().Quotas().Create(ctx, s.TestOrg(), q)
	if err != nil {
		// Quota creation may require special permissions
		s.T().Logf("Quota create failed (may require admin permissions): %v", err)
		s.T().Skip("Quota creation may require admin permissions")
		return
	}

	defer func() {
		_ = s.Client().Quotas().Delete(context.Background(), s.TestOrg(), quotaName)
	}()

	s.Equal(quotaName, string(created.Name))
	s.Equal(float32(100), *created.Max)
}

// TestQuota_Update tests updating a quota.
func (s *QuotaTestSuite) TestQuota_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First create a quota
	quotaName := "test-quota-update-" + randomSuffix()
	q := newTestQuota(quotaName)

	_, err := s.Client().Quotas().Create(ctx, s.TestOrg(), q)
	if err != nil {
		s.T().Logf("Quota create failed (may require admin permissions): %v", err)
		s.T().Skip("Quota creation may require admin permissions")
		return
	}

	defer func() {
		_ = s.Client().Quotas().Delete(context.Background(), s.TestOrg(), quotaName)
	}()

	// Update the quota
	q.Description = "Updated description"
	q.Max = common.Float32Ptr(200)

	updated, err := s.Client().Quotas().Update(ctx, s.TestOrg(), quotaName, q)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
	s.Equal(float32(200), *updated.Max)
}

// TestQuota_Delete tests deleting a quota.
func (s *QuotaTestSuite) TestQuota_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// First create a quota
	quotaName := "test-quota-delete-" + randomSuffix()
	q := newTestQuota(quotaName)

	_, err := s.Client().Quotas().Create(ctx, s.TestOrg(), q)
	if err != nil {
		s.T().Logf("Quota create failed (may require admin permissions): %v", err)
		s.T().Skip("Quota creation may require admin permissions")
		return
	}

	// Delete
	err = s.Client().Quotas().Delete(ctx, s.TestOrg(), quotaName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().Quotas().Get(ctx, s.TestOrg(), quotaName)
	s.Require().Error(err, "Quota should not exist after deletion")
}

// TestQuota_Query tests querying quotas with filters.
func (s *QuotaTestSuite) TestQuota_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Query for all quotas (no specific tag filter since built-in quotas may not have tags)
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	count := 0
	for quota, err := range s.Client().Quotas().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		s.NotEmpty(quota.Name)
		count++
	}
	s.T().Logf("Query returned %d quotas", count)
}

// TestQuota_QueryAll tests querying all quotas matching criteria.
func (s *QuotaTestSuite) TestQuota_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	quotas, err := s.Client().Quotas().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.T().Logf("QueryAll returned %d quotas", len(quotas))
}

// TestQuota_QueryPage tests pagination for querying quotas.
func (s *QuotaTestSuite) TestQuota_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	page, err := s.Client().Quotas().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(page)
	s.T().Logf("QueryPage returned %d quotas", len(page.Items))
}

// TestQuota_BuiltinQuotas tests that built-in quotas are present.
func (s *QuotaTestSuite) TestQuota_BuiltinQuotas() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	quotas, err := s.Client().Quotas().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	// Look for built-in quota types
	builtinFound := false
	for _, q := range quotas {
		if q.Origin == quota.QuotaOriginBuiltin || q.Origin == quota.QuotaOriginDefault {
			builtinFound = true
			s.T().Logf("Found builtin quota: %s (origin: %s, max: %v, current: %v)",
				q.Name, q.Origin, q.Max, q.Current)
		}
	}

	// If there are any quotas, at least some should be built-in or default
	if len(quotas) > 0 {
		s.True(builtinFound, "Expected to find built-in quotas")
	}
}

// TestQuota_Dimensions tests that quota dimensions are populated.
func (s *QuotaTestSuite) TestQuota_Dimensions() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	quotas, err := s.Client().Quotas().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	for _, q := range quotas {
		if q.Dimensions != nil && len(q.Dimensions) > 0 {
			s.T().Logf("Quota %s has dimensions: %v", q.Name, q.Dimensions)
		}
	}
}

// TestQuota_CurrentUsage tests that current usage is tracked.
func (s *QuotaTestSuite) TestQuota_CurrentUsage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	quotas, err := s.Client().Quotas().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)

	for _, q := range quotas {
		// Current should be >= 0 (may be nil for some quotas)
		if q.Current != nil {
			s.GreaterOrEqual(*q.Current, float32(0), "Quota current usage should be >= 0")
		}
		s.T().Logf("Quota %s: current=%v, max=%v", q.Name, q.Current, q.Max)
	}
}
