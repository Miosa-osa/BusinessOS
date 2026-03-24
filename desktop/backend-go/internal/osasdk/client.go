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
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// OrchestrateRequest is the payload for OSA orchestration.
type OrchestrateRequest struct {
	Input       string `json:"input"`
	UserID      string `json:"user_id"`
	WorkspaceID string `json:"workspace_id"`
	Phase       string `json:"phase"`
}

// OrchestrateResponse is the response from OSA orchestration.
type OrchestrateResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
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
		Timestamp: time.Now(),
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

func (c *cloudClient) Close() error {
	return nil
}
