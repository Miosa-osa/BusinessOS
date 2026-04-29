package handlers

import (
	"log/slog"

	"github.com/Miosa-osa/OptimalEngine-go/connectors/adapters"
	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/connectorbridges"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// registerIntegrationRoutes wires up all integration and module management routes:
// /api/integrations (provider-based router + user integration management),
// /api/modules (browse by module), /api/modules (custom modules CRUD),
// /api/onboarding.
//
// Returns the IntegrationRouter so that dependent routes (e.g. onboarding) can
// extract services from it.
func (h *Handlers) registerIntegrationRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, optionalAuth gin.HandlerFunc) {
	// Initialize the IntegrationRouter which manages all integration providers
	// (Google, Slack, Notion) with their OAuth flows, data sync, and API handlers.
	integrationRouter := NewIntegrationRouter(h.pool)
	h.integrationRouter = integrationRouter

	// Wire OptimalEngine connector adapters to their live OAuth providers.
	// Each setter is idempotent; missing providers leave the adapter in
	// stub mode (Sync returns ErrNotImplemented). Must run BEFORE
	// registerConnectorRoutes so the first sync request finds a wired
	// fetcher.
	if gmailSvc := integrationRouter.GetGoogleGmailService(); gmailSvc != nil {
		adapters.SetGmailFetcher(connectorbridges.NewGmailFetcher(gmailSvc))
		slog.Info("connectors: gmail fetcher wired to live OAuth service")
	}
	if slackProv := integrationRouter.GetSlackProvider(); slackProv != nil {
		adapters.SetSlackFetcher(connectorbridges.NewSlackFetcher(slackProv))
		slog.Info("connectors: slack fetcher wired to live OAuth service")
	}
	if notionProv := integrationRouter.GetNotionProvider(); notionProv != nil {
		adapters.SetNotionFetcher(connectorbridges.NewNotionFetcher(notionProv))
		slog.Info("connectors: notion fetcher wired to live OAuth service")
	}
	if linearProv := integrationRouter.GetLinearProvider(); linearProv != nil {
		adapters.SetLinearFetcher(connectorbridges.NewLinearFetcher(linearProv))
		slog.Info("connectors: linear fetcher wired to live OAuth service")
	}
	if hubspotProv := integrationRouter.GetHubSpotProvider(); hubspotProv != nil {
		adapters.SetHubspotFetcher(connectorbridges.NewHubspotFetcher(hubspotProv))
		slog.Info("connectors: hubspot fetcher wired to live OAuth service")
	}
	// Drive shares the Google provider (just needs drive.readonly scope).
	if googleProv := integrationRouter.GetGoogleProvider(); googleProv != nil {
		adapters.SetDriveFetcher(connectorbridges.NewDriveFetcher(googleProv))
		slog.Info("connectors: drive fetcher wired to live OAuth service")
	}
	// OneDrive + Teams share the Microsoft provider.
	if msProv := integrationRouter.GetMicrosoftProvider(); msProv != nil {
		adapters.SetOnedriveFetcher(connectorbridges.NewOnedriveFetcher(msProv))
		adapters.SetTeamsFetcher(connectorbridges.NewTeamsFetcher(msProv))
		slog.Info("connectors: onedrive + teams fetchers wired to live OAuth service")
	}

	// Wire calendar event creation into EngineSync so events appear in
	// the Pages graph view alongside everything else. Each integration
	// handler exposes an OnEventCreated callback we attach a closure to.
	if gh := integrationRouter.GetGoogleHandler(); gh != nil {
		gh.OnEventCreated = newGoogleCalendarEngineHook(h.engineSync)
	}
	if gth := integrationRouter.GetGoogleCalendarToolHandler(); gth != nil {
		gth.OnEventCreated = newGoogleCalendarEngineHook(h.engineSync)
	}
	if mh := integrationRouter.GetMicrosoftHandler(); mh != nil {
		mh.OnEventCreated = newMicrosoftCalendarEngineHook(h.engineSync)
	}

	// Register provider-based integration routes - /api/integrations/{provider}/*
	integrationsGroup := api.Group("/integrations")
	integrationRouter.RegisterRoutes(integrationsGroup, auth)

	// Integrations Module - /api/integrations (user integration management)
	integrationsHandler := NewIntegrationsHandler(h.pool, integrationRouter)
	integrationsModule := api.Group("/integrations")

	// Public endpoints - provider catalog (no auth required)
	integrationsModule.GET("/providers", integrationsHandler.GetProviders)
	integrationsModule.GET("/providers/:id", integrationsHandler.GetProvider)

	// Protected endpoints - user-specific data (auth required)
	integrationsProtected := integrationsModule.Group("")
	integrationsProtected.Use(auth, middleware.RequireAuth())
	{
		// Aggregated status (must be before :id to avoid matching)
		integrationsProtected.GET("/status", integrationsHandler.GetAllIntegrationsStatus)
		// User's connected integrations
		integrationsProtected.GET("/connected", integrationsHandler.GetConnectedIntegrations)
		integrationsProtected.GET("/:id", integrationsHandler.GetIntegration)
		integrationsProtected.PATCH("/:id/settings", integrationsHandler.UpdateIntegrationSettings)
		integrationsProtected.DELETE("/:id", integrationsHandler.DisconnectIntegration)
		integrationsProtected.POST("/:id/sync", integrationsHandler.TriggerSync)
		// AI Model preferences
		integrationsProtected.GET("/ai/preferences", integrationsHandler.GetModelPreferences)
		integrationsProtected.PUT("/ai/preferences", integrationsHandler.UpdateModelPreferences)
	}

	// MCP Server management - /api/integrations/mcp/connectors
	mcpServersHandler := NewMCPServersHandler(h.pool)
	mcpConnectors := integrationsModule.Group("/mcp/connectors")
	mcpConnectors.Use(auth, middleware.RequireAuth())
	{
		mcpConnectors.GET("", mcpServersHandler.ListMCPServers)
		mcpConnectors.POST("", mcpServersHandler.CreateMCPServer)
		mcpConnectors.GET("/:id", mcpServersHandler.GetMCPServer)
		mcpConnectors.PUT("/:id", mcpServersHandler.UpdateMCPServer)
		mcpConnectors.DELETE("/:id", mcpServersHandler.DeleteMCPServer)
		mcpConnectors.POST("/:id/test", mcpServersHandler.TestMCPServer)
		mcpConnectors.POST("/:id/discover", mcpServersHandler.DiscoverMCPTools)
	}
	slog.Info("MCP server routes registered at /api/integrations/mcp/connectors/*")

	// Module-specific integration endpoints - /api/modules/:id/integrations
	modules := api.Group("/modules")
	modules.Use(optionalAuth) // Optional auth for browsing available integrations
	{
		modules.GET("/:id/integrations", integrationsHandler.GetModuleIntegrations)
	}

	// Custom Modules routes - /api/modules (Feature: Custom Module System)
	customModulesHandler := NewCustomModulesHandler(h.pool, slog.Default())
	customModules := api.Group("/modules")
	customModules.Use(auth, middleware.RequireAuth())
	{
		// Module CRUD
		customModules.POST("", customModulesHandler.CreateModule)
		customModules.GET("", customModulesHandler.ListModules)
		customModules.GET("/:id", customModulesHandler.GetModule)
		customModules.PUT("/:id", customModulesHandler.UpdateModule)
		customModules.DELETE("/:id", customModulesHandler.DeleteModule)

		// Module operations
		customModules.POST("/:id/publish", customModulesHandler.PublishModule)
		customModules.POST("/:id/install", customModulesHandler.InstallModule)
		customModules.POST("/:id/share", customModulesHandler.ShareModule)

		// Module data
		customModules.GET("/installed", customModulesHandler.ListInstalledModules)
		customModules.GET("/stats", customModulesHandler.GetModuleStats)
		customModules.GET("/popular", customModulesHandler.GetPopularModules)

		// Module import/export
		customModules.GET("/export/:id", customModulesHandler.ExportModule)
		customModules.POST("/import", customModulesHandler.ImportModule)
	}
	slog.Info("Custom modules routes registered at /api/modules/*")

	// Onboarding routes - /api/onboarding
	// Conversational AI onboarding with workspace setup
	onboardingAIService := services.NewOnboardingAIService()
	gmailService := integrationRouter.GetGoogleGmailService()
	onboardingService := services.NewOnboardingService(h.pool, onboardingAIService, gmailService, h.osaSyncService)
	onboardingHandler := NewOnboardingHandler(onboardingService)
	onboardingHandler.RegisterOnboardingRoutes(api, auth)
	slog.Info("Onboarding routes registered", "ai_provider", onboardingAIService.GetProvider())
}
