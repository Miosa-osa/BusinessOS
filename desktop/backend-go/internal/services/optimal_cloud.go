package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/optimal"
)

// OptimalCloudService bridges the OptimalOS engine pipeline with PostgreSQL.
// All operations are scoped to a workspace_id for multi-tenant isolation.
// The classification, routing, tiering, and composition logic from the optimal
// package is reused unchanged — only the storage layer differs from the local
// filesystem+SQLite implementation.
type OptimalCloudService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewOptimalCloudService creates a new cloud-backed optimal engine service.
func NewOptimalCloudService(pool *pgxpool.Pool, logger *slog.Logger) *OptimalCloudService {
	return &OptimalCloudService{
		pool:   pool,
		logger: logger,
	}
}

// ── Request / Response Types ──────────────────────────────────────────────────

// IngestRequest carries the input for the full classify→route→store pipeline.
type IngestRequest struct {
	WorkspaceID string
	UserID      string
	Text        string
	Genre       string // optional override; if empty, classification determines genre
	NodeID      string // optional explicit target node UUID
}

// IngestResult is returned after a successful ingest.
type IngestResult struct {
	SignalID       string
	Classification optimal.SignalClassification
	Routing        optimal.RoutingResult
	L0Abstract     string
	L1Overview     string
}

// SearchRequest carries the parameters for a hybrid full-text search.
type SearchRequest struct {
	WorkspaceID string
	Query       string
	NodeID      string // optional: filter to a specific node UUID
	Genre       string // optional: filter by genre
	Limit       int
}

