package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// A2A Protocol Types
// ---------------------------------------------------------------------------

// AgentCard represents an A2A agent's identity and capabilities.
// Follows the Google A2A specification for agent discovery.
type AgentCard struct {
	Name         string       `json:"name"`
	DisplayName  string       `json:"display_name,omitempty"`
	Description  string       `json:"description"`
	Version      string       `json:"version"`
	URL          string       `json:"url"`
	Capabilities []string     `json:"capabilities"`
	Skills       []AgentSkill `json:"skills,omitempty"`
}

// AgentSkill describes a capability offered by an A2A agent.
type AgentSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

// Task represents an A2A task with its lifecycle state.
type Task struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id,omitempty"`
	Input     map[string]any `json:"input"`
	Status    string         `json:"status"`
	Output    map[string]any `json:"output,omitempty"`
}

// A2ATool represents a tool exposed by an A2A agent.
type A2ATool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
}

// ---------------------------------------------------------------------------
// A2A Client
// ---------------------------------------------------------------------------

// AgentConnection tracks a discovered and cached A2A agent.
type AgentConnection struct {
	URL      string
	Card     *AgentCard
	LastSeen time.Time
}

// A2AClient provides A2A (Agent-to-Agent) protocol communication.
// It follows Google's A2A specification for agent discovery, task
// submission, and tool execution over HTTP.
type A2AClient struct {
	httpClient *http.Client
	agents     map[string]*AgentConnection
	mu         sync.RWMutex
}

// NewA2AClient creates a new A2A client with sensible defaults.
func NewA2AClient() *A2AClient {
	return &A2AClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     5 * time.Minute,
				DisableCompression:  false,
				MaxConnsPerHost:     5,
				MaxIdleConnsPerHost: 5,
			},
		},
		agents: make(map[string]*AgentConnection),
	}
}

// ---------------------------------------------------------------------------
// Agent Discovery
// ---------------------------------------------------------------------------

// DiscoverAgent fetches the agent card from a remote A2A agent URL.
// The agent URL is expected to serve an AgentCard at its root (GET /).
func (c *A2AClient) DiscoverAgent(ctx context.Context, agentURL string) (*AgentCard, error) {
	agentURL = strings.TrimRight(agentURL, "/")

	if err := ValidateA2AAgentURL(agentURL); err != nil {
		return nil, fmt.Errorf("invalid agent URL: %w", err)
	}

	// A2A spec: GET the agent URL to retrieve the AgentCard
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, agentURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach agent at %s: %w", agentURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
	}

	var card AgentCard
	if err := json.NewDecoder(io.LimitReader(resp.Body, 512*1024)).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to parse agent card: %w", err)
	}

	if card.Name == "" {
		return nil, fmt.Errorf("agent card missing required field: name")
	}

	// Cache the connection
	c.mu.Lock()
	c.agents[agentURL] = &AgentConnection{
		URL:      agentURL,
		Card:     &card,
		LastSeen: time.Now(),
	}
	c.mu.Unlock()

	slog.Info("A2A agent discovered", "name", card.Name, "url", agentURL)

	return &card, nil
}

// ---------------------------------------------------------------------------
// Task Execution
// ---------------------------------------------------------------------------

// CallAgent sends a message to an A2A agent and returns the resulting task.
// The agent is expected to expose a POST /tasks endpoint that accepts a JSON
// message and returns a Task.
func (c *A2AClient) CallAgent(ctx context.Context, agentURL string, message string) (*Task, error) {
	agentURL = strings.TrimRight(agentURL, "/")

	if err := ValidateA2AAgentURL(agentURL); err != nil {
		return nil, fmt.Errorf("invalid agent URL: %w", err)
	}

	taskURL := agentURL + "/tasks"

	payload := map[string]any{
		"message": message,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, taskURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create task request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call agent at %s: %w", taskURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
	}

	var task Task
	if err := json.NewDecoder(io.LimitReader(resp.Body, 512*1024)).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to parse task response: %w", err)
	}

	// Update last seen
	c.mu.Lock()
	if conn, ok := c.agents[agentURL]; ok {
		conn.LastSeen = time.Now()
	}
	c.mu.Unlock()

	slog.Info("A2A task created", "task_id", task.ID, "agent_url", agentURL, "status", task.Status)

	return &task, nil
}

