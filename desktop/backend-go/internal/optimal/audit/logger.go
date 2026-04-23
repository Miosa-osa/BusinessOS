package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Logger writes events to the `audit_events` table. Stateless and safe for
// concurrent use; constructed once at app start with a *sql.DB handle.
type Logger struct {
	db *sql.DB
}

// New returns a Logger bound to the given database.
func NewLogger(db *sql.DB) *Logger {
	return &Logger{db: db}
}

// EnsureSchema creates the `audit_events` table if missing. Idempotent.
// Kept in-package so the audit log can be bootstrapped independently of
// the main migration pipeline — audit is load-bearing for compliance and
// should never be skipped.
func (l *Logger) EnsureSchema(ctx context.Context) error {
	const ddl = `
	CREATE TABLE IF NOT EXISTS audit_events (
		id          BIGSERIAL PRIMARY KEY,
		tenant_id   TEXT        NOT NULL DEFAULT 'default',
		ts          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		principal   TEXT        NOT NULL DEFAULT 'system',
		kind        TEXT        NOT NULL,
		target_uri  TEXT,
		latency_ms  BIGINT,
		metadata    JSONB       NOT NULL DEFAULT '{}'::jsonb
	);
	CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_ts ON audit_events(tenant_id, ts DESC);
	CREATE INDEX IF NOT EXISTS idx_audit_events_kind      ON audit_events(kind);
	CREATE INDEX IF NOT EXISTS idx_audit_events_principal ON audit_events(principal);
	`
	if _, err := l.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("audit: ensure schema: %w", err)
	}
	return nil
}

// Log writes an event. Audit logging is best-effort: errors are logged but
// never returned, so an audit failure cannot crash business logic.
func (l *Logger) Log(ctx context.Context, e Event) {
	if e.TS.IsZero() {
		e.TS = time.Now().UTC()
	}
	if e.TenantID == "" {
		e.TenantID = DefaultTenantID
	}
	if e.Principal == "" {
		e.Principal = "system"
	}
	metaJSON, err := json.Marshal(e.Metadata)
	if err != nil {
		slog.Warn("audit: marshal metadata", "kind", e.Kind, "err", err)
		metaJSON = []byte("{}")
	}

	var latency any
	if e.LatencyMS > 0 {
		latency = e.LatencyMS
	}

	const insert = `
		INSERT INTO audit_events (tenant_id, ts, principal, kind, target_uri, latency_ms, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = l.db.ExecContext(ctx, insert,
		e.TenantID, e.TS, e.Principal, e.Kind, nullable(e.TargetURI), latency, string(metaJSON),
	)
	if err != nil {
		slog.Warn("audit: write failed", "kind", e.Kind, "err", err)
	}
}

// LogKind is the convenience equivalent of Elixir's log/2: build-and-write
// in one call.
func (l *Logger) LogKind(ctx context.Context, kind string, opts ...Option) {
	l.Log(ctx, New(kind, opts...))
}

// QueryOptions filter Query results. Zero values are ignored.
type QueryOptions struct {
	TenantID  string
	Principal string
	Kind      string
	Since     time.Time
	Until     time.Time
	Limit     int
}

// Query returns recent events matching the filter, newest first.
func (l *Logger) Query(ctx context.Context, opts QueryOptions) ([]Event, error) {
	if opts.TenantID == "" {
		opts.TenantID = DefaultTenantID
	}
	if opts.Limit <= 0 {
		opts.Limit = 100
	}

	// Build predicates incrementally so placeholders match args.
	preds := []string{"tenant_id = $1"}
	args := []any{opts.TenantID}
	next := func(clause string, val any) {
		args = append(args, val)
		preds = append(preds, strings.Replace(clause, "?", fmt.Sprintf("$%d", len(args)), 1))
	}
	if opts.Principal != "" {
		next("principal = ?", opts.Principal)
	}
	if opts.Kind != "" {
		next("kind = ?", opts.Kind)
	}
	if !opts.Since.IsZero() {
		next("ts >= ?", opts.Since)
	}
	if !opts.Until.IsZero() {
		next("ts < ?", opts.Until)
	}
	args = append(args, opts.Limit)
	sql := fmt.Sprintf(`
		SELECT tenant_id, ts, principal, kind, COALESCE(target_uri, ''), COALESCE(latency_ms, 0), metadata
		FROM audit_events
		WHERE %s
		ORDER BY ts DESC
		LIMIT $%d
	`, strings.Join(preds, " AND "), len(args))

	rows, err := l.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("audit: query: %w", err)
	}
	defer rows.Close()

	out := []Event{}
	for rows.Next() {
		var e Event
		var metaJSON string
		if err := rows.Scan(&e.TenantID, &e.TS, &e.Principal, &e.Kind, &e.TargetURI, &e.LatencyMS, &metaJSON); err != nil {
			return nil, fmt.Errorf("audit: scan: %w", err)
		}
		if metaJSON != "" {
			_ = json.Unmarshal([]byte(metaJSON), &e.Metadata)
		}
		if e.Metadata == nil {
			e.Metadata = map[string]any{}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
