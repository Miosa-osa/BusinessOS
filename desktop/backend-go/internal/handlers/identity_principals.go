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

// PrincipalHandler exposes CRUD for non-human principals (agents, services).
// Human users are managed by better-auth — Upsert refuses to write Kind=user
// rows so the two stores can never drift.
type PrincipalHandler struct {
	store *identity.PrincipalAdapter
}

// NewPrincipalHandler wires the handler.
func NewPrincipalHandler(pool *pgxpool.Pool) *PrincipalHandler {
	db := stdlib.OpenDBFromPool(pool)
	return &PrincipalHandler{store: identity.NewPrincipalAdapter(db)}
}

// RegisterPrincipalRoutes attaches /api/identity/principals/* endpoints.
func RegisterPrincipalRoutes(api *gin.RouterGroup, h *PrincipalHandler, auth gin.HandlerFunc) {
	g := api.Group("/identity/principals")
	g.Use(auth, middleware.RequireAuth())
	{
		g.GET("", h.List)
		g.GET("/:id", h.Get)
		g.POST("", h.Upsert)
	}
}

// List returns every principal for the requested tenant.
//
//	GET /api/identity/principals?tenant_id=…
func (h *PrincipalHandler) List(c *gin.Context) {
	tenantID := tenancy.Resolve(c.Query("tenant_id"))
	principals, err := h.store.List(c.Request.Context(), tenantID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "identity: list principals failed",
			"tenant_id", tenantID, "error", err)
		utils.RespondInternalError(c, slog.Default(), "list principals", err)
		return
	}
	out := make([]map[string]any, len(principals))
	for i, p := range principals {
		out[i] = principalToJSON(p)
	}
	c.JSON(http.StatusOK, gin.H{"principals": out, "count": len(out)})
}

// Get fetches a principal by id, tenant-scoped.
//
//	GET /api/identity/principals/:id?tenant_id=…
func (h *PrincipalHandler) Get(c *gin.Context) {
	id := c.Param("id")
	tenantID := tenancy.Resolve(c.Query("tenant_id"))
	p, err := h.store.Get(c.Request.Context(), id, tenantID)
	if errors.Is(err, identity.ErrPrincipalNotFound) {
		utils.RespondNotFound(c, slog.Default(), "principal")
		return
	}
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "identity: get principal failed",
			"id", id, "tenant_id", tenantID, "error", err)
		utils.RespondInternalError(c, slog.Default(), "get principal", err)
		return
	}
	c.JSON(http.StatusOK, principalToJSON(p))
}

// Upsert creates or updates an agent / service principal. Human users are
// rejected here so they always flow through better-auth.
//
//	POST /api/identity/principals
//	{
//	  "id": "agent:alpha",
//	  "tenant_id": "…",
//	  "kind": "agent|service",
//	  "display_name": "Alpha Agent",
//	  "external_id": "…",
//	  "metadata": {...}
//	}
func (h *PrincipalHandler) Upsert(c *gin.Context) {
	if user := middleware.GetCurrentUser(c); user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req struct {
		ID          string         `json:"id"           binding:"required"`
		TenantID    string         `json:"tenant_id"`
		Kind        string         `json:"kind"         binding:"required"`
		DisplayName string         `json:"display_name" binding:"required"`
		ExternalID  string         `json:"external_id"`
		Metadata    map[string]any `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	kind := identity.Kind(req.Kind)
	if kind != identity.KindAgent && kind != identity.KindService {
		utils.RespondBadRequest(c, slog.Default(),
			"kind must be \"agent\" or \"service\" — human users are managed by auth")
		return
	}

	p := &identity.Principal{
		ID:          req.ID,
		TenantID:    req.TenantID,
		Kind:        kind,
		DisplayName: req.DisplayName,
		ExternalID:  req.ExternalID,
		Metadata:    req.Metadata,
	}
	if err := h.store.Upsert(c.Request.Context(), p); err != nil {
		slog.ErrorContext(c.Request.Context(), "identity: upsert principal failed",
			"id", req.ID, "kind", req.Kind, "error", err)
		utils.RespondInternalError(c, slog.Default(), "upsert principal", err)
		return
	}
	c.JSON(http.StatusOK, principalToJSON(p))
}

func principalToJSON(p *identity.Principal) map[string]any {
	return map[string]any{
		"id":           p.ID,
		"tenant_id":    p.TenantID,
		"kind":         string(p.Kind),
		"display_name": p.DisplayName,
		"external_id":  p.ExternalID,
		"metadata":     p.Metadata,
	}
}
