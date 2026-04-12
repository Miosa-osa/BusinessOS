package miosa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MIOSAComputer is the compute resource returned by the MIOSA platform API.
type MIOSAComputer struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Status       string   `json:"status"`
	TemplateType string   `json:"template_type"`
	Size         string   `json:"size"`
	VMID         string   `json:"vm_id"`
	DesktopURL   string   `json:"desktop_url"`
	IPAddress    string   `json:"ip_address"`
	VCPUs        int      `json:"vcpus"`
	MemoryMB     int      `json:"memory_mb"`
	DiskGB       int      `json:"disk_gb"`
	SelectedApps []string `json:"selected_apps"`
	Resolution   string   `json:"resolution"`
	TenantID     string   `json:"tenant_id"`
	OwnerUserID  string   `json:"owner_user_id"`
	InsertedAt   string   `json:"inserted_at"`
	UpdatedAt    string   `json:"updated_at"`
	CreatedAt    string   `json:"created_at"`
}

// TerminalSession is returned by the MIOSA terminal session endpoint.
type TerminalSession struct {
	SessionID  string `json:"session_id"`
	StreamAuth string `json:"stream_auth"`
	ExpiresAt  int64  `json:"expires_at"`
}

// createComputerRequest is the JSON body for POST /api/v1/computers.
type createComputerRequest struct {
	Name         string   `json:"name"`
	TemplateType string   `json:"template_type"`
	Size         string   `json:"size"`
	SelectedApps []string `json:"selected_apps"`
}

// listComputersResponse wraps the MIOSA list response.
type listComputersResponse struct {
	Computers []MIOSAComputer `json:"computers"`
	Total     int             `json:"total"`
}

// singleComputerResponse wraps the MIOSA single-computer response.
type singleComputerResponse struct {
	Computer MIOSAComputer `json:"computer"`
}

// ComputeClient calls the MIOSA platform API for computer provisioning and
// lifecycle management. All calls require Bearer token authentication.
type ComputeClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewComputeClient constructs a ComputeClient.
// baseURL defaults to http://localhost:4000 when empty.
// apiKey is the Bearer token sent on every request.
func NewComputeClient(baseURL, apiKey string) *ComputeClient {
	if baseURL == "" {
		baseURL = "http://localhost:4000"
	}
	return &ComputeClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateComputer calls POST /api/v1/computers to provision a new VM.
func (c *ComputeClient) CreateComputer(ctx context.Context, name, templateType, size string) (*MIOSAComputer, error) {
	body := createComputerRequest{
		Name:         name,
		TemplateType: templateType,
		Size:         size,
		SelectedApps: []string{},
	}

	var resp singleComputerResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/computers", body, &resp); err != nil {
		return nil, fmt.Errorf("miosa: create computer: %w", err)
	}
	return &resp.Computer, nil
}

// ListComputers calls GET /api/v1/computers and returns all computers for the
// authenticated user.
func (c *ComputeClient) ListComputers(ctx context.Context) ([]MIOSAComputer, error) {
	var resp listComputersResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/computers", nil, &resp); err != nil {
		return nil, fmt.Errorf("miosa: list computers: %w", err)
	}
	return resp.Computers, nil
}

// GetComputer calls GET /api/v1/computers/:id.
func (c *ComputeClient) GetComputer(ctx context.Context, id string) (*MIOSAComputer, error) {
	var resp singleComputerResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/computers/"+id, nil, &resp); err != nil {
		return nil, fmt.Errorf("miosa: get computer %s: %w", id, err)
	}
	return &resp.Computer, nil
}

// StopComputer calls POST /api/v1/computers/:id/stop.
func (c *ComputeClient) StopComputer(ctx context.Context, id string) error {
	if err := c.do(ctx, http.MethodPost, "/api/v1/computers/"+id+"/stop", nil, nil); err != nil {
		return fmt.Errorf("miosa: stop computer %s: %w", id, err)
	}
	return nil
}

// StartComputer calls POST /api/v1/computers/:id/start.
func (c *ComputeClient) StartComputer(ctx context.Context, id string) error {
	if err := c.do(ctx, http.MethodPost, "/api/v1/computers/"+id+"/start", nil, nil); err != nil {
		return fmt.Errorf("miosa: start computer %s: %w", id, err)
	}
	return nil
}

// CreditBalance is returned by GET /api/v1/credits/balance.
type CreditBalance struct {
	Balance        int    `json:"balance_credits"`
	LifetimeEarned int    `json:"lifetime_earned"`
	LifetimeSpent  int    `json:"lifetime_spent"`
	TenantID       string `json:"tenant_id"`
}

// GetCreditBalance calls GET /api/v1/credits/balance.
func (c *ComputeClient) GetCreditBalance(ctx context.Context) (*CreditBalance, error) {
	var resp CreditBalance
	if err := c.do(ctx, http.MethodGet, "/api/v1/credits/balance", nil, &resp); err != nil {
		return nil, fmt.Errorf("miosa: get credit balance: %w", err)
	}
	return &resp, nil
}

