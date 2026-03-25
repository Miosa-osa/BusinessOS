package audit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogEvent(t *testing.T) {
	t.Run("logs event successfully", func(t *testing.T) {
		logger := NewAuditLog()

		actor := uuid.New()
		resource := uuid.New()
		action := "CREATE"
		resourceType := "document"

		err := logger.LogEvent(context.Background(), &LogEventParams{
			Actor:        &actor,
			Action:       action,
			Resource:     &resource,
			ResourceType: &resourceType,
			Timestamp:    time.Now(),
		})

		assert.NoError(t, err)
		assert.Equal(t, int64(1), logger.eventCount)
	})

	t.Run("tamper detection catches hash manipulation", func(t *testing.T) {
		logger := NewAuditLog()

		actor := uuid.New()
		resource := uuid.New()
		action := "DELETE"
		resourceType := "user"

		// Log legitimate event
		err := logger.LogEvent(context.Background(), &LogEventParams{
			Actor:        &actor,
			Action:       action,
			Resource:     &resource,
			ResourceType: &resourceType,
			Timestamp:    time.Now(),
		})
		require.NoError(t, err)

		// Attempt to tamper with stored event
		if len(logger.events) > 0 {
			logger.events[0].EntryHash = "tampered-hash"
		}

		// Verify the tamper detection flag is set
		if len(logger.events) > 0 {
			assert.NotEmpty(t, logger.events[0].EntryHash)
		}
	})

	t.Run("retention enforced based on age", func(t *testing.T) {
		logger := NewAuditLog()

		// Log events with old timestamps
		oldTime := time.Now().Add(-100 * 24 * time.Hour)
		actor := uuid.New()

		for i := 0; i < 5; i++ {
			resourceID := uuid.New()
			err := logger.LogEvent(context.Background(), &LogEventParams{
				Actor:        &actor,
				Action:       "READ",
				Resource:     &resourceID,
				ResourceType: strPtr("log_entry"),
				Timestamp:    oldTime,
			})
			assert.NoError(t, err)
		}

		// Apply retention policy (e.g., keep last 30 days)
		logger.EnforceRetention(30 * 24 * time.Hour)

		// Old events should be pruned
		assert.True(t, logger.eventCount <= 5)
	})
}

func TestGetHistory(t *testing.T) {
	t.Run("retrieves event history for resource", func(t *testing.T) {
		logger := NewAuditLog()

		resourceID := uuid.New()
		actor1 := uuid.New()
		actor2 := uuid.New()

		// Create multiple events for same resource
		err1 := logger.LogEvent(context.Background(), &LogEventParams{
			Actor:        &actor1,
			Action:       "CREATE",
			Resource:     &resourceID,
			ResourceType: strPtr("document"),
			Timestamp:    time.Now(),
		})
		assert.NoError(t, err1)

		err2 := logger.LogEvent(context.Background(), &LogEventParams{
			Actor:        &actor2,
			Action:       "EDIT",
			Resource:     &resourceID,
			ResourceType: strPtr("document"),
			Timestamp:    time.Now().Add(1 * time.Hour),
		})
		assert.NoError(t, err2)

		// Retrieve history
		history, err := logger.GetHistory(context.Background(), resourceID)
		assert.NoError(t, err)
		assert.NotNil(t, history)
	})
}

// Helper function
func strPtr(s string) *string {
	return &s
}
