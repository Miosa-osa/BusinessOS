// Package audit ports OptimalEngine.Audit (Elixir) to Go.
//
// Every read and write in the engine emits an Event so SOC 2 / GDPR audit
// queries can answer questions like "who accessed X between T1 and T2".
// Events are append-only and persisted via Logger.Log.
package audit

import "time"

// Canonical event kinds. Using constants here so callers don't typo the
// string keys that SOC 2 / GDPR queries later filter on.
const (
	KindIngestStarted     = "ingest.started"
	KindIngestCompleted   = "ingest.completed"
	KindIngestFailed      = "ingest.failed"
	KindRetrievalExecuted = "retrieval.executed"
	KindWikiCurated       = "wiki.page.curated"
	KindWikiRead          = "wiki.page.read"
	KindPermissionDenied  = "permission.denied"
	KindPermissionGranted = "permission.granted"
	KindConnectorSynced   = "connector.synced"
	KindConnectorFailed   = "connector.failed"
	KindTenantCreated     = "tenant.created"
	KindPrincipalCreated  = "principal.created"
	KindPrincipalLogin    = "principal.logged_in"
	KindLegalHoldPlaced   = "legal_hold.placed"
	KindLegalHoldReleased = "legal_hold.released"
	KindRetentionApplied  = "retention.applied"
)

// DefaultTenantID is the fallback tenant used when no multi-tenant context
// exists yet. BusinessOS is single-tenant today — this keeps the audit table
// populated with consistent values so tenant filtering works by construction.
const DefaultTenantID = "default"

// Event is one row of the `events` table. Append-only; never updated.
type Event struct {
	TenantID  string
	TS        time.Time
	Principal string
	Kind      string
	TargetURI string
	LatencyMS int64
	Metadata  map[string]any
}

// Option configures a New call.
type Option func(*Event)

// WithTenant overrides the default tenant.
func WithTenant(id string) Option { return func(e *Event) { e.TenantID = id } }

// WithPrincipal sets the actor (user ID, service name, "system").
func WithPrincipal(p string) Option { return func(e *Event) { e.Principal = p } }

// WithTargetURI records what resource was touched.
func WithTargetURI(uri string) Option { return func(e *Event) { e.TargetURI = uri } }

// WithLatency records how long the operation took.
func WithLatency(ms int64) Option { return func(e *Event) { e.LatencyMS = ms } }

// WithMetadata merges the provided map into the event's metadata.
func WithMetadata(m map[string]any) Option {
	return func(e *Event) {
		if e.Metadata == nil {
			e.Metadata = map[string]any{}
		}
		for k, v := range m {
			e.Metadata[k] = v
		}
	}
}

// WithTS explicitly sets the event timestamp (otherwise time.Now().UTC()).
func WithTS(t time.Time) Option { return func(e *Event) { e.TS = t } }

// New builds an Event with sensible defaults for the fields not set by opts.
func New(kind string, opts ...Option) Event {
	e := Event{
		TenantID:  DefaultTenantID,
		TS:        time.Now().UTC(),
		Principal: "system",
		Kind:      kind,
		Metadata:  map[string]any{},
	}
	for _, o := range opts {
		o(&e)
	}
	return e
}
