package ontology

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	oxigraphStoreEndpoint = "/store"
	l0NamedGraph          = "http://businessos.local/l0"
	l0SyncInterval        = 15 * time.Minute
)

// BoardchairL0Sync syncs BusinessOS case data into Oxigraph as L0 RDF facts.
// WvdA: L0 = ground truth event log. Must be continuously updated.
// Armstrong: supervised background job, bounded query (LIMIT 10000), 30s timeout.
type BoardchairL0Sync struct {
	db          *sql.DB
	oxigraphURL string
	httpClient  *http.Client
	tracer      trace.Tracer
	ticker      *time.Ticker
	done        chan struct{}
}

// NewBoardchairL0Sync creates a new L0 sync service.
func NewBoardchairL0Sync(db *sql.DB, oxigraphURL string) *BoardchairL0Sync {
	return &BoardchairL0Sync{
		db:          db,
		oxigraphURL: oxigraphURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		tracer:      otel.Tracer("businessos.board"),
		done:        make(chan struct{}),
	}
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

// Sync performs one L0 sync: read cases from PostgreSQL → write Turtle to Oxigraph.
func (s *BoardchairL0Sync) Sync(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "board.l0_sync")
	defer span.End()

	turtle, caseCount, handoffCount, err := s.buildTurtle(ctx)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return fmt.Errorf("board.l0_sync build turtle: %w", err)
	}

	if err := s.writeTurtle(ctx, turtle); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return fmt.Errorf("board.l0_sync write turtle: %w", err)
	}

	span.SetAttributes(
		attribute.Int("case_count", caseCount),
		attribute.Int("handoff_count", handoffCount),
	)
	slog.Info("board.l0_sync complete",
		"cases", caseCount, "handoffs", handoffCount)
	return nil
}

func (s *BoardchairL0Sync) buildTurtle(ctx context.Context) (string, int, int, error) {
	var buf strings.Builder

	buf.WriteString(`@prefix bos: <http://businessos.local/ontology#> .
@prefix dcterms: <http://purl.org/dc/terms/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
@prefix prov: <http://www.w3.org/ns/prov#> .

`)

	caseCount, err := s.writeCases(ctx, &buf)
	if err != nil {
		return "", 0, 0, err
	}

	handoffCount, err := s.writeHandoffs(ctx, &buf)
	if err != nil {
		return "", 0, 0, err
	}

	return buf.String(), caseCount, handoffCount, nil
}

func (s *BoardchairL0Sync) writeCases(ctx context.Context, buf *strings.Builder) (int, error) {
	// WvdA: cases are the event log. Query active cases + recently completed.
	// Armstrong: bounded query — LIMIT 10000 prevents unbounded memory.
	query := `
		SELECT
			id::text,
			COALESCE(department, 'unknown') as department,
			COALESCE(status, 'unknown') as status,
			COALESCE(EXTRACT(EPOCH FROM (COALESCE(completed_at, NOW()) - created_at))::int, 0) as cycle_time_seconds,
			created_at,
			updated_at
		FROM cases
		WHERE created_at > NOW() - INTERVAL '90 days'
		ORDER BY created_at DESC
		LIMIT 10000
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		// Table may not exist yet — not a fatal error
		slog.Warn("board.l0_sync cases query failed (table may not exist)", "error", err)
		return 0, nil
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, dept, status string
		var cycleSeconds int
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &dept, &status, &cycleSeconds, &createdAt, &updatedAt); err != nil {
			continue
		}

		uri := fmt.Sprintf("http://businessos.local/cases/%s", sanitizeURI(id))
		fmt.Fprintf(buf, `<%s> a bos:Case ;
    bos:caseId "%s" ;
    bos:department "%s" ;
    bos:status "%s" ;
    bos:cycleTimeSeconds %d ;
    dcterms:created "%s"^^xsd:dateTime ;
    dcterms:modified "%s"^^xsd:dateTime ;
    prov:generatedAtTime "%s"^^xsd:dateTime .

`,
			uri, id, dept, status, cycleSeconds,
			createdAt.UTC().Format(time.RFC3339),
			updatedAt.UTC().Format(time.RFC3339),
			time.Now().UTC().Format(time.RFC3339),
		)
		count++
	}
	return count, rows.Err()
}

func (s *BoardchairL0Sync) writeHandoffs(ctx context.Context, buf *strings.Builder) (int, error) {
	query := `
		SELECT
			id::text,
			COALESCE(source_department, 'unknown') as source_dept,
			COALESCE(target_department, 'unknown') as target_dept,
			COALESCE(EXTRACT(EPOCH FROM duration)::int, 0) as duration_seconds,
			created_at
		FROM process_handoffs
		WHERE created_at > NOW() - INTERVAL '90 days'
		ORDER BY created_at DESC
		LIMIT 5000
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		slog.Warn("board.l0_sync handoffs query failed (table may not exist)", "error", err)
		return 0, nil
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, sourceDept, targetDept string
		var durationSeconds int
		var createdAt time.Time

		if err := rows.Scan(&id, &sourceDept, &targetDept, &durationSeconds, &createdAt); err != nil {
			continue
		}

		uri := fmt.Sprintf("http://businessos.local/handoffs/%s", sanitizeURI(id))
		fmt.Fprintf(buf, `<%s> a bos:ProcessHandoff ;
    bos:sourceDepartment "%s" ;
    bos:targetDepartment "%s" ;
    bos:durationSeconds %d ;
    bos:handoffAt "%s"^^xsd:dateTime .

`,
			uri, sourceDept, targetDept, durationSeconds,
			createdAt.UTC().Format(time.RFC3339),
		)
		count++
	}
	return count, rows.Err()
}

func (s *BoardchairL0Sync) writeTurtle(ctx context.Context, turtle string) error {
	url := fmt.Sprintf("%s%s?graph=%s", s.oxigraphURL, oxigraphStoreEndpoint, l0NamedGraph)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBufferString(turtle))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/turtle")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("oxigraph PUT failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("oxigraph PUT status %d", resp.StatusCode)
	}
	return nil
}

func sanitizeURI(s string) string {
	replacer := strings.NewReplacer(
		" ", "-", "/", "-", "\\", "-",
		"<", "", ">", "", "\"", "", "'", "",
	)
	return replacer.Replace(s)
}
