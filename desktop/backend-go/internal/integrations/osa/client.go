package osa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Client represents the OSA API client
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new OSA client
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// GenerateApp triggers application generation in OSA
func (c *Client) GenerateApp(ctx context.Context, req *AppGenerationRequest) (*AppGenerationResponse, error) {
	endpoint := fmt.Sprintf("%s/api/orchestrate", c.config.BaseURL)

	// Convert to OSA orchestrate format
	orchestrateReq := map[string]interface{}{
		"description": fmt.Sprintf("%s - %s", req.Name, req.Description),
	}

	var orchestrateResp struct {
		WorkflowID uuid.UUID `json:"workflow_id"`
		Success    bool      `json:"success"`
	}

	if err := c.makeRequest(ctx, "POST", endpoint, orchestrateReq, &orchestrateResp, req.UserID, &req.WorkspaceID); err != nil {
		return nil, fmt.Errorf("failed to generate app: %w", err)
	}

	// Convert response to expected format
	resp := &AppGenerationResponse{
		AppID:       orchestrateResp.WorkflowID.String(),
		WorkspaceID: req.WorkspaceID.String(),
		Status:      "processing",
		CreatedAt:   time.Now(),
	}

	if !orchestrateResp.Success {
		resp.Status = "failed"
	}

	return resp, nil
}

// GetAppStatus retrieves the status of an app generation
func (c *Client) GetAppStatus(ctx context.Context, appID string, userID uuid.UUID) (*AppStatusResponse, error) {
	endpoint := fmt.Sprintf("%s/api/apps/%s/status", c.config.BaseURL, appID)

	var resp AppStatusResponse
	if err := c.makeRequest(ctx, "GET", endpoint, nil, &resp, userID, nil); err != nil {
		return nil, fmt.Errorf("failed to get app status: %w", err)
	}

	return &resp, nil
}

// Orchestrate triggers the full 21-agent orchestration workflow
func (c *Client) Orchestrate(ctx context.Context, req *OrchestrateRequest) (*OrchestrateResponse, error) {
	endpoint := fmt.Sprintf("%s/api/orchestrate", c.config.BaseURL)

	var resp OrchestrateResponse
	var workspaceID *uuid.UUID
	if req.WorkspaceID != uuid.Nil {
		workspaceID = &req.WorkspaceID
	}

	if err := c.makeRequest(ctx, "POST", endpoint, req, &resp, req.UserID, workspaceID); err != nil {
		return nil, fmt.Errorf("failed to orchestrate: %w", err)
	}

	return &resp, nil
}

// GetWorkspaces retrieves the list of workspaces for a user
func (c *Client) GetWorkspaces(ctx context.Context, userID uuid.UUID) (*WorkspacesResponse, error) {
	endpoint := fmt.Sprintf("%s/api/workspaces", c.config.BaseURL)

	var resp WorkspacesResponse
	if err := c.makeRequest(ctx, "GET", endpoint, nil, &resp, userID, nil); err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	return &resp, nil
}

// HealthCheck checks if OSA is healthy and reachable
func (c *Client) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	endpoint := fmt.Sprintf("%s/health", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// OSA returns plain text "OK" instead of JSON
	bodyStr := string(body)
	if bodyStr == "OK" || bodyStr == "OK\n" {
		return &HealthResponse{
			Status:  "healthy",
			Version: "1.0.0",
		}, nil
	}

	// Try to parse as JSON (for future compatibility)
	var healthResp HealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		// If not JSON and not "OK", treat as unhealthy
		return nil, fmt.Errorf("unexpected response: %s", bodyStr)
	}

	return &healthResp, nil
}

// makeRequest is a helper method to make authenticated HTTP requests to OSA
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, reqBody interface{}, respBody interface{}, userID uuid.UUID, workspaceID *uuid.UUID) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Generate and set authentication token
	token, err := GenerateAuthToken(userID, workspaceID, c.config.SharedSecret)
	if err != nil {
		return fmt.Errorf("failed to generate auth token: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Set user context headers
	req.Header.Set("X-User-ID", userID.String())
	if workspaceID != nil {
		req.Header.Set("X-Workspace-ID", workspaceID.String())
	}

	// Execute request with retry logic
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}
		lastErr = err
	}

	if resp == nil {
		return fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error responses
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf("API error (%d): %s - %s", resp.StatusCode, errResp.Error, errResp.Details)
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Decode success response
	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Close closes the client and cleans up resources
func (c *Client) Close() error {
	// Close idle connections
	c.httpClient.CloseIdleConnections()
	return nil
}
