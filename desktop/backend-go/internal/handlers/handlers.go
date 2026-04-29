package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/cache"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/feedback"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/Miosa-osa/OptimalEngine-go"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/sorx"
	"github.com/rhl/businessos-backend/internal/subconscious"
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
	workspaceService        *services.WorkspaceService        // Workspace management
	workspaceVersionService *services.WorkspaceVersionService // Workspace versioning and snapshots
	roleContextService      *services.RoleContextService      // Role-based access control
	memoryHierarchyService  *services.MemoryHierarchyService  // Workspace memory hierarchy (Q1)
	inviteService           *services.WorkspaceInviteService  // Workspace invitation management
	auditService            *services.WorkspaceAuditService   // Workspace audit logging
	projectAccessService    *services.ProjectAccessService    // Project-level access control
	// Voice services (3D Desktop)
	whisperService    *services.WhisperService    // Local speech-to-text
	elevenLabsService *services.ElevenLabsService // Text-to-speech (OSA voice)
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
	osaDeploymentHandler *OSADeploymentHandler             // OSA app deployment handler
	osaAppsHandler       *OSAAppsHandler                   // OSA app management handler (ISR-4)
	osaPromptBuilder     *services.OSAPromptBuilder        // OSA prompt template builder
	// Sandbox services
	sandboxHandler     *SandboxHandler     // Sandbox deployment and management handler
	sandboxEditHandler *SandboxEditHandler // Sandbox module edit lifecycle handler
	// OSA Mode Router - mode classification and SORX engine
	sorxEngine        *sorx.Engine                      // SORX skill execution engine for EXECUTE mode
	protectionService *services.ModuleProtectionService // Module protection for BUILD mode
	// Caching services
	queryCache *cache.QueryCache // Query result caching service for Redis
	// Cognitive session services
	contextTracker    *services.ContextTrackerService // Per-conversation token budget tracking
	modeTransitionSvc *services.ModeTransitionService // OSA mode transition history
	sessionHealthSvc  *services.SessionHealthService  // Session cognitive load metrics
	// Signal Theory feedback
	signalHints          feedback.SignalHintProvider // Homeostatic feedback → prompt corrections
	subconsciousObserver *subconscious.Observer      // Subconscious observer for async pattern detection
	// OptimalOS local data bridge
	optimalHandler *OptimalHandler // read-only bridge to OptimalOS filesystem + engine
	// OptimalOS cloud engine (PostgreSQL-backed, multi-tenant)
	optimalCloudHandler *OptimalCloudHandler

	// OptimalEngine Phase 3-8 — lazily constructed in their register* methods
	// so dev / single-instance deployments without the optimal schemas pay no
	// startup cost.
	wikiHandler       *WikiHandler       // /api/wiki/* — Pages-backed wiki facade
	dsarHandler       *DSARHandler       // /api/compliance/dsar/* — GDPR Art.15 export
	erasureHandler    *ErasureHandler    // /api/compliance/erasure/* — GDPR Art.17 erase
	legalHoldHandler  *LegalHoldHandler  // /api/compliance/holds/* — litigation hold
	retentionHandler  *RetentionHandler  // /api/compliance/retention/* — TTL sweeps
	auditQueryHandler *AuditQueryHandler // /api/audit/* — read-only audit log
	principalHandler  *PrincipalHandler  // /api/identity/principals/* — agents + services
	aclHandler        *ACLHandler        // /api/identity/acls/* — resource ACLs

	// Phase 9 — connector sync orchestration
	integrationRouter *IntegrationRouter // referenced by connectors_sync to access GmailService etc.
	connectorsHandler *ConnectorsHandler // /api/connectors/:kind/sync
}

// SetOptimalHandler injects the OptimalOS handler after construction.
// Call this from main/server setup when the OptimalOS paths are available.
// dbPath is the absolute path to .system/index.db; pass an empty string to
// disable the SQLite-backed graph/search endpoints gracefully.
func (h *Handlers) SetOptimalHandler(nodesRoot, osRoot, enginePath, dbPath string) {
	h.optimalHandler = NewOptimalHandler(nodesRoot, osRoot, enginePath, dbPath)
}

// GetOptimalEngine returns the engine config from the optimal handler, or nil
// if the handler is not initialized. Used by bootstrap for Init/topology loading.
func (h *Handlers) GetOptimalEngine() *optimal.EngineConfig {
	if h.optimalHandler == nil {
		return nil
	}
	return h.optimalHandler.engine
}

// Auth returns the appropriate authentication middleware for this handler set.
// Uses Redis-cached auth when a session cache is configured (horizontal scaling),
// falling back to direct DB auth for single-instance mode.
// Call this to get the auth middleware for route groups defined outside of
// RegisterRoutes (e.g. computer, billing, sync, admin routes in cmd/server/).
func (h *Handlers) Auth() gin.HandlerFunc {
	if h.sessionCache != nil {
		return middleware.CachedAuthMiddleware(h.pool, h.sessionCache)
	}
	return middleware.AuthMiddleware(h.pool)
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
