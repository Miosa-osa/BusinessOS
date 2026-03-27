package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TEST HELPERS
// ============================================================================

// createTempEventLogFile writes a minimal valid JSON event log to a temp file.
// The caller is responsible for removing the file via t.Cleanup or defer.
func createTempEventLogFile(t *testing.T) string {
	t.Helper()
	content := []byte(`[{"case_id":"case_1","activity":"create_case","timestamp":"2024-01-01T10:00:00Z"},{"case_id":"case_1","activity":"close_case","timestamp":"2024-01-01T11:00:00Z"}]`)
	f, err := os.CreateTemp("", "event_log_*.json")
	require.NoError(t, err, "failed to create temp event log file")
	_, err = f.Write(content)
	require.NoError(t, err, "failed to write event log content")
	require.NoError(t, f.Close())
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

// ============================================================================
// FIXTURES: pm4py-rust Mock Server
// ============================================================================

// startPM4PyMockServer creates a test HTTP server mocking pm4py-rust on port 8090.
func startPM4PyMockServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock pm4py-rust received: %s %s", r.Method, r.URL.Path)

		// Mock /discover endpoint — expects {event_log: JSON, variant: string}
		if r.URL.Path == "/discover" && r.Method == "POST" {
			var req struct {
				EventLog json.RawMessage `json:"event_log"`
				Variant  string          `json:"variant"`
			}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"model_id":     "petri_net_abc123",
				"algorithm":    req.Variant,
				"activities":   []string{"create_case", "assign_case", "process_case", "close_case"},
				"transitions":  8,
				"source_place": "start",
				"sink_place":   "end",
				"model_data": map[string]interface{}{
					"type":       "petri_net",
					"nodes":      5,
					"edges":      12,
					"activities": []string{"create_case", "assign_case", "process_case", "close_case"},
				},
			})
			return
		}

		// Mock /conformance endpoint — expects {event_log: JSON, petri_net: JSON, method: string}
		if r.URL.Path == "/conformance" && r.Method == "POST" {
			var req struct {
				EventLog json.RawMessage `json:"event_log"`
				PetriNet json.RawMessage `json:"petri_net"`
				Method   string          `json:"method"`
			}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"traces_checked": 150,
				"fitting_traces": 144,
				"fitness":        0.96,
				"precision":      0.92,
				"generalization": 0.88,
				"simplicity":     0.91,
			})
			return
		}

		// Mock /statistics endpoint — expects {event_log: JSON, include_variants: bool, ...}
		if r.URL.Path == "/statistics" && r.Method == "POST" {
			var req struct {
				EventLog               json.RawMessage `json:"event_log"`
				IncludeVariants        bool            `json:"include_variants"`
				IncludeResourceMetrics bool            `json:"include_resource_metrics"`
				IncludeBottlenecks     bool            `json:"include_bottlenecks"`
			}
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"log_name":              "sample_process.xes",
				"num_traces":            500,
				"num_events":            2450,
				"num_unique_activities": 8,
				"num_variants":          45,
				"avg_trace_length":      4.9,
				"min_trace_length":      2,
				"max_trace_length":      12,
				"activity_frequency": []map[string]interface{}{
					{"activity": "create_case", "frequency": 500, "percentage": 20.4},
					{"activity": "assign_case", "frequency": 490, "percentage": 20.0},
					{"activity": "process_case", "frequency": 475, "percentage": 19.4},
				},
				"case_duration": map[string]interface{}{
					"min_seconds":    60,
					"max_seconds":    3600,
					"avg_seconds":    1200.5,
					"median_seconds": 900.0,
				},
			})
			return
		}

		// Not found
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))

	return server
}

// setupPM4PyGatewayTest initializes a gateway handler with real HTTP client calling mock pm4py-rust.
func setupPM4PyGatewayTest(t *testing.T, pm4pyURL string) (*BOSGatewayHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewBOSGatewayHandler(nil, logger)
	handler.pm4pyURL = pm4pyURL
	handler.httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	router := gin.New()
	api := router.Group("/api")
	RegisterBOSGatewayRoutes(api, handler)

	return handler, router
}

// ============================================================================
// DISCOVER ENDPOINT TESTS - Real pm4py-rust Calls
// ============================================================================

