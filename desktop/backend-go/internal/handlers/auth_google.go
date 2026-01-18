package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	integrations "github.com/rhl/businessos-backend/internal/integrations/google"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleAuthHandler handles Google OAuth for authentication
type GoogleAuthHandler struct {
	pool         *pgxpool.Pool
	cfg          *config.Config
	oauthConfig  *oauth2.Config
	sessionCache *middleware.SessionCache // Redis session cache for horizontal scaling
}

// GoogleUserInfo represents the user info from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// NewGoogleAuthHandler creates a new Google Auth handler
func NewGoogleAuthHandler(pool *pgxpool.Pool, cfg *config.Config, sessionCache *middleware.SessionCache) *GoogleAuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://mail.google.com/", // Full Gmail access (read, send, edit, delete)
		},
		Endpoint: google.Endpoint,
	}

	// Debug: Log the redirect URI being used
	log.Printf("🔧 [DEBUG] Google OAuth RedirectURL configured as: %s", cfg.GoogleRedirectURI)

	return &GoogleAuthHandler{
		sessionCache: sessionCache,
		pool:         pool,
		cfg:          cfg,
		oauthConfig:  oauthConfig,
	}
}

// InitiateGoogleLogin starts the Google OAuth flow for login
func (h *GoogleAuthHandler) InitiateGoogleLogin(c *gin.Context) {
	// Generate random state for CSRF protection
	state := generateRandomState()

	// Get redirect URL from query (for desktop app flow)
	redirectAfter := c.Query("redirect")
	if redirectAfter == "" {
		redirectAfter = "/dashboard"
	}

	// Store state in cookie with SameSite=Lax for OAuth flow
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_redirect", redirectAfter, 600, "/", "", false, true)

	// Force Google to show account picker every time (don't auto-login)
	authURL := h.oauthConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "select_account"))

	// Debug: Log the OAuth URL being generated
	log.Printf("🔧 [DEBUG] Generated OAuth URL with redirect_uri: %s", h.oauthConfig.RedirectURL)

	// Redirect to Google OAuth
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleGoogleLoginCallback handles the OAuth callback for login
func (h *GoogleAuthHandler) HandleGoogleLoginCallback(c *gin.Context) {
	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Get redirect URL
	redirectAfter, _ := c.Cookie("oauth_redirect")
	if redirectAfter == "" {
		redirectAfter = "/dashboard"
	}

	// Check for error from Google
	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/?error="+errMsg)
		return
	}

	// Exchange code for tokens
	code := c.Query("code")
	token, err := h.oauthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code: " + err.Error()})
		return
	}

	// Get user info from Google
	userInfo, err := h.getGoogleUserInfo(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	// Create or update user in database
	userID, isNewUser, err := h.upsertUser(c.Request.Context(), userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	// Store OAuth tokens for Gmail access + trigger background analysis
	if err := h.storeGmailTokensAndStartAnalysis(c.Request.Context(), userID, token); err != nil {
		log.Printf("⚠️  [WARNING] Failed to store Gmail tokens or start analysis: %v", err)
		// Don't fail the login - continue to session creation
		// Background analysis can be retried later
	}

	// Create session
	sessionToken, err := h.createSession(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session: " + err.Error()})
		return
	}

	// Clear OAuth cookies
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)
	c.SetCookie("oauth_redirect", "", -1, "/", "", false, true)

	// Set a temporary cookie to indicate if this is a new user
	// Frontend will check this to decide whether to redirect to onboarding
	if isNewUser {
		c.SetCookie("new_user", "true", 60, "/", "", false, true) // 60 seconds, enough for redirect
	}

	// Set session cookie with environment-dependent configuration
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "" // Current domain
	}

	sameSite := http.SameSiteLaxMode // Secure default for production
	if os.Getenv("ALLOW_CROSS_ORIGIN") == "true" {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    sessionToken,
		Path:     "/",
		Domain:   domain,
		MaxAge:   60 * 60 * 24 * 30, // 30 days - persistent login
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})

	// Redirect to app
	c.Redirect(http.StatusTemporaryRedirect, redirectAfter)
}

