package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ========== WRITE TOOLS ==========

// CreateTaskTool creates a new task
type CreateTaskTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *CreateTaskTool) Name() string { return "create_task" }
func (t *CreateTaskTool) Description() string {
	return "Create a new task. Use this when the user asks to create, add, or make a new task."
}
func (t *CreateTaskTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "The title of the task",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Detailed description of the task (optional)",
			},
			"project_id": map[string]interface{}{
				"type":        "string",
				"description": "The project UUID to assign the task to (optional)",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"low", "medium", "high", "critical"},
				"description": "Task priority (default: medium)",
			},
			"due_date": map[string]interface{}{
				"type":        "string",
				"description": "Due date in YYYY-MM-DD format (optional)",
			},
		},
		"required": []string{"title"},
	}
}
func (t *CreateTaskTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ProjectID   string `json:"project_id"`
		Priority    string `json:"priority"`
		DueDate     string `json:"due_date"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	if params.Priority == "" {
		params.Priority = "medium"
	}

	// Generate UUID for task
	taskID := uuid.New()

	// Build query dynamically
	query := `INSERT INTO tasks (id, user_id, title, description, status, priority, project_id, due_date, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, 'todo', $5, $6, $7, NOW(), NOW())`

	var projectID interface{} = nil
	if params.ProjectID != "" {
		if parsed, err := uuid.Parse(params.ProjectID); err == nil {
			projectID = parsed
		}
	}

	var dueDate interface{} = nil
	if params.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02", params.DueDate); err == nil {
			dueDate = parsed
		}
	}

	var description interface{} = nil
	if params.Description != "" {
		description = params.Description
	}

	_, err := t.pool.Exec(ctx, query, taskID, t.userID, params.Title, description, params.Priority, projectID, dueDate)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	return fmt.Sprintf("✅ Task created successfully!\n\n**Title:** %s\n**ID:** %s\n**Priority:** %s\n**Status:** todo",
		params.Title, taskID.String(), params.Priority), nil
}

// UpdateTaskTool updates an existing task
type UpdateTaskTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *UpdateTaskTool) Name() string { return "update_task" }
func (t *UpdateTaskTool) Description() string {
	return "Update an existing task's status, priority, or other fields. Use this to mark tasks as done, change priority, etc."
}
func (t *UpdateTaskTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "The UUID of the task to update",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"todo", "in_progress", "done"},
				"description": "New status (optional)",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"low", "medium", "high", "critical"},
				"description": "New priority (optional)",
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "New title (optional)",
			},
		},
		"required": []string{"task_id"},
	}
}
func (t *UpdateTaskTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		TaskID   string `json:"task_id"`
		Status   string `json:"status"`
		Priority string `json:"priority"`
		Title    string `json:"title"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	taskUUID, err := uuid.Parse(params.TaskID)
	if err != nil {
		return "", fmt.Errorf("invalid task ID")
	}

	// Build dynamic UPDATE query
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argNum := 1

	if params.Title != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argNum))
		args = append(args, params.Title)
		argNum++
	}
	if params.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argNum))
		args = append(args, params.Status)
		argNum++
		if params.Status == "done" {
			setClauses = append(setClauses, "completed_at = NOW()")
		}
	}
	if params.Priority != "" {
		setClauses = append(setClauses, fmt.Sprintf("priority = $%d", argNum))
		args = append(args, params.Priority)
		argNum++
	}

	query := fmt.Sprintf(`UPDATE tasks SET %s WHERE id = $%d AND user_id = $%d RETURNING title, status, priority`,
		joinStrings(setClauses, ", "), argNum, argNum+1)
	args = append(args, taskUUID, t.userID)

	var title, status, priority string
	err = t.pool.QueryRow(ctx, query, args...).Scan(&title, &status, &priority)
	if err != nil {
		return "", fmt.Errorf("failed to update task: %w", err)
	}

	statusIcon := "⬜"
	switch status {
	case "done":
		statusIcon = "✅"
	case "in_progress":
		statusIcon = "🔄"
	}

	return fmt.Sprintf("%s Task updated successfully!\n\n**Title:** %s\n**Status:** %s\n**Priority:** %s",
		statusIcon, title, status, priority), nil
}

