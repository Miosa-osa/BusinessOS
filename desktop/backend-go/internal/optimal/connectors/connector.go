// Package connectors ports OptimalEngine.Connectors (Elixir) to Go.
//
// A connector is a one-way pipe: "external system X" → "engine intake".
// It has no business logic of its own. The pipeline stages
// (parse / classify / embed / route / store) run downstream of the
// connector's Sync output; a connector's only job is to hand the
// engine a stream of Signals plus an opaque cursor that lets the next
// run pick up where this one stopped.
package connectors

import (
	"context"
	"time"
)

// AuthScheme enumerates the supported authentication modes for connectors.
type AuthScheme string

const (
	AuthOAuth2         AuthScheme = "oauth2"
	AuthToken          AuthScheme = "token"
	AuthBasic          AuthScheme = "basic"
	AuthWebhook        AuthScheme = "webhook"
	AuthServiceAccount AuthScheme = "service_account"
)

// Signal is the canonical output of a connector Sync. It is the same shape
// the engine persists to the contexts table and is classified on the 5
// Signal-Theory dimensions downstream. This struct is deliberately narrow:
// connectors emit the raw fields they know; classification fills the rest.
type Signal struct {
	ID         string   // external-system-stable ID (e.g. "slack:C123:16780.001")
	Path       string   // optimal:// URI pointing at the source
	Title      string   // 80-char-ish human label
	Content    string   // full raw body
	Genre      string   // "message", "email", "issue", "doc", …
	Entities   []string // people / orgs / tags lifted from payload
	CreatedAt  time.Time
	ModifiedAt time.Time
	Metadata   map[string]any // anything the adapter wants to attach
}

// Cursor is opaque to the engine. Each adapter defines its own shape
// (e.g. "<channel_id>:<oldest_ts>" for Slack). Empty string means "start".
type Cursor string

// ErrorCategory tells the Runner how to react to a sync failure.
type ErrorCategory string

const (
	// ErrRateLimited means back off by RetryAfter ms then retry with the same cursor.
	ErrRateLimited ErrorCategory = "rate_limited"
	// ErrAuthExpired means refresh the credential once and retry.
	ErrAuthExpired ErrorCategory = "auth_expired"
	// ErrTransient means exponential backoff up to 5 retries.
	ErrTransient ErrorCategory = "transient"
	// ErrFatal means disable the connector and surface to the operator.
	ErrFatal ErrorCategory = "fatal"
	// ErrNotImplemented is a sentinel for adapter scaffolds whose sync path is TBD.
	ErrNotImplemented ErrorCategory = "not_implemented"
)

// SyncError is the typed error returned by Sync. Runners switch on Category.
type SyncError struct {
	Category   ErrorCategory
	RetryAfter time.Duration // populated for ErrRateLimited
	Cause      error
}

func (e *SyncError) Error() string {
	if e.Cause != nil {
		return string(e.Category) + ": " + e.Cause.Error()
	}
	return string(e.Category)
}

// NewSyncError is a tiny ctor for typed errors.
func NewSyncError(cat ErrorCategory, cause error) *SyncError {
	return &SyncError{Category: cat, Cause: cause}
}

// SyncResult is what a Connector.Sync call hands back on success.
type SyncResult struct {
	Signals []Signal
	Next    Cursor
}

// Config is the per-connector JSON config persisted in the `connectors` table.
// The shape is adapter-defined; the Base helpers validate required keys.
type Config map[string]any

// Connector is the contract every enterprise connector must implement.
// It maps 1:1 to OptimalEngine.Connectors.Behaviour in the Elixir source.
type Connector interface {
	// Kind is the stable identifier ("slack", "gmail", …).
	Kind() string
	// DisplayName is the human label shown in UIs and logs.
	DisplayName() string
	// AuthScheme reports what kind of credential the adapter needs.
	AuthScheme() AuthScheme
	// RequiredConfigKeys are the keys the Base validator checks on Init.
	RequiredConfigKeys() []string
	// Init hydrates runtime state from Config (opens HTTP clients, decodes
	// credentials, validates scopes). The returned state is passed to Sync.
	Init(ctx context.Context, cfg Config) (State, error)
	// Sync fetches the next batch of signals for the given cursor. It must:
	//   * be idempotent — replaying the same cursor yields the same batch
	//   * bound its batch size (the Runner has no memory limit of its own)
	//   * never panic — return a *SyncError with a category instead
	Sync(ctx context.Context, state State, cursor Cursor) (SyncResult, error)
	// Transform maps one external payload to a Signal. Optional — some
	// adapters do the mapping inline inside Sync.
	Transform(raw map[string]any) (Signal, error)
}

// State is an adapter-private runtime handle returned from Init and threaded
// through Sync. Adapters cast it back to their concrete type.
type State any

// KnownErrorCategories mirrors Behaviour.known_error_reasons/0.
func KnownErrorCategories() []ErrorCategory {
	return []ErrorCategory{ErrRateLimited, ErrAuthExpired, ErrTransient, ErrFatal}
}
