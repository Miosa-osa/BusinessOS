package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rhl/businessos-backend/internal/config"
)

// AnthropicService handles LLM inference via Anthropic's Claude API
type AnthropicService struct {
	apiKey  string
	model   string
	client  *http.Client
	options LLMOptions
}

// AnthropicMessage represents a message in the Anthropic format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicThinking represents extended thinking configuration
type AnthropicThinking struct {
	Type         string `json:"type"`          // "enabled"
	BudgetTokens int    `json:"budget_tokens"` // Max tokens for thinking (1024-32768)
}

// AnthropicRequest represents a request to the Anthropic API
type AnthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []AnthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
	Thinking  *AnthropicThinking `json:"thinking,omitempty"` // Extended thinking support
}

// AnthropicContentBlock represents a content block in the response
type AnthropicContentBlock struct {
	Type     string `json:"type"`               // "text" or "thinking"
	Text     string `json:"text,omitempty"`     // For text blocks
	Thinking string `json:"thinking,omitempty"` // For thinking blocks
}

// AnthropicResponse represents a non-streaming response from Anthropic
type AnthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []AnthropicContentBlock `json:"content"`
	StopReason string                  `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		// Extended thinking usage
		CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	} `json:"usage"`
}

// AnthropicStreamEvent represents a streaming event from Anthropic
type AnthropicStreamEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Delta struct {
		Type     string `json:"type,omitempty"`
		Text     string `json:"text,omitempty"`
		Thinking string `json:"thinking,omitempty"` // For thinking_delta events
	} `json:"delta,omitempty"`
	ContentBlock struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		Thinking string `json:"thinking,omitempty"`
	} `json:"content_block,omitempty"`
	Message struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message,omitempty"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

// NewAnthropicService creates a new Anthropic service instance
func NewAnthropicService(cfg *config.Config, model string) *AnthropicService {
	if model == "" {
		model = cfg.AnthropicModel
	}
	if model == "" {
		model = "claude-sonnet-4-20250514" // Default to Claude Sonnet
	}

	return &AnthropicService{
		apiKey: cfg.AnthropicAPIKey,
		model:  model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		options: DefaultLLMOptions(),
	}
}

// SetOptions sets the LLM options for this service
func (s *AnthropicService) SetOptions(opts LLMOptions) {
	s.options = opts
}

// GetOptions returns the current LLM options
func (s *AnthropicService) GetOptions() LLMOptions {
	return s.options
}

// StreamChat sends a chat request and streams the response
func (s *AnthropicService) StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string) (<-chan string, <-chan error) {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		// Convert messages to Anthropic format (filter out system messages)
		anthropicMsgs := make([]AnthropicMessage, 0, len(messages))
		for _, msg := range messages {
			if msg.Role == "system" {
				// Combine with system prompt
				if systemPrompt != "" {
					systemPrompt = systemPrompt + "\n\n" + msg.Content
				} else {
					systemPrompt = msg.Content
				}
				continue
			}
			anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		maxTokens := s.options.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 8192
		}

		reqBody := AnthropicRequest{
			Model:     s.model,
			MaxTokens: maxTokens,
			System:    systemPrompt,
			Messages:  anthropicMsgs,
			Stream:    true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errs <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
		if err != nil {
			errs <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", s.apiKey)
		req.Header.Set("anthropic-version", "2023-09-01")

		resp, err := s.client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("anthropic API error: %s - %s", resp.Status, string(body))
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
				return
			}

			var event AnthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue // Skip malformed lines
			}

			// Handle different event types
			switch event.Type {
			case "content_block_delta":
				if event.Delta.Text != "" {
					select {
					case chunks <- event.Delta.Text:
					case <-ctx.Done():
						return
					}
				}
			case "message_stop":
				return
			case "error":
				errs <- fmt.Errorf("anthropic stream error")
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("error reading response: %w", err)
		}
	}()

	return chunks, errs
}

// ChatComplete sends a non-streaming chat request
func (s *AnthropicService) ChatComplete(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, error) {
	// Convert messages to Anthropic format
	anthropicMsgs := make([]AnthropicMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt = systemPrompt + "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
			continue
		}
		anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	maxTokens := s.options.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 8192
	}

	reqBody := AnthropicRequest{
		Model:     s.model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  anthropicMsgs,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-09-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic API error: %s - %s", resp.Status, string(body))
	}

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text from content blocks
	var result strings.Builder
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}

	return result.String(), nil
}

// HealthCheck checks if Anthropic API is available
func (s *AnthropicService) HealthCheck(ctx context.Context) bool {
	// Simple check - just verify API key is set
	return s.apiKey != ""
}

// GetModel returns the model name
func (s *AnthropicService) GetModel() string {
	return s.model
}

// GetProvider returns the provider name
func (s *AnthropicService) GetProvider() string {
	return "anthropic"
}

// StreamChatWithUsage streams chat and tracks token usage
func (s *AnthropicService) StreamChatWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) *StreamResult {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)
	result := &StreamResult{
		Chunks: chunks,
		Errors: errs,
	}

	go func() {
		defer close(chunks)
		defer close(errs)

		var inputTokens, outputTokens int

		// Convert messages to Anthropic format (filter out system messages)
		anthropicMsgs := make([]AnthropicMessage, 0, len(messages))
		for _, msg := range messages {
			if msg.Role == "system" {
				if systemPrompt != "" {
					systemPrompt = systemPrompt + "\n\n" + msg.Content
				} else {
					systemPrompt = msg.Content
				}
				continue
			}
			anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		maxTokens := s.options.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 8192
		}

		reqBody := AnthropicRequest{
			Model:     s.model,
			MaxTokens: maxTokens,
			System:    systemPrompt,
			Messages:  anthropicMsgs,
			Stream:    true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errs <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
		if err != nil {
			errs <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", s.apiKey)
		req.Header.Set("anthropic-version", "2023-09-01")

		resp, err := s.client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("anthropic API error: %s - %s", resp.Status, string(body))
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
				break
			}

			var event AnthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "message_start":
				inputTokens = event.Message.Usage.InputTokens
			case "content_block_delta":
				if event.Delta.Text != "" {
					select {
					case chunks <- event.Delta.Text:
					case <-ctx.Done():
						return
					}
				}
			case "message_delta":
				outputTokens = event.Usage.OutputTokens
			case "message_stop":
				// Stream complete
			case "error":
				errs <- fmt.Errorf("anthropic stream error")
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("error reading response: %w", err)
		}

		// Set final token usage
		result.SetTokenUsage(&TokenUsage{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			TotalTokens:  inputTokens + outputTokens,
			Model:        s.model,
			Provider:     "anthropic",
		})
	}()

	return result
}

// ChatCompleteWithUsage sends a non-streaming chat request and returns token usage
func (s *AnthropicService) ChatCompleteWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, *TokenUsage, error) {
	// Convert messages to Anthropic format
	anthropicMsgs := make([]AnthropicMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt = systemPrompt + "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
			continue
		}
		anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	maxTokens := s.options.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 8192
	}

	reqBody := AnthropicRequest{
		Model:     s.model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  anthropicMsgs,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-09-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("anthropic API error: %s - %s", resp.Status, string(body))
	}

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text from content blocks
	var resultText strings.Builder
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			resultText.WriteString(block.Text)
		}
	}

	usage := &TokenUsage{
		InputTokens:  anthropicResp.Usage.InputTokens,
		OutputTokens: anthropicResp.Usage.OutputTokens,
		TotalTokens:  anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		Model:        s.model,
		Provider:     "anthropic",
	}

	return resultText.String(), usage, nil
}

// SupportsExtendedThinking returns true if the current model supports extended thinking
func (s *AnthropicService) SupportsExtendedThinking() bool {
	// Extended thinking is supported on Claude 3.5 Sonnet and Claude 3 Opus and newer
	supportedModels := []string{
		"claude-sonnet-4",
		"claude-opus-4",
		"claude-3-7-sonnet",
		"claude-3-5-sonnet",
		"claude-3-opus",
	}
	for _, supported := range supportedModels {
		if strings.Contains(s.model, supported) {
			return true
		}
	}
	return false
}

// StreamChatWithThinking streams chat with extended thinking support
func (s *AnthropicService) StreamChatWithThinking(ctx context.Context, messages []ChatMessage, systemPrompt string) *ExtendedThinkingResult {
	result := &ExtendedThinkingResult{
		Chunks:         make(chan string, 100),
		ThinkingChunks: make(chan string, 100),
		Errors:         make(chan error, 1),
	}

	go func() {
		defer close(result.Chunks)
		defer close(result.ThinkingChunks)
		defer close(result.Errors)

		var inputTokens, outputTokens, thinkingTokens int
		var currentBlockType string

		// Convert messages to Anthropic format
		anthropicMsgs := make([]AnthropicMessage, 0, len(messages))
		for _, msg := range messages {
			if msg.Role == "system" {
				if systemPrompt != "" {
					systemPrompt = systemPrompt + "\n\n" + msg.Content
				} else {
					systemPrompt = msg.Content
				}
				continue
			}
			anthropicMsgs = append(anthropicMsgs, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		// Build request with extended thinking enabled
		reqBody := AnthropicRequest{
			Model:     s.model,
			MaxTokens: s.options.MaxTokens,
			System:    systemPrompt,
			Messages:  anthropicMsgs,
			Stream:    true,
		}

		// Enable extended thinking if supported and enabled in options
		if s.options.ThinkingEnabled && s.SupportsExtendedThinking() {
			budgetTokens := s.options.MaxThinkingTokens
			if budgetTokens < 1024 {
				budgetTokens = 1024
			}
			if budgetTokens > 32768 {
				budgetTokens = 32768
			}
			reqBody.Thinking = &AnthropicThinking{
				Type:         "enabled",
				BudgetTokens: budgetTokens,
			}
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			result.Errors <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
		if err != nil {
			result.Errors <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", s.apiKey)
		req.Header.Set("anthropic-version", "2023-09-01")

		resp, err := s.client.Do(req)
		if err != nil {
			result.Errors <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			result.Errors <- fmt.Errorf("anthropic API error: %s - %s", resp.Status, string(body))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		// Increase buffer size for large thinking blocks
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var event AnthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "message_start":
				inputTokens = event.Message.Usage.InputTokens

			case "content_block_start":
				// Track what type of block we're in
				currentBlockType = event.ContentBlock.Type

			case "content_block_delta":
				if currentBlockType == "thinking" {
					// This is thinking content
					if event.Delta.Thinking != "" {
						select {
						case result.ThinkingChunks <- event.Delta.Thinking:
							thinkingTokens++ // Approximate token count
						case <-ctx.Done():
							return
						}
					}
				} else if currentBlockType == "text" {
					// This is regular text content
					if event.Delta.Text != "" {
						select {
						case result.Chunks <- event.Delta.Text:
						case <-ctx.Done():
							return
						}
					}
				}

			case "content_block_stop":
				currentBlockType = ""

			case "message_delta":
				outputTokens = event.Usage.OutputTokens

			case "message_stop":
				// Stream complete

			case "error":
				result.Errors <- fmt.Errorf("anthropic stream error")
				return
			}
		}

		if err := scanner.Err(); err != nil {
			result.Errors <- fmt.Errorf("error reading response: %w", err)
		}

		// Set final token usage
		result.SetTokenUsage(&TokenUsage{
			InputTokens:    inputTokens,
			OutputTokens:   outputTokens,
			ThinkingTokens: thinkingTokens,
			TotalTokens:    inputTokens + outputTokens + thinkingTokens,
			Model:          s.model,
			Provider:       "anthropic",
		})
	}()

	return result
}
