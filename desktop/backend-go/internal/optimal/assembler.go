package optimal

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ReceiverProfile defines what a receiver can decode and how much they can handle.
type ReceiverProfile struct {
	Name            string   `json:"name"`
	Role            string   `json:"role"`             // executive, operator, technical, salesperson, agent
	GenreCompetence []string `json:"genre_competence"` // genres this receiver can decode
	TokenBudget     int      `json:"token_budget"`     // max tokens for this receiver
	PreferredGenre  string   `json:"preferred_genre"`  // what genre to compose in
	PreferredMode   string   `json:"preferred_mode"`   // linguistic, visual, code, data
}

// AssembledContext is the final output of the Context Assembler.
type AssembledContext struct {
	// Tier content
	L0 string `json:"l0"` // always-loaded headline context (~2K tokens)
	L1 string `json:"l1"` // task-relevant context (~10K tokens)
	L2 string `json:"l2"` // deep detail (only if requested/needed)

	// Token accounting
	L0Tokens    int     `json:"l0_tokens"`
	L1Tokens    int     `json:"l1_tokens"`
	L2Tokens    int     `json:"l2_tokens"`
	TotalTokens int     `json:"total_tokens"`
	BudgetUsed  float64 `json:"budget_used"` // 0.0-1.0 fraction of budget consumed

	// Retrieval trace (debuggable — "why was this included?")
	Sources []ContextSource `json:"sources"`

	// Assembly metadata
	Query           string          `json:"query"`
	ReceiverProfile ReceiverProfile `json:"receiver"`
	AssemblyTimeMs  int64           `json:"assembly_time_ms"`
}

// ContextSource traces why a piece of content was included.
type ContextSource struct {
	Path   string  `json:"path"`
	Title  string  `json:"title"`
	Genre  string  `json:"genre"`
	Tier   string  `json:"tier"`   // "l0", "l1", "l2"
	Score  float64 `json:"score"`  // relevance score
	Tokens int     `json:"tokens"` // tokens consumed by this source
	Reason string  `json:"reason"` // why included: "always_loaded", "query_match", "entity_connected", "drill_down"
}

// AssemblerConfig holds tunable parameters.
type AssemblerConfig struct {
	L0Budget    int     // tokens for L0 (default 2000)
	L1Budget    int     // tokens for L1 (default 10000)
	L2Budget    int     // tokens for L2 (default 50000)
	L1Threshold float64 // minimum relevance score to include in L1 (default 0.1)
	MaxSources  int     // max number of source documents to include (default 20)
}

// DefaultAssemblerConfig returns sensible defaults.
func DefaultAssemblerConfig() AssemblerConfig {
	return AssemblerConfig{
		L0Budget:    2000,
		L1Budget:    10000,
		L2Budget:    50000,
		L1Threshold: 0.1,
		MaxSources:  20,
	}
}

// genreImportance weights document genres by durable business value.
// Decisions and ADRs carry the most signal; raw notes the least.
var genreImportance = map[string]float64{
	"decision":     1.0,
	"decision-log": 1.0,
	"adr":          0.95,
	"spec":         0.9,
	"plan":         0.85,
	"report":       0.8,
	"brief":        0.75,
	"review":       0.7,
	"standup":      0.6,
	"transcript":   0.5,
	"note":         0.4,
	"pitch":        0.7,
}

// defaultGenreImportance is used when a genre is not in the table.
const defaultGenreImportance = 0.5

// estimateTokens returns an approximate token count (1 token ≈ 4 chars).
func estimateTokens(text string) int {
	return (len(text) + 3) / 4
}

