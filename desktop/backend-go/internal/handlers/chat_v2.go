package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
	"github.com/rhl/businessos-backend/internal/tools"
)

// AgentMention represents a parsed @agent mention
type AgentMention struct {
	AgentName string
	StartPos  int
	EndPos    int
}

// parseAgentMentions extracts @agent-name mentions from a message
func parseAgentMentions(message string) []AgentMention {
	var mentions []AgentMention
	mentionPattern := regexp.MustCompile(`@([a-z0-9][a-z0-9-]*[a-z0-9]|[a-z0-9])`)

	matches := mentionPattern.FindAllStringSubmatchIndex(message, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			mentions = append(mentions, AgentMention{
				AgentName: message[match[2]:match[3]],
				StartPos:  match[0],
				EndPos:    match[1],
			})
		}
	}
	return mentions
}

// stripMentions removes @mentions from message for cleaner processing
func stripMentions(message string, mentions []AgentMention) string {
	if len(mentions) == 0 {
		return message
	}
	result := message
	// Remove in reverse order to preserve indices
	for i := len(mentions) - 1; i >= 0; i-- {
		m := mentions[i]
		result = result[:m.StartPos] + result[m.EndPos:]
	}
	return strings.TrimSpace(result)
}

// SendMessageV2 handles chat messages using the new AgentV2 architecture
// This endpoint uses streaming events with artifact detection
func (h *Handlers) SendMessageV2(c *gin.Context) {
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

	// Debug: Log received request parameters
	focusModeStr := "nil"
	if req.FocusMode != nil {
		focusModeStr = *req.FocusMode
	}
	log.Printf("[ChatV2] Request: msg=%q focus=%s cot=%v", req.Message, focusModeStr, req.UseCOT)

	// Handle slash commands - route to specialized processing
	if req.Command != nil && *req.Command != "" {
		h.handleSlashCommandV2(c, user, req)
		return
	}

	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Parse IDs
	var contextID *uuid.UUID
	var projectID *uuid.UUID
	var nodeID *uuid.UUID
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

	// Determine model
	model := h.cfg.DefaultModel
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}
	// Normalize model name (fix common issues like spaces instead of dashes)
	model = normalizeModelName(model)

	// Build LLM options
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
	if req.ThinkingEnabled != nil && *req.ThinkingEnabled {
		llmOptions.ThinkingEnabled = true
		slog.Debug("ChatV2: ThinkingEnabled set from request.ThinkingEnabled")
	}
	// Also enable thinking if use_cot is true (frontend sends this)
	if req.UseCOT != nil && *req.UseCOT {
		llmOptions.ThinkingEnabled = true
		slog.Debug("ChatV2: ThinkingEnabled set from request.UseCOT")
	}

	// Apply reasoning template if thinking is enabled
	var appliedTemplateID *uuid.UUID
	if llmOptions.ThinkingEnabled {
		// First check if a specific template is requested
		if req.ReasoningTemplateID != nil && *req.ReasoningTemplateID != "" {
			if templateUUID, err := uuid.Parse(*req.ReasoningTemplateID); err == nil {
				template, err := queries.GetReasoningTemplate(ctx, sqlc.GetReasoningTemplateParams{
					ID:     pgtype.UUID{Bytes: templateUUID, Valid: true},
					UserID: user.ID,
				})
				if err == nil {
					applyReasoningTemplate(&llmOptions, template)
					appliedTemplateID = &templateUUID
					slog.Debug("ChatV2: Applied requested reasoning template", "name", template.Name)
				}
			}
		} else {
			// Check for user's default template
			defaultTemplate, err := queries.GetDefaultReasoningTemplate(ctx, user.ID)
			if err == nil {
				applyReasoningTemplate(&llmOptions, defaultTemplate)
				if defaultTemplate.ID.Valid {
					templateUUID := defaultTemplate.ID.Bytes
					appliedTemplateID = (*uuid.UUID)(&templateUUID)
				}
				slog.Debug("ChatV2: Applied default reasoning template", "name", defaultTemplate.Name)
			}
		}

		// Increment template usage counter
		if appliedTemplateID != nil {
			go func(templateID uuid.UUID) {
				queries.IncrementTemplateUsage(context.Background(), pgtype.UUID{Bytes: templateID, Valid: true})
			}(*appliedTemplateID)
		}
	}

	// Build tiered context
	var tieredCtx *services.TieredContext
	if h.tieredContextService != nil && (len(contextIDs) > 0 || projectID != nil || nodeID != nil) {
		tieredReq := services.TieredContextRequest{
			UserID:     user.ID,
			ContextIDs: contextIDs,
			ProjectID:  projectID,
			NodeID:     nodeID,
		}
		tieredCtx, _ = h.tieredContextService.BuildTieredContext(ctx, tieredReq)
	}

	// Create AgentV2 registry
	registry := agents.NewAgentRegistryV2(h.pool, h.cfg, h.embeddingService)

	// Check if COT (Chain of Thought) mode is enabled
	useCOT := req.UseCOT != nil && *req.UseCOT

	// Parse @agent mentions from message
	mentions := parseAgentMentions(strings.ToLower(req.Message))
	var customAgent *sqlc.CustomAgent
	var customAgentSystemPrompt string

	log.Printf("[ChatV2] @Agent parsing - message: %q, found %d mentions", req.Message, len(mentions))
	for i, m := range mentions {
		log.Printf("[ChatV2] @Agent mention %d: name=%q pos=%d-%d", i, m.AgentName, m.StartPos, m.EndPos)
	}

	// Try to resolve first @mention to a custom agent
	if len(mentions) > 0 {
		for _, mention := range mentions {
			log.Printf("[ChatV2] Looking up custom agent: name=%q user_id=%v", mention.AgentName, user.ID)
			agent, err := queries.GetCustomAgentByName(ctx, sqlc.GetCustomAgentByNameParams{
				Name:   mention.AgentName,
				UserID: user.ID,
			})
			if err != nil {
				log.Printf("[ChatV2] Agent lookup failed for @%s: %v", mention.AgentName, err)
				continue
			}
			customAgent = &agent
			customAgentSystemPrompt = agent.SystemPrompt
			// Increment usage counter
			go queries.IncrementAgentUsage(context.Background(), agent.ID)
			log.Printf("[ChatV2] Resolved @%s to custom agent: %s (prompt: %d chars)", mention.AgentName, agent.DisplayName, len(customAgentSystemPrompt))
			break
		}
	}

	// Determine agent type using SmartIntentRouter
	router := agents.NewSmartIntentRouter(h.pool, h.cfg)
	intent := router.ClassifyIntent(ctx, chatMessages, tieredCtx)

	// Focus mode can override intent and apply focus-specific settings
	var agentType agents.AgentTypeV2
	var focusSystemPrompt string
	var searchContextText string
	var searchResultCount int
	if req.FocusMode != nil && *req.FocusMode != "" {
		log.Printf("[ChatV2] FocusMode received: %s", *req.FocusMode)

		// Build preflight context with web search if enabled
		focusService := services.NewFocusService(h.pool)
		focusCtx, err := focusService.BuildPreflightContext(ctx, user.ID, *req.FocusMode, req.Message, nil, nil)
		if err == nil {
			// Apply focus mode settings to LLM options
			llmOptions.Temperature = focusCtx.LLMOptions.Temperature
			llmOptions.MaxTokens = focusCtx.LLMOptions.MaxTokens
			if focusCtx.LLMOptions.ThinkingEnabled {
				llmOptions.ThinkingEnabled = true
			}
			// Apply model override from focus mode (if set and not already overridden by request)
			if focusCtx.LLMOptions.Model != nil && *focusCtx.LLMOptions.Model != "" {
				// Only override if request didn't explicitly set a model
				if req.Model == nil || *req.Model == "" {
					model = *focusCtx.LLMOptions.Model
					log.Printf("[ChatV2] Focus mode model override: %s", model)
				}
			}
			// Build focus-specific system prompt
			focusSystemPrompt = focusCtx.SystemPrompt

			// Format search results if available
			if len(focusCtx.SearchContext) > 0 {
				searchContextText = focusService.FormatContextForPrompt(focusCtx)
				searchResultCount = len(focusCtx.SearchContext)
				log.Printf("[ChatV2] Web search returned %d results for focus mode", searchResultCount)
			}

			log.Printf("[ChatV2] Applied focus settings: temp=%.2f, maxTokens=%d, thinking=%v, model=%v, searchResults=%d",
				focusCtx.LLMOptions.Temperature, focusCtx.LLMOptions.MaxTokens, focusCtx.LLMOptions.ThinkingEnabled,
				focusCtx.LLMOptions.Model, len(focusCtx.SearchContext))
		} else {
			// Fallback to just settings if preflight fails
			focusSettings, settingsErr := focusService.GetEffectiveSettings(ctx, user.ID, *req.FocusMode)
			if settingsErr == nil {
				llmOptions.Temperature = focusSettings.Temperature
				llmOptions.MaxTokens = focusSettings.MaxTokens
				if focusSettings.ThinkingEnabled {
					llmOptions.ThinkingEnabled = true
				}
				// Apply model override from focus settings
				if focusSettings.EffectiveModel != nil && *focusSettings.EffectiveModel != "" {
					if req.Model == nil || *req.Model == "" {
						model = *focusSettings.EffectiveModel
						log.Printf("[ChatV2] Focus mode model override (fallback): %s", model)
					}
				}
				focusSystemPrompt = focusSettings.SystemPromptPrefix
			}
		}

		shouldDelegate, targetAgent := agents.ShouldDelegateForFocusMode(*req.FocusMode)
		if shouldDelegate {
			agentType = targetAgent
			log.Printf("[ChatV2] FocusMode delegating to agent: %v", targetAgent)
		} else {
			agentType = intent.TargetAgent
		}
	} else if intent.ShouldDelegate {
		agentType = intent.TargetAgent
	} else {
		agentType = agents.AgentTypeV2Orchestrator
	}

	// Get the agent (for non-COT mode)
	agent := registry.GetAgent(agentType, user.ID, user.Name, convUUID, tieredCtx)
	agent.SetModel(model)
	agent.SetOptions(llmOptions)

	// Apply focus mode system prompt prefix if set
	if focusSystemPrompt != "" {
		// Combine focus prompt with search context if available
		fullFocusPrompt := focusSystemPrompt
		if searchContextText != "" {
			fullFocusPrompt = focusSystemPrompt + "\n\n" + searchContextText
			log.Printf("[ChatV2] Injected search context (%d chars) into focus prompt", len(searchContextText))
		}
		agent.SetFocusModePrompt(fullFocusPrompt)
		log.Printf("[ChatV2] Applied focus mode prompt prefix (%d chars)", len(fullFocusPrompt))
	} else if searchContextText != "" {
		// If no focus prompt but we have search results, still inject them
		agent.SetFocusModePrompt(searchContextText)
		log.Printf("[ChatV2] Injected search context only (%d chars)", len(searchContextText))
	}

	// If custom agent found, override the system prompt
	log.Printf("[ChatV2] Custom agent check: customAgent=%v, promptLen=%d", customAgent != nil, len(customAgentSystemPrompt))
	if customAgent != nil && customAgentSystemPrompt != "" {
		log.Printf("[ChatV2] APPLYING custom prompt: %s", customAgentSystemPrompt[:min(100, len(customAgentSystemPrompt))])
		agent.SetCustomSystemPrompt(customAgentSystemPrompt)
		// Apply custom agent's model preference if set
		if customAgent.ModelPreference != nil && *customAgent.ModelPreference != "" {
			agent.SetModel(*customAgent.ModelPreference)
			model = *customAgent.ModelPreference
		}
		// Apply custom agent's thinking setting
		if customAgent.ThinkingEnabled != nil && *customAgent.ThinkingEnabled {
			llmOptions.ThinkingEnabled = true
			agent.SetOptions(llmOptions)
		}
		log.Printf("[ChatV2] Using custom agent: %s (model: %s, prompt: %d chars)", customAgent.DisplayName, model, len(customAgentSystemPrompt))
	} else {
		log.Printf("[ChatV2] NOT using custom agent - customAgent=%v, promptLen=%d", customAgent != nil, len(customAgentSystemPrompt))
	}

	// Create COT orchestrator if enabled (but NOT when using custom agents)
	// Custom agents have their own system prompts and should not be routed through multi-agent orchestration
	var cotOrchestrator *agents.OrchestratorCOT
	if useCOT && customAgent == nil {
		cotOrchestrator = agents.NewOrchestratorCOT(h.pool, h.cfg, registry)
	} else if customAgent != nil && useCOT {
		log.Printf("[ChatV2] COT mode disabled for custom agent: %s", customAgent.DisplayName)
	}

	// Set streaming headers
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("X-Conversation-Id", uuidToString(conversationID))
	c.Header("X-Agent-Type", string(agentType))
	c.Header("X-Intent-Category", intent.Category)
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Track timing
	startTime := time.Now()
	provider := h.cfg.GetActiveProvider()

	// Create stream context
	streamCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Augment user message with search results for non-Claude models
	isNonClaudeModel := !strings.Contains(strings.ToLower(model), "claude")
	if isNonClaudeModel && searchContextText != "" && len(chatMessages) > 0 {
		log.Printf("[ChatV2] Augmenting for non-Claude model: %s", model)
		for i := len(chatMessages) - 1; i >= 0; i-- {
			if strings.EqualFold(chatMessages[i].Role, "user") {
				augmentedContent := fmt.Sprintf(`Based on web search:

%s

---
Question: %s

INSTRUCTIONS:
1. Provide a comprehensive, detailed answer based on the search results above
2. Be thorough - do NOT stop mid-sentence
3. CRITICAL: You MUST end your response with a "## Sources" section
4. In the Sources section, list ALL sources you referenced as markdown links

Example ending:
## Sources
- [Source Title 1](url1)
- [Source Title 2](url2)`, searchContextText, chatMessages[i].Content)
				chatMessages[i].Content = augmentedContent
				log.Printf("[ChatV2] Augmented with %d chars", len(searchContextText))
				break
			}
		}
	}

	// Build agent input
	input := agents.AgentInput{
		Messages:       chatMessages,
		Context:        tieredCtx,
		FocusMode:      "",
		ConversationID: *convUUID,
		UserID:         user.ID,
		UserName:       user.Name,
	}
	if req.FocusMode != nil {
		input.FocusMode = *req.FocusMode
	}

	// Run agent (with or without COT)
	var events <-chan streaming.StreamEvent
	var errs <-chan error

	if useCOT && cotOrchestrator != nil {
		// Use Chain of Thought orchestration for multi-agent coordination
		events, errs, _ = cotOrchestrator.ProcessWithCOT(streamCtx, input, user.ID, user.Name, convUUID, llmOptions)
		c.Header("X-COT-Enabled", "true")
	} else {
		// Standard single-agent execution with thinking events
		events, errs = agent.Run(streamCtx, input)
	}

	var fullResponse string
	var detectedArtifacts []streaming.Artifact
	var firstTokenReceived bool
	var pendingArtifactType, pendingArtifactTitle string
	var artifactContentStart int = -1

	slog.Debug("ChatV2: Starting stream", "agent", agentType, "intent", intent.Category)

	// Track if we've sent the initial thinking event
	var thinkingEventSent bool

	// Thinking tag parsing state
	var insideThinking bool
	var thinkingStartSent bool
	var thinkingEndSent bool // Prevent duplicate thinking_end events
	var thinkingContent string // Accumulated thinking content for DB storage

	// Stream the response
	c.Stream(func(w io.Writer) bool {
		slog.Debug("ChatV2: Stream callback invoked", "responseLen", len(fullResponse))

		// Send initial thinking event at the start of stream
		if !thinkingEventSent {
			thinkingEventSent = true
			writeSSEEvent(w, streaming.StreamEvent{
				Type: streaming.EventTypeThinking,
				Data: streaming.ThinkingStep{
					Step:      "analyzing",
					Content:   "Processing your request...",
					Agent:     string(agentType),
					Completed: false,
				},
			})

			// Send web search notification if search was performed
			if searchResultCount > 0 {
				writeSSEEvent(w, streaming.StreamEvent{
					Type: streaming.EventTypeThinking,
					Data: streaming.ThinkingStep{
						Step:      "search_complete",
						Content:   fmt.Sprintf("Found %d sources from web search", searchResultCount),
						Agent:     string(agentType),
						Completed: true,
					},
				})
			}
		}

		select {
		case event, ok := <-events:
			if !ok {
				slog.Debug("ChatV2: Stream ended", "responseLen", len(fullResponse))
				// Stream ended - send artifact_complete if we have pending artifact from tool
				if pendingArtifactTitle != "" && artifactContentStart > 0 {
					artifactContent := fullResponse
					if artifactContentStart < len(fullResponse) {
						artifactContent = fullResponse[artifactContentStart:]
					}
					artifactContent = strings.TrimPrefix(artifactContent, "Now write the complete document content below. Everything you write will be saved to the artifact.")
					artifactContent = strings.TrimSpace(artifactContent)

					if len(artifactContent) > 50 {
						writeSSEEvent(w, streaming.StreamEvent{
							Type: streaming.EventTypeArtifactComplete,
							Data: streaming.Artifact{
								Type:    pendingArtifactType,
								Title:   pendingArtifactTitle,
								Content: artifactContent,
							},
						})
						slog.Debug("ChatV2: Artifact complete sent", "title", pendingArtifactTitle, "len", len(artifactContent))
					}
				} else if len(detectedArtifacts) == 0 {
					// Auto-detect artifacts based on content structure and keywords
					if artifact := detectStructuredArtifact(fullResponse, req.Message); artifact != nil {
						writeSSEEvent(w, streaming.StreamEvent{
							Type: streaming.EventTypeArtifactComplete,
							Data: *artifact,
						})
						slog.Debug("ChatV2: Auto-artifact detected", "title", artifact.Title, "type", artifact.Type, "len", len(artifact.Content))
					}
				}
				// Send usage and done
				sendUsageEvent(w, startTime, req.Message, messages, fullResponse, provider, model, len(thinkingContent)/4)
				return false
			}

			// Send thinking completed event on first token
			if !firstTokenReceived {
				firstTokenReceived = true
				writeSSEEvent(w, streaming.StreamEvent{
					Type: streaming.EventTypeThinking,
					Data: streaming.ThinkingStep{
						Step:      "responding",
						Content:   "Generating response...",
						Agent:     string(agentType),
						Completed: true,
					},
				})
			}

			// Process event
			switch event.Type {
			case streaming.EventTypeToken:
				tokenContent := event.Content
				fullResponse += tokenContent

				// Flexible thinking tag parsing using regex
				// Only process thinking tags if we haven't finished thinking yet
				if !thinkingEndSent {
					// Use regex to find thinking tags (matches <think...> variations)
					startRe := regexp.MustCompile(`<think[a-z]*\s*>`)
					endRe := regexp.MustCompile(`</think[a-z]*\s*>`)

					startMatch := startRe.FindStringIndex(fullResponse)
					endMatch := endRe.FindStringIndex(fullResponse)

					foundStart := startMatch != nil
					foundEnd := endMatch != nil

					// Check for thinking start
					if !insideThinking && foundStart && !foundEnd {
						insideThinking = true
						if !thinkingStartSent {
							thinkingStartSent = true
							writeSSEEvent(w, streaming.StreamEvent{
								Type: streaming.EventTypeThinkingStart,
								Data: map[string]interface{}{
									"step":  1,
									"agent": string(agentType),
								},
							})
							startTag := fullResponse[startMatch[0]:startMatch[1]]
							slog.Debug("ChatV2: Thinking started", "tag", startTag)
						}
					}

					// Check for thinking end
					if insideThinking && foundEnd {
						insideThinking = false
						thinkingEndSent = true
						// Extract thinking content between tags
						startTagEnd := startMatch[1]
						endTagStart := endMatch[0]
						if startTagEnd < endTagStart {
							thinkingContent = fullResponse[startTagEnd:endTagStart]
						}
						writeSSEEvent(w, streaming.StreamEvent{
							Type: streaming.EventTypeThinkingEnd,
							Data: map[string]interface{}{
								"step":    1,
								"content": sanitizeContent(thinkingContent),
							},
						})
						slog.Debug("ChatV2: Thinking ended", "chars", len(thinkingContent))
					} else if insideThinking {
						// Send thinking chunk (only the new token, not accumulated)
						writeSSEEvent(w, streaming.StreamEvent{
							Type: streaming.EventTypeThinkingChunk,
							Data: map[string]interface{}{
								"content": sanitizeContent(tokenContent),
								"step":    1,
							},
						})
					}
				}

				// Check for artifact start marker from tool call
				if strings.Contains(fullResponse, "ARTIFACT_START::") && pendingArtifactTitle == "" {
					// Parse the marker: ARTIFACT_START::type::title::message
					if idx := strings.Index(fullResponse, "ARTIFACT_START::"); idx != -1 {
						markerEnd := strings.Index(fullResponse[idx:], "::Now write")
						if markerEnd != -1 {
							marker := fullResponse[idx : idx+markerEnd]
							parts := strings.Split(marker, "::")
							if len(parts) >= 3 {
								pendingArtifactType = parts[1]
								pendingArtifactTitle = parts[2]
								artifactContentStart = len(fullResponse)
								// Send artifact_start event to frontend
								writeSSEEvent(w, streaming.StreamEvent{
									Type: streaming.EventTypeArtifactStart,
									Data: map[string]string{"type": pendingArtifactType, "title": pendingArtifactTitle},
								})
								slog.Debug("ChatV2: Artifact started", "title", pendingArtifactTitle, "type", pendingArtifactType)
							}
						}
					}
				}

				// If we're in artifact mode, don't send tokens to chat - they go to artifact panel
				// Also don't send tokens that are inside thinking tags
				if pendingArtifactTitle == "" && !insideThinking {
					// Only write token to chat if NOT in artifact mode and NOT in thinking mode
					// Skip tokens that contain thinking tags
					if !strings.Contains(tokenContent, "<thinking>") && !strings.Contains(tokenContent, "</thinking>") {
						writeSSEEvent(w, event)
					}
				}
				// When in artifact mode, content goes to panel only (via artifact_complete event)

			case streaming.EventTypeThinkingStart:
				// Thinking started from ArtifactDetector
				if !thinkingStartSent {
					thinkingStartSent = true
					insideThinking = true
					slog.Debug("ChatV2: Thinking started (from detector)")
				}
				writeSSEEvent(w, event)

			case streaming.EventTypeThinkingChunk:
				// Accumulate thinking content for token tracking
				if data, ok := event.Data.(map[string]interface{}); ok {
					if content, ok := data["content"].(string); ok {
						thinkingContent += content
					}
				}
				writeSSEEvent(w, event)

			case streaming.EventTypeThinkingEnd:
				// Thinking ended from ArtifactDetector
				thinkingEndSent = true
				insideThinking = false
				slog.Debug("ChatV2: Thinking ended (from detector)", "chars", len(thinkingContent))
				writeSSEEvent(w, event)

			case streaming.EventTypeArtifactStart:
				writeSSEEvent(w, event)

			case streaming.EventTypeArtifactComplete:
				if artifact, ok := event.Data.(streaming.Artifact); ok {
					detectedArtifacts = append(detectedArtifacts, artifact)
				}
				writeSSEEvent(w, event)

			case streaming.EventTypeArtifactError:
				writeSSEEvent(w, event)

			case streaming.EventTypeDone:
			slog.Debug("ChatV2: EventTypeDone received", "responseLen", len(fullResponse))
			// Send artifact_complete if we have pending artifact from tool
			if pendingArtifactTitle != "" && artifactContentStart > 0 {
				artifactContent := fullResponse
				if artifactContentStart < len(fullResponse) {
					artifactContent = fullResponse[artifactContentStart:]
				}
				artifactContent = strings.TrimPrefix(artifactContent, "Now write the complete document content below. Everything you write will be saved to the artifact.")
				artifactContent = strings.TrimSpace(artifactContent)
				if len(artifactContent) > 50 {
					writeSSEEvent(w, streaming.StreamEvent{
						Type: streaming.EventTypeArtifactComplete,
						Data: streaming.Artifact{
							Type:    pendingArtifactType,
							Title:   pendingArtifactTitle,
							Content: artifactContent,
						},
					})
					slog.Debug("ChatV2: Artifact complete sent on Done", "title", pendingArtifactTitle)
				}
			} else if len(detectedArtifacts) == 0 {
				// Auto-detect artifacts based on content structure
				if artifact := detectStructuredArtifact(fullResponse, req.Message); artifact != nil {
					writeSSEEvent(w, streaming.StreamEvent{
						Type: streaming.EventTypeArtifactComplete,
						Data: *artifact,
					})
					slog.Debug("ChatV2: Auto-artifact detected on Done", "title", artifact.Title, "type", artifact.Type)
				}
			}
			sendUsageEvent(w, startTime, req.Message, messages, fullResponse, provider, model, len(thinkingContent)/4)
			return false

		default:
				writeSSEEvent(w, event)
			}
			return true

		case err := <-errs:
			slog.Debug("ChatV2: Error channel received", "err", err, "responseLen", len(fullResponse))
			if err != nil {
				slog.Error("ChatV2: Error details", "err", err)
				writeSSEEvent(w, streaming.StreamEvent{
					Type:    streaming.EventTypeError,
					Content: err.Error(),
				})
			} else {
				// Stream completed via error channel (nil error = success)
				// Check for artifacts since this is a completion path
				if len(detectedArtifacts) == 0 && len(fullResponse) > 200 {
					slog.Debug("ChatV2: Checking for auto-artifact on error channel", "responseLen", len(fullResponse))
					if artifact := detectStructuredArtifact(fullResponse, req.Message); artifact != nil {
						detectedArtifacts = append(detectedArtifacts, *artifact)
						writeSSEEvent(w, streaming.StreamEvent{
							Type: streaming.EventTypeArtifactComplete,
							Data: *artifact,
						})
						slog.Debug("ChatV2: Auto-artifact detected on error channel", "title", artifact.Title, "type", artifact.Type)
					}
				}
				// Send usage event
				sendUsageEvent(w, startTime, req.Message, messages, fullResponse, provider, model, len(thinkingContent)/4)
			}
			return false

		case <-streamCtx.Done():
			slog.Debug("ChatV2: Context done", "reason", streamCtx.Err(), "responseLen", len(fullResponse))
			return false
		}
	})

	// Post-process: save artifacts and message
	if fullResponse != "" {
		// Strip thinking tags from the response for clean storage
		cleanResponse := stripThinkingTags(fullResponse)

		// Save thinking trace to database if thinking was present
		if thinkingContent != "" && convUUID != nil {
			saveThinkingTrace(ctx, h.pool, user.ID, *convUUID, thinkingContent, model, startTime)
		}

		// Save any detected artifacts
		for _, artifact := range detectedArtifacts {
			tools.CreateArtifact(ctx, h.pool, user.ID, convUUID, contextID, projectID, tools.ArtifactData{
				Type:    artifact.Type,
				Title:   artifact.Title,
				Content: artifact.Content,
			})
		}

		// Create artifact from tool call if pending
		if pendingArtifactTitle != "" && artifactContentStart > 0 {
			// Extract content after the marker
			artifactContent := fullResponse
			if artifactContentStart < len(fullResponse) {
				artifactContent = fullResponse[artifactContentStart:]
			}
			// Clean up the marker from the content
			artifactContent = strings.TrimPrefix(artifactContent, "Now write the complete document content below. Everything you write will be saved to the artifact.")
			artifactContent = strings.TrimSpace(artifactContent)

			if len(artifactContent) > 100 {
				artifact, err := tools.CreateArtifact(ctx, h.pool, user.ID, convUUID, contextID, projectID, tools.ArtifactData{
					Type:    pendingArtifactType,
					Title:   pendingArtifactTitle,
					Content: artifactContent,
				})
				if err == nil && artifact != nil {
					slog.Debug("ChatV2: Created artifact from tool", "title", pendingArtifactTitle, "type", pendingArtifactType, "len", len(artifactContent))
				}
			}
		}

		// Also parse artifacts from response (fallback)
		parsed, err := tools.SaveArtifactsFromResponse(ctx, h.pool, user.ID, convUUID, contextID, cleanResponse)
		if err == nil && len(parsed.Artifacts) > 0 {
			cleanResponse = parsed.CleanResponse

			// Link to project
			if projectID != nil {
				for _, artifactData := range parsed.Artifacts {
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
		}

		// Save assistant message (without thinking tags)
		queries.CreateMessage(ctx, sqlc.CreateMessageParams{
			ConversationID:  conversationID,
			Role:            sqlc.MessageroleASSISTANT,
			Content:         cleanResponse,
			MessageMetadata: []byte("{}"),
		})

		// Update conversation title
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
	}
}

// sanitizeContent replaces problematic Unicode characters with ASCII equivalents
func sanitizeContent(content string) string {
	// Replace Unicode bullet points with ASCII dashes
	content = strings.ReplaceAll(content, "\u2022", "-")  // BULLET
	content = strings.ReplaceAll(content, "\u25CF", "-")  // BLACK CIRCLE
	content = strings.ReplaceAll(content, "\u25CB", "-")  // WHITE CIRCLE
	content = strings.ReplaceAll(content, "\u25E6", "-")  // WHITE BULLET
	content = strings.ReplaceAll(content, "\u25AA", "-")  // BLACK SMALL SQUARE
	content = strings.ReplaceAll(content, "\u25B8", "-")  // BLACK RIGHT-POINTING SMALL TRIANGLE
	content = strings.ReplaceAll(content, "\u25BA", "-")  // BLACK RIGHT-POINTING POINTER
	content = strings.ReplaceAll(content, "\u2023", "-")  // TRIANGULAR BULLET
	content = strings.ReplaceAll(content, "\u2043", "-")  // HYPHEN BULLET
	content = strings.ReplaceAll(content, "\u2013", "-")  // EN DASH
	content = strings.ReplaceAll(content, "\u2014", "-")  // EM DASH
	content = strings.ReplaceAll(content, "\u201C", "\"") // LEFT DOUBLE QUOTATION MARK
	content = strings.ReplaceAll(content, "\u201D", "\"") // RIGHT DOUBLE QUOTATION MARK
	content = strings.ReplaceAll(content, "\u2018", "'")  // LEFT SINGLE QUOTATION MARK
	content = strings.ReplaceAll(content, "\u2019", "'")  // RIGHT SINGLE QUOTATION MARK
	content = strings.ReplaceAll(content, "\u2026", "...") // HORIZONTAL ELLIPSIS
	return content
}

// writeSSEEvent writes a streaming event in SSE format
func writeSSEEvent(w io.Writer, event streaming.StreamEvent) {
	// Sanitize content in the event
	if event.Content != "" {
		event.Content = sanitizeContent(event.Content)
	}
	if str, ok := event.Data.(string); ok {
		event.Data = sanitizeContent(str)
	}
	// Sanitize artifact content
	if artifact, ok := event.Data.(streaming.Artifact); ok {
		artifact.Content = sanitizeContent(artifact.Content)
		artifact.Title = sanitizeContent(artifact.Title)
		event.Data = artifact
	}
	// Sanitize map data (for thinking events)
	if mapData, ok := event.Data.(map[string]interface{}); ok {
		if content, exists := mapData["content"]; exists {
			if contentStr, isStr := content.(string); isStr {
				mapData["content"] = sanitizeContent(contentStr)
				event.Data = mapData
			}
		}
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(data))
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// extractDocumentTitle extracts a title from the document content or user message
func extractDocumentTitle(content string, userMessage string) string {
	// Try to find first heading
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
		if strings.HasPrefix(line, "## ") {
			return strings.TrimPrefix(line, "## ")
		}
	}

	// Fallback: use user message
	title := userMessage
	if len(title) > 60 {
		title = title[:60] + "..."
	}
	return title
}

// detectStructuredArtifact analyzes response content to detect if it should be an artifact
// Works with models that don't follow ```artifact format (like Llama 3.3 70B)
func detectStructuredArtifact(content string, userMessage string) *streaming.Artifact {
	slog.Debug("detectStructuredArtifact: Called", "contentLen", len(content), "messagePreview", userMessage[:min(50, len(userMessage))])

	contentLower := strings.ToLower(content)
	msgLower := strings.ToLower(userMessage)
	lines := strings.Split(content, "\n")

	// Count structural elements
	headingCount := 0
	listItemCount := 0
	numberedListCount := 0
	codeBlockCount := 0
	tableRowCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			headingCount++
		}
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			listItemCount++
		}
		if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && (trimmed[1] == '.' || trimmed[1] == ')') {
			numberedListCount++
		}
		if strings.HasPrefix(trimmed, "```") {
			codeBlockCount++
		}
		if strings.HasPrefix(trimmed, "|") && strings.Contains(trimmed, "|") {
			tableRowCount++
		}
	}

	// Calculate structure score
	structureScore := headingCount*3 + listItemCount + numberedListCount*2 + codeBlockCount*2 + tableRowCount

	// Detect document type based on keywords and structure
	docType := ""

	// Plan detection - lower threshold
	planKeywords := []string{"fase", "phase", "etapa", "step", "milestone", "roadmap", "timeline",
		"cronograma", "plano", "plan", "objetivo", "goal", "meta", "sprint", "iteration",
		"week 1", "week 2", "semana 1", "semana 2", "day 1", "day 2", "dia 1", "dia 2"}
	for _, kw := range planKeywords {
		if strings.Contains(msgLower, kw) || strings.Contains(contentLower, kw) {
			docType = "plan"
			break
		}
	}

	// Proposal detection
	if docType == "" {
		proposalKeywords := []string{"proposal", "proposta", "orçamento", "budget", "escopo", "scope",
			"deliverables", "entregáveis", "investimento", "investment", "pricing", "preço"}
		for _, kw := range proposalKeywords {
			if strings.Contains(msgLower, kw) || strings.Contains(contentLower, kw) {
				docType = "proposal"
				break
			}
		}
	}

	// Report/Analysis detection
	if docType == "" {
		reportKeywords := []string{"analysis", "análise", "report", "relatório", "findings", "conclusões",
			"recommendations", "recomendações", "metrics", "métricas", "results", "resultados", "assessment"}
		for _, kw := range reportKeywords {
			if strings.Contains(msgLower, kw) || strings.Contains(contentLower, kw) {
				docType = "report"
				break
			}
		}
	}

	// Code detection
	if docType == "" && codeBlockCount >= 2 {
		codeKeywords := []string{"code", "código", "implement", "implementar", "function", "função",
			"class", "classe", "component", "componente", "api", "endpoint"}
		for _, kw := range codeKeywords {
			if strings.Contains(msgLower, kw) {
				docType = "code"
				break
			}
		}
	}

	// Document detection (generic)
	if docType == "" {
		docKeywords := []string{"document", "documento", "write", "escrever", "create", "criar",
			"draft", "rascunho", "template", "modelo", "guide", "guia", "manual", "tutorial"}
		for _, kw := range docKeywords {
			if strings.Contains(msgLower, kw) {
				docType = "document"
				break
			}
		}
	}

	// Determine if content qualifies as artifact
	// Criteria: sufficient length + structure OR explicit document type detected
	minLength := 300 // Lower threshold for structured content
	if structureScore >= 5 {
		minLength = 200 // Even lower if well-structured
	}

	shouldCreateArtifact := false

	if docType != "" && len(content) >= minLength {
		shouldCreateArtifact = true
	} else if len(content) >= 500 && structureScore >= 8 {
		// Long, well-structured content even without explicit keywords
		shouldCreateArtifact = true
		docType = "document"
	} else if len(content) >= 800 && headingCount >= 2 {
		// Fallback: long content with multiple sections
		shouldCreateArtifact = true
		docType = "document"
	}

	slog.Debug("detectStructuredArtifact: Analysis", "structureScore", structureScore, "docType", docType, "contentLen", len(content), "shouldCreate", shouldCreateArtifact)

	if !shouldCreateArtifact {
		slog.Debug("detectStructuredArtifact: Returning nil - not creating artifact")
		return nil
	}

	// Extract title
	title := extractDocumentTitle(content, userMessage)
	slog.Debug("detectStructuredArtifact: Creating artifact", "title", title, "type", docType)

	return &streaming.Artifact{
		Type:    docType,
		Title:   title,
		Content: content,
	}
}

