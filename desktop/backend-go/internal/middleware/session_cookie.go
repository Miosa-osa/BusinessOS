package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SetSessionCookie sets the Better Auth session cookie with environment-dependent configuration
// This centralizes the duplicate cookie-setting logic across all auth handlers
func SetSessionCookie(c *gin.Context, token string) {
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")

	sameSite := http.SameSiteLaxMode
	secure := isProduction

	// For development, explicitly set domain to "localhost" to allow cross-port access
	// and use SameSite=None with Secure=false (allowed for localhost)
	if !isProduction {
		if domain == "" {
			domain = "localhost" // Explicitly set for cross-port compatibility
		}
		sameSite = http.SameSiteNoneMode
		secure = false // Chrome/Safari allow SameSite=None without Secure for localhost
	} else {
		// Production: use current domain if not specified
		if domain == "" {
			domain = ""
		}
		// Allow explicit cross-origin in production if needed
		if os.Getenv("ALLOW_CROSS_ORIGIN") == "true" {
			sameSite = http.SameSiteNoneMode
			secure = true // Required for SameSite=None in production
		}
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    token,
		Path:     "/",
		Domain:   domain,
		MaxAge:   60 * 60 * 24 * 30, // 30 days - persistent login
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

// ClearSessionCookie removes the Better Auth session cookie with environment-dependent configuration
// This must match the configuration used when setting the cookie
func ClearSessionCookie(c *gin.Context) {
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")

	sameSite := http.SameSiteLaxMode

	// Match the settings used in SetSessionCookie
	if !isProduction {
		if domain == "" {
			domain = "localhost"
		}
		sameSite = http.SameSiteNoneMode
	} else {
		if domain == "" {
			domain = ""
		}
		if os.Getenv("ALLOW_CROSS_ORIGIN") == "true" {
			sameSite = http.SameSiteNoneMode
		}
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
}
