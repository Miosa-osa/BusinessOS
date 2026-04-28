package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/optimal/audit"
	"github.com/rhl/businessos-backend/internal/optimal/compliance"
	"github.com/rhl/businessos-backend/internal/utils"
)

// RetentionHandler exposes CRUD for retention policies + sweep execution.
// Sweep walks every policy in scope and applies archive / delete / redact
// to expired signals, skipping content under legal hold.
type RetentionHandler struct {
	retention *compliance.Retention
}

// NewRetentionHandler wires the handler. Holds are mandatory — the sweep
// path consults legal_holds per signal and refuses to act on anything held.
func NewRetentionHandler(pool *pgxpool.Pool) *RetentionHandler {
	db := stdlib.OpenDBFromPool(pool)
	holds := compliance.NewLegalHold(db)
	return &RetentionHandler{
		retention: compliance.NewRetention(db, holds, audit.NewLogger(db)),
	}
}

// RegisterRetentionRoutes attaches /api/compliance/retention/* endpoints.
func RegisterRetentionRoutes(api *gin.RouterGroup, h *RetentionHandler, auth gin.HandlerFunc) {
	g := api.Group("/compliance/retention")
	g.Use(auth, middleware.RequireAuth())
	{
		g.POST("/policies", h.AddPolicy)
		g.DELETE("/policies/:id", h.RemovePolicy)
		g.POST("/sweep", h.Sweep)
	}
}

// AddPolicy registers a retention policy.
//
//	POST /api/compliance/retention/policies
//	{
//	  "tenant_id":  "…",
//	  "scope_type": "node|genre|tenant|global",
//	  "scope_value":"ai-masters",
//	  "ttl_days":   90,
//	  "action":     "archive|delete|redact"
//	}
func (h *RetentionHandler) AddPolicy(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req struct {
		TenantID   string `json:"tenant_id"`
		ScopeType  string `json:"scope_type"  binding:"required"`
		ScopeValue string `json:"scope_value"`
		TTLDays    int    `json:"ttl_days"    binding:"required"`
		Action     string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	id, err := h.retention.AddPolicy(c.Request.Context(), compliance.Policy{
		TenantID:   req.TenantID,
		ScopeType:  compliance.ScopeType(req.ScopeType),
		ScopeValue: req.ScopeValue,
		TTLDays:    req.TTLDays,
		Action:     compliance.Action(req.Action),
	})
	if err != nil {
		// Validation errors (bad action, zero ttl) are user-facing 400s;
		// the package returns a free-form error so we forward the message.
		utils.RespondBadRequest(c, slog.Default(), err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// RemovePolicy deletes a policy.
//
//	DELETE /api/compliance/retention/policies/:id
func (h *RetentionHandler) RemovePolicy(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.RespondBadRequest(c, slog.Default(), "policy id must be a positive integer")
		return
	}
	if err := h.retention.RemovePolicy(c.Request.Context(), id); err != nil {
		// Phase-1: the package returns "policy not found" or a wrapped DB error.
		// Treat the well-known message as 404, otherwise 500.
		if err.Error() == "compliance/retention: policy not found" {
			utils.RespondNotFound(c, slog.Default(), "retention policy")
			return
		}
		slog.ErrorContext(c.Request.Context(), "compliance/retention: remove failed",
			"id", id, "error", err)
		utils.RespondInternalError(c, slog.Default(), "remove retention policy", err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Sweep runs every policy for the requested tenant and applies actions to
// expired signals. dry_run=true is the safe default — explicit `dry_run=false`
// is required to mutate.
//
//	POST /api/compliance/retention/sweep
//	{ "tenant_id": "…", "dry_run": true }
func (h *RetentionHandler) Sweep(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req struct {
		TenantID string `json:"tenant_id"`
		DryRun   *bool  `json:"dry_run"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}
	dry := true
	if req.DryRun != nil {
		dry = *req.DryRun
	}

	report, err := h.retention.Sweep(c.Request.Context(), compliance.SweepOptions{
		TenantID: req.TenantID,
		DryRun:   dry,
	})
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "compliance/retention: sweep failed",
			"tenant_id", req.TenantID, "dry_run", dry, "error", err)
		utils.RespondInternalError(c, slog.Default(), "retention sweep", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"policies_evaluated": report.PoliciesEvaluated,
		"actions_taken":      report.ActionsTaken,
		"skipped_held":       report.SkippedHeld,
		"errors":             report.Errors,
		"by_action":          report.ByAction,
		"dry_run":            dry,
	})
}
