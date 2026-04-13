package optimal

import (
	"strings"
	"testing"
)

func TestL0Abstract_emptyInput(t *testing.T) {
	out := L0Abstract("")
	if out != "" {
		t.Errorf("expected empty, got %q", out)
	}
}

func TestL0Abstract_stripsToTitle(t *testing.T) {
	content := "# My Great Title\n\nFirst paragraph of the document here."
	out := L0Abstract(content)
	if !strings.Contains(out, "My Great Title") {
		t.Errorf("expected title in output, got %q", out)
	}
	if !strings.Contains(out, "First paragraph") {
		t.Errorf("expected first paragraph in output, got %q", out)
	}
}

func TestL0Abstract_stripsFrontmatter(t *testing.T) {
	content := "---\ntitle: Frontmatter Title\ngenre: note\n---\n\n# Heading\n\nContent here."
	out := L0Abstract(content)
	if strings.Contains(out, "genre: note") {
		t.Errorf("frontmatter should be stripped, got %q", out)
	}
	if !strings.Contains(out, "Heading") {
		t.Errorf("expected heading in output, got %q", out)
	}
}

func TestL0Abstract_clampsAt400Chars(t *testing.T) {
	// Build content that is much longer than 400 chars.
	long := "# Title\n\n" + strings.Repeat("word ", 200)
	out := L0Abstract(long)
	if len(out) > 450 { // small buffer for the "..." suffix
		t.Errorf("output too long: %d chars", len(out))
	}
}

func TestL1Overview_emptyInput(t *testing.T) {
	out := L1Overview("")
	if out != "" {
		t.Errorf("expected empty, got %q", out)
	}
}

func TestL1Overview_includesHeadingsAndFirstParagraphs(t *testing.T) {
	content := `# Document

Intro line.

## Section One

First paragraph of section one.
Second line of the paragraph.

Third paragraph — should not appear.

## Section Two

First paragraph of section two.
`
	out := L1Overview(content)
	if !strings.Contains(out, "## Section One") {
		t.Errorf("expected Section One heading, got %q", out)
	}
	if !strings.Contains(out, "First paragraph of section one") {
		t.Errorf("expected section one first para, got %q", out)
	}
	// Third paragraph after a blank line should be excluded.
	if strings.Contains(out, "Third paragraph") {
		t.Errorf("third paragraph should be excluded, got %q", out)
	}
	if !strings.Contains(out, "## Section Two") {
		t.Errorf("expected Section Two heading, got %q", out)
	}
}

func TestL1Overview_clampsAt8000Chars(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		sb.WriteString("## Heading\n\n")
		sb.WriteString(strings.Repeat("word ", 100))
		sb.WriteString("\n\n")
	}
	out := L1Overview(sb.String())
	if len(out) > 8100 { // small buffer for "..." suffix
		t.Errorf("output too long: %d chars", len(out))
	}
}

func TestL2Full_returnsUnchanged(t *testing.T) {
	content := "# Title\n\nSome content."
	if got := L2Full(content); got != content {
		t.Errorf("L2Full modified content: got %q, want %q", got, content)
	}
}

func TestStripFrontmatter_nofrontmatter(t *testing.T) {
	content := "# Title\n\nContent."
	if got := stripFrontmatter(content); got != content {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestStripFrontmatter_withfrontmatter(t *testing.T) {
	content := "---\nkey: val\n---\n\n# Title"
	got := stripFrontmatter(content)
	if strings.Contains(got, "key: val") {
		t.Errorf("frontmatter not stripped: %q", got)
	}
	if !strings.Contains(got, "# Title") {
		t.Errorf("body missing: %q", got)
	}
}

func TestIsHeading(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"# Title", true},
		{"## Sub", true},
		{"### Deep", true},
		{"regular line", false},
		{"", false},
		{" # not a heading (leading space)", false},
	}
	for _, tc := range cases {
		if got := isHeading(tc.line); got != tc.want {
			t.Errorf("isHeading(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}
