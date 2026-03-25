// Package models provides data types and integration layers for BOS (pm4py-rust) ↔ BusinessOS
//
// This package defines Go equivalents of Rust BOS types and provides:
// - Type-safe deserialization from JSON
// - Validation at integration boundaries
// - Conversion utilities for service layer
// - Comprehensive error handling
package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// BosError represents errors from BOS integration
type BosError struct {
	Code    string
	Message string
	Field   string
	Details map[string]interface{}
}

func (e BosError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes
const (
	ErrInvalidActivity      = "INVALID_ACTIVITY"
	ErrInvalidTimestamp     = "INVALID_TIMESTAMP"
	ErrInvalidUUID          = "INVALID_UUID"
	ErrInvalidTraceID       = "INVALID_TRACE_ID"
	ErrMaxSizeExceeded      = "MAX_SIZE_EXCEEDED"
	ErrEmptyCollection      = "EMPTY_COLLECTION"
	ErrValidationFailed     = "VALIDATION_FAILED"
	ErrReferenceNotFound    = "REFERENCE_NOT_FOUND"
	ErrOutOfRange           = "OUT_OF_RANGE"
	ErrMarshaling           = "MARSHALING_ERROR"
	ErrUnmarshaling         = "UNMARSHALING_ERROR"
	ErrUUID                 = "INVALID_UUID_FORMAT"
)

// ============================================================================
// Event Model (BOS → BusinessOS)
// ============================================================================

// Event represents a single activity occurrence in a business process
type Event struct {
	Activity   string            `json:"activity"`
	Timestamp  string            `json:"timestamp"` // RFC3339 format
	Resource   *string           `json:"resource,omitempty"`
	Attributes map[string]string `json:"attributes"`
	ID         string            `json:"id"` // UUID v4 string
}

// ParseEvent converts raw JSON to Event with validation
func ParseEvent(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, BosError{
			Code:    ErrUnmarshaling,
			Message: fmt.Sprintf("Failed to unmarshal event: %v", err),
		}
	}

	if err := event.Validate(); err != nil {
		return nil, err
	}

	return &event, nil
}

// Validate checks Event constraints
func (e *Event) Validate() error {
	// Activity validation
	if e.Activity == "" {
		return BosError{
			Code:    ErrInvalidActivity,
			Message: "Activity cannot be empty",
			Field:   "activity",
		}
	}
	if len(e.Activity) > 255 {
		return BosError{
			Code:    ErrInvalidActivity,
			Message: fmt.Sprintf("Activity exceeds 255 chars: %d", len(e.Activity)),
			Field:   "activity",
		}
	}
	if e.Activity != strings.TrimSpace(e.Activity) {
		return BosError{
			Code:    ErrInvalidActivity,
			Message: "Activity has leading/trailing whitespace",
			Field:   "activity",
		}
	}

	// Timestamp validation
	ts, err := parseRFC3339(e.Timestamp)
	if err != nil {
		return BosError{
			Code:    ErrInvalidTimestamp,
			Message: fmt.Sprintf("Invalid timestamp format: %v", err),
			Field:   "timestamp",
		}
	}
	if ts.After(time.Now()) {
		return BosError{
			Code:    ErrInvalidTimestamp,
			Message: "Event timestamp cannot be in the future",
			Field:   "timestamp",
		}
	}

	// Resource validation
	if e.Resource != nil && *e.Resource != "" {
		if len(*e.Resource) > 100 {
			return BosError{
				Code:    ErrValidationFailed,
				Message: "Resource exceeds 100 chars",
				Field:   "resource",
			}
		}
	}

	// Attributes validation
	if len(e.Attributes) > 50 {
		return BosError{
			Code:    ErrMaxSizeExceeded,
			Message: fmt.Sprintf("Max attributes: 50, got %d", len(e.Attributes)),
			Field:   "attributes",
		}
	}
	for k, v := range e.Attributes {
		if len(k) == 0 || len(k) > 50 {
			return BosError{
				Code:    ErrValidationFailed,
				Message: fmt.Sprintf("Attribute key must be 1-50 chars: %q", k),
				Field:   "attributes",
			}
		}
		if len(v) > 1000 {
			return BosError{
				Code:    ErrValidationFailed,
				Message: fmt.Sprintf("Attribute %q value exceeds 1000 chars: %d", k, len(v)),
				Field:   "attributes",
			}
		}
	}

	// UUID validation
	if !isValidUUID(e.ID) {
		return BosError{
			Code:    ErrUUID,
			Message: fmt.Sprintf("Invalid UUID format: %s", e.ID),
			Field:   "id",
		}
	}

	return nil
}

