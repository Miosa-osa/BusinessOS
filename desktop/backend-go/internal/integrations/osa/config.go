package osa

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

// Config holds configuration for the OSA client
type Config struct {
	// BaseURL is the base URL for the OSA API (e.g., "http://localhost:8089")
	BaseURL string

	// SharedSecret is the JWT secret shared between BusinessOS and OSA
	SharedSecret string

	// Timeout for API requests
	Timeout time.Duration

	// MaxRetries for failed requests
	MaxRetries int

	// RetryDelay between retries
	RetryDelay time.Duration
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "http://localhost:8089",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("OSA base URL is required")
	}

	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("OSA base URL must be an HTTP or HTTPS URL")
	}
	ip := net.ParseIP(u.Hostname())
	loopback := u.Hostname() == "localhost" || (ip != nil && ip.IsLoopback())
	if c.SharedSecret == "" && !loopback {
		return fmt.Errorf("OSA shared secret is required for remote runtimes")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	return nil
}
