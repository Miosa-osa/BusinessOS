package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/middleware"
	redisClient "github.com/rhl/businessos-backend/internal/redis"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/terminal"
)

// Note: middleware package now provides SessionCache for Redis-backed session validation

func main() {
	// Create background context for the application
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Generate unique instance ID for pub/sub (avoid message echo)
	instanceID := uuid.New().String()[:8]
	log.Printf("Server instance ID: %s", instanceID)

	// Connect to database (optional in dev)
	var pool *pgxpool.Pool
	dbConnected := false
	var dbErr error
	if cfg.DatabaseRequired {
		pool, err = database.Connect(cfg)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		dbConnected = true
		defer database.Close()
	} else {
		log.Printf("DATABASE_REQUIRED=false: starting without database (degraded mode)")
		p, err := database.Connect(cfg)
		if err != nil {
			dbErr = err
			log.Printf("Database unavailable (continuing): %v", err)
		} else {
			pool = p
			dbConnected = true
			defer database.Close()
		}
	}

	// Create a database/sql wrapper for services that use stdlib APIs.
	var sqlDB *sql.DB
	if dbConnected && pool != nil {
		sqlDB = stdlib.OpenDBFromPool(pool)
		defer sqlDB.Close()
	}

	// Connect to Redis (optional - for session storage and horizontal scaling)
	redisConnected := false
	var sessionCache *middleware.SessionCache
	var terminalPubSub *terminal.TerminalPubSub
	if cfg.RedisURL != "" {
		// Configure Redis with security settings
		redisCfg := redisClient.DefaultConfig()
		redisCfg.URL = cfg.RedisURL
		redisCfg.Password = cfg.RedisPassword
		redisCfg.TLSEnabled = cfg.RedisTLSEnabled

		// In development, allow insecure TLS for self-signed certs
		// NEVER do this in production
		if !cfg.IsProduction() && cfg.RedisTLSEnabled {
			redisCfg.TLSInsecure = true
		}

		if err := redisClient.Connect(ctx, redisCfg); err != nil {
			log.Printf("Warning: Redis unavailable: %v", err)
			log.Printf("Sessions will use direct DB auth (not optimal for horizontal scaling)")
		} else {
			redisConnected = true
			log.Printf("Redis connected successfully")
			defer redisClient.Close()

			// Create session cache for auth middleware with secure HMAC key
			sessionCacheConfig := &middleware.SessionCacheConfig{
				KeyPrefix:  "auth_session:",
				TTL:        15 * time.Minute,
				HMACSecret: cfg.RedisKeyHMACSecret, // Load from environment (CRITICAL for production)
			}
			sessionCache = middleware.NewSessionCache(redisClient.Client(), sessionCacheConfig)
			log.Printf("Session cache enabled (TTL=15m, HMAC-secured keys)")

			// Create terminal pub/sub for horizontal scaling
			terminalPubSub = terminal.NewTerminalPubSub(redisClient.Client(), instanceID)
			log.Printf("Terminal pub/sub enabled (instance=%s)", instanceID)
		}
	}

	// Initialize container manager (optional - for Docker-based terminal isolation)
	var containerMonitor *container.ContainerMonitor
	containerMgr, err := container.NewContainerManager(ctx, "businessos-workspace:latest")
	if err != nil {
		log.Printf("Warning: Container manager unavailable: %v", err)
		log.Printf("Terminal will use local PTY mode")
		containerMgr = nil
	} else {
		log.Printf("Container manager initialized successfully")

		// Initialize and start container monitor for cleanup and health checks
		containerMonitor = container.NewContainerMonitor(containerMgr, nil)
		if err := containerMonitor.StartMonitoring(ctx); err != nil {
			log.Printf("Warning: Container monitor failed to start: %v", err)
		} else {
			log.Printf("Container monitor started (cleanup=%v, idle_timeout=%v)",
				container.DefaultMonitorConfig().CleanupInterval,
				container.DefaultMonitorConfig().IdleTimeout)
		}
	}

	// Create gin router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware(cfg))

	// Apply global rate limiting (100 req/sec per IP, 200 req/sec per authenticated user)
	globalRateLimiter := middleware.GetGlobalHTTPRateLimiter()
	router.Use(middleware.RateLimitMiddleware(globalRateLimiter))
	log.Printf("Rate limiting enabled (100 req/s per IP, 200 req/s per user)")

	// Health check (no auth required)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Business OS API", "version": "1.0.0", "instance_id": instanceID})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Readiness check (includes dependencies)
	router.GET("/ready", func(c *gin.Context) {
		dbStatus := "disconnected"
		if dbConnected {
			dbStatus = "connected"
		} else if !cfg.DatabaseRequired {
			dbStatus = "disabled"
		}
		status := gin.H{
			"status":      "ready",
			"instance_id": instanceID,
			"database":    dbStatus,
			"redis":       "disconnected",
			"containers":  "unavailable",
		}

		// Check Redis
		if redisConnected && redisClient.IsConnected(c.Request.Context()) {
			status["redis"] = "connected"
		}

		// Check container manager
		if containerMgr != nil {
			status["containers"] = "available"
		}

		c.JSON(200, status)
	})

	// Detailed health check for monitoring
	router.GET("/health/detailed", func(c *gin.Context) {
		health := gin.H{
			"status":      "healthy",
			"instance_id": instanceID,
			"components":  gin.H{},
		}

		components := health["components"].(gin.H)

		// Database health
		dbComponent := gin.H{}
		if dbConnected {
			dbComponent["status"] = "connected"
		} else if !cfg.DatabaseRequired {
			dbComponent["status"] = "disabled"
			if dbErr != nil {
				dbComponent["error"] = dbErr.Error()
			}
		} else {
			dbComponent["status"] = "disconnected"
			if dbErr != nil {
				dbComponent["error"] = dbErr.Error()
			}
		}
		components["database"] = dbComponent

		// Redis health
		if redisConnected {
			redisHealth, err := redisClient.HealthCheck(c.Request.Context())
			if err != nil {
				components["redis"] = gin.H{"status": "error", "error": err.Error()}
			} else {
				components["redis"] = gin.H{
					"status":     "connected",
					"latency_ms": redisHealth.Latency.Milliseconds(),
					"pool_stats": redisHealth.PoolStats,
				}
			}
		} else {
			components["redis"] = gin.H{"status": "not_configured"}
		}

		// Container manager health
		if containerMgr != nil {
			components["containers"] = gin.H{"status": "available"}
		} else {
			components["containers"] = gin.H{"status": "unavailable"}
		}

		c.JSON(200, health)
	})

	// Serve uploaded files (profile photos, backgrounds) - no auth needed
	router.Static("/uploads", "./uploads")

	// API routes group
	api := router.Group("/api")

	// If DB isn't available, only expose basic health endpoints.
	if !dbConnected || pool == nil {
		api.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":            "degraded",
				"database":          "unavailable",
				"database_required": cfg.DatabaseRequired,
			})
		})

		log.Printf("Backend running in degraded mode (no database): only /, /health, /ready, /health/detailed, /api/status, and /uploads are available")
		// Start server
		go func() {
			log.Printf("Server starting on port %s", cfg.ServerPort)
			if err := router.Run(":" + cfg.ServerPort); err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		if containerMonitor != nil {
			log.Println("Stopping container monitor...")
			if err := containerMonitor.StopMonitoring(); err != nil {
				log.Printf("Warning: Error stopping container monitor: %v", err)
			}
		}
		if containerMgr != nil {
			log.Println("Shutting down container manager...")
			containerMgr.Shutdown()
		}
		database.Close()
		log.Println("Server stopped")
		return
	}

	// Initialize embedding service for RAG (semantic search)
	var embeddingService *services.EmbeddingService
	var contextBuilder *services.ContextBuilder
	var tieredContextService *services.TieredContextService
	embeddingService = services.NewEmbeddingService(pool, cfg.OllamaLocalURL)
	if embeddingService.HealthCheck(ctx) {
		contextBuilder = services.NewContextBuilder(pool, embeddingService)
		summarizer := services.NewSummarizerService(pool, cfg)
		tieredContextService = services.NewTieredContextService(pool, embeddingService, summarizer)
		log.Printf("Embedding service initialized (model=nomic-embed-text, dimensions=768)")
		log.Printf("Tiered context service enabled (scoped RAG, Level 1/2/3 context)")
	} else {
		log.Printf("Warning: Embedding service unavailable (Ollama not running or nomic-embed-text model not pulled)")
		log.Printf("RAG features will be disabled. Run: ollama pull nomic-embed-text")
		embeddingService = nil
	}

	// Initialize Pedro Tasks services (Memory, Context & Intelligence System)
	var documentProcessor *services.DocumentProcessor
	var learningService *services.LearningService
	var appProfilerService *services.AppProfilerService
	var conversationIntelligence *services.ConversationIntelligenceService
	var memoryExtractor *services.MemoryExtractorService
	var blockMapper *services.BlockMapperService

	// Document Processor - requires embedding service for semantic search
	if embeddingService != nil {
		documentProcessor = services.NewDocumentProcessor(pool, embeddingService, "./uploads/documents")
		log.Printf("Document processor initialized (chunking + semantic search)")
	}

	// Learning Service - always available
	learningService = services.NewLearningService(pool)
	log.Printf("Learning service initialized (feedback + personalization)")

	// App Profiler Service - always available
	appProfilerService = services.NewAppProfilerService(pool)
	log.Printf("App profiler service initialized (codebase analysis)")

	// Conversation Intelligence - requires embedding service
	if embeddingService != nil {
		conversationIntelligence = services.NewConversationIntelligenceService(pool, embeddingService)
		log.Printf("Conversation intelligence initialized (analysis + summarization)")
	}

	// Memory Extractor - requires embedding service
	if embeddingService != nil {
		memoryExtractor = services.NewMemoryExtractorService(pool, embeddingService)

		// Wire LLM service for enhanced memory extraction (using Groq for speed/cost)
		if cfg.GroqAPIKey != "" {
			groqLLM := services.NewGroqService(cfg, "llama-3.1-8b-instant") // Fast model for extraction
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

	// Block Mapper - requires sqlDB wrapper
	if sqlDB != nil {
		blockMapper = services.NewBlockMapperService(sqlDB, slog.Default())
		log.Printf("Block mapper initialized (markdown to structured blocks)")
	}

	// Initialize handlers with container manager, session cache, terminal pub/sub, and embedding services
	h := handlers.NewHandlers(pool, cfg, containerMgr, sessionCache, terminalPubSub, embeddingService, contextBuilder, tieredContextService)

	// Set Pedro Tasks services
	h.SetPedroServices(documentProcessor, learningService, appProfilerService, conversationIntelligence, memoryExtractor, blockMapper)
	log.Printf("Pedro Tasks services registered (documents, learning, profiles, intelligence, blocks)")

	// Optional background job: keep conversation_summaries fresh for context + semantic search.
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

		go func() {
			t := time.NewTicker(interval)
			defer t.Stop()
			log.Printf("Conversation summary job enabled (interval=%s batch=%d maxMessages=%d)", interval, batchSize, maxMessages)
			for {
				select {
				case <-t.C:
					count, err := conversationIntelligence.BackfillStaleSummaries(ctx, batchSize, maxMessages, false)
					if err != nil {
						log.Printf("Conversation summary job error: %v", err)
						continue
					}
					if count > 0 {
						log.Printf("Conversation summary job updated %d conversations", count)
					}
				}
			}
		}()
	}

	// Optional background job: detect behavior patterns and store them as user_facts for explicit confirmation.
	if learningService != nil && cfg.BehaviorPatternsJobEnabled {
		interval := time.Duration(cfg.BehaviorPatternsJobIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = 60 * time.Minute
		}
		batchSize := cfg.BehaviorPatternsJobUserBatchSize
		if batchSize <= 0 {
			batchSize = 50
		}

		go func() {
			t := time.NewTicker(interval)
			defer t.Stop()
			log.Printf("Behavior patterns job enabled (interval=%s userBatch=%d)", interval, batchSize)
			for {
				select {
				case <-t.C:
					usersProcessed, factsUpserted, err := learningService.BackfillRecentUsersBehaviorPatterns(ctx, batchSize)
					if err != nil {
						log.Printf("Behavior patterns job error: %v", err)
						continue
					}
					if usersProcessed > 0 || factsUpserted > 0 {
						log.Printf("Behavior patterns job processed %d users, upserted %d pattern facts", usersProcessed, factsUpserted)
					}
				}
			}
		}()
	}

	// Optional background job: auto-sync application profiles based on git HEAD or filesystem changes.
	if appProfilerService != nil && cfg.AppProfilerSyncJobEnabled {
		interval := time.Duration(cfg.AppProfilerSyncJobIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = 10 * time.Minute
		}
		batchSize := cfg.AppProfilerSyncJobBatchSize
		if batchSize <= 0 {
			batchSize = 5
		}

		go func() {
			t := time.NewTicker(interval)
			defer t.Stop()
			log.Printf("App profiler auto-sync job enabled (interval=%s batch=%d)", interval, batchSize)
			for {
				select {
				case <-t.C:
					checked, refreshed, err := appProfilerService.SyncAutoProfiles(ctx, batchSize)
					if err != nil {
						log.Printf("App profiler auto-sync job error: %v", err)
						continue
					}
					if checked > 0 || refreshed > 0 {
						log.Printf("App profiler auto-sync checked %d profiles, refreshed %d", checked, refreshed)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Register routes
	h.RegisterRoutes(api)

	// Start server
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := router.Run(":" + cfg.ServerPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop container monitor first (if available)
	if containerMonitor != nil {
		log.Println("Stopping container monitor...")
		if err := containerMonitor.StopMonitoring(); err != nil {
			log.Printf("Warning: Error stopping container monitor: %v", err)
		}
	}

	// Shutdown container manager (if available)
	if containerMgr != nil {
		log.Println("Shutting down container manager...")
		containerMgr.Shutdown()
	}

	// Close database connection
	database.Close()
	log.Println("Server stopped")
}
