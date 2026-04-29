package adapters

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

// GmailFetcher is the seam between the gmail adapter and the actual Gmail
// API. The adapter keeps no Gmail dependency; callers wire a concrete
// implementation via SetGmailFetcher. Without one, Sync returns
// ErrNotImplemented to preserve Phase-1 parity with the Elixir source.
//
// Two methods so the adapter can paginate and per-message-error-tolerantly
// fetch full payloads. ListIDs returns the next cursor empty-string when
// there are no more pages.
type GmailFetcher interface {
	ListIDs(ctx context.Context, userEmail, cursor string, max int) (ids []string, next string, err error)
	Get(ctx context.Context, userEmail, msgID string) (map[string]any, error)
}

// gmailAdapter holds an optional Fetcher. nil = stub mode. Mutex guards
// fetcher reads/writes since SetGmailFetcher may be called after init().
type gmailAdapter struct {
	mu      sync.RWMutex
	fetcher GmailFetcher
}

var globalGmailAdapter = &gmailAdapter{}

func init() { connectors.Register(globalGmailAdapter) }

// SetGmailFetcher wires a real Gmail fetcher to the registered gmail adapter.
// Idempotent; safe to call from bootstrap after the OAuth provider is ready.
// Pass nil to revert to stub behaviour (useful in tests).
func SetGmailFetcher(f GmailFetcher) {
	globalGmailAdapter.mu.Lock()
	defer globalGmailAdapter.mu.Unlock()
	globalGmailAdapter.fetcher = f
}

func (g *gmailAdapter) Kind() string                      { return "gmail" }
func (g *gmailAdapter) DisplayName() string               { return "Gmail" }
func (g *gmailAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (g *gmailAdapter) RequiredConfigKeys() []string      { return []string{"user_email"} }

// gmailState carries the better-auth user.id through Sync. Tokens live
// inside the fetcher implementation (which calls GetTokenSource on the
// google provider) — the adapter never touches them. user_email is a
// historical field name; semantically it's the user.id.
type gmailState struct {
	userEmail string
}

func (g *gmailAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, g.RequiredConfigKeys()); err != nil {
		return nil, err
	}
	return &gmailState{userEmail: connectors.StringField(cfg, "user_email")}, nil
}

// gmailBatchSize bounds how many message ids ListIDs returns per call.
// 50 keeps memory steady and matches Gmail's recommended page size.
const gmailBatchSize = 50

// Sync pulls the next batch of messages and converts each through Transform.
// Per-message Get/Transform errors are logged and skipped so a single bad
// message doesn't sink the whole batch — Gmail occasionally returns 404 for
// messages that were deleted between list and get, and that's fine.
//
// When fetcher is nil (stub mode) we return ErrNotImplemented unchanged,
// preserving parity with the Elixir source's :not_implemented tag.
func (g *gmailAdapter) Sync(ctx context.Context, state connectors.State, cursor connectors.Cursor) (connectors.SyncResult, error) {
	g.mu.RLock()
	fetcher := g.fetcher
	g.mu.RUnlock()
	if fetcher == nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
	}
	s, ok := state.(*gmailState)
	if !ok {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrFatal,
			errors.New("gmail: invalid state type"))
	}

	ids, next, err := fetcher.ListIDs(ctx, s.userEmail, string(cursor), gmailBatchSize)
	if err != nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrTransient, err)
	}

	signals := make([]connectors.Signal, 0, len(ids))
	for _, id := range ids {
		raw, err := fetcher.Get(ctx, s.userEmail, id)
		if err != nil {
			slog.WarnContext(ctx, "gmail: fetch message failed",
				"id", id, "user_email", s.userEmail, "error", err)
			continue
		}
		sig, err := g.Transform(raw)
		if err != nil {
			slog.WarnContext(ctx, "gmail: transform failed",
				"id", id, "user_email", s.userEmail, "error", err)
			continue
		}
		signals = append(signals, sig)
	}
	return connectors.SyncResult{Signals: signals, Next: connectors.Cursor(next)}, nil
}

// Transform maps a Gmail users.messages.get JSON payload onto a Signal.
// Matches the Elixir source header extraction (Subject, From) + internalDate.
func (g *gmailAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "id")
	subject := gmailHeader(raw, "Subject")
	if subject == "" {
		subject = "(no subject)"
	}
	from := gmailHeader(raw, "From")

	body := firstString(raw, "snippet", "body")
	content := connectors.StripHTML(body)

	sig := connectors.NewSignal(
		connectors.SignalID("gmail", extID),
		connectors.ClampTitle(subject, 120),
		content,
		connectors.SourceURI("gmail", extID),
		"email",
	)
	sig.Entities = connectors.ExtractEmails(from)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "internalDate"))
	return sig, nil
}

// gmailHeader walks raw["payload"]["headers"] which is a list of {name,value}.
func gmailHeader(raw map[string]any, name string) string {
	payload := nestedMap(raw, "payload")
	if payload == nil {
		return ""
	}
	headers := anyList(payload, "headers")
	for _, h := range headers {
		hm, ok := h.(map[string]any)
		if !ok {
			continue
		}
		if stringField(hm, "name") == name {
			return stringField(hm, "value")
		}
	}
	return ""
}