// CreateNoteTool creates a project note
type CreateNoteTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *CreateNoteTool) Name() string { return "create_note" }
func (t *CreateNoteTool) Description() string {
	return "Create a note for a project. Use this to add notes, comments, or updates to a project."
}
func (t *CreateNoteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": map[string]interface{}{
				"type":        "string",
				"description": "The project UUID to add the note to",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The note content",
			},
		},
		"required": []string{"project_id", "content"},
	}
}
func (t *CreateNoteTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ProjectID string `json:"project_id"`
		Content   string `json:"content"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	projectUUID, err := uuid.Parse(params.ProjectID)
	if err != nil {
		return "", fmt.Errorf("invalid project ID")
	}

	query := `INSERT INTO project_notes (project_id, content, created_at) VALUES ($1, $2, NOW()) RETURNING id`
	var noteID uuid.UUID
	err = t.pool.QueryRow(ctx, query, projectUUID, params.Content).Scan(&noteID)
	if err != nil {
		return "", fmt.Errorf("failed to create note: %w", err)
	}

	return fmt.Sprintf("📝 Note added to project!\n\n**Content:** %s", params.Content), nil
}

// UpdateClientPipelineTool updates client pipeline stage
type UpdateClientPipelineTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *UpdateClientPipelineTool) Name() string { return "update_client_pipeline" }
func (t *UpdateClientPipelineTool) Description() string {
	return "Move a client to a different pipeline stage. Use this to update client status in the sales pipeline."
}
func (t *UpdateClientPipelineTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "The client UUID",
			},
			"stage": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"lead", "prospect", "proposal", "negotiation", "won", "lost"},
				"description": "The new pipeline stage",
			},
		},
		"required": []string{"client_id", "stage"},
	}
}
func (t *UpdateClientPipelineTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ClientID string `json:"client_id"`
		Stage    string `json:"stage"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	clientUUID, err := uuid.Parse(params.ClientID)
	if err != nil {
		return "", fmt.Errorf("invalid client ID")
	}

	query := `UPDATE clients SET status = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3 RETURNING name`
	var clientName string
	err = t.pool.QueryRow(ctx, query, params.Stage, clientUUID, t.userID).Scan(&clientName)
	if err != nil {
		return "", fmt.Errorf("failed to update client: %w", err)
	}

	return fmt.Sprintf("📊 Client pipeline updated!\n\n**Client:** %s\n**New Stage:** %s", clientName, params.Stage), nil
}

