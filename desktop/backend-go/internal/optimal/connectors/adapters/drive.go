package adapters

import (
	"context"
	"strings"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

type driveAdapter struct{}

func init() { connectors.Register(&driveAdapter{}) }

func (d *driveAdapter) Kind() string                      { return "drive" }
func (d *driveAdapter) DisplayName() string               { return "Google Drive" }
func (d *driveAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (d *driveAdapter) RequiredConfigKeys() []string      { return []string{} }

type driveState struct {
	accessToken string
}

func (d *driveAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireCredentials(cfg, []string{"access_token"}); err != nil {
		return nil, err
	}
	return &driveState{accessToken: connectors.CredentialField(cfg, "access_token")}, nil
}

func (d *driveAdapter) Sync(_ context.Context, _ connectors.State, _ connectors.Cursor) (connectors.SyncResult, error) {
	return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
}

// Transform inspects mimeType to set genre/mode/format. The Elixir source
// special-cases Google's native types (document/spreadsheet/presentation)
// and image/*. Everything else lands as a generic "file".
func (d *driveAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "id")
	name := stringField(raw, "name")
	if name == "" {
		name = "Untitled"
	}
	mime := stringField(raw, "mimeType")

	content := firstString(raw, "exportedText", "content")

	sig := connectors.NewSignal(
		connectors.SignalID("drive", extID),
		name,
		content,
		connectors.SourceURI("drive", extID),
		driveGenre(mime),
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "modifiedTime"))
	sig.Metadata["mode"] = driveMode(mime)
	sig.Metadata["format"] = driveFormat(mime)
	sig.Metadata["mime_type"] = mime
	return sig, nil
}

func driveGenre(mime string) string {
	switch mime {
	case "application/vnd.google-apps.document":
		return "document"
	case "application/vnd.google-apps.spreadsheet":
		return "table"
	case "application/vnd.google-apps.presentation":
		return "slides"
	case "application/pdf":
		return "document"
	default:
		return "file"
	}
}

func driveMode(mime string) string {
	switch {
	case mime == "application/vnd.google-apps.spreadsheet":
		return "data"
	case strings.HasPrefix(mime, "image/"):
		return "visual"
	default:
		return "linguistic"
	}
}

func driveFormat(mime string) string {
	switch mime {
	case "application/vnd.google-apps.document":
		return "markdown"
	case "application/vnd.google-apps.spreadsheet":
		return "json"
	default:
		return "unknown"
	}
}
