package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// maskToken tests
// ---------------------------------------------------------------------------

func TestMaskToken_Short(t *testing.T) {
	// Tokens <= 12 chars should be fully masked
	result := maskToken("short")
	assert.Equal(t, "****", result)
}

func TestMaskToken_Empty(t *testing.T) {
	result := maskToken("")
	assert.Equal(t, "****", result)
}

func TestMaskToken_ExactBoundary(t *testing.T) {
	// Exactly 12 chars
	result := maskToken("12chars_tok")
	assert.Equal(t, "****", result)
}

func TestMaskToken_LongToken(t *testing.T) {
	result := maskToken("abcdefghijklmnopqrst")
	assert.Equal(t, "abcdefgh****qrst", result)
}

func TestMaskToken_VeryLongToken(t *testing.T) {
	token := "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := maskToken(token)
	assert.Equal(t, token[:8]+"****"+token[len(token)-4:], result)
}

// ---------------------------------------------------------------------------
// isValidRedirectURL tests
// ---------------------------------------------------------------------------

func TestIsValidRedirectURL_Empty(t *testing.T) {
	assert.False(t, isValidRedirectURL(""))
}

func TestIsValidRedirectURL_InternalPath(t *testing.T) {
	assert.True(t, isValidRedirectURL("/dashboard"))
	assert.True(t, isValidRedirectURL("/"))
	assert.True(t, isValidRedirectURL("/settings/profile"))
}

func TestIsValidRedirectURL_DoubleSlashBlocked(t *testing.T) {
	// Protocol-relative URLs should be blocked
	assert.False(t, isValidRedirectURL("//evil.com"))
	assert.False(t, isValidRedirectURL("//evil.com/dashboard"))
}

func TestIsValidRedirectURL_AbsoluteWithoutAllowlist(t *testing.T) {
	// Without ALLOWED_ORIGINS set, absolute URLs should be rejected
	assert.False(t, isValidRedirectURL("https://evil.com"))
	assert.False(t, isValidRedirectURL("http://evil.com"))
}

// ---------------------------------------------------------------------------
// generateRandomState tests
// ---------------------------------------------------------------------------

func TestGenerateRandomState_Length(t *testing.T) {
	state := generateRandomState()
	// base64 of 32 bytes = 44 chars
	assert.Len(t, state, 44)
}

func TestGenerateRandomState_Unique(t *testing.T) {
	state1 := generateRandomState()
	state2 := generateRandomState()
	assert.NotEqual(t, state1, state2)
}

func TestGenerateRandomState_URLSafe(t *testing.T) {
	state := generateRandomState()
	// base64.URLEncoding should not contain + or /
	assert.NotContains(t, state, "+")
	assert.NotContains(t, state, "/")
}

// ---------------------------------------------------------------------------
// generateUserID tests
// ---------------------------------------------------------------------------

func TestGenerateUserID_Length(t *testing.T) {
	id := generateUserID()
	// base64 of 16 bytes = 24 chars, truncated to 22
	assert.Len(t, id, 22)
}

func TestGenerateUserID_Unique(t *testing.T) {
	id1 := generateUserID()
	id2 := generateUserID()
	assert.NotEqual(t, id1, id2)
}

// ---------------------------------------------------------------------------
// generateSessionToken tests
// ---------------------------------------------------------------------------

func TestGenerateSessionToken_Length(t *testing.T) {
	token := generateSessionToken()
	// base64 of 32 bytes = 44 chars
	assert.Len(t, token, 44)
}

func TestGenerateSessionToken_Unique(t *testing.T) {
	t1 := generateSessionToken()
	t2 := generateSessionToken()
	assert.NotEqual(t, t1, t2)
}

// ---------------------------------------------------------------------------
// generateSessionID tests
// ---------------------------------------------------------------------------

func TestGenerateSessionID_Length(t *testing.T) {
	id := generateSessionID()
	// base64 of 16 bytes = 24 chars, truncated to 22
	assert.Len(t, id, 22)
}

func TestGenerateSessionID_Unique(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()
	assert.NotEqual(t, id1, id2)
}

// ---------------------------------------------------------------------------
// GoogleUserInfo struct tests
// ---------------------------------------------------------------------------

func TestGoogleUserInfo_Fields(t *testing.T) {
	info := GoogleUserInfo{
		ID:            "google-123",
		Email:         "user@gmail.com",
		VerifiedEmail: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "https://example.com/pic.jpg",
	}
	assert.Equal(t, "google-123", info.ID)
	assert.Equal(t, "user@gmail.com", info.Email)
	assert.True(t, info.VerifiedEmail)
	assert.Equal(t, "Test User", info.Name)
	assert.Equal(t, "Test", info.GivenName)
	assert.Equal(t, "User", info.FamilyName)
	assert.Equal(t, "https://example.com/pic.jpg", info.Picture)
}

func TestGoogleUserInfo_Defaults(t *testing.T) {
	info := GoogleUserInfo{}
	assert.Empty(t, info.ID)
	assert.Empty(t, info.Email)
	assert.False(t, info.VerifiedEmail)
}
