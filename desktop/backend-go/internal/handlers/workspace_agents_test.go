package handlers

import "testing"

func TestNormalizeAgentRuntime(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "default", in: "", want: "osa"},
		{name: "osa", in: "osa", want: "osa"},
		{name: "claude code", in: " Claude-Code ", want: "claude-code"},
		{name: "codex", in: "CODEX", want: "codex"},
		{name: "hermes", in: "hermes", want: "hermes"},
		{name: "unknown", in: "fake-runtime", want: "osa"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeAgentRuntime(tt.in); got != tt.want {
				t.Fatalf("normalizeAgentRuntime(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
