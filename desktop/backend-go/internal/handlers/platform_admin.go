package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// PlatformAdminHandler serves platform-level admin endpoints.
// All routes require platform_role = 'superadmin'.
type PlatformAdminHandler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

// NewPlatformAdminHandler constructs a PlatformAdminHandler.
func NewPlatformAdminHandler(pool *pgxpool.Pool, cfg *config.Config) *PlatformAdminHandler {
	return &PlatformAdminHandler{pool: pool, cfg: cfg}
}

// RegisterPlatformAdminRoutes mounts platform admin endpoints under /api/v1/admin.
// auth is the shared authentication middleware (sets user in context).
// Every route applies requireSuperAdmin via the group middleware — no per-handler
// boilerplate needed.
func RegisterPlatformAdminRoutes(api *gin.RouterGroup, h *PlatformAdminHandler, auth gin.HandlerFunc) {
	admin := api.Group("/v1/admin")
	admin.Use(auth, middleware.RequireAuth(), h.superAdminMiddleware())
	{
		admin.GET("/dashboard", h.GetDashboard)
		admin.GET("/users", h.ListUsers)
		admin.GET("/computers", h.ListComputers)
		admin.POST("/users/:id/role", h.SetUserRole)
		admin.POST("/users/:id/plan", h.SetUserPlan)
		admin.GET("/audit-log", h.ListAuditLog)
		admin.GET("/workspaces", h.ListWorkspaces)
		admin.GET("/workspaces/:id", h.GetWorkspace)
		admin.POST("/computers/:id/restart", h.RestartComputer)
		admin.POST("/computers/:id/stop", h.StopComputer)
		admin.POST("/computers/:id/hibernate", h.HibernateComputer)
		admin.GET("/settings", h.GetSettings)
	}
}

// ── superadmin enforcement ────────────────────────────────────────────────────

