package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// registerConnectorsRoutes mounts /api/connectors/* — list + per-kind sync.
// Lazily constructs the handler the first time it runs; skipped entirely
// when the DB pool isn't available (degraded mode).
func (h *Handlers) registerConnectorsRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.pool == nil {
		slog.Warn("connectors routes skipped: database pool not available")
		return
	}
	if h.connectorsHandler == nil {
		h.connectorsHandler = NewConnectorsHandler(h.pool)
	}
	RegisterConnectorsRoutes(api, h.connectorsHandler, auth)
	slog.Info("connectors routes registered at /api/connectors/*")
}
