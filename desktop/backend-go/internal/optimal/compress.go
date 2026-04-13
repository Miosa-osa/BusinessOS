package optimal

import (
	"regexp"
	"strings"
)

// CompressTranscript reduces a conversation transcript while preserving key
// information. It keeps first sentences, lines with decisions/actions/numbers,
// and removes filler/pleasantries. Output is capped at approximately maxTokens
// (~4 chars per token).
func CompressTranscript(transcript string, maxTokens int) string {
	if transcript == "" {
		return ""
	}
	if maxTokens <= 0 {
		maxTokens = 2000
	}

	lines := strings.Split(transcript, "\n")
	var kept []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if isFillerLine(line) {
			continue
		}
		if isSignalLine(line) {
			kept = append(kept, line)
			continue
		}
		// For non-signal lines keep only the first sentence.
		first := firstSentence(line)
		if first != "" {
			kept = append(kept, first)
		}
	}

	// Deduplicate adjacent identical lines.
	deduped := dedupLines(kept)

	// Join and cap at maxTokens (~4 chars each).
	out := strings.Join(deduped, "\n")
	maxChars := maxTokens * 4
	if len(out) > maxChars {
		out = out[:maxChars]
		// Snap to last newline.
		if idx := strings.LastIndexByte(out, '\n'); idx > maxChars/2 {
			out = out[:idx]
		}
		out += "\n..."
	}
	return strings.TrimSpace(out)
}

// ── line classifiers ──────────────────────────────────────────────────────────

var (
	reSignal = regexp.MustCompile(
		`(?i)(decided?|decision|action item|todo|will|must|need to|` +
			`deadline|due|urgent|revenue|\$\d|agreed|confirmed|` +
			`\b\d{4}-\d{2}-\d{2}\b|\b(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\b)`,
	)

	reFiller = regexp.MustCompile(
		`(?i)^(hi|hello|hey|thanks|thank you|ok|okay|sure|sounds good|` +
			`great|awesome|perfect|got it|understood|absolutely|` +
			`let me think|that'?s? a great|no problem|of course|` +
			`definitely|right|yeah|yep|nope|cool|alright)\b[.!,]?$`,
	)

	reSentenceEnd = regexp.MustCompile(`[.!?]+`)
)

// isSignalLine reports whether a line contains high-value signal content worth
// preserving in full (decisions, actions, dates, money amounts, names).
func isSignalLine(line string) bool {
	return reSignal.MatchString(line)
}

// isFillerLine reports whether a line is pure filler with no signal value.
func isFillerLine(line string) bool {
	return reFiller.MatchString(strings.TrimSpace(line))
}

// firstSentence returns the content up to (and including) the first sentence
// boundary. If no boundary is found, the full line is returned.
func firstSentence(line string) string {
	loc := reSentenceEnd.FindStringIndex(line)
	if loc == nil {
		return line
	}
	return strings.TrimSpace(line[:loc[1]])
}

// dedupLines removes consecutive duplicate lines.
func dedupLines(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	out := []string{lines[0]}
	for i := 1; i < len(lines); i++ {
		if lines[i] != lines[i-1] {
			out = append(out, lines[i])
		}
	}
	return out
}
