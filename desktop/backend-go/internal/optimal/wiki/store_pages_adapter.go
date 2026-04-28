package wiki

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// PagesAdapter implements Store on top of the existing BusinessOS `contexts`
// table — the storage layer for the Pages module. We deliberately do NOT
// create a parallel `wiki_pages` table: the Pages module already owns
// document storage with block editing, sharing, and the knowledge graph.
//
// Mapping:
//
//	Page.Slug          ↔ contexts.share_id  (when present)  | slugified contexts.name
//	Page.Body          ↔ contexts.content   (markdown view of blocks)
//	Page.Frontmatter   ↔ contexts.properties JSONB           (.wiki.* keys)
//	Page.Audience      ↔ contexts.properties.wiki_audience
//	Page.Version       ↔ contexts.properties.wiki_version
//	Page.LastCurated   ↔ contexts.properties.wiki_last_curated
//	Page.CuratedBy     ↔ contexts.properties.wiki_curated_by
//	Page.TenantID      ↔ contexts.user_id   (until Wave 4 maps workspaces)
//
// Citations live in their own `wiki_citations` table (created by EnsureSchema)
// because the contexts table has no place for them.
type PagesAdapter struct {
	DB *sql.DB
}

// NewPagesAdapter wires the adapter to an existing *sql.DB. Callers should
// invoke EnsureSchema once at startup so the citations table exists.
func NewPagesAdapter(db *sql.DB) *PagesAdapter { return &PagesAdapter{DB: db} }

// EnsureSchema creates the wiki_citations table if missing. Idempotent.
// The contexts table is owned by the Pages migrations; we do not touch it.
func (a *PagesAdapter) EnsureSchema(ctx context.Context) error {
	const ddl = `
	CREATE TABLE IF NOT EXISTS wiki_citations (
		id             BIGSERIAL PRIMARY KEY,
		tenant_id      TEXT        NOT NULL,
		slug           TEXT        NOT NULL,
		audience       TEXT        NOT NULL DEFAULT 'default',
		chunk_id       TEXT        NOT NULL,
		claim_hash     TEXT        NOT NULL,
		last_verified  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (tenant_id, slug, audience, chunk_id, claim_hash)
	);
	CREATE INDEX IF NOT EXISTS idx_wiki_citations_page ON wiki_citations(tenant_id, slug, audience);
	`
	if _, err := a.DB.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("wiki/pages adapter: ensure schema: %w", err)
	}
	return nil
}

// Put upserts a wiki page into contexts.properties, bumping wiki_version.
// The body is written to contexts.content; existing blocks are left intact
// so the editor view continues working.
func (a *PagesAdapter) Put(ctx context.Context, page *Page) error {
	if page == nil || page.Slug == "" {
		return ErrMissingSlug
	}
	props := wikiProperties(page)
	propsJSON, err := json.Marshal(props)
	if err != nil {
		return fmt.Errorf("wiki/pages adapter: marshal properties: %w", err)
	}

	// We try to update an existing context (matched by share_id == slug OR
	// slugified name), and INSERT only when no match exists.
	const update = `
		UPDATE contexts
		SET content = $3,
		    properties = COALESCE(properties, '{}'::jsonb) || $4::jsonb,
		    last_edited_at = NOW(),
		    updated_at = NOW()
		WHERE user_id = $1
		  AND (share_id = $2 OR LOWER(REPLACE(name, ' ', '-')) = $2)
		RETURNING id
	`
	var id string
	err = a.DB.QueryRowContext(ctx, update, page.TenantID, page.Slug, page.Body, string(propsJSON)).Scan(&id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("wiki/pages adapter: update: %w", err)
	}

	// No existing context — insert a new wiki-tagged one.
	const insert = `
		INSERT INTO contexts (user_id, name, type, content, properties, share_id, last_edited_at)
		VALUES ($1, $2, 'CUSTOM', $3, $4::jsonb, $5, NOW())
	`
	if _, err := a.DB.ExecContext(ctx, insert,
		page.TenantID, slugToName(page.Slug), page.Body, string(propsJSON), page.Slug,
	); err != nil {
		return fmt.Errorf("wiki/pages adapter: insert: %w", err)
	}
	return nil
}

// Latest fetches the most recently edited wiki page matching (tenantID, slug, audience).
// Audience is matched against properties.wiki_audience; missing audience falls
// back to "default" pages.
func (a *PagesAdapter) Latest(ctx context.Context, tenantID, slug, audience string) (*Page, error) {
	if audience == "" {
		audience = DefaultAudience
	}
	const q = `
		SELECT name, COALESCE(content, ''), COALESCE(properties, '{}'::jsonb), updated_at
		FROM contexts
		WHERE user_id = $1
		  AND (share_id = $2 OR LOWER(REPLACE(name, ' ', '-')) = $2)
		  AND COALESCE(properties->>'wiki_audience', 'default') = $3
		ORDER BY updated_at DESC
		LIMIT 1
	`
	var name, content, propsJSON string
	var updated time.Time
	if err := a.DB.QueryRowContext(ctx, q, tenantID, slug, audience).Scan(&name, &content, &propsJSON, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("wiki/pages adapter: latest: %w", err)
	}
	return rowToPage(tenantID, slug, audience, name, content, propsJSON, updated), nil
}

