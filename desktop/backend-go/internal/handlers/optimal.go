package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/optimal"
	"github.com/rhl/businessos-backend/internal/utils"
)

// OptimalHandler exposes the OptimalOS filesystem and engine as REST endpoints.
// It is stateless — all reads go through the optimal package helpers at request time.
type OptimalHandler struct {
	nodesRoot string // absolute path to OptimalOS nodes/ directory
	osRoot    string // absolute path to OptimalOS root (for rhythm/)
	dbPath    string // absolute path to the SQLite index database (.system/index.db)
	engine    *optimal.EngineConfig
}

// NewOptimalHandler constructs an OptimalHandler.
//   - nodesRoot  — absolute path to the nodes/ folder inside OptimalOS
//   - osRoot     — absolute path to the OptimalOS repo root (contains rhythm/)
//   - enginePath — absolute path to the engine/ directory (contains mix.exs)
//   - dbPath     — absolute path to .system/index.db (SQLite, may be empty string)
func NewOptimalHandler(nodesRoot, osRoot, enginePath, dbPath string) *OptimalHandler {
	return &OptimalHandler{
		nodesRoot: nodesRoot,
		osRoot:    osRoot,
		dbPath:    dbPath,
		engine:    optimal.NewEngineConfig(enginePath),
	}
}

// RegisterOptimalRoutes attaches all /api/optimal/* routes to api.
// These endpoints are intentionally unauthenticated — they serve local desktop
// data that the Tauri frontend reads without user sessions.
func RegisterOptimalRoutes(api *gin.RouterGroup, h *OptimalHandler) {
	g := api.Group("/optimal")
	{
		g.GET("/nodes", h.ListNodes)
		g.GET("/nodes/:slug", h.GetNode)
		g.GET("/nodes/:slug/signals", h.GetNodeSignals)
		g.GET("/rhythm/today", h.GetTodayRhythm)
		g.POST("/search", h.Search)
		g.POST("/ingest", h.Ingest)
		g.POST("/compose", h.Compose)
		g.GET("/nodes/:slug/files", h.GetNodeFileTree)
		g.GET("/nodes/:slug/file/*filepath", h.GetNodeFile)
		g.PUT("/nodes/:slug/file/*filepath", h.SaveNodeFile)
		g.POST("/nodes/:slug/file", h.CreateNodeFile)

		// Graph endpoints — read directly from SQLite for low latency.
		g.GET("/graph", h.GetGraph)
		g.GET("/graph/hubs", h.GetGraphHubs)
		g.GET("/graph/search", h.GraphSearch)

		// Aggregated data endpoints.
		g.GET("/dashboard", h.GetDashboard)
		g.GET("/team", h.GetTeam)
		g.GET("/revenue", h.GetRevenue)
		g.GET("/projects", h.GetProjects)
		g.GET("/rhythm/weekly", h.GetWeeklyRhythm)
		g.GET("/health", h.GetHealth)

		// L9 Operating System endpoints
		g.GET("/diagnose", h.Diagnose)
		g.POST("/reweave", h.Reweave)
		g.POST("/verify", h.VerifyL0)

		// L8 Learning & Feedback endpoints
		g.POST("/remember", h.Remember)
		g.GET("/observations", h.GetObservations)
		g.GET("/escalations", h.GetEscalations)
		g.POST("/rethink", h.Rethink)

		// Context Assembler — budget-aware tiered context retrieval
		g.POST("/assemble", h.AssembleContext)

		// Reindex — rebuild the full search index from the filesystem.
		g.POST("/reindex", h.Reindex)
	}
}

// ── handler methods ───────────────────────────────────────────────────────────