// LogClientInteractionTool logs a client interaction
type LogClientInteractionTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *LogClientInteractionTool) Name() string { return "log_client_interaction" }
func (t *LogClientInteractionTool) Description() string {
	return "Log an interaction with a client (meeting, call, email, etc.)"
}
func (t *LogClientInteractionTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "The client UUID",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"meeting", "call", "email", "note"},
				"description": "Type of interaction",
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Summary of the interaction",
			},
		},
		"required": []string{"client_id", "type", "summary"},
	}
}
func (t *LogClientInteractionTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ClientID string `json:"client_id"`
		Type     string `json:"type"`
		Summary  string `json:"summary"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	clientUUID, err := uuid.Parse(params.ClientID)
	if err != nil {
		return "", fmt.Errorf("invalid client ID")
	}

	query := `INSERT INTO client_interactions (client_id, user_id, type, summary, created_at)
	          VALUES ($1, $2, $3, $4, NOW()) RETURNING id`
	var interactionID uuid.UUID
	err = t.pool.QueryRow(ctx, query, clientUUID, t.userID, params.Type, params.Summary).Scan(&interactionID)
	if err != nil {
		return "", fmt.Errorf("failed to log interaction: %w", err)
	}

	typeIcon := "📝"
	switch params.Type {
	case "meeting":
		typeIcon = "🤝"
	case "call":
		typeIcon = "📞"
	case "email":
		typeIcon = "📧"
	}

	return fmt.Sprintf("%s Interaction logged!\n\n**Type:** %s\n**Summary:** %s", typeIcon, params.Type, params.Summary), nil
}

// CreateProjectTool creates a new project
type CreateProjectTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *CreateProjectTool) Name() string { return "create_project" }
func (t *CreateProjectTool) Description() string {
	return "Create a new project with name, description, and optional settings"
}
func (t *CreateProjectTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name":        map[string]interface{}{"type": "string", "description": "Project name"},
			"description": map[string]interface{}{"type": "string", "description": "Project description"},
			"status":      map[string]interface{}{"type": "string", "enum": []string{"ACTIVE", "PAUSED", "COMPLETED", "ARCHIVED"}, "description": "Project status (default: ACTIVE)"},
			"priority":    map[string]interface{}{"type": "string", "enum": []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}, "description": "Project priority (default: MEDIUM)"},
		},
		"required": []string{"name"},
	}
}
func (t *CreateProjectTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}
	if params.Status == "" {
		params.Status = "ACTIVE"
	}
	if params.Priority == "" {
		params.Priority = "MEDIUM"
	}

	// Generate UUID for project
	projectID := uuid.New()
	query := `INSERT INTO projects (id, user_id, name, description, status, priority, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`
	_, err := t.pool.Exec(ctx, query, projectID, t.userID, params.Name, params.Description, params.Status, params.Priority)
	if err != nil {
		return "", fmt.Errorf("failed to create project: %w", err)
	}

	return fmt.Sprintf("✅ Project created!\n\n**Name:** %s\n**ID:** %s\n**Status:** %s\n**Priority:** %s", params.Name, projectID, params.Status, params.Priority), nil
}

// UpdateProjectTool updates an existing project
type UpdateProjectTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *UpdateProjectTool) Name() string { return "update_project" }
func (t *UpdateProjectTool) Description() string {
	return "Update an existing project's name, description, or status"
}
func (t *UpdateProjectTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id":  map[string]interface{}{"type": "string", "description": "Project UUID"},
			"name":        map[string]interface{}{"type": "string", "description": "New project name"},
			"description": map[string]interface{}{"type": "string", "description": "New description"},
			"status":      map[string]interface{}{"type": "string", "enum": []string{"active", "planning", "on_hold", "completed"}},
		},
		"required": []string{"project_id"},
	}
}
func (t *UpdateProjectTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ProjectID   string `json:"project_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	projectUUID, err := uuid.Parse(params.ProjectID)
	if err != nil {
		return "", fmt.Errorf("invalid project ID")
	}

	// Build dynamic update
	updates := []string{}
	args := []interface{}{projectUUID, t.userID}
	argNum := 3

	if params.Name != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argNum))
		args = append(args, params.Name)
		argNum++
	}
	if params.Description != "" {
		updates = append(updates, fmt.Sprintf("description = $%d", argNum))
		args = append(args, params.Description)
		argNum++
	}
	if params.Status != "" {
		updates = append(updates, fmt.Sprintf("status = $%d", argNum))
		args = append(args, params.Status)
		argNum++
	}

	if len(updates) == 0 {
		return "No updates provided", nil
	}

	query := fmt.Sprintf("UPDATE projects SET %s, updated_at = NOW() WHERE id = $1 AND user_id = $2", joinStrings(updates, ", "))
	_, err = t.pool.Exec(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to update project: %w", err)
	}

	return "✅ Project updated successfully", nil
}

