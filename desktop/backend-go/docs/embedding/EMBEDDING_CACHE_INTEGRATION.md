# Embedding Cache Integration Guide

This guide shows how to integrate the `EmbeddingCacheService` into your BusinessOS backend.

## Step 1: Initialize Redis Client

In your `main.go` or initialization code:

```go
package main

import (
    "context"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/rhl/businessos-backend/internal/services"
)

func initializeRedis() *redis.Client {
    // Get Redis configuration from environment
    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        redisURL = "localhost:6379"
    }

    redisPassword := os.Getenv("REDIS_PASSWORD")
    redisDB := 0
    if db := os.Getenv("REDIS_DB"); db != "" {
        if parsed, err := strconv.Atoi(db); err == nil {
            redisDB = parsed
        }
    }

    // Create Redis client
    client := redis.NewClient(&redis.Options{
        Addr:         redisURL,
        Password:     redisPassword,
        DB:           redisDB,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolSize:     10,
        MinIdleConns: 5,
    })

    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        log.Printf("WARNING: Redis connection failed: %v", err)
        log.Println("Continuing without Redis cache")
        return nil // Return nil, services will handle gracefully
    }

    log.Println("Redis connected successfully")
    return client
}
```

## Step 2: Create Embedding Cache Service

```go
func initializeEmbeddingCache(redisClient *redis.Client, pool *pgxpool.Pool) *services.EmbeddingCacheService {
    // Configure cache
    cfg := &services.EmbeddingCacheConfig{
        KeyPrefix:        "embedding:",
        DefaultTTL:       24 * time.Hour,
        TextTTL:          24 * time.Hour,  // Text embeddings cached for 1 day
        ImageTTL:         48 * time.Hour,  // Image embeddings cached for 2 days
        Enabled:          true,
        GracefulFallback: true,
    }

    // Create cache service
    cacheService := services.NewEmbeddingCacheService(redisClient, pool, cfg)

    // Log initial stats
    ctx := context.Background()
    if cacheService.IsEnabled() {
        stats, err := cacheService.GetCacheStats(ctx)
        if err == nil {
            log.Printf("Embedding Cache initialized: %d cached embeddings", stats.Size)
        }
    } else {
        log.Println("Embedding Cache is disabled")
    }

    return cacheService
}
```

## Step 3: Integrate with EmbeddingService

```go
func main() {
    // Initialize database pool
    pool := initializeDatabase()
    defer pool.Close()

    // Initialize Redis
    redisClient := initializeRedis()
    if redisClient != nil {
        defer redisClient.Close()
    }

    // Initialize embedding cache
    embeddingCache := initializeEmbeddingCache(redisClient, pool)

    // Initialize embedding service
    ollamaURL := os.Getenv("OLLAMA_URL")
    if ollamaURL == "" {
        ollamaURL = "http://localhost:11434"
    }

    embeddingService := services.NewEmbeddingService(pool, ollamaURL)

    // Connect cache to embedding service
    if embeddingCache != nil && embeddingCache.IsEnabled() {
        embeddingService.SetCache(embeddingCache)
        log.Println("Embedding cache connected to embedding service")
    }

    // Initialize other services that use embeddings
    hybridSearch := services.NewHybridSearchService(
        pool,
        embeddingService,
        nil, // Will use cache through embeddingService
    )

    // ... rest of initialization
}
```

## Step 4: Update Existing RAGCacheService Integration (Optional)

If you're already using `RAGCacheService`, you can use both:

```go
func initializeCaching(redisClient *redis.Client, pool *pgxpool.Pool) (
    *services.EmbeddingCacheService,
    *services.RAGCacheService,
) {
    // Initialize embedding cache (for low-level embedding caching)
    embeddingCfg := services.DefaultEmbeddingCacheConfig()
    embeddingCache := services.NewEmbeddingCacheService(redisClient, pool, embeddingCfg)

    // Initialize RAG cache (for high-level query result caching)
    ragCfg := services.DefaultRAGCacheConfig()
    ragCache := services.NewRAGCacheService(redisClient, ragCfg)

    return embeddingCache, ragCache
}
```

**Note**: The `EmbeddingCacheService` and existing `RAGCacheService` serve different purposes:
- **EmbeddingCacheService**: Caches individual embedding vectors (low-level)
- **RAGCacheService**: Caches entire search results and RAG responses (high-level)

They can work together - the existing `RAGCacheService.GetEmbedding()` method returns `[]float32`, while the new service uses `[]float64` for precision.

## Step 5: Add Health Check Endpoint

```go
// In your handlers package
func HandleCacheHealth(embeddingCache *services.EmbeddingCacheService) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        stats, err := embeddingCache.GetCacheStats(ctx)
        if err != nil {
            c.JSON(500, gin.H{
                "error": "Failed to get cache stats",
                "details": err.Error(),
            })
            return
        }

        healthy := embeddingCache.HealthCheck(ctx)

        c.JSON(200, gin.H{
            "service": "embedding_cache",
            "enabled": embeddingCache.IsEnabled(),
            "healthy": healthy,
            "stats": gin.H{
                "hits":      stats.Hits,
                "misses":    stats.Misses,
                "size":      stats.Size,
                "hit_rate":  stats.HitRate,
                "hit_rate_percent": stats.HitRate * 100,
            },
        })
    }
}

// Register route
router.GET("/health/cache", HandleCacheHealth(embeddingCache))
```

