package optimal

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// StaleContext describes a context file that has not been updated recently
// relative to how active its topic appears in recent signals.
type StaleContext struct {
	Path      string  `json:"path"`
	Title     string  `json:"title"`
	Node      string  `json:"node"`
	Staleness float64 `json:"staleness"` // 0.0 = fresh, 1.0 = very stale
	LastMod   string  `json:"last_modified"`
	DaysSince int     `json:"days_since_modified"`
}

// defaultMaxDays is used when the caller passes maxDays <= 0.
const defaultMaxDays = 30

// Reweave finds stale contexts related to topic by:
//  1. Searching the FTS index for contexts matching topic.
//  2. For each hit, reading its indexed_at timestamp.
//  3. Computing staleness = min(1.0, days_since / maxDays).
//  4. Sorting by staleness descending (most stale first).
func Reweave(dbPath, topic string, maxDays int) ([]StaleContext, error) {
	if topic == "" {
		return nil, fmt.Errorf("optimal/reweave: topic must not be empty")
	}
	if maxDays <= 0 {
		maxDays = defaultMaxDays
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/reweave: open db: %w", err)
	}
	defer db.Close()

	// Search for contexts matching topic, including their last-modified timestamp.
	// We join against the main contexts table to get indexed_at.
	const q = `
		SELECT c.path, c.title, c.node, COALESCE(c.indexed_at, '')
		FROM contexts_fts f
		JOIN contexts c ON f.rowid = c.rowid
		WHERE contexts_fts MATCH ?
		ORDER BY rank
		LIMIT 50`

	rows, err := db.Query(q, topic)
	if err != nil {
		// If FTS table does not exist yet, fall back to a plain LIKE search.
		rows, err = db.Query(`
			SELECT path, title, node, COALESCE(indexed_at, '')
			FROM contexts
			WHERE content LIKE ? OR title LIKE ?
			LIMIT 50`,
			"%"+topic+"%", "%"+topic+"%",
		)
		if err != nil {
			return nil, fmt.Errorf("optimal/reweave: search contexts: %w", err)
		}
	}
	defer rows.Close()

	now := time.Now().UTC()
	var results []StaleContext

	for rows.Next() {
		var path, title, node, indexedAt string
		if err := rows.Scan(&path, &title, &node, &indexedAt); err != nil {
			return nil, fmt.Errorf("optimal/reweave: scan row: %w", err)
		}

		daysSince, lastMod := parseStaleness(indexedAt, now)
		staleness := computeStaleness(daysSince, maxDays)

		results = append(results, StaleContext{
			Path:      path,
			Title:     title,
			Node:      node,
			Staleness: staleness,
			LastMod:   lastMod,
			DaysSince: daysSince,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/reweave: iterate rows: %w", err)
	}

	// Sort most stale first.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Staleness > results[j].Staleness
	})

	return results, nil
}

// parseStaleness parses an RFC3339 timestamp and returns (daysSince, formatted).
// On parse failure it returns (maxDays, "unknown") so unknown timestamps rank
// as maximally stale.
func parseStaleness(ts string, now time.Time) (daysSince int, formatted string) {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return defaultMaxDays, "unknown"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Try date-only format used in some older records.
		t, err = time.Parse("2006-01-02", ts[:min10(len(ts), 10)])
		if err != nil {
			return defaultMaxDays, "unknown"
		}
	}
	days := int(now.Sub(t).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return days, t.Format("2006-01-02")
}

// computeStaleness returns a 0.0–1.0 staleness score.
// staleness = min(1.0, daysSince / maxDays)
func computeStaleness(daysSince, maxDays int) float64 {
	if maxDays <= 0 {
		return 1.0
	}
	s := float64(daysSince) / float64(maxDays)
	if s > 1.0 {
		s = 1.0
	}
	return s
}

// min10 returns the smaller of a and b (avoids importing math for a trivial clamp).
func min10(a, b int) int {
	if a < b {
		return a
	}
	return b
}
