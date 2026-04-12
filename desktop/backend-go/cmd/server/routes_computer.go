package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/handlers"
)

// registerComputerRoutes wires up all Computer and Billing API endpoints.
// These routes are intentionally unauthenticated at the transport layer —
// the handlers return mock data and will gain auth middleware once MIOSA
// integration is complete.
func registerComputerRoutes(api *gin.RouterGroup, ch *handlers.ComputerHandler, bh *handlers.BillingHandler) {
	comp := api.Group("/computer")
	{
		comp.GET("", ch.GetComputer)
		comp.POST("", ch.CreateComputer)
		comp.PUT("/upgrade", ch.UpgradeComputer)
		comp.DELETE("", ch.DeleteComputer)
		comp.GET("/metrics", ch.GetMetrics)
		comp.GET("/runtimes", ch.GetRuntimes)
		comp.POST("/runtimes/:name/start", ch.StartRuntime)
		comp.POST("/runtimes/:name/stop", ch.StopRuntime)
		comp.GET("/terminal-session", ch.GetTerminalSession)
		comp.GET("/desktop-stream", ch.GetDesktopStream)
	}

	billing := api.Group("/billing")
	{
		billing.GET("/plans", bh.GetPlans)
		billing.GET("/subscription", bh.GetSubscription)
		billing.POST("/subscribe", bh.Subscribe)
		billing.POST("/credits/purchase", bh.PurchaseCredits)
	}
}
