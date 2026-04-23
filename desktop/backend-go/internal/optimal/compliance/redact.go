package compliance

import (
	"fmt"
	"sort"
	"strings"
)

// Strategy picks how matches get replaced.
type Strategy string

const (
	StrategyPlaceholder Strategy = "placeholder" // "<REDACTED:email>"   (default)
	StrategyMask        Strategy = "mask"        // character-by-character "*"
	StrategyHash        Strategy = "hash"        // "<REDACTED:email:7a3c9f>"
	StrategyRemove      Strategy = "remove"      // drop the matched bytes
)

// RedactOpts tune a Redact call.
type RedactOpts struct {
	Strategy Strategy
	Only     []Kind // empty = all
	Except   []Kind
}

// Report bundles the rewritten text with the matches that drove the changes.
type Report struct {
	Redacted string
	Matches  []Match
	Strategy Strategy
}

// Redact produces a redacted copy of text plus a report of every match.
// Default strategy is StrategyPlaceholder — the safest for downstream LLMs
// that must not memorize PII.
func Redact(text string, opts RedactOpts) Report {
	if opts.Strategy == "" {
		opts.Strategy = StrategyPlaceholder
	}
	all := Scan(text)
	kept := all[:0]
	for _, m := range all {
		if included(m.Kind, opts.Only, opts.Except) {
			kept = append(kept, m)
		}
	}
	redacted := applyRedactions(text, kept, opts.Strategy)
	return Report{Redacted: redacted, Matches: kept, Strategy: opts.Strategy}
}

// RedactString is the shorthand when only the output text is wanted.
func RedactString(text string, opts RedactOpts) string { return Redact(text, opts).Redacted }

func included(k Kind, only, except []Kind) bool {
	for _, e := range except {
		if e == k {
			return false
		}
	}
	if len(only) == 0 {
		return true
	}
	for _, o := range only {
		if o == k {
			return true
		}
	}
	return false
}

// applyRedactions walks matches right-to-left so offsets remain valid after
// each splice — identical algorithm to the Elixir version.
func applyRedactions(text string, matches []Match, strategy Strategy) string {
	if len(matches) == 0 {
		return text
	}
	sorted := make([]Match, len(matches))
	copy(sorted, matches)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Offset > sorted[j].Offset })

	out := text
	for _, m := range sorted {
		out = splice(out, m, strategy)
	}
	return out
}

func splice(text string, m Match, strategy Strategy) string {
	if m.Offset+m.Length > len(text) {
		return text
	}
	before := text[:m.Offset]
	after := text[m.Offset+m.Length:]
	return before + replacement(m, strategy) + after
}

func replacement(m Match, strategy Strategy) string {
	switch strategy {
	case StrategyMask:
		return strings.Repeat("*", m.Length)
	case StrategyHash:
		return fmt.Sprintf("<REDACTED:%s:%s>", m.Kind, Fingerprint(m.Value))
	case StrategyRemove:
		return ""
	default:
		return fmt.Sprintf("<REDACTED:%s>", m.Kind)
	}
}
