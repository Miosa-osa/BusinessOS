package ontology

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	bossemconv "github.com/rhl/businessos-backend/internal/semconv"
)

const (
	// l0NamedGraph is the canonical Oxigraph named graph for L0 ground-truth facts.
	// L1-L3 SPARQL files reference this graph via FROM <l0NamedGraph>.
	l0NamedGraph   = "https://chatmangpt.com/ontology/businessos/l0"
	l0SyncInterval = 15 * time.Minute
	l0SubprocTimeout = 30 * time.Second
)

// bosExecuteResult mirrors the JSON stdout from `bos ontology execute`.
type bosExecuteResult struct {
	TotalRows            int `json:"total_rows"`
	TotalConstructTriples int `json:"total_construct_triples"`
	Tables               []struct {
		Table             string `json:"table"`
		RowsLoaded        int    `json:"rows_loaded"`
		TriplesGenerated  int    `json:"triples_generated"`
		ConstructTriples  int    `json:"construct_triples"`
	} `json:"tables"`
}

// CanopyIntelligencePusher is an optional side-car that receives the L0 sync
// results and forwards them to Canopy.  Implemented in integrations/canopy.
type CanopyIntelligencePusher interface {
	PushIntelligence(ctx context.Context, caseCount, handoffCount int) error
}

// BoardchairL0Sync syncs BusinessOS case data into Oxigraph as L0 RDF facts.
// Two paths:
//   1. Subprocess (`bos ontology execute`) for legacy BOS-managed tables.
//   2. Direct DB → Oxigraph for process_cases and process_discovery_results (Phase 4b).
//
// WvdA: L0 = ground truth event log. Must be continuously updated.
// Armstrong: supervised background job, subprocess bounded by l0SubprocTimeout.
type BoardchairL0Sync struct {
	bosPath      string // path to bos binary (env BOS_PATH)
	mappingFile  string // path to mapping JSON (env BOS_MAPPING_FILE)
	dbURL        string // PostgreSQL connection string (env DATABASE_URL)
	oxigraphURL  string // Oxigraph SPARQL endpoint (env OXIGRAPH_URL, default http://localhost:7878)
	pool         *pgxpool.Pool
	tracer       trace.Tracer
	ticker       *time.Ticker
	done         chan struct{}
	canopy       CanopyIntelligencePusher // nil → skip push
}

// NewBoardchairL0Sync creates a new L0 sync service.
//
// bosPath: path to the `bos` binary (default "bos").
// mappingFile: path to the mapping JSON file for `bos ontology execute`.
// dbURL: PostgreSQL connection string.
func NewBoardchairL0Sync(bosPath, mappingFile, dbURL string) *BoardchairL0Sync {
	if bosPath == "" {
		bosPath = "bos"
	}
	oxigraphURL := os.Getenv("OXIGRAPH_URL")
	if oxigraphURL == "" {
		oxigraphURL = "http://localhost:7878"
	}
	return &BoardchairL0Sync{
		bosPath:     bosPath,
		mappingFile: mappingFile,
		dbURL:       dbURL,
		oxigraphURL: oxigraphURL,
		tracer:      otel.Tracer("businessos.board"),
		done:        make(chan struct{}),
	}
}

// SetPool wires a pgxpool.Pool for direct DB → Oxigraph sync (Phase 4b).
// Without a pool, the direct-DB path is silently skipped.
func (s *BoardchairL0Sync) SetPool(pool *pgxpool.Pool) {
	s.pool = pool
}

// SetCanopyPusher wires an optional Canopy intelligence push after each sync.
func (s *BoardchairL0Sync) SetCanopyPusher(p CanopyIntelligencePusher) {
	s.canopy = p
}

