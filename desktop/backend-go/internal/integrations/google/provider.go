// Package google provides the Google Workspace integration (Calendar, Gmail, Drive).
package google

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/integrations"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googlecalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
)

const (
	ProviderID   = "google"
	ProviderName = "Google Workspace"
	Category     = "productivity"
)

// Scopes for different Google services - MINIMAL REQUIRED SCOPES
// Only request what's needed and what's enabled in Google Cloud Console
var (
	// Calendar: Basic calendar access (read events, create/modify events)
	CalendarScopes = []string{
		googlecalendar.CalendarReadonlyScope, // Read calendar list and events
		googlecalendar.CalendarEventsScope,   // Create/edit/delete events
	}

	// Gmail: Basic email access (read, send, modify)
	GmailScopes = []string{
		gmail.GmailReadonlyScope, // Read emails
		gmail.GmailSendScope,     // Send emails
		gmail.GmailModifyScope,   // Modify emails (archive, delete, labels)
	}

	// Drive: Basic file access
	DriveScopes = []string{
		"https://www.googleapis.com/auth/drive.readonly", // Read files and metadata
		"https://www.googleapis.com/auth/drive.file",     // Create/edit files created by app
	}

	// Contacts/People: Basic contacts access
	ContactsScopes = []string{
		"https://www.googleapis.com/auth/contacts.readonly", // Read contacts
	}

	// Tasks: Basic tasks access
	TasksScopes = []string{
		"https://www.googleapis.com/auth/tasks", // Create, edit, delete tasks
	}

	// Sheets: Basic spreadsheet access
	SheetsScopes = []string{
		"https://www.googleapis.com/auth/spreadsheets.readonly", // Read spreadsheets
	}

	// Docs: Basic document access
	DocsScopes = []string{
		"https://www.googleapis.com/auth/documents.readonly", // Read documents
	}

	// Slides: Basic presentation access
	SlidesScopes = []string{
		"https://www.googleapis.com/auth/presentations.readonly", // Read presentations
	}

	// Forms: Basic form access
	FormsScopes = []string{
		"https://www.googleapis.com/auth/forms.currentonly", // Manage forms the app is installed in
	}

	// Chat: Basic chat access (requires Chat API enabled)
	ChatScopes = []string{
		"https://www.googleapis.com/auth/chat.messages.readonly", // Read messages
	}

	// Photos: Basic photos access (requires Photos API enabled)
	PhotosScopes = []string{
		"https://www.googleapis.com/auth/photoslibrary.readonly", // View Google Photos
	}

	// YouTube: Basic YouTube access (requires YouTube API enabled)
	YouTubeScopes = []string{
		"https://www.googleapis.com/auth/youtube.readonly", // View YouTube account
	}

	// Blogger: Basic blog access (requires Blogger API enabled)
	BloggerScopes = []string{
		"https://www.googleapis.com/auth/blogger.readonly", // View Blogger account
	}

	// Classroom: Basic classroom access (requires Classroom API enabled)
	ClassroomScopes = []string{
		"https://www.googleapis.com/auth/classroom.courses.readonly", // View classes
	}

	// User info: Basic profile information
	UserInfoScopes = []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"openid",
	}

	// Cloud Platform: GCP access (requires Cloud Platform API)
	CloudScopes = []string{
		"https://www.googleapis.com/auth/cloud-platform.read-only", // Read GCP access
	}

	// Meet: Google Meet access (requires Meet API)
	MeetScopes = []string{
		"https://www.googleapis.com/auth/meetings.space.readonly", // View meeting space info
	}

	// Keep: Google Keep notes access (requires Keep API)
	KeepScopes = []string{
		"https://www.googleapis.com/auth/keep.readonly", // Read Keep notes
	}

	// Analytics: Google Analytics access (requires Analytics API)
	AnalyticsScopes = []string{
		"https://www.googleapis.com/auth/analytics.readonly", // Read Analytics data
	}

	// Admin SDK: Google Workspace Admin access (requires Admin SDK API + admin account)
	AdminScopes = []string{
		"https://www.googleapis.com/auth/admin.directory.user.readonly", // View users
	}

	// Fitness: Google Fit access (requires Fitness API)
	FitnessScopes = []string{
		"https://www.googleapis.com/auth/fitness.activity.read", // Read activity data
	}

	// Ads: Google Ads access (requires Google Ads API)
	AdsScopes = []string{
		"https://www.googleapis.com/auth/adwords", // Google Ads access
	}

	// Search Console: Google Search Console access (requires Search Console API)
	SearchConsoleScopes = []string{
		"https://www.googleapis.com/auth/webmasters.readonly", // Read Search Console data
	}

	// BigQuery: BigQuery access (requires BigQuery API)
	BigQueryScopes = []string{
		"https://www.googleapis.com/auth/bigquery.readonly", // Read BigQuery data
	}

	// Pub/Sub: Google Pub/Sub access (requires Pub/Sub API)
	PubSubScopes = []string{
		"https://www.googleapis.com/auth/pubsub", // Pub/Sub access
	}

	// Storage: Google Cloud Storage access (requires Cloud Storage API)
	StorageScopes = []string{
		"https://www.googleapis.com/auth/devstorage.read_only", // Read GCS objects
	}

	// AllGoogleScopes contains EVERY Google scope for maximum access
	AllGoogleScopes []string
)

