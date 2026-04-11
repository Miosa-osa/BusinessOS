// Package optimal provides a read-only bridge between the BusinessOS Go backend
// and the OptimalOS filesystem (markdown nodes at nodes/, rhythm files at rhythm/).
//
// File layout:
//   - reader.go  — Pure filesystem reads: nodes, signals, rhythm files
//   - engine.go  — Subprocess wrapper for the Elixir mix optimal.* CLI
package optimal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required
)

// NodeSummary holds the essential metadata and raw content for a single node folder.
type NodeSummary struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	ContextMD   string `json:"context_md"`
	SignalMD    string `json:"signal_md"`
	SignalCount int    `json:"signal_count"`
	HasSignals  bool   `json:"has_signals"`
}

// SignalFile represents a single dated signal file inside a node's signals/ subfolder.
type SignalFile struct {
	Slug    string    `json:"slug"`
	Date    time.Time `json:"date"`
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Content string    `json:"content"`
}

// RhythmDay holds today's daily working file from rhythm/daily/.
type RhythmDay struct {
	Date    time.Time `json:"date"`
	Content string    `json:"content"`
}

// ListNodes reads all numbered node folders under nodesRoot, returns summaries sorted
// by their numeric prefix (00, 01, 02, ...). An error is returned only if nodesRoot
// cannot be read; missing context.md or signal.md files are treated as empty content.
func ListNodes(nodesRoot string) ([]NodeSummary, error) {
	entries, err := os.ReadDir(nodesRoot)
	if err != nil {
		return nil, fmt.Errorf("optimal: read nodes dir %q: %w", nodesRoot, err)
	}

	type indexed struct {
		prefix int
		sum    NodeSummary
	}
	var items []indexed

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		prefix, ok := parseNumericPrefix(name)
		if !ok {
			continue
		}

		sum, err := readNodeDir(nodesRoot, name)
		if err != nil {
			// Skip unreadable nodes rather than aborting the whole list.
			continue
		}
		items = append(items, indexed{prefix: prefix, sum: sum})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].prefix < items[j].prefix
	})

	summaries := make([]NodeSummary, len(items))
	for i, item := range items {
		summaries[i] = item.sum
	}
	return summaries, nil
}

// GetNode returns the NodeSummary for the node identified by slug (e.g. "01-roberto").
// Returns an error if the folder does not exist.
func GetNode(nodesRoot, slug string) (NodeSummary, error) {
	nodeDir := filepath.Join(nodesRoot, slug)
	if _, err := os.Stat(nodeDir); err != nil {
		return NodeSummary{}, fmt.Errorf("optimal: node %q not found: %w", slug, err)
	}
	return readNodeDir(nodesRoot, slug)
}

// GetNodeSignals returns all dated signal files from the node's signals/ subfolder,
// sorted newest-first. Returns an empty slice (not an error) if signals/ is absent.
func GetNodeSignals(nodesRoot, slug string) ([]SignalFile, error) {
	signalsDir := filepath.Join(nodesRoot, slug, "signals")
	entries, err := os.ReadDir(signalsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("optimal: read signals dir for %q: %w", slug, err)
	}

	var files []SignalFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(signalsDir, e.Name())
		content, _ := readFileOrEmpty(path)
		fileSlug := strings.TrimSuffix(e.Name(), ".md")
		date := parseDatePrefix(fileSlug)
		name := extractFrontmatterField(content, "title")
		if name == "" {
			name = slugToTitle(fileSlug)
		}
		files = append(files, SignalFile{
			Slug:    fileSlug,
			Date:    date,
			Name:    name,
			Path:    path,
			Content: content,
		})
	}

	// Sort newest-first by filename (YYYY-MM-DD prefix sorts lexicographically).
	sort.Slice(files, func(i, j int) bool {
		return files[i].Slug > files[j].Slug
	})
	return files, nil
}

// GetTodayRhythm reads rhythm/daily/YYYY-MM-DD.md for today's date from osRoot.
// If the file does not exist, Content is empty and no error is returned.
func GetTodayRhythm(osRoot string) (RhythmDay, error) {
	today := time.Now()
	filename := today.Format("2006-01-02") + ".md"
	path := filepath.Join(osRoot, "rhythm", "daily", filename)

	content, _ := readFileOrEmpty(path)
	return RhythmDay{
		Date:    today.Truncate(24 * time.Hour),
		Content: content,
	}, nil
}

