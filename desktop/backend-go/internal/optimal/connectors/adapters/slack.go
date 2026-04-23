// Package adapters registers every concrete Connector. Importing this
// package for side-effects populates the registry:
//
//	import _ "github.com/rhl/businessos-backend/internal/optimal/connectors/adapters"
//
// Ported from OptimalEngine.Connectors.Adapters.* (Elixir).
package adapters

import (
	"context"
	"strconv"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type slackAdapter struct{}

func init() { connectors.Register(&slackAdapter{}) }

func (s *slackAdapter) Kind() string                      { return "slack" }
func (s *slackAdapter) DisplayName() string               { return "Slack" }
func (s *slackAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (s *slackAdapter) RequiredConfigKeys() []string      { return []string{"workspace_id", "channels"} }

// slackState is the runtime handle threaded through Sync. Kept tiny so the
// same pattern can be copied into every adapter.
type slackState struct {
	workspaceID string
	botToken    string
}

func (s *slackAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, s.RequiredConfigKeys()); err != nil {
		return nil, err
	}
	if err := connectors.RequireCredentials(cfg, []string{"bot_token"}); err != nil {
		return nil, err
	}
	return &slackState{
		workspaceID: connectors.StringField(cfg, "workspace_id"),
		botToken:    connectors.CredentialField(cfg, "bot_token"),
	}, nil
}

// Sync is scaffolded to match the Elixir source which also returns
// :not_implemented. The real implementation lands in a dedicated Slack
// port phase once conversations.history / files.list are wired.
func (s *slackAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

func (s *slackAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := firstString(raw, "client_msg_id", "ts")
	text := stringField(raw, "text")
	user := stringField(raw, "user")
	if user == "" {
		user = "slack-user"
	}

	sig := connectors.NewSignal(
		connectors.SignalID("slack", extID),
		connectors.ClampTitle(text, 80),
		text,
		connectors.SourceURI("slack", extID),
		"message",
	)
	sig.Entities = []string{user}
	sig.ModifiedAt = parseSlackTS(stringField(raw, "ts"))
	return sig, nil
}

// parseSlackTS converts Slack's "1672531200.000100" format to time.Time.
func parseSlackTS(ts string) time.Time {
	if ts == "" {
		return time.Now().UTC()
	}
	f, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Now().UTC()
	}
	return time.Unix(int64(f), 0).UTC()
}
