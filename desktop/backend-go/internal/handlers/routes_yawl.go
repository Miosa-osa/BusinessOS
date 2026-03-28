package handlers

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// registerYawlRoutes wires /api/yawl routes for the YAWL v6 engine integration.
//
// Optional static bearer token: set YAWLV6_API_TOKEN env var. If empty,
// bearer auth is skipped (dev mode — StaticBearerAuth("") is a pass-through).
// JWT auth is always applied (same as all other protected route groups).
func (h *Handlers) registerYawlRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	yawlHandler := NewYawlHandler(slog.Default())
	yawlGroup := api.Group("/yawl")

	token := os.Getenv("YAWLV6_API_TOKEN")
	if token != "" {
		yawlGroup.Use(middleware.StaticBearerAuth(token))
	}
	yawlGroup.Use(auth)

	yawlGroup.GET("/health", yawlHandler.GetHealth)
	yawlGroup.POST("/conformance", yawlHandler.CheckConformance)
	yawlGroup.POST("/spec", yawlHandler.BuildSpec)
	yawlGroup.GET("/spec/load", yawlHandler.LoadSpec)
	yawlGroup.GET("/specs", yawlHandler.ListSpecs)
	yawlGroup.GET("/real-data", yawlHandler.ListRealData)
	yawlGroup.GET("/real-data/:name", yawlHandler.GetRealData)
}
