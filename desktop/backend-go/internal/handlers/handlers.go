package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	pool         *pgxpool.Pool
	cfg          *config.Config
	containerMgr *container.ContainerManager
}

// NewHandlers creates a new Handlers instance
func NewHandlers(pool *pgxpool.Pool, cfg *config.Config, containerMgr *container.ContainerManager) *Handlers {
	return &Handlers{
		pool:         pool,
		cfg:          cfg,
		containerMgr: containerMgr,
	}
}

// RegisterRoutes registers all API routes
func (h *Handlers) RegisterRoutes(api *gin.RouterGroup) {
	// Auth middleware for protected routes
	auth := middleware.AuthMiddleware(h.pool)
	_ = middleware.OptionalAuthMiddleware(h.pool) // Available for routes that need optional auth

	// Chat routes - /api/chat
	chat := api.Group("/chat")
	chat.Use(auth)
	{
		chat.GET("/conversations", h.ListConversations)
		chat.POST("/conversations", h.CreateConversation)
		chat.GET("/conversations/:id", h.GetConversation)
		chat.PUT("/conversations/:id", h.UpdateConversation)
		chat.DELETE("/conversations/:id", h.DeleteConversation)
		chat.POST("/message", h.SendMessage)
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
		projects.GET("/:id", h.GetProject)
		projects.PUT("/:id", h.UpdateProject)
		projects.DELETE("/:id", h.DeleteProject)
		projects.POST("/:id/notes", h.AddProjectNote)
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
		// Agent prompts
		ai.GET("/agents", h.GetAgentPrompts)
		ai.GET("/agents/:id", h.GetAgentPrompt)
		// Slash commands (built-in + custom)
		ai.GET("/commands", h.ListCommands)
		// Custom user commands CRUD
		ai.POST("/commands", h.CreateUserCommand)
		ai.GET("/commands/:id", h.GetUserCommand)
		ai.PUT("/commands/:id", h.UpdateUserCommand)
		ai.DELETE("/commands/:id", h.DeleteUserCommand)
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

	// Google OAuth integration routes - /api/integrations/google
	integrations := api.Group("/integrations")
	{
		google := integrations.Group("/google")
		google.Use(auth)
		{
			google.GET("/auth", googleOAuthHandler.InitiateGoogleAuth)
			google.GET("/status", googleOAuthHandler.GetGoogleConnectionStatus)
			google.DELETE("", googleOAuthHandler.DisconnectGoogle)
		}
		// Callback doesn't need auth (user redirected from Google)
		integrations.GET("/google/callback", googleOAuthHandler.HandleGoogleCallback)
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
	terminalHandler := NewTerminalHandler(h.containerMgr)
	terminalRoutes := api.Group("/terminal")
	terminalRoutes.Use(auth)
	{
		terminalRoutes.GET("/ws", terminalHandler.HandleWebSocket)
		terminalRoutes.GET("/sessions", terminalHandler.ListSessions)
		terminalRoutes.DELETE("/sessions/:id", terminalHandler.CloseSession)
	}

	// Filesystem routes - /api/filesystem
	filesystem := api.Group("/filesystem")
	filesystem.Use(auth)
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

	// Authentication routes - /api/auth
	googleAuthHandler := NewGoogleAuthHandler(h.pool, h.cfg)
	emailAuthHandler := NewEmailAuthHandler(h.pool, h.cfg)
	authRoutes := api.Group("/auth")
	{
		// Email/Password auth (public)
		authRoutes.POST("/sign-up/email", emailAuthHandler.SignUp)
		authRoutes.POST("/sign-in/email", emailAuthHandler.SignIn)

		// Google OAuth (public)
		authRoutes.GET("/google", googleAuthHandler.InitiateGoogleLogin)
		authRoutes.GET("/google/callback/login", googleAuthHandler.HandleGoogleLoginCallback)

		// Session management (public)
		authRoutes.GET("/session", googleAuthHandler.GetCurrentSession)
		authRoutes.GET("/get-session", googleAuthHandler.GetCurrentSession) // Alias for better-auth compatibility
		authRoutes.POST("/logout", googleAuthHandler.Logout)
		authRoutes.POST("/sign-out", googleAuthHandler.Logout) // Alias for better-auth compatibility
	}
}
