package optimal

import (
	"regexp"
	"strings"
	"time"
)

// FailureMode represents a detected signal quality issue from Signal Theory.
type FailureMode struct {
	Code     string `json:"code"`     // e.g. "routing_failure", "genre_mismatch"
	Category string `json:"category"` // "shannon", "ashby", "beer", "wiener", "cross_cutting"
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error", "warning", "info"
}

// DetectFailureModes runs all 11 Signal Theory failure mode checks against
// a classified signal and its content. Returns detected failures (empty = clean).
func DetectFailureModes(cl SignalClassification, content string, topo *TopologyConfig) []FailureMode {
	var modes []FailureMode

	// === SHANNON VIOLATIONS (channel capacity) ===

	// 1. Routing failure — signal has no clear destination.
	// Triggered when genre defaulted to "note" AND no entities are extractable.
	if cl.Genre == "note" {
		entities := extractEntities(content)
		if len(entities) == 0 {
			modes = append(modes, FailureMode{
				Code:     "routing_failure",
				Category: "shannon",
				Message:  "Signal has no clear destination: genre defaulted to 'note' and no named entities were found.",
				Severity: "warning",
			})
		}
	}

	// 2. Bandwidth overload — content exceeds reasonable word count for its genre.
	wc := countWords(content)
	genreLimits := map[string]int{
		"standup": 500,
		"brief":   300,
		"note":    2000,
	}
	if limit, ok := genreLimits[cl.Genre]; ok && wc > limit {
		modes = append(modes, FailureMode{
			Code:     "bandwidth_overload",
			Category: "shannon",
			Message:  "Content exceeds bandwidth for genre '" + cl.Genre + "': " + wordCountStr(wc) + " words (limit " + wordCountStr(limit) + ").",
			Severity: "warning",
		})
	}

	// 3. Fidelity failure — S/N ratio too low, meaning likely lost.
	if cl.SNRatio < 0.4 && content != "" {
		modes = append(modes, FailureMode{
			Code:     "fidelity_failure",
			Category: "shannon",
			Message:  "Low signal-to-noise ratio: filler density or lack of structure reduces decodability.",
			Severity: "warning",
		})
	}

	// === ASHBY VIOLATIONS (requisite variety) ===

	// 4. Genre mismatch — detected genre is "spec" but no structural spec markers present.
	if cl.Genre == "spec" && !reSpecStructure.MatchString(content) {
		modes = append(modes, FailureMode{
			Code:     "genre_mismatch",
			Category: "ashby",
			Message:  "Genre classified as 'spec' but content lacks required structural sections (requirements, constraints, acceptance criteria).",
			Severity: "error",
		})
	}

	// 5. Variety failure — content has strong structural markers but was classified as "note".
	if cl.Genre == "note" && reStrongStructure.MatchString(content) && wc > 100 {
		modes = append(modes, FailureMode{
			Code:     "variety_failure",
			Category: "ashby",
			Message:  "Content appears structured but was classified as 'note' — possible genre misclassification.",
			Severity: "info",
		})
	}

	// 6. Structure failure — genre has a defined skeleton but content has no section headers.
	structuredGenres := map[string]bool{
		"transcript": true, "brief": true, "spec": true,
		"plan": true, "decision-log": true, "review": true,
		"report": true, "pitch": true,
	}
	if structuredGenres[cl.Genre] && !reMarkdownSection.MatchString(content) && wc > 50 {
		modes = append(modes, FailureMode{
			Code:     "structure_failure",
			Category: "ashby",
			Message:  "Genre '" + cl.Genre + "' requires structural sections, but none were detected in content.",
			Severity: "warning",
		})
	}

	// === BEER VIOLATIONS (viable structure) ===

	// 7. Bridge failure — content uses unexplained acronyms/jargon.
	acronyms := reAcronym.FindAllString(content, -1)
	if len(acronyms) >= 3 {
		// Check that acronyms are not defined inline (e.g. "RRF (Reciprocal Rank Fusion)")
		defined := reAcronymDefinition.FindAllString(content, -1)
		if len(defined) < len(acronyms)/2 {
			modes = append(modes, FailureMode{
				Code:     "bridge_failure",
				Category: "beer",
				Message:  "Content contains unexplained acronyms/jargon without shared context definitions.",
				Severity: "info",
			})
		}
	}

	// 8. Herniation failure — abstract (first sentence/heading) has low keyword overlap with full content.
	if wc > 80 {
		abstract := extractAbstract(content)
		if jaccardOverlap(abstract, content) < 0.15 {
			modes = append(modes, FailureMode{
				Code:     "herniation_failure",
				Category: "beer",
				Message:  "Abstract/opening and full content have low keyword overlap — signal may be incoherent across abstraction layers.",
				Severity: "warning",
			})
		}
	}

	// 9. Decay failure — content references dates older than 90 days without context.
	if hasStaleDate(content) {
		modes = append(modes, FailureMode{
			Code:     "decay_failure",
			Category: "beer",
			Message:  "Content references dates older than 90 days — signal may contain outdated information.",
			Severity: "info",
		})
	}

	// === WIENER VIOLATIONS (closed loops) ===

	// 10. Feedback failure — action items present but no assignee detected.
	if reActionItemMarker.MatchString(content) && !reAssignee.MatchString(content) {
		modes = append(modes, FailureMode{
			Code:     "feedback_failure",
			Category: "wiener",
			Message:  "Action items detected without an assignee or deadline — feedback loop is open.",
			Severity: "warning",
		})
	}

	// === CROSS-CUTTING ===

	// 11. Adversarial noise — extreme repetition or nonsense character sequences.
	if wc > 10 && isAdversarialNoise(content) {
		modes = append(modes, FailureMode{
			Code:     "adversarial_noise",
			Category: "cross_cutting",
			Message:  "Content exhibits abnormally high repetition or nonsense character density — possible spam or injection attempt.",
			Severity: "error",
		})
	}

	return modes
}

