package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/sync/metrics"
)

// NATSPublisher defines the interface for publishing messages to NATS JetStream.
// This abstraction allows for testing and different NATS client implementations.
type NATSPublisher interface {
	// Publish publishes a message to the specified subject.
	Publish(ctx context.Context, subject string, data []byte) error
}

// ProcessorConfig holds configuration for the outbox processor.
type ProcessorConfig struct {
	// BatchSize is the number of messages to process in each poll (default: 100)
	BatchSize int

	// PollInterval is how often to poll for new messages (default: 1 second)
	PollInterval time.Duration

	// MaxAttempts is the maximum number of retry attempts (default: 5)
	MaxAttempts int

	// InitialBackoff is the initial retry delay (default: 1 second)
	InitialBackoff time.Duration

	// MaxBackoff is the maximum retry delay (default: 5 minutes)
	MaxBackoff time.Duration

	// BackoffMultiplier is the exponential backoff multiplier (default: 2.0)
	BackoffMultiplier float64

	// JitterFactor is the amount of jitter to add to backoff (default: 0.25 = ±25%)
	JitterFactor float64
}

// DefaultProcessorConfig returns the default processor configuration.
func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		BatchSize:         100,
		PollInterval:      1 * time.Second,
		MaxAttempts:       5,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        5 * time.Minute,
		BackoffMultiplier: 2.0,
		JitterFactor:      0.25,
	}
}

// Processor handles background processing of outbox events.
// It polls the database for pending events, publishes them to NATS JetStream,
// and handles retry logic with exponential backoff.
type Processor struct {
	pool      *pgxpool.Pool
	nats      NATSPublisher
	logger    *slog.Logger
	metrics   *metrics.Metrics
	config    ProcessorConfig
	stopCh    chan struct{}
	stoppedCh chan struct{}
}

// NewProcessor creates a new outbox processor with default configuration.
func NewProcessor(pool *pgxpool.Pool, nats NATSPublisher, logger *slog.Logger) *Processor {
	return NewProcessorWithConfig(pool, nats, logger, DefaultProcessorConfig())
}

// NewProcessorWithConfig creates a new outbox processor with custom configuration.
func NewProcessorWithConfig(pool *pgxpool.Pool, nats NATSPublisher, logger *slog.Logger, config ProcessorConfig) *Processor {
	return &Processor{
		pool:      pool,
		nats:      nats,
		logger:    logger,
		metrics:   metrics.GetMetrics(),
		config:    config,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}
}

// Start begins processing outbox messages in a background loop.
// It returns when the context is cancelled or Stop() is called.
func (p *Processor) Start(ctx context.Context) error {
	p.logger.Info("starting outbox processor",
		"batch_size", p.config.BatchSize,
		"poll_interval", p.config.PollInterval,
		"max_attempts", p.config.MaxAttempts)

	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()
	defer close(p.stoppedCh)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("outbox processor stopped (context cancelled)")
			return ctx.Err()

		case <-p.stopCh:
			p.logger.Info("outbox processor stopped (stop signal)")
			return nil

		case <-ticker.C:
			if err := p.processBatch(ctx); err != nil {
				p.logger.Error("batch processing failed", "error", err)
				p.metrics.IncrementError("batch_processing")
			}
		}
	}
}

// Stop gracefully stops the processor.
func (p *Processor) Stop() {
	close(p.stopCh)
	<-p.stoppedCh // Wait for processor to finish
}

