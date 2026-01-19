// Package handlers provides HTTP handlers for BusinessOS.
package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/sorx"
)

// SorxHandler handles Sorx integration endpoints.
type SorxHandler struct {
	sorxService *services.SorxService
	engine      *sorx.Engine
}

// NewSorxHandler creates a new Sorx handler.
func NewSorxHandler(sorxService *services.SorxService, pool *pgxpool.Pool, logger *slog.Logger) *SorxHandler {
	return &SorxHandler{
		sorxService: sorxService,
		engine:      sorx.NewEngine(pool, logger),
	}
}

// ============================================================================
// Credential Endpoints
// ============================================================================

// RequestCredentialTicket handles POST /api/sorx/credential-ticket
// Sorx requests a ticket to retrieve a credential.
func (h *SorxHandler) RequestCredentialTicket(c *gin.Context) {
	var req services.CredentialTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate the request
	if err := h.sorxService.ValidateTicketRequest(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Issue ticket
	ticket, err := h.sorxService.IssueCredentialTicket(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to issue ticket: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"ticket":  ticket,
	})
}

// RedeemCredential handles POST /api/sorx/redeem-credential
// Sorx exchanges a ticket for the encrypted credential.
func (h *SorxHandler) RedeemCredential(c *gin.Context) {
	var ticket services.CredentialTicket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid ticket: " + err.Error(),
		})
		return
	}

	// Redeem the ticket
	response, err := h.sorxService.RedeemTicket(c.Request.Context(), &ticket)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"credential": response,
	})
}

// ============================================================================
// Callback Endpoint
// ============================================================================

// HandleCallback handles POST /api/sorx/callback
// Main callback endpoint for all Sorx skill callbacks.
func (h *SorxHandler) HandleCallback(c *gin.Context) {
	var req services.CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid callback: " + err.Error(),
		})
		return
	}

	// Process the callback
	response, err := h.sorxService.HandleCallback(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================================
// Decision Endpoints
// ============================================================================

// GetPendingDecisions handles GET /api/sorx/decisions
// Returns all pending decisions for the authenticated user.
func (h *SorxHandler) GetPendingDecisions(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	userID := user.ID

	decisions, err := h.sorxService.GetPendingDecisions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get decisions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"decisions": decisions,
		"count":     len(decisions),
	})
}

// GetDecision handles GET /api/sorx/decisions/:id
// Returns a single pending decision.
func (h *SorxHandler) GetDecision(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid decision ID",
		})
		return
	}

	decision, err := h.sorxService.GetPendingDecision(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Decision not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"decision": decision,
	})
}

// RespondToDecisionRequest represents the request body for responding to a decision.
type RespondToDecisionRequest struct {
	Decision string                 `json:"decision" binding:"required"`
	Inputs   map[string]interface{} `json:"inputs"`
	Comment  string                 `json:"comment"`
}

// RespondToDecision handles POST /api/sorx/decisions/:id/respond
// Submits a human's decision response.
func (h *SorxHandler) RespondToDecision(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	userID := user.ID

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid decision ID",
		})
		return
	}

	var req RespondToDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Record the decision
	if err := h.sorxService.RespondToDecision(c.Request.Context(), id, userID, req.Decision, req.Inputs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to record decision: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Decision recorded",
	})
}

// ============================================================================
// Skill Execution Endpoints
// ============================================================================

// TriggerSkillRequest represents a request to trigger a skill.
type TriggerSkillRequest struct {
	SkillID string                 `json:"skill_id" binding:"required"`
	Params  map[string]interface{} `json:"params"`
}

// TriggerSkill handles POST /api/sorx/execute
// Triggers a skill execution from BusinessOS.
func (h *SorxHandler) TriggerSkill(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	userID := user.ID

	var req TriggerSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Execute skill using the Sorx engine
	execution, err := h.engine.ExecuteSkill(c.Request.Context(), sorx.ExecuteRequest{
		SkillID: req.SkillID,
		UserID:  userID,
		Params:  req.Params,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"execution_id": execution.ID,
		"skill_id":     execution.SkillID,
		"status":       execution.Status,
		"message":      "Skill execution started",
	})
}