// superAdminMiddleware returns a gin middleware that aborts with 403 when the
// authenticated user does not hold platform_role = 'superadmin'.
// It fetches the role from the DB on every request so revocations take effect
// immediately (no stale session state).
func (h *PlatformAdminHandler) superAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil {
			utils.RespondUnauthorized(c, slog.Default())
			c.Abort()
			return
		}

		if !h.isSuperAdmin(c, user.ID) {
			utils.RespondForbidden(c, slog.Default(), "superadmin role required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// isSuperAdmin queries the DB for the user's platform_role.
// Returns false on any DB error — fail-closed is the safe default.
func (h *PlatformAdminHandler) isSuperAdmin(c *gin.Context, userID string) bool {
	var role string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT platform_role FROM "user" WHERE id = $1`, userID,
	).Scan(&role)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "platform_admin: failed to fetch platform_role",
			"user_id", userID, "error", err)
		return false
	}
	return role == "superadmin"
}

// ── response types ────────────────────────────────────────────────────────────

type adminUserRow struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	PlatformRole  string    `json:"platform_role"`
	CreatedAt     time.Time `json:"created_at"`
	WorkspaceID   *string   `json:"workspace_id"`
	WorkspaceName *string   `json:"workspace_name"`
	Plan          *string   `json:"plan"`
}

type adminComputerRow struct {
	ID              string     `json:"id"`
	WorkspaceID     string     `json:"workspace_id"`
	WorkspaceName   *string    `json:"workspace_name"`
	OwnerUserID     string     `json:"owner_user_id"`
	OwnerEmail      string     `json:"owner_email"`
	MiosaComputerID *string    `json:"miosa_computer_id"`
	Plan            string     `json:"plan"`
	Status          string     `json:"status"`
	RAMMB           int        `json:"ram_mb"`
	VCPUs           int        `json:"vcpus"`
	StorageGB       int        `json:"storage_gb"`
	CreatedAt       time.Time  `json:"created_at"`
	LastActiveAt    *time.Time `json:"last_active_at"`
}

type setRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type setPlanRequest struct {
	Plan string `json:"plan" binding:"required"`
}

type adminWorkspaceRow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	PlanType    *string   `json:"plan_type"`
	CreatedAt   time.Time `json:"created_at"`
	OwnerEmail  *string   `json:"owner_email"`
	OwnerName   *string   `json:"owner_name"`
	MemberCount int64     `json:"member_count"`
}

type adminWorkspaceMemberRow struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     *string   `json:"role"`
	Status   *string   `json:"status"`
	JoinedAt time.Time `json:"joined_at"`
}

type adminWorkspaceDetail struct {
	adminWorkspaceRow
	Members []adminWorkspaceMemberRow `json:"members"`
}

type auditLogEntry struct {
	ID         string            `json:"id"`
	ActorID    string            `json:"actor_id"`
	ActorRole  string            `json:"actor_role"`
	Action     string            `json:"action"`
	TargetType *string           `json:"target_type"`
	TargetID   *string           `json:"target_id"`
	Details    map[string]string `json:"details"`
	IPAddress  *string           `json:"ip_address"`
	CreatedAt  time.Time         `json:"created_at"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// GetDashboard handles GET /api/v1/admin/dashboard.
// Queries the local DB for authoritative counts; returns mock data for revenue
// and system utilization until Stripe and metrics pipelines are wired.
func (h *PlatformAdminHandler) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	// ── DB queries ──

	var totalUsers, newUsersToday int64
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&totalUsers); err != nil {
		slog.WarnContext(ctx, "platform_admin: dashboard user count failed", "error", err)
	}
	if err := h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM "user" WHERE "createdAt" >= NOW() - INTERVAL '24 hours'`,
	).Scan(&newUsersToday); err != nil {
		slog.WarnContext(ctx, "platform_admin: dashboard new users today failed", "error", err)
	}

	var totalWorkspaces int64
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM workspaces`).Scan(&totalWorkspaces); err != nil {
		slog.WarnContext(ctx, "platform_admin: dashboard workspace count failed", "error", err)
	}

	var totalComputers, runningComputers int64
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM platform_computers`).Scan(&totalComputers); err != nil {
		slog.WarnContext(ctx, "platform_admin: dashboard computer count failed", "error", err)
	}
	if err := h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM platform_computers WHERE status IN ('active', 'running')`,
	).Scan(&runningComputers); err != nil {
		slog.WarnContext(ctx, "platform_admin: dashboard running computers failed", "error", err)
	}

	var activeSubs int64
	if err := h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM platform_subscriptions WHERE status = 'active' AND plan != 'free'`,
	).Scan(&activeSubs); err != nil {
		slog.WarnContext(ctx, "platform_admin: dashboard active subscriptions failed", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": gin.H{
			"total":     totalUsers,
			"online":    0, // requires WebSocket presence tracking — not yet wired
			"new_today": newUsersToday,
		},
		"workspaces": gin.H{
			"total":  totalWorkspaces,
			"active": totalWorkspaces, // all workspaces are active until soft-delete is added
		},
		"computers": gin.H{
			"total":      totalComputers,
			"running":    runningComputers,
			"hibernated": totalComputers - runningComputers,
		},
		"revenue": gin.H{
			// TODO: wire Stripe webhook events into platform_subscriptions and sum mrr_cents
			"mrr_cents":            activeSubs * 4000, // naive estimate: assume all active = pro plan
			"active_subscriptions": activeSubs,
		},
		"system": gin.H{
			// TODO: wire MIOSA health check and compute utilization metrics
			"miosa_status":        "ok",
			"compute_utilization": 0.0,
		},
	})
}

// ListUsers handles GET /api/v1/admin/users.
// Returns a paginated user list joined with their workspace and subscription.
// Query params: page (1-based), page_size (max 100).
func (h *PlatformAdminHandler) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()
	pg := ParsePagination(c)

	const q = `
		SELECT
			u.id,
			u.name,
			u.email,
			COALESCE(u.platform_role, 'user') AS platform_role,
			u."createdAt",
			w.id         AS workspace_id,
			w.name       AS workspace_name,
			ps.plan      AS plan
		FROM "user" u
		LEFT JOIN workspaces w
			ON w.id = (
				SELECT workspace_id
				FROM   team_members
				WHERE  user_id = u.id
				ORDER  BY created_at
				LIMIT  1
			)
		LEFT JOIN platform_subscriptions ps ON ps.workspace_id = w.id
		ORDER BY u."createdAt" DESC
		LIMIT  $1
		OFFSET $2
	`

	rows, err := h.pool.Query(ctx, q, pg.Limit, pg.Offset)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list platform users", err)
		return
	}
	defer rows.Close()

	users := make([]adminUserRow, 0, pg.Limit)
	for rows.Next() {
		var row adminUserRow
		if err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.Email,
			&row.PlatformRole,
			&row.CreatedAt,
			&row.WorkspaceID,
			&row.WorkspaceName,
			&row.Plan,
		); err != nil {
			utils.RespondInternalError(c, slog.Default(), "scan platform users", err)
			return
		}
		users = append(users, row)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "iterate platform users", err)
		return
	}

	var total int64
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&total); err != nil {
		slog.WarnContext(ctx, "platform_admin: list users count failed", "error", err)
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data: users,
		Pagination: PaginationMeta{
			Page:       pg.Page,
			PageSize:   pg.PageSize,
			TotalItems: total,
			TotalPages: int((total + int64(pg.PageSize) - 1) / int64(pg.PageSize)),
			HasMore:    int64(pg.Offset)+int64(pg.Limit) < total,
		},
	})
}

// ListComputers handles GET /api/v1/admin/computers.
// Returns all platform_computers rows joined with workspace name and owner email.
// Query params: page (1-based), page_size (max 100).
func (h *PlatformAdminHandler) ListComputers(c *gin.Context) {
	ctx := c.Request.Context()
	pg := ParsePagination(c)

	const q = `
		SELECT
			pc.id,
			pc.workspace_id,
			w.name                AS workspace_name,
			pc.owner_user_id,
			u.email               AS owner_email,
			pc.miosa_computer_id,
			pc.plan,
			pc.status,
			pc.ram_mb,
			pc.vcpus,
			pc.storage_gb,
			pc.created_at,
			pc.last_active_at
		FROM platform_computers pc
		JOIN workspaces w ON w.id = pc.workspace_id
		JOIN "user"   u ON u.id = pc.owner_user_id
		ORDER BY pc.created_at DESC
		LIMIT  $1
		OFFSET $2
	`

	rows, err := h.pool.Query(ctx, q, pg.Limit, pg.Offset)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list platform computers", err)
		return
	}
	defer rows.Close()

	computers := make([]adminComputerRow, 0, pg.Limit)
	for rows.Next() {
		var row adminComputerRow
		if err := rows.Scan(
			&row.ID,
			&row.WorkspaceID,
			&row.WorkspaceName,
			&row.OwnerUserID,
			&row.OwnerEmail,
			&row.MiosaComputerID,
			&row.Plan,
			&row.Status,
			&row.RAMMB,
			&row.VCPUs,
			&row.StorageGB,
			&row.CreatedAt,
			&row.LastActiveAt,
		); err != nil {
			utils.RespondInternalError(c, slog.Default(), "scan platform computers", err)
			return
		}
		computers = append(computers, row)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "iterate platform computers", err)
		return
	}

	var total int64
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM platform_computers`).Scan(&total); err != nil {
		slog.WarnContext(ctx, "platform_admin: list computers count failed", "error", err)
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data: computers,
		Pagination: PaginationMeta{
			Page:       pg.Page,
			PageSize:   pg.PageSize,
			TotalItems: total,
			TotalPages: int((total + int64(pg.PageSize) - 1) / int64(pg.PageSize)),
			HasMore:    int64(pg.Offset)+int64(pg.Limit) < total,
		},
	})
}

