package handlers

// Chicago TDD: PM4PyRustHandler tests — RED → GREEN → REFACTOR
//
// Tests cover all 4 HTTP handler endpoints:
//   T1: Health success — upstream returns 200
//   T2: Health service-down — upstream returns 500, handler returns 500
//   T3: Discover success — upstream returns petri net
//   T4: Discover missing event_log — handler validates before calling upstream → 400
//   T5: Discover service-down — upstream returns 503, handler returns 500
//   T6: Conformance success — upstream returns fitness/precision
//   T7: Statistics success — upstream returns trace/variant counts
//   T8: Statistics service-down — error response has code + message

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/integrations/pm4py_rust"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rustSetupRouter creates a test router pointed at the given baseURL mock.
// Retries: 1 keeps service-down tests fast (~100ms backoff, not 700ms).
func rustSetupRouter(baseURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	client := pm4py_rust.NewClientWithConfig(baseURL, pm4py_rust.ClientConfig{
		Timeout: 2 * time.Second,
		Retries: 1,
	})
	handler := &PM4PyRustHandler{
		client: client,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	router := gin.New()
	api := router.Group("/api")
	handler.RegisterRoutes(api)
	return router
}

// ─── T1: Health success ──────────────────────────────────────────────────────

func TestPM4PyRustHandler_Health_Success(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok", "version": "1.2.3",
		})
	}))
	defer mock.Close()

	router := rustSetupRouter(mock.URL)

	req := httptest.NewRequest("GET", "/api/pm4py/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp pm4py_rust.HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Status)
	assert.NotEmpty(t, resp.Version)
}

// ─── T2: Health service-down ─────────────────────────────────────────────────

func TestPM4PyRustHandler_Health_ServiceDown(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "pm4py-rust unavailable"})
	}))
	defer mock.Close()

	router := rustSetupRouter(mock.URL)

	req := httptest.NewRequest("GET", "/api/pm4py/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// ─── T3: Discover success ─────────────────────────────────────────────────────

func TestPM4PyRustHandler_Discover_Success(t *testing.T) {
	label := "A"
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pm4py_rust.DiscoveryResponse{
			Algorithm: "alpha",
			PetriNet: pm4py_rust.PetriNetJSON{
				Places:      []pm4py_rust.PlaceJSON{{ID: "p0", Name: "start"}},
				Transitions: []pm4py_rust.TransitionJSON{{ID: "t0", Name: "A", Label: &label}},
				Arcs:        []pm4py_rust.ArcJSON{{From: "p0", To: "t0", Weight: 1}},
			},
			TraceCount: 5,
		})
	}))
	defer mock.Close()

	router := rustSetupRouter(mock.URL)

	body := `{"event_log":[{"case_id":"c1","activity":"A","timestamp":"2024-01-01T10:00:00Z"}]}`
	req := httptest.NewRequest("POST", "/api/pm4py/discover", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp pm4py_rust.DiscoveryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "alpha", resp.Algorithm)
	assert.NotNil(t, resp.PetriNet.Places)
	assert.NotNil(t, resp.PetriNet.Transitions)
}

// ─── T4: Discover missing event_log → 400 from binding validation ────────────

func TestPM4PyRustHandler_Discover_MissingEventLog(t *testing.T) {
	// Handler validates request before calling upstream; unreachable addr never contacted.
	router := rustSetupRouter("http://127.0.0.1:0")

	body := `{}` // missing required event_log field
	req := httptest.NewRequest("POST", "/api/pm4py/discover", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// ─── T5: Discover service-down → 500 ─────────────────────────────────────────

func TestPM4PyRustHandler_Discover_ServiceDown(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
	}))
	defer mock.Close()

	router := rustSetupRouter(mock.URL)

	body := `{"event_log":[{"case_id":"c1","activity":"A","timestamp":"2024-01-01T10:00:00Z"}]}`
	req := httptest.NewRequest("POST", "/api/pm4py/discover", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// ─── T6: Conformance success ──────────────────────────────────────────────────

func TestPM4PyRustHandler_Conformance_Success(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pm4py_rust.ConformanceResponse{
			IsConformant: true,
			Fitness:      0.95,
			Precision:    0.88,
			Method:       "token-replay",
		})
	}))
	defer mock.Close()

	router := rustSetupRouter(mock.URL)

	body := `{"event_log":[{"case_id":"c1","activity":"A","timestamp":"2024-01-01T10:00:00Z"}],"petri_net":{"places":[],"transitions":[],"arcs":[]}}`
	req := httptest.NewRequest("POST", "/api/pm4py/conformance", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp pm4py_rust.ConformanceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.IsConformant)
	assert.True(t, resp.Fitness >= 0.0 && resp.Fitness <= 1.0,
		"fitness must be in [0,1], got %f", resp.Fitness)
	assert.True(t, resp.Precision >= 0.0 && resp.Precision <= 1.0,
		"precision must be in [0,1], got %f", resp.Precision)
}

// ─── T7: Statistics success ───────────────────────────────────────────────────

func TestPM4PyRustHandler_Statistics_Success(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pm4py_rust.StatisticsResponse{
			TraceCount:           50,
			EventCount:           200,
			UniqueActivities:     8,
			VariantCount:         12,
			BottleneckActivities: []string{"ProcessApplication", "ReviewDecision"},
		})
	}))
	defer mock.Close()

	router := rustSetupRouter(mock.URL)

	body := `{"event_log":[{"case_id":"c1","activity":"A","timestamp":"2024-01-01T10:00:00Z"}],"include_bottlenecks":true}`
	req := httptest.NewRequest("POST", "/api/pm4py/statistics", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp pm4py_rust.StatisticsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 50, resp.TraceCount)
	assert.Equal(t, 12, resp.VariantCount)
	assert.NotEmpty(t, resp.BottleneckActivities)
}

// ─── T8: Statistics service-down → 500 with error.code + error.message ───────

func TestPM4PyRustHandler_Statistics_ServiceDown(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
	}))
	defer mock.Close()

	router := rustSetupRouter(mock.URL)

	body := `{"event_log":[{"case_id":"c1","activity":"A","timestamp":"2024-01-01T10:00:00Z"}]}`
	req := httptest.NewRequest("POST", "/api/pm4py/statistics", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp),
		"response body must be valid JSON: %s", w.Body.String())

	errDetail, ok := errResp["error"].(map[string]interface{})
	require.True(t, ok, "response must have 'error' object, got: %s", w.Body.String())
	assert.NotEmpty(t, errDetail["code"], "error.code must be present")
	assert.NotEmpty(t, errDetail["message"], "error.message must be present")
}
