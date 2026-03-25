package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIdempotencyTest creates a test store and context
func setupIdempotencyTest() (*IdempotencyStore, *gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{"data":"test"}`)))
	return store, c, w
}

func TestIdempotencyStore_Store_And_Get(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "test-key-123"
	entry := &IdempotencyEntry{
		Status:      http.StatusOK,
		Body:        `{"result":"success"}`,
		RequestHash: "abc123",
	}

	err := store.Store(key, entry)
	require.NoError(t, err)

	retrieved := store.Get(key)
	require.NotNil(t, retrieved)
	assert.Equal(t, http.StatusOK, retrieved.Status)
	assert.Equal(t, `{"result":"success"}`, retrieved.Body)
}

func TestIdempotencyStore_Get_Expired(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "expired-key"
	entry := &IdempotencyEntry{
		Status:      http.StatusOK,
		Body:        `{"result":"success"}`,
		StoredAt:    time.Now().Add(-25 * time.Hour),
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
		RequestHash: "abc123",
	}

	// Manually insert expired entry
	store.mu.Lock()
	store.cache[key] = entry
	store.mu.Unlock()

	// Get should return nil for expired key
	retrieved := store.Get(key)
	assert.Nil(t, retrieved)
}

func TestIdempotencyStore_Delete(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "delete-key"
	entry := &IdempotencyEntry{
		Status:      http.StatusOK,
		Body:        `{"result":"success"}`,
		RequestHash: "abc123",
	}

	store.Store(key, entry)
	assert.NotNil(t, store.Get(key))

	store.Delete(key)
	assert.Nil(t, store.Get(key))
}

func TestIdempotencyStore_LockUnlock(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "lock-key"

	// First lock should succeed
	locked := store.LockKey(key)
	assert.True(t, locked)

	// Second lock should fail (already locked)
	locked2 := store.LockKey(key)
	assert.False(t, locked2)

	// Unlock
	entry := &IdempotencyEntry{
		Status:      http.StatusOK,
		Body:        `{"result":"success"}`,
		RequestHash: "abc123",
	}
	store.UnlockKey(key, entry)

	// Key should be unlocked now
	locked3 := store.LockKey(key)
	assert.True(t, locked3)
}

func TestIdempotencyStore_Stats(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	// Store some entries
	for i := 0; i < 5; i++ {
		key := "key-" + string(rune(i))
		entry := &IdempotencyEntry{
			Status:      http.StatusOK,
			Body:        "body",
			RequestHash: "hash",
		}
		store.Store(key, entry)
	}

	stats := store.Stats()
	assert.Equal(t, 5, stats["cached_entries"])
	assert.Equal(t, 0, stats["processing_entries"])
}

func TestIdempotencyStore_Capacity(t *testing.T) {
	config := DefaultIdempotencyConfig()
	config.MaxConcurrentKeys = 2
	store := NewIdempotencyStore(config)
	defer store.Stop()

	// Store up to capacity
	for i := 0; i < 2; i++ {
		key := "key-" + string(rune(i))
		entry := &IdempotencyEntry{
			Status:      http.StatusOK,
			Body:        "body",
			RequestHash: "hash",
		}
		err := store.Store(key, entry)
		require.NoError(t, err)
	}

	// Exceed capacity
	entry := &IdempotencyEntry{
		Status:      http.StatusOK,
		Body:        "body",
		RequestHash: "hash",
	}
	err := store.Store("key-exceed", entry)
	assert.Error(t, err)
}

func TestIdempotencyStore_ConcurrentAccess(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	// Concurrent stores
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(index int) {
			key := "concurrent-" + string(rune(index))
			entry := &IdempotencyEntry{
				Status:      http.StatusOK,
				Body:        "body",
				RequestHash: "hash",
			}
			store.Store(key, entry)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	stats := store.Stats()
	assert.Equal(t, 100, stats["cached_entries"])
}

func TestIdempotencyMiddleware_WithoutKey(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{"data":"test"}`)))

	// No idempotency key header
	middleware := Idempotency(store)

	middleware(c)

	// After middleware, handler should proceed normally
	c.JSON(http.StatusOK, gin.H{"result": "success"})

	// Verify response was sent
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIdempotencyMiddleware_FirstRequest(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "first-request-key"
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{"data":"test"}`)))
	c.Request.Header.Set("Idempotency-Key", key)

	middleware := Idempotency(store)
	middleware(c)

	// Handler sends response
	c.JSON(http.StatusOK, gin.H{"result": "success"})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIdempotencyMiddleware_ReturnsCached(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "cached-request-key"

	// Pre-populate cache
	entry := &IdempotencyEntry{
		Status:      http.StatusCreated,
		Body:        `{"id":"123","result":"created"}`,
		RequestHash: "abc123",
		Headers: map[string]any{
			"Content-Type": "application/json",
		},
	}
	store.Store(key, entry)

	// Make a request with the same key
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{"data":"test"}`)))
	c.Request.Header.Set("Idempotency-Key", key)

	middleware := Idempotency(store)
	middleware(c)
	c.Next()

	// Should return cached response
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "cached")
	assert.Contains(t, w.Body.String(), "true")
}

