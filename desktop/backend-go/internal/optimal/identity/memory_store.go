package identity

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal/tenancy"
)

// MemoryPrincipalStore is for tests. Production deployments use
// PrincipalAdapter.
type MemoryPrincipalStore struct {
	mu         sync.Mutex
	principals map[string]*Principal // keyed by ID
}

func NewMemoryPrincipalStore() *MemoryPrincipalStore {
	return &MemoryPrincipalStore{principals: map[string]*Principal{}}
}

func (m *MemoryPrincipalStore) Upsert(_ context.Context, p *Principal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	clone := *p
	if clone.TenantID == "" {
		clone.TenantID = tenancy.DefaultID
	}
	m.principals[p.ID] = &clone
	return nil
}

func (m *MemoryPrincipalStore) Get(_ context.Context, id, tenantID string) (*Principal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.principals[id]
	if !ok {
		return nil, ErrPrincipalNotFound
	}
	if tenantID != "" && p.TenantID != tenantID {
		return nil, ErrPrincipalNotFound
	}
	clone := *p
	return &clone, nil
}

func (m *MemoryPrincipalStore) List(_ context.Context, tenantID string) ([]*Principal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []*Principal{}
	for _, p := range m.principals {
		if tenantID == "" || p.TenantID == tenantID {
			clone := *p
			out = append(out, &clone)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// MemoryACLStore is the in-process ACLStore for tests.
type MemoryACLStore struct {
	mu   sync.Mutex
	acls []*ACL
	next int64
}

func NewMemoryACLStore() *MemoryACLStore { return &MemoryACLStore{} }

func (m *MemoryACLStore) Grant(_ context.Context, acl *ACL) error {
	if err := validateACL(acl); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.acls {
		if e.TenantID == acl.TenantID && e.ResourceURI == acl.ResourceURI &&
			e.PrincipalID == acl.PrincipalID && e.GroupID == acl.GroupID &&
			e.Permission == acl.Permission {
			return nil
		}
	}
	m.next++
	clone := *acl
	clone.ID = m.next
	if clone.GrantedAt.IsZero() {
		clone.GrantedAt = time.Now().UTC()
	}
	if clone.TenantID == "" {
		clone.TenantID = tenancy.DefaultID
	}
	m.acls = append(m.acls, &clone)
	return nil
}

func (m *MemoryACLStore) Revoke(_ context.Context, tenantID, resourceURI, principalID, groupID string, perm Permission) error {
	if (principalID == "") == (groupID == "") {
		return ErrACLConflict
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	tenantID = tenancy.Resolve(tenantID)
	out := m.acls[:0]
	for _, e := range m.acls {
		match := e.TenantID == tenantID &&
			e.ResourceURI == resourceURI &&
			e.Permission == perm &&
			e.PrincipalID == principalID &&
			e.GroupID == groupID
		if !match {
			out = append(out, e)
		}
	}
	m.acls = out
	return nil
}

func (m *MemoryACLStore) ListForResource(_ context.Context, tenantID, resourceURI string) ([]*ACL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	tenantID = tenancy.Resolve(tenantID)
	out := []*ACL{}
	for _, e := range m.acls {
		if e.TenantID == tenantID && e.ResourceURI == resourceURI {
			clone := *e
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (m *MemoryACLStore) Can(ctx context.Context, principalID, resourceURI string, want Permission, opts CanOptions) (bool, error) {
	acls, err := m.ListForResource(ctx, opts.TenantID, resourceURI)
	if err != nil {
		return false, err
	}
	if len(acls) == 0 {
		return true, nil
	}
	groupSet := map[string]bool{}
	for _, g := range opts.Groups {
		groupSet[g] = true
	}
	for _, acl := range acls {
		if !permissionSatisfies(acl.Permission, want) {
			continue
		}
		if acl.PrincipalID != "" && acl.PrincipalID == principalID {
			return true, nil
		}
		if acl.GroupID != "" && groupSet[acl.GroupID] {
			return true, nil
		}
	}
	return false, nil
}
