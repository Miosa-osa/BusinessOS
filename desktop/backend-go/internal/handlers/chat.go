package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/prompts"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/tools"
)

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	Message        string            `json:"message" binding:"required"`
	ConversationID *string           `json:"conversation_id"`
	ContextID      *string           `json:"context_id"`       // Legacy: single context ID
	ContextIDs     []string          `json:"context_ids"`      // NEW: Multiple context IDs for tiered context
	ProjectID      *string           `json:"project_id"`
	NodeID         *string           `json:"node_id"`          // NEW: Business node context
	Model          *string           `json:"model"`
	AgentType      *string           `json:"agent_type"`       // orchestrator, document, analysis, planning
	FocusMode      *string           `json:"focus_mode"`       // research, analyze, write, build, general
	FocusOptions   map[string]string `json:"focus_options"`    // depth, output, searchScope, etc.
	Command        *string           `json:"command"`          // slash command: analyze, summarize, explain, etc.
	Temperature    *float64          `json:"temperature"`
	MaxTokens      *int              `json:"max_tokens"`
	TopP           *float64          `json:"top_p"`
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

// SendMessage handles streaming chat message using the agent system
func (h *Handlers) SendMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle slash commands with specialized processing
	if req.Command != nil && *req.Command != "" {
		h.handleSlashCommand(c, user, req)
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Parse optional IDs
	var contextID, projectID, nodeID *uuid.UUID
	var contextIDs []uuid.UUID

	// Legacy single context ID
	if req.ContextID != nil {
		if parsed, err := uuid.Parse(*req.ContextID); err == nil {
			contextID = &parsed
		}
	}

	// NEW: Multiple context IDs for tiered context
	for _, cidStr := range req.ContextIDs {
		if parsed, err := uuid.Parse(cidStr); err == nil {
			contextIDs = append(contextIDs, parsed)
		}
	}

	// If no contextIDs but legacy contextID exists, use it
	if len(contextIDs) == 0 && contextID != nil {
		contextIDs = append(contextIDs, *contextID)
	}

	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = &parsed
		}
	}

	// NEW: Business node ID
	if req.NodeID != nil {
		if parsed, err := uuid.Parse(*req.NodeID); err == nil {
			nodeID = &parsed
		}
	}

	// Get or create conversation
	var conversationID pgtype.UUID
	var convUUID *uuid.UUID
	if req.ConversationID != nil {
		parsed, err := uuid.Parse(*req.ConversationID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
			return
		}
		conversationID = pgtype.UUID{Bytes: parsed, Valid: true}
		convUUID = &parsed
	} else {
		// Create new conversation
		var ctxID pgtype.UUID
		if contextID != nil {
			ctxID = pgtype.UUID{Bytes: *contextID, Valid: true}
		}

		defaultTitle := "New Conversation"
		conv, err := queries.CreateConversation(ctx, sqlc.CreateConversationParams{
			UserID:    user.ID,
			Title:     &defaultTitle,
			ContextID: ctxID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
			return
		}
		conversationID = conv.ID
		parsed := uuid.UUID(conv.ID.Bytes)
		convUUID = &parsed
	}

	// Save user message
	_, err := queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		ConversationID:  conversationID,
		Role:            sqlc.MessageroleUSER,
		Content:         req.Message,
		MessageMetadata: []byte("{}"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	// Get conversation history
	messages, err := queries.ListMessages(ctx, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	// Convert to chat message format
	chatMessages := make([]services.ChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = services.ChatMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	// Apply focus mode prefix to the first user message if focus mode is set
	if req.FocusMode != nil && *req.FocusMode != "" {
		focusPrefix := prompts.FocusModePrefix(*req.FocusMode, req.FocusOptions)
		if focusPrefix != "" && len(chatMessages) > 0 {
			// Find the last user message and prepend the focus prefix
			for i := len(chatMessages) - 1; i >= 0; i-- {
				if chatMessages[i].Role == "user" {
					chatMessages[i].Content = focusPrefix + chatMessages[i].Content
					break
				}
			}
		}
	}

	// Determine which agent to use
	model := h.cfg.DefaultModel
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}

	// Normalize cloud model names - some old preferences may be missing the -cloud suffix
	cloudModelMigrations := map[string]string{
		"qwen3-coder:480b": "qwen3-coder:480b-cloud",
		"qwen3-coder:30b":  "qwen3-coder:30b-cloud",
	}
	if corrected, ok := cloudModelMigrations[model]; ok {
		fmt.Printf("[Chat] Normalizing model name: %s -> %s\n", model, corrected)
		model = corrected
	}

	agentType := agents.AgentTypeOrchestrator // default

	// Focus mode takes precedence over agent type if provided
	if req.FocusMode != nil && *req.FocusMode != "" {
		agentType = agents.GetAgentForFocusMode(*req.FocusMode)
	} else if req.AgentType != nil {
		switch *req.AgentType {
		case "document":
			agentType = agents.AgentTypeDocument
		case "analysis":
			agentType = agents.AgentTypeAnalysis
		case "planning":
			agentType = agents.AgentTypePlanning
		}
	}

	// Check if context has a custom system prompt (use default agent if so)
	customPrompt := false
	if contextID != nil {
		contextDoc, err := queries.GetContext(ctx, sqlc.GetContextParams{
			ID:     pgtype.UUID{Bytes: *contextID, Valid: true},
			UserID: user.ID,
		})
		if err == nil && contextDoc.SystemPromptTemplate != nil && *contextDoc.SystemPromptTemplate != "" {
			customPrompt = true
		}
	}

	// Set streaming headers
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("X-Conversation-Id", uuidToString(conversationID))
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Track timing for usage logging
	startTime := time.Now()
	var agentName string

	// Create agent and stream response
	streamCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var chunks <-chan string
	var errs <-chan error

	// Get provider info for usage logging
	provider := h.cfg.GetActiveProvider()

	// Build LLM options from request
	llmOptions := services.DefaultLLMOptions()
	if req.Temperature != nil {
		llmOptions.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		llmOptions.MaxTokens = *req.MaxTokens
	}
	if req.TopP != nil {
		llmOptions.TopP = *req.TopP
	}

	// Build context for AI - prefer tiered context if available and contexts are selected
	var ragContext string

	// Use tiered context system if available and user has selected contexts or project
	if h.tieredContextService != nil && (len(contextIDs) > 0 || projectID != nil || nodeID != nil) {
		tieredReq := services.TieredContextRequest{
			UserID:     user.ID,
			ContextIDs: contextIDs,
			ProjectID:  projectID,
			NodeID:     nodeID,
		}

		tieredCtx, err := h.tieredContextService.BuildTieredContext(streamCtx, tieredReq)
		if err == nil && tieredCtx != nil {
			// Perform scoped RAG search within selected contexts
			if len(contextIDs) > 0 {
				relevantBlocks, err := h.tieredContextService.ScopedRAGSearch(streamCtx, req.Message, contextIDs, user.ID, 5)
				if err == nil && len(relevantBlocks) > 0 {
					tieredCtx.Level1.RelevantRAG = relevantBlocks
					fmt.Printf("[Chat] Tiered context: scoped RAG found %d relevant blocks\n", len(relevantBlocks))
				}
			}

			ragContext = tieredCtx.FormatForAI()
			fmt.Printf("[Chat] Tiered context built: Level1(project=%v, contexts=%d), Level2(projects=%d, siblings=%d)\n",
				tieredCtx.Level1.Project != nil, len(tieredCtx.Level1.Contexts),
				len(tieredCtx.Level2.OtherProjects), len(tieredCtx.Level2.SiblingContexts))
		}
	} else if h.contextBuilder != nil {
		// Fallback to legacy global RAG search
		hc, err := h.contextBuilder.BuildContext(streamCtx, req.Message, user.ID, 5)
		if err == nil && len(hc.RelevantBlocks) > 0 {
			ragContext = hc.FormatForAI()
			fmt.Printf("[Chat] RAG context retrieved (legacy): %d relevant blocks\n", len(hc.RelevantBlocks))
		}
	}

	if customPrompt && contextID != nil {
		// Use direct LLM service with custom prompt
		contextDoc, _ := queries.GetContext(ctx, sqlc.GetContextParams{
			ID:     pgtype.UUID{Bytes: *contextID, Valid: true},
			UserID: user.ID,
		})
		// Combine RAG context with custom prompt
		systemPrompt := *contextDoc.SystemPromptTemplate
		if ragContext != "" {
			systemPrompt = ragContext + "\n\n---\n\n" + systemPrompt
		}
		llm := services.NewLLMService(h.cfg, model)
		llm.SetOptions(llmOptions)
		chunks, errs = llm.StreamChat(streamCtx, chatMessages, systemPrompt)
		agentName = "custom_prompt"
	} else {
		// Get project context if project is selected
		var projectCtx *agents.ProjectContext
		if projectID != nil {
			project, err := queries.GetProject(ctx, sqlc.GetProjectParams{
				ID:     pgtype.UUID{Bytes: *projectID, Valid: true},
				UserID: user.ID,
			})
			if err == nil {
				desc := ""
				if project.Description != nil {
					desc = *project.Description
				}
				projectCtx = &agents.ProjectContext{
					Name:        project.Name,
					Description: desc,
				}
			}
		}

		// Inject RAG context as a system message if available
		messagesWithRAG := chatMessages
		if ragContext != "" {
			// Prepend RAG context as a system message
			ragMessage := services.ChatMessage{
				Role:    "system",
				Content: ragContext,
			}
			messagesWithRAG = append([]services.ChatMessage{ragMessage}, chatMessages...)
		}

		// Use agent system with context
		agent := agents.GetAgentWithContext(agentType, h.pool, h.cfg, user.ID, convUUID, model, user.Name, projectCtx)
		agent.SetOptions(llmOptions)
		chunks, errs = agent.Run(streamCtx, messagesWithRAG)
		agentName = string(agentType)
	}

	var fullResponse string
	var streamErr error
	usageSent := false
	fmt.Println("[Chat] Starting stream...")

	// Stream the response - usage is sent as part of the stream before closing
	c.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				// Chunks channel closed - send usage data before ending stream
				if !usageSent && streamErr == nil && fullResponse != "" {
					usageSent = true
					endTime := time.Now()
					inputChars := len(req.Message)
					for _, msg := range messages {
						inputChars += len(msg.Content)
					}
					outputChars := len(fullResponse)
					inputTokens := inputChars / 4
					outputTokens := outputChars / 4
					totalTokens := inputTokens + outputTokens
					durationMs := endTime.Sub(startTime).Milliseconds()
					tps := float64(0)
					if durationMs > 0 {
						tps = float64(outputTokens) / (float64(durationMs) / 1000)
					}
					estimatedCost := services.CalculateEstimatedCost(provider, model, inputTokens, outputTokens)

					usageJSON := fmt.Sprintf("\n\n<!--USAGE:{\"input_tokens\":%d,\"output_tokens\":%d,\"total_tokens\":%d,\"duration_ms\":%d,\"tps\":%.1f,\"provider\":\"%s\",\"model\":\"%s\",\"estimated_cost\":%.6f}-->",
						inputTokens, outputTokens, totalTokens, durationMs, tps, provider, model, estimatedCost)
					fmt.Printf("[Chat] Sending usage in stream: %s\n", usageJSON)
					w.Write([]byte(usageJSON))
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				}
				return false
			}
			fullResponse += chunk
			w.Write([]byte(chunk))
			return true
		case err := <-errs:
			if err != nil {
				streamErr = err
				fmt.Printf("[Chat] Error received: %v\n", err)
				w.Write([]byte("\n\n[Error: " + err.Error() + "]"))
			}
			// Also try to send usage on error channel close (if no error)
			if err == nil && !usageSent && fullResponse != "" {
				usageSent = true
				endTime := time.Now()
				inputChars := len(req.Message)
				for _, msg := range messages {
					inputChars += len(msg.Content)
				}
				outputChars := len(fullResponse)
				inputTokens := inputChars / 4
				outputTokens := outputChars / 4
				totalTokens := inputTokens + outputTokens
				durationMs := endTime.Sub(startTime).Milliseconds()
				tps := float64(0)
				if durationMs > 0 {
					tps = float64(outputTokens) / (float64(durationMs) / 1000)
				}
				estimatedCost := services.CalculateEstimatedCost(provider, model, inputTokens, outputTokens)

				usageJSON := fmt.Sprintf("\n\n<!--USAGE:{\"input_tokens\":%d,\"output_tokens\":%d,\"total_tokens\":%d,\"duration_ms\":%d,\"tps\":%.1f,\"provider\":\"%s\",\"model\":\"%s\",\"estimated_cost\":%.6f}-->",
					inputTokens, outputTokens, totalTokens, durationMs, tps, provider, model, estimatedCost)
				fmt.Printf("[Chat] Sending usage on err close: %s\n", usageJSON)
				w.Write([]byte(usageJSON))
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			}
			return false
		case <-streamCtx.Done():
			fmt.Println("[Chat] Context done")
			return false
		}
	})

	// Post-process the response
	if fullResponse != "" {
		// Parse and save any artifacts from the response
		parsed, err := tools.SaveArtifactsFromResponse(ctx, h.pool, user.ID, convUUID, contextID, fullResponse)
		if err == nil && len(parsed.Artifacts) > 0 {
			// Use clean response (with artifact references) for the message
			fullResponse = parsed.CleanResponse
		}

		// Link artifacts to project if provided
		if projectID != nil && len(parsed.Artifacts) > 0 {
			for _, artifactData := range parsed.Artifacts {
				// artifactData.Summary contains the artifact ID after save
				if artifactData.Summary != "" {
					if artifactID, err := uuid.Parse(artifactData.Summary); err == nil {
						queries.LinkArtifact(ctx, sqlc.LinkArtifactParams{
							ID:        pgtype.UUID{Bytes: artifactID, Valid: true},
							ProjectID: pgtype.UUID{Bytes: *projectID, Valid: true},
						})
					}
				}
			}
		}

		// Save assistant message
		queries.CreateMessage(ctx, sqlc.CreateMessageParams{
			ConversationID:  conversationID,
			Role:            sqlc.MessageroleASSISTANT,
			Content:         fullResponse,
			MessageMetadata: []byte("{}"),
		})

		// Update conversation title if it's the first response
		if len(messages) <= 1 {
			title := req.Message
			if len(title) > 50 {
				title = title[:50] + "..."
			}
			queries.UpdateConversation(ctx, sqlc.UpdateConversationParams{
				ID:    conversationID,
				Title: &title,
			})
		}

		// Log usage (estimate tokens from content length - roughly 4 chars per token)
		endTime := time.Now()
		inputChars := len(req.Message)
		for _, msg := range messages {
			inputChars += len(msg.Content)
		}
		outputChars := len(fullResponse)

		// Rough token estimation (4 chars per token is a common approximation)
		inputTokens := inputChars / 4
		outputTokens := outputChars / 4
		totalTokens := inputTokens + outputTokens

		// Calculate estimated cost
		estimatedCost := services.CalculateEstimatedCost(provider, model, inputTokens, outputTokens)

		// Log usage asynchronously to not block response
		go func() {
			usageService := services.NewUsageService(h.pool)
			usageService.LogAIUsage(context.Background(), services.LogAIUsageParams{
				UserID:         user.ID,
				ConversationID: convUUID,
				Provider:       provider,
				Model:          model,
				InputTokens:    inputTokens,
				OutputTokens:   outputTokens,
				TotalTokens:    totalTokens,
				AgentName:      agentName,
				RequestType:    "chat",
				ProjectID:      projectID,
				DurationMs:     int(endTime.Sub(startTime).Milliseconds()),
				StartedAt:      startTime,
				CompletedAt:    endTime,
				EstimatedCost:  estimatedCost,
			})
		}()
	}
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

