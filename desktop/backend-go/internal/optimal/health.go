package optimal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HealthReport is the full diagnostic report returned by Diagnose.
type HealthReport struct {
	Status string      `json:"status"` // "healthy", "degraded", or "error"
	Score  float64     `json:"score"`  // 0.0–1.0
	Checks []DiagCheck `json:"checks"`
	Stats  HealthStats `json:"stats"`
}

// DiagCheck is a single named diagnostic check result.
// Named DiagCheck to avoid conflict with the existing HealthCheck type in reader.go.
type DiagCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "warn", or "fail"
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// HealthStats holds aggregate counts across the knowledge base.
type HealthStats struct {
	Contexts  int `json:"contexts"`
	Signals   int `json:"signals"`
	Entities  int `json:"entities"`
	Edges     int `json:"edges"`
	Decisions int `json:"decisions"`
	Sessions  int `json:"sessions"`
}

// Diagnose runs 10 diagnostic checks against the OptimalOS filesystem and
// SQLite index. Returns a HealthReport with per-check results and aggregate score.
func Diagnose(nodesRoot, dbPath string) (*HealthReport, error) {
	report := &HealthReport{}

	// Collect aggregate stats first — used by multiple checks.
	stats, err := collectHealthStats(dbPath)
	if err != nil {
		// Non-fatal: continue with empty stats.
		stats = HealthStats{}
	}
	report.Stats = stats

	var checks []DiagCheck

	// ── Check 1: Orphaned contexts (contexts with 0 edges) ───────────────────
	checks = append(checks, checkOrphanedContexts(dbPath, stats))

	// ── Check 2: Stale signals (not modified in 30+ days) ────────────────────
	checks = append(checks, checkStaleSignals(nodesRoot))

	// ── Check 3: Missing cross-references ────────────────────────────────────
	checks = append(checks, checkMissingCrossRefs(dbPath))

	// ── Check 4: FTS drift (contexts count ≠ FTS row count) ──────────────────
	checks = append(checks, checkFTSDrift(dbPath))

	// ── Check 5: Entity merge candidates (case-insensitive duplicates) ────────
	checks = append(checks, checkEntityMergeCandidates(dbPath))

	// ── Check 6: Node imbalance (any node > 3× mean context count) ───────────
	checks = append(checks, checkNodeImbalance(dbPath))

	// ── Check 7: Duplicate titles (same title in same node) ──────────────────
	checks = append(checks, checkDuplicateTitles(dbPath))

	// ── Check 8: Broken supersedes references ────────────────────────────────
	checks = append(checks, checkBrokenReferences(dbPath))

	// ── Check 9: Quality distribution (% of contexts with sn_ratio < 0.4) ────
	checks = append(checks, checkQualityDistribution(dbPath))

	// ── Check 10: Nodes directory health ─────────────────────────────────────
	checks = append(checks, checkNodesDirHealth(nodesRoot))

	report.Checks = checks

	// ── Derive overall status and score ──────────────────────────────────────
	var passes, warns, fails int
	for _, c := range checks {
		switch c.Status {
		case "pass":
			passes++
		case "warn":
			warns++
		case "fail":
			fails++
		}
	}

	total := len(checks)
	report.Score = float64(passes) / float64(total)

	switch {
	case fails > 2:
		report.Status = "error"
	case fails > 0 || warns > 3:
		report.Status = "degraded"
	default:
		report.Status = "healthy"
	}

	return report, nil
}

// ── individual checks ─────────────────────────────────────────────────────────

func checkOrphanedContexts(dbPath string, stats HealthStats) DiagCheck {
	if dbPath == "" {
		return DiagCheck{Name: "orphaned_contexts", Status: "warn", Message: "database not configured"}
	}
	db, err := openDB(dbPath)
	if err != nil {
		return DiagCheck{Name: "orphaned_contexts", Status: "warn", Message: err.Error()}
	}
	defer db.Close()

	var count int
	_ = db.QueryRow(`
		SELECT COUNT(*) FROM contexts c
		WHERE NOT EXISTS (
			SELECT 1 FROM edges e
			WHERE e.source_id = c.id OR e.target_id = c.id
		)`).Scan(&count)

	if count == 0 {
		return DiagCheck{Name: "orphaned_contexts", Status: "pass",
			Message: "all contexts have at least one edge", Count: 0}
	}
	status := "warn"
	if stats.Contexts > 0 && float64(count)/float64(stats.Contexts) > 0.5 {
		status = "fail"
	}
	return DiagCheck{Name: "orphaned_contexts", Status: status,
		Message: fmt.Sprintf("%d contexts have no edges", count), Count: count}
}

