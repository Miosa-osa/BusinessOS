package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OSAClientAdapter adapts existing OSA functionality for app generation worker
type OSAClientAdapter struct {
	pool *pgxpool.Pool
}

func NewOSAClientAdapter(pool *pgxpool.Pool) *OSAClientAdapter {
	return &OSAClientAdapter{
		pool: pool,
	}
}

// GenerateApp calls OSA to generate an app and stores result in osa_generated_apps
func (a *OSAClientAdapter) GenerateApp(
	ctx context.Context,
	prompt string,
	config map[string]interface{},
) (*OSAGenerateResult, error) {
	// For MVP, we create a placeholder entry in osa_generated_apps
	// In production, this would call the actual OSA-5 API

	// Extract workspace_id from config
	workspaceIDRaw, ok := config["workspace_id"]
	if !ok {
		return nil, fmt.Errorf("workspace_id not found in config")
	}

	var workspaceID uuid.UUID
	switch v := workspaceIDRaw.(type) {
	case uuid.UUID:
		workspaceID = v
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid workspace_id: %w", err)
		}
		workspaceID = parsed
	default:
		return nil, fmt.Errorf("workspace_id has invalid type")
	}

	// Ensure OSA workspace exists (maps from regular workspace to OSA workspace)
	osaWorkspaceID, err := a.ensureOSAWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("ensure osa workspace: %w", err)
	}

	// Create OSA app entry (this is where OSA-5 integration would go)
	var appID uuid.UUID
	var appName string
	var appDescription string

	// For now, extract name from prompt
	appName = extractAppNameFromPrompt(prompt)
	appDescription = fmt.Sprintf("Generated app: %s", appName)

	err = a.pool.QueryRow(ctx, `
		INSERT INTO osa_generated_apps (
			workspace_id,
			name,
			display_name,
			description,
			status,
			app_name,
			category,
			icon_type,
			app_description
		) VALUES ($1, $2, $3, $4, 'generated', $5, $6, $7, $8)
		RETURNING id
	`,
		osaWorkspaceID,
		makeSlug(appName),
		appName,
		appDescription,
		appName,
		"custom",
		"cube",
		appDescription,
	).Scan(&appID)

	if err != nil {
		return nil, fmt.Errorf("create osa_generated_app: %w", err)
	}

	return &OSAGenerateResult{
		AppID:       appID,
		Name:        appName,
		Description: appDescription,
		CodeBundle:  "", // Would be populated by actual OSA-5
		Metadata: map[string]interface{}{
			"generation_prompt": prompt,
			"config":            config,
		},
	}, nil
}

// ensureOSAWorkspace creates an OSA workspace if one doesn't exist for the given workspace ID
func (a *OSAClientAdapter) ensureOSAWorkspace(ctx context.Context, workspaceID uuid.UUID) (uuid.UUID, error) {
	// First, try to find existing OSA workspace
	var osaWorkspaceID uuid.UUID
	err := a.pool.QueryRow(ctx, `
		SELECT id FROM osa_workspaces WHERE id = $1
	`, workspaceID).Scan(&osaWorkspaceID)

	if err == nil {
		return osaWorkspaceID, nil
	}

	// Get user_id from the regular workspace to create OSA workspace
	var userID string
	err = a.pool.QueryRow(ctx, `
		SELECT COALESCE(user_id, (SELECT id FROM "user" LIMIT 1)) FROM workspaces WHERE id = $1
	`, workspaceID).Scan(&userID)

	if err != nil || userID == "" {
		// Use first user as fallback
		err = a.pool.QueryRow(ctx, `
			SELECT id FROM "user" LIMIT 1
		`).Scan(&userID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("no users found for osa workspace creation: %w", err)
		}
	}

	// Create OSA workspace with same ID as regular workspace
	// Use INSERT ... ON CONFLICT DO UPDATE to always return a row
	err = a.pool.QueryRow(ctx, `
		INSERT INTO osa_workspaces (id, user_id, name)
		VALUES ($1, $2, 'OSA Generated Apps')
		ON CONFLICT (id) DO UPDATE SET name = osa_workspaces.name
		RETURNING id
	`, workspaceID, userID).Scan(&osaWorkspaceID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("create or get osa workspace: %w", err)
	}

	return osaWorkspaceID, nil
}

// extractAppNameFromPrompt extracts app name from generation prompt
func extractAppNameFromPrompt(prompt string) string {
	// Simple extraction - in production this would be smarter
	// For now, take first few words
	words := []string{}
	for i, word := range splitWords(prompt) {
		if i >= 3 {
			break
		}
		words = append(words, word)
	}

	if len(words) == 0 {
		return "Generated App"
	}

	return joinWords(words)
}

// makeSlug creates URL-safe slug from name
func makeSlug(name string) string {
	slug := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			slug += string(r)
		} else if r == ' ' || r == '-' {
			slug += "-"
		}
	}
	return slug
}

func splitWords(s string) []string {
	words := []string{}
	word := ""
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(r)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

func joinWords(words []string) string {
	result := ""
	for i, word := range words {
		if i > 0 {
			result += " "
		}
		result += word
	}
	return result
}
