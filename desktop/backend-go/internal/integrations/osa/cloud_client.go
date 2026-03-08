package osa

import (
	"fmt"

	osasdk "github.com/Miosa-osa/sdk-go"
)

// NewCloudClient creates an OSA client backed by MIOSA Cloud (api.miosa.ai).
// It uses osasdk.NewCloudClient from the sdk-go module and wraps it in the
// same thin Client struct as NewClient (local mode), so all callers are
// mode-agnostic.
//
// The cloudConfig must supply a non-empty APIKey. BaseURL defaults to
// https://api.miosa.ai if left blank.
func NewCloudClient(cloudConfig *CloudConfig) (*Client, error) {
	if err := cloudConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cloud config: %w", err)
	}

	baseURL := cloudConfig.BaseURL
	if baseURL == "" {
		baseURL = "https://api.miosa.ai"
	}

	// CloudConfig does not carry a Resilience field (only LocalConfig does);
	// BOS wraps cloud calls in its own ResilientClient layer instead.
	sdkClient, err := osasdk.NewCloudClient(osasdk.CloudConfig{
		APIKey:  cloudConfig.APIKey,
		BaseURL: baseURL,
		Timeout: cloudConfig.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud SDK client: %w", err)
	}

	// Wrap in the same Config struct so the rest of BOS is unaware of mode.
	localEquivalent := &Config{
		BaseURL:      baseURL,
		SharedSecret: cloudConfig.APIKey, // used only in error messages for local mode
		Timeout:      cloudConfig.Timeout,
		MaxRetries:   cloudConfig.MaxRetries,
		RetryDelay:   cloudConfig.RetryDelay,
	}

	return &Client{
		config: localEquivalent,
		sdk:    sdkClient,
	}, nil
}