// Start begins the periodic L0 sync. Call in a goroutine.
// Armstrong: caller must supervise — if this crashes, restart it.
func (s *BoardchairL0Sync) Start(ctx context.Context) {
	s.ticker = time.NewTicker(l0SyncInterval)
	defer s.ticker.Stop()

	// Run immediately on start
	if err := s.Sync(ctx); err != nil {
		slog.Error("board.l0_sync initial sync failed", "error", err)
	}

	for {
		select {
		case <-s.ticker.C:
			if err := s.Sync(ctx); err != nil {
				slog.Error("board.l0_sync periodic sync failed", "error", err)
			}
		case <-s.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop halts the sync loop.
func (s *BoardchairL0Sync) Stop() {
	close(s.done)
}

// Sync performs one L0 sync by invoking `bos ontology execute` as a subprocess.
// The subprocess writes N-Triples directly to Oxigraph (port 7878); Go never
// touches Oxigraph directly.
func (s *BoardchairL0Sync) Sync(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, bossemconv.BoardL0SyncSpan)
	defer span.End()

	if s.mappingFile == "" {
		slog.Warn("board.l0_sync skipped: BOS_MAPPING_FILE not set")
		return nil
	}

	// Inject W3C traceparent so bos CLI participates in the same trace.
	traceparent := extractTraceparent(ctx)

	// WvdA: 30s hard timeout on the subprocess — deadlock freedom guarantee.
	subCtx, cancel := context.WithTimeout(ctx, l0SubprocTimeout)
	defer cancel()

	cmd := exec.CommandContext(subCtx, s.bosPath,
		"ontology", "execute",
		"--mapping", s.mappingFile,
		"--database", s.dbURL,
		"--graph", l0NamedGraph,
	)
	cmd.Env = append(os.Environ(),
		"TRACEPARENT="+traceparent,
		"WEAVER_LIVE_CHECK="+os.Getenv("WEAVER_LIVE_CHECK"),
		"WEAVER_OTLP_ENDPOINT="+os.Getenv("WEAVER_OTLP_ENDPOINT"),
		"CHATMANGPT_CORRELATION_ID="+os.Getenv("CHATMANGPT_CORRELATION_ID"),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		slog.Error("board.l0_sync bos subprocess failed",
			"error", err, "stderr", stderr.String())
		return fmt.Errorf("board.l0_sync bos execute: %w", err)
	}

	// Parse JSON output from bos CLI to get counts for span attributes.
	var result bosExecuteResult
	caseCount, handoffCount := 0, 0
	if err := json.Unmarshal(stdout.Bytes(), &result); err == nil {
		caseCount = result.TotalRows
		handoffCount = result.TotalConstructTriples
		// Attempt to split by table name if available.
		for _, t := range result.Tables {
			switch t.Table {
			case "cases":
				caseCount = t.RowsLoaded
			case "process_handoffs":
				handoffCount = t.RowsLoaded
			}
		}
	}

	span.SetAttributes(
		attribute.Int("board.l0_sync.case_count", caseCount),
		attribute.Int("board.l0_sync.handoff_count", handoffCount),
	)
	slog.Info("board.l0_sync complete", "cases", caseCount, "handoffs", handoffCount)

	// Direct DB → Oxigraph for process_cases + process_discovery_results (Phase 4b).
	// Non-fatal: subprocess result is already in Oxigraph if subprocess succeeded.
	if s.pool != nil {
		if err := s.syncDirectToOxigraph(ctx); err != nil {
			slog.Warn("board.l0_sync direct DB sync failed (non-fatal)", "error", err)
		}
	}

	// Fire-and-forget push to Canopy (Armstrong: never let failures propagate).
	if s.canopy != nil {
		go func() {
			pushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.canopy.PushIntelligence(pushCtx, caseCount, handoffCount); err != nil {
				slog.Warn("board.l0_sync canopy push failed", "error", err)
			}
		}()
	}

	return nil
}

// extractTraceparent reads the W3C traceparent from the current span context.
func extractTraceparent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if tp, ok := carrier["traceparent"]; ok {
		return tp
	}
	return ""
}

// ============================================================================
// DIRECT DB → OXIGRAPH SYNC (Phase 4b)
// ============================================================================

// syncDirectToOxigraph builds a SPARQL INSERT DATA statement from process_cases
// and process_discovery_results, then POSTs it to Oxigraph's /update endpoint.
// WvdA: 30s total timeout split across DB queries and Oxigraph POST.
func (s *BoardchairL0Sync) syncDirectToOxigraph(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf strings.Builder
	buf.WriteString("PREFIX bos: <https://chatmangpt.com/ontology/businessos/>\n")
	buf.WriteString("PREFIX org: <http://www.w3.org/ns/org#>\n")
	buf.WriteString("PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>\n")
	buf.WriteString("\n")
	buf.WriteString("INSERT DATA {\n")
	buf.WriteString("  GRAPH <" + l0NamedGraph + "> {\n")

	if err := s.writeCases(ctx, &buf); err != nil {
		return fmt.Errorf("writeCases: %w", err)
	}
	if err := s.writeDepartments(ctx, &buf); err != nil {
		return fmt.Errorf("writeDepartments: %w", err)
	}

	buf.WriteString("  }\n")
	buf.WriteString("}\n")

	oxURL := strings.TrimSuffix(s.oxigraphURL, "/") + "/update"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oxURL,
		strings.NewReader(buf.String()))
	if err != nil {
		return fmt.Errorf("build oxigraph request: %w", err)
	}
	req.Header.Set("Content-Type", "application/sparql-update")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("oxigraph POST /update: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("oxigraph /update returned %d", resp.StatusCode)
	}

	slog.Info("board.l0_sync direct DB → Oxigraph complete")
	return nil
}

