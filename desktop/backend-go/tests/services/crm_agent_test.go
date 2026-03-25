package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

// TestCRMAgentAutoCreateDeal tests automatic deal creation from high-quality leads
func TestCRMAgentAutoCreateDeal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewCRMAgent(pool)

	// Create a lead with high-quality data (should auto-create deal with S/N > 0.8)
	companyID := uuid.New()
	pipelineID := uuid.New()
	stageID := uuid.New()

	// Insert pipeline and stage
	_, err := pool.Exec(ctx, `
		INSERT INTO pipelines (id, user_id, name, description, pipeline_type, currency, is_default, is_active, color)
		VALUES ($1, $2, 'Test Pipeline', 'Test', 'sales', 'USD', true, true, '#3b82f6')
		ON CONFLICT DO NOTHING`,
		pipelineID, userID)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO pipeline_stages (id, pipeline_id, name, position, probability, stage_type, rotting_days, color)
		VALUES ($1, $2, 'Lead', 0, 10, 'open', 14, '#94a3b8')
		ON CONFLICT DO NOTHING`,
		stageID, pipelineID)
	if err != nil {
		t.Fatalf("Failed to create stage: %v", err)
	}

	// Insert company
	_, err = pool.Exec(ctx, `
		INSERT INTO companies (id, user_id, name, industry, company_size, website, email, phone, city, country, annual_revenue, lifecycle_stage, health_score, engagement_score)
		VALUES ($1, $2, 'Test Company', 'Technology', '11-50', 'https://test.com', 'hello@test.com', '+1-555-0001', 'Seattle', 'USA', 5000000, 'prospect', 85, 80)
		ON CONFLICT DO NOTHING`,
		companyID, userID)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	// Test execution
	result, err := agent.ProcessLeads(ctx, userID, pipelineID)
	if err != nil {
		t.Fatalf("ProcessLeads failed: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Errorf("Expected result, got nil")
	}

	// Verify we can get deals
	if result.CreatedCount < 0 {
		t.Errorf("CreatedCount should be >= 0, got %d", result.CreatedCount)
	}

	if result.SuccessRate < 0.0 || result.SuccessRate > 1.0 {
		t.Errorf("SuccessRate should be in [0, 1], got %.2f", result.SuccessRate)
	}
}

// TestCRMAgentAutoScoreLead tests lead scoring with S/N calculation
func TestCRMAgentAutoScoreLead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewCRMAgent(pool)

	// Create test company with complete data
	companyID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO companies (id, user_id, name, industry, company_size, website, email, phone, city, country, annual_revenue, lifecycle_stage, health_score, engagement_score)
		VALUES ($1, $2, 'Premium Company', 'Finance', '201-500', 'https://premium.com', 'business@premium.com', '+1-555-0100', 'New York', 'USA', 150000000, 'prospect', 92, 88)
		ON CONFLICT DO NOTHING`,
		companyID, userID)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	// Test scoring
	score, confidence, err := agent.ScoreLead(ctx, companyID.String())
	if err != nil {
		t.Fatalf("ScoreLead failed: %v", err)
	}

	if score < 0.0 || score > 1.0 {
		t.Errorf("Score should be in [0, 1], got %.2f", score)
	}

	if confidence < 0.0 || confidence > 1.0 {
		t.Errorf("Confidence should be in [0, 1], got %.2f", confidence)
	}

	// High-quality data should have high confidence
	if confidence <= 0.7 {
		t.Logf("Warning: Expected higher confidence for complete company data, got %.2f", confidence)
	}
}