// sendUsageEvent sends usage statistics as an SSE event
func sendUsageEvent(w io.Writer, startTime time.Time, userMessage string, messages []sqlc.Message, fullResponse string, provider string, model string, thinkingTokens int) {
	endTime := time.Now()
	inputChars := len(userMessage)
	for _, msg := range messages {
		inputChars += len(msg.Content)
	}
	outputChars := len(fullResponse)
	inputTokens := inputChars / 4
	outputTokens := outputChars / 4
	totalTokens := inputTokens + outputTokens + thinkingTokens
	durationMs := endTime.Sub(startTime).Milliseconds()
	tps := float64(0)
	if durationMs > 0 {
		tps = float64(outputTokens) / (float64(durationMs) / 1000)
	}
	estimatedCost := services.CalculateEstimatedCost(provider, model, inputTokens, outputTokens)

	usageData := map[string]interface{}{
		"input_tokens":    inputTokens,
		"output_tokens":   outputTokens,
		"thinking_tokens": thinkingTokens,
		"total_tokens":    totalTokens,
		"duration_ms":     durationMs,
		"tps":             tps,
		"provider":        provider,
		"model":           model,
		"estimated_cost":  estimatedCost,
	}

	event := streaming.StreamEvent{
		Type: streaming.EventTypeDone,
		Data: usageData,
	}
	writeSSEEvent(w, event)
}