// FileEntry represents a single file or directory in a node's file tree.
type FileEntry struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"` // relative to nodes root
	IsDir    bool        `json:"is_dir"`
	Size     int64       `json:"size"`
	Children []FileEntry `json:"children,omitempty"`
}

// GetNodeFileTree returns the recursive directory tree for a node, including all
// .md files and subdirectories. Paths are relative to the node folder.
func GetNodeFileTree(nodesRoot, slug string) ([]FileEntry, error) {
	nodeDir := filepath.Join(nodesRoot, slug)
	if _, err := os.Stat(nodeDir); err != nil {
		return nil, fmt.Errorf("optimal: node %q not found: %w", slug, err)
	}
	return readDirRecursive(nodeDir, slug)
}

// GetNodeFile reads a single file from inside a node. filePath is relative to the node folder.
func GetNodeFile(nodesRoot, slug, filePath string) (string, error) {
	// Prevent path traversal
	full := filepath.Join(nodesRoot, slug, filePath)
	clean := filepath.Clean(full)
	nodeDir := filepath.Clean(filepath.Join(nodesRoot, slug))
	if !strings.HasPrefix(clean, nodeDir) {
		return "", fmt.Errorf("optimal: path traversal denied")
	}
	b, err := os.ReadFile(clean)
	if err != nil {
		return "", fmt.Errorf("optimal: read file %q: %w", filePath, err)
	}
	return string(b), nil
}

// SaveNodeFile writes content to an existing or new file inside a node.
// Creates parent directories as needed. Prevents path traversal.
func SaveNodeFile(nodesRoot, slug, filePath, content string) error {
	full := filepath.Join(nodesRoot, slug, filePath)
	clean := filepath.Clean(full)
	nodeDir := filepath.Clean(filepath.Join(nodesRoot, slug))
	if !strings.HasPrefix(clean, nodeDir) {
		return fmt.Errorf("optimal: path traversal denied")
	}
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
		return fmt.Errorf("optimal: mkdir for %q: %w", filePath, err)
	}
	if err := os.WriteFile(clean, []byte(content), 0o644); err != nil {
		return fmt.Errorf("optimal: write file %q: %w", filePath, err)
	}
	return nil
}

// CreateNodeFile creates a new file inside a node. Returns an error if the file
// already exists. Creates parent directories as needed.
func CreateNodeFile(nodesRoot, slug, filePath, content string) error {
	full := filepath.Join(nodesRoot, slug, filePath)
	clean := filepath.Clean(full)
	nodeDir := filepath.Clean(filepath.Join(nodesRoot, slug))
	if !strings.HasPrefix(clean, nodeDir) {
		return fmt.Errorf("optimal: path traversal denied")
	}
	// Check node exists
	if _, err := os.Stat(nodeDir); err != nil {
		return fmt.Errorf("optimal: node %q not found: %w", slug, err)
	}
	// Check file doesn't already exist
	if _, err := os.Stat(clean); err == nil {
		return fmt.Errorf("optimal: file already exists: %s", filePath)
	}
	// Create parent dirs
	if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
		return fmt.Errorf("optimal: mkdir for %q: %w", filePath, err)
	}
	return os.WriteFile(clean, []byte(content), 0o644)
}

func readDirRecursive(dir, relBase string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []FileEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		relPath := filepath.Join(relBase, name)
		info, err := e.Info()
		if err != nil {
			continue
		}

		entry := FileEntry{
			Name:  name,
			Path:  relPath,
			IsDir: e.IsDir(),
			Size:  info.Size(),
		}

		if e.IsDir() {
			children, err := readDirRecursive(filepath.Join(dir, name), relPath)
			if err == nil {
				entry.Children = children
			}
		}

		result = append(result, entry)
	}

	// Sort: directories first, then alphabetically
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// ── internal helpers ──────────────────────────────────────────────────────────

