package adapters

import (
	"strings"
	"testing"

	"github.com/rhl/businessos-backend/internal/optimal/connectors"
)

func TestConfluenceTransform(t *testing.T) {
	a, _ := connectors.Lookup("confluence")
	sig, err := a.Transform(map[string]any{
		"id":    "page-1",
		"title": "Architecture",
		"body": map[string]any{
			"storage": map[string]any{"value": "<p>Hello <b>world</b></p>"},
		},
		"version": map[string]any{"when": "2026-04-22T10:00:00Z"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig.Genre != "document" || sig.Title != "Architecture" {
		t.Fatalf("unexpected: %+v", sig)
	}
	if !strings.Contains(sig.Content, "Hello world") {
		t.Fatalf("html not stripped: %q", sig.Content)
	}
}

func TestDriveGenreFromMime(t *testing.T) {
	a, _ := connectors.Lookup("drive")
	sig, err := a.Transform(map[string]any{
		"id":           "f-1",
		"name":         "Sheet",
		"mimeType":     "application/vnd.google-apps.spreadsheet",
		"modifiedTime": "2026-04-22T10:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig.Genre != "table" {
		t.Fatalf("genre: %q", sig.Genre)
	}
	if sig.Metadata["mode"] != "data" {
		t.Fatalf("mode: %v", sig.Metadata["mode"])
	}
}

func TestDropboxFallsBackToPathLower(t *testing.T) {
	a, _ := connectors.Lookup("dropbox")
	sig, _ := a.Transform(map[string]any{
		"path_lower":      "/folder/file.txt",
		"path_display":    "/Folder/File.txt",
		"server_modified": "2026-04-22T10:00:00Z",
	})
	if !strings.Contains(sig.ID, "/folder/file.txt") {
		t.Fatalf("ID should include path_lower fallback: %q", sig.ID)
	}
	if sig.Title != "File.txt" {
		t.Fatalf("title from basename: %q", sig.Title)
	}
}

func TestHubspotPicksFirstAvailableTitle(t *testing.T) {
	a, _ := connectors.Lookup("hubspot")
	sig, _ := a.Transform(map[string]any{
		"id": 42,
		"properties": map[string]any{
			"firstname": "Alice",
			// no dealname → should pick firstname
		},
	})
	if sig.Title != "Alice" {
		t.Fatalf("title: %q", sig.Title)
	}
	if sig.Genre != "crm" {
		t.Fatalf("genre: %q", sig.Genre)
	}
}

func TestJiraTitleIncludesKey(t *testing.T) {
	a, _ := connectors.Lookup("jira")
	sig, _ := a.Transform(map[string]any{
		"key": "PROJ-123",
		"fields": map[string]any{
			"summary":     "Ship it",
			"description": "<p>Body</p>",
			"updated":     "2026-04-22T10:00:00Z",
			"assignee":    map[string]any{"displayName": "Alice"},
			"reporter":    map[string]any{"displayName": "Bob"},
		},
	})
	if sig.Title != "[PROJ-123] Ship it" {
		t.Fatalf("title: %q", sig.Title)
	}
	if len(sig.Entities) != 2 {
		t.Fatalf("entities: %+v", sig.Entities)
	}
	if sig.Content != "Body" {
		t.Fatalf("html not stripped: %q", sig.Content)
	}
}

func TestOnedriveTransform(t *testing.T) {
	a, _ := connectors.Lookup("onedrive")
	sig, _ := a.Transform(map[string]any{
		"id":                   "od-1",
		"name":                 "Spec.docx",
		"lastModifiedDateTime": "2026-04-22T10:00:00Z",
	})
	if sig.Title != "Spec.docx" {
		t.Fatalf("title: %q", sig.Title)
	}
	if sig.Genre != "file" {
		t.Fatalf("genre: %q", sig.Genre)
	}
}

func TestSalesforceGenreFromAttributesType(t *testing.T) {
	a, _ := connectors.Lookup("salesforce")
	sig, _ := a.Transform(map[string]any{
		"Id":               "001abc",
		"Name":             "Acme Corp",
		"Description":      "big customer",
		"LastModifiedDate": "2026-04-22T10:00:00Z",
		"attributes":       map[string]any{"type": "Account"},
	})
	if sig.Genre != "account" {
		t.Fatalf("genre should be lower-cased sObject: %q", sig.Genre)
	}
	if sig.Title != "Acme Corp" {
		t.Fatalf("title: %q", sig.Title)
	}
}

func TestTeamsStripsHtmlForTitleAndContent(t *testing.T) {
	a, _ := connectors.Lookup("teams")
	sig, _ := a.Transform(map[string]any{
		"id":                   "msg-1",
		"body":                 map[string]any{"content": "<p>quick chat <b>ping</b></p>"},
		"lastModifiedDateTime": "2026-04-22T10:00:00Z",
		"from": map[string]any{
			"user": map[string]any{"displayName": "Alice"},
		},
	})
	if strings.Contains(sig.Title, "<") || strings.Contains(sig.Content, "<") {
		t.Fatalf("html leaked: title=%q content=%q", sig.Title, sig.Content)
	}
	if len(sig.Entities) != 1 || sig.Entities[0] != "Alice" {
		t.Fatalf("entities: %+v", sig.Entities)
	}
}

func TestZoomFallsBackToTopicTitle(t *testing.T) {
	a, _ := connectors.Lookup("zoom")
	sig, _ := a.Transform(map[string]any{
		"uuid":       "abc==",
		"topic":      "Weekly sync",
		"transcript": "Alice: hello.",
		"start_time": "2026-04-22T10:00:00Z",
	})
	if sig.Title != "Weekly sync" {
		t.Fatalf("title: %q", sig.Title)
	}
	if sig.Genre != "transcript" {
		t.Fatalf("genre: %q", sig.Genre)
	}
	if !strings.HasPrefix(sig.ID, "zoom/") {
		t.Fatalf("id: %q", sig.ID)
	}
}

func TestZoomDefaultTitleWhenTopicMissing(t *testing.T) {
	a, _ := connectors.Lookup("zoom")
	sig, _ := a.Transform(map[string]any{"uuid": "abc"})
	if sig.Title != "Zoom meeting" {
		t.Fatalf("title fallback: %q", sig.Title)
	}
}
