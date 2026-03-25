// Package middleware provides HTTP request idempotency handling for exactly-once delivery
package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// IdempotencyConfig holds idempotency middleware configuration
type IdempotencyConfig struct {
	// TTL for cached responses (24 hours by default)
	TTL time.Duration

	// HTTP status codes to cache (successful responses only)
	CacheableStatuses []int

	// Maximum number of concurrent requests to track
	MaxConcurrentKeys int

	// Paths to exclude from idempotency tracking
	ExcludePaths []string
}

// DefaultIdempotencyConfig returns production-safe defaults
func DefaultIdempotencyConfig() *IdempotencyConfig {
	return &IdempotencyConfig{
		TTL: 24 * time.Hour,
		CacheableStatuses: []int{
			http.StatusOK,
			http.StatusCreated,
			http.StatusAccepted,
			http.StatusNoContent,
		},
		MaxConcurrentKeys: 10000,
		ExcludePaths: []string{
			"/health",
			"/ready",
			"/metrics",
		},
	}
}

// IdempotencyEntry represents a cached response
type IdempotencyEntry struct {
	Status      int            `json:"status"`
	Headers     map[string]any `json:"headers"`
	Body        string         `json:"body"`
	StoredAt    time.Time      `json:"stored_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	RequestHash string         `json:"request_hash"` // Hash of request method + path + body
}

// IdempotencyStore manages in-memory idempotency key cache with TTL
type IdempotencyStore struct {
	mu          sync.RWMutex
	cache       map[string]*IdempotencyEntry
	processing  map[string]chan *IdempotencyEntry // In-flight request locks
	config      *IdempotencyConfig
	stopCleanup chan struct{}
}

// NewIdempotencyStore creates a new idempotency store
func NewIdempotencyStore(config *IdempotencyConfig) *IdempotencyStore {
	if config == nil {
		config = DefaultIdempotencyConfig()
	}

	store := &IdempotencyStore{
		cache:       make(map[string]*IdempotencyEntry),
		processing:  make(map[string]chan *IdempotencyEntry),
		config:      config,
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup routine
	go store.cleanupExpiredKeys()

	return store
}

// Get retrieves a cached response if it exists and hasn't expired
func (s *IdempotencyStore) Get(key string) *IdempotencyEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.cache[key]
	if !exists {
		return nil
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry
}

// Store caches a response with TTL
func (s *IdempotencyStore) Store(key string, entry *IdempotencyEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.cache) >= s.config.MaxConcurrentKeys {
		return fmt.Errorf("idempotency store at capacity: %d keys", s.config.MaxConcurrentKeys)
	}

	entry.StoredAt = time.Now()
	entry.ExpiresAt = time.Now().Add(s.config.TTL)
	s.cache[key] = entry

	return nil
}

// Delete removes a key from the cache
func (s *IdempotencyStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, key)
}

// WaitForProcessing blocks until an in-flight request completes, returning its result
func (s *IdempotencyStore) WaitForProcessing(key string) *IdempotencyEntry {
	s.mu.RLock()
	ch, processing := s.processing[key]
	s.mu.RUnlock()

	if !processing {
		return nil
	}

	// Wait for the channel to be closed or a result to be sent
	entry := <-ch
	return entry
}

// LockKey marks a key as being processed (in-flight)
func (s *IdempotencyStore) LockKey(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, processing := s.processing[key]; processing {
		// Already processing, return false
		return false
	}

	s.processing[key] = make(chan *IdempotencyEntry, 1)
	return true
}

// UnlockKey marks a key as no longer processing and broadcasts the result
func (s *IdempotencyStore) UnlockKey(key string, entry *IdempotencyEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, exists := s.processing[key]; exists {
		ch <- entry
		close(ch)
		delete(s.processing, key)
	}
}

// Stats returns current cache statistics
func (s *IdempotencyStore) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]int{
		"cached_entries":     len(s.cache),
		"processing_entries": len(s.processing),
	}
}

// cleanupExpiredKeys runs periodically to remove expired entries
func (s *IdempotencyStore) cleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			deleted := 0

			for key, entry := range s.cache {
				if now.After(entry.ExpiresAt) {
					delete(s.cache, key)
					deleted++
				}
			}

			s.mu.Unlock()

		case <-s.stopCleanup:
			return
		}
	}
}

// Stop halts the cleanup routine
func (s *IdempotencyStore) Stop() {
	close(s.stopCleanup)
}

// Idempotency is the Gin middleware handler
func Idempotency(store *IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for excluded paths
		for _, excludePath := range store.config.ExcludePaths {
			if c.Request.URL.Path == excludePath {
				c.Next()
				return
			}
		}

		// Extract idempotency key header
		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			// No idempotency key, proceed normally
			c.Next()
			return
		}

		// Check if we have a cached response
		if cachedEntry := store.Get(idempotencyKey); cachedEntry != nil {
			// Return cached response
			c.JSON(cachedEntry.Status, gin.H{
				"cached":     true,
				"status":     cachedEntry.Status,
				"stored_at":  cachedEntry.StoredAt,
				"expires_at": cachedEntry.ExpiresAt,
			})
			return
		}

		// Check if request is already in flight
		if !store.LockKey(idempotencyKey) {
			// Request is being processed, wait for it
			if result := store.WaitForProcessing(idempotencyKey); result != nil {
				c.JSON(result.Status, gin.H{
					"cached":     true,
					"status":     result.Status,
					"stored_at":  result.StoredAt,
					"expires_at": result.ExpiresAt,
				})
				return
			}
		}

		// Capture request body for hashing
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		// Restore request body for handler to read
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		requestHash := hashRequest(c.Request.Method, c.Request.URL.Path, bodyBytes)

		// Proceed with the request
		c.Next()

		// Get the status code
		statusCode := c.Writer.Status()

		// Cache successful responses
		if isCacheableStatus(statusCode, store.config.CacheableStatuses) {
			entry := &IdempotencyEntry{
				Status:      statusCode,
				Body:        c.Writer.Header().Get("Content"),
				RequestHash: requestHash,
				Headers:     extractHeaders(c),
			}

			if err := store.Store(idempotencyKey, entry); err != nil {
				// Log error but don't fail the request
				c.Error(err)
			}

			// Unlock and broadcast
			store.UnlockKey(idempotencyKey, entry)
		} else {
			// Don't cache error responses, just unlock
			store.UnlockKey(idempotencyKey, nil)
		}
	}
}


// hashRequest creates a hash of method + path + body for request validation
func hashRequest(method, path string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(method))
	h.Write([]byte(path))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// isCacheableStatus checks if a status code should be cached
func isCacheableStatus(status int, cacheableStatuses []int) bool {
	for _, cs := range cacheableStatuses {
		if status == cs {
			return true
		}
	}
	return false
}

// extractHeaders extracts relevant headers from response
func extractHeaders(c *gin.Context) map[string]any {
	relevantHeaders := []string{
		"Content-Type",
		"Content-Length",
		"Location",
		"ETag",
		"Cache-Control",
	}

	headers := make(map[string]any)
	for _, h := range relevantHeaders {
		if value := c.Writer.Header().Get(h); value != "" {
			headers[h] = value
		}
	}
	return headers
}