// getGoogleUserInfo fetches user info from Google API
func (h *GoogleAuthHandler) getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// upsertUser creates or updates a user based on Google info
// Returns (userID, isNewUser, error)
func (h *GoogleAuthHandler) upsertUser(ctx context.Context, info *GoogleUserInfo) (string, bool, error) {
	// Check if user exists by email
	var existingID string
	err := h.pool.QueryRow(ctx, `
		SELECT id FROM "user" WHERE email = $1
	`, info.Email).Scan(&existingID)

	if err == nil {
		// User exists, update their info
		_, err = h.pool.Exec(ctx, `
			UPDATE "user"
			SET name = $1, image = $2, "emailVerified" = $3, "updatedAt" = NOW()
			WHERE id = $4
		`, info.Name, info.Picture, info.VerifiedEmail, existingID)
		if err != nil {
			return "", false, fmt.Errorf("failed to update user: %w", err)
		}
		return existingID, false, nil // Existing user
	}

	// Create new user (with onboarding_completed = false by default)
	userID := generateUserID()
	_, err = h.pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", image, onboarding_completed, "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, $5, FALSE, NOW(), NOW())
	`, userID, info.Name, info.Email, info.VerifiedEmail, info.Picture)

	if err != nil {
		return "", false, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, true, nil // New user
}

// createSession creates a new session for the user
func (h *GoogleAuthHandler) createSession(ctx context.Context, userID string) (string, error) {
	sessionToken := generateSessionToken()
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days - persistent login

	_, err := h.pool.Exec(ctx, `
		INSERT INTO session (id, "userId", token, "expiresAt", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, sessionID, userID, sessionToken, expiresAt)

	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

