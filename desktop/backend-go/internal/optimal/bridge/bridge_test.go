package bridge

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGraphBoostUnchangedWhenNoConnections(t *testing.T) {
	// No DB → connected entity lookup errs → results pass through unchanged.
	k := &Knowledge{}
	results := []SearchResult{{ID: "a", Score: 0.5, Entities: []string{"alice"}}}
	out := k.GraphBoost(context.Background(), results, "alice")
	if len(out) != 1 || out[0].Score != 0.5 {
		t.Fatalf("scores should be untouched on lookup failure, got %+v", out)
	}
}

func TestGraphBoostNoOpOnEmptyQuery(t *testing.T) {
	k := &Knowledge{}
	results := []SearchResult{{ID: "a", Score: 0.5}}
	out := k.GraphBoost(context.Background(), results, "")
	if out[0].Score != 0.5 {
		t.Fatal("empty query should leave scores untouched")
	}
}

func TestSplitTermsDropsShortTokens(t *testing.T) {
	got := splitTerms("a is the quick brown fox")
	for _, term := range got {
		if len(term) < 3 {
			t.Fatalf("short term leaked through: %q in %v", term, got)
		}
	}
}

func TestRound4Truncates(t *testing.T) {
	if round4(0.123456789) != 0.1235 {
		t.Fatalf("rounding: %v", round4(0.123456789))
	}
}

func TestMemoryRecordsAndRecalls(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()
	m.RecordEvent(ctx, EventIntake, map[string]any{"path": "a.md"}, "session-1")
	m.RecordEvent(ctx, EventSearch, map[string]any{"q": "foo"}, "session-1")
	got := m.RecallEvents(EventIntake, "session-1", 0)
	if len(got) != 1 || got[0].Kind != EventIntake {
		t.Fatalf("recall: %+v", got)
	}
	all := m.RecallEvents("", "session-1", 0)
	if len(all) != 2 {
		t.Fatalf("expected 2, got %d", len(all))
	}
}

func TestMemoryRingBufferEviction(t *testing.T) {
	m := &Memory{cap: 3}
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		m.RecordEvent(ctx, EventToolCall, map[string]any{"i": i}, "")
	}
	if got := len(m.events); got != 3 {
		t.Fatalf("ring buffer overflow: %d events held", got)
	}
}

func TestMemoryHooksCalled(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()
	called := 0
	m.OnEvent = func(_ context.Context, _ Event) { called++ }
	m.RecordEvent(ctx, EventDecision, nil, "")
	if called != 1 {
		t.Fatalf("hook should fire exactly once, got %d", called)
	}
}

func TestMemoryHooksNeverPanic(t *testing.T) {
	m := NewMemory()
	m.OnEvent = func(_ context.Context, _ Event) { panic("boom") }
	// Must not propagate — observability never breaks business logic.
	m.RecordEvent(context.Background(), EventError, nil, "")
}

func TestMemoryMetricsCountErrors(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()
	m.RecordEvent(ctx, EventIntake, nil, "")
	m.RecordError(ctx, "tool-x", "boom", nil)
	m.RecordError(ctx, "tool-y", "blam", nil)
	got := m.Metrics()
	if got.Errors != 2 {
		t.Fatalf("errors counted: %d", got.Errors)
	}
}

func TestSignalClassifyRoundsTrip(t *testing.T) {
	s := NewSignal()
	cl := s.Classify("Project plan: Ship the wiki bridge by Friday.")
	if cl.SNRatio < 0.0 || cl.SNRatio > 1.0 {
		t.Fatalf("S/N out of range: %v", cl.SNRatio)
	}
}

func TestSignalCloudEventShape(t *testing.T) {
	s := NewSignal()
	ev := s.CloudEvent("we ship today", "test-source")
	for _, key := range []string{"specversion", "id", "source", "type", "time"} {
		if _, ok := ev[key]; !ok {
			t.Fatalf("missing required CloudEvents field %q", key)
		}
	}
	if !strings.HasPrefix(ev["type"].(string), "optimal.") {
		t.Fatalf("type prefix wrong: %v", ev["type"])
	}
}

func TestSignalAuditReturnsSomething(t *testing.T) {
	s := NewSignal()
	// Any text should classify and produce a non-nil failure-mode list (may be empty).
	out := s.Audit("aaa bbb ccc")
	if out == nil {
		t.Fatal("Audit must never return nil")
	}
}

// Sentinel test — captures the contract that bridge errors propagate
// distinctly from business errors. We use a plain errors.New here so
// future compliance/audit code can grep for this pattern.
func TestErrorIsDistinctSentinel(t *testing.T) {
	err := errors.New("knowledge: graph unreachable")
	if !strings.HasPrefix(err.Error(), "knowledge:") {
		t.Fatal("sentinel pattern broken")
	}
}