// handleSlashCommandV2 processes slash commands using the V2 architecture
func (h *Handlers) handleSlashCommandV2(c *gin.Context, user *middleware.BetterAuthUser, req SendMessageRequest) {
	command := *req.Command
	ctx := c.Request.Context()
	queries := sqlc.New(h.pool)

	// Map commands to agent types (6 agents per doc)
	commandAgentMap := map[string]agents.AgentTypeV2{
		"analyze":  agents.AgentTypeV2Analyst,
		"analysis": agents.AgentTypeV2Analyst,
		"document": agents.AgentTypeV2Document,
		"write":    agents.AgentTypeV2Document,
		"plan":     agents.AgentTypeV2Project,
		"project":  agents.AgentTypeV2Project,
		"task":     agents.AgentTypeV2Task,
		"tasks":    agents.AgentTypeV2Task,
		"todo":     agents.AgentTypeV2Task,
		"client":   agents.AgentTypeV2Client,
		"crm":      agents.AgentTypeV2Client,
	}

	// Determine agent type from command
	agentType, ok := commandAgentMap[command]
	if !ok {
		agentType = agents.AgentTypeV2Orchestrator
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
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
	registry := agents.NewAgentRegistryV2(h.pool, h.cfg, h.embeddingService)
	agent := registry.GetAgent(agentType, user.ID, user.Name, convUUID, tieredCtx)

	// Set model
	model := h.cfg.DefaultModel
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
			MessageMetadata: []byte("{}"),
		})
	}
}

