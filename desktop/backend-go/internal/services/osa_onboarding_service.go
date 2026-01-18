package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rhl/businessos-backend/internal/integrations/google"
	"github.com/rhl/businessos-backend/internal/integrations/osa"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OSAOnboardingService handles the "Build Your OS" onboarding flow
// This creates personalized starter apps based on user data analysis
type OSAOnboardingService struct {
	pool           *pgxpool.Pool
	osaClient      *osa.Client
	googleProvider *google.Provider
	aiProvider     string
}

// UserAnalysisResult represents the AI's analysis of the user
type UserAnalysisResult struct {
	Insights       []string               `json:"insights"`        // e.g. "No-code builder energy"
	Interests      []string               `json:"interests"`       // e.g. ["design tools", "automation"]
	ToolsUsed      []string               `json:"tools_used"`      // e.g. ["Figma", "Notion"]
	ProfileSummary string                 `json:"profile_summary"` // Full summary for context
	RawData        map[string]interface{} `json:"raw_data"`        // For debugging
}

// StarterApp represents a personalized starter app
type StarterApp struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconEmoji   string `json:"icon_emoji"`  // Emoji for icon (temporary)
	IconURL     string `json:"icon_url"`    // URL once generated
	Reasoning   string `json:"reasoning"`   // Why this app was created
	Category    string `json:"category"`    // Type of app
	Status      string `json:"status"`      // "generating", "ready", "failed"
	WorkflowID  string `json:"workflow_id"` // OSA workflow ID
}

// OnboardingAnalysisRequest represents the request to analyze user data
type OnboardingAnalysisRequest struct {
	UserID            uuid.UUID `json:"user_id"`
	Email             string    `json:"email"`
	GmailConnected    bool      `json:"gmail_connected"`
	CalendarConnected bool      `json:"calendar_connected"`
}

// OnboardingAnalysisResponse represents the analysis result
type OnboardingAnalysisResponse struct {
	Analysis      *UserAnalysisResult `json:"analysis"`
	StarterApps   []StarterApp        `json:"starter_apps"`
	ReadyToLaunch bool                `json:"ready_to_launch"` // All apps generated
}

func NewOSAOnboardingService(pool *pgxpool.Pool, osaClient *osa.Client, googleProvider *google.Provider, aiProvider string) *OSAOnboardingService {
	return &OSAOnboardingService{
		pool:           pool,
		osaClient:      osaClient,
		googleProvider: googleProvider,
		aiProvider:     aiProvider,
	}
}

// AnalyzeUser analyzes user's data and generates personalized insights
func (s *OSAOnboardingService) AnalyzeUser(ctx context.Context, req *OnboardingAnalysisRequest) (*UserAnalysisResult, error) {
	slog.Info("Analyzing user for OSA Build onboarding",
		"user_id", req.UserID,
		"gmail_connected", req.GmailConnected,
	)

	result := &UserAnalysisResult{
		Insights:  []string{},
		Interests: []string{},
		ToolsUsed: []string{},
		RawData:   make(map[string]interface{}),
	}

	// Step 1: Gather data from connected services
	var gmailData string
	if req.GmailConnected && s.googleProvider != nil {
		slog.Info("Fetching Gmail data for analysis")
		// TODO: Implement actual Gmail analysis
		// For now, use mock data
		gmailData = s.mockGmailAnalysis(req.Email)
		result.RawData["gmail_summary"] = "Analyzed recent emails"
	}

	// Step 2: Use AI to analyze the data and create insights
	insights, err := s.generateInsights(ctx, req.Email, gmailData)
	if err != nil {
		slog.Error("Failed to generate insights", "error", err)
		// Use fallback insights
		insights = s.fallbackInsights()
	}

	result.Insights = insights.Messages
	result.Interests = insights.Interests
	result.ToolsUsed = insights.Tools
	result.ProfileSummary = insights.Summary

	slog.Info("User analysis complete",
		"insights_count", len(result.Insights),
		"interests_count", len(result.Interests),
	)

	return result, nil
}

