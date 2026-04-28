package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/optimal/audit"
	"github.com/rhl/businessos-backend/internal/optimal/wiki"
	"github.com/rhl/businessos-backend/internal/utils"
)

// WikiHandler exposes the OptimalEngine wiki facade as REST endpoints.
// The store under it is wiki.PagesAdapter, which writes through the existing
// BusinessOS contexts table — wiki pages ARE Pages with extra metadata in
// contexts.properties. See internal/optimal/wiki/store_pages_adapter.go.
//
// Tenant scoping: every request derives the effective tenant_id from
// workspace_id (path or query). Falls back to the better-auth user.id
// for personal-workspace queries.
type WikiHandler struct {
	wiki  *wiki.Wiki
	audit *audit.Logger
}

// NewWikiHandler wires the handler to the existing connection pool.
// Construction is lazy: callers should funnel through registerWikiRoutes
// rather than embedding the constructor in NewHandlers (so single-instance
// desktop deployments without a DB don't pay startup cost).
func NewWikiHandler(pool *pgxpool.Pool) *WikiHandler {
	db := stdlib.OpenDBFromPool(pool)
	store := wiki.NewPagesAdapter(db)
	return &WikiHandler{
		wiki:  wiki.New(store),
		audit: audit.NewLogger(db),
	}
}

// RegisterWikiRoutes attaches all /api/wiki/* routes. auth is the request
// authentication middleware (cached or direct); RequireAuth is layered on
// top so a missing user always 401s before the handler runs.
func RegisterWikiRoutes(api *gin.RouterGroup, h *WikiHandler, auth gin.HandlerFunc) {
	g := api.Group("/wiki")
	g.Use(auth, middleware.RequireAuth())
	{
		g.GET("", h.List)
		g.GET("/:slug", h.Get)
		g.POST("/curate", h.Curate)
		g.POST("/:slug/render", h.Render)
		g.POST("/:slug/verify", h.Verify)
	}
}

// tenantFromContext resolves the effective tenant id for a wiki request.
// Priority: explicit ?tenant_id=… → user.ID. The reserved "default" tenant
// is never auto-selected here — callers wanting it must pass it explicitly.
func tenantFromContext(c *gin.Context) string {
	if t := c.Query("tenant_id"); t != "" {
		return t
	}
	if user := middleware.GetCurrentUser(c); user != nil {
		return user.ID
	}
	return ""
}

// List returns every wiki-tagged page for the effective tenant, latest
// version per (slug, audience).
//
//	GET /api/wiki?tenant_id=…
func (h *WikiHandler) List(c *gin.Context) {
	tenant := tenantFromContext(c)
	pages, err := h.wiki.List(c.Request.Context(), tenant)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "wiki: list failed",
			"tenant_id", tenant, "error", err)
		utils.RespondInternalError(c, slog.Default(), "list wiki pages", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"tenant_id": tenant,
		"pages":     pagesToJSON(pages),
		"count":     len(pages),
	})
}

// Get returns the latest version of a wiki page.
//
//	GET /api/wiki/:slug?audience=…&tenant_id=…
func (h *WikiHandler) Get(c *gin.Context) {
	tenant := tenantFromContext(c)
	slug := c.Param("slug")
	audience := c.DefaultQuery("audience", wiki.DefaultAudience)

	page, err := h.wiki.Get(c.Request.Context(), tenant, slug, audience)
	if errors.Is(err, wiki.ErrNotFound) {
		utils.RespondNotFound(c, slog.Default(), "wiki page")
		return
	}
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "wiki: get failed",
			"tenant_id", tenant, "slug", slug, "audience", audience, "error", err)
		utils.RespondInternalError(c, slog.Default(), "get wiki page", err)
		return
	}
	h.audit.LogKind(c.Request.Context(), audit.KindWikiRead,
		audit.WithTenant(tenant),
		audit.WithPrincipal(principalIDFromContext(c)),
		audit.WithTargetURI("wiki:"+slug+":"+audience),
	)
	c.JSON(http.StatusOK, pageToJSON(page))
}

