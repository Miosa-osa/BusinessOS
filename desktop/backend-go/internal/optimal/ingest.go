package optimal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required
)

// IndexParams holds all columns required for a full contexts INSERT.
type IndexParams struct {
	ID         string
	URI        string
	Type       string
	Path       string
	Title      string
	Content    string
	L0Abstract string
	L1Overview string
	Mode       string
	Genre      string
	SignalType string
	Format     string
	Structure  string
	Node       string
	SNRatio    float64
	Entities   string // JSON array string
	RoutedTo   string // JSON array string
	Metadata   string // JSON object string
}

// Ingest writes a new signal to the filesystem and indexes it in SQLite.
// It runs the full pipeline: Classify → Route → generate abstracts → write file
// → index with all 22 columns → extract entities → create edges.
func Ingest(osRoot, nodesRoot, dbPath, text, genre string) error {
	if text == "" {
		return fmt.Errorf("optimal/ingest: text must not be empty")
	}

	// 1. CLASSIFY
	classification := Classify(text)
	if genre == "" {
		genre = classification.Genre
	}

	// 2. ROUTE
	routing := Route(text)
	// Resolve the short numeric node ID (e.g. "09") to the full directory name
	// (e.g. "09-new-stuff"). When nodesRoot is set we scan for the matching dir;
	// otherwise we use the canonical fallback map so indexing still works.
	targetNode := resolveNodeDir(nodesRoot, routing.PrimaryNode)

	// 3. Generate L0/L1 abstracts
	l0 := L0Abstract(text)
	l1 := L1Overview(text)

	// 4. Write markdown file to correct node
	now := time.Now()
	slug := makeSlug(text, 50)
	filename := now.Format("2006-01-02") + "-" + slug + ".md"

	if nodesRoot != "" {
		content := buildSignalContent(genre, classification, routing, now, text)
		signalDir := filepath.Join(nodesRoot, targetNode, "signals")
		if err := os.MkdirAll(signalDir, 0o755); err != nil {
			return fmt.Errorf("optimal/ingest: mkdir signals: %w", err)
		}
		dest := filepath.Join(signalDir, filename)
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			return fmt.Errorf("optimal/ingest: write signal file: %w", err)
		}
	}

	// 5. Index in SQLite with FULL schema
	if dbPath != "" {
		uri := fmt.Sprintf("optimal://nodes/%s/signals/%s", targetNode, filename)
		path := filepath.Join("nodes", targetNode, "signals", filename)
		title := extractTitle(text)

		params := IndexParams{
			ID:         slug,
			URI:        uri,
			Type:       "signal",
			Path:       path,
			Title:      title,
			Content:    text,
			L0Abstract: l0,
			L1Overview: l1,
			Mode:       classification.Mode,
			Genre:      genre,
			SignalType: classification.Type,
			Format:     classification.Format,
			Structure:  classification.Structure,
			Node:       targetNode,
			SNRatio:    classification.SNRatio,
			Entities:   "[]",
			RoutedTo:   marshalJSON(routing.CrossRefNodes),
			Metadata:   "{}",
		}

		if err := indexSignalFull(dbPath, params, now); err != nil {
			// Non-fatal: file is already on disk.
			return fmt.Errorf("optimal/ingest: index in SQLite: %w", err)
		}

		// 6. Extract entities and create edges
		entities := extractEntitiesFromText(text)
		for _, entity := range entities {
			_ = insertEntity(dbPath, slug, entity, "person")
		}

		// 7. Create edges
		//   - lives_in: signal lives in its target node
		//   - cross_ref: signal cross-references other nodes
		_ = insertEdge(dbPath, slug, targetNode, "lives_in", 1.0)
		for _, xref := range routing.CrossRefNodes {
			_ = insertEdge(dbPath, slug, xref, "cross_ref", 1.2)
		}
	}

	return nil
}

// buildSignalContent formats the signal file with full YAML frontmatter
// matching the Elixir engine schema.
func buildSignalContent(genre string, cl SignalClassification, r RoutingResult, now time.Time, text string) string {
	crossRef := ""
	if len(r.CrossRefNodes) > 0 {
		crossRef = strings.Join(r.CrossRefNodes, ", ")
	}
	return fmt.Sprintf("---\nsignal:\n  mode: %s\n  genre: %s\n  type: %s\n  format: %s\n  structure: %s\n  sn_ratio: %.2f\nnode: %s\ncross_ref: %s\nvalid_from: \"%s\"\n---\n\n%s\n",
		cl.Mode, genre, cl.Type, cl.Format,
		cl.Structure, cl.SNRatio, r.PrimaryNode,
		crossRef, now.Format("2006-01-02"), text,
	)
}

