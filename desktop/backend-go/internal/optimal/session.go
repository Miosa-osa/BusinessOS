package optimal

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Session is a bounded conversation window. Sessions group messages for
// compaction, summary generation, and episodic memory.
type Session struct {
	ID           string `json:"id"`
	StartedAt    string `json:"started_at"`
	CommittedAt  string `json:"committed_at,omitempty"`
	Summary      string `json:"summary"`
	MessageCount int    `json:"message_count"`
	Metadata     string `json:"metadata,omitempty"`
}

// StartSession creates a new session record in SQLite and returns its ID.
// The ID is a UUID v4 string.
func StartSession(dbPath string) (string, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return "", fmt.Errorf("optimal/session: open db: %w", err)
	}
	defer db.Close()

	if err := ensureSessionTables(db); err != nil {
		return "", fmt.Errorf("optimal/session: ensure tables: %w", err)
	}

	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err = db.Exec(`
		INSERT INTO sessions (id, started_at, message_count, metadata)
		VALUES (?, ?, 0, '{}')`,
		id, now,
	)
	if err != nil {
		return "", fmt.Errorf("optimal/session: insert session: %w", err)
	}
	return id, nil
}

// AddMessage appends a message to an existing session.
// role should be one of: "user", "assistant", "system".
// It writes the message to session_messages and increments message_count on sessions.
func AddMessage(dbPath, sessionID, role, content string) error {
	if sessionID == "" {
		return fmt.Errorf("optimal/session: sessionID must not be empty")
	}
	if content == "" {
		return fmt.Errorf("optimal/session: content must not be empty")
	}

	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("optimal/session: open db: %w", err)
	}
	defer db.Close()

	if err := ensureSessionTables(db); err != nil {
		return fmt.Errorf("optimal/session: ensure tables: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Write to session_messages table (local, not part of core Elixir schema — used for detail).
	_, err = db.Exec(`
		INSERT INTO session_messages (session_id, role, content, created_at)
		VALUES (?, ?, ?, ?)`,
		sessionID, role, content, now,
	)
	if err != nil {
		return fmt.Errorf("optimal/session: insert message: %w", err)
	}

	// Increment message_count on the parent session.
	_, err = db.Exec(`
		UPDATE sessions SET message_count = message_count + 1 WHERE id = ?`,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("optimal/session: increment message_count: %w", err)
	}
	return nil
}

// CommitSession finalizes a session with a summary string and records
// the committed_at timestamp. Further messages may still be added, but
// committed sessions are treated as closed by the episodic memory layer.
func CommitSession(dbPath, sessionID, summary string) error {
	if sessionID == "" {
		return fmt.Errorf("optimal/session: sessionID must not be empty")
	}

	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("optimal/session: open db: %w", err)
	}
	defer db.Close()

	if err := ensureSessionTables(db); err != nil {
		return fmt.Errorf("optimal/session: ensure tables: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := db.Exec(`
		UPDATE sessions SET summary = ?, committed_at = ? WHERE id = ?`,
		summary, now, sessionID,
	)
	if err != nil {
		return fmt.Errorf("optimal/session: commit session: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("optimal/session: session %q not found", sessionID)
	}
	return nil
}

// GetSession retrieves a session by ID.
// Returns an error wrapping sql.ErrNoRows when the session does not exist.
func GetSession(dbPath, sessionID string) (*Session, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/session: open db: %w", err)
	}
	defer db.Close()

	if err := ensureSessionTables(db); err != nil {
		return nil, fmt.Errorf("optimal/session: ensure tables: %w", err)
	}

	var s Session
	var committedAt sql.NullString
	err = db.QueryRow(`
		SELECT id, started_at, committed_at, COALESCE(summary, ''),
		       message_count, COALESCE(metadata, '{}')
		FROM sessions
		WHERE id = ?`, sessionID,
	).Scan(&s.ID, &s.StartedAt, &committedAt, &s.Summary, &s.MessageCount, &s.Metadata)
	if err != nil {
		return nil, fmt.Errorf("optimal/session: get session: %w", err)
	}
	if committedAt.Valid {
		s.CommittedAt = committedAt.String
	}
	return &s, nil
}

// ListSessions returns the most recent limit sessions, ordered by start time
// descending.
func ListSessions(dbPath string, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = 20
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/session: open db: %w", err)
	}
	defer db.Close()

	if err := ensureSessionTables(db); err != nil {
		return nil, fmt.Errorf("optimal/session: ensure tables: %w", err)
	}

	rows, err := db.Query(`
		SELECT id, started_at, COALESCE(committed_at, ''), COALESCE(summary, ''),
		       message_count, COALESCE(metadata, '{}')
		FROM sessions
		ORDER BY started_at DESC
		LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("optimal/session: list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.StartedAt, &s.CommittedAt, &s.Summary, &s.MessageCount, &s.Metadata); err != nil {
			return nil, fmt.Errorf("optimal/session: scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/session: iterate sessions: %w", err)
	}
	return sessions, nil
}

// ensureSessionTables creates the sessions and session_messages tables if they
// do not exist. The sessions table matches the real Elixir engine schema with
// message_count and metadata columns. session_messages is a local extension for
// storing individual message records.
func ensureSessionTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id            TEXT PRIMARY KEY,
			started_at    TEXT    NOT NULL DEFAULT (datetime('now')),
			committed_at  TEXT,
			summary       TEXT    NOT NULL DEFAULT '',
			message_count INTEGER NOT NULL DEFAULT 0,
			metadata      TEXT    NOT NULL DEFAULT '{}'
		);
		CREATE TABLE IF NOT EXISTS session_messages (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT    NOT NULL REFERENCES sessions(id),
			role       TEXT    NOT NULL DEFAULT 'user',
			content    TEXT    NOT NULL,
			created_at TEXT    NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_session_messages_session_id
			ON session_messages(session_id);`)
	return err
}
