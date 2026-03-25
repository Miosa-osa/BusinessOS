package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
)

// BOSProgressHandler handles real-time progress streaming from BOS process mining operations.
//
// BOS processes large event logs (1M-100M events) and this handler receives streaming
// progress updates, aggregates metrics, and broadcasts to connected WebSocket clients.
//
// Architecture:
//   BOS (Rust) -> HTTP/SSE -> BOSProgressHandler -> WebSocket -> Browser UI
//
// Example curl to test:
//   curl -N -H "Authorization: Bearer <token>" \
//     http://localhost:8001/api/bos/stream/discover/550e8400-e29b-41d4-a716-446655440000
//
type BOSProgressHandler struct {
	streamService *services.BOSStreamingService
	logger        *slog.Logger
}

// Re-export service types for handler use
type BOSStreamEvent = services.BOSStreamEvent
type BOSProgressData = services.BOSProgressMetrics
type BOSMetricsData = services.BOSAggregatedMetrics
type BOSErrorData = services.BOSErrorInfo

// NewBOSProgressHandler creates a new BOS progress streaming handler
func NewBOSProgressHandler(
	streamService *services.BOSStreamingService,
	logger *slog.Logger,
) *BOSProgressHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &BOSProgressHandler{
		streamService: streamService,
		logger:        logger.With("component", "bos_progress"),
	}
}

// StreamDiscoveryProgress handles SSE streaming for BOS discovery progress
// GET /api/bos/stream/discover/:session_id
//
// This endpoint receives real-time progress events from BOS process discovery.
// Events are forwarded to connected WebSocket clients for real-time UI updates.
func (h *BOSProgressHandler) StreamDiscoveryProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	h.logger.Info("BOS discovery stream connected",
		"user_id", userUUID,
		"session_id", sessionID,
		"remote_addr", c.Request.RemoteAddr,
	)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Create cancellable context
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Subscribe to BOS events
	subscriber := h.streamService.Subscribe(ctx, userUUID, sessionID)
	defer h.streamService.Unsubscribe(subscriber.ID)

	// Get flusher for immediate writes
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.logger.Error("response writer does not support flushing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Send initial connection confirmation
	connectMsg := map[string]interface{}{
		"type":       "connected",
		"session_id": sessionID.String(),
	}
	if data, err := json.Marshal(connectMsg); err == nil {
		c.Writer.WriteString("data: " + string(data) + "\n\n")
		flusher.Flush()
	}

	// Create heartbeat ticker (30 seconds)
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Create metrics snapshot ticker (every 5 seconds)
	metricsTicker := time.NewTicker(5 * time.Second)
	defer metricsTicker.Stop()

	h.logger.Info("BOS stream started",
		"subscriber_id", subscriber.ID,
		"user_id", userUUID,
		"session_id", sessionID,
	)

	// Stream events to client
	for {
		select {
		case event, ok := <-subscriber.Events:
			if !ok {
				h.logger.Info("event channel closed, ending stream",
					"subscriber_id", subscriber.ID,
					"session_id", sessionID,
				)
				return
			}

			// Forward event to connected WebSocket clients
			h.streamService.BroadcastEvent(subscriber.ID, sessionID, event)

			// Format SSE message
			sseMessage := h.formatSSE(event)
			if _, err := c.Writer.WriteString(sseMessage); err != nil {
				if err == io.EOF || err == context.Canceled {
					h.logger.Info("client disconnected",
						"subscriber_id", subscriber.ID,
						"session_id", sessionID,
					)
				} else {
					h.logger.Error("failed to write event",
						"error", err,
						"subscriber_id", subscriber.ID,
					)
				}
				return
			}
			flusher.Flush()

			h.logger.Debug("sent BOS event to client",
				"subscriber_id", subscriber.ID,
				"event_type", event.EventType,
			)

		case <-metricsTicker.C:
			// Send periodic metrics aggregation
			aggregated := h.streamService.GetAggregatedMetrics(sessionID)
			if aggregated != nil {
				data, _ := json.Marshal(aggregated)
				c.Writer.WriteString("event: metrics_snapshot\ndata: " + string(data) + "\n\n")
				flusher.Flush()
			}

		case <-heartbeatTicker.C:
			// Send heartbeat
			heartbeat := map[string]interface{}{
				"type":      "heartbeat",
				"timestamp": time.Now().Unix(),
			}
			data, _ := json.Marshal(heartbeat)
			c.Writer.WriteString("data: " + string(data) + "\n\n")
			flusher.Flush()

		case <-ctx.Done():
			h.logger.Info("context cancelled, ending stream",
				"subscriber_id", subscriber.ID,
				"session_id", sessionID,
				"reason", ctx.Err(),
			)
			return
		}
	}
}

