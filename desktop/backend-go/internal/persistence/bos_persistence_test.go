// Package persistence provides comprehensive tests for the BOS persistence layer
package persistence

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDatabase represents a test database connection
type TestDatabase struct {
	pool *pgxpool.Pool
	t    *testing.T
}

// setupTestDB creates a test database connection
// Set DATABASE_URL environment variable to test against a real PostgreSQL instance
func setupTestDB(t *testing.T) *TestDatabase {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping persistence tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	return &TestDatabase{
		pool: pool,
		t:    t,
	}
}

// cleanup closes the database connection
func (td *TestDatabase) cleanup() {
	if td.pool != nil {
		td.pool.Close()
	}
}

// dropTestTables removes all BOS tables (for clean tests)
func (td *TestDatabase) dropTestTables(ctx context.Context) {
	tables := []string{
		"persistence_sync_checkpoints",
		"persistence_audit_log",
		"process_statistics",
		"conformance_results",
		"discovered_models",
		"discovery_sessions",
	}

	for _, table := range tables {
		sql := "DROP TABLE IF EXISTS " + table + " CASCADE"
		if _, err := td.pool.Exec(ctx, sql); err != nil {
			td.t.Logf("warning: failed to drop table %s: %v", table, err)
		}
	}
}

// initializeTestSchema creates all tables for testing
func (td *TestDatabase) initializeTestSchema(ctx context.Context) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	sm := NewSchemaManager(td.pool, log)

	if err := sm.InitializeSchema(ctx); err != nil {
		td.t.Fatalf("failed to initialize schema: %v", err)
	}
}

// ============================================================================
// Model Repository Tests
// ============================================================================

func TestModelRepository_SaveAndRetrieve(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	td := setupTestDB(t)
	defer td.cleanup()

	ctx := context.Background()
	td.dropTestTables(ctx)
	td.initializeTestSchema(ctx)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	repo := NewModelRepository(td.pool, log)

	workspaceID := uuid.New()
	logID := uuid.New()

	// Create test model
	places := json.RawMessage(`[{"id":"p1","name":"source","initial_marking":1}]`)
	transitions := json.RawMessage(`[{"id":"t1","label":"activity","name":"Activity"}]`)
	arcs := json.RawMessage(`[{"from":"p1","to":"t1","weight":1}]`)

	model := &DiscoveredModel{
		WorkspaceID:   workspaceID,
		Name:          "Test Model",
		Description:   stringPtr("A test Petri net"),
		ModelType:     "petri_net",
		Places:        &places,
		Transitions:   &transitions,
		Arcs:          &arcs,
		SourceLogID:   logID,
		CreatedBy:     stringPtr("test_user"),
		FitnessScore:  float64Ptr(0.95),
		PrecisionScore: float64Ptr(0.88),
	}

	// Save model
	modelID, err := repo.SaveModel(ctx, model)
	if err != nil {
		t.Fatalf("SaveModel failed: %v", err)
	}

	if modelID == uuid.Nil {
		t.Fatal("SaveModel returned nil ID")
	}

	// Retrieve model
	retrieved, err := repo.GetModel(ctx, modelID)
	if err != nil {
		t.Fatalf("GetModel failed: %v", err)
	}

	// Verify model
	if retrieved.ID != modelID {
		t.Errorf("ID mismatch: expected %s, got %s", modelID, retrieved.ID)
	}

	if retrieved.Name != "Test Model" {
		t.Errorf("Name mismatch: expected 'Test Model', got '%s'", retrieved.Name)
	}

	if retrieved.ModelType != "petri_net" {
		t.Errorf("ModelType mismatch: expected 'petri_net', got '%s'", retrieved.ModelType)
	}

	if retrieved.FitnessScore == nil || *retrieved.FitnessScore != 0.95 {
		t.Errorf("FitnessScore mismatch: expected 0.95, got %v", retrieved.FitnessScore)
	}
}