// Curate persists a new page version. The body is supplied verbatim; an
// LLM-driven curator that *generates* the body is wired separately at the
// agent layer (see chat/agent flow).
//
//	POST /api/wiki/curate
//	{
//	  "tenant_id": "…", "slug": "…", "audience": "default",
//	  "body": "# Page\n\nbody…",
//	  "citations": [{"chunk_id": "optimal://…", "claim_hash": "…"}]
//	}
func (h *WikiHandler) Curate(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	var req struct {
		TenantID  string `json:"tenant_id"`
		Slug      string `json:"slug"      binding:"required"`
		Audience  string `json:"audience"`
		Body      string `json:"body"      binding:"required"`
		Citations []struct {
			ChunkID   string `json:"chunk_id" binding:"required"`
			ClaimHash string `json:"claim_hash"`
		} `json:"citations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}
	if req.TenantID == "" {
		req.TenantID = user.ID
	}

	citations := make([]wiki.Citation, len(req.Citations))
	for i, ct := range req.Citations {
		citations[i] = wiki.Citation{ChunkID: ct.ChunkID, ClaimHash: ct.ClaimHash}
	}

	// The HTTP path supplies the body directly — no LLM round-trip. The
	// wiki facade still wants a CurateFn so it can compute version + persist
	// citations, so we wrap req.Body into a no-op curator.
	curator := func(_ context.Context, _ *wiki.Page, _ []wiki.Citation) (string, error) {
		return req.Body, nil
	}
	page, err := h.wiki.Curate(c.Request.Context(), req.TenantID, req.Slug, citations, curator,
		wiki.CurateOptions{Audience: req.Audience, CuratedBy: user.ID},
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "wiki: curate failed",
			"tenant_id", req.TenantID, "slug", req.Slug, "error", err)
		utils.RespondInternalError(c, slog.Default(), "curate wiki page", err)
		return
	}

	h.audit.LogKind(c.Request.Context(), audit.KindWikiCurated,
		audit.WithTenant(req.TenantID),
		audit.WithPrincipal(user.ID),
		audit.WithTargetURI("wiki:"+req.Slug+":"+page.Audience),
		audit.WithMetadata(map[string]any{
			"version":         page.Version,
			"citation_count":  len(citations),
			"body_byte_count": len(req.Body),
		}),
	)
	c.JSON(http.StatusOK, pageToJSON(page))
}

// Render walks a page's directives and resolves them into the requested
// output format (plain | markdown | claude | openai). Resolvers for `cite`
// echo the URI; richer resolvers (include, expand) return the raw directive
// unchanged with a metadata note — they require pipeline integration that
// lands separately.
//
//	POST /api/wiki/:slug/render
//	{ "audience": "default", "format": "markdown" }
func (h *WikiHandler) Render(c *gin.Context) {
	tenant := tenantFromContext(c)
	slug := c.Param("slug")
	var req struct {
		Audience string `json:"audience"`
		Format   string `json:"format"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Audience == "" {
		req.Audience = wiki.DefaultAudience
	}
	if req.Format == "" {
		req.Format = string(wiki.FormatMarkdown)
	}

	page, err := h.wiki.Get(c.Request.Context(), tenant, slug, req.Audience)
	if errors.Is(err, wiki.ErrNotFound) {
		utils.RespondNotFound(c, slog.Default(), "wiki page")
		return
	}
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "wiki: render get failed",
			"tenant_id", tenant, "slug", slug, "error", err)
		utils.RespondInternalError(c, slog.Default(), "load wiki page", err)
		return
	}

	resolver := func(d wiki.Directive) (wiki.Resolved, error) {
		// Phase 1 resolver: cite directives echo their URI; everything else
		// preserves the original directive in the output. Pipeline-backed
		// include/expand/search resolvers land in a follow-up handler that
		// has access to the engine + retrieval services.
		if d.Verb == "cite" {
			return wiki.Resolved{Source: d.Argument}, nil
		}
		return wiki.Resolved{Content: d.Raw}, nil
	}
	result := h.wiki.Render(page, resolver, wiki.Format(req.Format))
	c.JSON(http.StatusOK, gin.H{
		"slug":      page.Slug,
		"audience":  page.Audience,
		"version":   page.Version,
		"format":    req.Format,
		"body":      result.Body,
		"citations": result.Citations,
		"errors":    formatRenderErrors(result.Errors),
	})
}

