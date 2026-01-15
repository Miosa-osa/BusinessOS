package outbox

import (
	"time"

	"github.com/google/uuid"
)

// SyncStatus represents the processing status of an outbox event.
type SyncStatus string

const (
	// SyncStatusPending indicates the event has been written but not yet processed.
	SyncStatusPending SyncStatus = "pending"

	// SyncStatusProcessing indicates the event is currently being processed.
	SyncStatusProcessing SyncStatus = "processing"

	// SyncStatusCompleted indicates the event has been successfully processed and published.
	SyncStatusCompleted SyncStatus = "completed"

	// SyncStatusFailed indicates the event processing failed (will retry based on policy).
	SyncStatusFailed SyncStatus = "failed"
)

// AggregateType represents the type of entity being synchronized.
type AggregateType string

const (
	// AggregateTypeUser represents user entity synchronization events.
	AggregateTypeUser AggregateType = "user"

	// AggregateTypeWorkspace represents workspace entity synchronization events.
	AggregateTypeWorkspace AggregateType = "workspace"

	// AggregateTypeApp represents application entity synchronization events.
	AggregateTypeApp AggregateType = "app"

	// AggregateTypeProject represents project entity synchronization events.
	AggregateTypeProject AggregateType = "project"

	// AggregateTypeTask represents task entity synchronization events.
	AggregateTypeTask AggregateType = "task"
)

// EventType represents the type of change event.
type EventType string

const (
	// EventTypeCreated indicates an entity was created.
	EventTypeCreated EventType = "created"

	// EventTypeUpdated indicates an entity was updated.
	EventTypeUpdated EventType = "updated"

	// EventTypeDeleted indicates an entity was deleted (soft or hard).
	EventTypeDeleted EventType = "deleted"

	// EventTypeRestored indicates a previously deleted entity was restored.
	EventTypeRestored EventType = "restored"
)

// Event represents a synchronization event in the transactional outbox.
// Events are written within the same transaction as the business logic,
// ensuring exactly-once semantics and consistency.
type Event struct {
	// ID is the unique identifier for this event.
	ID uuid.UUID `json:"id"`

	// AggregateType identifies the entity type (user, workspace, app, etc.).
	AggregateType AggregateType `json:"aggregate_type"`

	// AggregateID is the unique identifier of the entity being synchronized.
	AggregateID uuid.UUID `json:"aggregate_id"`

	// EventType describes the type of change (created, updated, deleted, restored).
	EventType EventType `json:"event_type"`

	// Payload contains the entity data as JSON. For updates, this should be the
	// complete current state, not a delta.
	Payload map[string]interface{} `json:"payload"`

	// VectorClock is the logical timestamp for this event, used for conflict detection
	// and ordering in distributed systems.
	VectorClock map[string]int `json:"vector_clock"`

	// Status tracks the processing state of this event.
	Status SyncStatus `json:"status"`

	// Attempts tracks how many times processing has been attempted (for retry logic).
	Attempts int `json:"attempts"`

	// MaxAttempts is the maximum number of retry attempts before moving to DLQ.
	MaxAttempts int `json:"max_attempts"`

	// LastError stores the error message from the most recent failed processing attempt.
	LastError *string `json:"last_error,omitempty"`

	// CreatedAt is when the event was written to the outbox.
	CreatedAt time.Time `json:"created_at"`

	// ProcessedAt is when the event was successfully processed (null until completed).
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// ScheduledFor allows delayed processing (e.g., for backoff retry).
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
}

// ConflictType represents the type of synchronization conflict detected.
type ConflictType string

const (
	// ConflictTypeConcurrent indicates two updates happened concurrently
	// (vector clocks are incomparable).
	ConflictTypeConcurrent ConflictType = "concurrent"

	// ConflictTypeStale indicates an update is being applied to an older version
	// (incoming vector clock is before current).
	ConflictTypeStale ConflictType = "stale"

	// ConflictTypeDeleted indicates an update is being applied to a deleted entity.
	ConflictTypeDeleted ConflictType = "deleted"

	// ConflictTypeDuplicate indicates the exact same event was received multiple times.
	ConflictTypeDuplicate ConflictType = "duplicate"
)

// Conflict represents a detected synchronization conflict.
type Conflict struct {
	// ID is the unique identifier for this conflict.
	ID uuid.UUID `json:"id"`

	// AggregateType identifies the entity type.
	AggregateType AggregateType `json:"aggregate_type"`

	// AggregateID is the unique identifier of the conflicting entity.
	AggregateID uuid.UUID `json:"aggregate_id"`

	// ConflictType describes the type of conflict.
	ConflictType ConflictType `json:"conflict_type"`

	// LocalVersion is the current state in BusinessOS.
	LocalVersion map[string]interface{} `json:"local_version"`

	// RemoteVersion is the incoming state from OSA.
	RemoteVersion map[string]interface{} `json:"remote_version"`

	// LocalVectorClock is the vector clock of the local version.
	LocalVectorClock map[string]int `json:"local_vector_clock"`

	// RemoteVectorClock is the vector clock of the remote version.
	RemoteVectorClock map[string]int `json:"remote_vector_clock"`

	// Resolved indicates whether this conflict has been resolved.
	Resolved bool `json:"resolved"`

	// Resolution describes how the conflict was resolved (if resolved).
	Resolution *string `json:"resolution,omitempty"`

	// CreatedAt is when the conflict was detected.
	CreatedAt time.Time `json:"created_at"`

	// ResolvedAt is when the conflict was resolved (null if still pending).
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}
