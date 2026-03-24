package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// MCPTool struct tests
// ---------------------------------------------------------------------------

func TestMCPTool_MarshalJSON(t *testing.T) {
	tool := MCPTool{
		Name:        "search_conversations",
		Description: "Search through past conversations",
		Source:      "builtin",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
			},
			"required": []string{"query"},
		},
	}

	data, err := json.Marshal(tool)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "search_conversations", decoded["name"])
	assert.Equal(t, "builtin", decoded["source"])
}

func TestMCPTool_WithParameters(t *testing.T) {
	tool := MCPTool{
		Name:        "create_artifact",
		Description: "Create an artifact",
		Source:      "builtin",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title":   map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string"},
				"type": map[string]interface{}{
					"type": "string",
					"enum": []string{"code", "document", "markdown"},
				},
			},
			"required": []string{"title", "type", "content"},
		},
	}

	data, err := json.Marshal(tool)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	params, ok := decoded["parameters"].(map[string]interface{})
	require.True(t, ok)
	props, ok := params["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, props, "title")
	assert.Contains(t, props, "content")
	assert.Contains(t, props, "type")
}

func TestMCPTool_NilParameters(t *testing.T) {
	tool := MCPTool{
		Name:        "minimal_tool",
		Description: "A tool with no parameters",
		Source:      "external",
	}

	data, err := json.Marshal(tool)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Nil(t, decoded["parameters"])
}

// ---------------------------------------------------------------------------
// ToolResponse struct tests
// ---------------------------------------------------------------------------

func TestToolResponse_Success(t *testing.T) {
	resp := ToolResponse{
		Success: true,
		Result: map[string]interface{}{
			"found": true,
			"count": 42,
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, true, decoded["success"])
	assert.NotNil(t, decoded["result"])
}

func TestToolResponse_Error(t *testing.T) {
	resp := ToolResponse{
		Success: false,
		Error:   "tool not found: search_conversations",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, false, decoded["success"])
	assert.Equal(t, "tool not found: search_conversations", decoded["error"])
}

// ---------------------------------------------------------------------------
// MCPClientTool struct tests
// ---------------------------------------------------------------------------

func TestMCPClientTool_Marshal(t *testing.T) {
	tool := MCPClientTool{
		Name:        "github_create_issue",
		Description: "Create a GitHub issue",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{"type": "string"},
				"body":  map[string]interface{}{"type": "string"},
			},
		},
	}

	data, err := json.Marshal(tool)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "github_create_issue", decoded["name"])
	assert.NotNil(t, decoded["inputSchema"])
}

func TestMCPClientTool_EmptySchema(t *testing.T) {
	tool := MCPClientTool{
		Name:        "ping",
		Description: "Ping the server",
	}

	data, err := json.Marshal(tool)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Nil(t, decoded["inputSchema"])
}

// ---------------------------------------------------------------------------
// ValidateMCPServerURL tests
// ---------------------------------------------------------------------------

func TestValidateMCPServerURL_InvalidFormat(t *testing.T) {
	err := ValidateMCPServerURL("://not-a-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL format")
}

func TestValidateMCPServerURL_NoScheme(t *testing.T) {
	err := ValidateMCPServerURL("example.com/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only http and https protocols")
}

func TestValidateMCPServerURL_FTP(t *testing.T) {
	err := ValidateMCPServerURL("ftp://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only http and https protocols")
}

func TestValidateMCPServerURL_NoHostname(t *testing.T) {
	err := ValidateMCPServerURL("https://")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hostname")
}

func TestValidateMCPServerURL_AWSMetadata(t *testing.T) {
	err := ValidateMCPServerURL("http://169.254.169.254/latest/meta-data")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloud metadata endpoints")
}

func TestValidateMCPServerURL_GoogleMetadata(t *testing.T) {
	err := ValidateMCPServerURL("http://metadata.google.internal/computeMetadata/v1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloud metadata endpoints")
}