func init() {
	// Build the complete list of all scopes
	AllGoogleScopes = make([]string, 0)
	AllGoogleScopes = append(AllGoogleScopes, UserInfoScopes...)
	AllGoogleScopes = append(AllGoogleScopes, CalendarScopes...)
	AllGoogleScopes = append(AllGoogleScopes, GmailScopes...)
	AllGoogleScopes = append(AllGoogleScopes, DriveScopes...)
	AllGoogleScopes = append(AllGoogleScopes, ContactsScopes...)
	AllGoogleScopes = append(AllGoogleScopes, TasksScopes...)
	AllGoogleScopes = append(AllGoogleScopes, SheetsScopes...)
	AllGoogleScopes = append(AllGoogleScopes, DocsScopes...)
	AllGoogleScopes = append(AllGoogleScopes, SlidesScopes...)
	AllGoogleScopes = append(AllGoogleScopes, FormsScopes...)
	AllGoogleScopes = append(AllGoogleScopes, ChatScopes...)
	AllGoogleScopes = append(AllGoogleScopes, PhotosScopes...)
	AllGoogleScopes = append(AllGoogleScopes, YouTubeScopes...)
	AllGoogleScopes = append(AllGoogleScopes, BloggerScopes...)
	AllGoogleScopes = append(AllGoogleScopes, ClassroomScopes...)
	AllGoogleScopes = append(AllGoogleScopes, CloudScopes...)
	AllGoogleScopes = append(AllGoogleScopes, MeetScopes...)
	AllGoogleScopes = append(AllGoogleScopes, KeepScopes...)
	AllGoogleScopes = append(AllGoogleScopes, AnalyticsScopes...)
	AllGoogleScopes = append(AllGoogleScopes, AdminScopes...)
	AllGoogleScopes = append(AllGoogleScopes, FitnessScopes...)
	AllGoogleScopes = append(AllGoogleScopes, AdsScopes...)
	AllGoogleScopes = append(AllGoogleScopes, SearchConsoleScopes...)
	AllGoogleScopes = append(AllGoogleScopes, BigQueryScopes...)
	AllGoogleScopes = append(AllGoogleScopes, PubSubScopes...)
	AllGoogleScopes = append(AllGoogleScopes, StorageScopes...)
}

// Provider implements the integrations.Provider interface for Google Workspace.
type Provider struct {
	pool        *pgxpool.Pool
	oauthConfig *oauth2.Config
	features    []string // enabled features: "calendar", "gmail", "drive"
}

// AllFeatures contains all available Google feature identifiers
var AllFeatures = []string{
	"calendar", "gmail", "drive", "contacts", "tasks",
	"sheets", "docs", "slides", "forms", "chat",
	"photos", "youtube", "blogger", "classroom", "cloud",
	"meet", "keep", "analytics", "admin", "fitness",
	"ads", "searchconsole", "bigquery", "pubsub", "storage",
}