func checkStaleSignals(nodesRoot string) DiagCheck {
	if nodesRoot == "" {
		return DiagCheck{Name: "stale_signals", Status: "warn", Message: "nodesRoot not configured"}
	}
	cutoff := time.Now().AddDate(0, 0, -30)
	var stale int

	entries, err := os.ReadDir(nodesRoot)
	if err != nil {
		return DiagCheck{Name: "stale_signals", Status: "warn",
			Message: fmt.Sprintf("cannot read nodes dir: %v", err)}
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sigDir := filepath.Join(nodesRoot, entry.Name(), "signals")
		sigs, err := os.ReadDir(sigDir)
		if err != nil {
			continue
		}
		for _, sig := range sigs {
			if !strings.HasSuffix(sig.Name(), ".md") {
				continue
			}
			info, err := sig.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				stale++
			}
		}
	}

	if stale == 0 {
		return DiagCheck{Name: "stale_signals", Status: "pass",
			Message: "all signals modified within 30 days"}
	}
	status := "warn"
	if stale > 50 {
		status = "fail"
	}
	return DiagCheck{Name: "stale_signals", Status: status,
		Message: fmt.Sprintf("%d signals not modified in 30+ days", stale), Count: stale}
}

func checkMissingCrossRefs(dbPath string) DiagCheck {
	if dbPath == "" {
		return DiagCheck{Name: "missing_cross_refs", Status: "warn", Message: "database not configured"}
	}
	db, err := openDB(dbPath)
	if err != nil {
		return DiagCheck{Name: "missing_cross_refs", Status: "warn", Message: err.Error()}
	}
	defer db.Close()

	// Count entities mentioned in content but with no outgoing edge.
	var count int
	_ = db.QueryRow(`
		SELECT COUNT(DISTINCT e.name)
		FROM entities e
		WHERE NOT EXISTS (
			SELECT 1 FROM edges ed WHERE ed.source_id = e.context_id
		)`).Scan(&count)

	if count == 0 {
		return DiagCheck{Name: "missing_cross_refs", Status: "pass",
			Message: "all mentioned entities have outgoing references"}
	}
	status := "warn"
	if count > 20 {
		status = "fail"
	}
	return DiagCheck{Name: "missing_cross_refs", Status: status,
		Message: fmt.Sprintf("%d entities have no outgoing cross-reference edges", count), Count: count}
}

func checkFTSDrift(dbPath string) DiagCheck {
	if dbPath == "" {
		return DiagCheck{Name: "fts_drift", Status: "warn", Message: "database not configured"}
	}
	db, err := openDB(dbPath)
	if err != nil {
		return DiagCheck{Name: "fts_drift", Status: "warn", Message: err.Error()}
	}
	defer db.Close()

	var contextCount, ftsCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM contexts").Scan(&contextCount)

	// contexts_fts is a content FTS5 table — its row count should match contexts.
	err = db.QueryRow("SELECT COUNT(*) FROM contexts_fts").Scan(&ftsCount)
	if err != nil {
		return DiagCheck{Name: "fts_drift", Status: "warn",
			Message: fmt.Sprintf("cannot count FTS rows: %v", err)}
	}

	diff := contextCount - ftsCount
	if diff < 0 {
		diff = -diff
	}
	if diff == 0 {
		return DiagCheck{Name: "fts_drift", Status: "pass",
			Message: fmt.Sprintf("FTS in sync (%d rows)", ftsCount)}
	}
	status := "warn"
	if diff > 10 {
		status = "fail"
	}
	return DiagCheck{Name: "fts_drift", Status: status,
		Message: fmt.Sprintf("FTS has %d rows, contexts has %d (drift: %d)", ftsCount, contextCount, diff),
		Count:   diff}
}

