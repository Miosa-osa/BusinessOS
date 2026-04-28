package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// registerAuditRoutes lazily mounts the read-only /api/audit/events endpoint.
// Auth-gated; tenants without DB access skip the route group entirely.
func (h *Handlers) registerAuditRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.pool == nil {
		slog.Warn("audit routes skipped: database pool not available")
		return
	}
	if h.auditQueryHandler == nil {
		h.auditQueryHandler = NewAuditQueryHandler(h.pool)
	}
	RegisterAuditRoutes(api, h.auditQueryHandler, auth)
	slog.Info("audit routes registered at /api/audit/*")
}
