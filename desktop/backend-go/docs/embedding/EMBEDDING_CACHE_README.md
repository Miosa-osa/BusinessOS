# Embedding Cache Service

A production-ready Redis-backed caching layer for embeddings in the RAG system.

## Features

- **Redis-backed caching** with configurable TTL
- **Graceful degradation** when Redis is unavailable
- **Content-based hashing** for deterministic cache keys
- **Multi-modal support** (text, image embeddings)
- **Cache statistics** (hits, misses, size, hit rate)
- **Atomic operations** with proper timeouts
- **Type safety** with validation
- **Thread-safe** operations

## Installation

The service uses the existing Redis client:

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/rhl/businessos-backend/internal/services"
)
```

## Quick Start

```go
// Initialize Redis client
redisClient := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: os.Getenv("REDIS_PASSWORD"),
    DB:       0,
})

// Create embedding cache service
cacheService := services.NewEmbeddingCacheService(
    redisClient,
    nil, // Optional: pgxpool for DB fallback
    nil, // Use default config
)

// Use in your embedding service
embeddingService := services.NewEmbeddingService(pool, ollamaURL)
embeddingService.SetCache(cacheService)
```

## Configuration

### Default Configuration

```go
cfg := services.DefaultEmbeddingCacheConfig()
// Returns:
// {
//     KeyPrefix:        "embedding:",
//     DefaultTTL:       24 * time.Hour,
//     TextTTL:          24 * time.Hour,
//     ImageTTL:         48 * time.Hour,
//     Enabled:          true,
//     GracefulFallback: true,
// }
```

### Custom Configuration

```go
cfg := &services.EmbeddingCacheConfig{
    KeyPrefix:        "my_app:embeddings:",
    DefaultTTL:       12 * time.Hour,
    TextTTL:          6 * time.Hour,
    ImageTTL:         24 * time.Hour,
    Enabled:          true,
    GracefulFallback: true,
}

cacheService := services.NewEmbeddingCacheService(redisClient, pool, cfg)
```

## API Reference

### Core Operations

#### GetEmbedding

Retrieve a cached embedding.

```go
func (s *EmbeddingCacheService) GetEmbedding(
    ctx context.Context,
    content string,
    embeddingType string,
) ([]float64, bool, error)
```

**Parameters:**
- `content`: The content to get embedding for (will be hashed)
- `embeddingType`: Type of embedding ("text", "image", etc.)

**Returns:**
- `[]float64`: The embedding vector (nil if not found)
- `bool`: true if found in cache, false otherwise
- `error`: non-nil if there was an error

**Example:**

```go
embedding, found, err := cacheService.GetEmbedding(ctx, "hello world", "text")
if err != nil {
    log.Printf("Cache error: %v", err)
}

if found {
    fmt.Printf("Cache HIT: got %d-dimensional embedding\n", len(embedding))
} else {
    fmt.Println("Cache MISS: need to compute embedding")
    // Compute embedding...
}
```

#### SetEmbedding

Store an embedding in cache.

```go
func (s *EmbeddingCacheService) SetEmbedding(
    ctx context.Context,
    content string,
    embedding []float64,
    embeddingType string,
    ttl time.Duration,
) error
```

**Parameters:**
- `content`: The content being embedded
- `embedding`: The embedding vector
- `embeddingType`: Type of embedding ("text", "image", etc.)
- `ttl`: Time-to-live (0 = use default based on type)

**Example:**

```go
embedding := []float64{0.1, 0.2, 0.3, ..., 0.768}

// Use default TTL (24h for text)
err := cacheService.SetEmbedding(ctx, "hello world", embedding, "text", 0)

// Use custom TTL
err := cacheService.SetEmbedding(ctx, "image.jpg", embedding, "image", 48*time.Hour)
```

#### InvalidateEmbedding

Remove a specific embedding from cache.

```go
func (s *EmbeddingCacheService) InvalidateEmbedding(
    ctx context.Context,
    content string,
    embeddingType string,
) error
```

**Example:**

```go
// When content changes, invalidate its cached embedding
err := cacheService.InvalidateEmbedding(ctx, "updated document", "text")
```

#### ClearCache

Remove all cached embeddings.

```go
func (s *EmbeddingCacheService) ClearCache(ctx context.Context) error
```

**Example:**

```go
// Clear all embeddings (e.g., after model update)
err := cacheService.ClearCache(ctx)
```

### Statistics and Monitoring

#### GetCacheStats

Get cache performance metrics.

```go
func (s *EmbeddingCacheService) GetCacheStats(
    ctx context.Context,
) (services.CacheStats, error)
```

**Returns:**

```go
type CacheStats struct {
    Hits    int64   // Total cache hits
    Misses  int64   // Total cache misses
    Size    int64   // Number of cached embeddings
    HitRate float64 // Hit rate (0.0 to 1.0)
}
```

**Example:**

```go
stats, err := cacheService.GetCacheStats(ctx)
if err != nil {
    log.Printf("Failed to get stats: %v", err)
    return
}

