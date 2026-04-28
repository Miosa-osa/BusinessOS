package identity

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rhl/businessos-backend/internal/optimal/tenancy"
)

// ACLAdapter persists ACLs in `optimal_acls`. Phase 1 uses a permissive
// default — when no ACL exists for a resource, Can returns true. Phase 8
// will tighten this once wiki-first retrieval is fully wired.
type ACLAdapter struct {
	DB *sql.DB
}

// NewACLAdapter wires the adapter to *sql.DB. Call EnsureSchema once.
func NewACLAdapter(db *sql.DB) *ACLAdapter { return &ACLAdapter{DB: db} }

// EnsureSchema creates the optimal_acls table if missing.
func (a *ACLAdapter) EnsureSchema(ctx context.Context) error {
	const ddl = `
	CREATE TABLE IF NOT EXISTS optimal_acls (
		id            BIGSERIAL PRIMARY KEY,
		tenant_id     TEXT NOT NULL DEFAULT 'default',
		resource_uri  TEXT NOT NULL,
		principal_id  TEXT,
		group_id      TEXT,
		permission    TEXT NOT NULL,
		granted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CHECK ((principal_id IS NULL) <> (group_id IS NULL)),
		UNIQUE (tenant_id, resource_uri, principal_id, group_id, permission)
	);
	CREATE INDEX IF NOT EXISTS idx_optimal_acls_resource ON optimal_acls(tenant_id, resource_uri);
	CREATE INDEX IF NOT EXISTS idx_optimal_acls_principal ON optimal_acls(principal_id);
	`
	if _, err := a.DB.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("identity: ensure acl schema: %w", err)
	}
	return nil
}

// Grant inserts a permission row. The principal-XOR-group invariant is
// enforced both client-side (validateACL) and server-side (CHECK constraint).
func (a *ACLAdapter) Grant(ctx context.Context, acl *ACL) error {
	if err := validateACL(acl); err != nil {
		return err
	}
	tenantID := tenancy.Resolve(acl.TenantID)
	const insert = `
		INSERT INTO optimal_acls (tenant_id, resource_uri, principal_id, group_id, permission)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, resource_uri, principal_id, group_id, permission) DO NOTHING
	`
	if _, err := a.DB.ExecContext(ctx, insert, tenantID, acl.ResourceURI, nullable(acl.PrincipalID), nullable(acl.GroupID), string(acl.Permission)); err != nil {
		return fmt.Errorf("identity: grant: %w", err)
	}
	return nil
}

// Revoke removes a specific grant. Either principalID or groupID must be set.
func (a *ACLAdapter) Revoke(ctx context.Context, tenantID, resourceURI, principalID, groupID string, perm Permission) error {
	if (principalID == "") == (groupID == "") {
		return ErrACLConflict
	}
	tenantID = tenancy.Resolve(tenantID)
	const del = `
		DELETE FROM optimal_acls
		WHERE tenant_id = $1 AND resource_uri = $2 AND permission = $3
		  AND COALESCE(principal_id, '') = COALESCE($4, '')
		  AND COALESCE(group_id, '') = COALESCE($5, '')
	`
	if _, err := a.DB.ExecContext(ctx, del, tenantID, resourceURI, string(perm), nullable(principalID), nullable(groupID)); err != nil {
		return fmt.Errorf("identity: revoke: %w", err)
	}
	return nil
}

// ListForResource returns every ACL touching the resource URI.
func (a *ACLAdapter) ListForResource(ctx context.Context, tenantID, resourceURI string) ([]*ACL, error) {
	tenantID = tenancy.Resolve(tenantID)
	const q = `
		SELECT id, tenant_id, resource_uri, COALESCE(principal_id, ''), COALESCE(group_id, ''), permission, granted_at
		FROM optimal_acls
		WHERE tenant_id = $1 AND resource_uri = $2
		ORDER BY granted_at ASC
	`
	rows, err := a.DB.QueryContext(ctx, q, tenantID, resourceURI)
	if err != nil {
		return nil, fmt.Errorf("identity: list acls: %w", err)
	}
	defer rows.Close()

	out := []*ACL{}
	for rows.Next() {
		acl := &ACL{}
		var perm string
		if err := rows.Scan(&acl.ID, &acl.TenantID, &acl.ResourceURI, &acl.PrincipalID, &acl.GroupID, &perm, &acl.GrantedAt); err != nil {
			return nil, fmt.Errorf("identity: scan: %w", err)
		}
		acl.Permission = Permission(perm)
		out = append(out, acl)
	}
	return out, rows.Err()
}

// Can returns true when the principal (or any of their groups) holds the
// requested permission on the resource. With NO ACLs present for the
// resource, the Phase-1 permissive default returns true.
func (a *ACLAdapter) Can(ctx context.Context, principalID, resourceURI string, want Permission, opts CanOptions) (bool, error) {
	acls, err := a.ListForResource(ctx, opts.TenantID, resourceURI)
	if err != nil {
		return false, err
	}
	if len(acls) == 0 {
		return true, nil // permissive default — see ACLAdapter doc
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
