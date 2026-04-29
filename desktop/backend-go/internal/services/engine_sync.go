// Package services — engine_sync.go
//
// One service every BusinessOS module hooks into to push its writes into
// the OptimalEngine knowledge graph. After a module saves an entity (a
// Page, a Task, a Project, a CRM record, a chat message, a daily log,
// etc.), it calls EngineSync.Enqueue(entity). EngineSync renders the
// entity as a markdown file under OPTIMAL_NODES_ROOT/<module-node>/
// and a single background worker reindexes the SQLite the engine reads
// from — debounced so a burst of saves doesn't trigger N reindexes.
//
// Modules don't need to know engine internals. They produce a Signal
// struct; EngineSync handles paths, frontmatter, atomic writes, and
// the reindex cadence. When OPTIMAL_NODES_ROOT or OPTIMAL_DB_PATH are
// unset, every method is a silent no-op so module CRUD continues to
// work without an engine wired.
//
// Read-path access (RAG, search) is exposed via the existing
// /api/optimal/* HTTP endpoints — modules can query the engine for
// context-augmented responses without involving this service at all.
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	optimal "github.com/Miosa-osa/OptimalEngine-go"
)

// Signal is the universal shape every BusinessOS module produces. Modules
// fill the fields they have; EngineSync handles the rest. ID is the only
// hard requirement; everything else degrades gracefully.
type Signal struct {
	// Module identifies which BusinessOS module produced this entity.
	// Maps to a dedicated node folder under OPTIMAL_NODES_ROOT — see
	// moduleToNode for the canonical list.
	Module Module

	// ID is the entity's stable identifier (UUID or similar). The MD
	// filename is derived from it.
	ID string

	// Title shows up as the H1 in the rendered markdown and as the
	// signal's title in the engine. Falls back to "<module>-<id>".
	Title string

	// Body is the textual content the engine indexes for retrieval.
	// Markdown / plain text both work — the engine's classifier
	// strips HTML and computes its own structure features.
	Body string

	// Genre is the engine's S/N "genre" dimension (note, plan, brief,
	// transcript, decision-log, etc.). When empty, EngineSync uses a
	// per-module default.
	Genre string

	// AuthorID is the BusinessOS user.id who owns this entity. Stored
	// in frontmatter so per-user audit + filtering still work after
	// ingest.
	AuthorID string

	// ModifiedAt drives the engine's recency ranking. Falls back to
	// time.Now() when zero.
	ModifiedAt time.Time

	// Metadata is arbitrary key/value pairs that show up under
	// `metadata:` in the frontmatter. Useful for module-specific
	// fields the engine doesn't have a first-class slot for
	// (project_id, task_status, deal_stage, etc.).
	Metadata map[string]string
}

// Module enumerates every BusinessOS subsystem that pushes signals into
// the engine. Each maps to a dedicated <NN>-<slug> directory under
// OPTIMAL_NODES_ROOT so a module's signals stay namespaced — easy to
// filter in the graph view, easy to purge if a module is reset.
type Module string

const (
	ModulePages    Module = "pages"     // → 10-businessos (legacy: Pages were first)
	ModuleTasks    Module = "tasks"     // → 11-tasks
	ModuleProjects Module = "projects"  // → 12-projects
	ModuleDailyLog Module = "daily-log" // → 13-daily-log
	ModuleCRM      Module = "crm"       // → 14-crm
	ModuleChat     Module = "chat"      // → 15-chat
	ModuleCalendar Module = "calendar"  // → 16-calendar
	ModuleTeam     Module = "team"      // → 17-team
	ModuleClients  Module = "clients"   // → 18-clients
)

// moduleToNode is the canonical mapping. Keeping it in one place makes
// it impossible for a module to silently land in the wrong folder.
var moduleToNode = map[Module]string{
	ModulePages:    "10-businessos",
	ModuleTasks:    "11-tasks",
	ModuleProjects: "12-projects",
	ModuleDailyLog: "13-daily-log",
	ModuleCRM:      "14-crm",
	ModuleChat:     "15-chat",
	ModuleCalendar: "16-calendar",
	ModuleTeam:     "17-team",
	ModuleClients:  "18-clients",
}

