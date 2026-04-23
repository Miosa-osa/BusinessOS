package optimal

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReindexResult holds the outcome of a full reindex operation.
type ReindexResult struct {
	FilesScanned  int      `json:"files_scanned"`
	FilesIndexed  int      `json:"files_indexed"`
	FilesSkipped  int      `json:"files_skipped"`  // non-.md or unreadable
	FilesRejected int      `json:"files_rejected"` // S/N < 0.3
	Errors        []string `json:"errors"`
	DurationMs    int64    `json:"duration_ms"`
}

// Reindex walks the entire nodes/ directory tree, reads every .md file,
// parses frontmatter, classifies content, and UPSERTs into SQLite.
// This rebuilds the entire search index from the filesystem source of truth.
//
// Individual file errors are collected and returned in ReindexResult.Errors
// so one bad file does not abort the entire run. A non-nil error return is
// reserved for fatal setup failures (schema, db open).
func Reindex(nodesRoot, dbPath string) (*ReindexResult, error) {
	start := time.Now()
	result := &ReindexResult{}

	// 1. Ensure schema exists (idempotent).
	if dbPath != "" {
		if err := EnsureSchema(dbPath); err != nil {
			return nil, fmt.Errorf("optimal/reindex: ensure schema: %w", err)
		}
	}

	// 2. Walk the nodes/ directory recursively.
	walkErr := filepath.WalkDir(nodesRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("walk error at %s: %v", path, err))
			return nil // continue walking
		}

		// Only process regular .md files.
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		result.FilesScanned++

		if indexErr := reindexFile(nodesRoot, dbPath, path, result); indexErr != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("%s: %v", relOrAbs(nodesRoot, path), indexErr))
		}

		return nil
	})

	if walkErr != nil {
		return result, fmt.Errorf("optimal/reindex: walk %q: %w", nodesRoot, walkErr)
	}

	result.DurationMs = time.Since(start).Milliseconds()
	slog.Info("optimal/reindex: complete",
		"scanned", result.FilesScanned,
		"indexed", result.FilesIndexed,
		"skipped", result.FilesSkipped,
		"rejected", result.FilesRejected,
		"errors", len(result.Errors),
		"duration_ms", result.DurationMs,
	)
	return result, nil
}

// reindexFile processes a single markdown file: parse, classify, quality-gate, UPSERT.
func reindexFile(nodesRoot, dbPath, absPath string, result *ReindexResult) error {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		result.FilesSkipped++
		return fmt.Errorf("read file: %w", err)
	}

	content := string(raw)
	if strings.TrimSpace(content) == "" {
		result.FilesSkipped++
		return nil
	}

	// 3a. Parse frontmatter.
	fm, body := ParseFrontmatter(content)

	// 3b. Determine classification: use parsed dims when available,
	//     otherwise run the classifier on the body text.
	var cl SignalClassification
	if fm.HasSignalDims() {
		cl = SignalClassification{
			Mode:      fm.Mode,
			Genre:     fm.Genre,
			Type:      fm.Type,
			Format:    fm.Format,
			Structure: fm.Structure,
			SNRatio:   fm.SNRatio,
		}
		// Fill any missing dimensions from the classifier.
		if cl.Mode == "" || cl.Genre == "" || cl.SNRatio == 0 {
			auto := Classify(body)
			if cl.Mode == "" {
				cl.Mode = auto.Mode
			}
			if cl.Genre == "" {
				cl.Genre = auto.Genre
			}
			if cl.Type == "" {
				cl.Type = auto.Type
			}
			if cl.Format == "" {
				cl.Format = auto.Format
			}
			if cl.Structure == "" {
				cl.Structure = auto.Structure
			}
			if cl.SNRatio == 0 {
				cl.SNRatio = auto.SNRatio
			}
		}
	} else {
		cl = Classify(body)
		// Re-classify on full content when no frontmatter is present so format
		// detection (YAML frontmatter check) still fires correctly.
		if body == content {
			cl = Classify(content)
		}
	}

	// 3c. Genre override from frontmatter.
	genre := fm.Genre
	if genre == "" {
		genre = cl.Genre
	}

	// 3e. S/N quality gate — skip files below the minimum threshold.
	if cl.SNRatio < 0.3 {
		result.FilesRejected++
		slog.Debug("optimal/reindex: file rejected — S/N below threshold",
			"path", absPath, "sn_ratio", cl.SNRatio)
		_ = LogDecision(dbPath, "reindex", "reject", "context", slugFromPath(absPath),
			fmt.Sprintf("sn_ratio=%.2f < 0.3", cl.SNRatio),
			"signal quality below minimum threshold")
		return nil
	}

	// 3d. Generate L0/L1 abstracts.
	l0 := L0Abstract(content)
	l1 := L1Overview(content)

	// Derive slug, title, node from path.
	slug := slugFromPath(absPath)
	title := fm.Title
	if title == "" {
		title = extractTitle(body)
	}
	if title == "" || title == "untitled" {
		title = slugToTitle(filepath.Base(absPath[:len(filepath.Base(absPath))-3]))
	}

	node := fm.Node
	if node == "" {
		node = nodeFromPath(nodesRoot, absPath)
	}

	// Build relative path for the index.
	relPath, _ := filepath.Rel(filepath.Dir(nodesRoot), absPath)
	if relPath == "" {
		relPath = absPath
	}

	uri := "optimal://" + filepath.ToSlash(relPath)

	// Build cross-refs.
	crossRefNodes := fm.CrossRef
	if len(crossRefNodes) == 0 && node != "" {
		// No frontmatter cross-refs — create a minimal live-in edge.
		crossRefNodes = nil
	}

	validFrom := fm.ValidFrom
	if validFrom == "" {
		validFrom = time.Now().UTC().Format(time.RFC3339)
	}

	now := time.Now()
	nowStr := now.UTC().Format(time.RFC3339)

	params := IndexParams{
		ID:         slug,
		URI:        uri,
		Type:       "signal",
		Path:       filepath.ToSlash(relPath),
		Title:      title,
		Content:    content,
		L0Abstract: l0,
		L1Overview: l1,
		Mode:       cl.Mode,
		Genre:      genre,
		SignalType: cl.Type,
		Format:     cl.Format,
		Structure:  cl.Structure,
		Node:       node,
		SNRatio:    cl.SNRatio,
		Entities:   "[]",
		RoutedTo:   marshalJSON(crossRefNodes),
		Metadata:   "{}",
	}

	// 3f. UPSERT into contexts (INSERT OR REPLACE).
	if dbPath != "" {
		if err := indexSignalFullAt(dbPath, params, nowStr, validFrom); err != nil {
			result.FilesSkipped++
			return fmt.Errorf("index: %w", err)
		}

		// 3g. Extract entities and create edges.
		entities := extractEntitiesFromText(body)
		for _, entity := range entities {
			_ = insertEntity(dbPath, slug, entity, "person")
		}
		_ = insertEdge(dbPath, slug, node, "lives_in", 1.0)
		for _, xref := range crossRefNodes {
			_ = insertEdge(dbPath, slug, xref, "cross_ref", 1.2)
		}

		// 3h. Log to decision_log.
		_ = LogDecision(dbPath, "reindex", "upsert", "context", slug,
			fmt.Sprintf("indexed from filesystem: %s", filepath.ToSlash(relPath)),
			fmt.Sprintf("reindex run at %s", nowStr),
		)
	}

	result.FilesIndexed++
	return nil
}