// writeCases queries process_cases for the 90-day rolling window and appends
// one bos:ProcessCase triple per row to buf.
// WvdA: query bounded by ctx timeout; LIMIT 5000 enforces boundedness.
func (s *BoardchairL0Sync) writeCases(ctx context.Context, buf *strings.Builder) error {
	rows, err := s.pool.Query(ctx, `
		SELECT case_ref, department, status, COALESCE(cycle_time_seconds, 0)
		FROM process_cases
		WHERE created_at >= NOW() - INTERVAL '90 days'
		ORDER BY created_at DESC
		LIMIT 5000
	`)
	if err != nil {
		return fmt.Errorf("query process_cases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var caseRef, dept, status string
		var cycleTimeSec int
		if err := rows.Scan(&caseRef, &dept, &status, &cycleTimeSec); err != nil {
			return fmt.Errorf("scan process_cases row: %w", err)
		}
		caseRef = strings.ReplaceAll(caseRef, ">", "")
		dept = strings.ReplaceAll(dept, `"`, "'")
		status = strings.ReplaceAll(status, `"`, "'")
		buf.WriteString(fmt.Sprintf(
			"    bos:case/%s a bos:ProcessCase ;\n"+
				"      bos:department \"%s\" ;\n"+
				"      bos:caseStatus \"%s\" ;\n"+
				"      bos:cycleTimeSeconds %d .\n",
			caseRef, dept, status, cycleTimeSec,
		))
	}
	return rows.Err()
}

// writeDepartments queries distinct departments from process_cases joined with
// process_discovery_results, and appends one org:Organization triple per
// department with bos:eventLogFitness, bos:avgCycleTimeHours, and
// bos:bottleneckActivityName attributes.
// WvdA: query bounded by ctx timeout; departments are org-bounded (<1k rows).
func (s *BoardchairL0Sync) writeDepartments(ctx context.Context, buf *strings.Builder) error {
	rows, err := s.pool.Query(ctx, `
		SELECT
		    pc.department,
		    COALESCE(dr.fitness, -1.0)             AS fitness,
		    COALESCE(dr.avg_cycle_time_hours, 0.0) AS avg_cycle_time_hours,
		    COALESCE(dr.bottleneck_activity, '')    AS bottleneck_activity
		FROM (
		    SELECT DISTINCT department FROM process_cases
		    WHERE created_at >= NOW() - INTERVAL '90 days'
		) pc
		LEFT JOIN LATERAL (
		    SELECT fitness, avg_cycle_time_hours, bottleneck_activity
		    FROM process_discovery_results dr2
		    WHERE dr2.department = pc.department
		    ORDER BY dr2.discovered_at DESC
		    LIMIT 1
		) dr ON true
		ORDER BY pc.department
	`)
	if err != nil {
		return fmt.Errorf("query departments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dept, bottleneck string
		var fitness, avgCycleHours float64
		if err := rows.Scan(&dept, &fitness, &avgCycleHours, &bottleneck); err != nil {
			return fmt.Errorf("scan department row: %w", err)
		}
		dept = strings.ReplaceAll(dept, `"`, "'")
		bottleneck = strings.ReplaceAll(bottleneck, `"`, "'")
		deptID := strings.ReplaceAll(dept, " ", "_")
		buf.WriteString(fmt.Sprintf(
			"    bos:dept/%s a org:Organization ;\n"+
				"      rdfs:label \"%s\" ;\n"+
				"      bos:eventLogFitness %.4f ;\n"+
				"      bos:avgCycleTimeHours %.4f ;\n"+
				"      bos:bottleneckActivityName \"%s\" .\n",
			deptID, dept, fitness, avgCycleHours, bottleneck,
		))
	}
	return rows.Err()
}
