package optimal

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"
)

// HybridResult extends SearchResult with a per-mode scoring breakdown.
type HybridResult struct {
	Path          string  `json:"path"`
	Title         string  `json:"title"`
	Abstract      string  `json:"abstract"`
	Node          string  `json:"node"`
	Genre         string  `json:"genre"`
	Score         float64 `json:"score"`          // final RRF score
	BM25Score     float64 `json:"bm25_score"`     // FTS5 component
	TemporalScore float64 `json:"temporal_score"` // recency decay
	VectorScore   float64 `json:"vector_score"`   // cosine similarity (0 if no embeddings)
	GraphBoost    float64 `json:"graph_boost"`    // hub entity proximity boost
}

// HybridSearchConfig holds tunable parameters for hybrid search.
type HybridSearchConfig struct {
	MaxCandidates int     // how many candidates to pull from each mode (default 50)
	RRFConstant   int     // k constant in RRF formula (default 60)
	Alpha         float64 // BM25 weight (default 0.6)
	Beta          float64 // temporal weight (default 0.25)
	Gamma         float64 // vector weight (default 0.1)
	Delta         float64 // graph boost weight (default 0.05)
}

// DefaultHybridConfig returns sensible defaults.
func DefaultHybridConfig() HybridSearchConfig {
	return HybridSearchConfig{
		MaxCandidates: 50,
		RRFConstant:   60,
		Alpha:         0.6,
		Beta:          0.25,
		Gamma:         0.1,
		Delta:         0.05,
	}
}

// genreHalfLives maps document genres to half-life decay periods in hours.
// A document at exactly one half-life old decays to ~0.5 relevance.
var genreHalfLives = map[string]float64{
	"message":    168,   // 7 days
	"transcript": 168,   // 7 days
	"standup":    336,   // 14 days
	"status":     336,   // 14 days
	"note":       1440,  // 60 days
	"plan":       720,   // 30 days
	"spec":       4320,  // 180 days
	"decision":   8760,  // 365 days
	"adr":        8760,  // 365 days
	"pattern":    17520, // 730 days
}

const defaultHalfLife = 720.0 // 30 days, used for unknown genres

// temporalDecay computes an exponential decay score in [0, 1] based on age.
// score = exp(-ln(2) * ageHours / halfLife)
// A brand-new document returns 1.0; one half-life old returns ~0.5.
func temporalDecay(ts string, genre string) float64 {
	if ts == "" {
		return 0.5 // unknown age: neutral
	}

	// Accept common SQLite timestamp formats.
	var t time.Time
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if parsed, err := time.Parse(layout, ts); err == nil {
			t = parsed
			break
		}
	}
	if t.IsZero() {
		return 0.5
	}

	halfLife, ok := genreHalfLives[strings.ToLower(genre)]
	if !ok {
		halfLife = defaultHalfLife
	}

	ageHours := time.Since(t).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	return math.Exp(-0.693 * ageHours / halfLife)
}

// ftsCandidate is an internal representation of a BM25 search hit enriched
// with genre, node, and timestamp metadata needed for hybrid scoring.
type ftsCandidate struct {
	id       string
	path     string
	title    string
	node     string
	genre    string
	abstract string
	ts       string // indexed_at or created_at
}

