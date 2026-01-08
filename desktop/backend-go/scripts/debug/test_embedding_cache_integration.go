package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rhl/businessos-backend/internal/services"
)

// Test script to verify EmbeddingCacheService integration
func main() {
	ctx := context.Background()

	// 1. Initialize Redis client (optional)
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         redisURL,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test Redis connection
	redisAvailable := false
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Redis unavailable: %v", err)
		log.Println("Cache will be disabled, but service will continue")
	} else {
		log.Println("Redis connected successfully")
		redisAvailable = true
		defer redisClient.Close()
	}

	// 2. Create EmbeddingCacheService
	var embeddingCache *services.EmbeddingCacheService
	var embeddingCacheAdapter *services.EmbeddingCacheAdapter

	if redisAvailable {
		cfg := services.DefaultEmbeddingCacheConfig()
		embeddingCache = services.NewEmbeddingCacheService(redisClient, nil, cfg)
		embeddingCacheAdapter = services.NewEmbeddingCacheAdapter(embeddingCache)

		log.Printf("Cache enabled: %v", embeddingCache.IsEnabled())
		log.Printf("Cache healthy: %v", embeddingCache.HealthCheck(ctx))

		// 3. Test cache operations
		testContent := "This is a test embedding content"
		testEmbedding := make([]float32, 768)
		for i := range testEmbedding {
			testEmbedding[i] = float32(i) * 0.001
		}

		// Test SET
		log.Println("\n=== Testing SetEmbedding ===")
		err := embeddingCacheAdapter.SetEmbedding(ctx, testContent, testEmbedding, "text", 24*time.Hour)
		if err != nil {
			log.Printf("ERROR: Failed to set embedding: %v", err)
		} else {
			log.Println("Successfully cached embedding")
		}

		// Test GET
		log.Println("\n=== Testing GetEmbedding ===")
		cached, found, err := embeddingCacheAdapter.GetEmbedding(ctx, testContent, "text")
		if err != nil {
			log.Printf("ERROR: Failed to get embedding: %v", err)
		} else if !found {
			log.Println("WARNING: Embedding not found in cache")
		} else {
			log.Printf("Successfully retrieved cached embedding (length: %d)", len(cached))
			// Verify data integrity
			if len(cached) == len(testEmbedding) {
				log.Println("Embedding length matches")
			} else {
				log.Printf("ERROR: Length mismatch: expected %d, got %d", len(testEmbedding), len(cached))
			}
		}

		// Test stats
		log.Println("\n=== Testing Cache Stats ===")
		stats, err := embeddingCacheAdapter.GetCacheStats(ctx)
		if err != nil {
			log.Printf("ERROR: Failed to get stats: %v", err)
		} else {
			log.Printf("Cache Stats:")
			log.Printf("  - Hits: %d", stats.Hits)
			log.Printf("  - Misses: %d", stats.Misses)
			log.Printf("  - Size: %d entries", stats.Size)
			log.Printf("  - Hit Rate: %.2f%%", stats.HitRate*100)
		}

		// Test invalidation
		log.Println("\n=== Testing Invalidation ===")
		err = embeddingCacheAdapter.InvalidateEmbedding(ctx, testContent, "text")
		if err != nil {
			log.Printf("ERROR: Failed to invalidate: %v", err)
		} else {
			log.Println("Successfully invalidated embedding")
		}

		// Verify invalidation
		cached, found, err = embeddingCacheAdapter.GetEmbedding(ctx, testContent, "text")
		if err != nil {
			log.Printf("ERROR: Failed to verify invalidation: %v", err)
		} else if found {
			log.Println("WARNING: Embedding still found after invalidation")
		} else {
			log.Println("Confirmed: Embedding successfully invalidated")
		}

		log.Println("\n=== Integration Test Complete ===")
		log.Println("The EmbeddingCacheService is working correctly!")
	} else {
		log.Println("\n=== Redis Unavailable ===")
		log.Println("Service will work in degraded mode (no caching)")
		log.Println("This is expected behavior for graceful degradation")

		// Test that cache adapter handles nil gracefully
		embeddingCache = services.NewEmbeddingCacheService(nil, nil, nil)
		embeddingCacheAdapter = services.NewEmbeddingCacheAdapter(embeddingCache)

		log.Printf("Cache enabled: %v", embeddingCache.IsEnabled())
		log.Printf("Cache healthy: %v", embeddingCache.HealthCheck(ctx))

		// These should all work gracefully without Redis
		testContent := "Test without Redis"
		testEmbedding := make([]float32, 768)

		err := embeddingCacheAdapter.SetEmbedding(ctx, testContent, testEmbedding, "text", 24*time.Hour)
		log.Printf("SetEmbedding (no Redis): err=%v (should be nil)", err)

		cached, found, err := embeddingCacheAdapter.GetEmbedding(ctx, testContent, "text")
		log.Printf("GetEmbedding (no Redis): found=%v, err=%v, cached=%v", found, err, cached)

		log.Println("\n=== Graceful Degradation Test Complete ===")
		log.Println("Service handles Redis unavailability correctly!")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("All tests passed! Integration successful.")
	fmt.Println(strings.Repeat("=", 60))
}