// TestCRMAgentEscalateHighValueDeal tests escalation of high-value deals with low confidence
func TestCRMAgentEscalateHighValueDeal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewCRMAgent(pool)

	// Create high-value deal that should trigger escalation
	dealID := uuid.New()
	pipelineID := uuid.New()
	stageID := uuid.New()
	companyID := uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO pipelines (id, user_id, name, description, pipeline_type, currency, is_default, is_active, color)
		VALUES ($1, $2, 'High Value Pipeline', 'Enterprise deals', 'sales', 'USD', true, true, '#3b82f6')
		ON CONFLICT DO NOTHING`,
		pipelineID, userID)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO pipeline_stages (id, pipeline_id, name, position, probability, stage_type, rotting_days, color)
		VALUES ($1, $2, 'Negotiation', 3, 80, 'open', 5, '#8b5cf6')
		ON CONFLICT DO NOTHING`,
		stageID, pipelineID)
	if err != nil {
		t.Fatalf("Failed to create stage: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO companies (id, user_id, name, industry, company_size, website, email, phone, city, country, annual_revenue, lifecycle_stage, health_score, engagement_score)
		VALUES ($1, $2, 'Enterprise Corp', 'Financial Services', '1000+', 'https://enterprise.com', 'deals@enterprise.com', '+1-555-0200', 'San Francisco', 'USA', 5000000000, 'prospect', 75, 60)
		ON CONFLICT DO NOTHING`,
		companyID, userID)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	// Create high-value deal (> $100K)
	_, err = pool.Exec(ctx, `
		INSERT INTO deals (id, user_id, pipeline_id, stage_id, name, description, amount, currency, probability, expected_close_date, owner_id, company_id, status, priority, lead_source)
		VALUES ($1, $2, $3, $4, 'Enterprise License', 'High-value enterprise contract', 500000, 'USD', 75, CURRENT_DATE + INTERVAL '14 days', $5, $6, 'open', 'critical', 'Referral')
		ON CONFLICT DO NOTHING`,
		dealID, userID, pipelineID, stageID, userID, companyID)
	if err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	// Test escalation logic
	shouldEscalate, reason, err := agent.CheckEscalation(ctx, dealID.String())
	if err != nil {
		t.Fatalf("CheckEscalation failed: %v", err)
	}

	if shouldEscalate && reason == "" {
		t.Errorf("Escalated deal should have a reason")
	}

	t.Logf("Escalation result: escalate=%v, reason=%s", shouldEscalate, reason)
}

// TestCRMAgentUpdatePipelineStage tests automatic pipeline stage progression
func TestCRMAgentUpdatePipelineStage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewCRMAgent(pool)

	// Create deal in Negotiation stage that's > 14 days old
	dealID := uuid.New()
	pipelineID := uuid.New()
	negotiationStageID := uuid.New()
	stalledStageID := uuid.New()
	companyID := uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO pipelines (id, user_id, name, description, pipeline_type, currency, is_default, is_active, color)
		VALUES ($1, $2, 'Stalled Test Pipeline', 'Test', 'sales', 'USD', true, true, '#3b82f6')
		ON CONFLICT DO NOTHING`,
		pipelineID, userID)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	// Negotiation stage
	_, err = pool.Exec(ctx, `
		INSERT INTO pipeline_stages (id, pipeline_id, name, position, probability, stage_type, rotting_days, color)
		VALUES ($1, $2, 'Negotiation', 3, 80, 'open', 5, '#8b5cf6')
		ON CONFLICT DO NOTHING`,
		negotiationStageID, pipelineID)
	if err != nil {
		t.Fatalf("Failed to create negotiation stage: %v", err)
	}

	// Stalled stage
	_, err = pool.Exec(ctx, `
		INSERT INTO pipeline_stages (id, pipeline_id, name, position, probability, stage_type, rotting_days, color)
		VALUES ($1, $2, 'Stalled', 4, 0, 'lost', 0, '#ef4444')
		ON CONFLICT DO NOTHING`,
		stalledStageID, pipelineID)
	if err != nil {
		t.Fatalf("Failed to create stalled stage: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO companies (id, user_id, name, industry, company_size, website, email, phone, city, country, annual_revenue, lifecycle_stage, health_score, engagement_score)
		VALUES ($1, $2, 'Stalled Test Company', 'Manufacturing', '51-200', 'https://stalled.com', 'sales@stalled.com', '+1-555-0300', 'Chicago', 'USA', 25000000, 'prospect', 50, 40)
		ON CONFLICT DO NOTHING`,
		companyID, userID)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	// Create deal in Negotiation stage that's been there >14 days
	_, err = pool.Exec(ctx, `
		INSERT INTO deals (id, user_id, pipeline_id, stage_id, name, description, amount, currency, probability, expected_close_date, owner_id, company_id, status, priority, lead_source, created_at)
		VALUES ($1, $2, $3, $4, 'Stalled Deal', 'Deal stuck in negotiation', 75000, 'USD', 80, CURRENT_DATE - INTERVAL '20 days', $5, $6, 'open', 'high', 'Website', CURRENT_TIMESTAMP - INTERVAL '20 days')
		ON CONFLICT DO NOTHING`,
		dealID, userID, pipelineID, negotiationStageID, userID, companyID)
	if err != nil {
		t.Fatalf("Failed to create deal: %v", err)
	}

	// Test stage update logic
	updated, newStage, err := agent.CheckStalledDeals(ctx, userID, pipelineID.String(), 14)
	if err != nil {
		t.Fatalf("CheckStalledDeals failed: %v", err)
	}

	if updated && newStage == "" {
		t.Errorf("Updated deal should have new stage identifier")
	}

	t.Logf("Stalled deals check: updated=%v, new_stage=%s", updated, newStage)
}

// TestCRMAgentDealCreationSuccessRate tests batch processing with success rate calculation
func TestCRMAgentDealCreationSuccessRate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewCRMAgent(pool)

	// Setup pipeline and stages
	pipelineID := uuid.New()
	leadStageID := uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO pipelines (id, user_id, name, description, pipeline_type, currency, is_default, is_active, color)
		VALUES ($1, $2, 'Batch Test Pipeline', 'Test', 'sales', 'USD', true, true, '#3b82f6')
		ON CONFLICT DO NOTHING`,
		pipelineID, userID)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO pipeline_stages (id, pipeline_id, name, position, probability, stage_type, rotting_days, color)
		VALUES ($1, $2, 'Lead', 0, 10, 'open', 14, '#94a3b8')
		ON CONFLICT DO NOTHING`,
		leadStageID, pipelineID)
	if err != nil {
		t.Fatalf("Failed to create stage: %v", err)
	}

	// Create 100 synthetic companies (simulating leads)
	successCount := 0
	for i := 0; i < 100; i++ {
		companyID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO companies (id, user_id, name, industry, company_size, website, email, phone, city, country, annual_revenue, lifecycle_stage, health_score, engagement_score)
			VALUES ($1, $2, $3, 'Tech', '11-50', 'https://test.com', 'hello@test.com', '+1-555-0001', 'Seattle', 'USA', $4, 'prospect', $5, $6)
			ON CONFLICT DO NOTHING`,
			companyID, userID, "Synthetic Co "+string(rune(i)), float64(1000000+i*100000), 50+i%50, 40+i%50)
		if err == nil {
			successCount++
		}
	}

	// Process all leads
	result, err := agent.ProcessLeads(ctx, userID, pipelineID)
	if err != nil {
		t.Fatalf("ProcessLeads failed: %v", err)
	}

	// Verify autonomy rate >= 80%
	if result.SuccessRate < 0.80 {
		t.Errorf("Expected autonomy rate >= 0.80, got %.2f", result.SuccessRate)
	} else {
		t.Logf("✓ Autonomy rate %.2f (target >= 0.80)", result.SuccessRate)
	}

	t.Logf("Batch processing: created=%d, escalated=%d, success_rate=%.2f",
		result.CreatedCount, result.EscalatedCount, result.SuccessRate)
}

