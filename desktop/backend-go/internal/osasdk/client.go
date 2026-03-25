// Package osasdk provides a local stub for the MIOSA OSA SDK.
// This replaces the external github.com/Miosa-osa/sdk-go dependency for ChatmanGPT's
// BusinessOS fork. The original SDK is a private MIOSA repository.
package osasdk

import (
	"context"
	"fmt"
	"time"
)

// Event represents a streaming event from OSA.
type Event struct {
	Type EventType
	Data map[string]any
}

// EventType represents the type of OSA event.
type EventType string

const (
	EventThinking       EventType = "thinking"
	EventResponse      EventType = "response"
	EventSkillStarted  EventType = "skill_started"
	EventSkillCompleted EventType = "skill_completed"
	EventSkillFailed    EventType = "skill_failed"
	EventError         EventType = "error"
	EventConnected     EventType = "connected"
	EventSignal        EventType = "signal"
	EventStreamingToken EventType = "streaming_token"
)

// Client is the OSA cloud client interface.
type Client interface {
	Health(ctx context.Context) (*HealthResponse, error)
	Orchestrate(ctx context.Context, req OrchestrateRequest) (*OrchestrateResponse, error)
	GenerateApp(ctx context.Context, req AppGenerationRequest) (*AppGenerationResponse, error)
	GetAppStatus(ctx context.Context, appID string) (*AppStatusResponse, error)
	GetWorkspaces(ctx context.Context) (*WorkspacesResponse, error)
	GenerateAppFromTemplate(ctx context.Context, req GenerateFromTemplateRequest) (*AppGenerationResponse, error)
	Stream(ctx context.Context, sessionID string) (<-chan Event, error)
	LaunchSwarm(ctx context.Context, req SwarmRequest) (*SwarmResponse, error)
	ListSwarms(ctx context.Context) ([]SwarmStatus, error)
	GetSwarm(ctx context.Context, swarmID string) (*SwarmStatus, error)
	CancelSwarm(ctx context.Context, swarmID string) error
	DispatchInstruction(ctx context.Context, agentID string, instruction Instruction) error
	ListTools(ctx context.Context) ([]ToolDefinition, error)
	ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error)
	Close() error
}

// CloudConfig configures the cloud client connection.
type CloudConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// HealthResponse is returned by the health check endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Provider  string `json:"provider,omitempty"`
	Model     string `json:"model,omitempty"`
}

// OrchestrateRequest is the payload for OSA orchestration.
type OrchestrateRequest struct {
	UserID      string                 `json:"user_id"`
	Input       string                 `json:"input"`
	SessionID   string                 `json:"session_id,omitempty"`
	Phase       string                 `json:"phase,omitempty"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// OrchestrateResponse is the response from OSA orchestration.
type OrchestrateResponse struct {
	SessionID   string                 `json:"session_id"`
	Status      string                 `json:"status"`
	Success     bool                   `json:"success"`
	Output      string                 `json:"output"`
	AgentsUsed  []string               `json:"agents_used"`
	ExecutionMS int64                  `json:"execution_ms"`
	Metadata    map[string]interface{} `json:"metadata"`
	NextStep    string                 `json:"next_step"`
}

// AppGenerationRequest triggers application generation
type AppGenerationRequest struct {
	UserID      string                 `json:"user_id"`
	WorkspaceID string                 `json:"workspace_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// AppGenerationResponse is returned from app generation
type AppGenerationResponse struct {
	AppID       string                 `json:"app_id"`
	Status      string                 `json:"status"`
	WorkspaceID string                 `json:"workspace_id"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   string                 `json:"created_at"`
}

// AppStatusResponse contains app generation status
type AppStatusResponse struct {
	AppID       string                 `json:"app_id"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress"`
	CurrentStep string                 `json:"current_step"`
	Output      string                 `json:"output"`
	Error       string                 `json:"error"`
	Metadata    map[string]interface{} `json:"metadata"`
	UpdatedAt   string                 `json:"updated_at"`
}

// Workspace represents an OSA workspace
type Workspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// WorkspacesResponse contains workspace list
type WorkspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
	Total      int         `json:"total"`
}

// GenerateFromTemplateRequest generates app from template
type GenerateFromTemplateRequest struct {
	TemplateName string                 `json:"template_name"`
	Variables    map[string]interface{} `json:"variables"`
	UserID       string                 `json:"user_id"`
	WorkspaceID  string                 `json:"workspace_id,omitempty"`
}

// SwarmRequest initiates a multi-agent swarm
type SwarmRequest struct {
	Pattern    string                 `json:"pattern"`
	Task       string                 `json:"task"`
	Agents     []string               `json:"agents"`
	Config     map[string]interface{} `json:"config,omitempty"`
	MaxAgents  int                    `json:"max_agents,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
}

// SwarmResponse is returned from swarm launch
type SwarmResponse struct {
	SwarmID string    `json:"swarm_id"`
	Status  string    `json:"status"`
	Agents  []string  `json:"agents"`
	Started time.Time `json:"started"`
}

// SwarmStatus contains the status of a swarm
type SwarmStatus struct {
	SwarmID    string                 `json:"swarm_id"`
	Status     string                 `json:"status"`
	Progress   int                    `json:"progress"`
	Agents     []string               `json:"agents"`
	Results    map[string]interface{} `json:"results,omitempty"`
	Error      string                 `json:"error,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
}

