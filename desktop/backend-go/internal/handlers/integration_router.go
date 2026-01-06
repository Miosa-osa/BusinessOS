package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/integrations/airtable"
	"github.com/rhl/businessos-backend/internal/integrations/clickup"
	"github.com/rhl/businessos-backend/internal/integrations/fathom"
	"github.com/rhl/businessos-backend/internal/integrations/google"
	"github.com/rhl/businessos-backend/internal/integrations/hubspot"
	"github.com/rhl/businessos-backend/internal/integrations/linear"
	"github.com/rhl/businessos-backend/internal/integrations/microsoft"
	"github.com/rhl/businessos-backend/internal/integrations/notion"
	"github.com/rhl/businessos-backend/internal/integrations/slack"
)

// IntegrationRouter manages all integration providers and their HTTP handlers.
type IntegrationRouter struct {
	pool *pgxpool.Pool

	// Legacy Providers (unified)
	googleProvider    *google.Provider
	microsoftProvider *microsoft.Provider
	slackProvider     *slack.Provider
	notionProvider    *notion.Provider
	linearProvider    *linear.Provider
	hubspotProvider   *hubspot.Provider
	fathomProvider    *fathom.Provider
	clickupProvider   *clickup.Provider
	airtableProvider  *airtable.Provider

	// Legacy Handlers (unified)
	googleHandler    *google.Handler
	microsoftHandler *microsoft.Handler
	slackHandler     *slack.Handler
	notionHandler    *notion.Handler
	linearHandler    *linear.Handler
	hubspotHandler   *hubspot.Handler
	fathomHandler    *fathom.Handler
	clickupHandler   *clickup.Handler
	airtableHandler  *airtable.Handler

	// NEW: Tool-specific handlers for Google (separate scopes per tool)
	googleCalendarHandler *google.ToolHandler
	googleGmailHandler    *google.ToolHandler
}

// NewIntegrationRouter creates a new integration router with all providers initialized.
func NewIntegrationRouter(pool *pgxpool.Pool) *IntegrationRouter {
	// Initialize legacy providers (unified)
	// Only request basic scopes (calendar + gmail) - other APIs must be enabled in Google Cloud Console first
	googleProvider := google.NewProvider(pool, []string{"calendar", "gmail"})
	// Only request basic scopes for Microsoft
	microsoftProvider := microsoft.NewProvider(pool, []string{"calendar", "mail"})
	slackProvider := slack.NewProvider(pool)
	notionProvider := notion.NewProvider(pool)
	linearProvider := linear.NewProvider(pool)
	hubspotProvider := hubspot.NewProvider(pool)
	fathomProvider := fathom.NewProvider(pool)
	clickupProvider := clickup.NewProvider(pool)
	airtableProvider := airtable.NewProvider(pool)

	// Initialize legacy handlers (unified)
	googleHandler := google.NewHandler(googleProvider)
	microsoftHandler := microsoft.NewHandler(microsoftProvider)
	slackHandler := slack.NewHandler(slackProvider)
	notionHandler := notion.NewHandler(notionProvider)
	linearHandler := linear.NewHandler(linearProvider)
	hubspotHandler := hubspot.NewHandler(hubspotProvider)
	fathomHandler := fathom.NewHandler(fathomProvider)
	clickupHandler := clickup.NewHandler(clickupProvider)
	airtableHandler := airtable.NewHandler(airtableProvider)

	// NEW: Initialize tool-specific handlers for Google
	// Each tool has its own OAuth flow with specific scopes
	googleCalendarHandler, err := google.NewToolHandler(pool, "google_calendar")
	if err != nil {
		log.Printf("WARNING: Failed to create google_calendar handler: %v", err)
	}
	googleGmailHandler, err := google.NewToolHandler(pool, "google_gmail")
	if err != nil {
		log.Printf("WARNING: Failed to create google_gmail handler: %v", err)
	}

	return &IntegrationRouter{
		pool:              pool,
		googleProvider:    googleProvider,
		microsoftProvider: microsoftProvider,
		slackProvider:     slackProvider,
		notionProvider:    notionProvider,
		linearProvider:    linearProvider,
		hubspotProvider:   hubspotProvider,
		fathomProvider:    fathomProvider,
		clickupProvider:   clickupProvider,
		airtableProvider:  airtableProvider,
		googleHandler:     googleHandler,
		microsoftHandler:  microsoftHandler,
		slackHandler:      slackHandler,
		notionHandler:     notionHandler,
		linearHandler:     linearHandler,
		hubspotHandler:    hubspotHandler,
		fathomHandler:     fathomHandler,
		clickupHandler:    clickupHandler,
		airtableHandler:   airtableHandler,
		// NEW: Tool-specific handlers
		googleCalendarHandler: googleCalendarHandler,
		googleGmailHandler:    googleGmailHandler,
	}
}