// NewProvider creates a new Google provider with specified features.
// Available features: calendar, gmail, drive, contacts, tasks, sheets, docs, slides, forms, chat,
// photos, youtube, blogger, classroom, cloud, meet, keep, analytics, admin, fitness, ads, searchconsole, bigquery, pubsub, storage
// Use "all" to enable ALL features with maximum scope access.
func NewProvider(pool *pgxpool.Pool, features []string) *Provider {
	cfg := config.AppConfig

	// Check if "all" is requested
	for _, f := range features {
		if f == "all" {
			features = AllFeatures
			break
		}
	}

	// Build scopes based on enabled features
	scopes := append([]string{}, UserInfoScopes...)
	for _, feature := range features {
		switch feature {
		case "calendar":
			scopes = append(scopes, CalendarScopes...)
		case "gmail":
			scopes = append(scopes, GmailScopes...)
		case "drive":
			scopes = append(scopes, DriveScopes...)
		case "contacts":
			scopes = append(scopes, ContactsScopes...)
		case "tasks":
			scopes = append(scopes, TasksScopes...)
		case "sheets":
			scopes = append(scopes, SheetsScopes...)
		case "docs":
			scopes = append(scopes, DocsScopes...)
		case "slides":
			scopes = append(scopes, SlidesScopes...)
		case "forms":
			scopes = append(scopes, FormsScopes...)
		case "chat":
			scopes = append(scopes, ChatScopes...)
		case "photos":
			scopes = append(scopes, PhotosScopes...)
		case "youtube":
			scopes = append(scopes, YouTubeScopes...)
		case "blogger":
			scopes = append(scopes, BloggerScopes...)
		case "classroom":
			scopes = append(scopes, ClassroomScopes...)
		case "cloud":
			scopes = append(scopes, CloudScopes...)
		case "meet":
			scopes = append(scopes, MeetScopes...)
		case "keep":
			scopes = append(scopes, KeepScopes...)
		case "analytics":
			scopes = append(scopes, AnalyticsScopes...)
		case "admin":
			scopes = append(scopes, AdminScopes...)
		case "fitness":
			scopes = append(scopes, FitnessScopes...)
		case "ads":
			scopes = append(scopes, AdsScopes...)
		case "searchconsole":
			scopes = append(scopes, SearchConsoleScopes...)
		case "bigquery":
			scopes = append(scopes, BigQueryScopes...)
		case "pubsub":
			scopes = append(scopes, PubSubScopes...)
		case "storage":
			scopes = append(scopes, StorageScopes...)
		}
	}

	redirectURI := cfg.GoogleIntegrationRedirectURI
	if redirectURI == "" {
		redirectURI = cfg.GoogleRedirectURI
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return &Provider{
		pool:        pool,
		oauthConfig: oauthConfig,
		features:    features,
	}
}

// NewProviderWithAllFeatures creates a provider with ALL Google features enabled.
// This requests the maximum possible OAuth scopes for full API access.
func NewProviderWithAllFeatures(pool *pgxpool.Pool) *Provider {
	return NewProvider(pool, []string{"all"})
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return ProviderID
}

// DisplayName returns the human-readable provider name.
func (p *Provider) DisplayName() string {
	return ProviderName
}

// Category returns the provider category.
func (p *Provider) Category() string {
	return Category
}

// Icon returns the provider icon URL.
func (p *Provider) Icon() string {
	return "/icons/google.svg"
}

// Features returns the enabled features for this provider.
func (p *Provider) Features() []string {
	return p.features
}

// HasFeature checks if a specific feature is enabled.
func (p *Provider) HasFeature(feature string) bool {
	for _, f := range p.features {
		if f == feature {
			return true
		}
	}
	return false
}

// GetAuthURL returns the OAuth authorization URL.
func (p *Provider) GetAuthURL(state string) string {
	return p.oauthConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

// GetAuthURLWithFeatures returns an OAuth URL with specific feature scopes.
// Available features: calendar, gmail, drive, contacts, tasks, sheets, docs, slides, forms, chat,
// photos, youtube, blogger, classroom, cloud, meet, keep, analytics, admin, fitness, ads, searchconsole, bigquery, pubsub, storage
// Use "all" to request ALL scopes.
func (p *Provider) GetAuthURLWithFeatures(state string, features []string) string {
	// Check if "all" is requested
	for _, f := range features {
		if f == "all" {
			features = AllFeatures
			break
		}
	}

	scopes := append([]string{}, UserInfoScopes...)
	for _, feature := range features {
		switch feature {
		case "calendar":
			scopes = append(scopes, CalendarScopes...)
		case "gmail":
			scopes = append(scopes, GmailScopes...)
		case "drive":
			scopes = append(scopes, DriveScopes...)
		case "contacts":
			scopes = append(scopes, ContactsScopes...)
		case "tasks":
			scopes = append(scopes, TasksScopes...)
		case "sheets":
			scopes = append(scopes, SheetsScopes...)
		case "docs":
			scopes = append(scopes, DocsScopes...)
		case "slides":
			scopes = append(scopes, SlidesScopes...)
		case "forms":
			scopes = append(scopes, FormsScopes...)
		case "chat":
			scopes = append(scopes, ChatScopes...)
		case "photos":
			scopes = append(scopes, PhotosScopes...)
		case "youtube":
			scopes = append(scopes, YouTubeScopes...)
		case "blogger":
			scopes = append(scopes, BloggerScopes...)
		case "classroom":
			scopes = append(scopes, ClassroomScopes...)
		case "cloud":
			scopes = append(scopes, CloudScopes...)
		case "meet":
			scopes = append(scopes, MeetScopes...)
		case "keep":
			scopes = append(scopes, KeepScopes...)
		case "analytics":
			scopes = append(scopes, AnalyticsScopes...)
		case "admin":
			scopes = append(scopes, AdminScopes...)
		case "fitness":
			scopes = append(scopes, FitnessScopes...)
		case "ads":
			scopes = append(scopes, AdsScopes...)
		case "searchconsole":
			scopes = append(scopes, SearchConsoleScopes...)
		case "bigquery":
			scopes = append(scopes, BigQueryScopes...)
		case "pubsub":
			scopes = append(scopes, PubSubScopes...)
		case "storage":
			scopes = append(scopes, StorageScopes...)
		}
	}

	// Create a temporary config with the requested scopes
	tempConfig := &oauth2.Config{
		ClientID:     p.oauthConfig.ClientID,
		ClientSecret: p.oauthConfig.ClientSecret,
		RedirectURL:  p.oauthConfig.RedirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return tempConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

// GetAuthURLWithAllScopes returns an OAuth URL with ALL available Google scopes.
func (p *Provider) GetAuthURLWithAllScopes(state string) string {
	return p.GetAuthURLWithFeatures(state, []string{"all"})
}

// ExchangeCode exchanges an authorization code for tokens.
func (p *Provider) ExchangeCode(ctx context.Context, code string) (*integrations.TokenResponse, error) {
	token, err := p.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	email, err := p.getUserEmail(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user email: %w", err)
	}

	// Extract scopes
	var scopes []string
	if scopeStr, ok := token.Extra("scope").(string); ok {
		scopes = splitScopes(scopeStr)
	}

	return &integrations.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		Scopes:       scopes,
		AccountEmail: email,
		AccountName:  email, // Use email as name for now
		Metadata: map[string]interface{}{
			"google_email": email,
		},
	}, nil
}

// RefreshToken refreshes an expired access token.
func (p *Provider) RefreshToken(ctx context.Context, refreshToken string) (*integrations.TokenResponse, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := p.oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &integrations.TokenResponse{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		ExpiresAt:    newToken.Expiry,
	}, nil
}

// GetConnectionStatus returns the connection status for a user.
func (p *Provider) GetConnectionStatus(ctx context.Context, userID string) (*integrations.ConnectionStatus, error) {
	// Query the database for connection status
	var status integrations.ConnectionStatus

	err := p.pool.QueryRow(ctx, `
		SELECT
			COALESCE(status = 'connected', false) as connected,
			connected_at,
			external_account_id,
			external_account_name,
			scopes,
			updated_at
		FROM user_integrations
		WHERE user_id = $1 AND provider_id = $2
	`, userID, ProviderID).Scan(
		&status.Connected,
		&status.ConnectedAt,
		&status.AccountID,
		&status.AccountName,
		&status.Scopes,
		&status.LastSyncAt,
	)

	if err != nil {
		// Not connected
		return &integrations.ConnectionStatus{
			Connected: false,
		}, nil
	}

	return &status, nil
}

// Disconnect removes the user's connection to Google.
func (p *Provider) Disconnect(ctx context.Context, userID string) error {
	// Delete from user_integrations
	_, err := p.pool.Exec(ctx, `
		DELETE FROM user_integrations
		WHERE user_id = $1 AND provider_id = $2
	`, userID, ProviderID)
	if err != nil {
		return fmt.Errorf("failed to delete integration: %w", err)
	}

	// Delete from credential_vault
	_, err = p.pool.Exec(ctx, `
		DELETE FROM credential_vault
		WHERE user_id = $1 AND provider_id = $2
	`, userID, ProviderID)
	if err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	return nil
}

// SaveToken saves OAuth tokens to the credential vault.
func (p *Provider) SaveToken(ctx context.Context, userID string, token *integrations.TokenResponse) error {
	// Start transaction
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert or update user_integrations
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
	`, userID, ProviderID, token.AccountEmail, token.AccountName, token.Scopes, token.Metadata)
	if err != nil {
		return fmt.Errorf("failed to save integration: %w", err)
	}

	// Also save to legacy google_oauth_tokens for backward compatibility
	_, err = tx.Exec(ctx, `
		INSERT INTO google_oauth_tokens (
			user_id, access_token, refresh_token, expiry, scopes, google_email
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expiry = EXCLUDED.expiry,
			scopes = EXCLUDED.scopes,
			google_email = EXCLUDED.google_email,
			updated_at = NOW()
	`, userID, token.AccessToken, token.RefreshToken, token.ExpiresAt, token.Scopes, token.AccountEmail)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return tx.Commit(ctx)
}

// GetToken retrieves OAuth tokens for a user.
func (p *Provider) GetToken(ctx context.Context, userID string) (*integrations.Token, error) {
	var token integrations.Token
	var expiry time.Time

	err := p.pool.QueryRow(ctx, `
		SELECT access_token, refresh_token, expiry, scopes
		FROM google_oauth_tokens
		WHERE user_id = $1
	`, userID).Scan(&token.AccessToken, &token.RefreshToken, &expiry, &token.Scopes)

	if err != nil {
		return nil, fmt.Errorf("token not found: %w", err)
	}

	token.ExpiresAt = expiry
	return &token, nil
}

// SupportsSync returns true since Google supports sync.
func (p *Provider) SupportsSync() bool {
	return true
}

// Sync performs a sync operation for the specified resources.
func (p *Provider) Sync(ctx context.Context, userID string, options integrations.SyncOptions) (*integrations.SyncResult, error) {
	result := &integrations.SyncResult{
		Success: true,
	}
	start := time.Now()

	for _, resource := range options.Resources {
		switch resource {
		case "calendar":
			if p.HasFeature("calendar") {
				// Calendar sync will be handled by CalendarService
			}
		case "gmail":
			if p.HasFeature("gmail") {
				// Gmail sync will be handled by GmailService
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// GetOAuth2Token returns an oauth2.Token for use with Google APIs.
func (p *Provider) GetOAuth2Token(ctx context.Context, userID string) (*oauth2.Token, error) {
	token, err := p.GetToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.ExpiresAt,
		TokenType:    "Bearer",
	}, nil
}

// GetTokenSource returns a token source that auto-refreshes.
func (p *Provider) GetTokenSource(ctx context.Context, userID string) (oauth2.TokenSource, error) {
	token, err := p.GetOAuth2Token(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.oauthConfig.TokenSource(ctx, token), nil
}

// Pool returns the database pool.
func (p *Provider) Pool() *pgxpool.Pool {
	return p.pool
}

// OAuthConfig returns the OAuth config.
func (p *Provider) OAuthConfig() *oauth2.Config {
	return p.oauthConfig
}

// Helper functions

func (p *Provider) getUserEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	client := p.oauthConfig.Client(ctx, token)
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

func splitScopes(scopeStr string) []string {
	if scopeStr == "" {
		return nil
	}
	// Scopes are space-separated
	var scopes []string
	for _, s := range split(scopeStr, " ") {
		if s != "" {
			scopes = append(scopes, s)
		}
	}
	return scopes
}

func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}
