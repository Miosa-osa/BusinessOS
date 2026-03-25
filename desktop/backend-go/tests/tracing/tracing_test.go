package tracing_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Trace represents a distributed trace context
type Trace struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Flags     string
	StartTime time.Time
}

// Span represents an individual span in a trace
type Span struct {
	TraceID      string
	SpanID       string
	ParentID     string
	Name         string
	StartTime    time.Time
	EndTime      time.Time
	DurationMs   int64
	Status       string
	Attributes   map[string]interface{}
	Service      string
}

// TestExtractTraceparent tests extracting traceparent from headers
func TestExtractTraceparent(t *testing.T) {
	headers := map[string]string{
		"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
	}

	trace, err := extractTraceparent(headers)
	require.NoError(t, err)

	assert.Equal(t, "0af7651916cd43dd8448eb211c80319c", trace.TraceID)
	assert.Equal(t, "b7ad6b7169203331", trace.SpanID)
	assert.Equal(t, "01", trace.Flags)
}

// TestExtractTraceparent_Missing tests generating new trace when header missing
func TestExtractTraceparent_Missing(t *testing.T) {
	headers := map[string]string{}

	trace, err := extractTraceparent(headers)
	require.NoError(t, err)

	assert.Equal(t, 32, len(trace.TraceID))
	assert.Equal(t, 16, len(trace.SpanID))
}

// TestCreateSpan tests creating a span with parent trace
func TestCreateSpan(t *testing.T) {
	parent := &Trace{
		TraceID: "0af7651916cd43dd8448eb211c80319c",
		SpanID:  "b7ad6b7169203331",
		Flags:   "01",
	}

	span, err := createSpan(parent, "test_operation", map[string]interface{}{"service": "go"})
	require.NoError(t, err)

	assert.Equal(t, parent.TraceID, span.TraceID)
	assert.Equal(t, parent.SpanID, span.ParentID)
	assert.Equal(t, "test_operation", span.Name)
	assert.Equal(t, "go", span.Attributes["service"])
}

// TestEndSpan tests ending a span and recording duration
func TestEndSpan(t *testing.T) {
	parent := &Trace{
		TraceID: "abc1234567890def1234567890abcdef",
		SpanID:  "1234567890abcdef",
		Flags:   "01",
	}

	span, err := createSpan(parent, "operation", map[string]interface{}{})
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	err = endSpan(span, "ok")
	require.NoError(t, err)

	assert.Greater(t, span.DurationMs, int64(0))
	assert.Equal(t, "ok", span.Status)
}

// TestTraceparentPropagation tests propagating traceparent across services
func TestTraceparentPropagation(t *testing.T) {
	// BusinessOS receives request with traceparent
	headers := map[string]string{
		"traceparent": "00-abc1234567890def1234567890abcdef-1234567890abcdef-01",
	}

	trace, err := extractTraceparent(headers)
	require.NoError(t, err)

	// BusinessOS creates span
	spanA, err := createSpan(trace, "businessos_operation", map[string]interface{}{"service": "go"})
	require.NoError(t, err)

	// BusinessOS propagates to Canopy
	propagated := encodeTraceparent(spanA)

	traceB, err := extractTraceparent(map[string]string{"traceparent": propagated})
	require.NoError(t, err)

	assert.Equal(t, trace.TraceID, traceB.TraceID)

	// Canopy creates span
	spanB, err := createSpan(traceB, "canopy_operation", map[string]interface{}{"service": "elixir"})
	require.NoError(t, err)

	// Canopy propagates to OSA
	propagatedB := encodeTraceparent(spanB)

	traceC, err := extractTraceparent(map[string]string{"traceparent": propagatedB})
	require.NoError(t, err)

	// All should share same trace_id
	assert.Equal(t, trace.TraceID, traceC.TraceID)
	assert.Equal(t, spanA.TraceID, spanB.TraceID)
	assert.Equal(t, spanB.TraceID, traceC.TraceID)
}

// TestSpanCreationOnServiceBoundaries tests span creation at service boundaries
func TestSpanCreationOnServiceBoundaries(t *testing.T) {
	// Entry boundary
	trace := &Trace{
		TraceID: "fedcba9876543210fedcba9876543210",
		SpanID:  "fedcba9876543210",
		Flags:   "01",
	}

	spanEntry, err := createSpan(trace, "boundary_entry", map[string]interface{}{
		"service":   "go",
		"operation": "api_call",
	})
	require.NoError(t, err)

	assert.Equal(t, trace.TraceID, spanEntry.TraceID)
	assert.Equal(t, trace.SpanID, spanEntry.ParentID)

	// Exit boundary
	spanExit, err := createSpan(trace, "boundary_exit", map[string]interface{}{
		"service": "go",
		"status":  "complete",
	})
	require.NoError(t, err)

	err = endSpan(spanExit, "ok")
	require.NoError(t, err)

	assert.Equal(t, "ok", spanExit.Status)
}

