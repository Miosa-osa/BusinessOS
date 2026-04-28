package tenancy

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MemoryStore is the in-process implementation used by tests and dev-mode
// bring-up. Production deployments use WorkspaceAdapter.
type MemoryStore struct {
	mu      sync.Mutex
	tenants map[string]*Tenant
}

// NewMemoryStore returns a ready-to-use store.
func NewMemoryStore() *MemoryStore { return &MemoryStore{tenants: map[string]*Tenant{}} }

func (m *MemoryStore) EnsureDefault(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tenants[DefaultID]; ok {
		return nil
	}
	m.tenants[DefaultID] = &Tenant{
		ID:        DefaultID,
		Name:      "Default",
		Plan:      "system",
		OwnerID:   "system",
		CreatedAt: time.Now().UTC(),
		Metadata:  map[string]any{"system": true},
	}
	return nil
}

func (m *MemoryStore) Get(_ context.Context, id string) (*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tenants[Resolve(id)]
	if !ok {
		return nil, ErrNotFound
	}
	clone := *t
	return &clone, nil
}

func (m *MemoryStore) Create(_ context.Context, t *Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	clone := *t
	m.tenants[t.ID] = &clone
	return nil
}

func (m *MemoryStore) List(_ context.Context) ([]*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*Tenant, 0, len(m.tenants))
	for _, t := range m.tenants {
		clone := *t
		out = append(out, &clone)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