// readNodeDir builds a NodeSummary by reading context.md, signal.md, and signals/.
func readNodeDir(nodesRoot, slug string) (NodeSummary, error) {
	base := filepath.Join(nodesRoot, slug)

	contextMD, _ := readFileOrEmpty(filepath.Join(base, "context.md"))
	signalMD, _ := readFileOrEmpty(filepath.Join(base, "signal.md"))

	// Count signal files.
	signalCount := 0
	signalsDir := filepath.Join(base, "signals")
	if entries, err := os.ReadDir(signalsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				signalCount++
			}
		}
	}

	// Derive name and type from context.md frontmatter, falling back to the slug.
	name := extractFrontmatterField(contextMD, "name")
	if name == "" {
		name = slugToTitle(slug)
	}
	nodeType := extractFrontmatterField(contextMD, "type")

	return NodeSummary{
		Slug:        slug,
		Name:        name,
		Type:        nodeType,
		ContextMD:   contextMD,
		SignalMD:    signalMD,
		SignalCount: signalCount,
		HasSignals:  signalCount > 0,
	}, nil
}

// readFileOrEmpty returns the file contents, or an empty string if the file does not
// exist. Other errors (permission denied, etc.) are also silenced and return "".
func readFileOrEmpty(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// parseNumericPrefix extracts the leading integer from names like "01-roberto".
// Returns (0, false) if no numeric prefix is found.
func parseNumericPrefix(name string) (int, bool) {
	idx := strings.IndexByte(name, '-')
	var numStr string
	if idx > 0 {
		numStr = name[:idx]
	} else {
		numStr = name
	}
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, false
	}
	return n, true
}

// parseDatePrefix extracts a time.Time from the "YYYY-MM-DD" prefix of a filename
// slug such as "2026-03-18-ed-pricing-call". Returns zero value if unparseable.
func parseDatePrefix(slug string) time.Time {
	if len(slug) < 10 {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", slug[:10])
	if err != nil {
		return time.Time{}
	}
	return t
}

// extractFrontmatterField parses a simple "key: value" pair from YAML frontmatter
// delimited by "---" at the start of a markdown file. It does not use a full YAML
// parser to avoid adding a dependency for this narrow use case.
func extractFrontmatterField(content, key string) string {
	if !strings.HasPrefix(content, "---") {
		return ""
	}
	// Find the closing delimiter.
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return ""
	}
	fm := rest[:end]
	prefix := key + ":"
	for _, line := range strings.Split(fm, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			val := strings.TrimSpace(trimmed[len(prefix):])
			// Strip surrounding quotes if present.
			if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			}
			return val
		}
	}
	return ""
}

// slugToTitle converts a slug like "01-roberto" or "2026-03-18-ed-call" into a
// human-readable title by removing the numeric/date prefix and title-casing words.
func slugToTitle(slug string) string {
	// Strip leading numeric prefix ("01-", "12-").
	if idx := strings.IndexByte(slug, '-'); idx > 0 {
		if _, err := strconv.Atoi(slug[:idx]); err == nil {
			slug = slug[idx+1:]
		}
	}
	words := strings.Split(slug, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// ── Team ─────────────────────────────────────────────────────────────────────

// TeamMember represents a single person extracted from nodes/10-team/context.md.
type TeamMember struct {
	Name    string   `json:"name"`
	Role    string   `json:"role"`
	Nodes   []string `json:"nodes"`
	Channel string   `json:"channel"`
	Status  string   `json:"status"` // "active" | "inactive" | "exited"
}

// GetTeamRoster reads nodes/10-team/context.md and parses the people listed
// there. It degrades gracefully on missing or malformed content, returning an
// empty slice rather than an error.
func GetTeamRoster(nodesRoot string) ([]TeamMember, error) {
	path := filepath.Join(nodesRoot, "10-team", "context.md")
	content, _ := readFileOrEmpty(path)
	if content == "" {
		return nil, nil
	}
	return parseTeamMembers(content), nil
}

// parseTeamMembers extracts TeamMember records from markdown content.
// It handles two layouts:
//   - Markdown table rows:  | Name | Role | ... |
//   - Bullet list entries:  - **Name** — Role
func parseTeamMembers(content string) []TeamMember {
	var members []TeamMember
	lines := strings.Split(content, "\n")

	inTable := false
	headerSeen := false

	for _, raw := range lines {
		line := strings.TrimSpace(raw)

		// ── markdown table path ───────────────────────────────────────────────
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			cols := splitTableRow(line)
			if len(cols) < 2 {
				continue
			}
			// Detect separator row (cells are just dashes/colons).
			if isTableSeparator(cols) {
				inTable = true
				headerSeen = true
				continue
			}
			if !headerSeen {
				// First non-separator pipe row → likely the header row.
				inTable = true
				continue
			}
			if inTable {
				m := teamMemberFromTableRow(cols)
				if m.Name != "" {
					members = append(members, m)
				}
			}
			continue
		}

		// Leaving a table block.
		if inTable && !strings.HasPrefix(line, "|") {
			inTable = false
			headerSeen = false
		}

		// ── bullet list path ─────────────────────────────────────────────────
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			m := teamMemberFromBullet(line[2:])
			if m.Name != "" {
				members = append(members, m)
			}
		}
	}
	return members
}