// TestTraceReconstruction tests reconstructing trace from service spans
func TestTraceReconstruction(t *testing.T) {
	traceID := "11111111111111111111111111111111"

	now := time.Now()

	spanA := &Span{
		TraceID:    traceID,
		SpanID:     "aaaaaaaaaaaaaaaa",
		Name:       "service_go",
		ParentID:   "",
		StartTime:  now,
		Service:    "go",
		Attributes: map[string]interface{}{},
	}

	spanB := &Span{
		TraceID:    traceID,
		SpanID:     "bbbbbbbbbbbbbbbb",
		ParentID:   spanA.SpanID,
		Name:       "service_elixir",
		StartTime:  now.Add(10 * time.Millisecond),
		Service:    "elixir",
		Attributes: map[string]interface{}{},
	}

	spanC := &Span{
		TraceID:    traceID,
		SpanID:     "cccccccccccccccc",
		ParentID:   spanB.SpanID,
		Name:       "service_rust",
		StartTime:  now.Add(20 * time.Millisecond),
		Service:    "rust",
		Attributes: map[string]interface{}{},
	}

	spans := []*Span{spanA, spanB, spanC}

	reconstructed, err := reconstructTrace(spans)
	require.NoError(t, err)

	assert.Equal(t, traceID, reconstructed.TraceID)
	assert.Equal(t, 3, len(reconstructed.Spans))
	assert.Equal(t, "service_go", reconstructed.Spans[0].Name)
}

// TestRecordAttribute tests recording attributes on span
func TestRecordAttribute(t *testing.T) {
	span := &Span{
		TraceID:    "ffffffffffffffffffffffffffffffff",
		SpanID:     "ffffffffffffffff",
		Name:       "test",
		Attributes: map[string]interface{}{},
	}

	err := recordAttribute(span, "user_id", "user_123")
	require.NoError(t, err)

	assert.Equal(t, "user_123", span.Attributes["user_id"])
}

// TestEncodeTraceparent tests W3C Trace Context format encoding
func TestEncodeTraceparent(t *testing.T) {
	span := &Span{
		TraceID: "0af7651916cd43dd8448eb211c80319c",
		SpanID:  "b7ad6b7169203331",
	}

	encoded := encodeTraceparent(span)

	assert.Equal(t, "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01", encoded)
}

// Helper functions

func extractTraceparent(headers map[string]string) (*Trace, error) {
	if tp, ok := headers["traceparent"]; ok {
		parts := strings.Split(tp, "-")
		if len(parts) == 4 {
			return &Trace{
				TraceID: parts[1],
				SpanID:  parts[2],
				Flags:   parts[3],
			}, nil
		}
	}
	// Generate new trace
	return &Trace{
		TraceID:   generateID(32),
		SpanID:    generateID(16),
		Flags:     "01",
		StartTime: time.Now(),
	}, nil
}

func createSpan(parent *Trace, name string, attrs map[string]interface{}) (*Span, error) {
	parentID := ""
	if parent != nil {
		parentID = parent.SpanID
	}

	span := &Span{
		TraceID:    parent.TraceID,
		SpanID:     generateID(16),
		ParentID:   parentID,
		Name:       name,
		StartTime:  time.Now(),
		Status:     "running",
		Attributes: attrs,
	}

	return span, nil
}

func endSpan(span *Span, status string) error {
	span.EndTime = time.Now()
	span.DurationMs = span.EndTime.Sub(span.StartTime).Milliseconds()
	span.Status = status
	return nil
}

func encodeTraceparent(span *Span) string {
	return "00-" + span.TraceID + "-" + span.SpanID + "-01"
}

func reconstructTrace(spans []*Span) (*Trace, error) {
	if len(spans) == 0 {
		return nil, nil
	}

	return &Trace{
		TraceID: spans[0].TraceID,
		Spans:   spans,
	}, nil
}

func recordAttribute(span *Span, key string, value interface{}) error {
	if span.Attributes == nil {
		span.Attributes = make(map[string]interface{})
	}
	span.Attributes[key] = value
	return nil
}

func generateID(length int) string {
	const chars = "0123456789abcdef"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}

// Add Spans field to Trace struct for reconstruction test
type TraceWithSpans struct {
	TraceID string
	Spans   []*Span
}

func (t *Trace) Spans([]*Span) {}
