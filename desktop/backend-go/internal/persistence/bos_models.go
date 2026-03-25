// Package persistence implements the BOS ↔ BusinessOS persistence layer
// It provides schema management, model operations, and synchronization primitives
// for the unified data layer stored in PostgreSQL.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

// ============================================================================
// Error Types
// ============================================================================

type PersistenceError struct {
	Code    string
	Message string
	Err     error
}

func (e *PersistenceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeVersionMismatch = "VERSION_MISMATCH"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeDatabase        = "DATABASE_ERROR"
	ErrCodeSerialization   = "SERIALIZATION_ERROR"
)

// ============================================================================
// Data Models
// ============================================================================

// DiscoveredModel represents a process model discovered by BOS
type DiscoveredModel struct {
	ID                   uuid.UUID            `json:"id"`
	WorkspaceID          uuid.UUID            `json:"workspace_id"`
	Name                 string               `json:"name"`
	Description          *string              `json:"description"`
	ModelType            string               `json:"model_type"` // 'petri_net', 'process_tree', 'dfg'
	Places               *json.RawMessage     `json:"places"`
	Transitions          *json.RawMessage     `json:"transitions"`
	Arcs                 *json.RawMessage     `json:"arcs"`
	TreeJSON             *json.RawMessage     `json:"tree_json"`
	FitnessScore         *float64             `json:"fitness_score"`
	PrecisionScore       *float64             `json:"precision_score"`
	GeneralizationScore  *float64             `json:"generalization_score"`
	Version              int32                `json:"version"`
	SourceSessionID      *uuid.UUID           `json:"source_session_id"`
	SourceLogID          uuid.UUID            `json:"source_log_id"`
	IsArchived           bool                 `json:"is_archived"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
	CreatedBy            *string              `json:"created_by"`
}

// ConformanceResult represents conformance checking results
type ConformanceResult struct {
	ID                 uuid.UUID        `json:"id"`
	WorkspaceID        uuid.UUID        `json:"workspace_id"`
	ModelID            uuid.UUID        `json:"model_id"`
	LogID              uuid.UUID        `json:"log_id"`
	ConformanceType    string           `json:"conformance_type"` // 'token_replay', 'alignment'
	Fitness            float64          `json:"fitness"`
	Precision          *float64         `json:"precision"`
	Generalization     *float64         `json:"generalization"`
	IsFitting          bool             `json:"is_fitting"`
	TraceFitness       *json.RawMessage `json:"trace_fitness"`
	AlignedTraces      *json.RawMessage `json:"aligned_traces"`
	ViolatedConstraints *json.RawMessage `json:"violated_constraints"`
	TotalTraces        int32            `json:"total_traces"`
	FittingTraces      int32            `json:"fitting_traces"`
	NonFittingTraces   int32            `json:"non_fitting_traces"`
	ExecutionTimeMs    *int32           `json:"execution_time_ms"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// ProcessStatistics represents process mining statistics
type ProcessStatistics struct {
	ID              uuid.UUID        `json:"id"`
	WorkspaceID     uuid.UUID        `json:"workspace_id"`
	LogID           uuid.UUID        `json:"log_id"`
	ModelID         *uuid.UUID       `json:"model_id"`
	VariantCount    int32            `json:"variant_count"`
	TopVariants     *json.RawMessage `json:"top_variants"`
	ActivityCount   int32            `json:"activity_count"`
	Activities      *json.RawMessage `json:"activities"`
	ResourceCount   *int32           `json:"resource_count"`
	Resources       *json.RawMessage `json:"resources"`
	ReworkFrequency *float64         `json:"rework_frequency"`
	ReworkDetails   *json.RawMessage `json:"rework_details"`
	CustomMetrics   *json.RawMessage `json:"custom_metrics"`
	AnalysisType    string           `json:"analysis_type"`
	ExecutionTimeMs *int32           `json:"execution_time_ms"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// AuditLogEntry represents persistence audit trail
type AuditLogEntry struct {
	ID           uuid.UUID        `json:"id"`
	WorkspaceID  uuid.UUID        `json:"workspace_id"`
	EntityType   string           `json:"entity_type"`
	EntityID     uuid.UUID        `json:"entity_id"`
	Operation    string           `json:"operation"` // INSERT, UPDATE, DELETE
	OldValues    *json.RawMessage `json:"old_values"`
	NewValues    *json.RawMessage `json:"new_values"`
	SourceSystem string           `json:"source_system"` // 'bos', 'businessos'
	UserID       *string          `json:"user_id"`
	SourceIP     *string          `json:"source_ip"`
	CreatedAt    time.Time        `json:"created_at"`
}

// ============================================================================
// Schema Management
// ============================================================================

// SchemaManager handles database migrations
type SchemaManager struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(pool *pgxpool.Pool, log *slog.Logger) *SchemaManager {
	return &SchemaManager{
		pool: pool,
		log:  log,
	}
}

// InitializeSchema creates all BOS persistence tables if they don't exist
func (sm *SchemaManager) InitializeSchema(ctx context.Context) error {
	tables := []string{
		createDiscoverySessionsTable,
		createDiscoveredModelsTable,
		createConformanceResultsTable,
		createProcessStatisticsTable,
		createAuditLogTable,
		createSyncCheckpointsTable,
	}

	for _, createStmt := range tables {
		if _, err := sm.pool.Exec(ctx, createStmt); err != nil {
			sm.log.Error("failed to create table",
				"error", err,
				"statement", createStmt[:50]) // Log first 50 chars
			return &PersistenceError{
				Code:    ErrCodeDatabase,
				Message: "schema initialization failed",
				Err:     err,
			}
		}
	}

	sm.log.Info("BOS persistence schema initialized successfully")
	return nil
}

// ============================================================================
// Model Repository
// ============================================================================

// ModelRepository handles model CRUD operations
type ModelRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewModelRepository creates a new model repository
func NewModelRepository(pool *pgxpool.Pool, log *slog.Logger) *ModelRepository {
	return &ModelRepository{
		pool: pool,
		log:  log,
	}
}

// GetModel retrieves a model by ID
func (r *ModelRepository) GetModel(ctx context.Context, modelID uuid.UUID) (*DiscoveredModel, error) {
	var model DiscoveredModel

	err := r.pool.QueryRow(ctx, `
		SELECT id, workspace_id, name, description, model_type,
		       places, transitions, arcs, tree_json,
		       fitness_score, precision_score, generalization_score,
		       version, source_session_id, source_log_id, is_archived,
		       created_at, updated_at, created_by
		FROM discovered_models
		WHERE id = $1
	`, modelID).Scan(
		&model.ID, &model.WorkspaceID, &model.Name, &model.Description, &model.ModelType,
		&model.Places, &model.Transitions, &model.Arcs, &model.TreeJSON,
		&model.FitnessScore, &model.PrecisionScore, &model.GeneralizationScore,
		&model.Version, &model.SourceSessionID, &model.SourceLogID, &model.IsArchived,
		&model.CreatedAt, &model.UpdatedAt, &model.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, &PersistenceError{
			Code:    ErrCodeNotFound,
			Message: fmt.Sprintf("model %s not found", modelID),
			Err:     err,
		}
	} else if err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to query model",
			Err:     err,
		}
	}

	return &model, nil
}

// ListModels retrieves models for a workspace with pagination
func (r *ModelRepository) ListModels(ctx context.Context, workspaceID uuid.UUID, limit, offset int32) ([]*DiscoveredModel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workspace_id, name, description, model_type,
		       places, transitions, arcs, tree_json,
		       fitness_score, precision_score, generalization_score,
		       version, source_session_id, source_log_id, is_archived,
		       created_at, updated_at, created_by
		FROM discovered_models
		WHERE workspace_id = $1 AND is_archived = false
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`, workspaceID, limit, offset)

	if err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to list models",
			Err:     err,
		}
	}
	defer rows.Close()

	var models []*DiscoveredModel
	for rows.Next() {
		var model DiscoveredModel
		err := rows.Scan(
			&model.ID, &model.WorkspaceID, &model.Name, &model.Description, &model.ModelType,
			&model.Places, &model.Transitions, &model.Arcs, &model.TreeJSON,
			&model.FitnessScore, &model.PrecisionScore, &model.GeneralizationScore,
			&model.Version, &model.SourceSessionID, &model.SourceLogID, &model.IsArchived,
			&model.CreatedAt, &model.UpdatedAt, &model.CreatedBy,
		)
		if err != nil {
			r.log.Error("failed to scan model row", "error", err)
			continue
		}
		models = append(models, &model)
	}

	if err = rows.Err(); err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "error iterating models",
			Err:     err,
		}
	}

	return models, nil
}

