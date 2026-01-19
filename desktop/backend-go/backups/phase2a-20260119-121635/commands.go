package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/tools"
)

// CommandInfo contains metadata about a slash command
type CommandInfo struct {
	Name           string   `json:"name"`
	DisplayName    string   `json:"display_name"`
	Description    string   `json:"description"`
	Icon           string   `json:"icon"`
	Category       string   `json:"category"` // general, business, creative
	SystemPrompt   string   `json:"-"`        // Hidden from API response
	ContextSources []string `json:"context_sources"`
}

// Built-in commands with their configurations
var builtInCommands = map[string]CommandInfo{
	// General Commands
	"analyze": {
		Name:        "analyze",
		DisplayName: "Analyze",
		Description: "Analyze content, data, or patterns in context",
		Icon:        "search",
		Category:    "general",
		SystemPrompt: `You are an expert analyst. Your task is to deeply analyze the provided content and context.

ANALYSIS FRAMEWORK:
1. **Overview**: Provide a high-level summary of what you're analyzing
2. **Key Findings**: Identify the most important patterns, trends, or insights
3. **Deep Dive**: Examine specific details that warrant attention
4. **Implications**: What do these findings mean for the user?
5. **Recommendations**: Based on your analysis, what actions should be considered?

Be thorough, objective, and data-driven in your analysis. Support conclusions with evidence from the provided context.`,
		ContextSources: []string{"documents", "conversations", "artifacts"},
	},
	"summarize": {
		Name:        "summarize",
		DisplayName: "Summarize",
		Description: "Create a concise summary of content or context",
		Icon:        "list",
		Category:    "general",
		SystemPrompt: `You are a skilled summarizer. Create clear, concise summaries that capture essential information.

SUMMARY STRUCTURE:
- **Executive Summary**: 2-3 sentence overview
- **Key Points**: Bullet points of the most important information
- **Details**: Brief elaboration on significant items if needed
- **Action Items**: Any tasks or next steps identified

Keep summaries focused and actionable. Prioritize information by relevance and importance.`,
		ContextSources: []string{"documents", "conversations", "artifacts"},
	},
	"explain": {
		Name:        "explain",
		DisplayName: "Explain",
		Description: "Explain concepts, code, or content clearly",
		Icon:        "info",
		Category:    "general",
		SystemPrompt: `You are an expert explainer who makes complex topics accessible.

EXPLANATION APPROACH:
1. Start with a simple, jargon-free overview
2. Build up complexity gradually
3. Use analogies and examples when helpful
4. Anticipate and address common questions
5. Connect to practical applications

Adapt your explanation to the apparent expertise level of the question. For technical topics, include both conceptual understanding and practical details.`,
		ContextSources: []string{"documents"},
	},
	"generate": {
		Name:        "generate",
		DisplayName: "Generate",
		Description: "Generate content based on context and requirements",
		Icon:        "sparkles",
		Category:    "creative",
		SystemPrompt: `You are a versatile content generator. Create high-quality content based on the provided context and requirements.

GENERATION PRINCIPLES:
- Match the tone and style appropriate for the use case
- Ensure accuracy when referencing provided context
- Be creative while staying relevant
- Structure content logically
- Make content actionable when appropriate

Consider the context provided to inform your generation. Reference specific details from the context when relevant.`,
		ContextSources: []string{"documents", "conversations", "artifacts"},
	},
	"review": {
		Name:        "review",
		DisplayName: "Review",
		Description: "Review and provide feedback on content",
		Icon:        "check",
		Category:    "general",
		SystemPrompt: `You are an expert reviewer providing constructive feedback.

REVIEW FRAMEWORK:
1. **Strengths**: What's working well
2. **Areas for Improvement**: Specific, actionable suggestions
3. **Critical Issues**: Any problems that need immediate attention
4. **Recommendations**: Prioritized list of improvements
5. **Summary**: Overall assessment

Be constructive, specific, and balanced in your feedback. Provide examples and explanations for suggestions.`,
		ContextSources: []string{"documents", "artifacts"},
	},
	"brainstorm": {
		Name:        "brainstorm",
		DisplayName: "Brainstorm",
		Description: "Generate creative ideas and possibilities",
		Icon:        "lightbulb",
		Category:    "creative",
		SystemPrompt: `You are a creative brainstorming partner generating innovative ideas.

BRAINSTORMING APPROACH:
1. Generate multiple diverse ideas (aim for 5-10)
2. Include both conventional and unconventional options
3. Build on provided context and constraints
4. Consider different perspectives and approaches
5. Briefly explain the rationale for each idea

Don't filter ideas too early - include creative possibilities even if they seem ambitious. Group related ideas together.`,
		ContextSources: []string{"documents", "conversations"},
	},
	"task": {
		Name:        "task",
		DisplayName: "Create Task",
		Description: "Parse input and create tasks",
		Icon:        "check-square",
		Category:    "general",
		SystemPrompt: `You are a task manager helping to create clear, actionable tasks.

TASK CREATION:
1. Parse the user's input to identify distinct tasks
2. For each task, provide:
   - Clear title (action-oriented, starts with verb)
   - Brief description if needed
   - Priority suggestion (high/medium/low)
   - Any relevant tags or categories
3. Group related tasks together
4. Identify dependencies between tasks if any

Format tasks clearly so they can be easily added to a task management system.`,
		ContextSources: []string{"conversations"},
	},
	"image": {
		Name:        "image",
		DisplayName: "Image Search",
		Description: "Multimodal image search - find images or search with images",
		Icon:        "image",
		Category:    "general",
		SystemPrompt: `You are an image search assistant helping users find and search with images.

IMAGE SEARCH CAPABILITIES:
1. **Search by description**: Find images matching text descriptions
2. **Visual similarity**: Find similar images to uploaded images
3. **Cross-modal search**: Combine text and image queries
4. **Contextual search**: Search within specific projects or contexts

RESPONSE FORMAT:
- Acknowledge the search request
- Explain what type of search will be performed
- Guide the user on how to refine their search
- Suggest relevant filters or options

Note: The actual image search is performed by the multimodal search system. Your role is to help users understand and use the search effectively.`,
		ContextSources: []string{"documents"},
	},

	// Business Commands
	"proposal": {
		Name:        "proposal",
		DisplayName: "Proposal",
		Description: "Generate a professional proposal from context",
		Icon:        "file-text",
		Category:    "business",
		SystemPrompt: `You are an expert proposal writer creating professional business proposals.

PROPOSAL STRUCTURE:
1. **Executive Summary**: Brief overview of the proposal
2. **Understanding**: Demonstrate understanding of the client's needs
3. **Proposed Solution**: Clear description of what you're proposing
4. **Approach/Methodology**: How you'll deliver the solution
5. **Timeline**: Key milestones and deliverables
6. **Investment**: Pricing and terms (if applicable)
7. **Next Steps**: Clear call to action

Use professional language, be specific about deliverables, and reference relevant context from the project/client data.`,
		ContextSources: []string{"documents", "conversations", "artifacts", "clients", "projects"},
	},
	"report": {
		Name:        "report",
		DisplayName: "Report",
		Description: "Create a business report from data and context",
		Icon:        "bar-chart",
		Category:    "business",
		SystemPrompt: `You are a business analyst creating comprehensive reports.

REPORT STRUCTURE:
1. **Title & Date**
2. **Executive Summary**: Key findings and recommendations
3. **Background**: Context and purpose of the report
4. **Methodology**: How data was gathered/analyzed
5. **Findings**: Detailed results with supporting data
6. **Analysis**: Interpretation of findings
7. **Recommendations**: Actionable next steps
8. **Appendix**: Supporting data if needed

Use data from the provided context. Include specific numbers and metrics where available.`,
		ContextSources: []string{"documents", "conversations", "artifacts", "projects"},
	},
	"email": {
		Name:        "email",
		DisplayName: "Email",
		Description: "Draft a professional email based on context",
		Icon:        "mail",
		Category:    "business",
		SystemPrompt: `You are an expert email writer crafting professional communications.

EMAIL PRINCIPLES:
1. Clear subject line that summarizes the email
2. Appropriate greeting based on relationship
3. Concise, well-structured body
4. Clear call to action or next steps
5. Professional closing

Adapt tone based on the context (formal for clients, friendly for team). Reference relevant details from the provided context.`,
		ContextSources: []string{"conversations", "clients", "projects"},
	},
	"meeting": {
		Name:        "meeting",
		DisplayName: "Meeting Notes",
		Description: "Create meeting notes or agenda from context",
		Icon:        "users",
		Category:    "business",
		SystemPrompt: `You are a meeting facilitator creating clear meeting documentation.

For MEETING NOTES:
- Date, attendees, and purpose
- Key discussion points
- Decisions made
- Action items with owners and deadlines
- Next steps

For MEETING AGENDA:
- Meeting objective
- Agenda items with time allocations
- Required preparation
- Expected outcomes

Extract relevant information from the provided context to populate the notes/agenda.`,
		ContextSources: []string{"conversations", "documents", "projects"},
	},
	"timeline": {
		Name:        "timeline",
		DisplayName: "Timeline",
		Description: "Generate a project timeline from tasks and context",
		Icon:        "calendar",
		Category:    "business",
		SystemPrompt: `You are a project planner creating realistic timelines.

TIMELINE CREATION:
1. Identify all tasks/milestones from the context
2. Estimate duration for each item
3. Identify dependencies
4. Create a logical sequence
5. Add buffer time for unexpected delays
6. Highlight critical path items

Present the timeline in a clear format with dates/durations, dependencies noted, and key milestones highlighted.`,
		ContextSources: []string{"tasks", "projects", "documents"},
	},
	"swot": {
		Name:        "swot",
		DisplayName: "SWOT Analysis",
		Description: "Create a SWOT analysis from context",
		Icon:        "grid",
		Category:    "business",
		SystemPrompt: `You are a strategic analyst performing SWOT analysis.

SWOT FRAMEWORK:
**Strengths** (Internal, Positive)
- What advantages exist?
- What is done well?
- What unique resources are available?

**Weaknesses** (Internal, Negative)
- What could be improved?
- What should be avoided?
- What limitations exist?

**Opportunities** (External, Positive)
- What trends could be leveraged?
- What opportunities are emerging?
- What could be done that isn't being done?

**Threats** (External, Negative)
- What obstacles exist?
- What is the competition doing?
- What risks are present?

Provide specific, actionable insights based on the provided context.`,
		ContextSources: []string{"documents", "projects", "clients"},
	},
	"budget": {
		Name:        "budget",
		DisplayName: "Budget Analysis",
		Description: "Analyze and create budget breakdowns",
		Icon:        "dollar-sign",
		Category:    "business",
		SystemPrompt: `You are a financial analyst creating budget analysis.

BUDGET ANALYSIS:
1. **Summary**: Total budget and key allocations
2. **Line Items**: Detailed breakdown of costs
3. **Categories**: Group expenses logically
4. **Comparison**: Actual vs planned if applicable
5. **Recommendations**: Cost optimization opportunities
6. **Projections**: Future budget considerations

Present numbers clearly with totals and percentages. Identify any concerns or opportunities.`,
		ContextSources: []string{"documents", "projects"},
	},
	"contract": {
		Name:        "contract",
		DisplayName: "Contract",
		Description: "Draft contract terms from context",
		Icon:        "file-contract",
		Category:    "business",
		SystemPrompt: `You are a contract specialist drafting clear agreement terms.

CONTRACT SECTIONS:
1. **Parties**: Who is involved
2. **Scope of Work**: What is being provided
3. **Deliverables**: Specific outputs
4. **Timeline**: Key dates and milestones
5. **Terms**: Payment, duration, renewal
6. **Responsibilities**: Each party's obligations
7. **Conditions**: Key terms and conditions

Use clear, professional language. Reference specific details from the project/client context.

Note: This is a draft for review - recommend legal review before finalizing any contract.`,
		ContextSources: []string{"projects", "clients", "documents"},
	},
	"pitch": {
		Name:        "pitch",
		DisplayName: "Pitch",
		Description: "Create pitch deck content from context",
		Icon:        "presentation",
		Category:    "business",
		SystemPrompt: `You are a pitch expert creating compelling presentation content.

PITCH STRUCTURE:
1. **Hook**: Attention-grabbing opening
2. **Problem**: What pain point are you solving?
3. **Solution**: Your unique approach
4. **Value Proposition**: Why choose this solution?
5. **How It Works**: Brief explanation
6. **Traction/Proof**: Evidence of success
7. **Team**: Why you're qualified (if applicable)
8. **Ask**: What do you need?

Create slide-by-slide content with key points and talking notes. Keep each slide focused on one main idea.`,
		ContextSources: []string{"projects", "clients", "documents"},
	},
	"forecast": {
		Name:        "forecast",
		DisplayName: "Forecast",
		Description: "Generate forecasts from historical data",
		Icon:        "trending-up",
		Category:    "business",
		SystemPrompt: `You are a forecasting analyst making data-driven predictions.

FORECAST APPROACH:
1. **Current State**: Summary of historical data
2. **Trends**: Key patterns identified
3. **Assumptions**: What assumptions underlie the forecast
4. **Projections**: Detailed forecasts with ranges
5. **Scenarios**: Best/expected/worst case
6. **Risks**: Factors that could affect accuracy
7. **Recommendations**: Actions based on forecast

Be clear about confidence levels and the basis for projections.`,
		ContextSources: []string{"documents", "projects"},
	},
	"compare": {
		Name:        "compare",
		DisplayName: "Compare",
		Description: "Compare documents, options, or data",
		Icon:        "columns",
		Category:    "general",
		SystemPrompt: `You are an analyst creating comprehensive comparisons.

COMPARISON FRAMEWORK:
1. **Overview**: What is being compared
2. **Criteria**: Key dimensions for comparison
3. **Side-by-Side**: Clear comparison table/list
4. **Analysis**: Key differences and similarities
5. **Recommendation**: Which option is best for what scenario

Present comparisons in a clear, scannable format. Highlight the most important differences.`,
		ContextSources: []string{"documents", "artifacts"},
	},
}

