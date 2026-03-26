package handlers

// Board Intelligence HTTP handlers
//
// GET  /api/board/intelligence  → latest bos:BoardIntelligence node from Oxigraph
// GET  /api/board/briefing      → latest briefing text
// GET  /api/board/escalations   → departments with structuralIssueCount > 0 or operationalIssueCount > 0
// POST /api/board/conway/trigger → trigger one-shot Conway check (call OSA HTTP endpoint)
//
// WvdA: all Oxigraph/OSA calls carry explicit 10 s timeouts (deadlock freedom).
// Armstrong: errors surface as HTTP 500 + span status Error — no silent swallow.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/rhl/businessos-backend/internal/semconv"
)

const (
	boardTracerName     = "businessos.board"
	oxigraphQueryPath   = "/query"
	boardQueryTimeout   = 10 * time.Second
	osaConwayPath       = "/api/board/conway/check"
	defaultOSABaseURL   = "http://localhost:8089"
	defaultOxigraphURL  = "http://localhost:7878"
)

// BoardHandler handles the /api/board route group.
// It queries Oxigraph for materialized L3 board intelligence and delegates
// Conway triggers to the OSA backend.
type BoardHandler struct {
	httpClient  *http.Client
	tracer      trace.Tracer
	logger      *slog.Logger
	oxigraphURL string
	osaBaseURL  string
}

// NewBoardHandler constructs a BoardHandler.
// oxigraphURL defaults to OXIGRAPH_URL env var or http://localhost:7878.
// osaBaseURL defaults to OSA_BASE_URL env var or http://localhost:8089.
func NewBoardHandler() *BoardHandler {
	oxURL := os.Getenv("OXIGRAPH_URL")
	if oxURL == "" {
		oxURL = defaultOxigraphURL
	}
	osaURL := os.Getenv("OSA_BASE_URL")
	if osaURL == "" {
		osaURL = defaultOSABaseURL
	}
	return &BoardHandler{
		// WvdA: explicit transport timeout prevents connection leak.
		httpClient:  &http.Client{Timeout: boardQueryTimeout},
		tracer:      otel.Tracer(boardTracerName),
		logger:      slog.Default(),
		oxigraphURL: oxURL,
		osaBaseURL:  osaURL,
	}
}

// RegisterBoardRoutes wires /api/board routes on the given router group.
// Routes are intentionally unauthenticated for now — callers should add auth
// middleware at the top-level group before calling this function.
func RegisterBoardRoutes(api *gin.RouterGroup, h *BoardHandler) {
	board := api.Group("/board")
	{
		board.GET("/intelligence", h.GetIntelligence)
		board.GET("/briefing", h.GetBriefing)
		board.GET("/escalations", h.GetEscalations)
		board.POST("/conway/trigger", h.TriggerConwayCheck)
	}
}

// ─── SPARQL query helpers ────────────────────────────────────────────────────

