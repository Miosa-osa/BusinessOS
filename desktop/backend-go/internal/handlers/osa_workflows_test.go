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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/services"
)

// setupTestDB creates a test database connection
// In production, this would use testcontainers, but for now we'll use a simple in-memory approach
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Use DATABASE_TEST_URL environment variable or skip
	dbURL := "postgres://postgres:postgres@localhost:5432/businessos_test?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
		return nil
	}

	return pool
}

// cleanupWorkflows removes test workflows from database
func cleanupWorkflows(t *testing.T, pool *pgxpool.Pool, workspaceID uuid.UUID) {
	t.Helper()
	_, _ = pool.Exec(context.Background(),
		"DELETE FROM osa_generated_apps WHERE workspace_id = $1", workspaceID)
	_, _ = pool.Exec(context.Background(),
		"DELETE FROM osa_workspaces WHERE id = $1", workspaceID)
}

// createTestWorkspace creates a test workspace
func createTestWorkspace(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) uuid.UUID {
	t.Helper()

	var workspaceID uuid.UUID
	err := pool.QueryRow(context.Background(), `
		INSERT INTO osa_workspaces (user_id, name, workspace_path)
		VALUES ($1, 'test-workspace', '/tmp/test-workspace')
		RETURNING id
	`, userID).Scan(&workspaceID)

	require.NoError(t, err)
	return workspaceID
}

