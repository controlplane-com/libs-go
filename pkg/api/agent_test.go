//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/controlplane-com/libs-go/pkg/common"
	"github.com/controlplane-com/libs-go/pkg/schema/agent"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// AgentTestSuite tests agent CRUD operations.
// Agents are used for Bring Your Own Kubernetes (BYOK) clusters.
type AgentTestSuite struct {
	IntegrationSuite
}

func TestAgentSuite(t *testing.T) {
	suite.Run(t, new(AgentTestSuite))
}

// newTestAgent creates a minimal agent for testing.
func newTestAgent(name string) *agent.Agent {
	return &agent.Agent{
		Name:        base.Name(name),
		Description: "Test agent created by integration test",
		Version:     common.Float32Ptr(1),
		Tags: agent.AgentTags{
			"test":        "true",
			"integration": "libs-go",
		},
	}
}

// TestAgent_Create tests creating a new agent.
func (s *AgentTestSuite) TestAgent_Create() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-create-" + randomSuffix()
	a := newTestAgent(agentName)

	created, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err, "Failed to create agent")

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	s.Equal(agentName, string(created.Name))
	s.Equal("agent", string(created.Kind))
	s.NotEmpty(created.Status.BootstrapConfig.RegistrationToken, "Should have a registration token")
}

// TestAgent_Get tests getting a specific agent.
func (s *AgentTestSuite) TestAgent_Get() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-get-" + randomSuffix()
	a := newTestAgent(agentName)

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// Get the agent
	fetched, err := s.Client().Agents().Get(ctx, s.TestOrg(), agentName)
	s.Require().NoError(err)
	s.Equal(agentName, string(fetched.Name))
	s.Equal("agent", string(fetched.Kind))
	s.NotEmpty(fetched.Status.BootstrapConfig.AgentId, "Should have an agent ID")
}

// TestAgent_Get_NotFound tests getting a non-existent agent.
func (s *AgentTestSuite) TestAgent_Get_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	_, err := s.Client().Agents().Get(ctx, s.TestOrg(), "nonexistent-agent-xyz")
	s.Require().Error(err, "Should error for non-existent agent")
}

// TestAgent_Update tests updating an agent.
func (s *AgentTestSuite) TestAgent_Update() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-update-" + randomSuffix()
	a := newTestAgent(agentName)

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// Update the agent
	a.Description = "Updated description"
	a.Tags["updated"] = "true"

	updated, err := s.Client().Agents().Update(ctx, s.TestOrg(), agentName, a)
	s.Require().NoError(err)
	s.Equal("Updated description", updated.Description)
}

// TestAgent_Delete tests deleting an agent.
func (s *AgentTestSuite) TestAgent_Delete() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-delete-" + randomSuffix()
	a := newTestAgent(agentName)

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	// Delete
	err = s.Client().Agents().Delete(ctx, s.TestOrg(), agentName)
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.Client().Agents().Get(ctx, s.TestOrg(), agentName)
	s.Require().Error(err, "Agent should not exist after deletion")
}

// TestAgent_Delete_NotFound tests deleting a non-existent agent.
func (s *AgentTestSuite) TestAgent_Delete_NotFound() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	err := s.Client().Agents().Delete(ctx, s.TestOrg(), "nonexistent-agent-for-delete")
	s.Require().Error(err, "Should error when deleting non-existent agent")
}

// TestAgent_List tests listing agents with iterator.
func (s *AgentTestSuite) TestAgent_List() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-list-" + randomSuffix()
	a := newTestAgent(agentName)

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// List using iterator
	found := false
	for ag, err := range s.Client().Agents().List(ctx, s.TestOrg()) {
		s.Require().NoError(err)
		if string(ag.Name) == agentName {
			found = true
			break
		}
	}
	s.True(found, "Agent should appear in list")
}

// TestAgent_ListAll tests listing all agents at once.
func (s *AgentTestSuite) TestAgent_ListAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-listall-" + randomSuffix()
	a := newTestAgent(agentName)

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	agents, err := s.Client().Agents().ListAll(ctx, s.TestOrg())
	s.Require().NoError(err)
	s.NotEmpty(agents, "Should have at least one agent")

	found := false
	for _, ag := range agents {
		if string(ag.Name) == agentName {
			found = true
			break
		}
	}
	s.True(found, "Agent should be in list")
}