// splitTableRow splits a markdown table row like "| A | B | C |" into trimmed cells.
func splitTableRow(line string) []string {
	// Trim leading and trailing pipes.
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		cols = append(cols, strings.TrimSpace(p))
	}
	return cols
}

// isTableSeparator reports whether every non-empty cell in cols contains only
// dashes, colons, and spaces (a markdown table separator row).
func isTableSeparator(cols []string) bool {
	for _, c := range cols {
		if c == "" {
			continue
		}
		for _, ch := range c {
			if ch != '-' && ch != ':' && ch != ' ' {
				return false
			}
		}
	}
	return true
}

// teamMemberFromTableRow builds a TeamMember from a table row's cells.
// Column order assumed: Name, Role, [Nodes|Channel], [Channel|Status], [Status].
// Unknown columns are ignored; missing columns default to zero values.
func teamMemberFromTableRow(cols []string) TeamMember {
	get := func(i int) string {
		if i < len(cols) {
			return stripMarkdownEmphasis(cols[i])
		}
		return ""
	}
	m := TeamMember{
		Name:   get(0),
		Role:   get(1),
		Status: "active",
	}
	// Heuristically find channel and status from remaining columns.
	for i := 2; i < len(cols); i++ {
		v := get(i)
		lv := strings.ToLower(v)
		switch {
		case lv == "active" || lv == "inactive" || lv == "exited":
			m.Status = lv
		case strings.Contains(lv, "slack") || strings.Contains(lv, "email") ||
			strings.Contains(lv, "voice") || strings.Contains(lv, "github"):
			m.Channel = v
		case v != "":
			// Treat as nodes list (comma or slash separated).
			if m.Nodes == nil {
				m.Nodes = splitNodes(v)
			}
		}
	}
	if m.Nodes == nil {
		m.Nodes = []string{}
	}
	return m
}

// teamMemberFromBullet builds a TeamMember from a bullet list entry.
// Handles patterns like: **Name** — Role (status), or Name: Role
func teamMemberFromBullet(text string) TeamMember {
	text = strings.TrimSpace(text)
	m := TeamMember{Status: "active", Nodes: []string{}}

	// Extract status hint from parentheses at end: "(exited)"
	if idx := strings.LastIndex(text, "("); idx > 0 {
		inner := strings.ToLower(strings.Trim(text[idx:], "()"))
		if inner == "exited" || inner == "inactive" || inner == "active" {
			m.Status = inner
			text = strings.TrimSpace(text[:idx])
		}
	}

	text = stripMarkdownEmphasis(text)

	// Try "Name — Role" or "Name - Role".
	for _, sep := range []string{" — ", " – ", " - "} {
		if idx := strings.Index(text, sep); idx > 0 {
			m.Name = strings.TrimSpace(text[:idx])
			m.Role = strings.TrimSpace(text[idx+len(sep):])
			return m
		}
	}

	// Try "Name: Role".
	if idx := strings.IndexByte(text, ':'); idx > 0 {
		m.Name = strings.TrimSpace(text[:idx])
		m.Role = strings.TrimSpace(text[idx+1:])
		return m
	}

	// Fall back: whole line is the name.
	m.Name = text
	return m
}

