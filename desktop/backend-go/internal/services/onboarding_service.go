package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OnboardingService struct {
	pool      *pgxpool.Pool
	aiService *OnboardingAIService
	validator *OnboardingValidator
}

// OnboardingSession represents an onboarding session
type OnboardingSession struct {
	ID                 uuid.UUID              `json:"id"`
	UserID             string                 `json:"user_id"`
	Status             string                 `json:"status"`
	CurrentStep        string                 `json:"current_step"`
	StepsCompleted     []string               `json:"steps_completed"`
	ExtractedData      map[string]interface{} `json:"extracted_data"`
	LowConfidenceCount int                    `json:"low_confidence_count"`
	FallbackTriggered  bool                   `json:"fallback_triggered"`
	WorkspaceID        *uuid.UUID             `json:"workspace_id,omitempty"`
	StartedAt          time.Time              `json:"started_at"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt          time.Time              `json:"expires_at"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// ConversationMessage represents a message in the conversation
type ConversationMessage struct {
	ID              uuid.UUID              `json:"id"`
	SessionID       uuid.UUID              `json:"session_id"`
	Role            string                 `json:"role"` // "user", "agent", "system"
	Content         string                 `json:"content"`
	ConfidenceScore *float64               `json:"confidence_score,omitempty"`
	ExtractedFields map[string]interface{} `json:"extracted_fields,omitempty"`
	QuestionType    *string                `json:"question_type,omitempty"`
	SequenceNumber  int                    `json:"sequence_number"`
	CreatedAt       time.Time              `json:"created_at"`
}

// ExtractedOnboardingData represents the data extracted from conversation
type ExtractedOnboardingData struct {
	WorkspaceName string   `json:"workspace_name,omitempty"`
	BusinessType  string   `json:"business_type,omitempty"`  // Raw user input
	TeamSize      string   `json:"team_size,omitempty"`      // Raw user input
	Role          string   `json:"role,omitempty"`           // Raw user input
	Challenge     string   `json:"challenge,omitempty"`
	Integrations  []string `json:"integrations,omitempty"`
	// Normalized values for internal use (not stored in DB)
	NormalizedBusinessType string `json:"-"` // Used for logic only
	NormalizedTeamSize     string `json:"-"` // Used for logic only  
	NormalizedRole         string `json:"-"` // Used for logic only
}

// OnboardingStatus check result
type OnboardingStatus struct {
	NeedsOnboarding bool               `json:"needs_onboarding"`
	HasSession      bool               `json:"has_session"`
	Session         *OnboardingSession `json:"session,omitempty"`
	WorkspaceCount  int                `json:"workspace_count"`
}

// SendMessageRequest for sending a message
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// SendMessageResponse from AI
type SendMessageResponse struct {
	Message            ConversationMessage `json:"message"`
	NextStep           string              `json:"next_step"`
	IsComplete         bool                `json:"is_complete"`
	ShouldShowFallback bool                `json:"should_show_fallback"`
	ExtractedData      ExtractedOnboardingData `json:"extracted_data"`
	RecommendedIntegrations []string        `json:"recommended_integrations,omitempty"`
}

// FallbackFormData for manual form submission
type FallbackFormData struct {
	WorkspaceName string   `json:"workspace_name" binding:"required"`
	BusinessType  string   `json:"business_type" binding:"required"`
	TeamSize      string   `json:"team_size"`
	Role          string   `json:"role"`
	Challenge     string   `json:"challenge"`
	Integrations  []string `json:"integrations"`
}

// CompleteOnboardingResponse after finishing
type CompleteOnboardingResponse struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	WorkspaceSlug string    `json:"workspace_slug"`
	RedirectURL   string    `json:"redirect_url"`
}

func NewOnboardingService(pool *pgxpool.Pool, aiService *OnboardingAIService) *OnboardingService {
	return &OnboardingService{
		pool:      pool,
		aiService: aiService,
		validator: NewOnboardingValidator(),
	}
}