func checkEntityMergeCandidates(dbPath string) DiagCheck {
	if dbPath == "" {
		return DiagCheck{Name: "entity_merge_candidates", Status: "warn", Message: "database not configured"}
	}
	db, err := openDB(dbPath)
	if err != nil {
		return DiagCheck{Name: "entity_merge_candidates", Status: "warn", Message: err.Error()}
	}
	defer db.Close()

	rows, err := db.Query(`SELECT LOWER(name), COUNT(DISTINCT name) FROM entities GROUP BY LOWER(name) HAVING COUNT(DISTINCT name) > 1`)
	if err != nil {
		return DiagCheck{Name: "entity_merge_candidates", Status: "warn",
			Message: fmt.Sprintf("query failed: %v", err)}
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var lower string
		var variants int
		if err := rows.Scan(&lower, &variants); err == nil {
			count += variants - 1 // each extra variant is a merge candidate
		}
	}

	if count == 0 {
		return DiagCheck{Name: "entity_merge_candidates", Status: "pass",
			Message: "no case-variant duplicate entity names found"}
	}
	return DiagCheck{Name: "entity_merge_candidates", Status: "warn",
		Message: fmt.Sprintf("%d entity name variants could be merged", count), Count: count}
}

func checkNodeImbalance(dbPath string) DiagCheck {
	if dbPath == "" {
		return DiagCheck{Name: "node_imbalance", Status: "warn", Message: "database not configured"}
	}
	db, err := openDB(dbPath)
	if err != nil {
		return DiagCheck{Name: "node_imbalance", Status: "warn", Message: err.Error()}
	}
	defer db.Close()

	rows, err := db.Query(`SELECT node, COUNT(*) AS cnt FROM contexts GROUP BY node`)
	if err != nil {
		return DiagCheck{Name: "node_imbalance", Status: "warn",
			Message: fmt.Sprintf("query failed: %v", err)}
	}
	defer rows.Close()

	counts := map[string]int{}
	var total, nodeCount int
	for rows.Next() {
		var node string
		var cnt int
		if err := rows.Scan(&node, &cnt); err == nil {
			counts[node] = cnt
			total += cnt
			nodeCount++
		}
	}

	if nodeCount == 0 {
		return DiagCheck{Name: "node_imbalance", Status: "pass", Message: "no nodes with contexts"}
	}

	mean := float64(total) / float64(nodeCount)
	threshold := mean * 3
	var overloaded int
	for _, cnt := range counts {
		if float64(cnt) > threshold {
			overloaded++
		}
	}

	if overloaded == 0 {
		return DiagCheck{Name: "node_imbalance", Status: "pass",
			Message: fmt.Sprintf("node distribution balanced (mean %.0f contexts)", mean)}
	}
	return DiagCheck{Name: "node_imbalance", Status: "warn",
		Message: fmt.Sprintf("%d nodes exceed 3× mean (%0.f) context count", overloaded, mean),
		Count:   overloaded}
}

