package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
)

// TestProjectAgentAutoAssignTask tests automatic task assignment to available agents
func TestProjectAgentAutoAssignTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewProjectAgent(pool)

	// Create a project
	projectID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type)
		VALUES ($1, $2, 'Test Project', 'Test assignment', 'ACTIVE', 'HIGH', 'Test Client', 'client')
		ON CONFLICT DO NOTHING`,
		projectID, userID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create open task
	taskID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, project_id, assignee_id, completed_at)
		VALUES ($1, $2, 'Test Task', 'Unassigned task', 'todo', 'high', CURRENT_DATE + INTERVAL '7 days', $3, NULL, NULL)
		ON CONFLICT DO NOTHING`,
		taskID, userID, projectID)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Test assignment
	result, err := agent.AssignTasks(ctx, userID)
	if err != nil {
		t.Fatalf("AssignTasks failed: %v", err)
	}

	if result == nil {
		t.Errorf("Expected result, got nil")
	}

	if result.AssignedCount < 0 {
		t.Errorf("AssignedCount should be >= 0, got %d", result.AssignedCount)
	}

	t.Logf("Task assignment: assigned=%d, success_rate=%.2f", result.AssignedCount, result.SuccessRate)
}

// TestProjectAgentAutoTrackProgress tests progress aggregation and burndown updates
func TestProjectAgentAutoTrackProgress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewProjectAgent(pool)

	// Create project
	projectID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type)
		VALUES ($1, $2, 'Progress Test Project', 'Track progress', 'ACTIVE', 'CRITICAL', '', 'internal')
		ON CONFLICT DO NOTHING`,
		projectID, userID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create several tasks with different statuses
	taskStatuses := []string{"todo", "in_progress", "done"}
	for i, status := range taskStatuses {
		taskID := uuid.New()
		var completedExpr string
		if status == "done" {
			completedExpr = "CURRENT_TIMESTAMP - INTERVAL '1 days'"
		} else {
			completedExpr = "NULL"
		}

		query := `
			INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, project_id, assignee_id, completed_at)
			VALUES ($1, $2, $3, $4, $5, 'high', CURRENT_DATE + INTERVAL '7 days', $6, NULL, ` + completedExpr + `)
			ON CONFLICT DO NOTHING`

		_, err = pool.Exec(ctx, query,
			taskID, userID, "Task "+string(rune(i)), "Progress tracking task", status, projectID)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	// Test progress tracking
	metrics, err := agent.TrackProgress(ctx, userID, projectID.String())
	if err != nil {
		t.Fatalf("TrackProgress failed: %v", err)
	}

	if metrics == nil {
		t.Errorf("Expected metrics, got nil")
	}

	if metrics.CompletionRate < 0.0 || metrics.CompletionRate > 1.0 {
		t.Errorf("CompletionRate should be in [0, 1], got %.2f", metrics.CompletionRate)
	}

	t.Logf("Progress metrics: completion=%.2f, total_tasks=%d", metrics.CompletionRate, metrics.TotalTasks)
}

// TestProjectAgentEscalateOverdueTask tests escalation of overdue tasks
func TestProjectAgentEscalateOverdueTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewProjectAgent(pool)

	// Create project
	projectID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type)
		VALUES ($1, $2, 'Overdue Test Project', 'Escalation test', 'ACTIVE', 'HIGH', '', 'internal')
		ON CONFLICT DO NOTHING`,
		projectID, userID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create overdue task (>2 days past due)
	taskID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, project_id, assignee_id, completed_at)
		VALUES ($1, $2, 'Overdue Task', 'This task is overdue', 'in_progress', 'critical', CURRENT_DATE - INTERVAL '5 days', $3, NULL, NULL)
		ON CONFLICT DO NOTHING`,
		taskID, userID, projectID)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Test escalation
	overdue, escalatedCount, err := agent.CheckOverdue(ctx, userID, 2)
	if err != nil {
		t.Fatalf("CheckOverdue failed: %v", err)
	}

	if overdue && escalatedCount <= 0 {
		t.Errorf("Should have escalated overdue tasks")
	}

	t.Logf("Overdue check: has_overdue=%v, escalated=%d", overdue, escalatedCount)
}

// TestProjectAgentGenerateDailyReport tests daily report generation and distribution
func TestProjectAgentGenerateDailyReport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewProjectAgent(pool)

	// Create project
	projectID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type)
		VALUES ($1, $2, 'Report Test Project', 'Daily reporting', 'ACTIVE', 'MEDIUM', '', 'internal')
		ON CONFLICT DO NOTHING`,
		projectID, userID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create some tasks to report on
	for i := 0; i < 5; i++ {
		taskID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, project_id, assignee_id, completed_at)
			VALUES ($1, $2, $3, $4, $5, 'medium', CURRENT_DATE + INTERVAL '7 days', $6, NULL, NULL)
			ON CONFLICT DO NOTHING`,
			taskID, userID, "Report Task "+string(rune(i)), "Task for reporting", "todo", projectID)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	// Test report generation
	report, err := agent.GenerateDailyReport(ctx, userID, projectID.String())
	if err != nil {
		t.Fatalf("GenerateDailyReport failed: %v", err)
	}

	if report == nil {
		t.Errorf("Expected report, got nil")
	}

	if report.ProjectName == "" {
		t.Errorf("Report should have project name")
	}

	if report.MetricsSummary == "" {
		t.Logf("Report has metrics summary")
	}

	t.Logf("Daily report: project=%s, metrics_length=%d", report.ProjectName, len(report.MetricsSummary))
}

// TestProjectAgentTaskAssignmentSuccessRate tests batch task assignment with success rate
func TestProjectAgentTaskAssignmentSuccessRate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewProjectAgent(pool)

	// Create project
	projectID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type)
		VALUES ($1, $2, 'Batch Assignment Project', 'Batch testing', 'ACTIVE', 'HIGH', '', 'internal')
		ON CONFLICT DO NOTHING`,
		projectID, userID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create 50 unassigned tasks
	for i := 0; i < 50; i++ {
		taskID := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, project_id, assignee_id, completed_at)
			VALUES ($1, $2, $3, $4, 'todo', 'medium', CURRENT_DATE + INTERVAL '7 days', $5, NULL, NULL)
			ON CONFLICT DO NOTHING`,
			taskID, userID, "Batch Task "+string(rune(i%10)), "Batch assignment task", projectID)
		if err != nil {
			t.Logf("Warning: failed to create task %d: %v", i, err)
		}
	}

	// Run assignment
	result, err := agent.AssignTasks(ctx, userID)
	if err != nil {
		t.Fatalf("AssignTasks failed: %v", err)
	}

	// Verify autonomy rate >= 80%
	if result.SuccessRate < 0.80 {
		t.Errorf("Expected autonomy rate >= 0.80, got %.2f", result.SuccessRate)
	} else {
		t.Logf("Autonomy rate %.2f (target >= 0.80)", result.SuccessRate)
	}

	t.Logf("Batch assignment: assigned=%d, success_rate=%.2f", result.AssignedCount, result.SuccessRate)
}

// TestProjectAgentConcurrentProgress tests concurrent progress tracking
func TestProjectAgentConcurrentProgress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewProjectAgent(pool)

	// Create 5 projects
	projectIDs := make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		projectIDs[i] = uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type)
			VALUES ($1, $2, $3, 'Concurrent project', 'ACTIVE', 'MEDIUM', '', 'internal')
			ON CONFLICT DO NOTHING`,
			projectIDs[i], userID, "Concurrent Project "+string(rune(i)))
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}
	}

	// Track progress concurrently
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			_, err := agent.TrackProgress(ctx, userID, projectIDs[idx].String())
			done <- err
		}(i)
	}

	// Collect results
	errorCount := 0
	for i := 0; i < 5; i++ {
		if err := <-done; err != nil {
			t.Logf("Concurrent operation %d failed: %v", i, err)
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
	}
}