// stripMarkdownEmphasis removes **bold** and *italic* markers.
func stripMarkdownEmphasis(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.TrimSpace(s)
}

// splitNodes splits a comma- or slash-separated node list into trimmed strings.
func splitNodes(s string) []string {
	var sep string
	if strings.Contains(s, ",") {
		sep = ","
	} else {
		sep = "/"
	}
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// ── Revenue ───────────────────────────────────────────────────────────────────

// RevenueStream represents a single income stream parsed from node 11.
type RevenueStream struct {
	Name   string  `json:"name"`
	Model  string  `json:"model"`
	Amount float64 `json:"amount"`
	Target float64 `json:"target"`
	Node   string  `json:"node"`
}

// RevenueData is the payload returned by GetRevenueStreams.
type RevenueData struct {
	Streams  []RevenueStream `json:"streams"`
	TotalMRR float64         `json:"total_mrr"`
}

// GetRevenueStreams reads nodes/11-money-revenue/context.md and parses revenue
// streams from markdown tables or bullet lists. Returns empty RevenueData (not
// an error) when the file is absent or contains no parseable rows.
func GetRevenueStreams(nodesRoot string) (RevenueData, error) {
	path := filepath.Join(nodesRoot, "11-money-revenue", "context.md")
	content, _ := readFileOrEmpty(path)
	if content == "" {
		return RevenueData{Streams: []RevenueStream{}}, nil
	}
	streams := parseRevenueStreams(content)
	var total float64
	for _, s := range streams {
		total += s.Amount
	}
	return RevenueData{Streams: streams, TotalMRR: total}, nil
}

// parseRevenueStreams extracts RevenueStream records from markdown content.
// Handles markdown tables and "- Name: $amount" bullet patterns.
func parseRevenueStreams(content string) []RevenueStream {
	var streams []RevenueStream
	lines := strings.Split(content, "\n")

	inTable := false
	headerSeen := false

	for _, raw := range lines {
		line := strings.TrimSpace(raw)

		// ── markdown table path ───────────────────────────────────────────────
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			cols := splitTableRow(line)
			if len(cols) < 2 {
				continue
			}
			if isTableSeparator(cols) {
				inTable = true
				headerSeen = true
				continue
			}
			if !headerSeen {
				inTable = true
				continue
			}
			if inTable {
				s := revenueStreamFromTableRow(cols)
				if s.Name != "" {
					streams = append(streams, s)
				}
			}
			continue
		}

		if inTable && !strings.HasPrefix(line, "|") {
			inTable = false
			headerSeen = false
		}

		// ── bullet list path ─────────────────────────────────────────────────
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			s := revenueStreamFromBullet(line[2:])
			if s.Name != "" {
				streams = append(streams, s)
			}
		}
	}
	return streams
}

// revenueStreamFromTableRow builds a RevenueStream from a table row.
// Assumed column order: Name, Model, Amount, Target, Node (extras ignored).
func revenueStreamFromTableRow(cols []string) RevenueStream {
	get := func(i int) string {
		if i < len(cols) {
			return strings.TrimSpace(cols[i])
		}
		return ""
	}
	return RevenueStream{
		Name:   get(0),
		Model:  get(1),
		Amount: parseCurrency(get(2)),
		Target: parseCurrency(get(3)),
		Node:   get(4),
	}
}

// revenueStreamFromBullet parses "Name: $amount" or "Name — model — $amount".
func revenueStreamFromBullet(text string) RevenueStream {
	text = strings.TrimSpace(text)
	s := RevenueStream{}
	// Try "Name: $amount (model)"
	if idx := strings.IndexByte(text, ':'); idx > 0 {
		s.Name = strings.TrimSpace(text[:idx])
		rest := strings.TrimSpace(text[idx+1:])
		// Extract amount from the rest.
		s.Amount = parseCurrency(rest)
		// Use anything before the $ as the model hint.
		if dollarIdx := strings.IndexByte(rest, '$'); dollarIdx > 0 {
			s.Model = strings.TrimSpace(rest[:dollarIdx])
		}
		return s
	}
	// Try em-dash separated: "Name — model — $amount"
	parts := strings.SplitN(text, "—", 3)
	if len(parts) >= 2 {
		s.Name = strings.TrimSpace(stripMarkdownEmphasis(parts[0]))
		s.Model = strings.TrimSpace(parts[1])
		if len(parts) >= 3 {
			s.Amount = parseCurrency(parts[2])
		}
		return s
	}
	return s
}