// moduleDefaults seeds genre + the node-level context.md content for each
// module. Anything not listed here falls back to the catch-all "page".
var moduleDefaults = map[Module]struct {
	genre       string
	contextName string
	contextDesc string
}{
	ModulePages:    {"page", "BusinessOS Pages", "Pages created by users in the BusinessOS Pages module."},
	ModuleTasks:    {"task", "BusinessOS Tasks", "Tasks tracked in the BusinessOS Tasks module."},
	ModuleProjects: {"project", "BusinessOS Projects", "Projects tracked in the BusinessOS Projects module."},
	ModuleDailyLog: {"log_entry", "BusinessOS Daily Log", "Daily log entries from the BusinessOS Daily Log module."},
	ModuleCRM:      {"crm_record", "BusinessOS CRM", "Companies, contacts, deals, and interactions from the BusinessOS CRM."},
	ModuleChat:     {"message", "BusinessOS Chat", "Conversations and messages from the BusinessOS Chat module + desktop dock."},
	ModuleCalendar: {"event", "BusinessOS Calendar", "Calendar events from the BusinessOS Calendar module."},
	ModuleTeam:     {"team_member", "BusinessOS Team", "Team members from the BusinessOS Team module."},
	ModuleClients:  {"client", "BusinessOS Clients", "Clients from the BusinessOS Clients module."},
}

// reindexDebounce sets the maximum delay between a save and the engine
// seeing the new content. Tighter = faster graph updates but more
// reindex churn; looser = bursty saves coalesce into a single reindex.
// 5s is a comfortable middle for an interactive desktop app.
const reindexDebounce = 5 * time.Second

// EngineSync is the single entry point every module uses.
type EngineSync struct {
	nodesRoot string
	dbPath    string

	// pending is the set of module folders dirtied since the last
	// reindex. We don't actually need to track which signal changed —
	// optimal.Reindex walks the full tree — but we use the set to
	// decide whether to call Reindex at all.
	mu       sync.Mutex
	pending  map[string]struct{}
	dirty    chan struct{}
	stopOnce sync.Once
	stop     chan struct{}
}

// NewEngineSync wires the service from environment + starts the
// debounce worker. When OPTIMAL_NODES_ROOT or OPTIMAL_DB_PATH are
// unset, the worker exits immediately and every method is a no-op.
//
// The returned service is safe to construct from any number of call
// sites — only one worker runs per *EngineSync.
func NewEngineSync() *EngineSync {
	s := &EngineSync{
		nodesRoot: os.Getenv("OPTIMAL_NODES_ROOT"),
		dbPath:    os.Getenv("OPTIMAL_DB_PATH"),
		pending:   map[string]struct{}{},
		dirty:     make(chan struct{}, 1),
		stop:      make(chan struct{}),
	}
	if s.nodesRoot == "" || s.dbPath == "" {
		// Disabled: leave dirty channel unread; Enqueue silently no-ops.
		return s
	}
	go s.run()
	return s
}

// Stop halts the background worker. Idempotent. Pending writes are
// flushed by a final reindex before the worker exits so we don't lose
// the last second of saves on shutdown.
func (s *EngineSync) Stop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
}

// Enqueue is the single fire-and-forget entrypoint modules use. Errors
// are logged but never returned — the calling handler should not fail
// because the engine sync hit a snag.
func (s *EngineSync) Enqueue(ctx context.Context, sig Signal) {
	if s == nil || s.nodesRoot == "" || s.dbPath == "" {
		return
	}
	if sig.ID == "" {
		return
	}
	if err := s.writeSignal(sig); err != nil {
		slog.WarnContext(ctx, "engine_sync: write failed",
			"module", string(sig.Module), "id", sig.ID, "error", err)
		return
	}
	s.markDirty(string(sig.Module))
}

// EnqueueDelete removes the on-disk MD for the entity then triggers a
// reindex. The reindex re-walks the directory so the SQLite stays
// consistent with what's actually on disk.
func (s *EngineSync) EnqueueDelete(ctx context.Context, module Module, id string) {
	if s == nil || s.nodesRoot == "" || s.dbPath == "" || id == "" {
		return
	}
	node := nodeFor(module)
	path := filepath.Join(s.nodesRoot, node, id+".md")
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.WarnContext(ctx, "engine_sync: delete failed",
			"module", string(module), "id", id, "path", path, "error", err)
		return
	}
	s.markDirty(string(module))
}