// MarshalJSON converts Event to JSON with RFC3339 timestamp
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&e),
	})
}

// ============================================================================
// Trace Model (BOS → BusinessOS)
// ============================================================================

// Trace represents a sequence of events for a single process instance
type Trace struct {
	ID         string            `json:"id"`
	Events     []Event           `json:"events"`
	Attributes map[string]string `json:"attributes"`
}

// ParseTrace converts raw JSON to Trace with validation
func ParseTrace(data []byte) (*Trace, error) {
	var trace Trace
	if err := json.Unmarshal(data, &trace); err != nil {
		return nil, BosError{
			Code:    ErrUnmarshaling,
			Message: fmt.Sprintf("Failed to unmarshal trace: %v", err),
		}
	}

	if err := trace.Validate(); err != nil {
		return nil, err
	}

	return &trace, nil
}

// Validate checks Trace constraints
func (t *Trace) Validate() error {
	// Trace ID validation
	if t.ID == "" || len(t.ID) > 100 {
		return BosError{
			Code:    ErrInvalidTraceID,
			Message: fmt.Sprintf("Trace ID must be 1-100 chars, got %d", len(t.ID)),
			Field:   "id",
		}
	}

	// Events validation
	if len(t.Events) == 0 {
		return BosError{
			Code:    ErrEmptyCollection,
			Message: "Trace must have at least one event",
			Field:   "events",
		}
	}
	if len(t.Events) > 10000 {
		return BosError{
			Code:    ErrMaxSizeExceeded,
			Message: fmt.Sprintf("Trace exceeds 10000 events: %d", len(t.Events)),
			Field:   "events",
		}
	}

	// Validate all events
	for i, event := range t.Events {
		if err := event.Validate(); err != nil {
			return BosError{
				Code:    ErrValidationFailed,
				Message: fmt.Sprintf("Event[%d] validation failed: %v", i, err),
				Field:   fmt.Sprintf("events[%d]", i),
			}
		}
	}

	// Attributes validation
	if len(t.Attributes) > 50 {
		return BosError{
			Code:    ErrMaxSizeExceeded,
			Message: fmt.Sprintf("Max attributes: 50, got %d", len(t.Attributes)),
			Field:   "attributes",
		}
	}

	return nil
}

// ============================================================================
// EventLog Model (BOS → BusinessOS)
// ============================================================================

// EventLog represents a complete process log (collection of traces)
type EventLog struct {
	Traces     []Trace           `json:"traces"`
	Attributes map[string]string `json:"attributes"`
}

// ParseEventLog converts raw JSON to EventLog with validation
func ParseEventLog(data []byte) (*EventLog, error) {
	var log EventLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, BosError{
			Code:    ErrUnmarshaling,
			Message: fmt.Sprintf("Failed to unmarshal event log: %v", err),
		}
	}

	if err := log.Validate(); err != nil {
		return nil, err
	}

	return &log, nil
}

// Validate checks EventLog constraints
func (e *EventLog) Validate() error {
	// Traces validation
	if len(e.Traces) == 0 {
		return BosError{
			Code:    ErrEmptyCollection,
			Message: "EventLog must have at least one trace",
			Field:   "traces",
		}
	}
	if len(e.Traces) > 1_000_000 {
		return BosError{
			Code:    ErrMaxSizeExceeded,
			Message: fmt.Sprintf("EventLog exceeds 1M traces: %d", len(e.Traces)),
			Field:   "traces",
		}
	}

	// Validate all traces
	traceIds := make(map[string]bool)
	for i, trace := range e.Traces {
		if err := trace.Validate(); err != nil {
			return BosError{
				Code:    ErrValidationFailed,
				Message: fmt.Sprintf("Trace[%d] validation failed: %v", i, err),
				Field:   fmt.Sprintf("traces[%d]", i),
			}
		}

		if traceIds[trace.ID] {
			return BosError{
				Code:    ErrValidationFailed,
				Message: fmt.Sprintf("Duplicate trace ID: %s", trace.ID),
				Field:   fmt.Sprintf("traces[%d].id", i),
			}
		}
		traceIds[trace.ID] = true
	}

	// Attributes validation
	if len(e.Attributes) > 50 {
		return BosError{
			Code:    ErrMaxSizeExceeded,
			Message: fmt.Sprintf("Max attributes: 50, got %d", len(e.Attributes)),
			Field:   "attributes",
		}
	}

	return nil
}

