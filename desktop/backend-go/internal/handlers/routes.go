package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// RegisterRoutes is the top-level route registration entry point.
// It sets up shared auth middleware and delegates each domain to its own
// register* method, keeping this file under 50 lines.
func (h *Handlers) RegisterRoutes(api *gin.RouterGroup) {
	// Auth middleware for protected routes - uses Redis cache if available
	var auth gin.HandlerFunc
	if h.sessionCache != nil {
		// Use Redis-cached auth for horizontal scaling
		auth = middleware.CachedAuthMiddleware(h.pool, h.sessionCache)
	} else {
		// Fallback to direct DB auth (single instance mode)
		auth = middleware.AuthMiddleware(h.pool)
	}
	optionalAuth := middleware.OptionalAuthMiddleware(h.pool) // For dev-friendly routes

	h.registerChatRoutes(api, auth)
	h.registerProjectRoutes(api, auth)
	h.registerWorkspaceRoutes(api, auth)
	h.registerCRMRoutes(api, auth)
	h.registerAIRoutes(api, auth)
	h.registerSearchRoutes(api, auth)
	h.registerMemoryRoutes(api, auth)
	h.registerVoiceRoutes(api, auth)
	h.registerNotificationRoutes(api, auth)
	h.registerUserRoutes(api, auth)
	h.registerInfraRoutes(api, auth)
	h.registerAuthRoutes(api, auth)
	h.registerIntegrationRoutes(api, auth, optionalAuth)
	h.registerPlatformRoutes(api, auth)
	h.registerOSARoutes(api, auth)
	h.registerOntologyRoutes(api, auth)
}

// registerOntologyRoutes wires /api/ontology routes via bos CLI bridge.
func (h *Handlers) registerOntologyRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.bosOntologyService == nil {
		return // bos not configured — skip
	}
	ontologyHandler := NewOntologyHandler(h.bosOntologyService)
	RegisterOntologyRoutes(api, ontologyHandler, auth)
}
