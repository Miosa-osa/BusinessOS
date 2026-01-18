package handlers

import (
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
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
	notificationService  *services.NotificationService  // Notification service for real-time alerts
	notificationTriggers *services.NotificationTriggers // Notification triggers for events
	webPushService       *services.WebPushService       // Web Push notification service
	emailService         *services.EmailService         // Email service via Resend
	commentService       *services.CommentService       // Comment service with mentions
	// Pedro tasks services
	documentProcessor        *services.DocumentProcessor               // Document processing with chunking
	learningService          *services.LearningService                 // Learning and personalization
	autoLearningTriggers     *services.AutoLearningTriggers            // Automatic learning from conversations
	promptPersonalizer       *services.PromptPersonalizer              // Prompt personalization with user data
	appProfilerService       *services.AppProfilerService              // Application profiling
	conversationIntelligence *services.ConversationIntelligenceService // Conversation analysis
	memoryExtractor          *services.MemoryExtractorService          // Memory extraction
	blockMapper              *services.BlockMapperService              // Markdown to structured blocks
	// Day 2 RAG services
	hybridSearchService *services.HybridSearchService // Hybrid search (semantic + keyword)
	rerankerService     *services.ReRankerService     // Re-ranking with multi-signal scoring
	agenticRAGService   *services.AgenticRAGService   // Intelligent adaptive retrieval
	memoryService       *services.MemoryService       // Memory persistence
	// Feature 7: Multi-modal Search services
	multiModalHandler *MultiModalSearchHandler // Multi-modal search handler (text + image)
	// Feature 1: Workspace & Team Collaboration
	workspaceService       *services.WorkspaceService       // Workspace management
	roleContextService     *services.RoleContextService     // Role-based access control
	memoryHierarchyService *services.MemoryHierarchyService // Workspace memory hierarchy (Q1)
	inviteService          *services.WorkspaceInviteService // Workspace invitation management
	auditService           *services.WorkspaceAuditService  // Workspace audit logging
	projectAccessService   *services.ProjectAccessService   // Project-level access control
	// Voice services (3D Desktop)
	// NOTE: Voice processing now uses Python LiveKit agents (see python-voice-agent/)
	whisperService    *services.WhisperService    // Local speech-to-text (for voice notes)
	elevenLabsService *services.ElevenLabsService // Text-to-speech (OSA voice, for non-voice-agent TTS)
	// LiveKit token generation is in livekit.go handler (no service needed)
	// Agent Skills System
	skillsLoader *services.SkillsLoader // Skills loader for agent prompts
	// OSA Integration - AI Agent Orchestration
	osaClient            *osa.ResilientClient              // OSA resilient client with circuit breaker
	osaSyncService       *services.OSASyncService          // OSA sync service for bidirectional sync
	osaFileSyncService   *services.OSAFileSyncService      // OSA file polling service
	osaWorkspaceInit     *services.OSAWorkspaceInitService // OSA workspace initialization service
	osaWorkflowsHandler  *OSAWorkflowsHandler              // OSA workflows handler
	osaWebhooksHandler   *OSAWebhooksHandler               // OSA webhooks handler
	osaBuildEventBus     *services.BuildEventBus           // OSA build event bus for real-time streaming
	osaStreamingHandler  *OSAStreamingHandler              // OSA SSE streaming handler
	osaOnboardingHandler *OSAOnboardingHandler             // OSA onboarding handler (Build Your OS flow)
}

// NewHandlers creates a new Handlers instance
func NewHandlers(pool *pgxpool.Pool, cfg *config.Config, containerMgr *container.ContainerManager, sessionCache *middleware.SessionCache, terminalPubSub *terminal.TerminalPubSub, embeddingService *services.EmbeddingService, contextBuilder *services.ContextBuilder, tieredContextService *services.TieredContextService, notificationService *services.NotificationService, osaClient *osa.ResilientClient, osaSyncService *services.OSASyncService) *Handlers {
	var notifTriggers *services.NotificationTriggers
	if notificationService != nil {
		notifTriggers = services.NewNotificationTriggers(notificationService)
	}

	return &Handlers{
		pool:                 pool,
		cfg:                  cfg,
		containerMgr:         containerMgr,
		sessionCache:         sessionCache,
		terminalPubSub:       terminalPubSub,
		embeddingService:     embeddingService,
		contextBuilder:       contextBuilder,
		tieredContextService: tieredContextService,
		notificationService:  notificationService,
		notificationTriggers: notifTriggers,
		osaClient:            osaClient,
		osaSyncService:       osaSyncService,
	}
}

// SetWebPushService sets the Web Push service (optional)
func (h *Handlers) SetWebPushService(svc *services.WebPushService) {
	h.webPushService = svc
	// Also set on dispatcher if available
	if h.notificationService != nil && svc != nil {
		h.notificationService.Dispatcher().SetWebPushService(svc)
	}
}

// SetEmailService sets the Email service (optional)
func (h *Handlers) SetEmailService(svc *services.EmailService) {
	h.emailService = svc
	// Also set on dispatcher if available
	if h.notificationService != nil && svc != nil {
		h.notificationService.Dispatcher().SetEmailService(svc)
	}
}

// SetCommentService sets the Comment service (optional)
func (h *Handlers) SetCommentService(svc *services.CommentService) {
	h.commentService = svc
}

// SetVoiceServices sets the voice services (3D Desktop - Whisper + ElevenLabs)
func (h *Handlers) SetVoiceServices(whisper *services.WhisperService, elevenLabs *services.ElevenLabsService) {
	h.whisperService = whisper
	h.elevenLabsService = elevenLabs
}

// SetLiveKitAgent removed - token generation now handled directly in livekit.go handler
// No separate service needed for Python voice agents

