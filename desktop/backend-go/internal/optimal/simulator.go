package optimal

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// SimulationResult is the output of a "what if" mutation trace through the
// knowledge graph. It lists all nodes affected, their estimated impact
// (0.0–1.0), how the impact reaches them, and observed risk signals.
type SimulationResult struct {
	Mutation      string         `json:"mutation"`
	AffectedNodes []AffectedNode `json:"affected_nodes"`
	TotalImpact   float64        `json:"total_impact"`
	Risks         []string       `json:"risks"`
}

// AffectedNode is a single node reachable from the mutated entity with its
// estimated impact score and the traversal path that connects it.
type AffectedNode struct {
	Name   string  `json:"name"`
	Impact float64 `json:"impact"` // 0.0–1.0
	Path   string  `json:"path"`   // hop chain e.g. "Ed → AI Masters → Revenue"
}

// simLink is an internal adjacency list entry used during simulation.
type simLink struct {
	name, relation string
	weight         float64
}

// edgeRelationWeights maps edge relation types to base impact multipliers.
// Matches the Elixir Simulator weights.
var edgeRelationWeights = map[string]float64{
	"works_on":     1.5,
	"cross_ref":    1.2,
	"mentioned_in": 1.0,
	"lives_in":     0.6,
}

const defaultEdgeWeight = 0.8

// depthDecay factors per hop depth (index = depth, 0 unused).
var depthDecay = [4]float64{0, 1.0, 0.5, 0.25}

// maxSimDepth is the maximum BFS traversal depth (inclusive).
const maxSimDepth = 3