// Instruction is sent to a fleet agent
type Instruction struct {
	ID         string                 `json:"id"`
	Action     string                 `json:"action"`
	Params     map[string]interface{} `json:"params"`
	SpecVersion string                 `json:"spec_version,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Source     string                 `json:"source,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// ToolDefinition describes an available tool
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Params      map[string]interface{} `json:"params"`
	Version     string                 `json:"version"`
}

// ToolResult is returned from tool execution
type ToolResult struct {
	Success bool                   `json:"success"`
	Output  map[string]interface{} `json:"output"`
	Error   string                 `json:"error,omitempty"`
}

// LocalConfig configures local SDK connection
type LocalConfig struct {
	BaseURL      string
	SharedSecret string
	Timeout      time.Duration
	Resilience   *ResilienceConfig
}

// ResilienceConfig controls SDK resilience behavior
type ResilienceConfig struct {
	Enabled bool
}

// APIError represents an API error from OSA
type APIError struct {
	StatusCode int
	ErrorCode  string
	Details    string
}

// Error implements the error interface for APIError
func (e *APIError) Error() string {
	if e.Details != "" {
		return e.Details
	}
	return e.ErrorCode
}

// NewLocalClient creates a new local OSA client
func NewLocalClient(cfg LocalConfig) (Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8089"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &cloudClient{
		baseURL: cfg.BaseURL,
		timeout: cfg.Timeout,
	}, nil
}

// cloudClient is a minimal HTTP-based implementation of the OSA client.
type cloudClient struct {
	apiKey  string
	baseURL string
	timeout time.Duration
}

// NewCloudClient creates a new OSA cloud client.
func NewCloudClient(cfg CloudConfig) (Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.miosa.ai"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &cloudClient{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		timeout: cfg.Timeout,
	}, nil
}

func (c *cloudClient) Health(ctx context.Context) (*HealthResponse, error) {
	// In local mode (OSA_MODE=local), return a healthy status without
	// making network calls. The cloud sync service handles mode checking.
	return &HealthResponse{
		Status:    "ok",
		Version:   "local-stub",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (c *cloudClient) Orchestrate(ctx context.Context, req OrchestrateRequest) (*OrchestrateResponse, error) {
	// In local mode, return a mock response. The actual OSA integration
	// happens via HTTP to localhost:9089, not through this SDK path.
	return &OrchestrateResponse{
		SessionID: fmt.Sprintf("local-manifest-%d", time.Now().Unix()),
		Status:    "completed",
	}, nil
}

// GenerateApp implementation
func (c *cloudClient) GenerateApp(ctx context.Context, req AppGenerationRequest) (*AppGenerationResponse, error) {
	return &AppGenerationResponse{
		AppID:  fmt.Sprintf("app-%d", time.Now().Unix()),
		Status: "generating",
	}, nil
}

// GetAppStatus implementation
func (c *cloudClient) GetAppStatus(ctx context.Context, appID string) (*AppStatusResponse, error) {
	return &AppStatusResponse{
		AppID:  appID,
		Status: "completed",
	}, nil
}

// GetWorkspaces implementation
func (c *cloudClient) GetWorkspaces(ctx context.Context) (*WorkspacesResponse, error) {
	return &WorkspacesResponse{
		Workspaces: []Workspace{},
		Total:      0,
	}, nil
}

// GenerateAppFromTemplate implementation
func (c *cloudClient) GenerateAppFromTemplate(ctx context.Context, req GenerateFromTemplateRequest) (*AppGenerationResponse, error) {
	return &AppGenerationResponse{
		AppID:  fmt.Sprintf("app-%d", time.Now().Unix()),
		Status: "generating",
	}, nil
}

// Stream implementation
func (c *cloudClient) Stream(ctx context.Context, sessionID string) (<-chan Event, error) {
	ch := make(chan Event)
	close(ch)
	return ch, nil
}

// LaunchSwarm implementation
func (c *cloudClient) LaunchSwarm(ctx context.Context, req SwarmRequest) (*SwarmResponse, error) {
	return &SwarmResponse{
		SwarmID: fmt.Sprintf("swarm-%d", time.Now().Unix()),
		Status:  "running",
		Started: time.Now(),
	}, nil
}

// ListSwarms implementation
func (c *cloudClient) ListSwarms(ctx context.Context) ([]SwarmStatus, error) {
	return []SwarmStatus{}, nil
}

// GetSwarm implementation
func (c *cloudClient) GetSwarm(ctx context.Context, swarmID string) (*SwarmStatus, error) {
	return &SwarmStatus{
		SwarmID: swarmID,
		Status:  "running",
	}, nil
}

// CancelSwarm implementation
func (c *cloudClient) CancelSwarm(ctx context.Context, swarmID string) error {
	return nil
}

// DispatchInstruction implementation
func (c *cloudClient) DispatchInstruction(ctx context.Context, agentID string, instruction Instruction) error {
	return nil
}

// ListTools implementation
func (c *cloudClient) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	return []ToolDefinition{}, nil
}

// ExecuteTool implementation
func (c *cloudClient) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error) {
	return &ToolResult{
		Success: true,
		Output:  make(map[string]interface{}),
	}, nil
}

// Close releases resources held by the client.
func (c *cloudClient) Close() error {
	return nil
}
