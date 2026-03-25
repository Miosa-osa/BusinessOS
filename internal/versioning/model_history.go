package versioning

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcessModelVersion represents a version of a discovered process model
type ProcessModelVersion struct {
	ID                uuid.UUID
	ModelID           uuid.UUID
	Version           string // "2.1.3+a7c3e9f1"
	Major             int
	Minor             int
	Patch             int
	ContentHash       string
	CreatedAt         time.Time
	CreatedBy         string
	DiscoverySource   *string // 'inductive', 'heuristic', 'alpha', 'user_edit'
	PreviousVersionID *uuid.UUID
	ModelJSON         json.RawMessage
	DeltaJSON         *json.RawMessage
	NodesCount        int
	EdgesCount        int
	Variants          int
	Fitness           float64
	AverageDuration   float64
	CoveredTraces     int
	ChangeType        string // "major", "minor", "patch"
	NodesAdded        int
	NodesRemoved      int
	EdgesAdded        int
	EdgesRemoved      int
	Description       string
	Tags              []string
	IsReleased        bool
	ReleaseNotes      *string
	ReleasedAt        *time.Time
	ArchivedAt        *time.Time
}

// ModelMetrics holds quantitative metrics for a process model
type ModelMetrics struct {
	NodesCount      int     `json:"nodes_count"`
	EdgesCount      int     `json:"edges_count"`
	Variants        int     `json:"variants"`
	Fitness         float64 `json:"fitness"`
	AverageDuration float64 `json:"average_duration"`
	CoveredTraces   int     `json:"covered_traces"`
}

// ChangeSummary describes what changed between versions
type ChangeSummary struct {
	NodesAdded   int `json:"nodes_added"`
	NodesRemoved int `json:"nodes_removed"`
	EdgesAdded   int `json:"edges_added"`
	EdgesRemoved int `json:"edges_removed"`
}

// VersionDiffResult represents comparison between two versions
type VersionDiffResult struct {
	FromVersion    string            `json:"from_version"`
	ToVersion      string            `json:"to_version"`
	StructuralDiff StructuralDiff    `json:"structural_diff"`
	MetricsDiff    MetricsDiffResult `json:"metrics_diff"`
	ChangeSummary  ChangeSummary     `json:"change_summary"`
	BreakingChanges []string         `json:"breaking_changes"`
}

// StructuralDiff shows node and edge changes
type StructuralDiff struct {
	NodesAdded   []NodeChange `json:"nodes_added"`
	NodesRemoved []NodeChange `json:"nodes_removed"`
	EdgesAdded   []EdgeChange `json:"edges_added"`
	EdgesRemoved []EdgeChange `json:"edges_removed"`
}

// NodeChange represents a single node modification
type NodeChange struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // task, xor_gateway, and_gateway, event
	Label string `json:"label"`
}

// EdgeChange represents a single edge modification
type EdgeChange struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  *string `json:"label,omitempty"`
}

// MetricsDiffResult shows quantitative changes
type MetricsDiffResult struct {
	NodesCount      DiffValue `json:"nodes_count"`
	EdgesCount      DiffValue `json:"edges_count"`
	Variants        DiffValue `json:"variants"`
	Fitness         DiffValue `json:"fitness"`
	AverageDuration DiffValue `json:"average_duration"`
	CoveredTraces   DiffValue `json:"covered_traces"`
}

// DiffValue represents before/after change
type DiffValue struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
	Delta  interface{} `json:"delta"`
}

// RollbackRequest specifies parameters for rolling back to a previous version
type RollbackRequest struct {
	ModelID          uuid.UUID
	TargetVersion    string
	Reason           string
	ApprovedBy       string
	RunningInstances string // "pause", "continue", "replay"
	BackupCurrent    bool
}

// RollbackAudit records when and why a rollback occurred
type RollbackAudit struct {
	ID           uuid.UUID
	ModelID      uuid.UUID
	FromVersion  string
	ToVersion    string
	Reason       string
	ApprovedBy   string
	PerformedAt  time.Time
	InstancesAffected int
}

