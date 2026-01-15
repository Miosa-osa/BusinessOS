package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
)

// OSAStreamingHandler handles SSE streaming for OSA build events
type OSAStreamingHandler struct {
	eventBus *services.BuildEventBus
	logger   *slog.Logger
}

// NewOSAStreamingHandler creates a new SSE streaming handler
func NewOSAStreamingHandler(eventBus *services.BuildEventBus, logger *slog.Logger) *OSAStreamingHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &OSAStreamingHandler{
		eventBus: eventBus,
		logger:   logger.With("component", "osa_streaming"),
	}
}

// StreamBuildProgress handles SSE connections for build progress updates
// GET /api/osa/stream/build/:app_id
func (h *OSAStreamingHandler) StreamBuildProgress(c *gin.Context) {
	// Get user ID from auth middleware
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

	// Get app ID from URL params
	appIDStr := c.Param("app_id")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID"})
		return
	}

	h.logger.Info("SSE client connecting",
		"user_id", userUUID,
		"app_id", appID,
		"remote_addr", c.Request.RemoteAddr,
	)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create cancellable context
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Subscribe to events for this app
	subscriber := h.eventBus.Subscribe(ctx, userUUID, appID)
	defer h.eventBus.Unsubscribe(subscriber.ID)

	// Get response writer flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.logger.Error("response writer does not support flushing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Send initial connection confirmation
	c.Writer.WriteString("data: {\"type\":\"connected\",\"app_id\":\"" + appID.String() + "\"}\n\n")
	flusher.Flush()

	// Create heartbeat ticker (every 30 seconds)
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	h.logger.Info("SSE stream started",
		"subscriber_id", subscriber.ID,
		"user_id", userUUID,
		"app_id", appID,
	)

	// Stream events to client
	for {
		select {
		case event, ok := <-subscriber.Events:
			if !ok {
				// Channel closed, end stream
				h.logger.Info("event channel closed, ending stream",
					"subscriber_id", subscriber.ID,
					"app_id", appID,
				)
				return
			}

			// Format and send event
			sseMessage := services.FormatSSE(event)
			if _, err := c.Writer.WriteString(sseMessage); err != nil {
				if err == io.EOF || err == context.Canceled {
					h.logger.Info("client disconnected",
						"subscriber_id", subscriber.ID,
						"app_id", appID,
					)
				} else {
					h.logger.Error("failed to write event",
						"error", err,
						"subscriber_id", subscriber.ID,
						"app_id", appID,
					)
				}
				return
			}
			flusher.Flush()

			h.logger.Debug("sent event to client",
				"subscriber_id", subscriber.ID,
				"event_type", event.EventType,
				"progress", event.ProgressPercent,
			)

		case <-heartbeatTicker.C:
			// Send heartbeat to keep connection alive
			if _, err := c.Writer.WriteString(services.SendHeartbeat()); err != nil {
				h.logger.Info("failed to send heartbeat, client disconnected",
					"subscriber_id", subscriber.ID,
					"app_id", appID,
				)
				return
			}
			flusher.Flush()

		case <-ctx.Done():
			// Context cancelled (client disconnected or server shutdown)
			h.logger.Info("context cancelled, ending stream",
				"subscriber_id", subscriber.ID,
				"app_id", appID,
				"reason", ctx.Err(),
			)
			return
		}
	}
}

// GetStreamStats returns statistics about active SSE connections
// GET /api/osa/stream/stats
func (h *OSAStreamingHandler) GetStreamStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	totalSubscribers := h.eventBus.GetSubscriberCount()

	c.JSON(http.StatusOK, gin.H{
		"total_subscribers": totalSubscribers,
		"timestamp":         time.Now(),
		"user_id":           userID,
	})
}

// GetAppStreamStats returns statistics for a specific app
// GET /api/osa/stream/stats/:app_id
func (h *OSAStreamingHandler) GetAppStreamStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	appIDStr := c.Param("app_id")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID"})
		return
	}

	subscriberCount := h.eventBus.GetSubscriberCountForApp(appID)

	c.JSON(http.StatusOK, gin.H{
		"app_id":            appID,
		"subscriber_count":  subscriberCount,
		"timestamp":         time.Now(),
		"user_id":           userID,
	})
}
