package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BuildEvent represents a build progress event to be streamed to clients
type BuildEvent struct {
	ID              uuid.UUID              `json:"id"`
	AppID           uuid.UUID              `json:"app_id"`
	WorkspaceID     *uuid.UUID             `json:"workspace_id,omitempty"`
	EventType       string                 `json:"event_type"`
	Phase           string                 `json:"phase,omitempty"`
	ProgressPercent int                    `json:"progress_percent"`
	StatusMessage   string                 `json:"status_message,omitempty"`
	Data            map[string]interface{} `json:"data,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// BuildEventSubscriber represents a client subscribed to build events
type BuildEventSubscriber struct {
	ID       string
	AppID    uuid.UUID
	UserID   uuid.UUID
	Events   chan BuildEvent
	ctx      context.Context
	cancelFn context.CancelFunc
}

// BuildEventBus manages pub/sub for build progress events
type BuildEventBus struct {
	subscribers map[string]*BuildEventSubscriber // subscriber ID -> subscriber
	mu          sync.RWMutex
	logger      *slog.Logger
}

// NewBuildEventBus creates a new build event bus
func NewBuildEventBus(logger *slog.Logger) *BuildEventBus {
	if logger == nil {
		logger = slog.Default()
	}
	return &BuildEventBus{
		subscribers: make(map[string]*BuildEventSubscriber),
		logger:      logger.With("component", "build_event_bus"),
	}
}

// Subscribe creates a new subscription for build events for a specific app
// Returns a subscriber that can receive events via the Events channel
func (b *BuildEventBus) Subscribe(ctx context.Context, userID, appID uuid.UUID) *BuildEventSubscriber {
	// Create cancellable context for this subscription
	subCtx, cancel := context.WithCancel(ctx)

	subscriber := &BuildEventSubscriber{
		ID:       uuid.New().String(),
		AppID:    appID,
		UserID:   userID,
		Events:   make(chan BuildEvent, 100), // Buffered channel to prevent blocking
		ctx:      subCtx,
		cancelFn: cancel,
	}

	b.mu.Lock()
	b.subscribers[subscriber.ID] = subscriber
	b.mu.Unlock()

	b.logger.Info("client subscribed to build events",
		"subscriber_id", subscriber.ID,
		"user_id", userID,
		"app_id", appID,
		"total_subscribers", len(b.subscribers),
	)

	// Start cleanup goroutine
	go b.cleanupOnContextDone(subscriber)

	return subscriber
}

// Unsubscribe removes a subscriber from the bus
func (b *BuildEventBus) Unsubscribe(subscriberID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if sub, exists := b.subscribers[subscriberID]; exists {
		// Cancel the context
		sub.cancelFn()
		// Close the channel
		close(sub.Events)
		// Remove from map
		delete(b.subscribers, subscriberID)

		b.logger.Info("client unsubscribed from build events",
			"subscriber_id", subscriberID,
			"app_id", sub.AppID,
			"total_subscribers", len(b.subscribers),
		)
	}
}

// Publish broadcasts a build event to all subscribers for the given app
func (b *BuildEventBus) Publish(event BuildEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	subscriberCount := 0
	for _, sub := range b.subscribers {
		// Only send to subscribers watching this specific app
		if sub.AppID == event.AppID {
			select {
			case sub.Events <- event:
				subscriberCount++
			case <-sub.ctx.Done():
				// Subscriber context cancelled, skip
				b.logger.Debug("skipping cancelled subscriber",
					"subscriber_id", sub.ID,
					"app_id", event.AppID,
				)
			default:
				// Channel full, drop event to prevent blocking
				b.logger.Warn("subscriber channel full, dropping event",
					"subscriber_id", sub.ID,
					"app_id", event.AppID,
					"event_type", event.EventType,
				)
			}
		}
	}

	b.logger.Debug("published build event",
		"app_id", event.AppID,
		"event_type", event.EventType,
		"phase", event.Phase,
		"progress", event.ProgressPercent,
		"subscribers_notified", subscriberCount,
	)
}

// cleanupOnContextDone removes subscriber when context is cancelled
func (b *BuildEventBus) cleanupOnContextDone(sub *BuildEventSubscriber) {
	<-sub.ctx.Done()
	b.Unsubscribe(sub.ID)
}

// GetSubscriberCount returns the current number of active subscribers
func (b *BuildEventBus) GetSubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}

// GetSubscriberCountForApp returns the number of subscribers for a specific app
func (b *BuildEventBus) GetSubscriberCountForApp(appID uuid.UUID) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	for _, sub := range b.subscribers {
		if sub.AppID == appID {
			count++
		}
	}
	return count
}

// FormatSSE formats a BuildEvent as a Server-Sent Event message
func FormatSSE(event BuildEvent) string {
	data, err := json.Marshal(event)
	if err != nil {
		return ""
	}
	return "data: " + string(data) + "\n\n"
}

// SendHeartbeat sends a heartbeat/keep-alive event
func SendHeartbeat() string {
	return ": heartbeat\n\n"
}
