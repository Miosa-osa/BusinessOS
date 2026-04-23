package optimal

import (
	"strings"
	"testing"
)

// ── ParseFrontmatter ──────────────────────────────────────────────────────────

func TestParseFrontmatter_NestedSignal(t *testing.T) {
	content := "---\nsignal:\n  mode: linguistic\n  genre: transcript\n  type: inform\n  format: markdown\n  structure: transcript/skeleton\n  sn_ratio: 0.85\nnode: 04\ncross_ref: 11, 06\nvalid_from: \"2026-04-13\"\n---\n\n# Meeting Notes\nEd called about pricing."

	fm, body := ParseFrontmatter(content)

	if fm.Mode != "linguistic" {
		t.Errorf("Mode = %q, want %q", fm.Mode, "linguistic")
	}
	if fm.Genre != "transcript" {
		t.Errorf("Genre = %q, want %q", fm.Genre, "transcript")
	}
	if fm.Type != "inform" {
		t.Errorf("Type = %q, want %q", fm.Type, "inform")
	}
	if fm.Format != "markdown" {
		t.Errorf("Format = %q, want %q", fm.Format, "markdown")
	}
	if fm.SNRatio != 0.85 {
		t.Errorf("SNRatio = %.2f, want 0.85", fm.SNRatio)
	}
	if fm.Node != "04" {
		t.Errorf("Node = %q, want %q", fm.Node, "04")
	}

	// cross_ref must contain both "11" and "06"
	has11, has06 := false, false
	for _, cr := range fm.CrossRef {
		if cr == "11" {
			has11 = true
		}
		if cr == "06" {
			has06 = true
		}
	}
	if !has11 {
		t.Errorf("CrossRef missing %q: %v", "11", fm.CrossRef)
	}
	if !has06 {
		t.Errorf("CrossRef missing %q: %v", "06", fm.CrossRef)
	}

	if fm.ValidFrom != "2026-04-13" {
		t.Errorf("ValidFrom = %q, want %q", fm.ValidFrom, "2026-04-13")
	}
	// ParseFrontmatter may retain a leading newline before the body; trim it.
	trimmedBody := strings.TrimPrefix(body, "\n")
	if !strings.HasPrefix(trimmedBody, "# Meeting Notes") {
		t.Errorf("body should start with heading, got: %q", body[:min(40, len(body))])
	}
}