## Step 6: Add Cache Management Endpoints (Optional)

```go
// Clear cache endpoint (admin only)
func HandleClearCache(embeddingCache *services.EmbeddingCacheService) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // Get stats before clearing
        beforeStats, _ := embeddingCache.GetCacheStats(ctx)

        // Clear cache
        if err := embeddingCache.ClearCache(ctx); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "Cache cleared successfully",
            "cleared_entries": beforeStats.Size,
        })
    }
}

// Invalidate specific embedding
func HandleInvalidateEmbedding(embeddingCache *services.EmbeddingCacheService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req struct {
            Content       string `json:"content" binding:"required"`
            EmbeddingType string `json:"embedding_type"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        if req.EmbeddingType == "" {
            req.EmbeddingType = "text"
        }

        ctx := c.Request.Context()
        if err := embeddingCache.InvalidateEmbedding(ctx, req.Content, req.EmbeddingType); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "Embedding invalidated successfully",
            "content_hash": "...", // Optionally return hash for verification
        })
    }
}

// Register admin routes
adminRoutes := router.Group("/admin/cache")
adminRoutes.Use(middleware.RequireAdmin()) // Your admin middleware
{
    adminRoutes.POST("/clear", HandleClearCache(embeddingCache))
    adminRoutes.POST("/invalidate", HandleInvalidateEmbedding(embeddingCache))
}
```

## Step 7: Add Monitoring (Production)

```go
// Background goroutine for cache monitoring
func startCacheMonitoring(embeddingCache *services.EmbeddingCacheService) {
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        defer ticker.Stop()

        for range ticker.C {
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

            stats, err := embeddingCache.GetCacheStats(ctx)
            if err != nil {
                log.Printf("Cache monitoring: failed to get stats: %v", err)
                cancel()
                continue
            }

            // Log metrics
            log.Printf("Cache Stats: Size=%d, Hits=%d, Misses=%d, HitRate=%.2f%%",
                stats.Size, stats.Hits, stats.Misses, stats.HitRate*100)

            // Alert on low hit rate
            total := stats.Hits + stats.Misses
            if total > 100 && stats.HitRate < 0.5 {
                log.Printf("WARNING: Low cache hit rate: %.2f%%", stats.HitRate*100)
            }

            // Alert on high memory usage
            if stats.Size > 100000 {
                log.Printf("WARNING: Large cache size: %d entries", stats.Size)
            }

            // Health check
            if !embeddingCache.HealthCheck(ctx) {
                log.Println("WARNING: Cache health check failed")

                // Try to re-enable if it was disabled
                if !embeddingCache.IsEnabled() {
                    log.Println("Attempting to re-enable cache...")
                    embeddingCache.Enable()
                }
            }

            cancel()
        }
    }()
}

