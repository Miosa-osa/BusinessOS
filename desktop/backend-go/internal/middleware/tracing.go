package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Trace represents a distributed trace context (W3C Trace Context)
type Trace struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Flags     string
	StartTime time.Time
}

// Span represents an individual operation span
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

// Middleware for distributed tracing
type TracingMiddleware struct {
	serviceName string
}

// NewTracingMiddleware creates a new tracing middleware
func NewTracingMiddleware(serviceName string) *TracingMiddleware {
	return &TracingMiddleware{
		serviceName: serviceName,
	}
}

// Handler wraps an HTTP handler with tracing
func (tm *TracingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract or create trace
		trace, _ := ExtractTraceparent(r.Header)

		// Create span for this operation
		span := createSpan(trace, r.RequestURI, map[string]interface{}{
			"service": tm.serviceName,
			"method":  r.Method,
		})

		// Store in context
		ctx := context.WithValue(r.Context(), "trace", trace)
		ctx = context.WithValue(ctx, "span", span)
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(w, r)

		// End span
		endSpan(span, "ok")
	})
}

// ExtractTraceparent extracts traceparent from HTTP headers
func ExtractTraceparent(headers http.Header) (*Trace, error) {
	if tp := headers.Get("traceparent"); tp != "" {
		return parseTraceparent(tp)
	}

	// Generate new trace
	return generateTrace(), nil
}

// CreateSpan creates a new span with parent trace
func CreateSpan(parent *Trace, name string, attrs map[string]interface{}) *Span {
	return createSpan(parent, name, attrs)
}

// EndSpan ends a span and records duration
func EndSpan(span *Span, status string) {
	endSpan(span, status)
}

// EncodeTraceparent encodes a span as W3C Trace Context format
func EncodeTraceparent(span *Span) string {
	return fmt.Sprintf("00-%s-%s-01", span.TraceID, span.SpanID)
}

// RecordAttribute records an attribute on a span
func RecordAttribute(span *Span, key string, value interface{}) {
	if span.Attributes == nil {
		span.Attributes = make(map[string]interface{})
	}
	span.Attributes[key] = value
}

// ReconstructTrace reconstructs a complete trace from spans
func ReconstructTrace(spans []*Span) (*Trace, error) {
	if len(spans) == 0 {
		return nil, fmt.Errorf("no spans provided")
	}

	return &Trace{
		TraceID:   spans[0].TraceID,
		StartTime: spans[0].StartTime,
	}, nil
}

// PropagateHeaders adds traceparent header for downstream propagation
func PropagateHeaders(span *Span, headers http.Header) http.Header {
	if headers == nil {
		headers = make(http.Header)
	}

	tp := EncodeTraceparent(span)
	headers.Set("traceparent", tp)

	return headers
}

// ============================================================================
// Private helpers
// ============================================================================

func generateTrace() *Trace {
	return &Trace{
		TraceID:   generateID(32),
		SpanID:    generateID(16),
		Flags:     "01",
		StartTime: time.Now(),
	}
}

func createSpan(parent *Trace, name string, attrs map[string]interface{}) *Span {
	parentID := ""
	traceID := generateID(32)

	if parent != nil {
		parentID = parent.SpanID
		traceID = parent.TraceID
	}

	return &Span{
		TraceID:    traceID,
		SpanID:     generateID(16),
		ParentID:   parentID,
		Name:       name,
		StartTime:  time.Now(),
		Status:     "running",
		Attributes: attrs,
		Service:    "go",
	}
}

func endSpan(span *Span, status string) {
	span.EndTime = time.Now()
	span.DurationMs = span.EndTime.Sub(span.StartTime).Milliseconds()
	span.Status = status
}

func parseTraceparent(tp string) (*Trace, error) {
	parts := strings.Split(tp, "-")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid traceparent format")
	}

	if parts[0] != "00" {
		return nil, fmt.Errorf("unsupported trace version")
	}

	if len(parts[1]) != 32 || len(parts[2]) != 16 {
		return nil, fmt.Errorf("invalid trace or span ID length")
	}

	return &Trace{
		TraceID:   parts[1],
		SpanID:    parts[2],
		Flags:     parts[3],
		StartTime: time.Now(),
	}, nil
}

func generateID(length int) string {
	const chars = "0123456789abcdef"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// HashVote creates a SHA256 hash of a vote value
func HashVote(value interface{}) string {
	data := []byte(fmt.Sprintf("%v", value))
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)[:16]
}