func TestParseFrontmatter_FlatFormat(t *testing.T) {
	content := "---\nmode: code\ngenre: spec\ntype: direct\nformat: markdown\nstructure: spec/goal-requirements-constraints-architecture-acceptance\nsn_ratio: 0.75\nnode: 02\nvalid_from: 2026-01-01\n---\n\nThe body."

	fm, body := ParseFrontmatter(content)

	if fm.Mode != "code" {
		t.Errorf("Mode = %q, want %q", fm.Mode, "code")
	}
	if fm.Genre != "spec" {
		t.Errorf("Genre = %q, want %q", fm.Genre, "spec")
	}
	if fm.Type != "direct" {
		t.Errorf("Type = %q, want %q", fm.Type, "direct")
	}
	if fm.SNRatio != 0.75 {
		t.Errorf("SNRatio = %.2f, want 0.75", fm.SNRatio)
	}
	if fm.Node != "02" {
		t.Errorf("Node = %q, want %q", fm.Node, "02")
	}
	// ParseFrontmatter may retain a leading newline; trim before comparing.
	if strings.TrimPrefix(body, "\n") != "The body." {
		t.Errorf("body = %q, want %q", body, "The body.")
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	content := "# Just a heading\n\nSome plain body text."

	fm, body := ParseFrontmatter(content)

	// All fields should be zero-values.
	if fm.Mode != "" || fm.Genre != "" || fm.Node != "" {
		t.Errorf("expected empty FrontmatterData for content without frontmatter, got: %+v", fm)
	}
	if body != content {
		t.Errorf("body should equal original content when no frontmatter present")
	}
}

func TestParseFrontmatter_EmptyContent(t *testing.T) {
	fm, body := ParseFrontmatter("")

	if fm.Mode != "" || fm.Genre != "" {
		t.Errorf("expected empty FrontmatterData for empty input, got: %+v", fm)
	}
	if body != "" {
		t.Errorf("body = %q, want empty", body)
	}
}

func TestParseFrontmatter_MissingClosingDelimiter(t *testing.T) {
	// Opening --- but no closing --- → not valid frontmatter, return content as-is.
	content := "---\ngenre: note\n\nNo closing delimiter here."

	_, body := ParseFrontmatter(content)

	if body != content {
		t.Errorf("body should equal full content when closing delimiter is absent")
	}
}

func TestParseFrontmatter_InlineBracketCrossRef(t *testing.T) {
	content := "---\ncross_ref: [11, 06, 03]\n---\n\nBody."

	fm, _ := ParseFrontmatter(content)

	if len(fm.CrossRef) != 3 {
		t.Errorf("CrossRef len = %d, want 3; got %v", len(fm.CrossRef), fm.CrossRef)
	}
}

func TestParseFrontmatter_ListItemsCrossRef(t *testing.T) {
	content := "---\ncross_ref:\n  - 11\n  - 06\n---\n\nBody."

	fm, _ := ParseFrontmatter(content)

	if len(fm.CrossRef) != 2 {
		t.Errorf("CrossRef len = %d, want 2; got %v", len(fm.CrossRef), fm.CrossRef)
	}
}

func TestParseFrontmatter_QuotedValues(t *testing.T) {
	content := "---\ngenre: \"note\"\nnode: '09'\n---\n\nBody."

	fm, _ := ParseFrontmatter(content)

	if fm.Genre != "note" {
		t.Errorf("Genre = %q, want %q (quotes should be stripped)", fm.Genre, "note")
	}
	if fm.Node != "09" {
		t.Errorf("Node = %q, want %q (quotes should be stripped)", fm.Node, "09")
	}
}

// ── HasSignalDims ─────────────────────────────────────────────────────────────

func TestHasSignalDims_TrueWhenAnyFieldSet(t *testing.T) {
	cases := []struct {
		name string
		fd   FrontmatterData
		want bool
	}{
		{"all empty → false", FrontmatterData{}, false},
		{"mode set → true", FrontmatterData{Mode: "linguistic"}, true},
		{"genre set → true", FrontmatterData{Genre: "note"}, true},
		{"type set → true", FrontmatterData{Type: "inform"}, true},
		{"format set → true", FrontmatterData{Format: "markdown"}, true},
		{"structure set → true", FrontmatterData{Structure: "note/context-content-route"}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.fd.HasSignalDims(); got != tc.want {
				t.Errorf("HasSignalDims() = %v, want %v", got, tc.want)
			}
		})
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func TestSplitKV_Valid(t *testing.T) {
	cases := []struct {
		line    string
		wantKey string
		wantVal string
		wantOK  bool
	}{
		{"genre: note", "genre", "note", true},
		{"sn_ratio: 0.75", "sn_ratio", "0.75", true},
		{"signal:", "signal", "", true},
		{"  key : value  ", "key", "value", true},
		{"no colon here", "", "", false},
	}

	for _, tc := range cases {
		k, v, ok := splitKV(tc.line)
		if ok != tc.wantOK {
			t.Errorf("splitKV(%q) ok = %v, want %v", tc.line, ok, tc.wantOK)
			continue
		}
		if ok {
			if k != tc.wantKey {
				t.Errorf("splitKV(%q) key = %q, want %q", tc.line, k, tc.wantKey)
			}
			if v != tc.wantVal {
				t.Errorf("splitKV(%q) val = %q, want %q", tc.line, v, tc.wantVal)
			}
		}
	}
}

func TestUnquote(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{"plain", "plain"},
		{`""`, ""},
		{"single char no quotes: a", "single char no quotes: a"},
	}

	for _, tc := range cases {
		got := unquote(tc.input)
		if got != tc.want {
			t.Errorf("unquote(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestParseInlineArray(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"11, 06", []string{"11", "06"}},
		{"[11, 06, 03]", []string{"11", "06", "03"}},
		{"single", []string{"single"}},
		{"", nil},
		{`["a", "b"]`, []string{"a", "b"}},
	}

	for _, tc := range cases {
		got := parseInlineArray(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("parseInlineArray(%q) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Errorf("parseInlineArray(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

// min is a local helper so the test file compiles on Go <1.21.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