// ── compiled patterns ──────────────────────────────────────────────────────────

var (
	// Check 4: spec must have real structural sections.
	reSpecStructure = regexp.MustCompile(`(?i)(requirements?[:\s]|constraints?[:\s]|acceptance criteria|must have|should have)`)

	// Check 5: strong structural markers (multiple headers or numbered lists).
	reStrongStructure = regexp.MustCompile(`(?m)(^#{1,6}\s+\S.+\n|^\d+\.\s+\S)`)

	// Check 6: any markdown section header.
	reMarkdownSection = regexp.MustCompile(`(?m)^#{1,6}\s+\S`)

	// Check 7: acronyms (2–5 consecutive uppercase letters).
	reAcronym = regexp.MustCompile(`\b[A-Z]{2,5}\b`)
	// Inline definition pattern: "ACRONYM (Full Name)" or "Full Name (ACRONYM)".
	reAcronymDefinition = regexp.MustCompile(`\b[A-Z]{2,5}\s*\([^)]{3,}\)|\([A-Z]{2,5}\)`)

	// Check 9: ISO dates (YYYY-MM-DD) and written month+year patterns.
	reDatePattern = regexp.MustCompile(
		`\b(\d{4}-\d{2}-\d{2}|` +
			`(january|february|march|april|may|june|july|august|september|october|november|december)\s+\d{4})\b`,
	)

	// Check 10: action item markers.
	reActionItemMarker = regexp.MustCompile(`(?i)(todo[:\s]|\[ \]|action items?[:\s]|next steps?[:\s])`)
	// Assignee: @mention or "Name:" owner pattern before a task.
	reAssignee = regexp.MustCompile(`(?i)(@\w+|assigned to|owner[:\s]|\b[A-Z][a-z]+ will\b)`)
)

// ── helpers ───────────────────────────────────────────────────────────────────

// wordCountStr converts an int to a string without importing strconv at package level.
func wordCountStr(n int) string {
	if n == 0 {
		return "0"
	}
	// Simple itoa-style conversion for positive integers.
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

// extractAbstract returns the first meaningful sentence or heading from content.
func extractAbstract(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) >= 10 {
			// Strip markdown markers.
			line = strings.TrimLeft(line, "#- *")
			line = strings.TrimSpace(line)
			if len(line) >= 10 {
				return line
			}
		}
	}
	return content
}

// jaccardOverlap computes the Jaccard index between the word-sets of two strings.
// Returns a value in [0, 1].
func jaccardOverlap(a, b string) float64 {
	setA := tokenSet(a)
	setB := tokenSet(b)
	if len(setA) == 0 && len(setB) == 0 {
		return 1.0
	}
	var intersection, union int
	for tok := range setA {
		if _, ok := setB[tok]; ok {
			intersection++
		}
	}
	union = len(setA) + len(setB) - intersection
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

// tokenSet lowercases and splits text into a set of word tokens (≥3 chars).
func tokenSet(text string) map[string]struct{} {
	words := strings.Fields(strings.ToLower(text))
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		// Strip punctuation edges.
		w = strings.Trim(w, ".,!?:;\"'()[]{}")
		if len(w) >= 3 {
			set[w] = struct{}{}
		}
	}
	return set
}

// hasStaleDate returns true if the content contains any date reference older
// than 90 days from now.
func hasStaleDate(content string) bool {
	matches := reDatePattern.FindAllString(strings.ToLower(content), -1)
	cutoff := time.Now().AddDate(0, 0, -90)

	for _, m := range matches {
		m = strings.TrimSpace(m)
		// Try ISO format first.
		for _, layout := range []string{
			"2006-01-02",
			"january 2006", "february 2006", "march 2006", "april 2006",
			"may 2006", "june 2006", "july 2006", "august 2006",
			"september 2006", "october 2006", "november 2006", "december 2006",
		} {
			if t, err := time.Parse(layout, m); err == nil {
				if t.Before(cutoff) {
					return true
				}
				break
			}
		}
	}
	return false
}

// isAdversarialNoise detects extreme word repetition (any single word > 25% of
// total word count) or a high ratio of non-printable/non-ASCII characters.
func isAdversarialNoise(content string) bool {
	words := strings.Fields(strings.ToLower(content))
	total := len(words)
	if total < 10 {
		return false
	}

	// Repetition check: single word dominates.
	freq := make(map[string]int, total)
	for _, w := range words {
		freq[w]++
	}
	for _, count := range freq {
		if float64(count)/float64(total) > 0.25 {
			return true
		}
	}

	// Non-ASCII density check: > 30% non-printable/non-ASCII runes.
	runes := []rune(content)
	var nonASCII int
	for _, r := range runes {
		if r > 127 || (r < 32 && r != '\n' && r != '\t' && r != '\r') {
			nonASCII++
		}
	}
	if len(runes) > 0 && float64(nonASCII)/float64(len(runes)) > 0.30 {
		return true
	}

	return false
}
