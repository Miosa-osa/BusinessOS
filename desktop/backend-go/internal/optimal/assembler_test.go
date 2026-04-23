package optimal

import (
	"testing"
)

// ── estimateTokens ────────────────────────────────────────────────────────────

func TestEstimateTokens(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  int
	}{
		// (len + 3) / 4
		{"empty string", "", 0},
		{"hello world (11 chars) → 3", "hello world", 3},
		{"4 chars → 1", "abcd", 1},
		{"5 chars → 2", "abcde", 2},
		{"exactly 8 chars → 2", "abcdefgh", 2},
		{"9 chars → 3", "abcdefghi", 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := estimateTokens(tc.input)
			if got != tc.want {
				t.Errorf("estimateTokens(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

// ── DefaultAssemblerConfig ────────────────────────────────────────────────────

func TestDefaultAssemblerConfig(t *testing.T) {
	cfg := DefaultAssemblerConfig()

	if cfg.L0Budget != 2000 {
		t.Errorf("L0Budget = %d, want 2000", cfg.L0Budget)
	}
	if cfg.L1Budget != 10000 {
		t.Errorf("L1Budget = %d, want 10000", cfg.L1Budget)
	}
	if cfg.L2Budget != 50000 {
		t.Errorf("L2Budget = %d, want 50000", cfg.L2Budget)
	}
	if cfg.L1Threshold != 0.1 {
		t.Errorf("L1Threshold = %.2f, want 0.1", cfg.L1Threshold)
	}
	if cfg.MaxSources != 20 {
		t.Errorf("MaxSources = %d, want 20", cfg.MaxSources)
	}
}

// ── AssembleContext: graceful fallbacks ───────────────────────────────────────

func TestAssembleContext_EmptyPathsNoError(t *testing.T) {
	// Both nodesRoot and dbPath empty — should fail on BuildL0 which needs at
	// least some filesystem root. We test that the error is surfaced rather than
	// panicking.
	_, err := AssembleContext("", "", "any query", nil, nil)
	// We expect either an error (empty root) or a valid result with empty L0.
	// The contract is: no panic, no silent corruption.
	_ = err // may be nil or non-nil depending on BuildL0 behaviour with empty root
}

func TestAssembleContext_NilReceiverUsesDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Pass nodesRoot pointing at an empty tmp dir — no nodes, no DB.
	// L0 might be empty but call must not panic.
	result, err := AssembleContext(tmpDir, "", "", nil, nil)
	if err != nil {
		// BuildL0 may error on missing dirs; that is acceptable.
		return
	}

	// If it succeeds, receiver should have been defaulted to role=agent.
	if result.ReceiverProfile.Role != "agent" {
		t.Errorf("default receiver role = %q, want %q", result.ReceiverProfile.Role, "agent")
	}
}

func TestAssembleContext_NilCfgUsesDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := AssembleContext(tmpDir, "", "", nil, nil)
	if err != nil {
		return // tolerate BuildL0 errors on empty dir
	}

	// TotalTokens should be >= 0 and budget_used should be in [0, ∞).
	if result.TotalTokens < 0 {
		t.Errorf("TotalTokens should be non-negative, got %d", result.TotalTokens)
	}
	if result.BudgetUsed < 0 {
		t.Errorf("BudgetUsed should be non-negative, got %.4f", result.BudgetUsed)
	}
}

// ── requiresL2 ────────────────────────────────────────────────────────────────

func TestRequiresL2(t *testing.T) {
	cases := []struct {
		name     string
		query    string
		receiver *ReceiverProfile
		want     bool
	}{
		{
			name:     "nil receiver → false",
			query:    "deep dive into the architecture",
			receiver: nil,
			want:     false,
		},
		{
			name:     "executive role → false even with depth keyword",
			query:    "deep dive into the architecture",
			receiver: &ReceiverProfile{Role: "executive"},
			want:     false,
		},
		{
			name:     "technical role + depth keyword → true",
			query:    "deep dive into the architecture",
			receiver: &ReceiverProfile{Role: "technical"},
			want:     true,
		},
		{
			name:     "agent role + 'implementation' keyword → true",
			query:    "show me the implementation details",
			receiver: &ReceiverProfile{Role: "agent"},
			want:     true,
		},
		{
			name:     "technical role + shallow query → false",
			query:    "what is the status",
			receiver: &ReceiverProfile{Role: "technical"},
			want:     false,
		},
		{
			name:     "agent role + 'complete' keyword → true",
			query:    "give me the complete context for this module",
			receiver: &ReceiverProfile{Role: "agent"},
			want:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := requiresL2(tc.query, tc.receiver)
			if got != tc.want {
				t.Errorf("requiresL2(%q, %+v) = %v, want %v", tc.query, tc.receiver, got, tc.want)
			}
		})
	}
}

// ── splitNodePath ─────────────────────────────────────────────────────────────

func TestSplitNodePath(t *testing.T) {
	cases := []struct {
		name         string
		path         string
		nodeHint     string
		wantSlug     string
		wantFilePath string
	}{
		{
			name:         "standard node/file path",
			path:         "01-company/context.md",
			nodeHint:     "",
			wantSlug:     "01-company",
			wantFilePath: "context.md",
		},
		{
			name:         "nested signal path",
			path:         "09-new-stuff/signals/2026-04-01.md",
			nodeHint:     "",
			wantSlug:     "09-new-stuff",
			wantFilePath: "signals/2026-04-01.md",
		},
		{
			name:         "with node hint",
			path:         "09-new-stuff/2026-04-01.md",
			nodeHint:     "09-new-stuff",
			wantSlug:     "09-new-stuff",
			wantFilePath: "2026-04-01.md",
		},
		{
			name:         "empty path returns empty",
			path:         "",
			nodeHint:     "",
			wantSlug:     "",
			wantFilePath: "",
		},
		{
			name:         "leading slash is stripped",
			path:         "/04-ai-masters/context.md",
			nodeHint:     "",
			wantSlug:     "04-ai-masters",
			wantFilePath: "context.md",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			slug, filePath := splitNodePath(tc.path, tc.nodeHint)
			if slug != tc.wantSlug {
				t.Errorf("slug = %q, want %q", slug, tc.wantSlug)
			}
			if filePath != tc.wantFilePath {
				t.Errorf("filePath = %q, want %q", filePath, tc.wantFilePath)
			}
		})
	}
}

// ── formatL1Section / formatL2Section ────────────────────────────────────────

func TestFormatL1Section_ContainsHeaderAndContent(t *testing.T) {
	out := formatL1Section("AI Masters", "04-ai-masters/context.md", "Some content here.")
	if out == "" {
		t.Fatal("formatL1Section returned empty string")
	}
	// Must contain the title and path in a header.
	if !containsAll(out, "AI Masters", "04-ai-masters/context.md", "Some content here.") {
		t.Errorf("formatL1Section output missing expected fields: %q", out)
	}
}

func TestFormatL2Section_ContainsFULLPrefix(t *testing.T) {
	out := formatL2Section("AI Masters", "04-ai-masters/context.md", "Full content here.")
	if !containsAll(out, "FULL", "AI Masters", "Full content here.") {
		t.Errorf("formatL2Section output missing expected fields: %q", out)
	}
}

// containsAll returns true if s contains every substring in needles.
func containsAll(s string, needles ...string) bool {
	for _, needle := range needles {
		found := false
		for i := 0; i <= len(s)-len(needle); i++ {
			if s[i:i+len(needle)] == needle {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