// CheckOnboardingStatus checks if user needs onboarding
func (s *OnboardingService) CheckOnboardingStatus(ctx context.Context, userID string) (*OnboardingStatus, error) {
	status := &OnboardingStatus{}

	// Check workspace membership count
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM workspace_members 
		WHERE user_id = $1 AND status = 'active'
	`, userID).Scan(&status.WorkspaceCount)
	if err != nil {
		return nil, fmt.Errorf("check workspace count: %w", err)
	}

	// If user has workspaces, no onboarding needed
	if status.WorkspaceCount > 0 {
		status.NeedsOnboarding = false
		return status, nil
	}

	// Check for existing in-progress session
	session, err := s.GetResumeableSession(ctx, userID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("check session: %w", err)
	}

	status.NeedsOnboarding = true
	if session != nil {
		status.HasSession = true
		status.Session = session
	}

	return status, nil
}

// CreateSession creates a new onboarding session
func (s *OnboardingService) CreateSession(ctx context.Context, userID string) (*OnboardingSession, error) {
	// First, expire/abandon any existing sessions for this user
	_, err := s.pool.Exec(ctx, `
		UPDATE onboarding_sessions 
		SET status = 'abandoned', updated_at = NOW()
		WHERE user_id = $1 AND status = 'in_progress'
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("abandon old sessions: %w", err)
	}

	// Create new session
	var session OnboardingSession
	err = s.pool.QueryRow(ctx, `
		INSERT INTO onboarding_sessions (user_id)
		VALUES ($1)
		RETURNING id, user_id, status, current_step, steps_completed, extracted_data, 
		          low_confidence_count, fallback_triggered, workspace_id, started_at, 
		          completed_at, expires_at, created_at, updated_at
	`, userID).Scan(
		&session.ID, &session.UserID, &session.Status, &session.CurrentStep,
		&session.StepsCompleted, &session.ExtractedData, &session.LowConfidenceCount,
		&session.FallbackTriggered, &session.WorkspaceID, &session.StartedAt,
		&session.CompletedAt, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Add initial agent message
	initialMessage := "Hi! I'm here to help set up your workspace. What's your company called?"
	_, err = s.AddMessage(ctx, session.ID, "agent", initialMessage, nil, onboardingStrPtr("company_name"))
	if err != nil {
		return nil, fmt.Errorf("add initial message: %w", err)
	}

	return &session, nil
}

// GetSession retrieves a session by ID
func (s *OnboardingService) GetSession(ctx context.Context, sessionID uuid.UUID) (*OnboardingSession, error) {
	var session OnboardingSession
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, status, current_step, steps_completed, extracted_data, 
		       low_confidence_count, fallback_triggered, workspace_id, started_at, 
		       completed_at, expires_at, created_at, updated_at
		FROM onboarding_sessions
		WHERE id = $1
	`, sessionID).Scan(
		&session.ID, &session.UserID, &session.Status, &session.CurrentStep,
		&session.StepsCompleted, &session.ExtractedData, &session.LowConfidenceCount,
		&session.FallbackTriggered, &session.WorkspaceID, &session.StartedAt,
		&session.CompletedAt, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &session, nil
}

// GetResumeableSession gets an existing in-progress session for a user
func (s *OnboardingService) GetResumeableSession(ctx context.Context, userID string) (*OnboardingSession, error) {
	var session OnboardingSession
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, status, current_step, steps_completed, extracted_data, 
		       low_confidence_count, fallback_triggered, workspace_id, started_at, 
		       completed_at, expires_at, created_at, updated_at
		FROM onboarding_sessions
		WHERE user_id = $1 AND status = 'in_progress' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(
		&session.ID, &session.UserID, &session.Status, &session.CurrentStep,
		&session.StepsCompleted, &session.ExtractedData, &session.LowConfidenceCount,
		&session.FallbackTriggered, &session.WorkspaceID, &session.StartedAt,
		&session.CompletedAt, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get resumeable session: %w", err)
	}
	return &session, nil
}

// GetSessionWithHistory retrieves a session with its conversation history
func (s *OnboardingService) GetSessionWithHistory(ctx context.Context, sessionID uuid.UUID) (*OnboardingSession, []ConversationMessage, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	messages, err := s.GetConversationHistory(ctx, sessionID, 0) // 0 = all messages
	if err != nil {
		return nil, nil, err
	}

	return session, messages, nil
}

// AbandonSession marks a session as abandoned
func (s *OnboardingService) AbandonSession(ctx context.Context, sessionID uuid.UUID, userID string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE onboarding_sessions
		SET status = 'abandoned', updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'in_progress'
	`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("abandon session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found or already completed")
	}
	return nil
}