fmt.Printf("Cache Statistics:\n")
fmt.Printf("  Hits:     %d\n", stats.Hits)
fmt.Printf("  Misses:   %d\n", stats.Misses)
fmt.Printf("  Size:     %d embeddings\n", stats.Size)
fmt.Printf("  Hit Rate: %.2f%%\n", stats.HitRate*100)
```

### Health and Control

#### HealthCheck

Check if Redis is available.

```go
func (s *EmbeddingCacheService) HealthCheck(ctx context.Context) bool
```

**Example:**

```go
if !cacheService.HealthCheck(ctx) {
    log.Println("WARNING: Embedding cache is unhealthy")
}
```

#### Enable / Disable

Manually control cache state.

```go
func (s *EmbeddingCacheService) Enable()
func (s *EmbeddingCacheService) Disable()
func (s *EmbeddingCacheService) IsEnabled() bool
```

**Example:**

```go
// Temporarily disable cache during maintenance
cacheService.Disable()
// ... maintenance ...
cacheService.Enable()

// Check state
if cacheService.IsEnabled() {
    fmt.Println("Cache is active")
}
```

## Integration Examples

### With Existing EmbeddingService

The `EmbeddingService` already has cache support built-in:

```go
// Create services
embeddingService := services.NewEmbeddingService(pool, ollamaURL)
cacheService := services.NewEmbeddingCacheService(redisClient, pool, nil)

// Connect them
embeddingService.SetCache(cacheService)

// Now embeddings are automatically cached
embedding, err := embeddingService.GenerateEmbedding(ctx, "hello world")
// First call: cache miss, calls Ollama
// Second call: cache hit, instant return
```

### Manual Cache Pattern

```go
func GetOrComputeEmbedding(
    ctx context.Context,
    content string,
    embeddingType string,
) ([]float64, error) {
    // Try cache first
    embedding, found, err := cacheService.GetEmbedding(ctx, content, embeddingType)
    if err != nil {
        log.Printf("Cache error (continuing): %v", err)
    }

    if found {
        return embedding, nil
    }

    // Cache miss - compute embedding
    embedding, err = computeExpensiveEmbedding(content)
    if err != nil {
        return nil, err
    }

    // Store in cache for next time
    if err := cacheService.SetEmbedding(ctx, content, embedding, embeddingType, 0); err != nil {
        log.Printf("Failed to cache embedding: %v", err)
        // Don't fail - we still have the embedding
    }

    return embedding, nil
}
```

### Batch Caching

```go
func CacheBatchEmbeddings(
    ctx context.Context,
    documents []Document,
) error {
    for _, doc := range documents {
        embedding, err := generateEmbedding(doc.Content)
        if err != nil {
            log.Printf("Failed to embed %s: %v", doc.ID, err)
            continue
        }

        // Cache each embedding
        if err := cacheService.SetEmbedding(
            ctx,
            doc.Content,
            embedding,
            "text",
            24*time.Hour,
        ); err != nil {
            log.Printf("Failed to cache %s: %v", doc.ID, err)
        }
    }

    return nil
}
```

### Monitoring Endpoint

```go
func HandleCacheStats(c *gin.Context) {
    ctx := c.Request.Context()

    stats, err := cacheService.GetCacheStats(ctx)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "enabled":  cacheService.IsEnabled(),
        "healthy":  cacheService.HealthCheck(ctx),
        "stats":    stats,
    })
}
```

## Cache Key Generation

The service uses **content-based hashing** to generate deterministic cache keys:

```
Key Format: <prefix><type>:<content-hash>

Example:
  embedding:text:a3c5e8d9f... (SHA-256 hash of content)
  embedding:image:b7f2a9c1e...
```

This ensures:
- **Deduplication**: Same content = same key
- **Security**: Content not exposed in Redis
- **Collision resistance**: SHA-256 provides strong guarantees

## Performance Characteristics

### Latency

| Operation | Latency | Notes |
|-----------|---------|-------|
| Cache Hit | ~1-2ms | Redis GET + JSON unmarshal |
| Cache Miss | ~1-2ms | Redis GET (returns nil) |
| Cache Set | ~2-3ms | JSON marshal + Redis SET |
| Stats | ~10-50ms | SCAN operation (depends on cache size) |
| Clear | ~100ms-1s | SCAN + DELETE (depends on cache size) |

### Memory Usage

| Embedding Size | Per Entry | 1K Entries | 10K Entries |
|----------------|-----------|------------|-------------|
| 768-dim (text) | ~6 KB | ~6 MB | ~60 MB |
| 1536-dim (OpenAI) | ~12 KB | ~12 MB | ~120 MB |
| 512-dim (CLIP) | ~4 KB | ~4 MB | ~40 MB |

*Includes JSON overhead and Redis metadata*

## Error Handling

The service implements **graceful degradation**:

### Automatic Handling

1. **Redis unavailable at startup**: Cache disabled, all operations no-op
2. **Redis connection lost**: Operations logged, treated as cache misses
3. **Corrupted cache data**: Entry deleted, treated as cache miss
4. **Timeout**: Operation aborted after 2 seconds, treated as cache miss

### Manual Handling

```go
embedding, found, err := cacheService.GetEmbedding(ctx, content, "text")
if err != nil {
    // Log but don't fail - continue without cache
    log.Printf("Cache error: %v", err)
}

