// Package services — pages_engine_sync.go
//
// Pushes BusinessOS Page edits into the OptimalEngine knowledge graph.
//
// When a user creates or updates a context (Page) in BusinessOS, this
// service:
//  1. Renders the Page as a markdown file at
//     OPTIMAL_NODES_ROOT/businessos/<slug>.md
//  2. Triggers the engine's reindex routine to update the SQLite index
//     that both the Elixir engine and the Go fallback read from.
//
// After the sync, the Pages → Knowledge Graph view in the BusinessOS
// frontend (which proxies /api/optimal/graph to the Elixir engine) shows
// the user's Page alongside the engine's seed corpus and any other
// ingested signals.
//
// The service is fire-and-forget from the handler's perspective — failure
// is logged but never returned to the caller. A failed sync is a degraded
// graph view, not a failed save.
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	optimal "github.com/Miosa-osa/OptimalEngine-go"
)

// businessosNodeSlug is the dedicated node folder for user-created Pages.
// Keeping them under a single namespace makes it easy to filter, purge,
// or rebuild without touching seed data in other nodes.
const businessosNodeSlug = "10-businessos"

// PagesEngineSync writes Pages to disk + reindexes the engine.
//
// Construction requires the same OPTIMAL_NODES_ROOT and OPTIMAL_DB_PATH
// the engine reads from. When either is empty, every method becomes a
// silent no-op so the service is safe to construct in environments
// without an engine wired (the BusinessOS Page save still succeeds).
type PagesEngineSync struct {
	nodesRoot string
	dbPath    string

	// reindex serializes concurrent reindex calls — the SQLite index
	// can't safely handle two writers at once, and we'd rather queue a
	// few hundred ms than corrupt the db.
	reindex sync.Mutex
}

// NewPagesEngineSync wires the service from environment variables. When
// either path is empty, the returned service is a no-op (safe to call,
// does nothing).
func NewPagesEngineSync() *PagesEngineSync {
	return &PagesEngineSync{
		nodesRoot: os.Getenv("OPTIMAL_NODES_ROOT"),
		dbPath:    os.Getenv("OPTIMAL_DB_PATH"),
	}
}

// Page is the minimum shape we need to render a context as markdown.
// Callers fill it from whatever struct they have on hand (sqlc.Context,
// a request body, etc.) — keeping the input small here avoids dragging
// the full sqlc package into services.
type Page struct {
	ID        string
	UserID    string
	Title     string
	Body      string
	Genre     string
	UpdatedAt time.Time
}

// SyncPage materializes the page on disk + triggers reindex. Errors are
// logged via slog and swallowed — Page-save handlers should not fail
// because the engine sync hit a snag.
func (s *PagesEngineSync) SyncPage(ctx context.Context, p Page) {
	if s == nil || s.nodesRoot == "" || s.dbPath == "" {
		return
	}
	if p.ID == "" || p.Title == "" {
		return
	}

	if err := s.writeFile(p); err != nil {
		slog.WarnContext(ctx, "pages_engine_sync: write failed",
			"page_id", p.ID, "error", err)
		return
	}

	go s.reindexAsync(p.ID)
}

// DeletePage removes the on-disk md and re-indexes. We don't delete the
// row inside the SQLite directly — letting Reindex re-walk the directory
// keeps the index consistent with whatever's actually on disk.
func (s *PagesEngineSync) DeletePage(ctx context.Context, pageID string) {
	if s == nil || s.nodesRoot == "" || s.dbPath == "" || pageID == "" {
		return
	}
	path := filepath.Join(s.nodesRoot, businessosNodeSlug, pageID+".md")
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.WarnContext(ctx, "pages_engine_sync: delete failed",
			"page_id", pageID, "path", path, "error", err)
		return
	}
	go s.reindexAsync(pageID)
}

// writeFile renders the Page as markdown with frontmatter and writes it
// atomically (write to .tmp + rename) so a partial write never gets
// indexed.
func (s *PagesEngineSync) writeFile(p Page) error {
	dir := filepath.Join(s.nodesRoot, businessosNodeSlug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	if err := s.ensureNodeContextFile(dir); err != nil {
		return err
	}

	body := renderPageMarkdown(p)
	final := filepath.Join(dir, p.ID+".md")
	tmp := final + ".tmp"

	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, final); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// ensureNodeContextFile guarantees the businessos node has a context.md
// at its root — every node directory needs one for the engine's nodes
// listing to surface it. Idempotent: only writes when the file is
// missing so we never clobber operator edits.
func (s *PagesEngineSync) ensureNodeContextFile(dir string) error {
	path := filepath.Join(dir, "context.md")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	body := "---\n" +
		"name: BusinessOS Pages\n" +
		"type: workspace\n" +
		"---\n\n" +
		"# BusinessOS Pages\n\n" +
		"Pages created by users in the BusinessOS Pages module. " +
		"Auto-synced from the contexts table on every save.\n"
	return os.WriteFile(path, []byte(body), 0o644)
}

// reindexAsync runs the engine's reindex under a mutex so concurrent
// page saves don't trample one another. Errors are logged; we don't
// retry — the next save will reindex anyway.
func (s *PagesEngineSync) reindexAsync(pageID string) {
	s.reindex.Lock()
	defer s.reindex.Unlock()
	result, err := optimal.Reindex(s.nodesRoot, s.dbPath)
	if err != nil {
		slog.Warn("pages_engine_sync: reindex failed",
			"page_id", pageID, "error", err)
		return
	}
	slog.Debug("pages_engine_sync: reindex complete",
		"page_id", pageID,
		"files_scanned", result.FilesScanned,
		"files_indexed", result.FilesIndexed,
	)
}

// renderPageMarkdown serializes a Page as YAML-frontmatter + markdown
// body. The frontmatter keys match what the engine's classifier reads
// (id, title, genre, modified_at). Body is the user's content verbatim.
func renderPageMarkdown(p Page) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "id: %s\n", p.ID)
	fmt.Fprintf(&sb, "title: %q\n", p.Title)
	if p.Genre != "" {
		fmt.Fprintf(&sb, "genre: %s\n", p.Genre)
	} else {
		sb.WriteString("genre: page\n")
	}
	fmt.Fprintf(&sb, "source: businessos\n")
	if p.UserID != "" {
		fmt.Fprintf(&sb, "author: %s\n", p.UserID)
	}
	if !p.UpdatedAt.IsZero() {
		fmt.Fprintf(&sb, "modified_at: %s\n", p.UpdatedAt.UTC().Format(time.RFC3339))
	}
	sb.WriteString("---\n\n")
	sb.WriteString("# ")
	sb.WriteString(p.Title)
	sb.WriteString("\n\n")
	sb.WriteString(stripHTMLBlocks(p.Body))
	return sb.String()
}

// stripHTMLBlocks rips out script + style tags so HTML pasted into a
// Page doesn't poison the engine's classifier or pollute search results
// with markup. We deliberately keep inline tags (em, strong, code) —
// the engine's strip-html helpers handle those at retrieval time.
var (
	scriptRe = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	styleRe  = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
)

func stripHTMLBlocks(s string) string {
	s = scriptRe.ReplaceAllString(s, "")
	s = styleRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