func TestIdempotencyMiddleware_ExcludedPath(t *testing.T) {
	config := DefaultIdempotencyConfig()
	store := NewIdempotencyStore(config)
	defer store.Stop()

	// /health is in excluded paths
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/health", nil)
	c.Request.Header.Set("Idempotency-Key", "some-key")

	middleware := Idempotency(store)
	middleware(c)

	// Send response to verify handler proceeds normally
	c.JSON(http.StatusOK, gin.H{"status": "ok"})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIdempotencyMiddleware_DoesNotCacheErrors(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "error-request-key"

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{"data":"test"}`)))
	c.Request.Header.Set("Idempotency-Key", key)

	middleware := Idempotency(store)
	// Set status to 400 before middleware checks
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad request"})

	middleware(c)

	// Error response should not be cached
	assert.Nil(t, store.Get(key))
}

func TestIdempotencyMiddleware_CachesSuccessResponses(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	gin.SetMode(gin.TestMode)

	successStatuses := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent,
	}

	for i, status := range successStatuses {
		key := "success-key-" + string(rune(i))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte{}))
		c.Request.Header.Set("Idempotency-Key", key)

		// Set status BEFORE middleware
		c.AbortWithStatusJSON(status, gin.H{"result": "success"})

		middleware := Idempotency(store)
		middleware(c)

		// Should be cached
		cached := store.Get(key)
		require.NotNil(t, cached, "Status %d should be cached", status)
		assert.Equal(t, status, cached.Status)
	}
}

func TestIdempotencyMiddleware_ConcurrentRequests(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "concurrent-request-key"
	results := make(chan int, 10)

	// Pre-populate cache
	entry := &IdempotencyEntry{
		Status:      http.StatusOK,
		Body:        `{"result":"cached"}`,
		RequestHash: "abc123",
	}
	store.Store(key, entry)

	// Simulate 10 concurrent requests with same key
	for i := 0; i < 10; i++ {
		go func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte{}))
			c.Request.Header.Set("Idempotency-Key", key)

			middleware := Idempotency(store)
			middleware(c)

			// Send response
			c.JSON(http.StatusOK, gin.H{"result": "success"})

			results <- w.Code
		}()
	}

	// All should return cached response (200)
	for i := 0; i < 10; i++ {
		code := <-results
		assert.Equal(t, http.StatusOK, code)
	}
}

func TestIdempotencyMiddleware_TTLExpiration(t *testing.T) {
	config := DefaultIdempotencyConfig()
	config.TTL = 100 * time.Millisecond // Very short TTL for testing
	store := NewIdempotencyStore(config)
	defer store.Stop()

	key := "ttl-test-key"
	entry := &IdempotencyEntry{
		Status:      http.StatusOK,
		Body:        `{"result":"success"}`,
		RequestHash: "abc123",
	}

	store.Store(key, entry)
	assert.NotNil(t, store.Get(key))

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	assert.Nil(t, store.Get(key))
}

func TestIdempotencyMiddleware_RequestHashValidation(t *testing.T) {
	method := "POST"
	path := "/api/users"
	body1 := []byte(`{"name":"Alice"}`)
	body2 := []byte(`{"name":"Bob"}`)

	hash1 := hashRequest(method, path, body1)
	hash2 := hashRequest(method, path, body2)

	// Different bodies should produce different hashes
	assert.NotEqual(t, hash1, hash2)

	// Same inputs should produce same hash
	hash1_again := hashRequest(method, path, body1)
	assert.Equal(t, hash1, hash1_again)
}

func TestDefaultIdempotencyConfig(t *testing.T) {
	config := DefaultIdempotencyConfig()

	assert.Equal(t, 24*time.Hour, config.TTL)
	assert.Contains(t, config.CacheableStatuses, http.StatusOK)
	assert.Contains(t, config.CacheableStatuses, http.StatusCreated)
	assert.Equal(t, 10000, config.MaxConcurrentKeys)
	assert.NotEmpty(t, config.ExcludePaths)
}

func TestIdempotencyStore_CleanupExpiredKeys(t *testing.T) {
	config := DefaultIdempotencyConfig()
	config.TTL = 50 * time.Millisecond
	store := NewIdempotencyStore(config)
	defer store.Stop()

	// Store multiple entries
	for i := 0; i < 5; i++ {
		key := "cleanup-key-" + string(rune(i))
		entry := &IdempotencyEntry{
			Status:      http.StatusOK,
			Body:        "body",
			RequestHash: "hash",
		}
		store.Store(key, entry)
	}

	stats := store.Stats()
	assert.Equal(t, 5, stats["cached_entries"])

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Force cleanup (normally runs every hour)
	store.mu.Lock()
	now := time.Now()
	deleted := 0
	for key, entry := range store.cache {
		if now.After(entry.ExpiresAt) {
			delete(store.cache, key)
			deleted++
		}
	}
	store.mu.Unlock()

	assert.Greater(t, deleted, 0)
}

func TestIdempotencyStore_WaitForProcessing(t *testing.T) {
	store := NewIdempotencyStore(DefaultIdempotencyConfig())
	defer store.Stop()

	key := "processing-key"

	// Lock the key
	locked := store.LockKey(key)
	assert.True(t, locked)

	done := make(chan bool)

	// Simulate concurrent request waiting for first request
	go func() {
		// This should block until UnlockKey is called
		result := store.WaitForProcessing(key)
		assert.NotNil(t, result)
		assert.Equal(t, http.StatusOK, result.Status)
		done <- true
	}()

	// Give the goroutine time to start waiting
	time.Sleep(100 * time.Millisecond)

	// Unlock with a result
	entry := &IdempotencyEntry{
		Status:      http.StatusOK,
		Body:        `{"result":"success"}`,
		RequestHash: "abc123",
	}
	store.UnlockKey(key, entry)

	// Wait for the goroutine to complete
	<-done
}


func TestCacheableStatus(t *testing.T) {
	config := DefaultIdempotencyConfig()

	assert.True(t, isCacheableStatus(http.StatusOK, config.CacheableStatuses))
	assert.True(t, isCacheableStatus(http.StatusCreated, config.CacheableStatuses))
	assert.False(t, isCacheableStatus(http.StatusBadRequest, config.CacheableStatuses))
	assert.False(t, isCacheableStatus(http.StatusInternalServerError, config.CacheableStatuses))
}

func TestExtractHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Content-Length", "1024")
	c.Writer.Header().Set("X-Custom-Header", "should-not-be-extracted")

	headers := extractHeaders(c)

	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "1024", headers["Content-Length"])
	assert.NotContains(t, headers, "X-Custom-Header")
}
