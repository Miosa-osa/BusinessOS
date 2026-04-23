package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/optimal"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/utils"
)

// OptimalCloudHandler exposes the OptimalOS cloud engine for authenticated users.
// Every endpoint is workspace-scoped and auth-gated.
// The workspace ID is supplied by the caller as either a URL parameter (:workspace_id),
// a query string (?workspace_id=…), or a JSON body field ("workspace_id").
type OptimalCloudHandler struct {
	service *services.OptimalCloudService
	pool    *pgxpool.Pool
}

// NewOptimalCloudHandler creates a ready-to-use handler.
// pool is required for the operations that execute SQL not yet covered by
// OptimalCloudService (list/get/update/delete signals, entities, graph, rethink).
func NewOptimalCloudHandler(pool *pgxpool.Pool) *OptimalCloudHandler {
	logger := slog.Default().With("component", "optimal_cloud_handler")
	return &OptimalCloudHandler{
		service: services.NewOptimalCloudService(pool, logger),
		pool:    pool,
	}
}

// RegisterOptimalCloudRoutes attaches all /optimal-cloud routes to api.
// auth must be the AuthMiddleware (or CachedAuthMiddleware) for this server instance.
func RegisterOptimalCloudRoutes(api *gin.RouterGroup, h *OptimalCloudHandler, auth gin.HandlerFunc) {
	g := api.Group("/optimal-cloud")
	g.Use(auth, middleware.RequireAuth())
	{
		// ── Signals ───────────────────────────────────────────────────────────
		g.POST("/signals/ingest", h.IngestSignal)
		g.POST("/signals/search", h.SearchSignals)
		g.GET("/signals", h.ListSignals)
		g.GET("/signals/:id", h.GetSignal)
		g.PUT("/signals/:id", h.UpdateSignal)
		g.DELETE("/signals/:id", h.DeleteSignal)

		// ── Context assembly + composition ───────────────────────────────────
		g.POST("/assemble", h.AssembleContext)
		g.POST("/compose", h.Compose)

		// ── Knowledge graph ───────────────────────────────────────────────────
		g.GET("/entities", h.ListEntities)
		g.GET("/graph", h.GetGraph)
		g.GET("/graph/hubs", h.GetGraphHubs)

		// ── Learning loop ─────────────────────────────────────────────────────
		g.POST("/remember", h.Remember)
		g.GET("/escalations", h.GetEscalations)
		g.POST("/rethink", h.Rethink)

		// ── Topology ─────────────────────────────────────────────────────────
		g.GET("/topology", h.GetTopology)
		g.PUT("/topology", h.UpdateTopology)

		// ── Meta ──────────────────────────────────────────────────────────────
		g.GET("/stats", h.GetStats)
		g.POST("/initialize", h.InitializeWorkspace)
	}
}

// ── workspace ID resolution ───────────────────────────────────────────────────

// workspaceIDFromQuery reads workspace_id from the query string.
// Returns ("", false) when absent or invalid.
func workspaceIDFromQuery(c *gin.Context) (string, bool) {
	id := c.Query("workspace_id")
	if id == "" {
		return "", false
	}
	if _, err := uuid.Parse(id); err != nil {
		return "", false
	}
	return id, true
}

// ── Signal endpoints ──────────────────────────────────────────────────────────

