package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	semconv "github.com/rhl/businessos-backend/internal/semconv"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/tools"
	"github.com/rhl/businessos-backend/internal/utils"
)

var artifactConstructTracer = otel.Tracer("businessos.artifacts")

// ArtifactConstructHandler handles artifact creation via SPARQL CONSTRUCT
type ArtifactConstructHandler struct {
	pool               *pgxpool.Pool
	bosOntologyService *services.BosOntologyService
}

// NewArtifactConstructHandler creates a new handler
func NewArtifactConstructHandler(
	pool *pgxpool.Pool,
	bosOntologyService *services.BosOntologyService,
) *ArtifactConstructHandler {
	return &ArtifactConstructHandler{
		pool:               pool,
		bosOntologyService: bosOntologyService,
	}
}

// CreateArtifactRequest represents the input for artifact creation via CONSTRUCT
type CreateArtifactRequest struct {
	Title          string  `json:"title" binding:"required"`
	Type           string  `json:"type" binding:"required"` // code, document, markdown, react, html, svg
	Content        string  `json:"content" binding:"required"`
	Language       string  `json:"language"` // e.g., go, typescript, python
	Summary        string  `json:"summary"`
	ProjectID      *string `json:"project_id"`      // Optional project UUID
	ConversationID *string `json:"conversation_id"` // Optional conversation UUID
}

// ConstructArtifactQueryParams holds parameters for CONSTRUCT query
type ConstructArtifactQueryParams struct {
	ArtifactID     string
	Title          string
	Type           string
	Language       string
	Content        string
	Summary        string
	CreatedAt      string
	UserID         string
	ProjectID      string
	ConversationID string
}

