package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/handlers"
)

// registerSyncRoutes attaches the cloud-mode sync endpoints to the given API
// router group. Both push and pull are intentional stubs at this stage —
// they validate payloads and return correct shapes but perform no DB work.
//
// Routes registered:
//
//	POST /api/sync/push
//	GET  /api/sync/pull
func registerSyncRoutes(api *gin.RouterGroup, h *handlers.CloudSyncHandler) {
	sync := api.Group("/sync")
	sync.POST("/push", h.Push)
	sync.GET("/pull", h.Pull)
}
