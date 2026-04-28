package wiki

import (
	"context"
	"errors"
)

// Store is the persistence contract every wiki backend implements. The
// Elixir source has one backend (a SQLite table); BusinessOS has two:
//
//   - PagesAdapter (this repo) — bridges to the existing `contexts` table
//     so the Pages module is the source of truth.
//   - In-memory store (testing only).
//
// Implementations MUST upsert by (TenantID, Slug, Audience, Version) and
// return ErrNotFound — not nil — when a page is missing.
type Store interface {
	Put(ctx context.Context, page *Page) error
	Latest(ctx context.Context, tenantID, slug, audience string) (*Page, error)
	List(ctx context.Context, tenantID string) ([]*Page, error)
	PutCitations(ctx context.Context, tenantID, slug, audience string, citations []Citation) error
	ListCitations(ctx context.Context, tenantID, slug, audience string) ([]Citation, error)
}

// Citation is the metadata layered on top of a wiki page — one row per
// (tenantID, slug, audience, ChunkID, ClaimHash). Citations are what the
// integrity checker walks to verify wiki claims still resolve.
type Citation struct {
	ChunkID      string // pointer to a Tier-2 chunk (URI form `optimal://...`)
	ClaimHash    string // sha256 hex of the claim sentence; stable across edits
	LastVerified string // RFC3339; updated whenever the chunk is re-validated
}

// ErrNotFound is returned by Latest when no version of (tenant, slug, audience)
// exists. Wrap-aware: callers do `errors.Is(err, wiki.ErrNotFound)`.
var ErrNotFound = errors.New("wiki: page not found")
