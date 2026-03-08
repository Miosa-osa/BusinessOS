package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AppGenerationWorker processes the app generation queue
type AppGenerationWorker struct {
	pool      *pgxpool.Pool
	osaClient OSAClient // Interface for OSA integration
	deployer  *AppDeploymentService
	logger    *slog.Logger
}

// OSAClient interface for generating apps
type OSAClient interface {
	GenerateApp(ctx context.Context, prompt string, config map[string]interface{}) (*OSAGenerateResult, error)
}

// OSAGenerateResult represents result from OSA generation
type OSAGenerateResult struct {
	AppID       uuid.UUID
	Name        string
	Description string
	CodeBundle  string
	Metadata    map[string]interface{}
}

// QueueItem represents an item from app_generation_queue
type QueueItem struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	TemplateID        uuid.UUID
	Status            string
	Priority          int
	GenerationContext map[string]interface{}
	ErrorMessage      *string
	RetryCount        int
	MaxRetries        int
	CreatedAt         time.Time
}

func NewAppGenerationWorker(
	pool *pgxpool.Pool,
	osaClient OSAClient,
	deployer *AppDeploymentService,
	logger *slog.Logger,
) *AppGenerationWorker {
	return &AppGenerationWorker{
		pool:      pool,
		osaClient: osaClient,
		deployer:  deployer,
		logger:    logger,
	}
}

// ProcessQueue processes pending items in the generation queue
func (w *AppGenerationWorker) ProcessQueue(ctx context.Context) error {
	w.logger.Info("starting app generation queue processing")

	// Get next pending item (with row locking to prevent concurrent processing)
	item, err := w.getNextPendingItem(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.logger.Debug("no pending items in queue")
			return nil
		}
		return fmt.Errorf("get next pending item: %w", err)
	}

	if item == nil {
		w.logger.Debug("no pending items in queue")
		return nil
	}

	w.logger.Info("processing queue item",
		"item_id", item.ID,
		"workspace_id", item.WorkspaceID,
		"template_id", item.TemplateID,
	)

	// Mark as processing
	err = w.updateQueueStatus(ctx, item.ID, "processing")
	if err != nil {
		return fmt.Errorf("update status to processing: %w", err)
	}

	// Process the item
	err = w.processQueueItem(ctx, item)
	if err != nil {
		w.logger.Error("failed to process queue item",
			"item_id", item.ID,
			"error", err,
		)

		// Update error and retry logic
		err = w.handleProcessingError(ctx, item, err)
		return err
	}

	// Mark as completed
	err = w.markCompleted(ctx, item.ID)
	if err != nil {
		return fmt.Errorf("mark completed: %w", err)
	}

	w.logger.Info("queue item processed successfully",
		"item_id", item.ID,
		"workspace_id", item.WorkspaceID,
	)

	return nil
}

