package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Miosa-osa/OptimalEngine-go/audit"
	"github.com/Miosa-osa/OptimalEngine-go/connectors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ConnectorsHandler triggers a one-batch sync against any registered
// connector adapter. Phase 1 supports gmail end-to-end (real OAuth fetcher
// wired in registerIntegrationRoutes); other adapters return 501 until
// their Sync paths land.
//
// The handler is deliberately thin: it builds the per-kind Config from the
// authenticated user's stored credentials, then delegates everything else
// to connectors.Run. That keeps adapter logic in adapters/ and credential
// resolution in services/.
type ConnectorsHandler struct {
	pool  *pgxpool.Pool
	audit *audit.Logger
	// resolveConfig builds the per-kind Config from request context.
	// The adapter package handles the rest (Init, Sync, error categories).
	resolveConfig func(c *gin.Context, kind, userID string) (connectors.Config, error)
}

// NewConnectorsHandler wires the handler with the standard credential
// resolver. Tenants may override resolveConfig at runtime if they need
// per-kind shortcuts (e.g. a service-account flow), but the default is
// suitable for every adapter shipped today.
func NewConnectorsHandler(pool *pgxpool.Pool) *ConnectorsHandler {
	db := stdlib.OpenDBFromPool(pool)
	return &ConnectorsHandler{
		pool:          pool,
		audit:         audit.NewLogger(db),
		resolveConfig: defaultConfigResolver,
	}
}

// RegisterConnectorsRoutes attaches /api/connectors/* endpoints. Auth-gated.
func RegisterConnectorsRoutes(api *gin.RouterGroup, h *ConnectorsHandler, auth gin.HandlerFunc) {
	g := api.Group("/connectors")
	g.Use(auth, middleware.RequireAuth())
	{
		g.GET("", h.List)
		g.POST("/:kind/sync", h.Sync)
	}
}

// List returns every adapter kind registered in this binary. Useful for the
// admin UI to render the supported set without hard-coding it.
//
//	GET /api/connectors
func (h *ConnectorsHandler) List(c *gin.Context) {
	summaries := connectors.List()
	out := make([]map[string]any, len(summaries))
	for i, s := range summaries {
		out[i] = map[string]any{
			"kind":         s.Kind,
			"display_name": s.DisplayName,
			"auth_scheme":  string(s.AuthScheme),
		}
	}
	c.JSON(http.StatusOK, gin.H{"connectors": out, "count": len(out)})
}

// Sync runs one batch against a connector. The Phase-1 contract is
// deliberately small: the caller passes an optional cursor (for pagination
// across calls), the handler runs MaxBatches=1, and the response includes
// the resulting Signals + the next cursor for the client to feed back.
//
//	POST /api/connectors/:kind/sync
//	{ "cursor": "..." }
//
// Response:
//
//	{
//	  "kind":          "gmail",
//	  "batches":       1,
//	  "signal_count":  42,
//	  "next_cursor":   "...",
//	  "signals":       [{ id, title, content, genre, entities[], modified_at }]
//	}
func (h *ConnectorsHandler) Sync(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}
	kind := c.Param("kind")
	if _, err := connectors.Lookup(kind); err != nil {
		utils.RespondNotFound(c, slog.Default(), "connector kind \""+kind+"\"")
		return
	}

	var req struct {
		Cursor string `json:"cursor"`
	}
	_ = c.ShouldBindJSON(&req) // empty body is fine

	cfg, err := h.resolveConfig(c, kind, user.ID)
	if err != nil {
		// Credential-resolution failures are typically 412 — the user
		// hasn't completed the OAuth flow for this kind. We return 400
		// with the error message for clarity.
		utils.RespondBadRequest(c, slog.Default(), err.Error())
		return
	}

	// We collect signals in-memory so the response can include them. For
	// production-scale ingest, swap MakeOnBatch with a streaming pipeline.
	collected := []connectors.Signal{}
	opts := connectors.RunOptions{
		MaxBatches:    1,
		InitialCursor: connectors.Cursor(req.Cursor),
		OnBatch: func(_ context.Context, batch []connectors.Signal) error {
			collected = append(collected, batch...)
			return nil
		},
	}
	result, err := connectors.Run(c.Request.Context(), kind, cfg, opts)
	if err != nil {
		// ErrNotImplemented on a stub-only adapter. Surface as 501 so the
		// frontend can render "configure first" messaging instead of a
		// generic 500.
		var syncErr *connectors.SyncError
		if errors.As(err, &syncErr) && syncErr.Category == connectors.ErrNotImplemented {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "this connector's sync path is not yet wired",
				"kind":  kind,
			})
			return
		}
		slog.ErrorContext(c.Request.Context(), "connectors: sync failed",
			"kind", kind, "user_id", user.ID, "error", err)
		utils.RespondInternalError(c, slog.Default(), "connector sync", err)
		return
	}

	h.audit.LogKind(c.Request.Context(), audit.KindConnectorSynced,
		audit.WithPrincipal("user:"+user.ID),
		audit.WithTargetURI("connector:"+kind),
		audit.WithMetadata(map[string]any{
			"batches":      result.Batches,
			"signal_count": result.SignalCount,
		}),
	)

	signals := make([]map[string]any, len(collected))
	for i, s := range collected {
		signals[i] = map[string]any{
			"id":          s.ID,
			"path":        s.Path,
			"title":       s.Title,
			"content":     s.Content,
			"genre":       s.Genre,
			"entities":    s.Entities,
			"modified_at": s.ModifiedAt,
			"metadata":    s.Metadata,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"kind":         result.Kind,
		"batches":      result.Batches,
		"signal_count": result.SignalCount,
		"next_cursor":  string(result.FinalCursor),
		"signals":      signals,
	})
}

// defaultConfigResolver builds a Config for the given kind. Each kind that
// has a wired Fetcher (gmail, slack, notion, linear, hubspot) gets a
// uniform "user_email = better-auth user.id" config; the per-kind fetcher
// pulls live OAuth tokens directly from the provider's token store.
//
// Kinds without a Fetcher fall through to the catch-all error which the
// handler maps to 400 — preferable to a generic 500 because the message
// names the missing wiring explicitly.
func defaultConfigResolver(_ *gin.Context, kind, userID string) (connectors.Config, error) {
	switch kind {
	case "gmail", "slack", "notion", "linear", "hubspot", "drive", "onedrive", "teams":
		return connectors.Config{"user_email": userID}, nil
	default:
		return nil, fmt.Errorf("connector %q has no Sync wiring yet", kind)
	}
}
