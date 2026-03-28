package agent

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rhl/businessos-backend/internal/database"
)

// Experience represents a learned outcome from a past agent execution
type Experience struct {
	ID        string            `json:"id"`
	AgentID   string            `json:"agent_id"`
	TaskType  string            `json:"task_type"`
	InputHash string            `json:"input_hash"`
	Outcome   string            `json:"outcome"` // success, failure, timeout
	LearnedAt string            `json:"learned_at"`
	Metadata  map[string]any    `json:"metadata"`
}

// ExperienceStore manages agent learning from past outcomes
type ExperienceStore struct {
	db *database.DB
}

// NewExperienceStore creates a new experience store
func NewExperienceStore(db *database.DB) *ExperienceStore {
	return &ExperienceStore{db: db}
}

// RecordOutcome stores an agent execution outcome for future learning
func (e *ExperienceStore) RecordOutcome(ctx context.Context, agentID, taskType, inputHash, outcome string, metadata map[string]any) error {
	const query = `
		INSERT INTO agent_experience (agent_id, task_type, input_hash, outcome, metadata)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (agent_id, task_type, input_hash)
		DO UPDATE SET
			outcome = EXCLUDED.outcome,
			learned_at = NOW(),
			metadata = EXCLUDED.metadata
	`

	_, err := e.db.Pool.Exec(ctx, query, agentID, taskType, inputHash, outcome, metadata)
	if err != nil {
		return fmt.Errorf("failed to record agent outcome: %w", err)
	}

	return nil
}

// GetLearnedBehavior retrieves past experience for a specific task
func (e *ExperienceStore) GetLearnedBehavior(ctx context.Context, agentID, taskType, inputHash string) (*Experience, error) {
	const query = `
		SELECT id, agent_id, task_type, input_hash, outcome, learned_at, metadata
		FROM agent_experience
		WHERE agent_id = $1 AND task_type = $2 AND input_hash = $3
	`

	var exp Experience
	err := e.db.Pool.QueryRow(ctx, query, agentID, taskType, inputHash).Scan(
		&exp.ID,
		&exp.AgentID,
		&exp.TaskType,
		&exp.InputHash,
		&exp.Outcome,
		&exp.LearnedAt,
		&exp.Metadata,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // No prior experience
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query learned behavior: %w", err)
	}

	return &exp, nil
}

// ShouldRetry determines if a task should be retried based on past outcomes
func (e *ExperienceStore) ShouldRetry(ctx context.Context, agentID, taskType, inputHash string) (bool, error) {
	exp, err := e.GetLearnedBehavior(ctx, agentID, taskType, inputHash)
	if err != nil {
		return false, err
	}

	// No prior experience - allow retry
	if exp == nil {
		return true, nil
	}

	// Don't retry if previous attempt was a timeout (likely systemic issue)
	if exp.Outcome == "timeout" {
		return false, nil
	}

	// Allow retry for failures (might be transient)
	if exp.Outcome == "failure" {
		return true, nil
	}

	// Don't retry successful tasks (idempotency)
	return false, nil
}

// GetFailureRate returns the failure rate for an agent across all task types
func (e *ExperienceStore) GetFailureRate(ctx context.Context, agentID string) (float64, int64, error) {
	const query = `
		SELECT
			COUNT(*) FILTER (WHERE outcome = 'failure') AS failures,
			COUNT(*) AS total
		FROM agent_experience
		WHERE agent_id = $1
	`

	var failures, total int64
	err := e.db.Pool.QueryRow(ctx, query, agentID).Scan(&failures, &total)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to calculate failure rate: %w", err)
	}

	if total == 0 {
		return 0, 0, nil
	}

	rate := float64(failures) / float64(total)
	return rate, total, nil
}

// PruneOldExperiences removes experiences older than the specified days
func (e *ExperienceStore) PruneOldExperiences(ctx context.Context, days int) (int64, error) {
	const query = `
		DELETE FROM agent_experience
		WHERE learned_at < NOW() - INTERVAL '1 day' * $1
	`

	result, err := e.db.Pool.Exec(ctx, query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to prune old experiences: %w", err)
	}

	return result.RowsAffected(), nil
}
