package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/integrations/pm4py_rust"
	"github.com/rhl/businessos-backend/internal/utils"
)

// PM4PyRustHandler wraps pm4py-rust client operations
type PM4PyRustHandler struct {
	client *pm4py_rust.Client
	logger *slog.Logger
}

// NewPM4PyRustHandler creates a new pm4py-rust handler
func NewPM4PyRustHandler(baseURL string) *PM4PyRustHandler {
	return &PM4PyRustHandler{
		client: pm4py_rust.NewClient(baseURL),
		logger: slog.Default(),
	}
}

// RegisterRoutes registers pm4py-rust routes
func (h *PM4PyRustHandler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/pm4py")
	{
		group.GET("/health", h.Health)
		group.POST("/discover", h.Discover)
		group.POST("/conformance", h.CheckConformance)
		group.POST("/statistics", h.Statistics)
	}
}

// DiscoverRequest represents a discovery request
type DiscoverRequest struct {
	EventLog json.RawMessage `json:"event_log" binding:"required"`
	Variant  string          `json:"variant,omitempty"`
}

// Health checks pm4py-rust service availability
// @Summary Check pm4py-rust service health
// @Description Returns health status of the pm4py-rust service
// @Tags pm4py
// @Produce json
// @Success 200 {object} pm4py_rust.HealthResponse
// @Router /api/pm4py/health [get]
func (h *PM4PyRustHandler) Health(c *gin.Context) {
	resp, err := h.client.HealthCheckRequest(c.Request.Context())
	if err != nil {
		h.logger.Error("health check failed", "error", err)
		utils.RespondInternalError(c, h.logger, "pm4py health check", err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Discover mines a Petri net from an event log
// @Summary Discover Petri net from event log
// @Description Performs process discovery using Alpha Miner algorithm
// @Tags pm4py
// @Accept json
// @Produce json
// @Param request body DiscoverRequest true "Event log and discovery parameters"
// @Success 200 {object} pm4py_rust.DiscoveryResponse
// @Router /api/pm4py/discover [post]
func (h *PM4PyRustHandler) Discover(c *gin.Context) {
	var req DiscoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	discoveryReq := &pm4py_rust.DiscoveryRequest{
		EventLog: req.EventLog,
		Variant:  req.Variant,
	}

	resp, err := h.client.DiscoveryRequest(c.Request.Context(), discoveryReq)
	if err != nil {
		h.logger.Error("discovery failed", "error", err)
		utils.RespondInternalError(c, h.logger, "process discovery", err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ConformanceCheckRequest represents a conformance check request
type ConformanceCheckRequest struct {
	EventLog json.RawMessage            `json:"event_log" binding:"required"`
	PetriNet pm4py_rust.PetriNetJSON    `json:"petri_net" binding:"required"`
	Method   string                     `json:"method,omitempty"`
}

// CheckConformance checks conformance of an event log against a Petri net
// @Summary Check event log conformance
// @Description Performs conformance checking using Token Replay
// @Tags pm4py
// @Accept json
// @Produce json
// @Param request body ConformanceCheckRequest true "Event log and Petri net"
// @Success 200 {object} pm4py_rust.ConformanceResponse
// @Router /api/pm4py/conformance [post]
func (h *PM4PyRustHandler) CheckConformance(c *gin.Context) {
	var req ConformanceCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	conformanceReq := &pm4py_rust.ConformanceRequest{
		EventLog: req.EventLog,
		PetriNet: req.PetriNet,
		Method:   req.Method,
	}

	resp, err := h.client.ConformanceRequest(c.Request.Context(), conformanceReq)
	if err != nil {
		h.logger.Error("conformance check failed", "error", err)
		utils.RespondInternalError(c, h.logger, "conformance checking", err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// StatisticsAnalysisRequest represents a statistics request
type StatisticsAnalysisRequest struct {
	EventLog              json.RawMessage `json:"event_log" binding:"required"`
	IncludeVariants       bool            `json:"include_variants,omitempty"`
	IncludeResourceMetrics bool           `json:"include_resource_metrics,omitempty"`
	IncludeBottlenecks    bool            `json:"include_bottlenecks,omitempty"`
}

// Statistics analyzes an event log for statistical metrics
// @Summary Analyze event log statistics
// @Description Returns comprehensive statistical analysis of event log
// @Tags pm4py
// @Accept json
// @Produce json
// @Param request body StatisticsAnalysisRequest true "Event log and analysis options"
// @Success 200 {object} pm4py_rust.StatisticsResponse
// @Router /api/pm4py/statistics [post]
func (h *PM4PyRustHandler) Statistics(c *gin.Context) {
	var req StatisticsAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	statsReq := &pm4py_rust.StatisticsRequest{
		EventLog:              req.EventLog,
		IncludeVariants:       req.IncludeVariants,
		IncludeResourceMetrics: req.IncludeResourceMetrics,
		IncludeBottlenecks:    req.IncludeBottlenecks,
	}

	resp, err := h.client.StatisticsRequest(c.Request.Context(), statsReq)
	if err != nil {
		h.logger.Error("statistics analysis failed", "error", err)
		utils.RespondInternalError(c, h.logger, "statistics analysis", err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Close closes the pm4py-rust client
func (h *PM4PyRustHandler) Close() error {
	return h.client.Close()
}
