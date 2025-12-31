package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkflowExecutionMode defines how steps are executed
type WorkflowExecutionMode string

const (
	ExecutionModeSequential WorkflowExecutionMode = "sequential"
	ExecutionModeParallel   WorkflowExecutionMode = "parallel"
	ExecutionModeSmart      WorkflowExecutionMode = "smart" // Analyzes dependencies
)

// StepActionType defines the type of action a step performs
type StepActionType string

const (
	StepActionCommand   StepActionType = "command"
	StepActionAgent     StepActionType = "agent"
	StepActionTool      StepActionType = "tool"
	StepActionCondition StepActionType = "condition"
	StepActionWait      StepActionType = "wait"
)

// StepFailureAction defines what to do when a step fails
type StepFailureAction string

const (
	FailureActionStop     StepFailureAction = "stop"
	FailureActionContinue StepFailureAction = "continue"
	FailureActionRetry    StepFailureAction = "retry"
	FailureActionSkip     StepFailureAction = "skip"
)

// ExecutionStatus represents the status of an execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
	StatusSkipped   ExecutionStatus = "skipped"
)

// CommandWorkflow represents a multi-step workflow
type CommandWorkflow struct {
	ID             uuid.UUID             `json:"id"`
	UserID         string                `json:"user_id"`
	Name           string                `json:"name"`
	DisplayName    string                `json:"display_name"`
	Description    string                `json:"description,omitempty"`
	Trigger        string                `json:"trigger"`
	ExecutionMode  WorkflowExecutionMode `json:"execution_mode"`
	StopOnFailure  bool                  `json:"stop_on_failure"`
	TimeoutSeconds int                   `json:"timeout_seconds"`
	IsActive       bool                  `json:"is_active"`
	IsSystem       bool                  `json:"is_system"`
	Steps          []WorkflowStep        `json:"steps,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID                  uuid.UUID          `json:"id"`
	WorkflowID          uuid.UUID          `json:"workflow_id"`
	Name                string             `json:"name"`
	Description         string             `json:"description,omitempty"`
	StepOrder           int                `json:"step_order"`
	ActionType          StepActionType     `json:"action_type"`
	CommandTrigger      *string            `json:"command_trigger,omitempty"`
	CommandArgs         *string            `json:"command_args,omitempty"`
	TargetAgentID       *uuid.UUID         `json:"target_agent_id,omitempty"`
	PromptTemplate      *string            `json:"prompt_template,omitempty"`
	ToolName            *string            `json:"tool_name,omitempty"`
	ToolParams          map[string]any     `json:"tool_params,omitempty"`
	ConditionExpression *string            `json:"condition_expression,omitempty"`
	OnTrueStep          *uuid.UUID         `json:"on_true_step,omitempty"`
	OnFalseStep         *uuid.UUID         `json:"on_false_step,omitempty"`
	WaitSeconds         int                `json:"wait_seconds,omitempty"`
	DependsOn           []uuid.UUID        `json:"depends_on,omitempty"`
	CanParallel         bool               `json:"can_parallel"`
	OnFailure           StepFailureAction  `json:"on_failure"`
	MaxRetries          int                `json:"max_retries"`
	RetryDelaySeconds   int                `json:"retry_delay_seconds"`
	InputMapping        map[string]string  `json:"input_mapping,omitempty"`
	OutputKey           *string            `json:"output_key,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

// WorkflowExecution represents a running or completed workflow execution
type WorkflowExecution struct {
	ID             uuid.UUID        `json:"id"`
	WorkflowID     uuid.UUID        `json:"workflow_id"`
	UserID         string           `json:"user_id"`
	ConversationID *uuid.UUID       `json:"conversation_id,omitempty"`
	InitialInput   string           `json:"initial_input,omitempty"`
	Context        map[string]any   `json:"context"`
	Status         ExecutionStatus  `json:"status"`
	CurrentStepID  *uuid.UUID       `json:"current_step_id,omitempty"`
	Result         map[string]any   `json:"result,omitempty"`
	ErrorMessage   *string          `json:"error_message,omitempty"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
}

// StepExecution represents a step execution within a workflow
type StepExecution struct {
	ID            uuid.UUID       `json:"id"`
	ExecutionID   uuid.UUID       `json:"execution_id"`
	StepID        uuid.UUID       `json:"step_id"`
	Status        ExecutionStatus `json:"status"`
	AttemptNumber int             `json:"attempt_number"`
	Input         map[string]any  `json:"input,omitempty"`
	Output        map[string]any  `json:"output,omitempty"`
	ErrorMessage  *string         `json:"error_message,omitempty"`
	StartedAt     *time.Time      `json:"started_at,omitempty"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	DurationMs    *float64        `json:"duration_ms,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// WorkflowService handles workflow operations
type WorkflowService struct {
	pool           *pgxpool.Pool
	commandService *CommandService
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(pool *pgxpool.Pool) *WorkflowService {
	return &WorkflowService{
		pool:           pool,
		commandService: NewCommandService(pool),
	}
}

// CreateWorkflow creates a new workflow
func (s *WorkflowService) CreateWorkflow(ctx context.Context, workflow *CommandWorkflow) error {
	if s.pool == nil {
		return fmt.Errorf("no database connection")
	}

	// Ensure trigger starts with /
	if !strings.HasPrefix(workflow.Trigger, "/") {
		workflow.Trigger = "/" + workflow.Trigger
	}

	workflow.ID = uuid.New()
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	_, err := s.pool.Exec(ctx, `
		INSERT INTO command_workflows (
			id, user_id, name, display_name, description, trigger,
			execution_mode, stop_on_failure, timeout_seconds, is_active, is_system
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, workflow.ID, workflow.UserID, workflow.Name, workflow.DisplayName,
		workflow.Description, workflow.Trigger, workflow.ExecutionMode,
		workflow.StopOnFailure, workflow.TimeoutSeconds, workflow.IsActive, workflow.IsSystem)

	return err
}

// AddStep adds a step to a workflow
func (s *WorkflowService) AddStep(ctx context.Context, step *WorkflowStep) error {
	if s.pool == nil {
		return fmt.Errorf("no database connection")
	}

	step.ID = uuid.New()
	step.CreatedAt = time.Now()
	step.UpdatedAt = time.Now()

	toolParamsJSON, _ := json.Marshal(step.ToolParams)
	inputMappingJSON, _ := json.Marshal(step.InputMapping)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO workflow_steps (
			id, workflow_id, name, description, step_order, action_type,
			command_trigger, command_args, target_agent_id, prompt_template,
			tool_name, tool_params, condition_expression, on_true_step, on_false_step,
			wait_seconds, depends_on, can_parallel, on_failure, max_retries,
			retry_delay_seconds, input_mapping, output_key
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		)
	`, step.ID, step.WorkflowID, step.Name, step.Description, step.StepOrder,
		step.ActionType, step.CommandTrigger, step.CommandArgs, step.TargetAgentID,
		step.PromptTemplate, step.ToolName, toolParamsJSON, step.ConditionExpression,
		step.OnTrueStep, step.OnFalseStep, step.WaitSeconds, step.DependsOn,
		step.CanParallel, step.OnFailure, step.MaxRetries, step.RetryDelaySeconds,
		inputMappingJSON, step.OutputKey)

	return err
}

// GetWorkflow gets a workflow by ID with its steps
func (s *WorkflowService) GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*CommandWorkflow, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("no database connection")
	}

	var workflow CommandWorkflow
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, name, display_name, description, trigger,
			execution_mode, stop_on_failure, timeout_seconds, is_active, is_system,
			created_at, updated_at
		FROM command_workflows WHERE id = $1
	`, workflowID).Scan(
		&workflow.ID, &workflow.UserID, &workflow.Name, &workflow.DisplayName,
		&workflow.Description, &workflow.Trigger, &workflow.ExecutionMode,
		&workflow.StopOnFailure, &workflow.TimeoutSeconds, &workflow.IsActive,
		&workflow.IsSystem, &workflow.CreatedAt, &workflow.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get steps
	rows, err := s.pool.Query(ctx, `
		SELECT id, workflow_id, name, description, step_order, action_type,
			command_trigger, command_args, target_agent_id, prompt_template,
			tool_name, tool_params, condition_expression, on_true_step, on_false_step,
			wait_seconds, depends_on, can_parallel, on_failure, max_retries,
			retry_delay_seconds, input_mapping, output_key, created_at, updated_at
		FROM workflow_steps WHERE workflow_id = $1 ORDER BY step_order
	`, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var step WorkflowStep
		var toolParamsJSON, inputMappingJSON []byte
		err := rows.Scan(
			&step.ID, &step.WorkflowID, &step.Name, &step.Description, &step.StepOrder,
			&step.ActionType, &step.CommandTrigger, &step.CommandArgs, &step.TargetAgentID,
			&step.PromptTemplate, &step.ToolName, &toolParamsJSON, &step.ConditionExpression,
			&step.OnTrueStep, &step.OnFalseStep, &step.WaitSeconds, &step.DependsOn,
			&step.CanParallel, &step.OnFailure, &step.MaxRetries, &step.RetryDelaySeconds,
			&inputMappingJSON, &step.OutputKey, &step.CreatedAt, &step.UpdatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(toolParamsJSON, &step.ToolParams)
		json.Unmarshal(inputMappingJSON, &step.InputMapping)
		workflow.Steps = append(workflow.Steps, step)
	}

	return &workflow, nil
}

// ResolveWorkflow finds a workflow by trigger
func (s *WorkflowService) ResolveWorkflow(ctx context.Context, userID string, trigger string) (*CommandWorkflow, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("no database connection")
	}

	if !strings.HasPrefix(trigger, "/") {
		trigger = "/" + trigger
	}
	trigger = strings.ToLower(trigger)

	var workflowID uuid.UUID

	// Try user's workflows first
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM command_workflows
		WHERE user_id = $1 AND trigger = $2 AND is_active = TRUE
	`, userID, trigger).Scan(&workflowID)
	if err != nil {
		// Try system workflows
		err = s.pool.QueryRow(ctx, `
			SELECT id FROM command_workflows
			WHERE user_id = 'SYSTEM' AND trigger = $1 AND is_active = TRUE
		`, trigger).Scan(&workflowID)
		if err != nil {
			return nil, fmt.Errorf("workflow not found: %s", trigger)
		}
	}

	return s.GetWorkflow(ctx, workflowID)
}

// ListWorkflows lists all workflows for a user
func (s *WorkflowService) ListWorkflows(ctx context.Context, userID string) ([]CommandWorkflow, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("no database connection")
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, name, display_name, description, trigger,
			execution_mode, stop_on_failure, timeout_seconds, is_active, is_system,
			created_at, updated_at
		FROM command_workflows
		WHERE user_id = $1 OR user_id = 'SYSTEM'
		ORDER BY is_system DESC, name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []CommandWorkflow
	for rows.Next() {
		var w CommandWorkflow
		err := rows.Scan(
			&w.ID, &w.UserID, &w.Name, &w.DisplayName, &w.Description, &w.Trigger,
			&w.ExecutionMode, &w.StopOnFailure, &w.TimeoutSeconds, &w.IsActive,
			&w.IsSystem, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			continue
		}
		workflows = append(workflows, w)
	}

	return workflows, nil
}

// ExecuteWorkflow starts a workflow execution
func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, workflow *CommandWorkflow, userID string, input string, conversationID *uuid.UUID) (*WorkflowExecution, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("no database connection")
	}

	// Create execution record
	exec := &WorkflowExecution{
		ID:             uuid.New(),
		WorkflowID:     workflow.ID,
		UserID:         userID,
		ConversationID: conversationID,
		InitialInput:   input,
		Context:        make(map[string]any),
		Status:         StatusRunning,
		CreatedAt:      time.Now(),
	}
	now := time.Now()
	exec.StartedAt = &now

	// Store initial input in context
	exec.Context["input"] = input
	exec.Context["workflow_name"] = workflow.Name
	exec.Context["started_at"] = now.Format(time.RFC3339)

	contextJSON, _ := json.Marshal(exec.Context)
	resultJSON, _ := json.Marshal(exec.Result)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO workflow_executions (
			id, workflow_id, user_id, conversation_id, initial_input,
			context, status, started_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, exec.ID, exec.WorkflowID, exec.UserID, exec.ConversationID,
		exec.InitialInput, contextJSON, exec.Status, exec.StartedAt)
	if err != nil {
		return nil, err
	}

	// Execute steps based on mode
	switch workflow.ExecutionMode {
	case ExecutionModeSequential:
		err = s.executeSequential(ctx, workflow, exec)
	case ExecutionModeParallel:
		err = s.executeParallel(ctx, workflow, exec)
	case ExecutionModeSmart:
		err = s.executeSmart(ctx, workflow, exec)
	default:
		err = s.executeSequential(ctx, workflow, exec)
	}

	// Update final status
	completedAt := time.Now()
	exec.CompletedAt = &completedAt

	if err != nil {
		exec.Status = StatusFailed
		errMsg := err.Error()
		exec.ErrorMessage = &errMsg
	} else {
		exec.Status = StatusCompleted
	}

	resultJSON, _ = json.Marshal(exec.Result)
	contextJSON, _ = json.Marshal(exec.Context)

	_, _ = s.pool.Exec(ctx, `
		UPDATE workflow_executions
		SET status = $1, result = $2, context = $3, error_message = $4, completed_at = $5
		WHERE id = $6
	`, exec.Status, resultJSON, contextJSON, exec.ErrorMessage, exec.CompletedAt, exec.ID)

	return exec, err
}

// executeSequential runs steps one after another
func (s *WorkflowService) executeSequential(ctx context.Context, workflow *CommandWorkflow, exec *WorkflowExecution) error {
	for i, step := range workflow.Steps {
		// Update current step
		exec.CurrentStepID = &step.ID
		_, _ = s.pool.Exec(ctx, `
			UPDATE workflow_executions SET current_step_id = $1 WHERE id = $2
		`, step.ID, exec.ID)

		// Execute step
		stepExec, err := s.executeStep(ctx, &step, exec)
		if err != nil {
			slog.Error("Step failed", "step", step.Name, "error", err)

			switch step.OnFailure {
			case FailureActionStop:
				if workflow.StopOnFailure {
					return fmt.Errorf("step %d (%s) failed: %w", i+1, step.Name, err)
				}
			case FailureActionContinue:
				continue
			case FailureActionSkip:
				continue
			case FailureActionRetry:
				// Simple retry logic
				for attempt := 1; attempt <= step.MaxRetries; attempt++ {
					time.Sleep(time.Duration(step.RetryDelaySeconds) * time.Second)
					stepExec, err = s.executeStep(ctx, &step, exec)
					stepExec.AttemptNumber = attempt + 1
					if err == nil {
						break
					}
				}
				if err != nil && workflow.StopOnFailure {
					return fmt.Errorf("step %d (%s) failed after %d retries: %w",
						i+1, step.Name, step.MaxRetries, err)
				}
			}
		}

		// Store step output in context
		if step.OutputKey != nil && *step.OutputKey != "" && stepExec != nil {
			exec.Context[*step.OutputKey] = stepExec.Output
		}
	}

	return nil
}

// executeParallel runs all steps concurrently (respecting dependencies)
func (s *WorkflowService) executeParallel(ctx context.Context, workflow *CommandWorkflow, exec *WorkflowExecution) error {
	// For simple parallel execution without dependencies
	type result struct {
		step    *WorkflowStep
		exec    *StepExecution
		err     error
	}

	results := make(chan result, len(workflow.Steps))

	for i := range workflow.Steps {
		step := &workflow.Steps[i]
		go func(step *WorkflowStep) {
			stepExec, err := s.executeStep(ctx, step, exec)
			results <- result{step: step, exec: stepExec, err: err}
		}(step)
	}

	var firstError error
	for i := 0; i < len(workflow.Steps); i++ {
		r := <-results
		if r.err != nil && firstError == nil {
			firstError = fmt.Errorf("step %s failed: %w", r.step.Name, r.err)
		}
		if r.step.OutputKey != nil && *r.step.OutputKey != "" && r.exec != nil {
			exec.Context[*r.step.OutputKey] = r.exec.Output
		}
	}

	if firstError != nil && workflow.StopOnFailure {
		return firstError
	}
	return nil
}

// executeSmart analyzes dependencies and runs optimally
func (s *WorkflowService) executeSmart(ctx context.Context, workflow *CommandWorkflow, exec *WorkflowExecution) error {
	// Build dependency graph and execute in waves
	completed := make(map[uuid.UUID]bool)

	for len(completed) < len(workflow.Steps) {
		var wave []*WorkflowStep

		// Find steps that can run (all dependencies met)
		for i := range workflow.Steps {
			step := &workflow.Steps[i]
			if completed[step.ID] {
				continue
			}

			canRun := true
			for _, depID := range step.DependsOn {
				if !completed[depID] {
					canRun = false
					break
				}
			}

			if canRun {
				wave = append(wave, step)
			}
		}

		if len(wave) == 0 {
			return fmt.Errorf("circular dependency detected or no runnable steps")
		}

		// Execute wave in parallel
		type result struct {
			step *WorkflowStep
			exec *StepExecution
			err  error
		}
		results := make(chan result, len(wave))

		for _, step := range wave {
			go func(step *WorkflowStep) {
				stepExec, err := s.executeStep(ctx, step, exec)
				results <- result{step: step, exec: stepExec, err: err}
			}(step)
		}

		for i := 0; i < len(wave); i++ {
			r := <-results
			completed[r.step.ID] = true

			if r.err != nil && workflow.StopOnFailure {
				return fmt.Errorf("step %s failed: %w", r.step.Name, r.err)
			}

			if r.step.OutputKey != nil && *r.step.OutputKey != "" && r.exec != nil {
				exec.Context[*r.step.OutputKey] = r.exec.Output
			}
		}
	}

	return nil
}

// executeStep executes a single step
func (s *WorkflowService) executeStep(ctx context.Context, step *WorkflowStep, exec *WorkflowExecution) (*StepExecution, error) {
	stepExec := &StepExecution{
		ID:            uuid.New(),
		ExecutionID:   exec.ID,
		StepID:        step.ID,
		Status:        StatusRunning,
		AttemptNumber: 1,
		Input:         make(map[string]any),
		Output:        make(map[string]any),
		CreatedAt:     time.Now(),
	}
	now := time.Now()
	stepExec.StartedAt = &now

	// Map input from context using input_mapping
	for outputKey, inputPath := range step.InputMapping {
		if val, ok := exec.Context[inputPath]; ok {
			stepExec.Input[outputKey] = val
		}
	}

	// Save step execution start
	inputJSON, _ := json.Marshal(stepExec.Input)
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO workflow_step_executions (
			id, execution_id, step_id, status, attempt_number, input, started_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, stepExec.ID, stepExec.ExecutionID, stepExec.StepID, stepExec.Status,
		stepExec.AttemptNumber, inputJSON, stepExec.StartedAt)

	var err error

	switch step.ActionType {
	case StepActionCommand:
		err = s.executeCommandStep(ctx, step, exec, stepExec)
	case StepActionAgent:
		err = s.executeAgentStep(ctx, step, exec, stepExec)
	case StepActionTool:
		err = s.executeToolStep(ctx, step, exec, stepExec)
	case StepActionCondition:
		err = s.executeConditionStep(ctx, step, exec, stepExec)
	case StepActionWait:
		time.Sleep(time.Duration(step.WaitSeconds) * time.Second)
		stepExec.Output["waited_seconds"] = step.WaitSeconds
	default:
		err = fmt.Errorf("unknown action type: %s", step.ActionType)
	}

	// Update step execution
	completedAt := time.Now()
	stepExec.CompletedAt = &completedAt
	duration := float64(completedAt.Sub(*stepExec.StartedAt).Milliseconds())
	stepExec.DurationMs = &duration

	if err != nil {
		stepExec.Status = StatusFailed
		errMsg := err.Error()
		stepExec.ErrorMessage = &errMsg
	} else {
		stepExec.Status = StatusCompleted
	}

	outputJSON, _ := json.Marshal(stepExec.Output)
	_, _ = s.pool.Exec(ctx, `
		UPDATE workflow_step_executions
		SET status = $1, output = $2, error_message = $3, completed_at = $4, duration_ms = $5
		WHERE id = $6
	`, stepExec.Status, outputJSON, stepExec.ErrorMessage, stepExec.CompletedAt,
		stepExec.DurationMs, stepExec.ID)

	return stepExec, err
}

// executeCommandStep runs a command step
func (s *WorkflowService) executeCommandStep(ctx context.Context, step *WorkflowStep, exec *WorkflowExecution, stepExec *StepExecution) error {
	if step.CommandTrigger == nil {
		return fmt.Errorf("command step missing trigger")
	}

	args := ""
	if step.CommandArgs != nil {
		args = s.interpolate(*step.CommandArgs, exec.Context)
	}

	result, err := s.commandService.ExecuteCommand(ctx, exec.UserID, *step.CommandTrigger, args, nil)
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("command failed: %s", result.Error)
	}

	stepExec.Output["command"] = *step.CommandTrigger
	stepExec.Output["processed_prompt"] = result.ProcessedPrompt
	stepExec.Output["success"] = result.Success

	return nil
}

// executeAgentStep delegates to an agent
func (s *WorkflowService) executeAgentStep(ctx context.Context, step *WorkflowStep, exec *WorkflowExecution, stepExec *StepExecution) error {
	prompt := ""
	if step.PromptTemplate != nil {
		prompt = s.interpolate(*step.PromptTemplate, exec.Context)
	}

	stepExec.Output["prompt"] = prompt
	stepExec.Output["agent_id"] = step.TargetAgentID
	// Note: Actual agent execution would be done by the caller
	// This just prepares the step for execution

	return nil
}

// executeToolStep executes a tool
func (s *WorkflowService) executeToolStep(ctx context.Context, step *WorkflowStep, exec *WorkflowExecution, stepExec *StepExecution) error {
	if step.ToolName == nil {
		return fmt.Errorf("tool step missing tool name")
	}

	// Interpolate tool params
	params := make(map[string]any)
	for k, v := range step.ToolParams {
		if str, ok := v.(string); ok {
			params[k] = s.interpolate(str, exec.Context)
		} else {
			params[k] = v
		}
	}

	stepExec.Output["tool"] = *step.ToolName
	stepExec.Output["params"] = params
	// Note: Actual tool execution would be done by the caller

	return nil
}

// executeConditionStep evaluates a condition
func (s *WorkflowService) executeConditionStep(ctx context.Context, step *WorkflowStep, exec *WorkflowExecution, stepExec *StepExecution) error {
	if step.ConditionExpression == nil {
		return fmt.Errorf("condition step missing expression")
	}

	result := s.evaluateCondition(*step.ConditionExpression, exec.Context)
	stepExec.Output["condition"] = *step.ConditionExpression
	stepExec.Output["result"] = result

	if result {
		stepExec.Output["next_step"] = step.OnTrueStep
	} else {
		stepExec.Output["next_step"] = step.OnFalseStep
	}

	return nil
}

// interpolate replaces {{variable}} with values from context
func (s *WorkflowService) interpolate(template string, context map[string]any) string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.TrimSpace(match[2 : len(match)-2])
		if val, ok := context[key]; ok {
			switch v := val.(type) {
			case string:
				return v
			default:
				jsonBytes, _ := json.Marshal(v)
				return string(jsonBytes)
			}
		}
		return match
	})
}

// evaluateCondition evaluates a simple condition expression
func (s *WorkflowService) evaluateCondition(expr string, context map[string]any) bool {
	// Simple evaluation: {{var}} == value or {{var}} != value
	expr = strings.TrimSpace(expr)

	// Replace variables
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	evaluated := re.ReplaceAllStringFunc(expr, func(match string) string {
		key := strings.TrimSpace(match[2 : len(match)-2])
		if val, ok := context[key]; ok {
			switch v := val.(type) {
			case bool:
				if v {
					return "true"
				}
				return "false"
			case string:
				return v
			default:
				jsonBytes, _ := json.Marshal(v)
				return string(jsonBytes)
			}
		}
		return "null"
	})

	// Simple comparisons
	if strings.Contains(evaluated, "==") {
		parts := strings.SplitN(evaluated, "==", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) == strings.TrimSpace(parts[1])
		}
	}
	if strings.Contains(evaluated, "!=") {
		parts := strings.SplitN(evaluated, "!=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) != strings.TrimSpace(parts[1])
		}
	}

	// Truthy check
	return evaluated == "true" || evaluated == "1"
}

// GetExecution retrieves an execution by ID
func (s *WorkflowService) GetExecution(ctx context.Context, executionID uuid.UUID) (*WorkflowExecution, error) {
	var exec WorkflowExecution
	var contextJSON, resultJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT id, workflow_id, user_id, conversation_id, initial_input,
			context, status, current_step_id, result, error_message,
			started_at, completed_at, created_at
		FROM workflow_executions WHERE id = $1
	`, executionID).Scan(
		&exec.ID, &exec.WorkflowID, &exec.UserID, &exec.ConversationID,
		&exec.InitialInput, &contextJSON, &exec.Status, &exec.CurrentStepID,
		&resultJSON, &exec.ErrorMessage, &exec.StartedAt, &exec.CompletedAt,
		&exec.CreatedAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(contextJSON, &exec.Context)
	json.Unmarshal(resultJSON, &exec.Result)

	return &exec, nil
}

// ListExecutions lists executions for a user
func (s *WorkflowService) ListExecutions(ctx context.Context, userID string, limit, offset int) ([]WorkflowExecution, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, workflow_id, user_id, conversation_id, initial_input,
			status, current_step_id, error_message, started_at, completed_at, created_at
		FROM workflow_executions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []WorkflowExecution
	for rows.Next() {
		var exec WorkflowExecution
		err := rows.Scan(
			&exec.ID, &exec.WorkflowID, &exec.UserID, &exec.ConversationID,
			&exec.InitialInput, &exec.Status, &exec.CurrentStepID,
			&exec.ErrorMessage, &exec.StartedAt, &exec.CompletedAt, &exec.CreatedAt)
		if err != nil {
			continue
		}
		executions = append(executions, exec)
	}

	return executions, nil
}
