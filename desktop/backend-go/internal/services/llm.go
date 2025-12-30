package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/rhl/businessos-backend/internal/config"
)

// LLMOptions holds configurable parameters for LLM calls
type LLMOptions struct {
	Temperature float64
	MaxTokens   int
	TopP        float64
	// Thinking/COT options
	ThinkingEnabled      bool
	ThinkingInstruction  string
	MaxThinkingTokens    int
	ReasoningTemplateID  string
}

// DefaultLLMOptions returns sensible defaults
func DefaultLLMOptions() LLMOptions {
	return LLMOptions{
		Temperature:       0.7,
		MaxTokens:         8192,
		TopP:              0.9,
		ThinkingEnabled:   false,
		MaxThinkingTokens: 4096,
	}
}

// ThinkingChunk represents a chunk of thinking/reasoning output
type ThinkingChunk struct {
	Step      int    `json:"step"`
	Content   string `json:"content"`
	Type      string `json:"type"` // analysis, planning, reflection, tool_use, reasoning, evaluation
	Completed bool   `json:"completed"`
}

// StreamResultWithThinking extends StreamResult with thinking support
type StreamResultWithThinking struct {
	*StreamResult
	ThinkingChunks <-chan ThinkingChunk
}

// TokenUsage tracks token consumption for LLM calls
type TokenUsage struct {
	InputTokens    int    `json:"input_tokens"`
	OutputTokens   int    `json:"output_tokens"`
	ThinkingTokens int    `json:"thinking_tokens,omitempty"` // Tokens used for extended thinking/COT
	TotalTokens    int    `json:"total_tokens"`
	Model          string `json:"model"`
	Provider       string `json:"provider"`
}

// AgentTrace tracks the flow of an agent request
type AgentTrace struct {
	AgentName   string       `json:"agent_name"`
	DelegatedTo string       `json:"delegated_to,omitempty"`
	TokenUsage  *TokenUsage  `json:"token_usage,omitempty"`
	SubTraces   []AgentTrace `json:"sub_traces,omitempty"`
	StartTime   int64        `json:"start_time"`
	EndTime     int64        `json:"end_time,omitempty"`
	DurationMs  int64        `json:"duration_ms,omitempty"`
}

// StreamResult wraps streaming results with token usage
type StreamResult struct {
	Chunks     <-chan string
	Errors     <-chan error
	TokenUsage *TokenUsage
	mu         sync.Mutex
}

// SetTokenUsage safely sets token usage (called when stream completes)
func (sr *StreamResult) SetTokenUsage(usage *TokenUsage) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.TokenUsage = usage
}

// GetTokenUsage safely gets token usage
func (sr *StreamResult) GetTokenUsage() *TokenUsage {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.TokenUsage
}

// ExtendedThinkingResult wraps streaming results with separate thinking channel
type ExtendedThinkingResult struct {
	Chunks         chan string        // Regular response content
	ThinkingChunks chan string        // Extended thinking content
	Errors         chan error         // Errors
	TokenUsage     *TokenUsage        // Final token usage
	mu             sync.Mutex
}

// SetTokenUsage safely sets token usage (called when stream completes)
func (etr *ExtendedThinkingResult) SetTokenUsage(usage *TokenUsage) {
	etr.mu.Lock()
	defer etr.mu.Unlock()
	etr.TokenUsage = usage
}

// GetTokenUsage safely gets token usage
func (etr *ExtendedThinkingResult) GetTokenUsage() *TokenUsage {
	etr.mu.Lock()
	defer etr.mu.Unlock()
	return etr.TokenUsage
}

// LLMService interface for language model services
type LLMService interface {
	StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string) (<-chan string, <-chan error)
	StreamChatWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) *StreamResult
	ChatComplete(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, error)
	ChatCompleteWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, *TokenUsage, error)
	HealthCheck(ctx context.Context) bool
	GetModel() string
	GetProvider() string
	SetOptions(opts LLMOptions)
	GetOptions() LLMOptions
}

// ExtendedThinkingService interface for providers that support native extended thinking
type ExtendedThinkingService interface {
	LLMService
	// StreamChatWithThinking streams chat with extended thinking support
	// Returns separate channels for thinking content and response content
	StreamChatWithThinking(ctx context.Context, messages []ChatMessage, systemPrompt string) *ExtendedThinkingResult
	// SupportsExtendedThinking returns true if the current model supports extended thinking
	SupportsExtendedThinking() bool
}

// NewLLMService creates the appropriate LLM service based on configuration
func NewLLMService(cfg *config.Config, model string) LLMService {
	provider := cfg.GetActiveProvider()
	fmt.Printf("[LLM] Creating service for provider=%q, model=%q\n", provider, model)
	switch provider {
	case "ollama_cloud":
		return NewOllamaCloudService(cfg, model)
	case "anthropic":
		return NewAnthropicService(cfg, model)
	case "groq":
		return NewGroqService(cfg, model)
	default:
		// Default to local Ollama
		return NewOllamaService(cfg, model)
	}
}