// fetchFTSCandidates runs the FTS5 BM25 query and returns enriched candidates.
func fetchFTSCandidates(dbPath, query string, limit int) ([]ftsCandidate, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	const q = `
		SELECT c.id,
		       COALESCE(c.path, ''),
		       COALESCE(c.title, ''),
		       COALESCE(c.node, ''),
		       COALESCE(c.genre, ''),
		       snippet(contexts_fts, 2, '', '', '...', 30) AS abstract,
		       COALESCE(c.indexed_at, c.created_at, '') AS ts,
		       rank AS bm25_rank
		FROM contexts_fts f
		JOIN contexts c ON f.rowid = c.rowid
		WHERE contexts_fts MATCH ?
		ORDER BY rank
		LIMIT ?`

	rows, err := db.Query(q, query, limit)
	if err != nil {
		return nil, fmt.Errorf("optimal/hybrid: fts query: %w", err)
	}
	defer rows.Close()

	var candidates []ftsCandidate
	for rows.Next() {
		var c ftsCandidate
		var bm25Rank float64 // consumed but not stored; order is what matters
		if err := rows.Scan(&c.id, &c.path, &c.title, &c.node, &c.genre, &c.abstract, &c.ts, &bm25Rank); err != nil {
			return nil, fmt.Errorf("optimal/hybrid: scan fts row: %w", err)
		}
		// Synthesise path when the stored column is empty.
		if c.path == "" {
			c.path = c.node + "/" + c.title
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/hybrid: iterate fts rows: %w", err)
	}
	return candidates, nil
}

// hubEntitySet fetches hub entity names from the graph and returns them as a
// lowercase set for fast membership tests.
func hubEntitySet(dbPath string) map[string]struct{} {
	hubs, err := FindHubs(dbPath, 30)
	if err != nil {
		return nil
	}
	set := make(map[string]struct{}, len(hubs))
	for _, h := range hubs {
		set[strings.ToLower(h.Name)] = struct{}{}
	}
	return set
}

// graphBoostScore returns a boost in [0, 1] for a result whose path/title/node
// overlaps with known hub entities. Simple token-level matching.
func graphBoostScore(path, title, node string, hubSet map[string]struct{}) float64 {
	if len(hubSet) == 0 {
		return 0
	}
	// Tokenise the three text fields and count hub matches.
	combined := strings.ToLower(path + " " + title + " " + node)
	tokens := strings.FieldsFunc(combined, func(r rune) bool {
		return r == '/' || r == ' ' || r == '_' || r == '-'
	})
	if len(tokens) == 0 {
		return 0
	}
	var hits int
	for _, tok := range tokens {
		if _, ok := hubSet[tok]; ok {
			hits++
		}
	}
	if hits == 0 {
		return 0
	}
	// Cap the boost at 1.0; use diminishing returns via sqrt.
	return math.Min(1.0, math.Sqrt(float64(hits))/3.0)
}

// rrfScore computes the Reciprocal Rank Fusion contribution for one ranked list.
// score = weight / (k + rank)  where rank is 1-indexed.
func rrfScore(rank, k int, weight float64) float64 {
	return weight / float64(k+rank)
}

// HybridSearch runs multi-mode search with Reciprocal Rank Fusion fusion.
//
// Steps:
//  1. Run FTS5 BM25 search → ranked list of up to cfg.MaxCandidates results
//  2. Apply temporal decay to each result based on indexed_at / created_at
//  3. Optionally run vector search if queryEmbedding is non-nil
//  4. Optionally apply graph boost from hub entities
//  5. Merge all ranked lists using Reciprocal Rank Fusion (RRF)
//  6. Return the top limit results sorted by descending RRF score
//
// Graceful degradation: if vector search or graph boost fails, those modes
// are silently skipped and FTS5+temporal results are always returned.
func HybridSearch(dbPath string, query string, queryEmbedding []float32, limit int, cfg *HybridSearchConfig) ([]HybridResult, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("optimal/hybrid: dbPath must not be empty")
	}
	if query == "" {
		return nil, fmt.Errorf("optimal/hybrid: query must not be empty")
	}
	if limit <= 0 {
		limit = 10
	}
	if cfg == nil {
		defaults := DefaultHybridConfig()
		cfg = &defaults
	}

	// ── Step 0: Intent analysis + query expansion ─────────────────────────────

	intent := AnalyzeIntent(query)

	// Expand the FTS5 query with entity-based OR terms so that documents
	// mentioning a named entity from the query surface even when the raw
	// phrasing differs.
	expandedQuery := query
	if len(intent.Entities) > 0 {
		// Wrap each entity in double-quotes to treat it as a phrase in FTS5.
		quotedEntities := make([]string, 0, len(intent.Entities))
		for _, e := range intent.Entities {
			e = strings.TrimSpace(e)
			if e == "" {
				continue
			}
			quotedEntities = append(quotedEntities, `"`+strings.ReplaceAll(e, `"`, "")+`"`)
		}
		if len(quotedEntities) > 0 {
			expandedQuery = query + " OR " + strings.Join(quotedEntities, " OR ")
		}
	}

	// ── Step 1: FTS5 BM25 search ─────────────────────────────────────────────

	ftsCandidates, err := fetchFTSCandidates(dbPath, expandedQuery, cfg.MaxCandidates)
	if err != nil {
		// If query expansion produced an invalid FTS5 expression, fall back to
		// the original query so the search always returns results.
		ftsCandidates, err = fetchFTSCandidates(dbPath, query, cfg.MaxCandidates)
		if err != nil {
			return nil, fmt.Errorf("optimal/hybrid: fts search: %w", err)
		}
	}

	// Keyed accumulator for merging scores across modes.
	// Key is the normalised path (which uniquely identifies a document).
	type candidate struct {
		path          string
		title         string
		abstract      string
		node          string
		genre         string
		bm25Score     float64
		temporalScore float64
		vectorScore   float64
		graphBoost    float64
		rrfTotal      float64
	}
	pool := make(map[string]*candidate, len(ftsCandidates))

	// Build ranked BM25 list and apply temporal decay.
	// FTS5 returns rows already ordered by rank ascending (most relevant first).
	for i, fc := range ftsCandidates {
		decay := temporalDecay(fc.ts, fc.genre)
		rank := i + 1 // 1-indexed

		// BM25 RRF contribution (alpha weight).
		bm25RRF := rrfScore(rank, cfg.RRFConstant, cfg.Alpha)

		// Temporal RRF: treat temporal decay as a secondary ranked list.
		// We rank candidates by their decay score for this pass; since all
		// decay values are computed in a single pass we approximate by using
		// the same BM25 rank weighted by the decay value itself — a reasonable
		// proxy that avoids a second sort.
		temporalRRF := rrfScore(rank, cfg.RRFConstant, cfg.Beta) * decay

		pool[fc.path] = &candidate{
			path:          fc.path,
			title:         fc.title,
			abstract:      fc.abstract,
			node:          fc.node,
			genre:         fc.genre,
			bm25Score:     bm25RRF,
			temporalScore: decay,
			rrfTotal:      bm25RRF + temporalRRF,
		}
	}

	// ── Step 2b: Temporal filtering (optional) ───────────────────────────────
	// When the query carries a time range (e.g. "last week", "march 2025"),
	// down-rank documents whose timestamps fall outside that window by zeroing
	// their temporal score component. This is non-destructive: documents are
	// demoted rather than removed so the caller always gets `limit` results.
	if intent.TimeRange != "" {
		windowStart, windowEnd := parseTimeRange(intent.TimeRange)
		if !windowStart.IsZero() {
			for _, c := range pool {
				if c.path == "" {
					continue
				}
				// Re-locate the original timestamp from ftsCandidates.
				ts := candidateTS(ftsCandidates, c.path)
				if ts == "" {
					continue
				}
				inWindow := isInTimeWindow(ts, windowStart, windowEnd)
				if !inWindow {
					// Demote: remove temporal contribution from RRF total.
					demotedTemporal := rrfScore(1, cfg.RRFConstant, cfg.Beta) * c.temporalScore
					c.rrfTotal -= demotedTemporal
					c.temporalScore = 0
				}
			}
		}
	}

	// ── Step 3: Vector search (optional) ─────────────────────────────────────

	if len(queryEmbedding) > 0 {
		vecResults, vecErr := SearchByVector(dbPath, queryEmbedding, cfg.MaxCandidates)
		if vecErr == nil {
			for rank, vr := range vecResults {
				contribution := rrfScore(rank+1, cfg.RRFConstant, cfg.Gamma)
				if c, ok := pool[vr.Path]; ok {
					// Document already in pool from FTS5 — merge.
					c.vectorScore = vr.Score
					c.rrfTotal += contribution
				} else {
					// Vector-only hit: add to pool.
					pool[vr.Path] = &candidate{
						path:        vr.Path,
						title:       vr.Abstract, // SearchByVector stores title in Abstract
						abstract:    vr.Abstract,
						vectorScore: vr.Score,
						rrfTotal:    contribution,
					}
				}
			}
		}
		// Silently skip if vector search fails (graceful degradation).
	}

	// ── Step 4: Graph boost (optional) ───────────────────────────────────────

	hubSet := hubEntitySet(dbPath) // returns nil on failure → boost is 0
	if len(hubSet) > 0 {
		for _, c := range pool {
			boost := graphBoostScore(c.path, c.title, c.node, hubSet)
			if boost > 0 {
				c.graphBoost = boost
				// Graph boost is not a ranked list; apply as a direct additive
				// delta weighted by cfg.Delta.
				c.rrfTotal += cfg.Delta * boost
			}
		}
	}

	// ── Step 5: Sort and return top `limit` ──────────────────────────────────

	ranked := make([]*candidate, 0, len(pool))
	for _, c := range pool {
		ranked = append(ranked, c)
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].rrfTotal > ranked[j].rrfTotal
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	results := make([]HybridResult, 0, len(ranked))
	for _, c := range ranked {
		results = append(results, HybridResult{
			Path:          c.path,
			Title:         c.title,
			Abstract:      c.abstract,
			Node:          c.node,
			Genre:         c.genre,
			Score:         c.rrfTotal,
			BM25Score:     c.bm25Score,
			TemporalScore: c.temporalScore,
			VectorScore:   c.vectorScore,
			GraphBoost:    c.graphBoost,
		})
	}
	return results, nil
}