// OptimalSearchResult is a single hit from SearchSignals.
type OptimalSearchResult struct {
	SignalID    string    `json:"signal_id"`
	Title       string    `json:"title"`
	L0Abstract  string    `json:"l0_abstract"`
	Genre       string    `json:"genre"`
	PrimaryNode string    `json:"primary_node"`
	Score       float64   `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
}

// DashboardStats holds aggregate counts for the workspace dashboard widget.
type DashboardStats struct {
	TotalSignals      int            `json:"total_signals"`
	TotalEntities     int            `json:"total_entities"`
	TotalObservations int            `json:"total_observations"`
	SignalsByGenre    map[string]int `json:"signals_by_genre"`
}

// ── IngestSignal ──────────────────────────────────────────────────────────────

// IngestSignal runs the full OptimalOS pipeline and persists the result to
// PostgreSQL. Pipeline: Classify → RouteWithTopology → L0/L1 → INSERT signal
// → UpsertEntities → INSERT edges → append decision log.
func (s *OptimalCloudService) IngestSignal(ctx context.Context, req IngestRequest) (*IngestResult, error) {
	if strings.TrimSpace(req.Text) == "" {
		return nil, fmt.Errorf("optimal_cloud: ingest text must not be empty")
	}
	if req.WorkspaceID == "" || req.UserID == "" {
		return nil, fmt.Errorf("optimal_cloud: workspace_id and user_id are required")
	}

	// 1. CLASSIFY — deterministic regex, no I/O.
	cl := optimal.Classify(req.Text)
	genre := req.Genre
	if genre == "" {
		genre = cl.Genre
	}

	// 2. ROUTE — load topology from DB, fall back to defaultRoutingRules when absent.
	topo, err := s.loadTopology(ctx, req.WorkspaceID)
	if err != nil {
		// Non-fatal: degrade to default routing.
		s.logger.WarnContext(ctx, "optimal_cloud: topology load failed, using defaults",
			"workspace_id", req.WorkspaceID, "error", err)
	}
	routing := optimal.RouteWithTopology(req.Text, topo)

	// 3. TIERED ABSTRACTS
	l0 := optimal.L0Abstract(req.Text)
	l1 := optimal.L1Overview(req.Text)

	// 4. METADATA
	slug := cloudMakeSlug(req.Text, 50)
	title := cloudExtractTitle(req.Text)
	now := time.Now().UTC()
	uri := fmt.Sprintf("optimal://workspaces/%s/signals/%s-%s",
		req.WorkspaceID, now.Format("2006-01-02"), slug)
	path := fmt.Sprintf("workspaces/%s/signals/%s-%s.md",
		req.WorkspaceID, now.Format("2006-01-02"), slug)

	crossRefJSON, _ := json.Marshal(routing.CrossRefNodes)
	entitiesExtracted := cloudExtractEntities(req.Text)
	entitiesJSON, _ := json.Marshal(entitiesExtracted)

	// Resolve optional node_id (empty string → NULL).
	var nodeIDPtr *uuid.UUID
	if req.NodeID != "" {
		if parsed, err := uuid.Parse(req.NodeID); err == nil {
			nodeIDPtr = &parsed
		}
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: invalid workspace_id %q: %w", req.WorkspaceID, err)
	}

	// 5. INSERT signal into PostgreSQL.
	const insertSignal = `
		INSERT INTO optimal_signals (
			workspace_id, node_id, created_by,
			title, slug, uri, path, content,
			mode, genre, signal_type, format, structure,
			sn_ratio, l0_abstract, l1_overview,
			primary_node, cross_ref_nodes, entities
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15, $16,
			$17, $18::jsonb, $19::jsonb
		) RETURNING id`

	var signalID uuid.UUID
	err = s.pool.QueryRow(ctx, insertSignal,
		workspaceUUID, nodeIDPtr, req.UserID,
		title, slug, uri, path, req.Text,
		cl.Mode, genre, cl.Type, cl.Format, cl.Structure,
		cl.SNRatio, l0, l1,
		routing.PrimaryNode, string(crossRefJSON), string(entitiesJSON),
	).Scan(&signalID)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: insert signal: %w", err)
	}

	s.logger.InfoContext(ctx, "optimal_cloud: signal ingested",
		"signal_id", signalID,
		"workspace_id", req.WorkspaceID,
		"genre", genre,
		"primary_node", routing.PrimaryNode,
	)

	// 6. UPSERT entities (non-fatal: a failure here must not roll back the signal).
	for _, entityName := range entitiesExtracted {
		if err := s.upsertEntity(ctx, workspaceUUID, entityName, "person"); err != nil {
			s.logger.WarnContext(ctx, "optimal_cloud: entity upsert failed",
				"entity", entityName, "error", err)
		}
	}

	// 7. INSERT edges: signal lives_in its primary node; cross_ref for all others.
	_ = s.insertEdge(ctx, workspaceUUID, signalID, uuid.Nil, "signal", "node",
		routing.PrimaryNode, "lives_in", 1.0)
	for _, xref := range routing.CrossRefNodes {
		_ = s.insertEdge(ctx, workspaceUUID, signalID, uuid.Nil, "signal", "node",
			xref, "cross_ref", 1.2)
	}

	// 8. DECISION LOG — append-only audit.
	_ = s.appendDecisionLog(ctx, workspaceUUID, req.UserID, "created", "signal", signalID,
		fmt.Sprintf("ingested via pipeline: genre=%s primary_node=%s sn_ratio=%.2f",
			genre, routing.PrimaryNode, cl.SNRatio))

	return &IngestResult{
		SignalID:       signalID.String(),
		Classification: cl,
		Routing:        routing,
		L0Abstract:     l0,
		L1Overview:     l1,
	}, nil
}

// ── SearchSignals ─────────────────────────────────────────────────────────────

// SearchSignals performs a weighted full-text search using PostgreSQL tsvector
// and applies the same temporal decay formula as hybrid.go.
func (s *OptimalCloudService) SearchSignals(ctx context.Context, req SearchRequest) ([]OptimalSearchResult, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("optimal_cloud: search query must not be empty")
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: invalid workspace_id %q: %w", req.WorkspaceID, err)
	}

	// Build optional filter clauses and args.
	// $1=workspace_id, $2=tsquery, $3..=optional filters
	args := []any{workspaceUUID, req.Query}
	extraWhere := ""

	if req.NodeID != "" {
		if parsed, err := uuid.Parse(req.NodeID); err == nil {
			args = append(args, parsed)
			extraWhere += fmt.Sprintf(" AND node_id = $%d", len(args))
		}
	}
	if req.Genre != "" {
		args = append(args, req.Genre)
		extraWhere += fmt.Sprintf(" AND genre = $%d", len(args))
	}
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT id, title, l0_abstract, genre, primary_node, created_at,
		       ts_rank_cd(search_vector, websearch_to_tsquery('english', $2)) AS rank
		FROM optimal_signals
		WHERE workspace_id = $1
		  AND valid_until IS NULL
		  AND search_vector @@ websearch_to_tsquery('english', $2)
		  %s
		ORDER BY rank DESC
		LIMIT $%d`, extraWhere, len(args))

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: search query: %w", err)
	}
	defer rows.Close()

	var results []OptimalSearchResult
	for rows.Next() {
		var r OptimalSearchResult
		var id uuid.UUID
		var rank float64
		if err := rows.Scan(&id, &r.Title, &r.L0Abstract, &r.Genre, &r.PrimaryNode, &r.CreatedAt, &rank); err != nil {
			return nil, fmt.Errorf("optimal_cloud: scan search row: %w", err)
		}
		r.SignalID = id.String()
		// Apply temporal decay on top of BM25 rank — same formula as hybrid.go.
		decay := cloudTemporalDecay(r.CreatedAt, r.Genre)
		r.Score = rank * (0.75 + 0.25*decay) // BM25 dominates, decay adjusts recency
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal_cloud: search rows: %w", err)
	}
	return results, nil
}

