package osa

import (
	"time"

	"github.com/google/uuid"
)

// AppGenerationRequest represents a request to generate an application
type AppGenerationRequest struct {
	UserID      uuid.UUID              `json:"user_id"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // "full-stack", "module", "tool"
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// AppGenerationResponse represents the response from app generation
type AppGenerationResponse struct {
	AppID       string                 `json:"app_id"`
	Status      string                 `json:"status"` // "pending", "processing", "completed", "failed"
	WorkspaceID string                 `json:"workspace_id"`
	Message     string                 `json:"message,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// AppStatusResponse represents the status of an app generation
type AppStatusResponse struct {
	AppID       string                 `json:"app_id"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress"` // 0.0 to 1.0
	CurrentStep string                 `json:"current_step,omitempty"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// OrchestrateRequest represents a request to run the full 21-agent orchestration
type OrchestrateRequest struct {
	UserID      uuid.UUID              `json:"user_id"`
	Input       string                 `json:"input"`
	Phase       string                 `json:"phase,omitempty"` // "analysis", "strategy", "development", etc.
	Context     map[string]interface{} `json:"context,omitempty"`
	WorkspaceID uuid.UUID              `json:"workspace_id,omitempty"`
}

// OrchestrateResponse represents the response from orchestration
type OrchestrateResponse struct {
	Success       bool                   `json:"success"`
	Output        string                 `json:"output"`
	AgentsUsed    []string               `json:"agents_used,omitempty"`
	ExecutionTime int64                  `json:"execution_ms"`
	Data          map[string]interface{} `json:"data,omitempty"`
	NextStep      string                 `json:"next_step,omitempty"`
}

// Workspace represents an OSA workspace
type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkspacesResponse represents the list of workspaces
type WorkspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
	Total      int         `json:"total"`
}

// HealthResponse represents OSA health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ErrorResponse represents an error from OSA API
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