func checkDuplicateTitles(dbPath string) DiagCheck {
	if dbPath == "" {
		return DiagCheck{Name: "duplicate_titles", Status: "warn", Message: "database not configured"}
	}
	db, err := openDB(dbPath)
	if err != nil {
		return DiagCheck{Name: "duplicate_titles", Status: "warn", Message: err.Error()}
	}
	defer db.Close()

	var count int
	_ = db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT node, title, COUNT(*) AS cnt
			FROM contexts
			GROUP BY node, title
			HAVING cnt > 1
		)`).Scan(&count)

	if count == 0 {
		return DiagCheck{Name: "duplicate_titles", Status: "pass",
			Message: "no duplicate titles found within nodes"}
	}
	return DiagCheck{Name: "duplicate_titles", Status: "warn",
		Message: fmt.Sprintf("%d (node, title) pairs have duplicates", count), Count: count}
}

func checkBrokenReferences(dbPath string) DiagCheck {
	if dbPath == "" {
		return DiagCheck{Name: "broken_references", Status: "warn", Message: "database not configured"}
	}
	db, err := openDB(dbPath)
	if err != nil {
		return DiagCheck{Name: "broken_references", Status: "warn", Message: err.Error()}
	}
	defer db.Close()

	// Check edges where target_id does not exist as a context id.
	var count int
	_ = db.QueryRow(`
		SELECT COUNT(*) FROM edges e
		WHERE NOT EXISTS (
			SELECT 1 FROM contexts c WHERE c.id = e.target_id
		)`).Scan(&count)

	if count == 0 {
		return DiagCheck{Name: "broken_references", Status: "pass",
			Message: "all edge targets resolve to existing contexts"}
	}
	status := "warn"
	if count > 5 {
		status = "fail"
	}
	return DiagCheck{Name: "broken_references", Status: status,
		Message: fmt.Sprintf("%d edges point to nonexistent context IDs", count), Count: count}
}

func checkQualityDistribution(dbPath string) DiagCheck {
	if dbPath == "" {
		return DiagCheck{Name: "quality_distribution", Status: "warn", Message: "database not configured"}
	}
	db, err := openDB(dbPath)
	if err != nil {
		return DiagCheck{Name: "quality_distribution", Status: "warn", Message: err.Error()}
	}
	defer db.Close()

	// Check if sn_ratio column exists.
	row := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('contexts') WHERE name='sn_ratio'`)
	var hasCol int
	_ = row.Scan(&hasCol)
	if hasCol == 0 {
		return DiagCheck{Name: "quality_distribution", Status: "pass",
			Message: "sn_ratio column not present — quality tracking not enabled"}
	}

	var total, lowQuality int
	_ = db.QueryRow("SELECT COUNT(*) FROM contexts").Scan(&total)
	_ = db.QueryRow("SELECT COUNT(*) FROM contexts WHERE sn_ratio < 0.4").Scan(&lowQuality)

	if total == 0 {
		return DiagCheck{Name: "quality_distribution", Status: "pass", Message: "no contexts to evaluate"}
	}

	pct := float64(lowQuality) / float64(total) * 100
	if pct < 20 {
		return DiagCheck{Name: "quality_distribution", Status: "pass",
			Message: fmt.Sprintf("%.0f%% of contexts have sn_ratio < 0.4 (acceptable)", pct)}
	}
	status := "warn"
	if pct > 50 {
		status = "fail"
	}
	return DiagCheck{Name: "quality_distribution", Status: status,
		Message: fmt.Sprintf("%.0f%% of contexts have sn_ratio < 0.4 (%d of %d)", pct, lowQuality, total),
		Count:   lowQuality}
}

func checkNodesDirHealth(nodesRoot string) DiagCheck {
	if nodesRoot == "" {
		return DiagCheck{Name: "nodes_dir_health", Status: "fail", Message: "nodesRoot not configured"}
	}
	entries, err := os.ReadDir(nodesRoot)
	if err != nil {
		return DiagCheck{Name: "nodes_dir_health", Status: "fail",
			Message: fmt.Sprintf("cannot read nodes dir: %v", err)}
	}

	var missing int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		_, ok := parseNumericPrefix(entry.Name())
		if !ok {
			continue
		}
		ctxPath := filepath.Join(nodesRoot, entry.Name(), "context.md")
		if _, err := os.Stat(ctxPath); os.IsNotExist(err) {
			missing++
		}
	}

	if missing == 0 {
		return DiagCheck{Name: "nodes_dir_health", Status: "pass",
			Message: "all numbered nodes have context.md"}
	}
	status := "warn"
	if missing > 3 {
		status = "fail"
	}
	return DiagCheck{Name: "nodes_dir_health", Status: status,
		Message: fmt.Sprintf("%d numbered nodes missing context.md", missing), Count: missing}
}

// ── stats collector ───────────────────────────────────────────────────────────

func collectHealthStats(dbPath string) (HealthStats, error) {
	if dbPath == "" {
		return HealthStats{}, nil
	}
	db, err := openDB(dbPath)
	if err != nil {
		return HealthStats{}, err
	}
	defer db.Close()

	var s HealthStats
	_ = db.QueryRow("SELECT COUNT(*) FROM contexts").Scan(&s.Contexts)
	_ = db.QueryRow("SELECT COUNT(*) FROM entities").Scan(&s.Entities)
	_ = db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&s.Edges)

	// Optional tables — ignore errors if they don't exist.
	_ = db.QueryRow("SELECT COUNT(*) FROM contexts WHERE genre='decision_log'").Scan(&s.Decisions)
	_ = db.QueryRow("SELECT COUNT(*) FROM contexts WHERE node='09-new-stuff'").Scan(&s.Signals)

	return s, nil
}