// ListAuditLog handles GET /api/v1/admin/audit-log.
// Returns the 50 most recent platform audit log entries ordered newest-first.
// Response: { "entries": [...], "total": N }
func (h *PlatformAdminHandler) ListAuditLog(c *gin.Context) {
	ctx := c.Request.Context()

	const q = `
		SELECT
			id,
			actor_id,
			actor_role,
			action,
			target_type,
			target_id,
			details,
			ip_address,
			created_at
		FROM platform_audit_log
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := h.pool.Query(ctx, q)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list audit log", err)
		return
	}
	defer rows.Close()

	entries := make([]auditLogEntry, 0, 50)
	for rows.Next() {
		var e auditLogEntry
		var details []byte
		if err := rows.Scan(
			&e.ID,
			&e.ActorID,
			&e.ActorRole,
			&e.Action,
			&e.TargetType,
			&e.TargetID,
			&details,
			&e.IPAddress,
			&e.CreatedAt,
		); err != nil {
			utils.RespondInternalError(c, slog.Default(), "scan audit log", err)
			return
		}
		if len(details) > 0 {
			if err := json.Unmarshal(details, &e.Details); err != nil {
				slog.WarnContext(ctx, "platform_admin: audit log details unmarshal failed",
					"entry_id", e.ID, "error", err)
			}
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "iterate audit log", err)
		return
	}

	var total int64
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM platform_audit_log`).Scan(&total); err != nil {
		slog.WarnContext(ctx, "platform_admin: audit log count failed", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"total":   total,
	})
}

