package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
)

// OutboxProcessor processes events from the sync_outbox table using the transactional outbox pattern
type OutboxProcessor struct {
	pool      *pgxpool.Pool
	queries   *sqlc.Queries
	osaClient *osa.Client
	logger    *slog.Logger

	// Configuration
	workers  int           // Number of concurrent workers
	interval time.Duration // Polling interval
	batchSize int          // Number of events to fetch per poll

	// Retry configuration (from Q7 of SYNC_SPECIFICATION_ANSWERS.md)
	retrySchedule []time.Duration // [0s, 1s, 2s, 4s, 8s]
	maxRetries    int             // 5 retries before moving to DLQ

	// Worker pool management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	stopCh chan struct{}
	mu     sync.Mutex
	running bool
}

// NewOutboxProcessor creates a new OutboxProcessor instance
func NewOutboxProcessor(
	pool *pgxpool.Pool,
	osaClient *osa.Client,
	workers int,
	interval time.Duration,
) *OutboxProcessor {
	if workers <= 0 {
		workers = 4 // Default to 4 workers
	}
	if interval <= 0 {
		interval = 5 * time.Second // Default to 5 seconds
	}

	// Retry schedule as per Q7: immediate, 1s, 2s, 4s, 8s (exponential backoff)
	retrySchedule := []time.Duration{
		0 * time.Second,  // Retry 0: Immediate
		1 * time.Second,  // Retry 1: 1 second
		2 * time.Second,  // Retry 2: 2 seconds
		4 * time.Second,  // Retry 3: 4 seconds
		8 * time.Second,  // Retry 4: 8 seconds
	}

	return &OutboxProcessor{
		pool:          pool,
		queries:       sqlc.New(pool),
		osaClient:     osaClient,
		logger:        slog.Default().With("component", "outbox_processor"),
		workers:       workers,
		interval:      interval,
		batchSize:     100,
		retrySchedule: retrySchedule,
		maxRetries:    5,
		stopCh:        make(chan struct{}),
	}
}

// Start begins processing outbox events
func (p *OutboxProcessor) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("outbox processor already running")
	}
	p.running = true
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.mu.Unlock()

	p.logger.Info("starting outbox processor",
		"workers", p.workers,
		"interval", p.interval,
		"batch_size", p.batchSize)

	// Start worker pool
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Start main polling loop
	p.wg.Add(1)
	go p.pollLoop()

	// Start cleanup routine
	p.wg.Add(1)
	go p.cleanupLoop()

	return nil
}

// Stop gracefully stops the outbox processor
func (p *OutboxProcessor) Stop() error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return fmt.Errorf("outbox processor not running")
	}
	p.mu.Unlock()

	p.logger.Info("stopping outbox processor")

	// Signal all goroutines to stop
	close(p.stopCh)
	p.cancel()

	// Wait for all workers to finish
	p.wg.Wait()

	p.mu.Lock()
	p.running = false
	p.mu.Unlock()

	p.logger.Info("outbox processor stopped")
	return nil
}

// pollLoop is the main polling loop that fetches pending events
func (p *OutboxProcessor) pollLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			if err := p.pollAndProcess(); err != nil {
				p.logger.Error("poll cycle failed", "error", err)
			}
		}
	}
}

// pollAndProcess fetches and processes a batch of pending events
func (p *OutboxProcessor) pollAndProcess() error {
	// Fetch pending events (uses FOR UPDATE SKIP LOCKED to prevent races)
	events, err := p.queries.GetPendingOutboxEvents(p.ctx, int32(p.batchSize))
	if err != nil {
		return fmt.Errorf("failed to fetch pending events: %w", err)
	}

	if len(events) == 0 {
		return nil // No events to process
	}

	p.logger.Debug("fetched pending events", "count", len(events))

	// Process events concurrently using worker pool
	eventChan := make(chan sqlc.SyncOutbox, len(events))
	for _, event := range events {
		eventChan <- event
	}
	close(eventChan)

	return nil
}

