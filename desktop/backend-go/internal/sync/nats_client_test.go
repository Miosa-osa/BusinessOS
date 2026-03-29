package sync

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNATSClient_Disabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     "nats://localhost:4222",
		Enabled: false,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.False(t, client.enabled)
	assert.False(t, client.IsConnected())

	// Health check should always return true when disabled
	assert.True(t, client.HealthCheck())

	// Cleanup
	err = client.Close()
	assert.NoError(t, err)
}

func TestNewNATSClient_DefaultConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := DefaultNATSConfig()
	assert.Equal(t, DefaultURL, config.URL)
	assert.True(t, config.Enabled)
	assert.Equal(t, DefaultReconnectWait, config.ReconnectWait)
	assert.Equal(t, DefaultMaxReconnects, config.MaxReconnects)
	assert.Equal(t, DefaultTimeout, config.Timeout)

	// Don't actually connect in this test
	config.Enabled = false
	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.Close()
	assert.NoError(t, err)
}

func TestNATSClient_PublishSync_Disabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     "nats://localhost:4222",
		Enabled: false,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	defer client.Close()

	// Publishing should succeed but do nothing
	data := map[string]string{"key": "value"}
	err = client.PublishSync("test.subject", data)
	assert.NoError(t, err)
}

func TestNATSClient_Subscribe_Disabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     "nats://localhost:4222",
		Enabled: false,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	defer client.Close()

	// Subscribing should succeed but do nothing
	err = client.Subscribe("test.subject", func(msg *nats.Msg) {
		t.Fatal("should not receive messages when disabled")
	})
	assert.NoError(t, err)
}

// Integration tests - require NATS server running
// Set NATS_URL environment variable or skip these tests

func skipIfNoNATS(t *testing.T) string {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}
	return natsURL
}