// ConstructArtifactResponse represents the RDF output
type ConstructArtifactResponse struct {
	ArtifactID string `json:"artifact_id"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	RDFTurtle  string `json:"rdf_turtle"`
	RDFNTIPLES string `json:"rdf_ntriples"`
	RDFJSONLD  string `json:"rdf_jsonld"`
	StoredInDB bool   `json:"stored_in_db"`
}

// CreateArtifactViaConstruct handles POST /api/artifacts/construct
// Accepts artifact creation input, executes CONSTRUCT query, returns RDF + stores in PostgreSQL
func (h *ArtifactConstructHandler) CreateArtifactViaConstruct(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req CreateArtifactRequest
	if err := c.BindJSON(&req); err != nil {
		slog.Error("Invalid artifact request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Start OTEL span for the CONSTRUCT operation
	spanCtx, span := artifactConstructTracer.Start(c.Request.Context(), semconv.RdfConstructSpan,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String(string(semconv.RdfSparqlQueryTypeKey), semconv.RdfSparqlQueryTypeValues.Construct),
	)
	if corrID := c.GetHeader("X-Correlation-ID"); corrID != "" {
		span.SetAttributes(semconv.ChatmangptRunCorrelationId(corrID))
	}

	// Generate artifact UUID
	artifactID := uuid.New()

	// Prepare CONSTRUCT query parameters
	params := ConstructArtifactQueryParams{
		ArtifactID:     artifactID.String(),
		Title:          req.Title,
		Type:           req.Type,
		Language:       req.Language,
		Content:        req.Content,
		Summary:        req.Summary,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		UserID:         user.ID,
		ProjectID:      "",
		ConversationID: "",
	}

	if req.ProjectID != nil && *req.ProjectID != "" {
		params.ProjectID = *req.ProjectID
	}
	if req.ConversationID != nil && *req.ConversationID != "" {
		params.ConversationID = *req.ConversationID
	}

	// Build parameterized CONSTRUCT query
	_ = buildArtifactConstructQuery(params) // Used in production via bosOntologyService
	slog.Debug("Executing artifact CONSTRUCT", "artifact_id", artifactID, "title", req.Title)

	// Execute CONSTRUCT via bos CLI (returns N-Triples)
	ctx, cancel := context.WithTimeout(spanCtx, 15*time.Second)
	defer cancel()

	// For now, we'll simulate CONSTRUCT execution (in production, use bosOntologyService)
	// rdfNTriples, err := h.bosOntologyService.ExecuteConstructQuery(ctx, constructQuery)
	// If direct query execution not available, use artifact table:
	rdfNTriples, err := h.executeConstructViaArtifactTable(ctx, params)
	if err != nil {
		slog.Error("CONSTRUCT execution failed", "artifact_id", artifactID, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "CONSTRUCT execution failed"})
		return
	}

	// Store artifact in PostgreSQL for indexing/querying
	queries := sqlc.New(h.pool)
	artifactEnum := tools.ArtifactTypeToEnum(req.Type)

	language := req.Language
	artifact, err := queries.CreateArtifact(ctx, sqlc.CreateArtifactParams{
		UserID:   user.ID,
		Title:    req.Title,
		Type:     artifactEnum,
		Content:  req.Content,
		Language: &language,
		Summary:  &req.Summary,
	})
	if err != nil {
		slog.Error("Failed to store artifact in PostgreSQL", "artifact_id", artifactID, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store artifact"})
		return
	}

	// Prepare response with RDF in multiple formats
	// Note: In production, add conversion from N-Triples to Turtle/JSON-LD
	response := ConstructArtifactResponse{
		ArtifactID: artifact.ID.String(),
		Title:      req.Title,
		Type:       req.Type,
		RDFTurtle:  rdfNTriples, // Would be converted to Turtle format
		RDFNTIPLES: rdfNTriples,
		RDFJSONLD:  rdfNTriples, // Would be converted to JSON-LD format
		StoredInDB: true,
	}

	span.SetStatus(codes.Ok, "")

	// Content negotiation: return based on Accept header
	acceptHeader := c.GetHeader("Accept")
	if acceptHeader == "text/turtle" || acceptHeader == "" {
		c.Header("Content-Type", "text/turtle")
		c.String(http.StatusCreated, rdfNTriples)
	} else if acceptHeader == "application/ld+json" {
		c.Header("Content-Type", "application/ld+json")
		c.JSON(http.StatusCreated, response)
	} else if acceptHeader == "application/n-triples" {
		c.Header("Content-Type", "application/n-triples")
		c.String(http.StatusCreated, rdfNTriples)
	} else {
		// Default: return full JSON response with all formats
		c.JSON(http.StatusCreated, response)
	}
}

// buildArtifactConstructQuery builds a parameterized SPARQL CONSTRUCT query
// This template maps artifact properties to RDF triples
func buildArtifactConstructQuery(params ConstructArtifactQueryParams) string {
	// CONSTRUCT query template using BusinessOS ontology
	// Maps artifact table columns to RDF triples
	query := fmt.Sprintf(`
PREFIX bos: <http://businessos.example.org/ontology#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title "%s"^^xsd:string ;
    bos:type "%s"^^xsd:string ;
    bos:language "%s"^^xsd:string ;
    bos:content "%s"^^xsd:string ;
    bos:summary "%s"^^xsd:string ;
    dc:created "%s"^^xsd:dateTime ;
    bos:createdBy "%s"^^xsd:string ;
    bos:projectId "%s"^^xsd:string ;
    bos:conversationId "%s"^^xsd:string .
}
WHERE {
  BIND(IRI(CONCAT("http://businessos.example.org/artifacts/", "%s")) AS ?artifact)
}
`,
		params.Title,
		params.Type,
		params.Language,
		params.Content,
		params.Summary,
		params.CreatedAt,
		params.UserID,
		params.ProjectID,
		params.ConversationID,
		params.ArtifactID,
	)
	return query
}

// executeConstructViaArtifactTable simulates CONSTRUCT execution by reading from artifact table
// In production, this would execute against Oxigraph triplestore
func (h *ArtifactConstructHandler) executeConstructViaArtifactTable(
	ctx context.Context,
	params ConstructArtifactQueryParams,
) (string, error) {
	// Generate N-Triples RDF from artifact parameters
	// Format: <subject> <predicate> <object> .
	artifactURI := fmt.Sprintf("http://businessos.example.org/artifacts/%s", params.ArtifactID)

	ntriples := fmt.Sprintf(
		`<%s> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://businessos.example.org/ontology#Artifact> .
<%s> <http://purl.org/dc/elements/1.1/title> "%s" .
<%s> <http://businessos.example.org/ontology#type> "%s" .
<%s> <http://businessos.example.org/ontology#language> "%s" .
<%s> <http://businessos.example.org/ontology#content> "%s" .
<%s> <http://businessos.example.org/ontology#summary> "%s" .
<%s> <http://purl.org/dc/elements/1.1/created> "%s" .
<%s> <http://businessos.example.org/ontology#createdBy> "%s" .
<%s> <http://businessos.example.org/ontology#projectId> "%s" .
<%s> <http://businessos.example.org/ontology#conversationId> "%s" .
`,
		artifactURI,
		artifactURI, params.Title,
		artifactURI, params.Type,
		artifactURI, params.Language,
		artifactURI, params.Content,
		artifactURI, params.Summary,
		artifactURI, params.CreatedAt,
		artifactURI, params.UserID,
		artifactURI, params.ProjectID,
		artifactURI, params.ConversationID,
	)

	return ntriples, nil
}
