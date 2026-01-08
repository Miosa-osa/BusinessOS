# Embedding Cache - Quick Start Guide

Get the embedding cache up and running in 5 minutes.

## Prerequisites

- Go 1.24+
- Redis server (local or remote)
- Existing BusinessOS backend running

## 1. Start Redis (Local Development)

```bash
# Using Docker (recommended)
docker run -d --name redis-cache -p 6379:6379 redis:latest

# Or using Redis CLI
redis-server

# Verify Redis is running
redis-cli ping
# Should return: PONG
```

## 2. Add to Your Code

### Option A: Quick Integration (Minimal Changes)

Add to your existing initialization code:

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/rhl/businessos-backend/internal/services"
)

// In your main() or setup function:
func setupServices() {
    // Your existing code...
    pool := initDatabase()

    // NEW: Add Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // NEW: Create embedding cache
    embeddingCache := services.NewEmbeddingCacheService(
        redisClient,
        pool,
        nil, // Use defaults
    )

    // Your existing embedding service
    embeddingService := services.NewEmbeddingService(pool, "http://localhost:11434")

    // NEW: Connect cache to embedding service
    embeddingService.SetCache(embeddingCache)

    // Done! Your embeddings are now cached.
}
```

### Option B: Full Integration with Configuration

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/rhl/businessos-backend/internal/services"
)

func main() {
    // 1. Setup Redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     getEnv("REDIS_URL", "localhost:6379"),
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
    })

    // Test connection
    ctx := context.Background()
    if err := redisClient.Ping(ctx).Err(); err != nil {
        log.Printf("WARNING: Redis unavailable: %v", err)
        redisClient = nil // Will disable caching gracefully
    }

    // 2. Configure cache
    cacheConfig := &services.EmbeddingCacheConfig{
        KeyPrefix:  "embedding:",
        TextTTL:    24 * time.Hour,
        ImageTTL:   48 * time.Hour,
        Enabled:    true,
    }

    // 3. Create cache service
    embeddingCache := services.NewEmbeddingCacheService(
        redisClient,
        pool,
        cacheConfig,
    )

    // 4. Create embedding service with cache
    embeddingService := services.NewEmbeddingService(pool, "http://localhost:11434")
    embeddingService.SetCache(embeddingCache)

    // 5. Log initial status
    if embeddingCache.IsEnabled() {
        log.Println("✓ Embedding cache ENABLED")
        stats, _ := embeddingCache.GetCacheStats(ctx)
        log.Printf("  Cached: %d embeddings", stats.Size)
    } else {
        log.Println("✗ Embedding cache DISABLED")
    }

    // Continue with your app...
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
```

## 3. Test It Works

### Test 1: Check Cache Status

```go
func testCache() {
    ctx := context.Background()

    // Check if enabled
    if !embeddingCache.IsEnabled() {
        log.Println("Cache is disabled")
        return
    }

    // Check health
    if embeddingCache.HealthCheck(ctx) {
        log.Println("✓ Cache is healthy")
    } else {
        log.Println("✗ Cache health check failed")
    }

    // Get stats
    stats, err := embeddingCache.GetCacheStats(ctx)
    if err != nil {
        log.Printf("Error getting stats: %v", err)
        return
    }

    log.Printf("Cache Stats:")
    log.Printf("  Size: %d entries", stats.Size)
    log.Printf("  Hits: %d", stats.Hits)
    log.Printf("  Misses: %d", stats.Misses)
    log.Printf("  Hit Rate: %.2f%%", stats.HitRate*100)
}
```

### Test 2: Manual Cache Operations

```go
func testManualCaching() {
    ctx := context.Background()

    // Test data
    content := "Hello, world!"
    testEmbedding := []float64{0.1, 0.2, 0.3, 0.4, 0.5}

    // Set embedding
    err := embeddingCache.SetEmbedding(ctx, content, testEmbedding, "text", 24*time.Hour)
    if err != nil {
        log.Printf("✗ Failed to cache: %v", err)
        return
    }
    log.Println("✓ Embedding cached")

    // Get embedding
    retrieved, found, err := embeddingCache.GetEmbedding(ctx, content, "text")
    if err != nil {
        log.Printf("✗ Failed to retrieve: %v", err)
        return
    }

    if !found {
        log.Println("✗ Embedding not found in cache")
        return
    }

    log.Printf("✓ Embedding retrieved: %v", retrieved)

    // Verify it matches
    if len(retrieved) == len(testEmbedding) {
        log.Println("✓ Embedding matches!")
    }
}
```