// Verify runs the integrity checker against a page (whitelisted verbs,
// citation resolution, claim density, page size, optional schema rules).
//
//	POST /api/wiki/:slug/verify
//	{
//	  "audience": "default",
//	  "schema": { "required_sections": ["Summary"], "min_citations": 3, "max_bytes": 65536 }
//	}
func (h *WikiHandler) Verify(c *gin.Context) {
	tenant := tenantFromContext(c)
	slug := c.Param("slug")
	var req struct {
		Audience string `json:"audience"`
		Schema   *struct {
			RequiredSections []string `json:"required_sections"`
			MinCitations     int      `json:"min_citations"`
			MaxBytes         int      `json:"max_bytes"`
		} `json:"schema"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Audience == "" {
		req.Audience = wiki.DefaultAudience
	}

	page, err := h.wiki.Get(c.Request.Context(), tenant, slug, req.Audience)
	if errors.Is(err, wiki.ErrNotFound) {
		utils.RespondNotFound(c, slog.Default(), "wiki page")
		return
	}
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "wiki: verify get failed",
			"tenant_id", tenant, "slug", slug, "error", err)
		utils.RespondInternalError(c, slog.Default(), "load wiki page", err)
		return
	}

	opts := wiki.CheckOptions{}
	var report wiki.Report
	if req.Schema != nil {
		report = h.wiki.VerifyAgainstSchema(page, wiki.SchemaRules{
			RequiredSections: req.Schema.RequiredSections,
			MinCitations:     req.Schema.MinCitations,
			MaxBytes:         req.Schema.MaxBytes,
		}, opts)
	} else {
		report = h.wiki.Verify(page, opts)
	}
	c.JSON(http.StatusOK, gin.H{
		"slug":        report.PageSlug,
		"ok":          report.OK,
		"issue_count": len(report.Issues),
		"issues":      issuesToJSON(report.Issues),
	})
}

// ── view-model helpers ────────────────────────────────────────────────────

func pagesToJSON(pages []*wiki.Page) []map[string]any {
	out := make([]map[string]any, len(pages))
	for i, p := range pages {
		out[i] = pageToJSON(p)
	}
	return out
}

func pageToJSON(p *wiki.Page) map[string]any {
	return map[string]any{
		"tenant_id":    p.TenantID,
		"slug":         p.Slug,
		"audience":     p.Audience,
		"version":      p.Version,
		"frontmatter":  p.Frontmatter,
		"body":         p.Body,
		"last_curated": p.LastCurated,
		"curated_by":   p.CuratedBy,
	}
}

func issuesToJSON(issues []wiki.Issue) []map[string]any {
	out := make([]map[string]any, len(issues))
	for i, iss := range issues {
		out[i] = map[string]any{
			"severity": string(iss.Severity),
			"kind":     iss.Kind,
			"message":  iss.Message,
			"location": iss.Location,
		}
	}
	return out
}

func formatRenderErrors(errs []error) []string {
	out := make([]string, len(errs))
	for i, e := range errs {
		out[i] = e.Error()
	}
	return out
}

// principalIDFromContext returns the conventional "user:<id>" string used
// for audit principal fields. Empty when no user is set.
func principalIDFromContext(c *gin.Context) string {
	if user := middleware.GetCurrentUser(c); user != nil {
		return "user:" + user.ID
	}
	return "system"
}
