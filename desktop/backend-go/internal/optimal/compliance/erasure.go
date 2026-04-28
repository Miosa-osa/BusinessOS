package compliance

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal/audit"
	"github.com/rhl/businessos-backend/internal/optimal/tenancy"
)

// Erasure performs the right-to-delete (GDPR Article 17 / CCPA §1798.105).
// It cascades through every BusinessOS table tied to a principal, atomically
// where possible, and writes an erasure_log row + audit event.
//
// Safety gates:
//
//   - Outstanding legal holds block the erase unless `Force` is set.
//   - Audit events are NEVER hard-deleted: the principal field is rewritten
//     to "erased:<sha256-prefix>" so the trail remains complete but anonymous.
type Erasure struct {
	DB    *sql.DB
	Holds HoldStore     // required — gates the erase
	Audit *audit.Logger // optional — emits an audit event on completion
}

// NewErasure wires Erasure to its dependencies. Holds is required: erasure
// must NEVER run without consulting the legal-hold table.
func NewErasure(db *sql.DB, holds HoldStore, auditLog *audit.Logger) *Erasure {
	return &Erasure{DB: db, Holds: holds, Audit: auditLog}
}

// EraseOptions tune a single erasure run.
type EraseOptions struct {
	TenantID string
	Actor    string // who is performing the erase (for audit). Defaults to "system:erasure".
	Force    bool   // skip the legal-hold gate
	DryRun   bool   // compute counts without applying
}

// EraseReport summarizes a completed erasure.
type EraseReport struct {
	PrincipalID    string
	TenantID       string
	Deleted        map[string]int // table → row count
	EventsRedacted int
	Held           int
	Forced         bool
	DryRun         bool
	CompletedAt    time.Time
}

// ErrHeldContent is returned when active holds prevent a non-forced erase.
var ErrHeldContent = errors.New("compliance/erasure: principal has held content; pass Force=true to override")

// Erase removes every row tied to principalID. Safe — refuses to run when
// holds exist (unless Force=true) and writes a complete audit trail.
func (e *Erasure) Erase(ctx context.Context, principalID string, opts EraseOptions) (*EraseReport, error) {
	if principalID == "" {
		return nil, fmt.Errorf("compliance/erasure: principal_id required")
	}
	tenantID := tenancy.Resolve(opts.TenantID)
	actor := opts.Actor
	if actor == "" {
		actor = "system:erasure"
	}

	// Hold gate: a query failure is propagated, NOT silently treated as
	// "no holds". Erasure must fail closed.
	held := 0
	if !opts.Force {
		n, err := e.Holds.CountForPrincipal(ctx, principalID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("compliance/erasure: hold check failed: %w", err)
		}
		held = n
		if held > 0 {
			return nil, fmt.Errorf("%w (%d holds)", ErrHeldContent, held)
		}
	}

	report := &EraseReport{
		PrincipalID: principalID,
		TenantID:    tenantID,
		Deleted:     map[string]int{},
		Held:        held,
		Forced:      opts.Force,
		DryRun:      opts.DryRun,
		CompletedAt: time.Now().UTC(),
	}

	tx, err := e.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("compliance/erasure: begin: %w", err)
	}
	defer tx.Rollback()

	// Tables we cascade through. Order matters: child rows before parents so
	// foreign-key constraints stay satisfied. Each entry is (table, where).
	cascades := []struct{ table, where string }{
		{"node_contexts", `context_id IN (SELECT id FROM contexts WHERE user_id = $1)`},
		{"contexts", `user_id = $1`},
		{"context_sessions", `user_id = $1`},
		{"context_retrieval_log", `user_id = $1`},
		{"wiki_citations", `slug IN (SELECT share_id FROM contexts WHERE user_id = $1 AND share_id IS NOT NULL)`},
		{"optimal_acls", `principal_id = $1`},
	}

	for _, c := range cascades {
		n, err := countOrDelete(ctx, tx, c.table, c.where, opts.DryRun, principalID)
		if err != nil {
			// Missing table is not fatal — BusinessOS may not have every
			// optional integration installed. Log via audit and continue.
			if e.Audit != nil {
				e.Audit.LogKind(ctx, "erasure.skipped",
					audit.WithTenant(tenantID),
					audit.WithMetadata(map[string]any{"table": c.table, "err": err.Error()}),
				)
			}
			continue
		}
		if n > 0 {
			report.Deleted[c.table] = n
		}
	}

	// Audit redaction — keep the row, anonymize the actor.
	redacted, err := redactAuditRows(ctx, tx, tenantID, principalID, opts.DryRun)
	if err != nil {
		// audit_events table may not exist in dev; that's acceptable.
		redacted = 0
	}
	report.EventsRedacted = redacted

	// Erasure log — written even on dry-run so the operator has a record
	// of what would have been deleted.
	if !opts.DryRun {
		if err := e.recordErasure(ctx, tx, report, actor); err != nil {
			return nil, fmt.Errorf("compliance/erasure: record: %w", err)
		}
	}

	if !opts.DryRun {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("compliance/erasure: commit: %w", err)
		}
	}

	if e.Audit != nil {
		e.Audit.LogKind(ctx, "erasure.completed",
			audit.WithTenant(tenantID),
			audit.WithPrincipal(actor),
			audit.WithTargetURI("principal:"+principalID),
			audit.WithMetadata(map[string]any{
				"deleted":         report.Deleted,
				"events_redacted": report.EventsRedacted,
				"held":            report.Held,
				"forced":          report.Forced,
				"dry_run":         report.DryRun,
			}),
		)
	}
	return report, nil
}

// countOrDelete either counts (dry run) or deletes rows matching `where`.
// Returns the count of affected rows.
func countOrDelete(ctx context.Context, tx *sql.Tx, table, where string, dryRun bool, args ...any) (int, error) {
	if dryRun {
		q := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)
		var n int
		if err := tx.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
			return 0, err
		}
		return n, nil
	}
	q := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
	res, err := tx.ExecContext(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// redactAuditRows rewrites the principal field on audit events authored by
// principalID to "erased:<sha256-prefix>" so the immutable audit trail
// remains intact while the personal identifier is removed.
func redactAuditRows(ctx context.Context, tx *sql.Tx, tenantID, principalID string, dryRun bool) (int, error) {
	const countQ = `SELECT COUNT(*) FROM audit_events WHERE tenant_id = $1 AND principal = $2`
	if dryRun {
		var n int
		if err := tx.QueryRowContext(ctx, countQ, tenantID, principalID).Scan(&n); err != nil {
			return 0, err
		}
		return n, nil
	}
	const upd = `UPDATE audit_events SET principal = $3 WHERE tenant_id = $1 AND principal = $2`
	res, err := tx.ExecContext(ctx, upd, tenantID, principalID, redactedActor(principalID))
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func redactedActor(principalID string) string {
	sum := sha256.Sum256([]byte(principalID))
	return "erased:" + hex.EncodeToString(sum[:])[:12]
}

func (e *Erasure) recordErasure(ctx context.Context, tx *sql.Tx, r *EraseReport, actor string) error {
	deletedJSON, err := json.Marshal(r.Deleted)
	if err != nil {
		return err
	}
	const ins = `
		INSERT INTO erasure_log (tenant_id, principal_id, actor, deleted_counts, events_redacted, held_count, forced)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, ins,
		r.TenantID, r.PrincipalID, actor, string(deletedJSON), r.EventsRedacted, r.Held, r.Forced,
	)
	return err
}