// handleSlashCommand routes to the appropriate command handler
func (h *Handlers) handleSlashCommand(c *gin.Context, user *middleware.BetterAuthUser, req SendMessageRequest) {
	command := strings.ToLower(strings.TrimPrefix(*req.Command, "/"))

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check if it's a built-in command first
	cmdInfo, exists := builtInCommands[command]
	if !exists {
		// Check for custom user command
		customCmd, err := queries.GetUserCommandByName(ctx, sqlc.GetUserCommandByNameParams{
			Name:   command,
			UserID: user.ID,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unknown command: %s", command)})
			return
		}
		// Convert custom command to CommandInfo
		desc := ""
		if customCmd.Description != nil {
			desc = *customCmd.Description
		}
		icon := "sparkles"
		if customCmd.Icon != nil {
			icon = *customCmd.Icon
		}
		cmdInfo = CommandInfo{
			Name:           customCmd.Name,
			DisplayName:    customCmd.DisplayName,
			Description:    desc,
			Icon:           icon,
			Category:       "custom",
			SystemPrompt:   customCmd.SystemPrompt,
			ContextSources: customCmd.ContextSources,
		}
	}

	// Parse optional IDs
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
		// Create new conversation with command-specific title
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

	// Save user message (include command prefix for history)
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

	// Load context data based on command's context sources
	contextBundle := h.loadContextBundle(ctx, queries, user.ID, contextID, projectID, cmdInfo.ContextSources)

	// Build the enhanced prompt with context
	enhancedPrompt := buildCommandPrompt(cmdInfo, req.Message, contextBundle)

	// Get conversation history (excluding the just-added message)
	messages, _ := queries.ListMessages(ctx, conversationID)
	chatMessages := make([]services.ChatMessage, 0, len(messages))
	for i := 0; i < len(messages)-1; i++ { // Exclude the message we just added
		chatMessages = append(chatMessages, services.ChatMessage{
			Role:    string(messages[i].Role),
			Content: messages[i].Content,
		})
	}

	// Add the enhanced user message
	chatMessages = append(chatMessages, services.ChatMessage{
		Role:    "user",
		Content: enhancedPrompt,
	})

	// Determine model
	model := h.cfg.DefaultModel
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}

	// Set streaming headers
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("X-Conversation-Id", uuidToString(conversationID))
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Track timing
	startTime := time.Now()
	provider := h.cfg.GetActiveProvider()

	// Use LLM service with command's system prompt
	llm := services.NewLLMService(h.cfg, model)
	streamCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	chunks, errs := llm.StreamChat(streamCtx, chatMessages, cmdInfo.SystemPrompt)

	var fullResponse string
	var streamErr error
	c.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				return false
			}
			fullResponse += chunk
			w.Write([]byte(chunk))
			return true
		case err := <-errs:
			if err != nil {
				streamErr = err
				w.Write([]byte("\n\n[Error: " + err.Error() + "]"))
			}
			return false
		case <-streamCtx.Done():
			return false
		}
	})

	// Send usage data
	if streamErr == nil && fullResponse != "" {
		endTime := time.Now()
		inputChars := len(enhancedPrompt)
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

		usageJSON := fmt.Sprintf("\n\n<!--USAGE:{\"input_tokens\":%d,\"output_tokens\":%d,\"total_tokens\":%d,\"duration_ms\":%d,\"tps\":%.1f,\"provider\":\"%s\",\"model\":\"%s\",\"estimated_cost\":%.6f,\"command\":\"%s\"}-->",
			inputTokens, outputTokens, totalTokens, durationMs, tps, provider, model, estimatedCost, command)
		c.Writer.Write([]byte(usageJSON))
		// Ensure usage data is flushed to client
		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	// Save response and log usage
	if fullResponse != "" {
		// Parse and save artifacts
		parsed, err := tools.SaveArtifactsFromResponse(ctx, h.pool, user.ID, convUUID, contextID, fullResponse)
		if err == nil && len(parsed.Artifacts) > 0 {
			fullResponse = parsed.CleanResponse
		}

		// Save assistant message
		queries.CreateMessage(ctx, sqlc.CreateMessageParams{
			ConversationID:  conversationID,
			Role:            sqlc.MessageroleASSISTANT,
			Content:         fullResponse,
			MessageMetadata: []byte(fmt.Sprintf(`{"command":"%s"}`, command)),
		})

		// Log usage
		endTime := time.Now()
		inputChars := len(enhancedPrompt)
		for _, msg := range messages {
			inputChars += len(msg.Content)
		}
		inputTokens := inputChars / 4
		outputTokens := len(fullResponse) / 4
		estimatedCost := services.CalculateEstimatedCost(provider, model, inputTokens, outputTokens)

		go func() {
			usageService := services.NewUsageService(h.pool)
			usageService.LogAIUsage(context.Background(), services.LogAIUsageParams{
				UserID:         user.ID,
				ConversationID: convUUID,
				Provider:       provider,
				Model:          model,
				InputTokens:    inputTokens,
				OutputTokens:   outputTokens,
				TotalTokens:    inputTokens + outputTokens,
				AgentName:      "command_" + command,
				RequestType:    "command",
				ProjectID:      projectID,
				DurationMs:     int(endTime.Sub(startTime).Milliseconds()),
				StartedAt:      startTime,
				CompletedAt:    endTime,
				EstimatedCost:  estimatedCost,
			})
		}()
	}
}

