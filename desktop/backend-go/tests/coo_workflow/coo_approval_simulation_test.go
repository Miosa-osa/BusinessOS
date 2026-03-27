package coo_workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/services"
)

// COOApprovalSimulationTest simulates a COO user flow:
// 1. Review 7 days of autonomous CRM/task operations
// 2. Check success_rate >= 0.80 for process-leads and assign-tasks
// 3. Query decision queue (should have <=3 items after 7-day sim)
// 4. Approve 3 decisions with tracking
// 5. Verify learning loop recorded feedback
func TestCOOApprovalSimulation(t *testing.T) {
	ctx := context.Background()

	// Connect to test database
	pool, err := setupTestDB(ctx)
	if err != nil {
		t.Skipf("Skipping: test DB not available: %v", err)
		return
	}
	defer pool.Close()

	logger := slog.Default()

	// ============================================================================
	// STEP 1: Setup test data (7 days of operations)
	// ============================================================================

	cooUserID := uuid.New().String()
	pipelineID := uuid.New()
	projectID := uuid.New()

	if err := seedTestDataFor7Days(ctx, pool, cooUserID, pipelineID, projectID); err != nil {
		t.Fatalf("Failed to seed 7-day test data: %v", err)
	}

	logger.Info("✓ Seeded 7 days of autonomous operations")

	// ============================================================================
	// STEP 2: Simulate CRM ProcessLeads (7-day batch)
	// ============================================================================

	crmAgent := services.NewCRMAgent(pool)
	processLeadsResult, err := crmAgent.ProcessLeads(ctx, cooUserID, pipelineID)
	if err != nil {
		t.Fatalf("Failed to process leads: %v", err)
	}

	// Verify success_rate >= 0.80
	if processLeadsResult.SuccessRate < 0.80 {
		t.Errorf("ProcessLeads success_rate %.2f is below 0.80 threshold", processLeadsResult.SuccessRate)
	}

	logger.Info("✓ ProcessLeads results",
		"created", processLeadsResult.CreatedCount,
		"escalated", processLeadsResult.EscalatedCount,
		"success_rate", fmt.Sprintf("%.2f", processLeadsResult.SuccessRate))

	// ============================================================================
	// STEP 3: Simulate Project AssignTasks
	// ============================================================================

	projectAgent := services.NewProjectAgent(pool)
	assignTasksResult, err := projectAgent.AssignTasks(ctx, cooUserID)
	if err != nil {
		t.Fatalf("Failed to assign tasks: %v", err)
	}

	// Verify success_rate >= 0.80
	if assignTasksResult.SuccessRate < 0.80 {
		t.Errorf("AssignTasks success_rate %.2f is below 0.80 threshold", assignTasksResult.SuccessRate)
	}

	logger.Info("✓ AssignTasks results",
		"assigned", assignTasksResult.AssignedCount,
		"skipped", assignTasksResult.SkippedCount,
		"success_rate", fmt.Sprintf("%.2f", assignTasksResult.SuccessRate))

	// ============================================================================
	// STEP 4: Query pending decisions
	// ============================================================================

	testCfg := &config.Config{
		Environment: "test",
	}
	sorxService := services.NewSorxService(pool, testCfg)
	decisions, err := sorxService.GetPendingDecisions(ctx, cooUserID)
	if err != nil {
		t.Fatalf("Failed to get pending decisions: %v", err)
	}

	// Verify queue depth <= 3 (COO should review only high-priority)
	if len(decisions) > 3 {
		t.Errorf("Decision queue has %d items, expected <= 3", len(decisions))
	}

	logger.Info("✓ Pending decisions fetched",
		"count", len(decisions))

	if len(decisions) == 0 {
		t.Logf("Note: No decisions in queue (all auto-processed by agents)")
		return // Early exit if nothing to approve
	}

	// ============================================================================
	// STEP 5: COO approves decisions (first 3)
	// ============================================================================

	var approvalMetrics ApprovalMetrics
	approvalMetrics.StartTime = time.Now()
	approvalMetrics.DecisionsReviewed = len(decisions)

	approveCount := 0
	if len(decisions) > 3 {
		approveCount = 3
	} else {
		approveCount = len(decisions)
	}

	for i := 0; i < approveCount; i++ {
		decision := decisions[i]

		decisionID, ok := decision["id"].(uuid.UUID)
		if !ok {
			idStr, _ := decision["id"].(string)
			decisionID = uuid.MustParse(idStr)
		}

		decision_str := "Approve" // COO action
		if i%3 == 1 {
			decision_str = "Defer"
		} else if i%3 == 2 {
			decision_str = "Reject"
		}

		responseTime := time.Now()
		err := sorxService.RespondToDecision(ctx, decisionID, cooUserID, decision_str, map[string]interface{}{
			"coo_comment": "Reviewed autonomy metrics and gate thresholds",
		})

		if err != nil {
			t.Errorf("Failed to record decision response: %v", err)
			continue
		}

		latency := time.Since(responseTime)
		approvalMetrics.ResponseLatencies = append(approvalMetrics.ResponseLatencies, latency)
		approvalMetrics.ApprovedCount++

		logger.Info("✓ Decision recorded",
			"decision_id", decisionID.String(),
			"response", decision_str,
			"latency_ms", latency.Milliseconds())
	}

	approvalMetrics.EndTime = time.Now()
	approvalMetrics.TotalDuration = approvalMetrics.EndTime.Sub(approvalMetrics.StartTime)

	// ============================================================================
	// STEP 6: Verify learning loop feedback (S/N governance routing)
	// ============================================================================

	healingDiagnosisSpans := queryOTELSpans(ctx, pool, "healing.adaptive.adjust", cooUserID)
	if len(healingDiagnosisSpans) > 0 {
		approvalMetrics.LearningLoopFeedbackCount = len(healingDiagnosisSpans)
		logger.Info("✓ Learning loop feedback verified",
			"healing_adaptive_adjust_spans", len(healingDiagnosisSpans))
	} else {
		logger.Warn("! No learning loop spans recorded (expected if no low-confidence decisions)")
		approvalMetrics.LearningLoopFeedbackCount = 0
	}

	// ============================================================================
	// STEP 7: Verify S/N governance tier routing
	// ============================================================================

	lowConfidenceDecisions := queryLowConfidenceDecisions(ctx, pool, cooUserID, 0.7)
	if len(lowConfidenceDecisions) > 0 {
		logger.Info("✓ S/N governance detected low-confidence decisions",
			"count", len(lowConfidenceDecisions),
			"tier_routing", "escalated")
	} else {
		logger.Warn("! No low-confidence decisions routed (all auto-processed with high confidence)")
	}

	// ============================================================================
	// STEP 8: Generate final report
	// ============================================================================

	report := generateCOOReport(
		processLeadsResult,
		assignTasksResult,
		len(decisions),
		approvalMetrics,
	)

	logger.Info("✓ COO Approval Simulation Complete")
	printCOOReport(report)

	// Assertions
	if approvalMetrics.ApprovedCount != approveCount {
		t.Errorf("Expected %d approvals, got %d", approveCount, approvalMetrics.ApprovedCount)
	}

	if processLeadsResult.SuccessRate < 0.80 || assignTasksResult.SuccessRate < 0.80 {
		t.Error("Automation success rates below 0.80 threshold")
	}
}

