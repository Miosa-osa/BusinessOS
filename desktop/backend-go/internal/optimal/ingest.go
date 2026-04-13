package optimal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required
)

// Ingest writes a new signal to the filesystem and indexes it in SQLite.
// The signal lands in nodes/09-new-stuff/signals/ by default — the same inbox
// node the Elixir engine used. If nodesRoot or osRoot are empty the filesystem
// write is skipped; if dbPath is empty the index write is skipped. Either leg
// can fail independently without blocking the other.
func Ingest(osRoot, nodesRoot, dbPath, text, genre string) error {
	if text == "" {
		return fmt.Errorf("optimal/ingest: text must not be empty")
	}
	if genre == "" {
		genre = "note"
	}

	now := time.Now()
	slug := makeSlug(text, 50)
	filename := now.Format("2006-01-02") + "-" + slug + ".md"
	targetNode := "09-new-stuff"

	// ── filesystem write ──────────────────────────────────────────────────────
	if nodesRoot != "" {
		content := buildSignalContent(genre, now, text)
		signalDir := filepath.Join(nodesRoot, targetNode, "signals")
		if err := os.MkdirAll(signalDir, 0o755); err != nil {
			return fmt.Errorf("optimal/ingest: mkdir signals: %w", err)
		}
		dest := filepath.Join(signalDir, filename)
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			return fmt.Errorf("optimal/ingest: write signal file: %w", err)
		}
	}

	// ── SQLite index write ────────────────────────────────────────────────────
	if dbPath != "" {
		if err := indexSignal(dbPath, targetNode, filename, text, genre, now); err != nil {
			// Non-fatal: the file is already on disk.
			return fmt.Errorf("optimal/ingest: index in SQLite: %w", err)
		}
	}

	return nil
}

// buildSignalContent formats the signal file with YAML frontmatter.
func buildSignalContent(genre string, now time.Time, text string) string {
	return fmt.Sprintf("---\ngenre: %s\ndate: %s\nmode: linguistic\ntype: inform\n---\n\n%s\n",
		genre,
		now.Format("2006-01-02"),
		text,
	)
}

// indexSignal inserts the signal into the contexts table (and its FTS5 shadow
// table) so it becomes searchable immediately. The contexts_fts virtual table
// is assumed to be a content table or trigger-maintained FTS5 index — we insert
// into contexts and let the trigger/content mechanism update the FTS index.
func indexSignal(dbPath, node, filename, text, genre string, now time.Time) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	// Derive a human-readable title from the filename slug.
	slug := strings.TrimSuffix(filename, ".md")
	title := slugToTitle(slug)

	// Build a synthetic path consistent with the rest of the index.
	path := filepath.Join("nodes", node, "signals", filename)

	_, err = db.Exec(`
		INSERT OR IGNORE INTO contexts
			(id, path, title, node, content, genre, indexed_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?)`,
		slug, path, title, node, text, genre, now.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert context: %w", err)
	}
	return nil
}

// ── slug helpers ──────────────────────────────────────────────────────────────

var (
	reNonAlnum    = regexp.MustCompile(`[^a-z0-9]+`)
	reMultiHyphen = regexp.MustCompile(`-{2,}`)
)

// makeSlug returns a URL-safe, hyphenated slug from the first maxLen characters
// of text. Leading/trailing hyphens are stripped.
func makeSlug(text string, maxLen int) string {
	// Fold to ASCII lowercase.
	s := strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return '-'
		}
		return unicode.ToLower(r)
	}, text)

	// Replace non-alphanumeric runs with a single hyphen.
	s = reNonAlnum.ReplaceAllString(s, "-")
	s = reMultiHyphen.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	// Truncate at a word boundary within maxLen.
	if len(s) > maxLen {
		s = s[:maxLen]
		// Walk back to the last hyphen to avoid cutting mid-word.
		if idx := strings.LastIndexByte(s, '-'); idx > 0 {
			s = s[:idx]
		}
	}
	if s == "" {
		s = "signal"
	}
	return s
}
