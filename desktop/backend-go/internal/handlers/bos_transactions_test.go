package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/transactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Test Setup ────────────────────────────────────────────────────────────────

// setupBOSTestDB creates a database connection for testing
func setupBOSTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/businessos_test"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("skipping test: database not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping test: cannot connect to database: %v", err)
	}

	return pool
}

// setupBOSHandler creates a test handler with a test database
func setupBOSHandler(t *testing.T) (*BOSTransactionHandler, *pgxpool.Pool) {
	pool := setupBOSTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	handler := NewBOSTransactionHandler(pool, logger)
	return handler, pool
}

// setupBOSTestRouter creates a Gin router with the handler registered
func setupBOSTestRouter(t *testing.T) (*gin.Engine, *pgxpool.Pool) {
	handler, pool := setupBOSHandler(t)
	router := gin.New()
	router.Use(gin.Recovery())
	api := router.Group("/api")
	handler.RegisterRoutes(api)
	return router, pool
}

// ─── TDD Test 1: Prepare Phase ─────────────────────────────────────────────────

// TestPrepare_Success validates the prepare phase happy path
func TestPrepare_Success(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	// Create request
	req := PrepareRequestPayload{
		TransactionID: "tx-12345-test",
		Algorithm:     "alpha_miner",
		LogData: transactions.LogData{
			Type:     "xes",
			Encoding: "base64",
			Content:  "base64encodedlogcontent",
		},
		Parameters: transactions.AlgorithmParams{
			ActivityKey:  "activity",
			TimestampKey: "timestamp",
			CaseKey:      "case_id",
		},
		TimeoutMS: 30000,
	}

	body, err := json.Marshal(req)
	require.NoError(t, err)

	// Send request
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/bos/tx/prepare", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	// Assertions: HTTP 200
	assert.Equal(t, http.StatusOK, w.Code, "expected HTTP 200, got %d", w.Code)

	// Unmarshal response
	var resp PrepareResponsePayload
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Assertions: Response structure
	assert.NotEmpty(t, resp.TransactionID, "transaction_id should not be empty")
	assert.Equal(t, "prepared", resp.Status, "status should be 'prepared'")
	assert.Equal(t, "YES", resp.Vote, "vote should be 'YES'")
	assert.NotZero(t, resp.Version, "version should be non-zero")
	assert.NotNil(t, resp.Model, "model should not be nil")
	assert.Equal(t, "petri_net", resp.Model.Type, "model type should be 'petri_net'")
	assert.True(t, resp.Timestamp.Before(time.Now().Add(1*time.Second)), "timestamp should be recent")

	// Assertions: X-Request-ID header
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "X-Request-ID header should be set")
}

// TestPrepare_InvalidRequest validates error handling for invalid requests
func TestPrepare_InvalidRequest(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name:    "missing_transaction_id",
			payload: map[string]interface{}{"algorithm": "alpha_miner"},
		},
		{
			name:    "missing_algorithm",
			payload: map[string]interface{}{"transaction_id": "tx-123"},
		},
		{
			name:    "invalid_json",
			payload: "invalid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/bos/tx/prepare", bytes.NewBuffer(body))
			r.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code,
				"expected HTTP 400 for %s, got %d", tt.name, w.Code)
		})
	}
}

// ─── TDD Test 2: Commit Phase ──────────────────────────────────────────────────

// TestCommit_Success validates the commit phase happy path
func TestCommit_Success(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	// First, prepare a transaction
	prepareReq := PrepareRequestPayload{
		TransactionID: "tx-commit-test",
		Algorithm:     "alpha_miner",
		LogData: transactions.LogData{
			Type:     "xes",
			Encoding: "base64",
			Content:  "base64encodedlogcontent",
		},
		Parameters: transactions.AlgorithmParams{
			ActivityKey:  "activity",
			TimestampKey: "timestamp",
			CaseKey:      "case_id",
		},
	}

	body, _ := json.Marshal(prepareReq)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/bos/tx/prepare", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	// Extract transaction ID
	var prepareResp PrepareResponsePayload
	json.Unmarshal(w.Body.Bytes(), &prepareResp)
	txID := prepareResp.TransactionID

	// Now commit
	commitReq := CommitRequestPayload{
		TransactionID: txID,
	}

	body, _ = json.Marshal(commitReq)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/bos/tx/commit", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	// Assertions: HTTP 200
	assert.Equal(t, http.StatusOK, w.Code, "expected HTTP 200, got %d", w.Code)

	// Unmarshal response
	var commitResp CommitResponsePayload
	err := json.Unmarshal(w.Body.Bytes(), &commitResp)
	require.NoError(t, err)

	// Assertions: Response structure
	assert.Equal(t, txID, commitResp.TransactionID, "transaction_id should match")
	assert.Equal(t, "committed", commitResp.Status, "status should be 'committed'")
	assert.NotZero(t, commitResp.Version, "version should be non-zero")
	assert.True(t, commitResp.Timestamp.Before(time.Now().Add(1*time.Second)), "timestamp should be recent")

	// Assertions: X-Request-ID header
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "X-Request-ID header should be set")
}

