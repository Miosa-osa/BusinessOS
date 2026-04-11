package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// registerOptimalRoutes wires up all OptimalOS data endpoints at /api/optimal/*.
// These routes are intentionally unauthenticated — they serve local filesystem
// data for the Tauri desktop frontend and do not require a user session.
func (h *Handlers) registerOptimalRoutes(api *gin.RouterGroup) {
	if h.optimalHandler == nil {
		slog.Warn("OptimalOS routes skipped: optimalHandler not initialized")
		return
	}
	RegisterOptimalRoutes(api, h.optimalHandler)
	slog.Info("OptimalOS routes registered at /api/optimal/*")
}
