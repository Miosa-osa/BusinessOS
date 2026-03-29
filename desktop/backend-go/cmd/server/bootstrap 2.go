package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/rhl/businessos-backend/internal/carrier"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/feedback"
	"github.com/rhl/businessos-backend/internal/governance"
	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/middleware"
	redisClient "github.com/rhl/businessos-backend/internal/redis"
	"github.com/rhl/businessos-backend/internal/security"
	"github.com/rhl/businessos-backend/internal/services"
	bossignal "github.com/rhl/businessos-backend/internal/signal"
	"github.com/rhl/businessos-backend/internal/sorx"
	"github.com/rhl/businessos-backend/internal/subconscious"
	"github.com/rhl/businessos-backend/internal/terminal"
	"github.com/rhl/businessos-backend/internal/workers"
)

// AppServices holds all initialized application state. It is the single
// authoritative record of what was started so that main() can shut it all
// down cleanly without relying on closure captures.
type AppServices struct {
	// Core
	cfg        *config.Config
	instanceID string
	router     *gin.Engine
	handlers   *handlers.Handlers

	// Database
	pool        *pgxpool.Pool
	sqlDB       *sql.DB
	dbConnected bool
	dbErr       error

	// Redis
	redisConnected bool

	// Container / sandbox
	containerMgr     *container.ContainerManager
	containerMonitor *container.ContainerMonitor

	sandboxDeploymentService *services.SandboxDeploymentService
	sandboxHealthMonitor     *services.SandboxHealthMonitor
	sandboxCleanupService    *services.SandboxCleanupService
	imageWarmerService       *services.ImageWarmerService

	// Notification / batch
	batchWorker *workers.BatchWorker

	// OSA
	osaQueueWorker *services.OSAQueueWorker

	// Background jobs
	jobsHandler  *handlers.BackgroundJobsHandler
	jobWorkers   []*services.JobWorker
	jobScheduler *services.JobScheduler

	// SORX / CARRIER / Optimal
	sorxScheduler     *sorx.Scheduler
	proactiveConsumer *carrier.ProactiveConsumer
	carrierClient     *carrier.Client
}

