package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/models"
	"github.com/rhl/businessos-backend/internal/services"
)

// newMockA2AAgentServer creates an httptest.Server acting as a real A2A agent.
// The DiscoverAgent call does GET <agentURL> (root), so the root handler must
// return the AgentCard JSON. The SSRF bypass is handled by NewA2AClientForTest().
func newMockA2AAgentServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Root: agent card (DiscoverAgent does GET <agentURL>)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"name":         "mock-agent",
			"version":      "1.0.0",
			"description":  "Mock A2A agent for integration tests",
			"url":          "http://mock/tasks",
			"capabilities": []string{"chat"},
		})
	})

	// Task creation: POST /tasks
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":     "mock-task-001",
			"status": "completed",
			"input":  map[string]any{},
			"output": map[string]any{"response": "ok from mock"},
		})
	})

	// Tools list: GET /tools
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"tools": []map[string]any{
				{"name": "echo", "description": "Echo tool"},
			},
		})
	})

	return httptest.NewServer(mux)
}

// mockIntegrationAuditSvc satisfies the A2ARoutesHandler audit interface.
type mockIntegrationAuditSvc struct{}

func (m *mockIntegrationAuditSvc) LogA2ACall(agent, action, resourceType, resourceID string, snScore float64) (*models.AuditEntry, error) {
	return &models.AuditEntry{
		ID:             "audit-integ-001",
		Timestamp:      time.Now().UTC(),
		Agent:          agent,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		SNScore:        snScore,
		GovernanceTier: models.DetermineGovernanceTier(snScore).Tier,
		Result:         "success",
	}, nil
}

func (m *mockIntegrationAuditSvc) QueryAuditTrail(resourceType, resourceID string) ([]*models.AuditEntry, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// A2AHandler integration tests (round-trip through httptest mock agent)
// ---------------------------------------------------------------------------

func TestIntegration_DiscoverAgent_RoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := newMockA2AAgentServer(t)
	defer mock.Close()

	// NewA2AClientForTest bypasses SSRF protection so loopback httptest URLs work.
	client := services.NewA2AClientForTest()
	handler := NewA2AHandler(client)

	r := gin.New()
	r.POST("/discover", handler.DiscoverAgent)

	body, _ := json.Marshal(map[string]string{"agent_url": mock.URL})
	req := httptest.NewRequest(http.MethodPost, "/discover", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "discover must return 200; got body: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	agentCard, ok := resp["agent_card"].(map[string]any)
	require.True(t, ok, "response must have agent_card; got: %s", w.Body.String())
	assert.Equal(t, "mock-agent", agentCard["name"])
	assert.Equal(t, "1.0.0", agentCard["version"])
}

func TestIntegration_CallAgent_RoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := newMockA2AAgentServer(t)
	defer mock.Close()

	client := services.NewA2AClientForTest()
	handler := NewA2AHandler(client)

	r := gin.New()
	r.POST("/call", handler.CallAgent)

	body, _ := json.Marshal(map[string]string{
		"agent_url": mock.URL,
		"message":   "hello from integration test",
	})
	req := httptest.NewRequest(http.MethodPost, "/call", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "task", "response must contain task field")

	task, ok := resp["task"].(map[string]any)
	require.True(t, ok, "task must be an object")
	assert.Equal(t, "mock-task-001", task["id"])
	assert.Equal(t, "completed", task["status"])
}

func TestIntegration_GetAgentTools_RoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := newMockA2AAgentServer(t)
	defer mock.Close()

	client := services.NewA2AClientForTest()
	handler := NewA2AHandler(client)

	r := gin.New()
	r.GET("/tools", handler.GetAgentTools)

	req := httptest.NewRequest(http.MethodGet, "/tools?agent_url="+mock.URL, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	tools, ok := resp["tools"].([]any)
	require.True(t, ok, "response must have tools array; got: %s", w.Body.String())
	assert.Len(t, tools, 1)

	firstTool, ok := tools[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "echo", firstTool["name"])
}

