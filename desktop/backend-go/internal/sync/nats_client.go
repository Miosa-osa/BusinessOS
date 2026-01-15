package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NATS subject patterns for OSA Phase 3 sync
const (
	// Workspace sync subjects
	SubjectWorkspaceAll     = "osa.sync.workspace.*"
	SubjectWorkspaceCreated = "osa.sync.workspace.created"
	SubjectWorkspaceUpdated = "osa.sync.workspace.updated"
	SubjectWorkspaceDeleted = "osa.sync.workspace.deleted"

	// Project sync subjects
	SubjectProjectAll     = "osa.sync.project.*"
	SubjectProjectCreated = "osa.sync.project.created"
	SubjectProjectUpdated = "osa.sync.project.updated"
	SubjectProjectDeleted = "osa.sync.project.deleted"

	// Task sync subjects
	SubjectTaskAll     = "osa.sync.task.*"
	SubjectTaskCreated = "osa.sync.task.created"
	SubjectTaskUpdated = "osa.sync.task.updated"
	SubjectTaskDeleted = "osa.sync.task.deleted"

	// Conflict notification subject
	SubjectConflict = "osa.sync.conflict"

	// Default NATS connection settings
	DefaultURL            = "nats://localhost:4222"
	DefaultReconnectWait  = 2 * time.Second
	DefaultMaxReconnects  = 10
	DefaultTimeout        = 30 * time.Second
	DefaultPingInterval   = 20 * time.Second
	DefaultMaxPingsOut    = 3
)

// NATSClient manages NATS connections for real-time sync events
type NATSClient struct {
	conn    *nats.Conn
	url     string
	enabled bool
	logger  *slog.Logger

	mu            sync.RWMutex
	subscriptions map[string]*nats.Subscription
	closed        bool
}

// NATSConfig holds configuration for NATS client
type NATSConfig struct {
	URL            string
	Enabled        bool
	ReconnectWait  time.Duration
	MaxReconnects  int
	Timeout        time.Duration
	PingInterval   time.Duration
	MaxPingsOut    int
}

// DefaultNATSConfig returns default NATS configuration
func DefaultNATSConfig() *NATSConfig {
	return &NATSConfig{
		URL:           DefaultURL,
		Enabled:       true,
		ReconnectWait: DefaultReconnectWait,
		MaxReconnects: DefaultMaxReconnects,
		Timeout:       DefaultTimeout,
		PingInterval:  DefaultPingInterval,
		MaxPingsOut:   DefaultMaxPingsOut,
	}
}