// markDirty records that some signal in the named module changed and
// pings the worker to wake up. Non-blocking — if the worker is already
// behind, we coalesce by dropping duplicate pings.
func (s *EngineSync) markDirty(module string) {
	s.mu.Lock()
	s.pending[module] = struct{}{}
	s.mu.Unlock()
	select {
	case s.dirty <- struct{}{}:
	default:
	}
}

// run is the single goroutine that drains dirty pings and reindexes
// once per debounce window. Mutex on optimal.Reindex would be overkill
// — there's only one writer here.
func (s *EngineSync) run() {
	for {
		select {
		case <-s.stop:
			s.flush(true)
			return
		case <-s.dirty:
			// Wait for the burst to settle before reindexing.
			timer := time.NewTimer(reindexDebounce)
			drained := false
			for !drained {
				select {
				case <-s.stop:
					timer.Stop()
					s.flush(true)
					return
				case <-s.dirty:
					// Reset the debounce — more saves are coming.
					if !timer.Stop() {
						<-timer.C
					}
					timer.Reset(reindexDebounce)
				case <-timer.C:
					drained = true
				}
			}
			s.flush(false)
		}
	}
}

func (s *EngineSync) flush(final bool) {
	s.mu.Lock()
	dirty := len(s.pending) > 0
	modules := make([]string, 0, len(s.pending))
	for k := range s.pending {
		modules = append(modules, k)
	}
	s.pending = map[string]struct{}{}
	s.mu.Unlock()
	if !dirty {
		return
	}

	// When the live Elixir engine is in front (OPTIMAL_ENGINE_URL set),
	// hit its POST /api/reindex endpoint. The Go reindex would try to
	// upsert into Elixir's contexts_fts virtual table, which has a
	// different column set — letting the engine own its writes avoids
	// the schema drift entirely. When no engine URL is set, fall back
	// to Go's in-process Reindex (which created the db, so its schema
	// is the one on disk).
	if engineURL := os.Getenv("OPTIMAL_ENGINE_URL"); engineURL != "" {
		s.reindexViaEngine(engineURL, modules, final)
		return
	}
	result, err := optimal.Reindex(s.nodesRoot, s.dbPath)
	if err != nil {
		slog.Warn("engine_sync: reindex failed",
			"modules", modules, "final", final, "error", err)
		return
	}
	slog.Info("engine_sync: reindex complete (go port)",
		"modules", modules,
		"files_scanned", result.FilesScanned,
		"files_indexed", result.FilesIndexed,
		"final", final,
	)
}