// RollbackImpact analyzes the effect of rolling back
type RollbackImpact struct {
	CurrentVersion        string
	TargetVersion         string
	BreakingChanges       []string
	InstancesToPause      int
	CompatibleInstances   int
	IncompatibleInstances int
}

// ModelHistoryService manages process model versions
type ModelHistoryService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewModelHistoryService creates a new model history service
func NewModelHistoryService(pool *pgxpool.Pool, logger *slog.Logger) *ModelHistoryService {
	return &ModelHistoryService{
		pool:   pool,
		logger: logger,
	}
}

// CreateVersion saves a new model version to the database
func (s *ModelHistoryService) CreateVersion(
	ctx context.Context,
	modelID uuid.UUID,
	model json.RawMessage,
	metrics ModelMetrics,
	changeType string,
	changeSummary ChangeSummary,
	description string,
	createdBy string,
	discoverySource *string,
	tags []string,
) (*ProcessModelVersion, error) {
	// Get previous version for relationship
	prevVersion, err := s.getLatestVersion(ctx, modelID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("fetch previous version: %w", err)
	}

	// Compute content hash
	contentHash := computeModelHash(model)

	// Determine semantic version
	major, minor, patch := s.determineVersion(changeType, prevVersion)
	version := fmt.Sprintf("%d.%d.%d+%s", major, minor, patch, contentHash[:8])

	// Compute delta if there's a previous version
	var deltaJSON *json.RawMessage
	if prevVersion != nil {
		delta := s.computeDelta(model, prevVersion.ModelJSON)
		deltaJSON = &delta
	}

	// Identify breaking changes
	breaking := identifyBreakingChanges(changeType, changeSummary)

	versionID := uuid.New()

	// Insert into database
	query := `
		INSERT INTO process_model_versions (
			id, model_id, version, major, minor, patch, content_hash,
			created_at, created_by, discovery_source, previous_version_id,
			model_json, delta_json,
			nodes_count, edges_count, variants, fitness, average_duration, covered_traces,
			change_type, nodes_added, nodes_removed, edges_added, edges_removed,
			description, tags
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13,
			$14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24,
			$25, $26
		)
	`

	var prevVersionID *uuid.UUID
	if prevVersion != nil {
		prevVersionID = &prevVersion.ID
	}

	_, err = s.pool.Exec(ctx, query,
		versionID,
		modelID,
		version,
		major, minor, patch,
		contentHash,
		time.Now(),
		createdBy,
		discoverySource,
		prevVersionID,
		model,
		deltaJSON,
		metrics.NodesCount,
		metrics.EdgesCount,
		metrics.Variants,
		metrics.Fitness,
		metrics.AverageDuration,
		metrics.CoveredTraces,
		changeType,
		changeSummary.NodesAdded,
		changeSummary.NodesRemoved,
		changeSummary.EdgesAdded,
		changeSummary.EdgesRemoved,
		description,
		tags,
	)

	if err != nil {
		return nil, fmt.Errorf("insert version: %w", err)
	}

	s.logger.Info("process model version created",
		"model_id", modelID,
		"version", version,
		"change_type", changeType)

	// Fetch and return created version
	return s.GetVersion(ctx, modelID, version)
}

