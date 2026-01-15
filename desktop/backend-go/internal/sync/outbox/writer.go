package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/sync/vectorclock"
)

// Writer handles writing synchronization events to the transactional outbox.
// Events are written within the same database transaction as the business logic,
// ensuring exactly-once semantics and data consistency.
type Writer struct {
	pool *pgxpool.Pool
}

// NewWriter creates a new outbox writer.
func NewWriter(pool *pgxpool.Pool) *Writer {
	return &Writer{pool: pool}
}

// WriteRequest contains the parameters for writing an outbox event.
type WriteRequest struct {
	// AggregateType identifies the entity type (user, workspace, app, etc.).
	AggregateType AggregateType

	// AggregateID is the unique identifier of the entity being synchronized.
	AggregateID uuid.UUID

	// EventType describes the type of change (created, updated, deleted, restored).
	EventType EventType

	// Payload contains the entity data as JSON. For updates, this should be the
	// complete current state, not a delta.
	Payload map[string]interface{}

	// VectorClock is the logical timestamp for this event. If nil, a new clock
	// will be initialized.
	VectorClock *vectorclock.VectorClock

	// ScheduledFor allows delayed processing (optional, for backoff retry).
	ScheduledFor *time.Time
}

// Write writes a synchronization event to the outbox within the provided transaction.
// This method MUST be called within an active database transaction to ensure
// atomicity with the business logic.
//
// Example usage:
//
//	tx, err := pool.Begin(ctx)
//	if err != nil {
//	    return err
//	}
//	defer tx.Rollback(ctx)
//
//	// Update workspace in database
//	_, err = tx.Exec(ctx, "UPDATE workspaces SET name = $1 WHERE id = $2", newName, workspaceID)
//	if err != nil {
//	    return err
//	}
//
//	// Write sync event in same transaction
//	err = writer.Write(ctx, tx, WriteRequest{
//	    AggregateType: outbox.AggregateTypeWorkspace,
//	    AggregateID:   workspaceID,
//	    EventType:     outbox.EventTypeUpdated,
//	    Payload:       workspaceData,
//	})
//	if err != nil {
//	    return err
//	}
//
//	return tx.Commit(ctx)
func (w *Writer) Write(ctx context.Context, tx pgx.Tx, req WriteRequest) (*Event, error) {
	// Validate request
	if req.AggregateType == "" {
		return nil, fmt.Errorf("aggregate_type is required")
	}
	if req.AggregateID == uuid.Nil {
		return nil, fmt.Errorf("aggregate_id is required")
	}
	if req.EventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}
	if req.Payload == nil {
		return nil, fmt.Errorf("payload is required")
	}

	// Initialize vector clock if not provided
	vc := req.VectorClock
	if vc == nil {
		vc = vectorclock.New()
	}

	// Increment vector clock for this event
	nodeID := "businessos" // TODO: Make this configurable for multi-instance deployments
	vc.Increment(nodeID)

	// Serialize vector clock
	vcJSON, err := json.Marshal(vc.ToMap())
	if err != nil {
		return nil, fmt.Errorf("marshal vector clock: %w", err)
	}

	// Serialize payload
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// Generate event ID
	eventID := uuid.New()

	// Write event to outbox
	now := time.Now()
	query := `
		INSERT INTO sync_outbox (
			id, aggregate_type, aggregate_id, event_type,
			payload, vector_clock, status, attempts,
			created_at, scheduled_for
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	var createdAt time.Time
	err = tx.QueryRow(ctx, query,
		eventID,
		req.AggregateType,
		req.AggregateID,
		req.EventType,
		payloadJSON,
		vcJSON,
		SyncStatusPending,
		0, // initial attempts
		now,
		req.ScheduledFor,
	).Scan(&eventID, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("insert outbox event: %w", err)
	}

	// Construct and return the event
	event := &Event{
		ID:            eventID,
		AggregateType: req.AggregateType,
		AggregateID:   req.AggregateID,
		EventType:     req.EventType,
		Payload:       req.Payload,
		VectorClock:   vc.ToMap(),
		Status:        SyncStatusPending,
		Attempts:      0,
		CreatedAt:     createdAt,
		ScheduledFor:  req.ScheduledFor,
	}

	return event, nil
}

// WriteWithPool writes a synchronization event to the outbox, managing its own transaction.
// This is a convenience method when you don't have an existing transaction.
// For operations that need atomicity with business logic, use Write() instead.
func (w *Writer) WriteWithPool(ctx context.Context, req WriteRequest) (*Event, error) {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	event, err := w.Write(ctx, tx, req)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return event, nil
}

// GetPendingEvents retrieves pending events from the outbox for processing.
// Events are returned in creation order, limited by the specified batch size.
func (w *Writer) GetPendingEvents(ctx context.Context, limit int) ([]*Event, error) {
	query := `
		SELECT
			id, aggregate_type, aggregate_id, event_type,
			payload, vector_clock, status, attempts,
			last_error, created_at, processed_at, scheduled_for
		FROM sync_outbox
		WHERE status = $1
			AND (scheduled_for IS NULL OR scheduled_for <= NOW())
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := w.pool.Query(ctx, query, SyncStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("query pending events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		var payloadJSON, vcJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.AggregateType,
			&event.AggregateID,
			&event.EventType,
			&payloadJSON,
			&vcJSON,
			&event.Status,
			&event.Attempts,
			&event.LastError,
			&event.CreatedAt,
			&event.ProcessedAt,
			&event.ScheduledFor,
		)
		if err != nil {
			return nil, fmt.Errorf("scan event row: %w", err)
		}

		// Deserialize payload
		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("unmarshal payload for event %s: %w", event.ID, err)
		}

		// Deserialize vector clock
		if err := json.Unmarshal(vcJSON, &event.VectorClock); err != nil {
			return nil, fmt.Errorf("unmarshal vector clock for event %s: %w", event.ID, err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate event rows: %w", err)
	}

	return events, nil
}

// MarkProcessing marks an event as currently being processed.
// This prevents duplicate processing by other workers.
func (w *Writer) MarkProcessing(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE sync_outbox
		SET status = $1, attempts = attempts + 1, updated_at = NOW()
		WHERE id = $2 AND status = $3
	`

	result, err := w.pool.Exec(ctx, query, SyncStatusProcessing, eventID, SyncStatusPending)
	if err != nil {
		return fmt.Errorf("mark event processing: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event %s not found or already processed", eventID)
	}

	return nil
}

// MarkCompleted marks an event as successfully processed.
func (w *Writer) MarkCompleted(ctx context.Context, eventID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE sync_outbox
		SET status = $1, processed_at = $2, updated_at = NOW()
		WHERE id = $3
	`

	result, err := w.pool.Exec(ctx, query, SyncStatusCompleted, now, eventID)
	if err != nil {
		return fmt.Errorf("mark event completed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event %s not found", eventID)
	}

	return nil
}

// MarkFailed marks an event as failed with an error message.
// The event can be retried later based on retry policy.
func (w *Writer) MarkFailed(ctx context.Context, eventID uuid.UUID, errMsg string, scheduleRetryAt *time.Time) error {
	query := `
		UPDATE sync_outbox
		SET status = $1, last_error = $2, scheduled_for = $3, updated_at = NOW()
		WHERE id = $4
	`

	result, err := w.pool.Exec(ctx, query, SyncStatusFailed, errMsg, scheduleRetryAt, eventID)
	if err != nil {
		return fmt.Errorf("mark event failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event %s not found", eventID)
	}

	return nil
}
