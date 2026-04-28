package compliance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal/audit"
	"github.com/rhl/businessos-backend/internal/optimal/tenancy"
)

// Retention applies TTL-based archive / delete / redact rules to aging
// content. Operators register policies in `retention_policies`; Sweep walks
// every policy, finds matching signals older than ttl_days, and applies the
// action — except when the signal is under legal hold.
type Retention struct {
	DB    *sql.DB
	Holds HoldStore     // required — every action consults the hold table
	Audit *audit.Logger // optional — emits "retention.applied" per action
}

// NewRetention wires a Retention engine.
func NewRetention(db *sql.DB, holds HoldStore, auditLog *audit.Logger) *Retention {
	return &Retention{DB: db, Holds: holds, Audit: auditLog}
}

// Action enumerates the supported retention verbs.
type Action string

const (
	ActionArchive Action = "archive" // mark archived; row stays for audit
	ActionDelete  Action = "delete"  // hard delete the row
	ActionRedact  Action = "redact"  // rewrite content through PII redactor
)

// ScopeType narrows what content a policy applies to.
type ScopeType string

const (
	ScopeNode   ScopeType = "node"   // by Pages parent/node id
	ScopeGenre  ScopeType = "genre"  // by genre tag (transcript, email, …)
	ScopeTenant ScopeType = "tenant" // tenant-wide
	ScopeGlobal ScopeType = "global" // every tenant
)

// Policy is one row of retention_policies.
type Policy struct {
	ID         int64
	TenantID   string
	ScopeType  ScopeType
	ScopeValue string // empty for tenant/global
	TTLDays    int
	Action     Action
	CreatedAt  time.Time
}

// SweepOptions tune a single sweep.
type SweepOptions struct {
	TenantID string
	Now      time.Time // injected for deterministic tests; defaults to time.Now().UTC()
	DryRun   bool
}

// SweepReport summarizes a completed sweep.
type SweepReport struct {
	PoliciesEvaluated int
	ActionsTaken      int
	SkippedHeld       int
	Errors            int
	ByAction          map[Action]int
}

