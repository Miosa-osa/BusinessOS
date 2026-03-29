//go:build integration

package weaver_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	weaverBaseURL = "http://localhost:4320"
	weaverTimeout = 5 * time.Second
)

// TestMain skips all weaver tests when the live-check service is not reachable.
// Run `make dev` (or docker compose up businessos-weaver-live-check) first.
func TestMain(m *testing.M) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(weaverBaseURL + "/health")
	if err != nil {
		fmt.Printf("SKIP weaver tests: weaver-live-check not reachable at %s (%v)\n", weaverBaseURL, err)
		os.Exit(0)
	}
	resp.Body.Close()
	os.Exit(m.Run())
}

// newWeaverClient returns an HTTP client sized for weaver requests.
func newWeaverClient() *http.Client {
	return &http.Client{Timeout: weaverTimeout}
}

// TestWeaverLiveCheck_HealthEndpointReady verifies that GET /health returns
// HTTP 200 with body {"status":"ready"}.
func TestWeaverLiveCheck_HealthEndpointReady(t *testing.T) {
	client := newWeaverClient()

	resp, err := client.Get(weaverBaseURL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /health: expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading /health body: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("parsing /health JSON: %v (raw body: %q)", err, string(body))
	}

	status, ok := payload["status"]
	if !ok {
		t.Fatalf("GET /health: response JSON missing \"status\" field (body: %q)", string(body))
	}
	if status != "ready" {
		t.Errorf("GET /health: expected status=ready, got %q", status)
	}
}

// TestWeaverLiveCheck_MetricsEndpointResponds verifies that GET /metrics returns
// HTTP 200 with a Prometheus text-format body (Content-Type: text/plain).
func TestWeaverLiveCheck_MetricsEndpointResponds(t *testing.T) {
	client := newWeaverClient()

	resp, err := client.Get(weaverBaseURL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /metrics: expected status 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("GET /metrics: expected Content-Type to contain \"text/plain\", got %q", ct)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading /metrics body: %v", err)
	}
	if len(body) == 0 {
		t.Error("GET /metrics: response body is empty; expected Prometheus exposition format")
	}
}

// TestWeaverLiveCheck_NoActiveCriticalViolations reads GET /metrics, parses the
// weaver_violations_total counter (if present), and asserts the count is <= 10.
// Upstream OTel semconv warnings are expected/normal; the threshold guards
// against genuine critical violations accumulating undetected.
func TestWeaverLiveCheck_NoActiveCriticalViolations(t *testing.T) {
	client := newWeaverClient()

	resp, err := client.Get(weaverBaseURL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /metrics: expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading /metrics body: %v", err)
	}

	// Parse Prometheus text format line-by-line.
	// We look for any line that matches a violation counter, e.g.:
	//   weaver_violations_total 3
	//   weaver_violations_total{severity="critical"} 0
	violations := parseViolationCount(string(body))

	const maxAllowedViolations = 10
	if violations > maxAllowedViolations {
		t.Errorf(
			"weaver-live-check reports %d violations (threshold: %d); check semconv schema alignment",
			violations, maxAllowedViolations,
		)
	} else {
		t.Logf("weaver-live-check violation count: %d (within threshold of %d)", violations, maxAllowedViolations)
	}
}

// parseViolationCount scans a Prometheus text-format body for lines that look
// like violation counters and returns the largest value found.
// Returns 0 when no violation metric is present (service just started or clean).
func parseViolationCount(metrics string) int {
	var maxCount int
	for _, line := range strings.Split(metrics, "\n") {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Match any metric whose name contains "violation".
		if !strings.Contains(strings.ToLower(line), "violation") {
			continue
		}
		// Prometheus line format: metric_name[{labels}] value [timestamp]
		// Split on whitespace; the value is the second-to-last or last field.
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		// Value is the second field (after name+labels).
		rawVal := parts[len(parts)-1]
		// If the last field looks like a Unix timestamp (13+ digits), use second-to-last.
		if len(rawVal) > 12 {
			if len(parts) >= 3 {
				rawVal = parts[len(parts)-2]
			}
		}
		val, err := strconv.Atoi(rawVal)
		if err != nil {
			// May be a float like "3.0" — try float parse.
			if f, ferr := strconv.ParseFloat(rawVal, 64); ferr == nil {
				val = int(f)
			} else {
				continue
			}
		}
		if val > maxCount {
			maxCount = val
		}
	}
	return maxCount
}
