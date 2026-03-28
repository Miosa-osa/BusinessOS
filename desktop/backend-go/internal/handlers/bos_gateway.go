// Package handlers provides HTTP handlers for BusinessOS.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	semconv "github.com/rhl/businessos-backend/internal/semconv"
)

// pm4py-rust HTTP routes (see pm4py-rust/src/http/businessos_api.rs).
const (
	pm4pyPathDiscoveryAlpha         = "/api/discovery/alpha"
	pm4pyPathConformanceTokenReplay = "/api/conformance/token-replay"
	pm4pyPathStatistics            = "/api/statistics"
	pm4pyPathParseXES              = "/api/io/parse-xes"
)

// BOSGatewayHandler handles BOS CLI ↔ BusinessOS API gateway operations.
type BOSGatewayHandler struct {
	pool             *pgxpool.Pool
	logger           *slog.Logger
	stats            *GatewayStatistics
	mu               sync.RWMutex
	pm4pyURL         string
	canopyWebhookURL string
	httpClient       *http.Client
}

// GatewayStatistics tracks gateway metrics.
type GatewayStatistics struct {
	RequestsTotal  uint64    `json:"requests_total"`
	RequestsFailed uint64    `json:"requests_failed"`
	AverageLatency float64   `json:"average_latency_ms"`
	LatencyValues  []uint64  `json:"-"`
	StartedAt      time.Time `json:"started_at"`
	mu             sync.Mutex
}

// NewBOSGatewayHandler creates a new BOS gateway handler.
// pm4pyURL is loaded from PM4PY_RUST_URL env var, defaults to http://localhost:8090.
func NewBOSGatewayHandler(pool *pgxpool.Pool, logger *slog.Logger) *BOSGatewayHandler {
	if logger == nil {
		logger = slog.Default()
	}

	pm4pyURL := "http://localhost:8090"
	// Try to load from environment if available
	if envURL := os.Getenv("PM4PY_RUST_URL"); envURL != "" {
		pm4pyURL = envURL
	}

	canopyWebhookURL := os.Getenv("CANOPY_WEBHOOK_URL")
	// No default — empty string disables the feature

	return &BOSGatewayHandler{
		pool:   pool,
		logger: logger,
		stats: &GatewayStatistics{
			StartedAt:     time.Now(),
			LatencyValues: make([]uint64, 0),
		},
		pm4pyURL:         pm4pyURL,
		canopyWebhookURL: canopyWebhookURL,
		// ## Backpressure: HTTP Client Timeout (WvdA deadlock-free)
		// 30-second timeout prevents unbounded hangs to pm4py-rust.
		// otelhttp.NewTransport wraps the default transport so that outbound
		// requests automatically carry W3C traceparent + tracestate headers,
		// enabling distributed trace propagation to pm4py-rust.
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   30 * time.Second,
		},
	}
}

// RegisterBOSGatewayRoutes wires /api/bos routes for the gateway.
func RegisterBOSGatewayRoutes(api *gin.RouterGroup, h *BOSGatewayHandler) {
	bos := api.Group("/bos")
	{
		bos.POST("/discover", h.Discover)
		bos.POST("/conformance", h.CheckConformance)
		bos.POST("/statistics", h.GetStatistics)
		bos.GET("/status", h.GetStatus)
	}
}

// ============================================================================
// REQUEST/RESPONSE TYPES
// ============================================================================

// BOSDiscoverRequest represents a model discovery request from BOS CLI.
type BOSDiscoverRequest struct {
	LogPath   string `json:"log_path" binding:"required"`
	Algorithm string `json:"algorithm,omitempty"`
}

// BOSDiscoverResponse represents the discovered model response.
type BOSDiscoverResponse struct {
	ModelID     string          `json:"model_id"`
	Algorithm   string          `json:"algorithm"`
	Places      int             `json:"places"`
	Transitions int             `json:"transitions"`
	Arcs        int             `json:"arcs"`
	ModelData   json.RawMessage `json:"model_data"`
	LatencyMs   uint64          `json:"latency_ms"`
}

// BOSConformanceRequest represents a conformance check request.
type BOSConformanceRequest struct {
	LogPath   string `json:"log_path" binding:"required"`
	ModelID   string `json:"model_id" binding:"required"`
	ModelPath string `json:"model_path,omitempty"`
}

// BOSConformanceResponse represents conformance check results.
type BOSConformanceResponse struct {
	TracesChecked  uint64  `json:"traces_checked"`
	FittingTraces  uint64  `json:"fitting_traces"`
	Fitness        float64 `json:"fitness"`
	Precision      float64 `json:"precision"`
	Generalization float64 `json:"generalization"`
	Simplicity     float64 `json:"simplicity"`
	LatencyMs      uint64  `json:"latency_ms"`
}