// GenerateStarterApps creates 4 personalized starter apps using OSA orchestration
func (s *OSAOnboardingService) GenerateStarterApps(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, analysis *UserAnalysisResult) ([]StarterApp, error) {
	slog.Info("Generating starter apps",
		"user_id", userID,
		"workspace_id", workspaceID,
		"interests_count", len(analysis.Interests),
	)

	apps := make([]StarterApp, 0, 4)

	// Generate 4 personalized apps based on analysis
	appPrompts := s.createAppPrompts(analysis)

	for i, prompt := range appPrompts {
		app := StarterApp{
			ID:          uuid.New().String(),
			Title:       prompt.Title,
			Description: prompt.Description,
			IconEmoji:   prompt.IconEmoji,
			Reasoning:   prompt.Reasoning,
			Category:    prompt.Category,
			Status:      "generating",
		}

		// Call OSA to actually generate the app
		slog.Info("Triggering OSA app generation",
			"app_title", app.Title,
			"category", app.Category,
		)

		osaReq := &osa.AppGenerationRequest{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Name:        app.Title,
			Description: prompt.FullPrompt,
			Type:        "module",
			Parameters: map[string]interface{}{
				"category":   app.Category,
				"icon_emoji": app.IconEmoji,
				"priority":   "starter_app",
				"index":      i,
			},
		}

		resp, err := s.osaClient.GenerateApp(ctx, osaReq)
		if err != nil {
			slog.Error("Failed to generate app via OSA", "error", err, "app_title", app.Title)
			app.Status = "failed"
		} else {
			app.WorkflowID = resp.AppID
			app.Status = "ready" // In production, this would be async
			slog.Info("OSA app generation started", "workflow_id", resp.AppID)
		}

		apps = append(apps, app)
	}

	slog.Info("Starter apps generation complete", "apps_count", len(apps))

	return apps, nil
}

// GetAppGenerationStatus checks the status of starter app generation
func (s *OSAOnboardingService) GetAppGenerationStatus(ctx context.Context, userID uuid.UUID, apps []StarterApp) (bool, error) {
	allReady := true

	for i := range apps {
		if apps[i].Status == "generating" && apps[i].WorkflowID != "" {
			// Check status from OSA
			status, err := s.osaClient.GetAppStatus(ctx, apps[i].WorkflowID, userID)
			if err != nil {
				slog.Error("Failed to check app status", "error", err, "workflow_id", apps[i].WorkflowID)
				continue
			}

			if status.Status == "completed" {
				apps[i].Status = "ready"
			} else if status.Status == "failed" {
				apps[i].Status = "failed"
			} else {
				allReady = false
			}
		}

		if apps[i].Status != "ready" {
			allReady = false
		}
	}

	return allReady, nil
}

// ===== INTERNAL HELPERS =====

// appPromptTemplate represents a template for generating an app
type appPromptTemplate struct {
	Title       string
	Description string
	IconEmoji   string
	Reasoning   string
	Category    string
	FullPrompt  string
}

type insightResult struct {
	Messages  []string
	Interests []string
	Tools     []string
	Summary   string
}

// generateInsights uses AI to analyze user data and create conversational insights
func (s *OSAOnboardingService) generateInsights(ctx context.Context, email string, gmailData string) (*insightResult, error) {
	// TODO: Implement actual AI analysis
	// For now, return mock insights that match iOS tone

	// Extract domain from email to make insights
	domain := ""
	if parts := strings.Split(email, "@"); len(parts) == 2 {
		domain = parts[1]
	}

	result := &insightResult{
		Messages:  []string{},
		Interests: []string{},
		Tools:     []string{},
	}

	// Generate 3 conversational insights (matching iOS screenshots)
	if strings.Contains(email, "gmail.com") || strings.Contains(email, "hey.com") {
		result.Messages = []string{
			"Productivity enthusiast energy ✨",
			"Digital workspace optimizer",
			"Early adopter, loves trying new tools",
		}
		result.Interests = []string{"productivity", "automation", "workspace tools"}
		result.Tools = []string{"Gmail", "Calendar", "Notion"}
	} else if strings.Contains(domain, ".edu") {
		result.Messages = []string{
			"Academic researcher vibes 📚",
			"Knowledge organization is your thing",
			"Curious learner, always exploring",
		}
		result.Interests = []string{"research", "learning", "knowledge management"}
		result.Tools = []string{"Google Scholar", "Notion", "Zotero"}
	} else {
		// Generic insights
		result.Messages = []string{
			"Getting-things-done energy 💪",
			"Organized and intentional",
			"Focused on what matters",
		}
		result.Interests = []string{"productivity", "focus", "organization"}
		result.Tools = []string{"Email", "Calendar"}
	}

	result.Summary = fmt.Sprintf("A %s who values %s and uses %s regularly",
		result.Messages[0],
		strings.Join(result.Interests, ", "),
		strings.Join(result.Tools, ", "),
	)

	return result, nil
}

