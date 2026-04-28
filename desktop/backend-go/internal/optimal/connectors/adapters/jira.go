package adapters

import (
	"context"
	"fmt"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type jiraAdapter struct{}

func init() { connectors.Register(&jiraAdapter{}) }

func (j *jiraAdapter) Kind() string                      { return "jira" }
func (j *jiraAdapter) DisplayName() string               { return "Jira" }
func (j *jiraAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (j *jiraAdapter) RequiredConfigKeys() []string      { return []string{"site_url"} }

type jiraState struct {
	siteURL string
	token   string
}

func (j *jiraAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, j.RequiredConfigKeys()); err != nil {
		return nil, err
	}
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &jiraState{
		siteURL: connectors.StringField(cfg, "site_url"),
		token:   connectors.CredentialField(cfg, "access_token"),
	}, nil
}

func (j *jiraAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform reads a Jira issue payload. The key (e.g. "PROJ-123") is the
// stable external id; falls back to numeric id when key is missing.
func (j *jiraAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := firstString(raw, "key", "id")
	fields, _ := raw["fields"].(map[string]any)
	if fields == nil {
		fields = map[string]any{}
	}

	title := fmt.Sprintf("[%s] %s", extID, stringField(fields, "summary"))
	sig := connectors.NewSignal(
		connectors.SignalID("jira", extID),
		title,
		connectors.StripHTML(stringField(fields, "description")),
		connectors.SourceURI("jira", extID),
		"ticket",
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(fields, "updated"))

	entities := []string{}
	if name := nestedString(fields, "assignee", "displayName"); name != "" {
		entities = append(entities, name)
	}
	if name := nestedString(fields, "reporter", "displayName"); name != "" {
		entities = append(entities, name)
	}
	sig.Entities = entities
	return sig, nil
}
