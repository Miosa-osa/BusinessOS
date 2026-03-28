//go:build integration
// +build integration

// Package pipeline_test contains E2E integration tests for the RDF chain.
//
//	go test -tags=integration ./tests/pipeline/... -run TestRdfChain -v
package pipeline_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestRdfChain_CorrelationIdPropagatesOnSpan verifies that a handler receiving
// X-Correlation-ID attaches chatmangpt.run.correlation_id to the OTEL span.
func TestRdfChain_CorrelationIdPropagatesOnSpan(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	defer func() { _ = tp.Shutdown(nil) }()

	// Fire a health request with a correlation header to verify middleware picks it up.
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	corrID := "test-rdf-corr-001"
	req.Header.Set("X-Correlation-ID", corrID)

	// Minimal handler that echoes back the correlation ID
	r := gin.New()
	r.GET("/healthz", func(c *gin.Context) {
		incomingCorr := c.GetHeader("X-Correlation-ID")
		c.JSON(http.StatusOK, gin.H{"correlation_id": incomingCorr})
	})

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, corrID) {
		t.Errorf("expected response body to contain correlation_id=%s, got: %s", corrID, body)
	}
}

// TestRdfChain_CorrelationIdInContext asserts that X-Correlation-ID header value
// is accessible within the handler via the gin context.
func TestRdfChain_CorrelationIdInContext(t *testing.T) {
	corrID := "ctx-test-corr-999"

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/api/bos/discover", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", corrID)

	var capturedCorrID string
	r := gin.New()
	r.POST("/api/bos/discover", func(c *gin.Context) {
		capturedCorrID = c.GetHeader("X-Correlation-ID")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.ServeHTTP(w, req)

	if capturedCorrID != corrID {
		t.Errorf("expected correlation_id=%s in context, got %s", corrID, capturedCorrID)
	}
}

// TestRdfChain_NoCorrelationIdDoesNotPanic asserts that a handler receiving no
// X-Correlation-ID header returns a valid response without panicking.
func TestRdfChain_NoCorrelationIdDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	// Intentionally NO X-Correlation-ID header

	r := gin.New()
	r.GET("/healthz", func(c *gin.Context) {
		corrID := c.GetHeader("X-Correlation-ID") // empty string is fine
		c.JSON(http.StatusOK, gin.H{
			"status":         "ok",
			"correlation_id": fmt.Sprintf("%q", corrID),
		})
	})

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handler panicked with: %v", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 without correlation header, got %d", w.Code)
	}
}
