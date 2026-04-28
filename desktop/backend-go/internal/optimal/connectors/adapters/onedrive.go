package adapters

import (
	"context"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type onedriveAdapter struct{}

func init() { connectors.Register(&onedriveAdapter{}) }

func (o *onedriveAdapter) Kind() string                      { return "onedrive" }
func (o *onedriveAdapter) DisplayName() string               { return "OneDrive" }
func (o *onedriveAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (o *onedriveAdapter) RequiredConfigKeys() []string      { return []string{} }

type onedriveState struct {
	accessToken string
}

func (o *onedriveAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &onedriveState{accessToken: connectors.CredentialField(cfg, "access_token")}, nil
}

func (o *onedriveAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform reads a Microsoft Graph drive item.
func (o *onedriveAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "id")
	name := stringField(raw, "name")
	if name == "" {
		name = "Untitled"
	}
	sig := connectors.NewSignal(
		connectors.SignalID("onedrive", extID),
		name,
		stringField(raw, "content"),
		connectors.SourceURI("onedrive", extID),
		"file",
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "lastModifiedDateTime"))
	return sig, nil
}
