package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------------------------------------------------------------------------
// validateAgentURL unit tests
// ---------------------------------------------------------------------------

func TestValidateAgentURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"empty URL", "", true},
		{"valid https", "https://agent.example.com", false},
		{"valid http", "http://agent.example.com", false},
		{"ftp rejected", "ftp://agent.example.com", true},
		{"no scheme", "agent.example.com", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"with path", "https://agent.example.com/v1", false},
		{"with port", "https://agent.example.com:8080", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAgentURL(tc.url)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// A2AHandler endpoint tests
// ---------------------------------------------------------------------------

func setupA2ARouter() (*gin.Engine, *A2AHandler) {
	client := services.NewA2AClient()
	handler := NewA2AHandler(client)
	r := gin.New()
	return r, handler
}

func TestA2AHandler_DiscoverAgent_MissingBody(t *testing.T) {
	r, handler := setupA2ARouter()
	r.POST("/api/integrations/a2a/agents/discover", handler.DiscoverAgent)

	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/agents/discover", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_DiscoverAgent_InvalidURL(t *testing.T) {
	r, handler := setupA2ARouter()
	r.POST("/api/integrations/a2a/agents/discover", handler.DiscoverAgent)

	body, _ := json.Marshal(map[string]string{"agent_url": "ftp://bad-scheme.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/agents/discover", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_DiscoverAgent_EmptyURL(t *testing.T) {
	r, handler := setupA2ARouter()
	r.POST("/api/integrations/a2a/agents/discover", handler.DiscoverAgent)

	body, _ := json.Marshal(map[string]string{"agent_url": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/agents/discover", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_CallAgent_MissingBody(t *testing.T) {
	r, handler := setupA2ARouter()
	r.POST("/api/integrations/a2a/agents/call", handler.CallAgent)

	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/agents/call", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_CallAgent_InvalidURL(t *testing.T) {
	r, handler := setupA2ARouter()
	r.POST("/api/integrations/a2a/agents/call", handler.CallAgent)

	body, _ := json.Marshal(map[string]string{
		"agent_url": "not-a-url",
		"message":   "hello",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/agents/call", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_GetAgentTools_MissingParam(t *testing.T) {
	r, handler := setupA2ARouter()
	r.GET("/api/integrations/a2a/agents/tools", handler.GetAgentTools)

	req := httptest.NewRequest(http.MethodGet, "/api/integrations/a2a/agents/tools", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_GetAgentTools_InvalidURL(t *testing.T) {
	r, handler := setupA2ARouter()
	r.GET("/api/integrations/a2a/agents/tools", handler.GetAgentTools)

	req := httptest.NewRequest(http.MethodGet, "/api/integrations/a2a/agents/tools?agent_url=ftp://bad.com", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_ExecuteAgentTool_MissingName(t *testing.T) {
	r, handler := setupA2ARouter()
	r.POST("/api/integrations/a2a/agents/tools/:name", handler.ExecuteAgentTool)

	body, _ := json.Marshal(map[string]string{"agent_url": "https://agent.example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/agents/tools/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Empty tool name triggers 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_ExecuteAgentTool_InvalidURL(t *testing.T) {
	r, handler := setupA2ARouter()
	r.POST("/api/integrations/a2a/agents/tools/:name", handler.ExecuteAgentTool)

	body, _ := json.Marshal(map[string]string{"agent_url": "bad"})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/agents/tools/search", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestA2AHandler_ListConnectedAgents_Empty(t *testing.T) {
	r, handler := setupA2ARouter()
	r.GET("/api/integrations/a2a/agents", handler.ListConnectedAgents)

	req := httptest.NewRequest(http.MethodGet, "/api/integrations/a2a/agents", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string][]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp["agents"])
}

func TestA2AHandler_DisconnectAgent_NotFound(t *testing.T) {
	r, handler := setupA2ARouter()
	r.DELETE("/api/integrations/a2a/agents", handler.DisconnectAgent)

	body, _ := json.Marshal(map[string]string{"agent_url": "https://nonexistent.example.com"})
	req := httptest.NewRequest(http.MethodDelete, "/api/integrations/a2a/agents", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestA2AHandler_DisconnectAgent_MissingBody(t *testing.T) {
	r, handler := setupA2ARouter()
	r.DELETE("/api/integrations/a2a/agents", handler.DisconnectAgent)

	req := httptest.NewRequest(http.MethodDelete, "/api/integrations/a2a/agents", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// A2A Client service-level tests (no HTTP server needed)
// ---------------------------------------------------------------------------

func TestA2AClient_NewA2AClient(t *testing.T) {
	client := services.NewA2AClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.ListConnectedAgents())
}

func TestA2AClient_DisconnectAgent_NotConnected(t *testing.T) {
	client := services.NewA2AClient()
	err := client.DisconnectAgent("https://not-connected.example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not connected")
}

func TestA2AClient_ListConnectedAgents_Empty(t *testing.T) {
	client := services.NewA2AClient()
	agents := client.ListConnectedAgents()
	assert.NotNil(t, agents)
	assert.Empty(t, agents)
}

func TestA2AClient_DiscoverAgent_InvalidURL(t *testing.T) {
	client := services.NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), "not-a-url")
	assert.Error(t, err)
}

func TestA2AClient_CallAgent_InvalidURL(t *testing.T) {
	client := services.NewA2AClient()
	_, err := client.CallAgent(context.Background(), "ftp://bad.com", "hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid agent URL")
}