// worker processes events from the event channel
func (p *OutboxProcessor) worker(id int) {
	defer p.wg.Done()

	logger := p.logger.With("worker_id", id)
	logger.Debug("worker started")

	for {
		select {
		case <-p.ctx.Done():
			logger.Debug("worker stopped (context done)")
			return
		case <-p.stopCh:
			logger.Debug("worker stopped (stop signal)")
			return
		default:
			// Fetch a single event to process (FOR UPDATE SKIP LOCKED ensures no contention)
			events, err := p.queries.GetPendingOutboxEvents(p.ctx, 1)
			if err != nil {
				logger.Error("failed to fetch event", "error", err)
				time.Sleep(1 * time.Second) // Back off on error
				continue
			}

			if len(events) == 0 {
				// No events available, sleep briefly
				time.Sleep(100 * time.Millisecond)
				continue
			}

			event := events[0]
			if err := p.processEvent(p.ctx, event); err != nil {
				logger.Error("failed to process event",
					"event_id", event.ID,
					"aggregate_type", event.AggregateType,
					"error", err)
			}
		}
	}
}

// processEvent processes a single outbox event
func (p *OutboxProcessor) processEvent(ctx context.Context, event sqlc.SyncOutbox) error {
	logger := p.logger.With(
		"event_id", event.ID,
		"aggregate_type", event.AggregateType,
		"aggregate_id", event.AggregateID,
		"event_type", event.EventType,
		"attempt", event.Attempts,
	)

	logger.Info("processing event")

	// Mark as processing
	if err := p.queries.MarkOutboxEventProcessing(ctx, event.ID); err != nil {
		return fmt.Errorf("failed to mark event as processing: %w", err)
	}

	// Process based on aggregate type
	var err error
	switch event.AggregateType {
	case "user":
		err = p.processUserEvent(ctx, event)
	case "workspace":
		err = p.processWorkspaceEvent(ctx, event)
	case "app":
		err = p.processAppEvent(ctx, event)
	case "project":
		err = p.processProjectEvent(ctx, event)
	case "task":
		err = p.processTaskEvent(ctx, event)
	default:
		err = fmt.Errorf("unknown aggregate type: %s", event.AggregateType)
	}

	if err != nil {
		return p.handleProcessingError(ctx, event, err)
	}

	// Mark as completed
	if err := p.queries.MarkOutboxEventCompleted(ctx, event.ID); err != nil {
		return fmt.Errorf("failed to mark event as completed: %w", err)
	}

	logger.Info("event processed successfully")
	return nil
}

// handleProcessingError handles errors during event processing with retry logic
func (p *OutboxProcessor) handleProcessingError(ctx context.Context, event sqlc.SyncOutbox, processingErr error) error {
	logger := p.logger.With("event_id", event.ID, "attempts", event.Attempts)

	// Check if we've exceeded max retries
	if event.Attempts >= int32(p.maxRetries) {
		logger.Error("event exceeded max retries, moving to DLQ",
			"error", processingErr.Error())

		// Move to Dead Letter Queue
		if _, err := p.queries.MoveEventToDLQ(ctx, sqlc.MoveEventToDLQParams{
			ID:            event.ID,
			FailureReason: func() *string { s := processingErr.Error(); return &s }(),
		}); err != nil {
			return fmt.Errorf("failed to move event to DLQ: %w", err)
		}

		// Delete from outbox
		if err := p.queries.DeleteOutboxEvent(ctx, event.ID); err != nil {
			return fmt.Errorf("failed to delete event from outbox: %w", err)
		}

		return nil
	}

	// Calculate next retry time using exponential backoff schedule
	var nextRetryDelay time.Duration
	if int(event.Attempts) < len(p.retrySchedule) {
		nextRetryDelay = p.retrySchedule[event.Attempts]
	} else {
		// If we've exhausted the schedule, use the last delay
		nextRetryDelay = p.retrySchedule[len(p.retrySchedule)-1]
	}

	scheduledFor := time.Now().Add(nextRetryDelay)

	logger.Info("scheduling event retry",
		"next_attempt", event.Attempts+1,
		"retry_delay", nextRetryDelay,
		"scheduled_for", scheduledFor,
		"error", processingErr.Error())

	// Mark as failed and schedule retry
	if err := p.queries.MarkOutboxEventFailed(ctx, sqlc.MarkOutboxEventFailedParams{
		ID:           event.ID,
		LastError:    func() *string { s := processingErr.Error(); return &s }(),
		ScheduledFor: pgtype.Timestamptz{Time: scheduledFor, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	return nil
}

// processUserEvent processes a user sync event
func (p *OutboxProcessor) processUserEvent(ctx context.Context, event sqlc.SyncOutbox) error {
	// Parse payload
	var payload UserSyncPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal user payload: %w", err)
	}

	// TODO: Call OSA API to sync user
	// This would integrate with the OSA client to push user data
	// For now, we'll simulate the sync

	p.logger.Info("syncing user to OSA",
		"user_id", payload.UserID,
		"email", payload.Email)

	// Simulated OSA sync call (replace with actual implementation)
	// resp, err := p.osaClient.SyncUser(ctx, &osa.UserSyncRequest{
	// 	UserID:   payload.UserID,
	// 	Email:    payload.Email,
	// 	FullName: payload.FullName,
	// })
	// if err != nil {
	// 	return fmt.Errorf("OSA sync failed: %w", err)
	// }

	return nil
}

// processWorkspaceEvent processes a workspace sync event
func (p *OutboxProcessor) processWorkspaceEvent(ctx context.Context, event sqlc.SyncOutbox) error {
	// Parse payload
	var payload WorkspaceSyncPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal workspace payload: %w", err)
	}

	p.logger.Info("syncing workspace to OSA",
		"workspace_id", payload.WorkspaceID,
		"name", payload.Name)

	// TODO: Implement workspace sync with OSA
	// resp, err := p.osaClient.SyncWorkspace(ctx, &osa.WorkspaceSyncRequest{...})

	return nil
}

// processAppEvent processes an app sync event
func (p *OutboxProcessor) processAppEvent(ctx context.Context, event sqlc.SyncOutbox) error {
	// Parse payload
	var payload AppSyncPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal app payload: %w", err)
	}

	p.logger.Info("syncing app to OSA",
		"app_id", payload.AppID,
		"name", payload.Name)

	// TODO: Implement app sync with OSA

	return nil
}

