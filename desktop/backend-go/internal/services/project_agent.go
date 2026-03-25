package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectAgent handles autonomous project operations: task assignment, progress tracking, reporting
type ProjectAgent struct {
	db *pgxpool.Pool
}

// AssignTasksResult represents task assignment outcome
type AssignTasksResult struct {
	AssignedCount int     `json:"assigned_count"`
	SkippedCount  int     `json:"skipped_count"`
	SuccessRate   float64 `json:"success_rate"` // percentage of tasks auto-assigned
}

// ProgressMetrics represents project progress data
type ProgressMetrics struct {
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	CompletionRate  float64 `json:"completion_rate"` // [0, 1]
	OverdueTasks    int     `json:"overdue_tasks"`
	InProgressTasks int     `json:"in_progress_tasks"`
}

// DailyReport represents a daily project status report
type DailyReport struct {
	ProjectID      string `json:"project_id"`
	ProjectName    string `json:"project_name"`
	GeneratedAt    string `json:"generated_at"`
	MetricsSummary string `json:"metrics_summary"`
	KeyAlerts      []string `json:"key_alerts"`
}

// BurndownData represents burndown chart data
type BurndownData struct {
	ProjectID       string             `json:"project_id"`
	DailyProgress   []DailyBurndownPoint `json:"daily_progress"`
	ProjectedDays   int                `json:"projected_days"`
	CompletionDate  string             `json:"completion_date"`
}

// DailyBurndownPoint represents a single day's burndown
type DailyBurndownPoint struct {
	Date              string `json:"date"`
	RemainingTasks    int    `json:"remaining_tasks"`
	CompletedTasks    int    `json:"completed_tasks"`
	CompletionPercent float64 `json:"completion_percent"`
}

// NewProjectAgent creates a new project agent instance
func NewProjectAgent(db *pgxpool.Pool) *ProjectAgent {
	return &ProjectAgent{db: db}
}

// AssignTasks automatically assigns unassigned tasks to team members with lowest utilization
func (a *ProjectAgent) AssignTasks(ctx context.Context, userID string) (*AssignTasksResult, error) {
	logger := slog.With("agent", "project", "operation", "assign_tasks", "user_id", userID)

	// Get team utilization (how many tasks each person is working on)
	utilization, err := a.GetTeamUtilization(ctx, userID)
	if err != nil {
		logger.Error("failed to get team utilization", "error", err)
		return nil, err
	}

	// Query unassigned tasks
	rows, err := a.db.Query(ctx, `
		SELECT id, title, priority, project_id
		FROM tasks
		WHERE user_id = $1
		AND status IN ('todo', 'in_progress')
		AND assignee_id IS NULL
		ORDER BY priority DESC, created_at ASC
		LIMIT 50
	`, userID)
	if err != nil {
		logger.Error("failed to query unassigned tasks", "error", err)
		return nil, err
	}
	defer rows.Close()

	result := &AssignTasksResult{
		AssignedCount: 0,
		SkippedCount:  0,
	}

	type Task struct {
		ID        uuid.UUID
		Title     string
		Priority  string
		ProjectID *uuid.UUID
	}

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Priority, &t.ProjectID); err != nil {
			logger.Error("failed to scan task", "error", err)
			continue
		}
		tasks = append(tasks, t)
	}

	// For each unassigned task, assign to person with lowest utilization
	for _, task := range tasks {
		assigneeID, err := a.findLowestUtilizationTeamMember(utilization, userID)
		if err != nil {
			logger.Error("failed to find assignee", "task_id", task.ID, "error", err)
			result.SkippedCount++
			continue
		}

		_, err = a.db.Exec(ctx, `
			UPDATE tasks SET assignee_id = $1, updated_at = NOW() WHERE id = $2
		`, assigneeID, task.ID)

		if err != nil {
			logger.Error("failed to assign task", "task_id", task.ID, "error", err)
			result.SkippedCount++
			continue
		}

		// Update local utilization
		if util, ok := utilization[assigneeID]; ok {
			utilization[assigneeID] = util + 1
		} else {
			utilization[assigneeID] = 1
		}

		logger.Info("task assigned", "task_id", task.ID, "assignee_id", assigneeID)
		result.AssignedCount++
	}

	// Calculate success rate
	totalProcessed := result.AssignedCount + result.SkippedCount
	if totalProcessed > 0 {
		result.SuccessRate = float64(result.AssignedCount) / float64(totalProcessed)
	}

	logger.Info("tasks assigned", "assigned", result.AssignedCount, "skipped", result.SkippedCount, "success_rate", result.SuccessRate)
	return result, nil
}