// parseCurrency extracts a float64 from strings like "$2,500", "2500", "2.5K".
// Returns 0 on unparseable input.
func parseCurrency(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	// Handle "K" suffix (e.g. "2.5K").
	multiplier := 1.0
	if strings.HasSuffix(strings.ToUpper(s), "K") {
		multiplier = 1000
		s = s[:len(s)-1]
	}

	// Strip any trailing non-numeric characters (units, spaces).
	for len(s) > 0 {
		last := s[len(s)-1]
		if (last >= '0' && last <= '9') || last == '.' {
			break
		}
		s = s[:len(s)-1]
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v * multiplier
}

// ── Projects ──────────────────────────────────────────────────────────────────

// ProjectEntry represents a single project directory found under a node.
type ProjectEntry struct {
	Name  string   `json:"name"`
	Node  string   `json:"node"`  // parent node slug (e.g. "04-ai-masters")
	Path  string   `json:"path"`  // absolute filesystem path
	Files []string `json:"files"` // .md filenames within the project dir
}

// GetAllProjects scans every nodes/*/projects/ directory and returns a unified
// list of ProjectEntry records. Nodes with no projects/ folder are silently
// skipped. Returns an empty slice (not an error) when no projects exist.
func GetAllProjects(nodesRoot string) ([]ProjectEntry, error) {
	entries, err := os.ReadDir(nodesRoot)
	if err != nil {
		return nil, fmt.Errorf("optimal: read nodes dir %q: %w", nodesRoot, err)
	}

	var projects []ProjectEntry

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		nodeSlug := e.Name()
		projectsDir := filepath.Join(nodesRoot, nodeSlug, "projects")

		projEntries, err := os.ReadDir(projectsDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			// Unreadable projects dir — skip without aborting the whole scan.
			continue
		}

		for _, pe := range projEntries {
			if !pe.IsDir() {
				continue
			}
			projPath := filepath.Join(projectsDir, pe.Name())
			mdFiles := listMDFiles(projPath)
			projects = append(projects, ProjectEntry{
				Name:  slugToTitle(pe.Name()),
				Node:  nodeSlug,
				Path:  projPath,
				Files: mdFiles,
			})
		}
	}

	if projects == nil {
		projects = []ProjectEntry{}
	}
	return projects, nil
}

// listMDFiles returns the base filenames of all .md files directly inside dir.
// Returns an empty slice on any error.
func listMDFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, e.Name())
		}
	}
	if files == nil {
		files = []string{}
	}
	return files
}

// ── Weekly Rhythm ─────────────────────────────────────────────────────────────

// WeeklyPlan holds the most recent weekly rhythm file.
type WeeklyPlan struct {
	WeekOf         string   `json:"week_of"`
	Content        string   `json:"content"`
	NonNegotiables []string `json:"non_negotiables"`
}

// GetWeeklyRhythm reads the most recent rhythm/weekly/week-of-*.md file from
// osRoot (sorted lexicographically by filename, latest wins). Returns an empty
// WeeklyPlan (not an error) when no weekly files exist.
func GetWeeklyRhythm(osRoot string) (WeeklyPlan, error) {
	weeklyDir := filepath.Join(osRoot, "rhythm", "weekly")
	entries, err := os.ReadDir(weeklyDir)
	if os.IsNotExist(err) {
		return WeeklyPlan{NonNegotiables: []string{}}, nil
	}
	if err != nil {
		return WeeklyPlan{NonNegotiables: []string{}},
			fmt.Errorf("optimal: read weekly rhythm dir %q: %w", weeklyDir, err)
	}

	// Collect week-of-*.md files and sort lexicographically (latest last).
	var weekFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "week-of-") && strings.HasSuffix(e.Name(), ".md") {
			weekFiles = append(weekFiles, e.Name())
		}
	}
	if len(weekFiles) == 0 {
		return WeeklyPlan{NonNegotiables: []string{}}, nil
	}
	sort.Strings(weekFiles)
	latest := weekFiles[len(weekFiles)-1]

	// Extract week date from filename: "week-of-2026-04-07.md" → "2026-04-07"
	weekOf := strings.TrimPrefix(latest, "week-of-")
	weekOf = strings.TrimSuffix(weekOf, ".md")

	path := filepath.Join(weeklyDir, latest)
	content, _ := readFileOrEmpty(path)

	return WeeklyPlan{
		WeekOf:         weekOf,
		Content:        content,
		NonNegotiables: parseNonNegotiables(content),
	}, nil
}

