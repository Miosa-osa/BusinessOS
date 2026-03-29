package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newOCPMTestRouter builds a Gin engine with the OCPM routes registered,
// pointing at the given mock server URLs.
func newOCPMTestRouter(pm4pyURL, osaURL string) *gin.Engine {
	r := gin.New()
	h := NewOCPMHandler(pm4pyURL, osaURL)
	api := r.Group("/api")
	h.RegisterRoutes(api)
	return r
}

// TestOCPMThroughput_ProxiesRequest verifies that POST /api/ocpm/throughput
// forwards the body to pm4py-rust and returns its response verbatim.
func TestOCPMThroughput_ProxiesRequest(t *testing.T) {
	// Arrange — mock pm4py-rust server.
	pm4py := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/ocpm/performance/throughput", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"throughput":42.5,"object_types":["order","item"]}`))
	}))
	defer pm4py.Close()

	router := newOCPMTestRouter(pm4py.URL, "http://osa-not-used")

	body := strings.NewReader(`{"events":[{"id":"e1","activity":"place_order"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ocpm/throughput", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act.
	router.ServeHTTP(w, req)

	// Assert.
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 42.5, resp["throughput"])
}

// TestOCPMBottleneck_ProxiesRequest verifies POST /api/ocpm/bottleneck forwarding.
func TestOCPMBottleneck_ProxiesRequest(t *testing.T) {
	pm4py := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/ocpm/performance/bottleneck", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"bottlenecks":[{"activity":"approve","avg_duration_ms":1500}]}`))
	}))
	defer pm4py.Close()

	router := newOCPMTestRouter(pm4py.URL, "http://osa-not-used")

	body := strings.NewReader(`{"events":[],"top_n":5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ocpm/bottleneck", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	bottlenecks, ok := resp["bottlenecks"].([]any)
	require.True(t, ok, "expected bottlenecks array")
	assert.Len(t, bottlenecks, 1)
}

// TestOCPMQuery_ProxiesRequest verifies POST /api/ocpm/query forwarding.
func TestOCPMQuery_ProxiesRequest(t *testing.T) {
	pm4py := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/ocpm/llm/query", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"answer":"The bottleneck is approve.","grounded":true}`))
	}))
	defer pm4py.Close()

	router := newOCPMTestRouter(pm4py.URL, "http://osa-not-used")

	body := strings.NewReader(`{"question":"What is the bottleneck?","ocel":{},"api_key":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ocpm/query", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "The bottleneck is approve.", resp["answer"])
	assert.Equal(t, true, resp["grounded"])
}

// TestOCPMExportOCEL_ProxiesRequest verifies GET /api/ocpm/export forwards to OSA.
func TestOCPMExportOCEL_ProxiesRequest(t *testing.T) {
	// Arrange — mock OSA server.
	osa := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/ocel/export", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ocelVersion":"2.0","events":[],"objects":[]}`))
	}))
	defer osa.Close()

	router := newOCPMTestRouter("http://pm4py-not-used", osa.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/ocpm/export", nil)
	w := httptest.NewRecorder()

	// Act.
	router.ServeHTTP(w, req)

	// Assert.
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "2.0", resp["ocelVersion"])
}

// TestOCPMHandler_RegisterRoutes verifies all 4 routes are registered.
func TestOCPMHandler_RegisterRoutes(t *testing.T) {
	h := NewOCPMHandler("http://localhost:8090", "http://localhost:8089")
	r := gin.New()
	api := r.Group("/api")
	h.RegisterRoutes(api)

	routes := r.Routes()
	routeMap := make(map[string]bool, len(routes))
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = true
	}

	assert.True(t, routeMap["POST:/api/ocpm/throughput"], "POST /api/ocpm/throughput must be registered")
	assert.True(t, routeMap["POST:/api/ocpm/bottleneck"], "POST /api/ocpm/bottleneck must be registered")
	assert.True(t, routeMap["POST:/api/ocpm/query"], "POST /api/ocpm/query must be registered")
	assert.True(t, routeMap["GET:/api/ocpm/export"], "GET /api/ocpm/export must be registered")
}

// TestOCPMThroughput_Pm4pyUnreachable verifies 502 when pm4py-rust is down.
func TestOCPMThroughput_Pm4pyUnreachable(t *testing.T) {
	// Point at a port that refuses connections.
	router := newOCPMTestRouter("http://127.0.0.1:19999", "http://osa-not-used")

	body := strings.NewReader(`{"events":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ocpm/throughput", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "pm4py-rust unreachable")
}

// TestOCPMExportOCEL_OSAUnreachable verifies 502 when OSA is down.
func TestOCPMExportOCEL_OSAUnreachable(t *testing.T) {
	router := newOCPMTestRouter("http://pm4py-not-used", "http://127.0.0.1:19998")

	req := httptest.NewRequest(http.MethodGet, "/api/ocpm/export", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "OSA unreachable")
}

// TestNewOCPMHandler_DefaultURLs verifies env-var fallback defaults.
func TestNewOCPMHandler_DefaultURLs(t *testing.T) {
	// Unset both env vars to trigger the hardcoded defaults.
	t.Setenv("PM4PY_RUST_URL", "")
	t.Setenv("OSA_URL", "")

	h := NewOCPMHandler("", "")
	assert.Equal(t, "http://localhost:8090", h.pm4pyURL)
	assert.Equal(t, "http://localhost:8089", h.osaURL)
}

// TestNewOCPMHandler_EnvURLs verifies env vars are picked up when no explicit URL given.
func TestNewOCPMHandler_EnvURLs(t *testing.T) {
	t.Setenv("PM4PY_RUST_URL", "http://pm4py-custom:9000")
	t.Setenv("OSA_URL", "http://osa-custom:9001")

	h := NewOCPMHandler("", "")
	assert.Equal(t, "http://pm4py-custom:9000", h.pm4pyURL)
	assert.Equal(t, "http://osa-custom:9001", h.osaURL)
}