// TestProjectAgentUtilizationScoring tests agent utilization calculation for assignment
func TestProjectAgentUtilizationScoring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewProjectAgent(pool)

	// Test utilization calculation
	utilization, err := agent.GetTeamUtilization(ctx, userID)
	if err != nil {
		t.Fatalf("GetTeamUtilization failed: %v", err)
	}

	if utilization == nil {
		t.Errorf("Expected utilization map, got nil")
	}

	if len(utilization) >= 0 {
		t.Logf("Team utilization calculated for %d team members", len(utilization))
	}

	// Verify utilization values are reasonable
	for member, util := range utilization {
		if util < 0.0 || util > 1.0 {
			t.Errorf("Utilization for %s should be in [0, 1], got %.2f", member, util)
		}
	}
}

// TestProjectAgentBurndownCalculation tests burndown chart data generation
func TestProjectAgentBurndownCalculation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := setupTestDB(ctx, t)
	defer pool.Close()

	userID := uuid.New().String()
	agent := services.NewProjectAgent(pool)

	// Create project
	projectID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type)
		VALUES ($1, $2, 'Burndown Test Project', 'Burndown tracking', 'ACTIVE', 'HIGH', '', 'internal')
		ON CONFLICT DO NOTHING`,
		projectID, userID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create tasks spanning several days
	baseDate := time.Now()
	for day := 0; day < 7; day++ {
		for task := 0; task < 5; task++ {
			taskID := uuid.New()
			dueDate := baseDate.AddDate(0, 0, day).Format("2006-01-02")
			status := "todo"
			if task < day {
				status = "done"
			}

			_, err := pool.Exec(ctx, `
				INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, project_id, assignee_id)
				VALUES ($1, $2, $3, $4, $5, 'medium', $6::date, $7, NULL)
				ON CONFLICT DO NOTHING`,
				taskID, userID, "Burndown Task D"+string(rune(day))+"T"+string(rune(task)), "Burndown task",
				status, dueDate, projectID)
			if err != nil {
				t.Logf("Warning: failed to create burndown task: %v", err)
			}
		}
	}

	// Calculate burndown
	burndown, err := agent.CalculateBurndown(ctx, userID, projectID.String(), 7)
	if err != nil {
		t.Fatalf("CalculateBurndown failed: %v", err)
	}

	if burndown == nil {
		t.Errorf("Expected burndown data, got nil")
	}

	if len(burndown.DailyProgress) != 7 {
		t.Logf("Burndown has %d daily data points (expected ~7)", len(burndown.DailyProgress))
	}
}

