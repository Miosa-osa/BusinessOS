package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/prompts"
	prompts_agents "github.com/rhl/businessos-backend/internal/prompts/agents"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
	"github.com/rhl/businessos-backend/internal/tools"
)

// BaseAgentV2 provides common functionality for all V2 agents
type BaseAgentV2 struct {
	pool           *pgxpool.Pool
	cfg            *config.Config
	userID         string
	userName       string
	conversationID *uuid.UUID
	model          string
	agentType      AgentTypeV2
	agentName      string
	description    string
	systemPrompt   string
	contextReqs    ContextRequirements
	llmOptions     services.LLMOptions
	toolRegistry   *tools.AgentToolRegistry
	enabledTools   []string // Tool names this agent can use
}

// BaseAgentV2Config holds configuration for creating a BaseAgentV2
type BaseAgentV2Config struct {
	Pool           *pgxpool.Pool
	Config         *config.Config
	UserID         string
	UserName       string
	ConversationID *uuid.UUID
	Model          string
	AgentType      AgentTypeV2
	AgentName      string
	Description    string
	SystemPrompt   string
	ContextReqs    ContextRequirements
	EnabledTools   []string // Tool names this agent can use
}

// NewBaseAgentV2 creates a new base agent with the given configuration
func NewBaseAgentV2(cfg BaseAgentV2Config) *BaseAgentV2 {
	model := cfg.Model
	if model == "" && cfg.Config != nil {
		model = cfg.Config.DefaultModel
	}

	// Create tool registry if pool is available
	var toolRegistry *tools.AgentToolRegistry
	if cfg.Pool != nil && cfg.UserID != "" {
		toolRegistry = tools.NewAgentToolRegistry(cfg.Pool, cfg.UserID)
	}

	return &BaseAgentV2{
		pool:           cfg.Pool,
		cfg:            cfg.Config,
		userID:         cfg.UserID,
		userName:       cfg.UserName,
		conversationID: cfg.ConversationID,
		model:          model,
		agentType:      cfg.AgentType,
		agentName:      cfg.AgentName,
		description:    cfg.Description,
		systemPrompt:   cfg.SystemPrompt,
		contextReqs:    cfg.ContextReqs,
		llmOptions:     services.DefaultLLMOptions(),
		toolRegistry:   toolRegistry,
		enabledTools:   cfg.EnabledTools,
	}
}

// Type returns the agent type
func (a *BaseAgentV2) Type() AgentTypeV2 {
	return a.agentType
}

// Name returns the agent name
func (a *BaseAgentV2) Name() string {
	return a.agentName
}

// Description returns the agent description
func (a *BaseAgentV2) Description() string {
	return a.description
}

// GetSystemPrompt returns the system prompt
func (a *BaseAgentV2) GetSystemPrompt() string {
	return a.systemPrompt
}

// GetContextRequirements returns what context the agent needs
func (a *BaseAgentV2) GetContextRequirements() ContextRequirements {
	return a.contextReqs
}

// SetModel sets the model to use
func (a *BaseAgentV2) SetModel(model string) {
	a.model = model
}

// SetOptions sets the LLM options
func (a *BaseAgentV2) SetOptions(opts services.LLMOptions) {
	a.llmOptions = opts
}

// GetOptions returns the current LLM options
func (a *BaseAgentV2) GetOptions() services.LLMOptions {
	return a.llmOptions
}

// GetEnabledTools returns the list of tools this agent can use
func (a *BaseAgentV2) GetEnabledTools() []string {
	return a.enabledTools
}

// GetToolDefinitions returns tool definitions for LLM function calling
func (a *BaseAgentV2) GetToolDefinitions() []map[string]interface{} {
	if a.toolRegistry == nil || len(a.enabledTools) == 0 {
		return nil
	}

	defs := make([]map[string]interface{}, 0)
	for _, toolName := range a.enabledTools {
		if tool, ok := a.toolRegistry.GetTool(toolName); ok {
			defs = append(defs, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        tool.Name(),
					"description": tool.Description(),
					"parameters":  tool.InputSchema(),
				},
			})
		}
	}
	return defs
}

// ExecuteTool executes a tool by name with the given input
func (a *BaseAgentV2) ExecuteTool(ctx context.Context, toolName string, input json.RawMessage) (string, error) {
	if a.toolRegistry == nil {
		return "", fmt.Errorf("tool registry not available")
	}

	// Check if tool is enabled for this agent
	enabled := false
	for _, t := range a.enabledTools {
		if t == toolName {
			enabled = true
			break
		}
	}
	if !enabled {
		return "", fmt.Errorf("tool %s not enabled for this agent", toolName)
	}

	return a.toolRegistry.ExecuteTool(ctx, toolName, input)
}

// Run executes the agent with streaming output
func (a *BaseAgentV2) Run(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		// Build messages with context
		messages := a.buildMessages(input)

		// Create LLM service
		llm := services.NewLLMService(a.cfg, a.model)
		llm.SetOptions(a.llmOptions)

		// Create artifact detector for streaming
		detector := streaming.NewArtifactDetector()

		// Stream response
		chunks, llmErrs := llm.StreamChat(ctx, messages, a.systemPrompt)

		// Process chunks through artifact detector
		for {
			select {
			case chunk, ok := <-chunks:
				if !ok {
					// Stream ended - flush detector
					for _, event := range detector.Flush() {
						events <- event
					}
					events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
					return
				}
				// Process chunk through artifact detector
				for _, event := range detector.ProcessChunk(chunk) {
					events <- event
				}

			case err := <-llmErrs:
				if err != nil {
					errs <- err
				}
				return

			case <-ctx.Done():
				return
			}
		}
	}()

	return events, errs
}

