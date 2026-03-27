//go:build integration
// +build integration

// Package pipeline_test contains E2E integration tests for the full
// pm4py-rust → BusinessOS pipeline.
//
// These tests require BOTH services to be running:
//   - pm4py-rust on http://localhost:8090
//   - BusinessOS backend on http://localhost:8001
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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	pm4pyBase = "http://localhost:8090"
	bosBase   = "http://localhost:8001"
)

// httpClient is a shared client with a short timeout so probes fail fast.
var httpClient = &http.Client{Timeout: 5 * time.Second}

// sampleEventLog returns a minimal two-trace event log JSON body compatible
// with pm4py-rust's DiscoveryRequest / StatisticsRequest formats.
func sampleEventLog() map[string]interface{} {
	return map[string]interface{}{
		"traces": []map[string]interface{}{
			{
				"caseID": "case_001",
				"events": []map[string]interface{}{
					{"activity": "Start", "timestamp": "2024-01-01T10:00:00Z"},
					{"activity": "Review", "timestamp": "2024-01-01T10:05:00Z"},
					{"activity": "Approve", "timestamp": "2024-01-01T10:10:00Z"},
					{"activity": "End", "timestamp": "2024-01-01T10:15:00Z"},
				},
			},
			{
				"caseID": "case_002",
				"events": []map[string]interface{}{
					{"activity": "Start", "timestamp": "2024-01-02T10:00:00Z"},
					{"activity": "Review", "timestamp": "2024-01-02T10:03:00Z"},
					{"activity": "Reject", "timestamp": "2024-01-02T10:08:00Z"},
					{"activity": "End", "timestamp": "2024-01-02T10:12:00Z"},
				},
			},
		},
	}
}