// BulkCreateTasksTool creates multiple tasks at once
type BulkCreateTasksTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *BulkCreateTasksTool) Name() string { return "bulk_create_tasks" }
func (t *BulkCreateTasksTool) Description() string {
	return "Create multiple tasks at once for a project"
}
func (t *BulkCreateTasksTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": map[string]interface{}{"type": "string", "description": "Project UUID"},
			"tasks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title":       map[string]interface{}{"type": "string"},
						"description": map[string]interface{}{"type": "string"},
						"priority":    map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "critical"}},
					},
					"required": []string{"title"},
				},
			},
		},
		"required": []string{"tasks"},
	}
}
func (t *BulkCreateTasksTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ProjectID string `json:"project_id"`
		Tasks     []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Priority    string `json:"priority"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	var projectUUID *uuid.UUID
	if params.ProjectID != "" {
		parsed, err := uuid.Parse(params.ProjectID)
		if err == nil {
			projectUUID = &parsed
		}
	}

	created := 0
	for _, task := range params.Tasks {
		priority := task.Priority
		if priority == "" {
			priority = "medium"
		}

		// Generate UUID for each task
		taskID := uuid.New()
		query := `INSERT INTO tasks (id, user_id, project_id, title, description, priority, status, created_at, updated_at)
		          VALUES ($1, $2, $3, $4, $5, $6, 'todo', NOW(), NOW())`
		_, err := t.pool.Exec(ctx, query, taskID, t.userID, projectUUID, task.Title, task.Description, priority)
		if err == nil {
			created++
		}
	}

	return fmt.Sprintf("✅ Created %d/%d tasks", created, len(params.Tasks)), nil
}

// MoveTaskTool moves a task to a different status/column
type MoveTaskTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *MoveTaskTool) Name() string { return "move_task" }
func (t *MoveTaskTool) Description() string {
	return "Move a task to a different status (kanban column)"
}
func (t *MoveTaskTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": map[string]interface{}{"type": "string", "description": "Task UUID"},
			"status":  map[string]interface{}{"type": "string", "enum": []string{"todo", "in_progress", "done", "cancelled"}},
		},
		"required": []string{"task_id", "status"},
	}
}
func (t *MoveTaskTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	taskUUID, err := uuid.Parse(params.TaskID)
	if err != nil {
		return "", fmt.Errorf("invalid task ID")
	}

	query := `UPDATE tasks SET status = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`
	_, err = t.pool.Exec(ctx, query, params.Status, taskUUID, t.userID)
	if err != nil {
		return "", fmt.Errorf("failed to move task: %w", err)
	}

	statusIcon := "📋"
	switch params.Status {
	case "in_progress":
		statusIcon = "🔄"
	case "done":
		statusIcon = "✅"
	case "cancelled":
		statusIcon = "❌"
	}

	return fmt.Sprintf("%s Task moved to **%s**", statusIcon, params.Status), nil
}

// AssignTaskTool assigns a task to a team member
type AssignTaskTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *AssignTaskTool) Name() string { return "assign_task" }
func (t *AssignTaskTool) Description() string {
	return "Assign a task to a team member"
}
func (t *AssignTaskTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id":     map[string]interface{}{"type": "string", "description": "Task UUID"},
			"assignee_id": map[string]interface{}{"type": "string", "description": "Team member UUID to assign"},
		},
		"required": []string{"task_id", "assignee_id"},
	}
}
func (t *AssignTaskTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		TaskID     string `json:"task_id"`
		AssigneeID string `json:"assignee_id"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	taskUUID, err := uuid.Parse(params.TaskID)
	if err != nil {
		return "", fmt.Errorf("invalid task ID")
	}
	assigneeUUID, err := uuid.Parse(params.AssigneeID)
	if err != nil {
		return "", fmt.Errorf("invalid assignee ID")
	}

	query := `UPDATE tasks SET assignee_id = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`
	_, err = t.pool.Exec(ctx, query, assigneeUUID, taskUUID, t.userID)
	if err != nil {
		return "", fmt.Errorf("failed to assign task: %w", err)
	}

	return "👤 Task assigned successfully", nil
}

