package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
)

// CreateWorkflowRequest represents a request to create a workflow
type CreateWorkflowRequest struct {
	Name           string                        `json:"name" binding:"required"`
	DisplayName    string                        `json:"display_name" binding:"required"`
	Description    string                        `json:"description"`
	Trigger        string                        `json:"trigger" binding:"required"`
	ExecutionMode  services.WorkflowExecutionMode `json:"execution_mode"`
	StopOnFailure  bool                          `json:"stop_on_failure"`
	TimeoutSeconds int                           `json:"timeout_seconds"`
	Steps          []CreateWorkflowStepRequest   `json:"steps"`
}

// CreateWorkflowStepRequest represents a step in a workflow creation request
type CreateWorkflowStepRequest struct {
	Name                string                    `json:"name" binding:"required"`
	Description         string                    `json:"description"`
	StepOrder           int                       `json:"step_order" binding:"required"`
	ActionType          services.StepActionType   `json:"action_type" binding:"required"`
	CommandTrigger      *string                   `json:"command_trigger,omitempty"`
	CommandArgs         *string                   `json:"command_args,omitempty"`
	TargetAgentID       *uuid.UUID                `json:"target_agent_id,omitempty"`
	PromptTemplate      *string                   `json:"prompt_template,omitempty"`
	ToolName            *string                   `json:"tool_name,omitempty"`
	ToolParams          map[string]any            `json:"tool_params,omitempty"`
	ConditionExpression *string                   `json:"condition_expression,omitempty"`
	OnTrueStep          *uuid.UUID                `json:"on_true_step,omitempty"`
	OnFalseStep         *uuid.UUID                `json:"on_false_step,omitempty"`
	WaitSeconds         int                       `json:"wait_seconds,omitempty"`
	DependsOn           []uuid.UUID               `json:"depends_on,omitempty"`
	CanParallel         bool                      `json:"can_parallel"`
	OnFailure           services.StepFailureAction `json:"on_failure"`
	MaxRetries          int                       `json:"max_retries"`
	RetryDelaySeconds   int                       `json:"retry_delay_seconds"`
	InputMapping        map[string]string         `json:"input_mapping,omitempty"`
	OutputKey           *string                   `json:"output_key,omitempty"`
}

// ExecuteWorkflowRequest represents a request to execute a workflow
type ExecuteWorkflowRequest struct {
	Input          string     `json:"input"`
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
}

// ListWorkflows lists all workflows for the user
func (h *Handlers) ListWorkflows(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	workflowService := services.NewWorkflowService(h.pool)
	workflows, err := workflowService.ListWorkflows(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workflows", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workflows": workflows})
}

// GetWorkflow gets a workflow by ID
func (h *Handlers) GetWorkflow(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	workflowService := services.NewWorkflowService(h.pool)
	workflow, err := workflowService.GetWorkflow(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	// Check ownership
	if workflow.UserID != userID && workflow.UserID != "SYSTEM" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// CreateWorkflow creates a new workflow
func (h *Handlers) CreateWorkflow(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Set defaults
	if req.ExecutionMode == "" {
		req.ExecutionMode = services.ExecutionModeSequential
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 300
	}

	workflow := &services.CommandWorkflow{
		UserID:         userID,
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		Trigger:        req.Trigger,
		ExecutionMode:  req.ExecutionMode,
		StopOnFailure:  req.StopOnFailure,
		TimeoutSeconds: req.TimeoutSeconds,
		IsActive:       true,
		IsSystem:       false,
	}

	workflowService := services.NewWorkflowService(h.pool)

	// Create workflow
	err := workflowService.CreateWorkflow(c.Request.Context(), workflow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workflow", "details": err.Error()})
		return
	}

	// Add steps
	for _, stepReq := range req.Steps {
		step := &services.WorkflowStep{
			WorkflowID:          workflow.ID,
			Name:                stepReq.Name,
			Description:         stepReq.Description,
			StepOrder:           stepReq.StepOrder,
			ActionType:          stepReq.ActionType,
			CommandTrigger:      stepReq.CommandTrigger,
			CommandArgs:         stepReq.CommandArgs,
			TargetAgentID:       stepReq.TargetAgentID,
			PromptTemplate:      stepReq.PromptTemplate,
			ToolName:            stepReq.ToolName,
			ToolParams:          stepReq.ToolParams,
			ConditionExpression: stepReq.ConditionExpression,
			OnTrueStep:          stepReq.OnTrueStep,
			OnFalseStep:         stepReq.OnFalseStep,
			WaitSeconds:         stepReq.WaitSeconds,
			DependsOn:           stepReq.DependsOn,
			CanParallel:         stepReq.CanParallel,
			OnFailure:           stepReq.OnFailure,
			MaxRetries:          stepReq.MaxRetries,
			RetryDelaySeconds:   stepReq.RetryDelaySeconds,
			InputMapping:        stepReq.InputMapping,
			OutputKey:           stepReq.OutputKey,
		}

		if step.OnFailure == "" {
			step.OnFailure = services.FailureActionStop
		}

		err := workflowService.AddStep(c.Request.Context(), step)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to add step",
				"step":    stepReq.Name,
				"details": err.Error(),
			})
			return
		}
	}

	// Fetch complete workflow with steps
	complete, _ := workflowService.GetWorkflow(c.Request.Context(), workflow.ID)
	if complete != nil {
		workflow = complete
	}

	c.JSON(http.StatusCreated, workflow)
}

// DeleteWorkflow deletes a workflow
func (h *Handlers) DeleteWorkflow(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	// Delete with ownership check
	result, err := h.pool.Exec(c.Request.Context(), `
		DELETE FROM command_workflows WHERE id = $1 AND user_id = $2 AND is_system = FALSE
	`, workflowID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workflow"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found or cannot be deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow deleted"})
}

// ExecuteWorkflow executes a workflow
func (h *Handlers) ExecuteWorkflow(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	workflowService := services.NewWorkflowService(h.pool)

	workflow, err := workflowService.GetWorkflow(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	// Check access
	if workflow.UserID != userID && workflow.UserID != "SYSTEM" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	execution, err := workflowService.ExecuteWorkflow(c.Request.Context(), workflow, userID, req.Input, req.ConversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "workflow execution failed",
			"details":   err.Error(),
			"execution": execution,
		})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// ExecuteWorkflowByTrigger executes a workflow by trigger
func (h *Handlers) ExecuteWorkflowByTrigger(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	trigger := c.Param("trigger")
	if trigger == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trigger is required"})
		return
	}

	var req ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	workflowService := services.NewWorkflowService(h.pool)

	workflow, err := workflowService.ResolveWorkflow(c.Request.Context(), userID, trigger)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found", "trigger": trigger})
		return
	}

	execution, err := workflowService.ExecuteWorkflow(c.Request.Context(), workflow, userID, req.Input, req.ConversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "workflow execution failed",
			"details":   err.Error(),
			"execution": execution,
		})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// ListWorkflowExecutions lists workflow executions for the user
func (h *Handlers) ListWorkflowExecutions(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	workflowService := services.NewWorkflowService(h.pool)
	executions, err := workflowService.ListExecutions(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list executions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetWorkflowExecution gets a specific execution
func (h *Handlers) GetWorkflowExecution(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	executionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution ID"})
		return
	}

	workflowService := services.NewWorkflowService(h.pool)
	execution, err := workflowService.GetExecution(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	// Check ownership
	if execution.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, execution)
}