// MarshalJSON converts EventLog to JSON
func (e EventLog) MarshalJSON() ([]byte, error) {
	type Alias EventLog
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&e),
	})
}

// NumEvents returns total count of events across all traces
func (e *EventLog) NumEvents() int {
	count := 0
	for _, trace := range e.Traces {
		count += len(trace.Events)
	}
	return count
}

// ============================================================================
// PetriNet Model (BOS → BusinessOS)
// ============================================================================

// Place represents a place in a Petri net
type Place struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	InitialMarking int    `json:"initial_marking"`
	FinalMarking   *int   `json:"final_marking,omitempty"`
}

// Transition represents a transition in a Petri net
type Transition struct {
	ID    string  `json:"id"`
	Label *string `json:"label,omitempty"`
	Name  string  `json:"name"`
}

// Arc represents an arc (edge) in a Petri net
type Arc struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Weight int    `json:"weight"`
}

// PetriNet represents a complete Petri net model
type PetriNet struct {
	Places       []Place      `json:"places"`
	Transitions  []Transition `json:"transitions"`
	Arcs         []Arc        `json:"arcs"`
	InitialPlace *string      `json:"initial_place,omitempty"`
	FinalPlace   *string      `json:"final_place,omitempty"`
}

// ParsePetriNet converts raw JSON to PetriNet with validation
func ParsePetriNet(data []byte) (*PetriNet, error) {
	var net PetriNet
	if err := json.Unmarshal(data, &net); err != nil {
		return nil, BosError{
			Code:    ErrUnmarshaling,
			Message: fmt.Sprintf("Failed to unmarshal Petri net: %v", err),
		}
	}

	if err := net.Validate(); err != nil {
		return nil, err
	}

	return &net, nil
}

// Validate checks PetriNet constraints and referential integrity
func (p *PetriNet) Validate() error {
	// Empty collections check
	if len(p.Places) == 0 {
		return BosError{
			Code:    ErrEmptyCollection,
			Message: "PetriNet must have at least one place",
			Field:   "places",
		}
	}
	if len(p.Transitions) == 0 {
		return BosError{
			Code:    ErrEmptyCollection,
			Message: "PetriNet must have at least one transition",
			Field:   "transitions",
		}
	}

	// Validate places
	placeIDs := make(map[string]bool)
	for i, place := range p.Places {
		if place.ID == "" {
			return BosError{
				Code:    ErrValidationFailed,
				Message: "Place ID cannot be empty",
				Field:   fmt.Sprintf("places[%d].id", i),
			}
		}
		if placeIDs[place.ID] {
			return BosError{
				Code:    ErrValidationFailed,
				Message: fmt.Sprintf("Duplicate place ID: %s", place.ID),
				Field:   fmt.Sprintf("places[%d]", i),
			}
		}
		placeIDs[place.ID] = true

		if place.InitialMarking < 0 || place.InitialMarking > 1000 {
			return BosError{
				Code:    ErrOutOfRange,
				Message: fmt.Sprintf("Marking out of range [0,1000]: %d", place.InitialMarking),
				Field:   fmt.Sprintf("places[%d].initial_marking", i),
			}
		}
		if place.FinalMarking != nil && (*place.FinalMarking < 0 || *place.FinalMarking > 1000) {
			return BosError{
				Code:    ErrOutOfRange,
				Message: fmt.Sprintf("Marking out of range [0,1000]: %d", *place.FinalMarking),
				Field:   fmt.Sprintf("places[%d].final_marking", i),
			}
		}
	}

	// Validate transitions
	transIDs := make(map[string]bool)
	for i, trans := range p.Transitions {
		if trans.ID == "" {
			return BosError{
				Code:    ErrValidationFailed,
				Message: "Transition ID cannot be empty",
				Field:   fmt.Sprintf("transitions[%d].id", i),
			}
		}
		if transIDs[trans.ID] {
			return BosError{
				Code:    ErrValidationFailed,
				Message: fmt.Sprintf("Duplicate transition ID: %s", trans.ID),
				Field:   fmt.Sprintf("transitions[%d]", i),
			}
		}
		transIDs[trans.ID] = true
	}

	// Validate arcs (referential integrity)
	for i, arc := range p.Arcs {
		if !placeIDs[arc.From] && !transIDs[arc.From] {
			return BosError{
				Code:    ErrReferenceNotFound,
				Message: fmt.Sprintf("Arc source %q not found", arc.From),
				Field:   fmt.Sprintf("arcs[%d].from", i),
			}
		}
		if !placeIDs[arc.To] && !transIDs[arc.To] {
			return BosError{
				Code:    ErrReferenceNotFound,
				Message: fmt.Sprintf("Arc target %q not found", arc.To),
				Field:   fmt.Sprintf("arcs[%d].to", i),
			}
		}
		if arc.Weight < 0 || arc.Weight > 1000 {
			return BosError{
				Code:    ErrOutOfRange,
				Message: fmt.Sprintf("Weight out of range [0,1000]: %d", arc.Weight),
				Field:   fmt.Sprintf("arcs[%d].weight", i),
			}
		}
	}

	return nil
}