// CreateClientTool creates a new client
type CreateClientTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *CreateClientTool) Name() string { return "create_client" }
func (t *CreateClientTool) Description() string {
	return "Create a new client in the CRM"
}
func (t *CreateClientTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name":           map[string]interface{}{"type": "string", "description": "Client/company name"},
			"email":          map[string]interface{}{"type": "string", "description": "Primary email"},
			"pipeline_stage": map[string]interface{}{"type": "string", "enum": []string{"lead", "prospect", "proposal", "negotiation", "won", "lost"}},
			"notes":          map[string]interface{}{"type": "string", "description": "Initial notes"},
		},
		"required": []string{"name"},
	}
}
func (t *CreateClientTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Name          string `json:"name"`
		Email         string `json:"email"`
		PipelineStage string `json:"pipeline_stage"`
		Notes         string `json:"notes"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}
	if params.PipelineStage == "" {
		params.PipelineStage = "lead"
	}

	query := `INSERT INTO clients (user_id, name, email, pipeline_stage, notes, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id`
	var clientID uuid.UUID
	err := t.pool.QueryRow(ctx, query, t.userID, params.Name, params.Email, params.PipelineStage, params.Notes).Scan(&clientID)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	return fmt.Sprintf("✅ Client created!\n\n**Name:** %s\n**ID:** %s\n**Stage:** %s", params.Name, clientID, params.PipelineStage), nil
}

// UpdateClientTool updates client information
type UpdateClientTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *UpdateClientTool) Name() string { return "update_client" }
func (t *UpdateClientTool) Description() string {
	return "Update client information"
}
func (t *UpdateClientTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"client_id": map[string]interface{}{"type": "string", "description": "Client UUID"},
			"name":      map[string]interface{}{"type": "string"},
			"email":     map[string]interface{}{"type": "string"},
			"notes":     map[string]interface{}{"type": "string"},
		},
		"required": []string{"client_id"},
	}
}
func (t *UpdateClientTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ClientID string `json:"client_id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Notes    string `json:"notes"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	clientUUID, err := uuid.Parse(params.ClientID)
	if err != nil {
		return "", fmt.Errorf("invalid client ID")
	}

	updates := []string{}
	args := []interface{}{clientUUID, t.userID}
	argNum := 3

	if params.Name != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argNum))
		args = append(args, params.Name)
		argNum++
	}
	if params.Email != "" {
		updates = append(updates, fmt.Sprintf("email = $%d", argNum))
		args = append(args, params.Email)
		argNum++
	}
	if params.Notes != "" {
		updates = append(updates, fmt.Sprintf("notes = $%d", argNum))
		args = append(args, params.Notes)
		argNum++
	}

	if len(updates) == 0 {
		return "No updates provided", nil
	}

	query := fmt.Sprintf("UPDATE clients SET %s, updated_at = NOW() WHERE id = $1 AND user_id = $2", joinStrings(updates, ", "))
	_, err = t.pool.Exec(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to update client: %w", err)
	}

	return "✅ Client updated successfully", nil
}

// LogActivityTool logs an activity to the daily log
type LogActivityTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *LogActivityTool) Name() string { return "log_activity" }
func (t *LogActivityTool) Description() string {
	return "Log an activity or note to the daily log"
}
func (t *LogActivityTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"content": map[string]interface{}{"type": "string", "description": "Activity content/note"},
			"type":    map[string]interface{}{"type": "string", "enum": []string{"note", "task", "meeting", "idea", "decision"}, "description": "Type of activity"},
		},
		"required": []string{"content"},
	}
}
func (t *LogActivityTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Content string `json:"content"`
		Type    string `json:"type"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}
	if params.Type == "" {
		params.Type = "note"
	}

	// Generate UUID for log entry
	logID := uuid.New()

	// daily_logs table requires date field
	query := `INSERT INTO daily_logs (id, user_id, date, content, created_at, updated_at)
	          VALUES ($1, $2, CURRENT_DATE, $3, NOW(), NOW())`
	_, err := t.pool.Exec(ctx, query, logID, t.userID, params.Content)
	if err != nil {
		return "", fmt.Errorf("failed to log activity: %w", err)
	}

	typeIcon := "📝"
	switch params.Type {
	case "task":
		typeIcon = "✅"
	case "meeting":
		typeIcon = "🤝"
	case "idea":
		typeIcon = "💡"
	case "decision":
		typeIcon = "🎯"
	}

	return fmt.Sprintf("%s Activity logged: %s", typeIcon, params.Content), nil
}

// CreateArtifactTool starts a document artifact - content will be captured from chat response
type CreateArtifactTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *CreateArtifactTool) Name() string { return "create_artifact" }
func (t *CreateArtifactTool) Description() string {
	return "Start creating a document artifact. Call this FIRST with type and title, then write the document content in your response. The content you write after calling this tool will automatically be saved to the artifact."
}
func (t *CreateArtifactTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Type of document: proposal, plan, report, sop, framework, document",
				"enum":        []string{"proposal", "plan", "report", "sop", "framework", "document"},
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Title of the document",
			},
		},
		"required": []string{"type", "title"},
	}
}
func (t *CreateArtifactTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Type  string `json:"type"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Title == "" {
		return "", fmt.Errorf("title is required")
	}

	// Return a marker that the handler will use to capture content
	return fmt.Sprintf("ARTIFACT_START::%s::%s::Now write the complete document content below. Everything you write will be saved to the artifact.", params.Type, params.Title), nil
}