// NewNATSClient creates a new NATS client for sync operations
func NewNATSClient(config *NATSConfig, logger *slog.Logger) (*NATSClient, error) {
	if config == nil {
		config = DefaultNATSConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	client := &NATSClient{
		url:           config.URL,
		enabled:       config.Enabled,
		logger:        logger,
		subscriptions: make(map[string]*nats.Subscription),
	}

	if !config.Enabled {
		logger.Info("NATS client disabled, operating in standalone mode")
		return client, nil
	}

	if err := client.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// Connect establishes connection to NATS server
func (c *NATSClient) Connect(ctx context.Context) error {
	if !c.enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && c.conn.IsConnected() {
		return nil
	}

	c.logger.Info("connecting to NATS", "url", c.url)

	opts := []nats.Option{
		nats.Name("BusinessOS-OSA-Sync"),
		nats.Timeout(DefaultTimeout),
		nats.ReconnectWait(DefaultReconnectWait),
		nats.MaxReconnects(DefaultMaxReconnects),
		nats.PingInterval(DefaultPingInterval),
		nats.MaxPingsOutstanding(DefaultMaxPingsOut),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				c.logger.Warn("NATS disconnected", "error", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			c.logger.Info("NATS reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			c.logger.Info("NATS connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			c.logger.Error("NATS error",
				"subject", sub.Subject,
				"error", err)
		}),
	}

	conn, err := nats.Connect(c.url, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS at %s: %w", c.url, err)
	}

	c.conn = conn
	c.logger.Info("connected to NATS",
		"url", conn.ConnectedUrl(),
		"server_id", conn.ConnectedServerId())

	return nil
}

// PublishSync publishes a sync event to NATS
func (c *NATSClient) PublishSync(subject string, data interface{}) error {
	if !c.enabled {
		c.logger.Debug("NATS disabled, skipping publish", "subject", subject)
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	if c.conn == nil || !c.conn.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	// Marshal data to JSON
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Publish message
	if err := c.conn.Publish(subject, payload); err != nil {
		c.logger.Error("failed to publish message",
			"subject", subject,
			"error", err)
		return fmt.Errorf("failed to publish: %w", err)
	}

	c.logger.Debug("message published",
		"subject", subject,
		"size_bytes", len(payload))

	return nil
}

// Subscribe subscribes to a subject and processes messages with the provided handler
func (c *NATSClient) Subscribe(subject string, handler func(msg *nats.Msg)) error {
	if !c.enabled {
		c.logger.Debug("NATS disabled, skipping subscribe", "subject", subject)
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	if c.conn == nil || !c.conn.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	// Check if already subscribed
	if _, exists := c.subscriptions[subject]; exists {
		return fmt.Errorf("already subscribed to subject: %s", subject)
	}

	// Create subscription with error handling wrapper
	sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("panic in message handler",
					"subject", msg.Subject,
					"panic", r)
			}
		}()

		c.logger.Debug("message received",
			"subject", msg.Subject,
			"size_bytes", len(msg.Data))

		handler(msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Store subscription
	c.subscriptions[subject] = sub

	c.logger.Info("subscribed to subject", "subject", subject)
	return nil
}

// Unsubscribe removes a subscription from a subject
func (c *NATSClient) Unsubscribe(subject string) error {
	if !c.enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	sub, exists := c.subscriptions[subject]
	if !exists {
		return fmt.Errorf("not subscribed to subject: %s", subject)
	}

	if err := sub.Unsubscribe(); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	delete(c.subscriptions, subject)
	c.logger.Info("unsubscribed from subject", "subject", subject)

	return nil
}

// Close gracefully closes the NATS connection
func (c *NATSClient) Close() error {
	if !c.enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Unsubscribe from all subjects
	for subject, sub := range c.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			c.logger.Warn("failed to unsubscribe",
				"subject", subject,
				"error", err)
		}
	}
	c.subscriptions = make(map[string]*nats.Subscription)

	// Close connection
	if c.conn != nil {
		c.conn.Close()
		c.logger.Info("NATS connection closed")
	}

	return nil
}

// HealthCheck verifies NATS connection health
func (c *NATSClient) HealthCheck() bool {
	if !c.enabled {
		return true // Always healthy when disabled
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.conn != nil && c.conn.IsConnected()
}

// IsConnected returns whether the client is connected to NATS
func (c *NATSClient) IsConnected() bool {
	if !c.enabled {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.conn != nil && c.conn.IsConnected()
}

// Stats returns connection statistics
func (c *NATSClient) Stats() *NATSStats {
	if !c.enabled {
		return &NATSStats{Enabled: false}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &NATSStats{
		Enabled:       c.enabled,
		Connected:     false,
		Subscriptions: len(c.subscriptions),
	}

	if c.conn != nil && c.conn.IsConnected() {
		stats.Connected = true
		stats.URL = c.conn.ConnectedUrl()
		stats.ServerID = c.conn.ConnectedServerId()

		connStats := c.conn.Stats()
		stats.InMsgs = connStats.InMsgs
		stats.OutMsgs = connStats.OutMsgs
		stats.InBytes = connStats.InBytes
		stats.OutBytes = connStats.OutBytes
		stats.Reconnects = connStats.Reconnects
	}

	return stats
}

// NATSStats contains NATS connection statistics
type NATSStats struct {
	Enabled       bool
	Connected     bool
	URL           string
	ServerID      string
	InMsgs        uint64
	OutMsgs       uint64
	InBytes       uint64
	OutBytes      uint64
	Reconnects    uint64
	Subscriptions int
}

// SyncEvent represents a sync event payload
type SyncEvent struct {
	EventType   string      `json:"event_type"`
	EntityType  string      `json:"entity_type"`
	EntityID    string      `json:"entity_id"`
	UserID      string      `json:"user_id"`
	Data        interface{} `json:"data"`
	Source      string      `json:"source"`       // "businessos" or "osa"
	Timestamp   time.Time   `json:"timestamp"`
	SyncVersion int64       `json:"sync_version"`
}

// ConflictEvent represents a conflict notification
type ConflictEvent struct {
	EntityType      string          `json:"entity_type"`
	EntityID        string          `json:"entity_id"`
	ConflictFields  []string        `json:"conflict_fields"`
	LocalData       json.RawMessage `json:"local_data"`
	RemoteData      json.RawMessage `json:"remote_data"`
	LocalUpdatedAt  time.Time       `json:"local_updated_at"`
	RemoteUpdatedAt time.Time       `json:"remote_updated_at"`
	DetectedAt      time.Time       `json:"detected_at"`
}

// Helper functions for publishing specific event types

// PublishWorkspaceCreated publishes a workspace created event
func (c *NATSClient) PublishWorkspaceCreated(workspaceID, userID string, data interface{}) error {
	event := &SyncEvent{
		EventType:   "workspace.created",
		EntityType:  "workspace",
		EntityID:    workspaceID,
		UserID:      userID,
		Data:        data,
		Source:      "businessos",
		Timestamp:   time.Now(),
		SyncVersion: 1,
	}
	return c.PublishSync(SubjectWorkspaceCreated, event)
}

// PublishWorkspaceUpdated publishes a workspace updated event
func (c *NATSClient) PublishWorkspaceUpdated(workspaceID, userID string, data interface{}, syncVersion int64) error {
	event := &SyncEvent{
		EventType:   "workspace.updated",
		EntityType:  "workspace",
		EntityID:    workspaceID,
		UserID:      userID,
		Data:        data,
		Source:      "businessos",
		Timestamp:   time.Now(),
		SyncVersion: syncVersion,
	}
	return c.PublishSync(SubjectWorkspaceUpdated, event)
}

// PublishProjectCreated publishes a project created event
func (c *NATSClient) PublishProjectCreated(projectID, userID string, data interface{}) error {
	event := &SyncEvent{
		EventType:   "project.created",
		EntityType:  "project",
		EntityID:    projectID,
		UserID:      userID,
		Data:        data,
		Source:      "businessos",
		Timestamp:   time.Now(),
		SyncVersion: 1,
	}
	return c.PublishSync(SubjectProjectCreated, event)
}

// PublishTaskCreated publishes a task created event
func (c *NATSClient) PublishTaskCreated(taskID, userID string, data interface{}) error {
	event := &SyncEvent{
		EventType:   "task.created",
		EntityType:  "task",
		EntityID:    taskID,
		UserID:      userID,
		Data:        data,
		Source:      "businessos",
		Timestamp:   time.Now(),
		SyncVersion: 1,
	}
	return c.PublishSync(SubjectTaskCreated, event)
}

// PublishConflict publishes a conflict notification
func (c *NATSClient) PublishConflict(conflict *ConflictEvent) error {
	return c.PublishSync(SubjectConflict, conflict)
}

// SubscribeWorkspaceEvents subscribes to all workspace sync events
func (c *NATSClient) SubscribeWorkspaceEvents(handler func(event *SyncEvent) error) error {
	return c.Subscribe(SubjectWorkspaceAll, func(msg *nats.Msg) {
		var event SyncEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			c.logger.Error("failed to unmarshal workspace event",
				"error", err,
				"data", string(msg.Data))
			return
		}

		if err := handler(&event); err != nil {
			c.logger.Error("handler failed for workspace event",
				"event_type", event.EventType,
				"entity_id", event.EntityID,
				"error", err)
		}
	})
}

// SubscribeProjectEvents subscribes to all project sync events
func (c *NATSClient) SubscribeProjectEvents(handler func(event *SyncEvent) error) error {
	return c.Subscribe(SubjectProjectAll, func(msg *nats.Msg) {
		var event SyncEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			c.logger.Error("failed to unmarshal project event",
				"error", err,
				"data", string(msg.Data))
			return
		}

		if err := handler(&event); err != nil {
			c.logger.Error("handler failed for project event",
				"event_type", event.EventType,
				"entity_id", event.EntityID,
				"error", err)
		}
	})
}

// SubscribeConflictEvents subscribes to conflict notifications
func (c *NATSClient) SubscribeConflictEvents(handler func(conflict *ConflictEvent) error) error {
	return c.Subscribe(SubjectConflict, func(msg *nats.Msg) {
		var conflict ConflictEvent
		if err := json.Unmarshal(msg.Data, &conflict); err != nil {
			c.logger.Error("failed to unmarshal conflict event",
				"error", err,
				"data", string(msg.Data))
			return
		}

		if err := handler(&conflict); err != nil {
			c.logger.Error("handler failed for conflict event",
				"entity_type", conflict.EntityType,
				"entity_id", conflict.EntityID,
				"error", err)
		}
	})
}
