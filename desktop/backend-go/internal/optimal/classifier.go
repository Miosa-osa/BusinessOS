package optimal

import (
	"regexp"
	"strings"
	"unicode"
)

// SignalClassification represents the 5-dimensional Signal Theory decomposition
// S=(Mode, Genre, Type, Format, Structure) plus a signal-to-noise ratio score.
type SignalClassification struct {
	Mode      string  `json:"mode"`      // linguistic, visual, code, data, mixed
	Genre     string  `json:"genre"`     // transcript, brief, spec, plan, note, decision-log, standup, review, report, pitch
	Type      string  `json:"type"`      // direct, inform, commit, decide, express
	Format    string  `json:"format"`    // markdown, code, json, yaml, unknown
	Structure string  `json:"structure"` // genre-specific skeleton label
	SNRatio   float64 `json:"sn_ratio"`  // 0.0–1.0 signal quality
}

// ── compiled patterns (package-level to avoid repeated compilation) ───────────

var (
	// Mode detection
	reCodeBlock  = regexp.MustCompile("(?m)^```|^\\s{4}\\S|import\\s+[\"(]|func\\s+\\w+\\(|package\\s+\\w+|class\\s+\\w+|def\\s+\\w+\\(|#include|<script|<style")
	reTableOrCSV = regexp.MustCompile(`(?m)^\|.+\|.+\||\d+,\d+,\d+`)
	reImage      = regexp.MustCompile(`(?i)!\[.*?\]\(|<img\s`)

	// Genre: transcript
	reTranscript = regexp.MustCompile(`(?i)(attendees?|participants?|discussed|action items?|meeting notes?|call recap|transcri)`)
	// Genre: brief
	reBrief = regexp.MustCompile(`(?i)(key messages?|call to action|objective[:\s]|executive summary|tldr|tl;dr)`)
	// Genre: spec
	reSpec = regexp.MustCompile(`(?i)(requirements?[:\s#]|constraints?[:\s#]|acceptance criteria|user stor(y|ies)|must have|should have|given.+when.+then)`)
	// Genre: plan
	rePlan = regexp.MustCompile(`(?i)(non-negotiable|day-by-day|success criteria|time blocks?|dependencies[:\s]|milestones?[:\s]|q[1-4]\s+\d{4})`)
	// Genre: decision-log
	reDecisionLog = regexp.MustCompile(`(?i)(decided by|rationale[:\s]|decision[:\s]|we decided|the choice is|alternatives? considered|adr\s*\d|decision record)`)
	// Genre: standup
	reStandup = regexp.MustCompile(`(?i)(yesterday|today|blockers?|stand.?up|what did i|what will i)`)
	// Genre: review
	reReview = regexp.MustCompile(`(?i)(code review|pull request|pr #\d|approved|needs changes|lgtm|ship it|overall:|issues found)`)
	// Genre: report
	reReport = regexp.MustCompile(`(?i)(weekly report|monthly report|status report|summary[:\s]|findings?[:\s]|recommendations?[:\s])`)
	// Genre: pitch
	rePitch = regexp.MustCompile(`(?i)(problem statement|market size|value proposition|go-to-market|runway|burn rate|ask[:\s]\$|seed round|series [a-c])`)

	// Type detection
	reDirect   = regexp.MustCompile(`(?i)^(do|run|create|build|update|fix|write|implement|deploy|send|configure|set up|add|remove)\b`)
	reCommit   = regexp.MustCompile(`(?i)(by\s+\w+day|deadline[:\s]|due\s+(by|on|date)|will\s+(deliver|complete|ship|finish)|i\s+commit|i promise|we'll have)`)
	reDecide   = regexp.MustCompile(`(?i)(we decided|the choice is|final decision|going with|we're choosing|decided to|chosen approach)`)
	reExpress  = regexp.MustCompile(`(?i)(i feel|i think|i believe|honestly|frustrated|excited|worried|love\s+this|hate\s+this|not sure|hope|wish)`)
	reQuestion = regexp.MustCompile(`\?`)

	// Format detection
	reMarkdownHeader  = regexp.MustCompile(`(?m)^#{1,6}\s+\S`)
	reJSONStart       = regexp.MustCompile(`^\s*[\[{]`)
	reYAMLFrontmatter = regexp.MustCompile(`(?s)^---\n.+?\n---`)
	reYAMLKV          = regexp.MustCompile(`(?m)^\w[\w_-]*:\s+\S`)

	// S/N quality signals
	reProperNoun  = regexp.MustCompile(`\b[A-Z][a-z]{2,}\b`)
	reActionItem  = regexp.MustCompile(`(?i)(\[\s*[xX ]\]|action items?[:\s]|todo[:\s]|@\w+\s+will|next steps?)`)
	reFillerWords = regexp.MustCompile(`(?i)\b(basically|actually|literally|just|kind of|sort of|you know|i mean|like|um|uh|well|so)\b`)
	reStructure   = regexp.MustCompile(`(?m)^(#{1,6}\s|\*\s|-\s|\d+\.\s|\|\s)`)
)

