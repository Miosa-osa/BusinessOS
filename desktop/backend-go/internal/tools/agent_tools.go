package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentTool defines the interface for all agent tools
type AgentTool interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}

// AgentToolRegistry manages available tools for agents
type AgentToolRegistry struct {
	pool             *pgxpool.Pool
	userID           string
	tools            map[string]AgentTool
	embeddingService EmbeddingServiceInterface // Optional, enables semantic search tools
}

// EmbeddingServiceInterface defines the interface for embedding operations
type EmbeddingServiceInterface interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// NewAgentToolRegistry creates a new tool registry for an agent
func NewAgentToolRegistry(pool *pgxpool.Pool, userID string) *AgentToolRegistry {
	registry := &AgentToolRegistry{
		pool:   pool,
		userID: userID,
		tools:  make(map[string]AgentTool),
	}

	// Register all available tools
	registry.registerTools()

	return registry
}

// NewAgentToolRegistryWithEmbedding creates a registry with embedding service for semantic search
func NewAgentToolRegistryWithEmbedding(pool *pgxpool.Pool, userID string, embeddingService EmbeddingServiceInterface) *AgentToolRegistry {
	registry := &AgentToolRegistry{
		pool:             pool,
		userID:           userID,
		tools:            make(map[string]AgentTool),
		embeddingService: embeddingService,
	}

	// Register all available tools including context tools
	registry.registerTools()

	return registry
}

// SetEmbeddingService sets the embedding service to enable semantic search tools
func (r *AgentToolRegistry) SetEmbeddingService(embeddingService EmbeddingServiceInterface) {
	r.embeddingService = embeddingService
	// Re-register tools to include context tools
	r.registerContextTools()
}

func (r *AgentToolRegistry) registerTools() {
	// Read tools
	r.tools["get_project"] = &GetProjectTool{pool: r.pool, userID: r.userID}
	r.tools["get_task"] = &GetTaskTool{pool: r.pool, userID: r.userID}
	r.tools["get_client"] = &GetClientTool{pool: r.pool, userID: r.userID}
	r.tools["list_tasks"] = &ListTasksTool{pool: r.pool, userID: r.userID}
	r.tools["list_projects"] = &ListProjectsTool{pool: r.pool, userID: r.userID}
	r.tools["search_documents"] = &SearchDocumentsTool{pool: r.pool, userID: r.userID}
	r.tools["get_team_capacity"] = &GetTeamCapacityTool{pool: r.pool, userID: r.userID}
	r.tools["query_metrics"] = &QueryMetricsTool{pool: r.pool, userID: r.userID}

	// Write tools
	r.tools["create_task"] = &CreateTaskTool{pool: r.pool, userID: r.userID}
	r.tools["update_task"] = &UpdateTaskTool{pool: r.pool, userID: r.userID}
	r.tools["create_note"] = &CreateNoteTool{pool: r.pool, userID: r.userID}
	r.tools["update_client_pipeline"] = &UpdateClientPipelineTool{pool: r.pool, userID: r.userID}
	r.tools["log_client_interaction"] = &LogClientInteractionTool{pool: r.pool, userID: r.userID}
	r.tools["create_project"] = &CreateProjectTool{pool: r.pool, userID: r.userID}
	r.tools["update_project"] = &UpdateProjectTool{pool: r.pool, userID: r.userID}
	r.tools["bulk_create_tasks"] = &BulkCreateTasksTool{pool: r.pool, userID: r.userID}
	r.tools["move_task"] = &MoveTaskTool{pool: r.pool, userID: r.userID}
	r.tools["assign_task"] = &AssignTaskTool{pool: r.pool, userID: r.userID}
	r.tools["create_client"] = &CreateClientTool{pool: r.pool, userID: r.userID}
	r.tools["update_client"] = &UpdateClientTool{pool: r.pool, userID: r.userID}
	r.tools["log_activity"] = &LogActivityTool{pool: r.pool, userID: r.userID}
	r.tools["create_artifact"] = &CreateArtifactTool{pool: r.pool, userID: r.userID}

	// Dashboard tools
	r.tools["configure_dashboard"] = &ConfigureDashboardTool{pool: r.pool, userID: r.userID}

	// Search tools
	r.tools["web_search"] = &WebSearchTool{pool: r.pool, userID: r.userID}
}

// GetTool returns a tool by name
func (r *AgentToolRegistry) GetTool(name string) (AgentTool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetAllTools returns all registered tools
func (r *AgentToolRegistry) GetAllTools() []AgentTool {
	tools := make([]AgentTool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetToolDefinitions returns tool definitions for LLM
func (r *AgentToolRegistry) GetToolDefinitions() []map[string]interface{} {
	defs := make([]map[string]interface{}, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, map[string]interface{}{
			"name":         tool.Name(),
			"description":  tool.Description(),
			"input_schema": tool.InputSchema(),
		})
	}
	return defs
}

// ExecuteTool executes a tool by name
func (r *AgentToolRegistry) ExecuteTool(ctx context.Context, name string, input json.RawMessage) (string, error) {
	tool, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return tool.Execute(ctx, input)
}

// ========== READ TOOLS ==========

// GetProjectTool retrieves project details
type GetProjectTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *GetProjectTool) Name() string { return "get_project" }
func (t *GetProjectTool) Description() string {
	return "Get detailed information about a specific project including its tasks, notes, and status"
}
func (t *GetProjectTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": map[string]interface{}{
				"type":        "string",
				"description": "The UUID of the project",
			},
		},
		"required": []string{"project_id"},
	}
}
func (t *GetProjectTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	tool := NewGetEntityContextTool(t.pool, t.userID)
	result := tool.Execute(ctx, GetEntityContextInput{
		EntityType: EntityTypeProject,
		EntityID:   params.ProjectID,
	})

	if !result.Success {
		return "", fmt.Errorf("%s", result.Error)
	}
	return result.Content, nil
}