// ---------------------------------------------------------------------------
// Tool Discovery and Execution
// ---------------------------------------------------------------------------

// GetAgentTools retrieves the list of tools exposed by an A2A agent.
// The agent is expected to expose a GET /tools endpoint.
func (c *A2AClient) GetAgentTools(ctx context.Context, agentURL string) ([]A2ATool, error) {
	agentURL = strings.TrimRight(agentURL, "/")

	if err := ValidateA2AAgentURL(agentURL); err != nil {
		return nil, fmt.Errorf("invalid agent URL: %w", err)
	}

	toolsURL := agentURL + "/tools"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, toolsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tools request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach agent tools at %s: %w", toolsURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Tools []A2ATool `json:"tools"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 512*1024)).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %w", err)
	}

	// Update last seen
	c.mu.Lock()
	if conn, ok := c.agents[agentURL]; ok {
		conn.LastSeen = time.Now()
	}
	c.mu.Unlock()

	return result.Tools, nil
}

// ExecuteAgentTool invokes a specific tool on an A2A agent by name.
// The agent is expected to expose a POST /tools/:name endpoint.
func (c *A2AClient) ExecuteAgentTool(ctx context.Context, agentURL string, toolName string, args map[string]any) (any, error) {
	agentURL = strings.TrimRight(agentURL, "/")

	if err := ValidateA2AAgentURL(agentURL); err != nil {
		return nil, fmt.Errorf("invalid agent URL: %w", err)
	}

	toolURL := fmt.Sprintf("%s/tools/%s", agentURL, url.PathEscape(toolName))

	bodyBytes, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool arguments: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, toolURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create tool execution request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool on agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("agent tool returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 512*1024)).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse tool result: %w", err)
	}

	// Update last seen
	c.mu.Lock()
	if conn, ok := c.agents[agentURL]; ok {
		conn.LastSeen = time.Now()
	}
	c.mu.Unlock()

	slog.Info("A2A tool executed", "tool", toolName, "agent_url", agentURL)

	return result, nil
}

// ---------------------------------------------------------------------------
// Connection Management
// ---------------------------------------------------------------------------

// ListConnectedAgents returns all cached agent connections.
func (c *A2AClient) ListConnectedAgents() []*AgentConnection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	agents := make([]*AgentConnection, 0, len(c.agents))
	for _, conn := range c.agents {
		agents = append(agents, conn)
	}
	return agents
}

// DisconnectAgent removes a cached agent connection by URL.
func (c *A2AClient) DisconnectAgent(agentURL string) error {
	agentURL = strings.TrimRight(agentURL, "/")

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.agents[agentURL]; !ok {
		return fmt.Errorf("agent not connected: %s", agentURL)
	}

	delete(c.agents, agentURL)
	slog.Info("A2A agent disconnected", "url", agentURL)
	return nil
}

// ---------------------------------------------------------------------------
// URL Validation (SSRF protection)
// ---------------------------------------------------------------------------

// ValidateA2AAgentURL checks that a URL is safe to connect to.
// Prevents SSRF attacks by blocking private/loopback/link-local addresses.
func ValidateA2AAgentURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only http and https protocols are allowed")
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("URL must have a hostname")
	}

	// Block cloud metadata endpoints
	if hostname == "169.254.169.254" || hostname == "metadata.google.internal" {
		return fmt.Errorf("cloud metadata endpoints are not allowed")
	}

	ips, err := net.LookupHost(hostname)
	if err != nil {
		return fmt.Errorf("cannot resolve hostname: %s", hostname)
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() {
			return fmt.Errorf("loopback addresses are not allowed")
		}
		if ip.IsPrivate() {
			return fmt.Errorf("private network addresses are not allowed")
		}
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("link-local addresses are not allowed")
		}
		if ip.IsUnspecified() {
			return fmt.Errorf("unspecified addresses are not allowed")
		}
	}

	return nil
}
