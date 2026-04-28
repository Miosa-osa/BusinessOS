package compliance

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal/audit"
	"github.com/rhl/businessos-backend/internal/optimal/tenancy"
)

// DSAR exports every record the engine holds about a principal. GDPR
// Article 15 / CCPA §1798.110 grant individuals this right.
//
// Sections returned (each may be empty for a given principal):
//
//   - Principal      — the identity record itself
//   - Contexts       — signals authored by the principal
//   - Mentions       — signals whose content references the principal
//   - Memberships    — workspace memberships
//   - Events         — audit events where the principal acted
//
// In BusinessOS, "principal" is most often a better-auth user.id; the
// `contexts` table uses `user_id` for ownership and we treat that as the
// authorship dimension.
type DSAR struct {
	DB     *sql.DB
	Audit  *audit.Logger // optional; when set, every export emits an audit event
	Tenant tenancy.Store // optional; when set, validates the tenant before exporting
}

// NewDSAR wires a DSAR engine to its dependencies.
func NewDSAR(db *sql.DB, auditLog *audit.Logger, tenants tenancy.Store) *DSAR {
	return &DSAR{DB: db, Audit: auditLog, Tenant: tenants}
}

// Export is the report returned by DSAR.Export.
type Export struct {
	PrincipalID string           `json:"principal_id"`
	TenantID    string           `json:"tenant_id"`
	Principal   map[string]any   `json:"principal,omitempty"`
	Contexts    []map[string]any `json:"contexts"`
	Mentions    []map[string]any `json:"mentions"`
	Memberships []map[string]any `json:"memberships"`
	Events      []map[string]any `json:"events"`
	Counts      map[string]int   `json:"counts"`
	ExportedAt  time.Time        `json:"exported_at"`
}

// ErrPrincipalNotFound is returned when the principal can't be located.
var ErrPrincipalNotFound = errors.New("compliance: principal not found")

// Export produces a DSAR for principalID scoped to tenantID. Sections are
// loaded best-effort: a single section error is logged and the section
// returned empty, so an operator gets a partial-but-trustworthy report
// rather than a single failure on a deep table.
func (d *DSAR) Export(ctx context.Context, principalID, tenantID, actor string) (*Export, error) {
	if principalID == "" {
		return nil, fmt.Errorf("compliance/dsar: principal_id required")
	}
	tenantID = tenancy.Resolve(tenantID)
	if d.Tenant != nil {
		if _, err := d.Tenant.Get(ctx, tenantID); err != nil && !errors.Is(err, tenancy.ErrNotFound) {
			return nil, fmt.Errorf("compliance/dsar: validate tenant: %w", err)
		}
	}

	exp := &Export{
		PrincipalID: principalID,
		TenantID:    tenantID,
		Counts:      map[string]int{},
		ExportedAt:  time.Now().UTC(),
	}

	if p, err := d.fetchPrincipal(ctx, principalID); err == nil {
		exp.Principal = p
	} else if errors.Is(err, ErrPrincipalNotFound) {
		// Fall through with empty Principal — the principal may be a
		// non-human (agent/service) recorded only via foreign-key tables.
	} else {
		return nil, err
	}

	contexts, ctxErr := d.listContexts(ctx, principalID)
	exp.Contexts = d.bestEffort(ctx, "contexts", contexts, ctxErr)
	mentions, menErr := d.listMentions(ctx, principalID, principalDisplayName(exp.Principal))
	exp.Mentions = d.bestEffort(ctx, "mentions", mentions, menErr)
	members, mbrErr := d.listMemberships(ctx, principalID)
	exp.Memberships = d.bestEffort(ctx, "memberships", members, mbrErr)
	events, evErr := d.listEvents(ctx, principalID, tenantID)
	exp.Events = d.bestEffort(ctx, "events", events, evErr)

	exp.Counts["contexts"] = len(exp.Contexts)
	exp.Counts["mentions"] = len(exp.Mentions)
	exp.Counts["memberships"] = len(exp.Memberships)
	exp.Counts["events"] = len(exp.Events)

	if err := d.recordExport(ctx, exp, actor); err != nil {
		return exp, fmt.Errorf("compliance/dsar: record export: %w", err)
	}
	if d.Audit != nil {
		d.Audit.LogKind(ctx, "dsar.exported",
			audit.WithTenant(tenantID),
			audit.WithPrincipal(actor),
			audit.WithTargetURI("principal:"+principalID),
			audit.WithMetadata(map[string]any{"counts": exp.Counts}),
		)
	}
	return exp, nil
}

