package optimal

import (
	"fmt"
	"regexp"
	"strings"
)

// Compose reformats content for a specific receiver based on their genre
// competence. This implements the "genre-receiver alignment" principle from
// Signal Theory: match the genre to the receiver's decoding capability.
//
// Supported receiverGenre values:
//   - "brief"    → compress to objective + 3-5 bullet key messages + CTA
//   - "pitch"    → reframe as problem/solution/ask
//   - "standup"  → extract done/doing/blocked
//   - "spec"     → pass through (technical content needs no reframing)
//   - any other  → pass through unchanged
func Compose(content string, receiverGenre string) string {
	switch strings.ToLower(strings.TrimSpace(receiverGenre)) {
	case "brief":
		return composeBrief(content)
	case "pitch":
		return composePitch(content)
	case "standup":
		return composeStandup(content)
	case "spec":
		return content // technical content passes through unchanged
	default:
		return content
	}
}

// ── genre composers ───────────────────────────────────────────────────────────

// composeBrief compresses content into the brief skeleton:
// Objective | Key Messages (3-5 bullets) | Call to Action | Supporting Materials
func composeBrief(content string) string {
	sentences := extractSentences(content)
	if len(sentences) == 0 {
		return content
	}

	var sb strings.Builder
	sb.WriteString("## Brief\n\n")

	// Objective: first substantive sentence.
	sb.WriteString("**Objective:** ")
	sb.WriteString(sentences[0])
	sb.WriteString("\n\n")

	// Key Messages: up to 5 subsequent sentences as bullets.
	msgs := sentences[1:]
	if len(msgs) > 5 {
		msgs = msgs[:5]
	}
	if len(msgs) > 0 {
		sb.WriteString("**Key Messages:**\n")
		for _, m := range msgs {
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
		sb.WriteString("\n")
	}

	// Call to Action: extracted from imperative sentences, or a generic prompt.
	cta := extractCTA(content)
	sb.WriteString("**Call to Action:** ")
	sb.WriteString(cta)
	sb.WriteString("\n")

	return sb.String()
}

// composePitch reframes content as a pitch skeleton:
// Problem | Solution | Evidence | Ask
func composePitch(content string) string {
	sentences := extractSentences(content)
	if len(sentences) == 0 {
		return content
	}

	var sb strings.Builder
	sb.WriteString("## Pitch\n\n")

	// Problem: first sentence or two.
	end := 1
	if len(sentences) >= 2 {
		end = 2
	}
	sb.WriteString("**Problem:** ")
	sb.WriteString(strings.Join(sentences[:end], " "))
	sb.WriteString("\n\n")

	// Solution: next chunk.
	start := end
	end = start + 2
	if end > len(sentences) {
		end = len(sentences)
	}
	if start < len(sentences) {
		sb.WriteString("**Solution:** ")
		sb.WriteString(strings.Join(sentences[start:end], " "))
		sb.WriteString("\n\n")
	}

	// Evidence: remaining sentences.
	start = end
	if start < len(sentences) {
		sb.WriteString("**Evidence:**\n")
		for _, s := range sentences[start:] {
			sb.WriteString(fmt.Sprintf("- %s\n", s))
		}
		sb.WriteString("\n")
	}

	// Ask: extracted CTA.
	cta := extractCTA(content)
	sb.WriteString("**Ask:** ")
	sb.WriteString(cta)
	sb.WriteString("\n")

	return sb.String()
}

// composeStandup extracts done/doing/blocked structure.
func composeStandup(content string) string {
	var sb strings.Builder
	sb.WriteString("## Standup\n\n")

	done := extractSection(content, []string{"done", "completed", "finished", "shipped", "closed"})
	doing := extractSection(content, []string{"doing", "working on", "in progress", "currently", "today"})
	blocked := extractSection(content, []string{"blocked", "blocker", "waiting", "stuck", "need"})

	if done != "" {
		sb.WriteString("**Done:**\n")
		for _, line := range strings.Split(done, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", line))
			}
		}
		sb.WriteString("\n")
	}

	if doing != "" {
		sb.WriteString("**Doing:**\n")
		for _, line := range strings.Split(doing, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", line))
			}
		}
		sb.WriteString("\n")
	}

	if blocked != "" {
		sb.WriteString("**Blocked:**\n")
		for _, line := range strings.Split(blocked, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", line))
			}
		}
		sb.WriteString("\n")
	}

	// If we couldn't extract any structure, fall through to raw content.
	if done == "" && doing == "" && blocked == "" {
		sb.Reset()
		sb.WriteString("## Standup\n\n")
		sb.WriteString("**Done:** (not detected)\n\n")
		sb.WriteString("**Doing:** ")
		if sentences := extractSentences(content); len(sentences) > 0 {
			sb.WriteString(sentences[0])
		}
		sb.WriteString("\n\n**Blocked:** (none)\n")
	}

	return sb.String()
}

// ── helpers ───────────────────────────────────────────────────────────────────

var reSentenceSplit = regexp.MustCompile(`[.!?]+\s+`)
var reImperative = regexp.MustCompile(`(?i)^(please\s+)?(send|share|review|approve|confirm|schedule|reply|respond|complete|do|run|update|sign|click|join|book|call|email)\b`)

// extractSentences splits content into sentences, filtering short fragments.
func extractSentences(content string) []string {
	// Strip markdown headers and list markers first.
	content = regexp.MustCompile(`(?m)^#{1,6}\s+`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`(?m)^[-*]\s+`).ReplaceAllString(content, "")
	content = strings.TrimSpace(content)

	raw := reSentenceSplit.Split(content, -1)
	var sentences []string
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if len(s) >= 10 { // filter trivial fragments
			sentences = append(sentences, s)
		}
	}
	return sentences
}

// extractCTA finds the first imperative sentence in content, or returns a
// generic fallback.
func extractCTA(content string) string {
	for _, s := range extractSentences(content) {
		if reImperative.MatchString(s) {
			return s
		}
	}
	return "Review and respond."
}

// extractSection finds lines in content that follow a heading matching any of
// the keywords. Returns the associated bullet lines as a newline-joined string.
func extractSection(content string, keywords []string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inSection := false

	for _, line := range lines {
		lower := strings.ToLower(line)
		isHeader := strings.HasPrefix(strings.TrimSpace(line), "#") ||
			strings.HasSuffix(strings.TrimSpace(line), ":")

		if isHeader {
			inSection = false
			for _, kw := range keywords {
				if strings.Contains(lower, kw) {
					inSection = true
					break
				}
			}
			continue
		}

		if inSection {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				// Blank line ends the section.
				inSection = false
				continue
			}
			// Strip list markers.
			trimmed = strings.TrimLeft(trimmed, "-*• ")
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}

	return strings.Join(result, "\n")
}
