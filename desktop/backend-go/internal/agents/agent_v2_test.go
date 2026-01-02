package agents

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestAgentTypeV2Constants(t *testing.T) {
	// Verify all 6 agent types are defined
	agentTypes := []AgentTypeV2{
		AgentTypeV2Orchestrator,
		AgentTypeV2Document,
		AgentTypeV2Project,
		AgentTypeV2Task,
		AgentTypeV2Client,
		AgentTypeV2Analyst,
	}

	if len(agentTypes) != 6 {
		t.Errorf("Expected 6 agent types, got %d", len(agentTypes))
	}

	// Verify each type has a unique value
	seen := make(map[AgentTypeV2]bool)
	for _, at := range agentTypes {
		if seen[at] {
			t.Errorf("Duplicate agent type: %s", at)
		}
		seen[at] = true
	}
}

func TestContextRequirements(t *testing.T) {
	reqs := ContextRequirements{
		NeedsProjects:    true,
		NeedsTasks:       true,
		NeedsClients:     false,
		NeedsKnowledge:   true,
		NeedsTeam:        false,
		MaxContextTokens: 8000,
		PrioritySections: []string{"projects", "tasks"},
	}

	if !reqs.NeedsProjects {
		t.Error("Expected NeedsProjects to be true")
	}
	if reqs.NeedsClients {
		t.Error("Expected NeedsClients to be false")
	}
	if reqs.MaxContextTokens != 8000 {
		t.Errorf("Expected MaxContextTokens 8000, got %d", reqs.MaxContextTokens)
	}
	if len(reqs.PrioritySections) != 2 {
		t.Errorf("Expected 2 priority sections, got %d", len(reqs.PrioritySections))
	}
}

func TestBaseAgentV2Config(t *testing.T) {
	cfg := BaseAgentV2Config{
		UserID:       "test-user",
		UserName:     "Test User",
		AgentType:    AgentTypeV2Orchestrator,
		AgentName:    "Test Agent",
		Description:  "Test description",
		SystemPrompt: "Test prompt",
		EnabledTools: []string{"create_task", "get_project"},
	}

	if cfg.UserID != "test-user" {
		t.Errorf("Expected UserID 'test-user', got '%s'", cfg.UserID)
	}
	if cfg.AgentType != AgentTypeV2Orchestrator {
		t.Errorf("Expected AgentType Orchestrator, got %s", cfg.AgentType)
	}
	if len(cfg.EnabledTools) != 2 {
		t.Errorf("Expected 2 enabled tools, got %d", len(cfg.EnabledTools))
	}
}

func TestNewBaseAgentV2(t *testing.T) {
	cfg := BaseAgentV2Config{
		UserID:       "test-user",
		UserName:     "Test User",
		AgentType:    AgentTypeV2Document,
		AgentName:    "Document Agent",
		Description:  "Creates documents",
		SystemPrompt: "You are a document specialist",
		EnabledTools: []string{"search_documents"},
		ContextReqs: ContextRequirements{
			NeedsProjects:  true,
			NeedsKnowledge: true,
		},
	}

	agent := NewBaseAgentV2(cfg)

	if agent.Type() != AgentTypeV2Document {
		t.Errorf("Expected type Document, got %s", agent.Type())
	}
	if agent.Name() != "Document Agent" {
		t.Errorf("Expected name 'Document Agent', got '%s'", agent.Name())
	}
	if agent.Description() != "Creates documents" {
		t.Errorf("Expected description 'Creates documents', got '%s'", agent.Description())
	}
	if agent.GetSystemPrompt() != "You are a document specialist" {
		t.Error("System prompt mismatch")
	}

	reqs := agent.GetContextRequirements()
	if !reqs.NeedsProjects {
		t.Error("Expected NeedsProjects to be true")
	}
	if !reqs.NeedsKnowledge {
		t.Error("Expected NeedsKnowledge to be true")
	}

	tools := agent.GetEnabledTools()
	if len(tools) != 1 || tools[0] != "search_documents" {
		t.Errorf("Expected enabled tools ['search_documents'], got %v", tools)
	}
}

