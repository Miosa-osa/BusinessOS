package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// sessionCookieSameSite supports both the same-origin web client and the
// packaged app:// Electron renderer. CSRF protection is enforced separately by
// the double-submit cookie and explicit header middleware.
func sessionCookieSameSite(isProduction bool) http.SameSite {
	if isProduction {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func setSessionCookie(c *gin.Context, token string) (string, error) {
	value, err := middleware.SignSessionCookieValue(token)
	if err != nil {
		return "", err
	}
	writeSessionCookie(c, value, os.Getenv("COOKIE_DOMAIN"), 7*24*60*60)
	return value, nil
}

func writeSessionCookie(c *gin.Context, value, domain string, maxAge int) {
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	if !isProduction {
		domain = ""
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    value,
		Path:     "/",
		Domain:   domain,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sessionCookieSameSite(isProduction),
	})
}

func clearSessionCookies(c *gin.Context) {
	// Clear both scopes. A browser may have retained an older host-only cookie
	// alongside the current parent-domain cookie.
	writeSessionCookie(c, "", os.Getenv("COOKIE_DOMAIN"), -1)
	if os.Getenv("COOKIE_DOMAIN") != "" {
		writeSessionCookie(c, "", "", -1)
	}
}

func clearHostOnlySessionCookie(c *gin.Context) {
	writeSessionCookie(c, "", "", -1)
}
