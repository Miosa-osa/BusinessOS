package outbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNATSPublisher is a mock implementation of NATSPublisher for testing.
type MockNATSPublisher struct {
	published []PublishedMessage
	shouldFail bool
	failCount  int
}

type PublishedMessage struct {
	Subject string
	Data    []byte
}

func (m *MockNATSPublisher) Publish(ctx context.Context, subject string, data []byte) error {
	if m.shouldFail {
		m.failCount++
		return assert.AnError
	}
	m.published = append(m.published, PublishedMessage{
		Subject: subject,
		Data:    data,
	})
	return nil
}

func TestDefaultProcessorConfig(t *testing.T) {
	config := DefaultProcessorConfig()

	assert.Equal(t, 100, config.BatchSize)
	assert.Equal(t, 1*time.Second, config.PollInterval)
	assert.Equal(t, 5, config.MaxAttempts)
	assert.Equal(t, 1*time.Second, config.InitialBackoff)
	assert.Equal(t, 5*time.Minute, config.MaxBackoff)
	assert.Equal(t, 2.0, config.BackoffMultiplier)
	assert.Equal(t, 0.25, config.JitterFactor)
}

func TestCalculateBackoff(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	processor := NewProcessor(nil, nil, logger)

	tests := []struct {
		attempt      int
		expectedMin  time.Duration
		expectedMax  time.Duration
		description  string
	}{
		{
			attempt:     0,
			expectedMin: 0,
			expectedMax: 0,
			description: "Attempt 0: Immediate",
		},
		{
			attempt:     1,
			expectedMin: 750 * time.Millisecond,  // 1s - 25% jitter
			expectedMax: 1250 * time.Millisecond, // 1s + 25% jitter
			description: "Attempt 1: 1s ± 25%",
		},
		{
			attempt:     2,
			expectedMin: 1500 * time.Millisecond, // 2s - 25% jitter
			expectedMax: 2500 * time.Millisecond, // 2s + 25% jitter
			description: "Attempt 2: 2s ± 25%",
		},
		{
			attempt:     3,
			expectedMin: 3000 * time.Millisecond, // 4s - 25% jitter
			expectedMax: 5000 * time.Millisecond, // 4s + 25% jitter
			description: "Attempt 3: 4s ± 25%",
		},
		{
			attempt:     4,
			expectedMin: 6000 * time.Millisecond,  // 8s - 25% jitter
			expectedMax: 10000 * time.Millisecond, // 8s + 25% jitter
			description: "Attempt 4: 8s ± 25%",
		},
		{
			attempt:     5,
			expectedMin: 12000 * time.Millisecond, // 16s - 25% jitter
			expectedMax: 20000 * time.Millisecond, // 16s + 25% jitter
			description: "Attempt 5: 16s ± 25%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Test multiple times to account for jitter
			for i := 0; i < 10; i++ {
				backoff := processor.calculateBackoff(tt.attempt)
				assert.GreaterOrEqual(t, backoff, tt.expectedMin, "Backoff should be >= min")
				assert.LessOrEqual(t, backoff, tt.expectedMax, "Backoff should be <= max")
			}
		})
	}
}

func TestCalculateBackoff_MaxCap(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := DefaultProcessorConfig()
	processor := NewProcessorWithConfig(nil, nil, logger, config)

	// Attempt 10 would be 1s * 2^9 = 512s without cap
	// Should be capped at 5 minutes (300s) ± 25%
	backoff := processor.calculateBackoff(10)

	maxExpected := time.Duration(float64(config.MaxBackoff) * 1.25) // 5min + 25% jitter
	assert.LessOrEqual(t, backoff, maxExpected, "Backoff should be capped at max + jitter")
	assert.GreaterOrEqual(t, backoff, 100*time.Millisecond, "Backoff should be >= minimum")
}