// List returns every wiki-tagged page for a tenant, latest version of each
// (slug, audience) pair. We treat any context row with a `wiki_version`
// property as a wiki page.
func (a *PagesAdapter) List(ctx context.Context, tenantID string) ([]*Page, error) {
	const q = `
		SELECT name, share_id, COALESCE(content, ''), COALESCE(properties, '{}'::jsonb), updated_at
		FROM contexts
		WHERE user_id = $1
		  AND properties ? 'wiki_version'
		ORDER BY updated_at DESC
	`
	rows, err := a.DB.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("wiki/pages adapter: list: %w", err)
	}
	defer rows.Close()

	out := []*Page{}
	for rows.Next() {
		var name, content, propsJSON string
		var shareID sql.NullString
		var updated time.Time
		if err := rows.Scan(&name, &shareID, &content, &propsJSON, &updated); err != nil {
			return nil, fmt.Errorf("wiki/pages adapter: scan: %w", err)
		}
		slug := shareID.String
		if slug == "" {
			slug = nameToSlug(name)
		}
		props := decodeProps(propsJSON)
		audience := stringFromAny(props["wiki_audience"])
		if audience == "" {
			audience = DefaultAudience
		}
		out = append(out, rowToPage(tenantID, slug, audience, name, content, propsJSON, updated))
	}
	return out, rows.Err()
}

// PutCitations bulk-upserts citation rows for (tenantID, slug, audience).
// Each citation must include a non-empty ChunkID; the ClaimHash defaults to
// the ChunkID when callers haven't computed one yet (rough parity with the
// Elixir source's derive_claim_hash fallback).
func (a *PagesAdapter) PutCitations(ctx context.Context, tenantID, slug, audience string, citations []Citation) error {
	if audience == "" {
		audience = DefaultAudience
	}
	if len(citations) == 0 {
		return nil
	}
	const upsert = `
		INSERT INTO wiki_citations (tenant_id, slug, audience, chunk_id, claim_hash, last_verified)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (tenant_id, slug, audience, chunk_id, claim_hash) DO UPDATE SET
			last_verified = EXCLUDED.last_verified
	`
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("wiki/pages adapter: begin: %w", err)
	}
	defer tx.Rollback()
	for _, c := range citations {
		if c.ChunkID == "" {
			continue
		}
		hash := c.ClaimHash
		if hash == "" {
			hash = c.ChunkID
		}
		if _, err := tx.ExecContext(ctx, upsert, tenantID, slug, audience, c.ChunkID, hash); err != nil {
			return fmt.Errorf("wiki/pages adapter: upsert citation: %w", err)
		}
	}
	return tx.Commit()
}

// ListCitations returns every citation for (tenantID, slug, audience).
func (a *PagesAdapter) ListCitations(ctx context.Context, tenantID, slug, audience string) ([]Citation, error) {
	if audience == "" {
		audience = DefaultAudience
	}
	const q = `
		SELECT chunk_id, claim_hash, last_verified
		FROM wiki_citations
		WHERE tenant_id = $1 AND slug = $2 AND audience = $3
		ORDER BY last_verified DESC
	`
	rows, err := a.DB.QueryContext(ctx, q, tenantID, slug, audience)
	if err != nil {
		return nil, fmt.Errorf("wiki/pages adapter: list citations: %w", err)
	}
	defer rows.Close()

	out := []Citation{}
	for rows.Next() {
		var c Citation
		var ts time.Time
		if err := rows.Scan(&c.ChunkID, &c.ClaimHash, &ts); err != nil {
			return nil, err
		}
		c.LastVerified = ts.UTC().Format(time.RFC3339)
		out = append(out, c)
	}
	return out, rows.Err()
}

// rowToPage materializes a Page from contexts row columns.
func rowToPage(tenantID, slug, audience, name, content, propsJSON string, updated time.Time) *Page {
	props := decodeProps(propsJSON)
	page := &Page{
		TenantID:    tenantID,
		Slug:        slug,
		Audience:    audience,
		Version:     intFromAny(props["wiki_version"], 1),
		Frontmatter: filterFrontmatter(props),
		Body:        content,
		LastCurated: stringFromAny(props["wiki_last_curated"]),
		CuratedBy:   stringFromAny(props["wiki_curated_by"]),
	}
	if page.LastCurated == "" {
		page.LastCurated = updated.UTC().Format(time.RFC3339)
	}
	// Display name useful for UIs that don't surface the slug.
	if name != "" {
		page.Frontmatter["title"] = name
	}
	return page
}

// wikiProperties extracts the wiki-relevant keys out of the page so we can
// merge them into contexts.properties without clobbering Pages-owned keys.
func wikiProperties(page *Page) map[string]any {
	out := map[string]any{
		"wiki_version":  page.Version,
		"wiki_audience": page.Audience,
	}
	if page.LastCurated != "" {
		out["wiki_last_curated"] = page.LastCurated
	}
	if page.CuratedBy != "" {
		out["wiki_curated_by"] = page.CuratedBy
	}
	for k, v := range page.Frontmatter {
		if strings.HasPrefix(k, "wiki_") {
			out[k] = v
		}
	}
	return out
}

// filterFrontmatter exposes wiki_* properties as Page.Frontmatter, stripping
// the prefix so callers see the same keys the Elixir source emits.
func filterFrontmatter(props map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range props {
		if strings.HasPrefix(k, "wiki_") {
			out[strings.TrimPrefix(k, "wiki_")] = v
		}
	}
	return out
}

func decodeProps(propsJSON string) map[string]any {
	out := map[string]any{}
	if propsJSON == "" {
		return out
	}
	_ = json.Unmarshal([]byte(propsJSON), &out)
	return out
}

var spaceRe = regexp.MustCompile(`\s+`)

func nameToSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = spaceRe.ReplaceAllString(s, "-")
	return s
}

func slugToName(slug string) string {
	s := strings.ReplaceAll(slug, "-", " ")
	if s == "" {
		return "Untitled"
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