// IngestSignal runs the full OptimalOS classify→route→store pipeline.
//
//	POST /api/optimal-cloud/signals/ingest
func (h *OptimalCloudHandler) IngestSignal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		WorkspaceID string `json:"workspace_id" binding:"required"`
		Text        string `json:"text"         binding:"required"`
		Genre       string `json:"genre"`
		NodeID      string `json:"node_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	result, err := h.service.IngestSignal(c.Request.Context(), services.IngestRequest{
		WorkspaceID: req.WorkspaceID,
		UserID:      user.ID,
		Text:        req.Text,
		Genre:       req.Genre,
		NodeID:      req.NodeID,
	})
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: ingest failed",
			"error", err, "workspace_id", req.WorkspaceID, "user_id", user.ID)
		utils.RespondInternalError(c, slog.Default(), "ingest signal", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"signal_id":      result.SignalID,
		"classification": result.Classification,
		"routing":        result.Routing,
		"l0_abstract":    result.L0Abstract,
		"l1_overview":    result.L1Overview,
	})
}

// ListSignals returns paginated signals for a workspace, optionally filtered.
//
//	GET /api/optimal-cloud/signals?workspace_id=…&node_id=…&genre=…&limit=…&offset=…
func (h *OptimalCloudHandler) ListSignals(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}

	// Optional filters.
	nodeID := c.Query("node_id")
	genre := c.Query("genre")
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if o := c.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	workspaceUUID, _ := uuid.Parse(workspaceID)

	// Build dynamic query.
	args := []any{workspaceUUID}
	where := "workspace_id = $1 AND valid_until IS NULL"

	if nodeID != "" {
		if parsed, err := uuid.Parse(nodeID); err == nil {
			args = append(args, parsed)
			where += " AND node_id = $" + strconv.Itoa(len(args))
		}
	}
	if genre != "" {
		args = append(args, genre)
		where += " AND genre = $" + strconv.Itoa(len(args))
	}

	args = append(args, limit, offset)
	limitPos := len(args) - 1
	offsetPos := len(args)

	query := "SELECT id, title, l0_abstract, genre, primary_node, created_at, updated_at " +
		"FROM optimal_signals WHERE " + where +
		" ORDER BY created_at DESC" +
		" LIMIT $" + strconv.Itoa(limitPos) +
		" OFFSET $" + strconv.Itoa(offsetPos)

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: list signals failed",
			"error", err, "workspace_id", workspaceID)
		utils.RespondInternalError(c, slog.Default(), "list signals", err)
		return
	}
	defer rows.Close()

	type signalRow struct {
		ID          string    `json:"id"`
		Title       string    `json:"title"`
		L0Abstract  string    `json:"l0_abstract"`
		Genre       string    `json:"genre"`
		PrimaryNode string    `json:"primary_node"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	signals := make([]signalRow, 0)
	for rows.Next() {
		var r signalRow
		var id uuid.UUID
		if err := rows.Scan(&id, &r.Title, &r.L0Abstract, &r.Genre, &r.PrimaryNode, &r.CreatedAt, &r.UpdatedAt); err != nil {
			slog.ErrorContext(c.Request.Context(), "optimal_cloud: scan signal row failed", "error", err)
			utils.RespondInternalError(c, slog.Default(), "list signals", err)
			return
		}
		r.ID = id.String()
		signals = append(signals, r)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "list signals", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"signals": signals})
}

// GetSignal returns a single signal by ID (workspace-scoped).
//
//	GET /api/optimal-cloud/signals/:id?workspace_id=…
func (h *OptimalCloudHandler) GetSignal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	signalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "signal_id")
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}
	workspaceUUID, _ := uuid.Parse(workspaceID)

	type signalDetail struct {
		ID          string          `json:"id"`
		Title       string          `json:"title"`
		Content     string          `json:"content"`
		L0Abstract  string          `json:"l0_abstract"`
		L1Overview  string          `json:"l1_overview"`
		Genre       string          `json:"genre"`
		Mode        string          `json:"mode"`
		SignalType  string          `json:"signal_type"`
		PrimaryNode string          `json:"primary_node"`
		CrossRef    json.RawMessage `json:"cross_ref_nodes"`
		Entities    json.RawMessage `json:"entities"`
		SNRatio     float64         `json:"sn_ratio"`
		CreatedAt   time.Time       `json:"created_at"`
		UpdatedAt   time.Time       `json:"updated_at"`
	}

	var row signalDetail
	var id uuid.UUID
	err = h.pool.QueryRow(c.Request.Context(), `
		SELECT id, title, content, l0_abstract, l1_overview,
		       genre, mode, signal_type, primary_node,
		       cross_ref_nodes, entities, sn_ratio,
		       created_at, updated_at
		FROM optimal_signals
		WHERE id = $1 AND workspace_id = $2 AND valid_until IS NULL`,
		signalID, workspaceUUID,
	).Scan(&id, &row.Title, &row.Content, &row.L0Abstract, &row.L1Overview,
		&row.Genre, &row.Mode, &row.SignalType, &row.PrimaryNode,
		&row.CrossRef, &row.Entities, &row.SNRatio,
		&row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Signal")
		return
	}
	row.ID = id.String()

	c.JSON(http.StatusOK, row)
}