// processBatch processes a batch of outbox messages.
// It uses FOR UPDATE SKIP LOCKED for concurrent worker safety.
func (p *Processor) processBatch(ctx context.Context) error {
	startTime := time.Now()

	// Query for pending messages that are ready to be processed
	query := `
		SELECT
			id, aggregate_type, aggregate_id, event_type,
			payload, vector_clock, attempts, max_attempts
		FROM sync_outbox
		WHERE status = 'pending'
			AND attempts < max_attempts
			AND (scheduled_for IS NULL OR scheduled_for <= NOW())
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, query, p.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to query outbox: %w", err)
	}
	defer rows.Close()

	var processed, failed int
	for rows.Next() {
		var msg Event
		var payloadJSON, vcJSON []byte

		if err := rows.Scan(
			&msg.ID,
			&msg.AggregateType,
			&msg.AggregateID,
			&msg.EventType,
			&payloadJSON,
			&vcJSON,
			&msg.Attempts,
			&msg.MaxAttempts,
		); err != nil {
			p.logger.Error("failed to scan row", "error", err)
			p.metrics.IncrementError("row_scan")
			continue
		}

		// Deserialize payload and vector clock
		if err := json.Unmarshal(payloadJSON, &msg.Payload); err != nil {
			p.logger.Error("failed to unmarshal payload",
				"event_id", msg.ID,
				"error", err)
			p.recordFailure(ctx, tx, &msg, fmt.Errorf("invalid payload: %w", err))
			failed++
			continue
		}

		if err := json.Unmarshal(vcJSON, &msg.VectorClock); err != nil {
			p.logger.Error("failed to unmarshal vector clock",
				"event_id", msg.ID,
				"error", err)
			p.recordFailure(ctx, tx, &msg, fmt.Errorf("invalid vector clock: %w", err))
			failed++
			continue
		}

		// Attempt to publish to NATS
		if err := p.publishToNATS(ctx, &msg); err != nil {
			p.logger.Warn("failed to publish to NATS",
				"event_id", msg.ID,
				"aggregate_type", msg.AggregateType,
				"event_type", msg.EventType,
				"attempts", msg.Attempts+1,
				"error", err)

			p.recordFailure(ctx, tx, &msg, err)
			p.metrics.IncrementFailedEvents()
			failed++
			continue
		}

		// Mark as processed on success
		if err := p.markProcessed(ctx, tx, msg.ID); err != nil {
			p.logger.Error("failed to mark as processed",
				"event_id", msg.ID,
				"error", err)
			// Don't increment failed counter - message was published successfully
			// It will be retried and marked as duplicate by the consumer
			continue
		}

		p.metrics.IncrementCompletedEvents()
		p.metrics.DecrementPendingEvents()
		processed++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %w", err)
	}

	// Commit all changes
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Log and record metrics
	duration := time.Since(startTime)
	if processed > 0 || failed > 0 {
		p.logger.Debug("processed outbox batch",
			"processed", processed,
			"failed", failed,
			"duration_ms", duration.Milliseconds())
		p.metrics.RecordProcessingDuration(duration)
	}

	return nil
}

// publishToNATS publishes an event to NATS JetStream.
func (p *Processor) publishToNATS(ctx context.Context, msg *Event) error {
	// Construct NATS subject: businessos.<aggregate_type>.<event_type>
	// Example: businessos.user.created, businessos.workspace.updated
	subject := fmt.Sprintf("businessos.%s.%s", msg.AggregateType, msg.EventType)

	// Serialize full event as JSON for NATS
	eventJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish with timeout
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.nats.Publish(publishCtx, subject, eventJSON); err != nil {
		return fmt.Errorf("NATS publish failed: %w", err)
	}

	p.logger.Debug("published event to NATS",
		"event_id", msg.ID,
		"subject", subject,
		"aggregate_id", msg.AggregateID)

	return nil
}

// markProcessed marks an event as successfully processed.
func (p *Processor) markProcessed(ctx context.Context, tx pgx.Tx, eventID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE sync_outbox
		SET status = 'completed', processed_at = $1
		WHERE id = $2
	`

	result, err := tx.Exec(ctx, query, now, eventID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event %s not found", eventID)
	}

	return nil
}

