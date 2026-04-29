package adapters

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

// LinearFetcher is the seam between the linear adapter and Linear's GraphQL.
type LinearFetcher interface {
	ListIssues(ctx context.Context, userID, cursor string, max int) (issues []map[string]any, next string, err error)
}

type linearAdapter struct {
	mu      sync.RWMutex
	fetcher LinearFetcher
}

var globalLinearAdapter = &linearAdapter{}

func init() { connectors.Register(globalLinearAdapter) }

// SetLinearFetcher wires a real Linear fetcher to the registered adapter.
func SetLinearFetcher(f LinearFetcher) {
	globalLinearAdapter.mu.Lock()
	defer globalLinearAdapter.mu.Unlock()
	globalLinearAdapter.fetcher = f
}

func (l *linearAdapter) Kind() string                      { return "linear" }
func (l *linearAdapter) DisplayName() string               { return "Linear" }
func (l *linearAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (l *linearAdapter) RequiredConfigKeys() []string      { return []string{} }

type linearState struct {
	userID string
}

func (l *linearAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, []string{"user_email"}); err != nil {
		return nil, err
	}
	return &linearState{userID: connectors.StringField(cfg, "user_email")}, nil
}

func (l *linearAdapter) Sync(ctx context.Context, state connectors.State, cursor connectors.Cursor) (connectors.SyncResult, error) {
	l.mu.RLock()
	fetcher := l.fetcher
	l.mu.RUnlock()
	if fetcher == nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
	}
	st, ok := state.(*linearState)
	if !ok {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrFatal,
			errors.New("linear: invalid state type"))
	}
	issues, next, err := fetcher.ListIssues(ctx, st.userID, string(cursor), 50)
	if err != nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrTransient, err)
	}
	signals := make([]connectors.Signal, 0, len(issues))
	for _, i := range issues {
		sig, err := l.Transform(i)
		if err != nil {
			slog.WarnContext(ctx, "linear: transform failed", "user_id", st.userID, "error", err)
			continue
		}
		signals = append(signals, sig)
	}
	return connectors.SyncResult{Signals: signals, Next: connectors.Cursor(next)}, nil
}

// Transform reads a Linear Issue GraphQL node into a Signal.
func (l *linearAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "id")
	title := stringField(raw, "title")
	if title == "" {
		title = "Untitled"
	}
	sig := connectors.NewSignal(
		connectors.SignalID("linear", extID),
		title,
		stringField(raw, "description"),
		connectors.SourceURI("linear", extID),
		"ticket",
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "updatedAt"))

	entities := []string{}
	if name := nestedString(raw, "assignee", "name"); name != "" {
		entities = append(entities, name)
	}
	if name := nestedString(raw, "creator", "name"); name != "" {
		entities = append(entities, name)
	}
	sig.Entities = entities
	return sig, nil
}
