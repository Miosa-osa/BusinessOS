package osa

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	osasdk "github.com/Miosa-osa/sdk-go"
)

// Client is a thin wrapper around the SDK that preserves the BOS method signatures.
// uuid.UUID parameters are converted to strings before delegating to the SDK, and
// SDK response types are converted back to the BOS types defined in types.go.
type Client struct {
	config *Config
	sdk    osasdk.Client
}

// NewClient creates a new OSA client backed by the SDK.
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	sdkClient, err := osasdk.NewLocalClient(osasdk.LocalConfig{
		BaseURL:      config.BaseURL,
		SharedSecret: config.SharedSecret,
		Timeout:      config.Timeout,
		// BOS has its own resilience layer (ResilientClient), so disable the
		// SDK's built-in circuit breaker and retry to avoid double-wrapping.
		Resilience: &osasdk.ResilienceConfig{
			Enabled: false,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SDK client: %w", err)
	}

	return &Client{
		config: config,
		sdk:    sdkClient,
	}, nil
}

// GenerateApp triggers application generation in OSA.
func (c *Client) GenerateApp(ctx context.Context, req *AppGenerationRequest) (*AppGenerationResponse, error) {
	sdkResp, err := c.sdk.GenerateApp(ctx, osasdk.AppGenerationRequest{
		UserID:      req.UserID.String(),
		WorkspaceID: req.WorkspaceID.String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Parameters:  req.Parameters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate app: %w", adaptSDKError(err))
	}

	createdAt, _ := time.Parse(time.RFC3339, sdkResp.CreatedAt)

	return &AppGenerationResponse{
		AppID:       sdkResp.AppID,
		Status:      sdkResp.Status,
		WorkspaceID: sdkResp.WorkspaceID,
		Message:     sdkResp.Message,
		Data:        sdkResp.Data,
		CreatedAt:   createdAt,
	}, nil
}

// GetAppStatus retrieves the status of an app generation.
func (c *Client) GetAppStatus(ctx context.Context, appID string, userID uuid.UUID) (*AppStatusResponse, error) {
	// The SDK does not accept a userID parameter; the authenticated user is
	// implicit in the JWT token generated from the shared secret.
	sdkResp, err := c.sdk.GetAppStatus(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app status: %w", adaptSDKError(err))
	}

	updatedAt, _ := time.Parse(time.RFC3339, sdkResp.UpdatedAt)

	// Convert map output to JSON string for BOS type compatibility.
	outputStr := ""
	if sdkResp.Output != nil {
		if b, err := json.Marshal(sdkResp.Output); err == nil {
			outputStr = string(b)
		}
	}

	return &AppStatusResponse{
		AppID:       sdkResp.AppID,
		Status:      sdkResp.Status,
		Progress:    sdkResp.Progress,
		CurrentStep: sdkResp.CurrentStep,
		Output:      outputStr,
		Error:       sdkResp.Error,
		Metadata:    sdkResp.Metadata,
		UpdatedAt:   updatedAt,
	}, nil
}

// Orchestrate triggers the full agent orchestration workflow.
func (c *Client) Orchestrate(ctx context.Context, req *OrchestrateRequest) (*OrchestrateResponse, error) {
	sdkReq := osasdk.OrchestrateRequest{
		UserID:    req.UserID.String(),
		Input:     req.Input,
		SessionID: req.SessionID,
		Phase:     req.Phase,
		Context:   req.Context,
	}
	if req.WorkspaceID != uuid.Nil {
		sdkReq.WorkspaceID = req.WorkspaceID.String()
	}

	sdkResp, err := c.sdk.Orchestrate(ctx, sdkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to orchestrate: %w", adaptSDKError(err))
	}

	return &OrchestrateResponse{
		Success:       sdkResp.Success,
		Output:        sdkResp.Output,
		AgentsUsed:    sdkResp.AgentsUsed,
		ExecutionTime: sdkResp.ExecutionMS,
		Data:          sdkResp.Metadata,
		NextStep:      sdkResp.NextStep,
	}, nil
}

// GetWorkspaces retrieves the list of workspaces for a user.
// The userID is encoded in the JWT and used by OSA to scope the results;
// the SDK does not take it as an explicit parameter.
func (c *Client) GetWorkspaces(ctx context.Context, userID uuid.UUID) (*WorkspacesResponse, error) {
	sdkResp, err := c.sdk.GetWorkspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", adaptSDKError(err))
	}

	workspaces := make([]Workspace, len(sdkResp.Workspaces))
	for i, w := range sdkResp.Workspaces {
		createdAt, _ := time.Parse(time.RFC3339, w.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, w.UpdatedAt)
		workspaces[i] = Workspace{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			OwnerID:     w.OwnerID,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}
	}

	return &WorkspacesResponse{
		Workspaces: workspaces,
		Total:      sdkResp.Total,
	}, nil
}

// HealthCheck checks if OSA is healthy and reachable.
func (c *Client) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	sdkResp, err := c.sdk.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", adaptSDKError(err))
	}

	return &HealthResponse{
		Status:   sdkResp.Status,
		Version:  sdkResp.Version,
		Provider: sdkResp.Provider,
	}, nil
}

