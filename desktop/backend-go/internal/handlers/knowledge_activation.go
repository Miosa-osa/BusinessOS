package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// resolveWorkspaceIDBySlug looks up a BusinessOS workspace UUID from its slug in
// the workspaces table. This works even before any knowledge has been synced
// (knowledge_documents is empty pre-sync), so activation state can be keyed on a
// real workspace_id from the first activation onward. Returns ("", false) when
// the pool is nil or the slug is unknown.
func resolveWorkspaceIDBySlug(ctx context.Context, pool *pgxpool.Pool, slug string) (string, bool) {
	if pool == nil || strings.TrimSpace(slug) == "" {
		return "", false
	}
	var wsID string
	if err := pool.QueryRow(ctx, `SELECT id::text FROM workspaces WHERE slug = $1`, slug).Scan(&wsID); err != nil {
		return "", false
	}
	return wsID, true
}

// userIsWorkspaceMember reports whether the user is an active member of the
// workspace (by UUID) in the given pool. This is the multi-tenant gate for
// CLOUD knowledge paths (the shared knowledge_documents / workspace_storage
// copy). Local single-tenant file paths are intentionally NOT gated.
func userIsWorkspaceMember(ctx context.Context, pool *pgxpool.Pool, wsID, userID string) bool {
	return userIsWorkspaceMemberByIdentity(ctx, pool, wsID, userID, "")
}

// userIsWorkspaceMemberByIdentity also matches the authenticated email. Local
// and cloud databases can assign different IDs to the same account, while the
// verified email remains the stable identity across both projections.
func userIsWorkspaceMemberByIdentity(ctx context.Context, pool *pgxpool.Pool, wsID, userID, email string) bool {
	if pool == nil || strings.TrimSpace(wsID) == "" || strings.TrimSpace(userID) == "" {
		return false
	}
	var ok bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM workspace_members wm
			JOIN "user" u ON u.id = wm.user_id
			WHERE wm.workspace_id = $1
			  AND wm.status = 'active'
			  AND (wm.user_id = $2 OR ($3 <> '' AND lower(u.email) = lower($3)))
		)`,
		wsID, userID, strings.TrimSpace(email)).Scan(&ok); err != nil {
		return false
	}
	return ok
}

// callerMemberOfSlug resolves a workspace slug to its id in the given pool and
// reports whether the authenticated caller is an active member. Fail-closed:
// unknown user, unresolvable slug, or non-membership all return false. Used to
// gate the CLOUD knowledge paths; the local file paths do not call this.
func (h *KnowledgeHandler) callerMemberOfSlug(c *gin.Context, pool *pgxpool.Pool, slug string) bool {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		return false
	}
	wsID, ok := resolveWorkspaceIDBySlug(c.Request.Context(), pool, slug)
	if !ok {
		return false
	}
	return userIsWorkspaceMemberByIdentity(c.Request.Context(), pool, wsID, user.ID, user.Email)
}

// cloudPool returns the pool that backs the cloud copy for storage/activation
// reads and writes: the cloud DB when this machine owns the sync
// (CLOUD_DATABASE_URL set), otherwise this backend's own pool (i.e. this IS the
// cloud backend). The bool reports whether the returned pool must be closed by
// the caller (true only for a freshly-opened cloud pool).
func (h *KnowledgeHandler) cloudPool(ctx context.Context) (*pgxpool.Pool, bool) {
	if cloudURL := services.CloudDatabaseURL(); cloudURL != "" {
		cfg, err := pgxpool.ParseConfig(cloudURL)
		if err != nil {
			return nil, false
		}
		cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			return nil, false
		}
		return pool, true
	}
	return h.pool, false
}

// ActivateCloudSync flips the opt-in cloud-sync gate ON for a workspace. Until
// this is called, SyncToCloud refuses to write (nothing leaves the machine).
// Local knowledge is always on + free; this is the explicit opt-in to the
// cloud layer (the paid-later tier; billing is soft for now).
//
//	POST /api/knowledge/cloud/activate?workspace=<slug>
//	-> { activated: true, workspace: <slug> }
func (h *KnowledgeHandler) ActivateCloudSync(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		RespondUnauthorizedErr(c, "authentication required")
		return
	}
	slug := c.Query("workspace")
	if strings.TrimSpace(slug) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace required"})
		return
	}
	ctx := c.Request.Context()

	pool, mustClose := h.cloudPool(ctx)
	if pool == nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "cloud database not available"})
		return
	}
	if mustClose {
		defer pool.Close()
	}

	wsID, ok := resolveWorkspaceIDBySlug(ctx, pool, slug)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown workspace: " + slug})
		return
	}
	if !userIsWorkspaceMemberByIdentity(ctx, pool, wsID, user.ID, user.Email) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this workspace"})
		return
	}

	// Ensure the storage table + activation column exist on the cloud DB (fresh
	// cloud DBs may never have run migrations).
	if _, err := pool.Exec(ctx, workspaceStorageDDL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ensure storage table: " + err.Error()})
		return
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO workspace_storage (workspace_id, cloud_sync_activated, updated_at)
		 VALUES ($1, true, now())
		 ON CONFLICT (workspace_id) DO UPDATE SET cloud_sync_activated = true, updated_at = now()`,
		wsID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "activate: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activated": true, "workspace": slug})
}

// isCloudSyncActivated reports whether cloud sync is opted-in for a workspace.
// Reads cloud_sync_activated from workspace_storage keyed by the workspace UUID.
// Defaults to false (not activated) when there is no row, the column/table is
// missing, or the workspace id cannot be resolved.
func isCloudSyncActivated(ctx context.Context, pool *pgxpool.Pool, wsID string) bool {
	if pool == nil || strings.TrimSpace(wsID) == "" {
		return false
	}
	var activated bool
	if err := pool.QueryRow(ctx,
		`SELECT COALESCE(cloud_sync_activated, false) FROM workspace_storage WHERE workspace_id = $1`, wsID).Scan(&activated); err != nil {
		return false
	}
	return activated
}
