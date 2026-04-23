package optimal

import (
	"strings"
	"testing"
)

// ── Compose dispatch ──────────────────────────────────────────────────────────

func TestCompose_SpecPassThrough(t *testing.T) {
	content := "# Spec\n\n## Requirements\n- Must process signals\n\n## Acceptance Criteria\n- All tests pass"
	got := Compose(content, "spec")
	if got != content {
		t.Error("spec genre should pass through unchanged")
	}
}

func TestCompose_UnknownGenrePassThrough(t *testing.T) {
	content := "Some arbitrary content."
	got := Compose(content, "unknown-genre-xyz")
	if got != content {
		t.Error("unknown genre should pass through unchanged")
	}
}

func TestCompose_CaseInsensitive(t *testing.T) {
	content := "Deploy the engine now. It handles AI signals."
	got1 := Compose(content, "brief")
	got2 := Compose(content, "BRIEF")
	got3 := Compose(content, "Brief")

	if got1 != got2 || got1 != got3 {
		t.Error("Compose should be case-insensitive in receiverGenre")
	}
}

func TestCompose_EmptyContent_NoPanic(t *testing.T) {
	// Empty content — composers must not panic. Return values may vary per genre:
	// some (standup) emit a fallback skeleton; others return the empty string.
	genres := []string{
		"brief", "pitch", "standup", "transcript",
		"report", "decision-log", "review", "note", "proposal",
	}
	for _, g := range genres {
		// Should not panic — that is the only hard contract here.
		_ = Compose("", g)
	}
}

func TestCompose_EmptyContent_PassThroughGenres(t *testing.T) {
	// Genres that always pass through return the input unchanged.
	passThrough := []string{"spec", "unknown-xyz"}
	for _, g := range passThrough {
		got := Compose("", g)
		if got != "" {
			t.Errorf("Compose(%q, %q) = %q, want empty passthrough", "", g, got)
		}
	}
}

// ── composeBrief ─────────────────────────────────────────────────────────────

func TestComposeBrief_HasRequiredSections(t *testing.T) {
	content := "We need to close the AI Masters deal. The course is ready for delivery. Ed has confirmed interest. Sign the agreement by Friday."
	got := Compose(content, "brief")

	required := []string{"## Brief", "**Objective:**", "**Key Messages:**", "**Call to Action:**"}
	for _, req := range required {
		if !strings.Contains(got, req) {
			t.Errorf("brief missing section %q in output:\n%s", req, got)
		}
	}
}

func TestComposeBrief_ObjectiveIsFirstSentence(t *testing.T) {
	content := "Close the AI Masters deal. Secondary context here. Tertiary point."
	got := Compose(content, "brief")

	// Objective should contain the first sentence text.
	if !strings.Contains(got, "Close the AI Masters deal") {
		t.Errorf("brief Objective should contain first sentence: %s", got)
	}
}

func TestComposeBrief_KeyMessagesCapped(t *testing.T) {
	// Generate 10 sentences — Key Messages should be capped at 5.
	sentences := make([]string, 10)
	for i := range sentences {
		sentences[i] = "Sentence number placeholder text here."
	}
	content := strings.Join(sentences, " ")
	got := Compose(content, "brief")

	// Count bullet lines in Key Messages section.
	bulletCount := strings.Count(got, "\n- ")
	if bulletCount > 5 {
		t.Errorf("brief Key Messages should have at most 5 bullets, got %d", bulletCount)
	}
}

// ── composePitch ──────────────────────────────────────────────────────────────

func TestComposePitch_HasRequiredSections(t *testing.T) {
	content := "Agencies waste 10 hours per week on manual reporting. Our platform automates signal routing. We have 12 paying customers. Invest $250K for 10% equity."
	got := Compose(content, "pitch")

	required := []string{"## Pitch", "**Problem:**", "**Solution:**", "**Ask:**"}
	for _, req := range required {
		if !strings.Contains(got, req) {
			t.Errorf("pitch missing section %q:\n%s", req, got)
		}
	}
}

// ── composeStandup ────────────────────────────────────────────────────────────

func TestComposeStandup_ExtractsStructuredContent(t *testing.T) {
	content := "Done:\n- Finished the graph module\n- Deployed to staging\n\nDoing:\n- Writing test coverage\n\nBlockers:\n- Need DB migration script"
	got := Compose(content, "standup")

	if !strings.Contains(got, "## Standup") {
		t.Errorf("standup missing header")
	}
	if !strings.Contains(got, "**Done:**") {
		t.Errorf("standup missing Done section")
	}
	if !strings.Contains(got, "**Doing:**") {
		t.Errorf("standup missing Doing section")
	}
	if !strings.Contains(got, "**Blocked:**") {
		t.Errorf("standup missing Blocked section")
	}
}