// parseNonNegotiables extracts bullet-point lines following a "Non-Negotiables"
// or "non-negotiables" heading. Stops at the next markdown heading or blank line
// run of more than one. Returns an empty slice when the section is absent.
func parseNonNegotiables(content string) []string {
	var items []string
	lines := strings.Split(content, "\n")

	inSection := false
	blankCount := 0

	for _, raw := range lines {
		line := strings.TrimSpace(raw)

		// Detect the section header (any heading level).
		if strings.Contains(strings.ToLower(line), "non-negotiable") &&
			(strings.HasPrefix(line, "#") || strings.HasPrefix(line, "**")) {
			inSection = true
			blankCount = 0
			continue
		}

		if !inSection {
			continue
		}

		// Stop at the next heading.
		if strings.HasPrefix(line, "#") {
			break
		}

		if line == "" {
			blankCount++
			if blankCount > 1 {
				break
			}
			continue
		}
		blankCount = 0

		// Collect bullet lines.
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") ||
			strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") || strings.HasPrefix(line, "3.") {
			item := strings.TrimSpace(line)
			// Strip leading bullet markers.
			for _, pfx := range []string{"- ", "* ", "1. ", "2. ", "3. "} {
				item = strings.TrimPrefix(item, pfx)
			}
			item = stripMarkdownEmphasis(item)
			if item != "" {
				items = append(items, item)
			}
		}
	}

	if items == nil {
		items = []string{}
	}
	return items
}

// ── Dashboard ─────────────────────────────────────────────────────────────────

// NodeRef is a lightweight node summary used inside DashboardData.
type NodeRef struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// DashboardData aggregates key metrics from multiple sources for the UI.
type DashboardData struct {
	Nodes       []NodeRef `json:"nodes"`
	NodeCount   int       `json:"node_count"`
	CommandMD   string    `json:"command_md"` // nodes/00-command/context.md
	TotalMRR    float64   `json:"total_mrr"`
	SignalCount int       `json:"signal_count"`
	EntityCount int       `json:"entity_count"`
	EdgeCount   int       `json:"edge_count"`
}

// GetDashboard aggregates dashboard data from the filesystem and (optionally)
// the SQLite index at dbPath. If dbPath is empty or inaccessible, entity/edge
// counts default to zero — the rest of the dashboard still works.
func GetDashboard(nodesRoot, osRoot, dbPath string) (DashboardData, error) {
	var dash DashboardData

	// ── node list ─────────────────────────────────────────────────────────────
	entries, err := os.ReadDir(nodesRoot)
	if err != nil {
		return dash, fmt.Errorf("optimal: read nodes dir %q: %w", nodesRoot, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, ok := parseNumericPrefix(e.Name()); !ok {
			continue
		}
		name := slugToTitle(e.Name())
		// Try to read a better name from context.md frontmatter.
		ctx, _ := readFileOrEmpty(filepath.Join(nodesRoot, e.Name(), "context.md"))
		if n := extractFrontmatterField(ctx, "name"); n != "" {
			name = n
		}
		dash.Nodes = append(dash.Nodes, NodeRef{Slug: e.Name(), Name: name})
	}
	dash.NodeCount = len(dash.Nodes)
	if dash.Nodes == nil {
		dash.Nodes = []NodeRef{}
	}

	// ── command center ────────────────────────────────────────────────────────
	dash.CommandMD, _ = readFileOrEmpty(filepath.Join(nodesRoot, "00-command", "context.md"))

	// ── revenue ───────────────────────────────────────────────────────────────
	rev, _ := GetRevenueStreams(nodesRoot)
	dash.TotalMRR = rev.TotalMRR

	// ── signal count ─────────────────────────────────────────────────────────
	dash.SignalCount = countAllSignals(nodesRoot, entries)

	// ── SQLite entity/edge counts (optional) ─────────────────────────────────
	if dbPath != "" {
		dash.EntityCount, dash.EdgeCount = queryDBCounts(dbPath)
	}

	return dash, nil
}

// countAllSignals tallies .md files across all nodes/*/signals/ directories.
func countAllSignals(nodesRoot string, entries []os.DirEntry) int {
	total := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		signalsDir := filepath.Join(nodesRoot, e.Name(), "signals")
		sEntries, err := os.ReadDir(signalsDir)
		if err != nil {
			continue
		}
		for _, s := range sEntries {
			if !s.IsDir() && strings.HasSuffix(s.Name(), ".md") {
				total++
			}
		}
	}
	return total
}