// GetTaskTool retrieves task details
type GetTaskTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *GetTaskTool) Name() string { return "get_task" }
func (t *GetTaskTool) Description() string {
	return "Get detailed information about a specific task"
}
func (t *GetTaskTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "The UUID of the task",
			},
		},
		"required": []string{"task_id"},
	}
}
func (t *GetTaskTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	tool := NewGetEntityContextTool(t.pool, t.userID)
	result := tool.Execute(ctx, GetEntityContextInput{
		EntityType: EntityTypeTask,
		EntityID:   params.TaskID,
	})

	if !result.Success {
		return "", fmt.Errorf("%s", result.Error)
	}
	return result.Content, nil
}

// GetClientTool retrieves client details
type GetClientTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *GetClientTool) Name() string { return "get_client" }
func (t *GetClientTool) Description() string {
	return "Get detailed information about a client including contacts and interaction history"
}
func (t *GetClientTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "The UUID of the client",
			},
		},
		"required": []string{"client_id"},
	}
}
func (t *GetClientTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ClientID string `json:"client_id"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	tool := NewGetEntityContextTool(t.pool, t.userID)
	result := tool.Execute(ctx, GetEntityContextInput{
		EntityType: EntityTypeClient,
		EntityID:   params.ClientID,
	})

	if !result.Success {
		return "", fmt.Errorf("%s", result.Error)
	}
	return result.Content, nil
}

// ListTasksTool lists tasks with filters
type ListTasksTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *ListTasksTool) Name() string { return "list_tasks" }
func (t *ListTasksTool) Description() string {
	return "List tasks with optional filters by project, status, or priority. Use this to see what tasks exist."
}
func (t *ListTasksTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter by project UUID (optional)",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"pending", "in_progress", "done"},
				"description": "Filter by status (optional)",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"low", "medium", "high", "critical"},
				"description": "Filter by priority (optional)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of tasks to return (default 20)",
			},
		},
	}
}
func (t *ListTasksTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ProjectID string `json:"project_id"`
		Status    string `json:"status"`
		Priority  string `json:"priority"`
		Limit     int    `json:"limit"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	query := `SELECT t.id, t.title, t.status::text, COALESCE(t.priority, 'medium'), 
	          COALESCE(p.name, 'No Project') as project_name
	          FROM tasks t
	          LEFT JOIN projects p ON t.project_id = p.id
	          WHERE t.user_id = $1`
	args := []interface{}{t.userID}
	argNum := 2

	if params.ProjectID != "" {
		if projectUUID, err := uuid.Parse(params.ProjectID); err == nil {
			query += fmt.Sprintf(" AND t.project_id = $%d", argNum)
			args = append(args, projectUUID)
			argNum++
		}
	}
	if params.Status != "" {
		query += fmt.Sprintf(" AND t.status = $%d", argNum)
		args = append(args, params.Status)
		argNum++
	}
	if params.Priority != "" {
		query += fmt.Sprintf(" AND t.priority = $%d", argNum)
		args = append(args, params.Priority)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY CASE t.priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END LIMIT $%d", argNum)
	args = append(args, params.Limit)

	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var result string
	result = "## Tasks\n\n"
	count := 0
	for rows.Next() {
		var id uuid.UUID
		var title, status, priority, projectName string
		if err := rows.Scan(&id, &title, &status, &priority, &projectName); err == nil {
			statusIcon := "⬜"
			switch status {
			case "done":
				statusIcon = "✅"
			case "in_progress":
				statusIcon = "🔄"
			}
			result += fmt.Sprintf("- %s **%s** [%s] - %s (ID: %s)\n", statusIcon, title, priority, projectName, id.String())
			count++
		}
	}

	if count == 0 {
		result += "No tasks found matching the criteria.\n"
	}

	return result, nil
}

// ListProjectsTool lists projects
type ListProjectsTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *ListProjectsTool) Name() string { return "list_projects" }
func (t *ListProjectsTool) Description() string {
	return "List all projects with their status. Use this to see available projects."
}
func (t *ListProjectsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"active", "completed", "on_hold", "cancelled"},
				"description": "Filter by status (optional)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of projects to return (default 20)",
			},
		},
	}
}
func (t *ListProjectsTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Status string `json:"status"`
		Limit  int    `json:"limit"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	query := `SELECT id, name, status::text, COALESCE(description, ''), 
	          (SELECT COUNT(*) FROM tasks WHERE project_id = projects.id AND status != 'done') as pending_tasks
	          FROM projects WHERE user_id = $1`
	args := []interface{}{t.userID}

	if params.Status != "" {
		query += " AND status = $2"
		args = append(args, params.Status)
	}

	query += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT %d", params.Limit)

	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var result string
	result = "## Projects\n\n"
	count := 0
	for rows.Next() {
		var id uuid.UUID
		var name, status, description string
		var pendingTasks int
		if err := rows.Scan(&id, &name, &status, &description, &pendingTasks); err == nil {
			result += fmt.Sprintf("- **%s** [%s] - %d pending tasks (ID: %s)\n", name, status, pendingTasks, id.String())
			if description != "" {
				result += fmt.Sprintf("  %s\n", description)
			}
			count++
		}
	}

	if count == 0 {
		result += "No projects found.\n"
	}

	return result, nil
}

// SearchDocumentsTool searches documents/contexts
type SearchDocumentsTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *SearchDocumentsTool) Name() string { return "search_documents" }
func (t *SearchDocumentsTool) Description() string {
	return "Search through documents and contexts by name or content"
}
func (t *SearchDocumentsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query to find in document names or content",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"document", "profile", "knowledge_base"},
				"description": "Filter by document type (optional)",
			},
		},
		"required": []string{"query"},
	}
}
func (t *SearchDocumentsTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Query string `json:"query"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	query := `SELECT id, name, type::text, COALESCE(LEFT(content, 200), '') as preview
	          FROM contexts 
	          WHERE user_id = $1 AND is_archived = false
	          AND (name ILIKE $2 OR content ILIKE $2)`
	args := []interface{}{t.userID, "%" + params.Query + "%"}

	if params.Type != "" {
		query += " AND type = $3"
		args = append(args, params.Type)
	}

	query += " ORDER BY updated_at DESC LIMIT 10"

	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to search documents: %w", err)
	}
	defer rows.Close()

	var result string
	result = fmt.Sprintf("## Search Results for \"%s\"\n\n", params.Query)
	count := 0
	for rows.Next() {
		var id uuid.UUID
		var name, docType, preview string
		if err := rows.Scan(&id, &name, &docType, &preview); err == nil {
			result += fmt.Sprintf("- **%s** [%s] (ID: %s)\n", name, docType, id.String())
			if preview != "" {
				result += fmt.Sprintf("  Preview: %s...\n", preview)
			}
			count++
		}
	}

	if count == 0 {
		result += "No documents found matching the query.\n"
	}

	return result, nil
}

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

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
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