// ── temporal filtering helpers ────────────────────────────────────────────────

// candidateTS looks up the timestamp for a given path in the fts candidate list.
func candidateTS(candidates []ftsCandidate, path string) string {
	for _, c := range candidates {
		if c.path == path {
			return c.ts
		}
	}
	return ""
}

// parseTimeRange converts a natural-language time phrase (as extracted by
// extractTimeRange) into a [start, end) time window. Returns zero times when
// the phrase cannot be parsed — callers must check windowStart.IsZero().
func parseTimeRange(phrase string) (start, end time.Time) {
	now := time.Now()
	phrase = strings.ToLower(strings.TrimSpace(phrase))

	switch {
	case phrase == "today":
		start = truncateDay(now)
		end = start.AddDate(0, 0, 1)
	case phrase == "yesterday":
		start = truncateDay(now.AddDate(0, 0, -1))
		end = start.AddDate(0, 0, 1)
	case phrase == "this week":
		// Week starts on Monday.
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = truncateDay(now.AddDate(0, 0, -(weekday - 1)))
		end = start.AddDate(0, 0, 7)
	case phrase == "this month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)
	case phrase == "this year":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(1, 0, 0)
	case strings.HasPrefix(phrase, "last "):
		rest := strings.TrimPrefix(phrase, "last ")
		switch rest {
		case "week":
			end = truncateDay(now.AddDate(0, 0, -int(now.Weekday())+1))
			start = end.AddDate(0, 0, -7)
		case "month":
			end = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			start = end.AddDate(0, -1, 0)
		case "year":
			end = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			start = end.AddDate(-1, 0, 0)
		default:
			// "last N days" pattern.
			var n int
			if _, err := fmt.Sscanf(rest, "%d days", &n); err == nil && n > 0 {
				start = truncateDay(now.AddDate(0, 0, -n))
				end = now
			}
		}
	default:
		// Try "month year" (e.g. "march 2025") or ISO date "YYYY-MM-DD".
		for _, layout := range []string{"january 2006", "february 2006", "march 2006",
			"april 2006", "may 2006", "june 2006", "july 2006", "august 2006",
			"september 2006", "october 2006", "november 2006", "december 2006"} {
			if t, err := time.Parse(layout, phrase); err == nil {
				start = t
				end = t.AddDate(0, 1, 0)
				return
			}
		}
		if t, err := time.Parse("2006-01-02", phrase); err == nil {
			start = t
			end = t.AddDate(0, 0, 1)
		}
	}
	return
}

// truncateDay returns t with hour/min/sec/nsec zeroed.
func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// isInTimeWindow returns true when the timestamp string falls within [start, end).
// Returns true on parse failure to err on the side of inclusion.
func isInTimeWindow(ts string, start, end time.Time) bool {
	var t time.Time
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if parsed, err := time.Parse(layout, ts); err == nil {
			t = parsed
			break
		}
	}
	if t.IsZero() {
		return true // unknown timestamp → include
	}
	return !t.Before(start) && t.Before(end)
}

// HybridSearch runs multi-mode search on the engine's configured database.
// If DBPath is unset the call returns nil, nil (graceful no-op).
// Embedding-based re-ranking is skipped; pass a non-nil queryEmbedding to the
// package-level HybridSearch function directly when vector ranking is needed.
func (c *EngineConfig) HybridSearch(ctx context.Context, query string, limit int) ([]HybridResult, error) {
	if c.DBPath == "" {
		return nil, nil
	}
	cfg := DefaultHybridConfig()
	results, err := HybridSearch(c.DBPath, query, nil, limit, &cfg)
	if err != nil {
		slog.WarnContext(ctx, "optimal: hybrid search failed", "query", query, "error", err)
		return nil, nil
	}
	return results, nil
}
