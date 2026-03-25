package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupJWTTestContext creates a test context for JWT middleware
func setupJWTTestContext(method string, authHeader string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(method, "/test", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	c.Request = req
	return c, w
}

// generateValidJWT creates a valid JWT token for testing
func generateValidJWT(secretKey string, userID string, email string) string {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		panic(err)
	}
	return tokenString
}

// TestJWTAuth_ValidToken verifies JWTAuth middleware accepts valid token
func TestJWTAuth_ValidToken(t *testing.T) {
	secretKey := "test-secret-key-12345"
	userID := "user123"
	email := "test@example.com"

	// Generate valid token
	token := generateValidJWT(secretKey, userID, email)
	c, w := setupJWTTestContext("POST", "Bearer "+token)

	// Execute middleware
	middleware := JWTAuth(secretKey)
	middleware(c)

	// Assert: request was NOT aborted (200 response)
	assert.False(t, c.IsAborted(), "JWTAuth should not abort valid token")
	assert.NotEqual(t, http.StatusUnauthorized, w.Code, "Response code should not be 401")

	// Assert: claims are in context
	claims := GetJWTClaims(c)
	require.NotNil(t, claims, "Claims should be in context")
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, email, claims.Email, "Email should match")
}

// TestJWTAuth_MissingHeader verifies JWTAuth middleware rejects missing Authorization header
func TestJWTAuth_MissingHeader(t *testing.T) {
	secretKey := "test-secret-key-12345"

	c, w := setupJWTTestContext("POST", "")

	// Execute middleware
	middleware := JWTAuth(secretKey)
	middleware(c)

	// Assert: request was aborted with 401
	assert.True(t, c.IsAborted(), "JWTAuth should abort missing header")
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Response code should be 401")
}

// TestJWTAuth_InvalidFormat verifies JWTAuth middleware rejects invalid header format
func TestJWTAuth_InvalidFormat(t *testing.T) {
	secretKey := "test-secret-key-12345"

	testCases := []string{
		"InvalidToken",                    // Missing Bearer scheme
		"Bearer",                          // Missing token
		"Basic dXNlcjpwYXNz",             // Wrong scheme
		"Bearer token1 token2",            // Multiple tokens
	}

	for _, authHeader := range testCases {
		t.Run(authHeader, func(t *testing.T) {
			c, w := setupJWTTestContext("POST", authHeader)

			// Execute middleware
			middleware := JWTAuth(secretKey)
			middleware(c)

			// Assert: request was aborted with 401
			assert.True(t, c.IsAborted(), "JWTAuth should abort invalid format: %s", authHeader)
			assert.Equal(t, http.StatusUnauthorized, w.Code, "Response code should be 401")
		})
	}
}

// TestJWTAuth_InvalidToken verifies JWTAuth middleware rejects invalid token
func TestJWTAuth_InvalidToken(t *testing.T) {
	secretKey := "test-secret-key-12345"

	testCases := []struct {
		name     string
		token    string
		reason   string
	}{
		{
			name:   "MalformedToken",
			token:  "not.a.valid.token.structure",
			reason: "malformed token",
		},
		{
			name:   "WrongSignature",
			token:  generateValidJWT("different-secret-key", "user123", "test@example.com"),
			reason: "token signed with different secret",
		},
		{
			name:   "ExpiredToken",
			token:  func() string {
				claims := &JWTClaims{
					UserID: "user123",
					Email:  "test@example.com",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secretKey))
				return tokenString
			}(),
			reason: "token is expired",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := setupJWTTestContext("POST", "Bearer "+tc.token)

			// Execute middleware
			middleware := JWTAuth(secretKey)
			middleware(c)

			// Assert: request was aborted with 401
			assert.True(t, c.IsAborted(), "JWTAuth should abort invalid token: %s (%s)", tc.name, tc.reason)
			assert.Equal(t, http.StatusUnauthorized, w.Code, "Response code should be 401")

			// Assert: no claims in context
			claims := GetJWTClaims(c)
			assert.Nil(t, claims, "Claims should not be in context")
		})
	}
}