// ── ComposeForReceiver ────────────────────────────────────────────────────────

// ComposeForReceiver loads the workspace topology, looks up the receiver's
// preferred genre, and reformats the content accordingly.
func (s *OptimalCloudService) ComposeForReceiver(ctx context.Context, workspaceID, content, receiverName string) (*optimal.ComposeResult, error) {
	topo, err := s.loadTopology(ctx, workspaceID)
	if err != nil {
		s.logger.WarnContext(ctx, "optimal_cloud: topology unavailable for compose",
			"workspace_id", workspaceID, "error", err)
		// Degrade gracefully: compose without topology.
	}
	result := optimal.ComposeForReceiver(content, receiverName, topo)
	return &result, nil
}

// ── Remember ──────────────────────────────────────────────────────────────────

// Remember stores an observation in the learning loop table.
// Valid categories: process, people, tool, decision, pattern, friction.
// Valid sources: explicit, contextual, mining.
func (s *OptimalCloudService) Remember(ctx context.Context, workspaceID, userID, content, category, source string) error {
	validCategories := map[string]bool{
		"process": true, "people": true, "tool": true,
		"decision": true, "pattern": true, "friction": true,
	}
	validSources := map[string]bool{
		"explicit": true, "contextual": true, "mining": true,
	}

	if !validCategories[category] {
		return fmt.Errorf("optimal_cloud: invalid observation category %q", category)
	}
	if !validSources[source] {
		return fmt.Errorf("optimal_cloud: invalid observation source %q", source)
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("optimal_cloud: observation content must not be empty")
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return fmt.Errorf("optimal_cloud: invalid workspace_id %q: %w", workspaceID, err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO optimal_observations (workspace_id, created_by, category, content, source)
		VALUES ($1, $2, $3, $4, $5)`,
		workspaceUUID, userID, category, content, source,
	)
	if err != nil {
		return fmt.Errorf("optimal_cloud: insert observation: %w", err)
	}

	s.logger.InfoContext(ctx, "optimal_cloud: observation stored",
		"workspace_id", workspaceID, "category", category, "source", source)
	return nil
}

// ── GetEscalations ────────────────────────────────────────────────────────────

// GetEscalations returns categories that have 3 or more observations, indicating
// a recurring pattern that merits escalation to a higher-fidelity signal.
func (s *OptimalCloudService) GetEscalations(ctx context.Context, workspaceID string) (map[string]int, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: invalid workspace_id %q: %w", workspaceID, err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT category, COUNT(*) AS cnt
		FROM optimal_observations
		WHERE workspace_id = $1
		GROUP BY category
		HAVING COUNT(*) >= 3
		ORDER BY cnt DESC`,
		workspaceUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: escalations query: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var cat string
		var cnt int
		if err := rows.Scan(&cat, &cnt); err != nil {
			return nil, fmt.Errorf("optimal_cloud: scan escalation row: %w", err)
		}
		result[cat] = cnt
	}
	return result, rows.Err()
}

// ── GetDashboardStats ─────────────────────────────────────────────────────────

// GetDashboardStats returns aggregate counts for the OptimalOS dashboard widget.
func (s *OptimalCloudService) GetDashboardStats(ctx context.Context, workspaceID string) (*DashboardStats, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: invalid workspace_id %q: %w", workspaceID, err)
	}

	stats := &DashboardStats{SignalsByGenre: make(map[string]int)}

	// Parallel counts with a single CTE query.
	err = s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM optimal_signals      WHERE workspace_id = $1 AND valid_until IS NULL),
			(SELECT COUNT(*) FROM optimal_entities     WHERE workspace_id = $1),
			(SELECT COUNT(*) FROM optimal_observations WHERE workspace_id = $1)`,
		workspaceUUID,
	).Scan(&stats.TotalSignals, &stats.TotalEntities, &stats.TotalObservations)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: dashboard counts: %w", err)
	}

	// Genre breakdown.
	rows, err := s.pool.Query(ctx, `
		SELECT genre, COUNT(*) FROM optimal_signals
		WHERE workspace_id = $1 AND valid_until IS NULL
		GROUP BY genre ORDER BY COUNT(*) DESC`,
		workspaceUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: genre breakdown: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var genre string
		var cnt int
		if err := rows.Scan(&genre, &cnt); err != nil {
			return nil, fmt.Errorf("optimal_cloud: scan genre row: %w", err)
		}
		stats.SignalsByGenre[genre] = cnt
	}
	return stats, rows.Err()
}

// ── SaveTopology ──────────────────────────────────────────────────────────────

// SaveTopology persists (or replaces) the workspace topology configuration.
func (s *OptimalCloudService) SaveTopology(ctx context.Context, workspaceID string, config optimal.TopologyConfig) error {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return fmt.Errorf("optimal_cloud: invalid workspace_id %q: %w", workspaceID, err)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("optimal_cloud: marshal topology: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO optimal_topology (workspace_id, config)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (workspace_id) DO UPDATE
		  SET config = EXCLUDED.config,
		      updated_at = NOW()`,
		workspaceUUID, string(configJSON),
	)
	if err != nil {
		return fmt.Errorf("optimal_cloud: upsert topology: %w", err)
	}

	s.logger.InfoContext(ctx, "optimal_cloud: topology saved", "workspace_id", workspaceID)
	return nil
}

