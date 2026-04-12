package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rhl/businessos-backend/internal/config"
)

// MIOSAJWTClaims holds the custom claims embedded in MIOSA-issued JWTs.
type MIOSAJWTClaims struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"` // "owner", "admin", "member"
	jwt.RegisteredClaims
}

// MIOSAAuth returns a Gin middleware that validates MIOSA-issued RS256 JWTs.
//
// Behaviour:
//   - When DEPLOYMENT_MODE != "cloud" the middleware is a no-op (passes through).
//   - Expects an "Authorization: Bearer <jwt>" header.
//   - Validates the signature using the RS256 public key from config.
//   - On success sets "miosa_tenant_id" and "miosa_user_id" in the Gin context.
//   - On failure aborts with 401.
func MIOSAAuth(cfg *config.Config) gin.HandlerFunc {
	// In local mode the middleware is disabled; return a pass-through.
	if !cfg.IsCloudDeployment() {
		return func(c *gin.Context) { c.Next() }
	}

	// Parse the public key once at startup so we fail fast on misconfiguration.
	pubKey, err := parseMIOSAPublicKey(cfg.MIOSAJWTPublicKey)
	if err != nil {
		// Log a hard error — cloud mode without a valid key is a misconfiguration.
		slog.Error("miosa_auth: failed to parse MIOSA_JWT_PUBLIC_KEY; all requests will be rejected",
			"error", err,
		)
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "server misconfiguration: invalid JWT public key",
			})
		}
	}

	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" || !strings.HasPrefix(raw, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or malformed Authorization header",
				"code":  "MIOSA_AUTH_MISSING",
			})
			return
		}

		tokenStr := strings.TrimPrefix(raw, "Bearer ")

		claims := &MIOSAJWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return pubKey, nil
		})
		if err != nil || !token.Valid {
			slog.WarnContext(c.Request.Context(), "miosa_auth: JWT validation failed",
				"error", err,
				"path", c.Request.URL.Path,
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired MIOSA JWT",
				"code":  "MIOSA_AUTH_INVALID",
			})
			return
		}

		// Propagate identity into the Gin context for downstream handlers.
		c.Set("miosa_tenant_id", claims.TenantID)
		c.Set("miosa_user_id", claims.UserID)

		slog.DebugContext(c.Request.Context(), "miosa_auth: request authenticated",
			"tenant_id", claims.TenantID,
			"user_id", claims.UserID,
			"role", claims.Role,
		)

		c.Next()
	}
}

// parseMIOSAPublicKey decodes the base64-encoded DER public key from config and
// returns an *rsa.PublicKey ready for jwt.ParseWithClaims.
func parseMIOSAPublicKey(b64Key string) (*rsa.PublicKey, error) {
	if b64Key == "" {
		return nil, jwt.ErrTokenSignatureInvalid
	}

	der, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		// Try URL-safe base64 as a fallback.
		der, err = base64.RawURLEncoding.DecodeString(b64Key)
		if err != nil {
			return nil, err
		}
	}

	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	return rsaPub, nil
}

// GetMIOSATenantID retrieves the tenant ID set by MIOSAAuth middleware.
// Returns ("", false) when not in cloud mode or middleware was not applied.
func GetMIOSATenantID(c *gin.Context) (string, bool) {
	v, exists := c.Get("miosa_tenant_id")
	if !exists {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

// GetMIOSAUserID retrieves the user ID set by MIOSAAuth middleware.
// Returns ("", false) when not in cloud mode or middleware was not applied.
func GetMIOSAUserID(c *gin.Context) (string, bool) {
	v, exists := c.Get("miosa_user_id")
	if !exists {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}