// ========== ADDITIONAL TOOLS ==========

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

// GetTeamCapacityTool gets team workload and capacity
type GetTeamCapacityTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *GetTeamCapacityTool) Name() string { return "get_team_capacity" }
func (t *GetTeamCapacityTool) Description() string {
	return "Get team members' current workload and capacity"
}
func (t *GetTeamCapacityTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}
func (t *GetTeamCapacityTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	// Get team members with their task counts
	query := `
		SELECT 
			tm.id, tm.name, tm.role,
			COUNT(CASE WHEN t.status = 'in_progress' THEN 1 END) as active_tasks,
			COUNT(CASE WHEN t.status = 'todo' THEN 1 END) as pending_tasks
		FROM team_members tm
		LEFT JOIN tasks t ON t.assignee_id = tm.id AND t.status IN ('todo', 'in_progress')
		WHERE tm.user_id = $1
		GROUP BY tm.id, tm.name, tm.role
		ORDER BY active_tasks DESC`

	rows, err := t.pool.Query(ctx, query, t.userID)
	if err != nil {
		return "", fmt.Errorf("failed to get team capacity: %w", err)
	}
	defer rows.Close()

	var result string
	result = "## Team Capacity\n\n"
	result += "| Member | Role | Active | Pending |\n"
	result += "|--------|------|--------|--------|\n"

	for rows.Next() {
		var id uuid.UUID
		var name, role string
		var activeTasks, pendingTasks int
		if err := rows.Scan(&id, &name, &role, &activeTasks, &pendingTasks); err != nil {
			continue
		}
		result += fmt.Sprintf("| %s | %s | %d | %d |\n", name, role, activeTasks, pendingTasks)
	}

	return result, nil
}

