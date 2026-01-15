package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNATSClient_Connection tests basic connection functionality
func TestNATSClient_Connection(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := NewNATSClient(natsURL, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	assert.True(t, client.IsConnected())

	stats := client.Stats()
	assert.True(t, stats.Connected)
	assert.NotEmpty(t, stats.URL)
}

// TestNATSClient_CreateStreams tests stream creation
func TestNATSClient_CreateStreams(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := NewNATSClient(natsURL, logger)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	require.NoError(t, err)

	// Creating again should update, not error
	err = client.CreateStreams(ctx)
	require.NoError(t, err)
}

// TestNATSClient_PublishSubscribe tests message publishing and consumption
func TestNATSClient_PublishSubscribe(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := NewNATSClient(natsURL, logger)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	require.NoError(t, err)

	// Test data
	type TestEvent struct {
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		Timestamp time.Time `json:"timestamp"`
	}

	testEvent := TestEvent{
		ID:        uuid.New(),
		Name:      "test-user",
		Timestamp: time.Now(),
	}

	// Setup subscriber
	var receivedMsg *Message
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(msg *Message) error {
		receivedMsg = msg
		wg.Done()
		return nil
	}

	subject := SubjectUserCreated(testEvent.ID)
	err = client.Subscribe(ctx, subject, handler)
	require.NoError(t, err)

	// Give subscriber time to initialize
	time.Sleep(100 * time.Millisecond)

	// Publish message
	err = client.Publish(ctx, subject, testEvent)
	require.NoError(t, err)

	// Wait for message to be received
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	// Verify received message
	require.NotNil(t, receivedMsg)
	assert.Equal(t, subject, receivedMsg.Subject)
	assert.NotEmpty(t, receivedMsg.MessageID)

	// Unmarshal and verify data
	var received TestEvent
	err = json.Unmarshal(receivedMsg.Data, &received)
	require.NoError(t, err)
	assert.Equal(t, testEvent.ID, received.ID)
	assert.Equal(t, testEvent.Name, received.Name)

	// Unsubscribe
	err = client.Unsubscribe(subject)
	require.NoError(t, err)
}

// TestNATSClient_Idempotency tests message idempotency
func TestNATSClient_Idempotency(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := NewNATSClient(natsURL, logger)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	require.NoError(t, err)

	// Test publishing multiple times
	userID := uuid.New()
	subject := SubjectUserCreated(userID)

	data := map[string]string{"name": "test-user"}

	// Publish 3 times
	for i := 0; i < 3; i++ {
		err = client.Publish(ctx, subject, data)
		require.NoError(t, err)
	}

	// Each publish should have unique message ID (no deduplication at publish time)
	// Deduplication happens at consumer level via Ack tracking
}

// TestNATSClient_WorkspaceEvents tests workspace event subjects
func TestNATSClient_WorkspaceEvents(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := NewNATSClient(natsURL, logger)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	require.NoError(t, err)

	workspaceID := uuid.New()

	// Subscribe to all workspace events
	var receivedCount int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2) // Expect 2 messages

	handler := func(msg *Message) error {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		wg.Done()
		return nil
	}

	err = client.Subscribe(ctx, SubjectAllWorkspaces, handler)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Publish created event
	err = client.Publish(ctx, SubjectWorkspaceCreated(workspaceID), map[string]string{
		"event": "created",
	})
	require.NoError(t, err)

	// Publish updated event
	err = client.Publish(ctx, SubjectWorkspaceUpdated(workspaceID), map[string]string{
		"event": "updated",
	})
	require.NoError(t, err)

	// Wait for messages
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for messages")
	}

	assert.Equal(t, 2, receivedCount)
}