// bootstrap initializes every service, constructs the router, and returns the
// fully-wired AppServices struct. It does NOT start the HTTP server.
// Returns (nil, error) only for fatal misconfigurations; non-fatal failures
// are logged and continue with degraded state.
func bootstrap(ctx context.Context) (*AppServices, error) {
	app := &AppServices{}

	// ===== CONFIG =====
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	app.cfg = cfg

	// ===== SECURITY =====
	if err := security.ValidateAndFail(
		cfg.Environment,
		cfg.SecretKey,
		cfg.TokenEncryptionKey,
		cfg.RedisKeyHMACSecret,
	); err != nil {
		return nil, fmt.Errorf("SECURITY ERROR: %w", err)
	}

	if cfg.TokenEncryptionKey != "" {
		if err := security.InitGlobalEncryption(cfg.TokenEncryptionKey); err != nil {
			return nil, fmt.Errorf("failed to initialize token encryption: %w", err)
		}
		log.Printf("Token encryption initialized (AES-256-GCM)")
	} else {
		warnings := security.WarnDevelopmentInsecure(cfg.TokenEncryptionKey, cfg.RedisKeyHMACSecret)
		for _, w := range warnings {
			log.Printf("WARNING: %s", w)
		}
	}

	// ===== INSTANCE ID =====
	app.instanceID = uuid.New().String()[:8]
	log.Printf("Server instance ID: %s", app.instanceID)

	// ===== DATABASE =====
	if cfg.DatabaseRequired {
		pool, err := database.Connect(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		app.pool = pool
		app.dbConnected = true
	} else {
		log.Printf("DATABASE_REQUIRED=false: starting without database (degraded mode)")
		p, err := database.Connect(cfg)
		if err != nil {
			app.dbErr = err
			log.Printf("Database unavailable (continuing): %v", err)
		} else {
			app.pool = p
			app.dbConnected = true
		}
	}

	if app.dbConnected && app.pool != nil {
		app.sqlDB = stdlib.OpenDBFromPool(app.pool)
	}

	// ===== TEMPLATE SYNC =====
	if app.dbConnected && app.pool != nil && os.Getenv("SYNC_TEMPLATES_ON_STARTUP") == "true" {
		slog.Info("syncing YAML templates to database")
		templatesDir := os.Getenv("TEMPLATES_DIR")
		if templatesDir == "" {
			templatesDir = "internal/prompts/templates/osa"
		}
		syncSvc := services.NewTemplateSyncService(app.pool, slog.Default(), templatesDir)
		syncCtx, cancelSync := context.WithTimeout(ctx, 30*time.Second)
		defer cancelSync()
		if result, err := syncSvc.SyncTemplates(syncCtx); err != nil {
			slog.Warn("template sync failed", "error", err)
		} else {
			slog.Info("template sync completed",
				"inserted", result.Inserted,
				"updated", result.Updated,
				"errors", len(result.Errors),
			)
		}
	}

	// ===== REDIS =====
	var sessionCache *middleware.SessionCache
	var terminalPubSub *terminal.TerminalPubSub

	if cfg.RedisURL != "" {
		redisCfg := redisClient.DefaultConfig()
		redisCfg.URL = cfg.RedisURL
		redisCfg.Password = cfg.RedisPassword
		redisCfg.TLSEnabled = cfg.RedisTLSEnabled
		if !cfg.IsProduction() && cfg.RedisTLSEnabled {
			redisCfg.TLSInsecure = true
		}

		if err := redisClient.Connect(ctx, redisCfg); err != nil {
			log.Printf("Warning: Redis unavailable: %v", err)
			log.Printf("Sessions will use direct DB auth (not optimal for horizontal scaling)")
		} else {
			app.redisConnected = true
			log.Printf("Redis connected successfully")

			sessionCacheConfig := &middleware.SessionCacheConfig{
				KeyPrefix:  "auth_session:",
				TTL:        15 * time.Minute,
				HMACSecret: cfg.RedisKeyHMACSecret,
			}
			sc, err := middleware.NewSessionCache(redisClient.Client(), sessionCacheConfig)
			if err != nil {
				log.Printf("Warning: Session cache initialization failed: %v", err)
				log.Printf("Sessions will use direct DB auth (not optimal for horizontal scaling)")
			} else {
				sessionCache = sc
				log.Printf("Session cache enabled (TTL=15m, HMAC-secured keys)")
			}

			terminalPubSub = terminal.NewTerminalPubSub(redisClient.Client(), app.instanceID)
			log.Printf("Terminal pub/sub enabled (instance=%s)", app.instanceID)
		}
	}

	// ===== CONTAINER MANAGER =====
	containerMgr, err := container.NewContainerManager(ctx, "businessos-workspace:latest")
	if err != nil {
		log.Printf("Warning: Container manager unavailable: %v", err)
		log.Printf("Terminal will use local PTY mode")
	} else {
		app.containerMgr = containerMgr
		log.Printf("Container manager initialized successfully")

		containerMonitor := container.NewContainerMonitor(containerMgr, nil)
		if err := containerMonitor.StartMonitoring(ctx); err != nil {
			log.Printf("Warning: Container monitor failed to start: %v", err)
		} else {
			app.containerMonitor = containerMonitor
			log.Printf("Container monitor started (cleanup=%v, idle_timeout=%v)",
				container.DefaultMonitorConfig().CleanupInterval,
				container.DefaultMonitorConfig().IdleTimeout)
		}
	}

	// ===== SANDBOX SERVICES =====
	var sandboxDeploymentService *services.SandboxDeploymentService
	var sandboxHealthMonitor *services.SandboxHealthMonitor
	var sandboxCleanupService *services.SandboxCleanupService
	var sandboxIntegrationService *services.SandboxIntegrationService
	var imageWarmerService *services.ImageWarmerService
	var sandboxQuotaService *services.QuotaService

	dockerClient, dockerErr := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if dockerErr == nil {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, pingErr := dockerClient.Ping(pingCtx)
		cancel()

		if pingErr == nil && app.dbConnected && app.pool != nil {
			slog.Info("initializing sandbox services (Docker available)")

			sandboxDeploymentService, err = services.NewSandboxDeploymentService(app.pool, dockerClient, cfg, slog.Default())
			if err != nil {
				slog.Warn("failed to initialize sandbox deployment service", "error", err)
			} else {
				slog.Info("sandbox deployment service initialized",
					"port_range", fmt.Sprintf("%d-%d", cfg.SandboxPortMin, cfg.SandboxPortMax))

				appContainerMgr := container.NewAppContainerManager(dockerClient, slog.Default(), "")
				healthConfig := services.DefaultHealthMonitorConfig()
				sandboxHealthMonitor = services.NewSandboxHealthMonitor(appContainerMgr, app.pool, slog.Default(), healthConfig)

				if err := sandboxHealthMonitor.Start(ctx); err != nil {
					slog.Warn("failed to start sandbox health monitor", "error", err)
				} else {
					slog.Info("sandbox health monitor started",
						"check_interval", healthConfig.CheckInterval,
						"auto_restart", healthConfig.AutoRestart)
				}

				portAllocator, err := services.NewSandboxPortAllocator(nil, app.pool, cfg, slog.Default())
				if err != nil {
					slog.Warn("failed to initialize port allocator for cleanup", "error", err)
				} else {
					cleanupConfig := services.DefaultCleanupConfig()
					sandboxCleanupService = services.NewSandboxCleanupService(appContainerMgr, portAllocator, app.pool, slog.Default(), cleanupConfig)

					if err := sandboxCleanupService.Start(ctx); err != nil {
						slog.Warn("failed to start sandbox cleanup service", "error", err)
					} else {
						slog.Info("sandbox cleanup service started",
							"cleanup_interval", cleanupConfig.CleanupInterval,
							"stopped_grace_period", cleanupConfig.StoppedGracePeriod)
					}
				}

				sandboxIntegrationService = services.NewSandboxIntegrationService(sandboxDeploymentService, app.pool, cfg, slog.Default())
				slog.Info("sandbox integration service initialized (auto-deploy enabled)")

				imageWarmerService, err = services.NewImageWarmerService(slog.Default(), 24*time.Hour)
				if err != nil {
					slog.Warn("failed to initialize image warmer service", "error", err)
				} else {
					if err := imageWarmerService.Start(ctx); err != nil {
						slog.Warn("failed to start image warmer service", "error", err)
					} else {
						slog.Info("image warmer service started (pre-pulling common images)")
					}
				}

				sandboxQuotaService = services.NewQuotaService(app.pool, cfg, slog.Default())
				slog.Info("sandbox quota service initialized")

				_ = sandboxIntegrationService
				_ = sandboxQuotaService
			}
		} else if pingErr != nil {
			slog.Info("Docker daemon not available, sandbox services disabled", "error", pingErr)
		} else {
			slog.Info("sandbox services disabled (database not available)")
		}
	} else {
		slog.Info("Docker not available, sandbox services disabled", "error", dockerErr)
	}

	app.sandboxDeploymentService = sandboxDeploymentService
	app.sandboxHealthMonitor = sandboxHealthMonitor
	app.sandboxCleanupService = sandboxCleanupService
	app.imageWarmerService = imageWarmerService

	// ===== ROUTER + MIDDLEWARE =====
	router := gin.Default()
	app.router = router

	router.Use(func(c *gin.Context) {
		c.Request.Body = httpMaxBytesReader(c.Writer, c.Request.Body, 10<<20)
		c.Next()
	})
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(middleware.SecurityHeaders(cfg))
	log.Printf("Security headers enabled (X-Frame-Options, CSP, HSTS, etc.)")

	csrfConfig := buildCSRFConfig(cfg)
	router.Use(middleware.CSRF(csrfConfig))
	log.Printf("CSRF protection enabled (double submit cookie pattern; excluded: webhooks, osa/webhooks, sorx/callback, internal/osa, health)")

	globalRateLimiter := middleware.GetGlobalHTTPRateLimiter()
	router.Use(middleware.RateLimitMiddleware(globalRateLimiter))
	log.Printf("Rate limiting enabled (100 req/s per IP, 200 req/s per user)")

	// ===== SERVICES (require DB) =====
	// These are only initialized when DB is available. If not, we return early
	// from main() after registering a degraded status endpoint.

	if !app.dbConnected || app.pool == nil {
		// Caller will detect this and run degraded-mode server.
		return app, nil
	}

	// ===== EMBEDDING / RAG =====
	embeddingService := services.NewEmbeddingService(app.pool, cfg.OllamaLocalURL)
	var contextBuilder *services.ContextBuilder
	var tieredContextService *services.TieredContextService

	if embeddingService.HealthCheck(ctx) {
		contextBuilder = services.NewContextBuilder(app.pool, embeddingService)
		summarizer := services.NewSummarizerService(app.pool, cfg)
		tieredContextService = services.NewTieredContextService(app.pool, embeddingService, summarizer)
		log.Printf("Embedding service initialized (model=nomic-embed-text, dimensions=768)")
		log.Printf("Tiered context service enabled (scoped RAG, Level 1/2/3 context)")
	} else {
		log.Printf("Warning: Embedding service unavailable (Ollama not running or nomic-embed-text model not pulled)")
		log.Printf("RAG features will be disabled. Run: ollama pull nomic-embed-text")
		embeddingService = nil
	}

	// ===== NOTIFICATIONS =====
	sseBroadcaster := services.NewSSEBroadcaster()
	notificationService := services.NewNotificationService(app.pool, sseBroadcaster)
	log.Printf("Notification service initialized (SSE real-time enabled)")

	batchWorker := workers.NewBatchWorker(app.pool, notificationService.Dispatcher())
	go batchWorker.Start(ctx)
	app.batchWorker = batchWorker
	log.Printf("Notification batch worker started (interval: 10s)")
	log.Printf("App generation worker DISABLED (OSAQueueWorker handles queue with SSE events)")

	// ===== WEB PUSH =====
	var webPushService *services.WebPushService
	if cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != "" {
		webPushService = services.NewWebPushService(app.pool, &services.WebPushConfig{
			VAPIDPublicKey:  cfg.VAPIDPublicKey,
			VAPIDPrivateKey: cfg.VAPIDPrivateKey,
			VAPIDContact:    cfg.VAPIDContact,
		})
		log.Printf("Web Push service initialized (VAPID keys configured)")
	} else {
		log.Printf("Web Push service disabled (VAPID keys not configured)")
	}

	// ===== PEDRO TASKS SERVICES =====
	var documentProcessor *services.DocumentProcessor
	if embeddingService != nil {
		documentProcessor = services.NewDocumentProcessor(app.pool, embeddingService, "./uploads/documents")
		log.Printf("Document processor initialized (chunking + semantic search)")
	}

	learningService := services.NewLearningService(app.pool)
	log.Printf("Learning service initialized (feedback + personalization)")

	var memoryService *services.MemoryService
	if embeddingService != nil {
		memoryService = services.NewMemoryService(app.pool, embeddingService)
		log.Printf("Memory service initialized (memory persistence)")
	}

	var autoLearningTriggers *services.AutoLearningTriggers
	if learningService != nil && memoryService != nil && embeddingService != nil {
		autoLearningTriggers = services.NewAutoLearningTriggers(learningService, memoryService, embeddingService)
		log.Printf("Auto-learning triggers initialized (automatic pattern extraction)")
	}

	var promptPersonalizer *services.PromptPersonalizer
	if app.pool != nil && learningService != nil && memoryService != nil && embeddingService != nil {
		promptPersonalizer = services.NewPromptPersonalizer(app.pool, learningService, memoryService, embeddingService)
		log.Printf("Prompt personalizer initialized (user-specific context injection)")
	}

	// ===== ADVANCED RAG SERVICES =====
	var hybridSearchService *services.HybridSearchService
	if app.pool != nil && embeddingService != nil {
		hybridSearchService = services.NewHybridSearchService(app.pool, embeddingService)
		log.Printf("Hybrid search service initialized (semantic + keyword with RRF)")
	}

	var rerankerService *services.ReRankerService
	if app.pool != nil && embeddingService != nil {
		rerankerService = services.NewReRankerService(app.pool, embeddingService)
		log.Printf("Re-ranker service initialized (multi-signal relevance scoring)")
	}

	var agenticRAGService *services.AgenticRAGService
	if app.pool != nil && hybridSearchService != nil && rerankerService != nil && embeddingService != nil && learningService != nil {
		agenticRAGService = services.NewAgenticRAGService(app.pool, hybridSearchService, rerankerService, embeddingService, learningService)
		log.Printf("Agentic RAG service initialized (intelligent adaptive retrieval)")
	}

	// ===== EMBEDDING CACHE =====
	var embeddingCacheAdapter *services.EmbeddingCacheAdapter
	if app.redisConnected && redisClient.Client() != nil {
		embeddingCacheConfig := services.DefaultEmbeddingCacheConfig()
		embeddingCache := services.NewEmbeddingCacheService(redisClient.Client(), app.pool, embeddingCacheConfig)
		embeddingCacheAdapter = services.NewEmbeddingCacheAdapter(embeddingCache)

		if embeddingCache.IsEnabled() {
			log.Printf("Embedding cache service initialized (24h text, 48h images)")
			if embeddingService != nil {
				embeddingService.SetEmbeddingCache(embeddingCacheAdapter)
				log.Printf("Embedding service now using dedicated embedding cache")
			}
		} else {
			log.Printf("Embedding cache disabled (Redis unavailable)")
		}
	}

	if app.redisConnected && redisClient.Client() != nil {
		cacheConfig := services.DefaultRAGCacheConfig()
		ragCache := services.NewRAGCacheService(redisClient.Client(), cacheConfig)
		log.Printf("RAG cache service initialized (15min queries, 24hr embeddings - legacy)")

		if embeddingService != nil {
			embeddingService.SetCache(ragCache)
			log.Printf("Embedding service legacy cache enabled (fallback)")
		}
		if agenticRAGService != nil {
			agenticRAGService.SetCache(ragCache)
			log.Printf("Agentic RAG cache enabled")
		}
	} else {
		log.Printf("RAG cache disabled (Redis not available)")
	}

	queryExpansion := services.NewQueryExpansionService(nil)
	log.Printf("Query expansion service initialized (60+ synonym mappings)")
	if queryExpansion != nil && agenticRAGService != nil {
		agenticRAGService.SetQueryExpansion(queryExpansion)
		log.Printf("Agentic RAG query expansion enabled")
	}

	// ===== MULTI-MODAL SEARCH =====
	var imageEmbeddingService *services.ImageEmbeddingService
	var multiModalSearchService *services.MultiModalSearchService

	clipProvider := os.Getenv("CLIP_PROVIDER")
	if clipProvider == "" {
		clipProvider = "local"
	}
	imageEmbedConfig := services.ImageEmbeddingConfig{
		Provider:     clipProvider,
		APIKey:       os.Getenv("CLIP_API_KEY"),
		ModelName:    "clip-vit-base-patch32",
		Dimensions:   512,
		LocalBaseURL: os.Getenv("CLIP_LOCAL_URL"),
	}

	if clipProvider == "local" && imageEmbedConfig.LocalBaseURL == "" {
		log.Printf("Image embedding service disabled (CLIP_LOCAL_URL not set)")
		log.Printf("To enable: Set CLIP_LOCAL_URL=http://localhost:8000 and run CLIP server")
	} else if (clipProvider == "openai" || clipProvider == "replicate") && imageEmbedConfig.APIKey == "" {
		log.Printf("Image embedding service disabled (CLIP_API_KEY not set for %s)", clipProvider)
	} else {
		imageEmbeddingService = services.NewImageEmbeddingService(app.pool, imageEmbedConfig)
		log.Printf("Image embedding service initialized (provider=%s, model=%s)", clipProvider, imageEmbedConfig.ModelName)

		if embeddingCacheAdapter != nil {
			imageEmbeddingService.SetEmbeddingCache(embeddingCacheAdapter)
			log.Printf("Image embedding service cache enabled (48h TTL)")
		}

		if hybridSearchService != nil && rerankerService != nil && embeddingService != nil {
			multiModalSearchService = services.NewMultiModalSearchService(
				app.pool,
				hybridSearchService,
				rerankerService,
				imageEmbeddingService,
				embeddingService,
			)
			log.Printf("Multi-modal search service initialized (text + image + cross-modal)")
			log.Printf("Feature 7 (Multi-modal Embeddings) complete: SearchWithImage ready!")
		}
	}

	// ===== APP PROFILER + CONVERSATION INTELLIGENCE + MEMORY EXTRACTOR =====
	appProfilerService := services.NewAppProfilerService(app.pool)
	log.Printf("App profiler service initialized (codebase analysis)")

	var conversationIntelligence *services.ConversationIntelligenceService
	if embeddingService != nil {
		conversationIntelligence = services.NewConversationIntelligenceService(app.pool, embeddingService)
		log.Printf("Conversation intelligence initialized (analysis + summarization)")
	}

	var memoryExtractor *services.MemoryExtractorService
	if embeddingService != nil {
		memoryExtractor = services.NewMemoryExtractorService(app.pool, embeddingService)
		if cfg.GroqAPIKey != "" {
			groqLLM := services.NewGroqService(cfg, "llama-3.1-8b-instant")
			if groqLLM.HealthCheck(ctx) {
				memoryExtractor.SetLLMService(groqLLM)
				log.Printf("Memory extractor initialized with LLM-enhanced extraction (Groq llama-3.1-8b-instant)")
			} else {
				log.Printf("Memory extractor initialized (regex-only, Groq unavailable)")
			}
		} else {
			log.Printf("Memory extractor initialized (regex-only, no Groq API key)")
		}
	}

	var blockMapper *services.BlockMapperService
	if app.sqlDB != nil {
		blockMapper = services.NewBlockMapperService(app.sqlDB, slog.Default())
		log.Printf("Block mapper initialized (markdown to structured blocks)")
	}

	// ===== OSA INTEGRATION =====
	var osaClient *osa.ResilientClient
	var osaSyncService *services.OSASyncService
	var osaFileSyncService *services.OSAFileSyncService
	var osaWorkspaceInitService *services.OSAWorkspaceInitService
	var osaWorkflowsHandler *handlers.OSAWorkflowsHandler
	var osaWebhooksHandler *handlers.OSAWebhooksHandler
	var osaBuildEventBus *services.BuildEventBus
	var osaStreamingHandler *handlers.OSAStreamingHandler
	var osaDeploymentService *services.AppDeploymentService
	var osaDeploymentHandler *handlers.OSADeploymentHandler
	var osaQueueWorker *services.OSAQueueWorker

	if cfg.OSAEnabled {
		osaConfig := &osa.ResilientClientConfig{
			OSAConfig:            cfg.OSA,
			CircuitBreakerConfig: osa.DefaultCircuitBreakerConfig(),
			FallbackStrategy:     osa.FallbackStale,
			CacheTTL:             5 * time.Minute,
			HealthCheckCacheTTL:  30 * time.Second,
			QueueSize:            1000,
			EnableAutoRecovery:   true,
		}

		osaClientInst, err := osa.NewResilientClient(osaConfig)
		if err != nil {
			log.Printf("Failed to create OSA client: %v", err)
		} else {
			osaClient = osaClientInst
			log.Printf("OSA client initialized (base_url=%s)", cfg.OSA.BaseURL)

			healthResp, healthErr := osaClient.HealthCheck(context.Background())
			if healthErr != nil {
				log.Printf("OSA health check failed at %s: %v (will use fallback agents)", cfg.OSA.BaseURL, healthErr)
			} else {
				log.Printf("OSA connected at %s (status=%s)", cfg.OSA.BaseURL, healthResp.Status)
			}

			syncSvc, err := services.NewOSASyncService(app.pool, cfg)
			if err != nil {
				log.Printf("Failed to create OSA sync service: %v", err)
			} else {
				osaSyncService = syncSvc
				log.Printf("OSA sync service initialized (transactional outbox pattern)")
			}

			osaWorkspacePath := os.Getenv("OSA_WORKSPACE_PATH")
			if osaWorkspacePath == "" {
				osaWorkspacePath = filepath.Join(os.TempDir(), "osa-workspaces")
				slog.Warn("OSA_WORKSPACE_PATH not set, using temp directory", "path", osaWorkspacePath)
			}
			osaFileSyncService = services.NewOSAFileSyncService(app.pool, slog.Default(), osaWorkspacePath)
			osaWorkspaceInitService = services.NewOSAWorkspaceInitService(app.pool, slog.Default())

			osaBuildEventBus = services.NewBuildEventBus(slog.Default())
			log.Printf("OSA build event bus initialized")

			osaWorkflowsHandler = handlers.NewOSAWorkflowsHandler(app.pool, osaFileSyncService)
			osaWebhooksHandler, err = handlers.NewOSAWebhooksHandler(app.pool, cfg.OSA.SharedSecret, osaBuildEventBus, cfg.IsProduction())
			if err != nil {
				return nil, fmt.Errorf("failed to initialize OSA webhooks handler: %w", err)
			}
			osaStreamingHandler = handlers.NewOSAStreamingHandler(osaBuildEventBus, slog.Default())

			log.Printf("OSA file sync service initialized (workspace=%s)", osaWorkspacePath)

			go osaFileSyncService.Start(ctx)
			log.Printf("OSA file sync service started (polling every 30s)")

			osaPromptBuilder, err := services.NewOSAPromptBuilder(app.pool, slog.Default())
			if err != nil {
				log.Printf("Warning: Failed to initialize OSA prompt builder: %v", err)
			}
			osaQueries := sqlc.New(app.pool)
			osaQueueWorker = services.NewOSAQueueWorker(app.pool, osaQueries, osaClient, osaBuildEventBus, osaPromptBuilder)
			go osaQueueWorker.Start(ctx)
			log.Printf("OSA queue worker started (polling every 5s)")

			osaWorkspaceRoot := os.Getenv("OSA_DEPLOYMENT_ROOT")
			if osaWorkspaceRoot == "" {
				osaWorkspaceRoot = "/tmp/businessos-apps"
			}
			osaDeploymentService = services.NewAppDeploymentService(app.pool, slog.Default(), osaWorkspaceRoot)
			osaDeploymentHandler = handlers.NewOSADeploymentHandler(osaDeploymentService)
			log.Printf("OSA deployment service initialized (workspace=%s)", osaWorkspaceRoot)

			osaFileSyncService.SetDeploymentService(osaDeploymentService)
			log.Printf("Auto-deployment enabled - new workflows will deploy automatically")
		}
	}

	// OSA deployment service: also available independently of OSA-5 connection.
	if osaDeploymentService == nil {
		osaWorkspaceRoot := os.Getenv("OSA_DEPLOYMENT_ROOT")
		if osaWorkspaceRoot == "" {
			osaWorkspaceRoot = "/tmp/businessos-apps"
		}
		osaDeploymentService = services.NewAppDeploymentService(app.pool, slog.Default(), osaWorkspaceRoot)
		osaDeploymentHandler = handlers.NewOSADeploymentHandler(osaDeploymentService)
		log.Printf("OSA deployment service initialized independently (workspace=%s)", osaWorkspaceRoot)
	}

	app.osaQueueWorker = osaQueueWorker

	// ===== HANDLERS =====
	h := handlers.NewHandlers(app.pool, cfg, app.containerMgr, sessionCache, terminalPubSub, embeddingService, contextBuilder, tieredContextService, notificationService, osaClient, osaSyncService)
	app.handlers = h

	if webPushService != nil {
		h.SetWebPushService(webPushService)
	}

	emailService := services.NewEmailService()
	if emailService.IsEnabled() {
		h.SetEmailService(emailService)
		log.Printf("Email service initialized (Resend configured)")
	} else {
		log.Printf("Email service disabled (RESEND_API_KEY not set)")
	}

	queries := sqlc.New(app.pool)
	commentService := services.NewCommentService(app.pool, queries, notificationService)
	h.SetCommentService(commentService)
	log.Printf("Comment service initialized (mentions & notifications enabled)")

	h.SetPedroServices(documentProcessor, learningService, autoLearningTriggers, promptPersonalizer, appProfilerService, conversationIntelligence, memoryExtractor, blockMapper)
	log.Printf("Pedro Tasks services registered (documents, learning, auto-learning, personalizer, profiles, intelligence, blocks)")

	h.SetRAGServices(hybridSearchService, rerankerService, agenticRAGService, memoryService)
	log.Printf("RAG services registered (hybrid search, re-ranker, agentic RAG, memory)")

	if multiModalSearchService != nil && imageEmbeddingService != nil {
		h.SetMultiModalServices(multiModalSearchService, imageEmbeddingService)
		log.Printf("Multi-modal services registered (image embeddings, text+image search)")
	}

	whisperService := services.NewWhisperService()
	elevenLabsService := services.NewElevenLabsService()
	h.SetVoiceServices(whisperService, elevenLabsService)
	if whisperService.IsAvailable() {
		log.Printf("Whisper service initialized (local speech-to-text)")
	} else {
		log.Printf("Whisper service not fully configured (model/binary not found)")
	}
	if elevenLabsService.IsConfigured() {
		log.Printf("ElevenLabs service initialized (OSA voice enabled)")
	} else {
		log.Printf("ElevenLabs service not configured (API key/voice ID not set)")
	}

	if sandboxDeploymentService != nil {
		sandboxHandler := handlers.NewSandboxHandler(sandboxDeploymentService, slog.Default())
		h.SetSandboxHandler(sandboxHandler)
		log.Printf("Sandbox handler initialized (deploy, start, stop, delete endpoints)")
	}

	sandboxEditSvc := services.NewSandboxEditService(slog.Default())
	sandboxEditHandler := handlers.NewSandboxEditHandler(sandboxEditSvc, slog.Default())
	h.SetSandboxEditHandler(sandboxEditHandler)
	slog.Info("sandbox edit handler initialized", "routes", "/api/v1/sandbox/edit/*")

	workspaceService := services.NewWorkspaceService(app.pool)
	h.SetWorkspaceService(workspaceService)
	log.Printf("Workspace service registered (workspaces, members, roles)")

	workspaceVersionService := services.NewWorkspaceVersionService(app.pool, slog.Default())
	h.SetWorkspaceVersionService(workspaceVersionService)
	log.Printf("Workspace version service registered (snapshots, restore)")

	if osaQueueWorker != nil {
		osaQueueWorker.SetVersionService(workspaceVersionService)
		log.Printf("Workspace version service wired to queue worker (auto-snapshots)")
	}

	if osaDeploymentHandler != nil {
		h.SetOSADeploymentHandler(osaDeploymentHandler)
		log.Printf("OSA deployment handler registered")
	}

	if osaClient != nil {
		log.Printf("OSA integration enabled (API endpoints at /api/osa/*)")
		if osaFileSyncService != nil && osaWorkflowsHandler != nil && osaWebhooksHandler != nil {
			h.SetOSAFileServices(osaFileSyncService, osaWorkspaceInitService, osaWorkflowsHandler, osaWebhooksHandler, osaBuildEventBus, osaStreamingHandler, osaDeploymentHandler)
			log.Printf("OSA file sync, workflow, and streaming services registered")
		}
		if osaQueueWorker != nil {
			h.SetOSAQueueWorker(osaQueueWorker)
		}
		if promptBuilder, err := services.NewOSAPromptBuilder(app.pool, slog.Default()); err != nil {
			log.Printf("Warning: Failed to initialize OSA template loader: %v", err)
		} else {
			h.SetOSAPromptBuilder(promptBuilder)
			log.Printf("OSA template loader initialized (template-based app generation)")
		}
	}

	// ===== SORX ENGINE =====
	sorxEngine := sorx.NewEngine(app.pool, slog.Default())
	h.SetSorxEngine(sorxEngine)
	log.Printf("SORX engine initialized (skill execution for EXECUTE mode)")

	// ===== CARRIER =====
	var carrierClient *carrier.Client
	{
		carrierCfg, err := carrier.ConfigFromEnv(carrier.DefaultConfig())
		if err != nil {
			return nil, fmt.Errorf("carrier config: %w", err)
		}
		if carrierCfg.Enabled {
			cc, err := carrier.NewClient(carrierCfg, slog.Default())
			if err != nil {
				log.Printf("Warning: CARRIER connection failed: %v", err)
				log.Printf("SORX will operate in local-only mode (no SorxMain reasoning)")
			} else {
				carrierClient = cc
				log.Printf("CARRIER connected (exchange=%s, reply_queue=sorx.responses.%s)", carrierCfg.Exchange, carrierCfg.OSInstanceID)
			}
		} else {
			log.Printf("CARRIER disabled — SORX operating in local-only mode")
		}
		sorxEngine.SetCarrierClient(carrierClient)
	}
	app.carrierClient = carrierClient

	// ===== OPTIMAL INTEGRATION =====
	var proactiveConsumer *carrier.ProactiveConsumer
	{
		optimalCfg, err := config.OptimalConfigFromEnv(config.DefaultOptimalConfig())
		if err != nil {
			return nil, fmt.Errorf("optimal config: %w", err)
		}

		if optimalCfg.Enabled && carrierClient != nil {
			regCtx, regCancel := context.WithTimeout(ctx, 10*time.Second)
			if err := carrierClient.RegisterWithOptimal(regCtx, optimalCfg.InstalledModules, optimalCfg.Capabilities); err != nil {
				log.Printf("Warning: Optimal registration failed: %v", err)
			} else {
				log.Printf("Registered with Optimal (template=%s, modules=%v)", optimalCfg.TemplateType, optimalCfg.InstalledModules)
			}
			regCancel()

			carrierClient.StartHeartbeat(ctx, optimalCfg.HeartbeatInterval)
			log.Printf("Optimal heartbeat started (interval=%s)", optimalCfg.HeartbeatInterval)

			carrierCfg, _ := carrier.ConfigFromEnv(carrier.DefaultConfig())
			actionHandler := func(handlerCtx context.Context, cmd carrier.ActionCommand) (any, error) {
				return sorxEngine.ExecuteAction(handlerCtx, cmd.Action, cmd.Params)
			}
			pc, err := carrier.NewProactiveConsumer(carrierCfg, actionHandler, slog.Default())
			if err != nil {
				log.Printf("Warning: Proactive consumer failed to start: %v", err)
			} else {
				proactiveConsumer = pc
				pc.Start()
				log.Printf("Optimal proactive consumer started (command_queue=sorx.commands.%s)", carrierCfg.OSInstanceID)
			}
		} else if optimalCfg.Enabled && carrierClient == nil {
			log.Printf("Warning: OPTIMAL_ENABLED=true but CARRIER is not connected; skipping Optimal integration")
		} else {
			log.Printf("Optimal integration disabled (OPTIMAL_ENABLED=false or mode=local)")
		}
	}
	app.proactiveConsumer = proactiveConsumer

	// ===== SORX SCHEDULER =====
	if app.pool != nil {
		sorxScheduler := sorx.NewScheduler(sorxEngine, app.pool, slog.Default())
		if err := sorxScheduler.Start(); err != nil {
			log.Printf("Warning: SORX scheduler failed to start: %v", err)
			log.Printf("Proactive skill execution disabled; skills must be triggered manually")
		} else {
			app.sorxScheduler = sorxScheduler
			log.Printf("SORX scheduler started (proactive skill execution enabled)")
		}
	}

	protectionService := services.NewModuleProtectionService(app.pool, slog.Default())
	h.SetProtectionService(protectionService)
	log.Printf("Module protection service initialized (BUILD mode safety)")

	// ===== COGNITIVE SESSION SERVICES =====
	contextTracker := services.NewContextTrackerService(app.pool, slog.Default(), nil)
	h.SetContextTracker(contextTracker)
	log.Printf("Context tracker initialized (per-conversation token budget)")

	modeTransitionSvc := services.NewModeTransitionService(app.pool, slog.Default())
	h.SetModeTransitionService(modeTransitionSvc)
	log.Printf("Mode transition service initialized (OSA mode history)")

	sessionHealthSvc := services.NewSessionHealthService(modeTransitionSvc, slog.Default())
	h.SetSessionHealthService(sessionHealthSvc)
	log.Printf("Session health service initialized (cognitive load metrics)")

	// ===== SIGNAL THEORY / HOMEOSTATIC FEEDBACK LOOP =====
	promptActuator := feedback.NewPromptActuator(10*time.Minute, slog.Default())

	reEncodingMetric := subconscious.NewInMemoryReEncoding(30 * time.Minute)
	signalBounceMetric := subconscious.NewInMemorySignalBounce(30 * time.Minute)
	genreRecognitionMetric := subconscious.NewInMemoryGenreRecognition(30 * time.Minute)
	feedbackClosureMetric := subconscious.NewInMemoryFeedbackClosure(30 * time.Minute)

	metricsRegistry := &bossignal.ProxyMetricsRegistry{
		ActionCompletion: feedback.NewSignalLogActionCompletion(app.pool),
		ReEncoding:       reEncodingMetric,
		SignalBounce:     signalBounceMetric,
		GenreRecognition: genreRecognitionMetric,
		FeedbackClosure:  feedbackClosureMetric,
	}

	setpointStore := feedback.NewInMemorySetpointStore()

	autopoieticMonitor, err := feedback.NewAutopoieticMonitor(ctx, feedback.AutopoieticMonitorConfig{
		Logger: slog.Default(),
	})
	if err != nil {
		log.Printf("Warning: AutopoieticMonitor failed to initialize: %v", err)
		autopoieticMonitor = nil
	}

	var redisUniversal redis.UniversalClient
	if app.redisConnected && redisClient.Client() != nil {
		redisUniversal = redisClient.Client()
	}
	selfImprovementEngine := feedback.NewSelfImprovementEngine(redisUniversal, slog.Default())

	metricEmitter := subconscious.NewMetricEmitter(
		reEncodingMetric, signalBounceMetric, genreRecognitionMetric, feedbackClosureMetric,
		autopoieticMonitor, slog.Default(),
	)

	blockStore := subconscious.NewPostgresBlockStore(app.pool)
	blockAccumulator := subconscious.NewBlockAccumulator(blockStore, selfImprovementEngine, slog.Default())
	signalClassifier := subconscious.NewSignalClassifier(cfg, app.pool, slog.Default())
	patternExtractor := subconscious.NewPatternExtractor()

	subconsciousObserver := subconscious.NewObserver(
		signalClassifier, patternExtractor, metricEmitter, blockAccumulator, slog.Default(),
	)
	h.SetSubconsciousObserver(subconsciousObserver)

	subconsciousHintProvider := subconscious.NewSubconsciousHintProvider(promptActuator, blockStore, slog.Default())
	h.SetSignalHints(subconsciousHintProvider)

	// ===== L7: ALGEDONIC CHANNEL (created before homeo loop for callback wiring) =====
	algedonicLogging := governance.NewLoggingHandler(slog.Default())
	algedonicPostgres := governance.NewPostgresHandler(app.pool, slog.Default())
	algedonicChannel := governance.NewAlgedonicChannel(slog.Default(), algedonicLogging, algedonicPostgres)
	log.Printf("Algedonic channel initialized (logging + postgres handlers)")

	homeoLoop, err := feedback.NewHomeostaticLoop(ctx, feedback.HomeostaticLoopConfig{
		Metrics:   metricsRegistry,
		Setpoints: setpointStore,
		Actuator:  promptActuator,
		Interval:  30 * time.Second,
		Logger:    slog.Default(),
		OnPersistentFailure: func(ctx context.Context, source string, failures int) {
			algedonicChannel.Fire(ctx, governance.AlgedonicSignal{
				Source:      "homeostatic_loop/" + source,
				Description: fmt.Sprintf("Persistent correction failure: %d consecutive failures on %s", failures, source),
				Severity:    governance.AlgedonicMedium,
				Metadata:    map[string]any{"metric": source, "consecutive_failures": failures},
			})
		},
	})
	if err != nil {
		log.Printf("Warning: Homeostatic loop failed to initialize: %v", err)
	} else {
		go func() {
			if err := homeoLoop.Run(ctx); err != nil {
				slog.Error("homeostatic loop exited", "error", err)
			}
		}()
		log.Printf("Homeostatic feedback loop started (30s interval, all 6 metrics active)")
	}
	log.Printf("Subconscious Observer initialized (async classify → extract → emit → accumulate)")
	log.Printf("SubconsciousHintProvider wired (PromptActuator + memory blocks → agent prompts)")

	// ===== L6: TIME TO DECIDE METRIC =====
	timeToDecideMetric := subconscious.NewInMemoryTimeToDecide(30 * time.Minute)
	metricEmitter.SetTimeToDecide(timeToDecideMetric)
	metricsRegistry.TimeToDecide = timeToDecideMetric
	log.Printf("TimeToDecide metric initialized (30min sliding window)")

	// ===== L6: FAILURE DETECTORS =====
	bandwidthDetector := bossignal.NewBandwidthOverloadDetector(10.0)
	feedbackDetector := bossignal.NewFeedbackFailureDetector(5 * time.Minute)
	routingDetector := bossignal.NewRoutingFailureDetector()
	genreMismatchDetector := bossignal.NewGenreMismatchDetector()
	bridgeDetector := bossignal.NewBridgeFailureDetector()

	_, err = bossignal.NewFailureDetectorRegistry(ctx,
		routingDetector,       // Shannon: routing
		bandwidthDetector,     // Shannon: bandwidth
		nil,                   // Shannon: fidelity (not yet implemented)
		genreMismatchDetector, // Ashby: genre mismatch
		nil,                   // Ashby: variety (not yet implemented)
		nil,                   // Ashby: structure (not yet implemented)
		bridgeDetector,        // Beer: bridge
		nil,                   // Beer: herniation (not yet implemented)
		nil,                   // Beer: decay (not yet implemented)
		feedbackDetector,      // Wiener: feedback
		nil,                   // Adversarial (not yet implemented)
	)
	if err != nil {
		log.Printf("Warning: FailureDetectorRegistry failed: %v", err)
	} else {
		log.Printf("FailureDetectorRegistry initialized (5/11 detectors active)")
	}

	// ===== L6: DOUBLE-LOOP CONTROLLER =====
	doubleLoop := feedback.NewDoubleLoopController(feedback.DoubleLoopConfig{
		Setpoints: setpointStore,
		Interval:  5 * time.Minute,
		Logger:    slog.Default(),
	})
	go func() {
		if err := doubleLoop.Run(ctx); err != nil {
			slog.Error("double-loop controller exited", "error", err)
		}
	}()
	log.Printf("Double-loop controller started (5min interval, setpoint adjustment)")

	// Suppress unused variable warnings for detectors available for future wiring
	_ = bandwidthDetector
	_ = feedbackDetector
	_ = routingDetector
	_ = genreMismatchDetector
	_ = bridgeDetector

	// ===== REMAINING SERVICES =====
	roleContextService := services.NewRoleContextService(app.pool)
	h.SetRoleContextService(roleContextService)
	log.Printf("Role context service registered (permission checks, hierarchy)")

	memoryHierarchyService := services.NewMemoryHierarchyService(app.pool)
	h.SetMemoryHierarchyService(memoryHierarchyService)
	log.Printf("Memory hierarchy service registered (workspace memories)")

	inviteService := services.NewWorkspaceInviteService(app.pool)
	h.SetInviteService(inviteService)
	log.Printf("Workspace invite service registered (email invitations)")

	auditService := services.NewWorkspaceAuditService(app.pool)
	h.SetAuditService(auditService)
	log.Printf("Workspace audit service registered (audit logging)")

	projectAccessService := services.NewProjectAccessService(app.pool)
	h.SetProjectAccessService(projectAccessService)
	log.Printf("Project access service registered (project member management)")

	// ===== SKILLS =====
	var skillsHandler *handlers.SkillsHandler
	skillsConfigPath := "./skills/skills.yaml"
	skillsLoader := services.NewSkillsLoader(skillsConfigPath)
	if err := skillsLoader.LoadConfig(); err != nil {
		log.Printf("Warning: Skills loader failed to initialize: %v", err)
		log.Printf("Agent skills system will be disabled")
	} else {
		skillsHandler = handlers.NewSkillsHandler(skillsLoader, app.pool, sessionCache)
		h.SetSkillsLoader(skillsLoader)
		log.Printf("Skills loader initialized (%d skills loaded)", len(skillsLoader.GetEnabledSkills()))
	}

	// ===== BACKGROUND JOBS SYSTEM =====
	var jobsHandler *handlers.BackgroundJobsHandler
	var jobWorkers []*services.JobWorker
	var jobScheduler *services.JobScheduler

	if app.dbConnected && app.pool != nil {
		slog.Info("Initializing background jobs system...")
		jobsHandler = handlers.NewBackgroundJobsHandler(app.pool)
		jobsService := jobsHandler.GetService()
		jobScheduler = jobsHandler.GetScheduler()

		for i := 1; i <= 3; i++ {
			workerID := fmt.Sprintf("worker-%d", i)
			worker := services.NewJobWorker(jobsService, workerID, 5*time.Second)

			worker.RegisterHandler("email_send", services.ExampleEmailSendHandler)
			worker.RegisterHandler("report_generate", services.ExampleReportGenerateHandler)
			worker.RegisterHandler("sync_calendar", services.ExampleSyncCalendarHandler)
			worker.RegisterHandler("user_onboarding", handlers.UserOnboardingHandler)
			worker.RegisterHandler("workspace_export", handlers.WorkspaceExportHandler)
			worker.RegisterHandler("analytics_aggregation", handlers.AnalyticsAggregationHandler)
			worker.RegisterHandler("notification_batch", handlers.NotificationBatchHandler)
			worker.RegisterHandler("data_cleanup", handlers.DataCleanupHandler)
			worker.RegisterHandler("integration_sync", handlers.IntegrationSyncHandler)
			worker.RegisterHandler("backup", handlers.BackupHandler)

			jobWorkers = append(jobWorkers, worker)

			if err := worker.Start(ctx); err != nil {
				slog.Error("Failed to start worker", "worker_id", workerID, "error", err)
			} else {
				slog.Info("Worker started", "worker_id", workerID)
			}
		}

		if err := jobScheduler.Start(ctx); err != nil {
			slog.Error("Failed to start scheduler", "error", err)
		} else {
			slog.Info("Job scheduler started")
		}
	}

	app.jobsHandler = jobsHandler
	app.jobWorkers = jobWorkers
	app.jobScheduler = jobScheduler

	// ===== OPTIONAL BACKGROUND JOBS (goroutines with explicit params) =====
	if conversationIntelligence != nil && cfg.ConversationSummaryJobEnabled {
		interval := time.Duration(cfg.ConversationSummaryJobIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = 30 * time.Minute
		}
		batchSize := cfg.ConversationSummaryJobBatchSize
		if batchSize <= 0 {
			batchSize = 25
		}
		maxMessages := cfg.ConversationSummaryJobMaxMessages
		if maxMessages <= 0 {
			maxMessages = 200
		}
		startConversationSummaryJob(ctx, conversationIntelligence, conversationSummaryJobConfig{
			interval:    interval,
			batchSize:   batchSize,
			maxMessages: maxMessages,
		})
	}

	if learningService != nil && cfg.BehaviorPatternsJobEnabled {
		interval := time.Duration(cfg.BehaviorPatternsJobIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = 60 * time.Minute
		}
		batchSize := cfg.BehaviorPatternsJobUserBatchSize
		if batchSize <= 0 {
			batchSize = 50
		}
		startBehaviorPatternsJob(ctx, learningService, behaviorPatternsJobConfig{
			interval:  interval,
			batchSize: batchSize,
		})
	}

	if appProfilerService != nil && cfg.AppProfilerSyncJobEnabled {
		interval := time.Duration(cfg.AppProfilerSyncJobIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = 10 * time.Minute
		}
		batchSize := cfg.AppProfilerSyncJobBatchSize
		if batchSize <= 0 {
			batchSize = 5
		}
		startAppProfilerSyncJob(ctx, appProfilerService, appProfilerJobConfig{
			interval:  interval,
			batchSize: batchSize,
		})
	}

	// ===== ROUTE REGISTRATION =====
	// Register routes onto the router — delegates to routes.go.
	// The versioning groups and degraded-status endpoint are set up in registerRoutes.
	registerRoutes(app, skillsHandler, osaClient)

	return app, nil
}
