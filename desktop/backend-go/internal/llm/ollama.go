package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// OllamaProvider implements Provider for a local or remote Ollama instance.
// It makes direct HTTP calls to the Ollama /api/chat endpoint.
type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider.
// baseURL is the Ollama API root (e.g. "http://localhost:11434").
// model is the default model name (e.g. "llama3.1:8b").
func NewOllamaProvider(baseURL, model string) (*OllamaProvider, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("ollama provider: baseURL must not be empty")
	}
	if model == "" {
		return nil, fmt.Errorf("ollama provider: model must not be empty")
	}
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// Name implements Provider.
func (p *OllamaProvider) Name() string { return "ollama" }

// ollamaChatRequest is the Ollama /api/chat request body.
type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaChatMsg `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *ollamaChatOpts `json:"options,omitempty"`
}

type ollamaChatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatOpts struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
}

// ollamaChatResponse is the Ollama /api/chat non-streaming response.
type ollamaChatResponse struct {
	Model           string        `json:"model"`
	Message         ollamaChatMsg `json:"message"`
	Done            bool          `json:"done"`
	PromptEvalCount int           `json:"prompt_eval_count,omitempty"`
	EvalCount       int           `json:"eval_count,omitempty"`
}

// Chat implements Provider.
func (p *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	start := time.Now()

	msgs := make([]ollamaChatMsg, 0, len(req.Messages))
	// Prepend system message if provided.
	if req.System != "" {
		msgs = append(msgs, ollamaChatMsg{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, ollamaChatMsg{Role: m.Role, Content: m.Content})
	}

	ollamaReq := ollamaChatRequest{
		Model:    p.model,
		Messages: msgs,
		Stream:   false,
	}
	if req.Temperature > 0 || req.MaxTokens > 0 || req.TopP > 0 {
		ollamaReq.Options = &ollamaChatOpts{
			Temperature: req.Temperature,
			NumPredict:  req.MaxTokens,
			TopP:        req.TopP,
		}
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("ollama provider chat: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("ollama provider chat: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("ollama provider chat: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return ChatResponse{}, fmt.Errorf("ollama provider chat: API error %s: %s", resp.Status, string(respBody))
	}

	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return ChatResponse{}, fmt.Errorf("ollama provider chat: decode response: %w", err)
	}

	return ChatResponse{
		Content:    ollamaResp.Message.Content,
		TokensUsed: ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		Latency:    time.Since(start),
		Provider:   "ollama",
		Model:      ollamaResp.Model,
	}, nil
}

// HealthCheck implements Provider.
func (p *OllamaProvider) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("ollama provider health check: create request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ollama provider health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama provider health check: unexpected status %s", resp.Status)
	}

	slog.DebugContext(ctx, "ollama health check passed", slog.String("base_url", p.baseURL))
	return nil
}
