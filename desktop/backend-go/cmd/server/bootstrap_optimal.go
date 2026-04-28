package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal/audit"
	"github.com/rhl/businessos-backend/internal/optimal/compliance"
	"github.com/rhl/businessos-backend/internal/optimal/identity"
	"github.com/rhl/businessos-backend/internal/optimal/tenancy"
	"github.com/rhl/businessos-backend/internal/optimal/wiki"
)

// bootstrapOptimalSchemas creates every Phase 3-8 OptimalEngine table that
// doesn't yet exist and seeds the reserved "default" tenant. Idempotent —
// safe to call on every startup; existing tables are left untouched.
//
// Ordering: tenancy.EnsureDefault first so audit/compliance rows can reference
// the default tenant_id; identity adapters next so any Phase 1 ACL grant can
// land; audit + compliance + wiki adapters last (they depend on neither).
//
// Failures are logged via slog but never propagated — a partial bootstrap
// degrades feature-by-feature instead of crashing the whole server.
func bootstrapOptimalSchemas(ctx context.Context, db *sql.DB) {
	if db == nil {
		slog.Warn("optimal schemas: skipped (no database)")
		return
	}

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	steps := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"tenancy default", tenancy.NewWorkspaceAdapter(db).EnsureDefault},
		{"audit events", audit.NewLogger(db).EnsureSchema},
		{"compliance tables", func(ctx context.Context) error { return compliance.EnsureSchema(ctx, db) }},
		{"wiki citations", wiki.NewPagesAdapter(db).EnsureSchema},
		{"identity principals", identity.NewPrincipalAdapter(db).EnsureSchema},
		{"identity acls", identity.NewACLAdapter(db).EnsureSchema},
	}

	failed := 0
	for _, s := range steps {
		if err := s.fn(initCtx); err != nil {
			slog.Warn("optimal schemas: step failed", "step", s.name, "error", err)
			failed++
			continue
		}
		slog.Debug("optimal schemas: step ready", "step", s.name)
	}

	if failed == 0 {
		slog.Info("optimal schemas ready",
			"steps", len(steps),
			"subsystems", "audit, compliance, wiki, tenancy, identity",
		)
		return
	}
	slog.Warn(fmt.Sprintf("optimal schemas: %d/%d steps failed (continuing)", failed, len(steps)))
}
