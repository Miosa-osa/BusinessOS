package compliance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal/tenancy"
)

// Hold is one row of the legal_holds table.
type Hold struct {
	ID         int64
	TenantID   string
	SignalID   string
	HeldBy     string
	Reason     string
	PlacedAt   time.Time
	ReleasedAt *time.Time // nil = active
}

// HoldStore is the persistence contract for legal holds.
type HoldStore interface {
	Place(ctx context.Context, tenantID, signalID, heldBy, reason string) (int64, error)
	Release(ctx context.Context, holdID int64) error
	Active(ctx context.Context, tenantID string) ([]Hold, error)
	CountForPrincipal(ctx context.Context, principalID, tenantID string) (int, error)
	IsHeld(ctx context.Context, tenantID, signalID string) (bool, error)
}

// LegalHold is the SQL-backed HoldStore. Construct via NewLegalHold.
type LegalHold struct {
	DB *sql.DB
}

// NewLegalHold wires a HoldStore to *sql.DB.
func NewLegalHold(db *sql.DB) *LegalHold { return &LegalHold{DB: db} }

// ErrHoldNotFound is returned when Release targets a missing or already-released id.
var ErrHoldNotFound = errors.New("compliance: legal hold not found or already released")

// Place inserts a new hold. Returns the new hold id.
func (l *LegalHold) Place(ctx context.Context, tenantID, signalID, heldBy, reason string) (int64, error) {
	if signalID == "" || heldBy == "" || reason == "" {
		return 0, fmt.Errorf("compliance/legal_hold: signal_id, held_by, reason all required")
	}
	tenantID = tenancy.Resolve(tenantID)
	const ins = `
		INSERT INTO legal_holds (tenant_id, signal_id, held_by, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id int64
	if err := l.DB.QueryRowContext(ctx, ins, tenantID, signalID, heldBy, reason).Scan(&id); err != nil {
		return 0, fmt.Errorf("compliance/legal_hold: place: %w", err)
	}
	return id, nil
}

// Release marks a hold as released. Returns ErrHoldNotFound if the hold
// doesn't exist or was already released.
func (l *LegalHold) Release(ctx context.Context, holdID int64) error {
	const upd = `
		UPDATE legal_holds
		SET released_at = NOW()
		WHERE id = $1 AND released_at IS NULL
	`
	res, err := l.DB.ExecContext(ctx, upd, holdID)
	if err != nil {
		return fmt.Errorf("compliance/legal_hold: release: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrHoldNotFound
	}
	return nil
}

// Active returns every un-released hold for a tenant.
func (l *LegalHold) Active(ctx context.Context, tenantID string) ([]Hold, error) {
	tenantID = tenancy.Resolve(tenantID)
	const q = `
		SELECT id, tenant_id, signal_id, held_by, reason, placed_at, released_at
		FROM legal_holds
		WHERE tenant_id = $1 AND released_at IS NULL
		ORDER BY placed_at DESC
	`
	rows, err := l.DB.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("compliance/legal_hold: active: %w", err)
	}
	defer rows.Close()
	out := []Hold{}
	for rows.Next() {
		var h Hold
		var rel sql.NullTime
		if err := rows.Scan(&h.ID, &h.TenantID, &h.SignalID, &h.HeldBy, &h.Reason, &h.PlacedAt, &rel); err != nil {
			return nil, fmt.Errorf("compliance/legal_hold: scan: %w", err)
		}
		if rel.Valid {
			rt := rel.Time
			h.ReleasedAt = &rt
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// CountForPrincipal returns the number of active holds covering content
// authored by the principal. Erasure uses this as its "must I refuse?" gate.
//
// Critical: a query error returns (0, error), NOT (0, nil). Callers gating
// destructive actions on this count must distinguish "no holds" from
// "couldn't tell you".
func (l *LegalHold) CountForPrincipal(ctx context.Context, principalID, tenantID string) (int, error) {
	if principalID == "" {
		return 0, fmt.Errorf("compliance/legal_hold: principal_id required")
	}
	tenantID = tenancy.Resolve(tenantID)
	const q = `
		SELECT COUNT(*)
		FROM legal_holds h
		JOIN contexts c ON c.id::text = h.signal_id OR c.share_id = h.signal_id
		WHERE h.tenant_id = $1
		  AND h.released_at IS NULL
		  AND c.user_id = $2
	`
	var count int
	if err := l.DB.QueryRowContext(ctx, q, tenantID, principalID).Scan(&count); err != nil {
		return 0, fmt.Errorf("compliance/legal_hold: count: %w", err)
	}
	return count, nil
}

// IsHeld reports whether an individual signal currently has any active hold.
// Retention uses this per-signal during sweeps.
func (l *LegalHold) IsHeld(ctx context.Context, tenantID, signalID string) (bool, error) {
	tenantID = tenancy.Resolve(tenantID)
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM legal_holds
			WHERE tenant_id = $1 AND signal_id = $2 AND released_at IS NULL
		)
	`
	var held bool
	if err := l.DB.QueryRowContext(ctx, q, tenantID, signalID).Scan(&held); err != nil {
		return false, fmt.Errorf("compliance/legal_hold: is_held: %w", err)
	}
	return held, nil
}
