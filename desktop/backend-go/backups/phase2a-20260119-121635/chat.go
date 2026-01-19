package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/tools"
)

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	Message        string            `json:"message" binding:"required"`
	ConversationID *string           `json:"conversation_id"`
	ContextID      *string           `json:"context_id"`  // Legacy: single context ID
	ContextIDs     []string          `json:"context_ids"` // NEW: Multiple context IDs for tiered context
	ProjectID      *string           `json:"project_id"`
	NodeID         *string           `json:"node_id"`       // NEW: Business node context
	WorkspaceID    *string           `json:"workspace_id"`  // NEW (Feature 1): Workspace context for role-based permissions
	DocumentIDs    []string          `json:"document_ids"`  // NEW: Attached document IDs for RAG
	Model          *string           `json:"model"`
	AgentType      *string           `json:"agent_type"`    // orchestrator, document, analysis, planning
	FocusMode      *string           `json:"focus_mode"`    // research, analyze, write, build, general
	FocusOptions   map[string]string `json:"focus_options"` // depth, output, searchScope, etc.
	Command        *string           `json:"command"`       // slash command: analyze, summarize, explain, etc.
	Temperature    *float64          `json:"temperature"`
	MaxTokens      *int              `json:"max_tokens"`
	TopP           *float64          `json:"top_p"`
	UseCOT         *bool             `json:"use_cot"` // Enable Chain of Thought with multi-agent coordination
	// Thinking/COT settings
	ThinkingEnabled     *bool   `json:"thinking_enabled"`      // Enable thinking/reasoning display
	ReasoningTemplateID *string `json:"reasoning_template_id"` // Custom reasoning template to use
	SaveThinking        *bool   `json:"save_thinking"`         // Save thinking traces to database
	MaxThinkingTokens   *int    `json:"max_thinking_tokens"`   // Max tokens for thinking
	// Output Style settings
	OutputStyle      *string `json:"output_style"`      // technical, creative, executive, concise
	StructuredOutput *bool   `json:"structured_output"` // If true, backend will return structured Blocks
}