// normalizeModelName fixes common model name issues
// Maps display names to actual API model IDs
func normalizeModelName(model string) string {
	// Common mappings from display names to API IDs
	modelMappings := map[string]string{
		// Groq models - fix spaces and case issues
		"llama 3.3 70b":           "llama-3.3-70b-versatile",
		"llama 3.3 70b versatile": "llama-3.3-70b-versatile",
		"llama 3.1 70b":           "llama-3.1-70b-versatile",
		"llama 3.1 70b versatile": "llama-3.1-70b-versatile",
		"llama 3.1 8b":            "llama-3.1-8b-instant",
		"llama 3.1 8b instant":    "llama-3.1-8b-instant",
		"llama 3 70b":             "llama3-70b-8192",
		"llama 3 8b":              "llama3-8b-8192",
		"mixtral 8x7b":            "mixtral-8x7b-32768",
		"gemma 2 9b":              "gemma2-9b-it",
		"gemma2 9b":               "gemma2-9b-it",
	}

	// Check for exact match (case-insensitive)
	lowerModel := strings.ToLower(strings.TrimSpace(model))
	if mapped, ok := modelMappings[lowerModel]; ok {
		return mapped
	}

	// Return original if no mapping found
	return model
}

// stripThinkingTags removes <thinking>...</thinking> tags and variations from the response
func stripThinkingTags(content string) string {
	// Use a more flexible regex that matches any tag starting with <think
	re := regexp.MustCompile(`<think[^>]*>[\s\S]*?</think[^>]*>\s*`)
	result := re.ReplaceAllString(content, "")
	return strings.TrimSpace(result)
}

