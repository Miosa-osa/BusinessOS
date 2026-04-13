package optimal

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
)

// Triangle is an open triad in the knowledge graph: A shares edges with both B
// and C, but no edge exists between B and C. These are synthesis opportunities.
type Triangle struct {
	A string `json:"a"` // shared hub entity
	B string `json:"b"` // first spoke
	C string `json:"c"` // second spoke — no direct edge to B
}

// Cluster is a connected component discovered via BFS on the undirected graph.
type Cluster struct {
	ID       int      `json:"id"`
	Entities []string `json:"entities"`
	Size     int      `json:"size"`
}

// FindTriangles finds open triangles (A→B and A→C exist, B↔C does not).
// Returns at most limit results. When limit <= 0 it defaults to 20.
// Returns an empty slice (not an error) when the database is absent or empty.
func FindTriangles(dbPath string, limit int) ([]Triangle, error) {
	if limit <= 0 {
		limit = 20
	}
	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/analyzer: open db: %w", err)
	}
	defer db.Close()

	idToName, err := entityNameMap(db)
	if err != nil {
		return nil, fmt.Errorf("optimal/analyzer: load entity names: %w", err)
	}
	if len(idToName) == 0 {
		return []Triangle{}, nil
	}

	edgePairs, err := edgePairList(db)
	if err != nil {
		return nil, fmt.Errorf("optimal/analyzer: load edges: %w", err)
	}
	if len(edgePairs) == 0 {
		return []Triangle{}, nil
	}

	// Build undirected adjacency set and neighbour lists (by entity name).
	edgeSet := make(map[[2]string]struct{}, len(edgePairs)*2)
	neighbours := make(map[string][]string)

	for _, p := range edgePairs {
		srcName, okS := idToName[p[0]]
		dstName, okD := idToName[p[1]]
		if !okS || !okD || srcName == dstName {
			continue
		}
		addEdge(edgeSet, srcName, dstName)
		neighbours[srcName] = append(neighbours[srcName], dstName)
		neighbours[dstName] = append(neighbours[dstName], srcName)
	}

	// For each entity A examine every unordered pair (B, C) from its neighbours.
	// Emit a Triangle when the B↔C edge is absent.
	seen := make(map[[3]string]struct{})
	var triangles []Triangle

	// Sort entity names for deterministic output.
	hubs := make([]string, 0, len(neighbours))
	for a := range neighbours {
		hubs = append(hubs, a)
	}
	sort.Strings(hubs)

	for _, a := range hubs {
		nbrs := neighbours[a]
		for i := 0; i < len(nbrs) && len(triangles) < limit; i++ {
			for j := i + 1; j < len(nbrs) && len(triangles) < limit; j++ {
				b, c := nbrs[i], nbrs[j]
				if b == c {
					continue
				}
				// Canonicalise so we don't emit the same triangle twice.
				if b > c {
					b, c = c, b
				}
				key := [3]string{a, b, c}
				if _, dup := seen[key]; dup {
					continue
				}
				seen[key] = struct{}{}

				if !hasEdge(edgeSet, b, c) {
					triangles = append(triangles, Triangle{A: a, B: b, C: c})
				}
			}
		}
		if len(triangles) >= limit {
			break
		}
	}

	return triangles, nil
}

