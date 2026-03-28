package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticBearerAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := "super-secret-canopy-token"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/bos/discover", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req

	mw := StaticBearerAuth(token)
	mw(c)

	assert.False(t, c.IsAborted(), "middleware must not abort valid token")
}

func TestStaticBearerAuth_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/bos/discover", nil)
	c.Request = req

	mw := StaticBearerAuth("any-secret")
	mw(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
}

func TestStaticBearerAuth_WrongToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/bos/discover", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	c.Request = req

	mw := StaticBearerAuth("correct-token")
	mw(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStaticBearerAuth_EmptyConfigToken_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/bos/discover", nil)
	c.Request = req

	mw := StaticBearerAuth("") // dev mode
	mw(c)

	assert.False(t, c.IsAborted(), "dev mode must pass through all requests")
}
