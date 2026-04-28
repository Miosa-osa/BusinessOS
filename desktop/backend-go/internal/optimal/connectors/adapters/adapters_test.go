package adapters

import (
	"strings"
	"testing"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

// All 14 adapters must register themselves on package import. Exercising
// Lookup guarantees that side-effect registration is wired correctly — a
// missed init() would make a kind silently disappear at runtime.
func TestAdaptersRegistered(t *testing.T) {
	want := []string{
		// Wave 2 priority five
		"slack", "gmail", "notion", "github", "linear",
		// Wave 7 — remaining nine
		"confluence", "drive", "dropbox", "hubspot", "jira",
		"onedrive", "salesforce", "teams", "zoom",
	}
	for _, kind := range want {
		if _, err := connectors.Lookup(kind); err != nil {
			t.Errorf("adapter %s not registered: %v", kind, err)
		}
	}
	got := connectors.Kinds()
	if len(got) < len(want) {
		t.Fatalf("registry shorter than expected: got %d want >= %d (%v)", len(got), len(want), got)
	}
}

func TestSlackTransform(t *testing.T) {
	a, _ := connectors.Lookup("slack")
	sig, err := a.Transform(map[string]any{
		"ts":   "1672531200.000100",
		"user": "U123",
		"text": "shipping the port today",
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig.Genre != "message" || !strings.HasPrefix(sig.ID, "slack/") {
		t.Fatalf("unexpected signal: %+v", sig)
	}
	if sig.Entities[0] != "U123" {
		t.Fatalf("entity not extracted: %+v", sig.Entities)
	}
}

func TestGmailTransform(t *testing.T) {
	a, _ := connectors.Lookup("gmail")
	sig, err := a.Transform(map[string]any{
		"id": "msg-abc",
		"payload": map[string]any{
			"headers": []any{
				map[string]any{"name": "Subject", "value": "Project kickoff"},
				map[string]any{"name": "From", "value": "Alice <alice@example.com>"},
			},
		},
		"snippet": "<p>hello</p> world",
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig.Title != "Project kickoff" {
		t.Fatalf("expected subject as title, got %q", sig.Title)
	}
	if sig.Content != "hello world" {
		t.Fatalf("html not stripped: %q", sig.Content)
	}
	if len(sig.Entities) == 0 || sig.Entities[0] != "alice@example.com" {
		t.Fatalf("email not extracted: %+v", sig.Entities)
	}
}

func TestNotionTransformHeadings(t *testing.T) {
	a, _ := connectors.Lookup("notion")
	sig, err := a.Transform(map[string]any{
		"id":               "page-1",
		"last_edited_time": "2026-04-20T10:00:00Z",
		"properties": map[string]any{
			"Name": map[string]any{
				"title": []any{map[string]any{"plain_text": "Q2 Plan"}},
			},
		},
		"blocks": []any{
			map[string]any{
				"type":      "heading_1",
				"heading_1": map[string]any{"rich_text": []any{map[string]any{"plain_text": "Goals"}}},
			},
			map[string]any{
				"type":      "paragraph",
				"paragraph": map[string]any{"rich_text": []any{map[string]any{"plain_text": "Ship the port."}}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig.Title != "Q2 Plan" {
		t.Fatalf("title: %q", sig.Title)
	}
	if !strings.Contains(sig.Content, "# Goals") || !strings.Contains(sig.Content, "Ship the port.") {
		t.Fatalf("blocks not flattened: %q", sig.Content)
	}
}

func TestGithubTransformPR(t *testing.T) {
	a, _ := connectors.Lookup("github")
	sig, err := a.Transform(map[string]any{
		"node_id":      "PR_abc",
		"title":        "Add connectors",
		"body":         "desc",
		"updated_at":   "2026-04-22T01:00:00Z",
		"pull_request": map[string]any{},
		"user":         map[string]any{"login": "rluna"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig.Genre != "pr" {
		t.Fatalf("genre should be pr, got %q", sig.Genre)
	}
	if !strings.HasPrefix(sig.Title, "[pr]") {
		t.Fatalf("title prefix missing: %q", sig.Title)
	}
	if sig.Entities[0] != "rluna" {
		t.Fatalf("entity: %+v", sig.Entities)
	}
}

func TestLinearTransform(t *testing.T) {
	a, _ := connectors.Lookup("linear")
	sig, err := a.Transform(map[string]any{
		"id":          "LN-1",
		"title":       "Port OptimalEngine",
		"description": "do it",
		"updatedAt":   "2026-04-22T10:00:00Z",
		"assignee":    map[string]any{"name": "Roberto"},
		"creator":     map[string]any{"name": "System"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig.Genre != "ticket" {
		t.Fatalf("genre: %q", sig.Genre)
	}
	wantNames := map[string]bool{"Roberto": true, "System": true}
	for _, e := range sig.Entities {
		delete(wantNames, e)
	}
	if len(wantNames) != 0 {
		t.Fatalf("missing entities: %+v (got %+v)", wantNames, sig.Entities)
	}
}

func TestRegistryListSorted(t *testing.T) {
	kinds := connectors.Kinds()
	for i := 1; i < len(kinds); i++ {
		if kinds[i-1] > kinds[i] {
			t.Fatalf("registry not sorted: %+v", kinds)
		}
	}
}