// RegisterRoutes registers all integration routes under the given router group.
// Expected to be called with api.Group("/integrations")
func (r *IntegrationRouter) RegisterRoutes(integrationsGroup *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// ============================================================================
	// NEW: Tool-specific routes for Google (separate OAuth per tool)
	// ============================================================================

	// Google Calendar tool - /api/integrations/google_calendar/*
	// Only requests calendar scopes
	if r.googleCalendarHandler != nil {
		calGroup := integrationsGroup.Group("/google_calendar")
		{
			protected := calGroup.Group("")
			protected.Use(authMiddleware)
			{
				protected.GET("/auth", r.googleCalendarHandler.GetAuthURL)
				protected.GET("/status", r.googleCalendarHandler.GetStatus)
				protected.POST("/disconnect", r.googleCalendarHandler.Disconnect)

				// Calendar routes
				protected.GET("/events", r.googleCalendarHandler.GetCalendarEvents)
				protected.POST("/events", r.googleCalendarHandler.CreateCalendarEvent)
				protected.DELETE("/events/:id", r.googleCalendarHandler.DeleteCalendarEvent)
				protected.POST("/sync", r.googleCalendarHandler.SyncCalendar)
			}
			calGroup.GET("/callback", r.googleCalendarHandler.HandleCallback)
		}
	}

	// Gmail tool - /api/integrations/google_gmail/*
	// Only requests gmail scopes
	if r.googleGmailHandler != nil {
		gmailGroup := integrationsGroup.Group("/google_gmail")
		{
			protected := gmailGroup.Group("")
			protected.Use(authMiddleware)
			{
				protected.GET("/auth", r.googleGmailHandler.GetAuthURL)
				protected.GET("/status", r.googleGmailHandler.GetStatus)
				protected.POST("/disconnect", r.googleGmailHandler.Disconnect)

				// Gmail routes
				protected.GET("/emails", r.googleGmailHandler.GetEmails)
				protected.GET("/emails/:id", r.googleGmailHandler.GetEmail)
				protected.POST("/send", r.googleGmailHandler.SendEmail)
				protected.POST("/sync", r.googleGmailHandler.SyncGmail)
			}
			gmailGroup.GET("/callback", r.googleGmailHandler.HandleCallback)
		}
	}

	// ============================================================================
	// LEGACY: Unified Google routes (for backward compatibility)
	// ============================================================================

	// Google integration routes - /api/integrations/google/*
	googleGroup := integrationsGroup.Group("/google")
	{
		// Protected routes (require auth)
		protected := googleGroup.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/auth", r.googleHandler.GetAuthURL)
			protected.GET("/status", r.googleHandler.GetStatus)
			protected.POST("/disconnect", r.googleHandler.Disconnect)

			// Calendar routes
			protected.GET("/calendar/events", r.googleHandler.GetCalendarEvents)
			protected.POST("/calendar/events", r.googleHandler.CreateCalendarEvent)
			protected.DELETE("/calendar/events/:id", r.googleHandler.DeleteCalendarEvent)
			protected.POST("/calendar/sync", r.googleHandler.SyncCalendar)

			// Gmail routes
			protected.GET("/gmail/emails", r.googleHandler.GetEmails)
			protected.GET("/gmail/emails/:id", r.googleHandler.GetEmail)
			protected.POST("/gmail/emails/send", r.googleHandler.SendEmail)
			protected.POST("/gmail/emails/:id/read", r.googleHandler.MarkEmailRead)
			protected.POST("/gmail/emails/:id/archive", r.googleHandler.ArchiveEmail)
			protected.DELETE("/gmail/emails/:id", r.googleHandler.DeleteEmail)
			protected.POST("/gmail/sync", r.googleHandler.SyncGmail)
		}

		// Callback doesn't need auth (user redirected from Google)
		googleGroup.GET("/callback", r.googleHandler.HandleCallback)
	}

	// Slack integration routes - /api/integrations/slack/*
	slackGroup := integrationsGroup.Group("/slack")
	{
		// Protected routes (require auth)
		protected := slackGroup.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/auth", r.slackHandler.GetAuthURL)
			protected.GET("/status", r.slackHandler.GetStatus)
			protected.POST("/disconnect", r.slackHandler.Disconnect)

			// Channel routes
			protected.GET("/channels", r.slackHandler.GetChannels)
			protected.POST("/channels/sync", r.slackHandler.SyncChannels)

			// Message routes
			protected.GET("/messages/:channel_id", r.slackHandler.GetMessages)
			protected.POST("/messages/:channel_id", r.slackHandler.SendMessage)
			protected.POST("/messages/:channel_id/sync", r.slackHandler.SyncMessages)
		}

		// Callback doesn't need auth (user redirected from Slack)
		slackGroup.GET("/callback", r.slackHandler.HandleCallback)
	}

	// Notion integration routes - /api/integrations/notion/*
	notionGroup := integrationsGroup.Group("/notion")
	{
		// Protected routes (require auth)
		protected := notionGroup.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/auth", r.notionHandler.GetAuthURL)
			protected.GET("/status", r.notionHandler.GetStatus)
			protected.POST("/disconnect", r.notionHandler.Disconnect)

			// Database routes
			protected.GET("/databases", r.notionHandler.GetDatabases)
			protected.POST("/databases/sync", r.notionHandler.SyncDatabases)

			// Page routes
			protected.GET("/pages/:database_id", r.notionHandler.GetPages)
			protected.POST("/pages/:database_id/sync", r.notionHandler.SyncPages)
		}

		// Callback doesn't need auth (user redirected from Notion)
		notionGroup.GET("/callback", r.notionHandler.HandleCallback)
	}

	// Microsoft 365 integration routes - /api/integrations/microsoft/*
	microsoftGroup := integrationsGroup.Group("/microsoft")
	{
		protected := microsoftGroup.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/auth", r.microsoftHandler.GetAuthURL)
			protected.GET("/status", r.microsoftHandler.GetStatus)
			protected.POST("/disconnect", r.microsoftHandler.Disconnect)

			// Calendar routes
			protected.GET("/calendar/events", r.microsoftHandler.GetCalendarEvents)
			protected.POST("/calendar/events", r.microsoftHandler.CreateCalendarEvent)
			protected.POST("/calendar/sync", r.microsoftHandler.SyncCalendar)

			// Mail routes
			protected.GET("/mail/emails", r.microsoftHandler.GetEmails)
			protected.GET("/mail/emails/:id", r.microsoftHandler.GetEmail)
			protected.POST("/mail/send", r.microsoftHandler.SendEmail)
			protected.POST("/mail/sync", r.microsoftHandler.SyncMail)

			// Files routes (OneDrive)
			protected.GET("/files", r.microsoftHandler.GetFiles)
			protected.GET("/files/:id", r.microsoftHandler.GetFile)
			protected.POST("/files/sync", r.microsoftHandler.SyncFiles)

			// Tasks routes (To Do)
			protected.GET("/tasks/lists", r.microsoftHandler.GetTaskLists)
			protected.GET("/tasks/:list_id/tasks", r.microsoftHandler.GetTasks)
			protected.POST("/tasks/:list_id/tasks", r.microsoftHandler.CreateTask)
			protected.POST("/tasks/:list_id/tasks/:task_id/complete", r.microsoftHandler.CompleteTask)
			protected.POST("/tasks/sync", r.microsoftHandler.SyncTasks)
		}

		microsoftGroup.GET("/callback", r.microsoftHandler.HandleCallback)
	}

	// Linear integration routes - /api/integrations/linear/*
	linearGroup := integrationsGroup.Group("/linear")
	{
		protected := linearGroup.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/auth", r.linearHandler.GetAuthURL)
			protected.GET("/status", r.linearHandler.GetStatus)
			protected.POST("/disconnect", r.linearHandler.Disconnect)

			// Issue routes
			protected.GET("/issues", r.linearHandler.GetIssues)
			protected.GET("/issues/:id", r.linearHandler.GetIssue)
			protected.POST("/issues", r.linearHandler.CreateIssue)
			protected.PUT("/issues/:id", r.linearHandler.UpdateIssue)
			protected.POST("/issues/sync", r.linearHandler.SyncIssues)

			// Project routes
			protected.GET("/projects", r.linearHandler.GetProjects)
			protected.GET("/projects/:id", r.linearHandler.GetProject)
			protected.POST("/projects/sync", r.linearHandler.SyncProjects)

			// Team routes
			protected.GET("/teams", r.linearHandler.GetTeams)
			protected.POST("/teams/sync", r.linearHandler.SyncTeams)
		}

		linearGroup.GET("/callback", r.linearHandler.HandleCallback)
	}

	// HubSpot integration routes - /api/integrations/hubspot/*
	hubspotGroup := integrationsGroup.Group("/hubspot")
	{
		protected := hubspotGroup.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/auth", r.hubspotHandler.GetAuthURL)
			protected.GET("/status", r.hubspotHandler.GetStatus)
			protected.POST("/disconnect", r.hubspotHandler.Disconnect)

			// Contact routes
			protected.GET("/contacts", r.hubspotHandler.GetContacts)
			protected.GET("/contacts/:id", r.hubspotHandler.GetContact)
			protected.POST("/contacts", r.hubspotHandler.CreateContact)
			protected.PUT("/contacts/:id", r.hubspotHandler.UpdateContact)
			protected.POST("/contacts/sync", r.hubspotHandler.SyncContacts)

			// Company routes
			protected.GET("/companies", r.hubspotHandler.GetCompanies)
			protected.GET("/companies/:id", r.hubspotHandler.GetCompany)
			protected.POST("/companies", r.hubspotHandler.CreateCompany)
			protected.POST("/companies/sync", r.hubspotHandler.SyncCompanies)

			// Deal routes
			protected.GET("/deals", r.hubspotHandler.GetDeals)
			protected.GET("/deals/:id", r.hubspotHandler.GetDeal)
			protected.POST("/deals", r.hubspotHandler.CreateDeal)
			protected.POST("/deals/sync", r.hubspotHandler.SyncDeals)
		}

		hubspotGroup.GET("/callback", r.hubspotHandler.HandleCallback)
	}

	// Fathom Analytics integration routes - /api/integrations/fathom/*
	// Note: Fathom uses API key auth, not OAuth
	fathomGroup := integrationsGroup.Group("/fathom")
	{
		protected := fathomGroup.Group("")
		protected.Use(authMiddleware)
		{
			// API key connection (not OAuth)
			protected.POST("/connect", r.fathomHandler.Connect)
			protected.POST("/disconnect", r.fathomHandler.Disconnect)
			protected.GET("/status", r.fathomHandler.GetStatus)

			// Site routes
			protected.GET("/sites", r.fathomHandler.GetSites)
			protected.GET("/sites/:site_id", r.fathomHandler.GetSite)
			protected.POST("/sites/sync", r.fathomHandler.SyncSites)

			// Analytics routes
			protected.GET("/sites/:site_id/aggregations", r.fathomHandler.GetAggregations)
			protected.GET("/sites/:site_id/current-visitors", r.fathomHandler.GetCurrentVisitors)
			protected.GET("/sites/:site_id/events", r.fathomHandler.GetEvents)
		}
		// Note: No callback for Fathom (API key auth)
	}

	// ClickUp integration routes - /api/integrations/clickup/*
	clickupGroup := integrationsGroup.Group("/clickup")
	{
		protected := clickupGroup.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/auth", r.clickupHandler.GetAuthURL)
			protected.GET("/status", r.clickupHandler.GetStatus)
			protected.POST("/disconnect", r.clickupHandler.Disconnect)

			// Workspace routes
			protected.GET("/workspaces", r.clickupHandler.GetWorkspaces)

			// Space routes
			protected.GET("/workspaces/:workspace_id/spaces", r.clickupHandler.GetSpaces)

			// List routes
			protected.GET("/spaces/:space_id/lists", r.clickupHandler.GetListsFromSpace)
			protected.GET("/folders/:folder_id/lists", r.clickupHandler.GetListsFromFolder)

			// Task routes
			protected.GET("/lists/:list_id/tasks", r.clickupHandler.GetTasks)
			protected.GET("/tasks/:task_id", r.clickupHandler.GetTask)
			protected.POST("/lists/:list_id/tasks", r.clickupHandler.CreateTask)
			protected.PUT("/tasks/:task_id", r.clickupHandler.UpdateTask)
			protected.POST("/tasks/sync", r.clickupHandler.SyncTasks)
		}

		clickupGroup.GET("/callback", r.clickupHandler.HandleCallback)
	}

	// Airtable integration routes - /api/integrations/airtable/*
	airtableGroup := integrationsGroup.Group("/airtable")
	{
		protected := airtableGroup.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/auth", r.airtableHandler.GetAuthURL)
			protected.GET("/status", r.airtableHandler.GetStatus)
			protected.POST("/disconnect", r.airtableHandler.Disconnect)

			// Base routes
			protected.GET("/bases", r.airtableHandler.GetBases)
			protected.POST("/bases/sync", r.airtableHandler.SyncBases)

			// Table routes
			protected.GET("/bases/:base_id/tables", r.airtableHandler.GetTables)
			protected.GET("/bases/:base_id/tables/:table_id", r.airtableHandler.GetTable)

			// Record routes
			protected.GET("/records", r.airtableHandler.GetRecords)
			protected.GET("/records/:record_id", r.airtableHandler.GetRecord)
			protected.POST("/records", r.airtableHandler.CreateRecord)
			protected.PUT("/records/:record_id", r.airtableHandler.UpdateRecord)
			protected.DELETE("/records/:record_id", r.airtableHandler.DeleteRecord)
			protected.POST("/records/sync", r.airtableHandler.SyncRecords)
		}

		airtableGroup.GET("/callback", r.airtableHandler.HandleCallback)
	}
}

