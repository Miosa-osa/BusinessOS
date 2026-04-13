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
	eng := optimal.NewEngineConfig(enginePath)
	eng.DBPath = dbPath
	eng.NodesRoot = nodesRoot
	eng.OSRoot = osRoot
	return &OptimalHandler{
		nodesRoot: nodesRoot,
		osRoot:    osRoot,
		dbPath:    dbPath,
		engine:    eng,
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
		g.GET("/nodes/:slug/files", h.GetNodeFileTree)
		g.GET("/nodes/:slug/file/*filepath", h.GetNodeFile)
		g.PUT("/nodes/:slug/file/*filepath", h.SaveNodeFile)
		g.POST("/nodes/:slug/file", h.CreateNodeFile)

		// Graph endpoints — read directly from SQLite for low latency.
		g.GET("/graph", h.GetGraph)
		g.GET("/graph/hubs", h.GetGraphHubs)
		g.GET("/graph/search", h.GraphSearch)
		g.GET("/graph/triangles", h.GetGraphTriangles)
		g.GET("/graph/clusters", h.GetGraphClusters)

		// Signal Theory classifier.
		g.POST("/classify", h.ClassifyText)

		// Scenario simulation.
		g.POST("/simulate", h.Simulate)

		// Aggregated data endpoints.
		g.GET("/dashboard", h.GetDashboard)
		g.GET("/team", h.GetTeam)
		g.GET("/revenue", h.GetRevenue)
		g.GET("/projects", h.GetProjects)
		g.GET("/rhythm/weekly", h.GetWeeklyRhythm)
		g.GET("/health", h.GetHealth)
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

// ── phase-3 handlers ──────────────────────────────────────────────────────────

// classifyRequest is the request body for POST /api/optimal/classify.
type classifyRequest struct {
	Text string `json:"text" binding:"required"`
}

// ClassifyText handles POST /api/optimal/classify.
// Runs the Signal Theory classifier and returns S=(M,G,T,F,W) + S/N ratio.
func (h *OptimalHandler) ClassifyText(c *gin.Context) {
	var req classifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}
	result := optimal.Classify(req.Text)
	c.JSON(http.StatusOK, result)
}

// GetGraphTriangles handles GET /api/optimal/graph/triangles?limit=20.
// Returns open triangles (synthesis opportunities) from the knowledge graph.
func (h *OptimalHandler) GetGraphTriangles(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			limit = n
		}
	}
	triangles, err := optimal.FindTriangles(h.dbPath, limit)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: find triangles failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "find graph triangles", err)
		return
	}
	if triangles == nil {
		triangles = []optimal.Triangle{}
	}
	c.JSON(http.StatusOK, gin.H{"triangles": triangles, "count": len(triangles)})
}

// GetGraphClusters handles GET /api/optimal/graph/clusters.
// Returns connected components of the knowledge graph via BFS.
func (h *OptimalHandler) GetGraphClusters(c *gin.Context) {
	if !h.hasDB(c) {
		return
	}
	clusters, err := optimal.FindClusters(h.dbPath)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: find clusters failed", "error", err)
		utils.RespondInternalError(c, slog.Default(), "find graph clusters", err)
		return
	}
	if clusters == nil {
		clusters = []optimal.Cluster{}
	}
	c.JSON(http.StatusOK, gin.H{"clusters": clusters, "count": len(clusters)})
}

// simulateRequest is the request body for POST /api/optimal/simulate.
type simulateRequest struct {
	Mutation    string `json:"mutation"    binding:"required"`
	Simulations int    `json:"simulations"` // optional Monte Carlo count; 0 → default 1000
	Budget      int    `json:"budget"`      // optional MCTS budget; 0 → default 32
}

// simulateResponse is the full response for POST /api/optimal/simulate.
type simulateResponse struct {
	Simulation *optimal.SimulationResult `json:"simulation"`
	MonteCarlo *optimal.MonteCarloResult `json:"monte_carlo"`
	MCTS       *optimal.MCTSResult       `json:"mcts"`
}

// Simulate handles POST /api/optimal/simulate.
// Runs Simulate → Sample (Monte Carlo) → PlanResponse (MCTS) in sequence and
// returns all three result sets in a single response.
func (h *OptimalHandler) Simulate(c *gin.Context) {
	var req simulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}
	if !h.hasDB(c) {
		return
	}

	simResult, err := optimal.Simulate(h.dbPath, req.Mutation)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "optimal: simulate failed",
			"mutation", req.Mutation, "error", err)
		utils.RespondInternalError(c, slog.Default(), "optimal simulate", err)
		return
	}

	mcResult := optimal.Sample(simResult, req.Simulations)
	mctsResult := optimal.PlanResponse(simResult, req.Budget)

	c.JSON(http.StatusOK, simulateResponse{
		Simulation: simResult,
		MonteCarlo: mcResult,
		MCTS:       mctsResult,
	})
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