func TestModelRepository_ListModels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	td := setupTestDB(t)
	defer td.cleanup()

	ctx := context.Background()
	td.dropTestTables(ctx)
	td.initializeTestSchema(ctx)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	repo := NewModelRepository(td.pool, log)

	workspaceID := uuid.New()
	logID := uuid.New()

	// Create multiple models
	places := json.RawMessage(`[{"id":"p1","name":"source"}]`)
	transitions := json.RawMessage(`[{"id":"t1","label":"activity"}]`)
	arcs := json.RawMessage(`[{"from":"p1","to":"t1","weight":1}]`)

	for i := 0; i < 5; i++ {
		model := &DiscoveredModel{
			WorkspaceID: workspaceID,
			Name:        "Test Model " + string(rune(i)),
			ModelType:   "petri_net",
			Places:      &places,
			Transitions: &transitions,
			Arcs:        &arcs,
			SourceLogID: logID,
		}

		_, err := repo.SaveModel(ctx, model)
		if err != nil {
			t.Fatalf("SaveModel failed: %v", err)
		}
	}

	// List models
	models, err := repo.ListModels(ctx, workspaceID, 10, 0)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}

	if len(models) != 5 {
		t.Errorf("Expected 5 models, got %d", len(models))
	}

	// Test pagination
	page1, err := repo.ListModels(ctx, workspaceID, 2, 0)
	if err != nil {
		t.Fatalf("ListModels with limit failed: %v", err)
	}

	if len(page1) != 2 {
		t.Errorf("Expected 2 models on page 1, got %d", len(page1))
	}
}

func TestModelRepository_UpdateScores_OptimisticLocking(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	td := setupTestDB(t)
	defer td.cleanup()

	ctx := context.Background()
	td.dropTestTables(ctx)
	td.initializeTestSchema(ctx)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	repo := NewModelRepository(td.pool, log)

	workspaceID := uuid.New()
	logID := uuid.New()

	// Create model
	places := json.RawMessage(`[{"id":"p1","name":"source"}]`)
	transitions := json.RawMessage(`[{"id":"t1","label":"activity"}]`)
	arcs := json.RawMessage(`[{"from":"p1","to":"t1","weight":1}]`)

	model := &DiscoveredModel{
		WorkspaceID: workspaceID,
		Name:        "Test Model",
		ModelType:   "petri_net",
		Places:      &places,
		Transitions: &transitions,
		Arcs:        &arcs,
		SourceLogID: logID,
	}

	modelID, err := repo.SaveModel(ctx, model)
	if err != nil {
		t.Fatalf("SaveModel failed: %v", err)
	}

	// Update with correct version
	fitness := 0.95
	precision := 0.88
	err = repo.UpdateModelScores(ctx, modelID, 1, &fitness, &precision, nil)
	if err != nil {
		t.Fatalf("UpdateModelScores with correct version failed: %v", err)
	}

	// Try to update with wrong version (should fail)
	err = repo.UpdateModelScores(ctx, modelID, 1, &fitness, &precision, nil)
	if err == nil {
		t.Fatal("UpdateModelScores with wrong version should have failed")
	}

	if perr, ok := err.(*PersistenceError); !ok || perr.Code != ErrCodeVersionMismatch {
		t.Errorf("Expected VersionMismatch error, got %v", err)
	}
}

// ============================================================================
// Conformance Repository Tests
// ============================================================================

