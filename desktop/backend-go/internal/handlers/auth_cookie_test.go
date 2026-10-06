package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
)

func TestSessionCookieSameSiteSupportsPackagedDesktop(t *testing.T) {
	if got := sessionCookieSameSite(true); got != http.SameSiteNoneMode {
		t.Fatalf("production cookie must support app:// Electron requests, got %v", got)
	}
	if got := sessionCookieSameSite(false); got != http.SameSiteLaxMode {
		t.Fatalf("development cookie should remain SameSite=Lax, got %v", got)
	}
}

func TestSetSessionCookieIssuesSignedProductionCookie(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("COOKIE_DOMAIN", ".businessos.dev")
	t.Setenv("SECRET_KEY", "test-session-secret")
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	if _, err := setSessionCookie(ctx, "fresh-token"); err != nil {
		t.Fatalf("setSessionCookie returned an error: %v", err)
	}

	response := recorder.Result()
	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one session cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != middleware.SessionCookieName {
		t.Fatalf("unexpected cookie name: %s", cookie.Name)
	}
	if !strings.HasPrefix(cookie.Value, "fresh-token.") {
		t.Fatalf("production session cookie was not signed: %q", cookie.Value)
	}
	if cookie.Domain != "businessos.dev" {
		t.Fatalf("unexpected cookie domain: %q", cookie.Domain)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteNoneMode {
		t.Fatalf("unexpected production cookie attributes: %#v", cookie)
	}
}
