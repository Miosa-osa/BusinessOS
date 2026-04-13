package optimal

import (
	"regexp"
	"strings"
)

// ExtractedMemory is a single structured memory pulled from a conversation message.
type ExtractedMemory struct {
	Category string `json:"category"` // fact, preference, decision, relationship, skill, context
	Content  string `json:"content"`
	Source   string `json:"source"` // the original message text
}

// ExtractMemories scans a slice of conversation messages and returns structured
// memories categorised by type. Classification is regex-based — no LLM required.
func ExtractMemories(messages []string) []ExtractedMemory {
	var memories []ExtractedMemory
	for _, msg := range messages {
		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}
		cat := classifyMemory(msg)
		memories = append(memories, ExtractedMemory{
			Category: cat,
			Content:  summariseMemory(msg),
			Source:   msg,
		})
	}
	return memories
}

// ── category patterns ─────────────────────────────────────────────────────────

var (
	rePreference = regexp.MustCompile(
		`(?i)\b(i prefer|i like|i always|i never|i hate|i love|i enjoy|` +
			`i want|my preference|i usually|i tend to)\b`,
	)

	reDecisionMem = regexp.MustCompile(
		`(?i)\b(we decided|the decision is|i decided|agreed to|` +
			`we chose|going with|we'll go with|final decision|` +
			`decision:)\b`,
	)

	reRelationship = regexp.MustCompile(
		`(?i)\b(works with|reports to|manages|is the|is a|` +
			`partners with|owns|founded|leads|heads|joined)\b`,
	)

	reSkill = regexp.MustCompile(
		`(?i)\b(i know how to|i can|i'm able to|i've built|` +
			`i've implemented|skilled in|expert in|experienced with)\b`,
	)

	reFact = regexp.MustCompile(
		// Statements of fact: proper nouns followed by descriptive verb phrases.
		`(?i)\b(is|are|was|were|has|have|had|will|costs?|earns?|generates?|` +
			`launched|released|built|created|founded)\b`,
	)

	reProperNounPhrase = regexp.MustCompile(`\b[A-Z][a-z]{1,}(?:\s+[A-Z][a-z]{1,})*\b`)
)

// classifyMemory returns the memory category for a single message.
func classifyMemory(msg string) string {
	switch {
	case rePreference.MatchString(msg):
		return "preference"
	case reDecisionMem.MatchString(msg):
		return "decision"
	case reRelationship.MatchString(msg) && reProperNounPhrase.MatchString(msg):
		return "relationship"
	case reSkill.MatchString(msg):
		return "skill"
	case reFact.MatchString(msg) && reProperNounPhrase.MatchString(msg):
		return "fact"
	default:
		return "context"
	}
}

// summariseMemory returns a compact version of the message (first sentence, max 200 chars).
func summariseMemory(msg string) string {
	// Use firstSentence from compress.go (same package).
	s := firstSentence(msg)
	if len(s) > 200 {
		s = s[:200]
		if idx := strings.LastIndexByte(s, ' '); idx > 100 {
			s = s[:idx]
		}
		s += "..."
	}
	return strings.TrimSpace(s)
}
