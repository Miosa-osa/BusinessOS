package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/Miosa-osa/OptimalEngine-go/audit"
	"github.com/rhl/businessos-backend/internal/utils"
)

// AuditQueryHandler exposes a read-only window onto the audit_events table.
// Write access is intentionally NOT provided here — events are emitted by
// the subsystem they audit (wiki/compliance/connectors), never by an HTTP
// caller. That preserves the immutability invariant that gives an audit
// trail its compliance value.
type AuditQueryHandler struct {
	logger *audit.Logger
}

// NewAuditQueryHandler wires the handler.
func NewAuditQueryHandler(pool *pgxpool.Pool) *AuditQueryHandler {
	db := stdlib.OpenDBFromPool(pool)
	return &AuditQueryHandler{logger: audit.NewLogger(db)}
}

// RegisterAuditRoutes attaches /api/audit/* endpoints. Auth-gated.
func RegisterAuditRoutes(api *gin.RouterGroup, h *AuditQueryHandler, auth gin.HandlerFunc) {
	g := api.Group("/audit")
	g.Use(auth, middleware.RequireAuth())
	{
		g.GET("/events", h.Query)
	}
}

// Query returns recent audit events filtered by tenant / principal / kind /
// time range. Hard-capped at 500 rows so the endpoint can never become
// pathological for a misconfigured caller.
//
//	GET /api/audit/events?tenant_id=…&principal=…&kind=…&since=…&until=…&limit=…
func (h *AuditQueryHandler) Query(c *gin.Context) {
	opts := audit.QueryOptions{
		TenantID:  c.Query("tenant_id"),
		Principal: c.Query("principal"),
		Kind:      c.Query("kind"),
	}
	if v := c.Query("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			opts.Since = t
		} else {
			utils.RespondBadRequest(c, slog.Default(), "since must be RFC3339")
			return
		}
	}
	if v := c.Query("until"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			opts.Until = t
		} else {
			utils.RespondBadRequest(c, slog.Default(), "until must be RFC3339")
			return
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > 500 {
				n = 500
			}
			opts.Limit = n
		}
	}

	events, err := h.logger.Query(c.Request.Context(), opts)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "audit: query failed",
			"tenant_id", opts.TenantID, "kind", opts.Kind, "error", err)
		utils.RespondInternalError(c, slog.Default(), "query audit events", err)
		return
	}

	out := make([]map[string]any, len(events))
	for i, e := range events {
		out[i] = map[string]any{
			"tenant_id":  e.TenantID,
			"ts":         e.TS,
			"principal":  e.Principal,
			"kind":       e.Kind,
			"target_uri": e.TargetURI,
			"latency_ms": e.LatencyMS,
			"metadata":   e.Metadata,
		}
	}
	c.JSON(http.StatusOK, gin.H{"events": out, "count": len(out)})
}
