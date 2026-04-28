package compliance

import (
	"context"
	"database/sql"
	"fmt"
)

// EnsureSchema creates the four compliance tables idempotently. Call once
// at app startup alongside audit.Logger.EnsureSchema and tenancy/identity
// setup. Tables:
//
//   - legal_holds          — pause retention + block erasure
//   - retention_policies   — TTL-based archive/delete/redact rules
//   - dsar_exports         — historical export audit (lightweight)
//   - erasure_log          — record of who erased whom + counts
//
// We intentionally do NOT create the contexts/audit_events tables here —
// those are owned by their respective subsystems and migrations.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	const ddl = `
	CREATE TABLE IF NOT EXISTS legal_holds (
		id            BIGSERIAL PRIMARY KEY,
		tenant_id     TEXT        NOT NULL DEFAULT 'default',
		signal_id     TEXT        NOT NULL,
		held_by       TEXT        NOT NULL,
		reason        TEXT        NOT NULL,
		placed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		released_at   TIMESTAMPTZ,
		UNIQUE (tenant_id, signal_id, held_by, placed_at)
	);
	CREATE INDEX IF NOT EXISTS idx_legal_holds_active ON legal_holds(tenant_id, signal_id) WHERE released_at IS NULL;

	CREATE TABLE IF NOT EXISTS retention_policies (
		id            BIGSERIAL PRIMARY KEY,
		tenant_id     TEXT        NOT NULL DEFAULT 'default',
		scope_type    TEXT        NOT NULL,                                       -- 'node' | 'genre' | 'tenant' | 'global'
		scope_value   TEXT,                                                       -- e.g. 'ai-masters' or 'transcript'; NULL for global/tenant
		ttl_days      INTEGER     NOT NULL CHECK (ttl_days > 0),
		action        TEXT        NOT NULL DEFAULT 'archive',                     -- 'archive' | 'delete' | 'redact'
		created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_retention_policies_tenant ON retention_policies(tenant_id);

	CREATE TABLE IF NOT EXISTS dsar_exports (
		id            BIGSERIAL PRIMARY KEY,
		tenant_id     TEXT        NOT NULL DEFAULT 'default',
		principal_id  TEXT        NOT NULL,
		actor         TEXT        NOT NULL DEFAULT 'system',
		section_counts JSONB      NOT NULL DEFAULT '{}'::jsonb,
		exported_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_dsar_exports_principal ON dsar_exports(tenant_id, principal_id);

	CREATE TABLE IF NOT EXISTS erasure_log (
		id            BIGSERIAL PRIMARY KEY,
		tenant_id     TEXT        NOT NULL DEFAULT 'default',
		principal_id  TEXT        NOT NULL,
		actor         TEXT        NOT NULL DEFAULT 'system',
		deleted_counts JSONB      NOT NULL DEFAULT '{}'::jsonb,
		events_redacted INTEGER   NOT NULL DEFAULT 0,
		held_count    INTEGER     NOT NULL DEFAULT 0,
		forced        BOOLEAN     NOT NULL DEFAULT FALSE,
		completed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_erasure_log_principal ON erasure_log(tenant_id, principal_id);
	`
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("compliance: ensure schema: %w", err)
	}
	return nil
}
