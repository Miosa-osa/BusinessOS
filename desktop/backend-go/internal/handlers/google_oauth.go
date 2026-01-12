package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// GoogleOAuthHandler handles Google OAuth endpoints
type GoogleOAuthHandler struct {
	calendarService *services.GoogleCalendarService
}

// NewGoogleOAuthHandler creates a new Google OAuth handler
func NewGoogleOAuthHandler(calendarService *services.GoogleCalendarService) *GoogleOAuthHandler {
	return &GoogleOAuthHandler{
		calendarService: calendarService,
	}
}

// InitiateGoogleAuth initiates the Google OAuth flow
func (h *GoogleOAuthHandler) InitiateGoogleAuth(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Generate random state for CSRF protection
	state := generateRandomState()

	// Store state in session/cookie for verification
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)
	c.SetCookie("oauth_user", user.ID, 600, "/", "", false, true)

	authURL := h.calendarService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleGoogleCallback handles the OAuth callback from Google
func (h *GoogleOAuthHandler) HandleGoogleCallback(c *gin.Context) {
	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Get user ID from cookie
	userID, err := c.Cookie("oauth_user")
	if err != nil || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User session not found"})
		return
	}

	// Check for error from Google
	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/settings?google_error="+errMsg)
		return
	}

	// Exchange code for tokens
	code := c.Query("code")
	token, err := h.calendarService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code"})
		return
	}

	// Get user's email from Google
	email, err := getGoogleUserEmail(token.AccessToken)
	if err != nil {
		email = ""
	}

	// Save tokens to database
	if err := h.calendarService.SaveToken(c.Request.Context(), userID, token, email); err != nil {
		// If token already exists, update it
		if err := h.calendarService.UpdateToken(c.Request.Context(), userID, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
			return
		}
	}

	// Clear OAuth cookies
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)
	c.SetCookie("oauth_user", "", -1, "/", "", false, true)

	// Redirect to settings page with success
	c.Redirect(http.StatusTemporaryRedirect, "/settings?google_connected=true")
}

// GetGoogleConnectionStatus returns the Google connection status for a user
func (h *GoogleOAuthHandler) GetGoogleConnectionStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	status, err := h.calendarService.GetConnectionStatus(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":    true,
		"email":        status.GoogleEmail,
		"connected_at": status.CreatedAt,
	})
}

// DisconnectGoogle disconnects the user's Google account
func (h *GoogleOAuthHandler) DisconnectGoogle(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Delete OAuth tokens
	if err := h.calendarService.DeleteToken(c.Request.Context(), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect Google account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Google account disconnected"})
}

// Helper functions

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getGoogleUserEmail(accessToken string) (string, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", err
	}

	return userInfo.Email, nil
}
