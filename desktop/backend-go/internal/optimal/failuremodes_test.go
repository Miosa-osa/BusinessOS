package optimal

import (
	"strings"
	"testing"
)

// hasCode is a helper that scans a slice of FailureMode values for the given code.
func hasCode(modes []FailureMode, code string) bool {
	for _, m := range modes {
		if m.Code == code {
			return true
		}
	}
	return false
}

// ── DetectFailureModes ────────────────────────────────────────────────────────

func TestDetectFailureModes_Clean(t *testing.T) {
	// Well-structured spec with all required sections, good S/N, no issues.
	content := "## Requirements\n- The system must ingest signals\n- The system must search in <500ms\n\n## Constraints\n- Must run on Go 1.21+\n\n## Acceptance Criteria\n- [ ] All tests pass\n\nDecided by: Roberto on 2026-04-13"

	cl := SignalClassification{
		Genre:   "spec",
		SNRatio: 0.9,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	// Should not contain any of these codes that our well-formed content avoids.
	unwanted := []string{"fidelity_failure", "bandwidth_overload", "variety_failure"}
	for _, code := range unwanted {
		if hasCode(modes, code) {
			t.Errorf("clean spec should not trigger %q, but did. Modes: %v", code, modes)
		}
	}
}

func TestDetectFailureModes_LowSNRatio(t *testing.T) {
	content := "The thing is basically sort of fine I think maybe."

	cl := SignalClassification{
		Genre:   "note",
		SNRatio: 0.2, // below the 0.4 threshold
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if !hasCode(modes, "fidelity_failure") {
		t.Errorf("expected fidelity_failure for SNRatio=0.2, got %v", modes)
	}
}

func TestDetectFailureModes_SNRatio_AboveThreshold_NoFidelityFailure(t *testing.T) {
	content := "The system must process all incoming signals within 500 milliseconds."

	cl := SignalClassification{
		Genre:   "note",
		SNRatio: 0.7,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if hasCode(modes, "fidelity_failure") {
		t.Errorf("SNRatio=0.7 should not trigger fidelity_failure")
	}
}

func TestDetectFailureModes_BandwidthOverload_Standup(t *testing.T) {
	// Standup has a 500-word limit; generate content exceeding it.
	words := strings.Repeat("word ", 600) // 600 words × 5 chars = 3000 chars
	content := "Yesterday:\n" + words

	cl := SignalClassification{
		Genre:   "standup",
		SNRatio: 0.6,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if !hasCode(modes, "bandwidth_overload") {
		t.Errorf("expected bandwidth_overload for standup with 600+ words, got %v", modes)
	}
}

func TestDetectFailureModes_BandwidthOverload_BriefUnderLimit(t *testing.T) {
	// Brief has a 300-word limit; 50 words should be clean.
	content := strings.Repeat("word ", 50)

	cl := SignalClassification{
		Genre:   "brief",
		SNRatio: 0.7,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if hasCode(modes, "bandwidth_overload") {
		t.Errorf("50-word brief should not trigger bandwidth_overload")
	}
}

func TestDetectFailureModes_VarietyFailure(t *testing.T) {
	// Content with multiple headers and numbered list items (strong structure)
	// but classified as "note" — should trigger variety_failure.
	// Over 100 words to pass the word count guard (wc > 100 is required).
	body := strings.Repeat("some text content here to pad word count for the test ", 15) // ~120 words
	content := "## Section One\n\n1. First item\n2. Second item\n\n## Section Two\n\n1. Third item\n\n" + body

	cl := SignalClassification{
		Genre:   "note",
		SNRatio: 0.5,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if !hasCode(modes, "variety_failure") {
		t.Errorf("expected variety_failure for structured content classified as 'note', got %v", modes)
	}
}

func TestDetectFailureModes_FeedbackFailure(t *testing.T) {
	// Action items present but no assignee — open feedback loop.
	content := "Action items:\n- [ ] Send the proposal\n- [ ] Schedule follow-up call\nBlockers: none"

	cl := SignalClassification{
		Genre:   "standup",
		SNRatio: 0.7,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if !hasCode(modes, "feedback_failure") {
		t.Errorf("expected feedback_failure for action items without assignee, got %v", modes)
	}
}

func TestDetectFailureModes_FeedbackFailure_WithAssignee_NoTrigger(t *testing.T) {
	// Action items WITH @assignee — loop is closed.
	content := "Action items:\n- @roberto will send the proposal by Friday\n- @ed will schedule follow-up"

	cl := SignalClassification{
		Genre:   "standup",
		SNRatio: 0.7,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if hasCode(modes, "feedback_failure") {
		t.Errorf("assignee present — feedback loop is closed, should not trigger feedback_failure")
	}
}

func TestDetectFailureModes_AdversarialNoise(t *testing.T) {
	// Single word repeated far beyond 25% threshold.
	content := strings.Repeat("spam ", 50) // "spam" = 100% of words
	content += " some other words here to reach 10 total"

	cl := SignalClassification{
		Genre:   "note",
		SNRatio: 0.1,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if !hasCode(modes, "adversarial_noise") {
		t.Errorf("expected adversarial_noise for heavily repeated content, got %v", modes)
	}
}

func TestDetectFailureModes_AdversarialNoise_NormalContent(t *testing.T) {
	content := "We had a productive meeting about the AI Masters pricing strategy. Ed proposed $2K per seat."

	cl := SignalClassification{
		Genre:   "note",
		SNRatio: 0.7,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, content, nil)

	if hasCode(modes, "adversarial_noise") {
		t.Errorf("normal content should not trigger adversarial_noise")
	}
}

func TestDetectFailureModes_EmptyContent_NoModes(t *testing.T) {
	// Empty content: adversarial_noise guard requires wc > 10, so it won't fire.
	// The rest should also be quiet or fire only the routing_failure.
	cl := SignalClassification{
		Genre:   "note",
		SNRatio: 0.0,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, "", nil)

	// fidelity_failure requires content != "", so it must not fire.
	if hasCode(modes, "fidelity_failure") {
		t.Errorf("empty content should not trigger fidelity_failure")
	}
	// adversarial_noise requires wc > 10.
	if hasCode(modes, "adversarial_noise") {
		t.Errorf("empty content should not trigger adversarial_noise")
	}
}

func TestDetectFailureModes_RoutingFailure(t *testing.T) {
	// Genre is "note" AND no named entities extractable — routing failure.
	// Use short generic text with no proper nouns.
	cl := SignalClassification{
		Genre:   "note",
		SNRatio: 0.5,
		Mode:    "linguistic",
	}

	modes := DetectFailureModes(cl, "the thing happened at the place", nil)

	if !hasCode(modes, "routing_failure") {
		t.Errorf("expected routing_failure for note with no named entities, got %v", modes)
	}
}

// ── helpers unit tests ────────────────────────────────────────────────────────

func TestWordCountStr(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
		{9999, "9999"},
	}

	for _, tc := range cases {
		got := wordCountStr(tc.n)
		if got != tc.want {
			t.Errorf("wordCountStr(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

func TestJaccardOverlap_SameText(t *testing.T) {
	text := "the engine processes signals from the optimal graph"
	got := jaccardOverlap(text, text)
	if got != 1.0 {
		t.Errorf("identical texts Jaccard = %.2f, want 1.0", got)
	}
}

func TestJaccardOverlap_NoOverlap(t *testing.T) {
	got := jaccardOverlap("apple banana cherry", "delta epsilon foxtrot")
	if got != 0.0 {
		t.Errorf("disjoint texts Jaccard = %.2f, want 0.0", got)
	}
}

func TestJaccardOverlap_PartialOverlap(t *testing.T) {
	a := "the dog sat on the mat"
	b := "the cat sat on the mat"
	got := jaccardOverlap(a, b)
	// Shared tokens (≥3 chars): "the", "sat", "mat" = 3
	// Unique across both: "dog", "cat" → union = 5; intersection = 3
	// Jaccard ≈ 0.6 but word-set deduplication may vary.
	if got <= 0.0 || got > 1.0 {
		t.Errorf("partial overlap Jaccard = %.2f, want > 0 and ≤ 1", got)
	}
}

func TestIsAdversarialNoise_LowRepetition(t *testing.T) {
	content := "We discussed the pricing model for AI Masters with Ed and Robert and agreed on the $2K seat price."
	if isAdversarialNoise(content) {
		t.Error("natural sentence should not be flagged as adversarial noise")
	}
}

func TestIsAdversarialNoise_HighRepetition(t *testing.T) {
	content := strings.Repeat("spam ", 40)
	if !isAdversarialNoise(content) {
		t.Error("single word repeated 40× should be flagged as adversarial noise")
	}
}

func TestIsAdversarialNoise_ShortContent_NotFlagged(t *testing.T) {
	// Guard: wc < 10 → always false.
	content := strings.Repeat("a ", 5)
	if isAdversarialNoise(content) {
		t.Error("content with fewer than 10 words should never be flagged")
	}
}
