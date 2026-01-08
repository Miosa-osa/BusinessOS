package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/middleware"
	redisClient "github.com/rhl/businessos-backend/internal/redis"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/terminal"
	"github.com/rhl/businessos-backend/internal/workers"
)

// Note: middleware package now provides SessionCache for Redis-backed session validation

func main() {
	// Load .env file first (before config.Load) so os.Getenv works for all services
	// This is optional - in production, env vars are set directly
	if err := godotenv.Load(); err != nil {
		log.Printf("Note: No .env file found (this is fine in production)")
	}

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

	// Connect to database
	pool, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

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
		status := gin.H{
			"status":      "ready",
			"instance_id": instanceID,
			"database":    "connected",
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
		components["database"] = gin.H{
			"status": "connected",
		}

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

	// Initialize embedding service for RAG (semantic search)
	var embeddingService *services.EmbeddingService
	var contextBuilder *services.ContextBuilder
	var tieredContextService *services.TieredContextService
	embeddingService = services.NewEmbeddingService(pool, cfg.OllamaLocalURL)
	if embeddingService.HealthCheck(ctx) {
		contextBuilder = services.NewContextBuilder(pool, embeddingService)
		tieredContextService = services.NewTieredContextService(pool, embeddingService)
		log.Printf("Embedding service initialized (model=nomic-embed-text, dimensions=768)")
		log.Printf("Tiered context service enabled (scoped RAG, Level 1/2/3 context)")
	} else {
		log.Printf("Warning: Embedding service unavailable (Ollama not running or nomic-embed-text model not pulled)")
		log.Printf("RAG features will be disabled. Run: ollama pull nomic-embed-text")
		embeddingService = nil
	}

	// Initialize notification service with SSE broadcaster
	sseBroadcaster := services.NewSSEBroadcaster()
	notificationService := services.NewNotificationService(pool, sseBroadcaster)
	log.Printf("Notification service initialized (SSE real-time enabled)")

	// Start notification batch worker
	batchWorker := workers.NewBatchWorker(pool, notificationService.Dispatcher())
	go batchWorker.Start(ctx)
	log.Printf("Notification batch worker started (interval: 10s)")

	// Initialize Web Push service (optional - requires VAPID keys)
	var webPushService *services.WebPushService
	if cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != "" {
		webPushService = services.NewWebPushService(pool, &services.WebPushConfig{
			VAPIDPublicKey:  cfg.VAPIDPublicKey,
			VAPIDPrivateKey: cfg.VAPIDPrivateKey,
			VAPIDContact:    cfg.VAPIDContact,
		})
		log.Printf("Web Push service initialized (VAPID keys configured)")
	} else {
		log.Printf("Web Push service disabled (VAPID keys not configured)")
		log.Printf("To enable: Generate keys with `npx web-push generate-vapid-keys` and set VAPID_PUBLIC_KEY, VAPID_PRIVATE_KEY")
	}

	// Initialize handlers with container manager, session cache, terminal pub/sub, and embedding services
	h := handlers.NewHandlers(pool, cfg, containerMgr, sessionCache, terminalPubSub, embeddingService, contextBuilder, tieredContextService, notificationService)

	// Set optional services
	if webPushService != nil {
		h.SetWebPushService(webPushService)
	}

	// Initialize Email service (Resend)
	emailService := services.NewEmailService()
	if emailService.IsEnabled() {
		h.SetEmailService(emailService)
		log.Printf("Email service initialized (Resend configured)")
	} else {
		log.Printf("Email service disabled (RESEND_API_KEY not set)")
	}

	// Initialize Comment service
	queries := sqlc.New(pool)
	commentService := services.NewCommentService(pool, queries, notificationService)
	h.SetCommentService(commentService)
	log.Printf("Comment service initialized (mentions & notifications enabled)")

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

	// Stop batch worker
	log.Println("Stopping notification batch worker...")
	batchWorker.Stop()

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