// TrackProgress aggregates all task statuses and calculates completion metrics
func (a *ProjectAgent) TrackProgress(ctx context.Context, userID string, projectID string) (*ProgressMetrics, error) {
	logger := slog.With("agent", "project", "operation", "track_progress", "project_id", projectID)

	metrics := &ProgressMetrics{}

	// Query task statuses
	var totalTasks, completedTasks, inProgressTasks, overdueTasks int

	err := a.db.QueryRow(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'done' THEN 1 END) as completed,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress,
			COUNT(CASE WHEN status IN ('todo', 'in_progress') AND due_date < CURRENT_DATE THEN 1 END) as overdue
		FROM tasks
		WHERE user_id = $1
		AND (project_id = $2::uuid OR $2::uuid IS NULL)
	`, userID, projectID).Scan(&totalTasks, &completedTasks, &inProgressTasks, &overdueTasks)

	if err != nil && err != pgx.ErrNoRows {
		logger.Error("failed to calculate metrics", "error", err)
		return nil, err
	}

	metrics.TotalTasks = totalTasks
	metrics.CompletedTasks = completedTasks
	metrics.InProgressTasks = inProgressTasks
	metrics.OverdueTasks = overdueTasks

	if totalTasks > 0 {
		metrics.CompletionRate = float64(completedTasks) / float64(totalTasks)
	}

	logger.Info("progress tracked", "total", totalTasks, "completed", completedTasks, "completion_rate", metrics.CompletionRate)
	return metrics, nil
}

// CheckOverdue finds and escalates tasks that are overdue by >N days
func (a *ProjectAgent) CheckOverdue(ctx context.Context, userID string, overdueDays int) (bool, int, error) {
	logger := slog.With("agent", "project", "operation", "check_overdue", "user_id", userID)

	var overdueCount int
	err := a.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = $1
		AND status IN ('todo', 'in_progress')
		AND due_date < CURRENT_DATE - INTERVAL '1 day' * $2
	`, userID, overdueDays).Scan(&overdueCount)

	if err != nil {
		logger.Error("failed to count overdue tasks", "error", err)
		return false, 0, err
	}

	if overdueCount > 0 {
		logger.Warn("overdue tasks found", "count", overdueCount, "days_overdue", overdueDays)
	}

	return overdueCount > 0, overdueCount, nil
}

// GenerateDailyReport creates a daily status report for a project
func (a *ProjectAgent) GenerateDailyReport(ctx context.Context, userID string, projectID string) (*DailyReport, error) {
	logger := slog.With("agent", "project", "operation", "generate_report", "project_id", projectID)

	report := &DailyReport{
		ProjectID:   projectID,
		GeneratedAt: time.Now().Format(time.RFC3339),
		KeyAlerts:   []string{},
	}

	// Get project name
	err := a.db.QueryRow(ctx, `
		SELECT name FROM projects WHERE id = $1::uuid
	`, projectID).Scan(&report.ProjectName)

	if err != nil {
		logger.Error("failed to get project", "error", err)
		return nil, err
	}

	// Get metrics
	metrics, err := a.TrackProgress(ctx, userID, projectID)
	if err != nil {
		logger.Error("failed to get metrics", "error", err)
		return nil, err
	}

	// Build metrics summary
	report.MetricsSummary = fmt.Sprintf(
		"Total: %d | Completed: %d (%.1f%%) | In Progress: %d | Overdue: %d",
		metrics.TotalTasks, metrics.CompletedTasks, metrics.CompletionRate*100,
		metrics.InProgressTasks, metrics.OverdueTasks,
	)

	// Add alerts
	if metrics.OverdueTasks > 0 {
		report.KeyAlerts = append(report.KeyAlerts, fmt.Sprintf("%d tasks overdue", metrics.OverdueTasks))
	}

	if metrics.CompletionRate < 0.25 {
		report.KeyAlerts = append(report.KeyAlerts, "Project behind schedule")
	}

	logger.Info("daily report generated", "project_name", report.ProjectName, "alerts", len(report.KeyAlerts))
	return report, nil
}

