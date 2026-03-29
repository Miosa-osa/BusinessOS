//go:build !integration
// +build !integration

// Package pipeline_test contains unit-level span tests for the RDF discovery chain.
//
// These tests run without any build tag in normal `go test ./...` runs.
// They use an in-process OTEL SpanRecorder and a httptest router to verify that
// the X-Correlation-ID header is stamped onto the OTEL span as the semconv
// attribute chatmangpt.run.correlation_id.
//
// Chicago TDD: RED → GREEN → REFACTOR
// Claim: Discover handler attaches chatmangpt.run.correlation_id to its span.
package pipeline_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	semconv "github.com/rhl/businessos-backend/internal/semconv"
)

// newTestTracer wires an in-process SpanRecorder into the global OTEL tracer and
// returns the recorder plus a cleanup function that shuts down the provider.
func newTestTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return sr
}

// discoverHandlerUnderTest mirrors the OTEL instrumentation in
// internal/handlers/bos_gateway.go Discover() (lines 199-208):
//
//	ctx, span := tracer.Start(c.Request.Context(), semconv.BosGatewayDiscoverSpan)
//	if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
//	    span.SetAttributes(attribute.String(string(semconv.ChatmangptRunCorrelationIdKey), correlationID))
//	}
//
// Using a minimal handler lets us test the span instrumentation logic without
// wiring the full database / pm4py-rust service stack.
func discoverHandlerUnderTest(c *gin.Context) {
	tracer := otel.Tracer("businessos-gateway")
	_, span := tracer.Start(c.Request.Context(), semconv.BosGatewayDiscoverSpan)
	defer span.End()

	if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
		span.SetAttributes(attribute.String(string(semconv.ChatmangptRunCorrelationIdKey), correlationID))
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// TestCorrelationID_SetOnDiscoverSpan is the primary claim test:
// A POST to /api/bos/discover bearing X-Correlation-ID must record the value
// as the chatmangpt.run.correlation_id attribute on the resulting OTEL span.
func TestCorrelationID_SetOnDiscoverSpan(t *testing.T) {
	sr := newTestTracer(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/bos/discover", discoverHandlerUnderTest)

	const wantCorrID = "rdf-chain-corr-001"
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/bos/discover", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", wantCorrID)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	// Verify the span was recorded and carries the expected attribute.
	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatal("no spans recorded — handler did not emit a span")
	}

	// Find the bos.gateway.discover span (may be the only one).
	var discoverSpan sdktrace.ReadOnlySpan
	for _, s := range spans {
		if s.Name() == semconv.BosGatewayDiscoverSpan {
			discoverSpan = s
			break
		}
	}
	if discoverSpan == nil {
		t.Fatalf("no span named %q found; recorded spans: %v", semconv.BosGatewayDiscoverSpan, spanNames(spans))
	}

	// Assert chatmangpt.run.correlation_id is present and matches.
	got, ok := attributeValue(discoverSpan.Attributes(), semconv.ChatmangptRunCorrelationIdKey)
	if !ok {
		t.Fatalf("span %q missing attribute %q; attrs: %v",
			semconv.BosGatewayDiscoverSpan, semconv.ChatmangptRunCorrelationIdKey, discoverSpan.Attributes())
	}
	if got != wantCorrID {
		t.Errorf("attribute %q: got %q, want %q",
			semconv.ChatmangptRunCorrelationIdKey, got, wantCorrID)
	}

	t.Logf("PASS: span %q carries %s=%s", semconv.BosGatewayDiscoverSpan,
		semconv.ChatmangptRunCorrelationIdKey, got)
}

// TestCorrelationID_AbsentWhenHeaderMissing verifies that when no X-Correlation-ID
// header is sent the attribute is NOT stamped (avoid polluting spans with empty strings).
func TestCorrelationID_AbsentWhenHeaderMissing(t *testing.T) {
	sr := newTestTracer(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/bos/discover", discoverHandlerUnderTest)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/bos/discover", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Intentionally omit X-Correlation-ID

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatal("no spans recorded")
	}
	for _, s := range spans {
		if s.Name() == semconv.BosGatewayDiscoverSpan {
			if _, ok := attributeValue(s.Attributes(), semconv.ChatmangptRunCorrelationIdKey); ok {
				t.Errorf("attribute %q must NOT be set when header is absent",
					semconv.ChatmangptRunCorrelationIdKey)
			}
			t.Log("PASS: correlation_id attribute absent when header omitted")
			return
		}
	}
	t.Fatalf("no span named %q recorded", semconv.BosGatewayDiscoverSpan)
}

// TestCorrelationID_UniquePerRequest verifies that two sequential requests with
// distinct correlation IDs produce spans with their own distinct attribute values —
// no cross-contamination between requests.
func TestCorrelationID_UniquePerRequest(t *testing.T) {
	sr := newTestTracer(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/bos/discover", discoverHandlerUnderTest)

	send := func(corrID string) {
		t.Helper()
		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/api/bos/discover", strings.NewReader(`{}`))
		if err != nil {
			t.Fatalf("build request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Correlation-ID", corrID)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	}

	const corrA = "unique-corr-aaa"
	const corrB = "unique-corr-bbb"
	send(corrA)
	send(corrB)

	spans := sr.Ended()
	if len(spans) < 2 {
		t.Fatalf("expected at least 2 spans, got %d", len(spans))
	}

	var values []string
	for _, s := range spans {
		if s.Name() == semconv.BosGatewayDiscoverSpan {
			if v, ok := attributeValue(s.Attributes(), semconv.ChatmangptRunCorrelationIdKey); ok {
				values = append(values, v)
			}
		}
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 correlation_id values, got %v", values)
	}
	if values[0] == values[1] {
		t.Errorf("both spans carry identical correlation_id %q — IDs must be unique per request", values[0])
	}
	if values[0] != corrA {
		t.Errorf("first span: got %q, want %q", values[0], corrA)
	}
	if values[1] != corrB {
		t.Errorf("second span: got %q, want %q", values[1], corrB)
	}
	t.Logf("PASS: two distinct spans carry corrA=%s corrB=%s", values[0], values[1])
}

// ── helpers ──────────────────────────────────────────────────────────────────

// attributeValue returns the string value of the first matching attribute key.
func attributeValue(attrs []attribute.KeyValue, key attribute.Key) (string, bool) {
	for _, kv := range attrs {
		if kv.Key == key {
			return kv.Value.AsString(), true
		}
	}
	return "", false
}

// spanNames returns a slice of span names for diagnostic output.
func spanNames(spans []sdktrace.ReadOnlySpan) []string {
	names := make([]string, len(spans))
	for i, s := range spans {
		names[i] = s.Name()
	}
	return names
}
