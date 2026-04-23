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
//   - "brief"        → compress to objective + 3-5 bullet key messages + CTA
//   - "pitch"        → reframe as problem/solution/ask
//   - "standup"      → extract done/doing/blocked
//   - "transcript"   → extract participants, points, decisions, actions
//   - "report"       → extract summary, findings, recommendations, next steps
//   - "decision-log" → extract decision, rationale, alternatives
//   - "review"       → extract wins, issues, patterns, next actions
//   - "note"         → extract context, content, route
//   - "proposal"     → extract problem, solution, ask
//   - "spec"         → pass through (technical content needs no reframing)
//   - any other      → pass through unchanged
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
	case "transcript":
		return composeTranscript(content)
	case "report":
		return composeReport(content)
	case "decision-log", "decision_log", "decisionlog":
		return composeDecisionLog(content)
	case "review":
		return composeReview(content)
	case "note":
		return composeNote(content)
	case "proposal":
		return composeProposal(content)
	default:
		return content
	}
}

// ComposeResult holds the output of a receiver-aware composition call.
type ComposeResult struct {
	Content       string `json:"content"`
	ReceiverName  string `json:"receiver_name"`
	OriginalGenre string `json:"original_genre"`
	OutputGenre   string `json:"output_genre"`
	Converted     bool   `json:"converted"`
}

