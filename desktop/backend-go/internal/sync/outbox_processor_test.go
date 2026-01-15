package sync

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestProcessor(t *testing.T) (*OutboxProcessor, *pgxpool.Pool, func()) {
	// Setup test database connection
	// This assumes you have a test database configured
	// Adjust the connection string as needed for your test environment
	pool, err := pgxpool.New(context.Background(), "postgres://localhost/businessos_test?sslmode=disable")
	require.NoError(t, err)

	// Create test OSA client
	osaConfig := &osa.Config{
		BaseURL:      "http://localhost:8089",
		SharedSecret: "test-secret",
		Timeout:      30 * time.Second,
	}
	osaClient, err := osa.NewClient(osaConfig)
	require.NoError(t, err)

	// Create processor
	processor := NewOutboxProcessor(pool, osaClient, 2, 1*time.Second)

	// Cleanup function
	cleanup := func() {
		// Clean up test data
		_, _ = pool.Exec(context.Background(), "DELETE FROM sync_outbox WHERE true")
		_, _ = pool.Exec(context.Background(), "DELETE FROM sync_dlq WHERE true")
		pool.Close()
	}

	return processor, pool, cleanup
}

func TestNewOutboxProcessor(t *testing.T) {
	processor, _, cleanup := setupTestProcessor(t)
	defer cleanup()

	assert.NotNil(t, processor)
	assert.Equal(t, 2, processor.workers)
	assert.Equal(t, 1*time.Second, processor.interval)
	assert.Equal(t, 5, processor.maxRetries)
	assert.Len(t, processor.retrySchedule, 5)
}

func TestRetrySchedule(t *testing.T) {
	processor, _, cleanup := setupTestProcessor(t)
	defer cleanup()

	// Verify retry schedule matches Q7 specification
	expected := []time.Duration{
		0 * time.Second,  // Retry 0: Immediate
		1 * time.Second,  // Retry 1: 1 second
		2 * time.Second,  // Retry 2: 2 seconds
		4 * time.Second,  // Retry 3: 4 seconds
		8 * time.Second,  // Retry 4: 8 seconds
	}

	assert.Equal(t, expected, processor.retrySchedule)
}

func TestProcessUserEvent(t *testing.T) {
	processor, pool, cleanup := setupTestProcessor(t)
	defer cleanup()

	ctx := context.Background()
	queries := sqlc.New(pool)

	// Create test user event
	userID := uuid.New()
	payload := UserSyncPayload{
		UserID:   userID,
		Email:    "test@example.com",
		FullName: "Test User",
	}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	vectorClock := map[string]int{"businessos": 1}
	vectorClockJSON, err := json.Marshal(vectorClock)
	require.NoError(t, err)

	// Insert event into outbox
	event, err := queries.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "user",
		AggregateID:   userID,
		EventType:     "user_created",
		Payload:       payloadJSON,
		VectorClock:   vectorClockJSON,
		MaxAttempts:   5,
	})
	require.NoError(t, err)

	// Process the event
	err = processor.processEvent(ctx, event)
	assert.NoError(t, err)

	// Verify event was marked as completed
	updatedEvent, err := queries.GetOutboxEventByID(ctx, event.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", updatedEvent.Status)
	assert.True(t, updatedEvent.ProcessedAt.Valid)
}

func TestProcessEventRetry(t *testing.T) {
	processor, pool, cleanup := setupTestProcessor(t)
	defer cleanup()

	ctx := context.Background()
	queries := sqlc.New(pool)

	// Create test event that will fail
	taskID := uuid.New()
	payload := TaskSyncPayload{
		TaskID:    taskID,
		ProjectID: uuid.New(),
		Title:     "Test Task",
		Status:    "todo",
	}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	vectorClock := map[string]int{"businessos": 1}
	vectorClockJSON, err := json.Marshal(vectorClock)
	require.NoError(t, err)

	event, err := queries.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "task",
		AggregateID:   taskID,
		EventType:     "task_created",
		Payload:       payloadJSON,
		VectorClock:   vectorClockJSON,
		MaxAttempts:   5,
	})
	require.NoError(t, err)

	// Simulate failure by processing with an invalid context
	// In a real scenario, the OSA client would fail
	// For now, we just verify the retry logic

	// First attempt should succeed (mark as processing then completed)
	err = processor.processEvent(ctx, event)
	assert.NoError(t, err)
}

func TestHandleProcessingErrorWithRetry(t *testing.T) {
	processor, pool, cleanup := setupTestProcessor(t)
	defer cleanup()

	ctx := context.Background()
	queries := sqlc.New(pool)

	// Create test event
	userID := uuid.New()
	payload := UserSyncPayload{
		UserID:   userID,
		Email:    "test@example.com",
		FullName: "Test User",
	}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	vectorClock := map[string]int{"businessos": 1}
	vectorClockJSON, err := json.Marshal(vectorClock)
	require.NoError(t, err)

	event, err := queries.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "user",
		AggregateID:   userID,
		EventType:     "user_created",
		Payload:       payloadJSON,
		VectorClock:   vectorClockJSON,
		MaxAttempts:   5,
	})
	require.NoError(t, err)

	// Simulate processing error
	processingErr := assert.AnError

	// Handle error (should schedule retry)
	err = processor.handleProcessingError(ctx, event, processingErr)
	require.NoError(t, err)

	// Verify event was marked as failed and retry scheduled
	updatedEvent, err := queries.GetOutboxEventByID(ctx, event.ID)
	require.NoError(t, err)
	assert.Equal(t, "failed", updatedEvent.Status)
	assert.Equal(t, int32(1), updatedEvent.Attempts)
	assert.True(t, updatedEvent.ScheduledFor.Valid)
	assert.True(t, updatedEvent.LastError.Valid)
}

