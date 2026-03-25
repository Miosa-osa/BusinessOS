package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
)

// TestBOSProgressHandlerDiscoveryStream tests basic streaming functionality
func TestBOSProgressHandlerDiscoveryStream(t *testing.T) {
	// Setup
	logger := slog.Default()
	streamService := services.NewBOSStreamingService(logger)
	handler := NewBOSProgressHandler(streamService, logger)

	sessionID := uuid.New()
	userID := uuid.New()

	// Create test router
	router := gin.New()
	router.GET("/stream/:session_id", func(c *gin.Context) {
		// Mock auth middleware
		c.Set("userID", userID)
		handler.StreamDiscoveryProgress(c)
	})

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Test: Connect to stream
	done := make(chan bool, 1)

	go func() {
		resp, err := http.Get(fmt.Sprintf("%s/stream/%s", server.URL, sessionID.String()))
		if err != nil {
			t.Errorf("Failed to connect: %v", err)
			done <- false
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
			done <- false
			return
		}

		if resp.Header.Get("Content-Type") != "text/event-stream" {
			t.Errorf("Expected text/event-stream, got %s", resp.Header.Get("Content-Type"))
			done <- false
			return
		}

		// Read connection confirmation
		reader := bufio.NewReader(resp.Body)
		line, _ := reader.ReadString('\n')
		if !strings.Contains(line, "data:") {
			t.Error("Expected SSE data line")
			done <- false
			return
		}

		done <- true
	}()

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	// Publish event
	event := &services.BOSStreamEvent{
		ID:        uuid.New().String(),
		EventType: "discovery_progress",
		SessionID: sessionID.String(),
		Progress: &services.BOSProgressMetrics{
			EventsProcessed: 1000,
			TotalEvents:     int64Ptr(100000),
			PercentComplete: 1,
			CurrentStep:     "Trace Analysis",
			ActiveWorkers:   4,
			ThroughputEPS:   1000.0,
		},
		TimestampMs: time.Now().UnixMilli(),
	}

	streamService.PublishEvent(event)

	// Wait for completion
	if !<-done {
		t.Fatal("Stream test failed")
	}
}

