// Package google provides individual Google tool integrations.
// Each tool (Calendar, Gmail, Drive, etc.) is a separate integration with its own scopes.
package google

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/integrations"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ============================================================================
// TOOL DEFINITIONS
// Each Google service is a separate "tool" with its own provider ID and scopes
// ============================================================================

// ToolDefinition defines a single Google tool (Calendar, Gmail, etc.)
type ToolDefinition struct {
	ID          string   // e.g., "google_calendar"
	Name        string   // e.g., "Google Calendar"
	Description string   // Human-readable description
	Category    string   // e.g., "calendar", "email", "storage"
	Scopes      []string // OAuth scopes required for this tool only
	Modules     []string // BusinessOS modules this tool integrates with
}

// All available Google tools
var GoogleTools = map[string]*ToolDefinition{
	"google_calendar": {
		ID:          "google_calendar",
		Name:        "Google Calendar",
		Description: "Sync calendar events and manage schedules",
		Category:    "calendar",
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.readonly",
			"https://www.googleapis.com/auth/calendar.events",
		},
		Modules: []string{"calendar", "daily_log", "projects"},
	},
	"google_gmail": {
		ID:          "google_gmail",
		Name:        "Gmail",
		Description: "Read and send emails, manage inbox",
		Category:    "email",
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.readonly",
			"https://www.googleapis.com/auth/gmail.send",
			"https://www.googleapis.com/auth/gmail.modify",
		},
		Modules: []string{"chat", "daily_log", "clients"},
	},
	"google_drive": {
		ID:          "google_drive",
		Name:        "Google Drive",
		Description: "Access and manage files in Google Drive",
		Category:    "storage",
		Scopes: []string{
			"https://www.googleapis.com/auth/drive.readonly",
			"https://www.googleapis.com/auth/drive.file",
		},
		Modules: []string{"contexts", "projects"},
	},
	"google_contacts": {
		ID:          "google_contacts",
		Name:        "Google Contacts",
		Description: "Sync contacts and manage address book",
		Category:    "contacts",
		Scopes: []string{
			"https://www.googleapis.com/auth/contacts.readonly",
		},
		Modules: []string{"clients", "team"},
	},
	"google_tasks": {
		ID:          "google_tasks",
		Name:        "Google Tasks",
		Description: "Sync tasks and to-do lists",
		Category:    "tasks",
		Scopes: []string{
			"https://www.googleapis.com/auth/tasks.readonly",
			"https://www.googleapis.com/auth/tasks",
		},
		Modules: []string{"tasks", "projects"},
	},
}

// ============================================================================
// TOOL PROVIDER
// A provider instance for a specific Google tool
// ============================================================================

// ToolProvider provides OAuth and API access for a specific Google tool.
type ToolProvider struct {
	pool        *pgxpool.Pool
	tool        *ToolDefinition
	oauthConfig *oauth2.Config
}

// NewToolProvider creates a provider for a specific Google tool.
func NewToolProvider(pool *pgxpool.Pool, toolID string) (*ToolProvider, error) {
	tool, ok := GoogleTools[toolID]
	if !ok {
		return nil, fmt.Errorf("unknown Google tool: %s", toolID)
	}

	// Use the same config as the main Google provider
	cfg := config.AppConfig

	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return nil, fmt.Errorf("Google OAuth credentials not configured")
	}

	// Use tool-specific redirect URL
	redirectURL := cfg.GoogleIntegrationRedirectURI
	if redirectURL == "" {
		redirectURL = cfg.GoogleRedirectURI
	}
	// Override to tool-specific callback if needed
	if toolID == "google_calendar" {
		redirectURL = strings.Replace(redirectURL, "/google/callback", "/google_calendar/callback", 1)
	} else if toolID == "google_gmail" {
		redirectURL = strings.Replace(redirectURL, "/google/callback", "/google_gmail/callback", 1)
	}

	// Always include basic profile scopes + tool-specific scopes
	scopes := append([]string{
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}, tool.Scopes...)

	return &ToolProvider{
		pool: pool,
		tool: tool,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
	}, nil
}

// ID returns the tool's provider ID.
func (tp *ToolProvider) ID() string {
	return tp.tool.ID
}

// Name returns the tool's display name.
func (tp *ToolProvider) Name() string {
	return tp.tool.Name
}

// Category returns the tool's category.
func (tp *ToolProvider) Category() string {
	return tp.tool.Category
}

// Pool returns the database pool.
func (tp *ToolProvider) Pool() *pgxpool.Pool {
	return tp.pool
}

// GetAuthURL returns the OAuth authorization URL for this specific tool.
func (tp *ToolProvider) GetAuthURL(state string) string {
	return tp.oauthConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

// ExchangeCode exchanges an authorization code for tokens.
func (tp *ToolProvider) ExchangeCode(ctx context.Context, code string) (*integrations.TokenResponse, error) {
	token, err := tp.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user email
	email, err := getUserEmail(ctx, tp.oauthConfig, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user email: %w", err)
	}

	// Extract scopes
	var scopes []string
	if scopeStr, ok := token.Extra("scope").(string); ok {
		scopes = strings.Split(scopeStr, " ")
	}

	return &integrations.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		Scopes:       scopes,
		AccountEmail: email,
		AccountName:  email,
		Metadata: map[string]interface{}{
			"google_email": email,
			"tool_id":      tp.tool.ID,
		},
	}, nil
}

