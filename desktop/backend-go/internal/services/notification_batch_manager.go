package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

type BatchManager struct {
	pool *pgxpool.Pool
}

func NewBatchManager(pool *pgxpool.Pool) *BatchManager {
	return &BatchManager{pool: pool}
}

// Queue adds a notification to a batch or creates a new batch
// Returns the batch ID if batched, or empty UUID if immediate dispatch is needed
func (m *BatchManager) Queue(ctx context.Context, input CreateInput) (*uuid.UUID, error) {
	queries := sqlc.New(m.pool)

	// Get batch config for this notification type
	cfg, ok := GetTypeConfig(input.Type)
	if !ok || cfg.Batch == nil {
		return nil, nil
	}

	batchConfig := cfg.Batch

	// Generate batch key based on grouping strategy
	batchKey := m.generateBatchKey(input, batchConfig.GroupBy)

	// Look for existing pending batch
	existingBatch, err := queries.GetPendingBatch(ctx, sqlc.GetPendingBatchParams{
		UserID:   input.UserID,
		BatchKey: batchKey,
	})

	if err == nil {
		pendingCount := int32(0)
		if existingBatch.PendingCount != nil {
			pendingCount = *existingBatch.PendingCount
		}

		if pendingCount < int32(batchConfig.Max) {
			// Create a placeholder UUID for the pending notification
			pendingID := uuid.New()
			_, err = queries.AddToBatch(ctx, sqlc.AddToBatchParams{
				ID:          pgtype.UUID{Bytes: existingBatch.ID.Bytes, Valid: true},
				ArrayAppend: pgtype.UUID{Bytes: pendingID, Valid: true},
			})
			if err != nil {
				log.Printf("[BatchManager] Failed to add to batch: %v", err)
				return nil, nil
			}

			batchID := uuid.UUID(existingBatch.ID.Bytes)
			log.Printf("[BatchManager] Added to existing batch %s (count: %d)", batchID, pendingCount+1)
			return &batchID, nil
		}

		// Batch is full, dispatch it now and create new one
		log.Printf("[BatchManager] Batch %s is full, will be dispatched", existingBatch.ID.Bytes)
	}

	// Create new batch
	dispatchAt := time.Now().Add(batchConfig.Window)

	var entityID pgtype.UUID
	if input.EntityID != nil {
		entityID = pgtype.UUID{Bytes: *input.EntityID, Valid: true}
	}

	pendingID := uuid.New()
	newBatch, err := queries.CreateBatch(ctx, sqlc.CreateBatchParams{
		UserID:     input.UserID,
		BatchKey:   batchKey,
		Type:       input.Type,
		EntityType: strPtr(input.EntityType),
		EntityID:   entityID,
		PendingIds: []pgtype.UUID{{Bytes: pendingID, Valid: true}},
		DispatchAt: pgtype.Timestamptz{Time: dispatchAt, Valid: true},
	})
	if err != nil {
		log.Printf("[BatchManager] Failed to create batch: %v", err)
		return nil, nil
	}

	batchID := uuid.UUID(newBatch.ID.Bytes)
	log.Printf("[BatchManager] Created new batch %s for type %s, dispatching at %s", batchID, input.Type, dispatchAt.Format(time.RFC3339))
	return &batchID, nil
}

func (m *BatchManager) generateBatchKey(input CreateInput, groupBy string) string {
	base := fmt.Sprintf("%s:%s", input.UserID, input.Type)

	switch groupBy {
	case "entity_id":
		if input.EntityID != nil {
			return fmt.Sprintf("%s:%s", base, input.EntityID.String())
		}
		return base
	case "sender_id":
		if input.SenderID != "" {
			return fmt.Sprintf("%s:%s", base, input.SenderID)
		}
		return base
	case "type":
		return base
	default:
		return base
	}
}

// GetPendingBatchCount
func (m *BatchManager) GetPendingBatchCount(ctx context.Context, batchID uuid.UUID) (int, error) {
	queries := sqlc.New(m.pool)

	batch, err := queries.GetPendingBatch(ctx, sqlc.GetPendingBatchParams{
		UserID:   "",
		BatchKey: "",
	})
	if err != nil {
		return 0, err
	}

	if batch.PendingCount == nil {
		return 0, nil
	}
	return int(*batch.PendingCount), nil
}

// Batch title templates
var batchTitleTemplates = map[string]string{
	NotifTaskAssigned:      "You were assigned %d tasks",
	NotifTaskUnassigned:    "You were unassigned from %d tasks",
	NotifTaskCompleted:     "%d tasks were completed",
	NotifTaskComment:       "%d new comments on your tasks",
	NotifTaskStatusChanged: "%d task status updates",
	NotifProjectAdded:      "You were added to %d projects",
	NotifTeamMemberJoined:  "%d new team members joined",
	NotifClientDealUpdate:  "%d deal updates",
	NotifMentionTask:       "You were mentioned in %d tasks",
	NotifMentionProject:    "You were mentioned in %d projects",
	NotifMentionComment:    "You were mentioned in %d comments",
	NotifMentionDailyLog:   "You were mentioned in %d daily logs",
	NotifDailyLogMention:   "You were mentioned in %d daily logs",
}

func GetBatchTitle(notifType string, count int) string {
	if template, ok := batchTitleTemplates[notifType]; ok {
		return fmt.Sprintf(template, count)
	}
	return fmt.Sprintf("%d new notifications", count)
}
