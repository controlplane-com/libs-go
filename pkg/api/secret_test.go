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
	"github.com/controlplane-com/libs-go/pkg/schema/secret"
)

// SecretTestSuite tests secret CRUD and reveal operations.
type SecretTestSuite struct {
	IntegrationSuite
}

func TestSecretSuite(t *testing.T) {
	suite.Run(t, new(SecretTestSuite))
}

// newTestSecret creates a minimal opaque secret for testing.
func newTestSecret(name string) *secret.Secret {
	return &secret.Secret{
		Name:        base.Name(name),
		Description: "Test secret created by integration test",
		Version:     common.Float32Ptr(1),
		Type:        secret.SecretTypeOpaque,
		Tags: secret.SecretTags{
			"test":        "true",
			"integration": "libs-go",
		},
		Data: map[string]interface{}{
			"payload":  "test-secret-value",
			"encoding": "plain",
		},
	}
}

// TestSecret_Create tests creating a new secret.
func (s *SecretTestSuite) TestSecret_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-create-" + randomSuffix()
	sec := newTestSecret(secretName)

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err, "Failed to create secret")

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	// Verify it exists by fetching it (data should be hidden)
	fetched, err := s.Client().Secrets().Get(ctx, s.TestOrg(), secretName)
	s.Require().NoError(err)
	s.Equal(secretName, string(fetched.Name))
	s.Equal(secret.SecretTypeOpaque, fetched.Type)
}

// TestSecret_Get tests getting a specific secret.
func (s *SecretTestSuite) TestSecret_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-get-" + randomSuffix()
	sec := newTestSecret(secretName)

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	// Get the secret
	fetched, err := s.Client().Secrets().Get(ctx, s.TestOrg(), secretName)
	s.Require().NoError(err)
	s.Equal(secretName, string(fetched.Name))
	s.Equal("secret", string(fetched.Kind))
	s.Equal(secret.SecretTypeOpaque, fetched.Type)
}

// TestSecret_Update tests updating a secret.
func (s *SecretTestSuite) TestSecret_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-update-" + randomSuffix()
	sec := newTestSecret(secretName)

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	// Update the secret
	sec.Description = "Updated description"
	sec.Tags["updated"] = "true"
	sec.Data = map[string]interface{}{
		"payload":  "updated-secret-value",
		"encoding": "plain",
	}

	updated, err := s.Client().Secrets().Update(ctx, s.TestOrg(), secretName, sec)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestSecret_Delete tests deleting a secret.
func (s *SecretTestSuite) TestSecret_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-delete-" + randomSuffix()
	sec := newTestSecret(secretName)

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	// Delete
	err = s.Client().Secrets().Delete(ctx, s.TestOrg(), secretName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().Secrets().Get(ctx, s.TestOrg(), secretName)
	s.Require().Error(err, "Secret should not exist after deletion")
}

// TestSecret_List tests listing secrets with iterator.
func (s *SecretTestSuite) TestSecret_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-list-" + randomSuffix()
	sec := newTestSecret(secretName)

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	// List using iterator
	found := false
	for sec, err := range s.Client().Secrets().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		if string(sec.Name) == secretName {
			found = true
			break
		}
	}
	s.True(found, "Secret should appear in list")
}

// TestSecret_ListAll tests listing all secrets at once.
func (s *SecretTestSuite) TestSecret_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-listall-" + randomSuffix()
	sec := newTestSecret(secretName)

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	secrets, err := s.Client().Secrets().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(secrets, "Should have at least one secret")

	found := false
	for _, sec := range secrets {
		if string(sec.Name) == secretName {
			found = true
			break
		}
	}
	s.True(found, "Secret should be in list")
}

// TestSecret_Query tests querying secrets with filters.
func (s *SecretTestSuite) TestSecret_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-query-" + randomSuffix()
	sec := newTestSecret(secretName)
	sec.Tags["querytest"] = "yes"

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	// Query for secrets with the tag
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
	for sec, err := range s.Client().Secrets().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(sec.Name) == secretName {
			found = true
			break
		}
	}
	s.True(found, "Should find secret with query")
}

// TestSecret_Reveal tests revealing a secret's data.
func (s *SecretTestSuite) TestSecret_Reveal() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-reveal-" + randomSuffix()
	sec := newTestSecret(secretName)

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	// Reveal the secret
	revealed, err := s.Client().Secrets().Reveal(ctx, s.TestOrg(), secretName)
	s.Require().NoError(err)
	s.NotNil(revealed.Data, "Revealed secret should have data")

	// Verify the data contains the expected value
	if data, ok := revealed.Data.(map[string]interface{}); ok {
		s.Equal("test-secret-value", data["payload"], "Revealed payload should match")
	}
}

// TestSecret_Dictionary tests creating a dictionary secret.
func (s *SecretTestSuite) TestSecret_Dictionary() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	secretName := "test-secret-dict-" + randomSuffix()
	sec := &secret.Secret{
		Name:        base.Name(secretName),
		Description: "Test dictionary secret",
		Version:     common.Float32Ptr(1),
		Type:        secret.SecretTypeDictionary,
		Tags: secret.SecretTags{
			"test": "true",
		},
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}

	_, err := s.Client().Secrets().Create(ctx, s.TestOrg(), sec)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Secrets().Delete(context.Background(), s.TestOrg(), secretName)
	}()

	// Verify by fetching it
	fetched, err := s.Client().Secrets().Get(ctx, s.TestOrg(), secretName)
	s.Require().NoError(err)
	s.Equal(secret.SecretTypeDictionary, fetched.Type)

	// Reveal and verify data
	revealed, err := s.Client().Secrets().Reveal(ctx, s.TestOrg(), secretName)
	s.Require().NoError(err)

	if data, ok := revealed.Data.(map[string]interface{}); ok {
		s.Equal("value1", data["key1"])
		s.Equal("value2", data["key2"])
		s.Equal("value3", data["key3"])
	}
}

// TestSecret_ListPage_NextPage tests pagination using the NextPage method.
func (s *SecretTestSuite) TestSecret_ListPage_NextPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	// Get first page
	page, err := s.Client().Secrets().ListPage(ctx, s.TestOrg(), "")
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

// TestSecret_QueryPage_NextPage tests query pagination using the NextPage method.
func (s *SecretTestSuite) TestSecret_QueryPage_NextPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	page, err := s.Client().Secrets().QueryPage(ctx, s.TestOrg(), q)
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