// SetPedroServices sets the Pedro task services (optional, to avoid breaking existing code)
func (h *Handlers) SetPedroServices(
	documentProcessor *services.DocumentProcessor,
	learningService *services.LearningService,
	autoLearningTriggers *services.AutoLearningTriggers,
	promptPersonalizer *services.PromptPersonalizer,
	appProfilerService *services.AppProfilerService,
	conversationIntelligence *services.ConversationIntelligenceService,
	memoryExtractor *services.MemoryExtractorService,
	blockMapper *services.BlockMapperService,
) {
	h.documentProcessor = documentProcessor
	h.learningService = learningService
	h.autoLearningTriggers = autoLearningTriggers
	h.promptPersonalizer = promptPersonalizer
	h.appProfilerService = appProfilerService
	h.conversationIntelligence = conversationIntelligence
	h.memoryExtractor = memoryExtractor
	h.blockMapper = blockMapper
}

// SetRAGServices sets the RAG services (Day 2)
func (h *Handlers) SetRAGServices(
	hybridSearch *services.HybridSearchService,
	reranker *services.ReRankerService,
	agenticRAG *services.AgenticRAGService,
	memory *services.MemoryService,
) {
	h.hybridSearchService = hybridSearch
	h.rerankerService = reranker
	h.agenticRAGService = agenticRAG
	h.memoryService = memory
}

// SetMultiModalServices sets the multi-modal search services (Feature 7)
func (h *Handlers) SetMultiModalServices(
	multiModalSearch *services.MultiModalSearchService,
	imageEmbedding *services.ImageEmbeddingService,
) {
	// Create multimodal handler
	h.multiModalHandler = NewMultiModalSearchHandler(multiModalSearch, imageEmbedding)
}

// SetWorkspaceService sets the workspace service (Feature 1)
func (h *Handlers) SetWorkspaceService(workspaceService *services.WorkspaceService) {
	h.workspaceService = workspaceService
}

// SetRoleContextService sets the role context service (Feature 1 - Permissions)
func (h *Handlers) SetRoleContextService(roleContextService *services.RoleContextService) {
	h.roleContextService = roleContextService
}

// SetMemoryHierarchyService sets the memory hierarchy service (Q1 - Memory Hierarchy)
func (h *Handlers) SetMemoryHierarchyService(memoryHierarchyService *services.MemoryHierarchyService) {
	h.memoryHierarchyService = memoryHierarchyService
}

// SetInviteService sets the workspace invite service (Feature 1 - Email Invites)
func (h *Handlers) SetInviteService(inviteService *services.WorkspaceInviteService) {
	h.inviteService = inviteService
}

// SetAuditService sets the workspace audit service (Feature 1 - Audit Logging)
func (h *Handlers) SetAuditService(auditService *services.WorkspaceAuditService) {
	h.auditService = auditService
}

// SetProjectAccessService sets the project access service (Feature 1 - Project Access Control)
func (h *Handlers) SetProjectAccessService(projectAccessService *services.ProjectAccessService) {
	h.projectAccessService = projectAccessService
}

// SetSkillsLoader sets the skills loader (Agent Skills System)
func (h *Handlers) SetSkillsLoader(skillsLoader *services.SkillsLoader) {
	h.skillsLoader = skillsLoader
}

// SetOSAServices sets the OSA integration services
func (h *Handlers) SetOSAServices(client *osa.ResilientClient, syncService *services.OSASyncService) {
	h.osaClient = client
	h.osaSyncService = syncService
}

// SetOSAFileServices sets the OSA file sync and workflow services
func (h *Handlers) SetOSAFileServices(fileSyncService *services.OSAFileSyncService, workspaceInit *services.OSAWorkspaceInitService, workflowsHandler *OSAWorkflowsHandler, webhooksHandler *OSAWebhooksHandler, eventBus *services.BuildEventBus, streamingHandler *OSAStreamingHandler) {
	h.osaFileSyncService = fileSyncService
	h.osaWorkspaceInit = workspaceInit
	h.osaWorkflowsHandler = workflowsHandler
	h.osaWebhooksHandler = webhooksHandler
	h.osaBuildEventBus = eventBus
	h.osaStreamingHandler = streamingHandler
}

