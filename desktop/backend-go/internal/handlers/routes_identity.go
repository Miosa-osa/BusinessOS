package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// registerIdentityRoutes mounts /api/identity/{principals,acls}/* behind
// the request-auth middleware.  Each sub-handler is lazily constructed so
// dev / single-instance deployments without the optimal schemas pay no
// startup cost.
func (h *Handlers) registerIdentityRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.pool == nil {
		slog.Warn("identity routes skipped: database pool not available")
		return
	}
	if h.principalHandler == nil {
		h.principalHandler = NewPrincipalHandler(h.pool)
	}
	if h.aclHandler == nil {
		h.aclHandler = NewACLHandler(h.pool)
	}
	RegisterPrincipalRoutes(api, h.principalHandler, auth)
	RegisterACLRoutes(api, h.aclHandler, auth)
	slog.Info("identity routes registered at /api/identity/{principals,acls}/*")
}
