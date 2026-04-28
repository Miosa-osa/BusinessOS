package bridge

import (
	"context"
	"sync"
	"time"
)

// Memory is the bridge to the episodic / SICA / VIGIL / Cortex layer.
//
// Those subsystems exist as Elixir GenServers in the source repo and aren't
// ported to Go yet (they sit in Phase 8+ of the BusinessOS roadmap). This
// bridge gives callers a stable API that:
//
//  1. Records events into an in-process ring buffer the rest of the engine
//     can consume immediately (audit trail, debugging).
//  2. Routes every event to optional callback hooks so a future Go-native
//     SICA/VIGIL/Cortex implementation can subscribe without changing call
//     sites.
//  3. Defaults to no-op when no hook is registered — calling code never has
//     to check whether the learning loop is wired.
type Memory struct {
	mu     sync.Mutex
	events []Event
	cap    int

	OnEvent      func(ctx context.Context, e Event)
	OnMutation   func(ctx context.Context, m Mutation)
	OnCorrection func(ctx context.Context, original, corrected string)
	OnError      func(ctx context.Context, tool string, err string, meta map[string]any)
}

// EventKind enumerates the temporal event types Episodic memory tracks.
// Strings match the Elixir source so cross-system metric labels line up.
type EventKind string

const (
	EventIntake   EventKind = "intake"
	EventSearch   EventKind = "search"
	EventMutation EventKind = "mutation"
	EventDecision EventKind = "decision"
	EventError    EventKind = "error"
	EventToolCall EventKind = "tool_call"
)

// Event is one episodic record. Data is whatever the caller wants to log;
// kept as map[string]any so the same struct fits every kind.
type Event struct {
	Kind      EventKind
	SessionID string
	Data      map[string]any
	At        time.Time
}

// Mutation is the shape SICA observes. Fields match the Elixir
// observe_mutation/1 contract for forward compatibility.
type Mutation struct {
	Tool    string
	Input   any
	Output  any
	Success bool
}

// DefaultBufferSize bounds the in-process episodic ring buffer. 4096 keeps
// the last few hours of activity in memory under typical load.
const DefaultBufferSize = 4096

// NewMemory returns a Memory bridge with the default buffer size.
func NewMemory() *Memory { return &Memory{cap: DefaultBufferSize} }

// RecordEvent appends to the ring buffer and fires OnEvent. Both are
// best-effort: errors from the hook are swallowed so episodic recording
// can never crash a write path.
func (m *Memory) RecordEvent(ctx context.Context, kind EventKind, data map[string]any, sessionID string) {
	if sessionID == "" {
		sessionID = "optimal-engine"
	}
	e := Event{Kind: kind, SessionID: sessionID, Data: data, At: time.Now().UTC()}

	m.mu.Lock()
	if m.cap == 0 {
		m.cap = DefaultBufferSize
	}
	if len(m.events) >= m.cap {
		// Drop oldest to keep the buffer bounded.
		m.events = m.events[1:]
	}
	m.events = append(m.events, e)
	hook := m.OnEvent
	m.mu.Unlock()

	if hook != nil {
		safeCallEvent(hook, ctx, e)
	}
}

// RecallEvents returns events matching kind/session, newest first. Bounded
// by limit; pass 0 for all.
func (m *Memory) RecallEvents(kind EventKind, sessionID string, limit int) []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []Event{}
	for i := len(m.events) - 1; i >= 0; i-- {
		e := m.events[i]
		if kind != "" && e.Kind != kind {
			continue
		}
		if sessionID != "" && e.SessionID != sessionID {
			continue
		}
		out = append(out, e)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

// ObserveMutation routes to the SICA hook. Useful for pattern detection in
// "this tool was called with X and returned Y, success: bool" interactions.
func (m *Memory) ObserveMutation(ctx context.Context, mut Mutation) {
	hook := m.OnMutation
	if hook != nil {
		safeCallMutation(hook, ctx, mut)
	}
	m.RecordEvent(ctx, EventMutation, map[string]any{
		"tool":    mut.Tool,
		"success": mut.Success,
	}, "")
}

// RecordCorrection routes to the SICA correction hook for supervised
// learning ("the user changed our output from X to Y").
func (m *Memory) RecordCorrection(ctx context.Context, original, corrected string) {
	hook := m.OnCorrection
	if hook != nil {
		safeCallCorrection(hook, ctx, original, corrected)
	}
	m.RecordEvent(ctx, EventDecision, map[string]any{
		"original":  original,
		"corrected": corrected,
	}, "")
}

// RecordError routes to the VIGIL recovery hook ("tool X failed with Y").
func (m *Memory) RecordError(ctx context.Context, tool, errMsg string, meta map[string]any) {
	hook := m.OnError
	if hook != nil {
		safeCallError(hook, ctx, tool, errMsg, meta)
	}
	m.RecordEvent(ctx, EventError, map[string]any{
		"tool":  tool,
		"error": errMsg,
		"meta":  meta,
	}, "")
}

// LearningMetrics is the API surface the Elixir source exposes. Until
// SICA lands in Go, the metrics are derived from the in-process ring
// buffer so existing callers keep working.
type LearningMetrics struct {
	Interactions int
	Patterns     int
	Skills       int
	Errors       int
}

// Metrics returns counts derived from the buffer. SICA's pattern/skill
// counts default to zero until the learning loop is wired.
func (m *Memory) Metrics() LearningMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := LearningMetrics{Interactions: len(m.events)}
	for _, e := range m.events {
		if e.Kind == EventError {
			out.Errors++
		}
	}
	return out
}

// Recover-from-panic helpers. The Elixir source uses `rescue _ -> :ok` so
// observability hooks can never break business logic; we replicate that
// guarantee here.
func safeCallEvent(fn func(context.Context, Event), ctx context.Context, e Event) {
	defer func() { _ = recover() }()
	fn(ctx, e)
}
func safeCallMutation(fn func(context.Context, Mutation), ctx context.Context, m Mutation) {
	defer func() { _ = recover() }()
	fn(ctx, m)
}
func safeCallCorrection(fn func(context.Context, string, string), ctx context.Context, a, b string) {
	defer func() { _ = recover() }()
	fn(ctx, a, b)
}
func safeCallError(fn func(context.Context, string, string, map[string]any), ctx context.Context, tool, msg string, meta map[string]any) {
	defer func() { _ = recover() }()
	fn(ctx, tool, msg, meta)
}
