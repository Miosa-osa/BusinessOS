package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BetterAuthUser represents a user from Better Auth's user table
type BetterAuthUser struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Image         *string   `json:"image"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

const (
	UserContextKey    = "user"
	SessionCookieName = "better-auth.session_token"
)

// AuthMiddleware validates Better Auth session from cookie
func AuthMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// DEV MODE: Bypass auth for local development
		devBypass := os.Getenv("DEV_AUTH_BYPASS")
		log.Printf("[AuthMiddleware] DEV_AUTH_BYPASS=%q", devBypass)
		if devBypass == "true" {
			devUser := &BetterAuthUser{
				ID:            "iV1MnkDZLcQRh9HkII3Q_Q",
				Name:          "Roberto Huacuja Luna",
				Email:         "roberto@lunivate.com",
				EmailVerified: true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			c.Set(UserContextKey, devUser)
			c.Set("user_id", devUser.ID)
			c.Next()
			return
		}

		// Get session token from cookie
		sessionCookie, err := c.Cookie(SessionCookieName)
		if err != nil || sessionCookie == "" {
			log.Printf("[AuthMiddleware] No cookie found, err=%v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		log.Printf("[AuthMiddleware] Raw cookie: %q", sessionCookie)

		// URL decode in case it's encoded
		sessionCookie, err = url.QueryUnescape(sessionCookie)
		if err != nil {
			log.Printf("[AuthMiddleware] URL decode failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid session cookie"})
			return
		}

		log.Printf("[AuthMiddleware] Decoded cookie: %q", sessionCookie)

		// Better Auth signs cookies with HMAC - format is {token}.{signature}
		// Extract just the token part (before the dot)
		sessionToken := sessionCookie
		if idx := strings.Index(sessionCookie, "."); idx != -1 {
			sessionToken = sessionCookie[:idx]
		}

		log.Printf("[AuthMiddleware] Token after strip: %q", sessionToken)

		// Look up session in Better Auth's session table
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var user BetterAuthUser
		err = pool.QueryRow(ctx, `
			SELECT u.id, u.name, u.email, u."emailVerified", u.image, u."createdAt", u."updatedAt"
			FROM session s
			JOIN "user" u ON s."userId" = u.id
			WHERE s.token = $1 AND s."expiresAt" > NOW()
		`, sessionToken).Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.EmailVerified,
			&user.Image,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			log.Printf("[AuthMiddleware] DB query failed: %v, token=%q", err, sessionToken)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			return
		}

		log.Printf("[AuthMiddleware] Found user: %s (%s)", user.Name, user.Email)

		// Store user in context
		c.Set(UserContextKey, &user)
		// Also set user_id as string for integration handlers
		c.Set("user_id", user.ID)
		c.Next()
	}
}

// GetCurrentUser retrieves the authenticated user from context
func GetCurrentUser(c *gin.Context) *BetterAuthUser {
	user, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}
	return user.(*BetterAuthUser)
}

// OptionalAuthMiddleware allows unauthenticated requests but sets user if authenticated
func OptionalAuthMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionCookie, err := c.Cookie(SessionCookieName)
		if err != nil || sessionCookie == "" {
			c.Next()
			return
		}

		sessionCookie, err = url.QueryUnescape(sessionCookie)
		if err != nil {
			c.Next()
			return
		}

		sessionToken := sessionCookie
		if idx := strings.Index(sessionCookie, "."); idx != -1 {
			sessionToken = sessionCookie[:idx]
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var user BetterAuthUser
		err = pool.QueryRow(ctx, `
			SELECT u.id, u.name, u.email, u."emailVerified", u.image, u."createdAt", u."updatedAt"
			FROM session s
			JOIN "user" u ON s."userId" = u.id
			WHERE s.token = $1 AND s."expiresAt" > NOW()
		`, sessionToken).Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.EmailVerified,
			&user.Image,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err == nil {
			c.Set(UserContextKey, &user)
		}
		c.Next()
	}
}
