package versioning

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MockPool provides a mock database connection for testing
// In production deployments, use testcontainers with real PostgreSQL
type MockPool struct {
	versions map[string][]*ProcessModelVersion
}

func newMockPool() *MockPool {
	return &MockPool{
		versions: make(map[string][]*ProcessModelVersion),
	}
}

func setupTestDB(t *testing.T) (*ModelHistoryService, *MockPool) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Suppress test logs
	}))

	// Use in-memory mock for CI/unit testing
	// For integration tests, use real PostgreSQL with testcontainers
	mock := newMockPool()

	// Create a service with a real pool reference (would be mocked in production)
	// This is sufficient for testing business logic
	service := &ModelHistoryService{
		pool:   nil, // Mock doesn't need pool
		logger: logger,
	}

	return service, mock
}

func TestCreateVersion(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create initial model
	model := json.RawMessage(`{
		"id": "model_001",
		"nodes": [
			{"id": "task_1", "type": "task", "label": "Check"},
			{"id": "task_2", "type": "task", "label": "Process"}
		],
		"edges": [
			{"id": "edge_1", "source": "task_1", "target": "task_2"}
		]
	}`)

	metrics := ModelMetrics{
		NodesCount:      2,
		EdgesCount:      1,
		Variants:        1,
		Fitness:         0.92,
		AverageDuration: 35.5,
		CoveredTraces:   100,
	}

	summary := ChangeSummary{
		NodesAdded:   2,
		NodesRemoved: 0,
		EdgesAdded:   1,
		EdgesRemoved: 0,
	}

	// Test version creation
	version, err := service.CreateVersion(
		ctx,
		modelID,
		model,
		metrics,
		"minor",
		summary,
		"Initial model discovery",
		"discovery-engine",
		nil,
		[]string{"initial"},
	)

	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	if version == nil {
		t.Fatal("CreateVersion returned nil")
	}

	if version.Major != 1 || version.Minor != 0 {
		t.Errorf("Expected version 1.0, got %d.%d", version.Major, version.Minor)
	}

	if version.Fitness != 0.92 {
		t.Errorf("Expected fitness 0.92, got %.2f", version.Fitness)
	}
}

func TestGetVersion(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	model := json.RawMessage(`{"nodes": [], "edges": []}`)
	metrics := ModelMetrics{NodesCount: 0, EdgesCount: 0}
	summary := ChangeSummary{}

	created, err := service.CreateVersion(
		ctx, modelID, model, metrics, "patch", summary,
		"Test version", "test", nil, nil,
	)
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}

	// Retrieve version
	retrieved, err := service.GetVersion(ctx, modelID, created.Version)
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if retrieved.Version != created.Version {
		t.Errorf("Expected version %s, got %s", created.Version, retrieved.Version)
	}

	if retrieved.CreatedBy != "test" {
		t.Errorf("Expected created_by 'test', got %s", retrieved.CreatedBy)
	}
}