// SaveModel persists a model to the database
func (r *ModelRepository) SaveModel(ctx context.Context, model *DiscoveredModel) (uuid.UUID, error) {
	if model.ID == uuid.Nil {
		model.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO discovered_models
		(id, workspace_id, name, description, model_type,
		 places, transitions, arcs, tree_json,
		 fitness_score, precision_score, generalization_score,
		 source_session_id, source_log_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`,
		model.ID, model.WorkspaceID, model.Name, model.Description, model.ModelType,
		model.Places, model.Transitions, model.Arcs, model.TreeJSON,
		model.FitnessScore, model.PrecisionScore, model.GeneralizationScore,
		model.SourceSessionID, model.SourceLogID, model.CreatedBy,
	).Scan(&model.ID)

	if err != nil {
		return uuid.Nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to save model",
			Err:     err,
		}
	}

	return model.ID, nil
}

// UpdateModelScores updates model fitness/precision scores with optimistic locking
func (r *ModelRepository) UpdateModelScores(ctx context.Context, modelID uuid.UUID, expectedVersion int32, fitness, precision, generalization *float64) error {
	var newVersion int32

	err := r.pool.QueryRow(ctx, `
		UPDATE discovered_models
		SET fitness_score = COALESCE($2, fitness_score),
		    precision_score = COALESCE($3, precision_score),
		    generalization_score = COALESCE($4, generalization_score),
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $1 AND version = $5
		RETURNING version
	`, modelID, fitness, precision, generalization, expectedVersion).Scan(&newVersion)

	if err == sql.ErrNoRows {
		return &PersistenceError{
			Code:    ErrCodeVersionMismatch,
			Message: fmt.Sprintf("version mismatch for model %s", modelID),
		}
	} else if err != nil {
		return &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to update model scores",
			Err:     err,
		}
	}

	return nil
}

// ============================================================================
// Conformance Repository
// ============================================================================

// ConformanceRepository handles conformance result operations
type ConformanceRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewConformanceRepository creates a new conformance repository
func NewConformanceRepository(pool *pgxpool.Pool, log *slog.Logger) *ConformanceRepository {
	return &ConformanceRepository{
		pool: pool,
		log:  log,
	}
}

// SaveResult persists a conformance result
func (r *ConformanceRepository) SaveResult(ctx context.Context, result *ConformanceResult) (uuid.UUID, error) {
	if result.ID == uuid.Nil {
		result.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO conformance_results
		(id, workspace_id, model_id, log_id, conformance_type,
		 fitness, precision, generalization, is_fitting,
		 trace_fitness, aligned_traces, violated_constraints,
		 total_traces, fitting_traces, non_fitting_traces, execution_time_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
	`,
		result.ID, result.WorkspaceID, result.ModelID, result.LogID, result.ConformanceType,
		result.Fitness, result.Precision, result.Generalization, result.IsFitting,
		result.TraceFitness, result.AlignedTraces, result.ViolatedConstraints,
		result.TotalTraces, result.FittingTraces, result.NonFittingTraces, result.ExecutionTimeMs,
	).Scan(&result.ID)

	if err != nil {
		return uuid.Nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to save conformance result",
			Err:     err,
		}
	}

	return result.ID, nil
}