// GetCurrentSession returns the current user session
func (h *GoogleAuthHandler) GetCurrentSession(c *gin.Context) {
	sessionCookie, err := c.Cookie("better-auth.session_token")
	if err != nil || sessionCookie == "" {
		log.Printf("[GetCurrentSession] No cookie found, err=%v, cookie=%q", err, sessionCookie)
		c.JSON(http.StatusOK, gin.H{
			"user":    nil,
			"session": nil,
		})
		return
	}

	log.Printf("[GetCurrentSession] Raw cookie: %q", sessionCookie)

	// URL-decode the cookie (consistent with auth middleware)
	sessionCookie, err = url.QueryUnescape(sessionCookie)
	if err != nil {
		log.Printf("[GetCurrentSession] URL decode failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"user":    nil,
			"session": nil,
		})
		return
	}

	log.Printf("[GetCurrentSession] Decoded cookie: %q", sessionCookie)

	// Strip signature part if present (consistent with auth middleware)
	sessionToken := sessionCookie
	if idx := strings.Index(sessionCookie, "."); idx != -1 {
		sessionToken = sessionCookie[:idx]
	}

	log.Printf("[GetCurrentSession] Token after strip: %q", sessionToken)

	// Look up session
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var userID, userName, userEmail, sessionID string
	var userImage *string
	var emailVerified, onboardingCompleted bool
	var sessionExpiresAt time.Time

	err = h.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u."emailVerified", u.image, COALESCE(u.onboarding_completed, FALSE), s.id, s."expiresAt"
		FROM session s
		JOIN "user" u ON s."userId" = u.id
		WHERE s.token = $1 AND s."expiresAt" > NOW()
	`, sessionToken).Scan(
		&userID, &userName, &userEmail, &emailVerified, &userImage, &onboardingCompleted, &sessionID, &sessionExpiresAt,
	)

	if err != nil {
		log.Printf("[GetCurrentSession] DB query failed: %v, token=%q", err, sessionToken)
		c.JSON(http.StatusOK, gin.H{
			"user":    nil,
			"session": nil,
		})
		return
	}

	log.Printf("[GetCurrentSession] Found user: %s (%s), onboarding: %v", userName, userEmail, onboardingCompleted)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                  userID,
			"name":                userName,
			"email":               userEmail,
			"emailVerified":       emailVerified,
			"image":               userImage,
			"onboardingCompleted": onboardingCompleted,
		},
		"session": gin.H{
			"id":        sessionID,
			"userId":    userID,
			"expiresAt": sessionExpiresAt,
		},
	})
}

// Logout clears the current session
func (h *GoogleAuthHandler) Logout(c *gin.Context) {
	sessionCookie, err := c.Cookie("better-auth.session_token")
	if err == nil && sessionCookie != "" {
		// URL decode the cookie
		sessionCookie, _ = url.QueryUnescape(sessionCookie)

		// Extract token (before HMAC signature dot)
		sessionToken := sessionCookie
		if idx := strings.Index(sessionCookie, "."); idx != -1 {
			sessionToken = sessionCookie[:idx]
		}

		// Invalidate Redis cache first (if available)
		if h.sessionCache != nil {
			if err := h.sessionCache.Invalidate(c.Request.Context(), sessionToken); err != nil {
				log.Printf("Logout: cache invalidation error: %v", err)
			}
		}

		// Delete session from database
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		h.pool.Exec(ctx, `DELETE FROM session WHERE token = $1`, sessionToken)
	}

	// Clear cookie with environment-dependent configuration (must match how it was set)
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "" // Current domain
	}

	sameSite := http.SameSiteLaxMode // Secure default for production
	if os.Getenv("ALLOW_CROSS_ORIGIN") == "true" {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// LogoutAllSessions invalidates all sessions for the current user
// This is a critical security feature for:
// - Password changes
// - Suspected account compromise
// - Permission/role changes
// - User-initiated "logout from all devices"
func (h *GoogleAuthHandler) LogoutAllSessions(c *gin.Context) {
	// Get current user from context (set by auth middleware)
	userInterface, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user, ok := userInterface.(*middleware.BetterAuthUser)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Invalidate all Redis cached sessions first
	if h.sessionCache != nil {
		if err := h.sessionCache.InvalidateUserSessions(ctx, user.ID); err != nil {
			log.Printf("LogoutAllSessions: cache invalidation error for user %s: %v", user.ID, err)
			// Continue to database cleanup even if cache fails
		}
	}

	// Delete all sessions from database
	result, err := h.pool.Exec(ctx, `DELETE FROM session WHERE "userId" = $1`, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invalidate sessions"})
		return
	}

	rowsAffected := result.RowsAffected()
	log.Printf("LogoutAllSessions: deleted %d database sessions for user %s", rowsAffected, user.ID)

	// Clear current session cookie with environment-dependent configuration (must match how it was set)
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "" // Current domain
	}

	sameSite := http.SameSiteLaxMode // Secure default for production
	if os.Getenv("ALLOW_CROSS_ORIGIN") == "true" {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":          "All sessions invalidated",
		"sessions_removed": rowsAffected,
	})
}

// Helper functions
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateUserID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:22]
}

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:22]
}

// storeGmailTokensAndStartAnalysis stores OAuth tokens and triggers background Gmail analysis
func (h *GoogleAuthHandler) storeGmailTokensAndStartAnalysis(ctx context.Context, userID string, token *oauth2.Token) error {
	log.Printf("📧 [Gmail] Storing tokens and starting analysis for user: %s", userID)

	// Extract scopes from token
	scopes := []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://mail.google.com/", // Full Gmail access
	}
	if scopeStr, ok := token.Extra("scope").(string); ok && scopeStr != "" {
		scopes = strings.Split(scopeStr, " ")
	}

	// Store tokens in user_integrations table
	// NOTE: access_token_encrypted and refresh_token_encrypted require encryption
	// For now, we store plain tokens (TODO: add encryption using TOKEN_ENCRYPTION_KEY)
	_, err := h.pool.Exec(ctx, `
		INSERT INTO user_integrations (
			user_id,
			provider_id,
			status,
			access_token_encrypted,
			refresh_token_encrypted,
			token_expires_at,
			scopes,
			external_account_id,
			external_account_name,
			metadata,
			connected_at,
			created_at,
			updated_at
		) VALUES (
			$1, 'google_gmail', 'connected',
			$2::bytea, $3::bytea, $4,
			$5, $6, $7,
			'{"source": "onboarding_oauth"}'::jsonb,
			NOW(), NOW(), NOW()
		)
		ON CONFLICT (user_id, provider_id)
		DO UPDATE SET
			status = 'connected',
			access_token_encrypted = EXCLUDED.access_token_encrypted,
			refresh_token_encrypted = EXCLUDED.refresh_token_encrypted,
			token_expires_at = EXCLUDED.token_expires_at,
			scopes = EXCLUDED.scopes,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
	`, userID, []byte(token.AccessToken), []byte(token.RefreshToken), token.Expiry, scopes, userID, userID)

	if err != nil {
		return fmt.Errorf("failed to store Gmail tokens: %w", err)
	}

	log.Printf("✅ [Gmail] Tokens stored successfully for user: %s", userID)

	// Trigger background Gmail analysis in a goroutine (non-blocking)
	go func() {
		// Use background context with timeout (not the request context which will be cancelled)
		analysisCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		log.Printf("🔍 [Gmail Analysis] Starting background analysis for user: %s", userID)

		// Run the actual Gmail analysis
		if err := h.runGmailAnalysis(analysisCtx, userID); err != nil {
			log.Printf("❌ [Gmail Analysis] Failed for user %s: %v", userID, err)
			// Store error in analysis table
			h.pool.Exec(analysisCtx, `
				INSERT INTO onboarding_user_analysis (
					user_id, workspace_id, status, error_message,
					analysis_model, ai_provider, created_at, updated_at
				)
				VALUES ($1, $2, 'failed', $3, 'n/a', 'groq', NOW(), NOW())
				ON CONFLICT (user_id, workspace_id) DO UPDATE SET
					status = 'failed',
					error_message = EXCLUDED.error_message,
					updated_at = NOW()
			`, userID, "00000000-0000-0000-0000-000000000000", err.Error())
		} else {
			log.Printf("✅ [Gmail Analysis] Completed successfully for user: %s", userID)
		}
	}()

	return nil
}

// runGmailAnalysis performs the actual Gmail analysis
func (h *GoogleAuthHandler) runGmailAnalysis(ctx context.Context, userID string) error {
	log.Printf("📧 [Analysis] Initializing Gmail services for user: %s", userID)

	// Create Google Provider for Gmail access
	googleProvider := integrations.NewProvider(h.pool, []string{"gmail"})

	// Create Gmail service
	gmailService := integrations.NewGmailService(googleProvider)

	// Create email analyzer service
	emailAnalyzer := services.NewEmailAnalyzerService(h.pool, gmailService)

	log.Printf("📊 [Analysis] Analyzing recent emails for user: %s", userID)

	// Analyze recent emails (last 100 emails)
	metadata, err := emailAnalyzer.AnalyzeRecentEmails(ctx, userID, 100)
	if err != nil {
		return fmt.Errorf("failed to analyze emails: %w", err)
	}

	log.Printf("✅ [Analysis] Email analysis complete: %d emails analyzed, %d tools detected",
		metadata.TotalEmails, len(metadata.DetectedTools))

	// Store analysis results in database
	return h.storeAnalysisResults(ctx, userID, metadata)
}

// storeAnalysisResults stores the analysis results in onboarding_user_analysis table
func (h *GoogleAuthHandler) storeAnalysisResults(ctx context.Context, userID string, metadata *services.EmailAnalysisMetadata) error {
	log.Printf("💾 [Analysis] Storing results for user: %s", userID)

	// Convert metadata to JSON
	topicsJSON, _ := json.Marshal(metadata.TopicFrequency)
	domainsJSON, _ := json.Marshal(metadata.SenderDomains)

	// Create insights array (first 3 detected patterns)
	insights := []string{}
	for topic := range metadata.TopicFrequency {
		if len(insights) < 3 {
			insights = append(insights, topic)
		}
	}
	insightsJSON, _ := json.Marshal(insights)

	// Tools used array
	tools := []string{}
	for tool := range metadata.DetectedTools {
		tools = append(tools, tool)
	}
	toolsUsedJSON, _ := json.Marshal(tools)

	// Store in database
	_, err := h.pool.Exec(ctx, `
		INSERT INTO onboarding_user_analysis (
			user_id,
			workspace_id,
			insights,
			tools_used,
			total_emails_analyzed,
			sender_domains,
			detected_patterns,
			analysis_model,
			ai_provider,
			status,
			created_at,
			updated_at,
			completed_at
		)
		VALUES (
			$1,
			'00000000-0000-0000-0000-000000000000',
			$2::jsonb,
			$3::jsonb,
			$4,
			$5::jsonb,
			$6::jsonb,
			'email-metadata',
			'system',
			'completed',
			NOW(),
			NOW(),
			NOW()
		)
		ON CONFLICT (user_id, workspace_id) DO UPDATE SET
			insights = EXCLUDED.insights,
			tools_used = EXCLUDED.tools_used,
			total_emails_analyzed = EXCLUDED.total_emails_analyzed,
			sender_domains = EXCLUDED.sender_domains,
			detected_patterns = EXCLUDED.detected_patterns,
			status = 'completed',
			updated_at = NOW(),
			completed_at = NOW()
	`, userID, insightsJSON, toolsUsedJSON, metadata.TotalEmails, domainsJSON, topicsJSON)

	if err != nil {
		return fmt.Errorf("failed to store analysis: %w", err)
	}

	log.Printf("✅ [Analysis] Results stored successfully for user: %s", userID)
	return nil
}
