package miosa

import (
	"context"
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

// CreditBalance is returned by GET /api/v1/credits/balance.
type CreditBalance struct {
	Balance        int    `json:"balance_credits"`
	LifetimeEarned int    `json:"lifetime_earned"`
	LifetimeSpent  int    `json:"lifetime_spent"`
	TenantID       string `json:"tenant_id"`
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

// StreamToken is returned by POST /api/v1/computers/:id/stream-token.
type StreamToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// ComputeClient calls the MIOSA platform API for computer provisioning and
// lifecycle management. Open-source stub — all methods return empty/error results.
type ComputeClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewComputeClient constructs a ComputeClient.
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

// ListComputers returns an empty list (open-source stub).
func (c *ComputeClient) ListComputers(_ context.Context) ([]MIOSAComputer, error) {
	return nil, nil
}

// GetComputer returns nil (open-source stub).
func (c *ComputeClient) GetComputer(_ context.Context, _ string) (*MIOSAComputer, error) {
	return nil, nil
}

// CreateComputer returns nil (open-source stub).
func (c *ComputeClient) CreateComputer(_ context.Context, _, _, _ string) (*MIOSAComputer, error) {
	return nil, nil
}

// DeleteComputer is a no-op (open-source stub).
func (c *ComputeClient) DeleteComputer(_ context.Context, _ string) error {
	return nil
}

// StartComputer is a no-op (open-source stub).
func (c *ComputeClient) StartComputer(_ context.Context, _ string) error {
	return nil
}

// StopComputer is a no-op (open-source stub).
func (c *ComputeClient) StopComputer(_ context.Context, _ string) error {
	return nil
}

// ExecInVM returns empty output (open-source stub).
func (c *ComputeClient) ExecInVM(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

// GetCreditBalance returns nil (open-source stub).
func (c *ComputeClient) GetCreditBalance(_ context.Context) (*CreditBalance, error) {
	return nil, nil
}

// GetCurrentUser returns nil (open-source stub).
func (c *ComputeClient) GetCurrentUser(_ context.Context) (*UserInfo, error) {
	return nil, nil
}

// GetStreamToken returns nil (open-source stub).
func (c *ComputeClient) GetStreamToken(_ context.Context, _ string) (*StreamToken, error) {
	return nil, nil
}

// CreateTerminalSession returns nil (open-source stub).
func (c *ComputeClient) CreateTerminalSession(_ context.Context, _ string, _, _ int) (*TerminalSession, error) {
	return nil, nil
}
