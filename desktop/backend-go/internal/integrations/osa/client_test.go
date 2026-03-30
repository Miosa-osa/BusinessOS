package osa

import (
	"context"
	"errors"
	"testing"
	"time"

	osasdk "github.com/Miosa-osa/sdk-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noopSDK implements osasdk.Client with no-op methods.
// Embed in test-specific structs and override only the methods under test.
type noopSDK struct{}

func (n *noopSDK) Health(_ context.Context) (*osasdk.HealthResponse, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) Orchestrate(_ context.Context, _ osasdk.OrchestrateRequest) (*osasdk.OrchestrateResponse, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) GenerateApp(_ context.Context, _ osasdk.AppGenerationRequest) (*osasdk.AppGenerationResponse, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) GetAppStatus(_ context.Context, _ string) (*osasdk.AppStatusResponse, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) GenerateAppFromTemplate(_ context.Context, _ osasdk.GenerateFromTemplateRequest) (*osasdk.AppGenerationResponse, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) GetWorkspaces(_ context.Context) (*osasdk.WorkspacesResponse, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) Stream(_ context.Context, _ string) (<-chan osasdk.Event, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) LaunchSwarm(_ context.Context, _ osasdk.SwarmRequest) (*osasdk.SwarmResponse, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) ListSwarms(_ context.Context) ([]osasdk.SwarmStatus, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) GetSwarm(_ context.Context, _ string) (*osasdk.SwarmStatus, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) CancelSwarm(_ context.Context, _ string) error {
	return errors.New("not implemented")
}
func (n *noopSDK) DispatchInstruction(_ context.Context, _ string, _ osasdk.Instruction) error {
	return errors.New("not implemented")
}
func (n *noopSDK) ListTools(_ context.Context) ([]osasdk.ToolDefinition, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) ExecuteTool(_ context.Context, _ string, _ map[string]interface{}) (*osasdk.ToolResult, error) {
	return nil, errors.New("not implemented")
}
func (n *noopSDK) Close() error { return nil }

func testConfig() *Config {
	return &Config{
		BaseURL:      "http://osa.local",
		SharedSecret: "test-secret-key-min-32-bytes-long",
		Timeout:      5 * time.Second,
		MaxRetries:   1,
	}
}

// --- per-test mock SDK types ---

type mockHealthSDK struct {
	noopSDK
	resp *osasdk.HealthResponse
	err  error
}

func (m *mockHealthSDK) Health(_ context.Context) (*osasdk.HealthResponse, error) {
	return m.resp, m.err
}

type mockGenerateAppSDK struct {
	noopSDK
	fn func(osasdk.AppGenerationRequest) (*osasdk.AppGenerationResponse, error)
}

func (m *mockGenerateAppSDK) GenerateApp(_ context.Context, req osasdk.AppGenerationRequest) (*osasdk.AppGenerationResponse, error) {
	return m.fn(req)
}

type mockGetAppStatusSDK struct {
	noopSDK
	fn func(string) (*osasdk.AppStatusResponse, error)
}

func (m *mockGetAppStatusSDK) GetAppStatus(_ context.Context, appID string) (*osasdk.AppStatusResponse, error) {
	return m.fn(appID)
}

type mockOrchestrateSDK struct {
	noopSDK
	fn func(osasdk.OrchestrateRequest) (*osasdk.OrchestrateResponse, error)
}

func (m *mockOrchestrateSDK) Orchestrate(_ context.Context, req osasdk.OrchestrateRequest) (*osasdk.OrchestrateResponse, error) {
	return m.fn(req)
}

type mockGetWorkspacesSDK struct {
	noopSDK
	fn func() (*osasdk.WorkspacesResponse, error)
}

func (m *mockGetWorkspacesSDK) GetWorkspaces(_ context.Context) (*osasdk.WorkspacesResponse, error) {
	return m.fn()
}

// --- tests ---

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
	client := newClientWithSDK(testConfig(), &mockHealthSDK{
		resp: &osasdk.HealthResponse{Status: "healthy", Version: "1.0.0"},
	})

	health, err := client.HealthCheck(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, health)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "1.0.0", health.Version)
}

