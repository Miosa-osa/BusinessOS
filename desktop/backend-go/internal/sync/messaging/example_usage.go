package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

// Example demonstrates how to use the NATS client for sync operations

// ExamplePublishUserCreated shows how to publish a user created event
func ExamplePublishUserCreated() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Connect to NATS
	client, err := NewNATSClient("nats://localhost:4222", logger)
	if err != nil {
		logger.Error("failed to connect", "error", err)
		return
	}
	defer client.Close()

	// Create streams
	ctx := context.Background()
	if err := client.CreateStreams(ctx); err != nil {
		logger.Error("failed to create streams", "error", err)
		return
	}

	// Publish user created event
	userID := uuid.New()
	event := UserCreatedEvent{
		UserID:    userID,
		Email:     "user@example.com",
		FullName:  "John Doe",
		CreatedAt: time.Now(),
	}

	subject := SubjectUserCreated(userID)
	if err := client.Publish(ctx, subject, event); err != nil {
		logger.Error("failed to publish", "error", err)
		return
	}

	logger.Info("user created event published", "user_id", userID)
}

// ExampleSubscribeToWorkspaceEvents shows how to subscribe to workspace events
func ExampleSubscribeToWorkspaceEvents() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Connect to NATS
	client, err := NewNATSClient("nats://localhost:4222", logger)
	if err != nil {
		logger.Error("failed to connect", "error", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.CreateStreams(ctx); err != nil {
		logger.Error("failed to create streams", "error", err)
		return
	}

	// Define message handler
	handler := func(msg *Message) error {
		logger.Info("received workspace event",
			"subject", msg.Subject,
			"msg_id", msg.MessageID,
			"sequence", msg.Sequence)

		// Process event based on data
		var event WorkspaceEvent
		if err := unmarshalMessage(msg, &event); err != nil {
			return fmt.Errorf("failed to unmarshal: %w", err)
		}

		logger.Info("workspace event",
			"workspace_id", event.WorkspaceID,
			"action", event.Action)

		return nil
	}

	// Subscribe to all workspace events
	if err := client.Subscribe(ctx, SubjectAllWorkspaces, handler); err != nil {
		logger.Error("failed to subscribe", "error", err)
		return
	}

	logger.Info("subscribed to workspace events")

	// Keep running (in real app, use proper shutdown signal)
	select {}
}

// ExampleBuildStatusUpdates shows how to handle build status updates from OSA
func ExampleBuildStatusUpdates() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewNATSClient("nats://localhost:4222", logger)
	if err != nil {
		logger.Error("failed to connect", "error", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.CreateStreams(ctx); err != nil {
		logger.Error("failed to create streams", "error", err)
		return
	}

	// Subscribe to all build events
	handler := func(msg *Message) error {
		var buildEvent BuildEvent
		if err := unmarshalMessage(msg, &buildEvent); err != nil {
			return err
		}

		logger.Info("build event",
			"build_id", buildEvent.BuildID,
			"status", buildEvent.Status,
			"progress", buildEvent.Progress)

		// Update database with build status
		// ... database update logic ...

		// Publish SSE update to connected clients
		// ... SSE broadcast logic ...

		return nil
	}

	if err := client.Subscribe(ctx, SubjectAllBuilds, handler); err != nil {
		logger.Error("failed to subscribe", "error", err)
		return
	}

	logger.Info("listening for build status updates")
	select {}
}

// ExampleBidirectionalSync shows complete bidirectional sync pattern
func ExampleBidirectionalSync() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewNATSClient("nats://localhost:4222", logger)
	if err != nil {
		logger.Error("failed to connect", "error", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.CreateStreams(ctx); err != nil {
		logger.Error("failed to create streams", "error", err)
		return
	}

	// Handler for BusinessOS → OSA events
	businessOSHandler := func(msg *Message) error {
		logger.Info("BusinessOS event", "subject", msg.Subject)
		// Forward to OSA API
		return nil
	}

	// Handler for OSA → BusinessOS events
	osaHandler := func(msg *Message) error {
		logger.Info("OSA event", "subject", msg.Subject)
		// Update BusinessOS database
		return nil
	}

	// Subscribe to both directions
	if err := client.Subscribe(ctx, "businessos.>", businessOSHandler); err != nil {
		logger.Error("failed to subscribe to BusinessOS events", "error", err)
		return
	}

	if err := client.Subscribe(ctx, "osa.>", osaHandler); err != nil {
		logger.Error("failed to subscribe to OSA events", "error", err)
		return
	}

	logger.Info("bidirectional sync active")
	select {}
}

// ExampleWithRetry shows how to implement retry logic
func ExampleWithRetry() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewNATSClient("nats://localhost:4222", logger)
	if err != nil {
		logger.Error("failed to connect", "error", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	client.CreateStreams(ctx)

	// Handler with retry logic
	handler := func(msg *Message) error {
		logger.Info("processing message",
			"subject", msg.Subject,
			"delivered", msg.NumDelivered)

		// Simulate processing
		if msg.NumDelivered < 3 {
			// Fail first 2 attempts to trigger retry
			logger.Warn("simulating failure", "attempt", msg.NumDelivered)
			return fmt.Errorf("simulated error")
		}

		logger.Info("processing succeeded", "attempt", msg.NumDelivered)
		return nil
	}

	// Consumer will automatically retry up to MaxDeliver (5) times
	if err := client.Subscribe(ctx, SubjectAllUsers, handler); err != nil {
		logger.Error("failed to subscribe", "error", err)
		return
	}

	// Publish a message that will be retried
	userID := uuid.New()
	if err := client.Publish(ctx, SubjectUserCreated(userID), map[string]string{
		"id": userID.String(),
	}); err != nil {
		logger.Error("failed to publish", "error", err)
		return
	}

	time.Sleep(5 * time.Second) // Wait for retries
}

// ExampleMonitoring shows how to monitor client health
func ExampleMonitoring() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewNATSClient("nats://localhost:4222", logger)
	if err != nil {
		logger.Error("failed to connect", "error", err)
		return
	}
	defer client.Close()

	// Periodically check stats
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := client.Stats()

		logger.Info("NATS client stats",
			"connected", stats.Connected,
			"in_msgs", stats.InMsgs,
			"out_msgs", stats.OutMsgs,
			"in_bytes", stats.InBytes,
			"out_bytes", stats.OutBytes,
			"reconnects", stats.Reconnects,
			"active_subs", stats.ActiveSubscriptions)

		if !stats.Connected {
			logger.Error("NATS client disconnected!")
		}
	}
}

// Sample event types

type UserCreatedEvent struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type WorkspaceEvent struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Action      string    `json:"action"` // created, updated, deleted
	UserID      uuid.UUID `json:"user_id"`
	Timestamp   time.Time `json:"timestamp"`
}

type BuildEvent struct {
	BuildID  string `json:"build_id"`
	AppID    string `json:"app_id"`
	Status   string `json:"status"` // started, progress, completed, failed
	Progress int    `json:"progress"`
}

// Helper to unmarshal message data
func unmarshalMessage(msg *Message, v interface{}) error {
	if err := msg.Unmarshal(v); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}
	return nil
}
