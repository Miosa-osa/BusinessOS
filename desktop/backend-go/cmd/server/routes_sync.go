package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// registerSyncRoutes attaches the cloud-mode sync endpoints to the given API
// router group. auth is the shared authentication middleware — all sync routes
// require an authenticated session.
//
// Routes registered:
//
//	POST /api/sync/push    — receive a batch of changes from a local client
//	GET  /api/sync/pull    — return changes since a given timestamp
//	GET  /api/sync/status  — last push/pull timestamps + connectivity signal
func registerSyncRoutes(api *gin.RouterGroup, h *handlers.CloudSyncHandler, auth gin.HandlerFunc) {
	sync := api.Group("/sync")
	sync.Use(auth, middleware.RequireAuth())
	sync.POST("/push", h.Push)
	sync.GET("/pull", h.Pull)
	sync.GET("/status", h.Status)
}
