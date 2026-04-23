package optimal

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required
)

// EnsureSchema opens the SQLite database at dbPath, enables WAL mode, and
// creates all tables, virtual tables, triggers, and indexes required by the
// OptimalOS engine. It is idempotent — every statement uses IF NOT EXISTS so
// it is safe to call on every startup.
//
// The caller is responsible for supplying a valid, writable dbPath.
// Returns nil on success or the first error encountered.
func EnsureSchema(dbPath string) error {
	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("optimal/schema: open db: %w", err)
	}
	defer db.Close()

	return applySchema(db)
}

// applySchema runs all DDL statements against an already-open *sql.DB.
// Separated from EnsureSchema so it can be tested with an in-memory database.
func applySchema(db *sql.DB) error {
	// WAL mode — enables concurrent readers while a writer is active.
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return fmt.Errorf("optimal/schema: set WAL mode: %w", err)
	}

	stmts := []string{
		// ── contexts ─────────────────────────────────────────────────────────
		// Core signal storage: 22 columns matching ingest.go indexSignalFull.
		`CREATE TABLE IF NOT EXISTS contexts (
			id          TEXT PRIMARY KEY,
			uri         TEXT NOT NULL DEFAULT '',
			type        TEXT NOT NULL DEFAULT 'signal',
			path        TEXT NOT NULL DEFAULT '',
			title       TEXT NOT NULL DEFAULT '',
			l0_abstract TEXT NOT NULL DEFAULT '',
			l1_overview TEXT NOT NULL DEFAULT '',
			content     TEXT NOT NULL DEFAULT '',
			mode        TEXT NOT NULL DEFAULT 'linguistic',
			genre       TEXT NOT NULL DEFAULT 'note',
			signal_type TEXT NOT NULL DEFAULT 'inform',
			format      TEXT NOT NULL DEFAULT 'markdown',
			structure   TEXT NOT NULL DEFAULT '',
			node        TEXT NOT NULL DEFAULT '',
			sn_ratio    REAL NOT NULL DEFAULT 0.5,
			entities    TEXT NOT NULL DEFAULT '[]',
			created_at  TEXT NOT NULL DEFAULT '',
			modified_at TEXT NOT NULL DEFAULT '',
			valid_from  TEXT NOT NULL DEFAULT '',
			routed_to   TEXT NOT NULL DEFAULT '[]',
			metadata    TEXT NOT NULL DEFAULT '{}',
			indexed_at  TEXT NOT NULL DEFAULT ''
		)`,

		// ── FTS5 virtual table ────────────────────────────────────────────────
		// Full-text search index over title, abstracts, and content.
		// content= keeps the FTS index in sync with the contexts table.
		`CREATE VIRTUAL TABLE IF NOT EXISTS contexts_fts USING fts5(
			title, l0_abstract, l1_overview, content,
			content='contexts',
			content_rowid='rowid',
			tokenize='porter unicode61'
		)`,

		// ── FTS5 auto-sync triggers ───────────────────────────────────────────
		`CREATE TRIGGER IF NOT EXISTS contexts_ai AFTER INSERT ON contexts BEGIN
			INSERT INTO contexts_fts(rowid, title, l0_abstract, l1_overview, content)
			VALUES (new.rowid, new.title, new.l0_abstract, new.l1_overview, new.content);
		END`,

		`CREATE TRIGGER IF NOT EXISTS contexts_ad AFTER DELETE ON contexts BEGIN
			INSERT INTO contexts_fts(contexts_fts, rowid, title, l0_abstract, l1_overview, content)
			VALUES ('delete', old.rowid, old.title, old.l0_abstract, old.l1_overview, old.content);
		END`,

		`CREATE TRIGGER IF NOT EXISTS contexts_au AFTER UPDATE ON contexts BEGIN
			INSERT INTO contexts_fts(contexts_fts, rowid, title, l0_abstract, l1_overview, content)
			VALUES ('delete', old.rowid, old.title, old.l0_abstract, old.l1_overview, old.content);
			INSERT INTO contexts_fts(rowid, title, l0_abstract, l1_overview, content)
			VALUES (new.rowid, new.title, new.l0_abstract, new.l1_overview, new.content);
		END`,

		// ── entities ─────────────────────────────────────────────────────────
		// Named entities extracted from signals (ingest.go insertEntity).
		`CREATE TABLE IF NOT EXISTS entities (
			context_id TEXT NOT NULL,
			name       TEXT NOT NULL,
			type       TEXT NOT NULL DEFAULT 'person',
			UNIQUE(context_id, name)
		)`,

		// ── edges ────────────────────────────────────────────────────────────
		// Directed relationships between signals/nodes (ingest.go insertEdge,
		// graph.go queryEdges).
		`CREATE TABLE IF NOT EXISTS edges (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id TEXT    NOT NULL,
			target_id TEXT    NOT NULL,
			relation  TEXT    NOT NULL DEFAULT 'cross_ref',
			weight    REAL    NOT NULL DEFAULT 1.0,
			UNIQUE(source_id, target_id, relation)
		)`,

		// ── observations ─────────────────────────────────────────────────────
		// Friction patterns, lessons, and behavioral notes (remember.go).
		// remember.go also has its own ensureObservationsTable guard — this
		// definition is intentionally identical so both paths converge.
		`CREATE TABLE IF NOT EXISTS observations (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			category   TEXT NOT NULL DEFAULT 'pattern',
			content    TEXT NOT NULL,
			confidence REAL NOT NULL DEFAULT 0.5,
			source     TEXT NOT NULL DEFAULT 'explicit',
			created_at TEXT NOT NULL
		)`,

		// ── sessions ─────────────────────────────────────────────────────────
		// Bounded conversation windows (session.go).
		// Schema matches session.go ensureSessionTables exactly.
		`CREATE TABLE IF NOT EXISTS sessions (
			id            TEXT    PRIMARY KEY,
			started_at    TEXT    NOT NULL DEFAULT (datetime('now')),
			committed_at  TEXT,
			summary       TEXT    NOT NULL DEFAULT '',
			message_count INTEGER NOT NULL DEFAULT 0,
			metadata      TEXT    NOT NULL DEFAULT '{}'
		)`,

		// session_messages is a local extension used by session.go for per-message
		// storage — not part of the Elixir schema but required by this Go engine.
		`CREATE TABLE IF NOT EXISTS session_messages (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT    NOT NULL REFERENCES sessions(id),
			role       TEXT    NOT NULL DEFAULT 'user',
			content    TEXT    NOT NULL,
			created_at TEXT    NOT NULL
		)`,

		// ── vectors ──────────────────────────────────────────────────────────
		// IEEE 754 little-endian float32 embeddings (vector.go).
		// Includes the dimensions column used by vector.go StoreEmbedding.
		`CREATE TABLE IF NOT EXISTS vectors (
			context_id TEXT    PRIMARY KEY,
			embedding  BLOB    NOT NULL,
			model      TEXT    NOT NULL DEFAULT 'nomic-embed-text',
			dimensions INTEGER NOT NULL DEFAULT 768,
			created_at TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,

		// ── decision_log ─────────────────────────────────────────────────────────
		// Append-only audit trail for every mutation (reindex.go, ingest.go UPSERT
		// path). Matches the OptimalOS Python/Elixir spec.
		`CREATE TABLE IF NOT EXISTS decision_log (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			actor        TEXT    NOT NULL DEFAULT 'system',
			action       TEXT    NOT NULL,
			target_type  TEXT    NOT NULL DEFAULT '',
			target_id    TEXT    NOT NULL DEFAULT '',
			what_changed TEXT    NOT NULL DEFAULT '',
			why          TEXT    NOT NULL DEFAULT '',
			created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE INDEX IF NOT EXISTS idx_decision_log_target ON decision_log(target_type, target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_decision_log_action ON decision_log(action)`,

		// ── indexes ──────────────────────────────────────────────────────────
		`CREATE INDEX IF NOT EXISTS idx_contexts_node    ON contexts(node)`,
		`CREATE INDEX IF NOT EXISTS idx_contexts_genre   ON contexts(genre)`,
		`CREATE INDEX IF NOT EXISTS idx_contexts_created ON contexts(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_entities_name    ON entities(name)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_source     ON edges(source_id)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_target     ON edges(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_session_messages_session_id ON session_messages(session_id)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("optimal/schema: exec DDL: %w\nstatement: %.120s", err, stmt)
		}
	}

	return nil
}

// LogDecision writes an append-only audit entry to the decision_log table.
// It ensures the table exists before inserting so it is safe to call before
// EnsureSchema completes (e.g. from the reindex hot path).
//
//   - actor:       who triggered the change ("system", "reindex", "ingest", or a user ID)
//   - action:      what happened ("upsert", "reject", "skip", "create", etc.)
//   - targetType:  the kind of object changed ("context", "entity", "edge")
//   - targetID:    the primary key or slug of the changed object
//   - whatChanged: brief description of the change ("indexed from filesystem", etc.)
//   - why:         reason / rationale for the change
func LogDecision(dbPath, actor, action, targetType, targetID, whatChanged, why string) error {
	if dbPath == "" {
		return nil
	}
	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("optimal/schema: LogDecision open db: %w", err)
	}
	defer db.Close()

	// Ensure the table exists (idempotent; handles the bootstrap case).
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS decision_log (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		actor        TEXT    NOT NULL DEFAULT 'system',
		action       TEXT    NOT NULL,
		target_type  TEXT    NOT NULL DEFAULT '',
		target_id    TEXT    NOT NULL DEFAULT '',
		what_changed TEXT    NOT NULL DEFAULT '',
		why          TEXT    NOT NULL DEFAULT '',
		created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("optimal/schema: LogDecision ensure table: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		INSERT INTO decision_log (actor, action, target_type, target_id, what_changed, why, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		actor, action, targetType, targetID, whatChanged, why, now,
	)
	if err != nil {
		return fmt.Errorf("optimal/schema: LogDecision insert: %w", err)
	}
	return nil
}
