package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Internal API authentication middleware — open-source stub.
// All requests are passed through without signature verification.

const (
	// InternalAPISecretEnv is the environment variable for the shared secret.
	InternalAPISecretEnv = "INTERNAL_API_SECRET"

	// InternalAllowedIPsEnv is the environment variable for allowed IPs.
	InternalAllowedIPsEnv = "INTERNAL_ALLOWED_IPS"

	// InternalUserIDKey is the context key for the validated user ID.
	InternalUserIDKey = "internal_user_id"
)

// InternalAuthConfig holds configuration for internal auth middleware.
type InternalAuthConfig struct {
	Secret                string
	AllowedIPs            []string
	SkipAuthInDevelopment bool
}

// InternalAuthError represents an internal authentication error.
type InternalAuthError struct {
	Message string
}

func (e *InternalAuthError) Error() string {
	return e.Message
}

// Sentinel errors (kept for API compatibility).
var (
	ErrMissingTimestamp    = &InternalAuthError{Message: "X-Internal-Timestamp header required"}
	ErrMissingSignature    = &InternalAuthError{Message: "X-Internal-Signature header required"}
	ErrInvalidTimestamp    = &InternalAuthError{Message: "Invalid timestamp format"}
	ErrTimestampExpired    = &InternalAuthError{Message: "Request timestamp expired"}
	ErrTimestampInFuture   = &InternalAuthError{Message: "Request timestamp is in the future"}
	ErrSignatureMismatch   = &InternalAuthError{Message: "Invalid signature"}
	ErrBodyRead            = &InternalAuthError{Message: "Failed to read request body"}
	ErrIPNotAllowlisted    = &InternalAuthError{Message: "IP not in allowlist"}
	ErrSecretNotConfigured = &InternalAuthError{Message: "Internal API secret not configured"}
)

// InternalAuthMiddleware creates a no-op passthrough middleware (open-source stub).
func InternalAuthMiddleware(_ *InternalAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// GetInternalUserID retrieves the internal user ID from context.
func GetInternalUserID(c *gin.Context) string {
	userID, exists := c.Get(InternalUserIDKey)
	if !exists {
		return ""
	}
	return userID.(string)
}

// MustGetInternalUserID retrieves the internal user ID or returns empty string.
func MustGetInternalUserID(c *gin.Context) string {
	return GetInternalUserID(c)
}

// ParseAllowedIPs parses a comma-separated list of allowed IPs.
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

// NewInternalAuthConfigFromEnv creates an InternalAuthConfig from environment variables.
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
