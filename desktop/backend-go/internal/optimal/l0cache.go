package optimal

import (
	"fmt"
	"strings"
	"time"
)

// L0Context is the always-loaded context buffer (~2000 tokens). It gives the
// agent a bird's-eye view of the knowledge base without reading any full files.
type L0Context struct {
	Nodes         []NodeSummary `json:"nodes"`
	TotalContexts int           `json:"total_contexts"`
	TotalSignals  int           `json:"total_signals"`
	TotalEntities int           `json:"total_entities"`
	GeneratedAt   string        `json:"generated_at"`
	TokenEstimate int           `json:"token_estimate"`
}

// l0TokenCap is the maximum character count before truncation (~2000 tokens at
// 4 chars/token). Individual node entries are dropped tail-first when exceeded.
const l0TokenCap = 8000

// BuildL0 generates the always-loaded L0 context from the nodes filesystem and
// the SQLite index. Either leg may be unavailable — the function degrades
// gracefully and returns whatever it can assemble.
func BuildL0(nodesRoot, dbPath string) (*L0Context, error) {
	ctx := &L0Context{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// ── 1. Load node summaries from filesystem ────────────────────────────────
	if nodesRoot != "" {
		nodes, err := ListNodes(nodesRoot)
		if err == nil {
			ctx.Nodes = nodes
		}
		// Count signals across all nodes.
		for _, n := range ctx.Nodes {
			ctx.TotalSignals += n.SignalCount
		}
	}

	// ── 2. Pull aggregate counts from SQLite ──────────────────────────────────
	if dbPath != "" {
		counts, err := aggregateCounts(dbPath)
		if err == nil {
			ctx.TotalContexts = counts[0]
			ctx.TotalEntities = counts[1]
		}
	}

	// ── 3. Render to text and estimate token count ────────────────────────────
	rendered := renderL0(ctx)
	ctx.TokenEstimate = len(rendered) / 4 // ~4 chars per token

	// ── 4. Trim nodes list if over budget ────────────────────────────────────
	for ctx.TokenEstimate > 2000 && len(ctx.Nodes) > 0 {
		ctx.Nodes = ctx.Nodes[:len(ctx.Nodes)-1]
		rendered = renderL0(ctx)
		ctx.TokenEstimate = len(rendered) / 4
	}

	return ctx, nil
}

// aggregateCounts queries the SQLite database for (context_count, entity_count).
func aggregateCounts(dbPath string) ([2]int, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return [2]int{}, fmt.Errorf("optimal/l0cache: open db: %w", err)
	}
	defer db.Close()

	var counts [2]int

	if err := db.QueryRow(`SELECT COUNT(*) FROM contexts`).Scan(&counts[0]); err != nil {
		counts[0] = 0
	}
	if err := db.QueryRow(`SELECT COUNT(DISTINCT name) FROM entities`).Scan(&counts[1]); err != nil {
		counts[1] = 0
	}
	return counts, nil
}

// renderL0 produces the text representation used for token estimation.
// This matches what the agent would actually see when loading the L0 cache.
func renderL0(ctx *L0Context) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# OptimalOS L0 Context — %s\n\n", ctx.GeneratedAt))
	sb.WriteString(fmt.Sprintf(
		"Index: %d contexts | %d signals | %d entities\n\n",
		ctx.TotalContexts, ctx.TotalSignals, ctx.TotalEntities,
	))
	sb.WriteString("## Nodes\n\n")

	for _, n := range ctx.Nodes {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | signals:%d |\n",
			n.Slug, n.Name, n.Type, n.SignalCount,
		))
	}

	return sb.String()
}
