package adapters

import (
	"context"
	"fmt"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type githubAdapter struct{}

func init() { connectors.Register(&githubAdapter{}) }

func (g *githubAdapter) Kind() string                      { return "github" }
func (g *githubAdapter) DisplayName() string               { return "GitHub" }
func (g *githubAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (g *githubAdapter) RequiredConfigKeys() []string      { return []string{} }

type githubState struct {
	pat            string
	appID          string
	installationID string
	privateKeyPEM  string
}

// Init accepts either a Personal Access Token (pat) OR a GitHub App
// triple (app_id + installation_id + private_key_pem). Matches the Elixir
// require_github_auth/1 branching.
func (g *githubAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	pat := connectors.CredentialField(cfg, "pat")

	appID := connectors.CredentialField(cfg, "app_id")
	installID := connectors.CredentialField(cfg, "installation_id")
	pem := connectors.CredentialField(cfg, "private_key_pem")

	hasApp := appID != "" && installID != "" && pem != ""
	if pat == "" && !hasApp {
		return nil, fmt.Errorf("github: missing credentials — supply pat OR (app_id+installation_id+private_key_pem)")
	}
	return &githubState{
		pat:            pat,
		appID:          appID,
		installationID: installID,
		privateKeyPEM:  pem,
	}, nil
}

func (g *githubAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform accepts the REST issue/PR payload. When raw["pull_request"] is
// present the signal is tagged genre="pr" otherwise "issue".
func (g *githubAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "node_id")
	if extID == "" {
		extID = fmt.Sprintf("%v", raw["id"])
	}
	genre := "issue"
	if _, ok := raw["pull_request"]; ok {
		genre = "pr"
	}
	title := fmt.Sprintf("[%s] %s", genre, stringField(raw, "title"))

	sig := connectors.NewSignal(
		connectors.SignalID("github", extID),
		title,
		stringField(raw, "body"),
		connectors.SourceURI("github", extID),
		genre,
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "updated_at"))
	if login := nestedString(raw, "user", "login"); login != "" {
		sig.Entities = []string{login}
	}
	return sig, nil
}
