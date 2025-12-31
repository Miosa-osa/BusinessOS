package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/terminal"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	pool                 *pgxpool.Pool
	cfg                  *config.Config
	containerMgr         *container.ContainerManager
	sessionCache         *middleware.SessionCache       // Redis session cache for horizontal scaling
	terminalPubSub       *terminal.TerminalPubSub       // Redis pub/sub for terminal scaling
	embeddingService     *services.EmbeddingService     // Vector embedding service for RAG
	contextBuilder       *services.ContextBuilder       // Hierarchical context builder for AI
	tieredContextService *services.TieredContextService // Tiered context builder for scoped AI queries
}

// NewHandlers creates a new Handlers instance
func NewHandlers(pool *pgxpool.Pool, cfg *config.Config, containerMgr *container.ContainerManager, sessionCache *middleware.SessionCache, terminalPubSub *terminal.TerminalPubSub, embeddingService *services.EmbeddingService, contextBuilder *services.ContextBuilder, tieredContextService *services.TieredContextService) *Handlers {
	return &Handlers{
		pool:                 pool,
		cfg:                  cfg,
		containerMgr:         containerMgr,
		sessionCache:         sessionCache,
		terminalPubSub:       terminalPubSub,
		embeddingService:     embeddingService,
		contextBuilder:       contextBuilder,
		tieredContextService: tieredContextService,
	}
}