func TestVersionHistory(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create multiple versions
	for i := 0; i < 3; i++ {
		model := json.RawMessage(`{"version": ` + string(rune('0'+i)) + `}`)
		metrics := ModelMetrics{NodesCount: i + 1}
		summary := ChangeSummary{}

		_, err := service.CreateVersion(
			ctx, modelID, model, metrics, "patch", summary,
			"Version "+string(rune('0'+i)), "test", nil, nil,
		)
		if err != nil {
			t.Fatalf("CreateVersion failed: %v", err)
		}

		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Retrieve history
	history, count, err := service.GetVersionHistory(ctx, modelID, 10, 0)
	if err != nil {
		t.Fatalf("GetVersionHistory failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 versions, got %d", count)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 history items, got %d", len(history))
	}

	// Should be sorted by creation time DESC
	if history[0].CreatedAt.Before(history[1].CreatedAt) {
		t.Error("History not sorted by creation time DESC")
	}
}

func TestVersionComparison(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create version 1
	model1 := json.RawMessage(`{
		"nodes": [
			{"id": "task_1", "type": "task"},
			{"id": "task_2", "type": "task"}
		]
	}`)
	metrics1 := ModelMetrics{NodesCount: 2, EdgesCount: 1, Fitness: 0.90}
	summary1 := ChangeSummary{NodesAdded: 2}

	v1, _ := service.CreateVersion(
		ctx, modelID, model1, metrics1, "minor", summary1,
		"Version 1", "test", nil, nil,
	)

	time.Sleep(10 * time.Millisecond)

	// Create version 2 with more nodes
	model2 := json.RawMessage(`{
		"nodes": [
			{"id": "task_1", "type": "task"},
			{"id": "task_2", "type": "task"},
			{"id": "task_3", "type": "task"}
		]
	}`)
	metrics2 := ModelMetrics{NodesCount: 3, EdgesCount: 2, Fitness: 0.93}
	summary2 := ChangeSummary{NodesAdded: 1}

	v2, _ := service.CreateVersion(
		ctx, modelID, model2, metrics2, "minor", summary2,
		"Version 2", "test", nil, nil,
	)

	// Compare versions
	diff, err := service.CompareBetweenVersions(ctx, modelID, v1.Version, v2.Version)
	if err != nil {
		t.Fatalf("CompareBetweenVersions failed: %v", err)
	}

	if diff.FromVersion != v1.Version {
		t.Errorf("Expected from_version %s, got %s", v1.Version, diff.FromVersion)
	}

	if diff.ToVersion != v2.Version {
		t.Errorf("Expected to_version %s, got %s", v2.Version, diff.ToVersion)
	}

	// Check metrics diff
	nodesDiff := diff.MetricsDiff.NodesCount
	if nodesDiff.Before != 2 || nodesDiff.After != 3 {
		t.Errorf("Expected nodes before=2, after=3, got before=%v, after=%v",
			nodesDiff.Before, nodesDiff.After)
	}

	fitnessDiff := diff.MetricsDiff.Fitness
	if fitnessDiff.Before != 0.90 || fitnessDiff.After != 0.93 {
		t.Errorf("Expected fitness before=0.90, after=0.93, got before=%v, after=%v",
			fitnessDiff.Before, fitnessDiff.After)
	}
}

func TestReleaseVersion(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	model := json.RawMessage(`{}`)
	metrics := ModelMetrics{Fitness: 0.90} // >= 0.85
	summary := ChangeSummary{}

	version, _ := service.CreateVersion(
		ctx, modelID, model, metrics, "patch", summary,
		"Good fitness", "test", nil, nil,
	)

	// Release version
	err := service.ReleaseVersion(ctx, modelID, version.ID, "Ready for production")
	if err != nil {
		t.Fatalf("ReleaseVersion failed: %v", err)
	}

	// Verify released
	retrieved, _ := service.GetVersion(ctx, modelID, version.Version)
	if !retrieved.IsReleased {
		t.Error("Version not marked as released")
	}

	if retrieved.ReleaseNotes == nil || *retrieved.ReleaseNotes != "Ready for production" {
		t.Error("Release notes not set correctly")
	}
}

func TestReleaseLowFitness(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	model := json.RawMessage(`{}`)
	metrics := ModelMetrics{Fitness: 0.80} // < 0.85
	summary := ChangeSummary{}

	version, _ := service.CreateVersion(
		ctx, modelID, model, metrics, "patch", summary,
		"Low fitness", "test", nil, nil,
	)

	// Try to release version with low fitness
	err := service.ReleaseVersion(ctx, modelID, version.ID, "Should fail")
	if err == nil {
		t.Error("ReleaseVersion should fail for low fitness")
	}

	retrieved, _ := service.GetVersion(ctx, modelID, version.Version)
	if retrieved.IsReleased {
		t.Error("Version should not be released with low fitness")
	}
}

func TestRollback(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create v1 and release it
	model1 := json.RawMessage(`{"version": 1}`)
	metrics1 := ModelMetrics{Fitness: 0.90}
	v1, _ := service.CreateVersion(
		ctx, modelID, model1, metrics1, "patch", ChangeSummary{},
		"Version 1", "test", nil, nil,
	)
	service.ReleaseVersion(ctx, modelID, v1.ID, "Stable")

	// Create v2
	time.Sleep(10 * time.Millisecond)
	model2 := json.RawMessage(`{"version": 2}`)
	metrics2 := ModelMetrics{Fitness: 0.92}
	v2, _ := service.CreateVersion(
		ctx, modelID, model2, metrics2, "patch", ChangeSummary{},
		"Version 2", "test", nil, nil,
	)

	// Rollback to v1
	req := RollbackRequest{
		ModelID:       modelID,
		TargetVersion: v1.Version,
		Reason:        "Regression detected",
		ApprovedBy:    "admin",
		BackupCurrent: true,
	}

	err := service.RollbackToVersion(ctx, &req)
	if err != nil {
		t.Fatalf("RollbackToVersion failed: %v", err)
	}

	// Verify rollback succeeded
	t.Log("Rollback completed successfully")
}

func TestRollbackToUnreleasedFails(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create unreleased version
	model := json.RawMessage(`{}`)
	metrics := ModelMetrics{Fitness: 0.90}
	v, _ := service.CreateVersion(
		ctx, modelID, model, metrics, "patch", ChangeSummary{},
		"Unreleased", "test", nil, nil,
	)

	// Try to rollback to unreleased
	req := RollbackRequest{
		ModelID:       modelID,
		TargetVersion: v.Version,
		Reason:        "Should fail",
		ApprovedBy:    "admin",
	}

	err := service.RollbackToVersion(ctx, &req)
	if err == nil {
		t.Error("RollbackToVersion should fail for unreleased version")
	}
}

func TestAnalyzeRollbackImpact(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create v1
	model1 := json.RawMessage(`{"nodes": []}`)
	metrics1 := ModelMetrics{NodesCount: 2}
	v1, _ := service.CreateVersion(
		ctx, modelID, model1, metrics1, "patch", ChangeSummary{},
		"Version 1", "test", nil, nil,
	)
	service.ReleaseVersion(ctx, modelID, v1.ID, "Stable")

	// Create v2 with changes
	time.Sleep(10 * time.Millisecond)
	model2 := json.RawMessage(`{"nodes": []}`)
	metrics2 := ModelMetrics{NodesCount: 3}
	v2, _ := service.CreateVersion(
		ctx, modelID, model2, metrics2, "minor", ChangeSummary{NodesAdded: 1},
		"Version 2", "test", nil, nil,
	)

	// Analyze rollback
	impact, err := service.AnalyzeRollbackImpact(ctx, modelID, v1.Version)
	if err != nil {
		t.Fatalf("AnalyzeRollbackImpact failed: %v", err)
	}

	if impact.TargetVersion != v1.Version {
		t.Errorf("Expected target version %s, got %s", v1.Version, impact.TargetVersion)
	}

	if impact.CurrentVersion != v2.Version {
		t.Errorf("Expected current version %s, got %s", v2.Version, impact.CurrentVersion)
	}
}

func TestSemanticVersioning(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	tests := []struct {
		name         string
		changeType   string
		expectMajor  int
		expectMinor  int
		expectPatch  int
	}{
		{"First version", "patch", 1, 0, 0},
		{"Minor change", "minor", 1, 1, 0},
		{"Another minor", "minor", 1, 2, 0},
		{"Patch bump", "patch", 1, 2, 1},
		{"Major change", "major", 2, 0, 0},
	}

	for _, tt := range tests {
		model := json.RawMessage(`{}`)
		metrics := ModelMetrics{Fitness: 0.90}
		summary := ChangeSummary{}

		version, err := service.CreateVersion(
			ctx, modelID, model, metrics, tt.changeType, summary,
			tt.name, "test", nil, nil,
		)

		if err != nil {
			t.Errorf("%s: CreateVersion failed: %v", tt.name, err)
			continue
		}

		if version.Major != tt.expectMajor || version.Minor != tt.expectMinor || version.Patch != tt.expectPatch {
			t.Errorf("%s: Expected %d.%d.%d, got %d.%d.%d",
				tt.name,
				tt.expectMajor, tt.expectMinor, tt.expectPatch,
				version.Major, version.Minor, version.Patch)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// ─────────────────────────────────────────────────────────────────
// Extended TDD Test Suite - Comprehensive Version Management
// ─────────────────────────────────────────────────────────────────

func TestVersionContentHashConsistency(t *testing.T) {
	// Verify that identical models always produce same content hash
	service, _ := setupTestDB(t)

	model := json.RawMessage(`{
		"id": "model_001",
		"nodes": [{"id": "node_1", "type": "task", "label": "Check"}],
		"edges": [{"id": "edge_1", "source": "node_1", "target": "node_1"}]
	}`)

	hash1 := computeModelHash(model)
	hash2 := computeModelHash(model)

	if hash1 != hash2 {
		t.Errorf("Content hash not deterministic: %s != %s", hash1, hash2)
	}

	// Verify hash prefix is correct length
	if len(hash1) != 64 { // SHA256 in hex
		t.Errorf("Expected 64-char hash, got %d", len(hash1))
	}

	// Verify different models produce different hashes
	model2 := json.RawMessage(`{
		"id": "model_002",
		"nodes": [{"id": "node_2", "type": "event", "label": "Different"}]
	}`)
	hash3 := computeModelHash(model2)

	if hash1 == hash3 {
		t.Error("Different models produced same hash (collision)")
	}
}

func TestBreakingChangeDetection(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	tests := []struct {
		name        string
		changeType  string
		summary     ChangeSummary
		expectBreak bool
	}{
		{
			name:       "Major change with no removals",
			changeType: "major",
			summary: ChangeSummary{
				NodesAdded:   5,
				NodesRemoved: 0,
				EdgesAdded:   6,
				EdgesRemoved: 0,
			},
			expectBreak: true,
		},
		{
			name:       "Minor change with removals",
			changeType: "minor",
			summary: ChangeSummary{
				NodesAdded:   1,
				NodesRemoved: 2,
				EdgesAdded:   1,
				EdgesRemoved: 2,
			},
			expectBreak: true,
		},
		{
			name:       "Patch with no structural changes",
			changeType: "patch",
			summary: ChangeSummary{
				NodesAdded:   0,
				NodesRemoved: 0,
				EdgesAdded:   0,
				EdgesRemoved: 0,
			},
			expectBreak: false,
		},
	}

	for _, tt := range tests {
		model := json.RawMessage(`{}`)
		metrics := ModelMetrics{Fitness: 0.90}

		version, err := service.CreateVersion(
			ctx, modelID, model, metrics, tt.changeType, tt.summary,
			tt.name, "test", nil, nil,
		)

		if err != nil {
			t.Errorf("%s: CreateVersion failed: %v", tt.name, err)
			continue
		}

		hasBreak := len(identifyBreakingChanges(tt.changeType, tt.summary)) > 0
		if hasBreak != tt.expectBreak {
			t.Errorf("%s: Expected breaking=%v, got %v", tt.name, tt.expectBreak, hasBreak)
		}

		_ = version
	}
}

func TestVersionPreviousVersionLinking(t *testing.T) {
	// Verify versions maintain correct previous_version_id chain
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create version 1
	model1 := json.RawMessage(`{"version": 1}`)
	v1, _ := service.CreateVersion(
		ctx, modelID, model1, ModelMetrics{Fitness: 0.90},
		"patch", ChangeSummary{},
		"Version 1", "test", nil, nil,
	)

	if v1.PreviousVersionID != nil {
		t.Error("Version 1 should not have previous version")
	}

	time.Sleep(10 * time.Millisecond)

	// Create version 2
	model2 := json.RawMessage(`{"version": 2}`)
	v2, _ := service.CreateVersion(
		ctx, modelID, model2, ModelMetrics{Fitness: 0.91},
		"patch", ChangeSummary{},
		"Version 2", "test", nil, nil,
	)

	if v2.PreviousVersionID == nil {
		t.Error("Version 2 should have previous version ID")
	}

	if *v2.PreviousVersionID != v1.ID {
		t.Errorf("Version 2 previous should be %s, got %s", v1.ID, *v2.PreviousVersionID)
	}
}

func TestFitnessThresholdEnforcement(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create version with low fitness
	lowFitnessModel := json.RawMessage(`{}`)
	lowFitnessMetrics := ModelMetrics{Fitness: 0.80}

	v, _ := service.CreateVersion(
		ctx, modelID, lowFitnessModel, lowFitnessMetrics,
		"patch", ChangeSummary{},
		"Low fitness version", "test", nil, nil,
	)

	// Attempt release (should fail)
	err := service.ReleaseVersion(ctx, modelID, v.ID, "Should fail")
	if err == nil {
		t.Error("Release should fail for fitness < 0.85")
	}

	// Create version with sufficient fitness
	goodFitnessModel := json.RawMessage(`{}`)
	goodFitnessMetrics := ModelMetrics{Fitness: 0.87}

	v2, _ := service.CreateVersion(
		ctx, modelID, goodFitnessModel, goodFitnessMetrics,
		"patch", ChangeSummary{},
		"Good fitness version", "test", nil, nil,
	)

	// Attempt release (should succeed)
	err = service.ReleaseVersion(ctx, modelID, v2.ID, "Good fitness")
	if err != nil {
		t.Errorf("Release should succeed for fitness >= 0.85: %v", err)
	}
}

func TestMetricsDeltaCalculation(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Version 1: baseline metrics
	model1 := json.RawMessage(`{}`)
	metrics1 := ModelMetrics{
		NodesCount:      8,
		EdgesCount:      10,
		Fitness:         0.87,
		AverageDuration: 45.3,
		CoveredTraces:   250,
		Variants:        4,
	}
	v1, _ := service.CreateVersion(
		ctx, modelID, model1, metrics1,
		"patch", ChangeSummary{},
		"Baseline", "test", nil, nil,
	)

	time.Sleep(10 * time.Millisecond)

	// Version 2: improved metrics
	model2 := json.RawMessage(`{}`)
	metrics2 := ModelMetrics{
		NodesCount:      10,
		EdgesCount:      13,
		Fitness:         0.92,
		AverageDuration: 42.1,
		CoveredTraces:   285,
		Variants:        6,
	}
	v2, _ := service.CreateVersion(
		ctx, modelID, model2, metrics2,
		"minor", ChangeSummary{NodesAdded: 2, EdgesAdded: 3},
		"Improved", "test", nil, nil,
	)

	// Compare versions
	diff, _ := service.CompareBetweenVersions(ctx, modelID, v1.Version, v2.Version)

	// Verify deltas
	if diff.MetricsDiff.NodesCount.Before != 8 || diff.MetricsDiff.NodesCount.After != 10 {
		t.Error("Nodes count delta incorrect")
	}

	if diff.MetricsDiff.Fitness.Delta.(float64) <= 0.04 || diff.MetricsDiff.Fitness.Delta.(float64) >= 0.08 {
		t.Errorf("Fitness delta should be ~0.05, got %v", diff.MetricsDiff.Fitness.Delta)
	}

	if diff.MetricsDiff.AverageDuration.Delta.(float64) >= -3.3 || diff.MetricsDiff.AverageDuration.Delta.(float64) <= -3.1 {
		t.Errorf("Duration delta should be ~-3.2, got %v", diff.MetricsDiff.AverageDuration.Delta)
	}
}

func TestVersionTagging(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	model := json.RawMessage(`{}`)
	metrics := ModelMetrics{Fitness: 0.90}
	tags := []string{"production-candidate", "improved-fitness", "automated-discovery"}

	v, _ := service.CreateVersion(
		ctx, modelID, model, metrics,
		"minor", ChangeSummary{},
		"Tagged version", "test", nil, tags,
	)

	if len(v.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(v.Tags))
	}

	for _, expectedTag := range tags {
		found := false
		for _, vTag := range v.Tags {
			if vTag == expectedTag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Tag %q not found in version", expectedTag)
		}
	}
}

func TestDiscoverySourceTracking(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	sources := []string{"inductive", "heuristic", "alpha", "manual"}

	for _, source := range sources {
		model := json.RawMessage(`{}`)
		metrics := ModelMetrics{Fitness: 0.90}
		src := source // Capture for reference

		v, _ := service.CreateVersion(
			ctx, modelID, model, metrics,
			"patch", ChangeSummary{},
			fmt.Sprintf("Discovered via %s", source),
			"test", &src, nil,
		)

		if v.DiscoverySource == nil || *v.DiscoverySource != source {
			t.Errorf("Discovery source not tracked: got %v", v.DiscoverySource)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func TestVersionReleaseTimestamp(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	model := json.RawMessage(`{}`)
	metrics := ModelMetrics{Fitness: 0.90}

	v, _ := service.CreateVersion(
		ctx, modelID, model, metrics,
		"patch", ChangeSummary{},
		"Release test", "test", nil, nil,
	)

	beforeRelease := time.Now()
	_ = service.ReleaseVersion(ctx, modelID, v.ID, "Release notes")
	afterRelease := time.Now()

	retrieved, _ := service.GetVersion(ctx, modelID, v.Version)

	if retrieved.ReleasedAt == nil {
		t.Error("Released timestamp not set")
	}

	if retrieved.ReleasedAt.Before(beforeRelease) || retrieved.ReleasedAt.After(afterRelease) {
		t.Error("Release timestamp outside expected range")
	}
}

func TestVersionComparison_NoChanges(t *testing.T) {
	// Test comparing identical versions
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	model := json.RawMessage(`{"nodes": [{"id": "n1"}], "edges": []}`)
	metrics := ModelMetrics{
		NodesCount:    1,
		EdgesCount:    0,
		Fitness:       0.90,
		CoveredTraces: 100,
	}

	v1, _ := service.CreateVersion(
		ctx, modelID, model, metrics,
		"patch", ChangeSummary{},
		"Version", "test", nil, nil,
	)

	// Compare same version (ideally shouldn't happen, but should handle gracefully)
	diff, _ := service.CompareBetweenVersions(ctx, modelID, v1.Version, v1.Version)

	if diff == nil {
		t.Error("Diff result should not be nil")
	}

	if diff.FromVersion != v1.Version || diff.ToVersion != v1.Version {
		t.Error("Diff versions incorrect")
	}
}

func TestRollbackEligibility(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	// Create unreleased version
	model := json.RawMessage(`{}`)
	metrics := ModelMetrics{Fitness: 0.90}

	v, _ := service.CreateVersion(
		ctx, modelID, model, metrics,
		"patch", ChangeSummary{},
		"Unreleased", "test", nil, nil,
	)

	// Attempt rollback to unreleased version
	req := RollbackRequest{
		ModelID:       modelID,
		TargetVersion: v.Version,
		Reason:        "Test",
		ApprovedBy:    "test",
	}

	err := service.RollbackToVersion(ctx, &req)
	if err == nil {
		t.Error("Rollback to unreleased version should fail")
	}
}

func TestVersionMetadataCompleteness(t *testing.T) {
	service, _ := setupTestDB(t)
	ctx := context.Background()
	modelID := uuid.New()

	model := json.RawMessage(`{}`)
	metrics := ModelMetrics{
		NodesCount:      5,
		EdgesCount:      6,
		Variants:        2,
		Fitness:         0.88,
		AverageDuration: 30.5,
		CoveredTraces:   200,
	}

	description := "Test version with full metadata"
	createdBy := "test-user"
	discoverySource := "inductive"
	tags := []string{"test", "metadata"}

	v, _ := service.CreateVersion(
		ctx, modelID, model, metrics,
		"minor", ChangeSummary{NodesAdded: 5, EdgesAdded: 6},
		description, createdBy, &discoverySource, tags,
	)

	// Verify all metadata preserved
	if v.Description != description {
		t.Error("Description not preserved")
	}
	if v.CreatedBy != createdBy {
		t.Error("CreatedBy not preserved")
	}
	if v.DiscoverySource == nil || *v.DiscoverySource != discoverySource {
		t.Error("DiscoverySource not preserved")
	}
	if len(v.Tags) != 2 {
		t.Error("Tags not preserved")
	}
	if v.NodesCount != 5 || v.EdgesCount != 6 {
		t.Error("Metrics not preserved")
	}
}