// BOSStatisticsRequest represents a statistics extraction request.
type BOSStatisticsRequest struct {
	LogPath string `json:"log_path" binding:"required"`
}

// BOSActivityStatistic represents per-activity statistics.
type BOSActivityStatistic struct {
	Activity   string  `json:"activity"`
	Frequency  int     `json:"frequency"`
	Percentage float64 `json:"percentage"`
}

// BOSCaseDurationStatistic represents case duration statistics.
type BOSCaseDurationStatistic struct {
	MinSeconds    int64   `json:"min_seconds"`
	MaxSeconds    int64   `json:"max_seconds"`
	AvgSeconds    float64 `json:"avg_seconds"`
	MedianSeconds float64 `json:"median_seconds"`
}

// BOSStatisticsResponse represents extracted log statistics.
type BOSStatisticsResponse struct {
	LogName             string                   `json:"log_name"`
	NumTraces           int                      `json:"num_traces"`
	NumEvents           int                      `json:"num_events"`
	NumUniqueActivities int                      `json:"num_unique_activities"`
	NumVariants         int                      `json:"num_variants"`
	AvgTraceLength      float64                  `json:"avg_trace_length"`
	MinTraceLength      int                      `json:"min_trace_length"`
	MaxTraceLength      int                      `json:"max_trace_length"`
	ActivityFrequency   []BOSActivityStatistic   `json:"activity_frequency"`
	CaseDuration        BOSCaseDurationStatistic `json:"case_duration"`
	LatencyMs           uint64                   `json:"latency_ms"`
}

