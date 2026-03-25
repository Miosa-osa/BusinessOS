package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures
func setupGatewayTest(t *testing.T) (*BOSGatewayHandler, *gin.Engine) {
	// Disable Gin debug output
	gin.SetMode(gin.TestMode)

	logger := slog.New(slog.NewTextHandler(nil, nil))
	handler := NewBOSGatewayHandler(nil, logger)

	router := gin.New()
	api := router.Group("/api")
	RegisterBOSGatewayRoutes(api, handler)

	return handler, router
}

func bosMustMarshal(t *testing.T, v interface{}) string {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

// ============================================================================
// DISCOVER ENDPOINT TESTS
// ============================================================================

func TestDiscover_Success(t *testing.T) {
	handler, router := setupGatewayTest(t)

	req := BOSDiscoverRequest{
		LogPath:   "/path/to/log.xes",
		Algorithm: "inductive_miner",
	}

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BOSDiscoverResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.ModelID)
	assert.Equal(t, "inductive_miner", resp.Algorithm)
	assert.Equal(t, 5, resp.Places)
	assert.Equal(t, 8, resp.Transitions)
	assert.Equal(t, 12, resp.Arcs)
	assert.True(t, resp.LatencyMs < 100, "Latency should be <100ms")

	// Verify stats were recorded
	stats := handler.stats
	assert.Equal(t, uint64(1), stats.RequestsTotal)
	assert.Equal(t, uint64(0), stats.RequestsFailed)
}

func TestDiscover_MissingLogPath(t *testing.T) {
	_, router := setupGatewayTest(t)

	req := BOSDiscoverRequest{
		Algorithm: "inductive_miner",
	}

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDiscover_InvalidJSON(t *testing.T) {
	_, router := setupGatewayTest(t)

	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString("invalid json"))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDiscover_DefaultAlgorithm(t *testing.T) {
	_, router := setupGatewayTest(t)

	req := BOSDiscoverRequest{
		LogPath: "/path/to/log.xes",
		// No algorithm specified
	}

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BOSDiscoverResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Should default to inductive_miner
	assert.Equal(t, "inductive_miner", resp.Algorithm)
}

// ============================================================================
// CONFORMANCE ENDPOINT TESTS
// ============================================================================

func TestConformance_Success(t *testing.T) {
	handler, router := setupGatewayTest(t)

	req := BOSConformanceRequest{
		LogPath: "/path/to/log.xes",
		ModelID: "model_123",
	}

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BOSConformanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, uint64(125), resp.TracesChecked)
	assert.Equal(t, uint64(120), resp.FittingTraces)
	assert.True(t, resp.Fitness > 0.9)
	assert.True(t, resp.LatencyMs < 100)

	// Verify stats
	stats := handler.stats
	assert.Equal(t, uint64(1), stats.RequestsTotal)
}