// StreamConformanceProgress handles SSE streaming for conformance checking progress
// GET /api/bos/stream/conformance/:session_id
func (h *BOSProgressHandler) StreamConformanceProgress(c *gin.Context) {
	// Similar to discovery but marked as conformance phase
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	h.logger.Info("BOS conformance stream connected",
		"session_id", sessionID,
	)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	subscriber := h.streamService.Subscribe(ctx, userUUID, sessionID)
	defer h.streamService.Unsubscribe(subscriber.ID)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Send initial connection
	connectMsg := map[string]interface{}{
		"type":       "connected",
		"phase":      "conformance",
		"session_id": sessionID.String(),
	}
	if data, err := json.Marshal(connectMsg); err == nil {
		c.Writer.WriteString("data: " + string(data) + "\n\n")
		flusher.Flush()
	}

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case event, ok := <-subscriber.Events:
			if !ok {
				return
			}

			h.streamService.BroadcastEvent(subscriber.ID, sessionID, event)

			sseMessage := h.formatSSE(event)
			if _, err := c.Writer.WriteString(sseMessage); err != nil {
				return
			}
			flusher.Flush()

		case <-heartbeatTicker.C:
			heartbeat := map[string]interface{}{
				"type":      "heartbeat",
				"timestamp": time.Now().Unix(),
			}
			data, _ := json.Marshal(heartbeat)
			c.Writer.WriteString("data: " + string(data) + "\n\n")
			flusher.Flush()

		case <-ctx.Done():
			return
		}
	}
}

// GetSessionMetrics returns aggregated metrics for a BOS session
// GET /api/bos/session/:session_id/metrics
func (h *BOSProgressHandler) GetSessionMetrics(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	metrics := h.streamService.GetAggregatedMetrics(sessionID)
	if metrics == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetAllActiveSessions returns list of all active BOS sessions
// GET /api/bos/sessions
func (h *BOSProgressHandler) GetAllActiveSessions(c *gin.Context) {
	sessions := h.streamService.GetAllActiveSessions()
	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// CancelSession cancels a BOS processing session
// POST /api/bos/session/:session_id/cancel
func (h *BOSProgressHandler) CancelSession(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	cancelled := h.streamService.CancelSession(sessionID)
	if !cancelled {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found or already cancelled"})
		return
	}

	h.logger.Info("session cancelled", "session_id", sessionID)
	c.JSON(http.StatusOK, gin.H{"cancelled": true, "session_id": sessionID.String()})
}

// formatSSE formats a BOS event as Server-Sent Events
func (h *BOSProgressHandler) formatSSE(event *BOSStreamEvent) string {
	eventType := "default"

	switch event.EventType {
	case "discovery_started":
		eventType = "discovery_started"
	case "discovery_progress":
		eventType = "discovery_progress"
	case "conformance_started":
		eventType = "conformance_started"
	case "conformance_progress":
		eventType = "conformance_progress"
	case "processing_complete":
		eventType = "processing_complete"
	case "partial_results":
		eventType = "partial_results"
	case "error_recoverable":
		eventType = "error_recoverable"
	case "error_fatal":
		eventType = "error_fatal"
	case "metrics":
		eventType = "metrics"
	case "heartbeat":
		eventType = "heartbeat"
	}

	data, _ := json.Marshal(event)
	return "event: " + eventType + "\ndata: " + string(data) + "\n\n"
}

// BOSSessionInfo represents session metadata
type BOSSessionInfo struct {
	SessionID     uuid.UUID `json:"session_id"`
	UserID        uuid.UUID `json:"user_id"`
	StartTime     time.Time `json:"start_time"`
	Phase         string    `json:"phase"` // discovery, conformance, complete
	ProgressPct   int32     `json:"progress_percent"`
	EventsProcessed int64   `json:"events_processed"`
}
