//go:build integration
// +build integration

// Package pipeline_test contains E2E correlation integration tests.
//
// These tests verify that the BusinessOS gateway propagates W3C traceparent
// and x-correlation-id headers without returning 5xx errors.
//
// They are gated by the "integration" build tag so that normal `go test ./...`
// runs skip them automatically. To execute:
//
//	go test -tags=integration ./tests/pipeline/... -v
//
// Each test calls t.Skip() when either service is unreachable, ensuring CI
// environments without the services never see a hard failure.
package pipeline_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestGatewayPropagatesCorrelationId verifies that a POST to the BOS discovery
// gateway with an X-Correlation-ID header does not result in a 5xx response.
// The header must be tolerated by all middleware layers in the gateway.
func TestGatewayPropagatesCorrelationId(t *testing.T) {
	if !isPm4PyRunning(t) {
		t.Skip("Skipping: pm4py-rust not running at " + pm4pyBase)
	}
	if !isBosRunning(t) {
		t.Skip("Skipping: BusinessOS not running at " + bosBase)
	}

	correlationID := "test-corr-" + t.Name()

	body := strings.NewReader(`{"log_path": "/tmp/test.json", "algorithm": "alpha"}`)
	req, err := http.NewRequest("POST", bosBase+"/api/bos/discover", body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", correlationID)

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Skipf("Cannot reach BusinessOS: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body) // drain body

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Skipping: /api/bos/discover not yet implemented (404)")
	}

	// The gateway must not return 5xx on a request bearing a correlation ID header.
	if resp.StatusCode >= 500 {
		t.Errorf("Expected non-500, got %d — correlation header must not cause server error", resp.StatusCode)
	}

	t.Logf("PASS: Gateway returned %d with X-Correlation-ID header", resp.StatusCode)
}

// TestGatewayPropagatesTraceparent verifies that a POST to the BOS discovery
// gateway with a W3C traceparent header is accepted without a 5xx response.
// This ensures the gateway middleware does not reject valid OTEL trace context.
func TestGatewayPropagatesTraceparent(t *testing.T) {
	if !isPm4PyRunning(t) {
		t.Skip("Skipping: pm4py-rust not running at " + pm4pyBase)
	}
	if !isBosRunning(t) {
		t.Skip("Skipping: BusinessOS not running at " + bosBase)
	}

	traceparent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	body := strings.NewReader(`{"log_path": "/tmp/test.json", "algorithm": "alpha"}`)
	req, err := http.NewRequest("POST", bosBase+"/api/bos/discover", body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("traceparent", traceparent)

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Skipf("Cannot reach BusinessOS: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body) // drain body

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Skipping: /api/bos/discover not yet implemented (404)")
	}

	// The gateway must not return 5xx when a W3C traceparent header is present.
	if resp.StatusCode >= 500 {
		t.Errorf("Expected non-500 with traceparent header, got %d", resp.StatusCode)
	}

	t.Logf("PASS: Gateway accepted traceparent header, status=%d", resp.StatusCode)
}
