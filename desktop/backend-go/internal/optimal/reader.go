// Package optimal provides a read-only bridge between the BusinessOS Go backend
// and the OptimalOS filesystem (markdown nodes at nodes/, rhythm files at rhythm/).
//
// File layout:
//   - reader.go  — Pure filesystem reads: nodes, signals, rhythm files
//   - engine.go  — Subprocess wrapper for the Elixir mix optimal.* CLI
package optimal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
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
