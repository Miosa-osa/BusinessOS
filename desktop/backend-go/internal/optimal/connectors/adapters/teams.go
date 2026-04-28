package adapters

import (
	"context"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type teamsAdapter struct{}

func init() { connectors.Register(&teamsAdapter{}) }

func (t *teamsAdapter) Kind() string                      { return "teams" }
func (t *teamsAdapter) DisplayName() string               { return "Microsoft Teams" }
func (t *teamsAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (t *teamsAdapter) RequiredConfigKeys() []string      { return []string{} }

type teamsState struct {
	accessToken string
}

func (t *teamsAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &teamsState{accessToken: connectors.CredentialField(cfg, "access_token")}, nil
}

func (t *teamsAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform reads a Teams chatMessage. Body is HTML in the Graph API; we
// strip and use the first 80 chars as the title.
func (t *teamsAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "id")
	body := nestedString(raw, "body", "content")
	plain := connectors.StripHTML(body)

	sig := connectors.NewSignal(
		connectors.SignalID("teams", extID),
		connectors.ClampTitle(plain, 80),
		plain,
		connectors.SourceURI("teams", extID),
		"message",
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "lastModifiedDateTime"))
	if name := nestedString(raw, "from", "user", "displayName"); name != "" {
		sig.Entities = []string{name}
	}
	return sig, nil
}
