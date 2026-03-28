package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMCPHandler_MCPHealth(t *testing.T) {
	r := gin.New()
	// MCPHandler needs a pool, but MCPHealth doesn't use it
	handler := &MCPHandler{}
	r.GET("/api/mcp/health", handler.MCPHealth)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", resp["status"])
	assert.Equal(t, "mcp", resp["service"])
}

func TestMCPHandler_MCPHealth_MethodNotAllowed(t *testing.T) {
	r := gin.New()
	handler := &MCPHandler{}
	r.GET("/api/mcp/health", handler.MCPHealth)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Gin returns 404 for unregistered methods, not 405
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMCPHandler_ExecuteMCPTool_MissingBody(t *testing.T) {
	r := gin.New()
	handler := &MCPHandler{}
	r.POST("/api/mcp/tools/execute", handler.ExecuteMCPTool)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/execute", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 401 because no user in context
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMCPHandler_ExecuteMCPTool_NoUser(t *testing.T) {
	r := gin.New()
	handler := &MCPHandler{}
	r.POST("/api/mcp/tools/execute", handler.ExecuteMCPTool)

	body, _ := json.Marshal(map[string]interface{}{
		"tool":      "search_conversations",
		"arguments": map[string]interface{}{"query": "test"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/execute", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// No user in context -> 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMCPHandler_ListMCPTools_NoUser(t *testing.T) {
	r := gin.New()
	handler := &MCPHandler{}
	r.GET("/api/mcp/tools", handler.ListMCPTools)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/tools", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// No user in context -> 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMCPHandler_ExecuteMCPTool_MissingTool(t *testing.T) {
	r := gin.New()
	handler := &MCPHandler{}
	r.POST("/api/mcp/tools/execute", handler.ExecuteMCPTool)

	// Missing "tool" field (required by binding)
	body, _ := json.Marshal(map[string]interface{}{
		"arguments": map[string]interface{}{},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/execute", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Even without user context, binding should fail first but handler checks user first
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
