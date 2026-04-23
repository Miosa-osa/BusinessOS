package optimal

import (
	"os"
	"path/filepath"
	"testing"
)

// ── EnsureSchema ──────────────────────────────────────────────────────────────

func TestEnsureSchema_CreatesDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := EnsureSchema(dbPath)
	if err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	// File must exist after the call.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestEnsureSchema_AllTablesExist(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := EnsureSchema(dbPath); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("openDB failed: %v", err)
	}
	defer db.Close()

	tables := []string{
		"contexts",
		"entities",
		"edges",
		"observations",
		"sessions",
		"session_messages",
		"vectors",
		"decision_log",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestEnsureSchema_FTSVirtualTableExists(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := EnsureSchema(dbPath); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("openDB failed: %v", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='contexts_fts'",
	).Scan(&name)
	if err != nil {
		t.Errorf("FTS virtual table 'contexts_fts' not found: %v", err)
	}
}

func TestEnsureSchema_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// First call.
	if err := EnsureSchema(dbPath); err != nil {
		t.Fatalf("first EnsureSchema failed: %v", err)
	}

	// Second call on the same DB must not error (IF NOT EXISTS everywhere).
	if err := EnsureSchema(dbPath); err != nil {
		t.Fatalf("second EnsureSchema (idempotency check) failed: %v", err)
	}
}

func TestEnsureSchema_ContextsTableColumns(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := EnsureSchema(dbPath); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("openDB failed: %v", err)
	}
	defer db.Close()

	// PRAGMA table_info returns one row per column.
	rows, err := db.Query("PRAGMA table_info(contexts)")
	if err != nil {
		t.Fatalf("PRAGMA table_info failed: %v", err)
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType, notNull, pk string
		var dfltValue *string // nullable — columns without defaults return NULL
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		cols[name] = true
	}

	required := []string{
		"id", "uri", "type", "path", "title",
		"l0_abstract", "l1_overview", "content",
		"mode", "genre", "signal_type", "format", "structure",
		"node", "sn_ratio", "entities",
		"created_at", "modified_at", "valid_from",
		"routed_to", "metadata", "indexed_at",
	}
	for _, col := range required {
		if !cols[col] {
			t.Errorf("contexts table missing column %q", col)
		}
	}
}

func TestEnsureSchema_DecisionLogIndexesExist(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := EnsureSchema(dbPath); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("openDB failed: %v", err)
	}
	defer db.Close()

	indexes := []string{"idx_decision_log_target", "idx_decision_log_action"}
	for _, idx := range indexes {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx,
		).Scan(&name)
		if err != nil {
			t.Errorf("index %q not found: %v", idx, err)
		}
	}
}

// ── LogDecision ───────────────────────────────────────────────────────────────

func TestLogDecision_WritesRecord(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// EnsureSchema first so the table exists.
	if err := EnsureSchema(dbPath); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	err := LogDecision(dbPath, "test-agent", "upsert", "context", "ctx-001", "indexed from filesystem", "reindex run")
	if err != nil {
		t.Fatalf("LogDecision failed: %v", err)
	}

	// Verify the record was written.
	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("openDB failed: %v", err)
	}
	defer db.Close()

	var action, targetType, targetID string
	err = db.QueryRow(
		"SELECT action, target_type, target_id FROM decision_log WHERE target_id = ?", "ctx-001",
	).Scan(&action, &targetType, &targetID)
	if err != nil {
		t.Fatalf("decision_log record not found: %v", err)
	}
	if action != "upsert" {
		t.Errorf("action = %q, want %q", action, "upsert")
	}
	if targetType != "context" {
		t.Errorf("target_type = %q, want %q", targetType, "context")
	}
	if targetID != "ctx-001" {
		t.Errorf("target_id = %q, want %q", targetID, "ctx-001")
	}
}

func TestLogDecision_EmptyDBPathNoOp(t *testing.T) {
	// Empty dbPath should be a no-op and return nil.
	err := LogDecision("", "system", "skip", "context", "x", "reason", "why")
	if err != nil {
		t.Errorf("LogDecision with empty dbPath should return nil, got %v", err)
	}
}

func TestLogDecision_CreatesTableIfMissing(t *testing.T) {
	// LogDecision has its own CREATE TABLE IF NOT EXISTS guard — call it on a
	// fresh DB that has NOT had EnsureSchema run.
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "fresh.db")

	err := LogDecision(dbPath, "system", "create", "context", "fresh-001", "created", "bootstrap")
	if err != nil {
		t.Fatalf("LogDecision on fresh DB failed: %v", err)
	}

	// File should now exist.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("DB file not created by LogDecision")
	}
}

// ── applySchema (internal) ────────────────────────────────────────────────────

func TestApplySchema_InMemory(t *testing.T) {
	// Use an in-memory SQLite DB to test applySchema directly without I/O.
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("openDB(:memory:) failed: %v", err)
	}
	defer db.Close()

	if err := applySchema(db); err != nil {
		t.Fatalf("applySchema failed: %v", err)
	}

	// Quick check: insert into contexts and read back.
	_, err = db.Exec(`INSERT INTO contexts (id, title, content) VALUES (?, ?, ?)`,
		"test-id", "Test Title", "Test content")
	if err != nil {
		t.Fatalf("insert into contexts failed: %v", err)
	}

	var title string
	if err := db.QueryRow("SELECT title FROM contexts WHERE id=?", "test-id").Scan(&title); err != nil {
		t.Fatalf("select from contexts failed: %v", err)
	}
	if title != "Test Title" {
		t.Errorf("title = %q, want %q", title, "Test Title")
	}
}
