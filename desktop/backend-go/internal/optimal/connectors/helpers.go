package connectors

import (
	"regexp"
	"strings"
	"time"
)

// Helper functions shared by adapter Transform implementations. These mirror
// the public exports from OptimalEngine.Connectors.Transform in Elixir.

var (
	htmlTagRe     = regexp.MustCompile(`<[^>]+>`)
	whitespaceRe  = regexp.MustCompile(`\s+`)
	emailHeaderRe = regexp.MustCompile(`[\w.+\-]+@[\w.\-]+\.[A-Za-z]{2,}`)
)

// StripHTML removes HTML tags and collapses whitespace. Good-enough for
// body-stripping email snippets / issue descriptions. Not an HTML parser.
func StripHTML(s string) string {
	out := htmlTagRe.ReplaceAllString(s, " ")
	out = whitespaceRe.ReplaceAllString(out, " ")
	return strings.TrimSpace(out)
}

// ParseISO8601 parses common timestamp shapes emitted by external APIs:
// RFC3339 (most), RFC3339Nano, and Gmail's millisecond-epoch string.
// On failure it returns time.Now().UTC() — adapters need *some* modified_at.
func ParseISO8601(s string) time.Time {
	if s == "" {
		return time.Now().UTC()
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	// Gmail returns `internalDate` as a millisecond epoch string.
	if len(s) > 0 && allDigits(s) {
		var ms int64
		for _, r := range s {
			ms = ms*10 + int64(r-'0')
		}
		return time.UnixMilli(ms).UTC()
	}
	return time.Now().UTC()
}

// ExtractEmails pulls every RFC-ish email address out of a free-form header
// value such as Gmail's "From: Alice <alice@example.com>".
func ExtractEmails(s string) []string {
	return emailHeaderRe.FindAllString(s, -1)
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
