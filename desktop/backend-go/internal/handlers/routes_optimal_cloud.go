package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// registerOptimalCloudRoutes wires up all OptimalOS cloud engine endpoints at
// /api/optimal-cloud/*.  Unlike the local /api/optimal/* routes, these are
// auth-gated and backed by PostgreSQL for multi-tenant SaaS deployments.
// The handler is lazily constructed the first time this is called so that
// single-instance desktop deployments (no DB) incur no startup cost.
func (h *Handlers) registerOptimalCloudRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.pool == nil {
		slog.Warn("OptimalOS cloud routes skipped: database pool not available")
		return
	}

	if h.optimalCloudHandler == nil {
		h.optimalCloudHandler = NewOptimalCloudHandler(h.pool)
	}

	RegisterOptimalCloudRoutes(api, h.optimalCloudHandler, auth)
	slog.Info("OptimalOS cloud routes registered at /api/optimal-cloud/*")
}