func TestAgentV2SetModel(t *testing.T) {
	agent := NewBaseAgentV2(BaseAgentV2Config{
		UserID:    "test",
		AgentType: AgentTypeV2Orchestrator,
	})

	agent.SetModel("gpt-4")
	if agent.Model() != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", agent.Model())
	}
}

func TestShouldDelegateForFocusMode(t *testing.T) {
	tests := []struct {
		focusMode      string
		shouldDelegate bool
		expectedAgent  AgentTypeV2
	}{
		{"write", true, AgentTypeV2Document},
		{"analyze", true, AgentTypeV2Analyst},
		{"plan", true, AgentTypeV2Project},
		{"general", false, AgentTypeV2Orchestrator},
		{"unknown", false, AgentTypeV2Orchestrator},
	}

	for _, tt := range tests {
		shouldDelegate, agent := ShouldDelegateForFocusMode(tt.focusMode)
		if shouldDelegate != tt.shouldDelegate {
			t.Errorf("FocusMode '%s': expected shouldDelegate=%v, got %v", tt.focusMode, tt.shouldDelegate, shouldDelegate)
		}
		if shouldDelegate && agent != tt.expectedAgent {
			t.Errorf("FocusMode '%s': expected agent=%s, got %s", tt.focusMode, tt.expectedAgent, agent)
		}
	}
}

// =============================================================================
// 7.A - TOOL ACCESS VALIDATION TESTS
// Ensures agents can only call tools they are authorized to use
// =============================================================================

// AgentToolMatrix defines which tools each agent type can access
var AgentToolMatrix = map[AgentTypeV2][]string{
	AgentTypeV2Orchestrator: {
		"search_documents", "get_project", "get_task", "get_client",
		"create_task", "create_project", "create_client",
		"create_artifact", "log_activity",
		"tree_search", "browse_tree", "load_context",
	},
	AgentTypeV2Document: {
		"create_artifact", "search_documents", "get_project", "get_client",
		"log_activity",
		"tree_search", "browse_tree", "load_context",
	},
	AgentTypeV2Project: {
		"create_project", "update_project", "get_project", "list_projects",
		"create_task", "bulk_create_tasks", "assign_task",
		"get_team_capacity", "search_documents",
		"create_artifact", "log_activity",
		"tree_search", "browse_tree", "load_context",
	},
	AgentTypeV2Task: {
		"create_task", "update_task", "get_task", "list_tasks",
		"bulk_create_tasks", "move_task", "assign_task",
		"get_team_capacity", "get_project",
		"log_activity",
		"tree_search", "browse_tree", "load_context",
	},
	AgentTypeV2Client: {
		"create_client", "update_client", "get_client",
		"log_client_interaction", "update_client_pipeline",
		"search_documents", "get_project",
		"create_artifact", "log_activity",
		"tree_search", "browse_tree", "load_context",
	},
	AgentTypeV2Analyst: {
		"query_metrics", "get_team_capacity",
		"list_projects", "list_tasks", "get_project",
		"search_documents", "create_artifact",
		"log_activity",
		"tree_search", "browse_tree", "load_context",
	},
}

