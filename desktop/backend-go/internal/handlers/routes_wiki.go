package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// registerWikiRoutes lazily constructs the WikiHandler the first time it
// runs and mounts every /api/wiki/* endpoint behind the request-auth
// middleware.  Requires a non-nil pgxpool — when the DB isn't available
// we skip wiki entirely and log it once at startup.
func (h *Handlers) registerWikiRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.pool == nil {
		slog.Warn("wiki routes skipped: database pool not available")
		return
	}
	if h.wikiHandler == nil {
		h.wikiHandler = NewWikiHandler(h.pool)
	}
	RegisterWikiRoutes(api, h.wikiHandler, auth)
	slog.Info("wiki routes registered at /api/wiki/*")
}