func TestComposeStandup_FallbackWhenUnstructured(t *testing.T) {
	// No section headers — should fall back gracefully.
	content := "I worked on the engine. No blockers at the moment."
	got := Compose(content, "standup")

	if !strings.Contains(got, "## Standup") {
		t.Errorf("standup fallback should still emit header")
	}
	// Fall through branch emits (not detected) for Done.
	if !strings.Contains(got, "(not detected)") {
		t.Errorf("unstructured standup should emit '(not detected)' for missing sections")
	}
}

// ── composeTranscript ─────────────────────────────────────────────────────────

func TestComposeTranscript_ExtractsParticipantsAndActions(t *testing.T) {
	content := "Attendees:\n- Ed Honour\n- Roberto Luna\n\nKey Points:\n- Discussed pricing\n\nAction Items:\n- Send proposal by Friday"
	got := Compose(content, "transcript")

	if !strings.Contains(got, "## Transcript") {
		t.Errorf("transcript missing header")
	}
	if !strings.Contains(got, "**Participants:**") {
		t.Errorf("transcript missing Participants section")
	}
	if !strings.Contains(got, "**Action Items:**") {
		t.Errorf("transcript missing Action Items section")
	}
}

func TestComposeTranscript_FallbackWhenNoStructure(t *testing.T) {
	// Flat prose with no section headers — should return original content.
	content := "Ed called and we talked about pricing for ten minutes."
	got := Compose(content, "transcript")
	if got != content {
		t.Errorf("transcript with no extractable structure should return original content")
	}
}

// ── composeReport ─────────────────────────────────────────────────────────────

func TestComposeReport_HasRequiredSections(t *testing.T) {
	content := "The system processed 1000 signals in the last month. Response latency averaged 12ms. Findings: throughput is above target. Recommendations: increase cache TTL. Next Steps:\n- Monitor for 30 days"
	got := Compose(content, "report")

	required := []string{"## Report", "**Summary:**"}
	for _, req := range required {
		if !strings.Contains(got, req) {
			t.Errorf("report missing %q:\n%s", req, got)
		}
	}
}

// ── composeDecisionLog ────────────────────────────────────────────────────────

func TestComposeDecisionLog_ExtractsDecisionAndRationale(t *testing.T) {
	content := "Decision:\n- Go with $2K per seat pricing\n\nRationale:\n- Matches market rate\n\nDecided By:\n- Roberto Luna"
	got := Compose(content, "decision-log")

	if !strings.Contains(got, "## Decision Log") {
		t.Errorf("decision-log missing header")
	}
	if !strings.Contains(got, "**Decision:**") {
		t.Errorf("decision-log missing Decision field")
	}
	if !strings.Contains(got, "**Rationale:**") {
		t.Errorf("decision-log missing Rationale field")
	}
	if !strings.Contains(got, "**Decided By:**") {
		t.Errorf("decision-log missing Decided By field")
	}
}

func TestComposeDecisionLog_AliasesWork(t *testing.T) {
	content := "Decision:\n- Use PostgreSQL\n\nRationale:\n- Already in stack"
	got1 := Compose(content, "decision-log")
	got2 := Compose(content, "decision_log")
	got3 := Compose(content, "decisionlog")

	if got1 != got2 || got1 != got3 {
		t.Error("all three decision-log aliases should produce identical output")
	}
}

func TestComposeDecisionLog_FallbackWhenNoStructure(t *testing.T) {
	// Flat prose — should return original.
	content := "We chose PostgreSQL because it was already in the stack."
	got := Compose(content, "decision-log")
	// With no extractable sections, decision is filled from first sentence.
	if !strings.Contains(got, "## Decision Log") {
		t.Errorf("decision-log should always emit header; got: %q", got)
	}
}

// ── composeReview ─────────────────────────────────────────────────────────────

func TestComposeReview_WithStructuredContent(t *testing.T) {
	content := "Wins:\n- Engine is fast\n\nIssues:\n- Coverage gap in graph module\n\nNext Steps:\n- Add more tests"
	got := Compose(content, "review")

	if !strings.Contains(got, "## Review") {
		t.Errorf("review missing header")
	}
	if !strings.Contains(got, "**Wins:**") {
		t.Errorf("review missing Wins section")
	}
	if !strings.Contains(got, "**Issues:**") {
		t.Errorf("review missing Issues section")
	}
}

func TestComposeReview_FallbackSummary(t *testing.T) {
	// No extractable sections — minimal review with Summary.
	content := "The code looks good overall with minor concerns."
	got := Compose(content, "review")

	if !strings.Contains(got, "## Review") {
		t.Errorf("review fallback should still emit header")
	}
	if !strings.Contains(got, "**Summary:**") {
		t.Errorf("review fallback should emit Summary")
	}
}

// ── composeNote ───────────────────────────────────────────────────────────────