func TestNATSClient_Integration_Connect(t *testing.T) {
	natsURL := skipIfNoNATS(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     natsURL,
		Enabled: true,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	assert.True(t, client.IsConnected())
	assert.True(t, client.HealthCheck())

	stats := client.Stats()
	assert.True(t, stats.Enabled)
	assert.True(t, stats.Connected)
	assert.NotEmpty(t, stats.URL)
	assert.NotEmpty(t, stats.ServerID)
}

func TestNATSClient_Integration_PublishSubscribe(t *testing.T) {
	natsURL := skipIfNoNATS(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     natsURL,
		Enabled: true,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	subject := "test.publish.subscribe"
	received := make(chan []byte, 1)

	// Subscribe first
	err = client.Subscribe(subject, func(msg *nats.Msg) {
		received <- msg.Data
	})
	require.NoError(t, err)

	// Give subscription time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish message
	testData := map[string]string{
		"message": "hello world",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	err = client.PublishSync(subject, testData)
	require.NoError(t, err)

	// Wait for message
	select {
	case data := <-received:
		var receivedData map[string]string
		err := json.Unmarshal(data, &receivedData)
		require.NoError(t, err)
		assert.Equal(t, testData["message"], receivedData["message"])
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestNATSClient_Integration_WorkspaceEvents(t *testing.T) {
	natsURL := skipIfNoNATS(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     natsURL,
		Enabled: true,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	received := make(chan *SyncEvent, 1)

	// Subscribe to workspace events
	err = client.SubscribeWorkspaceEvents(func(event *SyncEvent) error {
		received <- event
		return nil
	})
	require.NoError(t, err)

	// Give subscription time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish workspace created event
	workspaceID := "ws-12345"
	userID := "user-67890"
	workspaceData := map[string]interface{}{
		"name": "Test Workspace",
		"mode": "2d",
	}

	err = client.PublishWorkspaceCreated(workspaceID, userID, workspaceData)
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-received:
		assert.Equal(t, "workspace.created", event.EventType)
		assert.Equal(t, "workspace", event.EntityType)
		assert.Equal(t, workspaceID, event.EntityID)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, "businessos", event.Source)
		assert.Equal(t, int64(1), event.SyncVersion)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for workspace event")
	}
}

func TestNATSClient_Integration_ConflictEvents(t *testing.T) {
	natsURL := skipIfNoNATS(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     natsURL,
		Enabled: true,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	received := make(chan *ConflictEvent, 1)

	// Subscribe to conflict events
	err = client.SubscribeConflictEvents(func(conflict *ConflictEvent) error {
		received <- conflict
		return nil
	})
	require.NoError(t, err)

	// Give subscription time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish conflict event
	conflictEvent := &ConflictEvent{
		EntityType:      "workspace",
		EntityID:        "ws-12345",
		ConflictFields:  []string{"name", "layout"},
		LocalData:       json.RawMessage(`{"name":"Local Name"}`),
		RemoteData:      json.RawMessage(`{"name":"Remote Name"}`),
		LocalUpdatedAt:  time.Now().Add(-1 * time.Minute),
		RemoteUpdatedAt: time.Now(),
		DetectedAt:      time.Now(),
	}

	err = client.PublishConflict(conflictEvent)
	require.NoError(t, err)

	// Wait for event
	select {
	case conflict := <-received:
		assert.Equal(t, "workspace", conflict.EntityType)
		assert.Equal(t, "ws-12345", conflict.EntityID)
		assert.ElementsMatch(t, []string{"name", "layout"}, conflict.ConflictFields)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for conflict event")
	}
}

func TestNATSClient_Integration_MultipleSubscriptions(t *testing.T) {
	natsURL := skipIfNoNATS(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     natsURL,
		Enabled: true,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	workspaceReceived := make(chan *SyncEvent, 1)
	projectReceived := make(chan *SyncEvent, 1)

	// Subscribe to workspace events
	err = client.SubscribeWorkspaceEvents(func(event *SyncEvent) error {
		workspaceReceived <- event
		return nil
	})
	require.NoError(t, err)

	// Subscribe to project events
	err = client.SubscribeProjectEvents(func(event *SyncEvent) error {
		projectReceived <- event
		return nil
	})
	require.NoError(t, err)

	// Give subscriptions time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish workspace event
	err = client.PublishWorkspaceCreated("ws-1", "user-1", map[string]string{"name": "Test"})
	require.NoError(t, err)

	// Publish project event
	err = client.PublishProjectCreated("proj-1", "user-1", map[string]string{"title": "Project"})
	require.NoError(t, err)

	// Verify both events received
	wsReceived := false
	projReceived := false

	timeout := time.After(2 * time.Second)
	for !wsReceived || !projReceived {
		select {
		case event := <-workspaceReceived:
			assert.Equal(t, "workspace.created", event.EventType)
			wsReceived = true
		case event := <-projectReceived:
			assert.Equal(t, "project.created", event.EventType)
			projReceived = true
		case <-timeout:
			t.Fatal("timeout waiting for events")
		}
	}
}

func TestNATSClient_Integration_Unsubscribe(t *testing.T) {
	natsURL := skipIfNoNATS(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     natsURL,
		Enabled: true,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	subject := "test.unsubscribe"
	received := make(chan []byte, 10)

	// Subscribe
	err = client.Subscribe(subject, func(msg *nats.Msg) {
		received <- msg.Data
	})
	require.NoError(t, err)

	// Give subscription time to set up
	time.Sleep(100 * time.Millisecond)

	// Publish first message
	err = client.PublishSync(subject, map[string]string{"message": "first"})
	require.NoError(t, err)

	// Should receive first message
	select {
	case <-received:
		// OK
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for first message")
	}

	// Unsubscribe
	err = client.Unsubscribe(subject)
	require.NoError(t, err)

	// Give time for unsubscribe to process
	time.Sleep(100 * time.Millisecond)

	// Publish second message
	err = client.PublishSync(subject, map[string]string{"message": "second"})
	require.NoError(t, err)

	// Should NOT receive second message
	select {
	case <-received:
		t.Fatal("should not receive message after unsubscribe")
	case <-time.After(500 * time.Millisecond):
		// OK - timeout means message not received
	}
}

func TestNATSClient_Integration_Stats(t *testing.T) {
	natsURL := skipIfNoNATS(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     natsURL,
		Enabled: true,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Get initial stats
	stats := client.Stats()
	assert.True(t, stats.Enabled)
	assert.True(t, stats.Connected)
	assert.Equal(t, 0, stats.Subscriptions)

	// Add a subscription
	err = client.Subscribe("test.stats", func(msg *nats.Msg) {})
	require.NoError(t, err)

	// Check stats again
	stats = client.Stats()
	assert.Equal(t, 1, stats.Subscriptions)

	// Publish some messages
	for i := 0; i < 5; i++ {
		err = client.PublishSync("test.stats", map[string]int{"count": i})
		require.NoError(t, err)
	}

	// Stats should show messages sent
	stats = client.Stats()
	assert.True(t, stats.OutMsgs >= 5)
}

func TestNATSClient_Close_Multiple(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := &NATSConfig{
		URL:     "nats://localhost:4222",
		Enabled: false,
	}

	client, err := NewNATSClient(config, logger)
	require.NoError(t, err)

	// Close multiple times should not error
	err = client.Close()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestSyncEvent_Marshal(t *testing.T) {
	event := &SyncEvent{
		EventType:   "workspace.created",
		EntityType:  "workspace",
		EntityID:    "ws-123",
		UserID:      "user-456",
		Data:        map[string]string{"name": "Test"},
		Source:      "businessos",
		Timestamp:   time.Now(),
		SyncVersion: 1,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded SyncEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, event.EventType, decoded.EventType)
	assert.Equal(t, event.EntityType, decoded.EntityType)
	assert.Equal(t, event.EntityID, decoded.EntityID)
}

func TestConflictEvent_Marshal(t *testing.T) {
	conflict := &ConflictEvent{
		EntityType:      "workspace",
		EntityID:        "ws-123",
		ConflictFields:  []string{"name", "layout"},
		LocalData:       json.RawMessage(`{"name":"Local"}`),
		RemoteData:      json.RawMessage(`{"name":"Remote"}`),
		LocalUpdatedAt:  time.Now().Add(-1 * time.Minute),
		RemoteUpdatedAt: time.Now(),
		DetectedAt:      time.Now(),
	}

	data, err := json.Marshal(conflict)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded ConflictEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, conflict.EntityType, decoded.EntityType)
	assert.Equal(t, conflict.EntityID, decoded.EntityID)
	assert.ElementsMatch(t, conflict.ConflictFields, decoded.ConflictFields)
}
