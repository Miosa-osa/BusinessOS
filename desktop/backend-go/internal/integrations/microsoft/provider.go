// Package microsoft provides Microsoft 365 integration (Outlook, OneDrive, Teams, etc.).
package microsoft

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/integrations"
	"golang.org/x/oauth2"
)

const (
	ProviderID   = "microsoft"
	ProviderName = "Microsoft 365"
	Category     = "productivity"

	// Microsoft OAuth endpoints
	AuthURL  = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	TokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	// Microsoft Graph API base URL
	GraphAPIBase = "https://graph.microsoft.com/v1.0"
)

// Microsoft Graph API Scopes - COMPREHENSIVE ACCESS
var (
	// User profile
	UserScopes = []string{
		"User.Read",              // Read user profile
		"User.ReadBasic.All",     // Read basic profiles of all users
		"User.Read.All",          // Read all users' full profiles
		"User.ReadWrite",         // Read and update user profile
	}

	// Mail/Outlook
	MailScopes = []string{
		"Mail.Read",              // Read user mail
		"Mail.ReadBasic",         // Read user mail (basic)
		"Mail.ReadWrite",         // Read and write user mail
		"Mail.Send",              // Send mail as user
		"MailboxSettings.Read",   // Read mailbox settings
		"MailboxSettings.ReadWrite", // Read and write mailbox settings
	}

	// Calendar
	CalendarScopes = []string{
		"Calendars.Read",         // Read user calendars
		"Calendars.Read.Shared",  // Read shared calendars
		"Calendars.ReadWrite",    // Read and write user calendars
		"Calendars.ReadWrite.Shared", // Read and write shared calendars
	}

	// Contacts
	ContactsScopes = []string{
		"Contacts.Read",          // Read user contacts
		"Contacts.Read.Shared",   // Read shared contacts
		"Contacts.ReadWrite",     // Read and write user contacts
		"Contacts.ReadWrite.Shared", // Read and write shared contacts
	}

	// OneDrive/Files
	FilesScopes = []string{
		"Files.Read",             // Read user files
		"Files.Read.All",         // Read all files user can access
		"Files.ReadWrite",        // Read and write user files
		"Files.ReadWrite.All",    // Read and write all files user can access
		"Files.Read.Selected",    // Read files selected by user
		"Files.ReadWrite.Selected", // Read and write files selected by user
	}

	// To Do Tasks
	TasksScopes = []string{
		"Tasks.Read",             // Read user tasks
		"Tasks.Read.Shared",      // Read shared tasks
		"Tasks.ReadWrite",        // Read and write user tasks
		"Tasks.ReadWrite.Shared", // Read and write shared tasks
	}

	// OneNote
	OneNoteScopes = []string{
		"Notes.Read",             // Read OneNote notebooks
		"Notes.Read.All",         // Read all OneNote notebooks
		"Notes.ReadWrite",        // Read and write OneNote notebooks
		"Notes.ReadWrite.All",    // Read and write all OneNote notebooks
		"Notes.Create",           // Create OneNote notebooks
	}

	// Teams
	TeamsScopes = []string{
		"Team.ReadBasic.All",     // Read basic team info
		"TeamSettings.Read.All",  // Read team settings
		"TeamSettings.ReadWrite.All", // Read and write team settings
		"Channel.ReadBasic.All",  // Read channel basic info
		"ChannelMessage.Read.All", // Read channel messages
		"ChannelMessage.Send",    // Send channel messages
		"Chat.Read",              // Read chat messages
		"Chat.ReadWrite",         // Read and write chat messages
		"ChatMessage.Read",       // Read chat messages
		"ChatMessage.Send",       // Send chat messages
	}

	// SharePoint/Sites
	SitesScopes = []string{
		"Sites.Read.All",         // Read all site collections
		"Sites.ReadWrite.All",    // Read and write all site collections
		"Sites.Manage.All",       // Create, edit, delete site collections
		"Sites.FullControl.All",  // Full control of all site collections
	}

	// Groups
	GroupsScopes = []string{
		"Group.Read.All",         // Read all groups
		"Group.ReadWrite.All",    // Read and write all groups
		"GroupMember.Read.All",   // Read group members
		"GroupMember.ReadWrite.All", // Read and write group members
	}

	// Planner
	PlannerScopes = []string{
		"Tasks.Read",             // Read Planner tasks (shared with To Do)
		"Tasks.ReadWrite",        // Read and write Planner tasks
		"Group.Read.All",         // Required for Planner groups
	}

	// Directory
	DirectoryScopes = []string{
		"Directory.Read.All",     // Read directory data
		"Directory.ReadWrite.All", // Read and write directory data
		"Directory.AccessAsUser.All", // Access directory as user
	}

	// People
	PeopleScopes = []string{
		"People.Read",            // Read user's relevant people
		"People.Read.All",        // Read all users' relevant people
	}

	// Bookings
	BookingsScopes = []string{
		"Bookings.Read.All",      // Read booking businesses
		"Bookings.ReadWrite.All", // Read and write booking businesses
		"Bookings.Manage.All",    // Manage booking businesses
		"BookingsAppointment.ReadWrite.All", // Read and write appointments
	}

	// Reports
	ReportsScopes = []string{
		"Reports.Read.All",       // Read all usage reports
	}

	// Security
	SecurityScopes = []string{
		"SecurityEvents.Read.All", // Read security events
		"SecurityEvents.ReadWrite.All", // Read and write security events
	}

	// Audit logs
	AuditScopes = []string{
		"AuditLog.Read.All",      // Read audit logs
	}

	// Offline access (for refresh tokens)
	OfflineScopes = []string{
		"offline_access",         // Get refresh tokens
	}

	// OpenID Connect
	OpenIDScopes = []string{
		"openid",
		"profile",
		"email",
	}

	// AllMicrosoftScopes contains all available Microsoft scopes
	AllMicrosoftScopes []string
)

