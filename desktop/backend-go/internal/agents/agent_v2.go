package agents

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// AgentTypeV2 identifies the type of agent in the new architecture
type AgentTypeV2 string

const (
	AgentTypeV2Orchestrator AgentTypeV2 = "orchestrator"
	AgentTypeV2Document     AgentTypeV2 = "document"
	AgentTypeV2Project      AgentTypeV2 = "project"
	AgentTypeV2Task         AgentTypeV2 = "task"
	AgentTypeV2Client       AgentTypeV2 = "client"
	AgentTypeV2Analyst      AgentTypeV2 = "analyst"
)

// AgentV2 defines the interface for the new agent architecture
type AgentV2 interface {
	// Identity
	Type() AgentTypeV2
	Name() string
	Description() string

	// Configuration
	GetSystemPrompt() string
	GetContextRequirements() ContextRequirements

	// Execution - returns streaming events instead of raw strings
	Run(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error)
	RunWithTools(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error)

	// Options
	SetModel(model string)
	SetOptions(opts services.LLMOptions)
	SetCustomSystemPrompt(prompt string) // For custom agents with user-defined prompts
	SetFocusModePrompt(prompt string)    // For focus mode specific prompt prefix
}

// ContextRequirements declares what context an agent needs
type ContextRequirements struct {
	NeedsProjects    bool
	NeedsTasks       bool
	NeedsClients     bool
	NeedsTeam        bool
	NeedsKnowledge   bool
	NeedsMetrics     bool
	NeedsFullHistory bool
	MaxContextTokens int
	PrioritySections []string
}

// AgentInput contains everything an agent needs to process a request
type AgentInput struct {
	Messages       []services.ChatMessage
	Context        *services.TieredContext
	Selections     UserSelections
	FocusMode      string
	FocusModeOpts  map[string]string
	ConversationID uuid.UUID
	UserID         string
	UserName       string
}

// UserSelections represents what the user selected in the context bar
type UserSelections struct {
	ProjectID  *uuid.UUID
	ContextIDs []uuid.UUID
	NodeID     *uuid.UUID
	ClientID   *uuid.UUID
}

// AgentContextV2 holds context information for agent initialization
type AgentContextV2 struct {
	Pool             *pgxpool.Pool
	Config           *config.Config
	UserID           string
	UserName         string
	ConversationID   *uuid.UUID
	TieredContext    *services.TieredContext
	EmbeddingService *services.EmbeddingService
}

// Intent represents the classified intent of a user message
type Intent struct {
	Category       string // "document", "project", "client", "analysis", "general"
	ShouldDelegate bool
	TargetAgent    AgentTypeV2
	Confidence     float64
	Reasoning      string
}

// AgentRegistryV2 manages agent creation and retrieval for the new architecture
type AgentRegistryV2 struct {
	pool             *pgxpool.Pool
	config           *config.Config
	embeddingService *services.EmbeddingService
}

// NewAgentRegistryV2 creates a new agent registry
func NewAgentRegistryV2(pool *pgxpool.Pool, cfg *config.Config, embeddingService *services.EmbeddingService) *AgentRegistryV2 {
	return &AgentRegistryV2{
		pool:             pool,
		config:           cfg,
		embeddingService: embeddingService,
	}
}

// GetAgent creates an agent of the specified type
func (r *AgentRegistryV2) GetAgent(
	agentType AgentTypeV2,
	userID string,
	userName string,
	conversationID *uuid.UUID,
	tieredContext *services.TieredContext,
) AgentV2 {
	ctx := &AgentContextV2{
		Pool:             r.pool,
		Config:           r.config,
		UserID:           userID,
		UserName:         userName,
		ConversationID:   conversationID,
		TieredContext:    tieredContext,
		EmbeddingService: r.embeddingService,
	}

	switch agentType {
	case AgentTypeV2Orchestrator:
		return NewOrchestratorV2(ctx)
	case AgentTypeV2Document:
		return NewDocumentAgentV2(ctx)
	case AgentTypeV2Project:
		return NewProjectAgentV2(ctx)
	case AgentTypeV2Task:
		return NewTaskAgentV2(ctx)
	case AgentTypeV2Client:
		return NewClientAgentV2(ctx)
	case AgentTypeV2Analyst:
		return NewAnalystAgentV2(ctx)
	default:
		return NewOrchestratorV2(ctx)
	}
}

// GetAgentForFocusModeV2 maps focus modes to appropriate agents
func GetAgentForFocusModeV2(focusMode string) AgentTypeV2 {
	switch focusMode {
	case "write":
		return AgentTypeV2Document
	case "analyze", "research":
		return AgentTypeV2Analyst
	case "plan", "build":
		return AgentTypeV2Project
	default:
		return AgentTypeV2Orchestrator
	}
}

// AgentTypeV2FromString converts a string to AgentTypeV2
func AgentTypeV2FromString(s string) AgentTypeV2 {
	switch s {
	case "orchestrator":
		return AgentTypeV2Orchestrator
	case "document":
		return AgentTypeV2Document
	case "project":
		return AgentTypeV2Project
	case "task":
		return AgentTypeV2Task
	case "client":
		return AgentTypeV2Client
	case "analyst", "analysis":
		return AgentTypeV2Analyst
	default:
		return AgentTypeV2Orchestrator
	}
}