// GetGoogleProvider returns the Google provider for use by other handlers.
func (r *IntegrationRouter) GetGoogleProvider() *google.Provider {
	return r.googleProvider
}

// GetSlackProvider returns the Slack provider for use by other handlers.
func (r *IntegrationRouter) GetSlackProvider() *slack.Provider {
	return r.slackProvider
}

// GetNotionProvider returns the Notion provider for use by other handlers.
func (r *IntegrationRouter) GetNotionProvider() *notion.Provider {
	return r.notionProvider
}

// GetGoogleCalendarService returns the Google Calendar service.
func (r *IntegrationRouter) GetGoogleCalendarService() *google.CalendarService {
	return google.NewCalendarService(r.googleProvider)
}

// GetGoogleGmailService returns the Google Gmail service.
func (r *IntegrationRouter) GetGoogleGmailService() *google.GmailService {
	return google.NewGmailService(r.googleProvider)
}

// GetSlackChannelService returns the Slack channel service.
func (r *IntegrationRouter) GetSlackChannelService() *slack.ChannelService {
	return slack.NewChannelService(r.slackProvider)
}

// GetSlackMessageService returns the Slack message service.
func (r *IntegrationRouter) GetSlackMessageService() *slack.MessageService {
	return slack.NewMessageService(r.slackProvider)
}

