package adapters

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

// NotionFetcher is the seam between the notion adapter and Notion's API.
type NotionFetcher interface {
	ListPages(ctx context.Context, userID, cursor string, max int) (pages []map[string]any, next string, err error)
}

type notionAdapter struct {
	mu      sync.RWMutex
	fetcher NotionFetcher
}

var globalNotionAdapter = &notionAdapter{}

func init() { connectors.Register(globalNotionAdapter) }

// SetNotionFetcher wires a real Notion fetcher to the registered adapter.
func SetNotionFetcher(f NotionFetcher) {
	globalNotionAdapter.mu.Lock()
	defer globalNotionAdapter.mu.Unlock()
	globalNotionAdapter.fetcher = f
}

func (n *notionAdapter) Kind() string                      { return "notion" }
func (n *notionAdapter) DisplayName() string               { return "Notion" }
func (n *notionAdapter) AuthScheme() connectors.AuthScheme { return connectors.AuthOAuth2 }
func (n *notionAdapter) RequiredConfigKeys() []string      { return []string{} }

type notionState struct {
	userID string
}

func (n *notionAdapter) Init(_ context.Context, cfg connectors.Config) (connectors.State, error) {
	if err := connectors.RequireKeys(cfg, []string{"user_email"}); err != nil {
		return nil, err
	}
	return &notionState{userID: connectors.StringField(cfg, "user_email")}, nil
}

func (n *notionAdapter) Sync(ctx context.Context, state connectors.State, cursor connectors.Cursor) (connectors.SyncResult, error) {
	n.mu.RLock()
	fetcher := n.fetcher
	n.mu.RUnlock()
	if fetcher == nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrNotImplemented, nil)
	}
	st, ok := state.(*notionState)
	if !ok {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrFatal,
			errors.New("notion: invalid state type"))
	}
	pages, next, err := fetcher.ListPages(ctx, st.userID, string(cursor), 50)
	if err != nil {
		return connectors.SyncResult{}, connectors.NewSyncError(connectors.ErrTransient, err)
	}
	signals := make([]connectors.Signal, 0, len(pages))
	for _, p := range pages {
		sig, err := n.Transform(p)
		if err != nil {
			slog.WarnContext(ctx, "notion: transform failed", "user_id", st.userID, "error", err)
			continue
		}
		signals = append(signals, sig)
	}
	return connectors.SyncResult{Signals: signals, Next: connectors.Cursor(next)}, nil
}

// Transform maps a Notion page (with optional `blocks`) onto a Signal.
func (n *notionAdapter) Transform(raw map[string]any) (connectors.Signal, error) {
	extID := stringField(raw, "id")
	title := notionTitle(raw)
	if title == "" {
		title = "Untitled"
	}

	blocks, _ := raw["blocks"].([]any)
	content := flattenNotionBlocks(blocks)

	sig := connectors.NewSignal(
		connectors.SignalID("notion", extID),
		title,
		content,
		connectors.SourceURI("notion", extID),
		"document",
	)
	sig.ModifiedAt = connectors.ParseISO8601(stringField(raw, "last_edited_time"))
	sig.Metadata["format"] = "markdown"
	return sig, nil
}

// notionTitle reaches into the known Notion property shapes to find a title.
// Notion exposes titles at properties.Name.title[0].plain_text or
// properties.title.title[0].plain_text depending on schema.
func notionTitle(raw map[string]any) string {
	props := nestedMap(raw, "properties")
	if props == nil {
		return ""
	}
	for _, key := range []string{"Name", "title"} {
		if t := firstPlainText(props, key); t != "" {
			return t
		}
	}
	return ""
}

func firstPlainText(props map[string]any, key string) string {
	holder, ok := props[key].(map[string]any)
	if !ok {
		return ""
	}
	rt, ok := holder["title"].([]any)
	if !ok || len(rt) == 0 {
		return ""
	}
	first, ok := rt[0].(map[string]any)
	if !ok {
		return ""
	}
	return stringField(first, "plain_text")
}

// flattenNotionBlocks converts an ordered block list into markdown-ish text.
// Supports paragraph, heading_1, heading_2 — same set as the Elixir source.
func flattenNotionBlocks(blocks []any) string {
	var parts []string
	for _, b := range blocks {
		bm, ok := b.(map[string]any)
		if !ok {
			continue
		}
		kind := stringField(bm, "type")
		rt := blockRichText(bm, kind)
		text := richTextToPlain(rt)
		if text == "" {
			continue
		}
		switch kind {
		case "heading_1":
			parts = append(parts, "# "+text)
		case "heading_2":
			parts = append(parts, "## "+text)
		case "paragraph":
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

func blockRichText(bm map[string]any, kind string) []any {
	if kind == "" {
		return nil
	}
	holder, ok := bm[kind].(map[string]any)
	if !ok {
		return nil
	}
	rt, _ := holder["rich_text"].([]any)
	return rt
}

func richTextToPlain(rt []any) string {
	if len(rt) == 0 {
		return ""
	}
	var b strings.Builder
	for _, r := range rt {
		rm, ok := r.(map[string]any)
		if !ok {
			continue
		}
		b.WriteString(stringField(rm, "plain_text"))
	}
	return b.String()
}
