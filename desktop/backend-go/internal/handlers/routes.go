package handlers

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/integrations/pm4py_rust"
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

	// JWT auth for API-to-API communication (e.g., pm4py-rust to /api/bos/progress)
	jwtAuth := middleware.JWTAuth(h.cfg.SecretKey)

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
	h.registerComplianceRoutes(api, auth)
	h.registerTransactionRoutes(api, auth)
	h.registerMeshRoutes(api, auth)
	h.registerBOSProgressRoutes(api, jwtAuth)
	h.registerBOSGatewayRoutes(api)
	h.registerLinkedInRoutes(api, auth)
	h.registerBoardRoutes(api)
	h.registerPM4PyDashboardRoutes(api)
}

// registerOntologyRoutes wires /api/ontology routes via bos CLI bridge.
func (h *Handlers) registerOntologyRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.bosOntologyService == nil {
		return // bos not configured — skip
	}
	ontologyHandler := NewOntologyHandler(h.bosOntologyService)
	RegisterOntologyRoutes(api, ontologyHandler, auth)
}

// registerTransactionRoutes wires /api/bos/tx routes for 2PC transaction management.
func (h *Handlers) registerTransactionRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h.transactionHandler == nil {
		return // transactions not configured — skip
	}
	h.transactionHandler.RegisterRoutes(api)
}

// registerMeshRoutes wires /api/mesh routes for data mesh federation.
func (h *Handlers) registerMeshRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	meshHandler := NewDataMeshHandler("http://localhost:3030", nil)
	RegisterDataMeshRoutes(api, meshHandler, auth)
}

// registerBoardRoutes wires /api/board routes for Board Chair Intelligence.
// Unauthenticated by design at the handler level — callers may add middleware via the api group.
func (h *Handlers) registerBoardRoutes(api *gin.RouterGroup) {
	boardHandler := NewBoardHandler()
	RegisterBoardRoutes(api, boardHandler)
}

// registerPM4PyDashboardRoutes wires POST /api/pm4py/dashboard-kpi.
// The pm4py-rust base URL is read from the PM4PY_RUST_URL environment variable
// (default: http://localhost:8090), consistent with the bos gateway pattern.
func (h *Handlers) registerPM4PyDashboardRoutes(api *gin.RouterGroup) {
	baseURL := os.Getenv("PM4PY_RUST_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8090"
	}
	pm4pyClient := pm4py_rust.NewClient(baseURL)
	dashHandler := NewPM4PyDashboardHandler(pm4pyClient)
	api.POST("/pm4py/dashboard-kpi", dashHandler.GetDashboardKPI)
}

// registerBOSProgressRoutes wires /api/bos/progress route for external progress event reception
// from pm4py-rust progress events.
func (h *Handlers) registerBOSProgressRoutes(api *gin.RouterGroup, jwtAuth gin.HandlerFunc) {
	// POST /api/bos/progress — receives progress events from pm4py-rust
	// Requires JWT Bearer token in Authorization header to prevent unauthorized progress injection
	api.POST("/bos/progress", jwtAuth, ReceiveExternalProgressEventHandler)
}
