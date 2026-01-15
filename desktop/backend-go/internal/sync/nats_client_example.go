package sync

import (
	"log/slog"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

// Example: Basic NATS Client Setup
func ExampleBasicSetup() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create NATS client with default configuration
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// Check connection health
	if client.HealthCheck() {
		logger.Info("NATS client is healthy")
	}
}

// Example: Disabled NATS Client (for development/testing)
func ExampleDisabledClient() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create disabled client - won't attempt connection
	config := &NATSConfig{
		URL:     "nats://localhost:4222",
		Enabled: false, // Disable NATS
	}

	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// All operations will no-op gracefully
	_ = client.PublishSync("test.subject", map[string]string{"key": "value"})
}

// Example: Custom Configuration
func ExampleCustomConfig() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Custom configuration
	config := &NATSConfig{
		URL:           "nats://nats-server:4222",
		Enabled:       true,
		ReconnectWait: 5 * time.Second,
		MaxReconnects: 20,
		Timeout:       60 * time.Second,
	}

	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()
}

// Example: Publishing Workspace Events
func ExamplePublishWorkspaceEvent() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// Publish workspace created event
	workspaceData := map[string]interface{}{
		"name":          "My Workspace",
		"mode":          "2d",
		"template_type": "business_os",
		"settings":      map[string]string{"theme": "dark"},
	}

	err = client.PublishWorkspaceCreated("ws-12345", "user-67890", workspaceData)
	if err != nil {
		logger.Error("failed to publish workspace created", "error", err)
		return
	}

	logger.Info("workspace created event published")

	// Publish workspace updated event
	err = client.PublishWorkspaceUpdated("ws-12345", "user-67890", workspaceData, 2)
	if err != nil {
		logger.Error("failed to publish workspace updated", "error", err)
		return
	}

	logger.Info("workspace updated event published")
}

// Example: Subscribing to Workspace Events
func ExampleSubscribeWorkspaceEvents() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// Subscribe to all workspace events
	err = client.SubscribeWorkspaceEvents(func(event *SyncEvent) error {
		logger.Info("received workspace event",
			"event_type", event.EventType,
			"entity_id", event.EntityID,
			"user_id", event.UserID,
			"sync_version", event.SyncVersion)

		// Process the event based on type
		switch event.EventType {
		case "workspace.created":
			// Handle workspace creation
			logger.Info("workspace created", "entity_id", event.EntityID)
		case "workspace.updated":
			// Handle workspace update
			logger.Info("workspace updated",
				"entity_id", event.EntityID,
				"version", event.SyncVersion)
		case "workspace.deleted":
			// Handle workspace deletion
			logger.Info("workspace deleted", "entity_id", event.EntityID)
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to subscribe to workspace events", "error", err)
		return
	}

	logger.Info("subscribed to workspace events")
}

// Example: Publishing and Subscribing to Conflict Events
func ExampleConflictEvents() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// Subscribe to conflict notifications
	err = client.SubscribeConflictEvents(func(conflict *ConflictEvent) error {
		logger.Warn("conflict detected",
			"entity_type", conflict.EntityType,
			"entity_id", conflict.EntityID,
			"conflict_fields", conflict.ConflictFields)

		// Handle conflict - e.g., queue for manual review
		// or apply automatic resolution strategy
		return nil
	})

	if err != nil {
		logger.Error("failed to subscribe to conflict events", "error", err)
		return
	}

	// Later, when a conflict is detected
	conflict := &ConflictEvent{
		EntityType:      "workspace",
		EntityID:        "ws-12345",
		ConflictFields:  []string{"name", "layout"},
		LocalData:       []byte(`{"name":"Local Name"}`),
		RemoteData:      []byte(`{"name":"Remote Name"}`),
		LocalUpdatedAt:  time.Now().Add(-1 * time.Minute),
		RemoteUpdatedAt: time.Now(),
		DetectedAt:      time.Now(),
	}

	err = client.PublishConflict(conflict)
	if err != nil {
		logger.Error("failed to publish conflict", "error", err)
		return
	}

	logger.Info("conflict notification published")
}

// Example: Custom Message Subscription
func ExampleCustomSubscription() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// Subscribe to custom subject with raw message handler
	subject := "osa.sync.custom.events"
	err = client.Subscribe(subject, func(msg *nats.Msg) {
		logger.Info("received message",
			"subject", msg.Subject,
			"data", string(msg.Data))

		// Process message...
	})

	if err != nil {
		logger.Error("failed to subscribe", "error", err)
		return
	}

	logger.Info("subscribed to custom subject", "subject", subject)
}

