// Package compliance ports OptimalEngine.Compliance (Elixir) to Go.
//
// Phase 1 here covers PII detection + redaction — the building blocks every
// other compliance workflow (DSAR, erasure, legal hold, retention) layers on
// top of. Those higher-level workflows depend on tenancy + audit which are
// scheduled for a subsequent port phase.
package compliance

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strconv"
)

// Kind is the type of PII a match represents.
type Kind string

const (
	KindEmail      Kind = "email"
	KindPhone      Kind = "phone"
	KindSSN        Kind = "ssn"
	KindCreditCard Kind = "credit_card"
	KindIPv4       Kind = "ipv4"
	KindURL        Kind = "url"
)

// Match is one PII occurrence in a document.
type Match struct {
	Kind   Kind
	Value  string
	Offset int // byte offset — matches Elixir semantics
	Length int
}

// Regexes: Go's RE2 lacks Perl lookaround, so the Elixir patterns with
// `(?!...)` negative lookahead for SSN are rewritten as a post-filter.
var (
	reEmail = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	rePhone = regexp.MustCompile(
		`(?:\+?1[\s.\-]?)?(?:\(?\d{3}\)?[\s.\-]?)\d{3}[\s.\-]?\d{4}\b|\+[1-9]\d{1,14}\b`,
	)
	reSSN  = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	reIPv4 = regexp.MustCompile(
		`\b(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.){3}(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\b`,
	)
	reURL  = regexp.MustCompile(`https?://[^\s<>"']+`)
	reCard = regexp.MustCompile(`\b(?:\d[ \-]?){13,19}\b`)
)

// Scan returns every PII match in text, sorted by offset.
func Scan(text string) []Match {
	var out []Match
	out = append(out, scanRegex(text, reEmail, KindEmail)...)
	out = append(out, scanRegex(text, rePhone, KindPhone)...)
	out = append(out, scanSSN(text)...)
	out = append(out, scanRegex(text, reIPv4, KindIPv4)...)
	out = append(out, scanRegex(text, reURL, KindURL)...)
	out = append(out, scanCreditCards(text)...)
	sort.Slice(out, func(i, j int) bool { return out[i].Offset < out[j].Offset })
	return dedupe(out)
}

// Any reports whether text contains any PII.
func Any(text string) bool { return len(Scan(text)) > 0 }

// KindsPresent returns the distinct set of PII kinds found in text.
func KindsPresent(text string) []Kind {
	seen := map[Kind]bool{}
	for _, m := range Scan(text) {
		seen[m.Kind] = true
	}
	out := make([]Kind, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func scanRegex(text string, re *regexp.Regexp, k Kind) []Match {
	locs := re.FindAllStringIndex(text, -1)
	out := make([]Match, 0, len(locs))
	for _, loc := range locs {
		out = append(out, Match{
			Kind:   k,
			Value:  text[loc[0]:loc[1]],
			Offset: loc[0],
			Length: loc[1] - loc[0],
		})
	}
	return out
}

// scanSSN applies the Elixir post-filter: reject SSNs whose area code is
// 000, 666, or starts with 9; group 00; serial 0000.
func scanSSN(text string) []Match {
	raw := scanRegex(text, reSSN, KindSSN)
	out := raw[:0]
	for _, m := range raw {
		area := m.Value[0:3]
		group := m.Value[4:6]
		serial := m.Value[7:11]
		if area == "000" || area == "666" || area[0] == '9' {
			continue
		}
		if group == "00" || serial == "0000" {
			continue
		}
		out = append(out, m)
	}
	return out
}

func scanCreditCards(text string) []Match {
	locs := reCard.FindAllStringIndex(text, -1)
	out := make([]Match, 0, len(locs))
	for _, loc := range locs {
		candidate := text[loc[0]:loc[1]]
		digits := stripNonDigits(candidate)
		if len(digits) < 13 {
			continue
		}
		if !luhnValid(digits) {
			continue
		}
		out = append(out, Match{
			Kind:   KindCreditCard,
			Value:  candidate,
			Offset: loc[0],
			Length: loc[1] - loc[0],
		})
	}
	return out
}

func stripNonDigits(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b = append(b, s[i])
		}
	}
	return string(b)
}

// Luhn: duplicate every second digit from the right; sum mod 10 == 0.
func luhnValid(digits string) bool {
	sum := 0
	for i, r := range reverseDigits(digits) {
		n, err := strconv.Atoi(string(r))
		if err != nil {
			return false
		}
		if i%2 == 0 {
			sum += n
		} else {
			d := n * 2
			if d > 9 {
				d -= 9
			}
			sum += d
		}
	}
	return sum%10 == 0
}

func reverseDigits(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

func dedupe(m []Match) []Match {
	if len(m) < 2 {
		return m
	}
	out := m[:1]
	for i := 1; i < len(m); i++ {
		prev := out[len(out)-1]
		if m[i].Offset == prev.Offset && m[i].Length == prev.Length {
			// Prefer more-specific kind (credit_card > phone when both match).
			if preferKind(m[i].Kind, prev.Kind) {
				out[len(out)-1] = m[i]
			}
			continue
		}
		out = append(out, m[i])
	}
	return out
}

// preferKind decides which of two overlapping matches to keep.
func preferKind(a, b Kind) bool {
	rank := map[Kind]int{
		KindCreditCard: 5,
		KindSSN:        4,
		KindEmail:      3,
		KindURL:        2,
		KindPhone:      1,
		KindIPv4:       1,
	}
	return rank[a] > rank[b]
}

// Redactable utilities used by redact.go. Exported for adapters that want
// to implement hash-based placeholders on their own.

// Fingerprint returns the first 6 hex chars of sha256(value). Used for the
// :hash redaction strategy.
func Fingerprint(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:6]
}