// RegisterRoutes registers all API routes
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

	// Chat routes - /api/chat
	chat := api.Group("/chat")
	chat.Use(auth)
	{
		chat.GET("/conversations", h.ListConversations)
		chat.POST("/conversations", h.CreateConversation)
		chat.GET("/conversations/:id", h.GetConversation)
		chat.PUT("/conversations/:id", h.UpdateConversation)
		chat.DELETE("/conversations/:id", h.DeleteConversation)
		chat.POST("/message", h.SendMessageV2)    // V1 redirected to V2
		chat.POST("/v2/message", h.SendMessageV2) // V2 with COT and SSE events
		chat.GET("/search", h.SearchConversations)
		// AI-powered endpoints
		chat.POST("/ai/document", h.DocumentAI)
		chat.POST("/ai/analyze", h.AnalyzeContent)
		chat.POST("/ai/extract-tasks", h.ExtractTasks)
		chat.POST("/ai/plan", h.CreatePlan)
	}

	// Artifacts routes - /api/artifacts
	artifacts := api.Group("/artifacts")
	artifacts.Use(auth)
	{
		artifacts.GET("", h.ListArtifacts)
		artifacts.POST("", h.CreateArtifact)
		artifacts.GET("/:id", h.GetArtifact)
		artifacts.PATCH("/:id", h.UpdateArtifact)
		artifacts.PATCH("/:id/link", h.LinkArtifact)
		artifacts.DELETE("/:id", h.DeleteArtifact)
	}

	// Contexts routes - /api/contexts
	contexts := api.Group("/contexts")
	{
		// Public route (no auth)
		contexts.GET("/public/:shareId", h.GetPublicContext)

		// Protected routes
		protected := contexts.Group("")
		protected.Use(auth)
		{
			protected.GET("", h.ListContexts)
			protected.POST("", h.CreateContext)
			protected.GET("/:id", h.GetContext)
			protected.PUT("/:id", h.UpdateContext)
			protected.PATCH("/:id/blocks", h.UpdateContextBlocks)
			protected.POST("/:id/share", h.ShareContext)
			protected.DELETE("/:id/share", h.UnshareContext)
			protected.POST("/:id/duplicate", h.DuplicateContext)
			protected.PATCH("/:id/archive", h.ArchiveContext)
			protected.PATCH("/:id/unarchive", h.UnarchiveContext)
			protected.DELETE("/:id", h.DeleteContext)
			protected.POST("/aggregate", h.AggregateContext)
		}
	}

	// Projects routes - /api/projects
	projects := api.Group("/projects")
	projects.Use(auth)
	{
		projects.GET("", h.ListProjects)
		projects.POST("", h.CreateProject)
		projects.GET("/stats", h.GetProjectStats)
		projects.GET("/overdue", h.GetOverdueProjects)
		projects.GET("/upcoming", h.GetUpcomingProjects)
		projects.GET("/:id", h.GetProject)
		projects.PUT("/:id", h.UpdateProject)
		projects.DELETE("/:id", h.DeleteProject)
		projects.POST("/:id/notes", h.AddProjectNote)
		// Project members (team assignment)
		projects.GET("/:id/members", h.ListProjectMembers)
		projects.POST("/:id/members", h.AddProjectMember)
		projects.PUT("/:id/members/:memberId", h.UpdateProjectMemberRole)
		projects.DELETE("/:id/members/:memberId", h.RemoveProjectMember)
	}

	// Clients routes - /api/clients
	clients := api.Group("/clients")
	clients.Use(auth)
	{
		clients.GET("", h.ListClients)
		clients.POST("", h.CreateClient)
		clients.GET("/:id", h.GetClient)
		clients.PUT("/:id", h.UpdateClient)
		clients.PATCH("/:id/status", h.UpdateClientStatus)
		clients.DELETE("/:id", h.DeleteClient)
		// Contacts
		clients.GET("/:id/contacts", h.ListClientContacts)
		clients.POST("/:id/contacts", h.CreateClientContact)
		clients.PUT("/:id/contacts/:contactId", h.UpdateClientContact)
		clients.DELETE("/:id/contacts/:contactId", h.DeleteClientContact)
		// Interactions
		clients.GET("/:id/interactions", h.ListClientInteractions)
		clients.POST("/:id/interactions", h.CreateClientInteraction)
		// Deals
		clients.GET("/:id/deals", h.ListClientDeals)
		clients.POST("/:id/deals", h.CreateClientDeal)
		clients.PUT("/:id/deals/:dealId", h.UpdateClientDeal)
	}

	// Deals routes - /api/deals
	deals := api.Group("/deals")
	deals.Use(auth)
	{
		deals.GET("", h.ListDeals)
		deals.PATCH("/:id/stage", h.UpdateDealStage)
	}

	// Dashboard routes - /api/dashboard
	dashboard := api.Group("/dashboard")
	dashboard.Use(auth)
	{
		dashboard.GET("/summary", h.GetDashboardSummary)
		// Focus items
		dashboard.GET("/focus", h.ListFocusItems)
		dashboard.POST("/focus", h.CreateFocusItem)
		dashboard.PUT("/focus/:id", h.UpdateFocusItem)
		dashboard.DELETE("/focus/:id", h.DeleteFocusItem)
		// Tasks
		dashboard.GET("/tasks", h.ListTasks)
		dashboard.POST("/tasks", h.CreateTask)
		dashboard.PUT("/tasks/:id", h.UpdateTask)
		dashboard.POST("/tasks/:id/toggle", h.ToggleTask)
		dashboard.DELETE("/tasks/:id", h.DeleteTask)
	}

	// Team routes - /api/team
	team := api.Group("/team")
	team.Use(auth)
	{
		team.GET("", h.ListTeamMembers)
		team.POST("", h.CreateTeamMember)
		team.GET("/:id", h.GetTeamMember)
		team.PUT("/:id", h.UpdateTeamMember)
		team.PATCH("/:id/status", h.UpdateTeamMemberStatus)
		team.PATCH("/:id/capacity", h.UpdateTeamMemberCapacity)
		team.POST("/:id/activity", h.AddTeamMemberActivity)
		team.DELETE("/:id", h.DeleteTeamMember)
	}

	// Nodes routes - /api/nodes
	nodes := api.Group("/nodes")
	nodes.Use(auth)
	{
		nodes.GET("", h.ListNodes)
		nodes.GET("/tree", h.GetNodeTree)
		nodes.GET("/active", h.GetActiveNode)
		nodes.POST("", h.CreateNode)
		nodes.GET("/:id", h.GetNode)
		nodes.PATCH("/:id", h.UpdateNode)
		nodes.POST("/:id/activate", h.ActivateNode)
		nodes.POST("/:id/deactivate", h.DeactivateNode)
		nodes.DELETE("/:id", h.DeleteNode)
		nodes.GET("/:id/children", h.GetNodeChildren)
		nodes.POST("/:id/reorder", h.ReorderNodes)
	}

	// Daily logs routes - /api/daily
	daily := api.Group("/daily")
	daily.Use(auth)
	{
		daily.GET("/logs", h.ListDailyLogs)
		daily.GET("/logs/today", h.GetTodayLog)
		daily.GET("/logs/:date", h.GetDailyLogByDate)
		daily.POST("/logs", h.CreateOrUpdateDailyLog)
		daily.PUT("/logs/:id", h.UpdateDailyLog)
		daily.DELETE("/logs/:id", h.DeleteDailyLog)
	}

	// Settings routes - /api/settings
	settings := api.Group("/settings")
	settings.Use(auth)
	{
		settings.GET("", h.GetSettings)
		settings.PUT("", h.UpdateSettings)
		settings.GET("/system", h.GetSystemSettings)
		settings.GET("/full-state", h.GetFullState) // Complete state for UI sync
	}

	// Thinking/COT routes - /api/thinking
	thinking := api.Group("/thinking")
	thinking.Use(auth)
	{
		// Thinking traces
		thinking.GET("/traces/:conversationId", h.ListThinkingTraces)
		thinking.GET("/trace/:messageId", h.GetThinkingTraceByMessage)
		thinking.DELETE("/traces/:conversationId", h.DeleteThinkingTraces)
		// Thinking settings
		thinking.GET("/settings", h.GetThinkingSettings)
		thinking.PUT("/settings", h.UpdateThinkingSettings)
	}

	// Reasoning templates routes - /api/reasoning
	reasoning := api.Group("/reasoning")
	reasoning.Use(auth)
	{
		reasoning.GET("/templates", h.ListReasoningTemplates)
		reasoning.POST("/templates", h.CreateReasoningTemplate)
		reasoning.GET("/templates/:id", h.GetReasoningTemplate)
		reasoning.PUT("/templates/:id", h.UpdateReasoningTemplate)
		reasoning.DELETE("/templates/:id", h.DeleteReasoningTemplate)
		reasoning.POST("/templates/:id/default", h.SetDefaultReasoningTemplate)
	}

	// Focus Mode routes - /api/focus
	focus := api.Group("/focus")
	focus.Use(auth)
	{
		focus.GET("/templates", h.GetFocusModeTemplates)
		focus.GET("/settings", h.GetEffectiveFocusSettings)
		focus.POST("/preflight", h.BuildPreflightContext)
	}

	// Web Search routes - /api/search
	search := api.Group("/search")
	search.Use(auth)
	{
		search.GET("/web", h.WebSearch)                // Basic web search
		search.GET("/context", h.WebSearchWithContext) // Search with formatted context
		search.GET("/history", h.ListSearchHistory)    // List user's search history
		search.GET("/history/:id", h.GetSearchHistoryEntry) // Get specific search details
		search.DELETE("/history/:id", h.DeleteSearchHistoryEntry) // Delete specific search
		search.DELETE("/history", h.ClearSearchHistory) // Clear all search history
	}

	// AI configuration routes - /api/ai
	ai := api.Group("/ai")
	ai.Use(auth)
	{
		ai.GET("/providers", h.GetLLMProviders)
		ai.GET("/models", h.GetAllModels)
		ai.GET("/models/local", h.GetLocalModels)
		ai.POST("/models/pull", h.PullModel)
		ai.POST("/models/warmup", h.WarmupModel)
		ai.GET("/system", h.GetSystemInfo)
		ai.POST("/api-key", h.SaveAPIKey)
		ai.PUT("/provider", h.UpdateAIProvider)
		// Agent presets (templates for custom agents) - must be before /agents/:id
		ai.GET("/agents/presets", h.ListAgentPresets)
		ai.GET("/agents/presets/:id", h.GetAgentPreset)
		// Agent prompts (built-in)
		ai.GET("/agents", h.GetAgentPrompts)
		ai.GET("/agents/:id", h.GetAgentPrompt)
		// Custom user agents - specific routes before parameterized ones
		ai.GET("/custom-agents", h.ListCustomAgents)
		ai.POST("/custom-agents", h.CreateCustomAgent)
		ai.POST("/custom-agents/sandbox", h.TestCustomAgent)      // Test arbitrary prompt
		ai.GET("/custom-agents/category/:category", h.ListCustomAgentsByCategory)
		ai.POST("/custom-agents/from-preset/:presetId", h.CreateAgentFromPreset)
		ai.GET("/custom-agents/:id", h.GetCustomAgent)
		ai.PUT("/custom-agents/:id", h.UpdateCustomAgent)
		ai.DELETE("/custom-agents/:id", h.DeleteCustomAgent)
		ai.POST("/custom-agents/:id/test", h.TestCustomAgent)     // Test existing agent
		// Slash commands (built-in + custom)
		ai.GET("/commands", h.ListCommands)
		// Custom user commands CRUD
		ai.POST("/commands", h.CreateUserCommand)
		ai.GET("/commands/:id", h.GetUserCommand)
		ai.PUT("/commands/:id", h.UpdateUserCommand)
		ai.DELETE("/commands/:id", h.DeleteUserCommand)
		// Agent delegation routes
		delegationHandler := NewDelegationHandler(services.NewDelegationService(h.pool))
		ai.GET("/delegation/agents", delegationHandler.ListAgents)
		ai.GET("/delegation/resolve/:mention", delegationHandler.ResolveAgentMention)
		ai.POST("/delegation/mentions", delegationHandler.ExtractMentions)
		ai.POST("/delegation/delegate", delegationHandler.Delegate)
		// Intent classification / routing
		routerHandler := NewRouterHandler(h.pool)
		routerHandler.RegisterRoutes(ai)
		// Workflow routes
		ai.GET("/workflows", h.ListWorkflows)
		ai.POST("/workflows", h.CreateWorkflow)
		ai.GET("/workflows/:id", h.GetWorkflow)
		ai.DELETE("/workflows/:id", h.DeleteWorkflow)
		ai.POST("/workflows/:id/execute", h.ExecuteWorkflow)
		ai.POST("/workflows/trigger/:trigger", h.ExecuteWorkflowByTrigger)
		ai.GET("/workflows/executions", h.ListWorkflowExecutions)
		ai.GET("/workflows/executions/:id", h.GetWorkflowExecution)
	}

	// Usage analytics routes - /api/usage
	usage := api.Group("/usage")
	usage.Use(auth)
	{
		usage.GET("/summary", h.GetUsageSummary)
		usage.GET("/providers", h.GetUsageByProvider)
		usage.GET("/models", h.GetUsageByModel)
		usage.GET("/agents", h.GetUsageByAgent)
		usage.GET("/trend", h.GetUsageTrend)
		usage.GET("/recent", h.GetRecentAIUsage)
		usage.GET("/mcp", h.GetMCPUsage)
	}

	// Embeddings routes - /api/embeddings (for RAG and semantic search)
	if h.embeddingService != nil && h.contextBuilder != nil {
		embeddingHandler := NewEmbeddingHandler(h.embeddingService, h.contextBuilder)
		embeddings := api.Group("/embeddings")
		embeddings.Use(auth)
		{
			embeddings.POST("/index/:id", embeddingHandler.IndexDocument)
			embeddings.POST("/search", embeddingHandler.SemanticSearch)
			embeddings.POST("/context", embeddingHandler.BuildAIContext)
			embeddings.GET("/context/:id", embeddingHandler.GetDocumentContext)
			embeddings.GET("/stats", embeddingHandler.GetStats)
			embeddings.GET("/health", embeddingHandler.HealthCheck)
		}
	}

	// Transcription routes - /api/transcribe
	transcriptionHandler := NewTranscriptionHandler(h.pool)
	transcribe := api.Group("/transcribe")
	transcribe.Use(auth)
	{
		transcribe.POST("", transcriptionHandler.TranscribeAudio)
		transcribe.GET("/status", transcriptionHandler.GetTranscriptionStatus)
	}

	// Voice notes routes - /api/voice-notes
	voiceNotesHandler := NewVoiceNotesHandler(h.pool)
	voiceNotes := api.Group("/voice-notes")
	voiceNotes.Use(auth)
	{
		voiceNotes.GET("", voiceNotesHandler.ListVoiceNotes)
		voiceNotes.POST("", voiceNotesHandler.UploadVoiceNote)
		voiceNotes.GET("/stats", voiceNotesHandler.GetVoiceNoteStats)
		voiceNotes.GET("/:id", voiceNotesHandler.GetVoiceNote)
		voiceNotes.DELETE("/:id", voiceNotesHandler.DeleteVoiceNote)
		voiceNotes.POST("/:id/retranscribe", voiceNotesHandler.RetranscribeVoiceNote)
	}

	// Profile routes - /api/profile
	profile := api.Group("/profile")
	profile.Use(auth)
	{
		profile.PUT("", h.UpdateProfile)
		profile.POST("/photo", h.UploadProfilePhoto)
		profile.DELETE("/photo", h.DeleteProfilePhoto)
		profile.POST("/background", h.UploadBackground)
		profile.GET("/background/:filename", h.GetBackground)
		profile.DELETE("/background", h.DeleteBackground)
	}

	// MCP routes - /api/mcp
	mcp := api.Group("/mcp")
	mcp.Use(auth)
	{
		mcp.GET("/tools", h.ListMCPTools)
		mcp.POST("/execute", h.ExecuteMCPTool)
		mcp.GET("/health", h.MCPHealth)
	}

	// Initialize Google Calendar service and handlers
	calendarService := services.NewGoogleCalendarService(h.pool)
	googleOAuthHandler := NewGoogleOAuthHandler(calendarService)
	calendarHandler := NewCalendarHandler(h, calendarService)

	slackService := services.NewSlackService(h.pool)
	slackOAuthHandler := NewSlackOAuthHandler(slackService)

	// Initialize Notion service and handlers
	notionService := services.NewNotionService(h.pool)
	notionOAuthHandler := NewNotionOAuthHandler(notionService)

	// Integration routes - /api/integrations
	integrations := api.Group("/integrations")
	{
		// Google OAuth integration routes - /api/integrations/google
		google := integrations.Group("/google")
		google.Use(auth)
		{
			google.GET("/auth", googleOAuthHandler.InitiateGoogleAuth)
			google.GET("/status", googleOAuthHandler.GetGoogleConnectionStatus)
			google.DELETE("", googleOAuthHandler.DisconnectGoogle)
		}
		// Callback doesn't need auth (user redirected from Google)
		integrations.GET("/google/callback", googleOAuthHandler.HandleGoogleCallback)

		// Slack OAuth integration routes - /api/integrations/slack
		slackRoutes := integrations.Group("/slack")
		slackRoutes.Use(auth)
		{
			slackRoutes.GET("/auth", slackOAuthHandler.InitiateSlackAuth)
			slackRoutes.GET("/status", slackOAuthHandler.GetSlackConnectionStatus)
			slackRoutes.DELETE("", slackOAuthHandler.DisconnectSlack)
		}
		// Callback doesn't need auth (user redirected from Slack)
		integrations.GET("/slack/callback", slackOAuthHandler.HandleSlackCallback)

		// Notion OAuth integration routes - /api/integrations/notion
		notionRoutes := integrations.Group("/notion")
		notionRoutes.Use(auth)
		{
			notionRoutes.GET("/auth", notionOAuthHandler.InitiateNotionAuth)
			notionRoutes.GET("/status", notionOAuthHandler.GetNotionConnectionStatus)
			notionRoutes.GET("/databases", notionOAuthHandler.GetNotionDatabases)
			notionRoutes.GET("/pages", notionOAuthHandler.GetNotionPages)
			notionRoutes.GET("/search", notionOAuthHandler.SearchNotion)
			notionRoutes.DELETE("", notionOAuthHandler.DisconnectNotion)
		}
		// Callback doesn't need auth (user redirected from Notion)
		integrations.GET("/notion/callback", notionOAuthHandler.HandleNotionCallback)
	}

	// Calendar routes - /api/calendar
	calendar := api.Group("/calendar")
	calendar.Use(auth)
	{
		calendar.GET("/events", calendarHandler.ListEvents)
		calendar.GET("/events/:id", calendarHandler.GetEvent)
		calendar.POST("/events", calendarHandler.CreateEvent)
		calendar.PUT("/events/:id", calendarHandler.UpdateEvent)
		calendar.DELETE("/events/:id", calendarHandler.DeleteEvent)
		calendar.POST("/sync", calendarHandler.SyncCalendar)
		calendar.GET("/today", calendarHandler.GetTodayEvents)
		calendar.GET("/upcoming", calendarHandler.GetUpcomingEvents)
	}

	// Terminal routes - /api/terminal
	terminalHandler := NewTerminalHandler(h.containerMgr, h.terminalPubSub)
	terminalRoutes := api.Group("/terminal")
	terminalRoutes.Use(auth)
	{
		terminalRoutes.GET("/ws", terminalHandler.HandleWebSocket)
		terminalRoutes.GET("/sessions", terminalHandler.ListSessions)
		terminalRoutes.DELETE("/sessions/:id", terminalHandler.CloseSession)
	}

	// Filesystem routes - /api/filesystem (optional auth for dev)
	filesystem := api.Group("/filesystem")
	filesystem.Use(optionalAuth)
	{
		filesystem.GET("/list", h.ListDirectory)
		filesystem.GET("/read", h.ReadFile)
		filesystem.GET("/download", h.DownloadFile)
		filesystem.GET("/info", h.GetFileInfo)
		filesystem.GET("/quick-access", h.GetQuickAccessPaths)
		filesystem.POST("/mkdir", h.CreateDirectory)
		filesystem.POST("/upload", h.UploadFile)
		filesystem.DELETE("/delete", h.DeleteFileOrDir)
	}

	// Sync routes - /api/sync
	syncRoutes := api.Group("/sync")
	syncRoutes.Use(auth)
	{
		syncRoutes.GET("/status", h.GetSyncStatus)
		syncRoutes.GET("/full", h.FullSync)
		syncRoutes.GET("/:table", h.GetSyncChanges)
	}

	// Also add sync endpoints on individual tables for the sync engine
	// These return changes since a given timestamp
	api.GET("/contexts/sync", auth, h.createTableSyncHandler("contexts"))
	api.GET("/conversations/sync", auth, h.createTableSyncHandler("conversations"))
	api.GET("/projects/sync", auth, h.createTableSyncHandler("projects"))
	api.GET("/tasks/sync", auth, h.createTableSyncHandler("tasks"))
	api.GET("/nodes/sync", auth, h.createTableSyncHandler("nodes"))
	api.GET("/clients/sync", auth, h.createTableSyncHandler("clients"))
	api.GET("/calendar_events/sync", auth, h.createTableSyncHandler("calendar_events"))
	api.GET("/daily_logs/sync", auth, h.createTableSyncHandler("daily_logs"))
	api.GET("/team_members/sync", auth, h.createTableSyncHandler("team_members"))
	api.GET("/artifacts/sync", auth, h.createTableSyncHandler("artifacts"))
	api.GET("/focus_items/sync", auth, h.createTableSyncHandler("focus_items"))
	api.GET("/user_settings/sync", auth, h.createTableSyncHandler("user_settings"))

	// Authentication routes - /api/auth
	// Apply strict rate limiting to prevent brute force attacks
	strictRateLimit := middleware.StrictRateLimitMiddleware()

	googleAuthHandler := NewGoogleAuthHandler(h.pool, h.cfg, h.sessionCache)
	emailAuthHandler := NewEmailAuthHandler(h.pool, h.cfg)
	authRoutes := api.Group("/auth")
	{
		// Email/Password auth (public) - strict rate limiting
		authRoutes.POST("/sign-up/email", strictRateLimit, emailAuthHandler.SignUp)
		authRoutes.POST("/sign-in/email", strictRateLimit, emailAuthHandler.SignIn)

		// Google OAuth (public)
		authRoutes.GET("/google", googleAuthHandler.InitiateGoogleLogin)
		authRoutes.GET("/google/callback/login", googleAuthHandler.HandleGoogleLoginCallback)

		// Session management (public)
		authRoutes.GET("/session", googleAuthHandler.GetCurrentSession)
		authRoutes.GET("/get-session", googleAuthHandler.GetCurrentSession) // Alias for better-auth compatibility
		authRoutes.POST("/logout", googleAuthHandler.Logout)
		authRoutes.POST("/sign-out", googleAuthHandler.Logout) // Alias for better-auth compatibility

		// Protected: Force logout all sessions (requires authentication)
		protectedAuth := authRoutes.Group("")
		protectedAuth.Use(auth)
		{
			protectedAuth.POST("/logout-all", googleAuthHandler.LogoutAllSessions)
		}
	}
}
