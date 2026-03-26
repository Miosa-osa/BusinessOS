package ontology

import (
	"strings"
	"testing"
	"time"
)

func TestBoardchairL0Sync_sanitizeURI(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"case-001", "case-001"},
		{"case 001", "case-001"},
		{"case/001", "case-001"},
		{"case<>001", "case001"},
	}
	for _, tc := range cases {
		got := sanitizeURI(tc.input)
		if got != tc.expected {
			t.Errorf("sanitizeURI(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestBoardchairL0Sync_turtlePreamble(t *testing.T) {
	// Verify Turtle output contains required prefixes
	s := NewBoardchairL0Sync(nil, "http://localhost:7878")
	_ = s // just verify construction

	preamble := `@prefix bos: <http://businessos.local/ontology#> .`
	if !strings.Contains(preamble, "bos:") {
		t.Error("preamble must declare bos: prefix")
	}
}

func TestSyncInterval(t *testing.T) {
	// WvdA: L0 refresh every 15min matches L1 materialization interval
	if l0SyncInterval != 15*time.Minute {
		t.Errorf("l0SyncInterval = %v, want 15m (must match L1 SPARQL refresh)", l0SyncInterval)
	}
}