// TestNATSClient_BuildEvents tests OSA build event subjects
func TestNATSClient_BuildEvents(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := NewNATSClient(natsURL, logger)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	require.NoError(t, err)

	buildID := "build-" + uuid.New().String()

	var receivedMessages []*Message
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3) // Expect 3 messages

	handler := func(msg *Message) error {
		mu.Lock()
		receivedMessages = append(receivedMessages, msg)
		mu.Unlock()
		wg.Done()
		return nil
	}

	// Subscribe to all build events
	err = client.Subscribe(ctx, SubjectAllBuilds, handler)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Publish build lifecycle events
	err = client.Publish(ctx, SubjectBuildStarted(buildID), map[string]interface{}{
		"build_id": buildID,
		"status":   "started",
	})
	require.NoError(t, err)

	err = client.Publish(ctx, SubjectBuildProgress(buildID), map[string]interface{}{
		"build_id": buildID,
		"progress": 50,
	})
	require.NoError(t, err)

	err = client.Publish(ctx, SubjectBuildCompleted(buildID), map[string]interface{}{
		"build_id": buildID,
		"status":   "completed",
	})
	require.NoError(t, err)

	// Wait for messages
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for messages")
	}

	assert.Equal(t, 3, len(receivedMessages))
}

// TestNATSClient_ErrorHandling tests error scenarios
func TestNATSClient_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Test invalid URL
	_, err := NewNATSClient("", logger)
	assert.Error(t, err)

	// Test connection to invalid server
	_, err = NewNATSClient("nats://invalid:4222", logger)
	assert.Error(t, err)
}

// TestNATSClient_Close tests graceful shutdown
func TestNATSClient_Close(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := NewNATSClient(natsURL, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	require.NoError(t, err)

	// Subscribe to a subject
	handler := func(msg *Message) error {
		return nil
	}
	err = client.Subscribe(ctx, SubjectAllUsers, handler)
	require.NoError(t, err)

	assert.True(t, client.IsConnected())

	// Close client
	err = client.Close()
	require.NoError(t, err)

	assert.False(t, client.IsConnected())

	// Closing again should not error
	err = client.Close()
	require.NoError(t, err)

	// Operations after close should error
	err = client.Publish(ctx, SubjectAllUsers, map[string]string{})
	assert.Error(t, err)
}

// TestNATSClient_Stats tests statistics retrieval
func TestNATSClient_Stats(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		t.Skip("NATS_URL not set, skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := NewNATSClient(natsURL, logger)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	require.NoError(t, err)

	// Initial stats
	stats := client.Stats()
	assert.True(t, stats.Connected)
	assert.Equal(t, 0, stats.ActiveSubscriptions)

	// Subscribe
	handler := func(msg *Message) error { return nil }
	err = client.Subscribe(ctx, SubjectAllUsers, handler)
	require.NoError(t, err)

	// Stats after subscription
	stats = client.Stats()
	assert.Equal(t, 1, stats.ActiveSubscriptions)

	// Publish some messages
	for i := 0; i < 5; i++ {
		err = client.Publish(ctx, SubjectUserCreated(uuid.New()), map[string]int{
			"count": i,
		})
		require.NoError(t, err)
	}

	// Stats after publishing
	stats = client.Stats()
	assert.Greater(t, stats.OutMsgs, uint64(0))
	assert.Greater(t, stats.OutBytes, uint64(0))
}

// Benchmark tests
func BenchmarkNATSClient_Publish(b *testing.B) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		b.Skip("NATS_URL not set, skipping benchmark")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Quiet for benchmarks
	}))

	client, err := NewNATSClient(natsURL, logger)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	if err != nil {
		b.Fatal(err)
	}

	data := map[string]string{
		"name":  "benchmark-user",
		"email": "bench@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := client.Publish(ctx, SubjectUserCreated(uuid.New()), data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNATSClient_PublishSubscribe(b *testing.B) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		b.Skip("NATS_URL not set, skipping benchmark")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	client, err := NewNATSClient(natsURL, logger)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	err = client.CreateStreams(ctx)
	if err != nil {
		b.Fatal(err)
	}

	var wg sync.WaitGroup
	handler := func(msg *Message) error {
		wg.Done()
		return nil
	}

	err = client.Subscribe(ctx, SubjectAllUsers, handler)
	if err != nil {
		b.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond) // Let subscriber initialize

	data := map[string]string{"name": "bench"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		err := client.Publish(ctx, SubjectUserCreated(uuid.New()), data)
		if err != nil {
			b.Fatal(err)
		}
	}
	wg.Wait()
}