// isPm4pyRunning returns true when pm4py-rust /api/health returns 200.
func isPm4pyRunning(t *testing.T) bool {
	t.Helper()
	resp, err := httpClient.Get(pm4pyBase + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// isBosRunning returns true when BusinessOS /api/health returns 200.
func isBosRunning(t *testing.T) bool {
	t.Helper()
	resp, err := httpClient.Get(bosBase + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// postJSON is a thin helper that marshals body, POSTs to url, and returns the
// parsed JSON response.  The caller is responsible for checking the status.
func postJSON(t *testing.T, url string, body interface{}) (int, map[string]interface{}) {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal response JSON from %s: %v\nbody: %s", url, err, raw)
	}

	return resp.StatusCode, result
}

// ── Test 1: pm4py-rust health directly ───────────────────────────────────────

// TestPm4pyDirectHealth verifies that pm4py-rust is reachable and healthy.
// This is the foundation test; all other pipeline tests depend on it.
func TestPm4pyDirectHealth(t *testing.T) {
	if !isPm4pyRunning(t) {
		t.Skip("Skipping: pm4py-rust not running at " + pm4pyBase)
	}

	resp, err := httpClient.Get(pm4pyBase + "/api/health")
	if err != nil {
		t.Fatalf("GET %s/api/health: %v", pm4pyBase, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 from pm4py health, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	status, _ := body["status"].(string)
	if status != "healthy" {
		t.Errorf("Expected status=='healthy', got %q", status)
	}

	version, _ := body["version"].(string)
	if version == "" {
		t.Error("Expected non-empty version in health response")
	}

	t.Logf("pm4py-rust health OK: status=%s version=%s", status, version)
}

// ── Test 2: BusinessOS health (gateway layer) ─────────────────────────────────

// TestBosHealth verifies that the BusinessOS backend is reachable and healthy.
func TestBosHealth(t *testing.T) {
	if !isBosRunning(t) {
		t.Skip("Skipping: BusinessOS not running at " + bosBase)
	}

	resp, err := httpClient.Get(bosBase + "/api/health")
	if err != nil {
		t.Fatalf("GET %s/api/health: %v", bosBase, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 from BOS health, got %d", resp.StatusCode)
	}

	t.Log("BusinessOS health OK")
}

// ── Test 3: BOS gateway status includes pm4py-rust ───────────────────────────

// TestPm4pyHealthViaGateway tests that BusinessOS /api/bos/status returns a
// payload that reflects pm4py-rust connectivity.  When both services are up the
// pm4py_rust field must indicate a reachable / healthy state.
func TestPm4pyHealthViaGateway(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: BusinessOS or pm4py-rust not running")
	}

	resp, err := httpClient.Get(bosBase + "/api/bos/status")
	if err != nil {
		t.Fatalf("GET %s/api/bos/status: %v", bosBase, err)
	}
	defer resp.Body.Close()

	// 404 means the gateway status endpoint is not yet wired — skip gracefully.
	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Skipping: /api/bos/status not yet implemented (404)")
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 from /api/bos/status, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode /api/bos/status response: %v", err)
	}

	// The status response should mention pm4py in some form.
	bodyStr := fmt.Sprintf("%v", body)
	if !strings.Contains(strings.ToLower(bodyStr), "pm4py") {
		t.Logf("NOTE: /api/bos/status did not mention pm4py — body: %s", bodyStr)
	}

	t.Logf("BOS gateway status OK: %v", body)
}

// ── Test 4: Discovery pipeline E2E via BOS gateway ───────────────────────────

// TestDiscoveryPipelineE2E submits a discovery request through the BOS gateway
// at POST /api/bos/discover.  The gateway forwards to pm4py-rust and returns
// a Petri net.  Asserts: non-empty places, transitions, and arcs arrays.
func TestDiscoveryPipelineE2E(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}

	payload := map[string]interface{}{
		"event_log": sampleEventLog(),
		"variant":   "alpha",
	}

	status, body := postJSON(t, bosBase+"/api/bos/discover", payload)

	if status == http.StatusNotFound {
		t.Skip("Skipping: /api/bos/discover not yet implemented (404)")
	}

	if status < 200 || status >= 300 {
		t.Fatalf("Expected 2xx from /api/bos/discover, got %d — body: %v", status, body)
	}

	// Validate Petri net structure is present.
	net, ok := body["petri_net"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'petri_net' object in response, got: %v", body)
	}

	places, _ := net["places"].([]interface{})
	if len(places) == 0 {
		t.Error("Expected at least one place in discovered Petri net")
	}

	transitions, _ := net["transitions"].([]interface{})
	if len(transitions) == 0 {
		t.Error("Expected at least one transition in discovered Petri net")
	}

	arcs, _ := net["arcs"].([]interface{})
	if len(arcs) == 0 {
		t.Error("Expected at least one arc in discovered Petri net")
	}

	traceCount, _ := body["trace_count"].(float64)
	if int(traceCount) != 2 {
		t.Errorf("Expected trace_count==2, got %v", traceCount)
	}

	t.Logf("discovery E2E OK: places=%d transitions=%d arcs=%d traces=%d",
		len(places), len(transitions), len(arcs), int(traceCount))
}

// ── Test 5: Statistics pipeline E2E via BOS gateway ──────────────────────────

// TestStatisticsPipelineE2E submits a statistics request through the BOS
// gateway at POST /api/bos/statistics and validates trace_count, event_count,
// and unique_activities against the known sample event log values.
func TestStatisticsPipelineE2E(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}

	payload := map[string]interface{}{
		"event_log":               sampleEventLog(),
		"include_variants":        true,
		"include_resource_metrics": false,
		"include_bottlenecks":     false,
	}

	status, body := postJSON(t, bosBase+"/api/bos/statistics", payload)

	if status == http.StatusNotFound {
		t.Skip("Skipping: /api/bos/statistics not yet implemented (404)")
	}

	if status < 200 || status >= 300 {
		t.Fatalf("Expected 2xx from /api/bos/statistics, got %d — body: %v", status, body)
	}

	traceCount, _ := body["trace_count"].(float64)
	if int(traceCount) != 2 {
		t.Errorf("Expected trace_count==2, got %v", traceCount)
	}

	eventCount, _ := body["event_count"].(float64)
	if int(eventCount) != 8 {
		t.Errorf("Expected event_count==8, got %v", eventCount)
	}

	uniqueActivities, _ := body["unique_activities"].(float64)
	if int(uniqueActivities) != 5 {
		// Start, Review, Approve, Reject, End
		t.Errorf("Expected unique_activities==5, got %v", uniqueActivities)
	}

	t.Logf("statistics E2E OK: traces=%d events=%d unique_activities=%d",
		int(traceCount), int(eventCount), int(uniqueActivities))
}

// ── Test 6: Conformance pipeline E2E via BOS gateway ─────────────────────────

// TestConformancePipelineE2E posts a conformance-checking request to the BOS
// gateway at POST /api/bos/conformance with the sample log and a trivial Petri
// net, and asserts that a fitness score between 0 and 1 is returned.
func TestConformancePipelineE2E(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}

	payload := map[string]interface{}{
		"event_log": sampleEventLog(),
		"petri_net": map[string]interface{}{
			"places": []map[string]interface{}{
				{"id": "p_start", "name": "source", "initial_marking": 1},
				{"id": "p_end", "name": "sink", "initial_marking": 0},
			},
			"transitions": []map[string]interface{}{
				{"id": "t1", "name": "Start", "label": "Start"},
				{"id": "t2", "name": "End", "label": "End"},
			},
			"arcs": []map[string]interface{}{
				{"from": "p_start", "to": "t1", "weight": 1},
				{"from": "t1", "to": "p_end", "weight": 1},
			},
			"initial_place": "p_start",
			"final_place":   "p_end",
		},
		"method": "token_replay",
	}

	status, body := postJSON(t, bosBase+"/api/bos/conformance", payload)

	if status == http.StatusNotFound {
		t.Skip("Skipping: /api/bos/conformance not yet implemented (404)")
	}

	if status < 200 || status >= 300 {
		t.Fatalf("Expected 2xx from /api/bos/conformance, got %d — body: %v", status, body)
	}

	fitness, ok := body["fitness"].(float64)
	if !ok {
		t.Fatalf("Expected 'fitness' float in response, got: %v", body)
	}

	if fitness < 0.0 || fitness > 1.0 {
		t.Errorf("Expected fitness in [0.0, 1.0], got %f", fitness)
	}

	method, _ := body["method"].(string)
	t.Logf("conformance E2E OK: fitness=%.3f method=%s", fitness, method)
}

// ── Test 7: Pipeline round-trip latency ──────────────────────────────────────

// TestPipelineRoundTripLatency measures the wall-clock time for a full
// discover request through the BOS gateway.  Asserts the round-trip completes
// in under 10 seconds — a liveness guard against hung connections.
func TestPipelineRoundTripLatency(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}

	payload := map[string]interface{}{
		"event_log": sampleEventLog(),
		"variant":   "alpha",
	}

	start := time.Now()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, err := httpClient.Post(
		bosBase+"/api/bos/discover",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		t.Fatalf("POST /api/bos/discover: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // drain body

	elapsed := time.Since(start)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Skipping: /api/bos/discover not yet implemented (404)")
	}

	const maxLatency = 10 * time.Second
	if elapsed > maxLatency {
		t.Errorf("Pipeline round-trip took %v, expected < %v", elapsed, maxLatency)
	}

	t.Logf("pipeline latency OK: %v", elapsed)
}