// Sweep runs every policy for tenantID, applying actions to expired signals.
// Held signals are skipped (counted in SkippedHeld). DryRun computes what
// would happen without mutating anything.
func (r *Retention) Sweep(ctx context.Context, opts SweepOptions) (*SweepReport, error) {
	tenantID := tenancy.Resolve(opts.TenantID)
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	policies, err := r.listPolicies(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	report := &SweepReport{
		PoliciesEvaluated: len(policies),
		ByAction:          map[Action]int{},
	}

	for _, p := range policies {
		targets, err := r.findTargets(ctx, p, now)
		if err != nil {
			report.Errors++
			if r.Audit != nil {
				r.Audit.LogKind(ctx, "retention.policy_failed",
					audit.WithTenant(tenantID),
					audit.WithMetadata(map[string]any{"policy_id": p.ID, "err": err.Error()}),
				)
			}
			continue
		}
		for _, signalID := range targets {
			held, err := r.Holds.IsHeld(ctx, tenantID, signalID)
			if err != nil {
				// Hold check failure must NOT silently apply the action.
				report.Errors++
				continue
			}
			if held {
				report.SkippedHeld++
				continue
			}
			if opts.DryRun {
				report.ActionsTaken++
				report.ByAction[p.Action]++
				continue
			}
			if err := r.applyAction(ctx, p.Action, signalID); err != nil {
				report.Errors++
				continue
			}
			report.ActionsTaken++
			report.ByAction[p.Action]++
			if r.Audit != nil {
				r.Audit.LogKind(ctx, audit.KindRetentionApplied,
					audit.WithTenant(tenantID),
					audit.WithTargetURI("signal:"+signalID),
					audit.WithMetadata(map[string]any{
						"action":    string(p.Action),
						"policy_id": p.ID,
						"ttl_days":  p.TTLDays,
					}),
				)
			}
		}
	}
	return report, nil
}

// AddPolicy registers a new retention policy. Returns the new policy id.
func (r *Retention) AddPolicy(ctx context.Context, p Policy) (int64, error) {
	if p.TTLDays <= 0 {
		return 0, fmt.Errorf("compliance/retention: ttl_days must be positive")
	}
	if p.Action == "" {
		p.Action = ActionArchive
	}
	switch p.Action {
	case ActionArchive, ActionDelete, ActionRedact:
	default:
		return 0, fmt.Errorf("compliance/retention: invalid action %q", p.Action)
	}
	tenantID := tenancy.Resolve(p.TenantID)
	const ins = `
		INSERT INTO retention_policies (tenant_id, scope_type, scope_value, ttl_days, action)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int64
	if err := r.DB.QueryRowContext(ctx, ins, tenantID, string(p.ScopeType), nullableScope(p.ScopeValue), p.TTLDays, string(p.Action)).Scan(&id); err != nil {
		return 0, fmt.Errorf("compliance/retention: add policy: %w", err)
	}
	return id, nil
}

// RemovePolicy deletes a policy by id.
func (r *Retention) RemovePolicy(ctx context.Context, id int64) error {
	const del = `DELETE FROM retention_policies WHERE id = $1`
	res, err := r.DB.ExecContext(ctx, del, id)
	if err != nil {
		return fmt.Errorf("compliance/retention: remove policy: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("compliance/retention: policy not found")
	}
	return nil
}

func (r *Retention) listPolicies(ctx context.Context, tenantID string) ([]Policy, error) {
	const q = `
		SELECT id, tenant_id, scope_type, COALESCE(scope_value, ''), ttl_days, action, created_at
		FROM retention_policies
		WHERE tenant_id = $1 OR scope_type = 'global'
		ORDER BY id ASC
	`
	rows, err := r.DB.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("compliance/retention: list policies: %w", err)
	}
	defer rows.Close()
	out := []Policy{}
	for rows.Next() {
		var p Policy
		var st, act string
		if err := rows.Scan(&p.ID, &p.TenantID, &st, &p.ScopeValue, &p.TTLDays, &act, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.ScopeType = ScopeType(st)
		p.Action = Action(act)
		out = append(out, p)
	}
	return out, rows.Err()
}

// findTargets returns the contexts.id values that match policy.Scope and are
// older than policy.TTLDays. The match clauses are deliberately small and
// pure-SQL so the planner can use the user_id and created_at indexes.
func (r *Retention) findTargets(ctx context.Context, p Policy, now time.Time) ([]string, error) {
	cutoff := now.AddDate(0, 0, -p.TTLDays)

	var query string
	args := []any{cutoff}

	switch p.ScopeType {
	case ScopeNode:
		query = `SELECT id::text FROM contexts WHERE parent_id::text = $2 AND created_at < $1 AND COALESCE(is_archived, FALSE) = FALSE`
		args = append(args, p.ScopeValue)
	case ScopeGenre:
		// Genre is stored in contexts.properties->>'genre' for OptimalEngine-ingested rows.
		query = `SELECT id::text FROM contexts WHERE COALESCE(properties->>'genre', '') = $2 AND created_at < $1 AND COALESCE(is_archived, FALSE) = FALSE`
		args = append(args, p.ScopeValue)
	case ScopeTenant:
		query = `SELECT c.id::text FROM contexts c JOIN workspaces w ON w.owner_id = c.user_id WHERE w.id::text = $2 AND c.created_at < $1 AND COALESCE(c.is_archived, FALSE) = FALSE`
		args = append(args, p.TenantID)
	case ScopeGlobal:
		query = `SELECT id::text FROM contexts WHERE created_at < $1 AND COALESCE(is_archived, FALSE) = FALSE`
	default:
		return nil, fmt.Errorf("compliance/retention: unknown scope %q", p.ScopeType)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// applyAction runs the actual mutation. Per-action SQL is kept tiny so an
// unexpected action never has a chance to misfire.
func (r *Retention) applyAction(ctx context.Context, action Action, signalID string) error {
	switch action {
	case ActionArchive:
		const upd = `UPDATE contexts SET is_archived = TRUE, updated_at = NOW() WHERE id::text = $1`
		_, err := r.DB.ExecContext(ctx, upd, signalID)
		return err
	case ActionDelete:
		const del = `DELETE FROM contexts WHERE id::text = $1`
		_, err := r.DB.ExecContext(ctx, del, signalID)
		return err
	case ActionRedact:
		// Read content, run through PII redactor, write back.
		const read = `SELECT COALESCE(content, '') FROM contexts WHERE id::text = $1`
		var content string
		if err := r.DB.QueryRowContext(ctx, read, signalID).Scan(&content); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		redacted := RedactString(content, RedactOpts{Strategy: StrategyPlaceholder})
		const write = `UPDATE contexts SET content = $2, updated_at = NOW() WHERE id::text = $1`
		_, err := r.DB.ExecContext(ctx, write, signalID, redacted)
		return err
	default:
		return fmt.Errorf("compliance/retention: unsupported action %q", action)
	}
}

func nullableScope(s string) any {
	if s == "" {
		return nil
	}
	return s
}