// TestCommit_InvalidRequest validates error handling
func TestCommit_InvalidRequest(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name:    "missing_transaction_id",
			payload: map[string]interface{}{},
		},
		{
			name:    "invalid_json",
			payload: "not json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/bos/tx/commit", bytes.NewBuffer(body))
			r.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code,
				"expected HTTP 400 for %s, got %d", tt.name, w.Code)
		})
	}
}

// ─── TDD Test 3: Abort Phase ───────────────────────────────────────────────────

// TestAbort_Success validates the abort phase happy path
func TestAbort_Success(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	// First, prepare a transaction
	prepareReq := PrepareRequestPayload{
		TransactionID: "tx-abort-test",
		Algorithm:     "alpha_miner",
		LogData: transactions.LogData{
			Type:     "xes",
			Encoding: "base64",
			Content:  "base64encodedlogcontent",
		},
		Parameters: transactions.AlgorithmParams{
			ActivityKey:  "activity",
			TimestampKey: "timestamp",
			CaseKey:      "case_id",
		},
	}

	body, _ := json.Marshal(prepareReq)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/bos/tx/prepare", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	// Extract transaction ID
	var prepareResp PrepareResponsePayload
	json.Unmarshal(w.Body.Bytes(), &prepareResp)
	txID := prepareResp.TransactionID

	// Now abort
	abortReq := AbortRequestPayload{
		TransactionID: txID,
		Reason:        "participant_failure",
	}

	body, _ = json.Marshal(abortReq)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/bos/tx/abort", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	// Assertions: HTTP 200
	assert.Equal(t, http.StatusOK, w.Code, "expected HTTP 200, got %d", w.Code)

	// Unmarshal response
	var abortResp AbortResponsePayload
	err := json.Unmarshal(w.Body.Bytes(), &abortResp)
	require.NoError(t, err)

	// Assertions: Response structure
	assert.Equal(t, txID, abortResp.TransactionID, "transaction_id should match")
	assert.Equal(t, "aborted", abortResp.Status, "status should be 'aborted'")
	assert.NotZero(t, abortResp.Version, "version should be non-zero")
	assert.True(t, abortResp.Timestamp.Before(time.Now().Add(1*time.Second)), "timestamp should be recent")

	// Assertions: X-Request-ID header
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "X-Request-ID header should be set")
}

// TestAbort_WithoutReason validates abort with optional reason
func TestAbort_WithoutReason(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	// Prepare a transaction
	prepareReq := PrepareRequestPayload{
		TransactionID: "tx-abort-no-reason",
		Algorithm:     "alpha_miner",
		LogData: transactions.LogData{
			Type:     "xes",
			Encoding: "base64",
			Content:  "base64encodedlogcontent",
		},
		Parameters: transactions.AlgorithmParams{
			ActivityKey:  "activity",
			TimestampKey: "timestamp",
			CaseKey:      "case_id",
		},
	}

	body, _ := json.Marshal(prepareReq)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/bos/tx/prepare", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	var prepareResp PrepareResponsePayload
	json.Unmarshal(w.Body.Bytes(), &prepareResp)
	txID := prepareResp.TransactionID

	// Abort without reason
	abortReq := AbortRequestPayload{
		TransactionID: txID,
	}

	body, _ = json.Marshal(abortReq)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/bos/tx/abort", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var abortResp AbortResponsePayload
	json.Unmarshal(w.Body.Bytes(), &abortResp)
	assert.Equal(t, "aborted", abortResp.Status)
}

// TestAbort_InvalidRequest validates error handling
func TestAbort_InvalidRequest(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name:    "missing_transaction_id",
			payload: map[string]interface{}{"reason": "test"},
		},
		{
			name:    "invalid_json",
			payload: "not json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/bos/tx/abort", bytes.NewBuffer(body))
			r.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code,
				"expected HTTP 400 for %s, got %d", tt.name, w.Code)
		})
	}
}

// ─── TDD Test 4: Status Endpoint ───────────────────────────────────────────────