// ContextBundle contains all loaded context data for a command
type ContextBundle struct {
	Documents     []ContextDocument     `json:"documents"`
	Conversations []ContextConversation `json:"conversations"`
	Artifacts     []ContextArtifact     `json:"artifacts"`
	Projects      []ContextProject      `json:"projects"`
	Clients       []ContextClient       `json:"clients"`
	Tasks         []ContextTask         `json:"tasks"`
}

type ContextDocument struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type ContextConversation struct {
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Messages int    `json:"messages"`
}

type ContextArtifact struct {
	Title   string `json:"title"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
	Content string `json:"content"`
}

type ContextProject struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type ContextClient struct {
	Name     string `json:"name"`
	Company  string `json:"company"`
	Status   string `json:"status"`
	Industry string `json:"industry"`
}

type ContextTask struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
}

// loadContextBundle loads relevant context data for a command
func (h *Handlers) loadContextBundle(
	ctx context.Context,
	queries *sqlc.Queries,
	userID string,
	contextID *uuid.UUID,
	projectID *uuid.UUID,
	sources []string,
) *ContextBundle {
	bundle := &ContextBundle{}

	for _, source := range sources {
		switch source {
		case "documents":
			if contextID != nil {
				// Load context document content
				contextDoc, err := queries.GetContext(ctx, sqlc.GetContextParams{
					ID:     pgtype.UUID{Bytes: *contextID, Valid: true},
					UserID: userID,
				})
				if err == nil && contextDoc.Name != "" {
					// Extract text content from blocks
					content := extractBlocksContent(contextDoc.Blocks)
					if content != "" {
						docType := "document"
						if contextDoc.Type.Valid {
							docType = string(contextDoc.Type.Contexttype)
						}
						bundle.Documents = append(bundle.Documents, ContextDocument{
							Title:   contextDoc.Name,
							Content: content,
							Type:    docType,
						})
					}
				}
			}

		case "conversations":
			// Load recent conversations in context
			if contextID != nil {
				convs, err := queries.ListConversationsByContext(ctx, sqlc.ListConversationsByContextParams{
					UserID:    userID,
					ContextID: pgtype.UUID{Bytes: *contextID, Valid: true},
				})
				if err == nil {
					limit := len(convs)
					if limit > 5 {
						limit = 5
					}
					for i := 0; i < limit; i++ {
						conv := convs[i]
						title := "Untitled"
						if conv.Title != nil {
							title = *conv.Title
						}
						bundle.Conversations = append(bundle.Conversations, ContextConversation{
							Title:    title,
							Messages: int(conv.MessageCount),
						})
					}
				}
			}

		case "artifacts":
			// Load artifacts linked to context or project
			if contextID != nil {
				artifacts, err := queries.ListArtifacts(ctx, sqlc.ListArtifactsParams{
					UserID:    userID,
					ContextID: pgtype.UUID{Bytes: *contextID, Valid: true},
				})
				if err == nil {
					limit := len(artifacts)
					if limit > 10 {
						limit = 10
					}
					for i := 0; i < limit; i++ {
						a := artifacts[i]
						summary := ""
						if a.Summary != nil {
							summary = *a.Summary
						}
						bundle.Artifacts = append(bundle.Artifacts, ContextArtifact{
							Title:   a.Title,
							Type:    string(a.Type),
							Summary: summary,
							Content: truncateString(a.Content, 2000),
						})
					}
				}
			}

		case "projects":
			// Load project details
			if projectID != nil {
				project, err := queries.GetProject(ctx, sqlc.GetProjectParams{
					ID:     pgtype.UUID{Bytes: *projectID, Valid: true},
					UserID: userID,
				})
				if err == nil {
					desc := ""
					if project.Description != nil {
						desc = *project.Description
					}
					status := "active"
					if project.Status.Valid {
						status = string(project.Status.Projectstatus)
					}
					bundle.Projects = append(bundle.Projects, ContextProject{
						Name:        project.Name,
						Description: desc,
						Status:      status,
					})
				}
			}

		case "clients":
			// Load client details if context has client_id
			if contextID != nil {
				contextDoc, err := queries.GetContext(ctx, sqlc.GetContextParams{
					ID:     pgtype.UUID{Bytes: *contextID, Valid: true},
					UserID: userID,
				})
				if err == nil && contextDoc.ClientID.Valid {
					client, err := queries.GetClient(ctx, sqlc.GetClientParams{
						ID:     contextDoc.ClientID,
						UserID: userID,
					})
					if err == nil {
						industry := ""
						if client.Industry != nil {
							industry = *client.Industry
						}
						status := "unknown"
						if client.Status.Valid {
							status = string(client.Status.Clientstatus)
						}
						clientType := "company"
						if client.Type.Valid {
							clientType = string(client.Type.Clienttype)
						}
						bundle.Clients = append(bundle.Clients, ContextClient{
							Name:     client.Name,
							Company:  clientType, // Use type (company/individual) since Name already contains the client/company name
							Status:   status,
							Industry: industry,
						})
					}
				}
			}

		case "tasks":
			// Load tasks
			if projectID != nil {
				tasks, err := queries.ListTasks(ctx, sqlc.ListTasksParams{
					UserID:    userID,
					ProjectID: pgtype.UUID{Bytes: *projectID, Valid: true},
				})
				if err == nil {
					limit := len(tasks)
					if limit > 20 {
						limit = 20
					}
					for i := 0; i < limit; i++ {
						t := tasks[i]
						desc := ""
						if t.Description != nil {
							desc = *t.Description
						}
						status := "pending"
						if t.Status.Valid {
							status = string(t.Status.Taskstatus)
						}
						priority := "medium"
						if t.Priority.Valid {
							priority = string(t.Priority.Taskpriority)
						}
						bundle.Tasks = append(bundle.Tasks, ContextTask{
							Title:       t.Title,
							Description: desc,
							Priority:    priority,
							Status:      status,
						})
					}
				}
			} else {
				// Get general tasks
				tasks, err := queries.ListTasks(ctx, sqlc.ListTasksParams{
					UserID: userID,
				})
				if err == nil {
					limit := len(tasks)
					if limit > 10 {
						limit = 10
					}
					for i := 0; i < limit; i++ {
						t := tasks[i]
						desc := ""
						if t.Description != nil {
							desc = *t.Description
						}
						status := "pending"
						if t.Status.Valid {
							status = string(t.Status.Taskstatus)
						}
						priority := "medium"
						if t.Priority.Valid {
							priority = string(t.Priority.Taskpriority)
						}
						bundle.Tasks = append(bundle.Tasks, ContextTask{
							Title:       t.Title,
							Description: desc,
							Priority:    priority,
							Status:      status,
						})
					}
				}
			}
		}
	}

	return bundle
}

// buildCommandPrompt creates an enhanced prompt with context
func buildCommandPrompt(cmdInfo CommandInfo, userMessage string, bundle *ContextBundle) string {
	var sb strings.Builder

	sb.WriteString("USER REQUEST:\n")
	sb.WriteString(userMessage)
	sb.WriteString("\n\n")

	// Add context sections
	if len(bundle.Documents) > 0 {
		sb.WriteString("=== RELEVANT DOCUMENTS ===\n")
		for _, doc := range bundle.Documents {
			sb.WriteString(fmt.Sprintf("\n[%s: %s]\n", doc.Type, doc.Title))
			sb.WriteString(truncateString(doc.Content, 4000))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(bundle.Artifacts) > 0 {
		sb.WriteString("=== RELATED ARTIFACTS ===\n")
		for _, a := range bundle.Artifacts {
			sb.WriteString(fmt.Sprintf("\n[%s: %s]\n", a.Type, a.Title))
			if a.Summary != "" {
				sb.WriteString(fmt.Sprintf("Summary: %s\n", a.Summary))
			}
			if a.Content != "" {
				sb.WriteString(truncateString(a.Content, 1000))
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	if len(bundle.Projects) > 0 {
		sb.WriteString("=== PROJECT CONTEXT ===\n")
		for _, p := range bundle.Projects {
			sb.WriteString(fmt.Sprintf("Project: %s\n", p.Name))
			sb.WriteString(fmt.Sprintf("Status: %s\n", p.Status))
			if p.Description != "" {
				sb.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
			}
		}
		sb.WriteString("\n")
	}

	if len(bundle.Clients) > 0 {
		sb.WriteString("=== CLIENT CONTEXT ===\n")
		for _, cl := range bundle.Clients {
			sb.WriteString(fmt.Sprintf("Client: %s\n", cl.Name))
			if cl.Company != "" {
				sb.WriteString(fmt.Sprintf("Company: %s\n", cl.Company))
			}
			sb.WriteString(fmt.Sprintf("Status: %s\n", cl.Status))
			if cl.Industry != "" {
				sb.WriteString(fmt.Sprintf("Industry: %s\n", cl.Industry))
			}
		}
		sb.WriteString("\n")
	}

	if len(bundle.Tasks) > 0 {
		sb.WriteString("=== TASKS ===\n")
		for _, t := range bundle.Tasks {
			sb.WriteString(fmt.Sprintf("- [%s] %s (%s priority)\n", t.Status, t.Title, t.Priority))
		}
		sb.WriteString("\n")
	}

	if len(bundle.Conversations) > 0 {
		sb.WriteString("=== RECENT CONVERSATIONS ===\n")
		for _, conv := range bundle.Conversations {
			sb.WriteString(fmt.Sprintf("- %s (%d messages)\n", conv.Title, conv.Messages))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func extractBlocksContent(blocksJSON []byte) string {
	// Simple extraction - in production, properly parse JSON blocks
	// For now, just return the raw content cleaned up
	content := string(blocksJSON)
	if content == "null" || content == "[]" || content == "" {
		return ""
	}
	// Remove JSON artifacts for simple extraction
	content = strings.ReplaceAll(content, `"type":"paragraph"`, "")
	content = strings.ReplaceAll(content, `"content":[`, "")
	content = strings.ReplaceAll(content, `"text":"`, " ")
	content = strings.ReplaceAll(content, `"}]`, "")
	content = strings.ReplaceAll(content, `[{`, "")
	content = strings.ReplaceAll(content, `}]`, "")
	content = strings.ReplaceAll(content, `},{`, " ")
	content = strings.ReplaceAll(content, `"`, "")
	return strings.TrimSpace(content)
}

// ListCommands returns all available slash commands (built-in + custom)
func (h *Handlers) ListCommands(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Start with built-in commands (include system prompts for customization)
	commands := make([]gin.H, 0, len(builtInCommands)+10)
	for _, cmd := range builtInCommands {
		commands = append(commands, gin.H{
			"name":            cmd.Name,
			"display_name":    cmd.DisplayName,
			"description":     cmd.Description,
			"icon":            cmd.Icon,
			"category":        cmd.Category,
			"context_sources": cmd.ContextSources,
			"system_prompt":   cmd.SystemPrompt,
			"is_custom":       false,
		})
	}

	// Add user's custom commands
	ctx := context.Background()
	queries := sqlc.New(h.pool)
	userCommands, err := queries.ListUserCommands(ctx, user.ID)
	if err == nil {
		for _, cmd := range userCommands {
			desc := ""
			if cmd.Description != nil {
				desc = *cmd.Description
			}
			icon := "sparkles"
			if cmd.Icon != nil {
				icon = *cmd.Icon
			}
			commands = append(commands, gin.H{
				"id":              cmd.ID,
				"name":            cmd.Name,
				"display_name":    cmd.DisplayName,
				"description":     desc,
				"icon":            icon,
				"category":        "custom",
				"context_sources": cmd.ContextSources,
				"system_prompt":   cmd.SystemPrompt,
				"is_custom":       true,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"commands": commands})
}

// CreateUserCommandRequest represents request to create custom command
type CreateUserCommandRequest struct {
	Name           string   `json:"name" binding:"required"`
	DisplayName    string   `json:"display_name" binding:"required"`
	Description    string   `json:"description"`
	Icon           string   `json:"icon"`
	SystemPrompt   string   `json:"system_prompt" binding:"required"`
	ContextSources []string `json:"context_sources"`
}

// CreateUserCommand creates a new custom slash command
func (h *Handlers) CreateUserCommand(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreateUserCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate command name (alphanumeric + hyphens only)
	name := strings.ToLower(strings.TrimSpace(req.Name))
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Command name can only contain lowercase letters, numbers, and hyphens"})
			return
		}
	}

	// Check if name conflicts with built-in command
	if _, exists := builtInCommands[name]; exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot use a built-in command name"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}
	var icon *string
	if req.Icon != "" {
		icon = &req.Icon
	}

	cmd, err := queries.CreateUserCommand(ctx, sqlc.CreateUserCommandParams{
		UserID:         user.ID,
		Name:           name,
		DisplayName:    req.DisplayName,
		Description:    desc,
		Icon:           icon,
		SystemPrompt:   req.SystemPrompt,
		ContextSources: req.ContextSources,
		IsActive:       boolPtr(true),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create command: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"command": cmd})
}

// GetUserCommand retrieves a specific custom command
func (h *Handlers) GetUserCommand(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command ID"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	cmd, err := queries.GetUserCommand(ctx, sqlc.GetUserCommandParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Command not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"command": cmd})
}

// UpdateUserCommandRequest represents request to update custom command
type UpdateUserCommandRequest struct {
	Name           *string  `json:"name"`
	DisplayName    *string  `json:"display_name"`
	Description    *string  `json:"description"`
	Icon           *string  `json:"icon"`
	SystemPrompt   *string  `json:"system_prompt"`
	ContextSources []string `json:"context_sources"`
	IsActive       *bool    `json:"is_active"`
}

// UpdateUserCommand updates an existing custom command
func (h *Handlers) UpdateUserCommand(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command ID"})
		return
	}

	var req UpdateUserCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	// Get existing command to use as defaults
	existing, err := queries.GetUserCommand(ctx, sqlc.GetUserCommandParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Command not found"})
		return
	}

	// Use existing values as defaults
	name := existing.Name
	displayName := existing.DisplayName
	systemPrompt := existing.SystemPrompt
	description := existing.Description
	icon := existing.Icon
	contextSources := existing.ContextSources
	isActive := existing.IsActive // *bool

	// If updating name, validate it
	if req.Name != nil {
		name = strings.ToLower(strings.TrimSpace(*req.Name))
		for _, r := range name {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Command name can only contain lowercase letters, numbers, and hyphens"})
				return
			}
		}
		if _, exists := builtInCommands[name]; exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot use a built-in command name"})
			return
		}
	}
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}
	if req.SystemPrompt != nil {
		systemPrompt = *req.SystemPrompt
	}
	if req.Description != nil {
		description = req.Description
	}
	if req.Icon != nil {
		icon = req.Icon
	}
	if req.ContextSources != nil {
		contextSources = req.ContextSources
	}
	if req.IsActive != nil {
		isActive = req.IsActive
	}

	cmd, err := queries.UpdateUserCommand(ctx, sqlc.UpdateUserCommandParams{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		UserID:         user.ID,
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Icon:           icon,
		SystemPrompt:   systemPrompt,
		ContextSources: contextSources,
		IsActive:       isActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update command: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"command": cmd})
}

// DeleteUserCommand deletes a custom command
func (h *Handlers) DeleteUserCommand(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command ID"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	err = queries.DeleteUserCommand(ctx, sqlc.DeleteUserCommandParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete command"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
