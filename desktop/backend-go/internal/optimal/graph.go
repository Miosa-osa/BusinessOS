package optimal

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required
)

// GraphEntity is a deduplicated entity with its total connection count.
type GraphEntity struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Connections int    `json:"connections"`
}

// GraphEdge represents a directed relationship between two entities.
type GraphEdge struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Relation string  `json:"relation"`
	Weight   float64 `json:"weight"`
}

// GraphStats holds aggregate counts for the graph.
type GraphStats struct {
	EntityCount int            `json:"entity_count"`
	EdgeCount   int            `json:"edge_count"`
	EdgeTypes   map[string]int `json:"edge_types"`
}

// GraphData is the full graph payload returned to callers.
type GraphData struct {
	Entities []GraphEntity `json:"entities"`
	Edges    []GraphEdge   `json:"edges"`
	Stats    GraphStats    `json:"stats"`
}

// GetGraph reads all entities (deduplicated by name, connection counts aggregated)
// and all edges from the SQLite database at dbPath.
func GetGraph(dbPath string) (GraphData, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return GraphData{}, err
	}
	defer db.Close()

	entities, err := queryEntities(db)
	if err != nil {
		return GraphData{}, fmt.Errorf("optimal/graph: query entities: %w", err)
	}

	edges, err := queryEdges(db)
	if err != nil {
		return GraphData{}, fmt.Errorf("optimal/graph: query edges: %w", err)
	}

	stats := buildStats(entities, edges)
	return GraphData{Entities: entities, Edges: edges, Stats: stats}, nil
}

// GetGraphHubs returns the top limit most-connected entities by total edge count.
func GetGraphHubs(dbPath string, limit int) ([]GraphEntity, error) {
	if limit <= 0 {
		limit = 20
	}
	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Sum outgoing + incoming edges per entity name, deduplicating across context_ids.
	const q = `
		SELECT name, type, SUM(degree) AS connections
		FROM (
			SELECT e.name, e.type,
				(
					SELECT COUNT(*) FROM edges ed
					WHERE ed.source_id IN (SELECT context_id FROM entities WHERE name = e.name)
					   OR ed.target_id IN (SELECT context_id FROM entities WHERE name = e.name)
				) AS degree
			FROM entities e
			GROUP BY e.name, e.type
		)
		GROUP BY name, type
		ORDER BY connections DESC
		LIMIT ?`

	rows, err := db.Query(q, limit)
	if err != nil {
		return nil, fmt.Errorf("optimal/graph: query hubs: %w", err)
	}
	defer rows.Close()

	var hubs []GraphEntity
	for rows.Next() {
		var g GraphEntity
		if err := rows.Scan(&g.Name, &g.Type, &g.Connections); err != nil {
			return nil, fmt.Errorf("optimal/graph: scan hub row: %w", err)
		}
		hubs = append(hubs, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/graph: iterate hubs: %w", err)
	}
	return hubs, nil
}

// SearchResult is a single FTS5 hit from the contexts table.
// The SearchResult type is also declared in engine.go (subprocess version);
// this definition is intentionally aligned with that shape.
// (Go allows only one declaration per package — engine.go defines SearchResult,
// so we reuse it here instead of redeclaring.)

// SearchContexts performs a full-text search directly against the FTS5 index
// in the SQLite database, bypassing the Elixir subprocess for lower latency.
// It returns at most limit results ordered by FTS5 relevance rank.
func SearchContexts(dbPath, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// snippet() column index 2 = content column in the FTS5 virtual table.
	const q = `
		SELECT c.id, COALESCE(c.path, ''), c.title, c.node,
		       snippet(contexts_fts, 2, '', '', '...', 30) AS abstract
		FROM contexts_fts f
		JOIN contexts c ON f.rowid = c.rowid
		WHERE contexts_fts MATCH ?
		ORDER BY rank
		LIMIT ?`

	rows, err := db.Query(q, query, limit)
	if err != nil {
		return nil, fmt.Errorf("optimal/graph: fts search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var id, path, title, node string
		if err := rows.Scan(&id, &path, &title, &node, &r.Abstract); err != nil {
			return nil, fmt.Errorf("optimal/graph: scan search row: %w", err)
		}
		// Build a human-readable path field. Prefer the stored path, fall back to
		// a synthetic "node/title" representation so callers always get a non-empty value.
		if path != "" {
			r.Path = path
		} else {
			r.Path = node + "/" + title
		}
		// Score is not available from FTS5 rank in a cross-join; default to 1.0.
		r.Score = 1.0
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/graph: iterate search results: %w", err)
	}
	return results, nil
}

// ── internal helpers ──────────────────────────────────────────────────────────

func openDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/graph: open db %q: %w", dbPath, err)
	}
	// Validate the connection immediately so callers get a clear error.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("optimal/graph: ping db %q: %w", dbPath, err)
	}
	return db, nil
}

// queryEntities returns entities deduplicated by name with summed connection counts.
func queryEntities(db *sql.DB) ([]GraphEntity, error) {
	const q = `
		SELECT e.name, e.type,
			COUNT(DISTINCT ed.id) AS connections
		FROM entities e
		LEFT JOIN edges ed
			ON ed.source_id = e.context_id OR ed.target_id = e.context_id
		GROUP BY e.name, e.type
		ORDER BY e.name`

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Deduplicate by name: keep the type with the highest connection count.
	type key = string
	byName := make(map[key]GraphEntity)

	for rows.Next() {
		var g GraphEntity
		if err := rows.Scan(&g.Name, &g.Type, &g.Connections); err != nil {
			return nil, err
		}
		if existing, ok := byName[g.Name]; !ok || g.Connections > existing.Connections {
			byName[g.Name] = g
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	entities := make([]GraphEntity, 0, len(byName))
	for _, e := range byName {
		entities = append(entities, e)
	}
	return entities, nil
}

// queryEdges returns all edges, mapping source_id/target_id back to entity names.
func queryEdges(db *sql.DB) ([]GraphEdge, error) {
	const q = `
		SELECT
			COALESCE((SELECT name FROM entities WHERE context_id = ed.source_id LIMIT 1), ed.source_id),
			COALESCE((SELECT name FROM entities WHERE context_id = ed.target_id LIMIT 1), ed.target_id),
			ed.relation,
			ed.weight
		FROM edges ed`

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []GraphEdge
	for rows.Next() {
		var e GraphEdge
		if err := rows.Scan(&e.Source, &e.Target, &e.Relation, &e.Weight); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return edges, nil
}

func buildStats(entities []GraphEntity, edges []GraphEdge) GraphStats {
	edgeTypes := make(map[string]int)
	for _, e := range edges {
		edgeTypes[e.Relation]++
	}
	return GraphStats{
		EntityCount: len(entities),
		EdgeCount:   len(edges),
		EdgeTypes:   edgeTypes,
	}
}
