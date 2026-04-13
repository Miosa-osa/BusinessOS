package optimal

import (
	"database/sql"
	"fmt"
	"time"
)

// Observation is a stored friction pattern, lesson, or behavioral note captured
// during sessions. Observations accumulate until they are synthesized by Rethink.
type Observation struct {
	ID         int64   `json:"id"`
	Category   string  `json:"category"` // process, people, tool, decision, pattern, friction
	Content    string  `json:"content"`
	Confidence float64 `json:"confidence"` // 0.0–1.0, set by source type
	Source     string  `json:"source"`     // explicit, contextual, mining
	CreatedAt  string  `json:"created_at"`
}

// validCategories is the canonical set of observation categories.
var validCategories = map[string]struct{}{
	"process":  {},
	"people":   {},
	"tool":     {},
	"decision": {},
	"pattern":  {},
	"friction": {},
}

// validSources maps source labels to their default confidence levels.
var validSources = map[string]float64{
	"explicit":   0.9,
	"contextual": 0.6,
	"mining":     0.4,
}

// Remember stores an observation in SQLite under the observations table.
// Category must be one of: process, people, tool, decision, pattern, friction.
// Source must be one of: explicit, contextual, mining.
// If category or source are empty/invalid, safe defaults are applied.
func Remember(dbPath, content, category, source string) error {
	if content == "" {
		return fmt.Errorf("optimal/remember: content must not be empty")
	}

	if _, ok := validCategories[category]; !ok {
		category = "pattern"
	}
	conf, ok := validSources[source]
	if !ok {
		source = "explicit"
		conf = validSources["explicit"]
	}

	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("optimal/remember: open db: %w", err)
	}
	defer db.Close()

	if err := ensureObservationsTable(db); err != nil {
		return fmt.Errorf("optimal/remember: ensure table: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		INSERT INTO observations (category, content, confidence, source, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		category, content, conf, source, now,
	)
	if err != nil {
		return fmt.Errorf("optimal/remember: insert observation: %w", err)
	}
	return nil
}

// ListObservations returns all observations, optionally filtered by category.
// Pass an empty string to return all categories.
func ListObservations(dbPath, category string) ([]Observation, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/remember: open db: %w", err)
	}
	defer db.Close()

	if err := ensureObservationsTable(db); err != nil {
		return nil, fmt.Errorf("optimal/remember: ensure table: %w", err)
	}

	var (
		rows *sql.Rows
		qErr error
	)
	if category == "" {
		rows, qErr = db.Query(`
			SELECT id, category, content, confidence, source, created_at
			FROM observations
			ORDER BY created_at DESC`)
	} else {
		rows, qErr = db.Query(`
			SELECT id, category, content, confidence, source, created_at
			FROM observations
			WHERE category = ?
			ORDER BY created_at DESC`, category)
	}
	if qErr != nil {
		return nil, fmt.Errorf("optimal/remember: query observations: %w", qErr)
	}
	defer rows.Close()

	var obs []Observation
	for rows.Next() {
		var o Observation
		if err := rows.Scan(&o.ID, &o.Category, &o.Content, &o.Confidence, &o.Source, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("optimal/remember: scan observation: %w", err)
		}
		obs = append(obs, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/remember: iterate observations: %w", err)
	}
	return obs, nil
}

// GetEscalations returns a map of category → observation count for categories
// that have 3 or more observations. These are ready for synthesis via Rethink.
func GetEscalations(dbPath string) (map[string]int, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/remember: open db: %w", err)
	}
	defer db.Close()

	if err := ensureObservationsTable(db); err != nil {
		return nil, fmt.Errorf("optimal/remember: ensure table: %w", err)
	}

	rows, err := db.Query(`
		SELECT category, COUNT(*) AS cnt
		FROM observations
		GROUP BY category
		HAVING cnt >= 3
		ORDER BY cnt DESC`)
	if err != nil {
		return nil, fmt.Errorf("optimal/remember: query escalations: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var cat string
		var cnt int
		if err := rows.Scan(&cat, &cnt); err != nil {
			return nil, fmt.Errorf("optimal/remember: scan escalation row: %w", err)
		}
		result[cat] = cnt
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/remember: iterate escalations: %w", err)
	}
	return result, nil
}

// ensureObservationsTable creates the observations table if it does not exist.
// Called before every operation so callers do not need to pre-run migrations.
func ensureObservationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS observations (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			category   TEXT    NOT NULL DEFAULT 'pattern',
			content    TEXT    NOT NULL,
			confidence REAL    NOT NULL DEFAULT 0.5,
			source     TEXT    NOT NULL DEFAULT 'explicit',
			created_at TEXT    NOT NULL
		)`)
	return err
}
