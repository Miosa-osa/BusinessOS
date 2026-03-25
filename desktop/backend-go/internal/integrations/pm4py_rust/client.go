package pm4py_rust

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Client wraps HTTP communication with pm4py-rust service
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
	retries    int
	timeout    time.Duration
}

// NewClient creates a new pm4py-rust HTTP client
func NewClient(baseURL string) *Client {
	return NewClientWithConfig(baseURL, ClientConfig{
		Timeout: 30 * time.Second,
		Retries: 3,
	})
}

// ClientConfig configures the pm4py-rust client
type ClientConfig struct {
	Timeout time.Duration
	Retries int
}

// NewClientWithConfig creates a new client with custom configuration
func NewClientWithConfig(baseURL string, config ClientConfig) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	retries := config.Retries
	if retries == 0 {
		retries = 3
	}

	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     100,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: timeout,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     slog.Default(),
		retries:    retries,
		timeout:    timeout,
	}
}

// HealthCheckRequest checks if pm4py-rust service is available
func (c *Client) HealthCheckRequest(ctx context.Context) (*HealthResponse, error) {
	resp := &HealthResponse{}
	_, err := c.doWithRetry(ctx, "GET", "/api/health", nil, resp)
	return resp, err
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// DiscoveryRequest sends an event log for Petri net discovery
func (c *Client) DiscoveryRequest(ctx context.Context, req *DiscoveryRequest) (*DiscoveryResponse, error) {
	resp := &DiscoveryResponse{}
	_, err := c.doWithRetry(ctx, "POST", "/api/discovery/alpha", req, resp)
	return resp, err
}

// DiscoveryRequest represents the discovery API request
type DiscoveryRequest struct {
	EventLog json.RawMessage `json:"event_log"`
	Variant  string          `json:"variant,omitempty"`
}

// DiscoveryResponse contains the discovered Petri net
type DiscoveryResponse struct {
	PetriNet      PetriNetJSON `json:"petri_net"`
	Algorithm     string       `json:"algorithm"`
	ExecutionTime uint64       `json:"execution_time_ms"`
	EventCount    int          `json:"event_count"`
	TraceCount    int          `json:"trace_count"`
}

// PetriNetJSON represents a Petri net in JSON format
type PetriNetJSON struct {
	Places       []PlaceJSON      `json:"places"`
	Transitions  []TransitionJSON `json:"transitions"`
	Arcs         []ArcJSON        `json:"arcs"`
	InitialPlace *string          `json:"initial_place"`
	FinalPlace   *string          `json:"final_place"`
}

// PlaceJSON represents a Petri net place
type PlaceJSON struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	InitialMarking  int    `json:"initial_marking"`
	FinalMarking    *int   `json:"final_marking,omitempty"`
}

// TransitionJSON represents a Petri net transition
type TransitionJSON struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Label *string `json:"label,omitempty"`
}

// ArcJSON represents a Petri net arc
type ArcJSON struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Weight int    `json:"weight"`
}

// ConformanceRequest checks conformance of an event log against a Petri net
func (c *Client) ConformanceRequest(ctx context.Context, req *ConformanceRequest) (*ConformanceResponse, error) {
	resp := &ConformanceResponse{}
	_, err := c.doWithRetry(ctx, "POST", "/api/conformance/token-replay", req, resp)
	return resp, err
}

// ConformanceRequest represents the conformance checking API request
type ConformanceRequest struct {
	EventLog json.RawMessage `json:"event_log"`
	PetriNet PetriNetJSON    `json:"petri_net"`
	Method   string          `json:"method,omitempty"`
}

// ConformanceResponse contains conformance checking results
type ConformanceResponse struct {
	IsConformant  bool   `json:"is_conformant"`
	Fitness       float64 `json:"fitness"`
	Precision     float64 `json:"precision"`
	Generalization float64 `json:"generalization"`
	Method        string  `json:"method"`
	ExecutionTime uint64  `json:"execution_time_ms"`
}

