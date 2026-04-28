package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// registerComplianceRoutes mounts every /api/compliance/* endpoint behind
// the request-auth middleware.  Each sub-handler is lazily constructed the
// first time we hit this method so single-instance / dev deployments without
// the optimal schemas pay no startup cost.
func (h *Handlers) registerComplianceRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.pool == nil {
		slog.Warn("compliance routes skipped: database pool not available")
		return
	}
	if h.dsarHandler == nil {
		h.dsarHandler = NewDSARHandler(h.pool)
	}
	if h.erasureHandler == nil {
		h.erasureHandler = NewErasureHandler(h.pool)
	}
	if h.legalHoldHandler == nil {
		h.legalHoldHandler = NewLegalHoldHandler(h.pool)
	}
	if h.retentionHandler == nil {
		h.retentionHandler = NewRetentionHandler(h.pool)
	}

	RegisterDSARRoutes(api, h.dsarHandler, auth)
	RegisterErasureRoutes(api, h.erasureHandler, auth)
	RegisterLegalHoldRoutes(api, h.legalHoldHandler, auth)
	RegisterRetentionRoutes(api, h.retentionHandler, auth)
	slog.Info("compliance routes registered at /api/compliance/{dsar,erasure,holds,retention}/*")
}