// indexSignalFullAt is a variant of indexSignalFull that accepts pre-computed
// timestamp strings and a separate valid_from value derived from frontmatter.
// Using INSERT OR REPLACE keeps existing created_at when a row already exists.
func indexSignalFullAt(dbPath string, p IndexParams, nowStr, validFrom string) error {
	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	// Preserve created_at for existing rows by reading it first.
	var existingCreatedAt string
	_ = db.QueryRow(`SELECT created_at FROM contexts WHERE id = ?`, p.ID).Scan(&existingCreatedAt)
	createdAt := existingCreatedAt
	if createdAt == "" {
		createdAt = nowStr
	}

	_, err = db.Exec(`
		INSERT OR REPLACE INTO contexts
			(id, uri, type, path, title, l0_abstract, l1_overview, content,
			 mode, genre, signal_type, format, structure, node, sn_ratio,
			 entities, created_at, modified_at, valid_from, routed_to, metadata, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.URI, p.Type, p.Path, p.Title,
		p.L0Abstract, p.L1Overview, p.Content,
		p.Mode, p.Genre, p.SignalType, p.Format, p.Structure,
		p.Node, p.SNRatio,
		p.Entities, createdAt, nowStr, validFrom,
		p.RoutedTo, p.Metadata, nowStr,
	)
	if err != nil {
		return fmt.Errorf("upsert context: %w", err)
	}
	return nil
}

// ── path helpers ──────────────────────────────────────────────────────────────

// slugFromPath derives a unique slug from a file path. It uses the file base
// name without extension, falling back to the full path if the base is empty.
func slugFromPath(absPath string) string {
	base := filepath.Base(absPath)
	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return makeSlug(base, 80)
}

// nodeFromPath extracts the node directory name (e.g. "04-ai-masters") from an
// absolute path by finding the first directory segment under nodesRoot.
func nodeFromPath(nodesRoot, absPath string) string {
	rel, err := filepath.Rel(nodesRoot, absPath)
	if err != nil || rel == "" {
		return ""
	}
	parts := strings.SplitN(filepath.ToSlash(rel), "/", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// relOrAbs returns a path relative to base, or absPath on error.
func relOrAbs(base, absPath string) string {
	rel, err := filepath.Rel(base, absPath)
	if err != nil {
		return absPath
	}
	return rel
}