// AddMessage adds a message to the conversation
func (s *OnboardingService) AddMessage(ctx context.Context, sessionID uuid.UUID, role, content string, confidenceScore *float64, questionType *string) (*ConversationMessage, error) {
	// Get next sequence number
	var seqNum int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(sequence_number), 0) + 1
		FROM onboarding_conversation_history
		WHERE session_id = $1
	`, sessionID).Scan(&seqNum)
	if err != nil {
		return nil, fmt.Errorf("get sequence number: %w", err)
	}

	var msg ConversationMessage
	err = s.pool.QueryRow(ctx, `
		INSERT INTO onboarding_conversation_history (session_id, role, content, confidence_score, question_type, sequence_number)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, session_id, role, content, confidence_score, extracted_fields, question_type, sequence_number, created_at
	`, sessionID, role, content, confidenceScore, questionType, seqNum).Scan(
		&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.ConfidenceScore,
		&msg.ExtractedFields, &msg.QuestionType, &msg.SequenceNumber, &msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	return &msg, nil
}

// GetConversationHistory retrieves conversation messages
func (s *OnboardingService) GetConversationHistory(ctx context.Context, sessionID uuid.UUID, limit int) ([]ConversationMessage, error) {
	query := `
		SELECT id, session_id, role, content, confidence_score, extracted_fields, question_type, sequence_number, created_at
		FROM onboarding_conversation_history
		WHERE session_id = $1
		ORDER BY sequence_number ASC
	`
	if limit > 0 {
		query = fmt.Sprintf(`
			SELECT * FROM (
				SELECT id, session_id, role, content, confidence_score, extracted_fields, question_type, sequence_number, created_at
				FROM onboarding_conversation_history
				WHERE session_id = $1
				ORDER BY sequence_number DESC
				LIMIT %d
			) sub ORDER BY sequence_number ASC
		`, limit)
	}

	rows, err := s.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get conversation history: %w", err)
	}
	defer rows.Close()

	var messages []ConversationMessage
	for rows.Next() {
		var msg ConversationMessage
		err := rows.Scan(
			&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.ConfidenceScore,
			&msg.ExtractedFields, &msg.QuestionType, &msg.SequenceNumber, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ProcessUserMessage processes a user message and returns AI response
func (s *OnboardingService) ProcessUserMessage(ctx context.Context, sessionID uuid.UUID, userID, content string) (*SendMessageResponse, error) {
	// Get session
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Check if session is still valid
	if session.Status != "in_progress" {
		return nil, fmt.Errorf("session is not in progress")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session has expired")
	}

	// Add user message
	_, err = s.AddMessage(ctx, sessionID, "user", content, nil, nil)
	if err != nil {
		return nil, err
	}

	// Process based on current step (chip selections vs chat)
	response, err := s.processStepResponse(ctx, session, content)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// processStepResponse handles the response based on current step
func (s *OnboardingService) processStepResponse(ctx context.Context, session *OnboardingSession, content string) (*SendMessageResponse, error) {
	response := &SendMessageResponse{}
	extractedData := ExtractedOnboardingData{}

	// Parse existing extracted data
	if session.ExtractedData != nil {
		dataBytes, _ := json.Marshal(session.ExtractedData)
		json.Unmarshal(dataBytes, &extractedData)
	}

	// Get conversation history for AI context
	history, err := s.GetConversationHistory(ctx, session.ID, 10)
	if err != nil {
		history = []ConversationMessage{} // Continue without history if error
	}

	// Convert to AI message format
	var aiHistory []OnboardingChatMessage
	for _, msg := range history {
		role := msg.Role
		if role == "agent" {
			role = "assistant"
		}
		aiHistory = append(aiHistory, OnboardingChatMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Call AI to process the message and get response
	aiResponse, aiErr := s.aiService.ProcessMessage(
		ctx,
		content,
		session.CurrentStep,
		structToMap(extractedData),
		aiHistory,
	)

	// Variables for step processing
	nextStep := session.CurrentStep
	agentMessage := ""
	var validationError *ValidationError

	// Process based on current step with validation
	switch session.CurrentStep {
	case "company_name":
		// Validate company name
		if err := s.validator.ValidateCompanyName(content); err != nil {
			validationError = err
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = fmt.Sprintf("Hmm, %s. Could you try a different name?", err.Message)
			}
		} else {
			extractedData.WorkspaceName = strings.TrimSpace(content)
			nextStep = "business_type"
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "Great name! What kind of work does " + extractedData.WorkspaceName + " do?"
			}
		}

	case "business_type":
		// Normalize for validation but store raw input
		normalized := s.validator.NormalizeBusinessType(content)
		if err := s.validator.ValidateBusinessType(normalized); err != nil {
			validationError = err
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "I didn't quite catch that. Are you an agency, startup, freelancer, consultant, or something else?"
			}
		} else {
			extractedData.BusinessType = strings.TrimSpace(content) // Store raw input
			extractedData.NormalizedBusinessType = normalized       // Keep normalized for logic
			if normalized == "freelance" {
				extractedData.TeamSize = "solo (freelancer)"
				nextStep = "role"
				if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
					agentMessage = aiResponse.AgentMessage
				} else {
					agentMessage = "Solo power! What's your role - founder, consultant, or something else?"
				}
			} else {
				nextStep = "team_size"
				if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
					agentMessage = aiResponse.AgentMessage
				} else {
					agentMessage = "Nice! How big is your team?"
				}
			}
		}

	case "team_size":
		// Normalize for validation but store raw input
		normalized := s.validator.NormalizeTeamSize(content)
		if err := s.validator.ValidateTeamSize(normalized); err != nil {
			validationError = err
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "Could you tell me roughly how many people are on your team? Just you, 2-5, 6-10, 11-50, or more?"
			}
		} else {
			extractedData.TeamSize = strings.TrimSpace(content) // Store raw input
			extractedData.NormalizedTeamSize = normalized       // Keep normalized for logic
			nextStep = "role"
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "Got it! And what's your role in the team?"
			}
		}

	case "role":
		// Store raw input and normalize for internal use
		extractedData.Role = strings.TrimSpace(content)       // Store raw input
		extractedData.NormalizedRole = s.validator.NormalizeRole(content) // Keep normalized for logic
		nextStep = "challenge"
		if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
			agentMessage = aiResponse.AgentMessage
		} else {
			agentMessage = "Awesome! What's the biggest challenge you're hoping to solve with Business OS?"
		}

	case "challenge":
		// Validate challenge
		if err := s.validator.ValidateChallenge(content); err != nil {
			validationError = err
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "Could you tell me a bit more about the challenges you're facing? A sentence or two would help me understand."
			}
		} else {
			extractedData.Challenge = strings.TrimSpace(content)
			nextStep = "integrations"
			response.RecommendedIntegrations = ComputeRecommendations(extractedData.Challenge, extractedData.BusinessType)
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "I hear you! Based on what you've shared, I've got some tool recommendations. Let's connect your favorites!"
			}
			response.IsComplete = false
		}

	case "integrations":
		nextStep = "complete"
		response.IsComplete = true
		agentMessage = "Perfect! Your workspace is ready. Let's get started!"
	}

	// If validation failed, don't advance the step
	if validationError != nil {
		response.Message = ConversationMessage{
			Role:    "agent",
			Content: agentMessage,
		}
		response.NextStep = session.CurrentStep // Stay on current step
		response.ExtractedData = extractedData
		
		// Add agent message for the retry prompt
		_, _ = s.AddMessage(ctx, session.ID, "agent", agentMessage, nil, nil)
		
		return response, nil
	}

	// Update session
	stepsCompleted := append(session.StepsCompleted, session.CurrentStep)
	extractedDataMap := structToMap(extractedData)

	_, err = s.pool.Exec(ctx, `
		UPDATE onboarding_sessions
		SET current_step = $1, steps_completed = $2, extracted_data = $3, updated_at = NOW()
		WHERE id = $4
	`, nextStep, stepsCompleted, extractedDataMap, session.ID)
	if err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	// Add agent message
	if agentMessage != "" {
		msg, err := s.AddMessage(ctx, session.ID, "agent", agentMessage, nil, &nextStep)
		if err != nil {
			return nil, err
		}
		response.Message = *msg
	}

	response.NextStep = nextStep
	response.ExtractedData = extractedData

	return response, nil
}

// CompleteOnboarding completes the onboarding and creates the workspace
func (s *OnboardingService) CompleteOnboarding(ctx context.Context, sessionID uuid.UUID, userID string, integrations []string) (*CompleteOnboardingResponse, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Parse extracted data
	var extractedData ExtractedOnboardingData
	if session.ExtractedData != nil {
		dataBytes, _ := json.Marshal(session.ExtractedData)
		json.Unmarshal(dataBytes, &extractedData)
	}

	// Add integrations
	extractedData.Integrations = integrations

	// Apply defaults for missing required fields (allows "skip" flow)
	if extractedData.WorkspaceName == "" {
		extractedData.WorkspaceName = "My Workspace"
	}
	if extractedData.BusinessType == "" {
		extractedData.BusinessType = "other"
	}
	if extractedData.TeamSize == "" {
		extractedData.TeamSize = "solo"
	}

	// Validate integrations if provided
	if len(integrations) > 0 {
		if err := s.validator.ValidateIntegrations(integrations); err != nil {
			return nil, fmt.Errorf("invalid integrations: %s", err.Message)
		}
	}

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create workspace
	workspaceName := extractedData.WorkspaceName
	if workspaceName == "" {
		workspaceName = "My Workspace"
	}
	slug := generateSlugFromName(workspaceName)

	var workspace struct {
		ID   uuid.UUID
		Name string
		Slug string
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, owner_id, onboarding_completed_at, onboarding_data)
		VALUES ($1, $2, $3, NOW(), $4)
		RETURNING id, name, slug
	`, workspaceName, slug, userID, structToMap(extractedData)).Scan(&workspace.ID, &workspace.Name, &workspace.Slug)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	// Seed default roles
	_, err = tx.Exec(ctx, "SELECT seed_default_workspace_roles($1)", workspace.ID)
	if err != nil {
		// Try without the function if it doesn't exist
		_, err = tx.Exec(ctx, `
			INSERT INTO workspace_roles (workspace_id, name, display_name, is_system, hierarchy_level, permissions)
			VALUES 
				($1, 'owner', 'Owner', true, 100, '{"all": true}'::jsonb),
				($1, 'admin', 'Admin', true, 80, '{"manage_members": true, "manage_settings": true}'::jsonb),
				($1, 'member', 'Member', true, 50, '{"read": true, "write": true}'::jsonb)
			ON CONFLICT DO NOTHING
		`, workspace.ID)
		if err != nil {
			return nil, fmt.Errorf("seed roles: %w", err)
		}
	}

	// Add owner as first member
	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role_name, status, joined_at)
		VALUES ($1, $2, 'owner', 'active', NOW())
	`, workspace.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("add owner: %w", err)
	}

	// Create onboarding profile
	recommendations := ComputeRecommendations(extractedData.Challenge, extractedData.BusinessType)
	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_onboarding_profiles 
			(workspace_id, business_type, team_size, owner_role, main_challenge, recommended_integrations, onboarding_session_id, onboarding_method)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'conversational')
	`, workspace.ID, extractedData.BusinessType, extractedData.TeamSize, extractedData.Role, extractedData.Challenge, recommendations, session.ID)
	if err != nil {
		return nil, fmt.Errorf("create onboarding profile: %w", err)
	}

	// Commit the transaction first so workspace exists
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Update session as completed
	_, err = s.pool.Exec(ctx, `
		UPDATE onboarding_sessions
		SET status = 'completed', workspace_id = $1, completed_at = NOW(), 
		    extracted_data = $2, current_step = 'complete', updated_at = NOW()
		WHERE id = $3
	`, workspace.ID, structToMap(extractedData), session.ID)
	if err != nil {
		// Log warning but don't fail - workspace was created successfully
		fmt.Printf("Warning: Failed to update session status: %v\n", err)
	}

	return &CompleteOnboardingResponse{
		WorkspaceID:   workspace.ID,
		WorkspaceName: workspace.Name,
		WorkspaceSlug: workspace.Slug,
		RedirectURL:   "/window",
	}, nil
}

