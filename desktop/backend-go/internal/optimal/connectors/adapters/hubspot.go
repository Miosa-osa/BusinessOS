package adapters

import (
	"context"
	"fmt"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type hubspotAdapter struct{}

func init() { connectors.Register(&hubspotAdapter{}) }

func (h *hubspotAdapter) Kind() string                      { return "hubspot" }
func (h *hubspotAdapter) DisplayName() string               { return "HubSpot" }
func (h *hubspotAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (h *hubspotAdapter) RequiredConfigKeys() []string      { return []string{} }

type hubspotState struct {
	accessToken  string
	clientID     string
	clientSecret string
	refreshToken string
}

// Init accepts either a static access_token OR an OAuth refresh triple
// (client_id + client_secret + refresh_token). Matches the Elixir
// require_hubspot_auth/1 branching.
func (h *hubspotAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	tok := connectors.CredentialField(cfg, "access_token")
	cID := connectors.CredentialField(cfg, "client_id")
	cSec := connectors.CredentialField(cfg, "client_secret")
	rTok := connectors.CredentialField(cfg, "refresh_token")

	hasOAuth := cID != "" && cSec != "" && rTok != ""
	if tok == "" && !hasOAuth {
		return nil, fmt.Errorf("hubspot: missing credentials — supply access_token OR (client_id+client_secret+refresh_token)")
	}
	return &hubspotState{
		accessToken:  tok,
		clientID:     cID,
		clientSecret: cSec,
		refreshToken: rTok,
	}, nil
}

func (h *hubspotAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform handles deals, contacts, companies, tickets — anything from
// HubSpot's CRM. We pull title from whichever name field exists, falling
// back to the bare id.
func (h *hubspotAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := fmt.Sprintf("%v", raw["id"])
	props, _ := raw["properties"].(map[string]any)
	if props == nil {
		props = map[string]any{}
	}

	title := firstString(props, "dealname", "firstname", "company", "subject")
	if title == "" {
		title = extID
	}
	content := firstString(props, "description", "notes_last_contacted")

	sig := connectors.NewSignal(
		connectors.SignalID("hubspot", extID),
		title,
		content,
		connectors.SourceURI("hubspot", extID),
		"crm",
	)
	modified := stringField(props, "hs_lastmodifieddate")
	if modified == "" {
		modified = stringField(raw, "updatedAt")
	}
	sig.ModifiedAt = connectors.ParseISO8601(modified)
	return sig, nil
}
