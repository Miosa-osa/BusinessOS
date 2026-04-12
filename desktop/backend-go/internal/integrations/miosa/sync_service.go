// Package miosa implements the BusinessOS <-> MIOSA Cloud sync service.
//
// Sync is always opt-in and user-initiated (or periodic if the user enables
// auto-sync in settings). The payload is a WorkspaceManifest that contains
// only configuration objects — never raw business data rows.
//
// Architecture: the Go backend is the single authority for all MIOSA cloud
// interactions. The SvelteKit frontend never contacts api.miosa.ai directly.
package miosa

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	osasdk "github.com/Miosa-osa/sdk-go"
)

// SyncMode mirrors the OSA_MODE environment variable.
type SyncMode string

const (
	ModeLocal SyncMode = "local"
	ModeCloud SyncMode = "cloud"
)

// ConnectionStatus is returned by GET /api/miosa/status.
type ConnectionStatus struct {
	Mode      SyncMode  `json:"mode"`
	Connected bool      `json:"connected"`
	APIKeySet bool      `json:"api_key_set"`
	LastSync  time.Time `json:"last_sync,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// WorkspaceManifest is the sync payload. It contains only configuration
// objects. Raw business data (tasks, contacts, deals, conversations) is
// deliberately excluded.
type WorkspaceManifest struct {
	Version     string               `json:"version"`
	ExportedAt  time.Time            `json:"exported_at"`
	WorkspaceID string               `json:"workspace_id"`
	Settings    WorkspaceSettings    `json:"settings"`
	Agents      []AgentConfig        `json:"agents"`
	Apps        []AppDefinition      `json:"apps"`
	Templates   []TemplateDefinition `json:"templates"`
}

// WorkspaceSettings contains the workspace-level configuration (non-data).
type WorkspaceSettings struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Icon        string          `json:"icon,omitempty"`
	Theme       string          `json:"theme"`
	Features    map[string]bool `json:"features"`
	OSAModel    string          `json:"osa_model,omitempty"`
	Custom      map[string]any  `json:"custom,omitempty"`
}

// AgentConfig is an agent definition (no conversation history).
type AgentConfig struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	SystemPrompt string         `json:"system_prompt"`
	Model        string         `json:"model"`
	Temperature  float64        `json:"temperature"`
	Tools        []string       `json:"tools"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// AppDefinition is a custom app schema/layout (no data rows).
type AppDefinition struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Schema      map[string]any `json:"schema"`
	Layout      map[string]any `json:"layout"`
	Permissions map[string]any `json:"permissions,omitempty"`
}

// TemplateDefinition is a template structure (no documents created from it).
type TemplateDefinition struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Category string         `json:"category"`
	Body     string         `json:"body"`
	Vars     []string       `json:"vars,omitempty"`
	Meta     map[string]any `json:"meta,omitempty"`
}

// SyncResult is returned by POST /api/miosa/sync.
type SyncResult struct {
	Success    bool      `json:"success"`
	SyncedAt   time.Time `json:"synced_at"`
	ManifestID string    `json:"manifest_id,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// ManifestProvider is implemented by the workspace service to build a manifest
// for a given workspace. This keeps the sync service free of direct database
// dependencies.
type ManifestProvider interface {
	BuildManifest(ctx context.Context, workspaceID string) (*WorkspaceManifest, error)
}

// SyncService pushes a WorkspaceManifest to MIOSA Cloud.
// It is only active when OSA_MODE=cloud and MIOSA_API_KEY is set.
type SyncService struct {
	mode     SyncMode
	apiKey   string
	cloudURL string
	provider ManifestProvider
	logger   *slog.Logger
}

// NewSyncService constructs the service. If mode is ModeLocal or apiKey is
// empty, all cloud operations return a no-op success so callers need not
// branch on mode.
func NewSyncService(
	mode SyncMode,
	apiKey string,
	cloudURL string,
	provider ManifestProvider,
	logger *slog.Logger,
) *SyncService {
	if cloudURL == "" {
		cloudURL = "https://api.miosa.ai"
	}
	return &SyncService{
		mode:     mode,
		apiKey:   apiKey,
		cloudURL: cloudURL,
		provider: provider,
		logger:   logger,
	}
}

// Status returns the current cloud connection status without performing any
// network calls (safe to call frequently).
func (s *SyncService) Status(ctx context.Context) ConnectionStatus {
	return ConnectionStatus{
		Mode:      s.mode,
		Connected: s.mode == ModeCloud && s.apiKey != "",
		APIKeySet: s.apiKey != "",
	}
}

// Ping verifies the API key is valid by calling the MIOSA Cloud health
// endpoint. Use this when the user first enters their API key.
func (s *SyncService) Ping(ctx context.Context) error {
	if s.mode != ModeCloud || s.apiKey == "" {
		return fmt.Errorf("cloud mode is not configured (OSA_MODE must be 'cloud' and MIOSA_API_KEY must be set)")
	}

	client, err := s.newCloudClient()
	if err != nil {
		return fmt.Errorf("cannot create cloud client: %w", err)
	}
	defer client.Close()

	_, err = client.Health(ctx)
	if err != nil {
		return fmt.Errorf("MIOSA Cloud ping failed: %w", err)
	}
	return nil
}

// Sync builds the workspace manifest and pushes it to MIOSA Cloud.
// It is safe to call in local mode; the function returns a no-op success.
func (s *SyncService) Sync(ctx context.Context, workspaceID string) (*SyncResult, error) {
	if s.mode != ModeCloud || s.apiKey == "" {
		s.logger.DebugContext(ctx, "sync skipped: not in cloud mode")
		return &SyncResult{Success: true, SyncedAt: time.Now()}, nil
	}

	manifest, err := s.provider.BuildManifest(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build workspace manifest: %w", err)
	}

	payload, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	client, err := s.newCloudClient()
	if err != nil {
		return nil, fmt.Errorf("cannot create cloud client: %w", err)
	}
	defer client.Close()

	// The SDK does not expose a dedicated SyncManifest RPC.
	// We push the manifest via Orchestrate: OSA receives the JSON payload,
	// persists it, and returns a session ID that we treat as the manifest ID.
	resp, err := client.Orchestrate(ctx, osasdk.OrchestrateRequest{
		Input:       string(payload),
		UserID:      "bos-sync",
		WorkspaceID: workspaceID,
		Phase:       "manifest-sync",
	})
	if err != nil {
		return &SyncResult{
			Success: false,
			Error:   err.Error(),
		}, fmt.Errorf("MIOSA Cloud sync failed: %w", err)
	}

	s.logger.InfoContext(ctx, "workspace manifest synced to MIOSA Cloud",
		slog.String("workspace_id", workspaceID),
		slog.String("manifest_id", resp.SessionID),
	)

	return &SyncResult{
		Success:    true,
		SyncedAt:   time.Now(),
		ManifestID: resp.SessionID,
	}, nil
}

// newCloudClient builds a configured sdk-go CloudClient. A new client is
// created per call to avoid holding long-lived connections.
func (s *SyncService) newCloudClient() (osasdk.Client, error) {
	// CloudConfig has no Resilience field; BOS manages retries at a higher layer.
	return osasdk.NewCloudClient(osasdk.CloudConfig{
		APIKey:  s.apiKey,
		BaseURL: s.cloudURL,
		Timeout: 30 * time.Second,
	})
}