// ListNodes handles GET /api/optimal/nodes
// Returns all numbered node folders sorted by numeric prefix.
func (h *OptimalHandler) ListNodes(c *gin.Context) {
	nodes, err := optimal.ListNodes(h.nodesRoot)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: list nodes failed",
			"nodes_root", h.nodesRoot,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "list optimal nodes", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// GetNode handles GET /api/optimal/nodes/:slug
// Returns context.md + signal.md content and signal count for a single node.
func (h *OptimalHandler) GetNode(c *gin.Context) {
	slug := c.Param("slug")
	node, err := optimal.GetNode(h.nodesRoot, slug)
	if err != nil {
		if isNotExist(err) {
			utils.RespondNotFound(c, slog.Default(), "node")
			return
		}
		slog.ErrorContext(c.Request.Context(), "optimal: get node failed",
			"slug", slug,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get optimal node", err)
		return
	}
	c.JSON(http.StatusOK, node)
}

// GetNodeSignals handles GET /api/optimal/nodes/:slug/signals
// Returns all dated signal files for the node, newest-first.
func (h *OptimalHandler) GetNodeSignals(c *gin.Context) {
	slug := c.Param("slug")
	signals, err := optimal.GetNodeSignals(h.nodesRoot, slug)
	if err != nil {
		if isNotExist(err) {
			utils.RespondNotFound(c, slog.Default(), "node")
			return
		}
		slog.ErrorContext(c.Request.Context(), "optimal: get node signals failed",
			"slug", slug,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get optimal node signals", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"signals": signals,
		"count":   len(signals),
	})
}

// GetTodayRhythm handles GET /api/optimal/rhythm/today
// Returns the daily working file for today's date from rhythm/daily/.
func (h *OptimalHandler) GetTodayRhythm(c *gin.Context) {
	day, err := optimal.GetTodayRhythm(h.osRoot)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get today rhythm failed",
			"os_root", h.osRoot,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get today rhythm", err)
		return
	}
	c.JSON(http.StatusOK, day)
}

// searchRequest is the request body for POST /api/optimal/search.
type searchRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit"`
}

// Search handles POST /api/optimal/search
// Delegates to the Elixir engine via mix optimal.search.
func (h *OptimalHandler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	results, err := h.engine.Search(c.Request.Context(), req.Query, req.Limit)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: search failed",
			"query", req.Query,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "optimal search", err)
		return
	}
	if results == nil {
		results = []optimal.SearchResult{}
	}
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// composeRequest is the body for POST /api/optimal/compose.
type composeRequest struct {
	Content  string `json:"content"  binding:"required"`
	Receiver string `json:"receiver"` // person name from topology
	Genre    string `json:"genre"`    // explicit genre override
}

// Compose handles POST /api/optimal/compose
// Reformats content for a specific receiver based on their preferred genre from
// topology. If genre is provided explicitly it takes precedence over the topology
// lookup. When neither receiver nor genre is provided, content is returned unchanged.
func (h *OptimalHandler) Compose(c *gin.Context) {
	var req composeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	// Explicit genre: bypass topology lookup.
	if req.Genre != "" {
		composed := optimal.Compose(req.Content, req.Genre)
		c.JSON(http.StatusOK, gin.H{
			"content":   composed,
			"genre":     req.Genre,
			"receiver":  req.Receiver,
			"converted": composed != req.Content,
		})
		return
	}

	// Receiver-aware: resolve genre from topology.
	var topo *optimal.TopologyConfig
	if h.engine != nil && h.engine.Topology != nil {
		topo = h.engine.Topology
	}

	result := optimal.ComposeForReceiver(req.Content, req.Receiver, topo)
	c.JSON(http.StatusOK, result)
}

// ingestRequest is the request body for POST /api/optimal/ingest.
type ingestRequest struct {
	Text  string `json:"text"  binding:"required"`
	Genre string `json:"genre"`
}

