package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/Miosa-osa/OptimalEngine-go/audit"
	"github.com/Miosa-osa/OptimalEngine-go/compliance"
	"github.com/rhl/businessos-backend/internal/utils"
)

// LegalHoldHandler exposes place / release / list-active for legal holds.
// A hold freezes retention and blocks erasure for the named signal; releasing
// re-arms both. Place + Release each emit a "legal_hold.placed" /
// "legal_hold.released" audit event so litigation discovery can reconstruct
// the timeline.
type LegalHoldHandler struct {
	holds *compliance.LegalHold
	audit *audit.Logger
}

// NewLegalHoldHandler wires the handler.
func NewLegalHoldHandler(pool *pgxpool.Pool) *LegalHoldHandler {
	db := stdlib.OpenDBFromPool(pool)
	return &LegalHoldHandler{
		holds: compliance.NewLegalHold(db),
		audit: audit.NewLogger(db),
	}
}

// RegisterLegalHoldRoutes attaches /api/compliance/holds/* endpoints.
func RegisterLegalHoldRoutes(api *gin.RouterGroup, h *LegalHoldHandler, auth gin.HandlerFunc) {
	g := api.Group("/compliance/holds")
	g.Use(auth, middleware.RequireAuth())
	{
		g.POST("", h.Place)
		g.GET("", h.Active)
		g.DELETE("/:id", h.Release)
	}
}

// Place creates a new hold.
//
//	POST /api/compliance/holds
//	{ "tenant_id": "…", "signal_id": "…", "reason": "…" }
func (h *LegalHoldHandler) Place(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req struct {
		TenantID string `json:"tenant_id"`
		SignalID string `json:"signal_id" binding:"required"`
		Reason   string `json:"reason"    binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	heldBy := "user:" + user.ID
	id, err := h.holds.Place(c.Request.Context(), req.TenantID, req.SignalID, heldBy, req.Reason)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "compliance/holds: place failed",
			"tenant_id", req.TenantID, "signal_id", req.SignalID, "error", err)
		utils.RespondInternalError(c, slog.Default(), "place legal hold", err)
		return
	}

	h.audit.LogKind(c.Request.Context(), audit.KindLegalHoldPlaced,
		audit.WithTenant(req.TenantID),
		audit.WithPrincipal(heldBy),
		audit.WithTargetURI("signal:"+req.SignalID),
		audit.WithMetadata(map[string]any{"hold_id": id, "reason": req.Reason}),
	)
	c.JSON(http.StatusCreated, gin.H{"id": id, "signal_id": req.SignalID, "held_by": heldBy})
}

// Active lists every un-released hold for a tenant.
//
//	GET /api/compliance/holds?tenant_id=…
func (h *LegalHoldHandler) Active(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	holds, err := h.holds.Active(c.Request.Context(), tenantID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "compliance/holds: list failed",
			"tenant_id", tenantID, "error", err)
		utils.RespondInternalError(c, slog.Default(), "list legal holds", err)
		return
	}
	out := make([]map[string]any, len(holds))
	for i, hd := range holds {
		out[i] = map[string]any{
			"id":          hd.ID,
			"tenant_id":   hd.TenantID,
			"signal_id":   hd.SignalID,
			"held_by":     hd.HeldBy,
			"reason":      hd.Reason,
			"placed_at":   hd.PlacedAt,
			"released_at": hd.ReleasedAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{"holds": out, "count": len(out)})
}

// Release marks a hold as released (sets released_at = NOW()). Returns 404
// when the hold doesn't exist or has already been released.
//
//	DELETE /api/compliance/holds/:id
func (h *LegalHoldHandler) Release(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.RespondBadRequest(c, slog.Default(), "hold id must be a positive integer")
		return
	}

	if err := h.holds.Release(c.Request.Context(), id); err != nil {
		if errors.Is(err, compliance.ErrHoldNotFound) {
			utils.RespondNotFound(c, slog.Default(), "legal hold")
			return
		}
		slog.ErrorContext(c.Request.Context(), "compliance/holds: release failed",
			"id", id, "error", err)
		utils.RespondInternalError(c, slog.Default(), "release legal hold", err)
		return
	}

	h.audit.LogKind(c.Request.Context(), audit.KindLegalHoldReleased,
		audit.WithPrincipal("user:"+user.ID),
		audit.WithMetadata(map[string]any{"hold_id": id}),
	)
	c.Status(http.StatusNoContent)
}