// Simulate traces a "what if" mutation through the knowledge graph.
//
//  1. Extracts the primary entity from the mutation string.
//  2. Builds an undirected adjacency list from the SQLite edge table.
//  3. BFS-propagates impact up to maxSimDepth hops with depth decay.
//  4. Returns affected nodes sorted by impact descending.
//
// Returns an empty result (not an error) when the database is unavailable.
func Simulate(dbPath string, mutation string) (*SimulationResult, error) {
	if mutation == "" {
		return &SimulationResult{Mutation: mutation}, nil
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/simulator: open db: %w", err)
	}
	defer db.Close()

	// Load all edges resolved to entity names.
	type rawEdge struct {
		src, dst, relation string
		weight             float64
	}

	const q = `
		SELECT
			COALESCE((SELECT name FROM entities WHERE context_id = e.source_id LIMIT 1), e.source_id),
			COALESCE((SELECT name FROM entities WHERE context_id = e.target_id LIMIT 1), e.target_id),
			e.relation,
			e.weight
		FROM edges e`

	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("optimal/simulator: query edges: %w", err)
	}
	defer rows.Close()

	var rawEdges []rawEdge
	for rows.Next() {
		var re rawEdge
		if err := rows.Scan(&re.src, &re.dst, &re.relation, &re.weight); err != nil {
			return nil, fmt.Errorf("optimal/simulator: scan edge: %w", err)
		}
		rawEdges = append(rawEdges, re)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/simulator: iterate edges: %w", err)
	}

	// Build undirected adjacency list keyed by entity name.
	adj := make(map[string][]simLink, len(rawEdges)*2)
	for _, re := range rawEdges {
		w := computeEdgeWeight(re.relation, re.weight)
		adj[re.src] = append(adj[re.src], simLink{re.dst, re.relation, w})
		adj[re.dst] = append(adj[re.dst], simLink{re.src, re.relation, w})
	}

	// Find the best-matching entity name for the extracted subject.
	subject := extractSubject(mutation)
	startEntity := findEntityInMap(subject, adj)
	if startEntity == "" {
		return &SimulationResult{
			Mutation: mutation,
			Risks:    []string{"no entity matching '" + subject + "' found in graph"},
		}, nil
	}

	// BFS impact propagation.
	type visit struct {
		name   string
		depth  int
		impact float64
		path   string
	}

	impactMap := make(map[string]float64)
	pathMap := make(map[string]string)
	visitedSet := map[string]bool{startEntity: true}

	queue := []visit{{name: startEntity, depth: 0, impact: 1.0, path: startEntity}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.depth >= maxSimDepth {
			continue
		}
		nextDepth := curr.depth + 1
		decay := depthDecay[nextDepth]

		for _, nb := range adj[curr.name] {
			if visitedSet[nb.name] {
				continue
			}
			visitedSet[nb.name] = true

			impact := curr.impact * nb.weight * decay
			if impact < 0.01 {
				continue // prune negligible
			}

			hopPath := curr.path + " → " + nb.name
			if prev, ok := impactMap[nb.name]; !ok || impact > prev {
				impactMap[nb.name] = impact
				pathMap[nb.name] = hopPath
			}
			queue = append(queue, visit{
				name: nb.name, depth: nextDepth, impact: impact, path: hopPath,
			})
		}
	}

	// Collect affected nodes, clamping impact to [0,1].
	affected := make([]AffectedNode, 0, len(impactMap))
	var totalImpact float64
	for name, imp := range impactMap {
		if name == startEntity {
			continue
		}
		if imp > 1.0 {
			imp = 1.0
		}
		affected = append(affected, AffectedNode{
			Name:   name,
			Impact: imp,
			Path:   pathMap[name],
		})
		totalImpact += imp
	}

	sort.Slice(affected, func(i, j int) bool {
		return affected[i].Impact > affected[j].Impact
	})
	if totalImpact > 1.0 {
		totalImpact = 1.0
	}

	return &SimulationResult{
		Mutation:      mutation,
		AffectedNodes: affected,
		TotalImpact:   totalImpact,
		Risks:         buildRisks(startEntity, affected),
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

var reEntityExtract = regexp.MustCompile(`"([^"]+)"|'([^']+)'|\b([A-Z][a-zA-Z]{1,})\b`)
var reSimStopWords = regexp.MustCompile(`(?i)^(what|if|when|how|why|who|the|a|an|is|are|would|could|should|we|i|you|they|it|this|that)$`)

// extractSubject pulls the primary subject from a mutation string.
// e.g. "What if Ed leaves?" → "Ed"
func extractSubject(mutation string) string {
	if m := reEntityExtract.FindStringSubmatch(mutation); m != nil {
		for _, g := range m[1:] {
			if g != "" {
				return g
			}
		}
	}
	// Fallback: first non-stop-word token.
	for _, w := range strings.Fields(mutation) {
		w = strings.Trim(w, `.,!?"'()[]{}`)
		if w != "" && !reSimStopWords.MatchString(w) {
			return w
		}
	}
	return mutation
}

// findEntityInMap returns the key in adj that best matches subject.
// Priority: exact (case-insensitive) > prefix > substring. Returns "" if none.
func findEntityInMap(subject string, adj map[string][]simLink) string {
	lower := strings.ToLower(subject)
	for name := range adj {
		if strings.EqualFold(name, subject) {
			return name
		}
	}
	for name := range adj {
		if strings.HasPrefix(strings.ToLower(name), lower) {
			return name
		}
	}
	for name := range adj {
		if strings.Contains(strings.ToLower(name), lower) {
			return name
		}
	}
	return ""
}

// computeEdgeWeight blends the stored weight with the relation multiplier.
func computeEdgeWeight(relation string, storedWeight float64) float64 {
	mult, ok := edgeRelationWeights[relation]
	if !ok {
		mult = defaultEdgeWeight
	}
	w := storedWeight * mult
	if w > 1.5 {
		w = 1.5
	}
	return w
}

// buildRisks produces a short list of textual risk signals.
func buildRisks(entity string, affected []AffectedNode) []string {
	if len(affected) == 0 {
		return []string{"no downstream impact detected"}
	}
	var risks []string
	if len(affected) > 5 {
		risks = append(risks, fmt.Sprintf("high blast radius: %d nodes affected", len(affected)))
	}
	for _, n := range affected {
		if n.Impact > 0.7 {
			risks = append(risks, fmt.Sprintf("critical dependency: %s (impact %.2f)", n.Name, n.Impact))
		}
	}
	if len(risks) == 0 {
		risks = append(risks, fmt.Sprintf("low risk: %s removal has contained impact", entity))
	}
	return risks
}