// ── InitializeWorkspace ───────────────────────────────────────────────────────

// InitializeWorkspace sets up a default topology for a new workspace with five
// generalised starter nodes covering the core business domains. This is the
// portable default — workspace owners can customise via SaveTopology.
func (s *OptimalCloudService) InitializeWorkspace(ctx context.Context, workspaceID, userID string) error {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return fmt.Errorf("optimal_cloud: invalid workspace_id %q: %w", workspaceID, err)
	}

	// Idempotent: if topology already exists, do nothing.
	var exists bool
	err = s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM optimal_topology WHERE workspace_id = $1)`,
		workspaceUUID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("optimal_cloud: check topology existence: %w", err)
	}
	if exists {
		s.logger.InfoContext(ctx, "optimal_cloud: workspace already initialized, skipping",
			"workspace_id", workspaceID)
		return nil
	}

	defaultTopology := optimal.TopologyConfig{
		Nodes: map[string]optimal.NodeConfig{
			"business": {Name: "My Business", Type: "entity", CrossRef: []string{"revenue"}},
			"projects": {Name: "Projects", Type: "operation", CrossRef: []string{"business"}},
			"team":     {Name: "Team", Type: "unit", CrossRef: []string{"business"}},
			"inbox":    {Name: "Inbox", Type: "catchall", CrossRef: nil},
			"revenue":  {Name: "Revenue", Type: "finance", CrossRef: []string{"business"}},
		},
		People: map[string]optimal.PersonConfig{},
		Routes: []optimal.RoutingRule{
			{Keywords: []string{"revenue", "money", "invoice", "payment", "deal", "arr"}, Target: "revenue", CrossRef: []string{"business"}},
			{Keywords: []string{"project", "milestone", "deadline", "sprint", "task"}, Target: "projects", CrossRef: []string{"business"}},
			{Keywords: []string{"team", "hire", "staff", "onboard", "sync", "meeting"}, Target: "team", CrossRef: []string{"business"}},
		},
	}

	configJSON, err := json.Marshal(defaultTopology)
	if err != nil {
		return fmt.Errorf("optimal_cloud: marshal default topology: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO optimal_topology (workspace_id, config)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (workspace_id) DO NOTHING`,
		workspaceUUID, string(configJSON),
	)
	if err != nil {
		return fmt.Errorf("optimal_cloud: insert default topology: %w", err)
	}

	// Append to decision log so the audit trail is complete from day 1.
	_ = s.appendDecisionLog(ctx, workspaceUUID, userID, "created", "topology", workspaceUUID,
		"workspace initialized with default 5-node topology")

	s.logger.InfoContext(ctx, "optimal_cloud: workspace initialized",
		"workspace_id", workspaceID, "user_id", userID)
	return nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// loadTopology fetches the workspace topology from PostgreSQL and deserializes