// TestGetStatus_Success validates status query for a transaction
func TestGetStatus_Success(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	// Prepare a transaction
	prepareReq := PrepareRequestPayload{
		TransactionID: "tx-status-test",
		Algorithm:     "alpha_miner",
		LogData: transactions.LogData{
			Type:     "xes",
			Encoding: "base64",
			Content:  "base64encodedlogcontent",
		},
		Parameters: transactions.AlgorithmParams{
			ActivityKey:  "activity",
			TimestampKey: "timestamp",
			CaseKey:      "case_id",
		},
	}

	body, _ := json.Marshal(prepareReq)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/bos/tx/prepare", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	var prepareResp PrepareResponsePayload
	json.Unmarshal(w.Body.Bytes(), &prepareResp)
	txID := prepareResp.TransactionID

	// Query status
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/bos/tx/status/"+txID, nil)
	router.ServeHTTP(w, r)

	// Assertions: HTTP 200
	assert.Equal(t, http.StatusOK, w.Code)

	// Unmarshal response
	var statusResp StatusResponsePayload
	err := json.Unmarshal(w.Body.Bytes(), &statusResp)
	require.NoError(t, err)

	// Assertions: Response structure
	assert.Equal(t, txID, statusResp.TransactionID)
	assert.NotEmpty(t, statusResp.Status)
	assert.False(t, statusResp.StartedAt.IsZero())
}

// TestGetStatus_NotFound validates error for non-existent transaction
func TestGetStatus_NotFound(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/bos/tx/status/tx-nonexistent", nil)
	router.ServeHTTP(w, r)

	// Should return 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ─── Integration Test: Complete 2PC Protocol ───────────────────────────────────

// TestIntegration_Complete2PCProtocol validates the full 2PC workflow
func TestIntegration_Complete2PCProtocol(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	// Phase 1: PREPARE
	prepareReq := PrepareRequestPayload{
		TransactionID: "tx-integration-e2e",
		Algorithm:     "alpha_miner",
		LogData: transactions.LogData{
			Type:     "xes",
			Encoding: "base64",
			Content:  "base64encodedlogcontent",
		},
		Parameters: transactions.AlgorithmParams{
			ActivityKey:  "activity",
			TimestampKey: "timestamp",
			CaseKey:      "case_id",
		},
	}

	body, _ := json.Marshal(prepareReq)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/bos/tx/prepare", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code, "prepare should succeed")
	var prepareResp PrepareResponsePayload
	json.Unmarshal(w.Body.Bytes(), &prepareResp)
	txID := prepareResp.TransactionID

	assert.Equal(t, "prepared", prepareResp.Status)
	assert.Equal(t, "YES", prepareResp.Vote)

	// Phase 2: CHECK STATUS (PREPARED)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/bos/tx/status/"+txID, nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var statusResp StatusResponsePayload
	json.Unmarshal(w.Body.Bytes(), &statusResp)
	assert.NotEmpty(t, statusResp.Status)

	// Phase 3: COMMIT
	commitReq := CommitRequestPayload{
		TransactionID: txID,
	}

	body, _ = json.Marshal(commitReq)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/bos/tx/commit", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code, "commit should succeed")
	var commitResp CommitResponsePayload
	json.Unmarshal(w.Body.Bytes(), &commitResp)

	assert.Equal(t, txID, commitResp.TransactionID)
	assert.Equal(t, "committed", commitResp.Status)

	// Phase 4: CHECK FINAL STATUS
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/bos/tx/status/"+txID, nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &statusResp)
	assert.NotEmpty(t, statusResp.Status)
}

// TestIntegration_AbortAfterPrepare validates abort after prepare phase
func TestIntegration_AbortAfterPrepare(t *testing.T) {
	router, pool := setupBOSTestRouter(t)
	defer pool.Close()

	// Phase 1: PREPARE
	prepareReq := PrepareRequestPayload{
		TransactionID: "tx-abort-after-prepare",
		Algorithm:     "alpha_miner",
		LogData: transactions.LogData{
			Type:     "xes",
			Encoding: "base64",
			Content:  "base64encodedlogcontent",
		},
		Parameters: transactions.AlgorithmParams{
			ActivityKey:  "activity",
			TimestampKey: "timestamp",
			CaseKey:      "case_id",
		},
	}

	body, _ := json.Marshal(prepareReq)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/bos/tx/prepare", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var prepareResp PrepareResponsePayload
	json.Unmarshal(w.Body.Bytes(), &prepareResp)
	txID := prepareResp.TransactionID

	// Phase 2: ABORT (instead of commit)
	abortReq := AbortRequestPayload{
		TransactionID: txID,
		Reason:        "other_participant_failed",
	}

	body, _ = json.Marshal(abortReq)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/bos/tx/abort", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var abortResp AbortResponsePayload
	json.Unmarshal(w.Body.Bytes(), &abortResp)

	assert.Equal(t, txID, abortResp.TransactionID)
	assert.Equal(t, "aborted", abortResp.Status)
}