// ComposeForReceiver looks up the receiver's preferred genre from topology and
// composes content accordingly. When the receiver is not found in topology, or
// topology is nil, the content is returned unchanged with Converted = false.
func ComposeForReceiver(content string, receiverName string, topo *TopologyConfig) ComposeResult {
	result := ComposeResult{
		Content:       content,
		ReceiverName:  receiverName,
		OriginalGenre: "unknown",
		OutputGenre:   "passthrough",
		Converted:     false,
	}

	if topo == nil || receiverName == "" {
		return result
	}

	person := topo.GetPersonConfig(receiverName)
	if person == nil {
		return result
	}

	result.OutputGenre = person.Genre
	// We do not classify the incoming genre — caller did not provide it.
	result.OriginalGenre = "source"

	composed := Compose(content, person.Genre)
	if composed != content {
		result.Content = composed
		result.Converted = true
	}

	return result
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

// ── additional genre composers ─────────────────────────────────────────────────

// composeTranscript extracts participants, key points, decisions, and action
// items from meeting/call content into the transcript skeleton.
func composeTranscript(content string) string {
	participants := extractSection(content, []string{"attendees", "participants", "present", "invited"})
	discussed := extractSection(content, []string{"discussed", "key points", "topics", "agenda"})
	decisions := extractSection(content, []string{"decisions", "decided", "resolved", "agreed"})
	actions := extractSection(content, []string{"action items", "next steps", "todos", "follow-up"})

	// Graceful fallback: nothing could be extracted.
	if participants == "" && discussed == "" && decisions == "" && actions == "" {
		return content
	}

	var sb strings.Builder
	sb.WriteString("## Transcript\n\n")

	if participants != "" {
		sb.WriteString("**Participants:**\n")
		for _, p := range strings.Split(participants, "\n") {
			if p = strings.TrimSpace(p); p != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", p))
			}
		}
		sb.WriteString("\n")
	}

	if discussed != "" {
		sb.WriteString("**Key Points:**\n")
		for _, line := range strings.Split(discussed, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", line))
			}
		}
		sb.WriteString("\n")
	} else {
		// Fall back to first 3 sentences.
		sentences := extractSentences(content)
		if len(sentences) > 3 {
			sentences = sentences[:3]
		}
		if len(sentences) > 0 {
			sb.WriteString("**Key Points:**\n")
			for _, s := range sentences {
				sb.WriteString(fmt.Sprintf("- %s\n", s))
			}
			sb.WriteString("\n")
		}
	}

	if decisions != "" {
		sb.WriteString("**Decisions Made:**\n")
		for _, d := range strings.Split(decisions, "\n") {
			if d = strings.TrimSpace(d); d != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", d))
			}
		}
		sb.WriteString("\n")
	}

	if actions != "" {
		sb.WriteString("**Action Items:**\n")
		for _, a := range strings.Split(actions, "\n") {
			if a = strings.TrimSpace(a); a != "" {
				sb.WriteString(fmt.Sprintf("- [ ] %s\n", a))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// composeReport formats content into the report skeleton:
// Summary | Findings | Recommendations | Next Steps
func composeReport(content string) string {
	sentences := extractSentences(content)
	if len(sentences) == 0 {
		return content
	}

	var sb strings.Builder
	sb.WriteString("## Report\n\n")

	// Summary: first 1-2 sentences.
	end := 2
	if len(sentences) < end {
		end = len(sentences)
	}
	sb.WriteString("**Summary:** ")
	sb.WriteString(strings.Join(sentences[:end], " "))
	sb.WriteString("\n\n")

	// Findings: extracted section or next chunk of sentences.
	findings := extractSection(content, []string{"findings", "results", "data", "metrics", "analysis"})
	if findings != "" {
		sb.WriteString("**Findings:**\n")
		for _, f := range strings.Split(findings, "\n") {
			if f = strings.TrimSpace(f); f != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", f))
			}
		}
		sb.WriteString("\n")
	} else if len(sentences) > end {
		mid := sentences[end:]
		if len(mid) > 4 {
			mid = mid[:4]
		}
		sb.WriteString("**Findings:**\n")
		for _, s := range mid {
			sb.WriteString(fmt.Sprintf("- %s\n", s))
		}
		sb.WriteString("\n")
	}

	// Recommendations.
	recs := extractSection(content, []string{"recommendations", "suggest", "propose", "recommend"})
	if recs != "" {
		sb.WriteString("**Recommendations:**\n")
		num := 1
		for _, r := range strings.Split(recs, "\n") {
			if r = strings.TrimSpace(r); r != "" {
				sb.WriteString(fmt.Sprintf("%d. %s\n", num, r))
				num++
			}
		}
		sb.WriteString("\n")
	}

	// Next Steps.
	next := extractSection(content, []string{"next steps", "action items", "todos", "follow-up"})
	if next != "" {
		sb.WriteString("**Next Steps:**\n")
		for _, n := range strings.Split(next, "\n") {
			if n = strings.TrimSpace(n); n != "" {
				sb.WriteString(fmt.Sprintf("- [ ] %s\n", n))
			}
		}
	}

	return sb.String()
}

// composeDecisionLog extracts the decision, rationale, and alternatives into
// the decision-log skeleton.
func composeDecisionLog(content string) string {
	sentences := extractSentences(content)

	decision := extractSection(content, []string{"decision", "decided", "we decided", "going with", "chosen"})
	if decision == "" && len(sentences) > 0 {
		decision = sentences[0]
	}
	rationale := extractSection(content, []string{"rationale", "because", "reason", "why"})
	decider := extractSection(content, []string{"decided by", "owner", "decider", "approved by"})
	alts := extractSection(content, []string{"alternatives", "options considered", "other options", "rejected"})

	// Graceful fallback.
	if decision == "" && rationale == "" && alts == "" {
		return content
	}

	var sb strings.Builder
	sb.WriteString("## Decision Log\n\n")

	if decision != "" {
		sb.WriteString("**Decision:** ")
		sb.WriteString(strings.SplitN(decision, "\n", 2)[0])
		sb.WriteString("\n\n")
	}

	if rationale != "" {
		sb.WriteString("**Rationale:** ")
		sb.WriteString(strings.SplitN(rationale, "\n", 2)[0])
		sb.WriteString("\n\n")
	}

	if decider != "" {
		sb.WriteString("**Decided By:** ")
		sb.WriteString(strings.SplitN(decider, "\n", 2)[0])
		sb.WriteString("\n\n")
	}

	if alts != "" {
		sb.WriteString("**Alternatives Considered:**\n")
		for _, a := range strings.Split(alts, "\n") {
			if a = strings.TrimSpace(a); a != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", a))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// composeReview formats content into the review skeleton:
// Wins | Issues | Patterns | Next
func composeReview(content string) string {
	wins := extractSection(content, []string{"wins", "positives", "good", "approved", "lgtm", "ship it"})
	issues := extractSection(content, []string{"issues", "problems", "concerns", "needs changes", "blockers", "critical", "major", "minor"})
	patterns := extractSection(content, []string{"patterns", "observations", "trends", "recurring"})
	next := extractSection(content, []string{"next", "next steps", "action items", "follow-up", "suggestions"})

	// Graceful fallback: nothing extractable — return a minimal one-liner summary.
	if wins == "" && issues == "" && patterns == "" && next == "" {
		sentences := extractSentences(content)
		if len(sentences) == 0 {
			return content
		}
		var sb strings.Builder
		sb.WriteString("## Review\n\n")
		sb.WriteString("**Summary:** ")
		sb.WriteString(sentences[0])
		sb.WriteString("\n")
		return sb.String()
	}

	var sb strings.Builder
	sb.WriteString("## Review\n\n")

	if wins != "" {
		sb.WriteString("**Wins:**\n")
		for _, w := range strings.Split(wins, "\n") {
			if w = strings.TrimSpace(w); w != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", w))
			}
		}
		sb.WriteString("\n")
	}

	if issues != "" {
		sb.WriteString("**Issues:**\n")
		for _, iss := range strings.Split(issues, "\n") {
			if iss = strings.TrimSpace(iss); iss != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", iss))
			}
		}
		sb.WriteString("\n")
	}

	if patterns != "" {
		sb.WriteString("**Patterns:**\n")
		for _, p := range strings.Split(patterns, "\n") {
			if p = strings.TrimSpace(p); p != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", p))
			}
		}
		sb.WriteString("\n")
	}

	if next != "" {
		sb.WriteString("**Next:**\n")
		for _, n := range strings.Split(next, "\n") {
			if n = strings.TrimSpace(n); n != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", n))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// composeNote formats content into the note skeleton: Context | Content | Route.
// Notes are intentionally lightweight — they are already minimal signals.
func composeNote(content string) string {
	sentences := extractSentences(content)
	if len(sentences) == 0 {
		return content
	}

	var sb strings.Builder
	sb.WriteString("## Note\n\n")

	// Context: first sentence is the trigger.
	sb.WriteString("**Context:** ")
	sb.WriteString(sentences[0])
	sb.WriteString("\n\n")

	// Content: remaining sentences as bullets.
	if len(sentences) > 1 {
		sb.WriteString("**Content:**\n")
		for _, s := range sentences[1:] {
			sb.WriteString(fmt.Sprintf("- %s\n", s))
		}
		sb.WriteString("\n")
	}

	// Route: extracted from routing keywords, omitted when absent.
	route := extractSection(content, []string{"route", "target", "send to", "belongs in", "file under"})
	if route != "" {
		sb.WriteString("**Route:** ")
		sb.WriteString(strings.SplitN(route, "\n", 2)[0])
		sb.WriteString("\n")
	}

	return sb.String()
}

// composeProposal formats content as a proposal: Problem | Solution | Evidence | Ask.
// Proposals differ from pitches by focusing on internal problem-solving rather
// than external fundraising framing.
func composeProposal(content string) string {
	sentences := extractSentences(content)
	if len(sentences) == 0 {
		return content
	}

	problem := extractSection(content, []string{"problem", "challenge", "issue", "pain point", "current state"})
	if problem == "" && len(sentences) > 0 {
		problem = sentences[0]
	}

	solution := extractSection(content, []string{"solution", "proposal", "approach", "recommendation", "proposed"})
	if solution == "" && len(sentences) > 1 {
		solution = sentences[1]
	}

	evidence := extractSection(content, []string{"evidence", "data", "results", "proof", "metrics", "rationale"})
	if evidence == "" && len(sentences) > 2 {
		supporting := sentences[2:]
		if len(supporting) > 3 {
			supporting = supporting[:3]
		}
		evidence = strings.Join(supporting, "\n")
	}

	var sb strings.Builder
	sb.WriteString("## Proposal\n\n")

	if problem != "" {
		sb.WriteString("**Problem:** ")
		sb.WriteString(strings.SplitN(problem, "\n", 2)[0])
		sb.WriteString("\n\n")
	}

	if solution != "" {
		sb.WriteString("**Solution:** ")
		sb.WriteString(strings.SplitN(solution, "\n", 2)[0])
		sb.WriteString("\n\n")
	}

	if evidence != "" {
		sb.WriteString("**Evidence:**\n")
		for _, e := range strings.Split(evidence, "\n") {
			if e = strings.TrimSpace(e); e != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", e))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("**Ask:** ")
	sb.WriteString(extractCTA(content))
	sb.WriteString("\n")

	return sb.String()
}
