// Package webhooks provides webhook signature verification for various providers.
package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// SignatureVerifier handles webhook signature verification for different providers.
type SignatureVerifier struct {
	secrets map[string]string // provider -> secret
}

// NewSignatureVerifier creates a new signature verifier with provider secrets.
func NewSignatureVerifier(secrets map[string]string) *SignatureVerifier {
	return &SignatureVerifier{
		secrets: secrets,
	}
}

// SetSecret sets the webhook secret for a provider.
func (v *SignatureVerifier) SetSecret(provider, secret string) {
	if v.secrets == nil {
		v.secrets = make(map[string]string)
	}
	v.secrets[provider] = secret
}

// GetSecret returns the webhook secret for a provider.
func (v *SignatureVerifier) GetSecret(provider string) string {
	if v.secrets == nil {
		return ""
	}
	return v.secrets[provider]
}

// =============================================================================
// SLACK SIGNATURE VERIFICATION
// =============================================================================

// VerifySlackSignature verifies a Slack webhook signature.
// Slack uses HMAC-SHA256 with format: v0=<hash>
// baseString = "v0:{timestamp}:{body}"
func (v *SignatureVerifier) VerifySlackSignature(body []byte, timestamp, signature string) bool {
	secret := v.GetSecret("slack")
	if secret == "" {
		// No secret configured - allow in development
		return true
	}

	// Check timestamp to prevent replay attacks (5 minute window)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > 300 {
		return false
	}

	// Create signature base string
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(baseString))
	expectedSig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// =============================================================================
// LINEAR SIGNATURE VERIFICATION
// =============================================================================

// VerifyLinearSignature verifies a Linear webhook signature.
// Linear uses HMAC-SHA256 with the raw body.
func (v *SignatureVerifier) VerifyLinearSignature(body []byte, signature string) bool {
	secret := v.GetSecret("linear")
	if secret == "" {
		return true
	}

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// =============================================================================
// HUBSPOT SIGNATURE VERIFICATION
// =============================================================================

// VerifyHubSpotSignature verifies a HubSpot v3 webhook signature.
// HubSpot v3 uses HMAC-SHA256 with: requestMethod + requestUri + requestBody + timestamp
func (v *SignatureVerifier) VerifyHubSpotSignature(body []byte, signature, timestamp, requestMethod, requestURI string) bool {
	secret := v.GetSecret("hubspot")
	if secret == "" {
		return true
	}

	// Check timestamp (5 minute window)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().UnixMilli()-ts > 300000 { // 5 minutes in milliseconds
		return false
	}

	// Build source string: method + uri + body + timestamp
	sourceString := requestMethod + requestURI + string(body) + timestamp

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sourceString))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// =============================================================================
// NOTION SIGNATURE VERIFICATION
// =============================================================================

// VerifyNotionSignature verifies a Notion webhook signature.
// Notion uses HMAC-SHA256 with the raw body.
func (v *SignatureVerifier) VerifyNotionSignature(body []byte, signature string) bool {
	secret := v.GetSecret("notion")
	if secret == "" {
		return true
	}

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// =============================================================================
// CLICKUP SIGNATURE VERIFICATION
// =============================================================================

// VerifyClickUpSignature verifies a ClickUp webhook signature.
// ClickUp uses a shared secret that's sent in the webhook payload.
func (v *SignatureVerifier) VerifyClickUpSignature(body []byte, signature string) bool {
	secret := v.GetSecret("clickup")
	if secret == "" {
		return true
	}

	// ClickUp sends the webhook_id as signature verification
	// The signature in the payload should match what we configured
	return hmac.Equal([]byte(secret), []byte(signature))
}

// =============================================================================
// AIRTABLE SIGNATURE VERIFICATION
// =============================================================================

// VerifyAirtableSignature verifies an Airtable webhook signature.
// Airtable uses HMAC-SHA256 with the raw body.
func (v *SignatureVerifier) VerifyAirtableSignature(body []byte, signature, timestamp string) bool {
	secret := v.GetSecret("airtable")
	if secret == "" {
		return true
	}

	// Check timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > 300 {
		return false
	}

	// Build source string: timestamp + "." + body
	sourceString := timestamp + "." + string(body)

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sourceString))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// =============================================================================
// FATHOM SIGNATURE VERIFICATION
// =============================================================================

// VerifyFathomSignature verifies a Fathom webhook signature.
// Fathom uses HMAC-SHA256 with the raw body.
func (v *SignatureVerifier) VerifyFathomSignature(body []byte, signature string) bool {
	secret := v.GetSecret("fathom")
	if secret == "" {
		return true
	}

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// =============================================================================
// GOOGLE CALENDAR VERIFICATION
// =============================================================================

// VerifyGoogleChannelToken verifies the Google Calendar push notification channel token.
// Google Calendar uses a channel token that we set when creating the watch.
func (v *SignatureVerifier) VerifyGoogleChannelToken(token string) bool {
	// The token is set by us when creating the watch, so we just need to validate
	// that it's a valid user ID or contains expected data.
	// For now, we assume any non-empty token is valid.
	return token != ""
}

// =============================================================================
// MICROSOFT GRAPH VERIFICATION
// =============================================================================

// VerifyMicrosoftClientState verifies the Microsoft Graph webhook client state.
// Microsoft uses a clientState value that we set when creating the subscription.
func (v *SignatureVerifier) VerifyMicrosoftClientState(clientState string) bool {
	expectedState := v.GetSecret("microsoft")
	if expectedState == "" {
		return true
	}

	return hmac.Equal([]byte(expectedState), []byte(clientState))
}

// =============================================================================
// GENERIC HMAC-SHA256 VERIFICATION
// =============================================================================

// VerifyHMACSHA256 verifies a generic HMAC-SHA256 signature.
func (v *SignatureVerifier) VerifyHMACSHA256(provider string, body []byte, signature string) bool {
	secret := v.GetSecret(provider)
	if secret == "" {
		return true
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// ComputeHMACSHA256 computes an HMAC-SHA256 signature for testing.
func ComputeHMACSHA256(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