// GetResultsForModel retrieves all conformance results for a model
func (r *ConformanceRepository) GetResultsForModel(ctx context.Context, modelID uuid.UUID) ([]*ConformanceResult, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workspace_id, model_id, log_id, conformance_type,
		       fitness, precision, generalization, is_fitting,
		       trace_fitness, aligned_traces, violated_constraints,
		       total_traces, fitting_traces, non_fitting_traces, execution_time_ms,
		       created_at, updated_at
		FROM conformance_results
		WHERE model_id = $1
		ORDER BY created_at DESC
	`, modelID)

	if err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to query conformance results",
			Err:     err,
		}
	}
	defer rows.Close()

	var results []*ConformanceResult
	for rows.Next() {
		var result ConformanceResult
		err := rows.Scan(
			&result.ID, &result.WorkspaceID, &result.ModelID, &result.LogID, &result.ConformanceType,
			&result.Fitness, &result.Precision, &result.Generalization, &result.IsFitting,
			&result.TraceFitness, &result.AlignedTraces, &result.ViolatedConstraints,
			&result.TotalTraces, &result.FittingTraces, &result.NonFittingTraces, &result.ExecutionTimeMs,
			&result.CreatedAt, &result.UpdatedAt,
		)
		if err != nil {
			r.log.Error("failed to scan conformance result", "error", err)
			continue
		}
		results = append(results, &result)
	}

	if err = rows.Err(); err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "error iterating conformance results",
			Err:     err,
		}
	}

	return results, nil
}

// ============================================================================
// Statistics Repository
// ============================================================================

// StatisticsRepository handles process statistics operations
type StatisticsRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewStatisticsRepository creates a new statistics repository
func NewStatisticsRepository(pool *pgxpool.Pool, log *slog.Logger) *StatisticsRepository {
	return &StatisticsRepository{
		pool: pool,
		log:  log,
	}
}

// SaveStatistics persists process statistics
func (r *StatisticsRepository) SaveStatistics(ctx context.Context, stats *ProcessStatistics) (uuid.UUID, error) {
	if stats.ID == uuid.Nil {
		stats.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO process_statistics
		(id, workspace_id, log_id, model_id,
		 variant_count, top_variants, activity_count, activities,
		 resource_count, resources,
		 rework_frequency, rework_details, custom_metrics,
		 analysis_type, execution_time_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`,
		stats.ID, stats.WorkspaceID, stats.LogID, stats.ModelID,
		stats.VariantCount, stats.TopVariants, stats.ActivityCount, stats.Activities,
		stats.ResourceCount, stats.Resources,
		stats.ReworkFrequency, stats.ReworkDetails, stats.CustomMetrics,
		stats.AnalysisType, stats.ExecutionTimeMs,
	).Scan(&stats.ID)

	if err != nil {
		return uuid.Nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to save statistics",
			Err:     err,
		}
	}

	return stats.ID, nil
}

