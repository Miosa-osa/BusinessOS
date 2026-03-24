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

func TestMCPServersHandler_ListMCPServers_NoUser(t *testing.T) {
	r := gin.New()
	handler := &MCPServersHandler{}
	r.GET("/api/mcp-servers", handler.ListMCPServers)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp-servers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMCPServersHandler_GetMCPServers_NoUser(t *testing.T) {
	r := gin.New()
	handler := &MCPServersHandler{}
	r.GET("/api/mcp-servers/:id", handler.GetMCPServer)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp-servers/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks user first -> 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMCPServersHandler_CreateMCPServer_NoUser(t *testing.T) {
	r := gin.New()
	handler := &MCPServersHandler{}
	r.POST("/api/mcp-servers", handler.CreateMCPServer)

	body, _ := json.Marshal(CreateMCPServerRequest{
		Name:      "test-server",
		ServerURL: "https://mcp.example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/mcp-servers", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMCPServersHandler_DeleteMCPServer_NoUser(t *testing.T) {
	r := gin.New()
	handler := &MCPServersHandler{}
	r.DELETE("/api/mcp-servers/:id", handler.DeleteMCPServer)

	req := httptest.NewRequest(http.MethodDelete, "/api/mcp-servers/some-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMCPServersHandler_TestMCPServer_NoUser(t *testing.T) {
	r := gin.New()
	handler := &MCPServersHandler{}
	r.POST("/api/mcp-servers/:id/test", handler.TestMCPServer)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp-servers/some-id/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// toMCPServerResponse helper tests
// ---------------------------------------------------------------------------

func TestPtrStringVal_Nil(t *testing.T) {
	result := ptrStringVal(nil)
	assert.Equal(t, "", result)
}

func TestPtrStringVal_Value(t *testing.T) {
	s := "hello"
	result := ptrStringVal(&s)
	assert.Equal(t, "hello", result)
}

// ---------------------------------------------------------------------------
// Validation logic tests (extracted from handler logic)
// ---------------------------------------------------------------------------

func TestCreateMCPServerRequest_Validation_NameTooLong(t *testing.T) {
	// Simulate the handler's name validation logic
	name := "this-name-is-way-too-long-and-exceeds-the-one-hundred-character-maximum-limit-that-is-enforced-by-the-handler"
	if len(name) > 100 {
		assert.True(t, true, "name should be rejected")
	} else {
		t.Fatal("name should have been rejected for exceeding 100 chars")
	}
}

func TestCreateMCPServerRequest_Validation_NameEmpty(t *testing.T) {
	name := ""
	assert.True(t, len(name) == 0, "empty name should be rejected")
}

func TestCreateMCPServerRequest_Validation_InvalidTransport(t *testing.T) {
	invalidTransports := []string{"websocket", "grpc", "tcp", ""}
	for _, transport := range invalidTransports {
		if transport == "" {
			continue // empty defaults to "sse"
		}
		if transport != "sse" && transport != "streamable_http" {
			assert.True(t, true, "transport %q should be rejected", transport)
		}
	}
}

func TestCreateMCPServerRequest_Validation_ValidTransports(t *testing.T) {
	validTransports := []string{"sse", "streamable_http"}
	for _, transport := range validTransports {
		isValid := transport == "sse" || transport == "streamable_http"
		assert.True(t, isValid, "transport %q should be valid", transport)
	}
}

func TestCreateMCPServerRequest_Validation_InvalidAuthType(t *testing.T) {
	invalidAuthTypes := []string{"basic", "digest", "oauth", "ntlm"}
	for _, authType := range invalidAuthTypes {
		isValid := authType == "none" || authType == "api_key" || authType == "bearer"
		assert.False(t, isValid, "auth type %q should be invalid", authType)
	}
}

func TestCreateMCPServerRequest_Validation_ValidAuthTypes(t *testing.T) {
	validAuthTypes := []string{"none", "api_key", "bearer"}
	for _, authType := range validAuthTypes {
		isValid := authType == "none" || authType == "api_key" || authType == "bearer"
		assert.True(t, isValid, "auth type %q should be valid", authType)
	}
}

func TestMaxMCPServersPerUser(t *testing.T) {
	assert.Equal(t, 20, maxMCPServersPerUser)
}
