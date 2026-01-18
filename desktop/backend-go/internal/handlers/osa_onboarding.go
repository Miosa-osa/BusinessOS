package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	integrationGoogle "github.com/rhl/businessos-backend/internal/integrations/google"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// OSAOnboardingHandler handles OSA Build onboarding analysis endpoints
type OSAOnboardingHandler struct {
	pool            *pgxpool.Pool
	queries         *sqlc.Queries
	cfg             *config.Config
	emailAnalyzer   *services.EmailAnalyzerService
	profileAnalyzer *services.ProfileAnalyzerAgent
	appCustomizer   *services.AppCustomizerAgent
	gmailService    *integrationGoogle.GmailService
}

// NewOSAOnboardingHandler creates a new OSA onboarding handler
func NewOSAOnboardingHandler(pool *pgxpool.Pool, cfg *config.Config, googleProvider *integrationGoogle.Provider) *OSAOnboardingHandler {
	gmailService := integrationGoogle.NewGmailService(googleProvider)

	return &OSAOnboardingHandler{
		pool:            pool,
		queries:         sqlc.New(pool),
		cfg:             cfg,
		emailAnalyzer:   services.NewEmailAnalyzerService(pool, gmailService),
		profileAnalyzer: services.NewProfileAnalyzerAgent(cfg),
		appCustomizer:   services.NewAppCustomizerAgent(cfg),
		gmailService:    gmailService,
	}
}

// AnalyzeUserRequest represents the request to analyze a user
type AnalyzeUserRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	WorkspaceID string `json:"workspace_id" binding:"required"`
	MaxEmails   int    `json:"max_emails"` // Default: 50
}