// sparqlSelectResult holds the raw JSON-from-Oxigraph SELECT response.
type sparqlSelectResult struct {
	Results struct {
		Bindings []map[string]struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"bindings"`
	} `json:"results"`
}

// runSPARQLSelect sends a SELECT query to Oxigraph and returns the bindings.
// WvdA: caller must pass a context with a deadline — this method does not add one.
func (h *BoardHandler) runSPARQLSelect(ctx context.Context, query string) ([]map[string]string, error) {
	endpoint := h.oxigraphURL + oxigraphQueryPath

	data := url.Values{}
	data.Set("query", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint,
		bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build SPARQL request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SPARQL request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read SPARQL response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SPARQL endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var result sparqlSelectResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse SPARQL JSON: %w", err)
	}

	// Flatten bindings to map[var]value for convenience.
	rows := make([]map[string]string, 0, len(result.Results.Bindings))
	for _, binding := range result.Results.Bindings {
		row := make(map[string]string, len(binding))
		for k, v := range binding {
			row[k] = v.Value
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// ─── GET /api/board/intelligence ────────────────────────────────────────────

// BoardIntelligenceResponse is the JSON shape returned by GET /api/board/intelligence.
type BoardIntelligenceResponse struct {
	OrganizationalHealthSummary string `json:"organizational_health_summary,omitempty"`
	TopRisk                     string `json:"top_risk,omitempty"`
	ProcessVelocityTrend        string `json:"process_velocity_trend,omitempty"`
	ComplianceStatus            string `json:"compliance_status,omitempty"`
	WeeklyROIDelta              string `json:"weekly_roi_delta,omitempty"`
	IssuesAutoResolved          string `json:"issues_auto_resolved,omitempty"`
	IssuesPendingEscalation     string `json:"issues_pending_escalation,omitempty"`
	StructuralIssueCount        string `json:"structural_issue_count,omitempty"`
	OperationalIssueCount       string `json:"operational_issue_count,omitempty"`
	HighestConwayScore          string `json:"highest_conway_score,omitempty"`
	WorstQueueStability         string `json:"worst_queue_stability,omitempty"`
	LastRefreshed               string `json:"last_refreshed,omitempty"`
	DerivationLevel             string `json:"derivation_level,omitempty"`
}

const intelligenceQuery = `
PREFIX bos: <https://chatmangpt.com/ontology/businessos/>
SELECT
  ?orgHealthSummary ?topRisk ?velocityTrend ?complianceStatus
  ?weeklyROIDelta ?issuesAutoResolved ?issuesPendingEscalation
  ?structuralIssueCount ?operationalIssueCount
  ?highestConwayScore ?worstQueueStability
  ?lastRefreshed ?derivationLevel
WHERE {
  ?bi a bos:BoardIntelligence ;
    bos:organizationalHealthSummary ?orgHealthSummary ;
    bos:processVelocityTrend        ?velocityTrend ;
    bos:complianceStatus            ?complianceStatus ;
    bos:issuesAutoResolved          ?issuesAutoResolved ;
    bos:issuesPendingEscalation     ?issuesPendingEscalation ;
    bos:derivationLevel             ?derivationLevel .
  OPTIONAL { ?bi bos:topRisk                ?topRisk . }
  OPTIONAL { ?bi bos:weeklyROIDelta         ?weeklyROIDelta . }
  OPTIONAL { ?bi bos:structuralIssueCount   ?structuralIssueCount . }
  OPTIONAL { ?bi bos:operationalIssueCount  ?operationalIssueCount . }
  OPTIONAL { ?bi bos:highestConwayScore     ?highestConwayScore . }
  OPTIONAL { ?bi bos:worstQueueStability    ?worstQueueStability . }
  OPTIONAL { ?bi bos:lastRefreshed          ?lastRefreshed . }
}
LIMIT 1
`

// GetIntelligence handles GET /api/board/intelligence.
// Emits span: board.briefing_render
func (h *BoardHandler) GetIntelligence(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), semconv.BoardBriefingRenderSpan)
	defer span.End()

	// WvdA: 10 s deadline on Oxigraph query.
	qCtx, cancel := context.WithTimeout(ctx, boardQueryTimeout)
	defer cancel()

	rows, err := h.runSPARQLSelect(qCtx, intelligenceQuery)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.logger.ErrorContext(ctx, "board.intelligence query failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "oxigraph_unavailable",
			"message": err.Error(),
		})
		return
	}

	if len(rows) == 0 {
		span.SetAttributes(semconv.BoardSectionCount(0))
		c.JSON(http.StatusOK, gin.H{
			"data":    nil,
			"message": "no board intelligence materialized yet",
		})
		return
	}

	row := rows[0]
	resp := BoardIntelligenceResponse{
		OrganizationalHealthSummary: row["orgHealthSummary"],
		TopRisk:                     row["topRisk"],
		ProcessVelocityTrend:        row["velocityTrend"],
		ComplianceStatus:            row["complianceStatus"],
		WeeklyROIDelta:              row["weeklyROIDelta"],
		IssuesAutoResolved:          row["issuesAutoResolved"],
		IssuesPendingEscalation:     row["issuesPendingEscalation"],
		StructuralIssueCount:        row["structuralIssueCount"],
		OperationalIssueCount:       row["operationalIssueCount"],
		HighestConwayScore:          row["highestConwayScore"],
		WorstQueueStability:         row["worstQueueStability"],
		LastRefreshed:               row["lastRefreshed"],
		DerivationLevel:             row["derivationLevel"],
	}

	span.SetAttributes(
		semconv.BoardSectionCount(5),
		semconv.BoardHasStructuralIssues(row["structuralIssueCount"] != "" && row["structuralIssueCount"] != "0"),
	)

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── GET /api/board/briefing ─────────────────────────────────────────────────

// GetBriefing handles GET /api/board/briefing.
// Returns the latest briefing narrative text derived from L3 intelligence.
// Emits span: board.briefing_render
func (h *BoardHandler) GetBriefing(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), semconv.BoardBriefingRenderSpan)
	defer span.End()

	qCtx, cancel := context.WithTimeout(ctx, boardQueryTimeout)
	defer cancel()

	rows, err := h.runSPARQLSelect(qCtx, intelligenceQuery)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.logger.ErrorContext(ctx, "board.briefing query failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "oxigraph_unavailable",
			"message": err.Error(),
		})
		return
	}

	if len(rows) == 0 {
		span.SetAttributes(semconv.BoardHasStructuralIssues(false))
		c.JSON(http.StatusOK, gin.H{
			"briefing": "No board intelligence available. Run the L3 SPARQL inference chain first.",
		})
		return
	}

	row := rows[0]
	hasStructural := row["structuralIssueCount"] != "" && row["structuralIssueCount"] != "0"

	briefing := buildBriefingText(row, hasStructural)

	sectionCount := int64(5)
	if hasStructural {
		sectionCount = 6
	}

	span.SetAttributes(
		semconv.BoardHasStructuralIssues(hasStructural),
		semconv.BoardSectionCount(sectionCount),
	)

	c.JSON(http.StatusOK, gin.H{
		"briefing":       briefing,
		"section_count":  sectionCount,
		"has_structural": hasStructural,
	})
}

// buildBriefingText renders the briefing narrative from an L3 intelligence row.
func buildBriefingText(row map[string]string, hasStructural bool) string {
	health := row["orgHealthSummary"]
	if health == "" {
		health = "unknown"
	}
	trend := row["velocityTrend"]
	if trend == "" {
		trend = "unknown"
	}
	compliance := row["complianceStatus"]
	if compliance == "" {
		compliance = "unknown"
	}
	topRisk := row["topRisk"]
	if topRisk == "" {
		topRisk = "none identified"
	}
	roi := row["weeklyROIDelta"]
	if roi == "" {
		roi = "0"
	}
	autoResolved := row["issuesAutoResolved"]
	if autoResolved == "" {
		autoResolved = "0"
	}

	text := fmt.Sprintf(
		"BOARD CHAIR INTELLIGENCE BRIEFING\n\n"+
			"1. ORGANIZATIONAL HEALTH: %s (trend: %s)\n"+
			"2. COMPLIANCE STATUS: %s\n"+
			"3. TOP RISK: %s\n"+
			"4. AUTONOMOUS HEALING: %s issues resolved (est. $%s USD value)\n"+
			"5. PROCESS INTELLIGENCE: Derived from Oxigraph L3 inference chain\n",
		health, trend, compliance, topRisk, autoResolved, roi,
	)

	if hasStructural {
		structCount := row["structuralIssueCount"]
		text += fmt.Sprintf("\n6. STRUCTURAL DECISIONS REQUIRED: %s department(s) have Conway violations requiring board-level decision\n", structCount)
	}

	return text
}

// ─── GET /api/board/escalations ─────────────────────────────────────────────

const escalationsQuery = `
PREFIX bos: <https://chatmangpt.com/ontology/businessos/>
SELECT ?dept ?structuralCount ?operationalCount ?conwayScore
WHERE {
  ?bi a bos:BoardIntelligence ;
    bos:structuralIssueCount  ?structuralCount ;
    bos:operationalIssueCount ?operationalCount .
  OPTIONAL { ?bi bos:highestConwayScore ?conwayScore . }
  OPTIONAL { ?bi bos:forOrganization ?dept . }
  FILTER(xsd:integer(?structuralCount) > 0 || xsd:integer(?operationalCount) > 0)
}
LIMIT 100
`

// EscalationItem represents one department escalation in the response.
type EscalationItem struct {
	Department       string `json:"department,omitempty"`
	StructuralCount  string `json:"structural_count"`
	OperationalCount string `json:"operational_count"`
	ConwayScore      string `json:"conway_score,omitempty"`
}

// GetEscalations handles GET /api/board/escalations.
// Returns departments with structuralIssueCount > 0 or operationalIssueCount > 0.
// Emits span: board.structural_escalation
func (h *BoardHandler) GetEscalations(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), semconv.BoardStructuralEscalationSpan)
	defer span.End()

	qCtx, cancel := context.WithTimeout(ctx, boardQueryTimeout)
	defer cancel()

	rows, err := h.runSPARQLSelect(qCtx, escalationsQuery)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.logger.ErrorContext(ctx, "board.escalations query failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "oxigraph_unavailable",
			"message": err.Error(),
		})
		return
	}

	escalations := make([]EscalationItem, 0, len(rows))
	for _, row := range rows {
		escalations = append(escalations, EscalationItem{
			Department:       row["dept"],
			StructuralCount:  row["structuralCount"],
			OperationalCount: row["operationalCount"],
			ConwayScore:      row["conwayScore"],
		})
	}

	span.SetAttributes(
		semconv.BoardEscalationsEmitted(int64(len(escalations))),
		semconv.BoardEscalationType(semconv.BoardEscalationTypeValues.ConwayViolation),
	)

	c.JSON(http.StatusOK, gin.H{
		"escalations": escalations,
		"count":       len(escalations),
	})
}

// ─── POST /api/board/conway/trigger ─────────────────────────────────────────

// TriggerConwayCheck handles POST /api/board/conway/trigger.
// Delegates to OSA's Conway check endpoint and relays the result.
// Emits span: board.conway_check
func (h *BoardHandler) TriggerConwayCheck(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), semconv.BoardConwayCheckSpan)
	defer span.End()

	processID := c.DefaultQuery("process_id", "default")
	span.SetAttributes(semconv.BoardProcessId(processID))

	// WvdA: 10 s timeout for OSA HTTP call.
	reqCtx, cancel := context.WithTimeout(ctx, boardQueryTimeout)
	defer cancel()

	osaURL := h.osaBaseURL + osaConwayPath

	reqBody := fmt.Sprintf(`{"process_id": %q}`, processID)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, osaURL,
		bytes.NewBufferString(reqBody))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.logger.ErrorContext(ctx, "board.conway_trigger build request failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "request_build_failed",
			"message": err.Error(),
		})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.logger.ErrorContext(ctx, "board.conway_trigger OSA call failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "osa_unavailable",
			"message": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "response_read_failed",
			"message": err.Error(),
		})
		return
	}

	span.SetAttributes(semconv.BoardIsViolation(resp.StatusCode != http.StatusOK))

	// Relay OSA's status code and body transparently.
	c.Data(resp.StatusCode, "application/json", body)
}