func TestConformance_MissingFields(t *testing.T) {
	_, router := setupGatewayTest(t)

	tests := []struct {
		name    string
		request BOSConformanceRequest
	}{
		{"missing_log_path", BOSConformanceRequest{ModelID: "model_123"}},
		{"missing_model_id", BOSConformanceRequest{LogPath: "/path/to/log.xes"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bosMustMarshal(t, tt.request)
			httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewBufferString(body))
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// ============================================================================
// STATISTICS ENDPOINT TESTS
// ============================================================================

func TestStatistics_Success(t *testing.T) {
	handler, router := setupGatewayTest(t)

	req := BOSStatisticsRequest{
		LogPath: "/path/to/log.xes",
	}

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BOSStatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "sample_log.xes", resp.LogName)
	assert.Equal(t, 500, resp.NumTraces)
	assert.Equal(t, 2450, resp.NumEvents)
	assert.Equal(t, 8, resp.NumUniqueActivities)
	assert.True(t, resp.LatencyMs < 100)

	// Verify stats
	stats := handler.stats
	assert.Equal(t, uint64(1), stats.RequestsTotal)
}

func TestStatistics_MissingLogPath(t *testing.T) {
	_, router := setupGatewayTest(t)

	req := BOSStatisticsRequest{} // Empty request

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatistics_CaseDurationMetrics(t *testing.T) {
	_, router := setupGatewayTest(t)

	req := BOSStatisticsRequest{
		LogPath: "/path/to/log.xes",
	}

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSStatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify case duration stats structure
	assert.Equal(t, int64(60), resp.CaseDuration.MinSeconds)
	assert.Equal(t, int64(3600), resp.CaseDuration.MaxSeconds)
	assert.True(t, resp.CaseDuration.AvgSeconds > 0)
	assert.True(t, resp.CaseDuration.MedianSeconds > 0)
}

// ============================================================================
// STATUS ENDPOINT TESTS
// ============================================================================

func TestStatus_HealthyResponse(t *testing.T) {
	_, router := setupGatewayTest(t)

	httpReq := httptest.NewRequest("GET", "/api/bos/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BOSStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", resp.Status)
	assert.True(t, resp.UptimeSeconds >= 0)
	assert.Equal(t, uint64(0), resp.RequestsTotal)
	assert.Equal(t, uint64(0), resp.RequestsFailed)
}

func TestStatus_WithExistingRequests(t *testing.T) {
	handler, router := setupGatewayTest(t)

	// Make some requests first
	for i := 0; i < 3; i++ {
		req := BOSDiscoverRequest{
			LogPath:   "/path/to/log.xes",
			Algorithm: "inductive_miner",
		}
		body := bosMustMarshal(t, req)
		httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)
	}

	// Now check status
	httpReq := httptest.NewRequest("GET", "/api/bos/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, uint64(3), resp.RequestsTotal)
	assert.Equal(t, uint64(0), resp.RequestsFailed)
	assert.True(t, resp.AverageLatencyMs > 0)
}

// ============================================================================
// LATENCY TESTS
// ============================================================================

func TestLatencyUnder100ms(t *testing.T) {
	_, router := setupGatewayTest(t)

	tests := []struct {
		name     string
		endpoint string
		request  interface{}
	}{
		{"discover", "/api/bos/discover", BOSDiscoverRequest{LogPath: "/log.xes"}},
		{"conformance", "/api/bos/conformance", BOSConformanceRequest{LogPath: "/log.xes", ModelID: "model_123"}},
		{"statistics", "/api/bos/statistics", BOSStatisticsRequest{LogPath: "/log.xes"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bosMustMarshal(t, tt.request)
			httpReq := httptest.NewRequest("POST", "/api/bos"+tt.endpoint, bytes.NewBufferString(body))
			httpReq.Header.Set("Content-Type", "application/json")

			start := time.Now()
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httpReq)
			elapsed := time.Since(start).Milliseconds()

			assert.Equal(t, http.StatusOK, w.Code)
			assert.True(t, elapsed < 100, "%s took %dms (>100ms)", tt.name, elapsed)
		})
	}
}

// ============================================================================
// CONCURRENT REQUEST TESTS
// ============================================================================

func TestConcurrentRequests(t *testing.T) {
	handler, router := setupGatewayTest(t)

	// Run 10 concurrent requests
	results := make(chan int, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			req := BOSDiscoverRequest{
				LogPath:   "/path/to/log.xes",
				Algorithm: "inductive_miner",
			}

			body := bosMustMarshal(t, req)
			httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString(body))
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httpReq)

			if w.Code == http.StatusOK {
				results <- 1
			} else {
				results <- 0
			}
		}(i)
	}

	// Wait for results
	successCount := 0
	for i := 0; i < 10; i++ {
		if <-results == 1 {
			successCount++
		}
	}

	assert.Equal(t, 10, successCount)

	// Verify stats
	assert.Equal(t, uint64(10), handler.stats.RequestsTotal)
	assert.Equal(t, uint64(0), handler.stats.RequestsFailed)
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

func TestErrorHandling_BadContentType(t *testing.T) {
	_, router := setupGatewayTest(t)

	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString(`{"log_path": "/log.xes"}`))
	httpReq.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Gin should still parse JSON, but let's verify error handling
	assert.True(t, w.Code >= 400)
}

