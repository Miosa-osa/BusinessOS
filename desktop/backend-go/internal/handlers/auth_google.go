package handlers

// Google OAuth Configuration & Gmail API Scope Setup
//
// IMPORTANT: This handler requests the gmail.readonly scope during OAuth.
// For this to work, you MUST complete the following steps in Google Cloud Console:
//
// 1. Enable Gmail API:
//    - Go to: https://console.cloud.google.com/apis/library
//    - Search for "Gmail API"
//    - Click "Enable"
//
// 2. Update OAuth Consent Screen:
//    - Go to: https://console.cloud.google.com/apis/credentials/consent
//    - Add "gmail.readonly" to the scopes list
//    - User-facing description: "Access your emails to analyze work patterns and create personalized recommendations"
//
// 3. User Experience:
//    - Users will see Gmail permission request during login
//    - Users CAN decline Gmail access (login continues without it)
//    - Frontend should detect missing scope and prompt re-authentication if email sync needed
//
// 4. Scope Verification:
//    - This handler checks granted_scopes and logs warning if Gmail denied
//    - internal/integrations/google/gmail.go checks IsConnected() before sync
//    - Returns error: "Gmail access not authorized" if scope missing
//
// 5. Token Refresh:
//    - Refresh tokens include all granted scopes
//    - Existing users upgrading scope: refresh token remains valid
//    - No special handling needed (OAuth2 library handles this)

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// maskToken returns a masked version of the token for safe logging.
// Shows first 8 and last 4 characters, masks the middle.
// Example: "abc123def456ghi789" -> "abc123de****i789"
func maskToken(token string) string {
	if len(token) <= 12 {
		return "****"
	}
	return token[:8] + "****" + token[len(token)-4:]
}

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
			// NOTE: gmail.readonly removed from login flow — it's a sensitive scope
			// that requires Google verification. Gmail sync will request its own
			// scope via incremental auth when the user enables email integration.
			// See: internal/handlers/oauth_integrations.go for integration-specific scopes
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleAuthHandler{
		sessionCache: sessionCache,
		pool:         pool,
		cfg:          cfg,
		oauthConfig:  oauthConfig,
	}
}

// isValidRedirectURL validates that the redirect URL is safe
func isValidRedirectURL(redirectURL string) bool {
	if redirectURL == "" {
		return false
	}

	// Allow internal paths (e.g., "/dashboard")
	if strings.HasPrefix(redirectURL, "/") && !strings.HasPrefix(redirectURL, "//") {
		return true
	}

	// Allow absolute URLs to known frontend origins (for dev + production)
	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" && strings.HasPrefix(redirectURL, origin) {
			return true
		}
	}

	return false
}

