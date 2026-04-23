package adapters

import (
	"context"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type linearAdapter struct{}

func init() { connectors.Register(&linearAdapter{}) }

func (l *linearAdapter) Kind() string                      { return "linear" }
func (l *linearAdapter) DisplayName() string               { return "Linear" }
func (l *linearAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthToken }
func (l *linearAdapter) RequiredConfigKeys() []string      { return []string{} }

type linearState struct {
	apiKey string
}

func (l *linearAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireCredentials(cfg, []string{"api_key"}); err != nil {
		return nil, err
	}
	return &linearState{apiKey: connectors.CredentialField(cfg, "api_key")}, nil
}

func (l *linearAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
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