// enrichGenerationContext merges AI insights into the base generation context
// to provide OSA with richer information for personalized app generation
func (w *AppGenerationWorker) enrichGenerationContext(
	ctx context.Context,
	workspaceID uuid.UUID,
	baseContext map[string]interface{},
) map[string]interface{} {
	w.logger.Info("🔍 enrichGenerationContext: Starting context enrichment",
		"workspace_id", workspaceID,
		"base_context_keys", getMapKeys(baseContext),
	)

	// Start with base context
	enrichedContext := make(map[string]interface{})
	for k, v := range baseContext {
		enrichedContext[k] = v
	}

	// CRITICAL: Always include workspace_id - required by OSAClientAdapter
	enrichedContext["workspace_id"] = workspaceID.String()

	// Try to fetch AI analysis from onboarding_user_analysis table
	var analysisID uuid.UUID
	var insights, interests, toolsUsed []byte
	var profileSummary string

	err := w.pool.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(insights, '[]'::jsonb),
			COALESCE(interests, '[]'::jsonb),
			COALESCE(tools_used, '[]'::jsonb),
			COALESCE(profile_summary, '')
		FROM onboarding_user_analysis
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, workspaceID).Scan(
		&analysisID,
		&insights,
		&interests,
		&toolsUsed,
		&profileSummary,
	)

	if err != nil {
		// Table might not exist yet or no analysis for this workspace
		w.logger.Warn("❌ No AI analysis found for workspace, using base context only",
			"workspace_id", workspaceID,
			"error", err.Error(),
		)
		return enrichedContext
	}

	w.logger.Info("✅ AI analysis found",
		"workspace_id", workspaceID,
		"analysis_id", analysisID,
	)

	// Unmarshal JSON fields and add to context
	var insightsArray []string
	if err := json.Unmarshal(insights, &insightsArray); err == nil && len(insightsArray) > 0 {
		enrichedContext["ai_insights"] = insightsArray
	}

	var interestsArray []string
	if err := json.Unmarshal(interests, &interestsArray); err == nil && len(interestsArray) > 0 {
		enrichedContext["user_interests"] = interestsArray
	}

	var toolsArray []string
	if err := json.Unmarshal(toolsUsed, &toolsArray); err == nil && len(toolsArray) > 0 {
		enrichedContext["tools_used"] = toolsArray
	}

	if profileSummary != "" {
		enrichedContext["profile_summary"] = profileSummary
	}

	// Try to fetch workspace profile for additional structured data
	var businessType, ownerRole, mainChallenge string
	var teamSize int
	var recommendedIntegrations []byte

	err = w.pool.QueryRow(ctx, `
		SELECT
			COALESCE(business_type, ''),
			COALESCE(team_size, 0),
			COALESCE(owner_role, ''),
			COALESCE(main_challenge, ''),
			COALESCE(recommended_integrations, '[]'::jsonb)
		FROM workspace_onboarding_profiles
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, workspaceID).Scan(
		&businessType,
		&teamSize,
		&ownerRole,
		&mainChallenge,
		&recommendedIntegrations,
	)

	if err == nil {
		// Successfully fetched profile data
		w.logger.Info("✅ Workspace profile found",
			"workspace_id", workspaceID,
			"business_type", businessType,
			"team_size", teamSize,
			"owner_role", ownerRole,
		)

		if businessType != "" {
			enrichedContext["business_type"] = businessType
		}
		if teamSize > 0 {
			enrichedContext["team_size"] = teamSize
		}
		if ownerRole != "" {
			enrichedContext["owner_role"] = ownerRole
		}
		if mainChallenge != "" {
			enrichedContext["main_challenge"] = mainChallenge
		}

		var integrationsArray []string
		if err := json.Unmarshal(recommendedIntegrations, &integrationsArray); err == nil && len(integrationsArray) > 0 {
			enrichedContext["recommended_integrations"] = integrationsArray
		}
	} else {
		w.logger.Warn("⚠️  Workspace profile not found",
			"workspace_id", workspaceID,
			"error", err.Error(),
		)
	}

	w.logger.Info("🎉 Generation context enrichment complete",
		"workspace_id", workspaceID,
		"insights_count", len(insightsArray),
		"interests_count", len(interestsArray),
		"tools_count", len(toolsArray),
		"has_profile", err == nil,
		"enriched_context_keys", getMapKeys(enrichedContext),
		"total_keys", len(enrichedContext),
	)

	return enrichedContext
}

// processQueueItem processes a single queue item
func (w *AppGenerationWorker) processQueueItem(ctx context.Context, item *QueueItem) error {
	var generationPrompt string
	var template *AppTemplate

	// Check if this is pure AI generation (no template) or template-based
	isNullTemplate := item.TemplateID == uuid.Nil || item.TemplateID == uuid.UUID{}

	if isNullTemplate {
		// Pure AI generation mode - build prompt from context
		w.logger.Info("pure AI generation mode (no template)",
			"workspace_id", item.WorkspaceID,
		)

		appName := "Generated App"
		description := ""

		if name, ok := item.GenerationContext["app_name"].(string); ok {
			appName = name
		}
		if desc, ok := item.GenerationContext["description"].(string); ok {
			description = desc
		}

		generationPrompt = fmt.Sprintf("Generate an application named '%s'.\n\nDescription: %s\n\nRequirements:\n- Modern architecture\n- Clean code\n- Proper error handling", appName, description)
	} else {
		// Template-based generation
		var err error
		template, err = w.getTemplate(ctx, item.TemplateID)
		if err != nil {
			return fmt.Errorf("fetch template: %w", err)
		}

		w.logger.Info("fetched template",
			"template_name", template.TemplateName,
			"category", template.Category,
		)

		generationPrompt = template.GenerationPrompt
	}

	// 2. Enrich context with AI insights before calling OSA
	enrichedContext := w.enrichGenerationContext(ctx, item.WorkspaceID, item.GenerationContext)

	// 3. Call OSA to generate app
	w.logger.Info("calling OSA to generate app",
		"workspace_id", item.WorkspaceID,
		"has_template", !isNullTemplate,
	)

	result, err := w.osaClient.GenerateApp(ctx, generationPrompt, enrichedContext)
	if err != nil {
		return fmt.Errorf("OSA generate app: %w", err)
	}

	w.logger.Info("OSA generation complete",
		"app_id", result.AppID,
		"app_name", result.Name,
	)

	// 4. Link to user_generated_apps table
	userAppID, err := w.createUserGeneratedAppPureAI(ctx, item.WorkspaceID, template, result, isNullTemplate)
	if err != nil {
		return fmt.Errorf("create user generated app: %w", err)
	}

	w.logger.Info("created user generated app link",
		"user_app_id", userAppID,
		"osa_app_id", result.AppID,
	)

	// 5. Optional: Deploy app automatically (if deployer is configured)
	if w.deployer != nil {
		w.logger.Info("deploying generated app", "osa_app_id", result.AppID)

		_, err = w.deployer.DeployApp(ctx, result.AppID)
		if err != nil {
			// Log but don't fail - deployment can be done later
			w.logger.Warn("auto-deploy failed",
				"osa_app_id", result.AppID,
				"error", err,
			)
		} else {
			w.logger.Info("app deployed successfully", "osa_app_id", result.AppID)
		}
	}

	return nil
}

// createUserGeneratedAppPureAI creates link in user_generated_apps table for pure AI generation
func (w *AppGenerationWorker) createUserGeneratedAppPureAI(
	ctx context.Context,
	workspaceID uuid.UUID,
	template *AppTemplate,
	result *OSAGenerateResult,
	isNullTemplate bool,
) (uuid.UUID, error) {
	var userAppID uuid.UUID

	if isNullTemplate {
		// Pure AI generation - no template
		err := w.pool.QueryRow(ctx, `
			INSERT INTO user_generated_apps (
				workspace_id,
				template_id,
				app_name,
				osa_app_id,
				is_visible,
				is_pinned,
				position_index
			) VALUES ($1, NULL, $2, $3, true, false, NULL)
			RETURNING id
		`, workspaceID, result.Name, result.AppID).Scan(&userAppID)

		if err != nil {
			return uuid.Nil, err
		}
	} else {
		// Template-based - use existing function
		return w.createUserGeneratedApp(ctx, workspaceID, template, result)
	}

	return userAppID, nil
}

// getNextPendingItem gets next item from queue with row locking
func (w *AppGenerationWorker) getNextPendingItem(ctx context.Context) (*QueueItem, error) {
	var item QueueItem
	var contextJSON []byte

	err := w.pool.QueryRow(ctx, `
		SELECT
			id,
			workspace_id,
			template_id,
			status,
			priority,
			COALESCE(generation_context, '{}'::jsonb),
			error_message,
			retry_count,
			max_retries,
			created_at
		FROM app_generation_queue
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`).Scan(
		&item.ID,
		&item.WorkspaceID,
		&item.TemplateID,
		&item.Status,
		&item.Priority,
		&contextJSON,
		&item.ErrorMessage,
		&item.RetryCount,
		&item.MaxRetries,
		&item.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Parse generation context
	if len(contextJSON) > 0 {
		json.Unmarshal(contextJSON, &item.GenerationContext)
	}

	return &item, nil
}

// getTemplate fetches template by ID
func (w *AppGenerationWorker) getTemplate(ctx context.Context, templateID uuid.UUID) (*AppTemplate, error) {
	var t AppTemplate
	var configJSON []byte

	err := w.pool.QueryRow(ctx, `
		SELECT
			id,
			template_name,
			category,
			display_name,
			COALESCE(description, ''),
			COALESCE(icon_type, ''),
			target_business_types,
			target_challenges,
			target_team_sizes,
			priority_score,
			COALESCE(template_config, '{}'::jsonb),
			required_modules,
			optional_features,
			COALESCE(generation_prompt, ''),
			scaffold_type
		FROM app_templates
		WHERE id = $1
	`, templateID).Scan(
		&t.ID,
		&t.TemplateName,
		&t.Category,
		&t.DisplayName,
		&t.Description,
		&t.IconType,
		&t.TargetBusinessTypes,
		&t.TargetChallenges,
		&t.TargetTeamSizes,
		&t.PriorityScore,
		&configJSON,
		&t.RequiredModules,
		&t.OptionalFeatures,
		&t.GenerationPrompt,
		&t.ScaffoldType,
	)

	if err != nil {
		return nil, err
	}

	// Parse config JSON
	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &t.TemplateConfig)
	}

	return &t, nil
}

// createUserGeneratedApp creates link in user_generated_apps table
func (w *AppGenerationWorker) createUserGeneratedApp(
	ctx context.Context,
	workspaceID uuid.UUID,
	template *AppTemplate,
	result *OSAGenerateResult,
) (uuid.UUID, error) {
	var userAppID uuid.UUID

	err := w.pool.QueryRow(ctx, `
		INSERT INTO user_generated_apps (
			workspace_id,
			template_id,
			app_name,
			osa_app_id,
			is_visible,
			is_pinned,
			position_index
		) VALUES ($1, $2, $3, $4, true, false, NULL)
		RETURNING id
	`, workspaceID, template.ID, result.Name, result.AppID).Scan(&userAppID)

	if err != nil {
		return uuid.Nil, err
	}

	return userAppID, nil
}

// updateQueueStatus updates queue item status
func (w *AppGenerationWorker) updateQueueStatus(ctx context.Context, itemID uuid.UUID, status string) error {
	_, err := w.pool.Exec(ctx, `
		UPDATE app_generation_queue
		SET
			status = $2::varchar,
			started_at = CASE WHEN $2::varchar = 'processing' THEN NOW() ELSE started_at END
		WHERE id = $1
	`, itemID, status)

	return err
}

// markCompleted marks queue item as completed
func (w *AppGenerationWorker) markCompleted(ctx context.Context, itemID uuid.UUID) error {
	_, err := w.pool.Exec(ctx, `
		UPDATE app_generation_queue
		SET
			status = 'completed',
			completed_at = NOW()
		WHERE id = $1
	`, itemID)

	return err
}

// handleProcessingError handles errors during processing
func (w *AppGenerationWorker) handleProcessingError(ctx context.Context, item *QueueItem, processErr error) error {
	_, err := w.pool.Exec(ctx, `
		UPDATE app_generation_queue
		SET
			error_message = $2,
			retry_count = retry_count + 1,
			status = CASE
				WHEN retry_count + 1 >= max_retries THEN 'failed'
				ELSE 'pending'
			END,
			completed_at = CASE
				WHEN retry_count + 1 >= max_retries THEN NOW()
				ELSE NULL
			END
		WHERE id = $1
	`, item.ID, processErr.Error())

	if err != nil {
		return fmt.Errorf("update error: %w", err)
	}

	return nil
}

// getMapKeys returns the keys of a map as a slice (helper for logging)
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