// TestOptionalJWT_ValidToken verifies OptionalJWT accepts valid token
func TestOptionalJWT_ValidToken(t *testing.T) {
	secretKey := "test-secret-key-12345"
	userID := "user123"
	email := "test@example.com"

	// Generate valid token
	token := generateValidJWT(secretKey, userID, email)
	c, _ := setupJWTTestContext("POST", "Bearer "+token)

	// Execute middleware
	middleware := OptionalJWT(secretKey)
	middleware(c)

	// Assert: request was NOT aborted
	assert.False(t, c.IsAborted(), "OptionalJWT should not abort valid token")

	// Assert: claims are in context
	claims := GetJWTClaims(c)
	require.NotNil(t, claims, "Claims should be in context")
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, email, claims.Email, "Email should match")
}

// TestOptionalJWT_MissingHeader verifies OptionalJWT allows missing Authorization header
func TestOptionalJWT_MissingHeader(t *testing.T) {
	secretKey := "test-secret-key-12345"

	c, _ := setupJWTTestContext("POST", "")

	// Execute middleware
	middleware := OptionalJWT(secretKey)
	middleware(c)

	// Assert: request was NOT aborted (request continues)
	assert.False(t, c.IsAborted(), "OptionalJWT should allow missing header")

	// Assert: no claims in context
	claims := GetJWTClaims(c)
	assert.Nil(t, claims, "Claims should not be in context for unauthenticated request")
}

// TestOptionalJWT_InvalidTokenRejected verifies OptionalJWT still rejects invalid token
func TestOptionalJWT_InvalidTokenRejected(t *testing.T) {
	secretKey := "test-secret-key-12345"

	// Create expired token
	claims := &JWTClaims{
		UserID: "user123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secretKey))

	c, w := setupJWTTestContext("POST", "Bearer "+tokenString)

	// Execute middleware
	middleware := OptionalJWT(secretKey)
	middleware(c)

	// Assert: request was aborted with 401 (even though token was optional, it was invalid)
	assert.True(t, c.IsAborted(), "OptionalJWT should reject invalid token when provided")
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Response code should be 401")

	// Assert: no claims in context
	claimsResult := GetJWTClaims(c)
	assert.Nil(t, claimsResult, "Claims should not be in context")
}

// TestJWTAuth_GetJWTClaims verifies GetJWTClaims retrieves claims correctly
func TestJWTAuth_GetJWTClaims(t *testing.T) {
	secretKey := "test-secret-key-12345"
	userID := "user456"
	email := "user456@example.com"

	// Generate valid token
	token := generateValidJWT(secretKey, userID, email)
	c, _ := setupJWTTestContext("POST", "Bearer "+token)

	// Execute middleware
	middleware := JWTAuth(secretKey)
	middleware(c)

	// Retrieve claims using helper function
	claims := GetJWTClaims(c)
	require.NotNil(t, claims, "GetJWTClaims should return non-nil claims")
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

// TestJWTAuth_GetJWTClaimsNilWhenMissing verifies GetJWTClaims returns nil when no claims
func TestJWTAuth_GetJWTClaimsNilWhenMissing(t *testing.T) {
	c, _ := setupJWTTestContext("POST", "")

	// Don't execute middleware, context has no claims
	claims := GetJWTClaims(c)
	assert.Nil(t, claims, "GetJWTClaims should return nil when no claims in context")
}

// TestJWTAuth_DifferentSecretRejects verifies token signed with different secret is rejected
func TestJWTAuth_DifferentSecretRejects(t *testing.T) {
	originalSecret := "original-secret-key"
	differentSecret := "different-secret-key"

	// Generate token with original secret
	token := generateValidJWT(originalSecret, "user123", "test@example.com")

	// Try to validate with different secret
	c, w := setupJWTTestContext("POST", "Bearer "+token)
	middleware := JWTAuth(differentSecret)
	middleware(c)

	// Assert: request was aborted
	assert.True(t, c.IsAborted(), "JWTAuth should reject token signed with different secret")
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Response code should be 401")
}
