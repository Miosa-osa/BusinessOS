package models

import (
	"encoding/json"
	"testing"
	"time"
)

// ============================================================================
// Event Tests
// ============================================================================

func TestEventValidation_Valid(t *testing.T) {
	ts := time.Now().UTC().Format(time.RFC3339)
	event := &Event{
		Activity:   "TestActivity",
		Timestamp:  ts,
		Attributes: make(map[string]string),
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	if err := event.Validate(); err != nil {
		t.Fatalf("Valid event failed validation: %v", err)
	}
}

func TestEventValidation_EmptyActivity(t *testing.T) {
	event := &Event{
		Activity:   "",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Attributes: make(map[string]string),
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	err := event.Validate()
	if err == nil {
		t.Fatal("Expected error for empty activity, got nil")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrInvalidActivity {
		t.Fatalf("Expected ErrInvalidActivity, got %v", err)
	}
}

func TestEventValidation_ActivityTooLong(t *testing.T) {
	longActivity := ""
	for i := 0; i < 256; i++ {
		longActivity += "a"
	}

	event := &Event{
		Activity:   longActivity,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Attributes: make(map[string]string),
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	err := event.Validate()
	if err == nil {
		t.Fatal("Expected error for activity exceeding 255 chars")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrInvalidActivity {
		t.Fatalf("Expected ErrInvalidActivity, got %v", err)
	}
}

func TestEventValidation_FutureTimestamp(t *testing.T) {
	futureTs := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	event := &Event{
		Activity:   "TestActivity",
		Timestamp:  futureTs,
		Attributes: make(map[string]string),
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	err := event.Validate()
	if err == nil {
		t.Fatal("Expected error for future timestamp")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrInvalidTimestamp {
		t.Fatalf("Expected ErrInvalidTimestamp, got %v", err)
	}
}

func TestEventValidation_InvalidUUID(t *testing.T) {
	event := &Event{
		Activity:   "TestActivity",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Attributes: make(map[string]string),
		ID:         "not-a-valid-uuid",
	}

	err := event.Validate()
	if err == nil {
		t.Fatal("Expected error for invalid UUID")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrUUID {
		t.Fatalf("Expected ErrUUID, got %v", err)
	}
}

func TestEventValidation_TooManyAttributes(t *testing.T) {
	attrs := make(map[string]string)
	for i := 0; i < 51; i++ {
		attrs["key"+string(rune(i))] = "value"
	}

	event := &Event{
		Activity:   "TestActivity",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Attributes: attrs,
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	err := event.Validate()
	if err == nil {
		t.Fatal("Expected error for too many attributes")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrMaxSizeExceeded {
		t.Fatalf("Expected ErrMaxSizeExceeded, got %v", err)
	}
}

func TestParseEvent_Valid(t *testing.T) {
	eventJSON := `{
		"activity": "TestActivity",
		"timestamp": "2026-03-24T14:30:45Z",
		"resource": "Agent1",
		"attributes": {"key": "value"},
		"id": "550e8400-e29b-41d4-a716-446655440000"
	}`

	event, err := ParseEvent([]byte(eventJSON))
	if err != nil {
		t.Fatalf("Failed to parse valid event: %v", err)
	}

	if event.Activity != "TestActivity" {
		t.Errorf("Expected activity 'TestActivity', got %s", event.Activity)
	}
	if event.Resource == nil || *event.Resource != "Agent1" {
		t.Errorf("Expected resource 'Agent1', got %v", event.Resource)
	}
}

// ============================================================================
// Trace Tests
// ============================================================================

func TestTraceValidation_Valid(t *testing.T) {
	trace := &Trace{
		ID: "case-1",
		Events: []Event{
			{
				Activity:   "Activity1",
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Attributes: make(map[string]string),
				ID:         "550e8400-e29b-41d4-a716-446655440000",
			},
		},
		Attributes: make(map[string]string),
	}

	if err := trace.Validate(); err != nil {
		t.Fatalf("Valid trace failed validation: %v", err)
	}
}

func TestTraceValidation_EmptyTraceID(t *testing.T) {
	trace := &Trace{
		ID: "",
		Events: []Event{
			{
				Activity:   "Activity1",
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Attributes: make(map[string]string),
				ID:         "550e8400-e29b-41d4-a716-446655440000",
			},
		},
		Attributes: make(map[string]string),
	}

	err := trace.Validate()
	if err == nil {
		t.Fatal("Expected error for empty trace ID")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrInvalidTraceID {
		t.Fatalf("Expected ErrInvalidTraceID, got %v", err)
	}
}

func TestTraceValidation_NoEvents(t *testing.T) {
	trace := &Trace{
		ID:         "case-1",
		Events:     []Event{},
		Attributes: make(map[string]string),
	}

	err := trace.Validate()
	if err == nil {
		t.Fatal("Expected error for trace with no events")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrEmptyCollection {
		t.Fatalf("Expected ErrEmptyCollection, got %v", err)
	}
}

func TestParseTrace_Valid(t *testing.T) {
	traceJSON := `{
		"id": "case-1",
		"events": [
			{
				"activity": "Activity1",
				"timestamp": "2026-03-24T14:30:45Z",
				"attributes": {},
				"id": "550e8400-e29b-41d4-a716-446655440000"
			}
		],
		"attributes": {}
	}`

	trace, err := ParseTrace([]byte(traceJSON))
	if err != nil {
		t.Fatalf("Failed to parse valid trace: %v", err)
	}

	if trace.ID != "case-1" {
		t.Errorf("Expected trace ID 'case-1', got %s", trace.ID)
	}
	if len(trace.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(trace.Events))
	}
}

// ============================================================================
// EventLog Tests
// ============================================================================

func TestEventLogValidation_Valid(t *testing.T) {
	log := &EventLog{
		Traces: []Trace{
			{
				ID: "case-1",
				Events: []Event{
					{
						Activity:   "Activity1",
						Timestamp:  time.Now().UTC().Format(time.RFC3339),
						Attributes: make(map[string]string),
						ID:         "550e8400-e29b-41d4-a716-446655440000",
					},
				},
				Attributes: make(map[string]string),
			},
		},
		Attributes: make(map[string]string),
	}

	if err := log.Validate(); err != nil {
		t.Fatalf("Valid event log failed validation: %v", err)
	}
}

func TestEventLogValidation_NoTraces(t *testing.T) {
	log := &EventLog{
		Traces:     []Trace{},
		Attributes: make(map[string]string),
	}

	err := log.Validate()
	if err == nil {
		t.Fatal("Expected error for event log with no traces")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrEmptyCollection {
		t.Fatalf("Expected ErrEmptyCollection, got %v", err)
	}
}

func TestEventLogValidation_DuplicateTraceIDs(t *testing.T) {
	log := &EventLog{
		Traces: []Trace{
			{
				ID: "case-1",
				Events: []Event{
					{
						Activity:   "Activity1",
						Timestamp:  time.Now().UTC().Format(time.RFC3339),
						Attributes: make(map[string]string),
						ID:         "550e8400-e29b-41d4-a716-446655440000",
					},
				},
				Attributes: make(map[string]string),
			},
			{
				ID: "case-1",
				Events: []Event{
					{
						Activity:   "Activity2",
						Timestamp:  time.Now().UTC().Format(time.RFC3339),
						Attributes: make(map[string]string),
						ID:         "550e8400-e29b-41d4-a716-446655440001",
					},
				},
				Attributes: make(map[string]string),
			},
		},
		Attributes: make(map[string]string),
	}

	err := log.Validate()
	if err == nil {
		t.Fatal("Expected error for duplicate trace IDs")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrValidationFailed {
		t.Fatalf("Expected ErrValidationFailed, got %v", err)
	}
}

func TestParseEventLog_Valid(t *testing.T) {
	logJSON := `{
		"traces": [
			{
				"id": "case-1",
				"events": [
					{
						"activity": "Activity1",
						"timestamp": "2026-03-24T14:30:45Z",
						"attributes": {},
						"id": "550e8400-e29b-41d4-a716-446655440000"
					}
				],
				"attributes": {}
			}
		],
		"attributes": {}
	}`

	log, err := ParseEventLog([]byte(logJSON))
	if err != nil {
		t.Fatalf("Failed to parse valid event log: %v", err)
	}

	if len(log.Traces) != 1 {
		t.Errorf("Expected 1 trace, got %d", len(log.Traces))
	}

	numEvents := log.NumEvents()
	if numEvents != 1 {
		t.Errorf("Expected 1 total event, got %d", numEvents)
	}
}

func TestEventLogNumEvents(t *testing.T) {
	log := &EventLog{
		Traces: []Trace{
			{
				ID: "case-1",
				Events: []Event{
					{Activity: "A", Timestamp: time.Now().UTC().Format(time.RFC3339), ID: "1"},
					{Activity: "B", Timestamp: time.Now().UTC().Format(time.RFC3339), ID: "2"},
				},
				Attributes: make(map[string]string),
			},
			{
				ID: "case-2",
				Events: []Event{
					{Activity: "A", Timestamp: time.Now().UTC().Format(time.RFC3339), ID: "3"},
				},
				Attributes: make(map[string]string),
			},
		},
		Attributes: make(map[string]string),
	}

	if log.NumEvents() != 3 {
		t.Errorf("Expected 3 total events, got %d", log.NumEvents())
	}
}

// ============================================================================
// PetriNet Tests
// ============================================================================

func TestPetriNetValidation_Valid(t *testing.T) {
	net := &PetriNet{
		Places: []Place{
			{ID: "p1", Name: "Start", InitialMarking: 1, FinalMarking: nil},
			{ID: "p2", Name: "End", InitialMarking: 0, FinalMarking: intPtr(1)},
		},
		Transitions: []Transition{
			{ID: "t1", Name: "Process", Label: nil},
		},
		Arcs: []Arc{
			{From: "p1", To: "t1", Weight: 1},
			{From: "t1", To: "p2", Weight: 1},
		},
		InitialPlace: stringPtr("p1"),
		FinalPlace:   stringPtr("p2"),
	}

	if err := net.Validate(); err != nil {
		t.Fatalf("Valid Petri net failed validation: %v", err)
	}
}

func TestPetriNetValidation_NoPlaces(t *testing.T) {
	net := &PetriNet{
		Places: []Place{},
		Transitions: []Transition{
			{ID: "t1", Name: "Process", Label: nil},
		},
		Arcs: []Arc{},
	}

	err := net.Validate()
	if err == nil {
		t.Fatal("Expected error for Petri net with no places")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrEmptyCollection {
		t.Fatalf("Expected ErrEmptyCollection, got %v", err)
	}
}

func TestPetriNetValidation_InvalidArcReference(t *testing.T) {
	net := &PetriNet{
		Places: []Place{
			{ID: "p1", Name: "Start", InitialMarking: 1, FinalMarking: nil},
		},
		Transitions: []Transition{
			{ID: "t1", Name: "Process", Label: nil},
		},
		Arcs: []Arc{
			{From: "p1", To: "nonexistent", Weight: 1},
		},
	}

	err := net.Validate()
	if err == nil {
		t.Fatal("Expected error for invalid arc reference")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrReferenceNotFound {
		t.Fatalf("Expected ErrReferenceNotFound, got %v", err)
	}
}

func TestParsePetriNet_Valid(t *testing.T) {
	netJSON := `{
		"places": [
			{"id": "p1", "name": "Start", "initial_marking": 1, "final_marking": null},
			{"id": "p2", "name": "End", "initial_marking": 0, "final_marking": 1}
		],
		"transitions": [
			{"id": "t1", "name": "Process", "label": null}
		],
		"arcs": [
			{"from": "p1", "to": "t1", "weight": 1},
			{"from": "t1", "to": "p2", "weight": 1}
		],
		"initial_place": "p1",
		"final_place": "p2"
	}`

	net, err := ParsePetriNet([]byte(netJSON))
	if err != nil {
		t.Fatalf("Failed to parse valid Petri net: %v", err)
	}

	if len(net.Places) != 2 {
		t.Errorf("Expected 2 places, got %d", len(net.Places))
	}
	if len(net.Transitions) != 1 {
		t.Errorf("Expected 1 transition, got %d", len(net.Transitions))
	}
}

func TestPetriNetString(t *testing.T) {
	net := &PetriNet{
		Places: []Place{
			{ID: "p1", Name: "Start", InitialMarking: 1, FinalMarking: nil},
		},
		Transitions: []Transition{
			{ID: "t1", Name: "Process", Label: nil},
		},
		Arcs: []Arc{
			{From: "p1", To: "t1", Weight: 1},
		},
	}

	str := net.String()
	if str == "" {
		t.Fatal("Expected non-empty string representation")
	}
}

// ============================================================================
// Round-Trip Tests
// ============================================================================

func TestEventRoundTrip(t *testing.T) {
	original := &Event{
		Activity:   "TestActivity",
		Timestamp:  "2026-03-24T14:30:45.123456Z",
		Resource:   stringPtr("Agent1"),
		Attributes: map[string]string{"key": "value"},
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	// Marshal
	jsonBytes, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Unmarshal
	restored, err := ParseEvent(jsonBytes)
	if err != nil {
		t.Fatalf("Failed to parse event: %v", err)
	}

	// Compare
	if restored.Activity != original.Activity {
		t.Errorf("Activity mismatch: %s != %s", restored.Activity, original.Activity)
	}
	if restored.ID != original.ID {
		t.Errorf("ID mismatch: %s != %s", restored.ID, original.ID)
	}
	if restored.Attributes["key"] != original.Attributes["key"] {
		t.Errorf("Attributes mismatch")
	}
}

func TestEventLogRoundTrip(t *testing.T) {
	original := &EventLog{
		Traces: []Trace{
			{
				ID: "case-1",
				Events: []Event{
					{
						Activity:   "Activity1",
						Timestamp:  "2026-03-24T14:30:45Z",
						Attributes: map[string]string{"key": "value"},
						ID:         "550e8400-e29b-41d4-a716-446655440000",
					},
				},
				Attributes: map[string]string{"trace_attr": "trace_value"},
			},
		},
		Attributes: map[string]string{"log_attr": "log_value"},
	}

	// Marshal
	jsonBytes, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal event log: %v", err)
	}

	// Unmarshal
	restored, err := ParseEventLog(jsonBytes)
	if err != nil {
		t.Fatalf("Failed to parse event log: %v", err)
	}

	// Compare
	if len(restored.Traces) != len(original.Traces) {
		t.Errorf("Trace count mismatch: %d != %d", len(restored.Traces), len(original.Traces))
	}
	if restored.Traces[0].ID != original.Traces[0].ID {
		t.Errorf("Trace ID mismatch: %s != %s", restored.Traces[0].ID, original.Traces[0].ID)
	}
	if restored.Attributes["log_attr"] != original.Attributes["log_attr"] {
		t.Errorf("Log attributes mismatch")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func TestNewTestEvent(t *testing.T) {
	ts := time.Now().UTC()
	event := NewTestEvent("TestActivity", ts)

	if event.Activity != "TestActivity" {
		t.Errorf("Expected activity 'TestActivity', got %s", event.Activity)
	}
	if event == nil {
		t.Fatal("NewTestEvent returned nil")
	}
}

func TestNewTestTrace(t *testing.T) {
	trace := NewTestTrace("case-1", 5)

	if trace.ID != "case-1" {
		t.Errorf("Expected trace ID 'case-1', got %s", trace.ID)
	}
	if len(trace.Events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(trace.Events))
	}
}

func TestNewTestEventLog(t *testing.T) {
	log := NewTestEventLog(3, 4)

	if len(log.Traces) != 3 {
		t.Errorf("Expected 3 traces, got %d", len(log.Traces))
	}
	if log.NumEvents() != 12 {
		t.Errorf("Expected 12 total events (3 traces * 4 events), got %d", log.NumEvents())
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestEventValidation_LeadingWhitespace(t *testing.T) {
	event := &Event{
		Activity:   " LeadingSpace",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Attributes: make(map[string]string),
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	err := event.Validate()
	if err == nil {
		t.Fatal("Expected error for activity with leading whitespace")
	}
}

func TestEventValidation_TrailingWhitespace(t *testing.T) {
	event := &Event{
		Activity:   "TrailingSpace ",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Attributes: make(map[string]string),
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	err := event.Validate()
	if err == nil {
		t.Fatal("Expected error for activity with trailing whitespace")
	}
}

func TestEventValidation_ResourceTooLong(t *testing.T) {
	longResource := ""
	for i := 0; i < 101; i++ {
		longResource += "a"
	}

	resource := longResource
	event := &Event{
		Activity:   "TestActivity",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Resource:   &resource,
		Attributes: make(map[string]string),
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}

	err := event.Validate()
	if err == nil {
		t.Fatal("Expected error for resource exceeding 100 chars")
	}
}

func TestPetriNetValidation_MarkingOutOfRange(t *testing.T) {
	net := &PetriNet{
		Places: []Place{
			{ID: "p1", Name: "Start", InitialMarking: 2000, FinalMarking: nil},
		},
		Transitions: []Transition{
			{ID: "t1", Name: "Process", Label: nil},
		},
		Arcs: []Arc{},
	}

	err := net.Validate()
	if err == nil {
		t.Fatal("Expected error for marking out of range")
	}
	bosErr, ok := err.(BosError)
	if !ok || bosErr.Code != ErrOutOfRange {
		t.Fatalf("Expected ErrOutOfRange, got %v", err)
	}
}

func TestConformanceResultFitness(t *testing.T) {
	result := &ConformanceResult{
		IsConformant: true,
		TracesFit:    50,
		TracesTotal:  100,
	}

	expected := 0.5
	actual := result.Fitness()
	if actual != expected {
		t.Errorf("Expected fitness %.2f, got %.2f", expected, actual)
	}
}

func TestConformanceResultFitness_ZeroTraces(t *testing.T) {
	result := &ConformanceResult{
		IsConformant: false,
		TracesFit:    0,
		TracesTotal:  0,
	}

	expected := 0.0
	actual := result.Fitness()
	if actual != expected {
		t.Errorf("Expected fitness %.2f, got %.2f", expected, actual)
	}
}