func init() {
	// Build the complete list of all scopes
	AllMicrosoftScopes = make([]string, 0)
	AllMicrosoftScopes = append(AllMicrosoftScopes, OpenIDScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, OfflineScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, UserScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, MailScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, CalendarScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, ContactsScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, FilesScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, TasksScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, OneNoteScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, TeamsScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, SitesScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, GroupsScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, PlannerScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, DirectoryScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, PeopleScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, BookingsScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, ReportsScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, SecurityScopes...)
	AllMicrosoftScopes = append(AllMicrosoftScopes, AuditScopes...)
}

// Provider implements the integrations.Provider interface for Microsoft 365.
type Provider struct {
	pool        *pgxpool.Pool
	oauthConfig *oauth2.Config
	features    []string
}

// AllFeatures contains all available Microsoft feature identifiers
var AllFeatures = []string{
	"mail", "calendar", "contacts", "files", "tasks",
	"onenote", "teams", "sites", "groups", "planner",
	"directory", "people", "bookings", "reports", "security",
}

// NewProvider creates a new Microsoft provider with specified features.
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
	scopes := append([]string{}, OpenIDScopes...)
	scopes = append(scopes, OfflineScopes...)
	scopes = append(scopes, UserScopes[:2]...) // Basic user scopes

	for _, feature := range features {
		switch feature {
		case "mail":
			scopes = append(scopes, MailScopes...)
		case "calendar":
			scopes = append(scopes, CalendarScopes...)
		case "contacts":
			scopes = append(scopes, ContactsScopes...)
		case "files":
			scopes = append(scopes, FilesScopes...)
		case "tasks":
			scopes = append(scopes, TasksScopes...)
		case "onenote":
			scopes = append(scopes, OneNoteScopes...)
		case "teams":
			scopes = append(scopes, TeamsScopes...)
		case "sites":
			scopes = append(scopes, SitesScopes...)
		case "groups":
			scopes = append(scopes, GroupsScopes...)
		case "planner":
			scopes = append(scopes, PlannerScopes...)
		case "directory":
			scopes = append(scopes, DirectoryScopes...)
		case "people":
			scopes = append(scopes, PeopleScopes...)
		case "bookings":
			scopes = append(scopes, BookingsScopes...)
		case "reports":
			scopes = append(scopes, ReportsScopes...)
		case "security":
			scopes = append(scopes, SecurityScopes...)
		}
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.MicrosoftClientID,
		ClientSecret: cfg.MicrosoftClientSecret,
		RedirectURL:  cfg.MicrosoftRedirectURI,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}

	return &Provider{
		pool:        pool,
		oauthConfig: oauthConfig,
		features:    features,
	}
}