// QueryMetricsTool queries business metrics for analysis
type QueryMetricsTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *QueryMetricsTool) Name() string { return "query_metrics" }
func (t *QueryMetricsTool) Description() string {
	return "Query business metrics like task completion rates, project progress, client pipeline"
}
func (t *QueryMetricsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"metric_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"tasks", "projects", "clients", "overview"},
				"description": "Type of metrics to query",
			},
			"time_range": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"week", "month", "quarter", "year"},
				"description": "Time range for metrics",
			},
		},
		"required": []string{"metric_type"},
	}
}
func (t *QueryMetricsTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		MetricType string `json:"metric_type"`
		TimeRange  string `json:"time_range"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}
	if params.TimeRange == "" {
		params.TimeRange = "month"
	}

	var interval string
	switch params.TimeRange {
	case "week":
		interval = "7 days"
	case "month":
		interval = "30 days"
	case "quarter":
		interval = "90 days"
	case "year":
		interval = "365 days"
	default:
		interval = "30 days"
	}

	var result string

	switch params.MetricType {
	case "tasks":
		query := `
			SELECT 
				COUNT(*) FILTER (WHERE status = 'done') as completed,
				COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress,
				COUNT(*) FILTER (WHERE status = 'todo') as todo,
				COUNT(*) as total
			FROM tasks 
			WHERE user_id = $1 AND created_at > NOW() - $2::interval`
		var completed, inProgress, todo, total int
		err := t.pool.QueryRow(ctx, query, t.userID, interval).Scan(&completed, &inProgress, &todo, &total)
		if err != nil {
			return "", err
		}
		completionRate := 0.0
		if total > 0 {
			completionRate = float64(completed) / float64(total) * 100
		}
		result = fmt.Sprintf("## Task Metrics (%s)\n\n", params.TimeRange)
		result += fmt.Sprintf("- **Total Tasks:** %d\n", total)
		result += fmt.Sprintf("- **Completed:** %d (%.1f%%)\n", completed, completionRate)
		result += fmt.Sprintf("- **In Progress:** %d\n", inProgress)
		result += fmt.Sprintf("- **Todo:** %d\n", todo)

	case "projects":
		query := `
			SELECT status, COUNT(*) 
			FROM projects 
			WHERE user_id = $1 
			GROUP BY status`
		rows, err := t.pool.Query(ctx, query, t.userID)
		if err != nil {
			return "", err
		}
		defer rows.Close()
		result = "## Project Metrics\n\n"
		for rows.Next() {
			var status string
			var count int
			rows.Scan(&status, &count)
			result += fmt.Sprintf("- **%s:** %d\n", status, count)
		}

	case "clients":
		query := `
			SELECT pipeline_stage, COUNT(*) 
			FROM clients 
			WHERE user_id = $1 
			GROUP BY pipeline_stage
			ORDER BY COUNT(*) DESC`
		rows, err := t.pool.Query(ctx, query, t.userID)
		if err != nil {
			return "", err
		}
		defer rows.Close()
		result = "## Client Pipeline\n\n"
		for rows.Next() {
			var stage string
			var count int
			rows.Scan(&stage, &count)
			result += fmt.Sprintf("- **%s:** %d\n", stage, count)
		}

	case "overview":
		result = "## Business Overview\n\n"
		// Tasks
		var taskCount int
		t.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = $1", t.userID).Scan(&taskCount)
		result += fmt.Sprintf("- **Total Tasks:** %d\n", taskCount)
		// Projects
		var projectCount int
		t.pool.QueryRow(ctx, "SELECT COUNT(*) FROM projects WHERE user_id = $1", t.userID).Scan(&projectCount)
		result += fmt.Sprintf("- **Total Projects:** %d\n", projectCount)
		// Clients
		var clientCount int
		t.pool.QueryRow(ctx, "SELECT COUNT(*) FROM clients WHERE user_id = $1", t.userID).Scan(&clientCount)
		result += fmt.Sprintf("- **Total Clients:** %d\n", clientCount)
	}

	return result, nil
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

// ========== SEARCH TOOLS ==========

// WebSearchTool performs web searches using DuckDuckGo and other providers
type WebSearchTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *WebSearchTool) Name() string { return "web_search" }
func (t *WebSearchTool) Description() string {
	return "Search the web for current information. Use this when you need up-to-date information, facts, news, or data that might not be in your training data. Returns search results with titles, URLs, and snippets."
}
func (t *WebSearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query. Be specific and use keywords for better results.",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return (default: 5, max: 10)",
			},
		},
		"required": []string{"query"},
	}
}
func (t *WebSearchTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	if params.MaxResults <= 0 {
		params.MaxResults = 5
	}
	if params.MaxResults > 10 {
		params.MaxResults = 10
	}

	// Use the WebSearchService from services package
	// We need to import the services package and use it here
	// For now, we'll use a simple HTTP call to DuckDuckGo Lite

	results, err := t.performDuckDuckGoSearch(ctx, params.Query, params.MaxResults)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("No results found for: %s", params.Query), nil
	}

	// Format results
	var output string
	output = fmt.Sprintf("## Web Search Results for: \"%s\"\n\n", params.Query)
	for i, result := range results {
		output += fmt.Sprintf("### %d. %s\n", i+1, result.Title)
		output += fmt.Sprintf("**URL:** %s\n", result.URL)
		output += fmt.Sprintf("%s\n\n", result.Snippet)
	}

	return output, nil
}

// SearchResult represents a single search result
type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

// performDuckDuckGoSearch performs a search using DuckDuckGo Lite HTML
func (t *WebSearchTool) performDuckDuckGoSearch(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	// Build the DuckDuckGo Lite URL
	searchURL := fmt.Sprintf("https://lite.duckduckgo.com/lite/?q=%s", url.QueryEscape(query))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse results from HTML
	results := t.parseDuckDuckGoLiteHTML(string(body), maxResults)

	return results, nil
}

// parseDuckDuckGoLiteHTML parses the DuckDuckGo Lite HTML response
func (t *WebSearchTool) parseDuckDuckGoLiteHTML(htmlContent string, maxResults int) []SearchResult {
	var results []SearchResult

	// DuckDuckGo Lite uses a table with class "result-link" for links
	// Pattern: <a rel="nofollow" href="URL" class="result-link">TITLE</a>
	// Snippet is in the next <td class="result-snippet">

	linkPattern := regexp.MustCompile(`<a[^>]*class="result-link"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`)
	snippetPattern := regexp.MustCompile(`<td[^>]*class="result-snippet"[^>]*>([^<]+)</td>`)

	linkMatches := linkPattern.FindAllStringSubmatch(htmlContent, -1)
	snippetMatches := snippetPattern.FindAllStringSubmatch(htmlContent, -1)

	for i := 0; i < len(linkMatches) && i < maxResults; i++ {
		result := SearchResult{
			URL:   html.UnescapeString(linkMatches[i][1]),
			Title: html.UnescapeString(strings.TrimSpace(linkMatches[i][2])),
		}

		// Get corresponding snippet if available
		if i < len(snippetMatches) {
			result.Snippet = html.UnescapeString(strings.TrimSpace(snippetMatches[i][1]))
		}

		// Skip empty results
		if result.URL != "" && result.Title != "" {
			results = append(results, result)
		}
	}

	return results
}

// ========== CONTEXT TOOLS (for AI Agent tree navigation) ==========

// ContextServiceInterface defines the interface for context operations
type ContextServiceInterface interface {
	SearchTree(ctx context.Context, userID string, params TreeSearchParams) ([]TreeSearchResult, error)
	GetContextTree(ctx context.Context, userID string, projectID, nodeID *uuid.UUID) (*ContextTree, error)
	LoadContextItem(ctx context.Context, userID string, itemID uuid.UUID, itemType string) (*ContextItem, error)
}

// TreeSearchParams for context search
type TreeSearchParams struct {
	Query       string   `json:"query"`
	SearchType  string   `json:"search_type"`  // 'title', 'content', 'semantic'
	EntityTypes []string `json:"entity_types"` // 'memories', 'contexts', 'artifacts', 'documents'
	MaxResults  int      `json:"max_results"`
}

// TreeSearchResult represents a search result
type TreeSearchResult struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	Summary        string    `json:"summary,omitempty"`
	RelevanceScore float64   `json:"relevance_score"`
	TreePath       []string  `json:"tree_path"`
	TokenEstimate  int       `json:"token_estimate"`
}

// ContextTree represents the hierarchical context structure
type ContextTree struct {
	RootNode    *ContextTreeNode `json:"root_node"`
	TotalItems  int              `json:"total_items"`
}

// ContextTreeNode represents a node in the context tree
type ContextTreeNode struct {
	ID          uuid.UUID          `json:"id"`
	Type        string             `json:"type"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Icon        string             `json:"icon,omitempty"`
	ItemCount   int                `json:"item_count"`
	Children    []*ContextTreeNode `json:"children,omitempty"`
}