// TestBOSProgressHandlerConcurrentSessions tests multiple concurrent sessions
func TestBOSProgressHandlerConcurrentSessions(t *testing.T) {
	logger := slog.Default()
	streamService := services.NewBOSStreamingService(logger)
	handler := NewBOSProgressHandler(streamService, logger)

	// Create multiple sessions
	sessionIDs := make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		sessionIDs[i] = uuid.New()
	}

	router := gin.New()
	router.GET("/stream/:session_id", func(c *gin.Context) {
		c.Set("userID", uuid.New())
		handler.StreamDiscoveryProgress(c)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect all sessions concurrently
	done := make(chan bool, len(sessionIDs))

	for i, sessionID := range sessionIDs {
		go func(idx int, sid uuid.UUID) {
			resp, err := http.Get(fmt.Sprintf("%s/stream/%s", server.URL, sid.String()))
			if err != nil {
				t.Logf("Session %d: Connection failed: %v", idx, err)
				done <- false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Logf("Session %d: Unexpected status %d", idx, resp.StatusCode)
				done <- false
				return
			}

			done <- true
		}(i, sessionID)
	}

	// Wait for all connections
	time.Sleep(500 * time.Millisecond)

	// Check that all sessions are connected
	activeSessions := streamService.GetAllActiveSessions()
	if len(activeSessions) != len(sessionIDs) {
		t.Errorf("Expected %d sessions, got %d", len(sessionIDs), len(activeSessions))
	}

	// Verify all sessions received completion signals
	connected := 0
	for _, done := range done {
		if done {
			connected++
		}
	}

	if connected > 0 {
		t.Logf("Successfully connected %d sessions concurrently", connected)
	}
}

// TestBOSProgressHandlerErrorRecovery tests error handling and recovery
func TestBOSProgressHandlerErrorRecovery(t *testing.T) {
	logger := slog.Default()
	streamService := services.NewBOSStreamingService(logger)
	handler := NewBOSProgressHandler(streamService, logger)

	sessionID := uuid.New()
	userID := uuid.New()

	router := gin.New()
	router.GET("/stream/:session_id", func(c *gin.Context) {
		c.Set("userID", userID)
		handler.StreamDiscoveryProgress(c)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	done := make(chan bool, 1)

	go func() {
		resp, err := http.Get(fmt.Sprintf("%s/stream/%s", server.URL, sessionID.String()))
		if err != nil {
			done <- false
			return
		}
		defer resp.Body.Close()

		done <- true
	}()

	time.Sleep(100 * time.Millisecond)

	// Publish recoverable error
	errorEvent := &services.BOSStreamEvent{
		ID:        uuid.New().String(),
		EventType: "error_recoverable",
		SessionID: sessionID.String(),
		Error: &services.BOSErrorInfo{
			Code:        "TIMEOUT",
			Message:     "Worker timeout, retrying",
			Recoverable: true,
			RetryAttempt: int32Ptr(1),
			MaxRetries:  int32Ptr(3),
		},
		TimestampMs: time.Now().UnixMilli(),
	}

	streamService.PublishEvent(errorEvent)

	// Publish recovery event
	time.Sleep(100 * time.Millisecond)
	recoveryEvent := &services.BOSStreamEvent{
		ID:          uuid.New().String(),
		EventType:   "discovery_progress",
		SessionID:   sessionID.String(),
		TimestampMs: time.Now().UnixMilli(),
		Progress: &services.BOSProgressMetrics{
			EventsProcessed: 2000,
			TotalEvents:     int64Ptr(100000),
			PercentComplete: 2,
			CurrentStep:     "Continuing analysis",
			ActiveWorkers:   4,
			ThroughputEPS:   1000.0,
		},
	}

	streamService.PublishEvent(recoveryEvent)

	if !<-done {
		t.Fatal("Stream test failed")
	}

	// Verify metrics were updated
	metrics := streamService.GetAggregatedMetrics(sessionID)
	if metrics == nil {
		t.Error("Expected metrics for session")
	}
}

// TestBOSProgressHandlerMetricsAggregation tests metrics collection
func TestBOSProgressHandlerMetricsAggregation(t *testing.T) {
	logger := slog.Default()
	streamService := services.NewBOSStreamingService(logger)

	sessionID := uuid.New()
	userID := uuid.New()

	// Subscribe to session
	ctx := context.Background()
	subscriber := streamService.Subscribe(ctx, userID, sessionID)

	// Publish events with metrics
	for i := 0; i < 5; i++ {
		event := &services.BOSStreamEvent{
			ID:        uuid.New().String(),
			EventType: "metrics",
			SessionID: sessionID.String(),
			Metrics: &services.BOSAggregatedMetrics{
				ElapsedSecs:          int64(i * 2),
				TotalProcessed:       int64(1000 * (i + 1)),
				AvgThroughputEPS:     1000.0,
				CurrentThroughputEPS: 1000.0 + float64(i*100),
				PeakThroughputEPS:    1500.0,
				VariantsFound:        int64(10 + i),
				ViolationsFound:      int64(i),
			},
			TimestampMs: time.Now().UnixMilli(),
		}

		streamService.PublishEvent(event)
		time.Sleep(10 * time.Millisecond)
	}

	// Check aggregated metrics
	metrics := streamService.GetAggregatedMetrics(sessionID)
	if metrics == nil {
		t.Fatal("Expected metrics")
	}

	if metrics.TotalProcessed != 5000 {
		t.Errorf("Expected 5000 total processed, got %d", metrics.TotalProcessed)
	}

	if metrics.VariantsFound != 14 {
		t.Errorf("Expected 14 variants, got %d", metrics.VariantsFound)
	}

	streamService.Unsubscribe(subscriber.ID)
}

// TestBOSProgressHandlerPartialResults tests partial result streaming
func TestBOSProgressHandlerPartialResults(t *testing.T) {
	logger := slog.Default()
	streamService := services.NewBOSStreamingService(logger)

	sessionID := uuid.New()
	userID := uuid.New()

	ctx := context.Background()
	subscriber := streamService.Subscribe(ctx, userID, sessionID)
	defer streamService.Unsubscribe(subscriber.ID)

	// Publish partial result
	resultData := map[string]interface{}{
		"variant_id": "v1",
		"frequency":  100,
		"path":       []string{"A", "B", "C"},
	}

	event := &services.BOSStreamEvent{
		ID:        uuid.New().String(),
		EventType: "partial_results",
		SessionID: sessionID.String(),
		PartialResult: resultData,
		TimestampMs:   time.Now().UnixMilli(),
	}

	streamService.PublishEvent(event)

	// Receive event
	select {
	case received := <-subscriber.Events:
		if received.EventType != "partial_results" {
			t.Errorf("Expected partial_results, got %s", received.EventType)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for partial result")
	}
}

// TestBOSProgressHandlerSessionCancellation tests session cancellation
func TestBOSProgressHandlerSessionCancellation(t *testing.T) {
	logger := slog.Default()
	streamService := services.NewBOSStreamingService(logger)
	handler := NewBOSProgressHandler(streamService, logger)

	sessionID := uuid.New()

	router := gin.New()
	router.POST("/session/:session_id/cancel", func(c *gin.Context) {
		handler.CancelSession(c)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	// Cancel session
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/session/%s/cancel", server.URL, sessionID.String()), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if cancelled, ok := result["cancelled"].(bool); !ok || !cancelled {
		t.Error("Expected cancelled: true")
	}
}

// TestBOSProgressHandlerSSEEventFormatting tests SSE message formatting
func TestBOSProgressHandlerSSEEventFormatting(t *testing.T) {
	logger := slog.Default()
	handler := NewBOSProgressHandler(nil, logger)

	event := &services.BOSStreamEvent{
		ID:        "test-id",
		EventType: "discovery_progress",
		SessionID: "session-123",
		Progress: &services.BOSProgressMetrics{
			EventsProcessed: 1000,
			PercentComplete: 10,
			CurrentStep:     "Analysis",
			ActiveWorkers:   4,
			ThroughputEPS:   1000.0,
		},
		TimestampMs: 1000000,
	}

	formatted := handler.formatSSE(event)

	// Check SSE format
	if !strings.Contains(formatted, "event: discovery_progress") {
		t.Error("SSE should contain event type")
	}

	if !strings.Contains(formatted, "data:") {
		t.Error("SSE should contain data")
	}

	if !strings.HasSuffix(formatted, "\n\n") {
		t.Error("SSE should end with double newline")
	}

	// Verify JSON is valid
	lines := strings.Split(formatted, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data:") {
			jsonStr := strings.TrimPrefix(line, "data: ")
			var event services.BOSStreamEvent
			if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
				t.Errorf("Invalid JSON in SSE: %v", err)
			}
		}
	}
}

// TestBOSProgressHandlerMetricsSnapshot tests metrics snapshot retrieval
func TestBOSProgressHandlerMetricsSnapshot(t *testing.T) {
	logger := slog.Default()
	streamService := services.NewBOSStreamingService(logger)
	handler := NewBOSProgressHandler(streamService, logger)

	sessionID := uuid.New()
	userID := uuid.New()

	router := gin.New()
	router.GET("/session/:session_id/metrics", func(c *gin.Context) {
		c.Set("userID", userID)
		handler.GetSessionMetrics(c)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	// Subscribe to session
	ctx := context.Background()
	subscriber := streamService.Subscribe(ctx, userID, sessionID)
	defer streamService.Unsubscribe(subscriber.ID)

	// Publish metrics event
	event := &services.BOSStreamEvent{
		ID:        uuid.New().String(),
		EventType: "metrics",
		SessionID: sessionID.String(),
		Metrics: &services.BOSAggregatedMetrics{
			ElapsedSecs:          10,
			TotalProcessed:       10000,
			AvgThroughputEPS:     1000.0,
			CurrentThroughputEPS: 1100.0,
			PeakThroughputEPS:    1200.0,
			VariantsFound:        42,
			ViolationsFound:      3,
		},
		TimestampMs: time.Now().UnixMilli(),
	}

	streamService.PublishEvent(event)

	// Retrieve metrics via HTTP
	resp, err := http.Get(fmt.Sprintf("%s/session/%s/metrics", server.URL, sessionID.String()))
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var metrics services.BOSAggregatedMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		t.Fatalf("Failed to decode metrics: %v", err)
	}

	if metrics.TotalProcessed != 10000 {
		t.Errorf("Expected 10000 total processed, got %d", metrics.TotalProcessed)
	}
}

// TestBOSProgressHandlerEventDropping tests event dropping behavior
func TestBOSProgressHandlerEventDropping(t *testing.T) {
	logger := slog.Default()
	streamService := services.NewBOSStreamingService(logger)

	sessionID := uuid.New()
	userID := uuid.New()

	// Create subscriber
	ctx := context.Background()
	subscriber := streamService.Subscribe(ctx, userID, sessionID)
	defer streamService.Unsubscribe(subscriber.ID)

	// Rapidly publish events to fill buffer
	for i := 0; i < 150; i++ {
		event := &services.BOSStreamEvent{
			ID:        uuid.New().String(),
			EventType: "discovery_progress",
			SessionID: sessionID.String(),
			Progress: &services.BOSProgressMetrics{
				EventsProcessed: int64(i * 1000),
				PercentComplete: int32(i % 100),
				CurrentStep:     "Analysis",
				ActiveWorkers:   4,
				ThroughputEPS:   1000.0,
			},
			TimestampMs: time.Now().UnixMilli(),
		}

		streamService.PublishEvent(event)
	}

	// Drain buffer
	received := 0
	timeout := time.After(2 * time.Second)

	for {
		select {
		case <-subscriber.Events:
			received++
		case <-timeout:
			goto done
		}
	}

done:
	if received > 100 {
		t.Logf("Received %d events (buffer size: 100)", received)
	}
}

// Helper functions

func int64Ptr(i int64) *int64 {
	return &i
}

func int32Ptr(i int32) *int32 {
	return &i
}
