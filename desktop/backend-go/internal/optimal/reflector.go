package optimal

import (
	"fmt"
	"sort"
)

// MissingEdge is a pair of entities that co-occur in one or more shared contexts
// but have no direct edge between them in the knowledge graph. These represent
// synthesis opportunities — the agent should consider whether a relationship
// should be made explicit.
type MissingEdge struct {
	EntityA       string   `json:"entity_a"`
	EntityB       string   `json:"entity_b"`
	CoOccurrences int      `json:"co_occurrences"`
	SharedNodes   []string `json:"shared_nodes"`
}

// Reflect finds entity pairs that appear together in contexts but lack an
// explicit graph edge. Only pairs with at least minCooccurrences shared contexts
// are returned, sorted by co-occurrence count descending.
//
// The algorithm:
//  1. Load all (entity_name, context_id, node) rows from the entities table.
//  2. Build a map: context_id → []entity_name.
//  3. For every unordered pair of entities sharing a context, increment their
//     co-occurrence counter and record the shared node.
//  4. Load all existing edges into a set.
//  5. Return pairs whose count >= minCooccurrences and that have no edge.
func Reflect(dbPath string, minCooccurrences int) ([]MissingEdge, error) {
	if minCooccurrences <= 0 {
		minCooccurrences = 2
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/reflector: open db: %w", err)
	}
	defer db.Close()

	// ── Step 1: load entity → context mapping ────────────────────────────────
	type entityRow struct {
		name      string
		contextID string
		node      string
	}

	rows, err := db.Query(`SELECT name, context_id, COALESCE(node, '') FROM entities`)
	if err != nil {
		return nil, fmt.Errorf("optimal/reflector: query entities: %w", err)
	}
	defer rows.Close()

	// context_id → list of entity names appearing in that context.
	byContext := make(map[string][]entityRow)
	for rows.Next() {
		var r entityRow
		if err := rows.Scan(&r.name, &r.contextID, &r.node); err != nil {
			return nil, fmt.Errorf("optimal/reflector: scan entity: %w", err)
		}
		byContext[r.contextID] = append(byContext[r.contextID], r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/reflector: iterate entities: %w", err)
	}

	// ── Step 2: build co-occurrence counts ────────────────────────────────────
	type pairKey = [2]string
	coCount := make(map[pairKey]int)
	coNodes := make(map[pairKey]map[string]struct{})

	for _, members := range byContext {
		for i := 0; i < len(members); i++ {
			for j := i + 1; j < len(members); j++ {
				a, b := members[i].name, members[j].name
				if a == b {
					continue
				}
				// Canonical order: alphabetically smaller first.
				if a > b {
					a, b = b, a
				}
				key := pairKey{a, b}
				coCount[key]++
				if coNodes[key] == nil {
					coNodes[key] = make(map[string]struct{})
				}
				// Use the node from entity a (both share the same context).
				node := members[i].node
				if node == "" {
					node = members[j].node
				}
				if node != "" {
					coNodes[key][node] = struct{}{}
				}
			}
		}
	}

	// ── Step 3: load existing edges ───────────────────────────────────────────
	edgePairs, err := edgePairList(db)
	if err != nil {
		return nil, fmt.Errorf("optimal/reflector: load edges: %w", err)
	}

	// Build the entity name → context_ids map to resolve edge source/target IDs.
	idToName, err := entityNameMap(db)
	if err != nil {
		return nil, fmt.Errorf("optimal/reflector: load entity names: %w", err)
	}

	existingEdges := make(map[pairKey]struct{}, len(edgePairs))
	for _, p := range edgePairs {
		src, dst := idToName[p[0]], idToName[p[1]]
		if src == "" || dst == "" {
			continue
		}
		if src > dst {
			src, dst = dst, src
		}
		existingEdges[pairKey{src, dst}] = struct{}{}
	}

	// ── Step 4: collect missing edges above threshold ─────────────────────────
	var missing []MissingEdge
	for key, count := range coCount {
		if count < minCooccurrences {
			continue
		}
		if _, exists := existingEdges[key]; exists {
			continue
		}
		nodes := make([]string, 0, len(coNodes[key]))
		for n := range coNodes[key] {
			nodes = append(nodes, n)
		}
		sort.Strings(nodes)
		missing = append(missing, MissingEdge{
			EntityA:       key[0],
			EntityB:       key[1],
			CoOccurrences: count,
			SharedNodes:   nodes,
		})
	}

	// Sort by co-occurrence count descending for highest-value synthesis first.
	sort.Slice(missing, func(i, j int) bool {
		return missing[i].CoOccurrences > missing[j].CoOccurrences
	})

	return missing, nil
}
