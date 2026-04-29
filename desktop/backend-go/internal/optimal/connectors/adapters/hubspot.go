package adapters

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

// HubspotFetcher is the seam between the hubspot adapter and HubSpot's API.
// Records may be contacts, companies, or deals — the fetcher picks the
// active sObjects and the adapter's Transform maps them uniformly.
type HubspotFetcher interface {
	ListRecords(ctx context.Context, userID, cursor string, max int) (records []map[string]any, next string, err error)
}

type hubspotAdapter struct {
	mu      sync.RWMutex
	fetcher HubspotFetcher
}

var globalHubspotAdapter = &hubspotAdapter{}

func init() { connectors.Register(globalHubspotAdapter) }

// SetHubspotFetcher wires a real HubSpot fetcher to the registered adapter.
func SetHubspotFetcher(f HubspotFetcher) {
	globalHubspotAdapter.mu.Lock()
	defer globalHubspotAdapter.mu.Unlock()
	globalHubspotAdapter.fetcher = f
}

func (h *hubspotAdapter) Kind() string                      { return "hubspot" }
func (h *hubspotAdapter) DisplayName() string               { return "HubSpot" }
func (h *hubspotAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (h *hubspotAdapter) RequiredConfigKeys() []string      { return []string{} }

type hubspotState struct {
	userID string
}

// Init only needs the better-auth user.id; OAuth state is owned by the
// HubSpot Provider in BusinessOS, not this adapter.
func (h *hubspotAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, []string{"user_email"}); err != nil {
		return nil, err
	}
	return &hubspotState{userID: connectors.StringField(cfg, "user_email")}, nil
}

func (h *hubspotAdapter) Sync(ctx context.Context, state connectors.State, cursor connectors.Cursor) (connectors.SyncResult, error) {
	h.mu.RLock()
	fetcher := h.fetcher
	h.mu.RUnlock()
	if fetcher == nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
	}
	st, ok := state.(*hubspotState)
	if !ok {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrFatal,
			errors.New("hubspot: invalid state type"))
	}
	records, next, err := fetcher.ListRecords(ctx, st.userID, string(cursor), 50)
	if err != nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrTransient, err)
	}
	signals := make([]connectors.Signal, 0, len(records))
	for _, r := range records {
		sig, err := h.Transform(r)
		if err != nil {
			slog.WarnContext(ctx, "hubspot: transform failed", "user_id", st.userID, "error", err)
			continue
		}
		signals = append(signals, sig)
	}
	return connectors.SyncResult{Signals: signals, Next: connectors.Cursor(next)}, nil
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
