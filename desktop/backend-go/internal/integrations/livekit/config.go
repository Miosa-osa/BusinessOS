package livekit

import (
	"fmt"
	"time"
)

// Config holds configuration for the LiveKit client
type Config struct {
	// URL is the LiveKit server URL (e.g., "wss://your-project.livekit.cloud")
	URL string

	// APIKey is the LiveKit API key
	APIKey string

	// APISecret is the LiveKit API secret for signing JWTs
	APISecret string

	// AgentIdentity is the default identity for agents
	AgentIdentity string

	// TokenTTL is the time-to-live for generated tokens
	TokenTTL time.Duration

	// Timeout for API operations
	Timeout time.Duration
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		AgentIdentity: "businessos-agent",
		TokenTTL:      24 * time.Hour, // 24 hours
		Timeout:       30 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("LiveKit URL is required")
	}

	if c.APIKey == "" {
		return fmt.Errorf("LiveKit API key is required")
	}

	if c.APISecret == "" {
		return fmt.Errorf("LiveKit API secret is required")
	}

	if c.TokenTTL <= 0 {
		return fmt.Errorf("token TTL must be positive")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}