// DocumentAI handles document writing assistance using the Document agent
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

	// Use Document agent
	agent := agents.NewDocumentAgent(h.pool, h.cfg, user.ID, nil, model)

	messages := []services.ChatMessage{
		{Role: "user", Content: req.Prompt},
	}

	llm := services.NewLLMService(h.cfg, model)
	response, err := llm.ChatComplete(c.Request.Context(), messages, agent.SystemPrompt())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	// Parse and save any artifacts
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

// AnalyzeContent handles data analysis using the Analysis agent
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

	// Use Analysis agent
	agent := agents.NewAnalysisAgent(h.pool, h.cfg, user.ID, nil, model)

	messages := []services.ChatMessage{
		{Role: "user", Content: "Analyze the following content and provide insights:\n\n" + req.Content},
	}

	llm := services.NewLLMService(h.cfg, model)
	response, err := llm.ChatComplete(c.Request.Context(), messages, agent.SystemPrompt())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": response})
}

// ExtractTasks extracts actionable tasks from content
func (h *Handlers) ExtractTasks(c *gin.Context) {
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

	// Use Planning agent for task extraction
	agent := agents.NewPlanningAgent(h.pool, h.cfg, user.ID, nil, model)

	prompt := `Extract actionable tasks from the following content. Return them as a JSON array of objects with "title", "description", and "priority" (high/medium/low) fields.

Content:
` + req.Content + `

Return only valid JSON, no other text.`

	messages := []services.ChatMessage{
		{Role: "user", Content: prompt},
	}

	llm := services.NewLLMService(h.cfg, model)
	response, err := llm.ChatComplete(c.Request.Context(), messages, agent.SystemPrompt())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": response})
}

// CreatePlan creates a strategic plan using the Planning agent
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

	// Use Planning agent
	agent := agents.NewPlanningAgent(h.pool, h.cfg, user.ID, nil, model)

	prompt := "Create a detailed strategic plan for the following goal:\n\n" + req.Goal
	if req.Timeframe != nil {
		prompt += "\n\nTimeframe: " + *req.Timeframe
	}
	prompt += "\n\nInclude milestones, success criteria, and potential risks."

	messages := []services.ChatMessage{
		{Role: "user", Content: prompt},
	}

	llm := services.NewLLMService(h.cfg, model)
	response, err := llm.ChatComplete(c.Request.Context(), messages, agent.SystemPrompt())
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