// SetOSAOnboardingHandler sets the OSA onboarding handler
func (h *Handlers) SetOSAOnboardingHandler(handler *OSAOnboardingHandler) {
	h.osaOnboardingHandler = handler
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
		artifacts.GET("/:id/versions", h.GetArtifactVersions)
		artifacts.POST("/:id/restore", h.RestoreArtifactVersion)
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
		// Project members (team assignment with role-based access)
		projects.GET("/:id/members", h.ListProjectMembers)
		projects.POST("/:id/members", h.AddProjectMember)
		projects.PUT("/:id/members/:memberId/role", h.UpdateProjectMemberRole)
		projects.DELETE("/:id/members/:memberId", h.RemoveProjectMember)
		projects.GET("/:id/access/:userId", h.CheckProjectAccess)
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

	// Deals routes - /api/deals (legacy simple deals)
	deals := api.Group("/deals")
	deals.Use(auth)
	{
		deals.GET("", h.ListDeals)
		deals.PATCH("/:id/stage", h.UpdateDealStage)
	}

	// CRM routes - /api/crm (full CRM pipeline system)
	crm := api.Group("/crm")
	crm.Use(auth)
	{
		// Companies
		crm.GET("/companies", h.ListCompanies)
		crm.POST("/companies", h.CreateCompany)
		crm.GET("/companies/search", h.SearchCompanies)
		crm.GET("/companies/:id", h.GetCompany)
		crm.PUT("/companies/:id", h.UpdateCompany)
		crm.DELETE("/companies/:id", h.DeleteCompany)
		// Company contacts (linking to clients)
		crm.GET("/companies/:id/contacts", h.ListCompanyContacts)
		crm.POST("/companies/:id/contacts", h.LinkContactToCompany)
		crm.DELETE("/companies/:id/contacts/:relationId", h.UnlinkContactFromCompany)

		// Pipelines
		crm.GET("/pipelines", h.ListPipelines)
		crm.POST("/pipelines", h.CreatePipeline)
		crm.GET("/pipelines/:id", h.GetPipeline)
		crm.PUT("/pipelines/:id", h.UpdatePipeline)
		crm.DELETE("/pipelines/:id", h.DeletePipeline)
		// Pipeline stages
		crm.GET("/pipelines/:id/stages", h.ListPipelineStages)
		crm.POST("/pipelines/:id/stages", h.CreatePipelineStage)
		crm.PUT("/pipelines/:id/stages/:stageId", h.UpdatePipelineStage)
		crm.DELETE("/pipelines/:id/stages/:stageId", h.DeletePipelineStage)
		crm.POST("/pipelines/:id/stages/reorder", h.ReorderPipelineStages)

		// Deals (CRM pipeline deals)
		crm.GET("/deals", h.ListCRMDeals)
		crm.POST("/deals", h.CreateCRMDeal)
		crm.GET("/deals/stats", h.GetCRMDealStats)
		crm.GET("/deals/:id", h.GetCRMDeal)
		crm.PUT("/deals/:id", h.UpdateCRMDeal)
		crm.PATCH("/deals/:id/stage", h.MoveCRMDealStage)
		crm.PATCH("/deals/:id/status", h.UpdateCRMDealStatus)
		crm.DELETE("/deals/:id", h.DeleteCRMDeal)
		// Deal activities
		crm.GET("/deals/:id/activities", h.ListDealActivities)

		// Activities
		crm.GET("/activities", h.ListCRMActivities)
		crm.POST("/activities", h.CreateCRMActivity)
		crm.POST("/activities/:id/complete", h.CompleteCRMActivity)
		crm.DELETE("/activities/:id", h.DeleteCRMActivity)
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

	// Workspaces routes - /api/workspaces (Feature 1: Team/Collaboration)
	workspaces := api.Group("/workspaces")
	workspaces.Use(auth)
	{
		// Workspace CRUD - no role context needed for create/list
		workspaces.POST("", h.CreateWorkspace)
		workspaces.GET("", h.ListWorkspaces)

		// Workspace-scoped routes - inject role context
		workspaceScoped := workspaces.Group("/:id")
		workspaceScoped.Use(middleware.InjectRoleContext(h.pool, h.roleContextService))
		{
			// Read operations - any member
			workspaceScoped.GET("", h.GetWorkspace)
			workspaceScoped.GET("/members", h.ListWorkspaceMembers)
			workspaceScoped.GET("/roles", h.ListWorkspaceRoles)
			workspaceScoped.GET("/profile", h.GetWorkspaceProfile)     // User's profile in workspace
			workspaceScoped.GET("/role-context", h.GetUserRoleContext) // User's role & permissions

			// Update user profile
			workspaceScoped.PUT("/profile", h.UpdateWorkspaceProfile)

			// Update workspace - requires admin or owner
			workspaceScoped.PUT("", middleware.RequireWorkspaceAdmin(), h.UpdateWorkspace)

			// Delete workspace - requires owner only
			workspaceScoped.DELETE("", middleware.RequireWorkspaceOwner(h.pool), h.DeleteWorkspace)

			// Invite members - requires manager, admin, or owner
			workspaceScoped.POST("/members/invite", middleware.RequireWorkspaceManager(), h.AddWorkspaceMember)

			// Update/remove members - requires admin or owner
			workspaceScoped.PUT("/members/:userId", middleware.RequireWorkspaceAdmin(), h.UpdateWorkspaceMemberRole)
			workspaceScoped.DELETE("/members/:userId", middleware.RequireWorkspaceAdmin(), h.RemoveWorkspaceMember)

			// Workspace invitations - manager+ can invite
			workspaceScoped.POST("/invites", middleware.RequireWorkspaceManager(), h.CreateWorkspaceInvite)
			workspaceScoped.GET("/invites", middleware.RequireWorkspaceAdmin(), h.ListWorkspaceInvites)
			workspaceScoped.DELETE("/invites/:inviteId", middleware.RequireWorkspaceAdmin(), h.RevokeWorkspaceInvite)

			// Audit logs - admin+ can view
			workspaceScoped.GET("/audit-logs", middleware.RequireWorkspaceAdmin(), h.ListAuditLogs)
			workspaceScoped.GET("/audit-logs/:logId", middleware.RequireWorkspaceAdmin(), h.GetAuditLog)
			workspaceScoped.GET("/audit-logs/user/:userId", middleware.RequireWorkspaceAdmin(), h.GetUserActivity)
			workspaceScoped.GET("/audit-logs/resource/:resourceType/:resourceId", middleware.RequireWorkspaceAdmin(), h.GetResourceHistory)
			workspaceScoped.GET("/audit-logs/stats/actions", middleware.RequireWorkspaceAdmin(), h.GetActionStats)
			workspaceScoped.GET("/audit-logs/stats/active-users", middleware.RequireWorkspaceAdmin(), h.GetMostActiveUsers)

			// Workspace memory routes - CUS-25
			memoryHandler := NewWorkspaceMemoryHandlers(h.pool)
			RegisterWorkspaceMemoryRoutes(workspaceScoped, memoryHandler)
		}

		// Public invite acceptance (no workspace context needed)
		workspaces.POST("/invites/accept", h.AcceptWorkspaceInvite)
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
		nodes.POST("/:id/archive", h.ArchiveNode)
		nodes.POST("/:id/unarchive", h.UnarchiveNode)
		// Node linking
		nodes.GET("/:id/links", h.GetNodeLinks)
		nodes.GET("/:id/links/counts", h.GetNodeLinkCounts)
		nodes.POST("/:id/links/projects", h.LinkNodeProject)
		nodes.DELETE("/:id/links/projects/:projectId", h.UnlinkNodeProject)
		nodes.POST("/:id/links/contexts", h.LinkNodeContext)
		nodes.DELETE("/:id/links/contexts/:contextId", h.UnlinkNodeContext)
		nodes.POST("/:id/links/conversations", h.LinkNodeConversation)
		nodes.DELETE("/:id/links/conversations/:conversationId", h.UnlinkNodeConversation)
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
		search.GET("/web", h.WebSearch)                           // Basic web search
		search.GET("/context", h.WebSearchWithContext)            // Search with formatted context
		search.GET("/history", h.ListSearchHistory)               // List user's search history
		search.GET("/history/:id", h.GetSearchHistoryEntry)       // Get specific search details
		search.DELETE("/history/:id", h.DeleteSearchHistoryEntry) // Delete specific search
		search.DELETE("/history", h.ClearSearchHistory)           // Clear all search history

		// Enhanced RAG search endpoints
		if h.hybridSearchService != nil {
			search.POST("/hybrid", h.HybridSearch) // Hybrid semantic + keyword search
			search.POST("/rerank", h.HybridSearch) // Re-rank search results (uses hybrid search)
		}
		if h.multiModalHandler != nil {
			search.POST("/multimodal", h.multiModalHandler.SearchWithImage) // Multi-modal search
		}
		if h.hybridSearchService != nil {
			search.GET("/explain", h.HybridSearchExplain) // Explain search results
		}
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
		// Output styles & preferences
		ai.GET("/output-styles", h.ListOutputStyles)
		ai.GET("/output-preferences", h.GetUserOutputPreference)
		ai.PUT("/output-preferences", h.UpsertUserOutputPreference)
		// Agent presets (templates for custom agents) - must be before /agents/:id
		ai.GET("/agents/presets", h.ListAgentPresets)
		ai.GET("/agents/presets/:id", h.GetAgentPreset)
		// Agent prompts (built-in)
		ai.GET("/agents", h.GetAgentPrompts)
		ai.GET("/agents/:id", h.GetAgentPrompt)
		// Custom user agents - specific routes before parameterized ones
		ai.GET("/custom-agents", h.ListCustomAgents)
		ai.POST("/custom-agents", h.CreateCustomAgent)
		ai.POST("/custom-agents/sandbox", h.TestCustomAgent) // Test arbitrary prompt
		ai.GET("/custom-agents/category/:category", h.ListCustomAgentsByCategory)
		ai.POST("/custom-agents/from-preset/:presetId", h.CreateAgentFromPreset)
		ai.GET("/custom-agents/:id", h.GetCustomAgent)
		ai.PUT("/custom-agents/:id", h.UpdateCustomAgent)
		ai.DELETE("/custom-agents/:id", h.DeleteCustomAgent)
		ai.POST("/custom-agents/:id/test", h.TestCustomAgent) // Test existing agent
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

	// Notification routes - /api/notifications
	if h.notificationService != nil {
		notifHandler := NewNotificationHandler(h.notificationService)
		notifications := api.Group("/notifications")
		notifications.Use(auth)
		{
			notifications.GET("", notifHandler.ListNotifications)
			notifications.GET("/unread-count", notifHandler.GetUnreadCount)
			notifications.GET("/stream", notifHandler.Stream)
			notifications.GET("/preferences", notifHandler.GetPreferences)
			notifications.PUT("/preferences", notifHandler.UpdatePreferences)
			notifications.POST("/:id/read", notifHandler.MarkAsRead)
			notifications.POST("/read", notifHandler.MarkMultipleAsRead)
			notifications.POST("/read-all", notifHandler.MarkAllAsRead)
			notifications.DELETE("/:id", notifHandler.DeleteNotification)

			// Web Push routes
			if h.webPushService != nil {
				pushHandler := NewWebPushHandler(h.webPushService)
				notifications.GET("/push/vapid-public-key", pushHandler.GetVAPIDPublicKey)
				notifications.POST("/push/subscribe", pushHandler.Subscribe)
				notifications.POST("/push/unsubscribe", pushHandler.Unsubscribe)
				notifications.POST("/push/test", pushHandler.TestPush)
			}
		}

		// DEV ONLY: Notification seeding routes - /api/dev/notifications
		if IsDevMode() {
			seedHandler := NewNotificationSeedHandler(h.pool, h.notificationService)
			devNotifications := api.Group("/dev/notifications")
			devNotifications.Use(auth)
			{
				devNotifications.POST("/seed", seedHandler.SeedNotifications)
				devNotifications.POST("/seed-full", seedHandler.SeedNotificationsWithTimestamps)
				devNotifications.DELETE("/seed", seedHandler.ClearSeedNotifications)
			}
		}
	}

	// Email routes - /api/email
	emailHandler := NewEmailHandler()
	email := api.Group("/email")
	email.Use(auth)
	{
		email.GET("/status", emailHandler.GetEmailStatus)
		email.POST("/test", emailHandler.HandleTestEmail)
		email.POST("/send/verification", emailHandler.HandleSendVerificationEmail)
		email.POST("/send/password-reset", emailHandler.HandleSendPasswordResetEmail)
		email.POST("/send/welcome", emailHandler.HandleSendWelcomeEmail)
	}

	// Custom Dashboards routes - /api/user-dashboards
	userDashboards := api.Group("/user-dashboards")
	userDashboards.Use(auth)
	{
		userDashboards.GET("", h.ListUserDashboards)
		userDashboards.POST("", h.CreateUserDashboard)
		userDashboards.GET("/:id", h.GetUserDashboard)
		userDashboards.PUT("/:id", h.UpdateUserDashboard)
		userDashboards.DELETE("/:id", h.DeleteUserDashboard)
		userDashboards.POST("/:id/duplicate", h.DuplicateUserDashboard)
		userDashboards.PUT("/:id/layout", h.UpdateDashboardLayout)
		userDashboards.POST("/:id/default", h.SetDefaultUserDashboard)
		userDashboards.POST("/:id/share", h.ShareUserDashboard)
	}
	// Public shared dashboard access (no auth)
	api.GET("/user-dashboards/shared/:token", h.GetSharedDashboard)

	// Dashboard Widgets routes - /api/widgets
	widgets := api.Group("/widgets")
	widgets.Use(auth)
	{
		widgets.GET("", h.ListWidgetTypes)
		widgets.GET("/:type/schema", h.GetWidgetSchema)
	}

	// Dashboard Templates routes - /api/dashboard-templates
	dashboardTemplates := api.Group("/dashboard-templates")
	dashboardTemplates.Use(auth)
	{
		dashboardTemplates.GET("", h.ListDashboardTemplates)
		dashboardTemplates.POST("/create-from/:id", h.CreateDashboardFromTemplate)
	}

	// Analytics routes - /api/analytics
	analytics := api.Group("/analytics")
	analytics.Use(auth)
	{
		analytics.GET("/summary", h.GetAnalyticsSummary)
		analytics.GET("/task-burndown", h.GetTaskBurndown)
		analytics.GET("/workload", h.GetWorkloadHeatmap)
		analytics.GET("/upcoming-deadlines", h.GetUpcomingDeadlines)
	}

	// Comments routes - /api/comments
	if h.commentService != nil {
		comments := api.Group("/comments")
		comments.Use(auth)
		{
			comments.GET("", h.GetComments)
			comments.POST("", h.CreateComment)
			comments.GET("/:id", h.GetComment)
			comments.PUT("/:id", h.UpdateComment)
			comments.DELETE("/:id", h.DeleteComment)
			comments.POST("/:id/reactions", h.AddCommentReaction)
			comments.DELETE("/:id/reactions/:emoji", h.RemoveCommentReaction)
		}

		// Task comments - /api/tasks/:id/comments
		api.GET("/tasks/:id/comments", auth, h.GetTaskComments)
		api.POST("/tasks/:id/comments", auth, h.CreateTaskComment)

		// Project comments - /api/projects/:id/comments
		api.GET("/projects/:id/comments", auth, h.GetProjectComments)
		api.POST("/projects/:id/comments", auth, h.CreateProjectComment)
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

	// RAG routes - /api/rag (Day 2: Advanced retrieval with hybrid search + agentic RAG)
	if h.hybridSearchService != nil || h.agenticRAGService != nil || h.memoryService != nil || h.rerankerService != nil || h.multiModalHandler != nil {
		rag := api.Group("/rag")
		rag.Use(auth)
		{
			// Hybrid search endpoints
			if h.hybridSearchService != nil {
				rag.POST("/search/hybrid", h.HybridSearch)
				rag.POST("/search/hybrid/explain", h.HybridSearchExplain)
			}

			// Re-ranking endpoints
			if h.rerankerService != nil {
				rag.POST("/search/rerank", h.ReRankResults)
				rag.POST("/search/rerank/explain", h.ReRankExplain)
			}

			// Search explanation endpoint (hybrid search only, multimodal is registered separately)
			if h.hybridSearchService != nil {
				rag.GET("/search/explain", h.SearchExplain)
			}

			// Agentic RAG endpoint
			if h.agenticRAGService != nil {
				rag.POST("/retrieve", h.AgenticRAGRetrieve)
			}

			// Memory endpoints
			if h.memoryService != nil {
				rag.GET("/memories", h.MemoryList)
				rag.GET("/memories/:id", h.MemoryGet)
				rag.POST("/memories", h.MemoryCreate)
			}
		}
	}

	// Multi-modal Search routes - /api/images, /api/search (Feature 7: Multi-modal Embeddings)
	if h.multiModalHandler != nil {
		h.RegisterMultiModalRoutes(api, h.multiModalHandler)
	}

	// Memory routes - /api/memories (episodic memory system)
	memoryHandler := NewMemoryHandler(h.pool, h.embeddingService)
	memories := api.Group("/memories")
	memories.Use(auth)
	{
		memories.GET("", memoryHandler.ListMemories)
		memories.POST("", memoryHandler.CreateMemory)
		memories.GET("/stats", memoryHandler.GetMemoryStats)
		memories.POST("/search", memoryHandler.SearchMemories)
		memories.POST("/relevant", memoryHandler.GetRelevantMemories)
		memories.GET("/project/:projectId", memoryHandler.GetProjectMemories)
		memories.GET("/node/:nodeId", memoryHandler.GetNodeMemories)
		memories.GET("/:id", memoryHandler.GetMemory)
		memories.PUT("/:id", memoryHandler.UpdateMemory)
		memories.DELETE("/:id", memoryHandler.DeleteMemory)
		memories.POST("/:id/pin", memoryHandler.PinMemory)
	}

	// User Facts routes - /api/user-facts
	userFacts := api.Group("/user-facts")
	userFacts.Use(auth)
	{
		userFacts.GET("", memoryHandler.ListUserFacts)
		userFacts.PUT("/:key", memoryHandler.UpdateUserFact)
		userFacts.POST("/:key/confirm", memoryHandler.ConfirmUserFact)
		userFacts.POST("/:key/reject", memoryHandler.RejectUserFact)
		userFacts.DELETE("/:key", memoryHandler.DeleteUserFact)
	}

	// Context Tree routes - /api/context-tree (hierarchical context for agents)
	contextTreeHandler := NewContextTreeHandler(h.pool, h.embeddingService)
	contextTree := api.Group("/context-tree")
	contextTree.Use(auth)
	{
		// Tree retrieval
		contextTree.GET("/:entityType/:entityId", contextTreeHandler.GetContextTree)
		contextTree.POST("/search", contextTreeHandler.SearchContextTree)
		contextTree.POST("/load", contextTreeHandler.LoadContextItem)
		contextTree.GET("/stats", contextTreeHandler.GetContextStats)
		// Loading rules
		contextTree.GET("/rules/:entityType/:entityId", contextTreeHandler.GetLoadingRules)
		// Context sessions
		contextTree.POST("/session", contextTreeHandler.CreateContextSession)
		contextTree.GET("/session/:sessionId", contextTreeHandler.GetContextSession)
		contextTree.PUT("/session/:sessionId", contextTreeHandler.UpdateContextSession)
		contextTree.DELETE("/session/:sessionId", contextTreeHandler.EndContextSession)
	}

	// Transcription routes - /api/transcribe
	transcriptionHandler := NewTranscriptionHandler(h.pool)
	transcribe := api.Group("/transcribe")
	transcribe.Use(auth)
	{
		transcribe.POST("", transcriptionHandler.TranscribeAudio)
		transcribe.GET("/status", transcriptionHandler.GetTranscriptionStatus)
		// ⚡ Removed deprecated realtime endpoint - LiveKit handles all voice now
	}
	slog.Info("Transcription routes registered")

	// OSA Voice routes - /api/osa
	// NOTE: These routes are deprecated - voice now uses Python LiveKit agent
	// Uncomment if you need standalone TTS endpoints
	// osa := api.Group("/osa")
	// osa.Use(auth)
	// {
	// 	osa.POST("/speak", h.HandleOSASpeak)              // Convert text to speech
	// 	osa.POST("/speak/stream", h.HandleOSASpeakStream) // Stream TTS for long text
	// }
	// slog.Info("OSA voice routes registered")

	// LiveKit Voice routes - /api/livekit (Real-time voice with LiveKit)
	livekit := api.Group("/livekit")
	livekit.Use(auth)
	{
		livekit.POST("/token", h.HandleLiveKitToken) // Generate LiveKit room token
		livekit.GET("/rooms", h.HandleLiveKitRooms)  // Get room info/status
		// NOTE: /transcribe and /greeting routes deprecated - use /api/voice/* instead
	}
	slog.Info("LiveKit voice routes registered")

	// User context endpoint for Python LiveKit voice agent (no auth - internal service call)
	// Called by python-voice-agent/ to fetch user context for personalized responses
	api.GET("/voice/user-context/:user_id", h.HandleVoiceUserContext)
	slog.Info("Voice user-context route registered")

	// Voice notes routes - /api/voice-notes
	voiceNotesHandler := NewVoiceNotesHandler(h.pool, h.embeddingService)
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

	// Username routes - /api/users
	usernameHandler := NewUsernameHandler(h.pool)
	users := api.Group("/users")
	{
		// Public route - check username availability (no auth required for UX)
		users.GET("/check-username/:username", usernameHandler.CheckUsernameAvailability)

		// Protected routes
		users.GET("/me", auth, usernameHandler.GetCurrentUser)
		users.PATCH("/me/username", auth, usernameHandler.SetUsername)
		users.POST("/me/complete-onboarding", auth, usernameHandler.CompleteOnboarding)
	}

	// MCP routes - /api/mcp
	mcp := api.Group("/mcp")
	mcp.Use(auth)
	{
		mcp.GET("/tools", h.ListMCPTools)
		mcp.POST("/execute", h.ExecuteMCPTool)
		mcp.GET("/health", h.MCPHealth)
	}

	// Tables routes - /api/tables (NocoDB-inspired flexible tables)
	tables := api.Group("/tables")
	tables.Use(auth)
	{
		// Tables CRUD
		tables.GET("", h.ListTables)
		tables.POST("", h.CreateTable)
		tables.GET("/:id", h.GetTable)
		tables.PUT("/:id", h.UpdateTable)
		tables.DELETE("/:id", h.DeleteTable)
		// Columns (Fields) CRUD - matches frontend API naming
		tables.GET("/:id/columns", h.ListFields)
		tables.POST("/:id/columns", h.CreateField)
		tables.PUT("/:id/columns/:columnId", h.UpdateField)
		tables.DELETE("/:id/columns/:columnId", h.DeleteField)
		tables.POST("/:id/columns/reorder", h.ReorderFields)
		// Rows (Records) CRUD - matches frontend API naming
		tables.GET("/:id/rows", h.ListRecords)
		tables.POST("/:id/rows", h.CreateRecord)
		tables.GET("/:id/rows/:rowId", h.GetRecord)
		tables.PUT("/:id/rows/:rowId", h.UpdateRecord)
		tables.DELETE("/:id/rows/:rowId", h.DeleteRecord)
		tables.POST("/:id/rows/bulk-delete", h.BulkDeleteRecords)
		// Views CRUD
		tables.GET("/:id/views", h.ListViews)
		tables.POST("/:id/views", h.CreateView)
		tables.PUT("/:id/views/:viewId", h.UpdateView)
		tables.DELETE("/:id/views/:viewId", h.DeleteView)
	}

	// ============================================================================
	// NEW Integration Architecture - Provider-based handlers
	// ============================================================================
	// Initialize the new IntegrationRouter which manages all integration providers
	// (Google, Slack, Notion) with their OAuth flows, data sync, and API handlers.
	integrationRouter := NewIntegrationRouter(h.pool)

	// Register new integration routes - /api/integrations/{provider}/*
	// This provides: OAuth flows, calendar, gmail, slack channels/messages, notion databases/pages
	integrationsGroup := api.Group("/integrations")
	integrationRouter.RegisterRoutes(integrationsGroup, auth)

	// Terminal routes - /api/terminal
	terminalHandler := NewTerminalHandler(h.containerMgr, h.terminalPubSub)
	terminalRoutes := api.Group("/terminal")
	terminalRoutes.Use(auth)
	{
		terminalRoutes.GET("/ws", terminalHandler.HandleWebSocket)
		terminalRoutes.GET("/sessions", terminalHandler.ListSessions)
		terminalRoutes.DELETE("/sessions/:id", terminalHandler.CloseSession)
	}

	// NOTE: Native Window Capture routes moved to feature/native-app-capture branch
	// Web apps (iframe) are the recommended approach - see user_apps.go

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

	// =============================================================================
	// MOBILE API - /api/mobile/v1
	// =============================================================================
	// Optimized endpoints for mobile clients (PWA, native apps):
	// - Lean payloads (~100 bytes vs ~2KB per item)
	// - Cursor-based pagination
	// - Field selection support (?fields=id,title,status)
	// - Unix timestamps instead of ISO strings
	// - Standardized error responses
	// =============================================================================
	mobileHandler := NewMobileHandler(h.pool, h.notificationService)
	mobile := api.Group("/mobile/v1")
	mobile.Use(auth)
	mobile.Use(middleware.DeviceIDMiddleware()) // Extract X-Device-ID header
	{
		// User profile
		mobile.GET("/me", mobileHandler.GetMe)

		// Tasks (Phase 1)
		mobile.GET("/tasks", mobileHandler.ListTasks)
		mobile.GET("/tasks/:id", mobileHandler.GetTask)
		mobile.POST("/tasks/quick", mobileHandler.QuickCreateTask)
		mobile.PUT("/tasks/:id/status", mobileHandler.UpdateTaskStatus)
		mobile.PUT("/tasks/:id/toggle", mobileHandler.ToggleTask)

		// Notifications (Phase 2)
		mobile.GET("/notifications", mobileHandler.ListNotifications)
		mobile.GET("/notifications/count", mobileHandler.GetNotificationCount)
		mobile.POST("/notifications/mark-read", mobileHandler.MarkNotificationsRead)

		// Daily Log (Phase 3)
		mobile.GET("/dailylog/today", mobileHandler.GetTodayLog)
		mobile.GET("/dailylog/history", mobileHandler.GetLogHistory)

		// Sync (Phase 3)
		mobile.GET("/sync", mobileHandler.DeltaSync)

		// Chat (Phase 4)
		mobile.GET("/chat/threads", mobileHandler.ListChatThreads)
		mobile.GET("/chat/history/:id", mobileHandler.GetChatHistory)
		mobile.POST("/chat/message", mobileHandler.SendChatMessage)

		// Capture (Phase 4 - pending team discussion)
		// mobile.POST("/capture", mobileHandler.SmartCapture)

		// Push Registration (Phase 5)
		mobile.POST("/push/register", mobileHandler.RegisterPushDevice)
		mobile.DELETE("/push/unregister", mobileHandler.UnregisterPushDevice)
	}

	// Calendar routes - /api/calendar (aggregate from calendar_events table)
	calendar := api.Group("/calendar")
	{
		calendar.GET("/stats", auth, h.GetCalendarStats)
		calendar.GET("/upcoming", auth, h.GetUpcomingCalendarEvents)
		calendar.GET("/today", auth, h.GetTodayCalendarEvents)
	}

	// Authentication routes - /api/auth
	// Apply strict rate limiting to prevent brute force attacks
	strictRateLimit := middleware.StrictRateLimitMiddleware()

	googleAuthHandler := NewGoogleAuthHandler(h.pool, h.cfg, h.sessionCache)
	logger := slog.Default()
	emailAuthHandler := NewEmailAuthHandler(h.pool, h.cfg, h.notificationTriggers, h.osaWorkspaceInit, logger)
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

	// Pedro Tasks - Document Processing routes - /api/documents
	if h.documentProcessor != nil {
		documentHandler := NewDocumentHandler(h.documentProcessor)
		protectedDocs := api.Group("")
		protectedDocs.Use(auth)
		RegisterDocumentRoutes(protectedDocs, documentHandler)
	}

	// Pedro Tasks - Learning/Personalization routes - /api/learning
	if h.learningService != nil {
		learningHandler := NewLearningHandler(h.learningService)
		protectedLearning := api.Group("")
		protectedLearning.Use(auth)
		RegisterLearningRoutes(protectedLearning, learningHandler)
	}

	// Pedro Tasks - App Profiler routes - /api/app-profiles
	if h.appProfilerService != nil {
		appProfilerHandler := NewAppProfilerHandler(h.appProfilerService)
		protectedProfiler := api.Group("")
		protectedProfiler.Use(auth)
		RegisterAppProfilerRoutes(protectedProfiler, appProfilerHandler)
	}

	// Pedro Tasks - Conversation Intelligence routes - /api/intelligence
	if h.conversationIntelligence != nil && h.memoryExtractor != nil {
		intelligenceHandler := NewConversationIntelligenceHandler(h.conversationIntelligence, h.memoryExtractor)
		protectedIntel := api.Group("")
		protectedIntel.Use(auth)
		RegisterConversationIntelligenceRoutes(protectedIntel, intelligenceHandler)
	}

	// ============================================================================
	// Sorx Integration Module
	// ============================================================================

	// Initialize Sorx service and handler with engine
	sorxService := services.NewSorxService(h.pool, h.cfg)
	sorxLogger := slog.Default().With("component", "sorx")
	sorxHandler := NewSorxHandler(sorxService, h.pool, sorxLogger)

	// Sorx routes - /api/sorx (for Sorx engine callbacks and credentials)
	sorxRoutes := api.Group("/sorx")
	{
		// Credential endpoints (internal API for Sorx engine)
		sorxRoutes.POST("/credential-ticket", sorxHandler.RequestCredentialTicket)
		sorxRoutes.POST("/redeem-credential", sorxHandler.RedeemCredential)
		// Callback endpoint (for skill execution callbacks)
		sorxRoutes.POST("/callback", sorxHandler.HandleCallback)

		// Public skill catalog (browse available skills)
		sorxRoutes.GET("/skills", sorxHandler.ListSkills)
		sorxRoutes.GET("/skills/:id", sorxHandler.GetSkill)

		// Public skill commands catalog (lists available skill commands)
		sorxRoutes.GET("/commands", sorxHandler.ListSkillCommands)

		// Protected endpoints (require user auth)
		sorxProtected := sorxRoutes.Group("")
		sorxProtected.Use(auth)
		{
			// Decision endpoints (human-in-the-loop)
			sorxProtected.GET("/decisions", sorxHandler.GetPendingDecisions)
			sorxProtected.GET("/decisions/:id", sorxHandler.GetDecision)
			sorxProtected.POST("/decisions/:id/respond", sorxHandler.RespondToDecision)
			// Skill execution
			sorxProtected.POST("/execute", sorxHandler.TriggerSkill)
			sorxProtected.GET("/executions/:id", sorxHandler.GetSkillExecution)
			// Skill command execution (slash commands that trigger skills)
			sorxProtected.POST("/commands/execute", sorxHandler.ExecuteSkillCommand)
		}
	}

	// Integrations Module - /api/integrations (for user integration management)
	// Uses new IntegrationRouter for provider catalog and sync operations
	integrationsHandler := NewIntegrationsHandler(h.pool, integrationRouter)
	integrationsModule := api.Group("/integrations")

	// Public endpoints - provider catalog (no auth required)
	// These just list what integrations are available, not user-specific data
	integrationsModule.GET("/providers", integrationsHandler.GetProviders)
	integrationsModule.GET("/providers/:id", integrationsHandler.GetProvider)

	// Protected endpoints - user-specific data (auth required)
	integrationsProtected := integrationsModule.Group("")
	integrationsProtected.Use(auth)
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

	// Module-specific integration endpoints - /api/modules/:module/integrations
	modules := api.Group("/modules")
	modules.Use(optionalAuth) // Optional auth for browsing available integrations
	{
		modules.GET("/:module/integrations", integrationsHandler.GetModuleIntegrations)
	}

	// Desktop 3D Layout routes - /api/desktop3d
	// Custom 3D Desktop layout persistence
	desktop3d := api.Group("/desktop3d")
	desktop3d.Use(auth)
	{
		desktop3d.GET("/layouts", h.ListDesktop3DLayouts)
		desktop3d.POST("/layouts", h.CreateDesktop3DLayout)
		desktop3d.GET("/layouts/active", h.GetActiveDesktop3DLayout)
		desktop3d.POST("/layouts/:id/activate", h.ActivateDesktop3DLayout)
		desktop3d.DELETE("/layouts/:id", h.DeleteDesktop3DLayout)
	}
	slog.Info("Desktop 3D layout routes registered")

	// Onboarding routes - /api/onboarding
	// Conversational AI onboarding with workspace setup
	onboardingAIService := services.NewOnboardingAIService()
	onboardingService := services.NewOnboardingService(h.pool, onboardingAIService)
	onboardingHandler := NewOnboardingHandler(onboardingService)
	onboardingHandler.RegisterOnboardingRoutes(api, auth)
	slog.Info("Onboarding routes registered", "ai_provider", onboardingAIService.GetProvider())

	// ============================================================================
	// OSA Integration Module (21-Agent Orchestration System)
	// ============================================================================

	log.Printf("DEBUG: About to check OSA routes, h.osaClient != nil: %v", h.osaClient != nil)
	if h.osaClient != nil {
		log.Printf("DEBUG: Registering OSA authenticated routes...")

		// OSA authenticated routes - /api/osa (auth required)
		// Note: /api/osa/health is registered publicly in main.go (no auth)
		osaAuth := api.Group("/osa")
		osaAuth.Use(auth)
		{
			// Generation endpoints
			osaAuth.POST("/generate", h.HandleGenerateApp)
			osaAuth.GET("/status/:app_id", h.HandleGetAppStatus)
			osaAuth.GET("/workspaces", h.HandleListWorkspaces)

			// Workflow endpoints (if handlers are initialized)
			if h.osaWorkflowsHandler != nil {
				osaAuth.GET("/workflows", h.osaWorkflowsHandler.ListWorkflows)
				osaAuth.GET("/workflows/:id", h.osaWorkflowsHandler.GetWorkflow)
				osaAuth.GET("/workflows/:id/files", h.osaWorkflowsHandler.GetWorkflowFiles)
				osaAuth.GET("/workflows/:id/files/:type", h.osaWorkflowsHandler.GetFileContent)
				osaAuth.GET("/files/:id/content", h.osaWorkflowsHandler.GetFileContentByID)
				osaAuth.POST("/modules/install", h.osaWorkflowsHandler.InstallModule)
				osaAuth.POST("/sync/trigger", h.osaWorkflowsHandler.TriggerSync)
			}

			// Webhook management endpoints (if handlers are initialized)
			if h.osaWebhooksHandler != nil {
				osaAuth.GET("/webhooks", h.osaWebhooksHandler.ListWebhooks)
				osaAuth.POST("/webhooks/register", h.osaWebhooksHandler.RegisterWebhook)
			}

			// SSE streaming endpoints (if handlers are initialized)
			if h.osaStreamingHandler != nil {
				osaAuth.GET("/stream/build/:app_id", h.osaStreamingHandler.StreamBuildProgress)
				osaAuth.GET("/stream/stats", h.osaStreamingHandler.GetStreamStats)
				osaAuth.GET("/stream/stats/:app_id", h.osaStreamingHandler.GetAppStreamStats)
			}
		}
		log.Printf("✅ OSA authenticated API routes registered at /api/osa/*")

		// OSA Internal routes - /api/internal/osa (for terminal containers)
		// No auth middleware - uses X-User-ID header from container environment
		osaInternal := api.Group("/internal/osa")
		{
			osaInternal.POST("/generate", h.HandleInternalGenerateApp)
			osaInternal.GET("/status/:app_id", h.HandleInternalGetAppStatus)
			osaInternal.GET("/workspaces", h.HandleInternalListWorkspaces)
			osaInternal.GET("/health", h.HandleInternalOSAHealth)
		}
		log.Printf("✅ OSA internal API routes registered at /api/internal/osa/*")
	} else {
		log.Printf("DEBUG: Skipping OSA routes, osaClient is nil")
	}

	// OSA Webhook receiver routes - /api/osa/webhooks (OSA-5 → BusinessOS)
	// Public endpoints with HMAC verification (no JWT auth)
	if h.osaWebhooksHandler != nil {
		osaWebhooks := api.Group("/osa/webhooks")
		{
			osaWebhooks.POST("/workflow-complete", h.osaWebhooksHandler.HandleWorkflowComplete)
			osaWebhooks.POST("/build-event", h.osaWebhooksHandler.HandleBuildEvent)
		}
		log.Printf("✅ OSA webhook receiver routes registered at /api/osa/webhooks/*")
	}

	// Proxy routes - for embedding external web apps in iframes
	// These routes strip X-Frame-Options and CSP headers to allow embedding
	proxy := api.Group("/proxy")
	{
		proxy.GET("", h.HandleProxyURL)   // GET /api/proxy?url=<encoded-url>
		proxy.POST("", h.HandleProxyPost) // POST /api/proxy with JSON body
	}
	log.Printf("✅ Proxy routes registered at /api/proxy/*")

	// Note: OSA Onboarding routes are registered in cmd/server/main.go
}
