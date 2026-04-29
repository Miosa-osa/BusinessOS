package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/Miosa-osa/OptimalEngine-go/identity"
	"github.com/Miosa-osa/OptimalEngine-go/tenancy"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ACLHandler exposes grant / revoke / list / can-check for resource ACLs.
// Phase 1 keeps a permissive default: a resource with NO ACLs is treated as
// world-readable; Phase 8's wiki-first retrieval will tighten that to
// deny-if-absent. The handler exposes the underlying primitive so admin
// tooling can audit + remediate ahead of the policy flip.
type ACLHandler struct {
	store *identity.ACLAdapter
}

// NewACLHandler wires the handler.
func NewACLHandler(pool *pgxpool.Pool) *ACLHandler {
	db := stdlib.OpenDBFromPool(pool)
	return &ACLHandler{store: identity.NewACLAdapter(db)}
}

// RegisterACLRoutes attaches /api/identity/acls/* endpoints.
func RegisterACLRoutes(api *gin.RouterGroup, h *ACLHandler, auth gin.HandlerFunc) {
	g := api.Group("/identity/acls")
	g.Use(auth, middleware.RequireAuth())
	{
		g.POST("", h.Grant)
		g.DELETE("", h.Revoke)
		g.GET("", h.ListForResource)
		g.POST("/check", h.Can)
	}
}

// Grant inserts a permission row.  Exactly one of principal_id / group_id
// must be set — the package enforces this with a CHECK constraint plus a
// client-side validator.
//
//	POST /api/identity/acls
//	{
//	  "tenant_id": "…",
//	  "resource_uri": "wiki:my-slug:default",
//	  "principal_id": "user:alice",
//	  "permission": "read|write|admin"
//	}
func (h *ACLHandler) Grant(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req aclRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}
	acl, err := req.toACL()
	if err != nil {
		utils.RespondBadRequest(c, slog.Default(), err.Error())
		return
	}
	if err := h.store.Grant(c.Request.Context(), acl); err != nil {
		if errors.Is(err, identity.ErrACLConflict) {
			utils.RespondBadRequest(c, slog.Default(), err.Error())
			return
		}
		slog.ErrorContext(c.Request.Context(), "identity: grant acl failed",
			"resource", req.ResourceURI, "error", err)
		utils.RespondInternalError(c, slog.Default(), "grant acl", err)
		return
	}
	c.Status(http.StatusCreated)
}

// Revoke removes a specific grant.
//
//	DELETE /api/identity/acls
//	{ "tenant_id": "…", "resource_uri": "…",
//	  "principal_id": "…" | "group_id": "…",
//	  "permission": "read|write|admin" }
func (h *ACLHandler) Revoke(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req aclRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}
	if (req.PrincipalID == "") == (req.GroupID == "") {
		utils.RespondBadRequest(c, slog.Default(),
			"exactly one of principal_id / group_id must be set")
		return
	}

	if err := h.store.Revoke(c.Request.Context(),
		req.TenantID, req.ResourceURI, req.PrincipalID, req.GroupID,
		identity.Permission(req.Permission),
	); err != nil {
		slog.ErrorContext(c.Request.Context(), "identity: revoke acl failed",
			"resource", req.ResourceURI, "error", err)
		utils.RespondInternalError(c, slog.Default(), "revoke acl", err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListForResource returns every ACL touching a resource URI.
//
//	GET /api/identity/acls?tenant_id=…&resource_uri=…
func (h *ACLHandler) ListForResource(c *gin.Context) {
	resource := c.Query("resource_uri")
	if resource == "" {
		utils.RespondBadRequest(c, slog.Default(), "resource_uri query parameter is required")
		return
	}
	tenantID := tenancy.Resolve(c.Query("tenant_id"))
	acls, err := h.store.ListForResource(c.Request.Context(), tenantID, resource)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "identity: list acls failed",
			"resource", resource, "error", err)
		utils.RespondInternalError(c, slog.Default(), "list acls", err)
		return
	}
	out := make([]map[string]any, len(acls))
	for i, a := range acls {
		out[i] = aclToJSON(a)
	}
	c.JSON(http.StatusOK, gin.H{"acls": out, "count": len(out)})
}

// Can answers a permission check.  Useful for the frontend to gate UI on
// the same primitive the backend uses for retrieval.
//
//	POST /api/identity/acls/check
//	{ "tenant_id": "…", "resource_uri": "…", "principal_id": "user:…",
//	  "permission": "read", "groups": ["engineers"] }
func (h *ACLHandler) Can(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req struct {
		TenantID    string   `json:"tenant_id"`
		ResourceURI string   `json:"resource_uri" binding:"required"`
		PrincipalID string   `json:"principal_id" binding:"required"`
		Permission  string   `json:"permission"   binding:"required"`
		Groups      []string `json:"groups"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	allowed, err := h.store.Can(c.Request.Context(),
		req.PrincipalID, req.ResourceURI, identity.Permission(req.Permission),
		identity.CanOptions{TenantID: req.TenantID, Groups: req.Groups},
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "identity: can check failed",
			"resource", req.ResourceURI, "principal", req.PrincipalID, "error", err)
		utils.RespondInternalError(c, slog.Default(), "can check", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"allowed": allowed})
}

// aclRequest is the shared shape for Grant + Revoke. It would be duplicated
// inline otherwise; centralizing the validation lets toACL handle both.
type aclRequest struct {
	TenantID    string `json:"tenant_id"`
	ResourceURI string `json:"resource_uri" binding:"required"`
	PrincipalID string `json:"principal_id"`
	GroupID     string `json:"group_id"`
	Permission  string `json:"permission"   binding:"required"`
}

func (r *aclRequest) toACL() (*identity.ACL, error) {
	if (r.PrincipalID == "") == (r.GroupID == "") {
		return nil, errors.New("exactly one of principal_id / group_id must be set")
	}
	perm := identity.Permission(r.Permission)
	if identity.PermissionRank[perm] == 0 {
		return nil, errors.New("permission must be read | write | admin")
	}
	return &identity.ACL{
		TenantID:    r.TenantID,
		ResourceURI: r.ResourceURI,
		PrincipalID: r.PrincipalID,
		GroupID:     r.GroupID,
		Permission:  perm,
	}, nil
}

func aclToJSON(a *identity.ACL) map[string]any {
	return map[string]any{
		"id":           a.ID,
		"tenant_id":    a.TenantID,
		"resource_uri": a.ResourceURI,
		"principal_id": a.PrincipalID,
		"group_id":     a.GroupID,
		"permission":   string(a.Permission),
		"granted_at":   a.GrantedAt,
	}
}
