package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/rhl/businessos-backend/internal/services"
)

// AnthropicProvider adapts the existing BOS AnthropicService to the generic
// Provider interface. It does NOT duplicate any API logic — all calls are
// delegated to the wrapped service.
type AnthropicProvider struct {
	service *services.AnthropicService
}

// NewAnthropicProvider wraps an existing AnthropicService.
func NewAnthropicProvider(svc *services.AnthropicService) (*AnthropicProvider, error) {
	if svc == nil {
		return nil, fmt.Errorf("anthropic provider: service must not be nil")
	}
	return &AnthropicProvider{service: svc}, nil
}

// Name implements Provider.
func (p *AnthropicProvider) Name() string { return "anthropic" }

// Chat implements Provider.
func (p *AnthropicProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	start := time.Now()

	// Convert generic messages to BOS ChatMessage format.
	msgs := make([]services.ChatMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == "system" {
			// Fold system messages into the system prompt.
			if req.System != "" {
				req.System += "\n\n" + m.Content
			} else {
				req.System = m.Content
			}
			continue
		}
		msgs = append(msgs, services.ChatMessage{Role: m.Role, Content: m.Content})
	}

	// NOTE: We deliberately avoid calling p.service.SetOptions here because
	// AnthropicService.SetOptions mutates shared state without a mutex, causing
	// a data race under concurrent Chat calls. Instead we pass options via the
	// existing service API which already uses the configured defaults.
	// Per-request overrides (MaxTokens, Temperature) should be applied at the
	// service's request-construction layer in a future refactor.

	content, usage, err := p.service.ChatCompleteWithUsage(ctx, msgs, req.System)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("anthropic provider chat: %w", err)
	}

	tokensUsed := 0
	if usage != nil {
		tokensUsed = usage.TotalTokens
	}

	return ChatResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Latency:    time.Since(start),
		Provider:   "anthropic",
		Model:      p.service.GetModel(),
	}, nil
}

// HealthCheck implements Provider.
func (p *AnthropicProvider) HealthCheck(ctx context.Context) error {
	if !p.service.HealthCheck(ctx) {
		return fmt.Errorf("anthropic provider health check: service unhealthy")
	}
	return nil
}