if !found {
    // Compute embedding without cache
    embedding, err = computeEmbedding(content)
}
```

## Best Practices

### 1. Set Appropriate TTLs

```go
// Short-lived content (24h)
cacheService.SetEmbedding(ctx, newsArticle, embedding, "text", 24*time.Hour)

// Long-lived content (7 days)
cacheService.SetEmbedding(ctx, documentation, embedding, "text", 7*24*time.Hour)

// Static content (30 days)
cacheService.SetEmbedding(ctx, staticImage, embedding, "image", 30*24*time.Hour)
```

### 2. Invalidate on Content Changes

```go
func UpdateDocument(id string, newContent string) error {
    // Invalidate old embedding
    if err := cacheService.InvalidateEmbedding(ctx, oldContent, "text"); err != nil {
        log.Printf("Failed to invalidate cache: %v", err)
    }

    // Update document
    // ...

    // Optionally pre-compute and cache new embedding
    embedding, _ := generateEmbedding(newContent)
    cacheService.SetEmbedding(ctx, newContent, embedding, "text", 0)

    return nil
}
```

### 3. Monitor Cache Performance

```go
// Periodically log cache stats
func MonitorCache() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        stats, err := cacheService.GetCacheStats(context.Background())
        if err != nil {
            log.Printf("Failed to get cache stats: %v", err)
            continue
        }

        log.Printf("Cache: %d entries, %.2f%% hit rate",
            stats.Size, stats.HitRate*100)

        // Alert if hit rate is low
        if stats.HitRate < 0.5 && stats.Hits+stats.Misses > 100 {
            log.Printf("WARNING: Low cache hit rate (%.2f%%)", stats.HitRate*100)
        }
    }
}
```

### 4. Handle Redis Failures Gracefully

```go
// Check health before critical operations
if !cacheService.HealthCheck(ctx) {
    log.Println("Cache unhealthy, operating without cache")
    // Continue without cache
}

// Re-enable after recovery
if !cacheService.IsEnabled() {
    if err := redisClient.Ping(ctx).Err(); err == nil {
        cacheService.Enable()
        log.Println("Cache re-enabled")
    }
}
```

### 5. Clear Cache After Model Updates

```go
func UpdateEmbeddingModel(newModel string) error {
    // Clear all cached embeddings (they're from old model)
    if err := cacheService.ClearCache(context.Background()); err != nil {
        log.Printf("Failed to clear cache: %v", err)
    }

    // Update model
    // ...

    return nil
}
```

## Testing

Run tests with Redis available:

```bash
# Start Redis
docker run -d -p 6379:6379 redis:latest

# Run tests
go test -v ./internal/services -run TestEmbedding

# Run benchmarks
go test -bench=BenchmarkEmbeddingCache ./internal/services
```

## Troubleshooting

### Cache Not Working

1. **Check Redis connection**:
   ```go
   healthy := cacheService.HealthCheck(ctx)
   fmt.Printf("Cache healthy: %v\n", healthy)
   ```

2. **Check if enabled**:
   ```go
   fmt.Printf("Cache enabled: %v\n", cacheService.IsEnabled())
   ```

3. **Check stats**:
   ```go
   stats, _ := cacheService.GetCacheStats(ctx)
   fmt.Printf("Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)
   ```

### Low Hit Rate

- Check TTL (might be too short)
- Verify content is identical between set/get
- Check if cache is being cleared too frequently
- Monitor cache size (might be evicted due to memory limits)

### High Memory Usage

- Reduce TTL
- Implement cache size limits
- Use Redis eviction policies (LRU)

## Production Deployment

### Environment Variables

```bash
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=your-secure-password
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10
```

### Redis Configuration

```conf
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru  # Evict least recently used
save 900 1                     # Persistence for cache warmth
```

### Monitoring

Monitor these metrics:
- Cache hit rate (target: >80%)
- Cache size (watch for memory limits)
- Redis CPU usage
- Redis connection pool saturation
- Average latency

### High Availability

For production, use Redis Cluster or Sentinel:

```go
client := redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{":26379", ":26380", ":26381"},
})
```

## License

Part of BusinessOS backend.