// SetUserRole handles POST /api/v1/admin/users/:id/role.
// Accepts {"role": "admin"} and updates the target user's platform_role.
// Only superadmin may call this; the middleware layer already enforces that.
func (h *PlatformAdminHandler) SetUserRole(c *gin.Context) {
	ctx := c.Request.Context()

	actor := middleware.GetCurrentUser(c)
	if actor == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	targetID := c.Param("id")
	if targetID == "" {
		utils.RespondInvalidID(c, slog.Default(), "user ID")
		return
	}

	var req setRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	// Validate role value — must match the DB check constraint.
	switch req.Role {
	case "superadmin", "admin", "user":
		// valid
	default:
		utils.RespondBadRequest(c, slog.Default(), "role must be one of: superadmin, admin, user")
		return
	}

	// Update target user's role.
	tag, err := h.pool.Exec(ctx,
		`UPDATE "user" SET platform_role = $1 WHERE id = $2`,
		req.Role, targetID,
	)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "set platform role", err)
		return
	}
	if tag.RowsAffected() == 0 {
		utils.RespondNotFound(c, slog.Default(), "user")
		return
	}

	// Append to platform audit log — best-effort (non-fatal on failure).
	_, auditErr := h.pool.Exec(ctx,
		`INSERT INTO platform_audit_log
		 (actor_id, actor_role, action, target_type, target_id, details, ip_address)
		 VALUES ($1, 'superadmin', 'set_platform_role', 'user', $2, $3, $4)`,
		actor.ID,
		targetID,
		map[string]string{"new_role": req.Role},
		c.ClientIP(),
	)
	if auditErr != nil {
		slog.WarnContext(ctx, "platform_admin: audit log insert failed",
			"actor_id", actor.ID, "target_id", targetID, "error", auditErr)
	}

	slog.InfoContext(ctx, "platform_admin: role updated",
		"actor_id", actor.ID, "target_id", targetID, "new_role", req.Role)

	c.JSON(http.StatusOK, gin.H{
		"user_id": targetID,
		"role":    req.Role,
		"message": "Platform role updated.",
	})
}

// SetUserPlan handles POST /api/v1/admin/users/:id/plan.
// Accepts {"plan": "pro"} and updates plan_type on the user's primary workspace.
func (h *PlatformAdminHandler) SetUserPlan(c *gin.Context) {
	ctx := c.Request.Context()

	actor := middleware.GetCurrentUser(c)
	if actor == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		utils.RespondInvalidID(c, slog.Default(), "user ID")
		return
	}

	var req setPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	// Validate plan value.
	switch req.Plan {
	case "free", "starter", "pro", "enterprise":
		// valid
	default:
		utils.RespondBadRequest(c, slog.Default(), "plan must be one of: free, starter, pro, enterprise")
		return
	}

	// Resolve the user's primary workspace — earliest membership.
	var workspaceID string
	err := h.pool.QueryRow(ctx,
		`SELECT workspace_id FROM workspace_members WHERE user_id = $1 ORDER BY created_at LIMIT 1`,
		targetUserID,
	).Scan(&workspaceID)
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "workspace for user")
		return
	}

	tag, err := h.pool.Exec(ctx,
		`UPDATE workspaces SET plan_type = $1 WHERE id = $2`,
		req.Plan, workspaceID,
	)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "set workspace plan", err)
		return
	}
	if tag.RowsAffected() == 0 {
		utils.RespondNotFound(c, slog.Default(), "workspace")
		return
	}

	// Audit log — best-effort.
	_, auditErr := h.pool.Exec(ctx,
		`INSERT INTO platform_audit_log
		 (actor_id, actor_role, action, target_type, target_id, details, ip_address)
		 VALUES ($1, 'superadmin', 'set_workspace_plan', 'workspace', $2, $3, $4)`,
		actor.ID,
		workspaceID,
		map[string]string{"new_plan": req.Plan, "user_id": targetUserID},
		c.ClientIP(),
	)
	if auditErr != nil {
		slog.WarnContext(ctx, "platform_admin: audit log insert failed",
			"actor_id", actor.ID, "workspace_id", workspaceID, "error", auditErr)
	}

	slog.InfoContext(ctx, "platform_admin: plan updated",
		"actor_id", actor.ID, "user_id", targetUserID, "workspace_id", workspaceID, "new_plan", req.Plan)

	c.JSON(http.StatusOK, gin.H{
		"user_id":      targetUserID,
		"workspace_id": workspaceID,
		"plan":         req.Plan,
		"message":      "Workspace plan updated.",
	})
}

