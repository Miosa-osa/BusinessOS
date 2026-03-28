package rdf

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/rhl/businessos-backend/internal/config"
	semconv "github.com/rhl/businessos-backend/internal/semconv"
)

// Client provides RDF operations against an Oxigraph triplestore.
type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	logger     *slog.Logger
	tracer     trace.Tracer
}

// NewClient creates a new RDF client.
// If cfg is nil, uses environment variable OXIGRAPH_URL or defaults to http://localhost:8890.
func NewClient(cfg *config.Config) *Client {
	baseURL := os.Getenv("OXIGRAPH_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8890"
	}

	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		timeout: 30 * time.Second,
		logger:  slog.Default(),
		tracer:  otel.Tracer("businessos.rdf"),
	}
}

// LoadOntology loads a Turtle ontology into Oxigraph.
// file: path to .ttl file
// Returns error if the operation fails.
func (c *Client) LoadOntology(ctx context.Context, file string) error {
	c.logger.InfoContext(ctx, "loading ontology", slog.String("file", file))

	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read file failed: %w", err)
	}

	return c.loadTurtle(ctx, content)
}

// LoadTurtleData loads Turtle data (bytes) into Oxigraph.
func (c *Client) LoadTurtleData(ctx context.Context, data []byte) error {
	return c.loadTurtle(ctx, data)
}

// loadTurtle POSTs Turtle data to Oxigraph /store endpoint.
func (c *Client) loadTurtle(ctx context.Context, data []byte) error {
	storeURL := fmt.Sprintf("%s/store", c.baseURL)

	ctx, span := c.tracer.Start(ctx, semconv.OxigraphWriteSpan,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String(string(semconv.RdfSparqlEndpointKey), c.baseURL),
		attribute.String(string(semconv.RdfWriteFormatKey), "text/turtle"),
		attribute.Int(string(semconv.RdfWriteTripleCountKey), len(data)/50), // rough estimate
	)

	req, err := http.NewRequestWithContext(ctx, "POST", storeURL, strings.NewReader(string(data)))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "text/turtle")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Success: 200 OK or 204 No Content
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	// Error response
	body, _ := io.ReadAll(resp.Body)
	err = fmt.Errorf("oxigraph returned HTTP %d: %s", resp.StatusCode, string(body))
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}

// LoadData inserts RDF data via SPARQL INSERT query.
// query: SPARQL INSERT DATA query
// Returns error if the operation fails.
func (c *Client) LoadData(ctx context.Context, query string) error {
	c.logger.DebugContext(ctx, "executing sparql insert", slog.String("query_len", fmt.Sprintf("%d bytes", len(query))))

	return c.executeSPARQL(ctx, query)
}

// QuerySPARQL executes a SELECT or CONSTRUCT query.
// Returns query results as string (JSON or XML depending on Oxigraph response).
func (c *Client) QuerySPARQL(ctx context.Context, query string) (string, error) {
	c.logger.DebugContext(ctx, "executing sparql query", slog.String("query_len", fmt.Sprintf("%d bytes", len(query))))

	ctx, span := c.tracer.Start(ctx, semconv.OxigraphQuerySpan,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String(string(semconv.RdfSparqlEndpointKey), c.baseURL),
		attribute.String(string(semconv.RdfSparqlQueryTypeKey), detectQueryType(query)),
	)

	queryURL := fmt.Sprintf("%s/query", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, strings.NewReader(query))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/sparql-query")
	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("oxigraph returned HTTP %d: %s", resp.StatusCode, string(body))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	span.SetStatus(codes.Ok, "")
	return string(body), nil
}

// detectQueryType returns the SPARQL query type (SELECT, CONSTRUCT, ASK, INSERT, DESCRIBE).
func detectQueryType(query string) string {
	q := strings.TrimSpace(strings.ToUpper(query))
	for _, prefix := range []string{"PREFIX", "#"} {
		for strings.HasPrefix(q, prefix) {
			if idx := strings.Index(q, "\n"); idx >= 0 {
				q = strings.TrimSpace(q[idx+1:])
			} else {
				break
			}
		}
	}
	for _, qt := range []string{"SELECT", "CONSTRUCT", "ASK", "DESCRIBE", "INSERT", "DELETE"} {
		if strings.HasPrefix(q, qt) {
			return qt
		}
	}
	return "SELECT"
}

// executeSPARQL sends a SPARQL query to the /query endpoint.
// Supports both SELECT and INSERT queries.
func (c *Client) executeSPARQL(ctx context.Context, query string) error {
	url := fmt.Sprintf("%s/query", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(query))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/sparql-query")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Success: 200 OK or 204 No Content
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	// Error response
	return fmt.Errorf("oxigraph returned HTTP %d: %s", resp.StatusCode, string(body))
}

// Health checks if Oxigraph is responding.
func (c *Client) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/query", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader("SELECT ?s WHERE { ?s ?p ?o . } LIMIT 1"))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/sparql-query")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("oxigraph unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("oxigraph returned HTTP %d", resp.StatusCode)
	}

	return nil
}

// GetBaseURL returns the configured Oxigraph base URL.
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// SetTimeout updates the client timeout duration.
func (c *Client) SetTimeout(d time.Duration) {
	c.timeout = d
	c.httpClient.Timeout = d
}

// SetLogger updates the logger instance.
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logger
}
