package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// AuthorizationHeader is the standard Authorization header name
	AuthorizationHeader = "Authorization"

	// BearerScheme is the standard Bearer token scheme
	BearerScheme = "Bearer"

	// JWTContextKey is the key used to store JWT claims in context
	JWTContextKey = "jwt_claims"
)

// JWTClaims represents standard JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTAuth returns a middleware that validates Bearer JWT tokens.
// Returns 401 if token is missing, invalid, or expired.
// Valid claims are stored in context under JWTContextKey.
func JWTAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := slog.Default().With("component", "jwt_auth")

		// Extract Authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			logger.Debug("JWT: missing Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing Authorization header",
				"code":  "JWT_MISSING",
			})
			return
		}

		// Parse Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != BearerScheme {
			logger.Debug("JWT: invalid Authorization header format",
				"header", authHeader)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header format. Expected: Bearer <token>",
				"code":  "JWT_INVALID_FORMAT",
			})
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			logger.Debug("JWT: token parsing failed",
				"error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "JWT_INVALID",
			})
			return
		}

		if !token.Valid {
			logger.Debug("JWT: token is invalid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "JWT_INVALID",
			})
			return
		}

		logger.Debug("JWT: token validated successfully",
			"user_id", claims.UserID,
			"email", claims.Email)

		// Store claims in context
		c.Set(JWTContextKey, claims)
		c.Set("user_id", claims.UserID) // Also set for compatibility with other middleware

		c.Next()
	}
}

// OptionalJWT returns a middleware that validates Bearer JWT tokens if present.
// Returns 401 only if token is present but invalid.
// Allows requests without Authorization header.
// Valid claims are stored in context under JWTContextKey if present.
func OptionalJWT(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := slog.Default().With("component", "optional_jwt")

		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			// No token present - allow request to continue
			logger.Debug("OptionalJWT: no Authorization header (request will be unauthenticated)")
			c.Next()
			return
		}

		// Parse Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != BearerScheme {
			logger.Debug("OptionalJWT: invalid Authorization header format",
				"header", authHeader)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header format. Expected: Bearer <token>",
				"code":  "JWT_INVALID_FORMAT",
			})
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			logger.Debug("OptionalJWT: token parsing failed",
				"error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "JWT_INVALID",
			})
			return
		}

		if !token.Valid {
			logger.Debug("OptionalJWT: token is invalid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "JWT_INVALID",
			})
			return
		}

		logger.Debug("OptionalJWT: token validated successfully",
			"user_id", claims.UserID,
			"email", claims.Email)

		// Store claims in context
		c.Set(JWTContextKey, claims)
		c.Set("user_id", claims.UserID)

		c.Next()
	}
}

// GetJWTClaims retrieves JWT claims from context.
// Returns nil if no claims are present (unauthenticated request).
func GetJWTClaims(c *gin.Context) *JWTClaims {
	claims, exists := c.Get(JWTContextKey)
	if !exists {
		return nil
	}
	if claims == nil {
		return nil
	}
	return claims.(*JWTClaims)
}
