package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Internal API authentication middleware
// Protects internal endpoints (e.g., /api/internal/osa/*) from unauthorized access
// by verifying HMAC-SHA256 signatures with timestamp validation
//
// Security model:
// 1. Primary: HMAC signature verification (X-Internal-Signature header)
// 2. Fallback: IP allowlisting for trusted internal sources
//
// Signature format: HMAC-SHA256(secret, timestamp + method + path + body)
// Headers required:
//   - X-Internal-Timestamp: Unix timestamp (seconds)
//   - X-Internal-Signature: Hex-encoded HMAC signature
//   - X-User-ID: The user ID to impersonate (validated by signature)

const (
	// InternalAPISecretEnv is the environment variable for the shared secret
	InternalAPISecretEnv = "INTERNAL_API_SECRET"
	
	// InternalAllowedIPsEnv is the environment variable for allowed IPs (comma-separated)
	InternalAllowedIPsEnv = "INTERNAL_ALLOWED_IPS"
	
	// TimestampValidityWindow is the maximum age of a request timestamp (5 minutes)
	TimestampValidityWindow = 5 * time.Minute
	
	// Context key for the validated user ID
	InternalUserIDKey = "internal_user_id"
)

// InternalAuthConfig holds configuration for internal auth middleware
type InternalAuthConfig struct {
	// Secret is the shared secret for HMAC signing (required for signature mode)
	Secret string
	
	// AllowedIPs is a list of IPs that can bypass signature verification
	// If empty, signature verification is required
	AllowedIPs []string
	
	// SkipAuthInDevelopment disables auth checks when true (only for local dev)
	SkipAuthInDevelopment bool
}

// InternalAuthMiddleware creates a middleware that validates internal API requests
// using HMAC signature verification with timestamp validation
func InternalAuthMiddleware(cfg *InternalAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth in development if configured (security risk in production!)
		if cfg.SkipAuthInDevelopment && os.Getenv("ENVIRONMENT") == "development" {
			// Still require X-User-ID header for development
			userID := c.GetHeader("X-User-ID")
			if userID == "" {
				slog.Warn("InternalAuthMiddleware: missing X-User-ID in development mode")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "X-User-ID header required",
					"code":  "MISSING_USER_ID",
				})
				return
			}
			c.Set(InternalUserIDKey, userID)
			c.Next()
			return
		}

		// Check IP allowlist first (fallback mechanism)
		if len(cfg.AllowedIPs) > 0 {
			clientIP := c.ClientIP()
			for _, allowedIP := range cfg.AllowedIPs {
				if clientIP == allowedIP {
					slog.Debug("InternalAuthMiddleware: IP allowlisted",
						"ip", clientIP,
						"user_id", c.GetHeader("X-User-ID"))
					
					userID := c.GetHeader("X-User-ID")
					if userID == "" {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
							"error": "X-User-ID header required",
							"code":  "MISSING_USER_ID",
						})
						return
					}
					c.Set(InternalUserIDKey, userID)
					c.Next()
					return
				}
			}
		}

		// Require signature verification if not IP allowlisted
		if cfg.Secret == "" {
			slog.Error("InternalAuthMiddleware: no secret configured and IP not allowlisted",
				"ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Internal authentication not configured",
				"code":  "AUTH_NOT_CONFIGURED",
			})
			return
		}

		// Validate signature
		if err := validateInternalSignature(c, cfg.Secret); err != nil {
			slog.Warn("InternalAuthMiddleware: signature validation failed",
				"ip", c.ClientIP(),
				"error", err.Error(),
				"path", c.Request.URL.Path)
			
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  "INVALID_SIGNATURE",
			})
			return
		}

		// Extract validated user ID
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "X-User-ID header required",
				"code":  "MISSING_USER_ID",
			})
			return
		}

		c.Set(InternalUserIDKey, userID)
		c.Next()
	}
}

