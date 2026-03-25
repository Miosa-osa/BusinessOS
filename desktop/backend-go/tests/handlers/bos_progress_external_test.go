package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// generateValidJWT creates a valid JWT token for testing
func generateValidJWT(secretKey string, userID string, email string) string {
	claims := &middleware.JWTClaims{
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

// TestProgressEndpointRequiresAuth verifies that /api/bos/progress requires JWT authentication
func TestProgressEndpointRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secretKey := "test-secret-key-12345"

	tests := []struct {
		name         string
		authHeader   string
		expectStatus int
		expectError  string
		description  string
	}{
		{
			name:         "NoToken",
			authHeader:   "",
			expectStatus: http.StatusUnauthorized,
			expectError:  "Missing Authorization header",
			description:  "POST without Authorization header should return 401",
		},
		{
			name:         "InvalidToken",
			authHeader:   "Bearer invalid.token.here",
			expectStatus: http.StatusUnauthorized,
			expectError:  "Invalid or expired token",
			description:  "POST with invalid JWT should return 401",
		},
		{
			name:         "InvalidFormat",
			authHeader:   "Basic dXNlcjpwYXNz",
			expectStatus: http.StatusUnauthorized,
			expectError:  "Invalid Authorization header format",
			description:  "POST with wrong Bearer scheme should return 401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			reqBody := handlers.ReceiveExternalProgressEventRequest{
				Progress:  50,
				Algorithm: "alpha",
				ElapsedMs: 2500,
			}

			bodyBytes, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/bos/progress", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Apply JWT middleware
			jwtMiddleware := middleware.JWTAuth(secretKey)
			jwtMiddleware(c)

			// Verify response
			assert.Equal(t, tt.expectStatus, w.Code, tt.description)

			// Parse response body
			var respBody map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &respBody)
			require.NoError(t, err)

			// Verify error message
			if tt.expectError != "" {
				errorMsg, ok := respBody["error"].(string)
				assert.True(t, ok, "Response should contain 'error' field")
				assert.Contains(t, errorMsg, tt.expectError,
					"Error message should contain expected text: %s", tt.expectError)
			}
		})
	}
}

// TestProgressEndpointAcceptsValidToken verifies that valid JWT is accepted
func TestProgressEndpointAcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secretKey := "test-secret-key-12345"
	userID := "user123"
	email := "test@example.com"

	// Generate valid JWT token
	token := generateValidJWT(secretKey, userID, email)

	// Create test request
	reqBody := handlers.ReceiveExternalProgressEventRequest{
		Progress:  50,
		Algorithm: "alpha",
		ElapsedMs: 2500,
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/bos/progress", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Apply JWT middleware
	jwtMiddleware := middleware.JWTAuth(secretKey)
	jwtMiddleware(c)

	// Verify middleware accepted the request (context not aborted)
	assert.False(t, c.IsAborted(), "Middleware should not abort valid token")
	assert.NotEqual(t, http.StatusUnauthorized, w.Code,
		"Response should not be 401 with valid JWT")

	// Verify JWT claims are in context
	claims := middleware.GetJWTClaims(c)
	require.NotNil(t, claims, "Claims should be in context after JWT validation")
	assert.Equal(t, userID, claims.UserID, "UserID should match token claims")
	assert.Equal(t, email, claims.Email, "Email should match token claims")
}

// TestProgressEndpointWithExpiredToken verifies that expired JWT is rejected
func TestProgressEndpointWithExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secretKey := "test-secret-key-12345"

	// Create expired token
	claims := &middleware.JWTClaims{
		UserID: "user123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(secretKey))
	require.NoError(t, err)

	// Create test request
	reqBody := handlers.ReceiveExternalProgressEventRequest{
		Progress:  50,
		Algorithm: "alpha",
		ElapsedMs: 2500,
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/bos/progress", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Apply JWT middleware
	jwtMiddleware := middleware.JWTAuth(secretKey)
	jwtMiddleware(c)

	// Verify expired token is rejected
	assert.True(t, c.IsAborted(), "Middleware should abort expired token")
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"Response should be 401 with expired JWT")

	// Parse response body
	var respBody map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)

	errorMsg, ok := respBody["error"].(string)
	assert.True(t, ok, "Response should contain 'error' field")
	assert.Contains(t, errorMsg, "Invalid or expired token")
}

// TestProgressEndpointWithWrongSecret verifies that token signed with different secret is rejected
func TestProgressEndpointWithWrongSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalSecret := "original-secret-key"
	differentSecret := "different-secret-key"

	// Generate token with original secret
	token := generateValidJWT(originalSecret, "user123", "test@example.com")

	// Create test request
	reqBody := handlers.ReceiveExternalProgressEventRequest{
		Progress:  50,
		Algorithm: "alpha",
		ElapsedMs: 2500,
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/bos/progress", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Apply JWT middleware with DIFFERENT secret
	jwtMiddleware := middleware.JWTAuth(differentSecret)
	jwtMiddleware(c)

	// Verify token is rejected when validated with wrong secret
	assert.True(t, c.IsAborted(), "Middleware should abort token with wrong secret")
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"Response should be 401 when validated with wrong secret")

	// Parse response body
	var respBody map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)

	errorMsg, ok := respBody["error"].(string)
	assert.True(t, ok, "Response should contain 'error' field")
	assert.Contains(t, errorMsg, "Invalid or expired token")
}
