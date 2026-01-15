package outbox

// This file demonstrates how to use the Outbox pattern in practice.
// DO NOT compile this file - it's for documentation purposes only.

/*

// ===== EXAMPLE 1: Writing to Outbox in Business Logic =====

package services

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rhl/businessos-backend/internal/sync/outbox"
)

type UserService struct {
    pool   *pgxpool.Pool
    writer *outbox.Writer
}

func NewUserService(pool *pgxpool.Pool) *UserService {
    return &UserService{
        pool:   pool,
        writer: outbox.NewWriter(pool),
    }
}

// UpdateUser updates a user and writes a sync event atomically.
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, name string) error {
    // Begin transaction using pgx.BeginFunc for automatic commit/rollback
    return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
        // 1. Update business data
        query := `UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2 RETURNING *`
        var user User
        err := tx.QueryRow(ctx, query, name, userID).Scan(
            &user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
        )
        if err != nil {
            return fmt.Errorf("failed to update user: %w", err)
        }

        // 2. Write sync event in same transaction
        _, err = s.writer.Write(ctx, tx, outbox.WriteRequest{
            AggregateType: outbox.AggregateTypeUser,
            AggregateID:   userID,
            EventType:     outbox.EventTypeUpdated,
            Payload: map[string]interface{}{
                "id":         user.ID.String(),
                "email":      user.Email,
                "name":       user.Name,
                "updated_at": user.UpdatedAt,
            },
        })
        if err != nil {
            return fmt.Errorf("failed to write outbox event: %w", err)
        }

        // Transaction commits automatically on success, rolls back on error
        return nil
    })
}


// ===== EXAMPLE 2: Starting the Outbox Processor =====

package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rhl/businessos-backend/internal/sync/messaging"
    "github.com/rhl/businessos-backend/internal/sync/outbox"
)

func main() {
    // Setup context with cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Initialize logger
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))

    // Connect to database
    pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        logger.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer pool.Close()

    // Connect to NATS
    natsClient, err := messaging.NewNATSClient(os.Getenv("NATS_URL"))
    if err != nil {
        logger.Error("failed to connect to NATS", "error", err)
        os.Exit(1)
    }
    defer natsClient.Close()

    // Create and start outbox processor
    processor := outbox.NewProcessor(pool, natsClient, logger)

    // Start processor in background
    processorDone := make(chan error, 1)
    go func() {
        logger.Info("starting outbox processor")
        processorDone <- processor.Start(ctx)
    }()

    // Wait for shutdown signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    select {
    case sig := <-sigCh:
        logger.Info("received shutdown signal", "signal", sig)
        cancel() // Cancel context
        processor.Stop() // Graceful shutdown
    case err := <-processorDone:
        logger.Error("processor stopped unexpectedly", "error", err)
    }

    logger.Info("shutdown complete")
}


// ===== EXAMPLE 3: Custom Configuration =====

package main

import (
    "time"

    "github.com/rhl/businessos-backend/internal/sync/outbox"
)

func setupProcessor() *outbox.Processor {
    // Custom configuration for high-throughput scenarios
    config := outbox.ProcessorConfig{
        BatchSize:         200,              // Process 200 messages per poll
        PollInterval:      500 * time.Millisecond, // Poll every 500ms
        MaxAttempts:       3,                // Max 3 retries (faster failure)
        InitialBackoff:    500 * time.Millisecond, // Start with 500ms delay
        MaxBackoff:        2 * time.Minute,  // Cap at 2 minutes
        BackoffMultiplier: 2.0,              // Double delay each retry
        JitterFactor:      0.15,             // ±15% jitter
    }

    return outbox.NewProcessorWithConfig(pool, natsClient, logger, config)
}


// ===== EXAMPLE 4: Multiple Processors for Scaling =====

package main

import (
    "context"
    "sync"

    "github.com/rhl/businessos-backend/internal/sync/outbox"
)

func startMultipleProcessors(ctx context.Context, count int) []*outbox.Processor {
    processors := make([]*outbox.Processor, count)
    var wg sync.WaitGroup

    for i := 0; i < count; i++ {
        processor := outbox.NewProcessor(pool, natsClient, logger)
        processors[i] = processor

        wg.Add(1)
        go func(p *outbox.Processor, id int) {
            defer wg.Done()
            logger.Info("starting processor", "id", id)
            if err := p.Start(ctx); err != nil {
                logger.Error("processor failed", "id", id, "error", err)
            }
        }(processor, i)
    }

    // Wait for all processors to start
    return processors
}

func stopMultipleProcessors(processors []*outbox.Processor) {
    var wg sync.WaitGroup
    for i, p := range processors {
        wg.Add(1)
        go func(proc *outbox.Processor, id int) {
            defer wg.Done()
            logger.Info("stopping processor", "id", id)
            proc.Stop()
        }(p, i)
    }
    wg.Wait()
}


// ===== EXAMPLE 5: Monitoring Metrics =====

package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/rhl/businessos-backend/internal/sync/metrics"
)

func RegisterMetricsEndpoint(r *gin.Engine) {
    r.GET("/api/sync/metrics", func(c *gin.Context) {
        m := metrics.GetMetrics()
        snapshot := m.GetSnapshot()

        c.JSON(http.StatusOK, gin.H{
            "outbox": gin.H{
                "pending":    snapshot.OutboxPendingEvents,
                "processing": snapshot.OutboxProcessingEvents,
                "completed":  snapshot.OutboxCompletedEvents,
                "failed":     snapshot.OutboxFailedEvents,
            },
            "performance": gin.H{
                "avg_duration_ms": snapshot.AvgProcessingDuration.Milliseconds(),
                "p95_duration_ms": snapshot.P95ProcessingDuration.Milliseconds(),
                "p99_duration_ms": snapshot.P99ProcessingDuration.Milliseconds(),
            },
            "health": gin.H{
                "last_processed_at": snapshot.LastProcessedAt,
                "last_error_at":     snapshot.LastErrorAt,
            },
            "errors":    snapshot.ErrorsTotal,
            "conflicts": snapshot.ConflictsTotal,
        })
    })
}


// ===== EXAMPLE 6: DLQ Management Endpoints =====

package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

func RegisterDLQEndpoints(r *gin.Engine, pool *pgxpool.Pool) {
    // List DLQ events
    r.GET("/api/sync/dlq", func(c *gin.Context) {
        query := `
            SELECT id, aggregate_type, event_type, attempts, last_error, moved_to_dlq_at
            FROM sync_dlq
            WHERE resolved = FALSE
            ORDER BY moved_to_dlq_at DESC
            LIMIT 100
        `

        rows, err := pool.Query(c.Request.Context(), query)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        var events []map[string]interface{}
        for rows.Next() {
            var event map[string]interface{}
            // Scan into event...
            events = append(events, event)
        }

        c.JSON(http.StatusOK, gin.H{"events": events})
    })

    // Reprocess DLQ event
    r.POST("/api/sync/dlq/:id/reprocess", func(c *gin.Context) {
        eventID, err := uuid.Parse(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
            return
        }

        // Copy event back to outbox with reset attempts
        query := `
            INSERT INTO sync_outbox (
                aggregate_type, aggregate_id, event_type,
                payload, vector_clock, status, attempts, max_attempts
            )
            SELECT
                aggregate_type, aggregate_id, event_type,
                payload, vector_clock, 'pending', 0, 5
            FROM sync_dlq
            WHERE id = $1 AND resolved = FALSE
            RETURNING id
        `

        var newEventID uuid.UUID
        err = pool.QueryRow(c.Request.Context(), query, eventID).Scan(&newEventID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reprocess"})
            return
        }

        // Mark as resolved
        _, _ = pool.Exec(c.Request.Context(),
            "UPDATE sync_dlq SET resolved = TRUE, resolved_at = NOW(), resolution_notes = 'Manually reprocessed' WHERE id = $1",
            eventID)

        c.JSON(http.StatusOK, gin.H{
            "message":      "Event reprocessed",
            "new_event_id": newEventID,
        })
    })
}


// ===== EXAMPLE 7: Graceful Shutdown Pattern =====

package main

import (
    "context"
    "time"
)

func gracefulShutdown() {
    // Create context with timeout for shutdown
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    // Stop accepting new work
    cancel() // Cancel main context

    // Wait for processor to finish current batch
    processor.Stop()

    // Wait for in-flight database operations
    time.Sleep(1 * time.Second)

    // Close database pool
    pool.Close()

    // Close NATS connection
    natsClient.Close()

    logger.Info("shutdown complete")
}

*/