// it into an optimal.TopologyConfig. Returns nil (not an error) when no topology
// row exists so callers can fall back to defaultRoutingRules without branching.
func (s *OptimalCloudService) loadTopology(ctx context.Context, workspaceID string) (*optimal.TopologyConfig, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("optimal_cloud: invalid workspace_id %q: %w", workspaceID, err)
	}

	var configJSON string
	err = s.pool.QueryRow(ctx,
		`SELECT config FROM optimal_topology WHERE workspace_id = $1`,
		workspaceUUID,
	).Scan(&configJSON)
	if err != nil {
		// pgx returns pgx.ErrNoRows; treat as missing topology, not an error.
		return nil, nil //nolint:nilerr
	}

	var topo optimal.TopologyConfig
	if err := json.Unmarshal([]byte(configJSON), &topo); err != nil {
		return nil, fmt.Errorf("optimal_cloud: unmarshal topology: %w", err)
	}
	// Ensure maps are non-nil after unmarshal.
	if topo.Nodes == nil {
		topo.Nodes = make(map[string]optimal.NodeConfig)
	}
	if topo.People == nil {
		topo.People = make(map[string]optimal.PersonConfig)
	}
	return &topo, nil
}

// upsertEntity inserts or increments an entity's mention count within the workspace.
func (s *OptimalCloudService) upsertEntity(ctx context.Context, workspaceID uuid.UUID, name, entityType string) error {
	canonical := strings.ToLower(strings.TrimSpace(name))
	_, err := s.pool.Exec(ctx, `
		INSERT INTO optimal_entities (workspace_id, name, canonical_name, type, mention_count)
		VALUES ($1, $2, $3, $4, 1)
		ON CONFLICT (workspace_id, canonical_name) DO UPDATE
		  SET mention_count = optimal_entities.mention_count + 1,
		      last_seen = NOW()`,
		workspaceID, name, canonical, entityType,
	)
	if err != nil {
		return fmt.Errorf("optimal_cloud: upsert entity %q: %w", name, err)
	}
	return nil
}

// insertEdge creates a directed edge in the knowledge graph. sourceID and
// targetSlug are the string identifiers; targetSlug is stored as the target_id
// cast to UUID when valid, otherwise as a text fallback via a generated UUID
// from the slug string (deterministic, so the same slug always maps to the same
// pseudo-UUID edge key).
func (s *OptimalCloudService) insertEdge(ctx context.Context, workspaceID, sourceID uuid.UUID,
	_ uuid.UUID, sourceType, targetType, targetSlug, relation string, weight float64) error {

	// For node slugs that are not UUIDs (e.g. "09", "inbox"), derive a
	// deterministic UUID via uuid.NewSHA1 so the UNIQUE constraint still works.
	targetUUID, err := uuid.Parse(targetSlug)
	if err != nil {
		targetUUID = uuid.NewSHA1(uuid.NameSpaceURL, []byte("optimal:node:"+targetSlug))
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO optimal_edges (workspace_id, source_id, source_type, target_id, target_type, relation, weight)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (workspace_id, source_id, target_id, relation) DO NOTHING`,
		workspaceID, sourceID, sourceType, targetUUID, targetType, relation, weight,
	)
	if err != nil {
		return fmt.Errorf("optimal_cloud: insert edge %s→%s (%s): %w", sourceID, targetSlug, relation, err)
	}
	return nil
}

// appendDecisionLog writes a single audit entry. Non-fatal: log the error and
// continue — the audit trail must never block primary operations.
func (s *OptimalCloudService) appendDecisionLog(ctx context.Context, workspaceID uuid.UUID,
	actor, action, targetType string, targetID uuid.UUID, whatChanged string) error {

	_, err := s.pool.Exec(ctx, `
		INSERT INTO optimal_decision_log (workspace_id, actor, action, target_type, target_id, what_changed)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		workspaceID, actor, action, targetType, targetID, whatChanged,
	)
	if err != nil {
		s.logger.WarnContext(ctx, "optimal_cloud: decision log append failed",
			"workspace_id", workspaceID, "action", action, "error", err)
		return fmt.Errorf("optimal_cloud: decision log: %w", err)
	}
	return nil
}

