package handlers

import (
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/cache"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/feedback"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/middleware"
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
	// Ontology / RDF bridge
	bosOntologyService *services.BosOntologyService // bos CLI bridge for RDF operations
	// Zero-Touch Compliance (Innovation 3)
	complianceService *services.ComplianceService // Compliance status, audit trail, gap analysis
	// 2-Phase Commit Transaction Management
	transactionHandler *BOSTransactionHandler // 2PC transaction coordinator (prepare, commit, abort)
	// FIBO Deal Management
	fiboDealsService *services.FIBODealsService // FIBO ontology deal management
}

// NewHandlers creates a new Handlers instance
func NewHandlers(pool *pgxpool.Pool, cfg *config.Config, containerMgr *container.ContainerManager, sessionCache *middleware.SessionCache, terminalPubSub *terminal.TerminalPubSub, embeddingService *services.EmbeddingService, contextBuilder *services.ContextBuilder, tieredContextService *services.TieredContextService, notificationService *services.NotificationService, osaClient *osa.ResilientClient, osaSyncService *services.OSASyncService, fiboDealsService *services.FIBODealsService) *Handlers {
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
		complianceService:    initComplianceService(),
		transactionHandler:   NewBOSTransactionHandler(pool, slog.Default()),
		fiboDealsService:     fiboDealsService,
	}
}

// initComplianceService creates a ComplianceService if OSA_BASE_URL is set.
func initComplianceService() *services.ComplianceService {
	osaBaseURL := os.Getenv("OSA_BASE_URL")
	if osaBaseURL == "" {
		osaBaseURL = "http://localhost:8089"
	}
	return services.NewComplianceService(osaBaseURL, slog.Default())
}
