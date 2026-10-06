package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWorkspaceInviteValidationDoesNotRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authCalled := false
	auth := func(c *gin.Context) {
		authCalled = true
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	api := router.Group("/api")
	RegisterWorkspaceRoutes(api, &WorkspaceHandler{}, auth)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/workspaces/invites/validate",
		strings.NewReader(`{"token":"test-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if authCalled {
		t.Fatal("invite validation must be public so recipients can preview the workspace before signing in")
	}
}