// GetVersion retrieves a specific model version
func (s *ModelHistoryService) GetVersion(
	ctx context.Context,
	modelID uuid.UUID,
	version string,
) (*ProcessModelVersion, error) {
	query := `
		SELECT id, model_id, version, major, minor, patch, content_hash,
		       created_at, created_by, discovery_source, previous_version_id,
		       model_json, delta_json,
		       nodes_count, edges_count, variants, fitness, average_duration, covered_traces,
		       change_type, nodes_added, nodes_removed, edges_added, edges_removed,
		       description, tags, is_released, release_notes, released_at, archived_at
		FROM process_model_versions
		WHERE model_id = $1 AND version = $2
	`

	var v ProcessModelVersion
	err := s.pool.QueryRow(ctx, query, modelID, version).Scan(
		&v.ID, &v.ModelID, &v.Version, &v.Major, &v.Minor, &v.Patch, &v.ContentHash,
		&v.CreatedAt, &v.CreatedBy, &v.DiscoverySource, &v.PreviousVersionID,
		&v.ModelJSON, &v.DeltaJSON,
		&v.NodesCount, &v.EdgesCount, &v.Variants, &v.Fitness, &v.AverageDuration, &v.CoveredTraces,
		&v.ChangeType, &v.NodesAdded, &v.NodesRemoved, &v.EdgesAdded, &v.EdgesRemoved,
		&v.Description, &v.Tags, &v.IsReleased, &v.ReleaseNotes, &v.ReleasedAt, &v.ArchivedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get version: %w", err)
	}

	return &v, nil
}