// StatisticsRequest analyzes an event log for statistics
func (c *Client) StatisticsRequest(ctx context.Context, req *StatisticsRequest) (*StatisticsResponse, error) {
	resp := &StatisticsResponse{}
	_, err := c.doWithRetry(ctx, "POST", "/api/statistics", req, resp)
	return resp, err
}

// StatisticsRequest represents the statistics API request
type StatisticsRequest struct {
	EventLog              json.RawMessage `json:"event_log"`
	IncludeVariants       bool            `json:"include_variants,omitempty"`
	IncludeResourceMetrics bool           `json:"include_resource_metrics,omitempty"`
	IncludeBottlenecks    bool            `json:"include_bottlenecks,omitempty"`
}

// StatisticsResponse contains statistical analysis results
type StatisticsResponse struct {
	TraceCount           int                      `json:"trace_count"`
	EventCount           int                      `json:"event_count"`
	UniqueActivities     int                      `json:"unique_activities"`
	ActivityFrequencies  map[string]int           `json:"activity_frequencies,omitempty"`
	VariantCount         int                      `json:"variant_count"`
	VariantFrequencies   map[string]int           `json:"variant_frequencies,omitempty"`
	BottleneckActivities []string                 `json:"bottleneck_activities,omitempty"`
	ResourceCount        *int                     `json:"resource_count,omitempty"`
	ExecutionTime        uint64                   `json:"execution_time_ms"`
}

// APIError represents an error response from pm4py-rust
type APIError struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"status"`
}

// doWithRetry performs an HTTP request with exponential backoff retry logic
func (c *Client) doWithRetry(ctx context.Context, method, path string, body interface{}, result interface{}) (interface{}, error) {
	var lastErr error
	backoff := time.Millisecond * 100

	for attempt := 0; attempt <= c.retries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		if attempt > 0 {
			c.logger.Debug("retrying pm4py-rust request",
				"attempt", attempt,
				"method", method,
				"path", path,
				"backoff_ms", backoff.Milliseconds(),
				"last_error", lastErr.Error(),
			)
			select {
			case <-time.After(backoff):
				backoff *= 2
				if backoff > 10*time.Second {
					backoff = 10 * time.Second
				}
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			}
		}

		resp, err := c.do(ctx, method, path, body)
		if err != nil {
			lastErr = err
			if isRetryable(err) {
				continue
			}
			return nil, err
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			if attempt < c.retries {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()
			return nil, lastErr
		}

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var apiErr APIError
			if err := json.Unmarshal(body, &apiErr); err == nil {
				return nil, fmt.Errorf("api error: %s (status %d): %s", apiErr.Error, resp.StatusCode, apiErr.Details)
			}
			return nil, fmt.Errorf("http error: %d: %s", resp.StatusCode, string(body))
		}

		if resp.StatusCode == http.StatusNoContent {
			resp.Body.Close()
			return result, nil
		}

		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		c.logger.Debug("pm4py-rust request succeeded",
			"method", method,
			"path", path,
			"status", resp.StatusCode,
			"attempts", attempt+1,
		)
		return result, nil
	}

	return nil, fmt.Errorf("max retries exceeded after %d attempts: %w", c.retries+1, lastErr)
}

// do performs a single HTTP request
func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add trace ID if available in context
	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		req.Header.Set("X-Trace-ID", traceID)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "businessos-pm4py-client/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// isRetryable determines if an error should trigger a retry
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Timeout and context errors are retryable
	switch err.(type) {
	case net.Error:
		// Network errors like connection refused, timeout
		return true
	}

	// Check error string for common retryable patterns
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"EOF",
		"timeout",
		"temporary failure",
		"net/http",
	}

	for _, pattern := range retryablePatterns {
		if bytes.Contains([]byte(errStr), []byte(pattern)) {
			return true
		}
	}

	return false
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// SetLogger sets the logger for the client
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logger
}