// saveThinkingTrace saves thinking content to the database
func saveThinkingTrace(ctx context.Context, pool *pgxpool.Pool, userID string, conversationID uuid.UUID, thinkingContent string, model string, startTime time.Time) {
	if thinkingContent == "" {
		return
	}

	queries := sqlc.New(pool)

	// Estimate token count (rough approximation)
	thinkingTokens := int32(len(thinkingContent) / 4)
	stepNumber := int32(1)

	// Create thinking trace
	_, err := queries.CreateThinkingTrace(ctx, sqlc.CreateThinkingTraceParams{
		UserID:         userID,
		ConversationID: pgtype.UUID{Bytes: conversationID, Valid: true},
		MessageID:      pgtype.UUID{Valid: false}, // Will be set later if needed
		ThinkingContent: thinkingContent,
		ThinkingType: sqlc.NullThinkingtype{
			Thinkingtype: sqlc.ThinkingtypeAnalysis,
			Valid:        true,
		},
		StepNumber: &stepNumber,
		StartedAt: pgtype.Timestamptz{
			Time:  startTime,
			Valid: true,
		},
		ThinkingTokens:      &thinkingTokens,
		ModelUsed:           &model,
		ReasoningTemplateID: pgtype.UUID{Valid: false},
		Metadata:            []byte("{}"),
	})

	if err != nil {
		slog.Error("ChatV2: Failed to save thinking trace", "err", err)
	} else {
		slog.Debug("ChatV2: Saved thinking trace", "chars", len(thinkingContent), "tokens", thinkingTokens)
	}
}

// applyReasoningTemplate applies a reasoning template to LLM options
func applyReasoningTemplate(opts *services.LLMOptions, template sqlc.ReasoningTemplate) {
	// Apply thinking instruction from template
	if template.ThinkingInstruction != nil && *template.ThinkingInstruction != "" {
		opts.ThinkingInstruction = *template.ThinkingInstruction
		slog.Debug("ChatV2: Applied template thinking instruction", "len", len(*template.ThinkingInstruction))
	}

	// Apply max thinking tokens if set
	if template.MaxThinkingTokens != nil && *template.MaxThinkingTokens > 0 {
		opts.MaxThinkingTokens = int(*template.MaxThinkingTokens)
	}

	// Store template ID for tracing
	if template.ID.Valid {
		templateID := template.ID.Bytes
		opts.ReasoningTemplateID = uuid.UUID(templateID).String()
	}
}