// SaveToken saves the OAuth token for this tool.
func (tp *ToolProvider) SaveToken(ctx context.Context, userID string, token *integrations.TokenResponse) error {
	tx, err := tp.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert or update user_integrations for this specific tool
	_, err = tx.Exec(ctx, `
		INSERT INTO user_integrations (
			user_id, provider_id, status, connected_at,
			external_account_id, external_account_name, scopes, metadata
		) VALUES ($1, $2, 'connected', NOW(), $3, $4, $5, $6)
		ON CONFLICT (user_id, provider_id) DO UPDATE SET
			status = 'connected',
			connected_at = NOW(),
			external_account_id = EXCLUDED.external_account_id,
			external_account_name = EXCLUDED.external_account_name,
			scopes = EXCLUDED.scopes,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
	`, userID, tp.tool.ID, token.AccountEmail, token.AccountName, token.Scopes, token.Metadata)
	if err != nil {
		return fmt.Errorf("failed to save integration: %w", err)
	}

	// Also save to legacy google_oauth_tokens for token storage (easier than encrypted vault)
	_, err = tx.Exec(ctx, `
		INSERT INTO google_oauth_tokens (
			user_id, access_token, refresh_token, expiry, scopes, google_email
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = COALESCE(EXCLUDED.refresh_token, google_oauth_tokens.refresh_token),
			expiry = EXCLUDED.expiry,
			scopes = EXCLUDED.scopes,
			google_email = EXCLUDED.google_email,
			updated_at = NOW()
	`, userID, token.AccessToken, token.RefreshToken, token.ExpiresAt, token.Scopes, token.AccountEmail)
	if err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return tx.Commit(ctx)
}

// GetToken retrieves the OAuth token for this tool.
func (tp *ToolProvider) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	var accessToken, refreshToken string
	var expiry time.Time

	err := tp.pool.QueryRow(ctx, `
		SELECT access_token, refresh_token, expiry
		FROM google_oauth_tokens
		WHERE user_id = $1
	`, userID).Scan(&accessToken, &refreshToken, &expiry)

	if err != nil {
		return nil, fmt.Errorf("no token found for tool %s: %w", tp.tool.ID, err)
	}

	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}, nil
}

// GetTokenSource returns a TokenSource that auto-refreshes.
func (tp *ToolProvider) GetTokenSource(ctx context.Context, userID string) (oauth2.TokenSource, error) {
	token, err := tp.GetToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create a token source that auto-refreshes
	tokenSource := tp.oauthConfig.TokenSource(ctx, token)

	// Get a fresh token to trigger refresh if needed
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// If token was refreshed, save the new one
	if newToken.AccessToken != token.AccessToken {
		log.Printf("Token refreshed for user %s, tool %s", userID, tp.tool.ID)
		tp.saveRefreshedToken(ctx, userID, newToken)
	}

	return tokenSource, nil
}

// saveRefreshedToken saves a refreshed token.
func (tp *ToolProvider) saveRefreshedToken(ctx context.Context, userID string, token *oauth2.Token) {
	_, err := tp.pool.Exec(ctx, `
		UPDATE credential_vault
		SET access_token = $1, expires_at = $2, updated_at = NOW()
		WHERE user_id = $3 AND provider_id = $4
	`, token.AccessToken, token.Expiry, userID, tp.tool.ID)
	if err != nil {
		log.Printf("Failed to save refreshed token: %v", err)
	}
}

// IsConnected checks if the user has this tool connected.
func (tp *ToolProvider) IsConnected(ctx context.Context, userID string) bool {
	var count int
	err := tp.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_integrations
		WHERE user_id = $1 AND provider_id = $2 AND status = 'connected'
	`, userID, tp.tool.ID).Scan(&count)
	return err == nil && count > 0
}

// Disconnect removes the user's connection to this tool.
func (tp *ToolProvider) Disconnect(ctx context.Context, userID string) error {
	tx, err := tp.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete from user_integrations
	_, err = tx.Exec(ctx, `
		DELETE FROM user_integrations
		WHERE user_id = $1 AND provider_id = $2
	`, userID, tp.tool.ID)
	if err != nil {
		return err
	}

	// Delete from credential_vault
	_, err = tx.Exec(ctx, `
		DELETE FROM credential_vault
		WHERE user_id = $1 AND provider_id = $2
	`, userID, tp.tool.ID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetConnectionStatus returns the connection status for this tool.
func (tp *ToolProvider) GetConnectionStatus(ctx context.Context, userID string) (*integrations.ConnectionStatus, error) {
	var status integrations.ConnectionStatus

	err := tp.pool.QueryRow(ctx, `
		SELECT
			COALESCE(status = 'connected', false) as connected,
			connected_at,
			external_account_id,
			external_account_name,
			scopes,
			updated_at
		FROM user_integrations
		WHERE user_id = $1 AND provider_id = $2
	`, userID, tp.tool.ID).Scan(
		&status.Connected,
		&status.ConnectedAt,
		&status.AccountID,
		&status.AccountName,
		&status.Scopes,
		&status.LastSyncAt,
	)

	if err != nil {
		return &integrations.ConnectionStatus{Connected: false}, nil
	}

	return &status, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getUserEmail fetches the user's email from Google.
func getUserEmail(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (string, error) {
	client := config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}

	if err := decodeJSON(resp.Body, &userInfo); err != nil {
		return "", err
	}

	return userInfo.Email, nil
}