// requiresL2 reports whether the receiver profile and query indicate that deep
// context (L2) should be assembled. Only technical and agent roles get L2, and
// only when the query contains depth-signalling phrases.
func requiresL2(query string, receiver *ReceiverProfile) bool {
	if receiver == nil {
		return false
	}
	role := strings.ToLower(receiver.Role)
	if role != "technical" && role != "agent" {
		return false
	}
	lower := strings.ToLower(query)
	depthKeywords := []string{
		"deep dive", "full detail", "complete", "entire", "all context",
		"implementation", "how does", "architecture", "internals", "breakdown",
	}
	for _, kw := range depthKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// AssembleContext builds a budget-aware, tiered context package for a query.
//
// The assembly pipeline:
//  1. Build L0 (always-loaded) from the L0 cache or generate fresh
//  2. Run hybrid search for query-relevant signals
//  3. Score each candidate by relevance + recency + importance
//  4. Fill L1 tier with top candidates within L1 budget (WHOLE documents only, never truncate)
//  5. If receiver profile indicates need for L2, expand highest-scoring L1 entries
//  6. Apply receiver genre conversion if needed (salesperson gets brief, not spec)
//  7. Return AssembledContext with full retrieval trace
//
// Token estimation: 1 token ≈ 4 characters (conservative estimate).
func AssembleContext(nodesRoot, dbPath, query string, receiver *ReceiverProfile, cfg *AssemblerConfig) (*AssembledContext, error) {
	start := time.Now()

	if cfg == nil {
		defaults := DefaultAssemblerConfig()
		cfg = &defaults
	}
	if receiver == nil {
		receiver = &ReceiverProfile{
			Role:        "agent",
			TokenBudget: cfg.L0Budget + cfg.L1Budget + cfg.L2Budget,
		}
	}

	assembled := &AssembledContext{
		Query:           query,
		ReceiverProfile: *receiver,
	}

	// ── Step 1: L0 (always loaded) ────────────────────────────────────────────

	l0ctx, err := BuildL0(nodesRoot, dbPath)
	if err != nil {
		// L0 is non-negotiable — fail hard if we can't build it.
		return nil, fmt.Errorf("optimal/assembler: build L0: %w", err)
	}
	l0Text := renderL0(l0ctx)
	assembled.L0 = l0Text
	assembled.L0Tokens = estimateTokens(l0Text)
	assembled.Sources = append(assembled.Sources, ContextSource{
		Path:   ".system/l0",
		Title:  "L0 Context Cache",
		Genre:  "index",
		Tier:   "l0",
		Score:  1.0,
		Tokens: assembled.L0Tokens,
		Reason: "always_loaded",
	})

	// ── Step 2: Hybrid search for L1 candidates ───────────────────────────────

	type scoredCandidate struct {
		result    HybridResult
		composite float64
		content   string // loaded on demand for L1 insertion
	}

	var candidates []scoredCandidate

	if dbPath != "" && query != "" {
		hybridCfg := DefaultHybridConfig()
		limit := cfg.MaxSources * 3
		if limit <= 0 {
			limit = 60
		}
		hybridResults, searchErr := HybridSearch(dbPath, query, nil, limit, &hybridCfg)
		if searchErr == nil && len(hybridResults) > 0 {
			for _, hr := range hybridResults {
				importance, ok := genreImportance[strings.ToLower(hr.Genre)]
				if !ok {
					importance = defaultGenreImportance
				}
				composite := hr.Score*0.6 + importance*0.2 + hr.TemporalScore*0.2
				candidates = append(candidates, scoredCandidate{
					result:    hr,
					composite: composite,
				})
			}
		}
		// On search failure or empty results, fall through to filesystem-only mode.
	}

	// Filesystem-only fallback: when no DB is configured or search returned nothing,
	// build candidates from node summaries with flat relevance scoring.
	if len(candidates) == 0 && nodesRoot != "" {
		nodes, listErr := ListNodes(nodesRoot)
		if listErr == nil {
			lowerQuery := strings.ToLower(query)
			for _, n := range nodes {
				// Simple keyword match for relevance when FTS is unavailable.
				score := 0.0
				combined := strings.ToLower(n.Name + " " + n.Slug + " " + n.ContextMD)
				queryWords := strings.Fields(lowerQuery)
				for _, w := range queryWords {
					if strings.Contains(combined, w) {
						score += 1.0 / float64(max(1, len(queryWords)))
					}
				}
				if score == 0 && len(nodes) <= cfg.MaxSources {
					// No match but small corpus — include everything at low score.
					score = 0.05
				}
				if score < cfg.L1Threshold {
					continue
				}
				candidates = append(candidates, scoredCandidate{
					result: HybridResult{
						Path:  n.Slug + "/context.md",
						Title: n.Name,
						Genre: "note",
						Score: score,
						Node:  n.Slug,
					},
					composite: score,
					content:   n.ContextMD,
				})
			}
		}
	}

	// ── Step 3: Sort candidates by composite score descending ─────────────────

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].composite > candidates[j].composite
	})

	// ── Step 4: Fill L1 within budget ─────────────────────────────────────────

	var l1Parts []string
	l1TokensUsed := 0
	l1SourceCount := 0
	maxL1Sources := cfg.MaxSources
	if maxL1Sources <= 0 {
		maxL1Sources = 20
	}

	for i := range candidates {
		if l1SourceCount >= maxL1Sources {
			break
		}
		cand := &candidates[i]

		if cand.composite < cfg.L1Threshold {
			// Sorted descending — once we're below threshold, everything after is too.
			break
		}

		// Load content if not already fetched (DB path candidates).
		if cand.content == "" {
			// Try to load via LoadTiered — extract slug and filePath from result path.
			slug, filePath := splitNodePath(cand.result.Path, cand.result.Node)
			if slug != "" && filePath != "" {
				loaded, loadErr := LoadTiered(nodesRoot, slug, filePath, TierL1)
				if loadErr == nil {
					cand.content = loaded
				}
			}
			// If load failed, fall back to the abstract from the search result.
			if cand.content == "" {
				cand.content = cand.result.Abstract
			}
		}

		if cand.content == "" {
			continue
		}

		// Apply receiver genre conversion if a preferred genre is configured.
		displayContent := cand.content
		if receiver.PreferredGenre != "" && receiver.PreferredGenre != cand.result.Genre {
			displayContent = Compose(cand.content, receiver.PreferredGenre)
		}

		tokens := estimateTokens(displayContent)
		if tokens == 0 {
			continue
		}

		// Respect L1 budget — include whole documents only, never truncate.
		if l1TokensUsed+tokens > cfg.L1Budget {
			continue
		}

		l1Parts = append(l1Parts, formatL1Section(cand.result.Title, cand.result.Path, displayContent))
		l1TokensUsed += tokens
		l1SourceCount++

		assembled.Sources = append(assembled.Sources, ContextSource{
			Path:   cand.result.Path,
			Title:  cand.result.Title,
			Genre:  cand.result.Genre,
			Tier:   "l1",
			Score:  cand.composite,
			Tokens: tokens,
			Reason: reasonForInclusion(cand.result),
		})
	}

	assembled.L1 = strings.Join(l1Parts, "\n\n")
	assembled.L1Tokens = l1TokensUsed

	// ── Step 5: L2 expansion (conditional) ───────────────────────────────────

	if requiresL2(query, receiver) && len(assembled.Sources) > 1 {
		var l2Parts []string
		l2TokensUsed := 0

		// Take top 3 L1 entries (already scored highest; skip L0 source at index 0).
		l2Candidates := assembled.Sources[1:]
		if len(l2Candidates) > 3 {
			l2Candidates = l2Candidates[:3]
		}

		for _, src := range l2Candidates {
			if src.Tier != "l1" {
				continue
			}
			slug, filePath := splitNodePath(src.Path, "")
			if slug == "" || filePath == "" {
				continue
			}
			full, loadErr := LoadTiered(nodesRoot, slug, filePath, TierFull)
			if loadErr != nil {
				continue
			}

			tokens := estimateTokens(full)
			if tokens == 0 {
				continue
			}
			if l2TokensUsed+tokens > cfg.L2Budget {
				continue
			}

			l2Parts = append(l2Parts, formatL2Section(src.Title, src.Path, full))
			l2TokensUsed += tokens

			assembled.Sources = append(assembled.Sources, ContextSource{
				Path:   src.Path,
				Title:  src.Title,
				Genre:  src.Genre,
				Tier:   "l2",
				Score:  src.Score,
				Tokens: tokens,
				Reason: "drill_down",
			})
		}

		assembled.L2 = strings.Join(l2Parts, "\n\n")
		assembled.L2Tokens = l2TokensUsed
	}

	// ── Step 7: Compute totals and return ─────────────────────────────────────

	assembled.TotalTokens = assembled.L0Tokens + assembled.L1Tokens + assembled.L2Tokens
	if receiver.TokenBudget > 0 {
		assembled.BudgetUsed = float64(assembled.TotalTokens) / float64(receiver.TokenBudget)
	}
	assembled.AssemblyTimeMs = time.Since(start).Milliseconds()

	return assembled, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