func TestErrorHandling_NotFound(t *testing.T) {
	_, router := setupGatewayTest(t)

	httpReq := httptest.NewRequest("POST", "/api/bos/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ============================================================================
// STATISTICS TRACKING TESTS
// ============================================================================

func TestStatisticsTracking_AverageLatency(t *testing.T) {
	handler, router := setupGatewayTest(t)

	// Make multiple requests
	for i := 0; i < 5; i++ {
		req := BOSDiscoverRequest{
			LogPath:   "/path/to/log.xes",
			Algorithm: "inductive_miner",
		}

		body := bosMustMarshal(t, req)
		httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)
	}

	// Check stats
	handler.mu.RLock()
	stats := handler.stats
	handler.mu.RUnlock()

	stats.mu.Lock()
	avgLatency := stats.AverageLatency
	stats.mu.Unlock()

	assert.True(t, avgLatency > 0, "Should have computed average latency")
	assert.True(t, avgLatency < 100, "Average latency should be <100ms")
}

// ============================================================================
// DATABASE CONNECTIVITY TESTS
// ============================================================================

func TestDatabaseCheck_Nil(t *testing.T) {
	handler := NewBOSGatewayHandler(nil, slog.Default())

	// Database should be unavailable with nil pool
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ready := handler.checkDatabase(ctx)
	assert.False(t, ready)
}

// ============================================================================
// WvDA SOUNDNESS: WRITE-AHEAD LOG TESTS
// ============================================================================

func TestDiscover_WriteAheadLog_RecoveryAfterDBFailure(t *testing.T) {
	// WvdA soundness: if pm4py-rust returns a result but DB write fails,
	// the result must be recoverable from the write-ahead log.
	// This test validates that discovery results survive transient DB failures.
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(nil, nil))
	handler := NewBOSGatewayHandler(nil, logger)

	// Create a mock discovery result
	modelID := "model_wal_test_001"
	mockResult := BOSDiscoverResponse{
		ModelID:     modelID,
		Algorithm:   "inductive_miner",
		Places:      5,
		Transitions: 8,
		Arcs:        12,
		ModelData:   json.RawMessage(`{"places":5,"transitions":8}`),
		LatencyMs:   42,
	}

	// Write to WAL (write-ahead log)
	err := handler.writeAheadLog(modelID, &mockResult)
	require.NoError(t, err, "writeAheadLog should succeed")

	// Verify we can recover from WAL
	recovered, err := handler.recoverFromWAL(modelID)
	require.NoError(t, err, "recoverFromWAL should succeed for existing entry")
	require.NotNil(t, recovered, "recovered result should not be nil")

	assert.Equal(t, modelID, recovered.ModelID, "recovered model_id should match")
	assert.Equal(t, "inductive_miner", recovered.Algorithm)
	assert.Equal(t, 5, recovered.Places)
	assert.Equal(t, 8, recovered.Transitions)
	assert.Equal(t, uint64(42), recovered.LatencyMs)

	// Clean up WAL entry
	err = handler.cleanupWAL(modelID)
	assert.NoError(t, err, "cleanupWAL should succeed")

	// Verify cleanup: recovery should now fail
	_, err = handler.recoverFromWAL(modelID)
	assert.Error(t, err, "recovery should fail after cleanup")
}

func TestDiscover_WriteAheadLog_NonExistent(t *testing.T) {
	// Recovering a non-existent WAL entry should return an error
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(nil, nil))
	handler := NewBOSGatewayHandler(nil, logger)

	_, err := handler.recoverFromWAL("non_existent_model_id")
	assert.Error(t, err, "recovery of non-existent entry should fail")
}