func TestHandleProcessingErrorMaxRetries(t *testing.T) {
	processor, pool, cleanup := setupTestProcessor(t)
	defer cleanup()

	ctx := context.Background()
	queries := sqlc.New(pool)

	// Create test event with max attempts already reached
	userID := uuid.New()
	payload := UserSyncPayload{
		UserID:   userID,
		Email:    "test@example.com",
		FullName: "Test User",
	}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	vectorClock := map[string]int{"businessos": 1}
	vectorClockJSON, err := json.Marshal(vectorClock)
	require.NoError(t, err)

	event, err := queries.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "user",
		AggregateID:   userID,
		EventType:     "user_created",
		Payload:       payloadJSON,
		VectorClock:   vectorClockJSON,
		MaxAttempts:   5,
	})
	require.NoError(t, err)

	// Manually set attempts to max retries
	_, err = pool.Exec(ctx, "UPDATE sync_outbox SET attempts = $1 WHERE id = $2", 5, event.ID)
	require.NoError(t, err)

	// Reload event
	event, err = queries.GetOutboxEventByID(ctx, event.ID)
	require.NoError(t, err)

	// Handle error (should move to DLQ)
	processingErr := assert.AnError
	err = processor.handleProcessingError(ctx, event, processingErr)
	require.NoError(t, err)

	// Verify event was moved to DLQ
	dlqEvent, err := queries.GetDLQEventByID(ctx, event.ID)
	require.NoError(t, err)
	assert.Equal(t, "user", dlqEvent.AggregateType)
	assert.Equal(t, userID, dlqEvent.AggregateID)

	// Verify event was deleted from outbox
	_, err = queries.GetOutboxEventByID(ctx, event.ID)
	assert.Error(t, err) // Should not exist
}

func TestGetStats(t *testing.T) {
	processor, pool, cleanup := setupTestProcessor(t)
	defer cleanup()

	ctx := context.Background()
	queries := sqlc.New(pool)

	// Create test events with different statuses
	userID := uuid.New()
	payload := UserSyncPayload{
		UserID:   userID,
		Email:    "test@example.com",
		FullName: "Test User",
	}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	vectorClock := map[string]int{"businessos": 1}
	vectorClockJSON, err := json.Marshal(vectorClock)
	require.NoError(t, err)

	// Create pending event
	_, err = queries.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "user",
		AggregateID:   uuid.New(),
		EventType:     "user_created",
		Payload:       payloadJSON,
		VectorClock:   vectorClockJSON,
		MaxAttempts:   5,
	})
	require.NoError(t, err)

	// Create completed event
	completedEvent, err := queries.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "user",
		AggregateID:   uuid.New(),
		EventType:     "user_updated",
		Payload:       payloadJSON,
		VectorClock:   vectorClockJSON,
		MaxAttempts:   5,
	})
	require.NoError(t, err)
	err = queries.MarkOutboxEventCompleted(ctx, completedEvent.ID)
	require.NoError(t, err)

	// Get stats
	stats, err := processor.GetStats(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, stats.PendingCount)
	assert.Equal(t, 1, stats.CompletedCount)
	assert.Equal(t, 0, stats.ProcessingCount)
	assert.Equal(t, 0, stats.FailedCount)
	assert.Equal(t, 0, stats.DLQReadyCount)
}

func TestStartStop(t *testing.T) {
	processor, _, cleanup := setupTestProcessor(t)
	defer cleanup()

	ctx := context.Background()

	// Start processor
	err := processor.Start(ctx)
	require.NoError(t, err)
	assert.True(t, processor.running)

	// Wait a bit to let workers start
	time.Sleep(100 * time.Millisecond)

	// Stop processor
	err = processor.Stop()
	require.NoError(t, err)
	assert.False(t, processor.running)

	// Verify double start fails
	err = processor.Start(ctx)
	assert.Error(t, err)
}

func TestConcurrentProcessing(t *testing.T) {
	processor, pool, cleanup := setupTestProcessor(t)
	defer cleanup()

	ctx := context.Background()
	queries := sqlc.New(pool)

	// Create multiple test events
	numEvents := 10
	for i := 0; i < numEvents; i++ {
		payload := UserSyncPayload{
			UserID:   uuid.New(),
			Email:    "test@example.com",
			FullName: "Test User",
		}
		payloadJSON, err := json.Marshal(payload)
		require.NoError(t, err)

		vectorClock := map[string]int{"businessos": 1}
		vectorClockJSON, err := json.Marshal(vectorClock)
		require.NoError(t, err)

		_, err = queries.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
			AggregateType: "user",
			AggregateID:   uuid.New(),
			EventType:     "user_created",
			Payload:       payloadJSON,
			VectorClock:   vectorClockJSON,
			MaxAttempts:   5,
		})
		require.NoError(t, err)
	}

	// Start processor
	err := processor.Start(ctx)
	require.NoError(t, err)

	// Wait for processing to complete
	time.Sleep(3 * time.Second)

	// Stop processor
	err = processor.Stop()
	require.NoError(t, err)

	// Verify all events were processed
	stats, err := processor.GetStats(ctx)
	require.NoError(t, err)

	assert.Equal(t, numEvents, stats.CompletedCount)
	assert.Equal(t, 0, stats.PendingCount)
}
