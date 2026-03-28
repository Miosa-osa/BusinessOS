package yawlv6

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rhl/businessos-backend/internal/circuitbreaker"
)

// Client communicates with the YAWLv6 engine over HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
	breaker    *circuitbreaker.CircuitBreaker
}

// ConformanceResult holds the result of a YAWL conformance check.
type ConformanceResult struct {
	Fitness    float64  `json:"fitness"`
	Violations []string `json:"violations"`
	IsSound    bool     `json:"is_sound"`
}

// NewClient creates a Client using YAWLV6_URL env var (default: http://localhost:8080).
func NewClient() *Client {
	baseURL := os.Getenv("YAWLV6_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	defaultConfig := circuitbreaker.Config{
		MaxAttempts:      3,
		BaseDelay:        100 * time.Millisecond,
		MaxDelay:         5 * time.Second,
		TimeoutDuration:  10 * time.Second,
		CooldownPeriod:   30 * time.Second,
		HalfOpenMaxCalls: 3,
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		breaker:    circuitbreaker.NewCircuitBreaker(defaultConfig),
	}
}

// Health calls GET /health.jsp and returns an error if the engine is unreachable or unhealthy.
// Wrapped with circuit breaker to prevent cascading failures.
func (c *Client) Health(ctx context.Context) error {
	err := c.breaker.Execute(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health.jsp", nil)
		if err != nil {
			return err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("yawl health check failed: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("yawl health check returned %d", resp.StatusCode)
		}
		return nil
	})

	// Circuit breaker errors are already wrapped
	return err
}

