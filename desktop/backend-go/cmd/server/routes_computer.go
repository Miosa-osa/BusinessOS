package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// registerComputerRoutes wires up all Computer and Billing API endpoints.
// auth is the shared authentication middleware; it is applied to all routes
// except GET /billing/plans (public pricing page) and the Stripe webhook
// (registered separately on the bare router with no auth).
func registerComputerRoutes(api *gin.RouterGroup, ch *handlers.ComputerHandler, bh *handlers.BillingHandler, auth gin.HandlerFunc) {
	comp := api.Group("/computer")
	comp.Use(auth, middleware.RequireAuth())
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

	// GET /billing/plans is intentionally public — no auth required to view pricing.
	billing := api.Group("/billing")
	billing.GET("/plans", bh.GetPlans)

	// All other billing routes require authentication.
	billingAuth := api.Group("/billing")
	billingAuth.Use(auth, middleware.RequireAuth())
	{
		billingAuth.GET("/subscription", bh.GetSubscription)
		billingAuth.POST("/subscribe", bh.Subscribe)
		billingAuth.POST("/credits/purchase", bh.PurchaseCredits)
		billingAuth.POST("/portal", bh.ManageSubscription)
		// NOTE: /billing/webhook is registered separately (no auth, raw body required).
	}
}

// RegisterStripeWebhookRoute registers the Stripe webhook endpoint on the
// bare router (no auth middleware, no body parsing) so Stripe's signature
// verification has access to the raw request body.
func RegisterStripeWebhookRoute(router gin.IRouter, bh *handlers.BillingHandler) {
	router.POST("/api/v1/billing/webhook", bh.HandleStripeWebhook)
	router.POST("/api/billing/webhook", bh.HandleStripeWebhook)
}