// indexSignalFull inserts a signal into the contexts table with all 22 columns
// matching the real Elixir engine schema.
func indexSignalFull(dbPath string, p IndexParams, now time.Time) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	nowStr := now.UTC().Format(time.RFC3339)

	_, err = db.Exec(`
		INSERT OR IGNORE INTO contexts
			(id, uri, type, path, title, l0_abstract, l1_overview, content,
			 mode, genre, signal_type, format, structure, node, sn_ratio,
			 entities, created_at, modified_at, valid_from, routed_to, metadata, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.URI, p.Type, p.Path, p.Title,
		p.L0Abstract, p.L1Overview, p.Content,
		p.Mode, p.Genre, p.SignalType, p.Format, p.Structure,
		p.Node, p.SNRatio,
		p.Entities, nowStr, nowStr, nowStr,
		p.RoutedTo, p.Metadata, nowStr,
	)
	if err != nil {
		return fmt.Errorf("insert context: %w", err)
	}
	return nil
}

// extractEntitiesFromText finds proper nouns (capitalized words not at sentence
// starts) and returns a deduplicated list. Simple regex — no NLP required.
func extractEntitiesFromText(text string) []string {
	// Match capitalized words that follow a space (not sentence-start).
	// e.g. "Ed called" → "Ed"; "Bennett closed" → "Bennett"
	re := regexp.MustCompile(`(?:^|\s)([A-Z][a-z]{2,})(?:\s+[A-Z][a-z]{2,})*`)
	rawMatches := re.FindAllStringSubmatch(text, -1)

	// Also scan for multi-word proper noun phrases (e.g. "Ed Honour", "AI Masters").
	rePhrase := regexp.MustCompile(`\b[A-Z][a-z]{1,}(?:\s+[A-Z][a-z]{1,})+\b`)
	phraseMatches := rePhrase.FindAllString(text, -1)

	seen := make(map[string]struct{})
	var result []string

	// Skip common sentence-start words and stop words.
	stopWords := map[string]struct{}{
		"The": {}, "A": {}, "An": {}, "In": {}, "On": {}, "At": {},
		"To": {}, "For": {}, "Of": {}, "And": {}, "Or": {}, "But": {},
		"It": {}, "He": {}, "She": {}, "We": {}, "They": {}, "You": {},
		"I": {}, "This": {}, "That": {}, "These": {}, "Those": {},
	}

	for _, m := range rawMatches {
		if len(m) < 2 {
			continue
		}
		word := strings.TrimSpace(m[1])
		if _, stop := stopWords[word]; stop {
			continue
		}
		if _, ok := seen[word]; !ok {
			seen[word] = struct{}{}
			result = append(result, word)
		}
	}

	for _, phrase := range phraseMatches {
		phrase = strings.TrimSpace(phrase)
		if _, ok := seen[phrase]; !ok {
			seen[phrase] = struct{}{}
			result = append(result, phrase)
		}
	}

	return result
}

// insertEntity inserts an entity into the entities table linked to a context.
func insertEntity(dbPath, contextID, name, entityType string) error {
	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("optimal/ingest: open db for entity: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT OR IGNORE INTO entities (context_id, name, type)
		VALUES (?, ?, ?)`,
		contextID, name, entityType,
	)
	if err != nil {
		return fmt.Errorf("optimal/ingest: insert entity %q: %w", name, err)
	}
	return nil
}

// insertEdge inserts a directed edge into the edges table.
func insertEdge(dbPath, sourceID, targetID, relation string, weight float64) error {
	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("optimal/ingest: open db for edge: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT OR IGNORE INTO edges (source_id, target_id, relation, weight)
		VALUES (?, ?, ?, ?)`,
		sourceID, targetID, relation, weight,
	)
	if err != nil {
		return fmt.Errorf("optimal/ingest: insert edge %q→%q: %w", sourceID, targetID, err)
	}
	return nil
}

// extractTitle derives a human-readable title from the first sentence of text.
// Falls back to the slug if text is empty.
func extractTitle(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "untitled"
	}
	// Use the first sentence (up to '.', '!', '?' or 80 chars).
	s := firstSentence(text)
	if len(s) > 80 {
		s = s[:80]
		if idx := strings.LastIndexByte(s, ' '); idx > 40 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}

// marshalJSON encodes a value to JSON string, returning "[]" on failure.
func marshalJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// formatCrossRefs returns a comma-separated string of cross-reference nodes.
func formatCrossRefs(nodes []string) string {
	return strings.Join(nodes, ", ")
}

// nodeCanonicalNames maps short numeric node IDs to the canonical directory name.
// Used as fallback when nodesRoot is empty or the directory doesn't exist yet.
var nodeCanonicalNames = map[string]string{
	"00": "00-command",
	"01": "01-roberto",
	"02": "02-miosa",
	"03": "03-lunivate",
	"04": "04-ai-masters",
	"05": "05-os-architect",
	"06": "06-agency-accelerants",
	"07": "07-accelerants-community",
	"08": "08-content-creators",
	"09": "09-new-stuff",
	"10": "10-team",
	"11": "11-money-revenue",
	"12": "12-os-accelerator",
}

// resolveNodeDir maps a short numeric node ID (e.g. "09") to the full directory
// name (e.g. "09-new-stuff"). When nodesRoot is set it finds the first directory
// whose name starts with the numeric prefix; otherwise it uses the canonical map.
// If neither source resolves the ID, the ID itself is returned unchanged.
func resolveNodeDir(nodesRoot, nodeID string) string {
	if nodesRoot != "" {
		entries, err := os.ReadDir(nodesRoot)
		if err == nil {
			prefix := nodeID + "-"
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
					return e.Name()
				}
				// Exact match (e.g. directory already named "09").
				if e.IsDir() && e.Name() == nodeID {
					return e.Name()
				}
			}
		}
	}
	if name, ok := nodeCanonicalNames[nodeID]; ok {
		return name
	}
	return nodeID
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