// reindexViaEngine POSTs to the live engine's /api/reindex endpoint.
// The engine's own indexer understands its SQLite schema (notably the
// contexts_fts virtual table) so we don't have to. 202 responses are
// fine — they mean "another reindex is already in flight; the new files
// will land on the next pass."
//
// reindexClient times out aggressively. A reindex of the bundled sample
// workspace takes ~1s; a Page-save batch is comparable. 30s gives plenty
// of headroom without dangling on a hung engine.
func (s *EngineSync) reindexViaEngine(engineURL string, modules []string, final bool) {
	url := strings.TrimRight(engineURL, "/") + "/api/reindex"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("{}"))
	if err != nil {
		slog.Warn("engine_sync: build reindex request failed",
			"engine_url", engineURL, "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := reindexClient.Do(req)
	if err != nil {
		slog.Warn("engine_sync: reindex POST failed",
			"engine_url", engineURL, "modules", modules, "error", err)
		return
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		slog.Info("engine_sync: reindex complete (elixir engine)",
			"modules", modules, "final", final)
	case http.StatusAccepted:
		slog.Info("engine_sync: reindex queued (already running)",
			"modules", modules, "final", final)
	default:
		slog.Warn("engine_sync: reindex unexpected status",
			"engine_url", engineURL, "status", resp.StatusCode,
			"modules", modules)
	}
}

// reindexClient is package-level so we don't pay TCP setup per save burst.
// HTTP/1.1 keep-alive keeps the conn warm for the next reindex.
var reindexClient = &http.Client{Timeout: 30 * time.Second}

// writeSignal renders the signal as markdown + frontmatter and writes
// it atomically to disk. Atomic = write to .tmp + rename, so a partial
// write never gets indexed.
func (s *EngineSync) writeSignal(sig Signal) error {
	node := nodeFor(sig.Module)
	dir := filepath.Join(s.nodesRoot, node)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	if err := s.ensureNodeContext(dir, sig.Module); err != nil {
		return err
	}

	body := renderSignalMarkdown(sig)
	final := filepath.Join(dir, sig.ID+".md")
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

// ensureNodeContext writes the node-level context.md the first time we
// see a module's directory. The engine treats every node directory as
// requiring a context.md at its root — without one, the node doesn't
// surface in /api/optimal/nodes.
func (s *EngineSync) ensureNodeContext(dir string, module Module) error {
	path := filepath.Join(dir, "context.md")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	d, ok := moduleDefaults[module]
	if !ok {
		d = struct {
			genre       string
			contextName string
			contextDesc string
		}{"note", "BusinessOS Module", "Auto-synced from BusinessOS."}
	}
	// Frontmatter shape matches the engine's existing nodes (see
	// sample-workspace/nodes/01-founder/context.md): `node:` is the
	// engine's primary key for ingest; without it the indexer skips
	// the file. `kind:` + `style:` are descriptive, not load-bearing.
	body := "---\n" +
		"node: " + nodeFor(module) + "\n" +
		"kind: workspace\n" +
		"style: internal\n" +
		"source: businessos\n" +
		"---\n\n" +
		"# " + d.contextName + "\n\n" +
		d.contextDesc + " " +
		"Auto-synced from the relevant BusinessOS table on every save.\n"
	return os.WriteFile(path, []byte(body), 0o644)
}

func nodeFor(m Module) string {
	if n, ok := moduleToNode[m]; ok {
		return n
	}
	return "10-businessos"
}

// renderSignalMarkdown emits the YAML-frontmatter shape the engine's
// pipeline indexer recognizes. The critical field is `node:` — without
// it, the indexer skips the file. Other fields (title, genre, mode,
// authored_at, sn_ratio) follow the engine's convention exactly.
func renderSignalMarkdown(sig Signal) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "node: %s\n", nodeFor(sig.Module))

	title := sig.Title
	if title == "" {
		title = fmt.Sprintf("%s-%s", sig.Module, sig.ID)
	}
	fmt.Fprintf(&sb, "title: %s\n", title)

	genre := sig.Genre
	if genre == "" {
		if d, ok := moduleDefaults[sig.Module]; ok {
			genre = d.genre
		} else {
			genre = "note"
		}
	}
	fmt.Fprintf(&sb, "genre: %s\n", genre)
	fmt.Fprintf(&sb, "mode: linguistic\n")
	fmt.Fprintf(&sb, "id: %s\n", sig.ID)
	fmt.Fprintf(&sb, "source: businessos\n")
	fmt.Fprintf(&sb, "module: %s\n", sig.Module)
	if sig.AuthorID != "" {
		fmt.Fprintf(&sb, "author: %s\n", sig.AuthorID)
	}
	modified := sig.ModifiedAt
	if modified.IsZero() {
		modified = time.Now().UTC()
	}
	fmt.Fprintf(&sb, "authored_at: %s\n", modified.UTC().Format(time.RFC3339))
	for k, v := range sig.Metadata {
		fmt.Fprintf(&sb, "%s: %s\n", k, v)
	}
	sb.WriteString("---\n\n")
	sb.WriteString("# ")
	sb.WriteString(title)
	sb.WriteString("\n\n")
	sb.WriteString(stripHTMLDangerous(sig.Body))
	return sb.String()
}

// stripHTMLDangerous removes <script>/<style> blocks from incoming body
// content. Inline tags pass through; the engine's strip-html helpers
// handle those at retrieval time. We deliberately do NOT sanitize the
// rest of the HTML — Pages and other rich-editor modules emit valid
// markup and we want the engine to see it.
var (
	syncScriptRe = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	syncStyleRe  = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
)

func stripHTMLDangerous(s string) string {
	s = syncScriptRe.ReplaceAllString(s, "")
	s = syncStyleRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
