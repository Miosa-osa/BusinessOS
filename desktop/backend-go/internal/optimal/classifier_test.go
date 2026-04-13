package optimal

import (
	"testing"
)

func TestClassify_Mode(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantMode string
	}{
		{
			name:     "code block triggers code mode",
			text:     "```go\nfunc main() {}\n```",
			wantMode: "code",
		},
		{
			name:     "import statement triggers code mode",
			text:     `import "fmt"\nfmt.Println("hello")`,
			wantMode: "code",
		},
		{
			name:     "markdown table triggers data mode",
			text:     "| Name | Role | Status |\n|------|------|--------|\n| Ed | Partner | Active |",
			wantMode: "data",
		},
		{
			name:     "plain prose defaults to linguistic",
			text:     "We discussed the pricing strategy for the AI Masters course.",
			wantMode: "linguistic",
		},
		{
			name:     "image markdown triggers visual mode",
			text:     "Here is the diagram: ![architecture](diagram.png)",
			wantMode: "visual",
		},
		{
			name:     "empty text defaults to linguistic",
			text:     "",
			wantMode: "linguistic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.text)
			if got.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", got.Mode, tt.wantMode)
			}
		})
	}
}

func TestClassify_Genre(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantGenre string
	}{
		{
			name:      "transcript: attendees keyword",
			text:      "Attendees: Ed, Robert, Roberto. We discussed the pricing model. Action items: send proposal by Friday.",
			wantGenre: "transcript",
		},
		{
			name:      "spec: requirements keyword",
			text:      "Requirements:\n1. The system must ingest signals\n2. Acceptance criteria: search returns results in <500ms",
			wantGenre: "spec",
		},
		{
			name:      "brief: objective and key messages",
			text:      "Objective: Close the AI Masters deal.\nKey Messages:\n- Course is ready\n- Call to action: sign by EOD",
			wantGenre: "brief",
		},
		{
			name:      "plan: non-negotiables keyword",
			text:      "Non-Negotiables:\n1. Ship the MVP\nSuccess Criteria: 10 paying users",
			wantGenre: "plan",
		},
		{
			name:      "decision-log: decided by keyword",
			text:      "Decision: Go with the $2K pricing.\nRationale: matches market.\nDecided by: Roberto",
			wantGenre: "decision-log",
		},
		{
			name:      "standup: yesterday/today/blockers",
			text:      "Yesterday: finished the graph module.\nToday: writing tests.\nBlockers: none",
			wantGenre: "standup",
		},
		{
			name:      "review: code review markers",
			text:      "Code review for PR #42.\nOverall: APPROVED\nIssues found: none significant.",
			wantGenre: "review",
		},
		{
			name:      "pitch: market size and ask",
			text:      "Problem statement: agencies waste 10h/week.\nMarket size: $2B.\nAsk: $250K seed round.",
			wantGenre: "pitch",
		},
		{
			name:      "note: default fallback",
			text:      "Roberto mentioned that Ed called about the course launch.",
			wantGenre: "note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.text)
			if got.Genre != tt.wantGenre {
				t.Errorf("Genre = %q, want %q (text: %.60s...)", got.Genre, tt.wantGenre, tt.text)
			}
		})
	}
}

func TestClassify_Type(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantType string
	}{
		{
			name:     "imperative first word → direct",
			text:     "Deploy the new engine to production.",
			wantType: "direct",
		},
		{
			name:     "decision language → decide",
			text:     "We decided to go with the $2K pricing model.",
			wantType: "decide",
		},
		{
			name:     "deadline/promise → commit",
			text:     "I will deliver the spec by Wednesday.",
			wantType: "commit",
		},
		{
			name:     "feeling/opinion → express",
			text:     "I feel frustrated with the slow progress on the frontend.",
			wantType: "express",
		},
		{
			name:     "question → inform",
			text:     "What is the status of the AI Masters launch?",
			wantType: "inform",
		},
		{
			name:     "neutral statement → inform",
			text:     "The engine currently has 47 Elixir modules.",
			wantType: "inform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.text)
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
		})
	}
}

func TestClassify_Format(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantFormat string
	}{
		{
			name:       "YAML frontmatter",
			text:       "---\ngenre: transcript\ndate: 2026-04-11\n---\n\nContent here.",
			wantFormat: "yaml",
		},
		{
			name:       "JSON object",
			text:       `{"key": "value", "count": 42}`,
			wantFormat: "json",
		},
		{
			name:       "JSON array",
			text:       `[{"name": "Ed"}, {"name": "Robert"}]`,
			wantFormat: "json",
		},
		{
			name:       "markdown with headers",
			text:       "# Title\n\n## Section\n\nContent here.",
			wantFormat: "markdown",
		},
		{
			name:       "plain text → unknown",
			text:       "Just a plain sentence with no special formatting.",
			wantFormat: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.text)
			if got.Format != tt.wantFormat {
				t.Errorf("Format = %q, want %q", got.Format, tt.wantFormat)
			}
		})
	}
}

func TestClassify_SNRatio(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantMinSN float64
		wantMaxSN float64
	}{
		{
			name:      "structured spec has high S/N",
			text:      "# Spec\n\n## Requirements\n- Must ingest signals\n- Must search in <500ms\n\n## Acceptance Criteria\n- [ ] All tests pass\n\nDecided by: Roberto on 2026-04-11",
			wantMinSN: 0.8,
			wantMaxSN: 1.0,
		},
		{
			name:      "empty text has zero S/N",
			text:      "",
			wantMinSN: 0.0,
			wantMaxSN: 0.0,
		},
		{
			name:      "filler-heavy text has lower S/N than base",
			text:      "Like, I just basically kind of feel like you know it's sort of fine um",
			wantMinSN: 0.0,
			wantMaxSN: 0.6,
		},
		{
			name:      "plain note gets base S/N",
			text:      "Roberto mentioned Ed called.",
			wantMinSN: 0.4,
			wantMaxSN: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.text)
			if got.SNRatio < tt.wantMinSN || got.SNRatio > tt.wantMaxSN {
				t.Errorf("SNRatio = %.2f, want [%.2f, %.2f]",
					got.SNRatio, tt.wantMinSN, tt.wantMaxSN)
			}
		})
	}
}

func TestClassify_Structure(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		wantStructure string
	}{
		{
			name:          "spec structure label",
			text:          "Requirements:\n1. Must work\nAcceptance criteria: passes tests",
			wantStructure: "spec/goal-requirements-constraints-architecture-acceptance",
		},
		{
			name:          "note structure label",
			text:          "Ed mentioned pricing.",
			wantStructure: "note/context-content-route",
		},
		{
			name:          "transcript structure label",
			text:          "Attendees: Ed, Roberto. Action items: send email.",
			wantStructure: "transcript/participants-points-decisions-actions-questions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.text)
			if got.Structure != tt.wantStructure {
				t.Errorf("Structure = %q, want %q", got.Structure, tt.wantStructure)
			}
		})
	}
}

func TestClassify_EmptyText(t *testing.T) {
	got := Classify("")
	if got.Mode != "linguistic" {
		t.Errorf("empty Mode = %q, want %q", got.Mode, "linguistic")
	}
	if got.Genre != "note" {
		t.Errorf("empty Genre = %q, want %q", got.Genre, "note")
	}
	if got.SNRatio != 0.0 {
		t.Errorf("empty SNRatio = %.2f, want 0.0", got.SNRatio)
	}
}