// NewProviderWithAllFeatures creates a provider with ALL Microsoft features enabled.
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
	return "/icons/microsoft.svg"
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
func (p *Provider) GetAuthURLWithFeatures(state string, features []string) string {
	// Check if "all" is requested
	for _, f := range features {
		if f == "all" {
			features = AllFeatures
			break
		}
	}

	scopes := append([]string{}, OpenIDScopes...)
	scopes = append(scopes, OfflineScopes...)
	scopes = append(scopes, UserScopes[:2]...)

	for _, feature := range features {
		switch feature {
		case "mail":
			scopes = append(scopes, MailScopes...)
		case "calendar":
			scopes = append(scopes, CalendarScopes...)
		case "contacts":
			scopes = append(scopes, ContactsScopes...)
		case "files":
			scopes = append(scopes, FilesScopes...)
		case "tasks":
			scopes = append(scopes, TasksScopes...)
		case "onenote":
			scopes = append(scopes, OneNoteScopes...)
		case "teams":
			scopes = append(scopes, TeamsScopes...)
		case "sites":
			scopes = append(scopes, SitesScopes...)
		case "groups":
			scopes = append(scopes, GroupsScopes...)
		case "planner":
			scopes = append(scopes, PlannerScopes...)
		case "directory":
			scopes = append(scopes, DirectoryScopes...)
		case "people":
			scopes = append(scopes, PeopleScopes...)
		case "bookings":
			scopes = append(scopes, BookingsScopes...)
		case "reports":
			scopes = append(scopes, ReportsScopes...)
		case "security":
			scopes = append(scopes, SecurityScopes...)
		}
	}

	tempConfig := &oauth2.Config{
		ClientID:     p.oauthConfig.ClientID,
		ClientSecret: p.oauthConfig.ClientSecret,
		RedirectURL:  p.oauthConfig.RedirectURL,
		Scopes:       scopes,
		Endpoint:     p.oauthConfig.Endpoint,
	}

	return tempConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

// ExchangeCode exchanges an authorization code for tokens.
func (p *Provider) ExchangeCode(ctx context.Context, code string) (*integrations.TokenResponse, error) {
	token, err := p.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Microsoft Graph
	userInfo, err := p.getUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
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
		AccountEmail: userInfo.Email,
		AccountName:  userInfo.DisplayName,
		Metadata: map[string]interface{}{
			"microsoft_id":    userInfo.ID,
			"microsoft_email": userInfo.Email,
			"microsoft_name":  userInfo.DisplayName,
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
		return &integrations.ConnectionStatus{
			Connected: false,
		}, nil
	}

	return &status, nil
}

// Disconnect removes the user's connection to Microsoft.
func (p *Provider) Disconnect(ctx context.Context, userID string) error {
	_, err := p.pool.Exec(ctx, `
		DELETE FROM user_integrations
		WHERE user_id = $1 AND provider_id = $2
	`, userID, ProviderID)
	if err != nil {
		return fmt.Errorf("failed to delete integration: %w", err)
	}

	_, err = p.pool.Exec(ctx, `
		DELETE FROM credential_vault
		WHERE user_id = $1 AND provider_id = $2
	`, userID, ProviderID)
	if err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	return nil
}

// SaveToken saves OAuth tokens.
func (p *Provider) SaveToken(ctx context.Context, userID string, token *integrations.TokenResponse) error {
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

	// Save to microsoft_oauth_tokens table
	_, err = tx.Exec(ctx, `
		INSERT INTO microsoft_oauth_tokens (
			user_id, access_token, refresh_token, expiry, scopes, microsoft_email
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expiry = EXCLUDED.expiry,
			scopes = EXCLUDED.scopes,
			microsoft_email = EXCLUDED.microsoft_email,
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
		FROM microsoft_oauth_tokens
		WHERE user_id = $1
	`, userID).Scan(&token.AccessToken, &token.RefreshToken, &expiry, &token.Scopes)

	if err != nil {
		return nil, fmt.Errorf("token not found: %w", err)
	}

	token.ExpiresAt = expiry
	return &token, nil
}

// SupportsSync returns true since Microsoft supports sync.
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
		case "mail":
			if p.HasFeature("mail") {
				// Mail sync will be handled by OutlookService
			}
		case "calendar":
			if p.HasFeature("calendar") {
				// Calendar sync will be handled by OutlookService
			}
		case "files":
			if p.HasFeature("files") {
				// Files sync will be handled by OneDriveService
			}
		case "tasks":
			if p.HasFeature("tasks") {
				// Tasks sync will be handled by ToDoService
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// GetOAuth2Token returns an oauth2.Token for use with Microsoft APIs.
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

// GetHTTPClient returns an HTTP client with auto-refreshing tokens.
func (p *Provider) GetHTTPClient(ctx context.Context, userID string) (*http.Client, error) {
	tokenSource, err := p.GetTokenSource(ctx, userID)
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(ctx, tokenSource), nil
}

// Pool returns the database pool.
func (p *Provider) Pool() *pgxpool.Pool {
	return p.pool
}

// OAuthConfig returns the OAuth config.
func (p *Provider) OAuthConfig() *oauth2.Config {
	return p.oauthConfig
}

// Helper types and functions

// MicrosoftUser represents basic Microsoft user info.
type MicrosoftUser struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	Email             string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
}

func (p *Provider) getUserInfo(ctx context.Context, token *oauth2.Token) (*MicrosoftUser, error) {
	client := p.oauthConfig.Client(ctx, token)
	resp, err := client.Get(GraphAPIBase + "/me")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo MicrosoftUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	// Use userPrincipalName as email fallback
	if userInfo.Email == "" {
		userInfo.Email = userInfo.UserPrincipalName
	}

	return &userInfo, nil
}

func splitScopes(scopeStr string) []string {
	if scopeStr == "" {
		return nil
	}
	var scopes []string
	for _, s := range splitString(scopeStr, " ") {
		if s != "" {
			scopes = append(scopes, s)
		}
	}
	return scopes
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}