// Ingest handles POST /api/optimal/ingest
// Fires mix optimal.ingest asynchronously and returns 202 Accepted immediately.
func (h *OptimalHandler) Ingest(c *gin.Context) {
	var req ingestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	h.engine.IngestAsync(req.Text, req.Genre)
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

// GetNodeFileTree handles GET /api/optimal/nodes/:slug/files
// Returns the recursive file tree for a node.
func (h *OptimalHandler) GetNodeFileTree(c *gin.Context) {
	slug := c.Param("slug")
	tree, err := optimal.GetNodeFileTree(h.nodesRoot, slug)
	if err != nil {
		if isNotExist(err) {
			utils.RespondNotFound(c, slog.Default(), "node")
			return
		}
		utils.RespondInternalError(c, slog.Default(), "get node file tree", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": tree, "slug": slug})
}

// GetNodeFile handles GET /api/optimal/nodes/:slug/file/*filepath
// Returns the raw content of a single file inside a node.
func (h *OptimalHandler) GetNodeFile(c *gin.Context) {
	slug := c.Param("slug")
	filePath := c.Param("filepath")
	if len(filePath) > 0 && filePath[0] == '/' {
		filePath = filePath[1:]
	}
	content, err := optimal.GetNodeFile(h.nodesRoot, slug, filePath)
	if err != nil {
		if isNotExist(err) {
			utils.RespondNotFound(c, slog.Default(), "file")
			return
		}
		utils.RespondInternalError(c, slog.Default(), "get node file", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": content, "path": filePath, "slug": slug})
}

// saveFileRequest is the body for PUT /api/optimal/nodes/:slug/file/*filepath.
type saveFileRequest struct {
	Content string `json:"content" binding:"required"`
}

// SaveNodeFile handles PUT /api/optimal/nodes/:slug/file/*filepath
// Writes content back to the filesystem. Creates parent dirs if needed.
func (h *OptimalHandler) SaveNodeFile(c *gin.Context) {
	slug := c.Param("slug")
	filePath := c.Param("filepath")
	if len(filePath) > 0 && filePath[0] == '/' {
		filePath = filePath[1:]
	}

	var req saveFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := optimal.SaveNodeFile(h.nodesRoot, slug, filePath, req.Content); err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: save file failed",
			"slug", slug, "path", filePath, "error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "save node file", err)
		return
	}

	// Fire async ingest so the engine indexes the updated file
	h.engine.IngestAsync(req.Content, "note")

	c.JSON(http.StatusOK, gin.H{"status": "saved", "path": filePath, "slug": slug})
}

// createFileRequest is the body for POST /api/optimal/nodes/:slug/file.
type createFileRequest struct {
	Path    string `json:"path"    binding:"required"`
	Content string `json:"content"`
}

// CreateNodeFile handles POST /api/optimal/nodes/:slug/file
// Creates a new file inside a node. Fails if file already exists.
func (h *OptimalHandler) CreateNodeFile(c *gin.Context) {
	slug := c.Param("slug")

	var req createFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := optimal.CreateNodeFile(h.nodesRoot, slug, req.Path, req.Content); err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: create file failed",
			"slug", slug, "path", req.Path, "error", err,
		)
		if isNotExist(err) {
			utils.RespondNotFound(c, slog.Default(), "node")
			return
		}
		utils.RespondInternalError(c, slog.Default(), "create node file", err)
		return
	}

	// Ingest the new file
	if req.Content != "" {
		h.engine.IngestAsync(req.Content, "note")
	}

	c.JSON(http.StatusCreated, gin.H{"status": "created", "path": req.Path, "slug": slug})
}

// ── graph handlers ────────────────────────────────────────────────────────────

// GetGraph handles GET /api/optimal/graph
// Returns the full entity+edge graph from the SQLite index.
func (h *OptimalHandler) GetGraph(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	data, err := optimal.GetGraph(h.dbPath)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get graph failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "get optimal graph", err)
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetGraphHubs handles GET /api/optimal/graph/hubs?limit=20
// Returns the top N most-connected entities.
func (h *OptimalHandler) GetGraphHubs(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			limit = n
		}
	}
	hubs, err := optimal.GetGraphHubs(h.dbPath, limit)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get graph hubs failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "get optimal graph hubs", err)
		return
	}
	if hubs == nil {
		hubs = []optimal.GraphEntity{}
	}
	c.JSON(http.StatusOK, gin.H{"hubs": hubs, "count": len(hubs)})
}

// GraphSearch handles GET /api/optimal/graph/search?q=query&limit=10
// Runs FTS5 search directly on SQLite — faster than the Elixir subprocess.
func (h *OptimalHandler) GraphSearch(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	q := c.Query("q")
	if q == "" {
		utils.RespondInvalidRequest(c, slog.Default(), errors.New("missing q parameter"))
		return
	}
	limit := 10
	if v := c.Query("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			limit = n
		}
	}
	results, err := optimal.SearchContexts(h.dbPath, q, limit)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: graph search failed",
			"query", q, "error", err)
		utils.RespondInternalError(c, slog.Default(), "optimal graph search", err)
		return
	}
	if results == nil {
		results = []optimal.SearchResult{}
	}
	c.JSON(http.StatusOK, gin.H{"results": results, "count": len(results)})
}

// ── aggregated data handlers ──────────────────────────────────────────────────

// GetDashboard handles GET /api/optimal/dashboard
// Returns aggregated counts and command-center content for the UI home screen.
func (h *OptimalHandler) GetDashboard(c *gin.Context) {
	data, err := optimal.GetDashboard(h.nodesRoot, h.osRoot, h.dbPath)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get dashboard failed",
			"nodes_root", h.nodesRoot,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get optimal dashboard", err)
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTeam handles GET /api/optimal/team
// Returns the team roster parsed from nodes/10-team/context.md.
func (h *OptimalHandler) GetTeam(c *gin.Context) {
	members, err := optimal.GetTeamRoster(h.nodesRoot)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get team failed",
			"nodes_root", h.nodesRoot,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get optimal team", err)
		return
	}
	if members == nil {
		members = []optimal.TeamMember{}
	}
	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"count":   len(members),
	})
}