// GetNotionDatabaseService returns the Notion database service.
func (r *IntegrationRouter) GetNotionDatabaseService() *notion.DatabaseService {
	return notion.NewDatabaseService(r.notionProvider)
}

// GetMicrosoftProvider returns the Microsoft provider for use by other handlers.
func (r *IntegrationRouter) GetMicrosoftProvider() *microsoft.Provider {
	return r.microsoftProvider
}

// GetLinearProvider returns the Linear provider for use by other handlers.
func (r *IntegrationRouter) GetLinearProvider() *linear.Provider {
	return r.linearProvider
}

// GetHubSpotProvider returns the HubSpot provider for use by other handlers.
func (r *IntegrationRouter) GetHubSpotProvider() *hubspot.Provider {
	return r.hubspotProvider
}

// GetFathomProvider returns the Fathom provider for use by other handlers.
func (r *IntegrationRouter) GetFathomProvider() *fathom.Provider {
	return r.fathomProvider
}

// GetClickUpProvider returns the ClickUp provider for use by other handlers.
func (r *IntegrationRouter) GetClickUpProvider() *clickup.Provider {
	return r.clickupProvider
}

// GetAirtableProvider returns the Airtable provider for use by other handlers.
func (r *IntegrationRouter) GetAirtableProvider() *airtable.Provider {
	return r.airtableProvider
}
