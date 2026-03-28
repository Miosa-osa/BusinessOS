// Package middleware provides HTTP middleware for Gin framework.
// This file implements global concurrency limiting (admission control).
package middleware

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/concurrency"
	"github.com/rhl/businessos-backend/internal/observability"
)

var (
	globalSemaphore     *concurrency.Semaphore
	globalSemaphoreOnce sync.Once
)

// InitGlobalSemaphore initializes the global concurrency semaphore.
// Must be called during bootstrap before ConcurrencyLimitMiddleware().
//
// Parameters:
//   - maxSlots: maximum concurrent requests (from config.GlobalMaxConcurrent)
//   - telemetry: observability.Telemetry instance for metrics
func InitGlobalSemaphore(maxSlots int, tel *observability.Telemetry) {
	globalSemaphoreOnce.Do(func() {
		globalSemaphore = concurrency.New(maxSlots, tel)
		slog.Info("Global semaphore initialized", "max_slots", maxSlots)
	})
}

// ConcurrencyLimitMiddleware enforces global concurrent request limits.
// Returns 503 Service Unavailable when semaphore is full.
//
// Usage in bootstrap.go (after InitGlobalSemaphore):
//   middleware.InitGlobalSemaphore(cfg.GlobalMaxConcurrent, telemetry)
//   router.Use(middleware.ConcurrencyLimitMiddleware())
//
// WvdA compliant: bounded wait (5s timeout) in semaphore.Acquire()
func ConcurrencyLimitMiddleware() gin.HandlerFunc {
	if globalSemaphore == nil {
		slog.Error("ConcurrencyLimitMiddleware: globalSemaphore is nil, middleware disabled")
		// Return a no-op middleware if semaphore not initialized
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Acquire semaphore slot
		err := globalSemaphore.Acquire(c.Request.Context())
		if err != nil {
			// Timeout or cancellation: semaphore full
			slog.Warn("ConcurrencyLimitMiddleware: request rejected",
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"error", err,
				"available_slots", globalSemaphore.Available(),
			)

			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service temporarily unavailable (concurrency limit exceeded)",
				"code":  "CONCURRENCY_LIMIT_EXCEEDED",
			})
			return
		}

		// Slot acquired: ensure release on handler exit
		defer globalSemaphore.Release()

		// Continue to next handler
		c.Next()
	}
}

// GetSemaphoreStats returns current semaphore statistics.
// Used for health checks and monitoring dashboards.
//
// Endpoint example:
//   GET /api/internal/health/semaphore
func GetSemaphoreStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalSemaphore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Semaphore not initialized",
			})
			return
		}

		stats := globalSemaphore.GetStats()
		c.JSON(http.StatusOK, stats)
	}
}
