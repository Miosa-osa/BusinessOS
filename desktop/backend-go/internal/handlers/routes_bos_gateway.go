package handlers

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/integrations/linkedin"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// registerBOSGatewayRoutes wires POST /api/bos/discover, /conformance, /statistics, /status.
// Applies StaticBearerAuth middleware reading BUSINESSOS_API_TOKEN.
func (h *Handlers) registerBOSGatewayRoutes(api *gin.RouterGroup) {
	token := os.Getenv("BUSINESSOS_API_TOKEN")
	bosAuth := middleware.StaticBearerAuth(token)
	bosHandler := NewBOSGatewayHandler(h.pool, slog.Default())
	bosGroup := api.Group("/bos")
	bosGroup.Use(bosAuth)
	bosGroup.POST("/discover", bosHandler.Discover)
	bosGroup.POST("/conformance", bosHandler.CheckConformance)
	bosGroup.POST("/statistics", bosHandler.GetStatistics)
	bosGroup.GET("/status", bosHandler.GetStatus)

	// Schema management endpoints (BOS CLI ↔ BusinessOS round-trip).
	schema := bosGroup.Group("/schema")
	schema.POST("/import", bosHandler.SchemaImport)
	schema.GET("/export/:schema_id", bosHandler.SchemaExport)
	schema.POST("/validate/:schema_id", bosHandler.SchemaValidate)
	schema.POST("/update", bosHandler.SchemaUpdate)
}

// registerLinkedInRoutes wires /api/linkedin routes for ICP scoring and outreach.
// Requires DATABASE_URL; if not set, logs warn and skips (no error).
func (h *Handlers) registerLinkedInRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Default().Warn("registerLinkedInRoutes: DATABASE_URL not set — skipping LinkedIn routes")
		return
	}
	stdDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		slog.Default().Error("registerLinkedInRoutes: failed to open stdDB", "error", err)
		return
	}
	logger := slog.Default()
	repo := linkedin.NewRepository(logger, stdDB)
	scorer := linkedin.NewICPScorer(logger)
	importer := linkedin.NewCSVImporter(logger, repo)
	outreachQueue := linkedin.NewOutreachQueueManager(logger, nil, repo)
	linkedInHandler := NewLinkedInHandler(logger, repo, scorer, importer, outreachQueue)

	li := api.Group("/linkedin")
	li.Use(auth)
	{
		li.POST("/import", linkedInHandler.ImportCSV)
		li.GET("/contacts", linkedInHandler.GetContacts)
		li.POST("/icp-score", linkedInHandler.ICPScoreContacts)
		li.POST("/outreach/enroll", linkedInHandler.EnrollOutreach)
	}
}