// SubmitFallbackForm handles fallback form submission
func (s *OnboardingService) SubmitFallbackForm(ctx context.Context, sessionID uuid.UUID, userID string, data *FallbackFormData) (*CompleteOnboardingResponse, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Update session with fallback flag
	_, err = s.pool.Exec(ctx, `
		UPDATE onboarding_sessions
		SET fallback_triggered = true, updated_at = NOW()
		WHERE id = $1
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("update fallback flag: %w", err)
	}

	// Complete onboarding with form data
	return s.CompleteOnboarding(ctx, sessionID, userID, data.Integrations)
}

func onboardingStrPtr(s string) *string {
	return &s
}

func structToMap(v interface{}) map[string]interface{} {
	data, _ := json.Marshal(v)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// GetRecommendations returns integration recommendations based on session data
func (s *OnboardingService) GetRecommendations(ctx context.Context, sessionID uuid.UUID, userID string) ([]string, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Parse extracted data
	var extractedData ExtractedOnboardingData
	if session.ExtractedData != nil {
		dataBytes, _ := json.Marshal(session.ExtractedData)
		json.Unmarshal(dataBytes, &extractedData)
	}

	return ComputeRecommendations(extractedData.Challenge, extractedData.BusinessType), nil
}

func generateSlugFromName(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()
	// Add random suffix for uniqueness
	slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
	return slug
}

// ComputeRecommendations returns integration recommendations based on challenge and business type.
// This is the single source of truth for recommendation logic.
func ComputeRecommendations(challenge, businessType string) []string {
	challengeLower := strings.ToLower(challenge)

	if strings.Contains(challengeLower, "organiz") || strings.Contains(challengeLower, "chaos") || strings.Contains(challengeLower, "mess") {
		return []string{"notion", "google", "linear"}
	}
	if strings.Contains(challengeLower, "scale") || strings.Contains(challengeLower, "grow") || strings.Contains(challengeLower, "automat") {
		return []string{"linear", "slack", "airtable"}
	}
	if strings.Contains(challengeLower, "client") || strings.Contains(challengeLower, "customer") || strings.Contains(challengeLower, "crm") {
		return []string{"hubspot", "slack", "google"}
	}
	if strings.Contains(challengeLower, "team") || strings.Contains(challengeLower, "collaborat") || strings.Contains(challengeLower, "communic") {
		return []string{"slack", "notion", "linear"}
	}
	if strings.Contains(challengeLower, "time") || strings.Contains(challengeLower, "busy") || strings.Contains(challengeLower, "meeting") {
		return []string{"google", "fathom", "slack"}
	}

	// Default by business type
	switch businessType {
	case "agency", "consulting":
		return []string{"hubspot", "slack", "notion"}
	case "startup":
		return []string{"linear", "slack", "notion"}
	case "freelance":
		return []string{"google", "notion", "fathom"}
	default:
		return []string{"google", "slack", "notion"}
	}
}