// ContextItem represents a loaded context item
type ContextItem struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	TokenCount int       `json:"token_count"`
}

// registerContextTools registers the context navigation tools
func (r *AgentToolRegistry) registerContextTools() {
	if r.embeddingService == nil {
		return // Context tools require embedding service for semantic search
	}

	r.tools["tree_search"] = &TreeSearchTool{pool: r.pool, userID: r.userID, embeddingService: r.embeddingService}
	r.tools["browse_tree"] = &BrowseTreeTool{pool: r.pool, userID: r.userID}
	r.tools["load_context"] = &LoadContextTool{pool: r.pool, userID: r.userID}
}

// ========== TreeSearchTool ==========

// TreeSearchTool searches the context tree (memories, documents, artifacts)
type TreeSearchTool struct {
	pool             *pgxpool.Pool
	userID           string
	embeddingService EmbeddingServiceInterface
}

func (t *TreeSearchTool) Name() string { return "tree_search" }
func (t *TreeSearchTool) Description() string {
	return "Search through the user's knowledge base including memories, documents, and artifacts. Use this to find relevant context before answering questions. Supports title search, content search, and semantic (meaning-based) search."
}
func (t *TreeSearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query - what you're looking for",
			},
			"search_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"title", "content", "semantic"},
				"description": "Type of search: 'title' for name matching, 'content' for text search, 'semantic' for meaning-based search (default: semantic)",
			},
			"entity_types": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{"memories", "documents", "artifacts", "contexts"},
				},
				"description": "Types of items to search (default: all types)",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return (default: 10, max: 25)",
			},
		},
		"required": []string{"query"},
	}
}

