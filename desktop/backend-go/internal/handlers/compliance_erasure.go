package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/Miosa-osa/OptimalEngine-go/audit"
	"github.com/Miosa-osa/OptimalEngine-go/compliance"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ErasureHandler exposes the GDPR Article 17 right-to-delete workflow.
// Every erase consults the legal-hold store first; held content blocks
// the operation unless `force` is explicitly set. Audit events stay
// behind, anonymized rather than deleted.
type ErasureHandler struct {
	erasure *compliance.Erasure
}

// NewErasureHandler wires the erasure engine. Holds + audit are mandatory.
func NewErasureHandler(pool *pgxpool.Pool) *ErasureHandler {
	db := stdlib.OpenDBFromPool(pool)
	holds := compliance.NewLegalHold(db)
	return &ErasureHandler{
		erasure: compliance.NewErasure(db, holds, audit.NewLogger(db)),
	}
}

// RegisterErasureRoutes attaches /api/compliance/erase routes.
// All routes are auth-gated AND require an explicit "confirm" payload field
// (validated inside the handler) so accidental double-tap deletions can't
// fire — the destructive nature of erasure warrants belt-and-braces.
func RegisterErasureRoutes(api *gin.RouterGroup, h *ErasureHandler, auth gin.HandlerFunc) {
	g := api.Group("/compliance/erasure")
	g.Use(auth, middleware.RequireAuth())
	{
		g.POST("/preview", h.Preview)
		g.POST("/erase", h.Erase)
	}
}

// Preview is the dry-run path: returns the same EraseReport shape with the
// counts that would be deleted, but performs no mutation.
//
//	POST /api/compliance/erasure/preview
//	{ "principal_id": "user:…", "tenant_id": "…", "force": false }
func (h *ErasureHandler) Preview(c *gin.Context) {
	h.run(c, true)
}

// Erase performs the destructive operation. Requires "confirm": true in
// the body — a literal, explicit acknowledgment that this is destructive.
//
//	POST /api/compliance/erasure/erase
//	{ "principal_id": "user:…", "tenant_id": "…", "force": false, "confirm": true }
func (h *ErasureHandler) Erase(c *gin.Context) {
	h.run(c, false)
}

func (h *ErasureHandler) run(c *gin.Context, preview bool) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req struct {
		PrincipalID string `json:"principal_id" binding:"required"`
		TenantID    string `json:"tenant_id"`
		Force       bool   `json:"force"`
		Confirm     bool   `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}
	if !preview && !req.Confirm {
		utils.RespondBadRequest(c, slog.Default(),
			"erasure requires \"confirm\": true in the request body")
		return
	}

	report, err := h.erasure.Erase(c.Request.Context(), req.PrincipalID, compliance.EraseOptions{
		TenantID: req.TenantID,
		Actor:    "user:" + user.ID,
		Force:    req.Force,
		DryRun:   preview,
	})
	if errors.Is(err, compliance.ErrHeldContent) {
		// Held content is a 409 — the request is valid; the system state
		// just doesn't allow it without the force flag.
		utils.RespondConflict(c, slog.Default(),
			"principal has content under legal hold — pass \"force\": true to override")
		return
	}
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "compliance/erasure: failed",
			"principal_id", req.PrincipalID, "tenant_id", req.TenantID,
			"force", req.Force, "preview", preview, "error", err)
		utils.RespondInternalError(c, slog.Default(), "erasure", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"principal_id":    report.PrincipalID,
		"tenant_id":       report.TenantID,
		"deleted":         report.Deleted,
		"events_redacted": report.EventsRedacted,
		"held":            report.Held,
		"forced":          report.Forced,
		"dry_run":         report.DryRun,
		"completed_at":    report.CompletedAt,
	})
}
