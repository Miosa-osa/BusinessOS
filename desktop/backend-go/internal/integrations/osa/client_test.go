package osa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				BaseURL:      "http://localhost:8089",
				SharedSecret: "test-secret-key-min-32-bytes-long",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: &Config{
				SharedSecret: "test-secret",
				Timeout:      30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing shared secret",
			config: &Config{
				BaseURL: "http://localhost:8089",
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_HealthCheck(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		resp := HealthResponse{
			Status:    "healthy",
			Version:   "1.0.0",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	config := &Config{
		BaseURL:      server.URL,
		SharedSecret: "test-secret-key-min-32-bytes-long",
		Timeout:      5 * time.Second,
		MaxRetries:   1,
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// Test health check
	ctx := context.Background()
	health, err := client.HealthCheck(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, health)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "1.0.0", health.Version)
}

func TestClient_GenerateApp(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	appID := uuid.New().String()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/apps/generate", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.Equal(t, userID.String(), r.Header.Get("X-User-ID"))
		assert.Equal(t, workspaceID.String(), r.Header.Get("X-Workspace-ID"))

		// Decode and validate request
		var req AppGenerationRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, userID, req.UserID)
		assert.Equal(t, workspaceID, req.WorkspaceID)
		assert.Equal(t, "Test App", req.Name)

		// Send response
		resp := AppGenerationResponse{
			AppID:       appID,
			Status:      "pending",
			WorkspaceID: workspaceID.String(),
			Message:     "App generation started",
			CreatedAt:   time.Now(),
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	config := &Config{
		BaseURL:      server.URL,
		SharedSecret: "test-secret-key-min-32-bytes-long",
		Timeout:      5 * time.Second,
		MaxRetries:   1,
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// Test generate app
	ctx := context.Background()
	req := &AppGenerationRequest{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        "Test App",
		Description: "A test application",
		Type:        "full-stack",
	}

	resp, err := client.GenerateApp(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, appID, resp.AppID)
	assert.Equal(t, "pending", resp.Status)
	assert.Equal(t, workspaceID.String(), resp.WorkspaceID)
}

func TestClient_GetAppStatus(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New().String()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/apps/")
		assert.Contains(t, r.URL.Path, "/status")
		assert.Equal(t, "GET", r.Method)
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.Equal(t, userID.String(), r.Header.Get("X-User-ID"))

		// Send response
		resp := AppStatusResponse{
			AppID:       appID,
			Status:      "completed",
			Progress:    1.0,
			CurrentStep: "Done",
			UpdatedAt:   time.Now(),
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	config := &Config{
		BaseURL:      server.URL,
		SharedSecret: "test-secret-key-min-32-bytes-long",
		Timeout:      5 * time.Second,
		MaxRetries:   1,
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// Test get app status
	ctx := context.Background()
	status, err := client.GetAppStatus(ctx, appID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, appID, status.AppID)
	assert.Equal(t, "completed", status.Status)
	assert.Equal(t, 1.0, status.Progress)
}

func TestClient_Orchestrate(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orchestrate", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		// Decode and validate request
		var req OrchestrateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, userID, req.UserID)
		assert.Equal(t, "Build me a task manager", req.Input)

		// Send response
		resp := OrchestrateResponse{
			Success:       true,
			Output:        "Task manager created successfully",
			AgentsUsed:    []string{"StrategyAgent", "ArchitectAgent", "DevelopmentAgent"},
			ExecutionTime: 5000,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	config := &Config{
		BaseURL:      server.URL,
		SharedSecret: "test-secret-key-min-32-bytes-long",
		Timeout:      10 * time.Second,
		MaxRetries:   1,
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// Test orchestrate
	ctx := context.Background()
	req := &OrchestrateRequest{
		UserID:      userID,
		Input:       "Build me a task manager",
		WorkspaceID: workspaceID,
	}

	resp, err := client.Orchestrate(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Task manager created successfully", resp.Output)
	assert.Len(t, resp.AgentsUsed, 3)
	assert.Equal(t, int64(5000), resp.ExecutionTime)
}

func TestClient_GetWorkspaces(t *testing.T) {
	userID := uuid.New()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/workspaces", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		// Send response
		resp := WorkspacesResponse{
			Workspaces: []Workspace{
				{
					ID:          uuid.New().String(),
					Name:        "Workspace 1",
					Description: "First workspace",
					OwnerID:     userID.String(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          uuid.New().String(),
					Name:        "Workspace 2",
					Description: "Second workspace",
					OwnerID:     userID.String(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
			Total: 2,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	config := &Config{
		BaseURL:      server.URL,
		SharedSecret: "test-secret-key-min-32-bytes-long",
		Timeout:      5 * time.Second,
		MaxRetries:   1,
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// Test get workspaces
	ctx := context.Background()
	workspaces, err := client.GetWorkspaces(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, workspaces)
	assert.Equal(t, 2, workspaces.Total)
	assert.Len(t, workspaces.Workspaces, 2)
	assert.Equal(t, "Workspace 1", workspaces.Workspaces[0].Name)
	assert.Equal(t, "Workspace 2", workspaces.Workspaces[1].Name)
}

func TestClient_ErrorHandling(t *testing.T) {
	userID := uuid.New()

	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errResp := ErrorResponse{
			Error:   "Invalid request",
			Code:    "INVALID_REQUEST",
			Details: "Missing required field: name",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	// Create client
	config := &Config{
		BaseURL:      server.URL,
		SharedSecret: "test-secret-key-min-32-bytes-long",
		Timeout:      5 * time.Second,
		MaxRetries:   1,
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// Test error response
	ctx := context.Background()
	req := &AppGenerationRequest{
		UserID:      userID,
		WorkspaceID: uuid.New(),
		Type:        "full-stack",
		// Missing Name field
	}

	resp, err := client.GenerateApp(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid request")
}

func TestClient_HTTPErrorNoRetry(t *testing.T) {
	attemptCount := 0

	// Create test server that always returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create client with retries
	config := &Config{
		BaseURL:      server.URL,
		SharedSecret: "test-secret-key-min-32-bytes-long",
		Timeout:      5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   100 * time.Millisecond,
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// HTTP status errors should NOT trigger retries (only network errors do)
	ctx := context.Background()
	_, err = client.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Equal(t, 1, attemptCount, "Should only attempt once for HTTP errors")
}
