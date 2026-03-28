package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/integrations/pm4py_rust"
	"github.com/stretchr/testify/assert"
)

// TestPM4PyDashboardKPI_MissingEventLog verifies 400 is returned when event_log is absent.
// Chicago TDD: test name matches claim — "missing event_log returns bad request".
func TestPM4PyDashboardKPI_MissingEventLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPM4PyDashboardHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/pm4py/dashboard-kpi", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetDashboardKPI(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Contains(t, body, "error")
}

// TestPM4PyDashboardKPI_ReturnsKPIShape verifies that a well-formed request returns either:
//   - 200 with a ProcessMiningKPIResponse shape (nil client returns zero values), or
//   - 502 if pm4py-rust is unavailable.
func TestPM4PyDashboardKPI_ReturnsKPIShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPM4PyDashboardHandler(nil) // nil client → stats returns nil+nil, yields 200

	body, err := json.Marshal(map[string]interface{}{
		"event_log": map[string]interface{}{
			"traces": []interface{}{},
		},
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/pm4py/dashboard-kpi", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetDashboardKPI(c)

	// Accept 200 or 502 (pm4py not running in CI).
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadGateway,
		"expected 200 or 502, got %d", w.Code)

	if w.Code == http.StatusOK {
		var resp ProcessMiningKPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.FetchedAt, "fetched_at must be set")
		assert.NotNil(t, resp.TopVariants, "top_variants must not be nil")
		assert.NotNil(t, resp.BottleneckActivities, "bottleneck_activities must not be nil")
		assert.NotNil(t, resp.ActivityFrequencies, "activity_frequencies must not be nil")
	}
}

// TestPM4PyDashboardKPI_WithPetriNet verifies that supplying a petri_net does not crash the handler.
func TestPM4PyDashboardKPI_WithPetriNet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPM4PyDashboardHandler(nil)

	body, err := json.Marshal(map[string]interface{}{
		"event_log": map[string]interface{}{
			"traces": []interface{}{},
		},
		"petri_net": map[string]interface{}{
			"places":        []interface{}{},
			"transitions":   []interface{}{},
			"arcs":          []interface{}{},
			"initial_place": nil,
			"final_place":   nil,
		},
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/pm4py/dashboard-kpi", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetDashboardKPI(c)

	// nil client → both calls return nil+nil → 200 with zero KPI values
	assert.Equal(t, http.StatusOK, w.Code)

	var resp ProcessMiningKPIResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.FetchedAt)
}

// TestPM4PyDashboardKPI_BuildResponse_MapsStatisticsFields verifies buildResponse correctly
// maps StatisticsResponse fields into ProcessMiningKPIResponse.
func TestPM4PyDashboardKPI_BuildResponse_MapsStatisticsFields(t *testing.T) {
	handler := NewPM4PyDashboardHandler(nil)

	stats := &pm4py_rust.StatisticsResponse{
		TraceCount:           10,
		EventCount:           100,
		VariantCount:         3,
		ActivityFrequencies:  map[string]int{"A": 5, "B": 3},
		BottleneckActivities: []string{"A"},
		VariantFrequencies:   map[string]int{"v1": 6, "v2": 4},
	}

	resp := handler.buildResponse(stats, nil)

	assert.Equal(t, 100, resp.EventCount)
	assert.Equal(t, 10, resp.TraceCount)
	assert.Equal(t, 3, resp.VariantCount)
	assert.Equal(t, 5, resp.ActivityFrequencies["A"])
	assert.Len(t, resp.BottleneckActivities, 1)
	assert.Equal(t, "A", resp.BottleneckActivities[0].Activity)
	assert.Equal(t, 5, resp.BottleneckActivities[0].Frequency)
	assert.Len(t, resp.TopVariants, 2)
	assert.NotEmpty(t, resp.FetchedAt)
}

// TestPM4PyDashboardKPI_BuildResponse_MapsConformanceFields verifies buildResponse
// correctly maps ConformanceResponse into ProcessMiningKPIResponse.
func TestPM4PyDashboardKPI_BuildResponse_MapsConformanceFields(t *testing.T) {
	handler := NewPM4PyDashboardHandler(nil)

	conf := &pm4py_rust.ConformanceResponse{
		Fitness:      0.95,
		Precision:    0.88,
		IsConformant: true,
	}

	resp := handler.buildResponse(nil, conf)

	assert.InDelta(t, 0.95, resp.ConformanceFitness, 0.001)
	assert.InDelta(t, 0.88, resp.ConformancePrecision, 0.001)
	assert.True(t, resp.IsConformant)
}

// TestPM4PyDashboardKPI_BuildResponse_NilInputsReturnEmptySlices verifies that nil
// stats and conf produce a valid (non-nil collections) response.
func TestPM4PyDashboardKPI_BuildResponse_NilInputsReturnEmptySlices(t *testing.T) {
	handler := NewPM4PyDashboardHandler(nil)

	resp := handler.buildResponse(nil, nil)

	assert.NotNil(t, resp.TopVariants)
	assert.NotNil(t, resp.BottleneckActivities)
	assert.NotNil(t, resp.ActivityFrequencies)
	assert.NotEmpty(t, resp.FetchedAt)
	assert.False(t, resp.IsConformant)
	assert.Equal(t, 0.0, resp.ConformanceFitness)
}

// TestPM4PyDashboardKPI_30sTimeout_DoesNotBlock verifies that a slow pm4py-rust
// service does not block the handler beyond the configured timeout.
func TestPM4PyDashboardKPI_30sTimeout_DoesNotBlock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Start a mock server that delays every response by 500ms.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := pm4py_rust.NewClient(mockServer.URL)
	handler := NewPM4PyDashboardHandlerWithTimeout(client, 100*time.Millisecond)

	body, err := json.Marshal(map[string]interface{}{
		"event_log": map[string]interface{}{
			"traces": []interface{}{},
		},
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/pm4py/dashboard-kpi", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	start := time.Now()
	handler.GetDashboardKPI(c)
	elapsed := time.Since(start)

	// Handler must return before the mock's 500ms delay.
	assert.Less(t, elapsed, 200*time.Millisecond,
		"handler blocked for %v; expected to return within 200ms on 100ms timeout", elapsed)

	// A timed-out statistics call surfaces as 502 Bad Gateway.
	assert.Equal(t, http.StatusBadGateway, w.Code)

	var respBody map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
	assert.Contains(t, respBody, "error")
}
