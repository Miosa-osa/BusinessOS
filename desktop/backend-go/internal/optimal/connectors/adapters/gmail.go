package adapters

import (
	"context"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type gmailAdapter struct{}

func init() { connectors.Register(&gmailAdapter{}) }

func (g *gmailAdapter) Kind() string                      { return "gmail" }
func (g *gmailAdapter) DisplayName() string               { return "Gmail" }
func (g *gmailAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (g *gmailAdapter) RequiredConfigKeys() []string      { return []string{"user_email"} }

type gmailState struct {
	userEmail    string
	accessToken  string
	refreshToken string
}

func (g *gmailAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, g.RequiredConfigKeys()); err != nil {
		return nil, err
	}
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &gmailState{
		userEmail:    connectors.StringField(cfg, "user_email"),
		accessToken:  connectors.CredentialField(cfg, "access_token"),
		refreshToken: connectors.CredentialField(cfg, "refresh_token"),
	}, nil
}

func (g *gmailAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform maps a Gmail users.messages.get JSON payload onto a Signal.
// Matches the Elixir source header extraction (Subject, From) + internalDate.
func (g *gmailAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "id")
	subject := gmailHeader(raw, "Subject")
	if subject == "" {
		subject = "(no subject)"
	}
	from := gmailHeader(raw, "From")

	body := firstString(raw, "snippet", "body")
	content := connectors.StripHTML(body)

	sig := connectors.NewSignal(
		connectors.SignalID("gmail", extID),
		connectors.ClampTitle(subject, 120),
		content,
		connectors.SourceURI("gmail", extID),
		"email",
	)
	sig.Entities = connectors.ExtractEmails(from)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "internalDate"))
	return sig, nil
}

// gmailHeader walks raw["payload"]["headers"] which is a list of {name,value}.
func gmailHeader(raw map[string]any, name string) string {
	payload := nestedMap(raw, "payload")
	if payload == nil {
		return ""
	}
	headers := anyList(payload, "headers")
	for _, h := range headers {
		hm, ok := h.(map[string]any)
		if !ok {
			continue
		}
		if stringField(hm, "name") == name {
			return stringField(hm, "value")
		}
	}
	return ""
}
