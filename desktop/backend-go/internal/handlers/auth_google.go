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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/middleware"
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
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleAuthHandler{
		sessionCache: sessionCache,
		pool:        pool,
		cfg:         cfg,
		oauthConfig: oauthConfig,
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

	// Store state in cookie
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)
	c.SetCookie("oauth_redirect", redirectAfter, 600, "/", "", false, true)

	authURL := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

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
	userID, err := h.upsertUser(c.Request.Context(), userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
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

	// Set session cookie (compatible with Better Auth format)
	c.SetCookie("better-auth.session_token", sessionToken, 60*60*24*7, "/", "", false, true)

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
	var emailVerified bool
	var sessionExpiresAt time.Time

	err = h.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u."emailVerified", u.image, s.id, s."expiresAt"
		FROM session s
		JOIN "user" u ON s."userId" = u.id
		WHERE s.token = $1 AND s."expiresAt" > NOW()
	`, sessionToken).Scan(
		&userID, &userName, &userEmail, &emailVerified, &userImage, &sessionID, &sessionExpiresAt,
	)

	if err != nil {
		log.Printf("[GetCurrentSession] DB query failed: %v, token=%q", err, sessionToken)
		c.JSON(http.StatusOK, gin.H{
			"user":    nil,
			"session": nil,
		})
		return
	}

	log.Printf("[GetCurrentSession] Found user: %s (%s)", userName, userEmail)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":            userID,
			"name":          userName,
			"email":         userEmail,
			"emailVerified": emailVerified,
			"image":         userImage,
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

	// Clear cookie
	c.SetCookie("better-auth.session_token", "", -1, "/", "", false, true)

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

	// Clear current session cookie
	c.SetCookie("better-auth.session_token", "", -1, "/", "", false, true)

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