// UserInfo is returned by GET /api/v1/auth/me.
type UserInfo struct {
	User struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		Role        string `json:"role"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
	} `json:"user"`
}

// GetCurrentUser calls GET /api/v1/auth/me.
func (c *ComputeClient) GetCurrentUser(ctx context.Context) (*UserInfo, error) {
	var resp UserInfo
	if err := c.do(ctx, http.MethodGet, "/api/v1/auth/me", nil, &resp); err != nil {
		return nil, fmt.Errorf("miosa: get current user: %w", err)
	}
	return &resp, nil
}

// StreamToken is returned by POST /api/v1/computers/:id/stream-token.
type StreamToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// GetStreamToken calls POST /api/v1/computers/:id/stream-token to get
// an HMAC auth token for VNC desktop streaming.
func (c *ComputeClient) GetStreamToken(ctx context.Context, id string) (*StreamToken, error) {
	var resp StreamToken
	if err := c.do(ctx, http.MethodPost, "/api/v1/computers/"+id+"/stream-token", nil, &resp); err != nil {
		return nil, fmt.Errorf("miosa: get stream token for %s: %w", id, err)
	}
	return &resp, nil
}

// CreateTerminalSession calls POST /api/v1/computers/:id/terminal to create
// a PTY session inside the VM. Returns the session_id and stream_auth token
// needed to open a WebSocket connection to the terminal.
func (c *ComputeClient) CreateTerminalSession(ctx context.Context, id string, cols, rows int) (*TerminalSession, error) {
	body := map[string]any{"cols": cols, "rows": rows, "shell": "/bin/bash"}
	// The MIOSA API returns nested: { data: { data: { session_id, stream_auth } } }
	var raw map[string]any
	if err := c.do(ctx, http.MethodPost, "/api/v1/computers/"+id+"/terminal", body, &raw); err != nil {
		return nil, fmt.Errorf("miosa: create terminal session for %s: %w", id, err)
	}

	// Unwrap nested response
	inner := raw
	if d, ok := raw["data"].(map[string]any); ok {
		inner = d
		if dd, ok := d["data"].(map[string]any); ok {
			inner = dd
		}
	}

	session := &TerminalSession{}
	if sid, ok := inner["session_id"].(string); ok {
		session.SessionID = sid
	}
	if sa, ok := inner["stream_auth"].(string); ok {
		session.StreamAuth = sa
	}
	if ea, ok := inner["expires_at"].(float64); ok {
		session.ExpiresAt = int64(ea)
	}
	return session, nil
}

// DeleteComputer calls DELETE /api/v1/computers/:id.
func (c *ComputeClient) DeleteComputer(ctx context.Context, id string) error {
	if err := c.do(ctx, http.MethodDelete, "/api/v1/computers/"+id, nil, nil); err != nil {
		return fmt.Errorf("miosa: delete computer %s: %w", id, err)
	}
	return nil
}

// execResponse is the envd /process/exec response envelope.
type execResponse struct {
	Output   string `json:"output"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	// envd may also wrap it under "data"
	Data struct {
		Output   string `json:"output"`
		Stdout   string `json:"stdout"`
		ExitCode int    `json:"exit_code"`
	} `json:"data"`
}

// ExecInVM runs a shell command inside the VM via the compute-proxy envd endpoint.
// Requests go to https://{slug}.sandbox.miosa.ai/api/process/exec using the same
// Bearer token as other MIOSA API calls.
// Returns stdout+stderr combined, or an error if the command fails to execute.
func (c *ComputeClient) ExecInVM(ctx context.Context, slug, command string) (string, error) {
	if slug == "" {
		return "", fmt.Errorf("miosa: exec in vm: slug is empty")
	}

	proxyURL := "https://" + slug + ".sandbox.miosa.ai/api/process/exec"
	body, err := json.Marshal(map[string]any{
		"command": command,
		"timeout": 10,
	})
	if err != nil {
		return "", fmt.Errorf("miosa: exec marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("miosa: exec build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("miosa: exec http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("miosa: exec envd error %d: %s", resp.StatusCode, string(raw))
	}

	var result execResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("miosa: exec decode: %w", err)
	}

	// Unwrap: try output, then stdout, then nested data
	output := result.Output
	if output == "" {
		output = result.Stdout
	}
	if output == "" && result.Data.Output != "" {
		output = result.Data.Output
	}
	if output == "" && result.Data.Stdout != "" {
		output = result.Data.Stdout
	}
	return output, nil
}

// do executes an HTTP request against the MIOSA API.
// body is JSON-encoded when non-nil; result is JSON-decoded into out when non-nil.
func (c *ComputeClient) do(ctx context.Context, method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("miosa api error %d: %s", resp.StatusCode, string(raw))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
