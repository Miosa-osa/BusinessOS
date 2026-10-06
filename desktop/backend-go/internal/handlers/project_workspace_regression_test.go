package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/stretchr/testify/require"
)

func TestProjectMetadataIsJSONObject(t *testing.T) {
	rows := TransformProjectRows([]sqlc.ListProjectsRow{{ProjectMetadata: []byte(`{"knowledge_path":"packages/guide.md"}`)}})
	encoded, err := json.Marshal(rows)
	require.NoError(t, err)
	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(encoded, &result))
	metadata, ok := result[0]["project_metadata"].(map[string]interface{})
	require.True(t, ok, "project metadata must not be base64 encoded")
	require.Equal(t, "packages/guide.md", metadata["knowledge_path"])
}

func TestCreatedTaskStaysInWorkspace(t *testing.T) {
	url := os.Getenv("TEST_TASK_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_TASK_DATABASE_URL required")
	}
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(url)
	require.NoError(t, err)
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()
	_, err = pool.Exec(ctx, `CREATE TEMP TABLE tasks (LIKE public.tasks INCLUDING DEFAULTS);
	CREATE TEMP TABLE workspace_members (workspace_id uuid,user_id text,status text);
	CREATE TEMP TABLE sync_policies (workspace_id uuid,module text,sync_mode text);
	INSERT INTO workspace_members VALUES ('11111111-1111-4111-8111-111111111111','editor','active');
	INSERT INTO sync_policies VALUES ('11111111-1111-4111-8111-111111111111','tasks','workspace');`)
	require.NoError(t, err)
	h := &DashboardItemHandler{pool: pool}
	r := setupDashboardRouter("editor")
	r.POST("/tasks", h.CreateTask)
	req := httptest.NewRequest("POST", "/tasks", strings.NewReader(`{"title":"Scoped task","status":"todo","priority":"medium"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workspace-ID", "11111111-1111-4111-8111-111111111111")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, 201, w.Code, w.Body.String())
	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM tasks WHERE workspace_id='11111111-1111-4111-8111-111111111111' AND title='Scoped task'`).Scan(&count))
	require.Equal(t, 1, count)
}