// Example: Publishing Generic Sync Events
func ExampleGenericSyncEvent() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// Publish a custom sync event
	event := &SyncEvent{
		EventType:   "project.created",
		EntityType:  "project",
		EntityID:    "proj-123",
		UserID:      "user-456",
		Data: map[string]interface{}{
			"title":       "New Project",
			"description": "Project description",
			"status":      "active",
		},
		Source:      "businessos",
		Timestamp:   time.Now(),
		SyncVersion: 1,
	}

	err = client.PublishSync(SubjectProjectCreated, event)
	if err != nil {
		logger.Error("failed to publish event", "error", err)
		return
	}

	logger.Info("sync event published")
}

// Example: Monitoring and Health Checks
func ExampleMonitoring() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// Check connection status
	if client.IsConnected() {
		logger.Info("NATS is connected")
	} else {
		logger.Warn("NATS is not connected")
	}

	// Health check
	if client.HealthCheck() {
		logger.Info("NATS health check passed")
	} else {
		logger.Error("NATS health check failed")
	}

	// Get detailed statistics
	stats := client.Stats()
	logger.Info("NATS statistics",
		"enabled", stats.Enabled,
		"connected", stats.Connected,
		"url", stats.URL,
		"in_msgs", stats.InMsgs,
		"out_msgs", stats.OutMsgs,
		"reconnects", stats.Reconnects,
		"subscriptions", stats.Subscriptions)
}

// Example: Integration with Sync Service
func ExampleSyncServiceIntegration() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	natsClient, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer natsClient.Close()

	// Subscribe to workspace events from OSA
	err = natsClient.SubscribeWorkspaceEvents(func(event *SyncEvent) error {
		// Only process events from OSA (not our own echoes)
		if event.Source != "osa" {
			return nil
		}

		logger.Info("processing OSA workspace event",
			"event_type", event.EventType,
			"entity_id", event.EntityID)

		// Integrate with your sync service
		// Example: Update local database, trigger conflict detection, etc.
		switch event.EventType {
		case "workspace.updated":
			// Handle remote workspace update
			// - Fetch local version
			// - Detect conflicts
			// - Apply resolution strategy
			// - Update local state
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to setup OSA event subscription", "error", err)
		return
	}

	// When local workspace is updated, publish to NATS
	localWorkspaceUpdated := func(workspaceID, userID string, data interface{}, version int64) {
		err := natsClient.PublishWorkspaceUpdated(workspaceID, userID, data, version)
		if err != nil {
			logger.Error("failed to publish workspace update",
				"workspace_id", workspaceID,
				"error", err)
		}
	}

	// Simulate local update
	localWorkspaceUpdated("ws-12345", "user-67890", map[string]string{
		"name": "Updated Workspace",
	}, 5)
}

// Example: Graceful Shutdown
func ExampleGracefulShutdown() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}

	// Subscribe to some events
	_ = client.SubscribeWorkspaceEvents(func(event *SyncEvent) error {
		return nil
	})

	// ... application runs ...

	// Graceful shutdown
	logger.Info("shutting down NATS client")
	if err := client.Close(); err != nil {
		logger.Error("error closing NATS client", "error", err)
	} else {
		logger.Info("NATS client closed successfully")
	}
}

// Example: Error Handling and Retries
func ExampleErrorHandling() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultNATSConfig()
	client, err := NewNATSClient(config, logger)
	if err != nil {
		logger.Error("failed to create NATS client", "error", err)
		return
	}
	defer client.Close()

	// Subscribe with error handling
	err = client.SubscribeWorkspaceEvents(func(event *SyncEvent) error {
		// Process event with error handling
		if err := processEvent(event); err != nil {
			logger.Error("failed to process event",
				"event_type", event.EventType,
				"entity_id", event.EntityID,
				"error", err)
			// Return error to trigger NATS retry
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error("failed to subscribe", "error", err)
		return
	}

	// Publish with automatic reconnection handling
	// NATS client automatically handles reconnections
	for i := 0; i < 10; i++ {
		err := client.PublishWorkspaceCreated("ws-"+string(rune(i)), "user-1", map[string]string{
			"name": "Workspace " + string(rune(i)),
		})

		if err != nil {
			// Log error but NATS will handle reconnection automatically
			logger.Error("publish failed, will retry automatically",
				"attempt", i,
				"error", err)
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

// Placeholder function for example
func processEvent(event *SyncEvent) error {
	// Implement your event processing logic
	return nil
}