// ListWorkspaces handles GET /api/v1/admin/workspaces.
// Returns all workspaces with owner info and member count, paginated.
func (h *PlatformAdminHandler) ListWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	pg := ParsePagination(c)

	const q = `
		SELECT
			w.id,
			w.name,
			w.slug,
			w.plan_type,
			w.created_at,
			u.email  AS owner_email,
			u.name   AS owner_name,
			(SELECT COUNT(*) FROM workspace_members wm WHERE wm.workspace_id = w.id) AS member_count
		FROM workspaces w
		LEFT JOIN "user" u ON u.id = w.owner_id
		ORDER BY w.created_at DESC
		LIMIT  $1
		OFFSET $2
	`

	rows, err := h.pool.Query(ctx, q, pg.Limit, pg.Offset)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list platform workspaces", err)
		return
	}
	defer rows.Close()

	workspaces := make([]adminWorkspaceRow, 0, pg.Limit)
	for rows.Next() {
		var row adminWorkspaceRow
		if err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.Slug,
			&row.PlanType,
			&row.CreatedAt,
			&row.OwnerEmail,
			&row.OwnerName,
			&row.MemberCount,
		); err != nil {
			utils.RespondInternalError(c, slog.Default(), "scan platform workspaces", err)
			return
		}
		workspaces = append(workspaces, row)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "iterate platform workspaces", err)
		return
	}

	var total int64
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM workspaces`).Scan(&total); err != nil {
		slog.WarnContext(ctx, "platform_admin: list workspaces count failed", "error", err)
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data: workspaces,
		Pagination: PaginationMeta{
			Page:       pg.Page,
			PageSize:   pg.PageSize,
			TotalItems: total,
			TotalPages: int((total + int64(pg.PageSize) - 1) / int64(pg.PageSize)),
			HasMore:    int64(pg.Offset)+int64(pg.Limit) < total,
		},
	})
}

// GetWorkspace handles GET /api/v1/admin/workspaces/:id.
// Returns detailed workspace info including the full member list.
func (h *PlatformAdminHandler) GetWorkspace(c *gin.Context) {
	ctx := c.Request.Context()

	workspaceID := c.Param("id")
	if workspaceID == "" {
		utils.RespondInvalidID(c, slog.Default(), "workspace ID")
		return
	}

	// Fetch core workspace row.
	var ws adminWorkspaceRow
	err := h.pool.QueryRow(ctx, `
		SELECT
			w.id,
			w.name,
			w.slug,
			w.plan_type,
			w.created_at,
			u.email  AS owner_email,
			u.name   AS owner_name,
			(SELECT COUNT(*) FROM workspace_members wm WHERE wm.workspace_id = w.id) AS member_count
		FROM workspaces w
		LEFT JOIN "user" u ON u.id = w.owner_id
		WHERE w.id = $1
	`, workspaceID).Scan(
		&ws.ID,
		&ws.Name,
		&ws.Slug,
		&ws.PlanType,
		&ws.CreatedAt,
		&ws.OwnerEmail,
		&ws.OwnerName,
		&ws.MemberCount,
	)
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "workspace")
		return
	}

	// Fetch members.
	memberRows, err := h.pool.Query(ctx, `
		SELECT
			u.id    AS user_id,
			u.email,
			u.name,
			wm.role,
			wm.status,
			wm.created_at AS joined_at
		FROM workspace_members wm
		JOIN "user" u ON u.id = wm.user_id
		WHERE wm.workspace_id = $1
		ORDER BY wm.created_at
	`, workspaceID)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list workspace members", err)
		return
	}
	defer memberRows.Close()

	members := make([]adminWorkspaceMemberRow, 0)
	for memberRows.Next() {
		var m adminWorkspaceMemberRow
		if err := memberRows.Scan(
			&m.UserID,
			&m.Email,
			&m.Name,
			&m.Role,
			&m.Status,
			&m.JoinedAt,
		); err != nil {
			utils.RespondInternalError(c, slog.Default(), "scan workspace members", err)
			return
		}
		members = append(members, m)
	}
	if err := memberRows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "iterate workspace members", err)
		return
	}

	c.JSON(http.StatusOK, adminWorkspaceDetail{
		adminWorkspaceRow: ws,
		Members:           members,
	})
}

// computeAction is the shared implementation for restart/stop/hibernate.
// It logs the action and writes an audit entry; the actual MIOSA API call
// will be wired in a follow-up once the MIOSA compute client is available.
func (h *PlatformAdminHandler) computeAction(c *gin.Context, action string) {
	ctx := c.Request.Context()

	actor := middleware.GetCurrentUser(c)
	if actor == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	computerID := c.Param("id")
	if computerID == "" {
		utils.RespondInvalidID(c, slog.Default(), "computer ID")
		return
	}

	// Verify the computer exists before acting.
	var exists bool
	if err := h.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM platform_computers WHERE id = $1)`, computerID,
	).Scan(&exists); err != nil {
		utils.RespondInternalError(c, slog.Default(), "check computer exists", err)
		return
	}
	if !exists {
		utils.RespondNotFound(c, slog.Default(), "computer")
		return
	}

	slog.InfoContext(ctx, "platform_admin: computer action requested",
		"actor_id", actor.ID, "computer_id", computerID, "action", action)

	// Audit log — best-effort.
	_, auditErr := h.pool.Exec(ctx,
		`INSERT INTO platform_audit_log
		 (actor_id, actor_role, action, target_type, target_id, details, ip_address)
		 VALUES ($1, 'superadmin', $2, 'computer', $3, $4, $5)`,
		actor.ID,
		action,
		computerID,
		map[string]string{"computer_id": computerID},
		c.ClientIP(),
	)
	if auditErr != nil {
		slog.WarnContext(ctx, "platform_admin: audit log insert failed",
			"actor_id", actor.ID, "computer_id", computerID, "action", action, "error", auditErr)
	}

	// TODO: wire MIOSA compute client call here (action: restart/stop/hibernate).

	c.JSON(http.StatusOK, gin.H{"status": action + "ing"})
}