// validateInternalSignature verifies the HMAC signature of an internal request
func validateInternalSignature(c *gin.Context, secret string) error {
	// Get required headers
	timestampStr := c.GetHeader("X-Internal-Timestamp")
	signature := c.GetHeader("X-Internal-Signature")
	method := c.Request.Method
	path := c.Request.URL.Path

	if timestampStr == "" {
		return ErrMissingTimestamp
	}
	if signature == "" {
		return ErrMissingSignature
	}

	// Parse and validate timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return ErrInvalidTimestamp
	}

	// Check timestamp is within valid window (prevent replay attacks)
	requestTime := time.Unix(timestamp, 0)
	now := time.Now()
	
	if now.Sub(requestTime) > TimestampValidityWindow {
		return ErrTimestampExpired
	}
	if requestTime.Sub(now) > TimestampValidityWindow {
		return ErrTimestampInFuture
	}

	// Read request body for signature computation
	// Note: We need to preserve the body for downstream handlers
	bodyBytes, err := c.GetRawData()
	if err != nil {
		slog.Error("InternalAuthMiddleware: failed to read request body", "error", err)
		return ErrBodyRead
	}

	// Compute expected signature
	// Signature = HMAC-SHA256(secret, timestamp + method + path + body)
	message := timestampStr + method + path + string(bodyBytes)
	expectedSig := computeHMAC(secret, message)

	// Compare signatures using constant-time comparison
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		slog.Debug("InternalAuthMiddleware: signature mismatch",
			"expected", expectedSig,
			"received", signature)
		return ErrSignatureMismatch
	}

	// Restore body for downstream handlers
	c.Request.Body = &readCloser{body: bodyBytes}

	return nil
}

// computeHMAC generates an HMAC-SHA256 signature
func computeHMAC(secret, message string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// readCloser wraps a byte slice to implement io.ReadCloser
type readCloser struct {
	body   []byte
	offset int
}

func (r *readCloser) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.body) {
		return 0, nil
	}
	n = copy(p, r.body[r.offset:])
	r.offset += n
	return n, nil
}

func (r *readCloser) Close() error {
	return nil
}

// Error types for internal auth
var (
	ErrMissingTimestamp    = &InternalAuthError{Message: "X-Internal-Timestamp header required"}
	ErrMissingSignature    = &InternalAuthError{Message: "X-Internal-Signature header required"}
	ErrInvalidTimestamp    = &InternalAuthError{Message: "Invalid timestamp format"}
	ErrTimestampExpired    = &InternalAuthError{Message: "Request timestamp expired (possible replay attack)"}
	ErrTimestampInFuture   = &InternalAuthError{Message: "Request timestamp is in the future"}
	ErrSignatureMismatch   = &InternalAuthError{Message: "Invalid signature"}
	ErrBodyRead            = &InternalAuthError{Message: "Failed to read request body"}
	ErrIPNotAllowlisted    = &InternalAuthError{Message: "IP not in allowlist"}
	ErrSecretNotConfigured = &InternalAuthError{Message: "Internal API secret not configured"}
)

// InternalAuthError represents an internal authentication error
type InternalAuthError struct {
	Message string
}

func (e *InternalAuthError) Error() string {
	return e.Message
}

// GetInternalUserID retrieves the validated internal user ID from context
// Returns empty string if not set
func GetInternalUserID(c *gin.Context) string {
	userID, exists := c.Get(InternalUserIDKey)
	if !exists {
		return ""
	}
	return userID.(string)
}

// MustGetInternalUserID retrieves the validated internal user ID from context
// Panics if not set (indicates middleware misconfiguration)
// Use only in handlers protected by InternalAuthMiddleware
func MustGetInternalUserID(c *gin.Context) string {
	userID := GetInternalUserID(c)
	if userID == "" {
		slog.Error("BUG: internal user ID not in context despite InternalAuthMiddleware",
			"path", c.Request.URL.Path,
			"method", c.Request.Method)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error: authentication misconfiguration",
		})
		return ""
	}
	return userID
}

// ParseAllowedIPs parses a comma-separated list of allowed IPs
func ParseAllowedIPs(ipList string) []string {
	if ipList == "" {
		return nil
	}
	
	ips := strings.Split(ipList, ",")
	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		trimmed := strings.TrimSpace(ip)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// NewInternalAuthConfigFromEnv creates an InternalAuthConfig from environment variables
func NewInternalAuthConfigFromEnv() *InternalAuthConfig {
	secret := os.Getenv(InternalAPISecretEnv)
	allowedIPs := ParseAllowedIPs(os.Getenv(InternalAllowedIPsEnv))
	env := os.Getenv("ENVIRONMENT")
	
	return &InternalAuthConfig{
		Secret:                secret,
		AllowedIPs:            allowedIPs,
		SkipAuthInDevelopment: env == "development" && secret == "" && len(allowedIPs) == 0,
	}
}