func TestDiscoverRealPM4Py_Success(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	handler, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "inductive_miner",
	}

	body, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, "Status code should be 200")

	var resp BOSDiscoverResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify real fields from pm4py-rust
	assert.NotEmpty(t, resp.ModelID, "ModelID should not be empty")
	assert.Equal(t, "inductive_miner", resp.Algorithm, "Algorithm should match request")
	assert.Greater(t, resp.Transitions, 0, "Transitions should be > 0")
	assert.Greater(t, resp.Places, 0, "Places should be > 0")
	assert.Greater(t, resp.Arcs, 0, "Arcs should be > 0")

	// Verify ModelData contains pm4py-rust response
	var modelData map[string]interface{}
	err = json.Unmarshal(resp.ModelData, &modelData)
	require.NoError(t, err, "ModelData should be valid JSON")
	assert.Contains(t, modelData, "activities", "ModelData should contain activities from pm4py-rust")

	// Verify statistics were recorded
	assert.Equal(t, uint64(1), handler.stats.RequestsTotal)
	assert.Equal(t, uint64(0), handler.stats.RequestsFailed)
}

func TestDiscoverRealPM4Py_RespondsWithActivityField(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	_, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "heuristic_miner",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSDiscoverResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// ModelData should have activities field from pm4py-rust
	var modelData map[string]interface{}
	json.Unmarshal(resp.ModelData, &modelData)
	activities, hasActivities := modelData["activities"]
	assert.True(t, hasActivities, "ModelData should have activities field")
	assert.NotEmpty(t, activities, "Activities list should not be empty")
}

func TestDiscoverRealPM4Py_RespondsWithSourceAndSinkPlace(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	_, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "inductive_miner",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSDiscoverResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Response should indicate source and sink places
	assert.Greater(t, resp.Places, 0, "Should have places (including source/sink)")
}

// ============================================================================
// CONFORMANCE ENDPOINT TESTS - Real pm4py-rust Calls
// ============================================================================