// createTestApp creates a test generated app
func createTestApp(t *testing.T, pool *pgxpool.Pool, workspaceID uuid.UUID, metadata map[string]interface{}) uuid.UUID {
	t.Helper()

	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	var appID uuid.UUID
	err = pool.QueryRow(context.Background(), `
		INSERT INTO osa_generated_apps (
			workspace_id, name, display_name, description,
			osa_workflow_id, status, files_created, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, workspaceID, "test-app", "Test App", "Test Description",
		"wf-"+uuid.New().String()[:8], "generated", 3, metadataJSON).Scan(&appID)

	require.NoError(t, err)
	return appID
}

// setupRouter creates a test router with middleware
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// TestListWorkflows tests workflow listing endpoint
func TestListWorkflows(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userID := uuid.New()
	workspaceID := createTestWorkspace(t, pool, userID)
	defer cleanupWorkflows(t, pool, workspaceID)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	tests := []struct {
		name           string
		setupApps      int
		userID         uuid.UUID
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "List workflows - empty",
			setupApps:      0,
			userID:         userID,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "List workflows - with data",
			setupApps:      3,
			userID:         userID,
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "List workflows - unauthorized",
			setupApps:      2,
			userID:         uuid.Nil,
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test apps
			for i := 0; i < tt.setupApps; i++ {
				metadata := map[string]interface{}{
					"analysis": "Test analysis content",
				}
				createTestApp(t, pool, workspaceID, metadata)
			}

			// Create request
			r := setupRouter()
			r.GET("/api/osa/workflows", handler.ListWorkflows)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/osa/workflows", nil)

			// Setup context with user_id
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			if tt.userID != uuid.Nil {
				c.Set("user_id", tt.userID)
			}

			// Execute
			handler.ListWorkflows(c)

			// Assert status
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				count := int(response["count"].(float64))
				assert.Equal(t, tt.expectedCount, count)

				workflows, ok := response["workflows"].([]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCount, len(workflows))
			}
		})
	}
}

// TestGetWorkflow tests getting a specific workflow
func TestGetWorkflow(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userID := uuid.New()
	workspaceID := createTestWorkspace(t, pool, userID)
	defer cleanupWorkflows(t, pool, workspaceID)

	metadata := map[string]interface{}{
		"analysis":      "Analysis content",
		"architecture":  "Architecture content",
		"code":          "Code content",
	}
	appID := createTestApp(t, pool, workspaceID, metadata)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	tests := []struct {
		name           string
		workflowID     string
		userID         uuid.UUID
		expectedStatus int
		checkFields    bool
	}{
		{
			name:           "Get workflow by UUID - success",
			workflowID:     appID.String(),
			userID:         userID,
			expectedStatus: http.StatusOK,
			checkFields:    true,
		},
		{
			name:           "Get workflow - not found",
			workflowID:     uuid.New().String(),
			userID:         userID,
			expectedStatus: http.StatusNotFound,
			checkFields:    false,
		},
		{
			name:           "Get workflow - unauthorized",
			workflowID:     appID.String(),
			userID:         uuid.New(), // Different user
			expectedStatus: http.StatusNotFound,
			checkFields:    false,
		},
		{
			name:           "Get workflow - invalid UUID",
			workflowID:     "invalid-uuid",
			userID:         userID,
			expectedStatus: http.StatusNotFound,
			checkFields:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupRouter()
			r.GET("/api/osa/workflows/:id", handler.GetWorkflow)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/osa/workflows/"+tt.workflowID, nil)

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.workflowID}}
			c.Set("user_id", tt.userID)

			handler.GetWorkflow(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkFields {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, "test-app", response["name"])
				assert.Equal(t, "Test App", response["display_name"])
				assert.NotNil(t, response["metadata"])
			}
		})
	}
}

// TestGetWorkflowFiles tests file listing for a workflow
func TestGetWorkflowFiles(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userID := uuid.New()
	workspaceID := createTestWorkspace(t, pool, userID)
	defer cleanupWorkflows(t, pool, workspaceID)

	// Test with multi-file bundle
	codeBundle := `=== FILE: src/app.js ===
console.log('Hello');
=== END FILE ===
=== FILE: package.json ===
{"name": "test"}
=== END FILE ===`

	metadata := map[string]interface{}{
		"analysis":     "Analysis content",
		"architecture": "Architecture design",
		"code":         codeBundle,
		"quality":      "Quality report",
	}
	appID := createTestApp(t, pool, workspaceID, metadata)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	tests := []struct {
		name           string
		workflowID     string
		userID         uuid.UUID
		expectedStatus int
		minFileCount   int
	}{
		{
			name:           "Get files - success",
			workflowID:     appID.String(),
			userID:         userID,
			expectedStatus: http.StatusOK,
			minFileCount:   4, // 2 code files + 2 metadata files
		},
		{
			name:           "Get files - not found",
			workflowID:     uuid.New().String(),
			userID:         userID,
			expectedStatus: http.StatusNotFound,
			minFileCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/osa/workflows/"+tt.workflowID+"/files", nil)

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.workflowID}}
			c.Set("user_id", tt.userID)

			handler.GetWorkflowFiles(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				files, ok := response["files"].([]interface{})
				assert.True(t, ok)
				assert.GreaterOrEqual(t, len(files), tt.minFileCount)

				// Validate file structure
				if len(files) > 0 {
					file := files[0].(map[string]interface{})
					assert.NotEmpty(t, file["id"])
					assert.NotEmpty(t, file["name"])
					assert.NotEmpty(t, file["type"])
				}
			}
		})
	}
}

// TestGetFileContent tests getting file content by type
func TestGetFileContent(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userID := uuid.New()
	workspaceID := createTestWorkspace(t, pool, userID)
	defer cleanupWorkflows(t, pool, workspaceID)

	metadata := map[string]interface{}{
		"analysis":      "# Analysis Report\nThis is the analysis.",
		"architecture":  "# System Architecture",
		"code":          "console.log('test');",
	}
	appID := createTestApp(t, pool, workspaceID, metadata)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	tests := []struct {
		name           string
		workflowID     string
		fileType       string
		userID         uuid.UUID
		expectedStatus int
		expectedSize   int
	}{
		{
			name:           "Get analysis file",
			workflowID:     appID.String(),
			fileType:       "analysis",
			userID:         userID,
			expectedStatus: http.StatusOK,
			expectedSize:   42,
		},
		{
			name:           "Get code file",
			workflowID:     appID.String(),
			fileType:       "code",
			userID:         userID,
			expectedStatus: http.StatusOK,
			expectedSize:   22,
		},
		{
			name:           "Get nonexistent file type",
			workflowID:     appID.String(),
			fileType:       "deployment",
			userID:         userID,
			expectedStatus: http.StatusNotFound,
			expectedSize:   0,
		},
		{
			name:           "Invalid file type",
			workflowID:     appID.String(),
			fileType:       "invalid_type",
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
			expectedSize:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			url := "/api/osa/workflows/" + tt.workflowID + "/files/" + tt.fileType
			req := httptest.NewRequest("GET", url, nil)

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "id", Value: tt.workflowID},
				{Key: "type", Value: tt.fileType},
			}
			c.Set("user_id", tt.userID)

			handler.GetFileContent(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.fileType, response["type"])
				assert.Equal(t, float64(tt.expectedSize), response["size"])
				assert.NotEmpty(t, response["content"])
			}
		})
	}
}

// TestGetFileContentByID tests getting file content by deterministic UUID
func TestGetFileContentByID(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userID := uuid.New()
	workspaceID := createTestWorkspace(t, pool, userID)
	defer cleanupWorkflows(t, pool, workspaceID)

	metadata := map[string]interface{}{
		"analysis": "Analysis content here",
	}
	appID := createTestApp(t, pool, workspaceID, metadata)

	// Generate deterministic file ID (same logic as handler)
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	fileID := uuid.NewSHA1(namespace, []byte(appID.String()+":analysis"))

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	tests := []struct {
		name           string
		fileID         string
		userID         uuid.UUID
		expectedStatus int
	}{
		{
			name:           "Get file by valid ID",
			fileID:         fileID.String(),
			userID:         userID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get file by invalid ID",
			fileID:         "invalid-uuid",
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get file by nonexistent ID",
			fileID:         uuid.New().String(),
			userID:         userID,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/osa/files/"+tt.fileID+"/content", nil)

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.fileID}}
			c.Set("user_id", tt.userID)

			handler.GetFileContentByID(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response["content"])
				assert.NotNil(t, response["file"])
			}
		})
	}
}

// TestInstallModule tests module installation
func TestInstallModule(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userID := uuid.New()
	workspaceID := createTestWorkspace(t, pool, userID)
	defer cleanupWorkflows(t, pool, workspaceID)

	metadata := map[string]interface{}{
		"architecture":    "System design",
		"code":            "Application code",
		"recommendations": "UI recommendations",
	}
	appID := createTestApp(t, pool, workspaceID, metadata)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		userID         uuid.UUID
		expectedStatus int
	}{
		{
			name: "Install module - success",
			requestBody: map[string]interface{}{
				"workflow_id": appID.String(),
				"module_name": "my-custom-module",
			},
			userID:         userID,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Install module - workflow not found",
			requestBody: map[string]interface{}{
				"workflow_id": uuid.New().String(),
				"module_name": "missing-module",
			},
			userID:         userID,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Install module - invalid request",
			requestBody:    map[string]interface{}{},
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/osa/modules/install", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", tt.userID)

			handler.InstallModule(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response["success"].(bool))
				assert.NotEmpty(t, response["module_id"])
			}
		})
	}
}

// TestTriggerSync tests manual sync triggering
func TestTriggerSync(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	userID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/osa/sync/trigger", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.TriggerSync(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Sync triggered", response["message"])
	assert.Equal(t, userID.String(), response["user_id"])
}

// TestConcurrentWorkflowAccess tests concurrent access to workflows
func TestConcurrentWorkflowAccess(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userID := uuid.New()
	workspaceID := createTestWorkspace(t, pool, userID)
	defer cleanupWorkflows(t, pool, workspaceID)

	metadata := map[string]interface{}{
		"analysis": "Test content",
	}
	appID := createTestApp(t, pool, workspaceID, metadata)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	// Simulate 10 concurrent requests
	concurrency := 10
	results := make(chan int, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/osa/workflows/"+appID.String(), nil)

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: appID.String()}}
			c.Set("user_id", userID)

			handler.GetWorkflow(c)
			results <- w.Code
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < concurrency; i++ {
		statusCode := <-results
		if statusCode == http.StatusOK {
			successCount++
		}
	}

	// All requests should succeed
	assert.Equal(t, concurrency, successCount)
}

// BenchmarkListWorkflows benchmarks workflow listing performance
func BenchmarkListWorkflows(b *testing.B) {
	pool, err := pgxpool.New(context.Background(),
		"postgres://postgres:postgres@localhost:5432/businessos_test?sslmode=disable")
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
		return
	}
	defer pool.Close()

	userID := uuid.New()
	ctx := context.Background()

	var workspaceID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO osa_workspaces (user_id, name, workspace_path)
		VALUES ($1, 'bench-workspace', '/tmp/bench')
		RETURNING id
	`, userID).Scan(&workspaceID)
	if err != nil {
		b.Skipf("Setup failed: %v", err)
		return
	}
	defer pool.Exec(ctx, "DELETE FROM osa_workspaces WHERE id = $1", workspaceID)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	syncService := services.NewOSAFileSyncService(pool, logger, "")
	handler := NewOSAWorkflowsHandler(pool, syncService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/osa/workflows", nil)

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.ListWorkflows(c)
	}
}