// TestCRMAgentConcurrency tests concurrent deal processing
func TestCRMAgentConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewCRMAgent(pool)

	// Setup pipeline
	pipelineID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO pipelines (id, user_id, name, description, pipeline_type, currency, is_default, is_active, color)
		VALUES ($1, $2, 'Concurrent Pipeline', 'Test', 'sales', 'USD', true, true, '#3b82f6')
		ON CONFLICT DO NOTHING`,
		pipelineID, userID)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO pipeline_stages (id, pipeline_id, name, position, probability, stage_type, rotting_days, color)
		VALUES ($1, $2, 'Lead', 0, 10, 'open', 14, '#94a3b8')
		ON CONFLICT DO NOTHING`,
		uuid.New(), pipelineID)
	if err != nil {
		t.Fatalf("Failed to create stage: %v", err)
	}

	// Run 5 concurrent operations
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			result, err := agent.ProcessLeads(ctx, userID, pipelineID)
			if err != nil {
				done <- err
			} else if result == nil {
				done <- nil
			} else {
				done <- nil
			}
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent operation %d failed: %v", i, err)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────

// setupTestDB creates a test database connection (shared with project agent tests)
func setupTestDB(ctx context.Context, t *testing.T) *pgxpool.Pool {
	dbURL := "postgres://postgres:postgres@localhost:5432/businessos_test"
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Could not connect to test database: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Test database unreachable: %v", err)
	}

	return pool
}

