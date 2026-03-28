package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// StaticBearerAuth returns a Gin middleware that validates a static Bearer token.
// token is the expected secret (read once at startup, captured in closure).
// If token is empty, returns a pass-through handler (dev mode).
func StaticBearerAuth(token string) gin.HandlerFunc {
	if token == "" {
		return func(c *gin.Context) { c.Next() }
	}

	tokenBytes := []byte(token)
	logger := slog.Default().With("component", "static_bearer_auth")

	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			logger.Debug("missing Authorization header", "path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    ErrCodeUnauthorized,
					"message": "Authorization header required",
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != BearerScheme {
			logger.Debug("malformed Authorization header", "path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    ErrCodeUnauthorized,
					"message": "Authorization header must be: Bearer <token>",
				},
			})
			return
		}

		if subtle.ConstantTimeCompare([]byte(parts[1]), tokenBytes) != 1 {
			logger.Warn("token mismatch", "path", c.Request.URL.Path, "remote_addr", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    ErrCodeUnauthorized,
					"message": "Invalid token",
				},
			})
			return
		}

		c.Next()
	}
}
