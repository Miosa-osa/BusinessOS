package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// registerAuthRoutes wires up all authentication routes under /api/auth:
// email sign-up/sign-in, Google OAuth, third-party OAuth integrations
// (Slack, Notion, Microsoft, Linear), session management, and CSRF.
func (h *Handlers) registerAuthRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	// Apply strict rate limiting to prevent brute force attacks
	strictRateLimit := middleware.StrictRateLimitMiddleware()

	googleAuthHandler := NewGoogleAuthHandler(h.pool, h.cfg, h.sessionCache)
	logger := slog.Default()
	emailAuthHandler := NewEmailAuthHandler(h.pool, h.cfg, h.notificationTriggers, h.osaWorkspaceInit, logger)
	authRoutes := api.Group("/auth")
	{
		// Email/Password auth (public) - strict rate limiting
		authRoutes.POST("/sign-up/email", strictRateLimit, emailAuthHandler.SignUp)
		authRoutes.POST("/sign-in/email", strictRateLimit, emailAuthHandler.SignIn)

		// Google OAuth (public) - rate limited to prevent abuse
		authRoutes.GET("/google", strictRateLimit, googleAuthHandler.InitiateGoogleLogin)
		authRoutes.GET("/google/callback/login", strictRateLimit, googleAuthHandler.HandleGoogleLoginCallback)

		// Third-party OAuth integrations (Slack, Notion, Outlook, Linear)
		// Apply strict rate limiting to prevent OAuth abuse/brute-force
		oauthHandler := NewOAuthIntegrationHandler(h.pool, h.cfg)
		authRoutes.GET("/slack", strictRateLimit, oauthHandler.InitiateSlackOAuth)
		authRoutes.GET("/slack/callback", strictRateLimit, oauthHandler.HandleSlackCallback)
		authRoutes.GET("/notion", strictRateLimit, oauthHandler.InitiateNotionOAuth)
		authRoutes.GET("/notion/callback", strictRateLimit, oauthHandler.HandleNotionCallback)
		authRoutes.GET("/microsoft", strictRateLimit, oauthHandler.InitiateMicrosoftOAuth)
		authRoutes.GET("/microsoft/callback", strictRateLimit, oauthHandler.HandleMicrosoftCallback)
		authRoutes.GET("/linear", strictRateLimit, oauthHandler.InitiateLinearOAuth)
		authRoutes.GET("/linear/callback", strictRateLimit, oauthHandler.HandleLinearCallback)

		// Session management (public)
		authRoutes.GET("/session", googleAuthHandler.GetCurrentSession)
		authRoutes.GET("/get-session", googleAuthHandler.GetCurrentSession) // Alias for better-auth compatibility
		authRoutes.POST("/logout", googleAuthHandler.Logout)
		authRoutes.POST("/sign-out", googleAuthHandler.Logout) // Alias for better-auth compatibility

		// CSRF token endpoint (public) - clients retrieve CSRF token before state-changing requests
		authRoutes.GET("/csrf", middleware.CSRFTokenEndpoint())

		// Protected: Force logout all sessions (requires authentication)
		protectedAuth := authRoutes.Group("")
		protectedAuth.Use(auth, middleware.RequireAuth())
		{
			protectedAuth.POST("/logout-all", googleAuthHandler.LogoutAllSessions)
		}
	}
}
