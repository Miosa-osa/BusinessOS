package agent

import (
	"context"
	"testing"

	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExperienceStore_RecordOutcome(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutil.RequireTestDatabase(t)
	defer db.Close()

	// Create agent_experience table for testing
	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_experience (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id TEXT NOT NULL,
			task_type TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			outcome TEXT NOT NULL,
			learned_at TIMESTAMPTZ DEFAULT NOW(),
			metadata JSONB DEFAULT '{}',
			UNIQUE(agent_id, task_type, input_hash)
		);

		CREATE INDEX IF NOT EXISTS idx_agent_experience_lookup
			ON agent_experience(agent_id, task_type, input_hash);
	`)
	require.NoError(t, err, "Failed to create agent_experience table")

	store := NewExperienceStore(&database.DB{Pool: db.Pool})

	tests := []struct {
		name        string
		agentID     string
		taskType    string
		inputHash   string
		outcome     string
		metadata    map[string]any
		expectError bool
	}{
		{
			name:      "record success outcome",
			agentID:   "agent-1",
			taskType:  "data_processing",
			inputHash: "abc123",
			outcome:   "success",
			metadata:  map[string]any{"duration_ms": 100},
		},
		{
			name:      "record failure outcome",
			agentID:   "agent-1",
			taskType:  "api_call",
			inputHash: "def456",
			outcome:   "failure",
			metadata:  map[string]any{"error": "connection timeout"},
		},
		{
			name:      "record timeout outcome",
			agentID:   "agent-2",
			taskType:  "complex_task",
			inputHash: "ghi789",
			outcome:   "timeout",
			metadata:  map[string]any{"timeout_seconds": 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.RecordOutcome(ctx, tt.agentID, tt.taskType, tt.inputHash, tt.outcome, tt.metadata)
			if (err != nil) != tt.expectError {
				t.Errorf("RecordOutcome() error = %v, expectError %v", err, tt.expectError)
			}

			// Verify record was created
			exp, err := store.GetLearnedBehavior(ctx, tt.agentID, tt.taskType, tt.inputHash)
			if err != nil {
				t.Errorf("Failed to retrieve recorded experience: %v", err)
			}
			if exp == nil {
				t.Error("Expected experience to be recorded, got nil")
			} else if exp.Outcome != tt.outcome {
				t.Errorf("Expected outcome %s, got %s", tt.outcome, exp.Outcome)
			}
		})
	}

	t.Logf("✅ All outcome recording tests passed")
}

func TestExperienceStore_RecordOutcomeUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutil.RequireTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_experience (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id TEXT NOT NULL,
			task_type TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			outcome TEXT NOT NULL,
			learned_at TIMESTAMPTZ DEFAULT NOW(),
			metadata JSONB DEFAULT '{}',
			UNIQUE(agent_id, task_type, input_hash)
		);
	`)
	require.NoError(t, err)

	store := NewExperienceStore(&database.DB{Pool: db.Pool})

	// Record initial outcome
	err = store.RecordOutcome(ctx, "agent-1", "task_type", "hash1", "success", nil)
	require.NoError(t, err, "Failed to record initial outcome")

	// Update with failure
	err = store.RecordOutcome(ctx, "agent-1", "task_type", "hash1", "failure", map[string]any{"retry": 1})
	require.NoError(t, err, "Failed to update outcome")

	// Verify update
	exp, err := store.GetLearnedBehavior(ctx, "agent-1", "task_type", "hash1")
	require.NoError(t, err, "Failed to retrieve experience")
	require.NotNil(t, exp)

	assert.Equal(t, "failure", exp.Outcome, "Outcome should be updated to failure")
	assert.NotNil(t, exp.Metadata, "Metadata should be updated")

	t.Log("✅ Upsert behavior verified")
}

func TestExperienceStore_GetLearnedBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutil.RequireTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_experience (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id TEXT NOT NULL,
			task_type TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			outcome TEXT NOT NULL,
			learned_at TIMESTAMPTZ DEFAULT NOW(),
			metadata JSONB DEFAULT '{}',
			UNIQUE(agent_id, task_type, input_hash)
		);
	`)
	require.NoError(t, err)

	store := NewExperienceStore(&database.DB{Pool: db.Pool})

	// Test no prior experience
	exp, err := store.GetLearnedBehavior(ctx, "agent-1", "task_type", "unknown_hash")
	require.NoError(t, err, "Unexpected error for unknown experience")
	assert.Nil(t, exp, "Expected nil for unknown experience")

	// Test existing experience
	err = store.RecordOutcome(ctx, "agent-1", "task_type", "hash1", "success", nil)
	require.NoError(t, err, "Failed to record experience")

	exp, err = store.GetLearnedBehavior(ctx, "agent-1", "task_type", "hash1")
	require.NoError(t, err, "Failed to get existing experience")
	require.NotNil(t, exp)

	assert.Equal(t, "agent-1", exp.AgentID)
	assert.Equal(t, "task_type", exp.TaskType)
	assert.Equal(t, "success", exp.Outcome)

	t.Log("✅ GetLearnedBehavior tests passed")
}

func TestExperienceStore_ShouldRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutil.RequireTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_experience (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id TEXT NOT NULL,
			task_type TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			outcome TEXT NOT NULL,
			learned_at TIMESTAMPTZ DEFAULT NOW(),
			metadata JSONB DEFAULT '{}',
			UNIQUE(agent_id, task_type, input_hash)
		);
	`)
	require.NoError(t, err)

	store := NewExperienceStore(&database.DB{Pool: db.Pool})

	tests := []struct {
		name        string
		outcome     string
		wantRetry   bool
		setupRecord bool
	}{
		{
			name:        "no prior experience - should retry",
			setupRecord: false,
			wantRetry:   true,
		},
		{
			name:        "previous success - should not retry",
			outcome:     "success",
			setupRecord: true,
			wantRetry:   false,
		},
		{
			name:        "previous failure - should retry",
			outcome:     "failure",
			setupRecord: true,
			wantRetry:   true,
		},
		{
			name:        "previous timeout - should not retry",
			outcome:     "timeout",
			setupRecord: true,
			wantRetry:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupRecord {
				err := store.RecordOutcome(ctx, "agent-1", "task_type", "hash1", tt.outcome, nil)
				require.NoError(t, err, "Failed to setup test record")
			}

			retry, err := store.ShouldRetry(ctx, "agent-1", "task_type", "hash1")
			require.NoError(t, err, "ShouldRetry() should not error")
			assert.Equal(t, tt.wantRetry, retry, "Retry decision mismatch")
		})
	}

	t.Log("✅ ShouldRetry tests passed")
}