func (t *TreeSearchTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Query       string   `json:"query"`
		SearchType  string   `json:"search_type"`
		EntityTypes []string `json:"entity_types"`
		MaxResults  int      `json:"max_results"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Query == "" {
		return "", fmt.Errorf("query is required")
	}
	if params.SearchType == "" {
		params.SearchType = "semantic"
	}
	if params.MaxResults <= 0 {
		params.MaxResults = 10
	}
	if params.MaxResults > 25 {
		params.MaxResults = 25
	}

	var results []TreeSearchResult
	var err error

	switch params.SearchType {
	case "semantic":
		results, err = t.semanticSearch(ctx, params)
	case "title":
		results, err = t.titleSearch(ctx, params)
	case "content":
		results, err = t.contentSearch(ctx, params)
	default:
		results, err = t.titleSearch(ctx, params)
	}

	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("No results found for: \"%s\"", params.Query), nil
	}

	// Format results
	var output strings.Builder
	output.WriteString(fmt.Sprintf("## Search Results for: \"%s\" (%s search)\n\n", params.Query, params.SearchType))
	output.WriteString(fmt.Sprintf("Found %d results:\n\n", len(results)))

	for i, r := range results {
		output.WriteString(fmt.Sprintf("### %d. %s\n", i+1, r.Title))
		output.WriteString(fmt.Sprintf("- **Type:** %s\n", r.Type))
		output.WriteString(fmt.Sprintf("- **ID:** %s\n", r.ID.String()))
		if r.Summary != "" {
			output.WriteString(fmt.Sprintf("- **Summary:** %s\n", r.Summary))
		}
		if params.SearchType == "semantic" {
			output.WriteString(fmt.Sprintf("- **Relevance:** %.2f\n", r.RelevanceScore))
		}
		output.WriteString("\n")
	}

	output.WriteString("\n*Use `load_context` tool with an ID to load the full content of any item.*")

	return output.String(), nil
}

func (t *TreeSearchTool) semanticSearch(ctx context.Context, params struct {
	Query       string   `json:"query"`
	SearchType  string   `json:"search_type"`
	EntityTypes []string `json:"entity_types"`
	MaxResults  int      `json:"max_results"`
}) ([]TreeSearchResult, error) {
	if t.embeddingService == nil {
		return t.titleSearch(ctx, params)
	}

	// Generate query embedding
	queryEmbedding, err := t.embeddingService.GenerateEmbedding(ctx, params.Query)
	if err != nil {
		return t.titleSearch(ctx, params) // Fallback to title search
	}

	var results []TreeSearchResult
	shouldSearchType := func(entityType string) bool {
		if len(params.EntityTypes) == 0 {
			return true
		}
		for _, et := range params.EntityTypes {
			if et == entityType {
				return true
			}
		}
		return false
	}

	// Search memories
	if shouldSearchType("memories") {
		rows, _ := t.pool.Query(ctx, `
			SELECT id, COALESCE(title, summary, LEFT(content, 100)), memory_type, COALESCE(summary, LEFT(content, 200)),
			       1 - (embedding <=> $1::vector) as similarity
			FROM memories
			WHERE user_id = $2 AND is_active = true AND embedding IS NOT NULL
			ORDER BY embedding <=> $1::vector
			LIMIT $3
		`, fmt.Sprintf("[%s]", floatsToString(queryEmbedding)), t.userID, params.MaxResults)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var r TreeSearchResult
				var memType string
				if err := rows.Scan(&r.ID, &r.Title, &memType, &r.Summary, &r.RelevanceScore); err != nil {
					continue
				}
				r.Type = "memory"
				r.TreePath = []string{"Memories", memType}
				results = append(results, r)
			}
		}
	}

	// Search documents
	if shouldSearchType("documents") {
		rows, _ := t.pool.Query(ctx, `
			SELECT id, COALESCE(display_name, filename), COALESCE(document_type, 'document'),
			       COALESCE(description, LEFT(extracted_text, 200)),
			       1 - (embedding <=> $1::vector) as similarity
			FROM uploaded_documents
			WHERE user_id = $2 AND embedding IS NOT NULL
			ORDER BY embedding <=> $1::vector
			LIMIT $3
		`, fmt.Sprintf("[%s]", floatsToString(queryEmbedding)), t.userID, params.MaxResults)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var r TreeSearchResult
				var docType string
				if err := rows.Scan(&r.ID, &r.Title, &docType, &r.Summary, &r.RelevanceScore); err != nil {
					continue
				}
				r.Type = "document"
				r.TreePath = []string{"Documents", docType}
				results = append(results, r)
			}
		}
	}

	// Sort by relevance
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].RelevanceScore > results[i].RelevanceScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > params.MaxResults {
		results = results[:params.MaxResults]
	}

	return results, nil
}

func (t *TreeSearchTool) titleSearch(ctx context.Context, params struct {
	Query       string   `json:"query"`
	SearchType  string   `json:"search_type"`
	EntityTypes []string `json:"entity_types"`
	MaxResults  int      `json:"max_results"`
}) ([]TreeSearchResult, error) {
	var results []TreeSearchResult
	searchPattern := "%" + params.Query + "%"

	shouldSearchType := func(entityType string) bool {
		if len(params.EntityTypes) == 0 {
			return true
		}
		for _, et := range params.EntityTypes {
			if et == entityType {
				return true
			}
		}
		return false
	}

	// Search memories
	if shouldSearchType("memories") {
		rows, _ := t.pool.Query(ctx, `
			SELECT id, COALESCE(title, summary, LEFT(content, 100)), memory_type, COALESCE(summary, LEFT(content, 200))
			FROM memories
			WHERE user_id = $1 AND is_active = true AND (title ILIKE $2 OR summary ILIKE $2)
			ORDER BY importance_score DESC
			LIMIT $3
		`, t.userID, searchPattern, params.MaxResults)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var r TreeSearchResult
				var memType string
				if err := rows.Scan(&r.ID, &r.Title, &memType, &r.Summary); err != nil {
					continue
				}
				r.Type = "memory"
				r.TreePath = []string{"Memories", memType}
				r.RelevanceScore = 0.8
				results = append(results, r)
			}
		}
	}

	// Search documents
	if shouldSearchType("documents") {
		rows, _ := t.pool.Query(ctx, `
			SELECT id, COALESCE(display_name, filename), COALESCE(document_type, 'document'),
			       COALESCE(description, LEFT(extracted_text, 200))
			FROM uploaded_documents
			WHERE user_id = $1 AND (display_name ILIKE $2 OR filename ILIKE $2 OR description ILIKE $2)
			ORDER BY created_at DESC
			LIMIT $3
		`, t.userID, searchPattern, params.MaxResults)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var r TreeSearchResult
				var docType string
				if err := rows.Scan(&r.ID, &r.Title, &docType, &r.Summary); err != nil {
					continue
				}
				r.Type = "document"
				r.TreePath = []string{"Documents", docType}
				r.RelevanceScore = 0.7
				results = append(results, r)
			}
		}
	}

	// Search artifacts
	if shouldSearchType("artifacts") {
		rows, _ := t.pool.Query(ctx, `
			SELECT id, title, type, LEFT(content, 200)
			FROM artifacts
			WHERE user_id = $1 AND (title ILIKE $2 OR content ILIKE $2)
			ORDER BY created_at DESC
			LIMIT $3
		`, t.userID, searchPattern, params.MaxResults)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var r TreeSearchResult
				var artType string
				if err := rows.Scan(&r.ID, &r.Title, &artType, &r.Summary); err != nil {
					continue
				}
				r.Type = "artifact"
				r.TreePath = []string{"Artifacts", artType}
				r.RelevanceScore = 0.7
				results = append(results, r)
			}
		}
	}

	if len(results) > params.MaxResults {
		results = results[:params.MaxResults]
	}

	return results, nil
}

func (t *TreeSearchTool) contentSearch(ctx context.Context, params struct {
	Query       string   `json:"query"`
	SearchType  string   `json:"search_type"`
	EntityTypes []string `json:"entity_types"`
	MaxResults  int      `json:"max_results"`
}) ([]TreeSearchResult, error) {
	var results []TreeSearchResult
	searchPattern := "%" + params.Query + "%"

	shouldSearchType := func(entityType string) bool {
		if len(params.EntityTypes) == 0 {
			return true
		}
		for _, et := range params.EntityTypes {
			if et == entityType {
				return true
			}
		}
		return false
	}

	// Search memories by content
	if shouldSearchType("memories") {
		rows, _ := t.pool.Query(ctx, `
			SELECT id, COALESCE(title, summary, LEFT(content, 100)), memory_type, LEFT(content, 200)
			FROM memories
			WHERE user_id = $1 AND is_active = true AND content ILIKE $2
			ORDER BY importance_score DESC
			LIMIT $3
		`, t.userID, searchPattern, params.MaxResults)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var r TreeSearchResult
				var memType string
				if err := rows.Scan(&r.ID, &r.Title, &memType, &r.Summary); err != nil {
					continue
				}
				r.Type = "memory"
				r.TreePath = []string{"Memories", memType}
				r.RelevanceScore = 0.7
				results = append(results, r)
			}
		}
	}

	// Search documents by extracted text
	if shouldSearchType("documents") {
		rows, _ := t.pool.Query(ctx, `
			SELECT id, COALESCE(display_name, filename), COALESCE(document_type, 'document'), LEFT(extracted_text, 200)
			FROM uploaded_documents
			WHERE user_id = $1 AND extracted_text ILIKE $2
			ORDER BY created_at DESC
			LIMIT $3
		`, t.userID, searchPattern, params.MaxResults)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var r TreeSearchResult
				var docType string
				if err := rows.Scan(&r.ID, &r.Title, &docType, &r.Summary); err != nil {
					continue
				}
				r.Type = "document"
				r.TreePath = []string{"Documents", docType}
				r.RelevanceScore = 0.6
				results = append(results, r)
			}
		}
	}

	if len(results) > params.MaxResults {
		results = results[:params.MaxResults]
	}

	return results, nil
}

// Helper function to convert float slice to string
func floatsToString(floats []float32) string {
	strs := make([]string, len(floats))
	for i, f := range floats {
		strs[i] = fmt.Sprintf("%f", f)
	}
	return strings.Join(strs, ",")
}

// ========== BrowseTreeTool ==========

// BrowseTreeTool browses the context tree hierarchy
type BrowseTreeTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *BrowseTreeTool) Name() string { return "browse_tree" }
func (t *BrowseTreeTool) Description() string {
	return "Browse the user's knowledge tree to see what projects, memories, documents, and artifacts are available. Use this to understand the structure of available context before searching."
}
func (t *BrowseTreeTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Filter to a specific project UUID",
			},
			"show_counts": map[string]interface{}{
				"type":        "boolean",
				"description": "Show item counts for each category (default: true)",
			},
		},
	}
}

func (t *BrowseTreeTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ProjectID  string `json:"project_id"`
		ShowCounts bool   `json:"show_counts"`
	}
	params.ShowCounts = true // default
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	var output strings.Builder
	output.WriteString("## Knowledge Tree\n\n")

	// Get overall statistics
	var memoryCount, docCount, artifactCount, projectCount int
	t.pool.QueryRow(ctx, `SELECT COUNT(*) FROM memories WHERE user_id = $1 AND is_active = true`, t.userID).Scan(&memoryCount)
	t.pool.QueryRow(ctx, `SELECT COUNT(*) FROM uploaded_documents WHERE user_id = $1`, t.userID).Scan(&docCount)
	t.pool.QueryRow(ctx, `SELECT COUNT(*) FROM artifacts WHERE user_id = $1`, t.userID).Scan(&artifactCount)
	t.pool.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE user_id = $1 AND is_archived = false`, t.userID).Scan(&projectCount)

	if params.ShowCounts {
		output.WriteString("### Overview\n")
		output.WriteString(fmt.Sprintf("- **Projects:** %d\n", projectCount))
		output.WriteString(fmt.Sprintf("- **Memories:** %d\n", memoryCount))
		output.WriteString(fmt.Sprintf("- **Documents:** %d\n", docCount))
		output.WriteString(fmt.Sprintf("- **Artifacts:** %d\n", artifactCount))
		output.WriteString("\n")
	}

	// List projects
	output.WriteString("### Projects\n")
	query := `SELECT id, name, COALESCE(description, ''), status FROM projects WHERE user_id = $1 AND is_archived = false ORDER BY updated_at DESC LIMIT 20`
	args := []interface{}{t.userID}

	if params.ProjectID != "" {
		if projectUUID, err := uuid.Parse(params.ProjectID); err == nil {
			query = `SELECT id, name, COALESCE(description, ''), status FROM projects WHERE user_id = $1 AND id = $2`
			args = append(args, projectUUID)
		}
	}

	rows, err := t.pool.Query(ctx, query, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			var name, description, status string
			if rows.Scan(&id, &name, &description, &status) == nil {
				output.WriteString(fmt.Sprintf("- **%s** [%s] (ID: %s)\n", name, status, id.String()))
				if description != "" {
					output.WriteString(fmt.Sprintf("  %s\n", description))
				}
			}
		}
	}

	if projectCount == 0 {
		output.WriteString("No projects found.\n")
	}

	// Memory types breakdown
	output.WriteString("\n### Memory Types\n")
	typeRows, _ := t.pool.Query(ctx, `
		SELECT memory_type, COUNT(*)
		FROM memories
		WHERE user_id = $1 AND is_active = true
		GROUP BY memory_type
		ORDER BY COUNT(*) DESC
	`, t.userID)
	if typeRows != nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var memType string
			var count int
			if typeRows.Scan(&memType, &count) == nil {
				output.WriteString(fmt.Sprintf("- %s: %d\n", memType, count))
			}
		}
	}

	output.WriteString("\n*Use `tree_search` to find specific items, or `load_context` to load an item by ID.*")

	return output.String(), nil
}