// GetStatisticsForLog retrieves statistics for a log
func (r *StatisticsRepository) GetStatisticsForLog(ctx context.Context, logID uuid.UUID) ([]*ProcessStatistics, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workspace_id, log_id, model_id,
		       variant_count, top_variants, activity_count, activities,
		       resource_count, resources,
		       rework_frequency, rework_details, custom_metrics,
		       analysis_type, execution_time_ms,
		       created_at, updated_at
		FROM process_statistics
		WHERE log_id = $1
		ORDER BY created_at DESC
	`, logID)

	if err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to query statistics",
			Err:     err,
		}
	}
	defer rows.Close()

	var stats []*ProcessStatistics
	for rows.Next() {
		var s ProcessStatistics
		err := rows.Scan(
			&s.ID, &s.WorkspaceID, &s.LogID, &s.ModelID,
			&s.VariantCount, &s.TopVariants, &s.ActivityCount, &s.Activities,
			&s.ResourceCount, &s.Resources,
			&s.ReworkFrequency, &s.ReworkDetails, &s.CustomMetrics,
			&s.AnalysisType, &s.ExecutionTimeMs,
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			r.log.Error("failed to scan statistics", "error", err)
			continue
		}
		stats = append(stats, &s)
	}

	if err = rows.Err(); err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "error iterating statistics",
			Err:     err,
		}
	}

	return stats, nil
}

// ============================================================================
// Audit Repository
// ============================================================================

// AuditRepository handles audit trail operations
type AuditRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(pool *pgxpool.Pool, log *slog.Logger) *AuditRepository {
	return &AuditRepository{
		pool: pool,
		log:  log,
	}
}

// RecordEntry logs an operation to audit trail
func (r *AuditRepository) RecordEntry(ctx context.Context, entry *AuditLogEntry) (uuid.UUID, error) {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO persistence_audit_log
		(id, workspace_id, entity_type, entity_id, operation,
		 old_values, new_values, source_system, user_id, source_ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`,
		entry.ID, entry.WorkspaceID, entry.EntityType, entry.EntityID, entry.Operation,
		entry.OldValues, entry.NewValues, entry.SourceSystem, entry.UserID, entry.SourceIP,
	).Scan(&entry.ID)

	if err != nil {
		return uuid.Nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to record audit entry",
			Err:     err,
		}
	}

	return entry.ID, nil
}

// GetAuditHistory retrieves audit history for an entity
func (r *AuditRepository) GetAuditHistory(ctx context.Context, entityID uuid.UUID, limit int32) ([]*AuditLogEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workspace_id, entity_type, entity_id, operation,
		       old_values, new_values, source_system, user_id, source_ip, created_at
		FROM persistence_audit_log
		WHERE entity_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, entityID, limit)

	if err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "failed to query audit history",
			Err:     err,
		}
	}
	defer rows.Close()

	var entries []*AuditLogEntry
	for rows.Next() {
		var entry AuditLogEntry
		err := rows.Scan(
			&entry.ID, &entry.WorkspaceID, &entry.EntityType, &entry.EntityID, &entry.Operation,
			&entry.OldValues, &entry.NewValues, &entry.SourceSystem, &entry.UserID, &entry.SourceIP,
			&entry.CreatedAt,
		)
		if err != nil {
			r.log.Error("failed to scan audit entry", "error", err)
			continue
		}
		entries = append(entries, &entry)
	}

	if err = rows.Err(); err != nil {
		return nil, &PersistenceError{
			Code:    ErrCodeDatabase,
			Message: "error iterating audit entries",
			Err:     err,
		}
	}

	return entries, nil
}

