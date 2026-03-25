package services

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CRMAgent handles autonomous CRM operations: lead scoring, deal creation, pipeline management
type CRMAgent struct {
	db *pgxpool.Pool
}

// ProcessLeadsResult represents the outcome of lead processing
type ProcessLeadsResult struct {
	CreatedCount   int     `json:"created_count"`
	EscalatedCount int     `json:"escalated_count"`
	SkippedCount   int     `json:"skipped_count"`
	SuccessRate    float64 `json:"success_rate"` // percentage of leads that were auto-processed (not escalated)
}

// NewCRMAgent creates a new CRM agent instance
func NewCRMAgent(db *pgxpool.Pool) *CRMAgent {
	return &CRMAgent{db: db}
}

// ProcessLeads processes all open leads and auto-creates deals where S/N > 0.8
// Returns: {created_count, escalated_count, success_rate}
func (a *CRMAgent) ProcessLeads(ctx context.Context, userID string, pipelineID uuid.UUID) (*ProcessLeadsResult, error) {
	logger := slog.With("agent", "crm", "operation", "process_leads", "user_id", userID)

	// Query all leads (companies with no associated deals in this pipeline)
	rows, err := a.db.Query(ctx, `
		SELECT id, name, health_score, engagement_score, annual_revenue, company_size, lifecycle_stage
		FROM companies
		WHERE user_id = $1
		AND lifecycle_stage = 'prospect'
		AND id NOT IN (
			SELECT DISTINCT company_id FROM deals WHERE pipeline_id = $2
		)
		LIMIT 100
	`, userID, pipelineID)
	if err != nil {
		logger.Error("failed to query leads", "error", err)
		return nil, err
	}
	defer rows.Close()

	result := &ProcessLeadsResult{
		CreatedCount:   0,
		EscalatedCount: 0,
		SkippedCount:   0,
	}

	var leads []struct {
		ID              uuid.UUID
		Name            string
		HealthScore     int
		EngagementScore int
		AnnualRevenue   float64
		CompanySize     string
		LifecycleStage  string
	}

	if err := rows.Scan(&leads); err != nil && err != pgx.ErrNoRows {
		// Handle batch scan
		for rows.Next() {
			var id uuid.UUID
			var name string
			var health, engagement int
			var revenue float64
			var size, stage string

			if err := rows.Scan(&id, &name, &health, &engagement, &revenue, &size, &stage); err != nil {
				logger.Error("failed to scan lead", "error", err)
				continue
			}

			leads = append(leads, struct {
				ID              uuid.UUID
				Name            string
				HealthScore     int
				EngagementScore int
				AnnualRevenue   float64
				CompanySize     string
				LifecycleStage  string
			}{id, name, health, engagement, revenue, size, stage})
		}
	} else if err == pgx.ErrNoRows {
		// No leads found
	}

	// Re-query if initial scan failed
	if len(leads) == 0 {
		rows2, _ := a.db.Query(ctx, `
			SELECT id, name, health_score, engagement_score, annual_revenue, company_size, lifecycle_stage
			FROM companies
			WHERE user_id = $1
			AND lifecycle_stage = 'prospect'
			LIMIT 100
		`, userID)
		if rows2 != nil {
			defer rows2.Close()
			for rows2.Next() {
				var id uuid.UUID
				var name string
				var health, engagement int
				var revenue float64
				var size, stage string

				if err := rows2.Scan(&id, &name, &health, &engagement, &revenue, &size, &stage); err != nil {
					continue
				}
				leads = append(leads, struct {
					ID              uuid.UUID
					Name            string
					HealthScore     int
					EngagementScore int
					AnnualRevenue   float64
					CompanySize     string
					LifecycleStage  string
				}{id, name, health, engagement, revenue, size, stage})
			}
		}
	}

	// Get pipeline's first stage for new deals
	var leadStageID uuid.UUID
	err = a.db.QueryRow(ctx, `
		SELECT id FROM pipeline_stages
		WHERE pipeline_id = $1
		ORDER BY position ASC
		LIMIT 1
	`, pipelineID).Scan(&leadStageID)
	if err != nil {
		logger.Error("failed to get lead stage", "error", err)
		return result, err
	}

	// Process each lead
	for _, lead := range leads {
		score, confidence, err := a.scoreLeadInternal(lead.HealthScore, lead.EngagementScore, lead.AnnualRevenue, lead.CompanySize)
		if err != nil {
			logger.Error("failed to score lead", "lead_id", lead.ID, "error", err)
			result.SkippedCount++
			continue
		}

		logger.Info("lead scored", "lead_id", lead.ID, "score", score, "confidence", confidence)

		// Auto-create deal if confidence (S/N) > 0.8
		if confidence > 0.8 {
			dealID := uuid.New()
			dealAmount := calculateDealAmount(lead.AnnualRevenue, score)

			_, err := a.db.Exec(ctx, `
				INSERT INTO deals (id, user_id, pipeline_id, stage_id, name, description, amount, currency, probability, expected_close_date, owner_id, company_id, status, priority, lead_source)
				VALUES ($1, $2, $3, $4, $5, $6, $7, 'USD', $8, CURRENT_DATE + INTERVAL '30 days', $9, $10, 'open', 'medium', 'Auto-Created')
				ON CONFLICT (id) DO NOTHING
			`, dealID, userID, pipelineID, leadStageID, "Deal for "+lead.Name, "Auto-created from lead scoring", dealAmount, int(score*100), userID, lead.ID)

			if err != nil {
				logger.Error("failed to create deal", "lead_id", lead.ID, "error", err)
				result.EscalatedCount++
			} else {
				logger.Info("deal created", "deal_id", dealID, "amount", dealAmount)
				result.CreatedCount++
			}
		} else {
			// Escalate to human if confidence < 0.8
			logger.Info("lead escalated", "lead_id", lead.ID, "confidence", confidence)
			result.EscalatedCount++
		}
	}

	// Calculate success rate (created / (created + escalated))
	totalProcessed := result.CreatedCount + result.EscalatedCount
	if totalProcessed > 0 {
		result.SuccessRate = float64(result.CreatedCount) / float64(totalProcessed)
	}

	logger.Info("leads processed", "created", result.CreatedCount, "escalated", result.EscalatedCount, "success_rate", result.SuccessRate)
	return result, nil
}

