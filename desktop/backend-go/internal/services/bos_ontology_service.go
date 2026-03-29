package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BosOntologyService wraps the bos CLI for ontology operations.
// All RDF data generation goes through bos — the Go backend has zero
// knowledge of the BusinessOS domain model.
type BosOntologyService struct {
	bosPath    string
	dbURL      string
	mapping    string
	queriesDir string
	timeout    time.Duration
}

// NewBosOntologyService creates a new service instance.
// bosPath should point to the bos binary (e.g., ./bos/target/release/bos).
// dbURL is the PostgreSQL connection string.
// mapping is the path to ontology-mappings.json.
func NewBosOntologyService(bosPath, dbURL, mapping string) *BosOntologyService {
	return &BosOntologyService{
		bosPath:    bosPath,
		dbURL:      dbURL,
		mapping:    mapping,
		queriesDir: "desktop/backend-go/bos/queries",
		timeout:    30 * time.Second,
	}
}

// SetQueriesDir overrides the default queries directory.
func (s *BosOntologyService) SetQueriesDir(dir string) {
	s.queriesDir = dir
}

// ExecuteConstruct runs SPARQL CONSTRUCT via bos CLI with PostgreSQL data.
// Returns N-Triples format RDF.
func (s *BosOntologyService) ExecuteConstruct(ctx context.Context, table string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	args := []string{
		"ontology", "execute",
		"--mapping", s.mapping,
		"--database", s.dbURL,
		"--table", table,
		"--format", "nt",
	}

	cmd := exec.CommandContext(cmdCtx, s.bosPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("bos execute failed for table %s: %w, output: %s", table, err, string(output))
	}

	return string(output), nil
}

// ExecuteAll runs CONSTRUCT for all mapped tables, returning combined RDF.
func (s *BosOntologyService) ExecuteAll(ctx context.Context, format string) (string, error) {
	if format == "" {
		format = "nt"
	}

	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout*3) // longer timeout for all tables
	defer cancel()

	args := []string{
		"ontology", "execute",
		"--mapping", s.mapping,
		"--database", s.dbURL,
		"--format", format,
	}

	cmd := exec.CommandContext(cmdCtx, s.bosPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("bos execute all failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// GetConstructQuery returns the CONSTRUCT query text for a table.
func (s *BosOntologyService) GetConstructQuery(ctx context.Context, table string) (string, error) {
	queryPath := filepath.Join(s.queriesDir, fmt.Sprintf("%s.rq", table))

	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "cat", queryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("query not found for table %s: %w", table, err)
	}

	return string(output), nil
}

// ListQueries returns all available CONSTRUCT query files.
func (s *BosOntologyService) ListQueries(ctx context.Context) ([]string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "ls", s.queriesDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If queries dir doesn't exist, return known tables
		return []string{
			"projects", "tasks", "clients", "team_members",
			"contexts", "conversations", "artifacts",
		}, nil
	}

	var tables []string
	for _, f := range strings.Split(string(output), "\n") {
		f = strings.TrimSpace(f)
		if strings.HasSuffix(f, ".rq") {
			tables = append(tables, strings.TrimSuffix(f, ".rq"))
		}
	}

	if len(tables) == 0 {
		return []string{
			"projects", "tasks", "clients", "team_members",
			"contexts", "conversations", "artifacts",
		}, nil
	}

	return tables, nil
}

// ExecuteSelect runs a SPARQL SELECT query via bos CLI and returns JSON results.
// The bos CLI is invoked as: bos ontology query --query <sparql>
// A 30-second context deadline is applied; the raw JSON output is decoded and returned.
func (s *BosOntologyService) ExecuteSelect(ctx context.Context, query string) (map[string]interface{}, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	args := []string{
		"ontology", "query",
		"--query", query,
	}

	cmd := exec.CommandContext(cmdCtx, s.bosPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bos query failed: %w, output: %s", err, string(output))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parse bos query output: %w", err)
	}

	return result, nil
}

// GenerateQueries runs bos ontology construct to generate .rq files.
func (s *BosOntologyService) GenerateQueries(ctx context.Context, outputDir string) (int, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	args := []string{
		"ontology", "construct",
		"--mapping", s.mapping,
		"--output", outputDir,
	}

	cmd := exec.CommandContext(cmdCtx, s.bosPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("bos construct failed: %w, output: %s", err, string(output))
	}

	// Count generated .rq files
	lsCmd := exec.CommandContext(cmdCtx, "ls", outputDir)
	lsOutput, _ := lsCmd.CombinedOutput()
	count := 0
	for _, f := range strings.Split(string(lsOutput), "\n") {
		if strings.HasSuffix(f, ".rq") {
			count++
		}
	}

	return count, nil
}