// FindClusters discovers connected components via BFS on the undirected graph.
// Returns clusters sorted descending by size (largest first).
// Returns an empty slice (not an error) when the database is absent or empty.
func FindClusters(dbPath string) ([]Cluster, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/analyzer: open db: %w", err)
	}
	defer db.Close()

	idToName, err := entityNameMap(db)
	if err != nil {
		return nil, fmt.Errorf("optimal/analyzer: load entity names: %w", err)
	}
	if len(idToName) == 0 {
		return []Cluster{}, nil
	}

	edgePairs, err := edgePairList(db)
	if err != nil {
		return nil, fmt.Errorf("optimal/analyzer: load edges: %w", err)
	}

	// Build undirected adjacency list (by entity name).
	adj := make(map[string][]string)
	for _, name := range idToName {
		if _, ok := adj[name]; !ok {
			adj[name] = nil // ensure isolated entities appear
		}
	}
	for _, p := range edgePairs {
		src, dst := idToName[p[0]], idToName[p[1]]
		if src == "" || dst == "" || src == dst {
			continue
		}
		adj[src] = append(adj[src], dst)
		adj[dst] = append(adj[dst], src)
	}

	// BFS over all entity names in sorted order for determinism.
	names := make([]string, 0, len(adj))
	for n := range adj {
		names = append(names, n)
	}
	sort.Strings(names)

	visited := make(map[string]bool, len(adj))
	var clusters []Cluster
	clusterID := 0

	for _, start := range names {
		if visited[start] {
			continue
		}
		var component []string
		queue := []string{start}
		visited[start] = true
		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			component = append(component, curr)
			for _, nb := range adj[curr] {
				if !visited[nb] {
					visited[nb] = true
					queue = append(queue, nb)
				}
			}
		}
		sort.Strings(component)
		clusters = append(clusters, Cluster{
			ID:       clusterID,
			Entities: component,
			Size:     len(component),
		})
		clusterID++
	}

	// Largest cluster first.
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Size > clusters[j].Size
	})

	return clusters, nil
}

// FindHubs returns entities whose total degree (in + out edges) exceeds 2
// standard deviations above the mean across all entities. Results are sorted
// descending by connection count and capped at limit.
// Reuses GraphEntity from graph.go.
func FindHubs(dbPath string, limit int) ([]GraphEntity, error) {
	if limit <= 0 {
		limit = 20
	}
	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/analyzer: open db: %w", err)
	}
	defer db.Close()

	entities, err := queryEntities(db)
	if err != nil {
		return nil, fmt.Errorf("optimal/analyzer: query entities: %w", err)
	}
	if len(entities) == 0 {
		return []GraphEntity{}, nil
	}

	// Mean and standard deviation of connection counts.
	var sum float64
	for _, e := range entities {
		sum += float64(e.Connections)
	}
	mean := sum / float64(len(entities))

	var variance float64
	for _, e := range entities {
		d := float64(e.Connections) - mean
		variance += d * d
	}
	variance /= float64(len(entities))
	stddev := math.Sqrt(variance)
	threshold := mean + 2*stddev

	var hubs []GraphEntity
	for _, e := range entities {
		if float64(e.Connections) > threshold {
			hubs = append(hubs, e)
		}
	}

	sort.Slice(hubs, func(i, j int) bool {
		return hubs[i].Connections > hubs[j].Connections
	})
	if len(hubs) > limit {
		hubs = hubs[:limit]
	}
	return hubs, nil
}

// ── internal helpers ──────────────────────────────────────────────────────────

// entityNameMap returns a map of context_id → entity name from the entities table.
func entityNameMap(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query(`SELECT context_id, name FROM entities`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		// Keep the first name seen for duplicate context_ids.
		if _, exists := m[id]; !exists {
			m[id] = name
		}
	}
	return m, rows.Err()
}

// edgePairList returns all (source_id, target_id) pairs from the edges table.
func edgePairList(db *sql.DB) ([][2]string, error) {
	rows, err := db.Query(`SELECT source_id, target_id FROM edges`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pairs [][2]string
	for rows.Next() {
		var src, dst string
		if err := rows.Scan(&src, &dst); err != nil {
			return nil, err
		}
		pairs = append(pairs, [2]string{src, dst})
	}
	return pairs, rows.Err()
}

// addEdge inserts an undirected edge into the set using a canonical (min,max) key.
func addEdge(set map[[2]string]struct{}, a, b string) {
	if a > b {
		a, b = b, a
	}
	set[[2]string{a, b}] = struct{}{}
}

// hasEdge reports whether an undirected edge between a and b is in the set.
func hasEdge(set map[[2]string]struct{}, a, b string) bool {
	if a > b {
		a, b = b, a
	}
	_, ok := set[[2]string{a, b}]
	return ok
}