// fetchPrincipal reads the better-auth user row when the ID is bare or
// prefixed with "user:"; non-human IDs return ErrPrincipalNotFound which
// callers treat as "no profile to include".
func (d *DSAR) fetchPrincipal(ctx context.Context, id string) (map[string]any, error) {
	bare := id
	if len(id) > 5 && id[:5] == "user:" {
		bare = id[5:]
	}
	const q = `SELECT id, name, email, "createdAt" FROM "user" WHERE id = $1 LIMIT 1`
	var uid, name, email string
	var created time.Time
	if err := d.DB.QueryRowContext(ctx, q, bare).Scan(&uid, &name, &email, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPrincipalNotFound
		}
		return nil, fmt.Errorf("compliance/dsar: fetch principal: %w", err)
	}
	return map[string]any{
		"id":         uid,
		"name":       name,
		"email":      email,
		"created_at": created.UTC().Format(time.RFC3339),
	}, nil
}

func (d *DSAR) listContexts(ctx context.Context, principalID string) ([]map[string]any, error) {
	const q = `
		SELECT id, name, COALESCE(type::text, 'CUSTOM'), COALESCE(content, ''), created_at
		FROM contexts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1000
	`
	rows, err := d.DB.QueryContext(ctx, q, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, name, ctype, content string
		var created time.Time
		if err := rows.Scan(&id, &name, &ctype, &content, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":         id,
			"name":       name,
			"type":       ctype,
			"content":    truncate(content, 1024),
			"created_at": created.UTC().Format(time.RFC3339),
		})
	}
	return out, rows.Err()
}

// listMentions runs a LIKE scan against contexts.content. With BusinessOS at
// hundreds of thousands of rows, this is acceptable for a DSAR (cold path,
// run-on-demand, must be exhaustive). For larger deployments swap in a
// FTS index in Phase 6.
func (d *DSAR) listMentions(ctx context.Context, principalID, displayName string) ([]map[string]any, error) {
	if displayName == "" {
		return []map[string]any{}, nil
	}
	const q = `
		SELECT id, name, created_at
		FROM contexts
		WHERE user_id <> $1
		  AND content ILIKE '%' || $2 || '%'
		ORDER BY created_at DESC
		LIMIT 500
	`
	rows, err := d.DB.QueryContext(ctx, q, principalID, displayName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, name string
		var created time.Time
		if err := rows.Scan(&id, &name, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":         id,
			"name":       name,
			"created_at": created.UTC().Format(time.RFC3339),
		})
	}
	return out, rows.Err()
}

func (d *DSAR) listMemberships(ctx context.Context, principalID string) ([]map[string]any, error) {
	const q = `
		SELECT id, name, plan_type, created_at
		FROM workspaces
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT 200
	`
	rows, err := d.DB.QueryContext(ctx, q, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, name, plan string
		var created time.Time
		if err := rows.Scan(&id, &name, &plan, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":         id,
			"name":       name,
			"plan":       plan,
			"created_at": created.UTC().Format(time.RFC3339),
		})
	}
	return out, rows.Err()
}

func (d *DSAR) listEvents(ctx context.Context, principalID, tenantID string) ([]map[string]any, error) {
	const q = `
		SELECT ts, kind, COALESCE(target_uri, ''), metadata
		FROM audit_events
		WHERE tenant_id = $1 AND principal = $2
		ORDER BY ts DESC
		LIMIT 1000
	`
	rows, err := d.DB.QueryContext(ctx, q, tenantID, principalID)
	if err != nil {
		// audit_events may not exist when the optional logger isn't bootstrapped.
		return []map[string]any{}, nil
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var ts time.Time
		var kind, target, metaJSON string
		if err := rows.Scan(&ts, &kind, &target, &metaJSON); err != nil {
			return nil, err
		}
		meta := map[string]any{}
		_ = json.Unmarshal([]byte(metaJSON), &meta)
		out = append(out, map[string]any{
			"ts":         ts.UTC().Format(time.RFC3339),
			"kind":       kind,
			"target_uri": target,
			"metadata":   meta,
		})
	}
	return out, rows.Err()
}

func (d *DSAR) recordExport(ctx context.Context, exp *Export, actor string) error {
	if actor == "" {
		actor = "system"
	}
	countsJSON, err := json.Marshal(exp.Counts)
	if err != nil {
		return err
	}
	const ins = `
		INSERT INTO dsar_exports (tenant_id, principal_id, actor, section_counts)
		VALUES ($1, $2, $3, $4::jsonb)
	`
	_, err = d.DB.ExecContext(ctx, ins, exp.TenantID, exp.PrincipalID, actor, string(countsJSON))
	return err
}

// bestEffort logs the failing section but returns an empty slice so the
// caller still gets a partial report. The audit event still records the
// final section_counts so the operator can spot which sections were short.
func (d *DSAR) bestEffort(ctx context.Context, section string, list []map[string]any, err error) []map[string]any {
	if err != nil {
		if d.Audit != nil {
			d.Audit.LogKind(ctx, "dsar.section_failed",
				audit.WithMetadata(map[string]any{"section": section, "err": err.Error()}),
			)
		}
		return []map[string]any{}
	}
	if list == nil {
		return []map[string]any{}
	}
	return list
}

func principalDisplayName(p map[string]any) string {
	if p == nil {
		return ""
	}
	if name, ok := p["name"].(string); ok {
		return name
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