func TestConformanceRepository_SaveAndRetrieve(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	td := setupTestDB(t)
	defer td.cleanup()

	ctx := context.Background()
	td.dropTestTables(ctx)
	td.initializeTestSchema(ctx)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	modelRepo := NewModelRepository(td.pool, log)
	confRepo := NewConformanceRepository(td.pool, log)

	workspaceID := uuid.New()
	logID := uuid.New()

	// Create a model first
	places := json.RawMessage(`[{"id":"p1","name":"source"}]`)
	model := &DiscoveredModel{
		WorkspaceID: workspaceID,
		Name:        "Test Model",
		ModelType:   "petri_net",
		Places:      &places,
		SourceLogID: logID,
	}

	modelID, _ := modelRepo.SaveModel(ctx, model)

	// Create conformance result
	traceFitness := json.RawMessage(`[{"trace_id":"t1","fitness":0.98}]`)
	result := &ConformanceResult{
		WorkspaceID:     workspaceID,
		ModelID:         modelID,
		LogID:           logID,
		ConformanceType: "token_replay",
		Fitness:         0.95,
		Precision:       float64Ptr(0.88),
		IsFitting:       true,
		TraceFitness:    &traceFitness,
		TotalTraces:     100,
		FittingTraces:   95,
		NonFittingTraces: 5,
	}

	// Save conformance result
	resultID, err := confRepo.SaveResult(ctx, result)
	if err != nil {
		t.Fatalf("SaveResult failed: %v", err)
	}

	// Retrieve results for model
	results, err := confRepo.GetResultsForModel(ctx, modelID)
	if err != nil {
		t.Fatalf("GetResultsForModel failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	retrieved := results[0]
	if retrieved.ID != resultID {
		t.Errorf("ID mismatch: expected %s, got %s", resultID, retrieved.ID)
	}

	if retrieved.Fitness != 0.95 {
		t.Errorf("Fitness mismatch: expected 0.95, got %f", retrieved.Fitness)
	}

	if retrieved.TotalTraces != 100 {
		t.Errorf("TotalTraces mismatch: expected 100, got %d", retrieved.TotalTraces)
	}
}

// ============================================================================
// Statistics Repository Tests
// ============================================================================

func TestStatisticsRepository_SaveAndRetrieve(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	td := setupTestDB(t)
	defer td.cleanup()

	ctx := context.Background()
	td.dropTestTables(ctx)
	td.initializeTestSchema(ctx)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	statsRepo := NewStatisticsRepository(td.pool, log)

	workspaceID := uuid.New()
	logID := uuid.New()

	// Create statistics
	topVariants := json.RawMessage(`[{"variant":["a","b","c"],"frequency":50}]`)
	activities := json.RawMessage(`[{"activity":"a","frequency":100}]`)

	stats := &ProcessStatistics{
		WorkspaceID:     workspaceID,
		LogID:           logID,
		VariantCount:    25,
		TopVariants:     &topVariants,
		ActivityCount:   5,
		Activities:      &activities,
		ReworkFrequency: float64Ptr(0.12),
		AnalysisType:    "variant",
	}

	// Save statistics
	statsID, err := statsRepo.SaveStatistics(ctx, stats)
	if err != nil {
		t.Fatalf("SaveStatistics failed: %v", err)
	}

	// Retrieve statistics
	retrieved, err := statsRepo.GetStatisticsForLog(ctx, logID)
	if err != nil {
		t.Fatalf("GetStatisticsForLog failed: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("Expected 1 statistic, got %d", len(retrieved))
	}

	s := retrieved[0]
	if s.ID != statsID {
		t.Errorf("ID mismatch: expected %s, got %s", statsID, s.ID)
	}

	if s.VariantCount != 25 {
		t.Errorf("VariantCount mismatch: expected 25, got %d", s.VariantCount)
	}

	if s.ActivityCount != 5 {
		t.Errorf("ActivityCount mismatch: expected 5, got %d", s.ActivityCount)
	}
}

// ============================================================================
// Audit Repository Tests
// ============================================================================

func TestAuditRepository_RecordAndRetrieve(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	td := setupTestDB(t)
	defer td.cleanup()

	ctx := context.Background()
	td.dropTestTables(ctx)
	td.initializeTestSchema(ctx)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	auditRepo := NewAuditRepository(td.pool, log)

	workspaceID := uuid.New()
	entityID := uuid.New()

	// Create audit entry
	oldValues := json.RawMessage(`{"fitness":0.85}`)
	newValues := json.RawMessage(`{"fitness":0.95}`)

	entry := &AuditLogEntry{
		WorkspaceID: workspaceID,
		EntityType:  "discovered_model",
		EntityID:    entityID,
		Operation:   "UPDATE",
		OldValues:   &oldValues,
		NewValues:   &newValues,
		SourceSystem: "bos",
		UserID:      stringPtr("test_user"),
	}

	// Record entry
	entryID, err := auditRepo.RecordEntry(ctx, entry)
	if err != nil {
		t.Fatalf("RecordEntry failed: %v", err)
	}

	// Retrieve history
	history, err := auditRepo.GetAuditHistory(ctx, entityID, 10)
	if err != nil {
		t.Fatalf("GetAuditHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 audit entry, got %d", len(history))
	}

	retrieved := history[0]
	if retrieved.ID != entryID {
		t.Errorf("ID mismatch: expected %s, got %s", entryID, retrieved.ID)
	}

	if retrieved.Operation != "UPDATE" {
		t.Errorf("Operation mismatch: expected UPDATE, got %s", retrieved.Operation)
	}

	if retrieved.SourceSystem != "bos" {
		t.Errorf("SourceSystem mismatch: expected bos, got %s", retrieved.SourceSystem)
	}
}

// ============================================================================
// Data Integrity Tests
// ============================================================================

func TestConcurrentAccessSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	td := setupTestDB(t)
	defer td.cleanup()

	ctx := context.Background()
	td.dropTestTables(ctx)
	td.initializeTestSchema(ctx)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	repo := NewModelRepository(td.pool, log)

	workspaceID := uuid.New()
	logID := uuid.New()

	// Create model
	places := json.RawMessage(`[{"id":"p1","name":"source"}]`)
	model := &DiscoveredModel{
		WorkspaceID: workspaceID,
		Name:        "Test Model",
		ModelType:   "petri_net",
		Places:      &places,
		SourceLogID: logID,
	}

	modelID, _ := repo.SaveModel(ctx, model)

	// Simulate concurrent updates
	done := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func(idx int) {
			fitness := float64(0.85 + float64(idx)*0.05)
			err := repo.UpdateModelScores(ctx, modelID, 1, &fitness, nil, nil)
			done <- err
		}(i)
	}

	// At least one should succeed, others should fail with version mismatch
	successCount := 0
	for i := 0; i < 3; i++ {
		err := <-done
		if err == nil {
			successCount++
		}
	}

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful concurrent update, got %d", successCount)
	}
}

func TestTransactionIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	td := setupTestDB(t)
	defer td.cleanup()

	ctx := context.Background()
	td.dropTestTables(ctx)
	td.initializeTestSchema(ctx)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	modelRepo := NewModelRepository(td.pool, log)
	confRepo := NewConformanceRepository(td.pool, log)
	auditRepo := NewAuditRepository(td.pool, log)

	workspaceID := uuid.New()
	logID := uuid.New()

	// Create model
	places := json.RawMessage(`[{"id":"p1","name":"source"}]`)
	model := &DiscoveredModel{
		WorkspaceID: workspaceID,
		Name:        "Isolation Test",
		ModelType:   "petri_net",
		Places:      &places,
		SourceLogID: logID,
	}

	modelID, _ := modelRepo.SaveModel(ctx, model)

	// Create conformance result
	result := &ConformanceResult{
		WorkspaceID:     workspaceID,
		ModelID:         modelID,
		LogID:           logID,
		ConformanceType: "token_replay",
		Fitness:         0.95,
		TotalTraces:     100,
		FittingTraces:   95,
	}

	_, err := confRepo.SaveResult(ctx, result)
	if err != nil {
		t.Fatalf("SaveResult failed: %v", err)
	}

	// Record audit entry
	entry := &AuditLogEntry{
		WorkspaceID: workspaceID,
		EntityType:  "discovered_model",
		EntityID:    modelID,
		Operation:   "CREATE",
		SourceSystem: "bos",
	}

	_, err = auditRepo.RecordEntry(ctx, entry)
	if err != nil {
		t.Fatalf("RecordEntry failed: %v", err)
	}

	// Verify all data is present
	retrieved, _ := modelRepo.GetModel(ctx, modelID)
	if retrieved == nil {
		t.Fatal("Model not found after isolation test")
	}

	results, _ := confRepo.GetResultsForModel(ctx, modelID)
	if len(results) != 1 {
		t.Errorf("Expected 1 conformance result, got %d", len(results))
	}

	history, _ := auditRepo.GetAuditHistory(ctx, modelID, 10)
	if len(history) != 1 {
		t.Errorf("Expected 1 audit entry, got %d", len(history))
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