// ── Pure-Go helpers (mirrors of ingest.go internals) ─────────────────────────

// cloudExtractEntities is a workspace-safe copy of extractEntitiesFromText.
// The function is unexported in the optimal package so we replicate the logic
// here rather than patch the package boundary.
func cloudExtractEntities(text string) []string {
	re := regexp.MustCompile(`(?:^|\s)([A-Z][a-z]{2,})(?:\s+[A-Z][a-z]{2,})*`)
	rePhrase := regexp.MustCompile(`\b[A-Z][a-z]{1,}(?:\s+[A-Z][a-z]{1,})+\b`)
	stopWords := map[string]struct{}{
		"The": {}, "A": {}, "An": {}, "In": {}, "On": {}, "At": {},
		"To": {}, "For": {}, "Of": {}, "And": {}, "Or": {}, "But": {},
		"It": {}, "He": {}, "She": {}, "We": {}, "They": {}, "You": {},
		"I": {}, "This": {}, "That": {}, "These": {}, "Those": {},
	}

	seen := make(map[string]struct{})
	var result []string

	for _, m := range re.FindAllStringSubmatch(text, -1) {
		if len(m) < 2 {
			continue
		}
		word := strings.TrimSpace(m[1])
		if _, stop := stopWords[word]; stop {
			continue
		}
		if _, ok := seen[word]; !ok {
			seen[word] = struct{}{}
			result = append(result, word)
		}
	}
	for _, phrase := range rePhrase.FindAllString(text, -1) {
		phrase = strings.TrimSpace(phrase)
		if _, ok := seen[phrase]; !ok {
			seen[phrase] = struct{}{}
			result = append(result, phrase)
		}
	}
	return result
}

// cloudExtractTitle derives a title from the first sentence of text.
func cloudExtractTitle(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "untitled"
	}
	s := cloudFirstSentence(text)
	if len(s) > 80 {
		s = s[:80]
		if idx := strings.LastIndexByte(s, ' '); idx > 40 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}

// cloudFirstSentence returns the first sentence terminated by '.', '!', or '?'.
func cloudFirstSentence(text string) string {
	for i, ch := range text {
		if ch == '.' || ch == '!' || ch == '?' {
			return text[:i+1]
		}
	}
	return text
}

var (
	cloudReNonAlnum    = regexp.MustCompile(`[^a-z0-9]+`)
	cloudReMultiHyphen = regexp.MustCompile(`-{2,}`)
)

// cloudMakeSlug is a workspace-safe copy of makeSlug from ingest.go.
func cloudMakeSlug(text string, maxLen int) string {
	s := strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return '-'
		}
		return unicode.ToLower(r)
	}, text)
	s = cloudReNonAlnum.ReplaceAllString(s, "-")
	s = cloudReMultiHyphen.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > maxLen {
		s = s[:maxLen]
		if idx := strings.LastIndexByte(s, '-'); idx > 0 {
			s = s[:idx]
		}
	}
	if s == "" {
		s = "signal"
	}
	return s
}

// cloudTemporalDecay applies the same exponential decay formula as hybrid.go.
// score = exp(-ln(2) * ageHours / halfLife), clamped to [0, 1].
func cloudTemporalDecay(ts time.Time, genre string) float64 {
	halfLives := map[string]float64{
		"message": 168, "transcript": 168, "standup": 336, "status": 336,
		"note": 1440, "plan": 720, "spec": 4320, "decision": 8760,
		"adr": 8760, "pattern": 17520,
	}
	const defaultHalfLife = 720.0

	hl, ok := halfLives[genre]
	if !ok {
		hl = defaultHalfLife
	}
	ageHours := time.Since(ts).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	return math.Exp(-math.Log(2) * ageHours / hl)
}
