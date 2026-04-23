package optimal

import (
	"context"
	"log/slog"
	"time"
)

// SearchResult is a single hit returned by the FTS5 search index.
type SearchResult struct {
	Path     string  `json:"path"`
	Abstract string  `json:"abstract"`
	Score    float64 `json:"score"`
}

// EngineConfig holds runtime parameters for the OptimalOS engine.
// All operations now use pure Go — no Elixir subprocess required.
type EngineConfig struct {
	// MixProjectPath is kept for backward compatibility. It is no longer used
	// for any operations; the pure Go implementations take precedence.
	MixProjectPath string
	// DBPath is the absolute path to the SQLite index (.system/index.db).
	// When empty, search and ingest indexing are silently skipped.
	DBPath string
	// NodesRoot is the absolute path to the nodes/ directory inside OptimalOS.
	NodesRoot string
	// OSRoot is the absolute path to the OptimalOS repo root.
	OSRoot string
	// Timeout is kept for future use (e.g. context propagation in slow queries).
	Timeout time.Duration
	// Topology holds the routing rules and node/person config loaded from
	// topology.yaml at startup. When nil, routing falls back to defaultRoutingRules.
	Topology *TopologyConfig
}

// NewEngineConfig returns an EngineConfig with a 10-second default timeout.
// mixPath is accepted for backward compatibility but is no longer used —
// callers should additionally set DBPath, NodesRoot, and OSRoot.
func NewEngineConfig(mixPath string) *EngineConfig {
	return &EngineConfig{
		MixProjectPath: mixPath,
		Timeout:        10 * time.Second,
	}
}

// LoadTopologyConfig reads a topology.yaml file and stores the result in
// c.Topology. Call this once at startup (after setting OSRoot or before the
// first Ingest call). When configPath is empty or the file does not exist the
// call is a no-op — routing falls back to defaultRoutingRules.
func (c *EngineConfig) LoadTopologyConfig(configPath string) error {
	if configPath == "" {
		return nil
	}
	topo, err := LoadTopology(configPath)
	if err != nil {
		return err
	}
	c.Topology = topo
	return nil
}

// Init initializes the engine by ensuring the SQLite schema is up to date.
// It should be called once during application startup, after DBPath is set.
// When DBPath is empty the call is a no-op — the engine degrades gracefully.
func (c *EngineConfig) Init() error {
	if c.DBPath == "" {
		return nil
	}
	return EnsureSchema(c.DBPath)
}

// Search runs a full-text search against the SQLite FTS5 index.
// Previously shelled out to `mix optimal.search` — now uses direct SQLite.
// Returns an empty slice (not an error) when the index is unavailable so
// callers degrade gracefully.
func (c *EngineConfig) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if c.DBPath == "" {
		slog.WarnContext(ctx, "optimal: search skipped — DBPath not configured", "query", query)
		return nil, nil
	}
	results, err := SearchContexts(c.DBPath, query, limit)
	if err != nil {
		slog.WarnContext(ctx, "optimal: SQLite search failed",
			"query", query,
			"error", err,
		)
		return nil, nil
	}
	return results, nil
}

// AssembleContext builds a budget-aware, tiered context package for a query.
// It scales L0/L1/L2 budgets proportionally when the receiver specifies a
// TokenBudget, then delegates to the package-level AssembleContext function.
func (c *EngineConfig) AssembleContext(ctx context.Context, query string, receiver *ReceiverProfile) (*AssembledContext, error) {
	cfg := DefaultAssemblerConfig()
	if receiver != nil && receiver.TokenBudget > 0 {
		total := receiver.TokenBudget
		cfg.L0Budget = max(2000, total*13/1000)   // ~1.3%
		cfg.L1Budget = max(5000, total*67/1000)   // ~6.7%
		cfg.L2Budget = max(10000, total*200/1000) // ~20%
	}
	return AssembleContext(c.NodesRoot, c.DBPath, query, receiver, &cfg)
}

// IngestAsync writes a new signal to the filesystem and SQLite index in a
// background goroutine. Uses c.Topology for routing when available, falling
// back to defaultRoutingRules when Topology is nil. Errors are logged but
// never surfaced — fire-and-forget so callers are never blocked or failed by
// ingest latency.
func (c *EngineConfig) IngestAsync(text, genre string) {
	go func() {
		if err := IngestWithConfig(c.OSRoot, c.NodesRoot, c.DBPath, text, genre, c); err != nil {
			slog.Warn("optimal: ingest failed",
				"genre", genre,
				"error", err,
			)
		}
	}()
}