// CalculateBurndown generates burndown chart data for the past N days
func (a *ProjectAgent) CalculateBurndown(ctx context.Context, userID string, projectID string, days int) (*BurndownData, error) {
	logger := slog.With("agent", "project", "operation", "calculate_burndown", "project_id", projectID, "days", days)

	burndown := &BurndownData{
		ProjectID:     projectID,
		DailyProgress: []DailyBurndownPoint{},
	}

	// Query daily progress for the past N days
	rows, err := a.db.Query(ctx, `
		SELECT
			DATE(CURRENT_DATE - INTERVAL '1 day' * (g::integer)) as date,
			COUNT(CASE WHEN status = 'done' AND completed_at::date <= (CURRENT_DATE - INTERVAL '1 day' * (g::integer)) THEN 1 END) as completed,
			COUNT(*) - COUNT(CASE WHEN status = 'done' AND completed_at::date <= (CURRENT_DATE - INTERVAL '1 day' * (g::integer)) THEN 1 END) as remaining
		FROM tasks
		CROSS JOIN LATERAL (SELECT * FROM generate_series(0, $3)) g(integer)
		WHERE user_id = $1 AND project_id = $2::uuid
		GROUP BY date
		ORDER BY date ASC
	`, userID, projectID, days)

	if err != nil {
		logger.Error("failed to query burndown data", "error", err)
		return nil, err
	}
	defer rows.Close()

	totalTasks := 0
	for rows.Next() {
		var dateStr string
		var completed, remaining int

		if err := rows.Scan(&dateStr, &completed, &remaining); err != nil {
			logger.Error("failed to scan burndown row", "error", err)
			continue
		}

		total := completed + remaining
		if total > totalTasks {
			totalTasks = total
		}

		var completionPct float64
		if total > 0 {
			completionPct = float64(completed) / float64(total) * 100
		}

		point := DailyBurndownPoint{
			Date:              dateStr,
			RemainingTasks:    remaining,
			CompletedTasks:    completed,
			CompletionPercent: completionPct,
		}

		burndown.DailyProgress = append(burndown.DailyProgress, point)
	}

	// Estimate completion date (if remaining tasks > 0)
	if len(burndown.DailyProgress) > 1 {
		lastPoint := burndown.DailyProgress[len(burndown.DailyProgress)-1]
		if lastPoint.RemainingTasks > 0 {
			// Simple linear projection
			if len(burndown.DailyProgress) >= 2 {
				prevPoint := burndown.DailyProgress[len(burndown.DailyProgress)-2]
				tasksPerDay := float64(prevPoint.RemainingTasks - lastPoint.RemainingTasks)
				if tasksPerDay > 0 {
					daysRemaining := float64(lastPoint.RemainingTasks) / tasksPerDay
					burndown.ProjectedDays = int(daysRemaining)
					projectedDate := time.Now().AddDate(0, 0, int(daysRemaining))
					burndown.CompletionDate = projectedDate.Format("2006-01-02")
				}
			}
		}
	}

	logger.Info("burndown calculated", "data_points", len(burndown.DailyProgress), "projected_days", burndown.ProjectedDays)
	return burndown, nil
}

// GetTeamUtilization calculates how many tasks each team member is assigned
func (a *ProjectAgent) GetTeamUtilization(ctx context.Context, userID string) (map[string]float64, error) {
	logger := slog.With("agent", "project", "operation", "get_utilization", "user_id", userID)

	utilization := make(map[string]float64)

	rows, err := a.db.Query(ctx, `
		SELECT assignee_id, COUNT(*) as task_count
		FROM tasks
		WHERE user_id = $1
		AND assignee_id IS NOT NULL
		AND status IN ('todo', 'in_progress')
		GROUP BY assignee_id
	`, userID)

	if err != nil && err != pgx.ErrNoRows {
		logger.Error("failed to query utilization", "error", err)
		return utilization, err
	}

	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var assigneeID string
			var taskCount int

			if err := rows.Scan(&assigneeID, &taskCount); err != nil {
				logger.Error("failed to scan utilization row", "error", err)
				continue
			}

			// Normalize to [0, 1] (cap at 10 concurrent tasks = 1.0)
			utilization[assigneeID] = float64(taskCount) / 10.0
			if utilization[assigneeID] > 1.0 {
				utilization[assigneeID] = 1.0
			}
		}
	}

	logger.Info("utilization calculated", "team_members", len(utilization))
	return utilization, nil
}

// ────────────────────────────────────────────────────────────────────────

func (a *ProjectAgent) findLowestUtilizationTeamMember(utilization map[string]float64, userID string) (string, error) {
	// Find the person with lowest utilization
	// If no utilization data, assign to user themselves

	lowestKey := userID
	lowestUtil := 10.0

	for key, util := range utilization {
		if util < lowestUtil {
			lowestUtil = util
			lowestKey = key
		}
	}

	// If no one has been assigned yet, use the user
	if len(utilization) == 0 {
		return userID, nil
	}

	return lowestKey, nil
}