// queryDBCounts returns entity and edge counts from SQLite, returning (0,0) on
// any error so callers always get a usable value.
func queryDBCounts(dbPath string) (entityCount, edgeCount int) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, 0
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return 0, 0
	}
	_ = db.QueryRow("SELECT COUNT(*) FROM entities").Scan(&entityCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&edgeCount)
	return entityCount, edgeCount
}

// ── Health Check ─────────────────────────────────────────────────────────────

// HealthCheck represents a single named diagnostic check.
type HealthCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// HealthStatus is the payload returned by GetHealthCheck.
type HealthStatus struct {
	Status       string        `json:"status"` // "healthy" | "degraded" | "error"
	Checks       []HealthCheck `json:"checks"`
	ContextCount int           `json:"context_count"`
	SignalCount  int           `json:"signal_count"`
}

// GetHealthCheck runs a series of diagnostic checks against the OptimalOS
// filesystem and (optionally) the SQLite index. Returns a HealthStatus with
// overall "healthy", "degraded", or "error" status.
func GetHealthCheck(nodesRoot, osRoot, dbPath string) (HealthStatus, error) {
	var checks []HealthCheck
	failCount := 0

	addCheck := func(name string, ok bool, detail string) {
		checks = append(checks, HealthCheck{Name: name, OK: ok, Detail: detail})
		if !ok {
			failCount++
		}
	}

	// ── nodes dir ────────────────────────────────────────────────────────────
	if _, err := os.Stat(nodesRoot); err == nil {
		addCheck("nodes_dir", true, nodesRoot)
	} else {
		addCheck("nodes_dir", false, fmt.Sprintf("not found: %s", nodesRoot))
	}

	// ── rhythm dir ───────────────────────────────────────────────────────────
	rhythmDir := filepath.Join(osRoot, "rhythm")
	if _, err := os.Stat(rhythmDir); err == nil {
		addCheck("rhythm_dir", true, rhythmDir)
	} else {
		addCheck("rhythm_dir", false, fmt.Sprintf("not found: %s", rhythmDir))
	}

	// ── SQLite db ─────────────────────────────────────────────────────────────
	var contextCount, signalCount int
	if dbPath == "" {
		addCheck("sqlite_db", false, "OPTIMAL_DB_PATH not configured")
	} else if _, err := os.Stat(dbPath); err != nil {
		addCheck("sqlite_db", false, fmt.Sprintf("not found: %s", dbPath))
	} else {
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			addCheck("sqlite_db", false, err.Error())
		} else {
			defer db.Close()
			if pingErr := db.Ping(); pingErr != nil {
				addCheck("sqlite_db", false, pingErr.Error())
			} else {
				addCheck("sqlite_db", true, dbPath)
				_ = db.QueryRow("SELECT COUNT(*) FROM contexts").Scan(&contextCount)
				_ = db.QueryRow("SELECT COUNT(*) FROM entities").Scan(&signalCount)
			}
		}
	}

	// ── derive overall status ─────────────────────────────────────────────────
	status := "healthy"
	if failCount > 0 {
		// Degrade rather than error if at least the nodes dir is present.
		if checks[0].OK {
			status = "degraded"
		} else {
			status = "error"
		}
	}

	return HealthStatus{
		Status:       status,
		Checks:       checks,
		ContextCount: contextCount,
		SignalCount:  signalCount,
	}, nil
}
