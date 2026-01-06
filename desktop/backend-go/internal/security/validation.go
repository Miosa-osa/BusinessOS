package security

import (
	"errors"
	"fmt"
	"strings"
)

// ConfigValidationError represents a security configuration error
type ConfigValidationError struct {
	Field   string
	Message string
}

func (e ConfigValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateProductionConfig validates security-critical configuration for production
// Returns a list of all validation errors found
func ValidateProductionConfig(
	environment string,
	secretKey string,
	tokenEncryptionKey string,
	redisHMACSecret string,
) []error {
	var errs []error

	if !isProduction(environment) {
		return nil // Skip validation in development
	}

	// SECRET_KEY validation
	if secretKey == "" || secretKey == "your-secret-key-change-this-in-production" {
		errs = append(errs, ConfigValidationError{
			Field:   "SECRET_KEY",
			Message: "must be set to a strong random value in production (use: openssl rand -base64 64)",
		})
	} else if len(secretKey) < 32 {
		errs = append(errs, ConfigValidationError{
			Field:   "SECRET_KEY",
			Message: "must be at least 32 characters in production",
		})
	}

	// TOKEN_ENCRYPTION_KEY validation
	if tokenEncryptionKey == "" {
		errs = append(errs, ConfigValidationError{
			Field:   "TOKEN_ENCRYPTION_KEY",
			Message: "must be set in production for OAuth token encryption (use: openssl rand -base64 32)",
		})
	}

	// REDIS_KEY_HMAC_SECRET validation
	if redisHMACSecret == "" {
		errs = append(errs, ConfigValidationError{
			Field:   "REDIS_KEY_HMAC_SECRET",
			Message: "must be set in production for Redis key derivation (use: openssl rand -base64 32)",
		})
	}

	return errs
}

// ValidateAndFail validates production config and returns a combined error if any validations fail
func ValidateAndFail(
	environment string,
	secretKey string,
	tokenEncryptionKey string,
	redisHMACSecret string,
) error {
	errs := ValidateProductionConfig(environment, secretKey, tokenEncryptionKey, redisHMACSecret)
	if len(errs) == 0 {
		return nil
	}

	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}

	return errors.New("SECURITY CONFIGURATION ERRORS:\n  - " + strings.Join(messages, "\n  - "))
}

func isProduction(env string) bool {
	env = strings.ToLower(env)
	return env == "production" || env == "prod"
}

// WarnDevelopmentInsecure logs a warning about insecure development configuration
func WarnDevelopmentInsecure(
	tokenEncryptionKey string,
	redisHMACSecret string,
) []string {
	var warnings []string

	if tokenEncryptionKey == "" {
		warnings = append(warnings, "TOKEN_ENCRYPTION_KEY not set - OAuth tokens will be stored in plaintext (OK for development)")
	}

	if redisHMACSecret == "" {
		warnings = append(warnings, "REDIS_KEY_HMAC_SECRET not set - using auto-generated key (OK for development)")
	}

	return warnings
}