// ScoreLead scores a single lead (company) using health, engagement, and revenue signals
// Returns: score [0-1], confidence (S/N) [0-1], error
func (a *CRMAgent) ScoreLead(ctx context.Context, companyID string) (float64, float64, error) {
	var health, engagement int
	var revenue float64
	var size string

	err := a.db.QueryRow(ctx, `
		SELECT health_score, engagement_score, annual_revenue, company_size
		FROM companies
		WHERE id = $1
	`, companyID).Scan(&health, &engagement, &revenue, &size)

	if err != nil {
		return 0, 0, err
	}

	return a.scoreLeadInternal(health, engagement, revenue, size)
}

func (a *CRMAgent) scoreLeadInternal(health, engagement int, revenue float64, size string) (float64, float64, error) {
	// Normalize inputs to [0, 1]
	healthNorm := float64(health) / 100.0
	engagementNorm := float64(engagement) / 100.0
	revenueFactor := math.Min(revenue/50000000, 1.0) // Cap at $50M

	sizeScore := 0.5 // Default
	switch size {
	case "1-10":
		sizeScore = 0.3
	case "11-50":
		sizeScore = 0.5
	case "51-200":
		sizeScore = 0.6
	case "201-500":
		sizeScore = 0.7
	case "501-1000":
		sizeScore = 0.8
	case "1000+":
		sizeScore = 0.9
	}

	// Composite score: weighted average
	score := (healthNorm * 0.3) + (engagementNorm * 0.3) + (revenueFactor * 0.25) + (sizeScore * 0.15)

	// Confidence: how much signal are we getting?
	// High health + engagement + revenue = high confidence
	signalStrength := (healthNorm + engagementNorm + revenueFactor + sizeScore) / 4.0
	confidence := signalStrength * 0.95 // Cap at 0.95

	return score, confidence, nil
}

// CheckEscalation determines if a deal needs human review (S/N < 0.7 or value > $100K)
func (a *CRMAgent) CheckEscalation(ctx context.Context, dealID string) (bool, string, error) {
	var amount float64
	var status string

	err := a.db.QueryRow(ctx, `
		SELECT amount, status FROM deals WHERE id = $1
	`, dealID).Scan(&amount, &status)

	if err != nil {
		return false, "", err
	}

	// High-value deals should be escalated
	if amount > 100000 {
		return true, fmt.Sprintf("High-value deal: $%.0f", amount), nil
	}

	return false, "", nil
}

// CheckStalledDeals detects deals in Negotiation stage for >N days and auto-moves them to Stalled
func (a *CRMAgent) CheckStalledDeals(ctx context.Context, userID string, pipelineID string, rotDays int) (bool, string, error) {
	logger := slog.With("agent", "crm", "operation", "check_stalled", "user_id", userID)

	// Find deals in Negotiation stage older than rotDays
	var dealID uuid.UUID
	var stageName string

	err := a.db.QueryRow(ctx, `
		SELECT d.id, ps.name
		FROM deals d
		JOIN pipeline_stages ps ON d.stage_id = ps.id
		WHERE d.user_id = $1
		AND ps.name = 'Negotiation'
		AND d.created_at < NOW() - INTERVAL '1 day' * $2
		ORDER BY d.created_at ASC
		LIMIT 1
	`, userID, rotDays).Scan(&dealID, &stageName)

	if err != nil {
		return false, "", nil // No stalled deals
	}

	// Find "Stalled" stage in same pipeline
	var stalledStageID uuid.UUID
	err = a.db.QueryRow(ctx, `
		SELECT id FROM pipeline_stages
		WHERE pipeline_id = (SELECT pipeline_id FROM deals WHERE id = $1)
		AND name IN ('Stalled', 'Lost')
		LIMIT 1
	`, dealID).Scan(&stalledStageID)

	if err != nil {
		logger.Error("failed to find stalled stage", "error", err)
		return false, "", err
	}

	// Update deal stage to Stalled
	_, err = a.db.Exec(ctx, `
		UPDATE deals SET stage_id = $1, updated_at = NOW() WHERE id = $2
	`, stalledStageID, dealID)

	if err != nil {
		logger.Error("failed to update deal stage", "error", err)
		return false, "", err
	}

	logger.Info("deal moved to stalled", "deal_id", dealID)
	return true, "Stalled", nil
}

// ────────────────────────────────────────────────────────────────────────

func calculateDealAmount(companyRevenue, score float64) float64 {
	// Base deal amount: 0.5% - 5% of company annual revenue, scaled by score
	basePct := 0.005 + (score * 0.045)
	return companyRevenue * basePct
}
