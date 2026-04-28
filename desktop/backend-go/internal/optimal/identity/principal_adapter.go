package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rhl/businessos-backend/internal/optimal/tenancy"
)

// PrincipalAdapter persists agents + service accounts in `optimal_principals`
// and reads human users from the better-auth `user` table.
//
// Read path:
//   - id starts with "user:"   → query better-auth user table by id
//   - id starts with "agent:"  → query optimal_principals
//   - id starts with "service:"→ query optimal_principals
//   - id has no prefix         → assume KindUser
//
// Write path: only Upsert into optimal_principals — better-auth users are
// created by the auth flow, not by the engine.
type PrincipalAdapter struct {
	DB *sql.DB
}

// NewPrincipalAdapter wires the adapter to *sql.DB. Call EnsureSchema once.
func NewPrincipalAdapter(db *sql.DB) *PrincipalAdapter { return &PrincipalAdapter{DB: db} }

// EnsureSchema creates the optimal_principals table if missing.
func (a *PrincipalAdapter) EnsureSchema(ctx context.Context) error {
	const ddl = `
	CREATE TABLE IF NOT EXISTS optimal_principals (
		id           TEXT PRIMARY KEY,
		tenant_id    TEXT NOT NULL DEFAULT 'default',
		kind         TEXT NOT NULL,
		display_name TEXT NOT NULL,
		external_id  TEXT,
		metadata     JSONB NOT NULL DEFAULT '{}'::jsonb,
		created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_optimal_principals_tenant ON optimal_principals(tenant_id);
	`
	if _, err := a.DB.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("identity: ensure schema: %w", err)
	}
	return nil
}

// Upsert writes a non-human principal. Human users are managed by better-auth
// and rejected here so the two stores don't drift.
func (a *PrincipalAdapter) Upsert(ctx context.Context, p *Principal) error {
	if p == nil || p.ID == "" || p.DisplayName == "" {
		return fmt.Errorf("identity: id and display_name are required")
	}
	if p.Kind == KindUser {
		return fmt.Errorf("identity: human users are managed by better-auth, not optimal_principals")
	}
	tenantID := tenancy.Resolve(p.TenantID)
	metaJSON, err := json.Marshal(p.Metadata)
	if err != nil {
		return fmt.Errorf("identity: marshal metadata: %w", err)
	}
	const upsert = `
		INSERT INTO optimal_principals (id, tenant_id, kind, display_name, external_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			external_id  = EXCLUDED.external_id,
			metadata     = EXCLUDED.metadata
	`
	if _, err := a.DB.ExecContext(ctx, upsert, p.ID, tenantID, string(p.Kind), p.DisplayName, nullable(p.ExternalID), string(metaJSON)); err != nil {
		return fmt.Errorf("identity: upsert principal: %w", err)
	}
	return nil
}

// Get resolves a principal id to a Principal struct. Routing is by prefix
// so the caller doesn't have to know which backend stores what.
func (a *PrincipalAdapter) Get(ctx context.Context, id, tenantID string) (*Principal, error) {
	tenantID = tenancy.Resolve(tenantID)
	kind, raw := ParseID(id)
	if kind == KindUser {
		return a.getUser(ctx, raw, tenantID)
	}
	return a.getOptimal(ctx, id, tenantID)
}

// List returns every principal recorded in optimal_principals for the tenant
// plus a synthetic entry for the system account so audit views always have
// it. Human users are NOT enumerated here — the BusinessOS workspace API
// handles that and the data set is large.
func (a *PrincipalAdapter) List(ctx context.Context, tenantID string) ([]*Principal, error) {
	tenantID = tenancy.Resolve(tenantID)
	const q = `
		SELECT id, tenant_id, kind, display_name, COALESCE(external_id, ''), metadata
		FROM optimal_principals
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`
	rows, err := a.DB.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("identity: list: %w", err)
	}
	defer rows.Close()

	out := []*Principal{
		{ID: SystemPrincipalID, TenantID: tenantID, Kind: KindService, DisplayName: "System"},
	}
	for rows.Next() {
		p := &Principal{}
		var metaJSON string
		var kind string
		if err := rows.Scan(&p.ID, &p.TenantID, &kind, &p.DisplayName, &p.ExternalID, &metaJSON); err != nil {
			return nil, fmt.Errorf("identity: scan: %w", err)
		}
		p.Kind = Kind(kind)
		p.Metadata = decodeMeta(metaJSON)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (a *PrincipalAdapter) getUser(ctx context.Context, userID, tenantID string) (*Principal, error) {
	const q = `SELECT id, name, email FROM "user" WHERE id = $1 LIMIT 1`
	var id, name, email string
	if err := a.DB.QueryRowContext(ctx, q, userID).Scan(&id, &name, &email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPrincipalNotFound
		}
		return nil, fmt.Errorf("identity: get user: %w", err)
	}
	return &Principal{
		ID:          FormatID(KindUser, id),
		TenantID:    tenantID,
		Kind:        KindUser,
		DisplayName: name,
		ExternalID:  email,
	}, nil
}

func (a *PrincipalAdapter) getOptimal(ctx context.Context, id, tenantID string) (*Principal, error) {
	const q = `
		SELECT id, tenant_id, kind, display_name, COALESCE(external_id, ''), metadata
		FROM optimal_principals
		WHERE id = $1 AND tenant_id = $2
		LIMIT 1
	`
	var metaJSON, kind string
	p := &Principal{}
	err := a.DB.QueryRowContext(ctx, q, id, tenantID).Scan(&p.ID, &p.TenantID, &kind, &p.DisplayName, &p.ExternalID, &metaJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPrincipalNotFound
		}
		return nil, fmt.Errorf("identity: get optimal: %w", err)
	}
	p.Kind = Kind(kind)
	p.Metadata = decodeMeta(metaJSON)
	return p, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func decodeMeta(s string) map[string]any {
	out := map[string]any{}
	if s == "" {
		return out
	}
	_ = json.Unmarshal([]byte(s), &out)
	return out
}
