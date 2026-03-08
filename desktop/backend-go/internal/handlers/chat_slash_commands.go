package handlers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
	"github.com/rhl/businessos-backend/internal/tools"
	"github.com/rhl/businessos-backend/internal/utils"
)

// handleSlashCommandStreaming processes slash commands via the streaming SSE architecture.
func (h *ChatHandler) handleSlashCommandStreaming(c *gin.Context, user *middleware.BetterAuthUser, req SendMessageRequest) {
	command := *req.Command
	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Map commands to agent types (6 agents per doc)
	commandAgentMap := map[string]agents.AgentType{
		"analyze":  agents.AgentTypeAnalyst,
		"analysis": agents.AgentTypeAnalyst,
		"document": agents.AgentTypeDocument,
		"write":    agents.AgentTypeDocument,
		"plan":     agents.AgentTypeProject,
		"project":  agents.AgentTypeProject,
		"task":     agents.AgentTypeTask,
		"tasks":    agents.AgentTypeTask,
		"todo":     agents.AgentTypeTask,
		"client":   agents.AgentTypeClient,
		"crm":      agents.AgentTypeClient,
	}

	// Determine agent type from command
	agentType, ok := commandAgentMap[command]
	if !ok {
		agentType = agents.AgentTypeOrchestrator
	}

	// Parse IDs
	var contextID *uuid.UUID
	var projectID *uuid.UUID
	var contextIDs []uuid.UUID

	if req.ContextID != nil {
		if parsed, err := uuid.Parse(*req.ContextID); err == nil {
			contextID = &parsed
		}
	}
	for _, cidStr := range req.ContextIDs {
		if parsed, err := uuid.Parse(cidStr); err == nil {
			contextIDs = append(contextIDs, parsed)
		}
	}
	if len(contextIDs) == 0 && contextID != nil {
		contextIDs = append(contextIDs, *contextID)
	}
	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = &parsed
		}
	}

	// Get or create conversation
	var conversationID pgtype.UUID
	var convUUID *uuid.UUID
	if req.ConversationID != nil {
		parsed, err := uuid.Parse(*req.ConversationID)
		if err != nil {
			utils.RespondInvalidID(c, slog.Default(), "conversation_id")
			return
		}
		conversationID = pgtype.UUID{Bytes: parsed, Valid: true}
		convUUID = &parsed
	} else {
		var ctxID pgtype.UUID
		if contextID != nil {
			ctxID = pgtype.UUID{Bytes: *contextID, Valid: true}
		}
		title := fmt.Sprintf("/%s: %s", command, truncateString(req.Message, 40))
		conv, err := queries.CreateConversation(ctx, sqlc.CreateConversationParams{
			UserID:    user.ID,
			Title:     &title,
			ContextID: ctxID,
		})
		if err != nil {
			utils.RespondInternalError(c, slog.Default(), "create conversation", err)
			return
		}
		conversationID = conv.ID
		parsed := uuid.UUID(conv.ID.Bytes)
		convUUID = &parsed
	}

	// Save user message with command prefix
	userMessage := fmt.Sprintf("/%s %s", command, req.Message)
	_, err := queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		ConversationID:  conversationID,
		Role:            sqlc.MessageroleUSER,
		Content:         userMessage,
		MessageMetadata: nil,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "save user message", err)
		return
	}

	// Get conversation history
	messages, err := queries.ListMessages(ctx, conversationID)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "get conversation messages", err)
		return
	}

	// Convert to chat messages
	chatMessages := make([]services.ChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = services.ChatMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	// Build tiered context
	var tieredCtx *services.TieredContext
	if h.tieredContextService != nil {
		tieredReq := services.TieredContextRequest{
			UserID:     user.ID,
			ContextIDs: contextIDs,
			ProjectID:  projectID,
		}
		tieredCtx, _ = h.tieredContextService.BuildTieredContext(ctx, tieredReq)
	}

	// Create agent registry and get agent
	registry := agents.NewAgentRegistry(h.pool, h.cfg, h.embeddingService, h.promptPersonalizer, h.signalHints)
	agent := registry.GetAgent(agentType, user.ID, user.Name, convUUID, tieredCtx)

	// Set model - use provider-aware default
	model := h.cfg.GetModelForProvider()
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}
	agent.SetModel(model)

	// Set LLM options
	llmOptions := services.DefaultLLMOptions()
	if req.Temperature != nil {
		llmOptions.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		llmOptions.MaxTokens = *req.MaxTokens
	}
	agent.SetOptions(llmOptions)

	// Inject user onboarding profile context for personalization
	slog.Info("[ChatV2-Slash] Profile injection check", "user_id", user.ID)
	onboardingSvc := services.NewOnboardingService(h.pool, nil, nil, nil)
	userProf, err := onboardingSvc.GetUserProfile(ctx, user.ID)
	if err == nil && userProf != nil {
		profileCtx := services.BuildProfilePrefix(userProf)
		agent.SetProfileContext(profileCtx)
		slog.Info("[ChatV2-Slash] Injected user profile context", "chars", len(profileCtx))
	}

	// Inject role context if workspace_id is provided (Feature 1: Role-based permissions)
	if req.WorkspaceID != nil && *req.WorkspaceID != "" && h.roleContextService != nil {
		workspaceID, err := uuid.Parse(*req.WorkspaceID)
		if err == nil {
			roleCtx, err := h.roleContextService.GetUserRoleContext(ctx, user.ID, workspaceID)
			if err == nil {
				// Use the service method to build role context prompt
				rolePrompt := roleCtx.GetRoleContextPrompt()
				agent.SetRoleContextPrompt(rolePrompt)
				slog.Info("[ChatV2-Slash] Injected role context",
					"role", roleCtx.RoleName, "level", roleCtx.HierarchyLevel, "permissions", len(roleCtx.Permissions))
			}
		}
	}

	// Set streaming headers
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Conversation-Id", uuid.UUID(conversationID.Bytes).String())
	c.Header("X-Agent-Type", string(agentType))
	c.Header("X-Command", command)

	// Track timing
	startTime := time.Now()
	provider := h.cfg.GetActiveProvider()

	// Run agent with streaming
	streamCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	agentInput := agents.AgentInput{
		Messages:       chatMessages,
		Context:        tieredCtx,
		FocusMode:      command,
		ConversationID: *convUUID,
		UserID:         user.ID,
		UserName:       user.Name,
	}

	events, errs := agent.Run(streamCtx, agentInput)

	var fullResponse string

	// Stream response
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-events:
			if !ok {
				// Send usage data (no thinking tokens for slash commands)
				sendUsageEvent(w, startTime, req.Message, messages, fullResponse, provider, model, 0)
				return false
			}

			if event.Type == streaming.EventTypeToken {
				if content, ok := event.Data.(string); ok {
					fullResponse += content
				}
			}

			writeSSEEvent(w, event)
			return true

		case err := <-errs:
			if err != nil {
				errorEvent := streaming.StreamEvent{
					Type: streaming.EventTypeError,
					Data: err.Error(),
				}
				writeSSEEvent(w, errorEvent)
			}
			return false

		case <-streamCtx.Done():
			return false
		}
	})

	// Save assistant response
	if fullResponse != "" {
		// Parse and save artifacts
		parsed, _ := tools.SaveArtifactsFromResponse(ctx, h.pool, user.ID, convUUID, contextID, fullResponse)
		if len(parsed.Artifacts) > 0 {
			fullResponse = parsed.CleanResponse
		}

		queries.CreateMessage(ctx, sqlc.CreateMessageParams{
			ConversationID:  conversationID,
			Role:            sqlc.MessageroleASSISTANT,
			Content:         fullResponse,
			MessageMetadata: nil,
		})
	}
}