// GetRevenue handles GET /api/optimal/revenue
// Returns revenue streams and total MRR parsed from nodes/11-money-revenue/context.md.
func (h *OptimalHandler) GetRevenue(c *gin.Context) {
	data, err := optimal.GetRevenueStreams(h.nodesRoot)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get revenue failed",
			"nodes_root", h.nodesRoot,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get optimal revenue", err)
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetProjects handles GET /api/optimal/projects
// Returns a unified list of projects scanned from all nodes/*/projects/ dirs.
func (h *OptimalHandler) GetProjects(c *gin.Context) {
	projects, err := optimal.GetAllProjects(h.nodesRoot)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get projects failed",
			"nodes_root", h.nodesRoot,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get optimal projects", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"count":    len(projects),
	})
}

// GetWeeklyRhythm handles GET /api/optimal/rhythm/weekly
// Returns the most recent weekly plan from rhythm/weekly/week-of-*.md.
func (h *OptimalHandler) GetWeeklyRhythm(c *gin.Context) {
	plan, err := optimal.GetWeeklyRhythm(h.osRoot)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get weekly rhythm failed",
			"os_root", h.osRoot,
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get optimal weekly rhythm", err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

// GetHealth handles GET /api/optimal/health
// Returns a diagnostic health check across the filesystem and SQLite index.
func (h *OptimalHandler) GetHealth(c *gin.Context) {
	status, err := optimal.GetHealthCheck(h.nodesRoot, h.osRoot, h.dbPath)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: health check failed",
			"error", err,
		)
		utils.RespondInternalError(c, slog.Default(), "get optimal health", err)
		return
	}
	httpStatus := http.StatusOK
	if status.Status == "error" {
		httpStatus = http.StatusServiceUnavailable
	}
	c.JSON(httpStatus, status)
}

// Diagnose handles GET /api/optimal/diagnose
// Runs the full 10-check diagnostic suite against filesystem and SQLite index.
func (h *OptimalHandler) Diagnose(c *gin.Context) {
	report, err := optimal.Diagnose(h.nodesRoot, h.dbPath)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: diagnose failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "optimal diagnose", err)
		return
	}
	httpStatus := http.StatusOK
	if report.Status == "error" {
		httpStatus = http.StatusServiceUnavailable
	}
	c.JSON(httpStatus, report)
}