func TestClient_GenerateApp(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	appID := uuid.New().String()

	client := newClientWithSDK(testConfig(), &mockGenerateAppSDK{
		fn: func(req osasdk.AppGenerationRequest) (*osasdk.AppGenerationResponse, error) {
			assert.Equal(t, "Test App", req.Name)
			assert.Equal(t, userID.String(), req.UserID)
			assert.Equal(t, workspaceID.String(), req.WorkspaceID)
			return &osasdk.AppGenerationResponse{
				AppID:       appID,
				Status:      "processing",
				WorkspaceID: workspaceID.String(),
			}, nil
		},
	})

	resp, err := client.GenerateApp(context.Background(), &AppGenerationRequest{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        "Test App",
		Description: "A test application",
		Type:        "full-stack",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, appID, resp.AppID)
	assert.Equal(t, "processing", resp.Status)
	assert.Equal(t, workspaceID.String(), resp.WorkspaceID)
}

func TestClient_GetAppStatus(t *testing.T) {
	appID := uuid.New().String()
	userID := uuid.New()

	client := newClientWithSDK(testConfig(), &mockGetAppStatusSDK{
		fn: func(gotAppID string) (*osasdk.AppStatusResponse, error) {
			assert.Equal(t, appID, gotAppID)
			return &osasdk.AppStatusResponse{
				AppID:    appID,
				Status:   "completed",
				Progress: 1.0,
			}, nil
		},
	})

	status, err := client.GetAppStatus(context.Background(), appID, userID)

	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, appID, status.AppID)
	assert.Equal(t, "completed", status.Status)
	assert.Equal(t, 1.0, status.Progress)
}

func TestClient_Orchestrate(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()

	client := newClientWithSDK(testConfig(), &mockOrchestrateSDK{
		fn: func(req osasdk.OrchestrateRequest) (*osasdk.OrchestrateResponse, error) {
			assert.Equal(t, userID.String(), req.UserID)
			assert.Equal(t, "Build me a task manager", req.Input)
			return &osasdk.OrchestrateResponse{
				SessionID:   "sess-123",
				Success:     true,
				Output:      "Task manager created successfully",
				AgentsUsed:  []string{"StrategyAgent", "ArchitectAgent", "DevelopmentAgent"},
				ExecutionMS: 5000,
			}, nil
		},
	})

	resp, err := client.Orchestrate(context.Background(), &OrchestrateRequest{
		UserID:      userID,
		Input:       "Build me a task manager",
		WorkspaceID: workspaceID,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Task manager created successfully", resp.Output)
	assert.Len(t, resp.AgentsUsed, 3)
	assert.Equal(t, int64(5000), resp.ExecutionTime)
}

func TestClient_GetWorkspaces(t *testing.T) {
	userID := uuid.New()
	ws1ID := uuid.New().String()
	ws2ID := uuid.New().String()

	client := newClientWithSDK(testConfig(), &mockGetWorkspacesSDK{
		fn: func() (*osasdk.WorkspacesResponse, error) {
			return &osasdk.WorkspacesResponse{
				Total: 2,
				Workspaces: []osasdk.WorkspaceInfo{
					{ID: ws1ID, Name: "Workspace 1", OwnerID: userID.String()},
					{ID: ws2ID, Name: "Workspace 2", OwnerID: userID.String()},
				},
			}, nil
		},
	})

	workspaces, err := client.GetWorkspaces(context.Background(), userID)

	require.NoError(t, err)
	assert.NotNil(t, workspaces)
	assert.Equal(t, 2, workspaces.Total)
	assert.Len(t, workspaces.Workspaces, 2)
	assert.Equal(t, "Workspace 1", workspaces.Workspaces[0].Name)
	assert.Equal(t, "Workspace 2", workspaces.Workspaces[1].Name)
}

func TestClient_ErrorHandling(t *testing.T) {
	userID := uuid.New()

	client := newClientWithSDK(testConfig(), &mockGenerateAppSDK{
		fn: func(_ osasdk.AppGenerationRequest) (*osasdk.AppGenerationResponse, error) {
			return nil, errors.New("Missing required field: name")
		},
	})

	resp, err := client.GenerateApp(context.Background(), &AppGenerationRequest{
		UserID:      userID,
		WorkspaceID: uuid.New(),
		Type:        "full-stack",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Missing required field: name")
}

func TestClient_HTTPErrorNoRetry(t *testing.T) {
	callCount := 0
	sdk := &mockHealthSDK{
		err: func() error {
			callCount++
			return errors.New("internal server error")
		}(),
	}

	// SDK resilience is disabled; each call to the base Client hits the SDK exactly once.
	client := newClientWithSDK(testConfig(), sdk)
	_, err := client.HealthCheck(context.Background())

	assert.Error(t, err)
	assert.Equal(t, 1, callCount, "base client makes exactly one attempt per call")
}