// GenerateAppFromTemplate generates an app using a predefined template.
func (c *Client) GenerateAppFromTemplate(
	ctx context.Context,
	templateName string,
	variables map[string]interface{},
	userID uuid.UUID,
	workspaceID *uuid.UUID,
) (*AppGenerationResponse, error) {
	sdkReq := osasdk.GenerateFromTemplateRequest{
		TemplateName: templateName,
		Variables:    variables,
		UserID:       userID.String(),
	}
	if workspaceID != nil {
		sdkReq.WorkspaceID = workspaceID.String()
	}

	sdkResp, err := c.sdk.GenerateAppFromTemplate(ctx, sdkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app from template: %w", adaptSDKError(err))
	}

	createdAt, _ := time.Parse(time.RFC3339, sdkResp.CreatedAt)

	return &AppGenerationResponse{
		AppID:       sdkResp.AppID,
		Status:      sdkResp.Status,
		WorkspaceID: sdkResp.WorkspaceID,
		Message:     sdkResp.Message,
		Data:        sdkResp.Data,
		CreatedAt:   createdAt,
	}, nil
}

// Stream connects to OSA's real-time event stream for a session.
// Returns a channel of SDK events that maps to SSE events from OSA.
func (c *Client) Stream(ctx context.Context, sessionID string) (<-chan osasdk.Event, error) {
	return c.sdk.Stream(ctx, sessionID)
}

// LaunchSwarm starts a multi-agent swarm with the given pattern and task.
func (c *Client) LaunchSwarm(ctx context.Context, req osasdk.SwarmRequest) (*osasdk.SwarmResponse, error) {
	sdkResp, err := c.sdk.LaunchSwarm(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to launch swarm: %w", adaptSDKError(err))
	}
	return sdkResp, nil
}

// ListSwarms returns all swarms for the current session.
func (c *Client) ListSwarms(ctx context.Context) ([]osasdk.SwarmStatus, error) {
	swarms, err := c.sdk.ListSwarms(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list swarms: %w", adaptSDKError(err))
	}
	return swarms, nil
}

// GetSwarm retrieves the status of a specific swarm by ID.
func (c *Client) GetSwarm(ctx context.Context, swarmID string) (*osasdk.SwarmStatus, error) {
	sdkResp, err := c.sdk.GetSwarm(ctx, swarmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swarm: %w", adaptSDKError(err))
	}
	return sdkResp, nil
}

// CancelSwarm cancels a running swarm.
func (c *Client) CancelSwarm(ctx context.Context, swarmID string) error {
	if err := c.sdk.CancelSwarm(ctx, swarmID); err != nil {
		return fmt.Errorf("failed to cancel swarm: %w", adaptSDKError(err))
	}
	return nil
}

// DispatchInstruction sends an instruction directly to a specific fleet agent.
func (c *Client) DispatchInstruction(ctx context.Context, agentID string, instruction osasdk.Instruction) error {
	if err := c.sdk.DispatchInstruction(ctx, agentID, instruction); err != nil {
		return fmt.Errorf("failed to dispatch instruction: %w", adaptSDKError(err))
	}
	return nil
}

// ListTools returns all available tools (built-in + TOOL.md + MCP).
func (c *Client) ListTools(ctx context.Context) ([]osasdk.ToolDefinition, error) {
	tools, err := c.sdk.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", adaptSDKError(err))
	}
	return tools, nil
}

// ExecuteTool runs a specific tool by name with the given parameters.
func (c *Client) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*osasdk.ToolResult, error) {
	sdkResp, err := c.sdk.ExecuteTool(ctx, toolName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool: %w", adaptSDKError(err))
	}
	return sdkResp, nil
}

// Close releases resources held by the client.
func (c *Client) Close() error {
	return c.sdk.Close()
}

// adaptSDKError converts an SDK error into a form that BOS's IsRetryableError
// can inspect. The BOS resilience layer pattern-matches on strings like
// "status 503"; the SDK's APIError.Error() returns only the Details field
// (which may be empty for bare HTTP errors), so we enrich the message here.
func adaptSDKError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *osasdk.APIError
	if asAPIError(err, &apiErr) && apiErr.StatusCode > 0 {
		details := apiErr.Details
		if details == "" {
			details = apiErr.ErrorCode
		}
		return fmt.Errorf("status %d: %s", apiErr.StatusCode, details)
	}
	return err
}

// asAPIError is a helper that unwraps the SDK APIError type from an error.
func asAPIError(err error, target **osasdk.APIError) bool {
	if e, ok := err.(*osasdk.APIError); ok {
		*target = e
		return true
	}
	return false
}
