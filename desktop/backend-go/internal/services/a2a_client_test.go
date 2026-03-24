package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestA2AServer() *httptest.Server {
	mux := http.NewServeMux()

	// Agent card endpoint (GET /)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		card := AgentCard{
			Name:         "test-agent",
			DisplayName:  "Test Agent",
			Description:  "A test A2A agent",
			Version:      "1.0.0",
			URL:          "http://example.com",
			Capabilities: []string{"chat", "tools"},
			Skills: []AgentSkill{
				{
					ID:          "skill-1",
					Name:        "greet",
					Description: "Greets the user",
					Tags:        []string{"hello", "greeting"},
					Examples:    []string{"Hello!", "Hi there!"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(card)
	})

	// Tasks endpoint (POST /tasks)
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		task := Task{
			ID:     "task-123",
			Input:  body,
			Status: "completed",
			Output: map[string]any{
				"response": "Hello! I received your message.",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	})

	// Tools endpoint (GET /tools)
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		result := struct {
			Tools []A2ATool `json:"tools"`
		}{
			Tools: []A2ATool{
				{
					Name:        "search",
					Description: "Search for information",
					InputSchema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{
								"type":        "string",
								"description": "Search query",
							},
						},
						"required": []string{"query"},
					},
				},
				{
					Name:        "calculate",
					Description: "Perform a calculation",
					InputSchema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"expression": map[string]any{
								"type":        "string",
								"description": "Math expression",
							},
						},
						"required": []string{"expression"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// Tool execution endpoint (POST /tools/{name})
	mux.HandleFunc("/tools/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Extract tool name from path
		toolName := r.URL.Path[len("/tools/"):]
		if toolName == "" {
			http.Error(w, "tool name required", http.StatusBadRequest)
			return
		}

		var args map[string]any
		json.NewDecoder(r.Body).Decode(&args)

		result := map[string]any{
			"tool":    toolName,
			"args":    args,
			"result":  "ok",
			"message": fmt.Sprintf("Executed %s successfully", toolName),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	return httptest.NewServer(mux)
}

// ---------------------------------------------------------------------------
// NewA2AClient
// ---------------------------------------------------------------------------

func TestNewA2AClient(t *testing.T) {
	client := NewA2AClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.agents)
	assert.Empty(t, client.ListConnectedAgents())
}

func TestNewA2AClient_DefaultTimeout(t *testing.T) {
	client := NewA2AClient()
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

// ---------------------------------------------------------------------------
// DiscoverAgent
// ---------------------------------------------------------------------------

func TestDiscoverAgent_Success(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	card, err := client.DiscoverAgent(context.Background(), server.URL)
	require.NoError(t, err)
	require.NotNil(t, card)

	assert.Equal(t, "test-agent", card.Name)
	assert.Equal(t, "Test Agent", card.DisplayName)
	assert.Equal(t, "A test A2A agent", card.Description)
	assert.Equal(t, "1.0.0", card.Version)
	assert.Contains(t, card.Capabilities, "chat")
	assert.Len(t, card.Skills, 1)
	assert.Equal(t, "greet", card.Skills[0].Name)
}

func TestDiscoverAgent_CachesConnection(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), server.URL)
	require.NoError(t, err)

	agents := client.ListConnectedAgents()
	assert.Len(t, agents, 1)
	assert.Equal(t, server.URL, agents[0].URL)
	assert.Equal(t, "test-agent", agents[0].Card.Name)
}

func TestDiscoverAgent_InvalidURL(t *testing.T) {
	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), "not-a-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid agent URL")
}

func TestDiscoverAgent_ServerUnreachable(t *testing.T) {
	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), "http://localhost:1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reach agent")
}

func TestDiscoverAgent_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not-json"))
	}))
	defer server.Close()

	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse agent card")
}

func TestDiscoverAgent_MissingName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AgentCard{
			Description: "No name provided",
		})
	}))
	defer server.Close()

	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: name")
}

func TestDiscoverAgent_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "503")
}

func TestDiscoverAgent_TrailingSlash(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	// Server URL with trailing slash should be normalized
	card, err := client.DiscoverAgent(context.Background(), server.URL+"/")
	require.NoError(t, err)
	assert.Equal(t, "test-agent", card.Name)
}

// ---------------------------------------------------------------------------
// CallAgent
// ---------------------------------------------------------------------------

func TestCallAgent_Success(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	task, err := client.CallAgent(context.Background(), server.URL, "Hello agent!")
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, "task-123", task.ID)
	assert.Equal(t, "completed", task.Status)
	assert.Equal(t, "Hello agent!", task.Input["message"])
	assert.Contains(t, task.Output["response"], "Hello!")
}

func TestCallAgent_InvalidURL(t *testing.T) {
	client := NewA2AClient()
	_, err := client.CallAgent(context.Background(), "://bad", "test")
	assert.Error(t, err)
}

