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

// IngestAsync writes a new signal to the filesystem and SQLite index in a
// background goroutine. Errors are logged but never surfaced — fire-and-forget
// so callers are never blocked or failed by ingest latency.
func (c *EngineConfig) IngestAsync(text, genre string) {
	go func() {
		if err := Ingest(c.OSRoot, c.NodesRoot, c.DBPath, text, genre); err != nil {
			slog.Warn("optimal: ingest failed",
				"genre", genre,
				"error", err,
			)
		}
	}()
}