// ========== LoadContextTool ==========

// LoadContextTool loads a specific context item by ID
type LoadContextTool struct {
	pool   *pgxpool.Pool
	userID string
}

func (t *LoadContextTool) Name() string { return "load_context" }
func (t *LoadContextTool) Description() string {
	return "Load the full content of a specific memory, document, or artifact by its ID. Use this after finding items with tree_search or browse_tree to get the complete content."
}
func (t *LoadContextTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"item_id": map[string]interface{}{
				"type":        "string",
				"description": "The UUID of the item to load",
			},
			"item_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"memory", "document", "artifact"},
				"description": "The type of item to load",
			},
		},
		"required": []string{"item_id", "item_type"},
	}
}

func (t *LoadContextTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ItemID   string `json:"item_id"`
		ItemType string `json:"item_type"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.ItemID == "" {
		return "", fmt.Errorf("item_id is required")
	}
	if params.ItemType == "" {
		return "", fmt.Errorf("item_type is required")
	}

	itemUUID, err := uuid.Parse(params.ItemID)
	if err != nil {
		return "", fmt.Errorf("invalid item_id: %w", err)
	}

	var output strings.Builder

	switch params.ItemType {
	case "memory":
		var title, content, memType, summary string
		var importance int
		var createdAt time.Time
		err := t.pool.QueryRow(ctx, `
			SELECT COALESCE(title, 'Untitled'), content, memory_type, COALESCE(summary, ''), importance_score, created_at
			FROM memories
			WHERE id = $1 AND user_id = $2
		`, itemUUID, t.userID).Scan(&title, &content, &memType, &summary, &importance, &createdAt)
		if err != nil {
			return "", fmt.Errorf("memory not found: %w", err)
		}

		output.WriteString(fmt.Sprintf("## Memory: %s\n\n", title))
		output.WriteString(fmt.Sprintf("- **Type:** %s\n", memType))
		output.WriteString(fmt.Sprintf("- **Importance:** %d/10\n", importance))
		output.WriteString(fmt.Sprintf("- **Created:** %s\n", createdAt.Format("2006-01-02 15:04")))
		if summary != "" {
			output.WriteString(fmt.Sprintf("- **Summary:** %s\n", summary))
		}
		output.WriteString("\n### Content\n\n")
		output.WriteString(content)

	case "document":
		var displayName, filename, docType, description, extractedText string
		var createdAt time.Time
		err := t.pool.QueryRow(ctx, `
			SELECT COALESCE(display_name, filename), filename, COALESCE(document_type, 'document'),
			       COALESCE(description, ''), COALESCE(extracted_text, ''), created_at
			FROM uploaded_documents
			WHERE id = $1 AND user_id = $2
		`, itemUUID, t.userID).Scan(&displayName, &filename, &docType, &description, &extractedText, &createdAt)
		if err != nil {
			return "", fmt.Errorf("document not found: %w", err)
		}

		output.WriteString(fmt.Sprintf("## Document: %s\n\n", displayName))
		output.WriteString(fmt.Sprintf("- **Filename:** %s\n", filename))
		output.WriteString(fmt.Sprintf("- **Type:** %s\n", docType))
		output.WriteString(fmt.Sprintf("- **Uploaded:** %s\n", createdAt.Format("2006-01-02 15:04")))
		if description != "" {
			output.WriteString(fmt.Sprintf("- **Description:** %s\n", description))
		}
		output.WriteString("\n### Extracted Content\n\n")
		if extractedText != "" {
			// Limit content to avoid token overflow
			if len(extractedText) > 10000 {
				output.WriteString(extractedText[:10000])
				output.WriteString("\n\n*[Content truncated - document is very large]*")
			} else {
				output.WriteString(extractedText)
			}
		} else {
			output.WriteString("*No text extracted from this document*")
		}

	case "artifact":
		var title, content, artType string
		var createdAt time.Time
		err := t.pool.QueryRow(ctx, `
			SELECT title, content, type, created_at
			FROM artifacts
			WHERE id = $1 AND user_id = $2
		`, itemUUID, t.userID).Scan(&title, &content, &artType, &createdAt)
		if err != nil {
			return "", fmt.Errorf("artifact not found: %w", err)
		}

		output.WriteString(fmt.Sprintf("## Artifact: %s\n\n", title))
		output.WriteString(fmt.Sprintf("- **Type:** %s\n", artType))
		output.WriteString(fmt.Sprintf("- **Created:** %s\n", createdAt.Format("2006-01-02 15:04")))
		output.WriteString("\n### Content\n\n")
		output.WriteString(content)

	default:
		return "", fmt.Errorf("unknown item_type: %s", params.ItemType)
	}

	return output.String(), nil
}