// Reweave handles POST /api/optimal/reweave
// Finds stale contexts related to a topic that need updating.
func (h *OptimalHandler) Reweave(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}

	var req struct {
		Topic   string `json:"topic" binding:"required"`
		MaxDays int    `json:"max_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	results, err := optimal.Reweave(h.dbPath, req.Topic, req.MaxDays)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: reweave failed",
			"topic", req.Topic, "error", err)
		utils.RespondInternalError(c, slog.Default(), "optimal reweave", err)
		return
	}
	if results == nil {
		results = []optimal.StaleContext{}
	}
	c.JSON(http.StatusOK, gin.H{
		"topic":   req.Topic,
		"results": results,
		"count":   len(results),
	})
}

// VerifyL0 handles POST /api/optimal/verify
// Runs L0 fidelity cold-read test on a random sample of contexts.
func (h *OptimalHandler) VerifyL0(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}

	var req struct {
		SampleSize int `json:"sample_size"`
	}
	// Don't require body — use defaults if not provided.
	_ = c.ShouldBindJSON(&req)
	if req.SampleSize <= 0 {
		req.SampleSize = 10
	}

	result, err := optimal.Verify(h.dbPath, req.SampleSize)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: verify failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "optimal verify", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── L8 learning & feedback handlers ──────────────────────────────────────────

// rememberRequest is the body for POST /api/optimal/remember.
type rememberRequest struct {
	Content  string `json:"content"  binding:"required"`
	Category string `json:"category"` // process, people, tool, decision, pattern, friction
	Source   string `json:"source"`   // explicit, contextual, mining
}

// Remember handles POST /api/optimal/remember
// Stores an observation for the learning feedback loop.
func (h *OptimalHandler) Remember(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	var req rememberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := optimal.Remember(h.dbPath, req.Content, req.Category, req.Source); err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: remember failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "optimal remember", err)
		return
	}

	// Check for escalations after storing — auto-trigger awareness.
	escalations, _ := optimal.GetEscalations(h.dbPath)

	c.JSON(http.StatusCreated, gin.H{
		"status":      "stored",
		"category":    req.Category,
		"escalations": escalations,
	})
}

// GetObservations handles GET /api/optimal/observations?category=...
// Returns stored observations, optionally filtered by category.
func (h *OptimalHandler) GetObservations(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	category := c.Query("category")
	obs, err := optimal.ListObservations(h.dbPath, category)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: list observations failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "list observations", err)
		return
	}
	if obs == nil {
		obs = []optimal.Observation{}
	}
	c.JSON(http.StatusOK, gin.H{
		"observations": obs,
		"count":        len(obs),
	})
}

// GetEscalations handles GET /api/optimal/escalations
// Returns categories with 3+ observations ready for Rethink synthesis.
func (h *OptimalHandler) GetEscalations(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	escalations, err := optimal.GetEscalations(h.dbPath)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: get escalations failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "get escalations", err)
		return
	}
	if escalations == nil {
		escalations = make(map[string]int)
	}
	c.JSON(http.StatusOK, gin.H{
		"escalations": escalations,
		"count":       len(escalations),
	})
}

// rethinkRequest is the body for POST /api/optimal/rethink.
type rethinkRequest struct {
	Topic string `json:"topic" binding:"required"`
}

// Rethink handles POST /api/optimal/rethink
// Synthesizes accumulated observations and search hits for a topic.
func (h *OptimalHandler) Rethink(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	var req rethinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	report, err := optimal.Rethink(h.dbPath, h.nodesRoot, req.Topic)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: rethink failed",
			"topic", req.Topic, "error", err)
		utils.RespondInternalError(c, slog.Default(), "optimal rethink", err)
		return
	}
	c.JSON(http.StatusOK, report)
}

// assembleRequest is the request body for POST /api/optimal/assemble.
type assembleRequest struct {
	Query    string                   `json:"query" binding:"required"`
	Receiver *optimal.ReceiverProfile `json:"receiver"`
	Budget   int                      `json:"budget"` // total token budget override
}

// AssembleContext handles POST /api/optimal/assemble
// Returns a budget-aware, tiered context package: L0 (always-loaded), L1
// (query-relevant), and L2 (deep detail for technical/agent receivers).
// A full retrieval trace is included so callers can inspect why each source
// was selected.
func (h *OptimalHandler) AssembleContext(c *gin.Context) {
	var req assembleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	receiver := req.Receiver
	if receiver == nil {
		receiver = &optimal.ReceiverProfile{
			Role:        "agent",
			TokenBudget: 150000,
		}
	}
	if req.Budget > 0 {
		receiver.TokenBudget = req.Budget
	}

	result, err := h.engine.AssembleContext(c.Request.Context(), req.Query, receiver)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: assemble context failed",
			"query", req.Query, "error", err)
		utils.RespondInternalError(c, slog.Default(), "assemble context", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Reindex handles POST /api/optimal/reindex
// Rebuilds the entire SQLite search index by walking the nodes/ directory tree
// and upserting every .md file. Files with S/N < 0.3 are rejected. Every
// mutation is written to the decision_log table.
// Returns 200 with a ReindexResult JSON body. Errors in individual files are
// collected into the result rather than aborting the entire run.
func (h *OptimalHandler) Reindex(c *gin.Context) {
	if h.nodesRoot == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "nodesRoot not configured (OPTIMAL_OS_ROOT not set)",
		})
		return
	}

	result, err := optimal.Reindex(h.nodesRoot, h.dbPath)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: reindex failed",
			"nodes_root", h.nodesRoot, "error", err)
		utils.RespondInternalError(c, slog.Default(), "optimal reindex", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// hasDB returns true when the handler has a configured database path, otherwise
// it writes a 503 response and returns false. Call at the top of graph handlers.
func (h *OptimalHandler) hasDB(c *gin.Context) bool {
	if h.dbPath == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "SQLite index not configured (OPTIMAL_OS_ROOT not set)",
		})
		return false
	}
	return true
}

// parsePositiveInt parses s as a positive integer.
func parsePositiveInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("not a positive integer: %q", s)
	}
	return n, nil
}

// isNotExist reports whether err (or any error in its chain) wraps os.ErrNotExist.
func isNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