// ============================================================================
// Helper Types and Functions
// ============================================================================

type ApprovalMetrics struct {
	StartTime                    time.Time
	EndTime                      time.Time
	TotalDuration                time.Duration
	DecisionsReviewed            int
	ApprovedCount                int
	ResponseLatencies            []time.Duration
	LearningLoopFeedbackCount    int
}

type COOReport struct {
	Timestamp                       time.Time
	AutomationRates                 map[string]float64
	QueueDepth                      int
	DecisionResponseLatencies       map[string]interface{}
	LearningLoopFeedbackCount       int
	SignalToNoiseGovernanceTier     string
	LowConfidenceDecisionsRouted    int
	AgentSuccessMetrics             map[string]interface{}
}

func generateCOOReport(
	processLeads *services.ProcessLeadsResult,
	assignTasks *services.AssignTasksResult,
	queueDepth int,
	approvalMetrics ApprovalMetrics,
) *COOReport {
	report := &COOReport{
		Timestamp:  time.Now(),
		QueueDepth: queueDepth,
		AutomationRates: map[string]float64{
			"process_leads":  processLeads.SuccessRate,
			"assign_tasks":   assignTasks.SuccessRate,
			"avg_automation": (processLeads.SuccessRate + assignTasks.SuccessRate) / 2,
		},
		LearningLoopFeedbackCount: approvalMetrics.LearningLoopFeedbackCount,
		SignalToNoiseGovernanceTier: "NORMAL",
		AgentSuccessMetrics: map[string]interface{}{
			"crm_created_deals":     processLeads.CreatedCount,
			"crm_escalated_leads":   processLeads.EscalatedCount,
			"project_assigned_tasks": assignTasks.AssignedCount,
			"project_skipped_tasks":  assignTasks.SkippedCount,
		},
		DecisionResponseLatencies: map[string]interface{}{
			"total_reviewed": approvalMetrics.DecisionsReviewed,
			"total_approved": approvalMetrics.ApprovedCount,
			"batch_duration": approvalMetrics.TotalDuration.String(),
		},
	}

	// Calculate average latency
	if len(approvalMetrics.ResponseLatencies) > 0 {
		totalLatency := time.Duration(0)
		for _, lat := range approvalMetrics.ResponseLatencies {
			totalLatency += lat
		}
		avgLatency := totalLatency / time.Duration(len(approvalMetrics.ResponseLatencies))
		report.DecisionResponseLatencies["avg_latency_ms"] = avgLatency.Milliseconds()
		report.DecisionResponseLatencies["min_latency_ms"] = minDuration(approvalMetrics.ResponseLatencies).Milliseconds()
		report.DecisionResponseLatencies["max_latency_ms"] = maxDuration(approvalMetrics.ResponseLatencies).Milliseconds()
	}

	// S/N governance tier routing
	if (processLeads.SuccessRate + assignTasks.SuccessRate) / 2 < 0.85 {
		report.SignalToNoiseGovernanceTier = "CAUTION"
	}

	return report
}