// createAppPrompts creates 4 app prompts based on user analysis
func (s *OSAOnboardingService) createAppPrompts(analysis *UserAnalysisResult) []appPromptTemplate {
	prompts := []appPromptTemplate{}

	// App 1: Based on interests
	if len(analysis.Interests) > 0 {
		interest := analysis.Interests[0]
		prompts = append(prompts, appPromptTemplate{
			Title:       fmt.Sprintf("%s Tracker", strings.Title(interest)),
			Description: fmt.Sprintf("Track and organize your %s projects", interest),
			IconEmoji:   "📚",
			Reasoning:   fmt.Sprintf("Because you're interested in %s", interest),
			Category:    "tracker",
			FullPrompt:  fmt.Sprintf("Create a simple tracker app for %s. Include a list view, add/edit functionality, and basic categorization.", interest),
		})
	}

	// App 2: Based on tools used
	if len(analysis.ToolsUsed) > 0 {
		tool := analysis.ToolsUsed[0]
		prompts = append(prompts, appPromptTemplate{
			Title:       fmt.Sprintf("%s Companion", tool),
			Description: fmt.Sprintf("Quick access to your %s workflows", tool),
			IconEmoji:   "🎨",
			Reasoning:   fmt.Sprintf("Because you use %s frequently", tool),
			Category:    "companion",
			FullPrompt:  fmt.Sprintf("Create a companion app that helps users work better with %s. Include quick links, shortcuts, and productivity tips.", tool),
		})
	}

	// App 3: Community/Feedback
	prompts = append(prompts, appPromptTemplate{
		Title:       "Idea Inbox",
		Description: "Capture ideas and get feedback",
		IconEmoji:   "💡",
		Reasoning:   "For collecting thoughts and feedback",
		Category:    "feedback",
		FullPrompt:  "Create a simple idea inbox app where users can capture ideas, add notes, and optionally share for feedback. Include a clean card-based UI.",
	})

	// App 4: Daily utility
	prompts = append(prompts, appPromptTemplate{
		Title:       "Daily Focus",
		Description: "Plan your day, track what matters",
		IconEmoji:   "🎯",
		Reasoning:   "For staying focused on priorities",
		Category:    "daily",
		FullPrompt:  "Create a daily focus app with today's top 3 priorities, a quick note area, and a simple progress tracker. Minimalist design.",
	})

	return prompts
}

// fallbackInsights provides default insights if AI analysis fails
func (s *OSAOnboardingService) fallbackInsights() *insightResult {
	return &insightResult{
		Messages: []string{
			"Ready to build something amazing 🚀",
			"Organized and intentional",
			"Focused on getting things done",
		},
		Interests: []string{"productivity", "organization", "creativity"},
		Tools:     []string{"Email", "Calendar"},
		Summary:   "A builder ready to create their personalized OS",
	}
}

// mockGmailAnalysis provides mock Gmail analysis
func (s *OSAOnboardingService) mockGmailAnalysis(email string) string {
	return fmt.Sprintf("User %s shows interest in productivity tools and software development", email)
}

// SaveOnboardingProfile saves the analysis and apps to the database
func (s *OSAOnboardingService) SaveOnboardingProfile(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, analysis *UserAnalysisResult, apps []StarterApp) error {
	slog.Info("Saving onboarding profile", "user_id", userID, "workspace_id", workspaceID)

	// Serialize data
	analysisJSON, _ := json.Marshal(analysis)
	appsJSON, _ := json.Marshal(apps)

	// Save to workspace_onboarding_profiles table
	_, err := s.pool.Exec(ctx, `
		INSERT INTO workspace_onboarding_profiles (
			workspace_id,
			user_id,
			analysis_data,
			starter_apps_data,
			onboarding_method,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, 'osa_build', NOW(), NOW())
		ON CONFLICT (workspace_id)
		DO UPDATE SET
			analysis_data = EXCLUDED.analysis_data,
			starter_apps_data = EXCLUDED.starter_apps_data,
			updated_at = NOW()
	`, workspaceID, userID, analysisJSON, appsJSON)

	if err != nil {
		slog.Error("Failed to save onboarding profile", "error", err)
		return fmt.Errorf("save onboarding profile: %w", err)
	}

	slog.Info("Onboarding profile saved successfully")
	return nil
}

// GetOnboardingProfile retrieves saved onboarding data
func (s *OSAOnboardingService) GetOnboardingProfile(ctx context.Context, workspaceID uuid.UUID) (*UserAnalysisResult, []StarterApp, error) {
	var analysisJSON, appsJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT analysis_data, starter_apps_data
		FROM workspace_onboarding_profiles
		WHERE workspace_id = $1
	`, workspaceID).Scan(&analysisJSON, &appsJSON)

	if err != nil {
		return nil, nil, fmt.Errorf("get onboarding profile: %w", err)
	}

	var analysis UserAnalysisResult
	var apps []StarterApp

	json.Unmarshal(analysisJSON, &analysis)
	json.Unmarshal(appsJSON, &apps)

	return &analysis, apps, nil
}