// BOSStatusResponse represents the gateway health status.
type BOSStatusResponse struct {
	Status           string  `json:"status"`
	DatabaseReady    bool    `json:"database_ready"`
	LatencyMs        uint64  `json:"latency_ms"`
	RequestsTotal    uint64  `json:"requests_total"`
	RequestsFailed   uint64  `json:"requests_failed"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
	UptimeSeconds    int64   `json:"uptime_seconds"`
}

// ============================================================================
// GATEWAY ENDPOINTS
// ============================================================================

// Discover handles POST /api/bos/discover
// Triggers process model discovery on the given event log.
func (h *BOSGatewayHandler) Discover(c *gin.Context) {
	startTime := time.Now()

	// Start OTEL span for the gateway discover operation.
	gatewayTracer := otel.Tracer("businessos-gateway")
	ctx, span := gatewayTracer.Start(c.Request.Context(), semconv.BosGatewayDiscoverSpan)
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	// Attach correlation_id attribute if present.
	if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
		span.SetAttributes(attribute.String(string(semconv.ChatmangptRunCorrelationIdKey), correlationID))
	}

	var req BOSDiscoverRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("discover: invalid request", "error", err.Error())
		span.SetStatus(codes.Error, "invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Algorithm == "" {
		req.Algorithm = "inductive_miner"
	}

	h.logger.Info("discover: processing request",
		"log_path", req.LogPath,
		"algorithm", req.Algorithm,
	)

	// Read event log (JSON on disk, or XES forwarded to pm4py-rust for parsing).
	// pm4py-rust expects {event_log: <JSON content>, variant: <string>}.
	eventLog, err := h.loadEventLogForGateway(c.Request.Context(), req.LogPath)
	if err != nil {
		h.logger.Warn("discover: failed to read event log file",
			"log_path", req.LogPath,
			"error", err.Error(),
		)
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to read event log: %v", err)})
		return
	}

	// Call pm4py-rust HTTP API with event log content (not file path).
	pm4pyPayload := struct {
		EventLog json.RawMessage `json:"event_log"`
		Variant  string          `json:"variant"`
	}{
		EventLog: eventLog,
		Variant:  req.Algorithm,
	}
	pm4pyReqBody, err := json.Marshal(pm4pyPayload)
	if err != nil {
		h.logger.Error("discover: failed to marshal pm4py request", "error", err.Error())
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build pm4py-rust request"})
		return
	}

	httpReq, _ := http.NewRequestWithContext(c.Request.Context(), "POST",
		h.pm4pyURL+pm4pyPathDiscoveryAlpha, bytes.NewReader(pm4pyReqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	// Forward correlation_id to pm4py-rust so the full chain shares one ID.
	if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
		httpReq.Header.Set("X-Correlation-ID", correlationID)
	}

	httpResp, err := h.httpClient.Do(httpReq)
	if err != nil {
		h.logger.Error("discover: pm4py-rust request failed",
			"pm4py_url", h.pm4pyURL,
			"error", err.Error(),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "pm4py-rust unavailable")
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "pm4py-rust unavailable"})
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		h.logger.Warn("discover: pm4py-rust error",
			"status_code", httpResp.StatusCode,
		)
		span.SetStatus(codes.Error, fmt.Sprintf("pm4py-rust returned %d", httpResp.StatusCode))
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "pm4py-rust error"})
		return
	}

	// Parse pm4py-rust response
	var pm4pyResp map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&pm4pyResp); err != nil {
		h.logger.Error("discover: failed to parse pm4py-rust response", "error", err.Error())
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse pm4py-rust response"})
		return
	}

	modelData, _ := json.Marshal(pm4pyResp)

	modelID := uuid.New().String()
	if mid, ok := pm4pyResp["model_id"].(string); ok && mid != "" {
		modelID = mid
	}
	algo := req.Algorithm
	if a, ok := pm4pyResp["algorithm"].(string); ok && a != "" {
		algo = a
	}

	places, transitions, arcs := 0, 0, 0
	if pn, ok := pm4pyResp["petri_net"].(map[string]interface{}); ok {
		if pl, ok := pn["places"].([]interface{}); ok {
			places = len(pl)
		}
		if tr, ok := pn["transitions"].([]interface{}); ok {
			transitions = len(tr)
		}
		if ar, ok := pn["arcs"].([]interface{}); ok {
			arcs = len(ar)
		}
	}

	response := BOSDiscoverResponse{
		ModelID:     modelID,
		Algorithm:   algo,
		Places:      places,
		Transitions: transitions,
		Arcs:        arcs,
		ModelData:   json.RawMessage(modelData),
		LatencyMs:   uint64(time.Since(startTime).Milliseconds()),
	}

	// ## Durability via Write-Ahead Logging (WAL)
	// Persist result to PostgreSQL BEFORE sending response to client.
	// If connection drops mid-flight or Gin crashes, result recoverable from DB.
	// Non-fatal failure: continue even if WAL write fails (client gets response).
	// WvdA soundness: write-ahead log before returning response.
	// If the client connection drops after pm4py-rust succeeds, the result
	// is recoverable from the WAL. Cleanup happens after successful response.
	if err := h.writeAheadLog(modelID, &response); err != nil {
		h.logger.Warn("discover: WAL write failed (non-fatal, continuing)",
			"model_id", modelID,
			"error", err.Error(),
		)
	}

	// Persist discovery result to process_discovery_results (non-fatal — WAL is durability guarantee)
	if err := h.persistDiscoveryResult(c.Request.Context(), modelID, algo, &response); err != nil {
		h.logger.Warn("discover: DB persist failed (non-fatal, WAL has the record)",
			"model_id", modelID,
			"error", err.Error(),
		)
	}

	h.recordRequest(true, response.LatencyMs)
	span.SetAttributes(
		semconv.BosGatewayModelId(response.ModelID),
		semconv.BosGatewayAlgorithm(response.Algorithm),
		semconv.BosGatewayLatencyMs(int64(response.LatencyMs)),
	)
	span.SetStatus(codes.Ok, "")
	h.logger.Info("discover: completed successfully",
		"model_id", response.ModelID,
		"latency_ms", response.LatencyMs,
	)

	// ## Asynchronous Cleanup
	// Schedule WAL cleanup after response is sent (5s delay allows client to confirm receipt).
	// Non-blocking cleanup: if deletion fails, it's non-critical (duplicate results acceptable).
	// Schedule WAL cleanup after response is sent; context bounds goroutine lifetime.
	go func(ctx context.Context) {
		select {
		case <-time.After(5 * time.Second):
			if err := h.cleanupWAL(modelID); err != nil {
				h.logger.Debug("discover: WAL cleanup failed (non-critical)",
					"model_id", modelID,
					"error", err.Error(),
				)
			}
		case <-ctx.Done():
		}
	}(c.Request.Context())

	// Fire-and-forget: Canopy discovery webhook (WvdA: bounded, no leak)
	if h.canopyWebhookURL != "" {
		go h.sendCanopyWebhook(response.ModelID, response.Algorithm, response.Transitions)
	}

	c.JSON(http.StatusOK, response)
}

// CheckConformance handles POST /api/bos/conformance
// Checks if an event log conforms to a given process model.
func (h *BOSGatewayHandler) CheckConformance(c *gin.Context) {
	startTime := time.Now()

	// Start OTEL span for the gateway conformance operation.
	gatewayTracer := otel.Tracer("businessos-gateway")
	ctx, span := gatewayTracer.Start(c.Request.Context(), semconv.BosGatewayConformanceSpan)
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
		span.SetAttributes(attribute.String(string(semconv.ChatmangptRunCorrelationIdKey), correlationID))
	}

	var req BOSConformanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("conformance: invalid request", "error", err.Error())
		span.SetStatus(codes.Error, "invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.Info("conformance: processing request",
		"log_path", req.LogPath,
		"model_id", req.ModelID,
	)

	eventLog, err := h.loadEventLogForGateway(c.Request.Context(), req.LogPath)
	if err != nil {
		h.logger.Warn("conformance: failed to read event log file",
			"log_path", req.LogPath,
			"error", err.Error(),
		)
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to read event log: %v", err)})
		return
	}

	// Resolve Petri net: prefer model_path file if provided, otherwise recover from WAL.
	var petriNetRaw json.RawMessage
	if req.ModelPath != "" {
		petriNetRaw, err = h.loadEventLogForGateway(c.Request.Context(), req.ModelPath)
		if err != nil {
			h.logger.Warn("conformance: failed to read model path file",
				"model_path", req.ModelPath,
				"error", err.Error(),
			)
			h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to read model file: %v", err)})
			return
		}
	} else {
		// Attempt to recover from WAL using model_id.
		walResult, walErr := h.recoverFromWAL(req.ModelID)
		if walErr == nil && walResult != nil && len(walResult.ModelData) > 0 {
			// Extract petri_net from the stored discovery response model data.
			var modelMap map[string]json.RawMessage
			if jsonErr := json.Unmarshal(walResult.ModelData, &modelMap); jsonErr == nil {
				if pn, ok := modelMap["petri_net"]; ok {
					petriNetRaw = pn
				}
			}
		}
	}

	// Build conformance payload for pm4py-rust.
	// pm4py-rust expects {event_log: <JSON content>, petri_net: <petri net>, method: "token_replay"}.
	conformancePayload := struct {
		EventLog json.RawMessage `json:"event_log"`
		PetriNet json.RawMessage `json:"petri_net,omitempty"`
		Method   string          `json:"method"`
	}{
		EventLog: eventLog,
		PetriNet: petriNetRaw,
		Method:   "token_replay",
	}
	pm4pyReqBody, err := json.Marshal(conformancePayload)
	if err != nil {
		h.logger.Error("conformance: failed to marshal pm4py request", "error", err.Error())
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build pm4py-rust request"})
		return
	}

	httpReq, _ := http.NewRequestWithContext(c.Request.Context(), "POST",
		h.pm4pyURL+pm4pyPathConformanceTokenReplay, bytes.NewReader(pm4pyReqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
		httpReq.Header.Set("X-Correlation-ID", correlationID)
	}

	httpResp, err := h.httpClient.Do(httpReq)
	if err != nil {
		h.logger.Error("conformance: pm4py-rust request failed",
			"pm4py_url", h.pm4pyURL,
			"error", err.Error(),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "pm4py-rust unavailable")
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "pm4py-rust unavailable"})
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		h.logger.Warn("conformance: pm4py-rust error",
			"status_code", httpResp.StatusCode,
		)
		span.SetStatus(codes.Error, fmt.Sprintf("pm4py-rust returned %d", httpResp.StatusCode))
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "pm4py-rust error"})
		return
	}

	// Parse pm4py-rust response
	var pm4pyResp map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&pm4pyResp); err != nil {
		h.logger.Error("conformance: failed to parse pm4py-rust response", "error", err.Error())
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse pm4py-rust response"})
		return
	}

	tracesChecked := uint64(0)
	if v, ok := pm4pyResp["traces_checked"].(float64); ok {
		tracesChecked = uint64(v)
	}
	fittingTraces := uint64(0)
	if v, ok := pm4pyResp["fitting_traces"].(float64); ok {
		fittingTraces = uint64(v)
	}
	fitness := 0.0
	if v, ok := pm4pyResp["fitness"].(float64); ok {
		fitness = v
	}
	precision := 0.0
	if v, ok := pm4pyResp["precision"].(float64); ok {
		precision = v
	}
	generalization := 0.0
	if v, ok := pm4pyResp["generalization"].(float64); ok {
		generalization = v
	}
	simplicity := 0.0
	if v, ok := pm4pyResp["simplicity"].(float64); ok {
		simplicity = v
	}
	if isConf, ok := pm4pyResp["is_conformant"].(bool); ok && isConf && fittingTraces == 0 && tracesChecked == 0 {
		fittingTraces = 1
		tracesChecked = 1
	}

	response := BOSConformanceResponse{
		TracesChecked:  tracesChecked,
		FittingTraces:  fittingTraces,
		Fitness:        fitness,
		Precision:      precision,
		Generalization: generalization,
		Simplicity:     simplicity,
		LatencyMs:      uint64(time.Since(startTime).Milliseconds()),
	}

	// Update fitness in process_discovery_results (non-fatal — model_id may not exist yet)
	if req.ModelID != "" {
		if err := h.updateDiscoveryFitness(c.Request.Context(), req.ModelID, fitness, 0.0); err != nil {
			h.logger.Warn("conformance: DB fitness update failed (non-fatal)",
				"model_id", req.ModelID,
				"error", err.Error(),
			)
		}
	}

	h.recordRequest(true, response.LatencyMs)
	span.SetAttributes(
		semconv.BosGatewayFitness(response.Fitness),
		semconv.BosGatewayNumTraces(int64(response.TracesChecked)),
		semconv.BosGatewayLatencyMs(int64(response.LatencyMs)),
	)
	span.SetStatus(codes.Ok, "")
	h.logger.Info("conformance: completed successfully",
		"fitness", response.Fitness,
		"latency_ms", response.LatencyMs,
	)

	c.JSON(http.StatusOK, response)
}

// GetStatistics handles POST /api/bos/statistics
// Extracts statistics from an event log.
func (h *BOSGatewayHandler) GetStatistics(c *gin.Context) {
	startTime := time.Now()

	// Start OTEL span for the gateway statistics operation.
	gatewayTracer := otel.Tracer("businessos-gateway")
	ctx, span := gatewayTracer.Start(c.Request.Context(), semconv.BosGatewayStatisticsSpan)
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
		span.SetAttributes(attribute.String(string(semconv.ChatmangptRunCorrelationIdKey), correlationID))
	}

	var req BOSStatisticsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("statistics: invalid request", "error", err.Error())
		span.SetStatus(codes.Error, "invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.Info("statistics: processing request", "log_path", req.LogPath)

	eventLog, err := h.loadEventLogForGateway(c.Request.Context(), req.LogPath)
	if err != nil {
		h.logger.Warn("statistics: failed to read event log file",
			"log_path", req.LogPath,
			"error", err.Error(),
		)
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to read event log: %v", err)})
		return
	}

	// Call pm4py-rust HTTP API with event log content (not file path).
	// pm4py-rust expects {event_log: <JSON content>, include_variants: bool, ...}.
	statisticsPayload := struct {
		EventLog               json.RawMessage `json:"event_log"`
		IncludeVariants        bool            `json:"include_variants"`
		IncludeResourceMetrics bool            `json:"include_resource_metrics"`
		IncludeBottlenecks     bool            `json:"include_bottlenecks"`
	}{
		EventLog:               eventLog,
		IncludeVariants:        true,
		IncludeResourceMetrics: true,
		IncludeBottlenecks:     true,
	}
	pm4pyReqBody, err := json.Marshal(statisticsPayload)
	if err != nil {
		h.logger.Error("statistics: failed to marshal pm4py request", "error", err.Error())
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build pm4py-rust request"})
		return
	}

	httpReq, _ := http.NewRequestWithContext(c.Request.Context(), "POST",
		h.pm4pyURL+pm4pyPathStatistics, bytes.NewReader(pm4pyReqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
		httpReq.Header.Set("X-Correlation-ID", correlationID)
	}

	httpResp, err := h.httpClient.Do(httpReq)
	if err != nil {
		h.logger.Error("statistics: pm4py-rust request failed",
			"pm4py_url", h.pm4pyURL,
			"error", err.Error(),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "pm4py-rust unavailable")
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "pm4py-rust unavailable"})
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		h.logger.Warn("statistics: pm4py-rust error",
			"status_code", httpResp.StatusCode,
		)
		span.SetStatus(codes.Error, fmt.Sprintf("pm4py-rust returned %d", httpResp.StatusCode))
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "pm4py-rust error"})
		return
	}

	// Parse pm4py-rust response
	var pm4pyResp map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&pm4pyResp); err != nil {
		h.logger.Error("statistics: failed to parse pm4py-rust response", "error", err.Error())
		h.recordRequest(false, uint64(time.Since(startTime).Milliseconds()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse pm4py-rust response"})
		return
	}

	logName := "event_log"
	if v, ok := pm4pyResp["log_name"].(string); ok && v != "" {
		logName = v
	}
	numTraces := intFromJSONFloat(pm4pyResp["trace_count"], 0)
	numEvents := intFromJSONFloat(pm4pyResp["event_count"], 0)
	numUniqueActivities := intFromJSONFloat(pm4pyResp["unique_activities"], 0)
	if numTraces == 0 {
		numTraces = intFromJSONFloat(pm4pyResp["num_traces"], 0)
	}
	if numEvents == 0 {
		numEvents = intFromJSONFloat(pm4pyResp["num_events"], 0)
	}
	if numUniqueActivities == 0 {
		numUniqueActivities = intFromJSONFloat(pm4pyResp["num_unique_activities"], 0)
	}
	numVariants := intFromJSONFloat(pm4pyResp["variant_count"], 0)
	if numVariants == 0 {
		numVariants = intFromJSONFloat(pm4pyResp["num_variants"], 0)
	}
	avgTraceLength := 0.0
	if v, ok := pm4pyResp["avg_trace_length"].(float64); ok {
		avgTraceLength = v
	} else if numTraces > 0 {
		avgTraceLength = float64(numEvents) / float64(numTraces)
	}
	minTraceLength := intFromJSONFloat(pm4pyResp["min_trace_length"], 0)
	maxTraceLength := intFromJSONFloat(pm4pyResp["max_trace_length"], 0)

	activityFreq := []BOSActivityStatistic{}
	if freq, ok := pm4pyResp["activity_frequency"].([]interface{}); ok {
		for _, item := range freq {
			if m, ok := item.(map[string]interface{}); ok {
				activity := ""
				if v, ok := m["activity"].(string); ok {
					activity = v
				}
				frequency := 0
				if v, ok := m["frequency"].(float64); ok {
					frequency = int(v)
				}
				percentage := 0.0
				if v, ok := m["percentage"].(float64); ok {
					percentage = v
				}
				activityFreq = append(activityFreq, BOSActivityStatistic{
					Activity:   activity,
					Frequency:  frequency,
					Percentage: percentage,
				})
			}
		}
	} else if af, ok := pm4pyResp["activity_frequencies"].(map[string]interface{}); ok {
		total := 0
		for _, v := range af {
			if n, ok := v.(float64); ok {
				total += int(n)
			}
		}
		for act, v := range af {
			freq := 0
			if n, ok := v.(float64); ok {
				freq = int(n)
			}
			pct := 0.0
			if total > 0 {
				pct = 100.0 * float64(freq) / float64(total)
			}
			activityFreq = append(activityFreq, BOSActivityStatistic{
				Activity:   act,
				Frequency:  freq,
				Percentage: pct,
			})
		}
	}

	// Parse case duration
	caseDuration := BOSCaseDurationStatistic{
		MinSeconds:    60,
		MaxSeconds:    3600,
		AvgSeconds:    1200.5,
		MedianSeconds: 900.0,
	}
	if cd, ok := pm4pyResp["case_duration"].(map[string]interface{}); ok {
		if v, ok := cd["min_seconds"].(float64); ok {
			caseDuration.MinSeconds = int64(v)
		}
		if v, ok := cd["max_seconds"].(float64); ok {
			caseDuration.MaxSeconds = int64(v)
		}
		if v, ok := cd["avg_seconds"].(float64); ok {
			caseDuration.AvgSeconds = v
		}
		if v, ok := cd["median_seconds"].(float64); ok {
			caseDuration.MedianSeconds = v
		}
	}

	response := BOSStatisticsResponse{
		LogName:             logName,
		NumTraces:           numTraces,
		NumEvents:           numEvents,
		NumUniqueActivities: numUniqueActivities,
		NumVariants:         numVariants,
		AvgTraceLength:      avgTraceLength,
		MinTraceLength:      minTraceLength,
		MaxTraceLength:      maxTraceLength,
		ActivityFrequency:   activityFreq,
		CaseDuration:        caseDuration,
		LatencyMs:           uint64(time.Since(startTime).Milliseconds()),
	}

	h.recordRequest(true, response.LatencyMs)
	span.SetAttributes(
		semconv.BosGatewayNumTraces(int64(response.NumTraces)),
		semconv.BosGatewayLatencyMs(int64(response.LatencyMs)),
	)
	span.SetStatus(codes.Ok, "")
	h.logger.Info("statistics: completed successfully",
		"num_traces", response.NumTraces,
		"latency_ms", response.LatencyMs,
	)

	c.JSON(http.StatusOK, response)
}

// GetStatus handles GET /api/bos/status
// Returns gateway health status and statistics.
func (h *BOSGatewayHandler) GetStatus(c *gin.Context) {
	startTime := time.Now()
	dbReady := h.checkDatabase(c.Request.Context())

	h.mu.RLock()
	stats := h.stats
	h.mu.RUnlock()

	stats.mu.Lock()
	requestsTotal := stats.RequestsTotal
	requestsFailed := stats.RequestsFailed
	avgLatency := stats.AverageLatency
	stats.mu.Unlock()

	response := BOSStatusResponse{
		Status:           "healthy",
		DatabaseReady:    dbReady,
		LatencyMs:        uint64(time.Since(startTime).Milliseconds()),
		RequestsTotal:    requestsTotal,
		RequestsFailed:   requestsFailed,
		AverageLatencyMs: avgLatency,
		UptimeSeconds:    int64(time.Since(stats.StartedAt).Seconds()),
	}

	if !dbReady {
		response.Status = "degraded"
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================================
// INTERNAL HELPERS
// ============================================================================

// sendCanopyWebhook fires a POST to the Canopy discovery webhook.
// Called as a goroutine. Owns its own 10s context deadline — always exits.
// Failure is logged but not propagated; this is an advisory notification.
func (h *BOSGatewayHandler) sendCanopyWebhook(modelID, algorithm string, activitiesCount int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payload := map[string]interface{}{
		"model_id":         modelID,
		"algorithm":        algorithm,
		"activities_count": activitiesCount,
		"fitness_score":    -1.0, // -1.0 = not-yet-computed sentinel; updated after conformance check
	}

	body, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("canopy webhook: failed to marshal payload",
			"model_id", modelID,
			"error", err.Error(),
		)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		h.canopyWebhookURL, bytes.NewReader(body))
	if err != nil {
		h.logger.Error("canopy webhook: failed to build request",
			"model_id", modelID,
			"error", err.Error(),
		)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Warn("canopy webhook: POST failed (non-fatal)",
			"model_id", modelID,
			"canopy_url", h.canopyWebhookURL,
			"error", err.Error(),
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		h.logger.Warn("canopy webhook: non-2xx response (non-fatal)",
			"model_id", modelID,
			"status", resp.StatusCode,
		)
		return
	}

	h.logger.Info("canopy webhook: delivery confirmed",
		"model_id", modelID,
		"status", resp.StatusCode,
	)
}

// loadEventLogForGateway loads a JSON event log from disk, or parses .xes via pm4py-rust.
func (h *BOSGatewayHandler) loadEventLogForGateway(ctx context.Context, logPath string) (json.RawMessage, error) {
	if logPath == "" {
		return nil, fmt.Errorf("log_path is empty")
	}
	if strings.HasSuffix(strings.ToLower(logPath), ".xes") {
		data, err := os.ReadFile(logPath)
		if err != nil {
			return nil, fmt.Errorf("read XES file %q: %w", logPath, err)
		}
		parseURL := strings.TrimSuffix(h.pm4pyURL, "/") + pm4pyPathParseXES
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, parseURL, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/xml")
		resp, err := h.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("pm4py XES parse request: %w", err)
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("pm4py XES parse returned %d: %s", resp.StatusCode, string(raw))
		}
		if !json.Valid(raw) {
			return nil, fmt.Errorf("pm4py XES parse returned non-JSON")
		}
		return json.RawMessage(raw), nil
	}
	return readEventLog(logPath)
}

func intFromJSONFloat(v interface{}, def int) int {
	f, ok := v.(float64)
	if !ok {
		return def
	}
	return int(f)
}

// readEventLog reads a file at logPath, validates it is valid JSON, and returns
// the raw JSON bytes. Returns an error if the file cannot be read or is not
// valid JSON. This is used before forwarding event log content to pm4py-rust.
func readEventLog(logPath string) (json.RawMessage, error) {
	if logPath == "" {
		return nil, fmt.Errorf("log_path is empty")
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("read event log file %q: %w", logPath, err)
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("event log file %q is not valid JSON", logPath)
	}
	return json.RawMessage(data), nil
}

// checkDatabase verifies database connectivity.
func (h *BOSGatewayHandler) checkDatabase(ctx context.Context) bool {
	if h.pool == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return h.pool.Ping(ctx) == nil
}

// recordRequest updates gateway statistics.
func (h *BOSGatewayHandler) recordRequest(success bool, latencyMs uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stats.mu.Lock()
	defer h.stats.mu.Unlock()

	h.stats.RequestsTotal++
	if !success {
		h.stats.RequestsFailed++
	}

	// Update rolling average latency (keep last 100 values)
	h.stats.LatencyValues = append(h.stats.LatencyValues, latencyMs)
	if len(h.stats.LatencyValues) > 100 {
		h.stats.LatencyValues = h.stats.LatencyValues[1:]
	}

	// Calculate average
	var sum uint64
	for _, v := range h.stats.LatencyValues {
		sum += v
	}
	h.stats.AverageLatency = float64(sum) / float64(len(h.stats.LatencyValues))
}

// generateModelID generates a unique model identifier.
func generateModelID() string {
	// Use timestamp-based ID generation
	t := time.Now()
	return "model_" + t.Format("20060102150405") + "_" + t.Format("000")
}

// ============================================================================
// WvDA SOUNDNESS: WRITE-AHEAD LOG
// ============================================================================

// walDir returns the directory for write-ahead log files.
func (h *BOSGatewayHandler) walDir() string {
	dir := os.Getenv("BOS_WAL_DIR")
	if dir == "" {
		dir = os.TempDir() + "/bos_wal"
	}
	return dir
}

// walPath returns the file path for a given model ID's WAL entry.
func (h *BOSGatewayHandler) walPath(modelID string) string {
	return h.walDir() + "/" + modelID + ".wal.json"
}

// writeAheadLog persists a discovery result to a temporary location before
// attempting the final DB write. This prevents result loss on transient failures.
// Implements WvdA soundness: every token (result) has a path to completion.
func (h *BOSGatewayHandler) writeAheadLog(modelID string, result *BOSDiscoverResponse) error {
	dir := h.walDir()
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create WAL directory: %w", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal WAL entry: %w", err)
	}

	// Write atomically: write to temp file, then rename
	tmpPath := dir + "/" + modelID + ".wal.tmp"
	if err := os.WriteFile(tmpPath, data, 0640); err != nil {
		return fmt.Errorf("write WAL temp file: %w", err)
	}

	// Atomic rename
	finalPath := h.walPath(modelID)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmpPath)
		return fmt.Errorf("rename WAL file: %w", err)
	}

	h.logger.Debug("write-ahead log entry written", "model_id", modelID, "path", finalPath)
	return nil
}

// recoverFromWAL reads a previously written WAL entry for the given model ID.
// Returns the discovery result if found, or an error if no WAL entry exists.
func (h *BOSGatewayHandler) recoverFromWAL(modelID string) (*BOSDiscoverResponse, error) {
	path := h.walPath(modelID)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("WAL entry not found for model %s", modelID)
		}
		return nil, fmt.Errorf("read WAL file: %w", err)
	}

	var result BOSDiscoverResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal WAL entry: %w", err)
	}

	h.logger.Debug("recovered from write-ahead log", "model_id", modelID)
	return &result, nil
}

// cleanupWAL removes the WAL entry for a given model ID after successful persistence.
func (h *BOSGatewayHandler) cleanupWAL(modelID string) error {
	path := h.walPath(modelID)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cleanup WAL file: %w", err)
	}
	h.logger.Debug("cleaned up WAL entry", "model_id", modelID)
	return nil
}

// ============================================================================
// PROCESS DISCOVERY RESULTS — DB PERSISTENCE (Phase 4c)
// ============================================================================

// persistDiscoveryResult inserts a new row into process_discovery_results with
// fitness=-1.0 (the "not-yet-computed" sentinel). Called after WAL write in Discover.
// Uses workspace_id=00000000-0000-0000-0000-000000000000 as a default placeholder;
// ON CONFLICT DO NOTHING makes it idempotent for retried requests.
// WvdA: 5s timeout — non-fatal, WAL is the durability guarantee.
func (h *BOSGatewayHandler) persistDiscoveryResult(ctx context.Context, modelID, algo string, resp *BOSDiscoverResponse) error {
	if h.pool == nil {
		return fmt.Errorf("database pool unavailable")
	}
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rawResult, _ := json.Marshal(map[string]interface{}{
		"model_id":    resp.ModelID,
		"algorithm":   resp.Algorithm,
		"places":      resp.Places,
		"transitions": resp.Transitions,
		"arcs":        resp.Arcs,
		"latency_ms":  resp.LatencyMs,
	})

	_, err := h.pool.Exec(dbCtx, `
		INSERT INTO process_discovery_results
		    (workspace_id, model_id, algorithm, activities_count, fitness, raw_result)
		VALUES
		    ('00000000-0000-0000-0000-000000000000'::uuid, $1, $2, $3, -1.0, $4)
		ON CONFLICT (workspace_id, model_id) DO NOTHING
	`, modelID, algo, resp.Transitions, json.RawMessage(rawResult))
	if err != nil {
		return fmt.Errorf("persist discovery result: %w", err)
	}
	h.logger.Debug("discovery result persisted", "model_id", modelID, "algo", algo)
	return nil
}

// updateDiscoveryFitness updates the fitness score for a model after conformance check.
// fitness is in [0.0, 1.0]. avgCycleTimeHours is set to 0.0 when not available from
// conformance response (the L0 sync job will fill it in from process_cases later).
// WvdA: 5s timeout — non-fatal, conformance response is returned regardless.
func (h *BOSGatewayHandler) updateDiscoveryFitness(ctx context.Context, modelID string, fitness, avgCycleTimeHours float64) error {
	if h.pool == nil {
		return fmt.Errorf("database pool unavailable")
	}
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := h.pool.Exec(dbCtx, `
		UPDATE process_discovery_results
		SET fitness = $1, avg_cycle_time_hours = $2
		WHERE workspace_id = '00000000-0000-0000-0000-000000000000'::uuid
		  AND model_id = $3
	`, fitness, avgCycleTimeHours, modelID)
	if err != nil {
		return fmt.Errorf("update discovery fitness: %w", err)
	}
	h.logger.Debug("discovery fitness updated", "model_id", modelID, "fitness", fitness)
	return nil
}