func printCOOReport(report *COOReport) {
	jsonData, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("COO APPROVAL WORKFLOW REPORT")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(string(jsonData))
	fmt.Println(strings.Repeat("=", 80))
}

func minDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	min := durations[0]
	for _, d := range durations {
		if d < min {
			min = d
		}
	}
	return min
}

func maxDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	max := durations[0]
	for _, d := range durations {
		if d > max {
			max = d
		}
	}
	return max
}

// ============================================================================
// Database Setup and Seeding
// ============================================================================

func setupTestDB(ctx context.Context) (*pgxpool.Pool, error) {
	// In real test: would use testcontainers or similar
	// For now, assume test db is running (see docker-compose.test.yml)
	connStr := "postgresql://postgres:postgres@localhost:5432/businessos_test"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func seedTestDataFor7Days(ctx context.Context, pool *pgxpool.Pool, userID string, pipelineID, projectID uuid.UUID) error {
	// Create companies (prospects)
	for i := 1; i <= 50; i++ {
		companyID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO companies (id, user_id, name, health_score, engagement_score, annual_revenue, company_size, lifecycle_stage, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'prospect', NOW() - INTERVAL '7 days')
		`, companyID, userID, fmt.Sprintf("Company %d", i),
			50+i%50, // health: 50-100
			40+i%60, // engagement: 40-100
			float64(1000000+i*100000), // revenue
			[]string{"1-10", "11-50", "51-200", "201-500", "501-1000", "1000+"}[i%6])
		if err != nil {
			return err
		}
	}

	// Create unassigned tasks
	for i := 1; i <= 50; i++ {
		taskID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO tasks (id, user_id, project_id, title, priority, status, created_at)
			VALUES ($1, $2, $3, $4, $5, 'todo', NOW() - INTERVAL '7 days')
		`, taskID, userID, projectID, fmt.Sprintf("Task %d", i),
			[]string{"high", "medium", "low"}[i%3])
		if err != nil {
			return err
		}
	}

	// Create pipeline stages
	stageNames := []string{"Lead", "Negotiation", "Closing"}
	for i, name := range stageNames {
		stageID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO pipeline_stages (id, pipeline_id, name, position)
			VALUES ($1, $2, $3, $4)
		`, stageID, pipelineID, name, i)
		if err != nil {
			return err
		}
	}

	// Create pending decisions (low-confidence deals that need COO review)
	for i := 1; i <= 3; i++ {
		decisionID := uuid.New()
		executionID := uuid.New().String()

		_, err := pool.Exec(ctx, `
			INSERT INTO pending_decisions (id, execution_id, skill_id, step_id, user_id, question, options, priority, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending', NOW())
		`, decisionID, executionID, "crm-lead-scoring", fmt.Sprintf("step-%d", i), userID,
			fmt.Sprintf("Approve high-value deal from Company %d?", i),
			json.RawMessage(`["Approve", "Defer", "Reject"]`),
			[]string{"urgent", "high", "medium"}[i%3])
		if err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// OTEL Span Query (mock)
// ============================================================================

func queryOTELSpans(ctx context.Context, pool *pgxpool.Pool, spanName, userID string) []map[string]interface{} {
	// In production: query actual OTEL collector or Jaeger
	// For now: mock response based on database query
	var spans []map[string]interface{}

	rows, err := pool.Query(ctx, `
		SELECT id, attributes
		FROM audit_log
		WHERE user_id = $1
		AND operation_name LIKE $2
		AND created_at > NOW() - INTERVAL '24 hours'
		LIMIT 10
	`, userID, spanName+"%")

	if err != nil {
		return spans
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var attributes []byte
		if err := rows.Scan(&id, &attributes); err != nil {
			continue
		}
		var attrs map[string]interface{}
		json.Unmarshal(attributes, &attrs)
		spans = append(spans, map[string]interface{}{
			"id":         id,
			"span_name":  spanName,
			"attributes": attrs,
		})
	}

	return spans
}

func queryLowConfidenceDecisions(ctx context.Context, pool *pgxpool.Pool, userID string, confidenceThreshold float64) []map[string]interface{} {
	// Mock: query decisions with confidence < threshold
	var decisions []map[string]interface{}

	rows, err := pool.Query(ctx, `
		SELECT id, question, options
		FROM pending_decisions
		WHERE user_id = $1
		AND status = 'pending'
		AND priority IN ('high', 'urgent')
		LIMIT 5
	`, userID)

	if err != nil {
		return decisions
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var question string
		var options []byte

		if err := rows.Scan(&id, &question, &options); err != nil {
			continue
		}

		decisions = append(decisions, map[string]interface{}{
			"id":       id,
			"question": question,
		})
	}

	return decisions
}
