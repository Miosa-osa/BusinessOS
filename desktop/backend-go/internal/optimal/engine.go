package optimal

import (
	"context"
	"encoding/json"
	"log/slog"
	"os/exec"
	"strconv"
	"time"
)

// SearchResult is a single hit returned by mix optimal.search --json.
type SearchResult struct {
	Path     string  `json:"path"`
	Abstract string  `json:"abstract"`
	Score    float64 `json:"score"`
}

// EngineConfig holds runtime parameters for invoking the Elixir mix CLI.
type EngineConfig struct {
	// MixProjectPath is the absolute path to the Elixir project directory
	// (the one containing mix.exs), e.g. "/Users/rhl/Desktop/OptimalOS/engine".
	MixProjectPath string
	// Timeout is the maximum duration allowed for a synchronous engine call.
	Timeout time.Duration
}

// NewEngineConfig returns an EngineConfig with a 10-second default timeout.
func NewEngineConfig(mixPath string) *EngineConfig {
	return &EngineConfig{
		MixProjectPath: mixPath,
		Timeout:        10 * time.Second,
	}
}

// Search runs `mix optimal.search "<query>" --limit N --json` inside MixProjectPath
// and parses the JSON output into a slice of SearchResult. If the engine process
// fails, times out, or returns unparseable output the function returns an empty
// slice rather than an error — callers degrade gracefully when the engine is down.
func (c *EngineConfig) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	tctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	cmd := exec.CommandContext(tctx, "mix", "optimal.search", query,
		"--limit", strconv.Itoa(limit),
		"--json",
	)
	cmd.Dir = c.MixProjectPath

	out, err := cmd.Output()
	if err != nil {
		// Engine unavailable or timed out — degrade gracefully.
		slog.WarnContext(ctx, "optimal engine search failed",
			"query", query,
			"error", err,
		)
		return nil, nil
	}

	var results []SearchResult
	if jsonErr := json.Unmarshal(out, &results); jsonErr != nil {
		slog.WarnContext(ctx, "optimal engine search: unparseable JSON",
			"query", query,
			"error", jsonErr,
		)
		return nil, nil
	}
	return results, nil
}

// IngestAsync fires `mix optimal.ingest "<text>" --genre <genre>` in a background
// goroutine. Errors are logged but never surfaced — this is intentionally
// fire-and-forget so callers are never blocked or failed by engine latency.
func (c *EngineConfig) IngestAsync(text, genre string) {
	go func() {
		cmd := exec.Command("mix", "optimal.ingest", text, "--genre", genre)
		cmd.Dir = c.MixProjectPath

		if out, err := cmd.CombinedOutput(); err != nil {
			slog.Warn("optimal engine ingest failed",
				"genre", genre,
				"error", err,
				"output", string(out),
			)
		}
	}()
}