// In main():
if embeddingCache.IsEnabled() {
    startCacheMonitoring(embeddingCache)
}
```

## Step 8: Environment Configuration

Add to your `.env` file:

```bash
# Redis Configuration
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Embedding Cache Configuration
EMBEDDING_CACHE_ENABLED=true
EMBEDDING_CACHE_TEXT_TTL=24h
EMBEDDING_CACHE_IMAGE_TTL=48h
```

Parse in code:

```go
func getEmbeddingCacheConfig() *services.EmbeddingCacheConfig {
    cfg := services.DefaultEmbeddingCacheConfig()

    // Parse enabled
    if enabled := os.Getenv("EMBEDDING_CACHE_ENABLED"); enabled != "" {
        cfg.Enabled = enabled == "true"
    }

    // Parse TTLs
    if textTTL := os.Getenv("EMBEDDING_CACHE_TEXT_TTL"); textTTL != "" {
        if duration, err := time.ParseDuration(textTTL); err == nil {
            cfg.TextTTL = duration
        }
    }

    if imageTTL := os.Getenv("EMBEDDING_CACHE_IMAGE_TTL"); imageTTL != "" {
        if duration, err := time.ParseDuration(imageTTL); err == nil {
            cfg.ImageTTL = duration
        }
    }

    return cfg
}
```

## Complete Integration Example

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
    "github.com/rhl/businessos-backend/internal/services"
)

func main() {
    // 1. Initialize database
    pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer pool.Close()

    // 2. Initialize Redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_URL"),
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    if err := redisClient.Ping(ctx).Err(); err != nil {
        log.Printf("WARNING: Redis unavailable: %v", err)
        redisClient = nil // Will disable caching
    }
    cancel()

    if redisClient != nil {
        defer redisClient.Close()
    }

    // 3. Initialize embedding cache
    embeddingCfg := getEmbeddingCacheConfig()
    embeddingCache := services.NewEmbeddingCacheService(redisClient, pool, embeddingCfg)

    // 4. Initialize embedding service with cache
    embeddingService := services.NewEmbeddingService(pool, os.Getenv("OLLAMA_URL"))
    if embeddingCache.IsEnabled() {
        embeddingService.SetCache(embeddingCache)
        log.Println("Embedding cache enabled")

        // Start monitoring
        go startCacheMonitoring(embeddingCache)
    }

    // 5. Initialize other services
    hybridSearch := services.NewHybridSearchService(pool, embeddingService, nil)
    agenticRAG := services.NewAgenticRAGService(pool, embeddingService, hybridSearch)

    // 6. Setup routes
    router := gin.Default()

    // Health endpoints
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    router.GET("/health/cache", HandleCacheHealth(embeddingCache))

    // Admin endpoints
    adminRoutes := router.Group("/admin/cache")
    {
        adminRoutes.POST("/clear", HandleClearCache(embeddingCache))
        adminRoutes.POST("/invalidate", HandleInvalidateEmbedding(embeddingCache))
        adminRoutes.GET("/stats", HandleCacheStats(embeddingCache))
    }

    // 7. Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Starting server on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func getEmbeddingCacheConfig() *services.EmbeddingCacheConfig {
    cfg := services.DefaultEmbeddingCacheConfig()

    if enabled := os.Getenv("EMBEDDING_CACHE_ENABLED"); enabled == "false" {
        cfg.Enabled = false
    }

    if textTTL := os.Getenv("EMBEDDING_CACHE_TEXT_TTL"); textTTL != "" {
        if duration, err := time.ParseDuration(textTTL); err == nil {
            cfg.TextTTL = duration
        }
    }

    if imageTTL := os.Getenv("EMBEDDING_CACHE_IMAGE_TTL"); imageTTL != "" {
        if duration, err := time.ParseDuration(imageTTL); err == nil {
            cfg.ImageTTL = duration
        }
    }

    return cfg
}

func startCacheMonitoring(cache *services.EmbeddingCacheService) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        stats, err := cache.GetCacheStats(ctx)
        cancel()

        if err != nil {
            log.Printf("Cache monitoring error: %v", err)
            continue
        }

        log.Printf("Cache: %d entries, %.2f%% hit rate",
            stats.Size, stats.HitRate*100)
    }
}

func HandleCacheHealth(cache *services.EmbeddingCacheService) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        stats, err := cache.GetCacheStats(ctx)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{
            "enabled": cache.IsEnabled(),
            "healthy": cache.HealthCheck(ctx),
            "stats":   stats,
        })
    }
}

func HandleClearCache(cache *services.EmbeddingCacheService) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := cache.ClearCache(c.Request.Context()); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"message": "Cache cleared"})
    }
}

func HandleInvalidateEmbedding(cache *services.EmbeddingCacheService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req struct {
            Content string `json:"content" binding:"required"`
            Type    string `json:"type"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        if req.Type == "" {
            req.Type = "text"
        }

        if err := cache.InvalidateEmbedding(c.Request.Context(), req.Content, req.Type); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"message": "Invalidated"})
    }
}

func HandleCacheStats(cache *services.EmbeddingCacheService) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        stats, err := cache.GetCacheStats(ctx)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, stats)
    }
}
```

## Testing the Integration

```bash
# 1. Start Redis
docker run -d -p 6379:6379 redis:latest

# 2. Start your server
go run cmd/server/main.go

# 3. Test cache health
curl http://localhost:8080/health/cache

# 4. Use the API (embeddings will be cached automatically)
curl -X POST http://localhost:8080/api/rag/search \
  -H "Content-Type: application/json" \
  -d '{"query": "test query", "user_id": "user123"}'

# 5. Check cache stats
curl http://localhost:8080/admin/cache/stats

# 6. Clear cache if needed
curl -X POST http://localhost:8080/admin/cache/clear
```

## Migration from RAGCacheService

If you're currently using `RAGCacheService` for embeddings:

### Before (using RAGCacheService):
```go
// In embedding.go
if s.cache != nil {
    if cached, err := s.cache.GetEmbedding(ctx, text); err == nil && cached != nil {
        return cached, nil // Returns []float32
    }
}
```

### After (using EmbeddingCacheService):
```go
// In embedding.go
if s.cache != nil {
    if cached, found, err := s.cache.GetEmbedding(ctx, text, "text"); err == nil && found {
        // Convert []float64 to []float32 if needed
        result := make([]float32, len(cached))
        for i, v := range cached {
            result[i] = float32(v)
        }
        return result, nil
    }
}
```

**Note**: The new service uses `[]float64` for better precision. Convert as needed for your use case.

## Next Steps

1. Deploy with Redis in production
2. Monitor cache performance
3. Tune TTLs based on usage patterns
4. Set up alerts for cache failures
5. Consider Redis clustering for high availability