// UpdateSignal replaces the mutable content of a signal.
//
//	PUT /api/optimal-cloud/signals/:id
func (h *OptimalCloudHandler) UpdateSignal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	signalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "signal_id")
		return
	}

	var req struct {
		WorkspaceID string `json:"workspace_id" binding:"required"`
		Content     string `json:"content"      binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "workspace_id")
		return
	}

	tag, err := h.pool.Exec(c.Request.Context(), `
		UPDATE optimal_signals
		SET content = $1, updated_at = NOW()
		WHERE id = $2 AND workspace_id = $3 AND valid_until IS NULL`,
		req.Content, signalID, workspaceUUID,
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: update signal failed",
			"error", err, "signal_id", signalID)
		utils.RespondInternalError(c, slog.Default(), "update signal", err)
		return
	}
	if tag.RowsAffected() == 0 {
		utils.RespondNotFound(c, slog.Default(), "Signal")
		return
	}

	slog.InfoContext(c.Request.Context(), "optimal_cloud: signal updated",
		"signal_id", signalID, "workspace_id", req.WorkspaceID, "user_id", user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Signal updated"})
}

// DeleteSignal soft-deletes a signal by setting valid_until.
//
//	DELETE /api/optimal-cloud/signals/:id
func (h *OptimalCloudHandler) DeleteSignal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	signalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "signal_id")
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}
	workspaceUUID, _ := uuid.Parse(workspaceID)

	tag, err := h.pool.Exec(c.Request.Context(), `
		UPDATE optimal_signals
		SET valid_until = NOW(), updated_at = NOW()
		WHERE id = $1 AND workspace_id = $2 AND valid_until IS NULL`,
		signalID, workspaceUUID,
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: delete signal failed",
			"error", err, "signal_id", signalID)
		utils.RespondInternalError(c, slog.Default(), "delete signal", err)
		return
	}
	if tag.RowsAffected() == 0 {
		utils.RespondNotFound(c, slog.Default(), "Signal")
		return
	}

	slog.InfoContext(c.Request.Context(), "optimal_cloud: signal deleted",
		"signal_id", signalID, "workspace_id", workspaceID, "user_id", user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Signal deleted"})
}

// SearchSignals performs weighted full-text search with temporal decay.
//
//	POST /api/optimal-cloud/signals/search
func (h *OptimalCloudHandler) SearchSignals(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		WorkspaceID string `json:"workspace_id" binding:"required"`
		Query       string `json:"query"        binding:"required"`
		NodeID      string `json:"node_id"`
		Genre       string `json:"genre"`
		Limit       int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	results, err := h.service.SearchSignals(c.Request.Context(), services.SearchRequest{
		WorkspaceID: req.WorkspaceID,
		Query:       req.Query,
		NodeID:      req.NodeID,
		Genre:       req.Genre,
		Limit:       req.Limit,
	})
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: search failed",
			"error", err, "workspace_id", req.WorkspaceID, "user_id", user.ID)
		utils.RespondInternalError(c, slog.Default(), "search signals", err)
		return
	}

	if results == nil {
		results = []services.OptimalSearchResult{}
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// ── Context assembly + composition ────────────────────────────────────────────

// AssembleContext builds a budget-aware tiered context from workspace signals.
//
//	POST /api/optimal-cloud/assemble
func (h *OptimalCloudHandler) AssembleContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		WorkspaceID     string                  `json:"workspace_id" binding:"required"`
		Query           string                  `json:"query"        binding:"required"`
		ReceiverProfile optimal.ReceiverProfile `json:"receiver"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "workspace_id")
		return
	}

	// Default budget when receiver profile is incomplete.
	if req.ReceiverProfile.TokenBudget <= 0 {
		req.ReceiverProfile.TokenBudget = 10000
	}

	// Fetch ranked signals from the workspace.
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, title, l0_abstract, l1_overview, genre, created_at,
		       ts_rank_cd(search_vector, websearch_to_tsquery('english', $2)) AS rank
		FROM optimal_signals
		WHERE workspace_id = $1
		  AND valid_until IS NULL
		  AND search_vector @@ websearch_to_tsquery('english', $2)
		ORDER BY rank DESC
		LIMIT 50`,
		workspaceUUID, req.Query,
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: assemble query failed",
			"error", err, "workspace_id", req.WorkspaceID)
		utils.RespondInternalError(c, slog.Default(), "assemble context", err)
		return
	}
	defer rows.Close()

	type sourceSignal struct {
		id         uuid.UUID
		title      string
		l0Abstract string
		l1Overview string
		genre      string
		createdAt  time.Time
		rank       float64
	}

	var sources []sourceSignal
	for rows.Next() {
		var s sourceSignal
		if err := rows.Scan(&s.id, &s.title, &s.l0Abstract, &s.l1Overview,
			&s.genre, &s.createdAt, &s.rank); err != nil {
			continue
		}
		sources = append(sources, s)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "assemble context", err)
		return
	}

	// Assemble L0 (always-loaded abstracts) and L1 (task-relevant detail)
	// within the token budget using a simple greedy fill strategy.
	// Each token ≈ 4 chars (GPT/Claude approximation).
	const charsPerToken = 4
	l0Budget := req.ReceiverProfile.TokenBudget / 5
	l1Budget := req.ReceiverProfile.TokenBudget - l0Budget

	var l0Parts, l1Parts []string
	l0Used, l1Used := 0, 0

	type ctxSource struct {
		ID     string  `json:"id"`
		Title  string  `json:"title"`
		Genre  string  `json:"genre"`
		Tier   string  `json:"tier"`
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	}

	var ctxSources []ctxSource

	for _, s := range sources {
		l0Chars := len(s.l0Abstract)
		l0Tokens := l0Chars / charsPerToken

		if l0Used+l0Tokens <= l0Budget {
			l0Parts = append(l0Parts, s.l0Abstract)
			l0Used += l0Tokens
			ctxSources = append(ctxSources, ctxSource{
				ID: s.id.String(), Title: s.title, Genre: s.genre,
				Tier: "l0", Score: s.rank, Reason: "query_match",
			})
		}

		l1Chars := len(s.l1Overview)
		l1Tokens := l1Chars / charsPerToken
		if l1Used+l1Tokens <= l1Budget {
			l1Parts = append(l1Parts, s.l1Overview)
			l1Used += l1Tokens
			ctxSources = append(ctxSources, ctxSource{
				ID: s.id.String(), Title: s.title, Genre: s.genre,
				Tier: "l1", Score: s.rank, Reason: "query_match",
			})
		}
	}

	import_strings_join := func(parts []string, sep string) string {
		result := ""
		for i, p := range parts {
			if i > 0 {
				result += sep
			}
			result += p
		}
		return result
	}

	l0Content := import_strings_join(l0Parts, "\n\n")
	l1Content := import_strings_join(l1Parts, "\n\n")

	if ctxSources == nil {
		ctxSources = []ctxSource{}
	}

	c.JSON(http.StatusOK, gin.H{
		"l0":           l0Content,
		"l1":           l1Content,
		"l0_tokens":    l0Used,
		"l1_tokens":    l1Used,
		"total_tokens": l0Used + l1Used,
		"budget_used":  float64(l0Used+l1Used) / float64(req.ReceiverProfile.TokenBudget),
		"query":        req.Query,
		"receiver":     req.ReceiverProfile,
		"sources":      ctxSources,
	})
}

// Compose reformats content for a named receiver's preferred genre.
//
//	POST /api/optimal-cloud/compose
func (h *OptimalCloudHandler) Compose(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		WorkspaceID  string `json:"workspace_id"  binding:"required"`
		Content      string `json:"content"       binding:"required"`
		ReceiverName string `json:"receiver_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	result, err := h.service.ComposeForReceiver(c.Request.Context(), req.WorkspaceID, req.Content, req.ReceiverName)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: compose failed",
			"error", err, "workspace_id", req.WorkspaceID, "user_id", user.ID)
		utils.RespondInternalError(c, slog.Default(), "compose", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ── Knowledge graph ────────────────────────────────────────────────────────────

// ListEntities returns all known entities in the workspace graph.
//
//	GET /api/optimal-cloud/entities?workspace_id=…
func (h *OptimalCloudHandler) ListEntities(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}
	workspaceUUID, _ := uuid.Parse(workspaceID)

	type entityRow struct {
		ID            string    `json:"id"`
		Name          string    `json:"name"`
		CanonicalName string    `json:"canonical_name"`
		Type          string    `json:"type"`
		MentionCount  int       `json:"mention_count"`
		LastSeen      time.Time `json:"last_seen"`
	}

	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, name, canonical_name, type, mention_count, last_seen
		FROM optimal_entities
		WHERE workspace_id = $1
		ORDER BY mention_count DESC`,
		workspaceUUID,
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: list entities failed",
			"error", err, "workspace_id", workspaceID)
		utils.RespondInternalError(c, slog.Default(), "list entities", err)
		return
	}
	defer rows.Close()

	entities := make([]entityRow, 0)
	for rows.Next() {
		var r entityRow
		var id uuid.UUID
		if err := rows.Scan(&id, &r.Name, &r.CanonicalName, &r.Type, &r.MentionCount, &r.LastSeen); err != nil {
			slog.ErrorContext(c.Request.Context(), "optimal_cloud: scan entity row", "error", err)
			utils.RespondInternalError(c, slog.Default(), "list entities", err)
			return
		}
		r.ID = id.String()
		entities = append(entities, r)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "list entities", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"entities": entities})
}

// GetGraph returns the full entity + edge graph for a workspace.
//
//	GET /api/optimal-cloud/graph?workspace_id=…
func (h *OptimalCloudHandler) GetGraph(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}
	workspaceUUID, _ := uuid.Parse(workspaceID)

	type edgeRow struct {
		ID         string    `json:"id"`
		SourceID   string    `json:"source_id"`
		SourceType string    `json:"source_type"`
		TargetID   string    `json:"target_id"`
		TargetType string    `json:"target_type"`
		Relation   string    `json:"relation"`
		Weight     float64   `json:"weight"`
		CreatedAt  time.Time `json:"created_at"`
	}

	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, source_id, source_type, target_id, target_type, relation, weight, created_at
		FROM optimal_edges
		WHERE workspace_id = $1
		ORDER BY weight DESC`,
		workspaceUUID,
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: get graph failed",
			"error", err, "workspace_id", workspaceID)
		utils.RespondInternalError(c, slog.Default(), "get graph", err)
		return
	}
	defer rows.Close()

	edges := make([]edgeRow, 0)
	for rows.Next() {
		var r edgeRow
		var id, srcID, tgtID uuid.UUID
		if err := rows.Scan(&id, &srcID, &r.SourceType, &tgtID, &r.TargetType, &r.Relation, &r.Weight, &r.CreatedAt); err != nil {
			slog.ErrorContext(c.Request.Context(), "optimal_cloud: scan edge row", "error", err)
			utils.RespondInternalError(c, slog.Default(), "get graph", err)
			return
		}
		r.ID = id.String()
		r.SourceID = srcID.String()
		r.TargetID = tgtID.String()
		edges = append(edges, r)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "get graph", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"edges": edges})
}