// Classify performs a deterministic, regex-based Signal Theory classification
// of the given text. No external calls or allocations beyond the string itself.
func Classify(text string) SignalClassification {
	if text == "" {
		return SignalClassification{
			Mode:      "linguistic",
			Genre:     "note",
			Type:      "inform",
			Format:    "unknown",
			Structure: "note/capture",
			SNRatio:   0.0,
		}
	}

	mode := detectMode(text)
	genre := detectGenre(text)
	sigType := detectType(text)
	format := detectFormat(text)
	structure := genreStructure(genre)
	snRatio := computeSNRatio(text)

	return SignalClassification{
		Mode:      mode,
		Genre:     genre,
		Type:      sigType,
		Format:    format,
		Structure: structure,
		SNRatio:   snRatio,
	}
}

// ── dimension detectors ───────────────────────────────────────────────────────

func detectMode(text string) string {
	hasCode := reCodeBlock.MatchString(text)
	hasTable := reTableOrCSV.MatchString(text)
	hasImage := reImage.MatchString(text)

	switch {
	case hasCode && (hasTable || hasImage):
		return "mixed"
	case hasCode:
		return "code"
	case hasTable:
		return "data"
	case hasImage:
		return "visual"
	default:
		return "linguistic"
	}
}

func detectGenre(text string) string {
	// Priority order: most-specific first.
	switch {
	case reTranscript.MatchString(text):
		return "transcript"
	case reDecisionLog.MatchString(text):
		return "decision-log"
	case reSpec.MatchString(text):
		return "spec"
	case rePitch.MatchString(text):
		return "pitch"
	case rePlan.MatchString(text):
		return "plan"
	case reBrief.MatchString(text):
		return "brief"
	case reReview.MatchString(text):
		return "review"
	case reStandup.MatchString(text):
		return "standup"
	case reReport.MatchString(text):
		return "report"
	default:
		return "note"
	}
}

func detectType(text string) string {
	lower := strings.ToLower(strings.TrimSpace(text))

	// Check first non-blank line for imperative commands.
	firstLine := firstNonBlankLine(lower)
	if reDirect.MatchString(firstLine) {
		return "direct"
	}

	switch {
	case reDecide.MatchString(text):
		return "decide"
	case reCommit.MatchString(text):
		return "commit"
	case reExpress.MatchString(text):
		return "express"
	case reQuestion.MatchString(text):
		return "inform"
	default:
		return "inform"
	}
}

func detectFormat(text string) string {
	trimmed := strings.TrimSpace(text)

	// YAML frontmatter takes precedence over generic markdown.
	if reYAMLFrontmatter.MatchString(trimmed) {
		return "yaml"
	}
	// JSON/array literal.
	if reJSONStart.MatchString(trimmed) {
		return "json"
	}
	// Markdown (headers or lists).
	if reMarkdownHeader.MatchString(trimmed) {
		return "markdown"
	}
	// Standalone YAML key-value blocks (no frontmatter delimiters).
	if reYAMLKV.MatchString(trimmed) {
		kvCount := len(reYAMLKV.FindAllString(trimmed, -1))
		if kvCount >= 3 {
			return "yaml"
		}
	}
	// Code detection.
	if reCodeBlock.MatchString(trimmed) {
		return "code"
	}
	return "unknown"
}

// genreStructure returns the canonical genre skeleton label for a genre value.
// These mirror the genre skeletons defined in the OptimalOS signal theory spec.
func genreStructure(genre string) string {
	switch genre {
	case "transcript":
		return "transcript/participants-points-decisions-actions-questions"
	case "brief":
		return "brief/objective-messages-cta-materials"
	case "spec":
		return "spec/goal-requirements-constraints-architecture-acceptance"
	case "plan":
		return "plan/objective-nonnegotiables-timeblocks-dependencies-success"
	case "decision-log":
		return "decision-log/context-options-rationale-decision-consequences"
	case "standup":
		return "standup/yesterday-today-blockers"
	case "review":
		return "review/summary-issues-suggestions-positives"
	case "report":
		return "report/summary-findings-recommendations-next-steps"
	case "pitch":
		return "pitch/problem-solution-market-ask-team"
	default:
		return "note/context-content-route"
	}
}

// ── S/N ratio ─────────────────────────────────────────────────────────────────

// computeSNRatio scores signal quality on a 0.0–1.0 scale.
// Base is 0.5; bonuses are added for structural markers, entities, and action items;
// penalty for high filler-word density.
func computeSNRatio(text string) float64 {
	score := 0.5

	// Structural markers (headers, lists, tables) → +0.2
	if reStructure.MatchString(text) {
		score += 0.2
	}

	// Proper nouns / named entities → +0.1
	properNouns := reProperNoun.FindAllString(text, -1)
	if len(properNouns) >= 2 {
		score += 0.1
	}

	// Action items / task markers → +0.1
	if reActionItem.MatchString(text) {
		score += 0.1
	}

	// Filler-word ratio < 0.05 of total words → +0.1 (low noise)
	wordCount := countWords(text)
	if wordCount > 0 {
		fillerMatches := reFillerWords.FindAllString(text, -1)
		fillerRatio := float64(len(fillerMatches)) / float64(wordCount)
		if fillerRatio < 0.05 {
			score += 0.1
		}
	}

	// Clamp to [0.0, 1.0].
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}
	return score
}

// ── string helpers ────────────────────────────────────────────────────────────

// firstNonBlankLine returns the first non-empty line from text (lowercased by caller).
func firstNonBlankLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// countWords counts whitespace-separated tokens that contain at least one letter.
func countWords(text string) int {
	count := 0
	for _, field := range strings.Fields(text) {
		for _, r := range field {
			if unicode.IsLetter(r) {
				count++
				break
			}
		}
	}
	return count
}
