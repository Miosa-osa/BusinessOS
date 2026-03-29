package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider is a test double for the Provider interface.
type mockProvider struct {
	name       string
	chatResp   ChatResponse
	chatErr    error
	healthErr  error
	chatCalled int
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Chat(_ context.Context, _ ChatRequest) (ChatResponse, error) {
	m.chatCalled++
	return m.chatResp, m.chatErr
}
func (m *mockProvider) HealthCheck(_ context.Context) error { return m.healthErr }

func TestNewRouter_RequiresProvider(t *testing.T) {
	_, err := NewRouter(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one provider")
}

func TestRouter_Chat_FirstProviderSucceeds(t *testing.T) {
	primary := &mockProvider{
		name:     "primary",
		chatResp: ChatResponse{Content: "hello from primary", Provider: "primary"},
	}
	fallback := &mockProvider{name: "fallback"}

	r, err := NewRouter([]Provider{primary, fallback}, nil)
	require.NoError(t, err)

	resp, err := r.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "hello from primary", resp.Content)
	assert.Equal(t, 1, primary.chatCalled)
	assert.Equal(t, 0, fallback.chatCalled)
}

func TestRouter_Chat_FallbackOnPrimaryFailure(t *testing.T) {
	primary := &mockProvider{
		name:    "primary",
		chatErr: errors.New("primary down"),
	}
	fallback := &mockProvider{
		name:     "fallback",
		chatResp: ChatResponse{Content: "hello from fallback", Provider: "fallback"},
	}

	r, err := NewRouter([]Provider{primary, fallback}, nil)
	require.NoError(t, err)

	resp, err := r.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "hello from fallback", resp.Content)
	assert.Equal(t, 1, primary.chatCalled)
	assert.Equal(t, 1, fallback.chatCalled)
}

func TestRouter_Chat_AllProvidersFail(t *testing.T) {
	p1 := &mockProvider{name: "p1", chatErr: errors.New("p1 fail")}
	p2 := &mockProvider{name: "p2", chatErr: errors.New("p2 fail")}

	r, err := NewRouter([]Provider{p1, p2}, nil)
	require.NoError(t, err)

	_, err = r.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "all providers failed")
	assert.Contains(t, err.Error(), "p2 fail")
}

func TestRouter_Select_EmptyTask(t *testing.T) {
	r, _ := NewRouter([]Provider{&mockProvider{name: "p"}}, nil)
	_, err := r.Select(context.Background(), SelectOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task must not be empty")
}

func TestRouter_Select_FiltersAndScores(t *testing.T) {
	catalog := []ModelSpec{
		{
			Name: "small-fast", ProviderName: "ollama", MaxInputTokens: 4096,
			SupportsFunctions: false, Quality: 6, Speed: 9, Cost: 9,
			Fit: map[Task]int{TaskChat: 8},
		},
		{
			Name: "large-quality", ProviderName: "anthropic", MaxInputTokens: 200000,
			SupportsFunctions: true, Quality: 10, Speed: 5, Cost: 3,
			Fit: map[Task]int{TaskChat: 9, TaskReason: 10},
		},
		{
			Name: "too-small", ProviderName: "ollama", MaxInputTokens: 1024,
			SupportsFunctions: false, Quality: 5, Speed: 10, Cost: 10,
			Fit: map[Task]int{TaskChat: 5},
		},
	}

	r, _ := NewRouter([]Provider{&mockProvider{name: "p"}}, catalog)

	candidates, err := r.Select(context.Background(), SelectOptions{
		Task:        TaskChat,
		InputTokens: 2048,
		Priority:    PriorityQuality,
	})
	require.NoError(t, err)
	assert.Len(t, candidates, 2) // "too-small" filtered out (1024 < 2048)
	assert.Equal(t, "large-quality", candidates[0].Spec.Name)
}

func TestRouter_Select_NeedFunctionCalls(t *testing.T) {
	catalog := []ModelSpec{
		{Name: "no-funcs", ProviderName: "a", MaxInputTokens: 100000, SupportsFunctions: false, Quality: 9, Speed: 9, Cost: 9, Fit: map[Task]int{TaskChat: 9}},
		{Name: "has-funcs", ProviderName: "b", MaxInputTokens: 100000, SupportsFunctions: true, Quality: 7, Speed: 7, Cost: 7, Fit: map[Task]int{TaskChat: 7}},
	}
	r, _ := NewRouter([]Provider{&mockProvider{name: "p"}}, catalog)

	candidates, err := r.Select(context.Background(), SelectOptions{
		Task:              TaskChat,
		NeedFunctionCalls: true,
		Priority:          PriorityBalance,
	})
	require.NoError(t, err)
	assert.Len(t, candidates, 1)
	assert.Equal(t, "has-funcs", candidates[0].Spec.Name)
}

func TestRouter_UpdateStats(t *testing.T) {
	r, _ := NewRouter([]Provider{&mockProvider{name: "p"}}, nil)

	r.UpdateStats(context.Background(), "model-a", true, 100*time.Millisecond, 0.9)
	r.UpdateStats(context.Background(), "model-a", true, 200*time.Millisecond, 0.8)
	r.UpdateStats(context.Background(), "model-a", false, 300*time.Millisecond, 0.5)

	stats := r.getStats("model-a")
	require.NotNil(t, stats)
	assert.Equal(t, int64(3), stats.TotalRequests)
	assert.InDelta(t, 0.667, stats.SuccessRate, 0.01) // 2/3
}

func TestRouter_Select_NoCatalog(t *testing.T) {
	r, _ := NewRouter([]Provider{&mockProvider{name: "p"}}, nil)
	_, err := r.Select(context.Background(), SelectOptions{Task: TaskChat})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no compatible model")
}
