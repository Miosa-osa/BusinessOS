//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// groqChatRequest mirrors the Groq API request format.
type groqChatRequest struct {
	Model       string            `json:"model"`
	Messages    []groqChatMessage `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Stream      bool              `json:"stream"`
	Tools       []groqToolDef     `json:"tools,omitempty"`
	ToolChoice  string            `json:"tool_choice,omitempty"`
}

type groqChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
}

type groqToolDef struct {
	Type     string          `json:"type"`
	Function groqToolDefFunc `json:"function"`
}

type groqToolDefFunc struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type groqChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

const groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"

// getGroqAPIKey reads the key from env or skips the test.
func getGroqAPIKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("GROQ_API_KEY")
	if key == "" {
		t.Skip("GROQ_API_KEY not set — skipping real Groq integration test")
	}
	return key
}

// getGroqModel returns the model to use for testing.
func getGroqModel() string {
	m := os.Getenv("GROQ_MODEL")
	if m == "" {
		return "openai/gpt-oss-20b"
	}
	return m
}

// doGroqChat performs a synchronous (non-streaming) chat completion.
func doGroqChat(ctx context.Context, apiKey, model string, messages []groqChatMessage, tools []groqToolDef) (*groqChatResponse, time.Duration, error) {
	reqBody := groqChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   256,
		Temperature: 0.0,
		Stream:      false,
		Tools:       tools,
	}
	if len(tools) > 0 {
		reqBody.ToolChoice = "auto"
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return nil, elapsed, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, elapsed, fmt.Errorf("groq API %s: %s", resp.Status, string(raw))
	}

	var chatResp groqChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, elapsed, fmt.Errorf("decode: %w", err)
	}
	return &chatResp, elapsed, nil
}

// ---------------------------------------------------------------------------
// Test 1: Simple chat completion — "What is 2+2?"
// ---------------------------------------------------------------------------

func TestGroqRealAPI_SimpleChatComplete(t *testing.T) {
	apiKey := getGroqAPIKey(t)
	model := getGroqModel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []groqChatMessage{
		{Role: "user", Content: "What is 2+2? Answer with just the number."},
	}

	resp, elapsed, err := doGroqChat(ctx, apiKey, model, messages, nil)
	require.NoError(t, err, "Groq API call should succeed")

	t.Logf("Groq simple chat completed in %s", elapsed)
	t.Logf("Response: %q", resp.Choices[0].Message.Content)
	t.Logf("Tokens — prompt: %d  completion: %d  total: %d",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)

	require.NotEmpty(t, resp.Choices, "should have at least one choice")
	content := resp.Choices[0].Message.Content
	assert.NotEmpty(t, content, "response content should not be empty")
	assert.Contains(t, content, "4", "response should contain the number 4")
	assert.Greater(t, resp.Usage.TotalTokens, 0, "should report token usage")
}

// ---------------------------------------------------------------------------
// Test 2: Streaming chat — collect all chunks
// ---------------------------------------------------------------------------

func TestGroqRealAPI_StreamChat(t *testing.T) {
	apiKey := getGroqAPIKey(t)
	model := getGroqModel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := groqChatRequest{
		Model: model,
		Messages: []groqChatMessage{
			{Role: "user", Content: "Count from 1 to 5, each on a new line."},
		},
		MaxTokens:   256,
		Temperature: 0.0,
		Stream:      true,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqAPIURL, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "streaming request should return 200")

	// Read SSE stream and collect chunks
	var chunks []string
	var fullContent strings.Builder
	buf := make([]byte, 4096)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					goto done
				}

				var streamResp struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}
				if json.Unmarshal([]byte(data), &streamResp) == nil && len(streamResp.Choices) > 0 {
					chunk := streamResp.Choices[0].Delta.Content
					if chunk != "" {
						chunks = append(chunks, chunk)
						fullContent.WriteString(chunk)
					}
				}
			}
		}
		if readErr != nil {
			break
		}
	}

done:
	elapsed := time.Since(start)

	t.Logf("Groq streaming completed in %s", elapsed)
	t.Logf("Received %d chunks", len(chunks))
	t.Logf("Full streamed content: %q", fullContent.String())

	assert.Greater(t, len(chunks), 1, "streaming should produce multiple chunks")
	assembled := fullContent.String()
	assert.NotEmpty(t, assembled, "assembled content should not be empty")
	assert.Contains(t, assembled, "1", "streamed content should contain number 1")
	assert.Contains(t, assembled, "5", "streamed content should contain number 5")
}

// ---------------------------------------------------------------------------
// Test 3: Tool use — calculator tool definition
// ---------------------------------------------------------------------------

func TestGroqRealAPI_ToolUse(t *testing.T) {
	apiKey := getGroqAPIKey(t)
	model := getGroqModel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tools := []groqToolDef{
		{
			Type: "function",
			Function: groqToolDefFunc{
				Name:        "calculator",
				Description: "Perform basic arithmetic. Returns the numeric result.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"expression": map[string]interface{}{
							"type":        "string",
							"description": "The arithmetic expression to evaluate, e.g. '2+2'",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
	}

	messages := []groqChatMessage{
		{Role: "system", Content: "You must use the calculator tool for any math question."},
		{Role: "user", Content: "What is 15 * 7?"},
	}

	resp, elapsed, err := doGroqChat(ctx, apiKey, model, messages, tools)
	require.NoError(t, err, "Groq tool-use call should succeed")

	t.Logf("Groq tool-use call completed in %s", elapsed)
	t.Logf("Finish reason: %s", resp.Choices[0].FinishReason)

	require.NotEmpty(t, resp.Choices, "should have at least one choice")

	choice := resp.Choices[0]

	// Groq should either return tool_calls or a direct answer.
	// Both are acceptable; the key is proving the API round-trip works.
	if len(choice.Message.ToolCalls) > 0 {
		tc := choice.Message.ToolCalls[0]
		t.Logf("Tool call: id=%s name=%s args=%s", tc.ID, tc.Function.Name, tc.Function.Arguments)

		assert.Equal(t, "calculator", tc.Function.Name, "tool name should be calculator")
		assert.NotEmpty(t, tc.Function.Arguments, "tool arguments should not be empty")
		assert.Equal(t, "tool_calls", choice.FinishReason, "finish reason should be tool_calls")

		// Verify the arguments parse as valid JSON
		var args map[string]interface{}
		err := json.Unmarshal([]byte(tc.Function.Arguments), &args)
		assert.NoError(t, err, "tool arguments should be valid JSON")
		assert.Contains(t, args, "expression", "arguments should contain 'expression'")
	} else {
		// Direct answer — still valid, log it
		t.Logf("Groq answered directly instead of using tool: %q", choice.Message.Content)
		assert.NotEmpty(t, choice.Message.Content, "direct answer should not be empty")
	}

	assert.Greater(t, resp.Usage.TotalTokens, 0, "should report token usage")
}

// ---------------------------------------------------------------------------
// Test 4: Chat with usage tracking (non-streaming with usage metrics)
// ---------------------------------------------------------------------------

func TestGroqRealAPI_ChatCompleteWithUsage(t *testing.T) {
	apiKey := getGroqAPIKey(t)
	model := getGroqModel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []groqChatMessage{
		{Role: "system", Content: "You are a concise assistant."},
		{Role: "user", Content: "What is the capital of France? Answer in one word."},
	}

	resp, elapsed, err := doGroqChat(ctx, apiKey, model, messages, nil)
	require.NoError(t, err)

	t.Logf("Groq usage-tracking call completed in %s", elapsed)
	t.Logf("Response: %q", resp.Choices[0].Message.Content)
	t.Logf("Usage — prompt: %d  completion: %d  total: %d",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)

	content := resp.Choices[0].Message.Content
	assert.Contains(t, strings.ToLower(content), "paris", "answer should contain Paris")

	// Verify token counts are positive
	assert.Greater(t, resp.Usage.PromptTokens, 0, "prompt tokens should be positive")
	assert.Greater(t, resp.Usage.CompletionTokens, 0, "completion tokens should be positive")
	assert.Equal(t, resp.Usage.PromptTokens+resp.Usage.CompletionTokens, resp.Usage.TotalTokens,
		"total should equal prompt + completion")
}

// ---------------------------------------------------------------------------
// Test 5: Error handling — context timeout
// ---------------------------------------------------------------------------

func TestGroqRealAPI_ContextTimeout(t *testing.T) {
	apiKey := getGroqAPIKey(t)
	model := getGroqModel()

	// Use an absurdly short timeout to force a context deadline exceeded
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	<-ctx.Done()

	messages := []groqChatMessage{
		{Role: "user", Content: "test"},
	}

	_, _, err := doGroqChat(ctx, apiKey, model, messages, nil)
	assert.Error(t, err, "should fail with expired context")
	t.Logf("Expected error: %v", err)
}
