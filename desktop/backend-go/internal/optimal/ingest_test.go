package optimal

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestMakeSlug_basic(t *testing.T) {
	cases := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"Ed called about pricing", 50, "ed-called-about-pricing"},
		{"Hello World!", 50, "hello-world"},
		{"  leading spaces  ", 50, "leading-spaces"},
		{"", 50, "signal"},
		{"UPPERCASE TEXT", 50, "uppercase-text"},
	}
	for _, tc := range cases {
		got := makeSlug(tc.input, tc.maxLen)
		if got != tc.want {
			t.Errorf("makeSlug(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
		}
	}
}

func TestMakeSlug_truncatesAtWordBoundary(t *testing.T) {
	input := "this is a fairly long sentence that should be truncated nicely"
	got := makeSlug(input, 20)
	if len(got) > 20 {
		t.Errorf("slug too long: %q (%d chars)", got, len(got))
	}
	if strings.HasSuffix(got, "-") {
		t.Errorf("slug ends with hyphen: %q", got)
	}
}

func TestBuildSignalContent_containsExpectedFields(t *testing.T) {
	ts := mustParseDate("2026-04-11")
	content := buildSignalContent("note", ts, "some text here")

	checks := []string{
		"genre: note",
		"date: 2026-04-11",
		"mode: linguistic",
		"type: inform",
		"some text here",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("content missing %q:\n%s", check, content)
		}
	}
}

func TestIngest_emptyTextReturnsError(t *testing.T) {
	err := Ingest("", "", "", "", "note")
	if err == nil {
		t.Error("expected error for empty text, got nil")
	}
}

func TestIngest_writesToFilesystem(t *testing.T) {
	tmpRoot := t.TempDir()
	nodesRoot := tmpRoot + "/nodes"

	err := Ingest(tmpRoot, nodesRoot, "", "Ed called about pricing", "note")
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	entries := readDirSafe(nodesRoot + "/09-new-stuff/signals")
	if len(entries) == 0 {
		t.Fatal("expected signal file in 09-new-stuff/signals, found none")
	}
}

func TestIngest_defaultsGenreToNote(t *testing.T) {
	tmpRoot := t.TempDir()
	nodesRoot := tmpRoot + "/nodes"

	// Empty genre should default to "note" and not error.
	err := Ingest(tmpRoot, nodesRoot, "", "Some signal text", "")
	if err != nil {
		t.Fatalf("Ingest with empty genre failed: %v", err)
	}

	entries := readDirSafe(nodesRoot + "/09-new-stuff/signals")
	if len(entries) == 0 {
		t.Fatal("expected signal file written even with empty genre")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func readDirSafe(path string) []os.DirEntry {
	entries, _ := os.ReadDir(path)
	return entries
}