// buildMessages prepares messages for the LLM, including context
func (a *BaseAgentV2) buildMessages(input AgentInput) []services.ChatMessage {
	messages := make([]services.ChatMessage, 0, len(input.Messages)+1)

	// Prepend context as system message if available
	if input.Context != nil {
		contextContent := input.Context.FormatForAI()
		if contextContent != "" {
			contextMsg := services.ChatMessage{
				Role:    "system",
				Content: contextContent,
			}
			messages = append(messages, contextMsg)
		}
	}

	// Add conversation messages
	messages = append(messages, input.Messages...)

	return messages
}

// Pool returns the database pool
func (a *BaseAgentV2) Pool() *pgxpool.Pool {
	return a.pool
}

// Config returns the configuration
func (a *BaseAgentV2) Config() *config.Config {
	return a.cfg
}

// UserID returns the user ID
func (a *BaseAgentV2) UserID() string {
	return a.userID
}

// UserName returns the user name
func (a *BaseAgentV2) UserName() string {
	return a.userName
}

// ConversationID returns the conversation ID
func (a *BaseAgentV2) ConversationID() *uuid.UUID {
	return a.conversationID
}

// Model returns the current model
func (a *BaseAgentV2) Model() string {
	return a.model
}

// Agent constructors for the V2 architecture

// NewOrchestratorV2 creates a new orchestrator agent
func NewOrchestratorV2(ctx *AgentContextV2) AgentV2 {
	systemPrompt := prompts.ComposeWithUserContext(
		prompts_agents.OrchestratorAgentPrompt,
		ctx.UserName, "", "",
	)
	return NewBaseAgentV2(BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      AgentTypeV2Orchestrator,
		AgentName:      "OSA Orchestrator",
		Description:    "Primary interface that handles general requests and routes to specialists",
		SystemPrompt:   systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects:  true,
			NeedsTasks:     true,
			NeedsClients:   true,
			NeedsKnowledge: true,
		},
		EnabledTools: []string{
			"search_documents", "get_project", "get_task", "get_client",
			"create_task", "create_project", "create_client",
			"log_activity",
		},
	})
}

// NewDocumentAgentV2 creates a new document agent
func NewDocumentAgentV2(ctx *AgentContextV2) AgentV2 {
	systemPrompt := prompts.DefaultComposer.ComposeForDocument(prompts_agents.DocumentAgentPrompt)
	return NewBaseAgentV2(BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      AgentTypeV2Document,
		AgentName:      "Document Specialist",
		Description:    "Creates formal business documents: proposals, SOPs, reports, frameworks",
		SystemPrompt:   systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects:  true,
			NeedsKnowledge: true,
			NeedsClients:   true,
		},
		EnabledTools: []string{
			"search_documents", "get_project", "get_client",
			"log_activity",
		},
	})
}

// NewProjectAgentV2 creates a new project/planning agent
func NewProjectAgentV2(ctx *AgentContextV2) AgentV2 {
	systemPrompt := prompts.Compose(prompts_agents.ProjectAgentPrompt)
	return NewBaseAgentV2(BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      AgentTypeV2Project,
		AgentName:      "Project Specialist",
		Description:    "Project management and planning specialist",
		SystemPrompt:   systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects: true,
			NeedsTasks:    true,
			NeedsTeam:     true,
			NeedsClients:  true,
		},
		EnabledTools: []string{
			"create_project", "update_project", "get_project", "list_projects",
			"create_task", "bulk_create_tasks", "assign_task",
			"get_team_capacity", "search_documents",
			"log_activity",
		},
	})
}

// NewClientAgentV2 creates a new client agent
func NewClientAgentV2(ctx *AgentContextV2) AgentV2 {
	systemPrompt := prompts.Compose(prompts_agents.ClientAgentPrompt)
	return NewBaseAgentV2(BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      AgentTypeV2Client,
		AgentName:      "Client Specialist",
		Description:    "Client relationship and pipeline specialist",
		SystemPrompt:   systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsClients:   true,
			NeedsProjects:  true,
			NeedsKnowledge: true,
		},
		EnabledTools: []string{
			"create_client", "update_client", "get_client",
			"log_client_interaction", "update_client_pipeline",
			"search_documents", "get_project",
			"log_activity",
		},
	})
}

// NewAnalystAgentV2 creates a new analyst agent
func NewAnalystAgentV2(ctx *AgentContextV2) AgentV2 {
	systemPrompt := prompts.DefaultComposer.ComposeForAnalysis(prompts_agents.AnalystAgentPrompt)
	return NewBaseAgentV2(BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      AgentTypeV2Analyst,
		AgentName:      "Analyst Specialist",
		Description:    "Data analysis and insights specialist",
		SystemPrompt:   systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects: true,
			NeedsTasks:    true,
			NeedsClients:  true,
			NeedsTeam:     true,
		},
		EnabledTools: []string{
			"query_metrics", "get_team_capacity",
			"list_projects", "list_tasks", "get_project",
			"search_documents",
			"log_activity",
		},
	})
}

// NewTaskAgentV2 creates a new task management agent
func NewTaskAgentV2(ctx *AgentContextV2) AgentV2 {
	systemPrompt := prompts.Compose(prompts_agents.TaskAgentPrompt)
	return NewBaseAgentV2(BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      AgentTypeV2Task,
		AgentName:      "Task Specialist",
		Description:    "Task management, prioritization, scheduling, and dependencies",
		SystemPrompt:   systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects: true,
			NeedsTasks:    true,
			NeedsTeam:     true,
		},
		EnabledTools: []string{
			"create_task", "update_task", "get_task", "list_tasks",
			"bulk_create_tasks", "move_task", "assign_task",
			"get_team_capacity", "get_project",
			"log_activity",
		},
	})
}