// AnalyzeUserResponse represents the analysis result
type AnalyzeUserResponse struct {
	AnalysisID string   `json:"analysis_id"`
	Status     string   `json:"status"`
	Insights   []string `json:"insights,omitempty"`
	Interests  []string `json:"interests,omitempty"`
	ToolsUsed  []string `json:"tools_used,omitempty"`
	Summary    string   `json:"summary,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// StartAnalysis starts the email analysis and profile generation process
// POST /api/v1/osa-onboarding/analyze
func (h *OSAOnboardingHandler) StartAnalysis(c *gin.Context) {
	var req AnalyzeUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to 50 emails
	if req.MaxEmails == 0 {
		req.MaxEmails = 50
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id format"})
		return
	}

	slog.Info("Starting OSA onboarding analysis",
		"user_id", req.UserID,
		"workspace_id", req.WorkspaceID,
		"max_emails", req.MaxEmails,
	)

	// Check if analysis already exists
	existing, _ := h.queries.GetUserAnalysisByUserAndWorkspace(c.Request.Context(), sqlc.GetUserAnalysisByUserAndWorkspaceParams{
		UserID:      req.UserID,
		WorkspaceID: uuidToPgtype(workspaceUUID),
	})

	if existing.ID.Valid {
		// Return existing analysis
		c.JSON(http.StatusOK, AnalyzeUserResponse{
			AnalysisID: pgtypeToUUID(existing.ID).String(),
			Status:     derefString(existing.Status),
			Insights:   parseJSONArray(existing.Insights),
			Interests:  parseJSONArray(existing.Interests),
			ToolsUsed:  parseJSONArray(existing.ToolsUsed),
			Summary:    derefString(existing.ProfileSummary),
		})
		return
	}

	// Create initial analysis record
	analysis, err := h.queries.CreateUserAnalysis(c.Request.Context(), sqlc.CreateUserAnalysisParams{
		UserID:           req.UserID,
		WorkspaceID:      uuidToPgtype(workspaceUUID),
		Insights:         []byte("[]"),
		Interests:        []byte("[]"),
		ToolsUsed:        []byte("[]"),
		EmailMetadata:    []byte("{}"),
		SenderDomains:    []byte("[]"),
		DetectedPatterns: []byte("{}"),
		AnalysisModel:    "llama-3.3-70b-versatile",
		AiProvider:       "groq",
		Status:           stringPtr("analyzing"),
	})
	if err != nil {
		slog.Error("Failed to create analysis record", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create analysis"})
		return
	}

	// Start async analysis
	go h.performAnalysis(context.Background(), pgtypeToUUID(analysis.ID), userUUID, workspaceUUID, req.UserID, req.MaxEmails)

	c.JSON(http.StatusOK, AnalyzeUserResponse{
		AnalysisID: pgtypeToUUID(analysis.ID).String(),
		Status:     "analyzing",
	})
}

// performAnalysis performs the actual analysis asynchronously
func (h *OSAOnboardingHandler) performAnalysis(ctx context.Context, analysisID uuid.UUID, userUUID uuid.UUID, workspaceUUID uuid.UUID, userID string, maxEmails int) {
	startTime := time.Now()

	// Step 1: Extract email metadata
	slog.Info("Step 1: Extracting email metadata", "analysis_id", analysisID.String())
	emailMetadata, err := h.emailAnalyzer.AnalyzeRecentEmails(ctx, userID, maxEmails)
	if err != nil {
		slog.Error("Email analysis failed", "error", err, "analysis_id", analysisID.String())
		errStr := err.Error()
		h.queries.UpdateUserAnalysisStatus(ctx, sqlc.UpdateUserAnalysisStatusParams{
			ID:           uuidToPgtype(analysisID),
			Status:       stringPtr("failed"),
			ErrorMessage: &errStr,
		})
		return
	}

	// Step 2: Convert metadata to profile analyzer input
	profileInput := &services.EmailMetadataInput{
		TotalEmails:     emailMetadata.TotalEmails,
		SenderDomains:   emailMetadata.SenderDomains,
		SubjectKeywords: emailMetadata.SubjectKeywords,
		BodyKeywords:    emailMetadata.BodyKeywords,
		DetectedTools:   emailMetadata.DetectedTools,
		TopicFrequency:  emailMetadata.TopicFrequency,
		EmailDates:      emailMetadata.EmailDates,
	}

	// Step 3: AI profile analysis
	slog.Info("Step 2: Running AI profile analysis", "analysis_id", analysisID.String())
	profile, err := h.profileAnalyzer.AnalyzeProfile(ctx, profileInput)
	if err != nil {
		slog.Error("Profile analysis failed", "error", err, "analysis_id", analysisID.String())
		errStr := err.Error()
		h.queries.UpdateUserAnalysisStatus(ctx, sqlc.UpdateUserAnalysisStatusParams{
			ID:           uuidToPgtype(analysisID),
			Status:       stringPtr("failed"),
			ErrorMessage: &errStr,
		})
		return
	}

	// Step 4: Save analysis results
	slog.Info("Step 3: Saving analysis results", "analysis_id", analysisID.String())
	insightsJSON, _ := json.Marshal(profile.Insights)
	interestsJSON, _ := json.Marshal(profile.Interests)
	toolsJSON, _ := json.Marshal(profile.ToolsUsed)
	emailMetaJSON, _ := json.Marshal(emailMetadata)
	senderDomainsJSON, _ := json.Marshal(emailMetadata.SenderDomains)

	totalEmails := int32(emailMetadata.TotalEmails)
	tokensUsed := int32(0)
	durationMs := int32(time.Since(startTime).Milliseconds())

	_, err = h.queries.UpdateUserAnalysisResults(ctx, sqlc.UpdateUserAnalysisResultsParams{
		ID:                  uuidToPgtype(analysisID),
		Insights:            insightsJSON,
		Interests:           interestsJSON,
		ToolsUsed:           toolsJSON,
		ProfileSummary:      &profile.ProfileSummary,
		EmailMetadata:       emailMetaJSON,
		TotalEmailsAnalyzed: &totalEmails,
		SenderDomains:       senderDomainsJSON,
		DetectedPatterns:    []byte("{}"),
		AnalysisTokensUsed:  &tokensUsed,
		AnalysisDurationMs:  &durationMs,
	})

	if err != nil {
		slog.Error("Failed to save analysis results", "error", err, "analysis_id", analysisID.String())
		return
	}

	slog.Info("Analysis complete",
		"analysis_id", analysisID.String(),
		"duration_ms", time.Since(startTime).Milliseconds(),
	)
}

// GetAnalysisProgress returns the current analysis status
// GET /api/v1/osa-onboarding/analyze/:analysis_id
func (h *OSAOnboardingHandler) GetAnalysisProgress(c *gin.Context) {
	analysisIDStr := c.Param("analysis_id")
	analysisID, err := uuid.Parse(analysisIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid analysis_id"})
		return
	}

	analysis, err := h.queries.GetUserAnalysisByID(c.Request.Context(), uuidToPgtype(analysisID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found"})
		return
	}

	c.JSON(http.StatusOK, AnalyzeUserResponse{
		AnalysisID: pgtypeToUUID(analysis.ID).String(),
		Status:     derefString(analysis.Status),
		Insights:   parseJSONArray(analysis.Insights),
		Interests:  parseJSONArray(analysis.Interests),
		ToolsUsed:  parseJSONArray(analysis.ToolsUsed),
		Summary:    derefString(analysis.ProfileSummary),
		Error:      derefString(analysis.ErrorMessage),
	})
}

// StreamAnalysisProgress streams analysis progress via SSE
// GET /api/v1/osa-onboarding/analyze/:analysis_id/stream
func (h *OSAOnboardingHandler) StreamAnalysisProgress(c *gin.Context) {
	analysisIDStr := c.Param("analysis_id")
	analysisID, err := uuid.Parse(analysisIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid analysis_id"})
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	writer := streaming.NewSSEWriter(c.Writer)

	// Poll for updates every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(2 * time.Minute) // 2 minute timeout

	for {
		select {
		case <-ticker.C:
			analysis, err := h.queries.GetUserAnalysisByID(c.Request.Context(), uuidToPgtype(analysisID))
			if err != nil {
				writer.WriteEvent(streaming.StreamEvent{
					Type:    streaming.EventTypeError,
					Content: "Failed to get analysis status",
				})
				return
			}

			// Send progress update
			status := derefString(analysis.Status)
			eventData := map[string]interface{}{
				"status":    status,
				"insights":  parseJSONArray(analysis.Insights),
				"interests": parseJSONArray(analysis.Interests),
			}

			writer.WriteEvent(streaming.StreamEvent{
				Type: "progress",
				Data: eventData,
			})

			// If completed or failed, send final event and close
			if status == "completed" || status == "failed" {
				writer.WriteEvent(streaming.StreamEvent{
					Type: streaming.EventTypeDone,
					Data: eventData,
				})
				return
			}

			c.Writer.Flush()

		case <-timeout:
			writer.WriteEvent(streaming.StreamEvent{
				Type:    streaming.EventTypeError,
				Content: "Analysis timeout",
			})
			return

		case <-c.Request.Context().Done():
			return
		}
	}
}

// GenerateAppsRequest represents the request to generate starter apps
type GenerateAppsRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	WorkspaceID string `json:"workspace_id" binding:"required"`
	AnalysisID  string `json:"analysis_id" binding:"required"`
}

// GenerateStarterApps generates personalized starter apps based on analysis
// POST /api/v1/osa-onboarding/generate-apps
func (h *OSAOnboardingHandler) GenerateStarterApps(c *gin.Context) {
	var req GenerateAppsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analysisID, err := uuid.Parse(req.AnalysisID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid analysis_id"})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id"})
		return
	}

	// Get analysis results
	analysis, err := h.queries.GetUserAnalysisByID(c.Request.Context(), uuidToPgtype(analysisID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found"})
		return
	}

	if derefString(analysis.Status) != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analysis not completed yet"})
		return
	}

	// Convert analysis to profile result
	profile := &services.ProfileAnalysisResult{
		Insights:       parseJSONArray(analysis.Insights),
		Interests:      parseJSONArray(analysis.Interests),
		ToolsUsed:      parseJSONArray(analysis.ToolsUsed),
		ProfileSummary: derefString(analysis.ProfileSummary),
	}

	// Generate app recommendations
	slog.Info("Generating starter apps", "analysis_id", req.AnalysisID)
	recommendations, err := h.appCustomizer.RecommendApps(c.Request.Context(), profile)
	if err != nil {
		slog.Error("Failed to generate apps", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate apps"})
		return
	}

	// Save starter apps to database
	for _, app := range recommendations.Apps {
		basedOnInterestsJSON, _ := json.Marshal(app.BasedOnInterests)
		basedOnToolsJSON, _ := json.Marshal(app.BasedOnTools)
		moduleCustomizationsJSON, _ := json.Marshal(app.ModuleCustomizations)

		displayOrder := int32(app.Priority)
		status := "ready"

		_, err := h.queries.CreateStarterApp(c.Request.Context(), sqlc.CreateStarterAppParams{
			UserID:               req.UserID,
			WorkspaceID:          uuidToPgtype(workspaceID),
			AnalysisID:           uuidToPgtype(analysisID),
			Title:                app.Title,
			Description:          &app.Description,
			IconEmoji:            &app.IconEmoji,
			Category:             &app.Category,
			Reasoning:            &app.Reasoning,
			CustomizationPrompt:  app.CustomizationPrompt,
			BasedOnInterests:     basedOnInterestsJSON,
			BasedOnTools:         basedOnToolsJSON,
			BaseModule:           &app.BaseModule,
			ModuleCustomizations: moduleCustomizationsJSON,
			GenerationModel:      stringPtr("llama-3.3-70b-versatile"),
			AiProvider:           stringPtr("groq"),
			DisplayOrder:         &displayOrder,
			Status:               &status,
		})

		if err != nil {
			slog.Error("Failed to save starter app", "error", err, "app_title", app.Title)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"apps":       recommendations.Apps,
		"total_apps": recommendations.TotalApps,
	})
}

// GetUserAnalysisStatus returns the onboarding analysis status for a user
// GET /api/osa-onboarding/user-analysis/:user_id
func (h *OSAOnboardingHandler) GetUserAnalysisStatus(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Query the onboarding_user_analysis table directly
	var status, errorMessage *string
	var insights, toolsUsed, senderDomains, detectedPatterns []byte
	var totalEmails *int

	err := h.pool.QueryRow(c.Request.Context(), `
		SELECT
			status,
			insights,
			tools_used,
			total_emails_analyzed,
			sender_domains,
			detected_patterns,
			error_message
		FROM onboarding_user_analysis
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&status, &insights, &toolsUsed, &totalEmails, &senderDomains, &detectedPatterns, &errorMessage)

	if err != nil {
		// Analysis not started yet - return not_started status
		c.JSON(http.StatusOK, gin.H{
			"status":       "not_started",
			"insights":     []string{},
			"tools":        []string{},
			"total_emails": 0,
		})
		return
	}

	// Parse JSON arrays
	insightsArray := parseJSONArray(insights)
	toolsArray := parseJSONArray(toolsUsed)
	domainsArray := parseJSONArray(senderDomains)
	patternsArray := parseJSONArray(detectedPatterns)

	response := gin.H{
		"status":            derefString(status),
		"insights":          insightsArray,
		"tools":             toolsArray,
		"sender_domains":    domainsArray,
		"detected_patterns": patternsArray,
	}

	if totalEmails != nil {
		response["total_emails"] = *totalEmails
	} else {
		response["total_emails"] = 0
	}

	if errorMessage != nil && *errorMessage != "" {
		response["error"] = *errorMessage
	}

	c.JSON(http.StatusOK, response)
}

// Helper: parseJSONArray parses a JSON byte array into []string
func parseJSONArray(data []byte) []string {
	var result []string
	json.Unmarshal(data, &result)
	return result
}

// Helper: uuidToPgtype converts uuid.UUID to pgtype.UUID
func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// Helper: pgtypeToUUID converts pgtype.UUID to uuid.UUID
func pgtypeToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return u.Bytes
}

// Helper: derefString safely dereferences string pointer
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