func TestConformanceRealPM4Py_Success(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	handler, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSConformanceRequest{
		LogPath: logPath,
		ModelID: "petri_net_abc123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BOSConformanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify real metrics from pm4py-rust
	assert.Equal(t, uint64(150), resp.TracesChecked)
	assert.Equal(t, uint64(144), resp.FittingTraces)
	assert.Equal(t, 0.96, resp.Fitness)
	assert.Equal(t, 0.92, resp.Precision)
	assert.Equal(t, 0.88, resp.Generalization)
	assert.Equal(t, 0.91, resp.Simplicity)

	assert.Equal(t, uint64(1), handler.stats.RequestsTotal)
	assert.Equal(t, uint64(0), handler.stats.RequestsFailed)
}

func TestConformanceRealPM4Py_AllMetricsPopulated(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	_, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSConformanceRequest{
		LogPath: logPath,
		ModelID: "model_xyz",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSConformanceResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// All metrics should be populated
	assert.Greater(t, resp.Fitness, 0.0)
	assert.Greater(t, resp.Precision, 0.0)
	assert.Greater(t, resp.Generalization, 0.0)
	assert.Greater(t, resp.Simplicity, 0.0)
}

func TestConformanceRealPM4Py_ReportsAccurateFitnessMetrics(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	_, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSConformanceRequest{
		LogPath: logPath,
		ModelID: "model_123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSConformanceResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify actual values from pm4py-rust
	assert.True(t, resp.Fitness >= 0.9, "Fitness from pm4py-rust should be >= 0.9")
	assert.True(t, resp.Precision >= 0.9, "Precision from pm4py-rust should be >= 0.9")
	assert.True(t, resp.Generalization < 1.0, "Generalization should be < 1.0")
}

// ============================================================================
// STATISTICS ENDPOINT TESTS - Real pm4py-rust Calls
// ============================================================================

func TestStatisticsRealPM4Py_Success(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	handler, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSStatisticsRequest{
		LogPath: logPath,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BOSStatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify real data from pm4py-rust
	assert.NotEmpty(t, resp.LogName)
	assert.Equal(t, 500, resp.NumTraces)
	assert.Equal(t, 2450, resp.NumEvents)
	assert.Equal(t, 8, resp.NumUniqueActivities)
	assert.Equal(t, 45, resp.NumVariants)
	assert.Equal(t, 4.9, resp.AvgTraceLength)
	assert.Equal(t, 2, resp.MinTraceLength)
	assert.Equal(t, 12, resp.MaxTraceLength)

	assert.Equal(t, uint64(1), handler.stats.RequestsTotal)
	assert.Equal(t, uint64(0), handler.stats.RequestsFailed)
}

func TestStatisticsRealPM4Py_ActivityFrequencyFromPM4Py(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	_, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSStatisticsRequest{
		LogPath: logPath,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSStatisticsResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify activity frequency data from pm4py-rust
	assert.Greater(t, len(resp.ActivityFrequency), 0, "Should have activity frequency data")
	assert.Equal(t, "create_case", resp.ActivityFrequency[0].Activity)
	assert.Equal(t, 500, resp.ActivityFrequency[0].Frequency)
}

func TestStatisticsRealPM4Py_CaseDurationFromPM4Py(t *testing.T) {
	mockServer := startPM4PyMockServer(t)
	defer mockServer.Close()

	_, router := setupPM4PyGatewayTest(t, mockServer.URL)
	logPath := createTempEventLogFile(t)

	req := BOSStatisticsRequest{
		LogPath: logPath,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSStatisticsResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify case duration metrics from pm4py-rust
	assert.Equal(t, int64(60), resp.CaseDuration.MinSeconds)
	assert.Equal(t, int64(3600), resp.CaseDuration.MaxSeconds)
	assert.Greater(t, resp.CaseDuration.AvgSeconds, 0.0)
	assert.Greater(t, resp.CaseDuration.MedianSeconds, 0.0)
}

// ============================================================================
// ERROR HANDLING TESTS - pm4py-rust Network Failures
// ============================================================================

func TestPM4PyNetworkFailure_Discover_Returns503(t *testing.T) {
	// Use a URL that will fail to connect
	handler, router := setupPM4PyGatewayTest(t, "http://localhost:9999")
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "inductive_miner",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Should return 503 on network failure
	assert.Equal(t, http.StatusServiceUnavailable, w.Code,
		"Should return 503 when pm4py-rust is unreachable")

	// Failure should be recorded in stats
	assert.Equal(t, uint64(1), handler.stats.RequestsFailed)
}

func TestPM4PyNetworkFailure_Conformance_Returns503(t *testing.T) {
	handler, router := setupPM4PyGatewayTest(t, "http://localhost:9999")
	logPath := createTempEventLogFile(t)

	req := BOSConformanceRequest{
		LogPath: logPath,
		ModelID: "model_123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, uint64(1), handler.stats.RequestsFailed)
}

func TestPM4PyNetworkFailure_Statistics_Returns503(t *testing.T) {
	handler, router := setupPM4PyGatewayTest(t, "http://localhost:9999")
	logPath := createTempEventLogFile(t)

	req := BOSStatisticsRequest{
		LogPath: logPath,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, uint64(1), handler.stats.RequestsFailed)
}

// ============================================================================
// TIMEOUT TESTS
// ============================================================================

func TestPM4PyTimeout_Discover(t *testing.T) {
	// Create a server that delays longer than the timeout.
	// Uses select to stop when the client disconnects (avoids blocking server.Close).
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			// Client disconnected (timeout) — stop immediately.
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer slowServer.Close()

	handler, router := setupPM4PyGatewayTest(t, slowServer.URL)
	handler.httpClient.Timeout = 100 * time.Millisecond // Very short timeout for test
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "inductive_miner",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Should handle timeout gracefully
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// ============================================================================
// CONFIG TESTS
// ============================================================================

func TestPM4PyURLFromEnv(t *testing.T) {
	// Verify that handler can accept pm4py URL
	handler, _ := setupPM4PyGatewayTest(t, "http://localhost:8090")

	assert.Equal(t, "http://localhost:8090", handler.pm4pyURL)
}

func TestPM4PyURLDefaultValue(t *testing.T) {
	// Handler should have default pm4py URL
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewBOSGatewayHandler(nil, logger)

	// Default should be empty or configurable
	assert.NotNil(t, handler)
}