// ============================================================================
// DDL Statements
// ============================================================================

const createDiscoverySessionsTable = `
CREATE TABLE IF NOT EXISTS discovery_sessions (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL,
    log_id UUID NOT NULL,
    algorithm VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    model_id UUID,
    metadata JSONB DEFAULT '{}',
    created_by VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_discovery_sessions_workspace ON discovery_sessions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_discovery_sessions_status ON discovery_sessions(status);
CREATE INDEX IF NOT EXISTS idx_discovery_sessions_log_id ON discovery_sessions(log_id);
`

const createDiscoveredModelsTable = `
CREATE TABLE IF NOT EXISTS discovered_models (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    model_type VARCHAR(50) NOT NULL,
    places JSONB,
    transitions JSONB,
    arcs JSONB,
    tree_json JSONB,
    initial_marking JSONB,
    final_marking JSONB,
    fitness_score FLOAT8,
    precision_score FLOAT8,
    generalization_score FLOAT8,
    source_session_id UUID,
    source_log_id UUID NOT NULL,
    is_archived BOOLEAN DEFAULT false,
    version INTEGER DEFAULT 1,
    parent_model_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_discovered_models_workspace ON discovered_models(workspace_id);
CREATE INDEX IF NOT EXISTS idx_discovered_models_type ON discovered_models(model_type);
CREATE INDEX IF NOT EXISTS idx_discovered_models_log_id ON discovered_models(source_log_id);
`

const createConformanceResultsTable = `
CREATE TABLE IF NOT EXISTS conformance_results (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL,
    model_id UUID NOT NULL,
    log_id UUID NOT NULL,
    conformance_type VARCHAR(50) NOT NULL,
    fitness FLOAT8,
    precision FLOAT8,
    generalization FLOAT8,
    is_fitting BOOLEAN,
    trace_fitness JSONB,
    aligned_traces JSONB,
    violated_constraints JSONB,
    total_traces INTEGER,
    fitting_traces INTEGER,
    non_fitting_traces INTEGER,
    execution_time_ms INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conformance_results_workspace ON conformance_results(workspace_id);
CREATE INDEX IF NOT EXISTS idx_conformance_results_model ON conformance_results(model_id);
CREATE INDEX IF NOT EXISTS idx_conformance_results_log ON conformance_results(log_id);
`

const createProcessStatisticsTable = `
CREATE TABLE IF NOT EXISTS process_statistics (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL,
    log_id UUID NOT NULL,
    model_id UUID,
    variant_count INTEGER,
    top_variants JSONB,
    activity_count INTEGER,
    activities JSONB,
    resource_count INTEGER,
    resources JSONB,
    rework_frequency FLOAT8,
    rework_details JSONB,
    custom_metrics JSONB DEFAULT '{}',
    analysis_type VARCHAR(100),
    execution_time_ms INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_process_statistics_workspace ON process_statistics(workspace_id);
CREATE INDEX IF NOT EXISTS idx_process_statistics_log ON process_statistics(log_id);
CREATE INDEX IF NOT EXISTS idx_process_statistics_model ON process_statistics(model_id);
`

const createAuditLogTable = `
CREATE TABLE IF NOT EXISTS persistence_audit_log (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    operation VARCHAR(50) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    source_system VARCHAR(50),
    user_id VARCHAR(255),
    source_ip VARCHAR(45),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_workspace ON persistence_audit_log(workspace_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_entity ON persistence_audit_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON persistence_audit_log(created_at);
`

const createSyncCheckpointsTable = `
CREATE TABLE IF NOT EXISTS persistence_sync_checkpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    entity_type VARCHAR(100),
    last_sync_at TIMESTAMPTZ,
    last_entity_id UUID,
    total_synced INTEGER DEFAULT 0,
    source_system VARCHAR(50),
    destination_system VARCHAR(50),
    status VARCHAR(50) DEFAULT 'completed',
    error_message TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, entity_type, source_system, destination_system)
);

CREATE INDEX IF NOT EXISTS idx_sync_checkpoints_workspace ON persistence_sync_checkpoints(workspace_id);
`
