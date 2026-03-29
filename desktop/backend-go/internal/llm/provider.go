// Package llm provides a generic multi-provider LLM interface with priority-based
// fallback routing. Cherry-picked from OSA and adapted to BOS conventions.
package llm

import (
	"context"
	"time"
)

// Provider is the generic interface that every LLM backend must implement.
// Adapters for Anthropic and Ollama live in this package; additional providers
// can be added by implementing this interface.
type Provider interface {
	// Name returns the provider identifier (e.g. "anthropic", "ollama").
	Name() string

	// Chat sends a non-streaming chat completion request and returns the response.
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)

	// HealthCheck returns nil when the provider is reachable and ready.
	HealthCheck(ctx context.Context) error
}

// ChatRequest is the provider-agnostic request envelope.
type ChatRequest struct {
	Messages    []Message              `json:"messages"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	System      string                 `json:"system,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse is the provider-agnostic response envelope.
type ChatResponse struct {
	Content    string        `json:"content"`
	TokensUsed int           `json:"tokens_used"`
	Latency    time.Duration `json:"latency"`
	Provider   string        `json:"provider"`
	Model      string        `json:"model"`
}
