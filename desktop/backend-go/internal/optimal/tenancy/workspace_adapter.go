package tenancy

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// WorkspaceAdapter implements Store on top of the existing BusinessOS
// `workspaces` table. We do NOT add a parallel `tenants` table — every
// workspace IS the tenant. The reserved "default" tenant is materialized
// on demand by EnsureDefault so audit and wiki rows always have a valid
// tenant_id even before the user creates a workspace.
type WorkspaceAdapter struct {
	DB *sql.DB
}

// NewWorkspaceAdapter wires the adapter to the *sql.DB used elsewhere by
// BusinessOS. Call EnsureDefault once at startup.
func NewWorkspaceAdapter(db *sql.DB) *WorkspaceAdapter { return &WorkspaceAdapter{DB: db} }

// EnsureDefault inserts the reserved "default" tenant row if missing. The
// row is keyed by slug (the only UNIQUE column on workspaces) so this is
// idempotent across restarts.
func (a *WorkspaceAdapter) EnsureDefault(ctx context.Context) error {
	const upsert = `
		INSERT INTO workspaces (id, name, slug, plan_type, owner_id, settings)
		VALUES ('00000000-0000-0000-0000-000000000000', 'Default', $1, 'system', 'system', '{"system":true}'::jsonb)
		ON CONFLICT (slug) DO NOTHING
	`
	if _, err := a.DB.ExecContext(ctx, upsert, DefaultID); err != nil {
		return fmt.Errorf("tenancy: ensure default: %w", err)
	}
	return nil
}

// Get looks the tenant up by either workspaces.id (UUID) OR workspaces.slug.
// The reserved DefaultID always resolves to the system row.
func (a *WorkspaceAdapter) Get(ctx context.Context, id string) (*Tenant, error) {
	const q = `
		SELECT id, name, COALESCE(plan_type, 'free'), owner_id, COALESCE(settings, '{}'::jsonb), created_at
		FROM workspaces
		WHERE id::text = $1 OR slug = $1
		LIMIT 1
	`
	var settingsJSON string
	t := &Tenant{}
	err := a.DB.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.Name, &t.Plan, &t.OwnerID, &settingsJSON, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("tenancy: get: %w", err)
	}
	t.Metadata = decodeSettings(settingsJSON)
	if region, ok := t.Metadata["region"].(string); ok {
		t.Region = region
	}
	return t, nil
}

// Create inserts a new workspace row. ID + Name + OwnerID are required;
// slug defaults to ID when not in Metadata. Plan defaults to "free".
func (a *WorkspaceAdapter) Create(ctx context.Context, t *Tenant) error {
	if t == nil || t.ID == "" || t.Name == "" || t.OwnerID == "" {
		return fmt.Errorf("tenancy: id, name, and owner_id are required")
	}
	plan := t.Plan
	if plan == "" {
		plan = "free"
	}
	settings := cloneMeta(t.Metadata)
	if t.Region != "" {
		settings["region"] = t.Region
	}
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("tenancy: marshal settings: %w", err)
	}
	slug, _ := settings["slug"].(string)
	if slug == "" {
		slug = t.ID
	}

	const insert = `
		INSERT INTO workspaces (id, name, slug, plan_type, owner_id, settings)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
	`
	if _, err := a.DB.ExecContext(ctx, insert, t.ID, t.Name, slug, plan, t.OwnerID, string(settingsJSON)); err != nil {
		return fmt.Errorf("tenancy: create: %w", err)
	}
	return nil
}

// List returns every workspace as a Tenant, ordered by created_at.
func (a *WorkspaceAdapter) List(ctx context.Context) ([]*Tenant, error) {
	const q = `
		SELECT id, name, COALESCE(plan_type, 'free'), owner_id, COALESCE(settings, '{}'::jsonb), created_at
		FROM workspaces
		ORDER BY created_at ASC
	`
	rows, err := a.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("tenancy: list: %w", err)
	}
	defer rows.Close()

	out := []*Tenant{}
	for rows.Next() {
		var settingsJSON string
		t := &Tenant{}
		var created time.Time
		if err := rows.Scan(&t.ID, &t.Name, &t.Plan, &t.OwnerID, &settingsJSON, &created); err != nil {
			return nil, fmt.Errorf("tenancy: scan: %w", err)
		}
		t.CreatedAt = created
		t.Metadata = decodeSettings(settingsJSON)
		if region, ok := t.Metadata["region"].(string); ok {
			t.Region = region
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func decodeSettings(s string) map[string]any {
	out := map[string]any{}
	if s == "" {
		return out
	}
	_ = json.Unmarshal([]byte(s), &out)
	return out
}

func cloneMeta(m map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		out[k] = v
	}
	return out
}
