// Stub implementation of the MIOSA SDK for local/Docker builds.
package osasdk

import (
	"context"
	"fmt"
	"time"
)

// Event type constants
const (
	EventThinking       = "thinking"
	EventResponse       = "response"
	EventSkillStarted   = "skill_started"
	EventSkillCompleted = "skill_completed"
	EventSkillFailed    = "skill_failed"
	EventError          = "error"
	EventConnected      = "connected"
	EventSignal         = "signal"
)

// Event represents an OSA streaming event.
type Event struct {
	Type string
	Data map[string]interface{}
}

// SwarmRequest is the payload for launching a swarm.
type SwarmRequest struct {
	Pattern   string
	Task      string
	Config    map[string]interface{}
	MaxAgents int
	SessionID string
}

// SwarmResponse is returned after launching a swarm.
type SwarmResponse struct {
	SwarmID   string `json:"swarm_id"`
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
}

// SwarmStatus represents the current state of a swarm.
type SwarmStatus struct {
	SwarmID    string                 `json:"swarm_id"`
	Status     string                 `json:"status"`
	Pattern    string                 `json:"pattern"`
	AgentCount int                    `json:"agent_count"`
	Progress   float64                `json:"progress"`
	Results    map[string]interface{} `json:"results,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// ToolDefinition describes an available tool.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Source      string                 `json:"source,omitempty"`
}

// ToolResult is the output of executing a tool.
type ToolResult struct {
	Success bool                   `json:"success"`
	Output  interface{}            `json:"output,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// Instruction is a CloudEvent-like dispatch payload.
type Instruction struct {
	SpecVersion string
	Type        string
	Source      string
	ID          string
	Data        map[string]interface{}
}

// OrchestrateRequest is the payload for the Orchestrate RPC.
type OrchestrateRequest struct {
	Input       string
	UserID      string
	WorkspaceID string
	SessionID   string
	Phase       string
	Context     map[string]interface{}
}

// OrchestrateResponse is the response from Orchestrate.
type OrchestrateResponse struct {
	Success     bool
	Output      string
	SessionID   string
	AgentsUsed  []string
	ExecutionMS int64
	Metadata    map[string]interface{}
	NextStep    string
}

// AppGenerationRequest is the payload for generating an app.
type AppGenerationRequest struct {
	UserID      string
	WorkspaceID string
	Name        string
	Description string
	Type        string
	Parameters  map[string]interface{}
}

// AppGenerationResponse is returned after generating an app.
type AppGenerationResponse struct {
	AppID       string
	Status      string
	WorkspaceID string
	Message     string
	Data        map[string]interface{}
	CreatedAt   string
}

// AppStatusResponse is returned when querying app generation status.
type AppStatusResponse struct {
	AppID       string
	Status      string
	Progress    float64
	CurrentStep string
	Output      map[string]interface{}
	Error       string
	Metadata    map[string]interface{}
	UpdatedAt   string
}

// GenerateFromTemplateRequest is the payload for template-based generation.
type GenerateFromTemplateRequest struct {
	TemplateName string
	Variables    map[string]interface{}
	UserID       string
	WorkspaceID  string
}

// HealthResponse is returned from health checks.
type HealthResponse struct {
	Status   string
	Version  string
	Provider string
}

// WorkspaceInfo represents a single workspace.
type WorkspaceInfo struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	CreatedAt   string
	UpdatedAt   string
}

// WorkspacesResponse is returned from workspace listing.
type WorkspacesResponse struct {
	Workspaces []WorkspaceInfo
	Total      int
}

// CloudConfig configures a cloud client.
type CloudConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// LocalConfig configures a local client.
type LocalConfig struct {
	BaseURL      string
	SharedSecret string
	Timeout      time.Duration
	Resilience   *ResilienceConfig
}

// ResilienceConfig controls retry/circuit-breaker behavior.
type ResilienceConfig struct {
	Enabled bool
}

// APIError represents an error from the OSA API.
type APIError struct {
	StatusCode int
	ErrorCode  string
	Details    string
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return e.Details
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.ErrorCode)
}

// Client is the interface for OSA interactions.
type Client interface {
	Health(ctx context.Context) (*HealthResponse, error)
	Orchestrate(ctx context.Context, req OrchestrateRequest) (*OrchestrateResponse, error)
	GenerateApp(ctx context.Context, req AppGenerationRequest) (*AppGenerationResponse, error)
	GetAppStatus(ctx context.Context, appID string) (*AppStatusResponse, error)
	GenerateAppFromTemplate(ctx context.Context, req GenerateFromTemplateRequest) (*AppGenerationResponse, error)
	GetWorkspaces(ctx context.Context) (*WorkspacesResponse, error)
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

// stubClient is a no-op implementation for builds without the real SDK.
type stubClient struct{}

func (c *stubClient) Health(ctx context.Context) (*HealthResponse, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) Orchestrate(ctx context.Context, req OrchestrateRequest) (*OrchestrateResponse, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) GenerateApp(ctx context.Context, req AppGenerationRequest) (*AppGenerationResponse, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) GetAppStatus(ctx context.Context, appID string) (*AppStatusResponse, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) GenerateAppFromTemplate(ctx context.Context, req GenerateFromTemplateRequest) (*AppGenerationResponse, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) GetWorkspaces(ctx context.Context) (*WorkspacesResponse, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) Stream(ctx context.Context, sessionID string) (<-chan Event, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) LaunchSwarm(ctx context.Context, req SwarmRequest) (*SwarmResponse, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) ListSwarms(ctx context.Context) ([]SwarmStatus, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) GetSwarm(ctx context.Context, swarmID string) (*SwarmStatus, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) CancelSwarm(ctx context.Context, swarmID string) error {
	return fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) DispatchInstruction(ctx context.Context, agentID string, instruction Instruction) error {
	return fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error) {
	return nil, fmt.Errorf("stub: OSA not available")
}
func (c *stubClient) Close() error { return nil }

// NewCloudClient returns a stub client for local builds.
func NewCloudClient(cfg CloudConfig) (Client, error) {
	return &stubClient{}, nil
}

// NewLocalClient returns a stub client for local builds.
func NewLocalClient(cfg LocalConfig) (Client, error) {
	return &stubClient{}, nil
}
