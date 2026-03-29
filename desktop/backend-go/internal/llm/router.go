package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Task defines the types of tasks for model selection scoring.
type Task string

const (
	TaskChat          Task = "chat"
	TaskCode          Task = "code"
	TaskSummarize     Task = "summarize"
	TaskExtract       Task = "extract"
	TaskReason        Task = "reason"
	TaskOrchestration Task = "orchestration"
)

// Priority controls the quality/speed/cost weighting during selection.
type Priority string

const (
	PrioritySpeed   Priority = "speed"
	PriorityQuality Priority = "quality"
	PriorityCost    Priority = "cost"
	PriorityBalance Priority = "balance"
)

// ModelSpec describes a model available in the catalog.
type ModelSpec struct {
	Name              string       `json:"name"`
	ProviderName      string       `json:"provider"`
	MaxInputTokens    int          `json:"max_input_tokens"`
	SupportsFunctions bool         `json:"supports_functions"`
	Quality           int          `json:"quality"` // 1-10
	Speed             int          `json:"speed"`   // 1-10
	Cost              int          `json:"cost"`    // 1-10 (higher = cheaper)
	Fit               map[Task]int `json:"fit"`     // task → fit score 1-10
}

// Candidate is a scored model from the selection process.
type Candidate struct {
	Spec  ModelSpec `json:"spec"`
	Score float64   `json:"score"`
	Why   string    `json:"why"`
}

// SelectOptions configures a model selection query.
type SelectOptions struct {
	Task              Task
	InputTokens       int
	NeedFunctionCalls bool
	Priority          Priority
}

// Stats tracks model performance for self-improvement scoring.
type Stats struct {
	TotalRequests int64         `json:"total_requests"`
	SuccessRate   float64       `json:"success_rate"`
	AvgLatency    time.Duration `json:"avg_latency"`
	AvgConfidence float64       `json:"avg_confidence"`
	LastImproved  time.Time     `json:"last_improved"`
}

// Router manages provider selection and priority-based fallback routing.
// It holds an ordered list of providers and a model catalog for scored selection.
type Router struct {
	providers []Provider
	catalog   []ModelSpec
	stats     map[string]*Stats
	mu        sync.RWMutex
}

// NewRouter creates a Router with the given providers in priority order.
// The first provider is highest priority and will be tried first during Chat.
func NewRouter(providers []Provider, catalog []ModelSpec) (*Router, error) {
	if len(providers) == 0 {
		return nil, errors.New("llm router: at least one provider is required")
	}
	return &Router{
		providers: providers,
		catalog:   catalog,
		stats:     make(map[string]*Stats),
	}, nil
}

// Chat routes a request through providers in priority order with fallback.
// It tries each provider sequentially; if one fails, the next is attempted.
func (r *Router) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	var lastErr error
	for _, p := range r.providers {
		resp, err := p.Chat(ctx, req)
		if err != nil {
			slog.WarnContext(ctx, "llm router: provider failed, trying next",
				slog.String("provider", p.Name()),
				slog.String("error", err.Error()),
			)
			lastErr = err
			continue
		}
		return resp, nil
	}
	return ChatResponse{}, fmt.Errorf("llm router: all providers failed, last error: %w", lastErr)
}

// Select scores catalog models and returns the top candidates for the given options.
func (r *Router) Select(ctx context.Context, opts SelectOptions) ([]Candidate, error) {
	if opts.Task == "" {
		return nil, errors.New("llm router select: task must not be empty")
	}

	wq, ws, wc := weightsForPriority(opts.Priority)
	candidates := make([]Candidate, 0, len(r.catalog))

	for _, spec := range r.catalog {
		if opts.InputTokens > 0 && opts.InputTokens > spec.MaxInputTokens {
			continue
		}
		if opts.NeedFunctionCalls && !spec.SupportsFunctions {
			continue
		}

		fit := 6.0
		if f, ok := spec.Fit[opts.Task]; ok {
			fit = float64(f)
		}

		score := 0.3*(fit/10.0) +
			wq*(float64(spec.Quality)/10.0) +
			ws*(float64(spec.Speed)/10.0) +
			wc*(float64(spec.Cost)/10.0)

		// Self-improvement bonus for proven performers.
		if stats := r.getStats(spec.Name); stats != nil && stats.SuccessRate > 0.8 {
			score *= 1.1
		}

		candidates = append(candidates, Candidate{
			Spec:  spec,
			Score: score,
			Why:   fmt.Sprintf("fit=%d q=%d s=%d c=%d", spec.Fit[opts.Task], spec.Quality, spec.Speed, spec.Cost),
		})
	}

	if len(candidates) == 0 {
		return nil, errors.New("llm router select: no compatible model found")
	}

	// Sort descending by score (simple insertion for small catalogs).
	for i := 1; i < len(candidates); i++ {
		for j := i; j > 0 && candidates[j].Score > candidates[j-1].Score; j-- {
			candidates[j], candidates[j-1] = candidates[j-1], candidates[j]
		}
	}

	if len(candidates) > 3 {
		candidates = candidates[:3]
	}

	slog.DebugContext(ctx, "llm router: model selection complete",
		slog.Int("candidates", len(candidates)),
		slog.String("task", string(opts.Task)),
	)
	return candidates, nil
}

// UpdateStats records model performance for self-improvement scoring.
func (r *Router) UpdateStats(ctx context.Context, modelName string, success bool, latency time.Duration, confidence float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	stats, ok := r.stats[modelName]
	if !ok {
		stats = &Stats{}
		r.stats[modelName] = stats
	}

	stats.TotalRequests++

	if success {
		stats.SuccessRate = (stats.SuccessRate*float64(stats.TotalRequests-1) + 1.0) / float64(stats.TotalRequests)
	} else {
		stats.SuccessRate = (stats.SuccessRate * float64(stats.TotalRequests-1)) / float64(stats.TotalRequests)
	}

	stats.AvgLatency = (stats.AvgLatency*time.Duration(stats.TotalRequests-1) + latency) / time.Duration(stats.TotalRequests)
	stats.AvgConfidence = (stats.AvgConfidence*float64(stats.TotalRequests-1) + confidence) / float64(stats.TotalRequests)

	if stats.TotalRequests%100 == 0 && stats.SuccessRate > 0.85 {
		stats.LastImproved = time.Now()
		slog.InfoContext(ctx, "llm router: model performance improved",
			slog.String("model", modelName),
			slog.Float64("success_rate", stats.SuccessRate),
			slog.Float64("avg_confidence", stats.AvgConfidence),
		)
	}
}

// getStats returns stats for a model (read-locked).
func (r *Router) getStats(modelName string) *Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stats[modelName]
}

// Providers returns the ordered provider list (for inspection/testing).
func (r *Router) Providers() []Provider {
	return r.providers
}

// weightsForPriority returns (quality, speed, cost) weights.
func weightsForPriority(p Priority) (float64, float64, float64) {
	switch p {
	case PrioritySpeed:
		return 0.3, 0.5, 0.2
	case PriorityQuality:
		return 0.55, 0.25, 0.2
	case PriorityCost:
		return 0.25, 0.25, 0.5
	default: // Balance
		return 0.4, 0.35, 0.25
	}
}