func TestIntegration_DiscoverThenCall_CachesConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := newMockA2AAgentServer(t)
	defer mock.Close()

	// A single client shared across discover + call, verifying cache behavior.
	client := services.NewA2AClientForTest()
	handler := NewA2AHandler(client)

	r := gin.New()
	r.POST("/discover", handler.DiscoverAgent)
	r.POST("/call", handler.CallAgent)
	r.GET("/agents", handler.ListConnectedAgents)

	// 1. Discover the agent
	discoverBody, _ := json.Marshal(map[string]string{"agent_url": mock.URL})
	discoverReq := httptest.NewRequest(http.MethodPost, "/discover", bytes.NewReader(discoverBody))
	discoverReq.Header.Set("Content-Type", "application/json")
	discoverW := httptest.NewRecorder()
	r.ServeHTTP(discoverW, discoverReq)
	require.Equal(t, http.StatusOK, discoverW.Code)

	// 2. List agents — should now have one entry
	listReq := httptest.NewRequest(http.MethodGet, "/agents", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	require.Equal(t, http.StatusOK, listW.Code)

	var listResp map[string]any
	require.NoError(t, json.Unmarshal(listW.Body.Bytes(), &listResp))
	agents, ok := listResp["agents"].([]any)
	require.True(t, ok)
	assert.Len(t, agents, 1, "one agent should be cached after discovery")

	// 3. Call the agent — should succeed using cached URL
	callBody, _ := json.Marshal(map[string]string{
		"agent_url": mock.URL,
		"message":   "follow-up call",
	})
	callReq := httptest.NewRequest(http.MethodPost, "/call", bytes.NewReader(callBody))
	callReq.Header.Set("Content-Type", "application/json")
	callW := httptest.NewRecorder()
	r.ServeHTTP(callW, callReq)
	require.Equal(t, http.StatusOK, callW.Code)
}

// ---------------------------------------------------------------------------
// A2ARoutesHandler CRM integration tests (shared-secret auth + audit trail)
// ---------------------------------------------------------------------------

func TestIntegration_CRMCreateDeal_ValidSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("BOS_A2A_SHARED_SECRET", "test-integ-secret-xyz")

	auditSvc := &mockIntegrationAuditSvc{}
	handler := NewA2ARoutesHandler(auditSvc)

	r := gin.New()
	r.POST("/crm/deals", handler.CreateDeal)

	body, _ := json.Marshal(map[string]any{"name": "Integration Deal", "value": 1500.0})
	req := httptest.NewRequest(http.MethodPost, "/crm/deals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shared-Secret", "test-integ-secret-xyz")
	req.Header.Set("X-Agent-ID", "osa-agent")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	deal, ok := resp["deal"].(map[string]any)
	require.True(t, ok, "response must contain deal object")
	assert.Equal(t, "Integration Deal", deal["name"])
	assert.NotEmpty(t, deal["id"])
	assert.Equal(t, float64(1500), deal["value"])

	auditEntry, ok := resp["audit_entry"].(map[string]any)
	require.True(t, ok, "response must contain audit_entry")
	assert.Equal(t, "audit-integ-001", auditEntry["id"])
}

func TestIntegration_CRMCreateDeal_WrongSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("BOS_A2A_SHARED_SECRET", "correct-secret")

	auditSvc := &mockIntegrationAuditSvc{}
	handler := NewA2ARoutesHandler(auditSvc)

	r := gin.New()
	r.POST("/crm/deals", handler.CreateDeal)

	body, _ := json.Marshal(map[string]any{"name": "Bad Deal"})
	req := httptest.NewRequest(http.MethodPost, "/crm/deals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shared-Secret", "wrong-secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_CRMCreateDeal_MissingSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("BOS_A2A_SHARED_SECRET", "required-secret")

	auditSvc := &mockIntegrationAuditSvc{}
	handler := NewA2ARoutesHandler(auditSvc)

	r := gin.New()
	r.POST("/crm/deals", handler.CreateDeal)

	body, _ := json.Marshal(map[string]any{"name": "No Secret Deal"})
	req := httptest.NewRequest(http.MethodPost, "/crm/deals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No X-Shared-Secret header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_CRMCreateDeal_MissingName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("BOS_A2A_SHARED_SECRET", "test-secret")

	auditSvc := &mockIntegrationAuditSvc{}
	handler := NewA2ARoutesHandler(auditSvc)

	r := gin.New()
	r.POST("/crm/deals", handler.CreateDeal)

	// name is required, omitting it should fail binding
	body, _ := json.Marshal(map[string]any{"value": 500.0})
	req := httptest.NewRequest(http.MethodPost, "/crm/deals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shared-Secret", "test-secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
