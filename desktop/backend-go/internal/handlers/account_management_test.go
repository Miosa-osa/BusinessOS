package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// DeleteAccountRequest validation tests
// ---------------------------------------------------------------------------

func TestDeleteAccountRequest_Unconfirmed(t *testing.T) {
	req := DeleteAccountRequest{Confirm: false}
	assert.False(t, req.Confirm)
}

func TestDeleteAccountRequest_Confirmed(t *testing.T) {
	req := DeleteAccountRequest{Confirm: true}
	assert.True(t, req.Confirm)
}

func TestDeleteAccountRequest_Default(t *testing.T) {
	req := DeleteAccountRequest{}
	assert.False(t, req.Confirm)
}

// ---------------------------------------------------------------------------
// DeleteAccount handler tests (no user in context)
// ---------------------------------------------------------------------------

func TestDeleteAccount_NoUser(t *testing.T) {
	r := gin.New()
	handler := &ProfileHandler{}
	r.POST("/api/account/delete", handler.DeleteAccount)

	body, _ := json.Marshal(DeleteAccountRequest{Confirm: true})
	req := httptest.NewRequest(http.MethodPost, "/api/account/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteAccount_NotConfirmed(t *testing.T) {
	r := gin.New()
	handler := &ProfileHandler{}
	r.POST("/api/account/delete", handler.DeleteAccount)

	body, _ := json.Marshal(DeleteAccountRequest{Confirm: false})
	req := httptest.NewRequest(http.MethodPost, "/api/account/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks user first -> 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteAccount_InvalidJSON(t *testing.T) {
	r := gin.New()
	handler := &ProfileHandler{}
	r.POST("/api/account/delete", handler.DeleteAccount)

	req := httptest.NewRequest(http.MethodPost, "/api/account/delete", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks user first -> 401 (binding check happens after auth check)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// ExportAccountData handler tests (no user in context)
// ---------------------------------------------------------------------------

func TestExportAccountData_NoUser(t *testing.T) {
	r := gin.New()
	handler := &ProfileHandler{}
	r.GET("/api/account/export", handler.ExportAccountData)

	req := httptest.NewRequest(http.MethodGet, "/api/account/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExportAccountData_MethodNotAllowed(t *testing.T) {
	r := gin.New()
	handler := &ProfileHandler{}
	r.GET("/api/account/export", handler.ExportAccountData)

	req := httptest.NewRequest(http.MethodPost, "/api/account/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// ---------------------------------------------------------------------------
// DeleteAccountRequest JSON round-trip
// ---------------------------------------------------------------------------

func TestDeleteAccountRequest_JSONRoundTrip(t *testing.T) {
	original := DeleteAccountRequest{Confirm: true}
	data, err := json.Marshal(original)
	assert.NoError(t, err)

	var decoded DeleteAccountRequest
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, original.Confirm, decoded.Confirm)
}

func TestDeleteAccountRequest_JSONUnmarshal_False(t *testing.T) {
	data := []byte(`{"confirm": false}`)
	var req DeleteAccountRequest
	err := json.Unmarshal(data, &req)
	assert.NoError(t, err)
	assert.False(t, req.Confirm)
}

func TestDeleteAccountRequest_JSONUnmarshal_MissingField(t *testing.T) {
	data := []byte(`{}`)
	var req DeleteAccountRequest
	err := json.Unmarshal(data, &req)
	assert.NoError(t, err)
	assert.False(t, req.Confirm) // Default is false
}