// GetVersionHistory retrieves all versions of a model
func (s *ModelHistoryService) GetVersionHistory(
	ctx context.Context,
	modelID uuid.UUID,
	limit int,
	offset int,
) ([]*ProcessModelVersion, int, error) {
	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM process_model_versions WHERE model_id = $1 AND archived_at IS NULL`
	err := s.pool.QueryRow(ctx, countQuery, modelID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("count versions: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, model_id, version, major, minor, patch, content_hash,
		       created_at, created_by, discovery_source, previous_version_id,
		       model_json, delta_json,
		       nodes_count, edges_count, variants, fitness, average_duration, covered_traces,
		       change_type, nodes_added, nodes_removed, edges_added, edges_removed,
		       description, tags, is_released, release_notes, released_at, archived_at
		FROM process_model_versions
		WHERE model_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, modelID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query versions: %w", err)
	}
	defer rows.Close()

	var versions []*ProcessModelVersion
	for rows.Next() {
		var v ProcessModelVersion
		err := rows.Scan(
			&v.ID, &v.ModelID, &v.Version, &v.Major, &v.Minor, &v.Patch, &v.ContentHash,
			&v.CreatedAt, &v.CreatedBy, &v.DiscoverySource, &v.PreviousVersionID,
			&v.ModelJSON, &v.DeltaJSON,
			&v.NodesCount, &v.EdgesCount, &v.Variants, &v.Fitness, &v.AverageDuration, &v.CoveredTraces,
			&v.ChangeType, &v.NodesAdded, &v.NodesRemoved, &v.EdgesAdded, &v.EdgesRemoved,
			&v.Description, &v.Tags, &v.IsReleased, &v.ReleaseNotes, &v.ReleasedAt, &v.ArchivedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan version: %w", err)
		}
		versions = append(versions, &v)
	}

	return versions, totalCount, nil
}

// CompareBetweenVersions returns differences between two versions
func (s *ModelHistoryService) CompareBetweenVersions(
	ctx context.Context,
	modelID uuid.UUID,
	fromVersion string,
	toVersion string,
) (*VersionDiffResult, error) {
	from, err := s.GetVersion(ctx, modelID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("get from version: %w", err)
	}

	to, err := s.GetVersion(ctx, modelID, toVersion)
	if err != nil {
		return nil, fmt.Errorf("get to version: %w", err)
	}

	// Parse models
	var fromModel, toModel map[string]interface{}
	if err := json.Unmarshal(from.ModelJSON, &fromModel); err != nil {
		return nil, fmt.Errorf("parse from model: %w", err)
	}
	if err := json.Unmarshal(to.ModelJSON, &toModel); err != nil {
		return nil, fmt.Errorf("parse to model: %w", err)
	}

	// Compute structural differences
	structDiff := computeStructuralDiff(fromModel, toModel)

	// Compute metrics differences
	metricsDiff := MetricsDiffResult{
		NodesCount: DiffValue{
			Before: from.NodesCount,
			After:  to.NodesCount,
			Delta:  to.NodesCount - from.NodesCount,
		},
		EdgesCount: DiffValue{
			Before: from.EdgesCount,
			After:  to.EdgesCount,
			Delta:  to.EdgesCount - from.EdgesCount,
		},
		Variants: DiffValue{
			Before: from.Variants,
			After:  to.Variants,
			Delta:  to.Variants - from.Variants,
		},
		Fitness: DiffValue{
			Before: from.Fitness,
			After:  to.Fitness,
			Delta:  to.Fitness - from.Fitness,
		},
		AverageDuration: DiffValue{
			Before: from.AverageDuration,
			After:  to.AverageDuration,
			Delta:  to.AverageDuration - from.AverageDuration,
		},
		CoveredTraces: DiffValue{
			Before: from.CoveredTraces,
			After:  to.CoveredTraces,
			Delta:  to.CoveredTraces - from.CoveredTraces,
		},
	}

	breaking := identifyBreakingChanges(to.ChangeType, ChangeSummary{
		NodesAdded:   to.NodesAdded,
		NodesRemoved: to.NodesRemoved,
		EdgesAdded:   to.EdgesAdded,
		EdgesRemoved: to.EdgesRemoved,
	})

	return &VersionDiffResult{
		FromVersion:     fromVersion,
		ToVersion:       toVersion,
		StructuralDiff:  structDiff,
		MetricsDiff:     metricsDiff,
		ChangeSummary: ChangeSummary{
			NodesAdded:   to.NodesAdded,
			NodesRemoved: to.NodesRemoved,
			EdgesAdded:   to.EdgesAdded,
			EdgesRemoved: to.EdgesRemoved,
		},
		BreakingChanges: breaking,
	}, nil
}

// ReleaseVersion marks a version as ready for production use
func (s *ModelHistoryService) ReleaseVersion(
	ctx context.Context,
	modelID uuid.UUID,
	versionID uuid.UUID,
	notes string,
) error {
	// Check fitness threshold
	query := `SELECT fitness FROM process_model_versions WHERE id = $1 AND model_id = $2`
	var fitness float64
	err := s.pool.QueryRow(ctx, query, versionID, modelID).Scan(&fitness)
	if err != nil {
		return fmt.Errorf("fetch version: %w", err)
	}

	if fitness < 0.85 {
		return fmt.Errorf("insufficient fitness: %.2f < 0.85", fitness)
	}

	// Update to released
	updateQuery := `
		UPDATE process_model_versions
		SET is_released = true, released_at = $1, release_notes = $2
		WHERE id = $3 AND model_id = $4
	`

	_, err = s.pool.Exec(ctx, updateQuery, time.Now(), notes, versionID, modelID)
	if err != nil {
		return fmt.Errorf("release version: %w", err)
	}

	s.logger.Info("model version released", "model_id", modelID, "version_id", versionID)
	return nil
}

// RollbackToVersion restores a model to a previous version
func (s *ModelHistoryService) RollbackToVersion(
	ctx context.Context,
	req RollbackRequest,
) error {
	// Get target version
	target, err := s.GetVersion(ctx, req.ModelID, req.TargetVersion)
	if err != nil {
		return fmt.Errorf("get target version: %w", err)
	}

	// Only allow rollback to released versions
	if !target.IsReleased {
		return fmt.Errorf("cannot rollback to unreleased version: %s", req.TargetVersion)
	}

	// Create audit record
	auditID := uuid.New()
	auditQuery := `
		INSERT INTO model_version_rollback_audits (id, model_id, from_version, to_version, reason, approved_by, performed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = s.pool.Exec(ctx, auditQuery,
		auditID,
		req.ModelID,
		"", // Will be populated from current version
		req.TargetVersion,
		req.Reason,
		req.ApprovedBy,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("create audit record: %w", err)
	}

	// Update model to point to this version
	updateQuery := `
		UPDATE process_models
		SET current_version_id = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = s.pool.Exec(ctx, updateQuery, target.ID, time.Now(), req.ModelID)
	if err != nil {
		return fmt.Errorf("update model: %w", err)
	}

	s.logger.Info("model rolled back",
		"model_id", req.ModelID,
		"target_version", req.TargetVersion,
		"reason", req.Reason)

	return nil
}

// AnalyzeRollbackImpact determines effects of rolling back to a version
func (s *ModelHistoryService) AnalyzeRollbackImpact(
	ctx context.Context,
	modelID uuid.UUID,
	targetVersion string,
) (*RollbackImpact, error) {
	target, err := s.GetVersion(ctx, modelID, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("get target version: %w", err)
	}

	// Get current version
	current, err := s.getLatestVersion(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("get current version: %w", err)
	}

	// Compute diff
	diff, err := s.CompareBetweenVersions(ctx, modelID, target.Version, current.Version)
	if err != nil {
		return nil, fmt.Errorf("compute diff: %w", err)
	}

	return &RollbackImpact{
		CurrentVersion:        current.Version,
		TargetVersion:         target.Version,
		BreakingChanges:       diff.BreakingChanges,
		InstancesToPause:      len(diff.BreakingChanges), // Approximation
		CompatibleInstances:   0,
		IncompatibleInstances: len(diff.BreakingChanges),
	}, nil
}

// Helper methods

func (s *ModelHistoryService) getLatestVersion(
	ctx context.Context,
	modelID uuid.UUID,
) (*ProcessModelVersion, error) {
	query := `
		SELECT id, model_id, version, major, minor, patch, content_hash,
		       created_at, created_by, discovery_source, previous_version_id,
		       model_json, delta_json,
		       nodes_count, edges_count, variants, fitness, average_duration, covered_traces,
		       change_type, nodes_added, nodes_removed, edges_added, edges_removed,
		       description, tags, is_released, release_notes, released_at, archived_at
		FROM process_model_versions
		WHERE model_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var v ProcessModelVersion
	err := s.pool.QueryRow(ctx, query, modelID).Scan(
		&v.ID, &v.ModelID, &v.Version, &v.Major, &v.Minor, &v.Patch, &v.ContentHash,
		&v.CreatedAt, &v.CreatedBy, &v.DiscoverySource, &v.PreviousVersionID,
		&v.ModelJSON, &v.DeltaJSON,
		&v.NodesCount, &v.EdgesCount, &v.Variants, &v.Fitness, &v.AverageDuration, &v.CoveredTraces,
		&v.ChangeType, &v.NodesAdded, &v.NodesRemoved, &v.EdgesAdded, &v.EdgesRemoved,
		&v.Description, &v.Tags, &v.IsReleased, &v.ReleaseNotes, &v.ReleasedAt, &v.ArchivedAt,
	)

	return &v, err
}

func (s *ModelHistoryService) determineVersion(
	changeType string,
	prevVersion *ProcessModelVersion,
) (major, minor, patch int) {
	if prevVersion == nil {
		return 1, 0, 0
	}

	major = prevVersion.Major
	minor = prevVersion.Minor
	patch = prevVersion.Patch

	switch changeType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	return
}

func (s *ModelHistoryService) computeDelta(current, previous json.RawMessage) json.RawMessage {
	// Simplified delta: just return empty for now
	// In production, compute actual differences
	return json.RawMessage(`{}`)
}

// Utility functions

func computeModelHash(model json.RawMessage) string {
	h := sha256.New()
	h.Write(model)
	return hex.EncodeToString(h.Sum(nil))
}

func computeStructuralDiff(
	from map[string]interface{},
	to map[string]interface{},
) StructuralDiff {
	// Simplified structural comparison
	// In production, parse BPMN/Petri net structures
	return StructuralDiff{
		NodesAdded:   []NodeChange{},
		NodesRemoved: []NodeChange{},
		EdgesAdded:   []EdgeChange{},
		EdgesRemoved: []EdgeChange{},
	}
}

func identifyBreakingChanges(changeType string, summary ChangeSummary) []string {
	var breaking []string

	if changeType == "major" || summary.NodesRemoved > 0 {
		breaking = append(breaking, "nodes_removed")
	}
	if changeType == "major" || summary.EdgesRemoved > 0 {
		breaking = append(breaking, "edges_removed")
	}

	return breaking
}
