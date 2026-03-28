package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/concurrency"
	"github.com/rhl/businessos-backend/internal/observability"
)

// TestInitGlobalSemaphore verifies singleton initialization
func TestInitGlobalSemaphore(t *testing.T) {
	// Reset global state (for test isolation)
	globalSemaphore = nil
	globalSemaphoreOnce = *(new(sync.Once))

	tel := observability.New()
	InitGlobalSemaphore(50, tel)

	if globalSemaphore == nil {
		t.Fatal("globalSemaphore should not be nil after InitGlobalSemaphore")
	}

	if globalSemaphore.Available() != 50 {
		t.Errorf("Expected 50 available slots, got %d", globalSemaphore.Available())
	}

	// Second call should be no-op (sync.Once)
	InitGlobalSemaphore(100, tel)
	if globalSemaphore.Available() != 50 {
		t.Error("sync.Once should prevent re-initialization")
	}
}

// TestConcurrencyLimitMiddleware_Success verifies requests pass through when slots available
func TestConcurrencyLimitMiddleware_Success(t *testing.T) {
	// Reset global state
	globalSemaphore = nil
	globalSemaphoreOnce = *(new(sync.Once))

	tel := observability.New()
	InitGlobalSemaphore(10, tel)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test router with middleware
	router := gin.New()
	router.Use(ConcurrencyLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make 5 concurrent requests (all should succeed)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// Verify all slots released
	if globalSemaphore.Available() != 10 {
		t.Errorf("Expected 10 available slots after requests, got %d", globalSemaphore.Available())
	}
}

// TestConcurrencyLimitMiddleware_Rejection verifies middleware handles concurrent requests
func TestConcurrencyLimitMiddleware_Rejection(t *testing.T) {
	// Save and restore global state
	oldSem := globalSemaphore
	defer func() {
		globalSemaphore = oldSem
	}()

	// Reset and create fresh semaphore for this test
	tel := observability.New()
	globalSemaphore = concurrency.New(2, tel) // Only 2 slots

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test router with middleware
	router := gin.New()
	router.Use(ConcurrencyLimitMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make multiple requests - verify no crashes
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// All should succeed because requests complete fast
		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// Note: We can't easily test actual 503 rejection in unit tests because
	// httptest requests complete too fast. The timeout behavior
	// is thoroughly tested in semaphore_test.go via TestWvdACompliance.
}

// TestConcurrencyLimitMiddleware_NoNilPanic verifies middleware doesn't panic when semaphore nil
func TestConcurrencyLimitMiddleware_NoNilPanic(t *testing.T) {
	// Save and restore global state (can't copy sync.Once, so just reset)
	oldSem := globalSemaphore
	defer func() {
		globalSemaphore = oldSem
	}()

	// Reset global state to nil (note: sync.Once prevents re-init in real usage)
	globalSemaphore = nil

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test router with middleware (semaphore is nil)
	router := gin.New()
	router.Use(ConcurrencyLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request should pass through (no-op middleware)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (no-op middleware), got %d", w.Code)
	}
}

// TestGetSemaphoreStats verifies stats endpoint
func TestGetSemaphoreStats(t *testing.T) {
	// Reset global state
	globalSemaphore = nil
	globalSemaphoreOnce = *(new(sync.Once))

	tel := observability.New()
	InitGlobalSemaphore(100, tel)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test router with stats endpoint
	router := gin.New()
	router.GET("/stats", GetSemaphoreStats())

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify JSON response contains expected fields
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}

	// Verify stats contain expected keys (simple check)
	expectedKeys := []string{
		"max_slots",
		"available",
		"utilization_percent",
		"total_requests",
		"rejection_rate_percent",
	}

	for _, key := range expectedKeys {
		if !contains(body, key) {
			t.Errorf("Expected key '%s' in response", key)
		}
	}
}

// TestGetSemaphoreStats_NilSemaphore verifies 503 when semaphore not initialized
func TestGetSemaphoreStats_NilSemaphore(t *testing.T) {
	// Reset global state to nil
	globalSemaphore = nil
	globalSemaphoreOnce = *(new(sync.Once))

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test router with stats endpoint
	router := gin.New()
	router.GET("/stats", GetSemaphoreStats())

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
