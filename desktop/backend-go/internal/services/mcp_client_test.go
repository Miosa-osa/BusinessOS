package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewMCPClient tests
// ---------------------------------------------------------------------------

func TestNewMCPClient_Defaults(t *testing.T) {
	client := NewMCPClient("https://mcp.example.com", "none", "", nil)
	assert.NotNil(t, client)
	assert.Equal(t, "https://mcp.example.com", client.serverURL)
	assert.Equal(t, "none", client.authType)
	assert.Empty(t, client.authToken)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestNewMCPClient_TrailingSlash(t *testing.T) {
	client := NewMCPClient("https://mcp.example.com/", "none", "", nil)
	assert.Equal(t, "https://mcp.example.com", client.serverURL)
}

func TestNewMCPClient_WithAuth(t *testing.T) {
	client := NewMCPClient("https://mcp.example.com", "api_key", "secret123", map[string]string{"X-Custom": "value"})
	assert.Equal(t, "api_key", client.authType)
	assert.Equal(t, "secret123", client.authToken)
	assert.Equal(t, "value", client.headers["X-Custom"])
}

func TestNewMCPClient_BearerAuth(t *testing.T) {
	client := NewMCPClient("https://mcp.example.com", "bearer", "tok_abc", nil)
	assert.Equal(t, "bearer", client.authType)
	assert.Equal(t, "tok_abc", client.authToken)
}

// ---------------------------------------------------------------------------
// DiscoverTools with mock server
// ---------------------------------------------------------------------------

func TestMCPClient_DiscoverTools_JSON(t *testing.T) {
	tools := []MCPClientTool{
		{Name: "tool_a", Description: "Tool A", InputSchema: map[string]interface{}{"type": "object"}},
		{Name: "tool_b", Description: "Tool B", InputSchema: map[string]interface{}{"type": "object"}},
	}

	rpcResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"tools": tools,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rpcResponse)
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", nil)
	discovered, err := client.DiscoverTools(context.Background())
	require.NoError(t, err)
	require.Len(t, discovered, 2)
	assert.Equal(t, "tool_a", discovered[0].Name)
	assert.Equal(t, "tool_b", discovered[1].Name)
}

func TestMCPClient_DiscoverTools_SSE(t *testing.T) {
	tools := []MCPClientTool{
		{Name: "sse_tool", Description: "SSE Tool", InputSchema: map[string]interface{}{"type": "object"}},
	}

	rpcResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"tools": tools,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		data, _ := json.Marshal(rpcResponse)
		w.Write([]byte("data: " + string(data) + "\n\n"))
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", nil)
	discovered, err := client.DiscoverTools(context.Background())
	require.NoError(t, err)
	require.Len(t, discovered, 1)
	assert.Equal(t, "sse_tool", discovered[0].Name)
}

func TestMCPClient_DiscoverTools_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", nil)
	_, err := client.DiscoverTools(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestMCPClient_DiscoverTools_RPCError(t *testing.T) {
	rpcResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"error": map[string]interface{}{
			"code":    -32601,
			"message": "Method not found",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rpcResponse)
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", nil)
	_, err := client.DiscoverTools(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "-32601")
}

// ---------------------------------------------------------------------------
// ExecuteTool with mock server
// ---------------------------------------------------------------------------

func TestMCPClient_ExecuteTool_Success(t *testing.T) {
	rpcResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"status":  "ok",
			"created": true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "tools/call", req["method"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rpcResponse)
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", nil)
	result, err := client.ExecuteTool(context.Background(), "create_thing", map[string]interface{}{"name": "test"})
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ok", resultMap["status"])
}

func TestMCPClient_ExecuteTool_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", nil)
	_, err := client.ExecuteTool(context.Background(), "bad_tool", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

// ---------------------------------------------------------------------------
// Auth header tests
// ---------------------------------------------------------------------------

func TestMCPClient_SetHeaders_APIKey(t *testing.T) {
	var receivedHeaders http.Header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  map[string]interface{}{"tools": []interface{}{}},
		})
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "api_key", "my-secret-key", nil)
	_, err := client.DiscoverTools(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "my-secret-key", receivedHeaders.Get("X-API-Key"))
}

func TestMCPClient_SetHeaders_Bearer(t *testing.T) {
	var receivedHeaders http.Header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  map[string]interface{}{"tools": []interface{}{}},
		})
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "bearer", "my-bearer-token", nil)
	_, err := client.DiscoverTools(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "Bearer my-bearer-token", receivedHeaders.Get("Authorization"))
}

func TestMCPClient_SetHeaders_CustomHeaders(t *testing.T) {
	var receivedHeaders http.Header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  map[string]interface{}{"tools": []interface{}{}},
		})
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", map[string]string{
		"X-Custom-Header": "custom-value",
		"X-Request-ID":    "req-123",
	})
	_, err := client.DiscoverTools(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "custom-value", receivedHeaders.Get("X-Custom-Header"))
	assert.Equal(t, "req-123", receivedHeaders.Get("X-Request-ID"))
}

func TestMCPClient_SetHeaders_BlockedHeaders(t *testing.T) {
	var receivedHeaders http.Header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  map[string]interface{}{"tools": []interface{}{}},
		})
	}))
	defer ts.Close()

	// Security-critical headers should be blocked
	client := NewMCPClient(ts.URL, "none", "", map[string]string{
		"Host":             "evil.com",
		"Content-Length":   "99999",
		"Transfer-Encoding": "chunked",
	})
	_, err := client.DiscoverTools(context.Background())
	require.NoError(t, err)

	// The blocked headers should NOT have been overridden
	assert.NotEqual(t, "evil.com", receivedHeaders.Get("Host"))
}

// ---------------------------------------------------------------------------
// HealthCheck tests
// ---------------------------------------------------------------------------

func TestMCPClient_HealthCheck_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  map[string]interface{}{"tools": []interface{}{}},
		})
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", nil)
	err := client.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestMCPClient_HealthCheck_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	client := NewMCPClient(ts.URL, "none", "", nil)
	err := client.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}