func TestCallAgent_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewA2AClient()
	_, err := client.CallAgent(context.Background(), server.URL, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestCallAgent_UpdatesLastSeen(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()

	// Discover first to cache
	_, err := client.DiscoverAgent(context.Background(), server.URL)
	require.NoError(t, err)

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Call agent
	_, err = client.CallAgent(context.Background(), server.URL, "test")
	require.NoError(t, err)

	// Check last seen was updated
	agents := client.ListConnectedAgents()
	require.Len(t, agents, 1)
	assert.False(t, agents[0].LastSeen.IsZero())
}

// ---------------------------------------------------------------------------
// GetAgentTools
// ---------------------------------------------------------------------------

func TestGetAgentTools_Success(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	tools, err := client.GetAgentTools(context.Background(), server.URL)
	require.NoError(t, err)
	require.Len(t, tools, 2)

	assert.Equal(t, "search", tools[0].Name)
	assert.Equal(t, "Search for information", tools[0].Description)
	assert.NotNil(t, tools[0].InputSchema)

	assert.Equal(t, "calculate", tools[1].Name)
}

func TestGetAgentTools_InvalidURL(t *testing.T) {
	client := NewA2AClient()
	_, err := client.GetAgentTools(context.Background(), "not-a-url")
	assert.Error(t, err)
}

func TestGetAgentTools_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewA2AClient()
	_, err := client.GetAgentTools(context.Background(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// ---------------------------------------------------------------------------
// ExecuteAgentTool
// ---------------------------------------------------------------------------

func TestExecuteAgentTool_Success(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	args := map[string]any{"query": "test search"}
	result, err := client.ExecuteAgentTool(context.Background(), server.URL, "search", args)
	require.NoError(t, err)
	require.NotNil(t, result)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "search", resultMap["tool"])
	assert.Equal(t, "ok", resultMap["result"])
}

func TestExecuteAgentTool_NilArgs(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	result, err := client.ExecuteAgentTool(context.Background(), server.URL, "search", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestExecuteAgentTool_InvalidURL(t *testing.T) {
	client := NewA2AClient()
	_, err := client.ExecuteAgentTool(context.Background(), "bad", "tool", nil)
	assert.Error(t, err)
}

func TestExecuteAgentTool_ToolNameEscaped(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	// Tool name with special characters should be escaped
	result, err := client.ExecuteAgentTool(context.Background(), server.URL, "search things", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestExecuteAgentTool_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewA2AClient()
	_, err := client.ExecuteAgentTool(context.Background(), server.URL, "tool", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

// ---------------------------------------------------------------------------
// ListConnectedAgents / DisconnectAgent
// ---------------------------------------------------------------------------

func TestListConnectedAgents_Empty(t *testing.T) {
	client := NewA2AClient()
	agents := client.ListConnectedAgents()
	assert.Empty(t, agents)
}

func TestListConnectedAgents_AfterDiscovery(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), server.URL)
	require.NoError(t, err)

	agents := client.ListConnectedAgents()
	assert.Len(t, agents, 1)
	assert.Equal(t, server.URL, agents[0].URL)
}

func TestDisconnectAgent_Success(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), server.URL)
	require.NoError(t, err)

	err = client.DisconnectAgent(server.URL)
	require.NoError(t, err)

	agents := client.ListConnectedAgents()
	assert.Empty(t, agents)
}

func TestDisconnectAgent_NotConnected(t *testing.T) {
	client := NewA2AClient()
	err := client.DisconnectAgent("http://not-connected.example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not connected")
}

func TestDisconnectAgent_TrailingSlash(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()
	_, err := client.DiscoverAgent(context.Background(), server.URL)
	require.NoError(t, err)

	// Disconnect with trailing slash should still work
	err = client.DisconnectAgent(server.URL + "/")
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Concurrency
// ---------------------------------------------------------------------------

func TestConcurrentAccess(t *testing.T) {
	server := newTestA2AServer()
	defer server.Close()

	client := NewA2AClient()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Concurrent discoveries
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.DiscoverAgent(context.Background(), server.URL)
			if err != nil {
				errors <- err
			}
		}()
	}

	// Concurrent calls
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.CallAgent(context.Background(), server.URL, "concurrent test")
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent operation failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ValidateA2AAgentURL
// ---------------------------------------------------------------------------

func TestValidateA2AAgentURL_ValidHTTP(t *testing.T) {
	err := ValidateA2AAgentURL("http://example.com")
	assert.NoError(t, err)
}

func TestValidateA2AAgentURL_ValidHTTPS(t *testing.T) {
	err := ValidateA2AAgentURL("https://example.com")
	assert.NoError(t, err)
}

func TestValidateA2AAgentURL_InvalidScheme(t *testing.T) {
	err := ValidateA2AAgentURL("ftp://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only http and https")
}

func TestValidateA2AAgentURL_Empty(t *testing.T) {
	err := ValidateA2AAgentURL("")
	assert.Error(t, err)
}

func TestValidateA2AAgentURL_NoHostname(t *testing.T) {
	err := ValidateA2AAgentURL("http://")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hostname")
}

func TestValidateA2AAgentURL_Loopback(t *testing.T) {
	err := ValidateA2AAgentURL("http://127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loopback")
}

func TestValidateA2AAgentURL_LoopbackIPv6(t *testing.T) {
	err := ValidateA2AAgentURL("http://[::1]")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loopback")
}

func TestValidateA2AAgentURL_Localhost(t *testing.T) {
	err := ValidateA2AAgentURL("http://localhost")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loopback")
}

func TestValidateA2AAgentURL_PrivateNetwork(t *testing.T) {
	err := ValidateA2AAgentURL("http://192.168.1.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "private")
}

func TestValidateA2AAgentURL_LinkLocal(t *testing.T) {
	err := ValidateA2AAgentURL("http://169.254.1.1")
	assert.Error(t, err)
}

func TestValidateA2AAgentURL_CloudMetadata(t *testing.T) {
	err := ValidateA2AAgentURL("http://169.254.169.254")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloud metadata")
}

func TestValidateA2AAgentURL_GoogleMetadata(t *testing.T) {
	err := ValidateA2AAgentURL("http://metadata.google.internal")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloud metadata")
}

func TestValidateA2AAgentURL_Unparseable(t *testing.T) {
	err := ValidateA2AAgentURL("://")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL format")
}

func TestValidateA2AAgentURL_Unresolvable(t *testing.T) {
	err := ValidateA2AAgentURL("http://this-domain-definitely-does-not-exist.invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot resolve")
}