// TestAgent_ListPage tests paginated listing.
func (s *AgentTestSuite) TestAgent_ListPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-listpage-" + randomSuffix()
	a := newTestAgent(agentName)

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// Get first page
	resp, err := s.Client().Agents().ListPage(ctx, s.TestOrg(), "")
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotEmpty(resp.Items, "Should have at least one agent")
}

// TestAgent_Query tests querying agents with filters.
func (s *AgentTestSuite) TestAgent_Query() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-query-" + randomSuffix()
	a := newTestAgent(agentName)
	a.Tags["querytest"] = "yes"

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// Query for agents with the tag
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
	for ag, err := range s.Client().Agents().Query(ctx, s.TestOrg(), q) {
		s.Require().NoError(err)
		if string(ag.Name) == agentName {
			found = true
			break
		}
	}
	s.True(found, "Should find agent with query")
}

// TestAgent_QueryAll tests querying all agents matching criteria.
func (s *AgentTestSuite) TestAgent_QueryAll() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-queryall-" + randomSuffix()
	a := newTestAgent(agentName)
	a.Tags["queryalltest"] = "yes"

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// Query for agents with the tag
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

	agents, err := s.Client().Agents().QueryAll(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotEmpty(agents, "Should have at least one agent matching query")

	found := false
	for _, ag := range agents {
		if string(ag.Name) == agentName {
			found = true
			break
		}
	}
	s.True(found, "Agent should be in query results")
}

// TestAgent_QueryPage tests paginated querying.
func (s *AgentTestSuite) TestAgent_QueryPage() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-querypage-" + randomSuffix()
	a := newTestAgent(agentName)

	_, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// Query for all agents
	q := &query.Query{
		Spec: &query.Spec{
			Match: query.SpecMatchAll,
			Terms: []query.Term{},
		},
	}

	resp, err := s.Client().Agents().QueryPage(ctx, s.TestOrg(), q)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotEmpty(resp.Items, "Should have at least one agent")
}

// TestAgent_VerifyBootstrapConfig tests that created agents have valid bootstrap config.
func (s *AgentTestSuite) TestAgent_VerifyBootstrapConfig() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-bootstrap-" + randomSuffix()
	a := newTestAgent(agentName)

	created, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// Verify bootstrap config fields
	bc := created.Status.BootstrapConfig
	s.NotEmpty(bc.RegistrationToken, "Should have a registration token")
	s.NotEmpty(bc.AgentId, "Should have an agent ID")
	s.NotEmpty(bc.AgentLink, "Should have an agent link")
	s.NotEmpty(bc.HubEndpoint, "Should have a hub endpoint")

	s.T().Logf("Agent bootstrap config: agentId=%s, hubEndpoint=%s, protocolVersion=%s",
		bc.AgentId, bc.HubEndpoint, bc.ProtocolVersion)
}

// TestAgent_MultipleTags tests creating an agent with multiple tags.
func (s *AgentTestSuite) TestAgent_MultipleTags() {
	ctx, cancel := context.WithTimeout(s.Ctx(), 30*time.Second)
	defer cancel()

	agentName := "test-agent-tags-" + randomSuffix()
	a := &agent.Agent{
		Name:        base.Name(agentName),
		Description: "Test agent with multiple tags",
		Version:     common.Float32Ptr(1),
		Tags: agent.AgentTags{
			"env":         "test",
			"team":        "platform",
			"region":      "us-west",
			"integration": "libs-go",
		},
	}

	created, err := s.Client().Agents().Create(ctx, s.TestOrg(), a)
	s.Require().NoError(err)

	defer func() {
		_ = s.Client().Agents().Delete(context.Background(), s.TestOrg(), agentName)
	}()

	// Verify tags are preserved
	s.Equal("test", created.Tags["env"])
	s.Equal("platform", created.Tags["team"])
	s.Equal("us-west", created.Tags["region"])
}
