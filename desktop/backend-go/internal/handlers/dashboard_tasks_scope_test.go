package handlers

import (
	"context"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestTaskWorkspaceMutation(t *testing.T) {
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
	// Connection-local tables ensure no real workspace records are mutated.
	_, err = pool.Exec(ctx, `CREATE TEMP TABLE tasks AS SELECT * FROM public.tasks WHERE false;
	CREATE TEMP TABLE workspace_members (workspace_id uuid, user_id text, status text, role text);
	INSERT INTO workspace_members VALUES
	('11111111-1111-4111-8111-111111111111','editor','active','member'),
	('11111111-1111-4111-8111-111111111111','viewer','active','viewer'),
	('11111111-1111-4111-8111-111111111111','inactive','inactive','member');
	INSERT INTO tasks(id,user_id,title,status,priority,workspace_id) VALUES
	('22222222-2222-4222-8222-222222222222','creator','Shared task','todo','high','11111111-1111-4111-8111-111111111111');`)
	require.NoError(t, err)
	h := &DashboardItemHandler{pool: pool}
	for _, tc := range []struct {
		name, user, workspace string
		want                  int
	}{
		{"active editor", "editor", "11111111-1111-4111-8111-111111111111", 200},
		{"viewer", "viewer", "11111111-1111-4111-8111-111111111111", 404},
		{"inactive", "inactive", "11111111-1111-4111-8111-111111111111", 404},
		{"outsider", "outsider", "11111111-1111-4111-8111-111111111111", 404},
		{"wrong workspace", "editor", "33333333-3333-4333-8333-333333333333", 404},
		{"invalid workspace", "editor", "invalid", 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := setupDashboardRouter(tc.user)
			r.PUT("/tasks/:id", h.UpdateTask)
			req := httptest.NewRequest("PUT", "/tasks/22222222-2222-4222-8222-222222222222", strings.NewReader(`{"status":"in_progress"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Workspace-ID", tc.workspace)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, tc.want, w.Code, w.Body.String())
		})
	}
}
