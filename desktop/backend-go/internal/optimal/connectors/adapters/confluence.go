package adapters

import (
	"context"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type confluenceAdapter struct{}

func init() { connectors.Register(&confluenceAdapter{}) }

func (c *confluenceAdapter) Kind() string                      { return "confluence" }
func (c *confluenceAdapter) DisplayName() string               { return "Confluence" }
func (c *confluenceAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (c *confluenceAdapter) RequiredConfigKeys() []string      { return []string{"site_url"} }

type confluenceState struct {
	siteURL string
	token   string
}

func (c *confluenceAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, c.RequiredConfigKeys()); err != nil {
		return nil, err
	}
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &confluenceState{
		siteURL: connectors.StringField(cfg, "site_url"),
		token:   connectors.CredentialField(cfg, "access_token"),
	}, nil
}

func (c *confluenceAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform reads a Confluence page payload (storage representation).
// Body lives at body.storage.value as HTML; we strip and store as plain text.
func (c *confluenceAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "id")
	title := stringField(raw, "title")
	if title == "" {
		title = "Untitled"
	}
	body := nestedString(raw, "body", "storage", "value")
	sig := connectors.NewSignal(
		connectors.SignalID("confluence", extID),
		title,
		connectors.StripHTML(body),
		connectors.SourceURI("confluence", extID),
		"document",
	)
	sig.ModifiedAt = connectors.ParseISO8601(nestedString(raw, "version", "when"))
	return sig, nil
}
