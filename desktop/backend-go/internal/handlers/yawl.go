package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/rhl/businessos-backend/internal/integrations/yawlv6"
	semconv "github.com/rhl/businessos-backend/internal/semconv"
)

const yawlTracerName = "businessos.yawl"

// YawlHandler handles YAWL v6 engine proxy endpoints.
type YawlHandler struct {
	client *yawlv6.Client
	logger *slog.Logger
	tracer trace.Tracer
}

// NewYawlHandler constructs a YawlHandler.
// The YAWL engine URL is read from YAWLV6_URL env var inside yawlv6.NewClient().
func NewYawlHandler(logger *slog.Logger) *YawlHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &YawlHandler{
		client: yawlv6.NewClient(),
		logger: logger,
		tracer: otel.Tracer(yawlTracerName),
	}
}

// ============================================================================
// REQUEST / RESPONSE TYPES
// ============================================================================

// yawlConformanceRequest is the body accepted by POST /yawl/conformance.
type yawlConformanceRequest struct {
	SpecXML      string          `json:"spec_xml" binding:"required"`
	EventLogJSON json.RawMessage `json:"event_log" binding:"required"`
}

// yawlBuildSpecRequest is the body accepted by POST /yawl/spec.
type yawlBuildSpecRequest struct {
	Type     string   `json:"type" binding:"required"` // "sequence" or "parallel"
	Tasks    []string `json:"tasks"`
	Trigger  string   `json:"trigger"`
	Branches []string `json:"branches"`
}

// ============================================================================
// HANDLERS
// ============================================================================

// GetHealth handles GET /api/yawl/health.
// Returns 200 with {"status":"ok"} when the YAWL engine is reachable,
// or 502 with an error when it is not.
func (h *YawlHandler) GetHealth(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), semconv.YawlCaseSpan)
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	span.SetAttributes(semconv.YawlEventType("health_check"))

	if err := h.client.Health(ctx); err != nil {
		h.logger.Warn("yawl health check failed", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "yawl engine unreachable")
		c.JSON(http.StatusBadGateway, gin.H{"error": "YAWL engine unreachable"})
		return
	}

	span.SetStatus(codes.Ok, "")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CheckConformance handles POST /api/yawl/conformance.
// Body: {"spec_xml":"<xml...>","event_log":[...]}
// Returns yawlv6.ConformanceResult on success.
func (h *YawlHandler) CheckConformance(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), semconv.YawlTaskExecutionSpan)
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	span.SetAttributes(semconv.YawlEventType("conformance_check"))

	var req yawlConformanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("yawl conformance: invalid request body", "error", err)
		span.SetStatus(codes.Error, "invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "spec_xml and event_log are required"})
		return
	}

	result, err := h.client.CheckConformance(ctx, req.SpecXML, []byte(req.EventLogJSON))
	if err != nil {
		h.logger.Error("yawl conformance: client error", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "conformance check failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "YAWL conformance check failed"})
		return
	}

	span.SetStatus(codes.Ok, "")
	c.JSON(http.StatusOK, result)
}

// BuildSpec handles POST /api/yawl/spec.
// Body: {"type":"sequence","tasks":["A","B"]}
//    or {"type":"parallel","trigger":"Start","branches":["A","B"]}
// Returns generated YAWL XML as text/xml.
func (h *YawlHandler) BuildSpec(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), semconv.YawlCaseSpan)
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	var req yawlBuildSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("yawl build spec: invalid request body", "error", err)
		span.SetStatus(codes.Error, "invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
		return
	}

	span.SetAttributes(semconv.YawlSpecUri(req.Type))

	var xml string
	switch req.Type {
	case "sequence":
		if len(req.Tasks) == 0 {
			span.SetStatus(codes.Error, "tasks empty")
			c.JSON(http.StatusBadRequest, gin.H{"error": "tasks must be a non-empty array for type=sequence"})
			return
		}
		xml = yawlv6.BuildSequenceSpec(req.Tasks)
	case "parallel":
		if req.Trigger == "" || len(req.Branches) == 0 {
			span.SetStatus(codes.Error, "trigger or branches missing")
			c.JSON(http.StatusBadRequest, gin.H{"error": "trigger and branches are required for type=parallel"})
			return
		}
		xml = yawlv6.BuildParallelSplitSpec(req.Trigger, req.Branches)
	default:
		span.SetStatus(codes.Error, "unknown spec type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'sequence' or 'parallel'"})
		return
	}

	span.SetStatus(codes.Ok, "")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}

// LoadSpec handles GET /api/yawl/spec/load.
// Query param: pattern_id (e.g. "WCP-1", "WCP1", "WCP01")
// Returns the raw YAWL spec XML as text/xml.
func (h *YawlHandler) LoadSpec(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), semconv.YawlCaseSpan)
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	patternID := c.Query("pattern_id")
	if patternID == "" {
		span.SetStatus(codes.Error, "missing pattern_id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "pattern_id query parameter is required"})
		return
	}

	span.SetAttributes(semconv.YawlSpecUri(patternID))

	xml, err := h.client.LoadSpec(patternID)
	if err != nil {
		h.logger.Warn("yawl load spec: not found or unreadable", "pattern_id", patternID, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "spec load failed")
		c.JSON(http.StatusNotFound, gin.H{"error": "spec not found: " + patternID})
		return
	}

	span.SetStatus(codes.Ok, "")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}

// ListSpecs handles GET /api/yawl/specs.
// Returns JSON array of all WCP pattern specs found in exampleSpecs/wcp-patterns/.
func (h *YawlHandler) ListSpecs(c *gin.Context) {
	_, span := h.tracer.Start(c.Request.Context(), semconv.YawlCaseSpan)
	defer span.End()

	entries, err := h.client.ListPatterns()
	if err != nil {
		h.logger.Error("yawl list specs: scan failed", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "list specs failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list WCP pattern specs"})
		return
	}

	span.SetStatus(codes.Ok, "")
	c.JSON(http.StatusOK, gin.H{"specs": entries, "count": len(entries)})
}

// ListRealData handles GET /api/yawl/real-data.
// Returns JSON array of available real-world process spec datasets.
func (h *YawlHandler) ListRealData(c *gin.Context) {
	_, span := h.tracer.Start(c.Request.Context(), semconv.YawlCaseSpan)
	defer span.End()

	entries, err := h.client.ListRealData()
	if err != nil {
		h.logger.Error("yawl list real-data: scan failed", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "list real-data failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list real-data specs"})
		return
	}

	span.SetStatus(codes.Ok, "")
	c.JSON(http.StatusOK, gin.H{"datasets": entries, "count": len(entries)})
}

// GetRealData handles GET /api/yawl/real-data/:name.
// Returns the raw YAWL spec XML for a named real-world dataset.
func (h *YawlHandler) GetRealData(c *gin.Context) {
	_, span := h.tracer.Start(c.Request.Context(), semconv.YawlCaseSpan)
	defer span.End()

	name := c.Param("name")
	span.SetAttributes(semconv.YawlSpecUri(name))

	xml, err := h.client.LoadRealData(name)
	if err != nil {
		h.logger.Warn("yawl get real-data: not found", "name", name, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "real-data spec not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "real-data spec not found: " + name})
		return
	}

	span.SetStatus(codes.Ok, "")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xml))
}
