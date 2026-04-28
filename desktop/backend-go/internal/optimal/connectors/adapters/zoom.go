package adapters

import (
	"context"
	"fmt"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type zoomAdapter struct{}

func init() { connectors.Register(&zoomAdapter{}) }

func (z *zoomAdapter) Kind() string                      { return "zoom" }
func (z *zoomAdapter) DisplayName() string               { return "Zoom" }
func (z *zoomAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (z *zoomAdapter) RequiredConfigKeys() []string      { return []string{} }

type zoomState struct {
	accessToken string
}

func (z *zoomAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &zoomState{accessToken: connectors.CredentialField(cfg, "access_token")}, nil
}

func (z *zoomAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform reads a Zoom recording / transcript payload. Genre is
// "transcript" because the only useful body Zoom exposes via API is the
// transcript text — meeting metadata flows through other genres.
func (z *zoomAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	rawID := raw["uuid"]
	if rawID == nil {
		rawID = raw["id"]
	}
	extID := fmt.Sprintf("%v", rawID)

	title := stringField(raw, "topic")
	if title == "" {
		title = "Zoom meeting"
	}
	sig := connectors.NewSignal(
		connectors.SignalID("zoom", extID),
		title,
		stringField(raw, "transcript"),
		connectors.SourceURI("zoom", extID),
		"transcript",
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "start_time"))
	return sig, nil
}
