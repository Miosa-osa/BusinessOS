package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

// ConflictDetector handles conflict detection for bidirectional sync operations
type ConflictDetector struct {
	pool   *pgxpool.Pool
	db     *sqlc.Queries
	logger Logger
}

// Logger interface for structured logging
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Conflict represents a detected synchronization conflict
type Conflict struct {
	EntityType      string          `json:"entity_type"`
	EntityID        uuid.UUID       `json:"entity_id"`
	LocalData       json.RawMessage `json:"local_data"`
	RemoteData      json.RawMessage `json:"remote_data"`
	LocalUpdatedAt  time.Time       `json:"local_updated_at"`
	RemoteUpdatedAt time.Time       `json:"remote_updated_at"`
	ConflictFields  []string        `json:"conflict_fields"`
	DetectedAt      time.Time       `json:"detected_at"`
}

// Resolution represents a resolved conflict
type Resolution struct {
	Strategy     string          `json:"strategy"`
	ResolvedData json.RawMessage `json:"resolved_data"`
	ResolvedBy   *uuid.UUID      `json:"resolved_by"` // NULL = automatic
	ResolvedAt   time.Time       `json:"resolved_at"`
	Reasoning    string          `json:"reasoning"`
}

// Workspace represents an OSA workspace for conflict detection
type Workspace struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	Name          string          `json:"name"`
	Mode          string          `json:"mode"`
	Layout        json.RawMessage `json:"layout"`
	ActiveModules []uuid.UUID     `json:"active_modules"`
	TemplateType  string          `json:"template_type"`
	Settings      json.RawMessage `json:"settings"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

const (
	// ResolutionTimestampBased uses timestamps to determine winner (>5 second gap)
	ResolutionTimestampBased = "timestamp_based"

	// ResolutionFieldLevelMerge merges non-conflicting fields automatically
	ResolutionFieldLevelMerge = "field_level_merge"

	// ResolutionManualReview requires user intervention
	ResolutionManualReview = "manual_review"

	// TimestampThreshold is the minimum time difference for automatic resolution
	TimestampThreshold = 5 * time.Second
)

// NewConflictDetector creates a new conflict detector instance
func NewConflictDetector(pool *pgxpool.Pool, logger Logger) *ConflictDetector {
	return &ConflictDetector{
		pool:   pool,
		db:     sqlc.New(pool),
		logger: logger,
	}
}

// DetectWorkspaceConflict detects conflicts between local and remote workspace data
// Returns nil if no conflict exists
func (cd *ConflictDetector) DetectWorkspaceConflict(
	ctx context.Context,
	local, remote *Workspace,
) (*Conflict, error) {
	// No conflict if IDs don't match
	if local.ID != remote.ID {
		return nil, fmt.Errorf("workspace IDs don't match: local=%s, remote=%s", local.ID, remote.ID)
	}

	// Check if there's a clear timestamp-based winner (>5 seconds difference)
	timeDiff := remote.UpdatedAt.Sub(local.UpdatedAt)
	if timeDiff > TimestampThreshold {
		cd.logger.Info("no conflict: remote clearly newer",
			"workspace_id", local.ID,
			"time_diff", timeDiff)
		return nil, nil // Remote wins
	}
	if timeDiff < -TimestampThreshold {
		cd.logger.Info("no conflict: local clearly newer",
			"workspace_id", local.ID,
			"time_diff", -timeDiff)
		return nil, nil // Local wins
	}

	// Concurrent update detected - check field-level conflicts
	conflictFields := cd.findConflictingFields(local, remote)

	if len(conflictFields) == 0 {
		cd.logger.Info("no conflict: data is identical",
			"workspace_id", local.ID)
		return nil, nil
	}

	// Marshal data for storage
	localJSON, err := json.Marshal(local)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal local data: %w", err)
	}

	remoteJSON, err := json.Marshal(remote)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal remote data: %w", err)
	}

	conflict := &Conflict{
		EntityType:      "workspace",
		EntityID:        local.ID,
		LocalData:       localJSON,
		RemoteData:      remoteJSON,
		LocalUpdatedAt:  local.UpdatedAt,
		RemoteUpdatedAt: remote.UpdatedAt,
		ConflictFields:  conflictFields,
		DetectedAt:      time.Now(),
	}

	cd.logger.Warn("conflict detected",
		"workspace_id", local.ID,
		"fields", conflictFields,
		"time_diff", timeDiff)

	return conflict, nil
}

// findConflictingFields identifies which fields differ between local and remote
func (cd *ConflictDetector) findConflictingFields(local, remote *Workspace) []string {
	conflicts := []string{}

	if local.Name != remote.Name {
		conflicts = append(conflicts, "name")
	}

	if local.Mode != remote.Mode {
		conflicts = append(conflicts, "mode")
	}

	if !jsonEqual(local.Layout, remote.Layout) {
		conflicts = append(conflicts, "layout")
	}

	if !uuidSliceEqual(local.ActiveModules, remote.ActiveModules) {
		conflicts = append(conflicts, "active_modules")
	}

	if !jsonEqual(local.Settings, remote.Settings) {
		conflicts = append(conflicts, "settings")
	}

	return conflicts
}

// ResolveConflict applies the 3-tier resolution strategy
// Tier 1: Timestamp-based (>5 second gap) → Auto-resolve
// Tier 2: Field-level merge (non-critical fields) → Auto-merge
// Tier 3: Manual review (critical fields conflict) → Queue for user
func (cd *ConflictDetector) ResolveConflict(
	ctx context.Context,
	conflict *Conflict,
) (*Resolution, error) {
	if conflict.EntityType != "workspace" {
		return nil, fmt.Errorf("unsupported entity type: %s", conflict.EntityType)
	}

	var local, remote Workspace
	if err := json.Unmarshal(conflict.LocalData, &local); err != nil {
		return nil, fmt.Errorf("failed to unmarshal local data: %w", err)
	}
	if err := json.Unmarshal(conflict.RemoteData, &remote); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote data: %w", err)
	}

	// Tier 1: Timestamp-based resolution
	// (Should rarely reach here since DetectWorkspaceConflict already checks this)
	timeDiff := conflict.RemoteUpdatedAt.Sub(conflict.LocalUpdatedAt)
	if timeDiff > TimestampThreshold || timeDiff < -TimestampThreshold {
		return cd.resolveByTimestamp(&local, &remote, conflict)
	}

	// Tier 2: Field-level merge for non-critical fields
	if cd.canAutoMerge(conflict.ConflictFields) {
		return cd.resolveByFieldMerge(&local, &remote, conflict)
	}

	// Tier 3: Manual review required for critical field conflicts
	return cd.queueForManualReview(conflict)
}

// resolveByTimestamp resolves conflict based on timestamp (Tier 1)
func (cd *ConflictDetector) resolveByTimestamp(
	local, remote *Workspace,
	conflict *Conflict,
) (*Resolution, error) {
	var winner *Workspace
	var reasoning string

	if conflict.RemoteUpdatedAt.After(conflict.LocalUpdatedAt) {
		winner = remote
		reasoning = fmt.Sprintf("Remote updated more recently (%s vs %s, diff: %s)",
			conflict.RemoteUpdatedAt.Format(time.RFC3339),
			conflict.LocalUpdatedAt.Format(time.RFC3339),
			conflict.RemoteUpdatedAt.Sub(conflict.LocalUpdatedAt))
	} else {
		winner = local
		reasoning = fmt.Sprintf("Local updated more recently (%s vs %s, diff: %s)",
			conflict.LocalUpdatedAt.Format(time.RFC3339),
			conflict.RemoteUpdatedAt.Format(time.RFC3339),
			conflict.LocalUpdatedAt.Sub(conflict.RemoteUpdatedAt))
	}

	resolvedJSON, err := json.Marshal(winner)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resolved data: %w", err)
	}

	cd.logger.Info("conflict resolved by timestamp",
		"workspace_id", conflict.EntityID,
		"strategy", ResolutionTimestampBased,
		"winner", func() string {
			if winner == remote {
				return "remote"
			}
			return "local"
		}())

	return &Resolution{
		Strategy:     ResolutionTimestampBased,
		ResolvedData: resolvedJSON,
		ResolvedBy:   nil, // Automatic
		ResolvedAt:   time.Now(),
		Reasoning:    reasoning,
	}, nil
}

// resolveByFieldMerge merges non-conflicting fields automatically (Tier 2)
func (cd *ConflictDetector) resolveByFieldMerge(
	local, remote *Workspace,
	conflict *Conflict,
) (*Resolution, error) {
	merged := &Workspace{
		ID:           local.ID,
		UserID:       local.UserID,
		TemplateType: local.TemplateType,
		CreatedAt:    local.CreatedAt,
		UpdatedAt:    time.Now(),
	}

	// For each field, take the version from whoever updated more recently
	useLocal := conflict.LocalUpdatedAt.After(conflict.RemoteUpdatedAt)

	// Non-critical fields - can be merged independently
	if contains(conflict.ConflictFields, "layout") {
		merged.Layout = cd.mergeJSON(local.Layout, remote.Layout, useLocal)
	} else {
		merged.Layout = local.Layout // No conflict, use local
	}

	if contains(conflict.ConflictFields, "active_modules") {
		// For active_modules, merge both lists (union)
		merged.ActiveModules = mergeUUIDSlices(local.ActiveModules, remote.ActiveModules)
	} else {
		merged.ActiveModules = local.ActiveModules
	}

	if contains(conflict.ConflictFields, "settings") {
		merged.Settings = cd.mergeJSON(local.Settings, remote.Settings, useLocal)
	} else {
		merged.Settings = local.Settings
	}

	// For simple string fields, use the more recent one
	if useLocal {
		merged.Name = local.Name
		merged.Mode = local.Mode
	} else {
		merged.Name = remote.Name
		merged.Mode = remote.Mode
	}

	resolvedJSON, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged data: %w", err)
	}

	reasoning := fmt.Sprintf("Field-level merge: Layout and settings merged, simple fields taken from %s (more recent)",
		func() string {
			if useLocal {
				return "local"
			}
			return "remote"
		}())

	cd.logger.Info("conflict resolved by field merge",
		"workspace_id", conflict.EntityID,
		"strategy", ResolutionFieldLevelMerge)

	return &Resolution{
		Strategy:     ResolutionFieldLevelMerge,
		ResolvedData: resolvedJSON,
		ResolvedBy:   nil, // Automatic
		ResolvedAt:   time.Now(),
		Reasoning:    reasoning,
	}, nil
}

// queueForManualReview queues conflict for user review (Tier 3)
func (cd *ConflictDetector) queueForManualReview(conflict *Conflict) (*Resolution, error) {
	reasoning := fmt.Sprintf("Critical fields in conflict: %v. Manual review required.", conflict.ConflictFields)

	cd.logger.Warn("conflict requires manual review",
		"workspace_id", conflict.EntityID,
		"fields", conflict.ConflictFields)

	return &Resolution{
		Strategy:     ResolutionManualReview,
		ResolvedData: nil, // No automatic resolution
		ResolvedBy:   nil,
		ResolvedAt:   time.Now(),
		Reasoning:    reasoning,
	}, nil
}

// canAutoMerge determines if conflicts can be auto-merged
// Returns false if critical fields (name, mode) are in conflict
func (cd *ConflictDetector) canAutoMerge(conflictFields []string) bool {
	criticalFields := []string{"name", "mode"}

	for _, field := range conflictFields {
		if contains(criticalFields, field) {
			return false // Critical field conflict requires manual review
		}
	}

	return true // Only non-critical fields conflicted
}

// mergeJSON merges two JSON objects, taking newer values for shared keys
func (cd *ConflictDetector) mergeJSON(local, remote json.RawMessage, useLocal bool) json.RawMessage {
	var localMap, remoteMap map[string]interface{}

	if err := json.Unmarshal(local, &localMap); err != nil {
		return local
	}
	if err := json.Unmarshal(remote, &remoteMap); err != nil {
		return remote
	}

	merged := make(map[string]interface{})

	// Add all keys from both maps
	for k, v := range localMap {
		merged[k] = v
	}
	for k, v := range remoteMap {
		if _, exists := localMap[k]; exists {
			// Key exists in both - use preference
			if !useLocal {
				merged[k] = v
			}
		} else {
			// Key only in remote
			merged[k] = v
		}
	}

	result, _ := json.Marshal(merged)
	return result
}

// StoreConflict saves a conflict to the database
func (cd *ConflictDetector) StoreConflict(ctx context.Context, conflict *Conflict) error {
	_, err := cd.db.CreateSyncConflict(ctx, sqlc.CreateSyncConflictParams{
		EntityType:      conflict.EntityType,
		EntityID:        uuidToPgtype(conflict.EntityID),
		LocalData:       conflict.LocalData,
		RemoteData:      conflict.RemoteData,
		LocalUpdatedAt:  timeToPgtype(conflict.LocalUpdatedAt),
		RemoteUpdatedAt: timeToPgtype(conflict.RemoteUpdatedAt),
		ConflictFields:  conflict.ConflictFields,
	})

	if err != nil {
		return fmt.Errorf("failed to store conflict: %w", err)
	}

	cd.logger.Info("conflict stored",
		"entity_type", conflict.EntityType,
		"entity_id", conflict.EntityID)

	return nil
}

// GetUnresolvedConflicts retrieves all unresolved conflicts with pagination
func (cd *ConflictDetector) GetUnresolvedConflicts(
	ctx context.Context,
	limit, offset int32,
) ([]sqlc.SyncConflict, error) {
	return cd.db.GetUnresolvedConflicts(ctx, sqlc.GetUnresolvedConflictsParams{
		Limit:  limit,
		Offset: offset,
	})
}

// GetUnresolvedConflictsByEntity retrieves unresolved conflicts for a specific entity
func (cd *ConflictDetector) GetUnresolvedConflictsByEntity(
	ctx context.Context,
	entityType string,
	entityID uuid.UUID,
) ([]sqlc.SyncConflict, error) {
	return cd.db.GetUnresolvedConflictsByEntity(ctx, sqlc.GetUnresolvedConflictsByEntityParams{
		EntityType: entityType,
		EntityID:   uuidToPgtype(entityID),
	})
}

// Helper functions

func jsonEqual(a, b json.RawMessage) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	var aMap, bMap map[string]interface{}
	if err := json.Unmarshal(a, &aMap); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bMap); err != nil {
		return false
	}

	return reflect.DeepEqual(aMap, bMap)
}

func uuidSliceEqual(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[uuid.UUID]bool)
	for _, id := range a {
		aMap[id] = true
	}

	for _, id := range b {
		if !aMap[id] {
			return false
		}
	}

	return true
}

func mergeUUIDSlices(a, b []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool)
	result := []uuid.UUID{}

	for _, id := range a {
		if !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}

	for _, id := range b {
		if !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}

	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// uuidToPgtype converts uuid.UUID to pgtype.UUID
func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

// timeToPgtype converts time.Time to pgtype.Timestamptz
func timeToPgtype(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