func TestComposeNote_HasContextAndContent(t *testing.T) {
	content := "Ed called about pricing. He wants to move forward with $2K. We agreed to send the contract by Monday."
	got := Compose(content, "note")

	if !strings.Contains(got, "## Note") {
		t.Errorf("note missing header")
	}
	if !strings.Contains(got, "**Context:**") {
		t.Errorf("note missing Context field")
	}
}

func TestComposeNote_SingleSentenceNoContentSection(t *testing.T) {
	// One sentence — Content section (remaining sentences) should be absent.
	content := "Ed called about pricing strategy for the launch."
	got := Compose(content, "note")

	if strings.Contains(got, "**Content:**") {
		t.Errorf("single-sentence note should not emit empty Content section")
	}
}

// ── composeProposal ───────────────────────────────────────────────────────────

func TestComposeProposal_HasRequiredSections(t *testing.T) {
	content := "Problem:\n- Manual signal routing wastes 2h/day\n\nSolution:\n- Build automated routing engine\n\nEvidence:\n- Saved 40% time in pilot"
	got := Compose(content, "proposal")

	required := []string{"## Proposal", "**Problem:**", "**Solution:**", "**Ask:**"}
	for _, req := range required {
		if !strings.Contains(got, req) {
			t.Errorf("proposal missing %q:\n%s", req, got)
		}
	}
}

// ── ComposeForReceiver ────────────────────────────────────────────────────────

func TestComposeForReceiver_NilTopologyNoConversion(t *testing.T) {
	content := "Signal content that should pass through unchanged."
	result := ComposeForReceiver(content, "Ed", nil)

	if result.Converted {
		t.Error("nil topology should never convert")
	}
	if result.Content != content {
		t.Errorf("nil topology should return original content")
	}
}

func TestComposeForReceiver_EmptyReceiverNameNoConversion(t *testing.T) {
	topo := &TopologyConfig{
		People: map[string]PersonConfig{
			"Ed": {Role: "partner", Genre: "brief"},
		},
	}
	content := "Some signal content."
	result := ComposeForReceiver(content, "", topo)

	if result.Converted {
		t.Error("empty receiver name should not convert")
	}
}

func TestComposeForReceiver_KnownReceiverConverts(t *testing.T) {
	topo := &TopologyConfig{
		People: map[string]PersonConfig{
			"Ed": {Role: "partner", Genre: "brief"},
		},
	}
	// Content long enough that composeBrief actually transforms it.
	content := "Deploy the new system now. We have secured three new clients this month. Revenue is up 40 percent. Sign the contract today to move forward quickly."
	result := ComposeForReceiver(content, "Ed", topo)

	if !result.Converted {
		t.Error("known receiver with genre 'brief' should trigger conversion")
	}
	if !strings.Contains(result.Content, "## Brief") {
		t.Errorf("converted content should contain Brief header, got: %q", result.Content)
	}
}

func TestComposeForReceiver_UnknownReceiverNoConversion(t *testing.T) {
	topo := &TopologyConfig{
		People: map[string]PersonConfig{
			"Ed": {Role: "partner", Genre: "brief"},
		},
	}
	content := "Some signal content."
	result := ComposeForReceiver(content, "UnknownPerson", topo)

	if result.Converted {
		t.Error("unknown receiver should not convert")
	}
}

// ── extractSentences ─────────────────────────────────────────────────────────

func TestExtractSentences_Basic(t *testing.T) {
	content := "First sentence here. Second sentence follows. Third one ends it."
	got := extractSentences(content)
	if len(got) < 2 {
		t.Errorf("extractSentences returned %d sentences, want ≥ 2", len(got))
	}
}

func TestExtractSentences_FiltersShortFragments(t *testing.T) {
	content := "Ok. This is a longer sentence that should survive. Hi."
	got := extractSentences(content)
	for _, s := range got {
		if len(s) < 10 {
			t.Errorf("short fragment %q should have been filtered out", s)
		}
	}
}

func TestExtractSentences_EmptyReturnsNil(t *testing.T) {
	got := extractSentences("")
	if len(got) != 0 {
		t.Errorf("extractSentences(\"\") = %v, want empty", got)
	}
}

// ── extractSection ────────────────────────────────────────────────────────────

func TestExtractSection_MatchesHeader(t *testing.T) {
	content := "## Done:\n- Finished the engine\n- Deployed to staging\n\n## Other:\n- Unrelated"
	got := extractSection(content, []string{"done"})
	if !strings.Contains(got, "Finished the engine") {
		t.Errorf("extractSection did not find 'done' content: %q", got)
	}
	if strings.Contains(got, "Unrelated") {
		t.Errorf("extractSection leaked content from other section: %q", got)
	}
}

func TestExtractSection_NotFound_ReturnsEmpty(t *testing.T) {
	content := "## Summary:\nSome content here."
	got := extractSection(content, []string{"blockers"})
	if got != "" {
		t.Errorf("extractSection for missing section = %q, want empty", got)
	}
}