// GetGraphHubs returns the most connected entities (hub nodes) in the graph.
//
//	GET /api/optimal-cloud/graph/hubs?workspace_id=…&limit=…
func (h *OptimalCloudHandler) GetGraphHubs(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}
	workspaceUUID, _ := uuid.Parse(workspaceID)

	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	type hubRow struct {
		NodeID      string  `json:"node_id"`
		NodeType    string  `json:"node_type"`
		EdgeCount   int     `json:"edge_count"`
		TotalWeight float64 `json:"total_weight"`
	}

	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT source_id, source_type, COUNT(*) AS edge_count, SUM(weight) AS total_weight
		FROM optimal_edges
		WHERE workspace_id = $1
		GROUP BY source_id, source_type
		ORDER BY edge_count DESC
		LIMIT $2`,
		workspaceUUID, limit,
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: get hubs failed",
			"error", err, "workspace_id", workspaceID)
		utils.RespondInternalError(c, slog.Default(), "get graph hubs", err)
		return
	}
	defer rows.Close()

	hubs := make([]hubRow, 0)
	for rows.Next() {
		var r hubRow
		var nodeID uuid.UUID
		if err := rows.Scan(&nodeID, &r.NodeType, &r.EdgeCount, &r.TotalWeight); err != nil {
			slog.ErrorContext(c.Request.Context(), "optimal_cloud: scan hub row", "error", err)
			utils.RespondInternalError(c, slog.Default(), "get graph hubs", err)
			return
		}
		r.NodeID = nodeID.String()
		hubs = append(hubs, r)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "get graph hubs", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"hubs": hubs})
}

// ── Learning loop ──────────────────────────────────────────────────────────────

// Remember stores an observation in the learning loop table.
//
//	POST /api/optimal-cloud/remember
func (h *OptimalCloudHandler) Remember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		WorkspaceID string `json:"workspace_id" binding:"required"`
		Content     string `json:"content"      binding:"required"`
		Category    string `json:"category"     binding:"required"`
		Source      string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if req.Source == "" {
		req.Source = "explicit"
	}

	if err := h.service.Remember(c.Request.Context(),
		req.WorkspaceID, user.ID, req.Content, req.Category, req.Source); err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: remember failed",
			"error", err, "workspace_id", req.WorkspaceID, "user_id", user.ID)
		utils.RespondBadRequest(c, slog.Default(), err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Observation stored"})
}

// GetEscalations returns categories with 3+ observations — patterns ready for synthesis.
//
//	GET /api/optimal-cloud/escalations?workspace_id=…
func (h *OptimalCloudHandler) GetEscalations(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}

	result, err := h.service.GetEscalations(c.Request.Context(), workspaceID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: get escalations failed",
			"error", err, "workspace_id", workspaceID)
		utils.RespondInternalError(c, slog.Default(), "get escalations", err)
		return
	}

	if result == nil {
		result = map[string]int{}
	}

	c.JSON(http.StatusOK, gin.H{"escalations": result})
}

// Rethink synthesises repeated observations in a category into a higher-fidelity signal.
//
//	POST /api/optimal-cloud/rethink
func (h *OptimalCloudHandler) Rethink(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		WorkspaceID string `json:"workspace_id" binding:"required"`
		Category    string `json:"category"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "workspace_id")
		return
	}

	// Load observations for the category.
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT content, source, created_at
		FROM optimal_observations
		WHERE workspace_id = $1 AND category = $2
		ORDER BY created_at ASC`,
		workspaceUUID, req.Category,
	)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: rethink query failed",
			"error", err, "workspace_id", req.WorkspaceID)
		utils.RespondInternalError(c, slog.Default(), "rethink", err)
		return
	}
	defer rows.Close()

	type obs struct {
		Content   string    `json:"content"`
		Source    string    `json:"source"`
		CreatedAt time.Time `json:"created_at"`
	}

	var observations []obs
	for rows.Next() {
		var o obs
		if err := rows.Scan(&o.Content, &o.Source, &o.CreatedAt); err != nil {
			continue
		}
		observations = append(observations, o)
	}
	if err := rows.Err(); err != nil {
		utils.RespondInternalError(c, slog.Default(), "rethink", err)
		return
	}

	if len(observations) == 0 {
		utils.RespondNotFound(c, slog.Default(), "Observations for category")
		return
	}

	// Synthesise into a single signal via the OptimalOS rethink engine.
	// Build a combined text and run the ingest pipeline to produce a durable signal.
	var combined string
	for i, o := range observations {
		if i > 0 {
			combined += "\n\n"
		}
		combined += o.Content
	}

	result, err := h.service.IngestSignal(c.Request.Context(), services.IngestRequest{
		WorkspaceID: req.WorkspaceID,
		UserID:      user.ID,
		Text:        combined,
		Genre:       "pattern", // escalated observations become patterns
	})
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: rethink ingest failed",
			"error", err, "workspace_id", req.WorkspaceID)
		utils.RespondInternalError(c, slog.Default(), "rethink synthesis", err)
		return
	}

	slog.InfoContext(c.Request.Context(), "optimal_cloud: rethink synthesised",
		"signal_id", result.SignalID,
		"workspace_id", req.WorkspaceID,
		"category", req.Category,
		"observation_count", len(observations),
	)

	c.JSON(http.StatusCreated, gin.H{
		"signal_id":         result.SignalID,
		"category":          req.Category,
		"observation_count": len(observations),
		"classification":    result.Classification,
		"routing":           result.Routing,
		"l0_abstract":       result.L0Abstract,
	})
}

// ── Topology ──────────────────────────────────────────────────────────────────

// GetTopology returns the current workspace topology configuration.
//
//	GET /api/optimal-cloud/topology?workspace_id=…
func (h *OptimalCloudHandler) GetTopology(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}
	workspaceUUID, _ := uuid.Parse(workspaceID)

	var configJSON string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT config FROM optimal_topology WHERE workspace_id = $1`,
		workspaceUUID,
	).Scan(&configJSON)
	if err != nil {
		// No topology row → workspace not yet initialized.
		c.JSON(http.StatusOK, gin.H{"topology": nil, "initialized": false})
		return
	}

	var topo optimal.TopologyConfig
	if err := json.Unmarshal([]byte(configJSON), &topo); err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: unmarshal topology failed",
			"error", err, "workspace_id", workspaceID)
		utils.RespondInternalError(c, slog.Default(), "get topology", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"topology": topo, "initialized": true})
}