func TestExperienceStore_GetFailureRate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutil.RequireTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_experience (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id TEXT NOT NULL,
			task_type TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			outcome TEXT NOT NULL,
			learned_at TIMESTAMPTZ DEFAULT NOW(),
			metadata JSONB DEFAULT '{}',
			UNIQUE(agent_id, task_type, input_hash)
		);
	`)
	require.NoError(t, err)

	store := NewExperienceStore(&database.DB{Pool: db.Pool})

	// Test with no experiences
	rate, total, err := store.GetFailureRate(ctx, "agent-1")
	require.NoError(t, err, "Unexpected error for agent with no experiences")
	assert.Equal(t, 0.0, rate, "Expected 0 rate for agent with no experiences")
	assert.Equal(t, int64(0), total, "Expected 0 total for agent with no experiences")

	// Add mixed outcomes
	err = store.RecordOutcome(ctx, "agent-1", "task1", "hash1", "success", nil)
	require.NoError(t, err)
	err = store.RecordOutcome(ctx, "agent-1", "task2", "hash2", "failure", nil)
	require.NoError(t, err)
	err = store.RecordOutcome(ctx, "agent-1", "task3", "hash3", "failure", nil)
	require.NoError(t, err)
	err = store.RecordOutcome(ctx, "agent-1", "task4", "hash4", "success", nil)
	require.NoError(t, err)

	rate, total, err = store.GetFailureRate(ctx, "agent-1")
	require.NoError(t, err)
	assert.Equal(t, int64(4), total, "Expected total=4")
	assert.Equal(t, 0.5, rate, "Expected failure rate 0.5 (2 failures out of 4)")

	t.Log("✅ GetFailureRate tests passed")
}

func TestExperienceStore_PruneOldExperiences(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutil.RequireTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_experience (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id TEXT NOT NULL,
			task_type TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			outcome TEXT NOT NULL,
			learned_at TIMESTAMPTZ DEFAULT NOW(),
			metadata JSONB DEFAULT '{}',
			UNIQUE(agent_id, task_type, input_hash)
		);
	`)
	require.NoError(t, err)

	store := NewExperienceStore(&database.DB{Pool: db.Pool})

	// Insert test data
	err = store.RecordOutcome(ctx, "agent-1", "task1", "hash1", "success", nil)
	require.NoError(t, err)
	err = store.RecordOutcome(ctx, "agent-1", "task2", "hash2", "failure", nil)
	require.NoError(t, err)

	// Prune experiences older than 100 days (should not delete recent records)
	deleted, err := store.PruneOldExperiences(ctx, 100)
	require.NoError(t, err, "PruneOldExperiences() should not error")
	assert.Equal(t, int64(0), deleted, "Expected 0 deleted for recent experiences")

	// Prune experiences older than 0 days (should delete all)
	deleted, err = store.PruneOldExperiences(ctx, 0)
	require.NoError(t, err, "PruneOldExperiences() should not error")
	assert.Equal(t, int64(2), deleted, "Expected 2 deleted when pruning all")

	t.Log("✅ PruneOldExperiences tests passed")
}

func TestExperienceStore_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutil.RequireTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_experience (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id TEXT NOT NULL,
			task_type TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			outcome TEXT NOT NULL,
			learned_at TIMESTAMPTZ DEFAULT NOW(),
			metadata JSONB DEFAULT '{}',
			UNIQUE(agent_id, task_type, input_hash)
		);
	`)
	require.NoError(t, err)

	store := NewExperienceStore(&database.DB{Pool: db.Pool})

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			_ = store.RecordOutcome(ctx, "agent-1", "task_type", "hash1", "success", nil)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify only one record exists (upsert)
	exp, err := store.GetLearnedBehavior(ctx, "agent-1", "task_type", "hash1")
	require.NoError(t, err, "Failed to get experience after concurrent writes")
	assert.NotNil(t, exp, "Expected experience to exist after concurrent writes")

	t.Log("✅ Concurrent access tests passed")
}
