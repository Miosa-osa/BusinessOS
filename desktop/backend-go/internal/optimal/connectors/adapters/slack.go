// Package adapters registers every concrete Connector. Importing this
// package for side-effects populates the registry:
//
//	import _ "github.com/rhl/businessos-backend/internal/optimal/connectors/adapters"
//
// Ported from OptimalEngine.Connectors.Adapters.* (Elixir).
package adapters

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

// SlackFetcher is the seam between the slack adapter and Slack's REST API.
// Bootstrap wires a concrete implementation via SetSlackFetcher; without
// one, Sync returns ErrNotImplemented to preserve Phase-1 stub parity.
type SlackFetcher interface {
	ListMessages(ctx context.Context, userID, cursor string, max int) (msgs []map[string]any, next string, err error)
}

type slackAdapter struct {
	mu      sync.RWMutex
	fetcher SlackFetcher
}

var globalSlackAdapter = &slackAdapter{}

func init() { connectors.Register(globalSlackAdapter) }

// SetSlackFetcher wires a real Slack fetcher to the registered adapter.
// Idempotent; safe to call from bootstrap after the OAuth provider is ready.
func SetSlackFetcher(f SlackFetcher) {
	globalSlackAdapter.mu.Lock()
	defer globalSlackAdapter.mu.Unlock()
	globalSlackAdapter.fetcher = f
}

func (s *slackAdapter) Kind() string                      { return "slack" }
func (s *slackAdapter) DisplayName() string               { return "Slack" }
func (s *slackAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (s *slackAdapter) RequiredConfigKeys() []string      { return []string{} }

// slackState carries the better-auth user id through Sync. The bot token
// stays inside the fetcher implementation; the adapter never touches it.
type slackState struct {
	userID string
}

func (s *slackAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	// user_email is set to the better-auth user.id by the handler — same
	// convention as the gmail adapter (the field name is historical).
	if err := connectors.RequireKeys(cfg, []string{"user_email"}); err != nil {
		return nil, err
	}
	return &slackState{userID: connectors.StringField(cfg, "user_email")}, nil
}

// Sync pulls the next batch of messages and converts each through Transform.
// Per-message Transform errors are logged and skipped; a single malformed
// payload should not sink the whole batch.
func (s *slackAdapter) Sync(ctx context.Context, state connectors.State, cursor connectors.Cursor) (connectors.SyncResult, error) {
	s.mu.RLock()
	fetcher := s.fetcher
	s.mu.RUnlock()
	if fetcher == nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
	}
	st, ok := state.(*slackState)
	if !ok {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrFatal,
			errors.New("slack: invalid state type"))
	}
	msgs, next, err := fetcher.ListMessages(ctx, st.userID, string(cursor), 50)
	if err != nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrTransient, err)
	}
	signals := make([]connectors.Signal, 0, len(msgs))
	for _, m := range msgs {
		sig, err := s.Transform(m)
		if err != nil {
			slog.WarnContext(ctx, "slack: transform failed", "user_id", st.userID, "error", err)
			continue
		}
		signals = append(signals, sig)
	}
	return connectors.SyncResult{Signals: signals, Next: connectors.Cursor(next)}, nil
}

func (s *slackAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := firstString(raw, "client_msg_id", "ts")
	text := stringField(raw, "text")
	user := stringField(raw, "user")
	if user == "" {
		user = "slack-user"
	}

	sig := connectors.NewSignal(
		connectors.SignalID("slack", extID),
		connectors.ClampTitle(text, 80),
		text,
		connectors.SourceURI("slack", extID),
		"message",
	)
	sig.Entities = []string{user}
	sig.ModifiedAt = parseSlackTS(stringField(raw, "ts"))
	return sig, nil
}

// parseSlackTS converts Slack's "1672531200.000100" format to time.Time.
func parseSlackTS(ts string) time.Time {
	if ts == "" {
		return time.Now().UTC()
	}
	f, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Now().UTC()
	}
	return time.Unix(int64(f), 0).UTC()
}
