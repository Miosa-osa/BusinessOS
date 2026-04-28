package bridge

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Knowledge is the bridge to the graph subsystem (internal/optimal/graph.go).
// All queries hit the `edges` table directly — no GenServer hop, no ETS-like
// in-memory copy. Stateless and concurrency-safe.
type Knowledge struct {
	DB *sql.DB
}

// NewKnowledge wires the bridge to *sql.DB.
func NewKnowledge(db *sql.DB) *Knowledge { return &Knowledge{DB: db} }

// SearchResult is the shape a caller hands GraphBoost. We don't import the
// optimal package here to avoid a circular dep (optimal → bridge → optimal);
// callers convert to/from their own struct.
type SearchResult struct {
	ID       string
	Score    float64
	Entities []string
}

// GraphBoost re-scores a result list by graph connectivity. Results whose
// entities appear connected to the query terms via the `edges` table get a
// score multiplier of 1.0 + 0.1×(connection_count). Identical algorithm to
// OptimalEngine.Bridge.Knowledge.graph_boost/2.
//
// On any DB error the original results are returned unchanged — boosting
// must never reduce a caller's result quality below the un-boosted baseline.
func (k *Knowledge) GraphBoost(ctx context.Context, results []SearchResult, query string) []SearchResult {
	if len(results) == 0 || strings.TrimSpace(query) == "" {
		return results
	}
	terms := splitTerms(query)
	if len(terms) == 0 {
		return results
	}
	connected, err := k.findConnectedEntities(ctx, terms)
	if err != nil || len(connected) == 0 {
		return results
	}
	out := make([]SearchResult, len(results))
	for i, r := range results {
		count := 0
		for _, e := range r.Entities {
			if connected[strings.ToLower(e)] {
				count++
			}
		}
		boost := 1.0 + float64(count)*0.1
		out[i] = r
		out[i].Score = round4(r.Score * boost)
	}
	return out
}

// ContextFor builds a markdown snapshot of a topic's graph relationships.
// Returns "No knowledge graph context found for: <entity>" when no edges
// reference the entity (matching the Elixir source).
func (k *Knowledge) ContextFor(ctx context.Context, entity string) (string, error) {
	if entity == "" {
		return "", errors.New("bridge/knowledge: entity required")
	}
	if k == nil || k.DB == nil {
		return "No knowledge graph context found for: " + entity, nil
	}
	pattern := "%" + entity + "%"
	rows, err := k.DB.QueryContext(ctx, `
		SELECT source_id, relation, target_id FROM edges WHERE source_id LIKE $1
		UNION ALL
		SELECT source_id, relation, target_id FROM edges WHERE target_id LIKE $1
	`, pattern)
	if err != nil {
		return "", fmt.Errorf("bridge/knowledge: query edges: %w", err)
	}
	defer rows.Close()

	type triple struct{ src, rel, dst string }
	var edges []triple
	for rows.Next() {
		var t triple
		if err := rows.Scan(&t.src, &t.rel, &t.dst); err != nil {
			return "", err
		}
		edges = append(edges, t)
	}
	if len(edges) == 0 {
		return "No knowledge graph context found for: " + entity, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Knowledge graph context: %s\n\n", entity)
	for _, e := range edges {
		fmt.Fprintf(&sb, "- %s — %s → %s\n", e.src, e.rel, e.dst)
	}
	return sb.String(), nil
}

// findConnectedEntities returns the set of entity ids reachable from any
// query term in two hops (term → entity → entity), lower-cased for matching.
//
// A nil DB is treated as "no connections found" rather than an error, so
// GraphBoost is safe to call before the engine has finished bootstrapping.
func (k *Knowledge) findConnectedEntities(ctx context.Context, terms []string) (map[string]bool, error) {
	if k == nil || k.DB == nil {
		return map[string]bool{}, nil
	}
	out := map[string]bool{}
	for _, term := range terms {
		pattern := "%" + term + "%"
		rows, err := k.DB.QueryContext(ctx, `
			SELECT target_id FROM edges WHERE source_id ILIKE $1
			UNION
			SELECT source_id FROM edges WHERE target_id ILIKE $1
		`, pattern)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			out[strings.ToLower(id)] = true
		}
		rows.Close()
	}
	return out, nil
}

func splitTerms(query string) []string {
	parts := strings.Fields(strings.ToLower(query))
	out := parts[:0]
	for _, p := range parts {
		if len(p) >= 3 { // skip short stop-words
			out = append(out, p)
		}
	}
	return out
}

func round4(f float64) float64 {
	return float64(int(f*10000+0.5)) / 10000
}