// recordFailure increments the attempt count and schedules a retry.
// If max attempts is exceeded, the message is moved to the dead letter queue.
func (p *Processor) recordFailure(ctx context.Context, tx pgx.Tx, msg *Event, err error) {
	msg.Attempts++
	errMsg := err.Error()

	// Check if we've exceeded max attempts
	if msg.Attempts >= p.config.MaxAttempts {
		p.logger.Warn("max attempts exceeded, moving to DLQ",
			"event_id", msg.ID,
			"aggregate_type", msg.AggregateType,
			"attempts", msg.Attempts,
			"error", errMsg)

		if dlqErr := p.moveToDLQ(ctx, tx, msg, errMsg); dlqErr != nil {
			p.logger.Error("failed to move to DLQ",
				"event_id", msg.ID,
				"error", dlqErr)
			// Still update the outbox record even if DLQ fails
		}
		return
	}

	// Calculate exponential backoff with jitter
	retryDelay := p.calculateBackoff(msg.Attempts)
	scheduledFor := time.Now().Add(retryDelay)

	p.logger.Debug("scheduling retry",
		"event_id", msg.ID,
		"attempt", msg.Attempts,
		"max_attempts", p.config.MaxAttempts,
		"retry_after", retryDelay,
		"scheduled_for", scheduledFor)

	query := `
		UPDATE sync_outbox
		SET
			status = 'failed',
			attempts = $1,
			last_error = $2,
			scheduled_for = $3
		WHERE id = $4
	`

	if _, err := tx.Exec(ctx, query, msg.Attempts, errMsg, scheduledFor, msg.ID); err != nil {
		p.logger.Error("failed to record failure",
			"event_id", msg.ID,
			"error", err)
	}
}

// calculateBackoff calculates the retry delay with exponential backoff and jitter.
// Formula: delay = min(initialBackoff * multiplier^attempt, maxBackoff) ± jitter
func (p *Processor) calculateBackoff(attempt int) time.Duration {
	if attempt == 0 {
		return 0 // Immediate retry on first failure
	}

	// Calculate exponential backoff: initialBackoff * multiplier^(attempt-1)
	backoff := float64(p.config.InitialBackoff) * math.Pow(p.config.BackoffMultiplier, float64(attempt-1))

	// Cap at max backoff
	if backoff > float64(p.config.MaxBackoff) {
		backoff = float64(p.config.MaxBackoff)
	}

	// Add jitter: backoff ± (backoff * jitterFactor * random[-1, 1])
	jitter := backoff * p.config.JitterFactor * (rand.Float64()*2 - 1)
	finalBackoff := time.Duration(backoff + jitter)

	// Ensure minimum delay of 100ms
	if finalBackoff < 100*time.Millisecond {
		finalBackoff = 100 * time.Millisecond
	}

	return finalBackoff
}

// moveToDLQ moves a failed message to the dead letter queue.
func (p *Processor) moveToDLQ(ctx context.Context, tx pgx.Tx, msg *Event, lastError string) error {
	// Insert into DLQ
	dlqQuery := `
		INSERT INTO sync_dlq (
			id, aggregate_type, aggregate_id, event_type,
			payload, vector_clock, attempts, last_error,
			failure_reason, original_created_at, moved_to_dlq_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`

	payloadJSON, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	vcJSON, err := json.Marshal(msg.VectorClock)
	if err != nil {
		return fmt.Errorf("failed to marshal vector clock: %w", err)
	}

	failureReason := fmt.Sprintf("Max attempts (%d) exceeded", p.config.MaxAttempts)

	if _, err := tx.Exec(ctx, dlqQuery,
		msg.ID,
		msg.AggregateType,
		msg.AggregateID,
		msg.EventType,
		payloadJSON,
		vcJSON,
		msg.Attempts,
		lastError,
		failureReason,
		msg.CreatedAt,
	); err != nil {
		return fmt.Errorf("failed to insert into DLQ: %w", err)
	}

	// Delete from outbox
	deleteQuery := `DELETE FROM sync_outbox WHERE id = $1`
	if _, err := tx.Exec(ctx, deleteQuery, msg.ID); err != nil {
		return fmt.Errorf("failed to delete from outbox: %w", err)
	}

	p.logger.Info("moved event to DLQ",
		"event_id", msg.ID,
		"aggregate_type", msg.AggregateType,
		"event_type", msg.EventType,
		"attempts", msg.Attempts)

	return nil
}
