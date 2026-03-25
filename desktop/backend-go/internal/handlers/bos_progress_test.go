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
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

// generateTestJWT creates a JWT token for testing
func generateTestJWT(secretKey string, userID string, email string) string {
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
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}

// TestBOSProgressEndpoint_RequiresJWT verifies /api/bos/progress requires valid JWT
func TestBOSProgressEndpoint_RequiresJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the route with JWT middleware
	secretKey := "test-secret-key-12345"
	jwtAuth := middleware.JWTAuth(secretKey)
	router.POST("/api/bos/progress", jwtAuth, ReceiveExternalProgressEventHandler)

	requestBody := ReceiveExternalProgressEventRequest{
		Progress:  50,
		Algorithm: "alpha",
		ElapsedMs: 2500,
	}
	body, _ := json.Marshal(requestBody)

	t.Run("WithoutJWT_Returns401", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/bos/progress", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Should reject request without JWT")
	})

	t.Run("WithValidJWT_Returns200", func(t *testing.T) {
		token := generateTestJWT(secretKey, "service_pm4py", "pm4py@internal")
		req := httptest.NewRequest("POST", "/api/bos/progress", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should be 200 (OK) from handler, not 401 from JWT
		assert.NotEqual(t, http.StatusUnauthorized, w.Code, "Should accept request with valid JWT")
	})

	t.Run("WithInvalidJWT_Returns401", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/bos/progress", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer invalid-token-xyz")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Should reject request with invalid JWT")
	})
}

// TestBOSProgressEndpoint_JWTClaimsInContext verifies JWT claims are stored in context
func TestBOSProgressEndpoint_JWTClaimsInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	secretKey := "test-secret-key-12345"
	jwtAuth := middleware.JWTAuth(secretKey)

	// Custom handler that checks context
	router.POST("/api/test/jwt-context", jwtAuth, func(c *gin.Context) {
		claims := middleware.GetJWTClaims(c)
		if claims == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "claims not in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id": claims.UserID,
			"email":   claims.Email,
		})
	})

	token := generateTestJWT(secretKey, "test_user_123", "test@example.com")
	req := httptest.NewRequest("POST", "/api/test/jwt-context", nil)
	req.Header.Set("Authorization", "Bearer " + token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Request should succeed")

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test_user_123", response["user_id"], "UserID should match")
	assert.Equal(t, "test@example.com", response["email"], "Email should match")
}