### Test 3: Via API (HTTP)

```bash
# Generate an embedding (will cache it)
curl -X POST http://localhost:8080/api/rag/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is machine learning?",
    "user_id": "test-user"
  }'

# Same query again (should be faster - cache hit)
time curl -X POST http://localhost:8080/api/rag/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is machine learning?",
    "user_id": "test-user"
  }'
```

## 4. Add Health Check Endpoint (Optional)

```go
// In your router setup:
router.GET("/health/cache", func(c *gin.Context) {
    ctx := c.Request.Context()

    stats, err := embeddingCache.GetCacheStats(ctx)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "enabled": embeddingCache.IsEnabled(),
        "healthy": embeddingCache.HealthCheck(ctx),
        "stats": gin.H{
            "size":     stats.Size,
            "hits":     stats.Hits,
            "misses":   stats.Misses,
            "hit_rate": stats.HitRate,
        },
    })
})
```

Test it:
```bash
curl http://localhost:8080/health/cache
```

Expected response:
```json
{
  "enabled": true,
  "healthy": true,
  "stats": {
    "size": 42,
    "hits": 150,
    "misses": 30,
    "hit_rate": 0.833
  }
}
```

## 5. Environment Variables (Production)

Add to your `.env` file:

```bash
# Redis
REDIS_URL=localhost:6379
REDIS_PASSWORD=your-secure-password
REDIS_DB=0

# Cache Configuration
EMBEDDING_CACHE_ENABLED=true
EMBEDDING_CACHE_TEXT_TTL=24h
EMBEDDING_CACHE_IMAGE_TTL=48h
```

## 6. Verify Cache is Working

### Method 1: Check Logs

You should see log entries like:
```
EmbeddingCache: Redis connected successfully
Embedding cache enabled
Cache Stats: Size=15, Hits=45, Misses=5, HitRate=90.00%
```

### Method 2: Check Redis Directly

```bash
# Connect to Redis
redis-cli

# List all embedding keys
KEYS embedding:*

# Example output:
# 1) "embedding:text:a3c5e8d9f2b1c4a7"
# 2) "embedding:text:b7f2a9c1e5d3b8f6"
# 3) "embedding:image:c9d4e1a7f6b8c3d2"

# Get a specific embedding
GET embedding:text:a3c5e8d9f2b1c4a7

# Check TTL (time to live)
TTL embedding:text:a3c5e8d9f2b1c4a7
# Returns seconds until expiration (e.g., 86400 = 24 hours)
```

### Method 3: Performance Test

```go
func benchmarkCache() {
    ctx := context.Background()
    text := "benchmark test query"

    // First call - should be slow (calls Ollama)
    start := time.Now()
    _, err := embeddingService.GenerateEmbedding(ctx, text)
    firstCall := time.Since(start)

    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    // Second call - should be fast (cache hit)
    start = time.Now()
    _, err = embeddingService.GenerateEmbedding(ctx, text)
    secondCall := time.Since(start)

    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    log.Printf("Performance:")
    log.Printf("  First call:  %v (cache miss + Ollama)", firstCall)
    log.Printf("  Second call: %v (cache hit)", secondCall)
    log.Printf("  Speedup:     %.1fx", float64(firstCall)/float64(secondCall))

    // Expected output:
    // First call:  250ms (cache miss + Ollama)
    // Second call: 2ms (cache hit)
    // Speedup:     125.0x
}
```

## 7. Monitor in Production

Add monitoring to track cache performance:

```go
func startCacheMonitoring() {
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        defer ticker.Stop()

        for range ticker.C {
            ctx := context.Background()
            stats, err := embeddingCache.GetCacheStats(ctx)
            if err != nil {
                log.Printf("Cache stats error: %v", err)
                continue
            }

            log.Printf("Cache: %d entries, %.1f%% hit rate, %d hits, %d misses",
                stats.Size,
                stats.HitRate*100,
                stats.Hits,
                stats.Misses,
            )

            // Alert if hit rate is low
            total := stats.Hits + stats.Misses
            if total > 100 && stats.HitRate < 0.5 {
                log.Printf("WARNING: Low cache hit rate: %.1f%%", stats.HitRate*100)
            }
        }
    }()
}

// In main():
startCacheMonitoring()
```

## Troubleshooting

### Cache Not Working?

1. **Check Redis is running**:
   ```bash
   redis-cli ping
   # Should return: PONG
   ```

2. **Check logs for errors**:
   ```
   Look for: "Redis connected successfully"
   Or: "Redis unavailable"
   ```

3. **Verify cache is enabled**:
   ```go
   log.Printf("Cache enabled: %v", embeddingCache.IsEnabled())
   ```

4. **Check health**:
   ```go
   log.Printf("Cache healthy: %v", embeddingCache.HealthCheck(ctx))
   ```

### Low Hit Rate?

- Cache might be clearing too frequently (check TTL)
- Different queries each time (expected)
- Cache size limited (check Redis maxmemory)

### Redis Connection Errors?

```
Error: "connection refused"
→ Redis not running. Start Redis: docker run -d -p 6379:6379 redis

Error: "authentication failed"
→ Set REDIS_PASSWORD environment variable

Error: "timeout"
→ Redis overloaded or network issue. Check Redis CPU/memory
```

## Next Steps

1. ✓ Cache is working
2. Monitor hit rates in production
3. Tune TTLs based on your data patterns
4. Set up Redis persistence (optional)
5. Configure Redis clustering for HA (production)

## Common Usage Patterns

### Pattern 1: Cache-Aside (Automatic)

```go
// This happens automatically with SetCache():
embedding, err := embeddingService.GenerateEmbedding(ctx, text)
// 1. Checks cache
// 2. If miss, calls Ollama
// 3. Caches result
// 4. Returns embedding
```

### Pattern 2: Manual Cache Control

```go
// Manual cache check
cached, found, err := embeddingCache.GetEmbedding(ctx, text, "text")
if found {
    // Use cached embedding
    return cached, nil
}

// Compute and cache manually
embedding := computeExpensiveEmbedding(text)
embeddingCache.SetEmbedding(ctx, text, embedding, "text", 24*time.Hour)
return embedding, nil
```

### Pattern 3: Bulk Pre-warming

```go
// Pre-warm cache with common queries
func warmCache(commonQueries []string) {
    for _, query := range commonQueries {
        embedding, err := embeddingService.GenerateEmbedding(ctx, query)
        if err == nil {
            log.Printf("Cached: %s", query)
        }
    }
}

// In main():
warmCache([]string{
    "What is AI?",
    "Machine learning basics",
    "Deep learning tutorial",
})
```

## Performance Expectations

| Operation | Latency | Notes |
|-----------|---------|-------|
| Cache Hit | 1-2ms | Redis GET + unmarshal |
| Cache Miss | 1-2ms | Redis GET (returns nil) |
| Cache Set | 2-3ms | Marshal + Redis SET |
| Ollama Call | 100-300ms | Depends on model/hardware |
| **Total Speedup** | **50-150x** | When cache is warm |

## Success Checklist

- [ ] Redis running and accessible
- [ ] Cache service initialized
- [ ] Connected to EmbeddingService
- [ ] Health check endpoint working
- [ ] Test query shows cache hit on second call
- [ ] Monitoring enabled
- [ ] Hit rate >70% after warmup

You're done! Embeddings are now cached and your RAG system is significantly faster.

## Need Help?

- Check: `EMBEDDING_CACHE_README.md` for detailed API docs
- Check: `EMBEDDING_CACHE_INTEGRATION.md` for integration patterns
- Check: `EMBEDDING_CACHE_ARCHITECTURE.md` for system design
- Run tests: `go test -v ./internal/services -run TestEmbedding`
