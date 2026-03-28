package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestBOSGateway_Status_ReturnsOK verifies GET /api/bos/status returns 200.
func TestBOSGateway_Status_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBOSGatewayHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	router := gin.New()
	api := router.Group("/api")
	bosGroup := api.Group("/bos")
	bosGroup.GET("/status", h.GetStatus)

	req := httptest.NewRequest(http.MethodGet, "/api/bos/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Contains(t, body, "status")
}

// TestBOSGateway_Discover_MissingLogPath_Returns400 verifies binding validation fires.
func TestBOSGateway_Discover_MissingLogPath_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBOSGatewayHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	router := gin.New()
	api := router.Group("/api")
	bosGroup := api.Group("/bos")
	bosGroup.POST("/discover", h.Discover)

	req := httptest.NewRequest(http.MethodPost, "/api/bos/discover", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBOSGateway_Conformance_MissingFields_Returns400 verifies both required fields are validated.
func TestBOSGateway_Conformance_MissingFields_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBOSGatewayHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	router := gin.New()
	api := router.Group("/api")
	bosGroup := api.Group("/bos")
	bosGroup.POST("/conformance", h.CheckConformance)

	req := httptest.NewRequest(http.MethodPost, "/api/bos/conformance",
		bytes.NewBufferString(`{"log_path":"/tmp/log.csv"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBOSGateway_Statistics_MissingLogPath_Returns400 verifies binding validation.
func TestBOSGateway_Statistics_MissingLogPath_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBOSGatewayHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	router := gin.New()
	api := router.Group("/api")
	bosGroup := api.Group("/bos")
	bosGroup.POST("/statistics", h.GetStatistics)

	req := httptest.NewRequest(http.MethodPost, "/api/bos/statistics", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestLinkedIn_ImportCSV_MissingContent_Returns400 verifies binding requires csv_content.
func TestLinkedIn_ImportCSV_MissingContent_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkedInHandler(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil, nil, nil, nil,
	)
	router := gin.New()
	api := router.Group("/api")
	liGroup := api.Group("/linkedin")
	liGroup.POST("/import", h.ImportCSV)

	req := httptest.NewRequest(http.MethodPost, "/api/linkedin/import", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestLinkedIn_EnrollOutreach_MissingSequenceID_Returns400 verifies binding requires sequence_id.
func TestLinkedIn_EnrollOutreach_MissingSequenceID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkedInHandler(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil, nil, nil, nil,
	)
	router := gin.New()
	api := router.Group("/api")
	liGroup := api.Group("/linkedin")
	liGroup.POST("/outreach/enroll", h.EnrollOutreach)

	body := map[string]interface{}{"min_score": 0.7}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/linkedin/outreach/enroll", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
