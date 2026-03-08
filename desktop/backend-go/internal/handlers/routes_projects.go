package handlers

import "github.com/gin-gonic/gin"

// registerProjectRoutes wires up project and node routes:
// /api/projects, /api/nodes.
func (h *Handlers) registerProjectRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	// Projects routes - /api/projects (extracted handler)
	RegisterProjectRoutes(api, NewProjectHandler(h.pool, h.queryCache, h.notificationTriggers, h.projectAccessService), auth)

	// Nodes routes - /api/nodes
	RegisterNodeRoutes(api, NewNodeHandler(h.pool), auth)
}
