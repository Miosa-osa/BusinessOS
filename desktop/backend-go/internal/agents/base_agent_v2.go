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
	pool            *pgxpool.Pool
	cfg             *config.Config
	userID          string
	userName        string
	conversationID  *uuid.UUID
	model           string
	agentType       AgentTypeV2
	agentName       string
	description     string
	systemPrompt    string
	focusModePrompt string // Focus mode specific prompt prefix
	contextReqs     ContextRequirements
	llmOptions      services.LLMOptions
	toolRegistry    *tools.AgentToolRegistry
	enabledTools    []string // Tool names this agent can use
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

// SetCustomSystemPrompt overrides the system prompt with a custom one (for custom agents)
func (a *BaseAgentV2) SetCustomSystemPrompt(prompt string) {
	fmt.Printf("[Agent %p] SetCustomSystemPrompt called with %d chars\n", a, len(prompt))
	if prompt != "" {
		fmt.Printf("[Agent %p] Setting custom systemPrompt: %s\n", a, prompt[:min(100, len(prompt))])
		a.systemPrompt = prompt
		fmt.Printf("[Agent %p] systemPrompt now has %d chars\n", a, len(a.systemPrompt))
	} else {
		fmt.Printf("[Agent %p] SetCustomSystemPrompt called with empty prompt - NOT setting\n", a)
	}
}

// SetFocusModePrompt sets a focus mode specific prompt prefix
func (a *BaseAgentV2) SetFocusModePrompt(prompt string) {
	a.focusModePrompt = prompt
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

// RunWithTools executes the agent with tool calling support (non-streaming)
func (a *BaseAgentV2) RunWithTools(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		// Build messages with context
		messages := a.buildMessages(input)

		// Check if we have tools enabled
		if a.toolRegistry == nil || len(a.enabledTools) == 0 {
			// No tools, fall back to regular streaming
			a.runStreaming(ctx, input, events, errs)
			return
		}

		// Get tool definitions for enabled tools
		toolDefs := make([]services.ToolDefinition, 0)
		for _, toolName := range a.enabledTools {
			if tool, ok := a.toolRegistry.GetTool(toolName); ok {
				toolDefs = append(toolDefs, services.ToolDefinition{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  tool.InputSchema(),
				})
			}
		}

		// Create Groq service for tool calling
		groqService := services.NewGroqService(a.cfg, a.model)
		groqService.SetOptions(a.llmOptions)

		// Build system prompt with thinking instructions if enabled
		systemPrompt := a.buildSystemPromptWithThinking()

		// First call with tools
		resp, err := groqService.ChatWithTools(ctx, messages, systemPrompt, toolDefs)
		if err != nil {
			errs <- err
			return
		}

		// Process tool calls if any
		if len(resp.ToolCalls) > 0 {
			// Send thinking event
			events <- streaming.StreamEvent{
				Type: streaming.EventTypeThinking,
				Data: "Executing tools...",
			}

			toolResults := make(map[string]string)
			for _, tc := range resp.ToolCalls {
				// Send tool execution event
				events <- streaming.StreamEvent{
					Type: streaming.EventTypeThinking,
					Data: fmt.Sprintf("Running tool: %s", tc.Name),
				}

				// Execute the tool
				result, err := a.toolRegistry.ExecuteTool(ctx, tc.Name, json.RawMessage(tc.Arguments))
				if err != nil {
					toolResults[tc.ID] = fmt.Sprintf("Error: %s", err.Error())
				} else {
					toolResults[tc.ID] = result
				}

				// Send tool result as event
				events <- streaming.StreamEvent{
					Type: streaming.EventTypeToken,
					Data: fmt.Sprintf("\n\n**Tool Result (%s):**\n%s\n\n", tc.Name, toolResults[tc.ID]),
				}
			}

			// Continue conversation with tool results
			finalResponse, err := groqService.ContinueWithToolResults(ctx, messages, systemPrompt, toolResults)
			if err != nil {
				errs <- err
				return
			}

			// Stream the final response
			for _, chunk := range splitIntoChunks(finalResponse, 50) {
				events <- streaming.StreamEvent{
					Type: streaming.EventTypeToken,
					Data: chunk,
				}
			}
		} else {
			// No tool calls, stream the response directly
			for _, chunk := range splitIntoChunks(resp.Content, 50) {
				events <- streaming.StreamEvent{
					Type: streaming.EventTypeToken,
					Data: chunk,
				}
			}
		}

		events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
	}()

	return events, errs
}

// runStreaming handles regular streaming without tools
func (a *BaseAgentV2) runStreaming(ctx context.Context, input AgentInput, events chan<- streaming.StreamEvent, errs chan<- error) {
	messages := a.buildMessages(input)
	llm := services.NewLLMService(a.cfg, a.model)
	llm.SetOptions(a.llmOptions)
	detector := streaming.NewArtifactDetector()

	// Build system prompt with thinking instructions if enabled
	systemPrompt := a.buildSystemPromptWithThinking()
	chunks, llmErrs := llm.StreamChat(ctx, messages, systemPrompt)

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				for _, event := range detector.Flush() {
					events <- event
				}
				events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
				return
			}
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
}

// splitIntoChunks splits a string into chunks for streaming
func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
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

		// Build system prompt with thinking instructions if enabled
		systemPrompt := a.buildSystemPromptWithThinking()

		// Create artifact detector for streaming
		detector := streaming.NewArtifactDetector()

		// Stream response
		chunks, llmErrs := llm.StreamChat(ctx, messages, systemPrompt)

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

// buildSystemPromptWithThinking returns the system prompt with thinking instructions if enabled
func (a *BaseAgentV2) buildSystemPromptWithThinking() string {
	// Debug: log the system prompt being used
	fmt.Printf("[Agent %p] buildSystemPromptWithThinking called - systemPrompt length: %d\n", a, len(a.systemPrompt))
	if len(a.systemPrompt) > 0 {
		fmt.Printf("[Agent %p] systemPrompt preview: %s\n", a, a.systemPrompt[:min(150, len(a.systemPrompt))])
	}

	// Start with base system prompt
	result := a.systemPrompt

	// Prepend focus mode prompt if set
	if a.focusModePrompt != "" {
		result = a.focusModePrompt + "\n\n" + result
		fmt.Printf("[Agent] Applied focus mode prompt prefix (%d chars)\n", len(a.focusModePrompt))
	}

	if a.llmOptions.ThinkingEnabled {
		// Use custom thinking instruction from template if provided, otherwise use default
		thinkingInstruction := prompts.ThinkingInstruction
		if a.llmOptions.ThinkingInstruction != "" {
			thinkingInstruction = a.llmOptions.ThinkingInstruction
			fmt.Printf("[Agent] ThinkingEnabled=true, using custom template instruction (%d chars)\n", len(thinkingInstruction))
		} else {
			fmt.Printf("[Agent] ThinkingEnabled=true, using default thinking instruction (%d chars)\n", len(thinkingInstruction))
		}
		return result + "\n\n" + thinkingInstruction
	}
	fmt.Printf("[Agent] ThinkingEnabled=false, using base prompt\n")
	return result
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
		prompts_agents.OrchestratorAgentPrompt+prompts.ArtifactInstruction,
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
			"create_artifact", "log_activity",
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
			"create_artifact", "search_documents", "get_project", "get_client",
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
			"create_artifact", "log_activity",
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
			"create_artifact", "log_activity",
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
			"search_documents", "create_artifact",
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
