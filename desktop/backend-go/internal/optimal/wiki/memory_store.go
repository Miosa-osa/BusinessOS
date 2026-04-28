package wiki

import (
	"context"
	"sort"
	"sync"
)

// MemoryStore is an in-process Store for tests and dev-mode bring-up. Real
// BusinessOS deployments use PagesAdapter against the contexts table.
type MemoryStore struct {
	mu        sync.Mutex
	pages     map[memKey][]*Page
	citations map[memKey][]Citation
}

type memKey struct{ tenant, slug, audience string }

// NewMemoryStore returns a ready-to-use in-process Store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		pages:     map[memKey][]*Page{},
		citations: map[memKey][]Citation{},
	}
}

// Put appends a new version of the page (version monotonic per key).
func (m *MemoryStore) Put(_ context.Context, p *Page) error {
	if p == nil || p.Slug == "" {
		return ErrMissingSlug
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	k := memKey{p.TenantID, p.Slug, audienceOr(p.Audience)}
	versions := m.pages[k]
	clone := *p
	if clone.Audience == "" {
		clone.Audience = DefaultAudience
	}
	versions = append(versions, &clone)
	m.pages[k] = versions
	return nil
}

func (m *MemoryStore) Latest(_ context.Context, tenantID, slug, audience string) (*Page, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := memKey{tenantID, slug, audienceOr(audience)}
	versions, ok := m.pages[k]
	if !ok || len(versions) == 0 {
		return nil, ErrNotFound
	}
	return versions[len(versions)-1], nil
}

func (m *MemoryStore) List(_ context.Context, tenantID string) ([]*Page, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []*Page{}
	for k, versions := range m.pages {
		if k.tenant != tenantID || len(versions) == 0 {
			continue
		}
		out = append(out, versions[len(versions)-1])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Slug != out[j].Slug {
			return out[i].Slug < out[j].Slug
		}
		return out[i].Audience < out[j].Audience
	})
	return out, nil
}

func (m *MemoryStore) PutCitations(_ context.Context, tenantID, slug, audience string, citations []Citation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := memKey{tenantID, slug, audienceOr(audience)}
	existing := m.citations[k]
	for _, c := range citations {
		seen := false
		for _, e := range existing {
			if e.ChunkID == c.ChunkID && e.ClaimHash == c.ClaimHash {
				seen = true
				break
			}
		}
		if !seen {
			existing = append(existing, c)
		}
	}
	m.citations[k] = existing
	return nil
}

func (m *MemoryStore) ListCitations(_ context.Context, tenantID, slug, audience string) ([]Citation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := memKey{tenantID, slug, audienceOr(audience)}
	out := append([]Citation{}, m.citations[k]...)
	return out, nil
}

func audienceOr(a string) string {
	if a == "" {
		return DefaultAudience
	}
	return a
}