// ============================================================================
// ConformanceResult Model
// ============================================================================

// ConformanceResult represents the result of conformance checking
type ConformanceResult struct {
	IsConformant bool `json:"is_conformant"`
	TracesFit    int  `json:"traces_fit"`
	TracesTotal  int  `json:"traces_total"`
	Violations   []struct {
		TraceIndex int `json:"trace_index"`
		EventIndex int `json:"event_index"`
	} `json:"violations"`
}

// Fitness returns the fitness ratio (0.0 to 1.0)
func (c *ConformanceResult) Fitness() float64 {
	if c.TracesTotal == 0 {
		return 0.0
	}
	return float64(c.TracesFit) / float64(c.TracesTotal)
}

// ============================================================================
// LogStatistics Model
// ============================================================================

// LogStatistics represents aggregated log statistics
type LogStatistics struct {
	NumTraces           int              `json:"num_traces"`
	NumEvents           int              `json:"num_events"`
	NumVariants         int              `json:"num_variants"`
	AvgTraceLength      float64          `json:"avg_trace_length"`
	MinTraceLength      int              `json:"min_trace_length"`
	MaxTraceLength      int              `json:"max_trace_length"`
	Activities          []string         `json:"activities"`
	ActivityFrequencies map[string]int   `json:"activity_frequencies"`
	Variants            map[string]int   `json:"variants"`
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseRFC3339 parses RFC3339 timestamp string to time.Time
func parseRFC3339(ts string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid RFC3339 format: %w", err)
	}
	// Ensure UTC
	return t.UTC(), nil
}

// isValidUUID checks if string is a valid UUID v4
func isValidUUID(u string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(u)
}

// EventLogToConformanceCheckRequest converts EventLog to a format suitable for PM algorithms
func EventLogToConformanceCheckRequest(log *EventLog) map[string]interface{} {
	return map[string]interface{}{
		"num_traces":      len(log.Traces),
		"num_events":      log.NumEvents(),
		"traces":          log.Traces,
		"metadata":        log.Attributes,
	}
}

// PetriNetToString returns a human-readable representation of the Petri net
func (p *PetriNet) String() string {
	return fmt.Sprintf(
		"PetriNet{places: %d, transitions: %d, arcs: %d, initial: %v, final: %v}",
		len(p.Places), len(p.Transitions), len(p.Arcs),
		p.InitialPlace, p.FinalPlace,
	)
}

// ============================================================================
// Tests
// ============================================================================

// Test types to ensure validation works correctly

func NewTestEvent(activity string, timestamp time.Time) *Event {
	ts := timestamp.UTC().Format(time.RFC3339)
	return &Event{
		Activity:   activity,
		Timestamp:  ts,
		Attributes: make(map[string]string),
		ID:         "550e8400-e29b-41d4-a716-446655440000",
	}
}

func NewTestTrace(id string, eventCount int) *Trace {
	trace := &Trace{
		ID:         id,
		Events:     make([]Event, eventCount),
		Attributes: make(map[string]string),
	}

	for i := 0; i < eventCount; i++ {
		ts := time.Now().UTC().Add(-time.Duration(eventCount-i) * time.Minute)
		trace.Events[i] = Event{
			Activity:   fmt.Sprintf("Activity%d", i+1),
			Timestamp:  ts.Format(time.RFC3339),
			Attributes: make(map[string]string),
			ID:         fmt.Sprintf("550e8400-e29b-41d4-a716-44665544000%d", i),
		}
	}

	return trace
}

func NewTestEventLog(traceCount int, eventsPerTrace int) *EventLog {
	log := &EventLog{
		Traces:     make([]Trace, traceCount),
		Attributes: make(map[string]string),
	}

	for i := 0; i < traceCount; i++ {
		log.Traces[i] = *NewTestTrace(fmt.Sprintf("case-%d", i+1), eventsPerTrace)
	}

	return log
}