// InitiateGoogleLogin starts the Google OAuth flow for login
func (h *GoogleAuthHandler) InitiateGoogleLogin(c *gin.Context) {
	// Generate random state for CSRF protection
	state := generateRandomState()

	// Get redirect URL from query (for desktop app flow)
	redirectAfter := c.Query("redirect")

	// SECURITY: Validate redirect URL to prevent open redirect attacks
	if !isValidRedirectURL(redirectAfter) {
		slog.Warn("InitiateGoogleLogin: invalid redirect URL blocked", "redirect", redirectAfter)
		redirectAfter = "/dashboard"
	}

	// Store state in cookie with strict security settings
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	c.SetCookie("oauth_state", state, 600, "/", "", isProduction, true)
	c.SetCookie("oauth_redirect", redirectAfter, 600, "/", "", isProduction, true)

	// Force Google to show account picker every time (don't auto-login)
	authURL := h.oauthConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "select_account"))

	// Redirect to Google OAuth
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleGoogleLoginCallback handles the OAuth callback for login
func (h *GoogleAuthHandler) HandleGoogleLoginCallback(c *gin.Context) {
	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		utils.RespondBadRequest(c, slog.Default(), "Invalid state parameter")
		return
	}

	// Get redirect URL
	redirectAfter, _ := c.Cookie("oauth_redirect")

	// SECURITY: Validate redirect URL to prevent open redirect attacks
	if !isValidRedirectURL(redirectAfter) {
		slog.Warn("HandleGoogleLoginCallback: invalid redirect URL blocked", "redirect", redirectAfter)
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
		utils.RespondInternalError(c, slog.Default(), "exchange OAuth code", err)
		return
	}

	// Check granted scopes (OAuth2 token may contain this info)
	// If gmail.readonly scope was not granted, log warning but continue
	grantedScopes := []string{}
	if scopeInterface := token.Extra("scope"); scopeInterface != nil {
		if scopeStr, ok := scopeInterface.(string); ok {
			grantedScopes = strings.Split(scopeStr, " ")
		}
	}

	hasGmailScope := false
	for _, scope := range grantedScopes {
		if strings.Contains(scope, "gmail.readonly") {
			hasGmailScope = true
			break
		}
	}

	if !hasGmailScope {
		slog.Warn("Gmail scope not granted by user",
			"granted_scopes", grantedScopes)
		// Continue anyway - user can still use basic auth without Gmail
	}

	// Get user info from Google
	userInfo, err := h.getGoogleUserInfo(token.AccessToken)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "get user info from Google", err)
		return
	}

	// Create or update user in database
	userID, err := h.upsertUser(c.Request.Context(), userInfo)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create/update user", err)
		return
	}

	// Create session
	sessionToken, err := h.createSession(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create session", err)
		return
	}

	// Clear OAuth cookies and set session cookie
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	c.SetCookie("oauth_state", "", -1, "/", "", isProduction, true)
	c.SetCookie("oauth_redirect", "", -1, "/", "", isProduction, true)
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "" // Current domain
	}

	// SECURITY: Always use SameSite=Strict in production for CSRF protection
	// In development, use Lax for easier testing across localhost ports
	sameSite := http.SameSiteStrictMode
	if !isProduction {
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    sessionToken,
		Path:     "/",
		Domain:   domain,
		MaxAge:   60 * 60 * 24 * 7, // 7 days
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
func (h *GoogleAuthHandler) upsertUser(ctx context.Context, info *GoogleUserInfo) (string, error) {
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
			return "", fmt.Errorf("failed to update user: %w", err)
		}
		return existingID, nil
	}

	// Create new user
	userID := generateUserID()
	_, err = h.pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", image, "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`, userID, info.Name, info.Email, info.VerifiedEmail, info.Picture)

	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// createSession creates a new session for the user
func (h *GoogleAuthHandler) createSession(ctx context.Context, userID string) (string, error) {
	sessionToken := generateSessionToken()
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

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
		c.JSON(http.StatusOK, gin.H{
			"user":    nil,
			"session": nil,
		})
		return
	}

	// SECURITY: Never log session tokens, even masked versions in debug mode
	// Session tokens are sensitive credentials that can be used for account takeover

	// URL-decode the cookie (consistent with auth middleware)
	sessionCookie, err = url.QueryUnescape(sessionCookie)
	if err != nil {
		slog.Warn("get_session: URL decode failed", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"user":    nil,
			"session": nil,
		})
		return
	}

	// Strip signature part if present (consistent with auth middleware)
	sessionToken := sessionCookie
	if idx := strings.Index(sessionCookie, "."); idx != -1 {
		sessionToken = sessionCookie[:idx]
	}

	// Look up session
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var userID, userName, userEmail, sessionID string
	var userImage *string
	var emailVerified bool
	var sessionExpiresAt time.Time
	var userCreatedAt time.Time

	err = h.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u."emailVerified", u.image, u."createdAt", s.id, s."expiresAt"
		FROM session s
		JOIN "user" u ON s."userId" = u.id
		WHERE s.token = $1 AND s."expiresAt" > NOW()
	`, sessionToken).Scan(
		&userID, &userName, &userEmail, &emailVerified, &userImage, &userCreatedAt, &sessionID, &sessionExpiresAt,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"user":    nil,
			"session": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":            userID,
			"name":          userName,
			"email":         userEmail,
			"emailVerified": emailVerified,
			"image":         userImage,
			"createdAt":     userCreatedAt,
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
		cacheInvalidated := false
		if h.sessionCache != nil {
			if err := h.sessionCache.Invalidate(c.Request.Context(), sessionToken); err != nil {
				slog.Warn("Logout: cache invalidation error", "error", err)
			} else {
				cacheInvalidated = true
			}
		}

		// Delete session from database
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		_, err := h.pool.Exec(ctx, `DELETE FROM session WHERE token = $1`, sessionToken)
		if err != nil {
			slog.Error("Logout: database session deletion failed", "error", err)
			// SECURITY: If DB delete fails but cache was invalidated, session is partially logged out
			// This is a security concern - the session may still be valid in DB but not in cache
			if cacheInvalidated {
				slog.Warn("Logout: inconsistent state - cache invalidated but DB delete failed")
			}
			utils.RespondInternalError(c, slog.Default(), "logout session", err)
			return
		}
	}

	// Clear cookie with strict security configuration (must match how it was set)
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "" // Current domain
	}

	// SECURITY: Match SameSite mode used when setting cookie
	sameSite := http.SameSiteStrictMode
	if !isProduction {
		sameSite = http.SameSiteLaxMode
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
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	user, ok := userInterface.(*middleware.BetterAuthUser)
	if !ok {
		utils.RespondInternalError(c, slog.Default(), "get user context", fmt.Errorf("invalid user context type"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Invalidate all Redis cached sessions first
	if h.sessionCache != nil {
		if err := h.sessionCache.InvalidateUserSessions(ctx, user.ID); err != nil {
			slog.Warn("LogoutAllSessions: cache invalidation error", "user_id", user.ID, "error", err)
			// Continue to database cleanup even if cache fails
		}
	}

	// Delete all sessions from database
	result, err := h.pool.Exec(ctx, `DELETE FROM session WHERE "userId" = $1`, user.ID)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "invalidate sessions", err)
		return
	}

	rowsAffected := result.RowsAffected()
	slog.Info("LogoutAllSessions: sessions invalidated", "sessions_deleted", rowsAffected, "user_id", user.ID)

	// Clear current session cookie with strict security configuration (must match how it was set)
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "" // Current domain
	}

	// SECURITY: Match SameSite mode used when setting cookie
	sameSite := http.SameSiteStrictMode
	if !isProduction {
		sameSite = http.SameSiteLaxMode
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
	if _, err := rand.Read(b); err != nil {
		// CRITICAL: rand.Read failure means cryptographic randomness is compromised
		panic(fmt.Sprintf("crypto/rand.Read failed: %v - system entropy exhausted", err))
	}
	return base64.URLEncoding.EncodeToString(b)
}

func generateUserID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// CRITICAL: rand.Read failure means cryptographic randomness is compromised
		panic(fmt.Sprintf("crypto/rand.Read failed: %v - system entropy exhausted", err))
	}
	return base64.URLEncoding.EncodeToString(b)[:22]
}

func generateSessionToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// CRITICAL: rand.Read failure means cryptographic randomness is compromised
		panic(fmt.Sprintf("crypto/rand.Read failed: %v - system entropy exhausted", err))
	}
	return base64.URLEncoding.EncodeToString(b)
}

func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// CRITICAL: rand.Read failure means cryptographic randomness is compromised
		panic(fmt.Sprintf("crypto/rand.Read failed: %v - system entropy exhausted", err))
	}
	return base64.URLEncoding.EncodeToString(b)[:22]
}
