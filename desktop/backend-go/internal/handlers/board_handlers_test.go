package handlers

// Chicago TDD tests for BoardHandler.
// Each test describes a single behavior claim.
// WvdA: tests use short timeouts so they are fast and bounded.
// Armstrong: mock HTTP servers let it crash; no silent recovery.

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

// newTestBoardHandler returns a BoardHandler wired to mock servers.
// oxigraphURL and osaURL override the defaults.
func newTestBoardHandler(oxigraphURL, osaURL string) *BoardHandler {
	h := NewBoardHandler()
	h.oxigraphURL = oxigraphURL
	h.osaBaseURL = osaURL
	return h
}

// sparqlSelectJSON builds a minimal SPARQL SELECT JSON response with one binding row.
func sparqlSelectJSON(vars map[string]string) string {
	bindings := make([]string, 0, len(vars))
	for k, v := range vars {
		bindings = append(bindings, `"`+k+`":{"type":"literal","value":"`+v+`"}`)
	}
	return `{"results":{"bindings":[{` + strings.Join(bindings, ",") + `}]}}`
}

// emptySelectJSON returns a SPARQL SELECT response with no bindings.
func emptySelectJSON() string {
	return `{"results":{"bindings":[]}}`
}

// ─── TestGetBoardIntelligence_ReturnsJSON ────────────────────────────────────

// TestGetBoardIntelligence_ReturnsJSON verifies that a valid Oxigraph response
// is translated into a 200 JSON response with a "data" key.
func TestGetBoardIntelligence_ReturnsJSON(t *testing.T) {
	// Arrange: mock Oxigraph returning one BoardIntelligence binding.
	mockOxigraph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/sparql-results+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sparqlSelectJSON(map[string]string{
			"orgHealthSummary":       "0.82",
			"velocityTrend":          "improving",
			"complianceStatus":       "compliant",
			"issuesAutoResolved":     "12",
			"issuesPendingEscalation": "2",
			"derivationLevel":        "L3",
		})))
	}))
	defer mockOxigraph.Close()

	h := newTestBoardHandler(mockOxigraph.URL, "http://localhost:9999")

	// Act: call the handler via Gin test context.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/board/intelligence", nil)
	h.GetIntelligence(c)

	// Assert: HTTP 200 + JSON body with "data" key.
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "data", "response must contain 'data' key")

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "'data' must be an object")
	assert.Equal(t, "0.82", data["organizational_health_summary"])
	assert.Equal(t, "improving", data["process_velocity_trend"])
	assert.Equal(t, "L3", data["derivation_level"])
}

// ─── TestGetBoardEscalations_ReturnsEmptyWhenNoViolations ───────────────────

// TestGetBoardEscalations_ReturnsEmptyWhenNoViolations verifies that when
// Oxigraph returns no escalation rows the handler returns 200 + empty array.
func TestGetBoardEscalations_ReturnsEmptyWhenNoViolations(t *testing.T) {
	// Arrange: Oxigraph returns empty bindings.
	mockOxigraph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/sparql-results+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptySelectJSON()))
	}))
	defer mockOxigraph.Close()

	h := newTestBoardHandler(mockOxigraph.URL, "http://localhost:9999")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/board/escalations", nil)
	h.GetEscalations(c)

	// Assert: 200 with escalations=[] and count=0.
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	escalations, ok := resp["escalations"].([]interface{})
	require.True(t, ok, "'escalations' must be an array")
	assert.Empty(t, escalations, "escalations should be empty when no violations")

	count, _ := resp["count"].(float64)
	assert.Equal(t, float64(0), count)
}

// ─── TestGetBriefing_Returns503WhenOxigraphDown ──────────────────────────────

// TestGetBriefing_Returns503WhenOxigraphDown verifies that when Oxigraph is
// unreachable the handler returns a non-200 status with an error JSON body.
func TestGetBriefing_Returns503WhenOxigraphDown(t *testing.T) {
	// Arrange: point handler at an address with no server.
	// Using a closed server ensures connection refused.
	closedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedServer.Close() // Close immediately so all connections fail.

	h := newTestBoardHandler(closedServer.URL, "http://localhost:9999")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/board/briefing", nil)
	h.GetBriefing(c)

	// Assert: non-200 with error key in body.
	assert.NotEqual(t, http.StatusOK, w.Code,
		"should return non-200 when Oxigraph is down")

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "error", "response must contain 'error' key on failure")
}

// ─── TestTriggerConwayCheck_DelegatesToOSA ───────────────────────────────────

// TestTriggerConwayCheck_DelegatesToOSA verifies that the handler calls the
// OSA Conway check endpoint and relays the response body.
func TestTriggerConwayCheck_DelegatesToOSA(t *testing.T) {
	// Track whether OSA was called.
	osaCalled := false

	mockOSA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == osaConwayPath && r.Method == http.MethodPost {
			osaCalled = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","conway_violation":false}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockOSA.Close()

	// Oxigraph not needed for this test.
	h := newTestBoardHandler("http://localhost:9999", mockOSA.URL)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/board/conway/trigger", nil)
	h.TriggerConwayCheck(c)

	// Assert: OSA was called and response is relayed.
	assert.True(t, osaCalled, "handler must call the OSA Conway check endpoint")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
}

// ─── TestBoardHandlerDefaults ────────────────────────────────────────────────

// TestBoardHandlerDefaults verifies that NewBoardHandler applies correct defaults.
func TestBoardHandlerDefaults(t *testing.T) {
	// Unset env vars to confirm defaults kick in.
	t.Setenv("OXIGRAPH_URL", "")
	t.Setenv("OSA_BASE_URL", "")

	h := NewBoardHandler()
	assert.Equal(t, defaultOxigraphURL, h.oxigraphURL)
	assert.Equal(t, defaultOSABaseURL, h.osaBaseURL)
	assert.NotNil(t, h.httpClient)
	assert.Equal(t, boardQueryTimeout, h.httpClient.Timeout)
}