func TestPublishToNATS_SubjectFormat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mockNATS := &MockNATSPublisher{}
	processor := NewProcessor(nil, mockNATS, logger)

	tests := []struct {
		aggregateType string
		eventType     string
		expectedSubject string
	}{
		{
			aggregateType:   "user",
			eventType:       "created",
			expectedSubject: "businessos.user.created",
		},
		{
			aggregateType:   "workspace",
			eventType:       "updated",
			expectedSubject: "businessos.workspace.updated",
		},
		{
			aggregateType:   "app",
			eventType:       "deleted",
			expectedSubject: "businessos.app.deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expectedSubject, func(t *testing.T) {
			mockNATS.published = nil // Reset

			msg := &Event{
				ID:            uuid.New(),
				AggregateType: AggregateType(tt.aggregateType),
				AggregateID:   uuid.New(),
				EventType:     EventType(tt.eventType),
				Payload:       map[string]interface{}{"test": "data"},
				VectorClock:   map[string]int{"businessos": 1},
			}

			err := processor.publishToNATS(context.Background(), msg)
			require.NoError(t, err)

			require.Len(t, mockNATS.published, 1)
			assert.Equal(t, tt.expectedSubject, mockNATS.published[0].Subject)

			// Verify message can be deserialized
			var deserializedMsg Event
			err = json.Unmarshal(mockNATS.published[0].Data, &deserializedMsg)
			require.NoError(t, err)
			assert.Equal(t, msg.ID, deserializedMsg.ID)
			assert.Equal(t, msg.AggregateType, deserializedMsg.AggregateType)
		})
	}
}

func TestPublishToNATS_Timeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Mock that simulates slow publish
	slowNATS := &SlowMockNATSPublisher{delay: 10 * time.Second}
	processor := NewProcessor(nil, slowNATS, logger)

	msg := &Event{
		ID:            uuid.New(),
		AggregateType: AggregateTypeUser,
		AggregateID:   uuid.New(),
		EventType:     EventTypeCreated,
		Payload:       map[string]interface{}{"test": "data"},
		VectorClock:   map[string]int{"businessos": 1},
	}

	// Should timeout after 5 seconds (hardcoded in publishToNATS)
	start := time.Now()
	err := processor.publishToNATS(context.Background(), msg)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, duration, 7*time.Second, "Should timeout before 7 seconds")
	assert.Greater(t, duration, 4*time.Second, "Should take at least 4 seconds")
}

// SlowMockNATSPublisher simulates a slow NATS server
type SlowMockNATSPublisher struct {
	delay time.Duration
}

func (s *SlowMockNATSPublisher) Publish(ctx context.Context, subject string, data []byte) error {
	select {
	case <-time.After(s.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestProcessorConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config ProcessorConfig
		valid  bool
	}{
		{
			name:   "Valid default config",
			config: DefaultProcessorConfig(),
			valid:  true,
		},
		{
			name: "Custom valid config",
			config: ProcessorConfig{
				BatchSize:         50,
				PollInterval:      2 * time.Second,
				MaxAttempts:       3,
				InitialBackoff:    500 * time.Millisecond,
				MaxBackoff:        1 * time.Minute,
				BackoffMultiplier: 2.0,
				JitterFactor:      0.1,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			mockNATS := &MockNATSPublisher{}

			processor := NewProcessorWithConfig(nil, mockNATS, logger, tt.config)

			assert.NotNil(t, processor)
			assert.Equal(t, tt.config.BatchSize, processor.config.BatchSize)
			assert.Equal(t, tt.config.PollInterval, processor.config.PollInterval)
			assert.Equal(t, tt.config.MaxAttempts, processor.config.MaxAttempts)
		})
	}
}

func TestProcessorRetrySequence(t *testing.T) {
	// This test verifies the retry delay sequence matches the requirements
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	processor := NewProcessor(nil, nil, logger)

	// Expected retry delays (without jitter, center values)
	expectedDelays := []time.Duration{
		0,                 // Attempt 0: Immediate
		1 * time.Second,   // Attempt 1: 1s
		2 * time.Second,   // Attempt 2: 2s
		4 * time.Second,   // Attempt 3: 4s
		8 * time.Second,   // Attempt 4: 8s
		16 * time.Second,  // Attempt 5: 16s
	}

	for attempt, expected := range expectedDelays {
		backoff := processor.calculateBackoff(attempt)

		// Allow for jitter (±25%)
		minExpected := time.Duration(float64(expected) * 0.75)
		maxExpected := time.Duration(float64(expected) * 1.25)

		if expected == 0 {
			assert.Equal(t, expected, backoff, "Attempt %d should be immediate", attempt)
		} else {
			assert.GreaterOrEqual(t, backoff, minExpected, "Attempt %d should be >= %s", attempt, minExpected)
			assert.LessOrEqual(t, backoff, maxExpected, "Attempt %d should be <= %s", attempt, maxExpected)
		}
	}
}