// CheckConformance sends a YAWL spec XML and event log JSON to the engine and
// returns a ConformanceResult.
// Wrapped with circuit breaker to prevent cascading failures.
func (c *Client) CheckConformance(ctx context.Context, specXML string, eventLogJSON []byte) (*ConformanceResult, error) {
	var result *ConformanceResult
	err := c.breaker.Execute(ctx, func() error {
		body, _ := json.Marshal(map[string]interface{}{
			"spec":      specXML,
			"event_log": json.RawMessage(eventLogJSON),
		})
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/process-mining/conformance", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("yawl conformance check failed: %w", err)
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		var conformanceResult ConformanceResult
		if err := json.Unmarshal(data, &conformanceResult); err != nil {
			return fmt.Errorf("yawl response parse failed: %w", err)
		}
		result = &conformanceResult
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// wcpCategories is the ordered list of subdirectories under wcp-patterns/.
// The order matches the Elixir SpecLibrary for consistent discovery.
var wcpCategories = []string{
	"basic", "branching", "iteration", "cancellation", "multiinstance",
	"state", "structural", "resource", "termination", "trigger",
}

// LoadSpec reads a WCP pattern spec XML file from the local yawlv6 exampleSpecs directory.
// Dynamically scans all wcp-patterns/* subdirectories rather than using a static mapping,
// so it works for all 43 WCP patterns (not just the original 5).
//
// Accepted pattern ID formats: "WCP-1", "WCP1", "WCP01" (case-insensitive).
// The directory is controlled by YAWLV6_SPECS_PATH env var (default: ~/yawlv6/exampleSpecs).
func (c *Client) LoadSpec(patternID string) (string, error) {
	specsPath := os.Getenv("YAWLV6_SPECS_PATH")
	if specsPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		specsPath = filepath.Join(home, "chatmangpt", "yawlv6", "exampleSpecs")
	}

	normalized := normalizePatternID(patternID)

	for _, cat := range wcpCategories {
		catDir := filepath.Join(specsPath, "wcp-patterns", cat)
		entries, err := os.ReadDir(catDir)
		if err != nil {
			// Directory may not exist — skip and continue.
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(strings.ToUpper(name), normalized) && strings.HasSuffix(name, ".xml") {
				data, err := os.ReadFile(filepath.Join(catDir, name))
				if err != nil {
					return "", fmt.Errorf("cannot read spec %q: %w", name, err)
				}
				return string(data), nil
			}
		}
	}

	return "", fmt.Errorf("spec %q not found in %s", patternID, specsPath)
}

// PatternEntry describes one WCP pattern spec found on disk.
type PatternEntry struct {
	ID       string `json:"id"`       // e.g. "WCP-1"
	Name     string `json:"name"`     // e.g. "Sequence"
	Category string `json:"category"` // e.g. "basic"
	Path     string `json:"path"`     // absolute path to XML file
}

// RealDataEntry describes one real-world process spec dataset.
type RealDataEntry struct {
	Name string `json:"name"` // canonical name, e.g. "order-management"
	Path string `json:"path"` // absolute path to XML file
}

// realDataNames maps canonical dataset names to their filename stems (case-insensitive prefix).
var realDataNames = map[string]string{
	"order-management":         "OrderManagement",
	"repair-process":           "RepairProcess",
	"traffic-fine-management":  "TrafficFineManagement",
}

// specsBasePath returns the resolved exampleSpecs base directory.
func specsBasePath() (string, error) {
	if p := os.Getenv("YAWLV6_SPECS_PATH"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, "chatmangpt", "yawlv6", "exampleSpecs"), nil
}

// ListPatterns scans all wcp-patterns/* subdirectories and returns a sorted
// slice of all WCP pattern specs found on disk.
func (c *Client) ListPatterns() ([]PatternEntry, error) {
	base, err := specsBasePath()
	if err != nil {
		return nil, err
	}
	patternsDir := filepath.Join(base, "wcp-patterns")

	var entries []PatternEntry
	for _, cat := range wcpCategories {
		catDir := filepath.Join(patternsDir, cat)
		files, err := os.ReadDir(catDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			name := f.Name()
			if !strings.HasSuffix(name, ".xml") {
				continue
			}
			id, label := parsePatternFilename(name)
			entries = append(entries, PatternEntry{
				ID:       id,
				Name:     label,
				Category: cat,
				Path:     filepath.Join(catDir, name),
			})
		}
	}
	return entries, nil
}

// ListRealData returns all known real-data specs found on disk under real-data/.
func (c *Client) ListRealData() ([]RealDataEntry, error) {
	base, err := specsBasePath()
	if err != nil {
		return nil, err
	}
	realDir := filepath.Join(base, "real-data")

	files, err := os.ReadDir(realDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read real-data directory: %w", err)
	}
	fileNames := make([]string, 0, len(files))
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}

	var entries []RealDataEntry
	for canonical, stem := range realDataNames {
		for _, fn := range fileNames {
			if strings.HasPrefix(strings.ToLower(fn), strings.ToLower(stem)) && strings.HasSuffix(fn, ".xml") {
				entries = append(entries, RealDataEntry{
					Name: canonical,
					Path: filepath.Join(realDir, fn),
				})
				break
			}
		}
	}
	return entries, nil
}

// LoadRealData reads a real-world process spec XML by canonical dataset name.
// Known names: "order-management", "repair-process", "traffic-fine-management".
func (c *Client) LoadRealData(name string) (string, error) {
	base, err := specsBasePath()
	if err != nil {
		return "", err
	}
	stem, ok := realDataNames[strings.ToLower(name)]
	if !ok {
		return "", fmt.Errorf("unknown real-data dataset %q", name)
	}
	realDir := filepath.Join(base, "real-data")
	files, err := os.ReadDir(realDir)
	if err != nil {
		return "", fmt.Errorf("cannot read real-data directory: %w", err)
	}
	for _, f := range files {
		fn := f.Name()
		if strings.HasPrefix(strings.ToLower(fn), strings.ToLower(stem)) && strings.HasSuffix(fn, ".xml") {
			data, err := os.ReadFile(filepath.Join(realDir, fn))
			if err != nil {
				return "", fmt.Errorf("cannot read spec %q: %w", fn, err)
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("real-data spec %q not found in %s", name, realDir)
}

// parsePatternFilename converts "WCP01_Sequence.xml" → ("WCP-1", "Sequence").
func parsePatternFilename(filename string) (id, name string) {
	base := strings.TrimSuffix(filename, ".xml")
	re := regexp.MustCompile(`(?i)^WCP(\d+)_(.+)$`)
	m := re.FindStringSubmatch(base)
	if len(m) != 3 {
		return base, base
	}
	n := strings.TrimLeft(m[1], "0")
	if n == "" {
		n = "0"
	}
	label := regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(
		strings.ReplaceAll(m[2], "_", " "), "$1 $2")
	return "WCP-" + n, label
}

// normalizePatternID converts any supported pattern ID format to the zero-padded
// filename prefix used by the exampleSpecs naming convention.
// Examples: "WCP-1" → "WCP01", "WCP1" → "WCP01", "1" → "WCP01"
func normalizePatternID(id string) string {
	reNonAlnum := regexp.MustCompile(`[^a-zA-Z0-9]`)
	s := strings.ToUpper(reNonAlnum.ReplaceAllString(id, ""))

	reWCP := regexp.MustCompile(`^WCP(\d+)$`)
	if m := reWCP.FindStringSubmatch(s); len(m) == 2 {
		num := m[1]
		if len(num) == 1 {
			num = "0" + num
		}
		return "WCP" + num
	}

	// Bare number like "1" or "01"
	reNum := regexp.MustCompile(`^(\d+)$`)
	if m := reNum.FindStringSubmatch(s); len(m) == 2 {
		num := m[1]
		if len(num) == 1 {
			num = "0" + num
		}
		return "WCP" + num
	}

	return s
}