// splitNodePath extracts the slug and relative file path from a hybrid result
// path like "01-company/context.md" or "node-slug/signals/2026-04-01.md".
// When nodeHint is non-empty it is used as the slug directly (faster path).
func splitNodePath(path, nodeHint string) (slug, filePath string) {
	if path == "" {
		return "", ""
	}
	// Clean leading slash.
	path = strings.TrimPrefix(path, "/")

	if nodeHint != "" {
		// The hybrid result already carries the node slug — strip it from path.
		prefix := nodeHint + "/"
		if strings.HasPrefix(path, prefix) {
			return nodeHint, strings.TrimPrefix(path, prefix)
		}
		// path might already be just the file name.
		if !strings.Contains(path, "/") {
			return nodeHint, path
		}
	}

	// Fall back: split on first slash.
	idx := strings.Index(path, "/")
	if idx < 0 {
		// No slash — treat whole string as slug with no file path.
		return path, ""
	}
	return path[:idx], path[idx+1:]
}

// reasonForInclusion derives the human-readable reason a result was selected.
func reasonForInclusion(hr HybridResult) string {
	if hr.GraphBoost > 0.3 {
		return "entity_connected"
	}
	if hr.VectorScore > 0 {
		return "query_match"
	}
	if hr.BM25Score > 0 {
		return "query_match"
	}
	return "query_match"
}

// formatL1Section wraps a document's L1 content with a clear header for
// inclusion in the assembled L1 payload.
func formatL1Section(title, path, content string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### [%s] (%s)\n\n", title, path))
	sb.WriteString(content)
	return sb.String()
}

// formatL2Section wraps a document's full content for inclusion in L2.
func formatL2Section(title, path, content string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### FULL: [%s] (%s)\n\n", title, path))
	sb.WriteString(content)
	return sb.String()
}
