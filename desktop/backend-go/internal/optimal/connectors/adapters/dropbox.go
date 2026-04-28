package adapters

import (
	"context"
	"path"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type dropboxAdapter struct{}

func init() { connectors.Register(&dropboxAdapter{}) }

func (d *dropboxAdapter) Kind() string                      { return "dropbox" }
func (d *dropboxAdapter) DisplayName() string               { return "Dropbox" }
func (d *dropboxAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (d *dropboxAdapter) RequiredConfigKeys() []string      { return []string{} }

type dropboxState struct {
	accessToken string
}

func (d *dropboxAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &dropboxState{accessToken: connectors.CredentialField(cfg, "access_token")}, nil
}

func (d *dropboxAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform reads a Dropbox files/list_folder entry. ext_id falls back to
// path_lower when id is missing (folder entries don't always have ids).
// Title falls back to basename of path_display.
func (d *dropboxAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := firstString(raw, "id", "path_lower")
	name := stringField(raw, "name")
	if name == "" {
		name = path.Base(stringField(raw, "path_display"))
	}
	sig := connectors.NewSignal(
		connectors.SignalID("dropbox", extID),
		name,
		stringField(raw, "content"),
		connectors.SourceURI("dropbox", extID),
		"file",
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "server_modified"))
	return sig, nil
}
