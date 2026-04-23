package connectors

import (
	"fmt"
	"strings"
	"time"
)

// Transform helpers shared by every adapter. These mirror
// OptimalEngine.Connectors.Transform in the Elixir source.

// SignalID builds the engine-global signal ID from a kind and an external
// identifier. The prefix is what downstream classifiers use to route to the
// right node ("slack/…", "gmail/…").
func SignalID(kind, externalID string) string {
	if externalID == "" {
		return fmt.Sprintf("%s/unknown-%d", kind, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s/%s", kind, externalID)
}

// SourceURI builds the optimal:// URI that points back at the source record.
// It is stored on the Signal so the graph can link back to the origin.
func SourceURI(kind, externalID string) string {
	return fmt.Sprintf("optimal://%s/%s", kind, externalID)
}

// ClampTitle trims a body to a bounded length for titling. The Elixir side
// uses 80 chars; we use 80 here too for parity. Handles Unicode by rune.
func ClampTitle(body string, max int) string {
	if max <= 0 {
		max = 80
	}
	r := []rune(strings.TrimSpace(body))
	if len(r) <= max {
		return string(r)
	}
	return strings.TrimSpace(string(r[:max]))
}

// NewSignal is the factory used by Transform implementations. It fills
// sensible defaults (timestamps, empty slices) so adapters only set what
// differs. Callers override any field after the call returns.
func NewSignal(id, title, content, path, genre string) Signal {
	now := time.Now().UTC()
	return Signal{
		ID:         id,
		Title:      title,
		Content:    content,
		Path:       path,
		Genre:      genre,
		Entities:   []string{},
		CreatedAt:  now,
		ModifiedAt: now,
		Metadata:   map[string]any{},
	}
}

// RequireKeys validates that every key in required is present and non-zero
// in cfg. Matches the Base macro's require_keys/2.
func RequireKeys(cfg Config, required []string) error {
	var missing []string
	for _, k := range required {
		if v, ok := cfg[k]; !ok || isZero(v) {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("connectors: missing required config keys: %s", strings.Join(missing, ", "))
	}
	return nil
}

// RequireCredentials checks a nested `credentials` submap — the same shape
// DecryptCredentials returns. Used by Base to validate adapters have what
// they need before Sync is called.
func RequireCredentials(cfg Config, required []string) error {
	raw, ok := cfg["credentials"]
	if !ok {
		if len(required) == 0 {
			return nil
		}
		return fmt.Errorf("connectors: credentials missing; required %s", strings.Join(required, ", "))
	}
	creds, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("connectors: credentials must be an object")
	}
	var missing []string
	for _, k := range required {
		if v, ok := creds[k]; !ok || isZero(v) {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("connectors: missing credential keys: %s", strings.Join(missing, ", "))
	}
	return nil
}

// StringField looks up a key in Config returning "" when missing or non-string.
// Adapters use this during Init to pull workspace IDs, domains, etc.
func StringField(cfg Config, key string) string {
	v, ok := cfg[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// CredentialField reaches into cfg["credentials"][key]. Returns "" on miss.
func CredentialField(cfg Config, key string) string {
	raw, ok := cfg["credentials"]
	if !ok {
		return ""
	}
	creds, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	v, ok := creds[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func isZero(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	}
	return false
}
