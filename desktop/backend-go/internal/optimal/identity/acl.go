package identity

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Permission gates what kind of access a grant confers. Strings match
// the Elixir source's atom values.
type Permission string

const (
	PermRead  Permission = "read"
	PermWrite Permission = "write"
	PermAdmin Permission = "admin"
)

// ACL is one access-control row. Exactly one of PrincipalID / GroupID is set.
type ACL struct {
	ID          int64
	TenantID    string
	ResourceURI string // "optimal://...", "wiki:slug:audience", "tenant:id"
	PrincipalID string
	GroupID     string
	Permission  Permission
	GrantedAt   time.Time
}

// Resource URI prefixes — kept here as constants so callers don't typo them.
const (
	ResourcePrefixWiki   = "wiki:"   // wiki:<slug>:<audience>
	ResourcePrefixChunk  = "chunk:"  // chunk:<id>
	ResourcePrefixTenant = "tenant:" // tenant:<id>
)

// PermissionRank maps a permission to its implication weight: admin implies
// write implies read. Used by Can to short-circuit upgrade checks.
var PermissionRank = map[Permission]int{
	PermRead:  1,
	PermWrite: 2,
	PermAdmin: 3,
}

// ACLStore is the persistence contract.
type ACLStore interface {
	Grant(ctx context.Context, acl *ACL) error
	Revoke(ctx context.Context, tenantID, resourceURI, principalID, groupID string, perm Permission) error
	ListForResource(ctx context.Context, tenantID, resourceURI string) ([]*ACL, error)
	Can(ctx context.Context, principalID, resourceURI string, perm Permission, opts CanOptions) (bool, error)
}

// CanOptions carry caller-provided context so Can can resolve group
// membership without the engine having to embed a group store.
type CanOptions struct {
	TenantID string
	Groups   []string // groups the principal belongs to (caller resolves)
}

// ErrACLConflict is returned when a Grant violates the principal-XOR-group
// invariant or other shape requirements.
var ErrACLConflict = errors.New("identity: ACL must set exactly one of principal_id / group_id")

// validateACL enforces the principal-XOR-group invariant.
func validateACL(a *ACL) error {
	if a == nil {
		return ErrACLConflict
	}
	if a.ResourceURI == "" {
		return fmt.Errorf("identity: resource_uri is required")
	}
	if (a.PrincipalID == "") == (a.GroupID == "") {
		return ErrACLConflict
	}
	if PermissionRank[a.Permission] == 0 {
		return fmt.Errorf("identity: invalid permission %q", a.Permission)
	}
	return nil
}

// permissionSatisfies reports whether `held` covers a request for `wanted`.
// admin satisfies write satisfies read.
func permissionSatisfies(held, wanted Permission) bool {
	return PermissionRank[held] >= PermissionRank[wanted]
}
