package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rhl/businessos-backend/internal/config"
)

// sanitizeUTF8 ensures the string contains only valid UTF-8 characters
// Invalid bytes are replaced with empty string to avoid encoding issues
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// Replace invalid UTF-8 sequences
	var result strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid byte, skip it
			i++
			continue
		}
		result.WriteRune(r)
		i += size
	}
	return result.String()
}

// GroqService handles LLM inference via Groq API
type GroqService struct {
	apiKey  string
	model   string
	client  *http.Client
	options LLMOptions
}

// GroqToolCall represents a tool call from the LLM
type GroqToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// GroqTool represents a tool definition for the LLM
type GroqTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"function"`
}

// GroqMessage represents a message in the Groq format
type GroqMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolCalls  []GroqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

// GroqRequest represents a request to the Groq API
type GroqRequest struct {
	Model       string        `json:"model"`
	Messages    []GroqMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
	Tools       []GroqTool    `json:"tools,omitempty"`
	ToolChoice  string        `json:"tool_choice,omitempty"`
}

// GroqResponse represents a non-streaming response from Groq
type GroqResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Message struct {
			Role      string         `json:"role"`
			Content   string         `json:"content"`
			ToolCalls []GroqToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// GroqStreamResponse represents a streaming chunk from Groq
type GroqStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// NewGroqService creates a new Groq service instance
func NewGroqService(cfg *config.Config, model string) *GroqService {
	if model == "" {
		model = cfg.GroqModel
	}
	if model == "" {
		model = "llama-3.3-70b-versatile" // Default Groq model
	}

	return &GroqService{
		apiKey: cfg.GroqAPIKey,
		model:  model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		options: DefaultLLMOptions(),
	}
}

// SetOptions sets the LLM options for this service
func (s *GroqService) SetOptions(opts LLMOptions) {
	s.options = opts
}

// GetOptions returns the current LLM options
func (s *GroqService) GetOptions() LLMOptions {
	return s.options
}

// StreamChat sends a chat request and streams the response
func (s *GroqService) StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string) (<-chan string, <-chan error) {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		// Convert messages to Groq format
		groqMsgs := make([]GroqMessage, 0, len(messages)+1)

		for _, msg := range messages {
			role := strings.ToLower(msg.Role)
			slog.Debug("groq message role", "original", msg.Role, "normalized", role)
			if role == "system" {
				// Combine with existing system prompt
				if systemPrompt != "" {
					systemPrompt = systemPrompt + "\n\n" + msg.Content
				} else {
					systemPrompt = msg.Content
				}
				continue
			}
			// Ensure valid role for Groq API
			if role != "user" && role != "assistant" {
				slog.Debug("groq invalid role, defaulting to user", "role", role)
				role = "user" // Default to user for unknown roles
			}
			groqMsgs = append(groqMsgs, GroqMessage{
				Role:    role,
				Content: msg.Content,
			})
		}

		// Add system message first if provided/combined
		if systemPrompt != "" {
			groqMsgs = append([]GroqMessage{{
				Role:    "system",
				Content: systemPrompt,
			}}, groqMsgs...)
		}
		slog.Debug("groq sending messages to API", "count", len(groqMsgs))

		maxTokens := s.options.MaxTokens
		if maxTokens < 1000 {
			maxTokens = 8192 // Default to 8192 if not set properly
		}
		slog.Debug("groq streaming request", "max_tokens", maxTokens)

		reqBody := GroqRequest{
			Model:     s.model,
			Messages:  groqMsgs,
			MaxTokens: maxTokens,
			Stream:    true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errs <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
		if err != nil {
			errs <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.apiKey)

		resp, err := s.client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				slog.Debug("groq stream completed")
				return
			}

			var streamResp GroqStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue // Skip malformed lines
			}

			if len(streamResp.Choices) > 0 {
				// Log finish_reason when set (indicates why stream ended)
				if streamResp.Choices[0].FinishReason != "" {
					slog.Debug("groq stream finish", "reason", streamResp.Choices[0].FinishReason)
				}

				if streamResp.Choices[0].Delta.Content != "" {
					content := sanitizeUTF8(streamResp.Choices[0].Delta.Content)
					select {
					case chunks <- content:
					case <-ctx.Done():
						return
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			slog.Error("groq scanner error", "error", err)
			errs <- fmt.Errorf("error reading response: %w", err)
		}
	}()

	return chunks, errs
}

// ChatComplete sends a non-streaming chat request
func (s *GroqService) ChatComplete(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, error) {
	// Convert messages to Groq format
	groqMsgs := make([]GroqMessage, 0, len(messages)+1)

	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt = systemPrompt + "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
			continue
		}
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if systemPrompt != "" {
		groqMsgs = append([]GroqMessage{{
			Role:    "system",
			Content: systemPrompt,
		}}, groqMsgs...)
	}

	reqBody := GroqRequest{
		Model:     s.model,
		Messages:  groqMsgs,
		MaxTokens: 8192,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return groqResp.Choices[0].Message.Content, nil
}

// HealthCheck checks if Groq API is available
func (s *GroqService) HealthCheck(ctx context.Context) bool {
	return s.apiKey != ""
}

// GetModel returns the model name
func (s *GroqService) GetModel() string {
	return s.model
}

// GetProvider returns the provider name
func (s *GroqService) GetProvider() string {
	return "groq"
}

// StreamChatWithUsage streams chat and tracks token usage
func (s *GroqService) StreamChatWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) *StreamResult {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)
	result := &StreamResult{
		Chunks: chunks,
		Errors: errs,
	}

	go func() {
		defer close(chunks)
		defer close(errs)

		groqMsgs := make([]GroqMessage, 0, len(messages)+1)
		for _, msg := range messages {
			if msg.Role == "system" {
				if systemPrompt != "" {
					systemPrompt = systemPrompt + "\n\n" + msg.Content
				} else {
					systemPrompt = msg.Content
				}
				continue
			}
			groqMsgs = append(groqMsgs, GroqMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		if systemPrompt != "" {
			groqMsgs = append([]GroqMessage{{
				Role:    "system",
				Content: systemPrompt,
			}}, groqMsgs...)
		}

		maxTokens := s.options.MaxTokens
		if maxTokens < 1000 {
			maxTokens = 8192
		}

		reqBody := GroqRequest{
			Model:     s.model,
			Messages:  groqMsgs,
			MaxTokens: maxTokens,
			Stream:    true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errs <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
		if err != nil {
			errs <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.apiKey)

		resp, err := s.client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
			return
		}

		var estimatedTokens int
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var streamResp GroqStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				content := sanitizeUTF8(streamResp.Choices[0].Delta.Content)
				estimatedTokens += len(content) / 4
				select {
				case chunks <- content:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("error reading response: %w", err)
		}

		result.SetTokenUsage(&TokenUsage{
			InputTokens:  0,
			OutputTokens: estimatedTokens,
			TotalTokens:  estimatedTokens,
			Model:        s.model,
			Provider:     "groq",
		})
	}()

	return result
}

// ChatCompleteWithUsage sends a non-streaming chat request and returns token usage
func (s *GroqService) ChatCompleteWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, *TokenUsage, error) {
	groqMsgs := make([]GroqMessage, 0, len(messages)+1)

	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt = systemPrompt + "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
			continue
		}
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if systemPrompt != "" {
		groqMsgs = append([]GroqMessage{{
			Role:    "system",
			Content: systemPrompt,
		}}, groqMsgs...)
	}

	reqBody := GroqRequest{
		Model:     s.model,
		Messages:  groqMsgs,
		MaxTokens: 8192,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from Groq")
	}

	usage := &TokenUsage{
		InputTokens:  groqResp.Usage.PromptTokens,
		OutputTokens: groqResp.Usage.CompletionTokens,
		TotalTokens:  groqResp.Usage.TotalTokens,
		Model:        s.model,
		Provider:     "groq",
	}

	return groqResp.Choices[0].Message.Content, usage, nil
}

// ToolDefinition represents a tool for the LLM
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCallResult represents a tool call from the LLM
type ToolCallResult struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatWithToolsResponse represents the response from a chat with tools
type ChatWithToolsResponse struct {
	Content   string           `json:"content"`
	ToolCalls []ToolCallResult `json:"tool_calls,omitempty"`
	Usage     *TokenUsage      `json:"usage,omitempty"`
}

// ChatWithTools sends a chat request with tool definitions and returns tool calls if any
func (s *GroqService) ChatWithTools(ctx context.Context, messages []ChatMessage, systemPrompt string, tools []ToolDefinition) (*ChatWithToolsResponse, error) {
	groqMsgs := make([]GroqMessage, 0, len(messages)+1)

	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt = systemPrompt + "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
			continue
		}
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:    strings.ToLower(msg.Role),
			Content: msg.Content,
		})
	}

	if systemPrompt != "" {
		groqMsgs = append([]GroqMessage{{
			Role:    "system",
			Content: systemPrompt,
		}}, groqMsgs...)
	}

	// Convert tool definitions to Groq format
	groqTools := make([]GroqTool, 0, len(tools))
	for _, t := range tools {
		gt := GroqTool{Type: "function"}
		gt.Function.Name = t.Name
		gt.Function.Description = t.Description
		gt.Function.Parameters = t.Parameters
		groqTools = append(groqTools, gt)
	}

	reqBody := GroqRequest{
		Model:       s.model,
		Messages:    groqMsgs,
		MaxTokens:   s.options.MaxTokens,
		Temperature: s.options.Temperature,
		Stream:      false,
		Tools:       groqTools,
		ToolChoice:  "auto",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Groq")
	}

	result := &ChatWithToolsResponse{
		Content: groqResp.Choices[0].Message.Content,
		Usage: &TokenUsage{
			InputTokens:  groqResp.Usage.PromptTokens,
			OutputTokens: groqResp.Usage.CompletionTokens,
			TotalTokens:  groqResp.Usage.TotalTokens,
			Model:        s.model,
			Provider:     "groq",
		},
	}

	// Extract tool calls if any
	for _, tc := range groqResp.Choices[0].Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCallResult{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return result, nil
}

// ContinueWithToolResults continues the conversation after tool execution
func (s *GroqService) ContinueWithToolResults(ctx context.Context, messages []ChatMessage, systemPrompt string, toolResults map[string]string) (string, error) {
	groqMsgs := make([]GroqMessage, 0, len(messages)+len(toolResults)+1)

	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt = systemPrompt + "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
			continue
		}
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:    strings.ToLower(msg.Role),
			Content: msg.Content,
		})
	}

	if systemPrompt != "" {
		groqMsgs = append([]GroqMessage{{
			Role:    "system",
			Content: systemPrompt,
		}}, groqMsgs...)
	}

	// Add tool results as tool messages
	for toolCallID, result := range toolResults {
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCallID,
		})
	}

	reqBody := GroqRequest{
		Model:       s.model,
		Messages:    groqMsgs,
		MaxTokens:   s.options.MaxTokens,
		Temperature: s.options.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return groqResp.Choices[0].Message.Content, nil
}
