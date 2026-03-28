package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ComplianceHandler handles zero-touch compliance endpoints.
type ComplianceHandler struct {
	complianceService *services.ComplianceService
	logger            *slog.Logger
}

// NewComplianceHandler constructs a ComplianceHandler.
func NewComplianceHandler(complianceService *services.ComplianceService, logger *slog.Logger) *ComplianceHandler {
	return &ComplianceHandler{
		complianceService: complianceService,
		logger:            logger,
	}
}

// registerComplianceRoutes wires /api/compliance routes (authenticated).
func (h *Handlers) registerComplianceRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.complianceService == nil {
		slog.Debug("Skipping compliance routes, service not initialized")
		return
	}
	// Skip when called from the /v1 group: the legacy /compliance/* routes live on /api only.
	// The /api/v1/compliance/* routes are registered by registerComplianceAPIRoutes (called below
	// when !v1) to prevent duplicate registration when RegisterRoutes is invoked for both prefixes.
	if strings.Contains(api.BasePath(), "/v1") {
		return
	}

	complianceH := NewComplianceHandler(h.complianceService, slog.Default())
	compliance := api.Group("/compliance")
	compliance.Use(auth)
	{
		compliance.GET("/status", complianceH.GetComplianceStatus)
		compliance.GET("/audit-trail", complianceH.GetAuditTrail)
		compliance.GET("/audit-trail/verify/:session_id", complianceH.VerifyAuditChain)
		compliance.POST("/evidence/collect", complianceH.CollectEvidence)
		compliance.GET("/gap-analysis", complianceH.GetGapAnalysis)
		compliance.POST("/remediation", complianceH.CreateRemediation)
		compliance.POST("/verify", complianceH.VerifyCompliance)
	}
	slog.Info("Compliance routes registered at /api/compliance/*")

	// Alias: Canopy adapter calls /api/bos/compliance/verify (documented in CLAUDE.md)
	api.POST("/bos/compliance/verify", auth, complianceH.VerifyCompliance)

	// Register Compliance REST API routes (Agent 27: Compliance REST API)
	// Guard: only register the /v1/compliance sub-group when the parent is NOT already /v1
	// to prevent duplicate registration when RegisterRoutes is called for both /api and /api/v1.
	if !strings.Contains(api.BasePath(), "/v1") {
		h.registerComplianceAPIRoutes(api, auth)
	}
}

// GetComplianceStatus handles GET /api/compliance/status.
// Returns the overall compliance score, per-domain breakdown, and certificates.
func (h *ComplianceHandler) GetComplianceStatus(c *gin.Context) {
	status, err := h.complianceService.GetStatus(c.Request.Context())
	if err != nil {
		utils.RespondInternalError(c, h.logger, "get compliance status", err)
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAuditTrail handles GET /api/compliance/audit-trail.
// Query params: session_id, from, to, tool_name, limit, offset.
// Returns hash-chain verified audit entries from OSA.
// Returns 503 if OSA is unavailable and no cache exists.
func (h *ComplianceHandler) GetAuditTrail(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		utils.RespondBadRequest(c, h.logger, "session_id query parameter is required")
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var from, to time.Time
	if f := c.Query("from"); f != "" {
		parsed, err := time.Parse(time.RFC3339, f)
		if err == nil {
			from = parsed
		}
	}
	if t := c.Query("to"); t != "" {
		parsed, err := time.Parse(time.RFC3339, t)
		if err == nil {
			to = parsed
		}
	}

	params := services.AuditTrailParams{
		SessionID: sessionID,
		From:      from,
		To:        to,
		ToolName:  c.Query("tool_name"),
		Limit:     limit,
		Offset:    offset,
	}

	result, err := h.complianceService.GetAuditTrail(c.Request.Context(), params)
	if err != nil {
		// Check if error indicates OSA unavailability
		if strings.Contains(err.Error(), "OSA unavailable") {
			h.logger.Warn("OSA unavailable, returning 503", "session_id", sessionID, "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "OSA audit trail service unavailable",
				"details": err.Error(),
			})
			return
		}
		utils.RespondInternalError(c, h.logger, "get audit trail", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// VerifyAuditChain handles GET /api/compliance/audit-trail/verify/:session_id.
// Verifies the integrity of the audit chain for a session.
func (h *ComplianceHandler) VerifyAuditChain(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		utils.RespondBadRequest(c, h.logger, "session_id path parameter is required")
		return
	}

	result, err := h.complianceService.VerifyAuditChain(c.Request.Context(), sessionID)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "verify audit chain", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CollectEvidence handles POST /api/compliance/evidence/collect.
// Body: {"domain": "data_security", "period": "2026-Q1"}
func (h *ComplianceHandler) CollectEvidence(c *gin.Context) {
	var req services.EvidenceCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	result, err := h.complianceService.CollectEvidence(c.Request.Context(), req)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "collect evidence", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGapAnalysis handles GET /api/compliance/gap-analysis.
// Query param: framework (SOC2, HIPAA, GDPR, SOX). Defaults to SOC2.
func (h *ComplianceHandler) GetGapAnalysis(c *gin.Context) {
	framework := c.Query("framework")

	result, err := h.complianceService.GetGapAnalysis(c.Request.Context(), framework)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "get gap analysis", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateRemediation handles POST /api/compliance/remediation.
// Body: {"gap_id": "...", "priority": "high", "assignee": "...", "due_date": "2026-04-01"}
func (h *ComplianceHandler) CreateRemediation(c *gin.Context) {
	var req services.RemediationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	task, err := h.complianceService.CreateRemediation(c.Request.Context(), req)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "create remediation", err)
		return
	}

	c.JSON(http.StatusCreated, task)
}

// VerifyCompliance handles POST /api/compliance/verify.
// JTBD Wave 12 scenario 3 compliance verification endpoint.
// Body: {"workspace_id": "...", "framework": "soc2"}
// Returns compliance status, findings count, and remediation progress.
func (h *ComplianceHandler) VerifyCompliance(c *gin.Context) {
	var req services.ComplianceVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	if req.WorkspaceID == "" || req.Framework == "" {
		utils.RespondBadRequest(c, h.logger, "workspace_id and framework are required")
		return
	}

	result, err := h.complianceService.VerifyCompliance(c.Request.Context(), req)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "verify compliance", err)
		return
	}

	c.JSON(http.StatusOK, result)
}