// UpdateTopology replaces the workspace topology configuration.
//
//	PUT /api/optimal-cloud/topology
func (h *OptimalCloudHandler) UpdateTopology(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		WorkspaceID string                 `json:"workspace_id" binding:"required"`
		Topology    optimal.TopologyConfig `json:"topology"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := h.service.SaveTopology(c.Request.Context(), req.WorkspaceID, req.Topology); err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: save topology failed",
			"error", err, "workspace_id", req.WorkspaceID, "user_id", user.ID)
		utils.RespondInternalError(c, slog.Default(), "update topology", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Topology updated"})
}

// ── Meta ──────────────────────────────────────────────────────────────────────

// GetStats returns aggregate dashboard statistics for a workspace.
//
//	GET /api/optimal-cloud/stats?workspace_id=…
func (h *OptimalCloudHandler) GetStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	workspaceID, ok := workspaceIDFromQuery(c)
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "workspace_id query parameter is required and must be a valid UUID")
		return
	}

	stats, err := h.service.GetDashboardStats(c.Request.Context(), workspaceID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: get stats failed",
			"error", err, "workspace_id", workspaceID)
		utils.RespondInternalError(c, slog.Default(), "get stats", err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// InitializeWorkspace seeds a new workspace with a default 5-node topology.
// Idempotent: safe to call multiple times.
//
//	POST /api/optimal-cloud/initialize
func (h *OptimalCloudHandler) InitializeWorkspace(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		WorkspaceID string `json:"workspace_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := h.service.InitializeWorkspace(c.Request.Context(), req.WorkspaceID, user.ID); err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal_cloud: initialize workspace failed",
			"error", err, "workspace_id", req.WorkspaceID, "user_id", user.ID)
		utils.RespondInternalError(c, slog.Default(), "initialize workspace", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace initialized", "workspace_id": req.WorkspaceID})
}