// processProjectEvent processes a project sync event
func (p *OutboxProcessor) processProjectEvent(ctx context.Context, event sqlc.SyncOutbox) error {
	// Parse payload
	var payload ProjectSyncPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project payload: %w", err)
	}

	p.logger.Info("syncing project to OSA",
		"project_id", payload.ProjectID,
		"name", payload.Name)

	// TODO: Implement project sync with OSA

	return nil
}

// processTaskEvent processes a task sync event
func (p *OutboxProcessor) processTaskEvent(ctx context.Context, event sqlc.SyncOutbox) error {
	// Parse payload
	var payload TaskSyncPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	p.logger.Info("syncing task to OSA",
		"task_id", payload.TaskID,
		"title", payload.Title)

	// TODO: Implement task sync with OSA

	return nil
}

// cleanupLoop periodically cleans up old completed events and resets stuck processing events
func (p *OutboxProcessor) cleanupLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Hour) // Run cleanup every hour
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.runCleanup()
		}
	}
}

// runCleanup performs cleanup operations
func (p *OutboxProcessor) runCleanup() {
	ctx := context.Background()

	// Reset stuck processing events (older than 5 minutes)
	if err := p.queries.ResetStuckProcessingEvents(ctx); err != nil {
		p.logger.Error("failed to reset stuck processing events", "error", err)
	} else {
		p.logger.Debug("reset stuck processing events")
	}

	// Clean up old completed events (older than 7 days)
	if err := p.queries.CleanupOldCompletedEvents(ctx); err != nil {
		p.logger.Error("failed to cleanup old completed events", "error", err)
	} else {
		p.logger.Debug("cleaned up old completed events")
	}
}

// GetStats returns statistics about the outbox processor
func (p *OutboxProcessor) GetStats(ctx context.Context) (*OutboxStats, error) {
	stats, err := p.queries.GetOutboxEventStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &OutboxStats{
		PendingCount:    int(stats.PendingCount),
		ProcessingCount: int(stats.ProcessingCount),
		CompletedCount:  int(stats.CompletedCount),
		FailedCount:     int(stats.FailedCount),
		DLQReadyCount:   int(stats.DlqReadyCount),
	}, nil
}

// OutboxStats represents statistics about the outbox
type OutboxStats struct {
	PendingCount    int `json:"pending_count"`
	ProcessingCount int `json:"processing_count"`
	CompletedCount  int `json:"completed_count"`
	FailedCount     int `json:"failed_count"`
	DLQReadyCount   int `json:"dlq_ready_count"`
}

// Payload types for different aggregate types

type UserSyncPayload struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
}

type WorkspaceSyncPayload struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Mode        string    `json:"mode"`
}

type AppSyncPayload struct {
	AppID       uuid.UUID `json:"app_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
}

type ProjectSyncPayload struct {
	ProjectID uuid.UUID `json:"project_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
}

type TaskSyncPayload struct {
	TaskID    uuid.UUID `json:"task_id"`
	ProjectID uuid.UUID `json:"project_id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
}
