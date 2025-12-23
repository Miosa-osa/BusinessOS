package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	pool, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize container manager (optional - for Docker-based terminal isolation)
	ctx := context.Background()
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

	// Health check (no auth required)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Business OS API", "version": "1.0.0"})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Serve uploaded files (profile photos, backgrounds) - no auth needed
	router.Static("/uploads", "./uploads")

	// API routes group
	api := router.Group("/api")

	// Initialize handlers with container manager
	h := handlers.NewHandlers(pool, cfg, containerMgr)

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