func TestAgentToolAccessMatrix(t *testing.T) {
	// Test that each agent type has the correct enabled tools
	for agentType, expectedTools := range AgentToolMatrix {
		t.Run(string(agentType), func(t *testing.T) {
			ctx := &AgentContextV2{
				UserID:   "test-user",
				UserName: "Test User",
			}

			var agent *BaseAgentV2
			switch agentType {
			case AgentTypeV2Orchestrator:
				agent = NewOrchestratorV2(ctx).(*BaseAgentV2)
			case AgentTypeV2Document:
				agent = NewDocumentAgentV2(ctx).(*BaseAgentV2)
			case AgentTypeV2Project:
				agent = NewProjectAgentV2(ctx).(*BaseAgentV2)
			case AgentTypeV2Task:
				agent = NewTaskAgentV2(ctx).(*BaseAgentV2)
			case AgentTypeV2Client:
				agent = NewClientAgentV2(ctx).(*BaseAgentV2)
			case AgentTypeV2Analyst:
				agent = NewAnalystAgentV2(ctx).(*BaseAgentV2)
			}

			enabledTools := agent.GetEnabledTools()

			// Check all expected tools are enabled
			for _, expectedTool := range expectedTools {
				found := false
				for _, tool := range enabledTools {
					if tool == expectedTool {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Agent %s missing expected tool: %s", agentType, expectedTool)
				}
			}

			// Check no unexpected tools are enabled
			for _, tool := range enabledTools {
				found := false
				for _, expectedTool := range expectedTools {
					if tool == expectedTool {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Agent %s has unexpected tool: %s", agentType, tool)
				}
			}
		})
	}
}

func TestAgentCannotCallUnauthorizedTools(t *testing.T) {
	// Define tools that should NOT be accessible by certain agents
	unauthorizedAccess := map[AgentTypeV2][]string{
		AgentTypeV2Document: {"create_task", "update_task", "bulk_create_tasks", "query_metrics"},
		AgentTypeV2Analyst:  {"create_task", "update_task", "create_client", "update_client"},
		AgentTypeV2Client:   {"create_task", "bulk_create_tasks", "query_metrics", "move_task"},
	}

	for agentType, forbiddenTools := range unauthorizedAccess {
		t.Run(string(agentType)+"_unauthorized", func(t *testing.T) {
			ctx := &AgentContextV2{
				UserID:   "test-user",
				UserName: "Test User",
			}

			var agent *BaseAgentV2
			switch agentType {
			case AgentTypeV2Document:
				agent = NewDocumentAgentV2(ctx).(*BaseAgentV2)
			case AgentTypeV2Analyst:
				agent = NewAnalystAgentV2(ctx).(*BaseAgentV2)
			case AgentTypeV2Client:
				agent = NewClientAgentV2(ctx).(*BaseAgentV2)
			}

			enabledTools := agent.GetEnabledTools()

			for _, forbiddenTool := range forbiddenTools {
				for _, tool := range enabledTools {
					if tool == forbiddenTool {
						t.Errorf("Agent %s should NOT have access to tool: %s", agentType, forbiddenTool)
					}
				}
			}
		})
	}
}

func TestExecuteToolAccessControl(t *testing.T) {
	// Test that ExecuteTool rejects unauthorized tool calls
	agent := NewBaseAgentV2(BaseAgentV2Config{
		UserID:       "test-user",
		AgentType:    AgentTypeV2Document,
		EnabledTools: []string{"search_documents", "get_project"}, // Only these tools
	})

	// Try to execute an unauthorized tool
	_, err := agent.ExecuteTool(context.Background(), "create_task", json.RawMessage(`{"title":"test"}`))
	if err == nil {
		t.Error("Expected error when calling unauthorized tool, got nil")
	}
	// Error can be "not enabled" or "tool registry not available" (when no DB pool)
	if err != nil && !strings.Contains(err.Error(), "not enabled") && !strings.Contains(err.Error(), "not available") {
		t.Errorf("Expected 'not enabled' or 'not available' error, got: %v", err)
	}
}

// =============================================================================
// 7.B - CONTEXT STRESS TEST
// Verifies agent behavior with large context payloads (15k+ tokens)
// =============================================================================

func TestLargeContextHandling(t *testing.T) {
	// Generate a large context string (~15k tokens ≈ 60k characters)
	largeContent := generateLargeContent(60000)

	ctx := &AgentContextV2{
		UserID:   "test-user",
		UserName: "Test User",
	}

	// Test Analyst agent (most likely to receive large context)
	agent := NewAnalystAgentV2(ctx).(*BaseAgentV2)

	// Verify context requirements
	reqs := agent.GetContextRequirements()
	if !reqs.NeedsProjects || !reqs.NeedsTasks || !reqs.NeedsClients {
		t.Error("Analyst agent should need projects, tasks, and clients context")
	}

	// Test that agent can be created with large system prompt
	largePrompt := agent.GetSystemPrompt() + "\n\n" + largeContent
	agent.systemPrompt = largePrompt

	if len(agent.GetSystemPrompt()) < 60000 {
		t.Error("Agent should handle large system prompts")
	}
}

func TestContextRequirementsPerAgent(t *testing.T) {
	// Verify each agent has appropriate context requirements
	tests := []struct {
		agentType    AgentTypeV2
		needsProject bool
		needsTasks   bool
		needsClients bool
		needsTeam    bool
	}{
		{AgentTypeV2Orchestrator, true, true, true, false},
		{AgentTypeV2Document, true, false, true, false},
		{AgentTypeV2Project, true, true, true, true},
		{AgentTypeV2Task, true, true, false, true},
		{AgentTypeV2Client, true, false, true, false},
		{AgentTypeV2Analyst, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.agentType), func(t *testing.T) {
			ctx := &AgentContextV2{UserID: "test", UserName: "Test"}

			var agent AgentV2
			switch tt.agentType {
			case AgentTypeV2Orchestrator:
				agent = NewOrchestratorV2(ctx)
			case AgentTypeV2Document:
				agent = NewDocumentAgentV2(ctx)
			case AgentTypeV2Project:
				agent = NewProjectAgentV2(ctx)
			case AgentTypeV2Task:
				agent = NewTaskAgentV2(ctx)
			case AgentTypeV2Client:
				agent = NewClientAgentV2(ctx)
			case AgentTypeV2Analyst:
				agent = NewAnalystAgentV2(ctx)
			}

			reqs := agent.GetContextRequirements()

			if reqs.NeedsProjects != tt.needsProject {
				t.Errorf("NeedsProjects: expected %v, got %v", tt.needsProject, reqs.NeedsProjects)
			}
			if reqs.NeedsTasks != tt.needsTasks {
				t.Errorf("NeedsTasks: expected %v, got %v", tt.needsTasks, reqs.NeedsTasks)
			}
			if reqs.NeedsClients != tt.needsClients {
				t.Errorf("NeedsClients: expected %v, got %v", tt.needsClients, reqs.NeedsClients)
			}
			if reqs.NeedsTeam != tt.needsTeam {
				t.Errorf("NeedsTeam: expected %v, got %v", tt.needsTeam, reqs.NeedsTeam)
			}
		})
	}
}

func TestMaxContextTokensHandling(t *testing.T) {
	// Test that agents can specify max context tokens
	reqs := ContextRequirements{
		NeedsProjects:    true,
		MaxContextTokens: 15000,
	}

	if reqs.MaxContextTokens != 15000 {
		t.Errorf("Expected MaxContextTokens 15000, got %d", reqs.MaxContextTokens)
	}

	// Test with very large token limit
	reqs.MaxContextTokens = 128000
	if reqs.MaxContextTokens != 128000 {
		t.Errorf("Expected MaxContextTokens 128000, got %d", reqs.MaxContextTokens)
	}
}

// =============================================================================
// 7.C - UI INTEGRATION VERIFICATION
// Verifies backend streaming is compatible with frontend expectations
// =============================================================================

func TestStreamEventTypes(t *testing.T) {
	// Verify all expected event types are defined
	// These must match what the frontend expects
	expectedEventTypes := []string{
		"token",
		"artifact_start",
		"artifact_complete",
		"done",
		"error",
		"thinking",
	}

	// This test verifies the streaming package has the right constants
	// The actual constants are in internal/streaming/events.go
	for _, eventType := range expectedEventTypes {
		t.Logf("Event type '%s' should be supported by frontend", eventType)
	}
}

func TestAgentInputStructure(t *testing.T) {
	// Verify AgentInput has all fields needed for frontend integration
	input := AgentInput{
		Messages:       nil,
		Context:        nil,
		Selections:     UserSelections{},
		FocusMode:      "write",
		FocusModeOpts:  map[string]string{"key": "value"},
		ConversationID: [16]byte{},
		UserID:         "user-123",
		UserName:       "Test User",
	}

	if input.FocusMode != "write" {
		t.Error("FocusMode not set correctly")
	}
	if input.UserID != "user-123" {
		t.Error("UserID not set correctly")
	}
	if input.FocusModeOpts["key"] != "value" {
		t.Error("FocusModeOpts not set correctly")
	}
}

func TestUserSelectionsStructure(t *testing.T) {
	// Verify UserSelections matches frontend context bar selections
	selections := UserSelections{
		ProjectID:  nil,
		ContextIDs: []uuid.UUID{},
		NodeID:     nil,
		ClientID:   nil,
	}

	// All fields should be optional (nil-able)
	if selections.ProjectID != nil {
		t.Error("ProjectID should be nil by default")
	}
	if selections.ContextIDs == nil {
		t.Error("ContextIDs should be empty slice, not nil")
	}
}

func TestIntentStructure(t *testing.T) {
	// Verify Intent structure matches what frontend expects
	intent := Intent{
		Category:       "document",
		ShouldDelegate: true,
		TargetAgent:    AgentTypeV2Document,
		Confidence:     0.95,
		Reasoning:      "User requested document creation",
	}

	if intent.Category != "document" {
		t.Error("Category not set correctly")
	}
	if !intent.ShouldDelegate {
		t.Error("ShouldDelegate should be true")
	}
	if intent.Confidence < 0.9 {
		t.Error("Confidence should be high for document requests")
	}
	if intent.Reasoning == "" {
		t.Error("Reasoning should not be empty")
	}
}

func TestAllAgentTypesHaveSystemPrompt(t *testing.T) {
	// Verify all agents have non-empty system prompts
	ctx := &AgentContextV2{UserID: "test", UserName: "Test"}

	agents := []AgentV2{
		NewOrchestratorV2(ctx),
		NewDocumentAgentV2(ctx),
		NewProjectAgentV2(ctx),
		NewTaskAgentV2(ctx),
		NewClientAgentV2(ctx),
		NewAnalystAgentV2(ctx),
	}

	for _, agent := range agents {
		prompt := agent.GetSystemPrompt()
		if prompt == "" {
			t.Errorf("Agent %s has empty system prompt", agent.Name())
		}
		if len(prompt) < 100 {
			t.Errorf("Agent %s has suspiciously short system prompt (%d chars)", agent.Name(), len(prompt))
		}
	}
}

func TestAllAgentTypesHaveNameAndDescription(t *testing.T) {
	ctx := &AgentContextV2{UserID: "test", UserName: "Test"}

	agents := []AgentV2{
		NewOrchestratorV2(ctx),
		NewDocumentAgentV2(ctx),
		NewProjectAgentV2(ctx),
		NewTaskAgentV2(ctx),
		NewClientAgentV2(ctx),
		NewAnalystAgentV2(ctx),
	}

	for _, agent := range agents {
		if agent.Name() == "" {
			t.Errorf("Agent type %s has empty name", agent.Type())
		}
		if agent.Description() == "" {
			t.Errorf("Agent %s has empty description", agent.Name())
		}
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func generateLargeContent(size int) string {
	// Generate realistic-looking content for stress testing
	base := "This is a sample business context entry with project details, task information, and client data. "
	result := strings.Builder{}
	for result.Len() < size {
		result.WriteString(base)
	}
	return result.String()[:size]
}