// GetSkillExecution handles GET /api/sorx/executions/:id
// Returns the status of a skill execution.
func (h *SorxHandler) GetSkillExecution(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid execution ID",
		})
		return
	}

	// Get execution from engine
	execution, ok := h.engine.GetExecution(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Execution not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"execution": gin.H{
			"id":           execution.ID,
			"skill_id":     execution.SkillID,
			"status":       execution.Status,
			"current_step": execution.CurrentStep,
			"params":       execution.Params,
			"result":       execution.Result,
			"error":        execution.Error,
			"step_results": execution.StepResults,
			"started_at":   execution.StartedAt,
			"completed_at": execution.CompletedAt,
		},
	})
}

// ListSkills handles GET /api/sorx/skills
// Returns all available skills.
func (h *SorxHandler) ListSkills(c *gin.Context) {
	skills := h.engine.ListSkills()

	// Convert to response format
	var skillList []gin.H
	for _, skill := range skills {
		skillList = append(skillList, gin.H{
			"id":                    skill.ID,
			"name":                  skill.Name,
			"description":           skill.Description,
			"category":              skill.Category,
			"required_integrations": skill.RequiredIntegrations,
			"steps_count":           len(skill.Steps),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"skills":  skillList,
		"count":   len(skillList),
	})
}

// GetSkill handles GET /api/sorx/skills/:id
// Returns a single skill definition.
func (h *SorxHandler) GetSkill(c *gin.Context) {
	skillID := c.Param("id")

	skills := h.engine.ListSkills()
	for _, skill := range skills {
		if skill.ID == skillID {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"skill":   skill,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Skill not found",
	})
}

// ============================================================================
// Skill Command Endpoints
// ============================================================================

// ListSkillCommands handles GET /api/sorx/commands
// Returns all skill-based commands that can trigger Sorx skills.
func (h *SorxHandler) ListSkillCommands(c *gin.Context) {
	commands := sorx.ListSkillCommands()

	// Enrich with skill availability info
	var enrichedCommands []gin.H
	for _, cmd := range commands {
		// Check if the associated skill exists
		skills := h.engine.ListSkills()
		skillExists := false
		var skillInfo *sorx.SkillDefinition
		for _, skill := range skills {
			if skill.ID == cmd.SkillID {
				skillExists = true
				skillInfo = skill
				break
			}
		}

		enriched := gin.H{
			"name":         cmd.Name,
			"display_name": cmd.DisplayName,
			"description":  cmd.Description,
			"icon":         cmd.Icon,
			"category":     cmd.Category,
			"skill_id":     cmd.SkillID,
			"params":       cmd.Params,
			"skill_exists": skillExists,
		}

		if skillInfo != nil {
			enriched["required_integrations"] = skillInfo.RequiredIntegrations
			enriched["steps_count"] = len(skillInfo.Steps)
		}

		enrichedCommands = append(enrichedCommands, enriched)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"commands": enrichedCommands,
		"count":    len(enrichedCommands),
	})
}

// ExecuteSkillCommandRequest represents a request to execute a skill command.
type ExecuteSkillCommandRequest struct {
	Command string                 `json:"command" binding:"required"`
	Params  map[string]interface{} `json:"params"`
}

// ExecuteSkillCommand handles POST /api/sorx/commands/execute
// Executes a skill command by triggering the associated Sorx skill.
func (h *SorxHandler) ExecuteSkillCommand(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	userID := user.ID

	var req ExecuteSkillCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get the skill command
	cmd, ok := sorx.GetSkillCommand(req.Command)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Unknown skill command: " + req.Command,
		})
		return
	}

	// Execute the associated skill
	execution, err := h.engine.ExecuteSkill(c.Request.Context(), sorx.ExecuteRequest{
		SkillID: cmd.SkillID,
		UserID:  userID,
		Params:  req.Params,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"execution_id": execution.ID,
		"skill_id":     execution.SkillID,
		"command":      req.Command,
		"status":       execution.Status,
		"message":      "Skill command execution started",
	})
}
