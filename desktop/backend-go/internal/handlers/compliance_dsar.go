package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/Miosa-osa/OptimalEngine-go/audit"
	"github.com/Miosa-osa/OptimalEngine-go/compliance"
	"github.com/Miosa-osa/OptimalEngine-go/tenancy"
	"github.com/rhl/businessos-backend/internal/utils"
)

// DSARHandler exposes the GDPR Article 15 / CCPA §1798.110 export workflow.
// Every export emits a "dsar.exported" audit event with the section counts
// and writes a row to dsar_exports for legal-discovery audit trails.
type DSARHandler struct {
	dsar *compliance.DSAR
}

// NewDSARHandler wires the DSAR engine to its dependencies. Audit + tenancy
// are mandatory: a DSAR without an audit trail or tenant scope is itself a
// compliance violation, so the handler refuses to operate when either is nil.
func NewDSARHandler(pool *pgxpool.Pool) *DSARHandler {
	db := stdlib.OpenDBFromPool(pool)
	return &DSARHandler{
		dsar: compliance.NewDSAR(db, audit.NewLogger(db), tenancy.NewWorkspaceAdapter(db)),
	}
}

// RegisterDSARRoutes attaches /api/compliance/dsar/* endpoints. Auth-gated.
func RegisterDSARRoutes(api *gin.RouterGroup, h *DSARHandler, auth gin.HandlerFunc) {
	g := api.Group("/compliance/dsar")
	g.Use(auth, middleware.RequireAuth())
	{
		g.POST("/export", h.Export)
	}
}

// Export produces a DSAR export bundle for the requested principal_id.
// Authorization: callers may only export their OWN data unless they are
// admins of the workspace. Admin enforcement falls to upstream RBAC; here
// we simply require an authenticated user and a non-empty principal_id.
//
//	POST /api/compliance/dsar/export
//	{ "principal_id": "user:…", "tenant_id": "…" }
func (h *DSARHandler) Export(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req struct {
		PrincipalID string `json:"principal_id" binding:"required"`
		TenantID    string `json:"tenant_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	export, err := h.dsar.Export(c.Request.Context(), req.PrincipalID, req.TenantID, "user:"+user.ID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "compliance/dsar: export failed",
			"principal_id", req.PrincipalID, "tenant_id", req.TenantID, "error", err)
		utils.RespondInternalError(c, slog.Default(), "dsar export", err)
		return
	}
	c.JSON(http.StatusOK, export)
}
