package wiki

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// Wiki is the top-level facade — the equivalent of OptimalEngine.Wiki in
// Elixir. Most call sites should use this rather than touching Store /
// Directives / Integrity directly so behaviour stays uniform.
type Wiki struct {
	Store Store
}

// New returns a Wiki bound to the given Store.
func New(s Store) *Wiki { return &Wiki{Store: s} }

// Get returns the latest page for (tenantID, slug, audience) or ErrNotFound.
func (w *Wiki) Get(ctx context.Context, tenantID, slug, audience string) (*Page, error) {
	if audience == "" {
		audience = DefaultAudience
	}
	return w.Store.Latest(ctx, tenantID, slug, audience)
}

// Put persists a new page version.
func (w *Wiki) Put(ctx context.Context, page *Page) error { return w.Store.Put(ctx, page) }

// List returns every wiki page for a tenant (latest version per slug+audience).
func (w *Wiki) List(ctx context.Context, tenantID string) ([]*Page, error) {
	return w.Store.List(ctx, tenantID)
}

// Curate is the lightweight Go equivalent of Wiki.curate/3. The full Elixir
// curator drives an LLM round-trip; that part lives in the existing
// BusinessOS chat/agent pipeline. Here we expose the data half: bump the
// version, persist the new body + citations, and stamp curated_by + ts.
//
// The CurateFn callback receives the existing page (or nil for new) and
// returns the new body. Callers wire CurateFn to whatever LLM they want.
type CurateFn func(ctx context.Context, existing *Page, citations []Citation) (string, error)

// CurateOptions tune a Curate call.
type CurateOptions struct {
	CuratedBy string // user / agent ID; defaults to "system"
	Audience  string // defaults to DefaultAudience
}

// Curate is the high-level write path. It:
//  1. Loads any existing page for (tenantID, slug, audience).
//  2. Calls curate to compute the new body.
//  3. Persists the new page with version = existing.version + 1 (or 1).
//  4. Persists the citation rows (idempotent — replaying same set is safe).
//
// Returns the persisted Page so callers can publish a "wiki.page.curated"
// audit event with stable metadata.
func (w *Wiki) Curate(ctx context.Context, tenantID, slug string, citations []Citation, curate CurateFn, opts CurateOptions) (*Page, error) {
	if opts.Audience == "" {
		opts.Audience = DefaultAudience
	}
	if opts.CuratedBy == "" {
		opts.CuratedBy = "system"
	}

	existing, err := w.Store.Latest(ctx, tenantID, slug, opts.Audience)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	body, err := curate(ctx, existing, citations)
	if err != nil {
		return nil, err
	}

	version := 1
	frontmatter := map[string]any{}
	if existing != nil {
		version = existing.Version + 1
		frontmatter = cloneMap(existing.Frontmatter)
	}

	page := &Page{
		TenantID:    tenantID,
		Slug:        slug,
		Audience:    opts.Audience,
		Version:     version,
		Frontmatter: frontmatter,
		Body:        body,
		LastCurated: time.Now().UTC().Format(time.RFC3339),
		CuratedBy:   opts.CuratedBy,
	}

	if err := w.Store.Put(ctx, page); err != nil {
		return nil, err
	}

	// Auto-derive claim hashes for citations missing one. Hash inputs are
	// the chunk URI + first 256 bytes of body, so re-curates with the same
	// citations on identical bodies de-duplicate naturally.
	for i, c := range citations {
		if c.ClaimHash == "" {
			citations[i].ClaimHash = ClaimHash(c.ChunkID, body)
		}
	}
	if err := w.Store.PutCitations(ctx, tenantID, slug, opts.Audience, citations); err != nil {
		return nil, err
	}
	return page, nil
}

// Render is a thin wrapper around the package-level Render so callers can use
// the facade exclusively.
func (w *Wiki) Render(page *Page, resolver Resolver, format Format) RenderResult {
	return Render(page.Body, resolver, format)
}

// Verify runs the full integrity suite against a page.
func (w *Wiki) Verify(page *Page, opts CheckOptions) Report { return Check(page, opts) }

// VerifyAgainstSchema layers schema rules on top of Verify.
func (w *Wiki) VerifyAgainstSchema(page *Page, schema SchemaRules, opts CheckOptions) Report {
	return AgainstSchema(page, schema, opts)
}

// ClaimHash derives a stable claim hash from a chunk URI + body sample. Used
// when callers don't pre-compute one — matches the Elixir source's fallback.
func ClaimHash(chunkID, body string) string {
	sample := body
	if len(sample) > 256 {
		sample = sample[:256]
	}
	sum := sha256.Sum256([]byte(chunkID + "\n" + strings.TrimSpace(sample)))
	return hex.EncodeToString(sum[:])
}