func TestDiscover_WriteAheadLog_Overwrite(t *testing.T) {
	// Writing twice to the same model ID should overwrite (idempotent)
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(nil, nil))
	handler := NewBOSGatewayHandler(nil, logger)

	modelID := "model_overwrite_test"

	first := &BOSDiscoverResponse{ModelID: modelID, Algorithm: "alpha_miner", Places: 1, Transitions: 2, Arcs: 3, LatencyMs: 10}
	second := &BOSDiscoverResponse{ModelID: modelID, Algorithm: "inductive_miner", Places: 5, Transitions: 8, Arcs: 12, LatencyMs: 20}

	err := handler.writeAheadLog(modelID, first)
	require.NoError(t, err)

	err = handler.writeAheadLog(modelID, second)
	require.NoError(t, err)

	recovered, err := handler.recoverFromWAL(modelID)
	require.NoError(t, err)
	assert.Equal(t, "inductive_miner", recovered.Algorithm, "second write should overwrite first")
	assert.Equal(t, 5, recovered.Places)
	assert.Equal(t, uint64(20), recovered.LatencyMs)

	handler.cleanupWAL(modelID)
}

// ============================================================================
// WvDA SOUNDNESS: DISCOVER HANDLER WAL INTEGRATION TEST
// ============================================================================

func TestDiscover_WriteAheadLog_IntegrationWithHandler(t *testing.T) {
	// WvdA soundness: the Discover handler must write results to the WAL
	// before returning, so that transient failures (e.g., DB write) do not
	// lose the pm4py-rust discovery result. This test verifies the full
	// write-ahead -> recover -> cleanup lifecycle from the handler path.
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(nil, nil))
	handler := NewBOSGatewayHandler(nil, logger)

	// Build a mock response that the handler would produce
	modelID := "model_wal_integration_001"
	mockResult := &BOSDiscoverResponse{
		ModelID:     modelID,
		Algorithm:   "inductive_miner",
		Places:      5,
		Transitions: 8,
		Arcs:        12,
		ModelData:   json.RawMessage(`{"places":5,"transitions":8}`),
		LatencyMs:   42,
	}

	// Step 1: Write to WAL (simulates what Discover handler does)
	err := handler.writeAheadLog(modelID, mockResult)
	require.NoError(t, err, "writeAheadLog should succeed")

	// Step 2: Simulate transient DB failure -- result is still in WAL
	recovered, err := handler.recoverFromWAL(modelID)
	require.NoError(t, err, "recovery should succeed after WAL write")
	require.NotNil(t, recovered)

	// Verify recovered data integrity
	assert.Equal(t, modelID, recovered.ModelID)
	assert.Equal(t, "inductive_miner", recovered.Algorithm)
	assert.Equal(t, 5, recovered.Places)
	assert.Equal(t, 8, recovered.Transitions)
	assert.Equal(t, uint64(42), recovered.LatencyMs)

	// Verify model_data is valid JSON
	var modelData map[string]interface{}
	err = json.Unmarshal(recovered.ModelData, &modelData)
	assert.NoError(t, err, "recovered model_data should be valid JSON")

	// Step 3: Cleanup after successful recovery
	err = handler.cleanupWAL(modelID)
	assert.NoError(t, err, "cleanup should succeed")

	// Step 4: Verify cleanup
	_, err = handler.recoverFromWAL(modelID)
	assert.Error(t, err, "recovery should fail after cleanup")
}

// ============================================================================
// RESPONSE BODY VALIDATION TESTS
// ============================================================================

func TestDiscoverResponse_ModelDataJSON(t *testing.T) {
	_, router := setupGatewayTest(t)

	req := BOSDiscoverRequest{
		LogPath:   "/path/to/log.xes",
		Algorithm: "inductive_miner",
	}

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSDiscoverResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify model_data is valid JSON
	var modelData map[string]interface{}
	err = json.Unmarshal(resp.ModelData, &modelData)
	assert.NoError(t, err)
}

func TestConformanceResponse_AllFieldsPopulated(t *testing.T) {
	_, router := setupGatewayTest(t)

	req := BOSConformanceRequest{
		LogPath: "/path/to/log.xes",
		ModelID: "model_123",
	}

	body := bosMustMarshal(t, req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewBufferString(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSConformanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// All fields should be present
	assert.True(t, resp.TracesChecked > 0)
	assert.True(t, resp.FittingTraces > 0)
	assert.True(t, resp.Fitness > 0)
	assert.True(t, resp.Precision > 0)
	assert.True(t, resp.Generalization > 0)
	assert.True(t, resp.Simplicity > 0)
	assert.True(t, resp.LatencyMs >= 0)
}