// RestartComputer handles POST /api/v1/admin/computers/:id/restart.
func (h *PlatformAdminHandler) RestartComputer(c *gin.Context) {
	h.computeAction(c, "restart")
}

// StopComputer handles POST /api/v1/admin/computers/:id/stop.
func (h *PlatformAdminHandler) StopComputer(c *gin.Context) {
	h.computeAction(c, "stop")
}

// HibernateComputer handles POST /api/v1/admin/computers/:id/hibernate.
func (h *PlatformAdminHandler) HibernateComputer(c *gin.Context) {
	h.computeAction(c, "hibernate")
}

// GetSettings handles GET /api/v1/admin/settings.
// Exposes non-secret platform configuration values for the superadmin UI.
func (h *PlatformAdminHandler) GetSettings(c *gin.Context) {
	cfg := h.cfg
	if cfg == nil {
		c.JSON(http.StatusOK, gin.H{"error": "config not available"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"allowed_origins": cfg.AllowedOrigins,
		"environment":     cfg.Environment,
		"deployment_mode": cfg.DeploymentMode,
		"ai_provider":     cfg.AIProvider,
		"default_model":   cfg.DefaultModel,
		"osa_enabled":     cfg.OSAEnabled,
		"carrier_enabled": false, // not yet implemented
		"nats_enabled":    cfg.NATSEnabled,
	})
}