// ListConversations returns all conversations for the current user
// Optional query parameter: context_id to filter by context
func (h *Handlers) ListConversations(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Check for optional context_id filter
	contextIDStr := c.Query("context_id")
	if contextIDStr != "" {
		contextID, err := uuid.Parse(contextIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context_id"})
			return
		}

		// Filter conversations by context
		conversations, err := queries.ListConversationsByContext(c.Request.Context(), sqlc.ListConversationsByContextParams{
			UserID:    user.ID,
			ContextID: pgtype.UUID{Bytes: contextID, Valid: true},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list conversations"})
			return
		}
		c.JSON(http.StatusOK, TransformConversationsByContextRows(conversations))
		return
	}

	// No filter, return all conversations
	conversations, err := queries.ListConversations(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list conversations"})
		return
	}

	c.JSON(http.StatusOK, TransformConversationListRows(conversations))
}

// CreateConversation creates a new conversation
func (h *Handlers) CreateConversation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Title     string  `json:"title"`
		ContextID *string `json:"context_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title == "" {
		req.Title = "New Conversation"
	}

	queries := sqlc.New(h.pool)

	var contextID pgtype.UUID
	if req.ContextID != nil {
		parsed, err := uuid.Parse(*req.ContextID)
		if err == nil {
			contextID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	conversation, err := queries.CreateConversation(c.Request.Context(), sqlc.CreateConversationParams{
		UserID:    user.ID,
		Title:     &req.Title,
		ContextID: contextID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	c.JSON(http.StatusCreated, TransformConversation(conversation))
}

// GetConversation returns a single conversation with messages
func (h *Handlers) GetConversation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	queries := sqlc.New(h.pool)

	conversation, err := queries.GetConversation(c.Request.Context(), sqlc.GetConversationParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	messages, err := queries.ListMessages(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": TransformConversation(conversation),
		"messages":     TransformMessages(messages),
	})
}

// DeleteConversation deletes a conversation
func (h *Handlers) DeleteConversation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteConversation(c.Request.Context(), sqlc.DeleteConversationParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}

// UpdateConversation updates a conversation's title or context
func (h *Handlers) UpdateConversation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req struct {
		Title     *string `json:"title"`
		ContextID *string `json:"context_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Build update params
	var contextID pgtype.UUID
	if req.ContextID != nil {
		if *req.ContextID == "" {
			// Explicitly unlinking - set to NULL
			contextID = pgtype.UUID{Valid: false}
		} else {
			parsed, err := uuid.Parse(*req.ContextID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context_id"})
				return
			}
			contextID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	// Get current conversation to preserve existing values
	current, err := queries.GetConversation(c.Request.Context(), sqlc.GetConversationParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Use provided title or keep existing
	title := current.Title
	if req.Title != nil {
		title = req.Title
	}

	// Use provided contextID or keep existing
	if req.ContextID == nil {
		contextID = current.ContextID
	}

	conversation, err := queries.UpdateConversation(c.Request.Context(), sqlc.UpdateConversationParams{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		Title:     title,
		ContextID: contextID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update conversation"})
		return
	}

	c.JSON(http.StatusOK, TransformConversation(conversation))
}

// SearchConversations searches across conversations
func (h *Handlers) SearchConversations(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
		return
	}

	queries := sqlc.New(h.pool)
	conversations, err := queries.SearchConversations(c.Request.Context(), sqlc.SearchConversationsParams{
		UserID:  user.ID,
		Column2: &query,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

// DocumentAI handles document writing assistance using the Document agent (V2)
func (h *Handlers) DocumentAI(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Prompt    string  `json:"prompt" binding:"required"`
		Model     *string `json:"model"`
		ContextID *string `json:"context_id"`
		ProjectID *string `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model := h.cfg.DefaultModel
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}

	// Parse IDs
	var contextID, projectID *uuid.UUID
	if req.ContextID != nil {
		if parsed, err := uuid.Parse(*req.ContextID); err == nil {
			contextID = &parsed
		}
	}
	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = &parsed
		}
	}

	// Use Document agent V2
	registry := agents.NewAgentRegistryV2(h.pool, h.cfg, h.embeddingService, h.promptPersonalizer)
	agent := registry.GetAgent(agents.AgentTypeV2Document, user.ID, user.Name, nil, nil)
	agent.SetModel(model)

	messages := []services.ChatMessage{
		{Role: "user", Content: req.Prompt},
	}

	llm := services.NewLLMService(h.cfg, model)
	response, err := llm.ChatComplete(c.Request.Context(), messages, agent.GetSystemPrompt())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	parsed, _ := tools.SaveArtifactsFromResponse(c.Request.Context(), h.pool, user.ID, nil, contextID, response)

	// Link to project if provided
	if projectID != nil && len(parsed.Artifacts) > 0 {
		queries := sqlc.New(h.pool)
		for _, artifactData := range parsed.Artifacts {
			if artifactData.Summary != "" {
				if artifactID, err := uuid.Parse(artifactData.Summary); err == nil {
					queries.LinkArtifact(c.Request.Context(), sqlc.LinkArtifactParams{
						ID:        pgtype.UUID{Bytes: artifactID, Valid: true},
						ProjectID: pgtype.UUID{Bytes: *projectID, Valid: true},
					})
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"content":   parsed.CleanResponse,
		"artifacts": parsed.Artifacts,
	})
}

// AnalyzeContent handles data analysis using the Analyst agent (V2)
func (h *Handlers) AnalyzeContent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Content string  `json:"content" binding:"required"`
		Model   *string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model := h.cfg.DefaultModel
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}

	// Use Analyst agent V2
	registry := agents.NewAgentRegistryV2(h.pool, h.cfg, h.embeddingService, h.promptPersonalizer)
	agent := registry.GetAgent(agents.AgentTypeV2Analyst, user.ID, user.Name, nil, nil)
	agent.SetModel(model)

	messages := []services.ChatMessage{
		{Role: "user", Content: "Analyze the following content and provide insights:\n\n" + req.Content},
	}

	llm := services.NewLLMService(h.cfg, model)
	response, err := llm.ChatComplete(c.Request.Context(), messages, agent.GetSystemPrompt())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": response})
}

// ExtractTasks extracts actionable tasks from content using Task agent (V2)
func (h *Handlers) ExtractTasks(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Content         string                   `json:"content"`
		ArtifactContent string                   `json:"artifact_content"`
		ArtifactTitle   string                   `json:"artifact_title"`
		ArtifactType    string                   `json:"artifact_type"`
		Model           *string                  `json:"model"`
		TeamMembers     []map[string]interface{} `json:"team_members"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ExtractTasks] Bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[ExtractTasks] Received: title=%s, type=%s, content_len=%d", req.ArtifactTitle, req.ArtifactType, len(req.ArtifactContent))

	// Use artifact_content if content is empty
	content := req.Content
	if content == "" {
		content = req.ArtifactContent
	}
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content or artifact_content is required"})
		return
	}

	model := h.cfg.DefaultModel
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}

	// Use Task agent V2 for task extraction
	registry := agents.NewAgentRegistryV2(h.pool, h.cfg, h.embeddingService, h.promptPersonalizer)
	agent := registry.GetAgent(agents.AgentTypeV2Task, user.ID, user.Name, nil, nil)
	agent.SetModel(model)

	prompt := fmt.Sprintf(`Extract actionable tasks from the following %s titled "%s".
Return them as a JSON array of objects with "title", "description", and "priority" (high/medium/low) fields.

Focus on concrete, actionable items that can be assigned to team members.

Content:
%s

Return ONLY a valid JSON array, no other text. Example format:
[{"title": "Task name", "description": "What needs to be done", "priority": "high"}]`,
		req.ArtifactType, req.ArtifactTitle, content)

	messages := []services.ChatMessage{
		{Role: "user", Content: prompt},
	}

	llm := services.NewLLMService(h.cfg, model)
	response, err := llm.ChatComplete(c.Request.Context(), messages, agent.GetSystemPrompt())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	// Try to parse the response as JSON array
	var tasks []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &tasks); err != nil {
		// Try to extract JSON from response
		start := strings.Index(response, "[")
		end := strings.LastIndex(response, "]")
		if start >= 0 && end > start {
			jsonStr := response[start : end+1]
			if err := json.Unmarshal([]byte(jsonStr), &tasks); err != nil {
				log.Printf("[ExtractTasks] Failed to parse tasks JSON: %v", err)
				c.JSON(http.StatusOK, gin.H{"tasks": []interface{}{}})
				return
			}
		} else {
			log.Printf("[ExtractTasks] No JSON array found in response")
			c.JSON(http.StatusOK, gin.H{"tasks": []interface{}{}})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// CreatePlan creates a strategic plan using the Project agent (V2)
func (h *Handlers) CreatePlan(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Goal      string  `json:"goal" binding:"required"`
		Timeframe *string `json:"timeframe"`
		Model     *string `json:"model"`
		ProjectID *string `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model := h.cfg.DefaultModel
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}

	// Use Project agent V2 for planning
	registry := agents.NewAgentRegistryV2(h.pool, h.cfg, h.embeddingService, h.promptPersonalizer)
	agent := registry.GetAgent(agents.AgentTypeV2Project, user.ID, user.Name, nil, nil)
	agent.SetModel(model)

	prompt := "Create a detailed strategic plan for the following goal:\n\n" + req.Goal
	if req.Timeframe != nil {
		prompt += "\n\nTimeframe: " + *req.Timeframe
	}
	prompt += "\n\nInclude milestones, success criteria, and potential risks."

	messages := []services.ChatMessage{
		{Role: "user", Content: prompt},
	}

	llm := services.NewLLMService(h.cfg, model)
	response, err := llm.ChatComplete(c.Request.Context(), messages, agent.GetSystemPrompt())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	// Parse and save any artifacts (plans often generate artifacts)
	var projectID *uuid.UUID
	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = &parsed
		}
	}

	parsed, _ := tools.SaveArtifactsFromResponse(c.Request.Context(), h.pool, user.ID, nil, nil, response)

	// Link to project if provided
	if projectID != nil && len(parsed.Artifacts) > 0 {
		queries := sqlc.New(h.pool)
		for _, artifactData := range parsed.Artifacts {
			if artifactData.Summary != "" {
				if artifactID, err := uuid.Parse(artifactData.Summary); err == nil {
					queries.LinkArtifact(c.Request.Context(), sqlc.LinkArtifactParams{
						ID:        pgtype.UUID{Bytes: artifactID, Valid: true},
						ProjectID: pgtype.UUID{Bytes: *projectID, Valid: true},
					})
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"plan":      parsed.CleanResponse,
		"artifacts": parsed.Artifacts,
	})
}

// Helper function to convert pgtype.UUID to string
func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}
