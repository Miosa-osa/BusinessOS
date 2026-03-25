package agents

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// AgentType defines the type of agent
type AgentType string

const (
	AgentTypeOrchestrator AgentType = "orchestrator"
	AgentTypeAnalyst      AgentType = "analyst"
	AgentTypeDocument     AgentType = "document"
	AgentTypeProject      AgentType = "project"
	AgentTypeTask         AgentType = "task"
	AgentTypeClient       AgentType = "client"
)

// Agent interface for different agent types
type Agent interface {
	Process(input interface{}) (interface{}, error)
	SetModel(model string)
	GetSystemPrompt() string
	SetOptions(options interface{})
	SetProfileContext(context interface{})
	SetRoleContextPrompt(prompt string)
	Execute(input AgentInput) (interface{}, error)
	Run(ctx interface{}, input AgentInput) (<-chan streaming.StreamEvent, <-chan error)
	RegisterExternalTool(tool interface{})
	SetMemoryContext(context string)
	SetSkillsPrompt(prompt string)
	SetFocusModePrompt(prompt string)
	SetCustomSystemPrompt(prompt string)
	SetOutputStylePrompt(prompt string)
	SetGenreContext(context string)
	SetTieredContext(context interface{})
}

// AgentInput defines input for agents
type AgentInput struct {
	UserID          string
	UserName        string
	ProjectID       interface{}
	MessageText     string
	Context         interface{} // map[string]interface{} or *services.TieredContext
	Messages        interface{} // []ChatMessage
	TieredContext   interface{} // *services.TieredContext or map[string]interface{}
	FocusMode       interface{} // bool or string
	ConversationID  interface{} // *uuid.UUID
	MemoryContext   string
	RoleContext     interface{}
	SignalEnvelope  interface{}
}

// OrchestratorCOT handles chain-of-thought orchestration
type OrchestratorCOT struct {
	Steps []string
}

// NewOrchestratorCOT creates a new orchestrator
func NewOrchestratorCOT(pool interface{}, cfg interface{}, registry interface{}) *OrchestratorCOT {
	return &OrchestratorCOT{
		Steps: make([]string, 0),
	}
}

// ProcessWithCOT processes input with chain-of-thought
func (cot *OrchestratorCOT) ProcessWithCOT(ctx interface{}, input interface{}, userID interface{}, userName interface{}, convUUID interface{}, llmOptions interface{}) (<-chan streaming.StreamEvent, <-chan error, interface{}) {
	eventsCh := make(chan streaming.StreamEvent, 10)
	errsCh := make(chan error, 1)
	go func() {
		close(eventsCh)
		close(errsCh)
	}()
	return eventsCh, errsCh, nil
}

// BuildSignalAnnotation builds a signal annotation string
func BuildSignalAnnotation(signal interface{}, tieredContext interface{}) string {
	return ""
}

// BaseAgent is exported for handler references
type BaseAgent = baseAgent

// AgentRegistry manages agent instances
type AgentRegistry struct {
	pool               *pgxpool.Pool
	cfg                *config.Config
	embeddingService   interface{} // *services.EmbeddingService
	promptPersonalizer interface{} // *services.PromptPersonalizer
	signalHints        interface{} // map[string]interface{} or feedback.SignalHintProvider
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry(
	pool *pgxpool.Pool,
	cfg *config.Config,
	embeddingService interface{},
	promptPersonalizer interface{},
	signalHints interface{},
) *AgentRegistry {
	return &AgentRegistry{
		pool:               pool,
		cfg:                cfg,
		embeddingService:   embeddingService,
		promptPersonalizer: promptPersonalizer,
		signalHints:        signalHints,
	}
}

// GetAgent retrieves an agent of the specified type
func (ar *AgentRegistry) GetAgent(
	agentType AgentType,
	userID string,
	userName string,
	embedService interface{},
	personalizer interface{},
) Agent {
	// Parse userID to uuid if it's a valid UUID, otherwise use a nil UUID
	var parsedID uuid.UUID
	if uid, err := uuid.Parse(userID); err == nil {
		parsedID = uid
	}

	return &baseAgent{
		agentType: agentType,
		userID:    parsedID,
		userName:  userName,
	}
}

// baseAgent is a simple implementation
type baseAgent struct {
	agentType      AgentType
	userID         uuid.UUID
	userName       string
	model          string
	systemPrompt   string
	options        interface{}
	profileContext interface{}
}

// Process implements Agent interface
func (ba *baseAgent) Process(input interface{}) (interface{}, error) {
	return nil, nil
}

// SetModel sets the LLM model for the agent
func (ba *baseAgent) SetModel(model string) {
	ba.model = model
}

// GetSystemPrompt returns the system prompt for the agent
func (ba *baseAgent) GetSystemPrompt() string {
	if ba.systemPrompt == "" {
		return "You are a helpful AI assistant."
	}
	return ba.systemPrompt
}

// SetOptions sets agent options
func (ba *baseAgent) SetOptions(options interface{}) {
	ba.options = options
}

// SetProfileContext sets user profile context
func (ba *baseAgent) SetProfileContext(context interface{}) {
	ba.profileContext = context
}

// SetRoleContextPrompt sets the role context prompt
func (ba *baseAgent) SetRoleContextPrompt(prompt string) {
	ba.systemPrompt = prompt
}

// Execute runs the agent with the given input
func (ba *baseAgent) Execute(input AgentInput) (interface{}, error) {
	return ba.Process(input)
}

// Run executes the agent with context and returns events and errors channels
func (ba *baseAgent) Run(ctx interface{}, input AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	// Create stub channels
	eventsCh := make(chan streaming.StreamEvent, 10)
	errsCh := make(chan error, 1)
	go func() {
		close(eventsCh)
		close(errsCh)
	}()
	return eventsCh, errsCh
}

// RegisterExternalTool registers an external tool with the agent
func (ba *baseAgent) RegisterExternalTool(tool interface{}) {
	// Stub implementation
}

// SetMemoryContext sets memory context for the agent
func (ba *baseAgent) SetMemoryContext(context string) {
	ba.profileContext = context
}

// SetSkillsPrompt sets the skills prompt for the agent
func (ba *baseAgent) SetSkillsPrompt(prompt string) {
	ba.systemPrompt = ba.systemPrompt + "\n\n" + prompt
}

// SetFocusModePrompt sets the focus mode prompt
func (ba *baseAgent) SetFocusModePrompt(prompt string) {
	ba.systemPrompt = ba.systemPrompt + "\n\n" + prompt
}

// SetCustomSystemPrompt sets a custom system prompt
func (ba *baseAgent) SetCustomSystemPrompt(prompt string) {
	ba.systemPrompt = prompt
}

// SetOutputStylePrompt sets the output style prompt
func (ba *baseAgent) SetOutputStylePrompt(prompt string) {
	ba.systemPrompt = ba.systemPrompt + "\n\n" + prompt
}

// SetGenreContext sets the genre context
func (ba *baseAgent) SetGenreContext(context string) {
	ba.options = context
}

// SetTieredContext sets the tiered context
func (ba *baseAgent) SetTieredContext(context interface{}) {
	ba.profileContext = context
}
