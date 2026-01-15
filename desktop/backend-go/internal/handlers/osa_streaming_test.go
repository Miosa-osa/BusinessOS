package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestOSAStreamingHandler_StreamBuildProgress(t *testing.T) {
	logger := slog.Default()
	eventBus := services.NewBuildEventBus(logger)
	handler := NewOSAStreamingHandler(eventBus, logger)

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set userID (simulating auth)
	userID := uuid.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})

	appID := uuid.New()
	router.GET("/stream/build/:app_id", handler.StreamBuildProgress)

	// Create test server
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Create HTTP client with SSE support
	url := ts.URL + "/stream/build/" + appID.String()
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Verify SSE headers
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
	assert.Equal(t, "keep-alive", resp.Header.Get("Connection"))

	// Start reading SSE stream in goroutine
	receivedEvents := make([]string, 0)
	var mu sync.Mutex
	done := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil {
				close(done)
				return
			}
			if n > 0 {
				mu.Lock()
				receivedEvents = append(receivedEvents, string(buf[:n]))
				mu.Unlock()
			}
		}
	}()

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	// Publish test event
	event := services.BuildEvent{
		ID:              uuid.New(),
		AppID:           appID,
		EventType:       "build_progress",
		Phase:           "building",
		ProgressPercent: 50,
		StatusMessage:   "Building project...",
		Timestamp:       time.Now(),
	}
	eventBus.Publish(event)

	// Wait for event to be received
	time.Sleep(200 * time.Millisecond)

	// Verify events were received
	mu.Lock()
	assert.Greater(t, len(receivedEvents), 0, "should receive at least one event")
	eventData := strings.Join(receivedEvents, "")
	assert.Contains(t, eventData, "data:")
	mu.Unlock()

	// Cancel context to close connection
	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for stream to close")
	}
}

func TestOSAStreamingHandler_GetStreamStats(t *testing.T) {
	logger := slog.Default()
	eventBus := services.NewBuildEventBus(logger)
	handler := NewOSAStreamingHandler(eventBus, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	userID := uuid.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})

	router.GET("/stream/stats", handler.GetStreamStats)

	// Create test request
	req := httptest.NewRequest("GET", "/stream/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "total_subscribers")
}

func TestOSAStreamingHandler_GetAppStreamStats(t *testing.T) {
	logger := slog.Default()
	eventBus := services.NewBuildEventBus(logger)
	handler := NewOSAStreamingHandler(eventBus, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	userID := uuid.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})

	appID := uuid.New()
	router.GET("/stream/stats/:app_id", handler.GetAppStreamStats)

	// Subscribe to app to have stats
	ctx := context.Background()
	sub := eventBus.Subscribe(ctx, userID, appID)
	defer eventBus.Unsubscribe(sub.ID)

	// Create test request
	req := httptest.NewRequest("GET", "/stream/stats/"+appID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), appID.String())
	assert.Contains(t, w.Body.String(), "subscriber_count")
}

func TestOSAStreamingHandler_Unauthorized(t *testing.T) {
	logger := slog.Default()
	eventBus := services.NewBuildEventBus(logger)
	handler := NewOSAStreamingHandler(eventBus, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	appID := uuid.New()
	router.GET("/stream/build/:app_id", handler.StreamBuildProgress)

	// Request without userID in context
	req := httptest.NewRequest("GET", "/stream/build/"+appID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOSAStreamingHandler_InvalidAppID(t *testing.T) {
	logger := slog.Default()
	eventBus := services.NewBuildEventBus(logger)
	handler := NewOSAStreamingHandler(eventBus, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	userID := uuid.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})

	router.GET("/stream/build/:app_id", handler.StreamBuildProgress)

	// Request with invalid app_id
	req := httptest.NewRequest("GET", "/stream/build/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// BenchmarkSSEStreaming benchmarks SSE streaming with multiple concurrent clients
func BenchmarkSSEStreaming(b *testing.B) {
	logger := slog.Default()
	eventBus := services.NewBuildEventBus(logger)

	appID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	// Subscribe multiple clients
	subscribers := make([]*services.BuildEventSubscriber, 100)
	for i := 0; i < 100; i++ {
		subscribers[i] = eventBus.Subscribe(ctx, userID, appID)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		event := services.BuildEvent{
			ID:              uuid.New(),
			AppID:           appID,
			EventType:       "build_progress",
			Phase:           "building",
			ProgressPercent: i % 100,
			StatusMessage:   "Building...",
			Timestamp:       time.Now(),
		}
		eventBus.Publish(event)
	}

	b.StopTimer()

	// Cleanup
	for _, sub := range subscribers {
		eventBus.Unsubscribe(sub.ID)
	}
}

// BenchmarkEventBusPublish benchmarks raw event bus publishing
func BenchmarkEventBusPublish(b *testing.B) {
	logger := slog.Default()
	eventBus := services.NewBuildEventBus(logger)

	appID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	// Single subscriber
	sub := eventBus.Subscribe(ctx, userID, appID)
	defer eventBus.Unsubscribe(sub.ID)

	// Drain events in background
	go func() {
		for range sub.Events {
			// Discard
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		event := services.BuildEvent{
			ID:              uuid.New(),
			AppID:           appID,
			EventType:       "build_progress",
			ProgressPercent: i % 100,
			Timestamp:       time.Now(),
		}
		eventBus.Publish(event)
	}
}
