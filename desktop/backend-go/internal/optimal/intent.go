package optimal

import (
	"regexp"
	"strings"
)

// QueryIntent describes the parsed intent behind a search query.
type QueryIntent struct {
	Intent     string   `json:"intent"`     // lookup, comparison, temporal, decision, exploration
	Entities   []string `json:"entities"`   // extracted entity names
	TimeRange  string   `json:"time_range"` // e.g. "last week", "march", "" if not temporal
	Confidence float64  `json:"confidence"` // 0.0–1.0
}

// AnalyzeIntent determines the intent behind a search query using regex heuristics.
// No LLM is invoked — classification is purely rule-based.
func AnalyzeIntent(query string) QueryIntent {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)

	intent, confidence := classifyIntent(lower)
	timeRange := extractTimeRange(lower)
	entities := extractEntities(q)

	return QueryIntent{
		Intent:     intent,
		Entities:   entities,
		TimeRange:  timeRange,
		Confidence: confidence,
	}
}

// ── intent classification ─────────────────────────────────────────────────────

var (
	reLookup     = regexp.MustCompile(`\b(who|what is|what are|tell me about|find|show me|get)\b`)
	reComparison = regexp.MustCompile(`\b(compare|vs|versus|difference between|vs\.)\b`)
	reTemporal   = regexp.MustCompile(`\b(when|last|recent|this week|this month|yesterday|today|ago|since|before|after)\b`)
	reDecision   = regexp.MustCompile(`\b(should|decide|choose|pick|recommend|best option|what to)\b`)
	reExplore    = regexp.MustCompile(`\b(how|why|explain|understand|explore|overview|summary of)\b`)
)

func classifyIntent(lower string) (intent string, confidence float64) {
	type candidate struct {
		name       string
		re         *regexp.Regexp
		confidence float64
	}
	// Ordered by specificity — first match wins.
	candidates := []candidate{
		{"comparison", reComparison, 0.9},
		{"decision", reDecision, 0.85},
		{"temporal", reTemporal, 0.8},
		{"exploration", reExplore, 0.75},
		{"lookup", reLookup, 0.7},
	}

	for _, c := range candidates {
		if c.re.MatchString(lower) {
			return c.name, c.confidence
		}
	}
	// Default: treat as lookup with low confidence.
	return "lookup", 0.4
}

// ── time range extraction ─────────────────────────────────────────────────────

var reTimePhrase = regexp.MustCompile(
	`(last\s+\w+|this\s+week|this\s+month|this\s+year|yesterday|today|` +
		`(january|february|march|april|may|june|july|august|september|october|november|december)` +
		`(?:\s+\d{4})?|\d+\s+days?\s+ago|\d{4}-\d{2}-\d{2})`,
)

func extractTimeRange(lower string) string {
	m := reTimePhrase.FindString(lower)
	return strings.TrimSpace(m)
}

// ── entity extraction ─────────────────────────────────────────────────────────

var (
	// Quoted strings are high-confidence entities.
	reQuoted = regexp.MustCompile(`"([^"]+)"`)
	// Capitalized words (2+ chars, not at sentence start after punctuation).
	reCap = regexp.MustCompile(`\b([A-Z][a-z]{1,}(?:\s+[A-Z][a-z]{1,}){0,3})\b`)
	// Common stop words to exclude from capitalized hits.
	stopWords = map[string]struct{}{
		"The": {}, "A": {}, "An": {}, "In": {}, "On": {}, "At": {}, "By": {},
		"For": {}, "Of": {}, "To": {}, "And": {}, "Or": {}, "But": {},
		"When": {}, "What": {}, "Who": {}, "How": {}, "Why": {},
		"Show": {}, "Find": {}, "Get": {}, "Tell": {},
	}
)

func extractEntities(query string) []string {
	seen := make(map[string]struct{})
	var entities []string

	addEntity := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, dup := seen[s]; dup {
			return
		}
		seen[s] = struct{}{}
		entities = append(entities, s)
	}

	// Quoted strings first (highest confidence).
	for _, m := range reQuoted.FindAllStringSubmatch(query, -1) {
		if len(m) > 1 {
			addEntity(m[1])
		}
	}

	// Capitalized multi-word phrases.
	for _, m := range reCap.FindAllString(query, -1) {
		if _, stop := stopWords[m]; stop {
			continue
		}
		addEntity(m)
	}

	return entities
}
