package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
)

// ontologyBosService abstracts bos CLI operations needed by OntologyHandler.
// *services.BosOntologyService satisfies this interface automatically.
type ontologyBosService interface {
	ListQueries(ctx context.Context) ([]string, error)
	GetConstructQuery(ctx context.Context, table string) (string, error)
	ExecuteConstruct(ctx context.Context, table string) (string, error)
	ExecuteAll(ctx context.Context, format string) (string, error)
	GenerateQueries(ctx context.Context, outputDir string) (int, error)
	ExecuteSelect(ctx context.Context, query string) (map[string]interface{}, error)
}

// OntologyHandler handles RDF ontology endpoints via bos CLI.
type OntologyHandler struct {
	bosService ontologyBosService
}

// NewOntologyHandler constructs an OntologyHandler backed by a real BosOntologyService.
func NewOntologyHandler(bosService *services.BosOntologyService) *OntologyHandler {
	return &OntologyHandler{bosService: bosService}
}

// newOntologyHandlerFromInterface constructs an OntologyHandler from any implementation
// of ontologyBosService. Used in unit tests to inject stubs.
func newOntologyHandlerFromInterface(svc ontologyBosService) *OntologyHandler {
	return &OntologyHandler{bosService: svc}
}

// RegisterOntologyRoutes wires /api/ontology routes.
// All RDF data flows through bos — no domain logic in Go.
func RegisterOntologyRoutes(api *gin.RouterGroup, h *OntologyHandler, auth gin.HandlerFunc) {
	ontology := api.Group("/ontology")
	ontology.Use(auth)
	{
		// List available CONSTRUCT queries
		ontology.GET("/queries", h.ListQueries)

		// Get CONSTRUCT query text for a specific table
		ontology.GET("/queries/:table", h.GetQuery)

		// Execute CONSTRUCT and return RDF triples for a table
		ontology.GET("/data/:table", h.GetData)

		// Execute CONSTRUCT for all mapped tables
		ontology.GET("/export", h.ExportAll)

		// Generate .rq files from mappings
		ontology.POST("/generate", h.Generate)

		// Execute a raw SPARQL SELECT query and return JSON results
		ontology.POST("/query", h.QuerySPARQL)
	}
}

// ListQueries returns all available CONSTRUCT query files.
func (h *OntologyHandler) ListQueries(c *gin.Context) {
	queries, err := h.bosService.ListQueries(c.Request.Context())
	if err != nil {
		slog.Error("Failed to list queries", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list queries"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"queries": queries, "count": len(queries)})
}

// GetQuery returns the CONSTRUCT query text for a specific table.
func (h *OntologyHandler) GetQuery(c *gin.Context) {
	table := c.Param("table")
	if table == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table parameter required"})
		return
	}

	query, err := h.bosService.GetConstructQuery(c.Request.Context(), table)
	if err != nil {
		slog.Error("Query not found", "table", table, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Query not found for table: %s", table)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"table": table,
		"query": query,
	})
}

// GetData executes CONSTRUCT for a specific table and returns RDF triples.
// All data flows through bos ontology execute — PostgreSQL → RDF → CONSTRUCT.
func (h *OntologyHandler) GetData(c *gin.Context) {
	table := c.Param("table")
	if table == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table parameter required"})
		return
	}

	rdf, err := h.bosService.ExecuteConstruct(c.Request.Context(), table)
	if err != nil {
		slog.Error("Failed to execute CONSTRUCT", "table", table, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Execution failed for table %s: %v", table, err)})
		return
	}

	// Return N-Triples format
	c.Header("Content-Type", "application/n-triples")
	c.String(http.StatusOK, rdf)
}

// ExportAll executes CONSTRUCT for all mapped tables and returns combined RDF.
func (h *OntologyHandler) ExportAll(c *gin.Context) {
	format := c.DefaultQuery("format", "ttl")

	rdf, err := h.bosService.ExecuteAll(c.Request.Context(), format)
	if err != nil {
		slog.Error("Failed to export all", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Export failed: %v", err)})
		return
	}

	switch format {
	case "ttl":
		c.Header("Content-Type", "text/turtle")
	case "json":
		c.Header("Content-Type", "application/json")
	default:
		c.Header("Content-Type", "application/n-triples")
	}

	c.String(http.StatusOK, rdf)
}

// QuerySPARQL executes a raw SPARQL SELECT query via the bos CLI.
// POST /api/v1/ontology/query
// Body: {"query": "SELECT ..."}
// Returns JSON result from bos, or a structured error response.
func (h *OntologyHandler) QuerySPARQL(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid query request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "field 'query' is required"})
		return
	}

	result, err := h.bosService.ExecuteSelect(c.Request.Context(), req.Query)
	if err != nil {
		slog.Error("SPARQL query execution failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("query execution failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Generate triggers bos ontology construct to regenerate .rq files.
func (h *OntologyHandler) Generate(c *gin.Context) {
	outputDir := c.DefaultQuery("output", "desktop/backend-go/bos/queries")

	count, err := h.bosService.GenerateQueries(c.Request.Context(), outputDir)
	if err != nil {
		slog.Error("Failed to generate queries", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Generation failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "generated",
		"queries":    count,
		"output_dir": outputDir,
	})
}
