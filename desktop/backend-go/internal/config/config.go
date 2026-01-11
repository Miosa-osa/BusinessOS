package config

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// Environment (development, production)
	Environment string `mapstructure:"ENVIRONMENT"`

	// Database
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	DatabaseRequired bool `mapstructure:"DATABASE_REQUIRED"`

	// Server
	ServerPort string `mapstructure:"SERVER_PORT"`
	BaseURL    string `mapstructure:"BASE_URL"`

	// JWT Auth (kept for compatibility, Better Auth handles auth)
	SecretKey                string `mapstructure:"SECRET_KEY"`
	Algorithm                string `mapstructure:"ALGORITHM"`
	AccessTokenExpireMinutes int    `mapstructure:"ACCESS_TOKEN_EXPIRE_MINUTES"`

	// AI Provider Configuration
	// Options: "ollama_cloud", "ollama_local", "anthropic", "groq"
	AIProvider string `mapstructure:"AI_PROVIDER"`

	// Ollama Local Configuration
	OllamaLocalURL string `mapstructure:"OLLAMA_LOCAL_URL"`

	// Ollama Cloud Configuration (api.ollama.com)
	OllamaCloudAPIKey string `mapstructure:"OLLAMA_CLOUD_API_KEY"`
	OllamaCloudModel  string `mapstructure:"OLLAMA_CLOUD_MODEL"`

	// Anthropic Configuration
	AnthropicAPIKey string `mapstructure:"ANTHROPIC_API_KEY"`
	AnthropicModel  string `mapstructure:"ANTHROPIC_MODEL"`

	// Groq Configuration
	GroqAPIKey string `mapstructure:"GROQ_API_KEY"`
	GroqModel  string `mapstructure:"GROQ_MODEL"`

	// Default Model (for local Ollama)
	DefaultModel string `mapstructure:"DEFAULT_MODEL"`

	// Legacy - kept for compatibility
	OllamaMode string `mapstructure:"OLLAMA_MODE"`

	// Redis
	RedisURL        string `mapstructure:"REDIS_URL"`
	RedisPassword   string `mapstructure:"REDIS_PASSWORD"`
	RedisTLSEnabled bool   `mapstructure:"REDIS_TLS_ENABLED"`

	// Security: HMAC secret for Redis key derivation (prevents token enumeration attacks)
	// CRITICAL: Must be set in production to a strong random value (min 32 bytes)
	// Used to hash session tokens before storing as Redis keys
	RedisKeyHMACSecret string `mapstructure:"REDIS_KEY_HMAC_SECRET"`

	// Security: Token encryption key for OAuth tokens stored in database
	// CRITICAL: Must be set in production - 32-byte base64-encoded key
	// Generate with: openssl rand -base64 32
	TokenEncryptionKey string `mapstructure:"TOKEN_ENCRYPTION_KEY"`

	// Supermemory
	SupermemoryAPIKey string `mapstructure:"SUPERMEMORY_API_KEY"`

	// Google OAuth
	GoogleClientID               string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret           string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURI            string `mapstructure:"GOOGLE_REDIRECT_URI"`              // For login flow
	GoogleIntegrationRedirectURI string `mapstructure:"GOOGLE_INTEGRATION_REDIRECT_URI"` // For calendar integration

	// Slack OAuth
	SlackClientID     string `mapstructure:"SLACK_CLIENT_ID"`
	SlackClientSecret string `mapstructure:"SLACK_CLIENT_SECRET"`
	SlackRedirectURI  string `mapstructure:"SLACK_REDIRECT_URI"`

	// Notion OAuth
	NotionClientID     string `mapstructure:"NOTION_CLIENT_ID"`
	NotionClientSecret string `mapstructure:"NOTION_CLIENT_SECRET"`
	NotionRedirectURI  string `mapstructure:"NOTION_REDIRECT_URI"`

	// HubSpot OAuth
	HubSpotClientID     string `mapstructure:"HUBSPOT_CLIENT_ID"`
	HubSpotClientSecret string `mapstructure:"HUBSPOT_CLIENT_SECRET"`
	HubSpotRedirectURI  string `mapstructure:"HUBSPOT_REDIRECT_URI"`

	// Linear OAuth
	LinearClientID     string `mapstructure:"LINEAR_CLIENT_ID"`
	LinearClientSecret string `mapstructure:"LINEAR_CLIENT_SECRET"`
	LinearRedirectURI  string `mapstructure:"LINEAR_REDIRECT_URI"`

	// ClickUp OAuth
	ClickUpClientID     string `mapstructure:"CLICKUP_CLIENT_ID"`
	ClickUpClientSecret string `mapstructure:"CLICKUP_CLIENT_SECRET"`
	ClickUpRedirectURI  string `mapstructure:"CLICKUP_REDIRECT_URI"`

	// Airtable OAuth
	AirtableClientID     string `mapstructure:"AIRTABLE_CLIENT_ID"`
	AirtableClientSecret string `mapstructure:"AIRTABLE_CLIENT_SECRET"`
	AirtableRedirectURI  string `mapstructure:"AIRTABLE_REDIRECT_URI"`

	// Microsoft 365 OAuth
	MicrosoftClientID     string `mapstructure:"MICROSOFT_CLIENT_ID"`
	MicrosoftClientSecret string `mapstructure:"MICROSOFT_CLIENT_SECRET"`
	MicrosoftRedirectURI  string `mapstructure:"MICROSOFT_REDIRECT_URI"`

	// Web Search Providers
	// Priority: Brave > Serper > Tavily > DuckDuckGo (fallback)
	BraveSearchAPIKey string `mapstructure:"BRAVE_SEARCH_API_KEY"` // Free: 2000 queries/month
	SerperAPIKey      string `mapstructure:"SERPER_API_KEY"`       // Google results via API
	TavilyAPIKey      string `mapstructure:"TAVILY_API_KEY"`       // AI-focused search API
	SearchProvider    string `mapstructure:"SEARCH_PROVIDER"`      // Override: brave, serper, tavily, duckduckgo, auto

	// CORS
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`

	// Feature Flags
	EnableLocalModels bool `mapstructure:"ENABLE_LOCAL_MODELS"`

	// Web Push (VAPID)
	VAPIDPublicKey  string `mapstructure:"VAPID_PUBLIC_KEY"`
	VAPIDPrivateKey string `mapstructure:"VAPID_PRIVATE_KEY"`
	VAPIDContact    string `mapstructure:"VAPID_CONTACT"` // Email: mailto:admin@example.com

	// Background Jobs (disabled by default)
	ConversationSummaryJobEnabled         bool `mapstructure:"CONVERSATION_SUMMARY_JOB_ENABLED"`
	ConversationSummaryJobIntervalMinutes int  `mapstructure:"CONVERSATION_SUMMARY_JOB_INTERVAL_MINUTES"`
	ConversationSummaryJobBatchSize       int  `mapstructure:"CONVERSATION_SUMMARY_JOB_BATCH_SIZE"`
	ConversationSummaryJobMaxMessages     int  `mapstructure:"CONVERSATION_SUMMARY_JOB_MAX_MESSAGES"`

	BehaviorPatternsJobEnabled         bool `mapstructure:"BEHAVIOR_PATTERNS_JOB_ENABLED"`
	BehaviorPatternsJobIntervalMinutes int  `mapstructure:"BEHAVIOR_PATTERNS_JOB_INTERVAL_MINUTES"`
	BehaviorPatternsJobUserBatchSize   int  `mapstructure:"BEHAVIOR_PATTERNS_JOB_USER_BATCH_SIZE"`

	AppProfilerSyncJobEnabled         bool `mapstructure:"APP_PROFILER_SYNC_JOB_ENABLED"`
	AppProfilerSyncJobIntervalMinutes int  `mapstructure:"APP_PROFILER_SYNC_JOB_INTERVAL_MINUTES"`
	AppProfilerSyncJobBatchSize       int  `mapstructure:"APP_PROFILER_SYNC_JOB_BATCH_SIZE"`
}

var AppConfig *Config

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Set defaults
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SERVER_PORT", "8001")
	viper.SetDefault("BASE_URL", "http://localhost:8001")
	viper.SetDefault("DATABASE_URL", "postgres://postgres:password@localhost:5432/business_os")
	viper.SetDefault("DATABASE_REQUIRED", true)
	viper.SetDefault("ENABLE_LOCAL_MODELS", true) // Disable in production
	viper.SetDefault("SECRET_KEY", "your-secret-key-change-this-in-production")
	viper.SetDefault("ALGORITHM", "HS256")
	viper.SetDefault("ACCESS_TOKEN_EXPIRE_MINUTES", 1440)

	// AI Provider - Default to Ollama Cloud for best out-of-box experience
	// Options: "ollama_cloud", "ollama_local", "anthropic", "groq"
	viper.SetDefault("AI_PROVIDER", "ollama_cloud")

	// Ollama Local
	viper.SetDefault("OLLAMA_LOCAL_URL", "http://localhost:11434")
	viper.SetDefault("DEFAULT_MODEL", "llama3.2:3b")

	// Ollama Cloud
	viper.SetDefault("OLLAMA_CLOUD_API_KEY", "")
	viper.SetDefault("OLLAMA_CLOUD_MODEL", "llama3.2")

	// Anthropic
	viper.SetDefault("ANTHROPIC_API_KEY", "")
	viper.SetDefault("ANTHROPIC_MODEL", "claude-sonnet-4-20250514")

	// Groq
	viper.SetDefault("GROQ_API_KEY", "")
	viper.SetDefault("GROQ_MODEL", "llama-3.3-70b-versatile")

	// Legacy
	viper.SetDefault("OLLAMA_MODE", "cloud")

	// Redis
	viper.SetDefault("REDIS_URL", "redis://localhost:6379/0")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_TLS_ENABLED", false)
	viper.SetDefault("REDIS_KEY_HMAC_SECRET", "") // CRITICAL: Set strong value in production (min 32 bytes)
	viper.SetDefault("TOKEN_ENCRYPTION_KEY", "") // CRITICAL: Set in production for OAuth token encryption

	// Other services
	viper.SetDefault("SUPERMEMORY_API_KEY", "")
	viper.SetDefault("GOOGLE_CLIENT_ID", "")
	viper.SetDefault("GOOGLE_CLIENT_SECRET", "")
	viper.SetDefault("GOOGLE_REDIRECT_URI", "http://localhost:8001/api/auth/google/callback/login")
	viper.SetDefault("GOOGLE_INTEGRATION_REDIRECT_URI", "http://localhost:8001/api/integrations/google/callback")
	viper.SetDefault("SLACK_CLIENT_ID", "")
	viper.SetDefault("SLACK_CLIENT_SECRET", "")
	viper.SetDefault("SLACK_REDIRECT_URI", "http://localhost:8001/api/integrations/slack/callback")
	viper.SetDefault("NOTION_CLIENT_ID", "")
	viper.SetDefault("NOTION_CLIENT_SECRET", "")
	viper.SetDefault("NOTION_REDIRECT_URI", "http://localhost:8001/api/integrations/notion/callback")
	viper.SetDefault("HUBSPOT_CLIENT_ID", "")
	viper.SetDefault("HUBSPOT_CLIENT_SECRET", "")
	viper.SetDefault("HUBSPOT_REDIRECT_URI", "http://localhost:8001/api/integrations/hubspot/callback")
	viper.SetDefault("LINEAR_CLIENT_ID", "")
	viper.SetDefault("LINEAR_CLIENT_SECRET", "")
	viper.SetDefault("LINEAR_REDIRECT_URI", "http://localhost:8001/api/integrations/linear/callback")

	// Web Search Providers
	viper.SetDefault("BRAVE_SEARCH_API_KEY", "")
	viper.SetDefault("SERPER_API_KEY", "")
	viper.SetDefault("TAVILY_API_KEY", "")
	viper.SetDefault("SEARCH_PROVIDER", "auto") // auto, brave, serper, tavily, duckduckgo

	// Web Push (VAPID) - Generate keys: npx web-push generate-vapid-keys
	viper.SetDefault("VAPID_PUBLIC_KEY", "")
	viper.SetDefault("VAPID_PRIVATE_KEY", "")
	viper.SetDefault("VAPID_CONTACT", "mailto:admin@businessos.app")

	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174,http://localhost:3000,app://localhost")

	// Background jobs
	viper.SetDefault("CONVERSATION_SUMMARY_JOB_ENABLED", false)
	viper.SetDefault("CONVERSATION_SUMMARY_JOB_INTERVAL_MINUTES", 30)
	viper.SetDefault("CONVERSATION_SUMMARY_JOB_BATCH_SIZE", 25)
	viper.SetDefault("CONVERSATION_SUMMARY_JOB_MAX_MESSAGES", 200)

	viper.SetDefault("APP_PROFILER_SYNC_JOB_ENABLED", false)
	viper.SetDefault("APP_PROFILER_SYNC_JOB_INTERVAL_MINUTES", 10)
	viper.SetDefault("APP_PROFILER_SYNC_JOB_BATCH_SIZE", 5)

	viper.SetDefault("BEHAVIOR_PATTERNS_JOB_ENABLED", false)
	viper.SetDefault("BEHAVIOR_PATTERNS_JOB_INTERVAL_MINUTES", 60)
	viper.SetDefault("BEHAVIOR_PATTERNS_JOB_USER_BATCH_SIZE", 50)

	// Read from environment variables first (takes priority in production)
	viper.AutomaticEnv()

	// Try to read config file (optional - for local development)
	// Ignore all errors - we can run without a config file in production
	_ = viper.ReadInConfig()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	// In local development it's common to have a globally-set DATABASE_URL (e.g. Supabase)
	// that should NOT override this repo's local .env. Viper's AutomaticEnv takes
	// precedence over config files, so we explicitly re-apply .env values in development.
	//
	// Production still uses environment variables as the source of truth.
	dotenvVars := readDotenvFile(".env")
	dotenvApplied := false
	if len(dotenvVars) > 0 {
		dotenvEnv := strings.ToLower(strings.TrimSpace(dotenvVars["ENVIRONMENT"]))
		if dotenvEnv == "" || dotenvEnv == "development" {
			applyDotenvOverrides(config, dotenvVars)
			dotenvApplied = true
		}
	}

	// Default CORS origins for development
	defaultOrigins := []string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:3000",
		"app://localhost",
	}

	// Clear AllowedOrigins set by Unmarshal (may have garbage) and parse from string
	config.AllowedOrigins = nil
	originsStr := viper.GetString("ALLOWED_ORIGINS")
	if dotenvApplied {
		if v, ok := dotenvVars["ALLOWED_ORIGINS"]; ok && strings.TrimSpace(v) != "" {
			originsStr = v
		}
	}
	if originsStr != "" {
		origins := strings.Split(originsStr, ",")
		for _, o := range origins {
			trimmed := strings.TrimSpace(o)
			// Only include valid origins (starts with http:// or https:// or is *)
			if trimmed == "*" || strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
				config.AllowedOrigins = append(config.AllowedOrigins, trimmed)
			}
		}
	}

	// Fallback to defaults if no valid origins found
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = defaultOrigins
	}

	// Convert DATABASE_URL from asyncpg format to pgx format if needed
	// Python uses: postgresql+asyncpg://...
	// Go pgx uses: postgres://... or postgresql://...
	config.DatabaseURL = strings.Replace(config.DatabaseURL, "postgresql+asyncpg://", "postgres://", 1)

	AppConfig = config
	return config, nil
}

func readDotenvFile(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		eq := strings.Index(line, "=")
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if key != "" {
			vars[key] = val
		}
	}
	return vars
}

func applyDotenvOverrides(cfg *Config, vars map[string]string) {
	if v := strings.TrimSpace(vars["ENVIRONMENT"]); v != "" {
		cfg.Environment = v
	}
	if v := strings.TrimSpace(vars["SERVER_PORT"]); v != "" {
		cfg.ServerPort = v
	}
	if v := strings.TrimSpace(vars["DATABASE_URL"]); v != "" {
		cfg.DatabaseURL = v
	}
	if v := strings.TrimSpace(vars["DATABASE_REQUIRED"]); v != "" {
		cfg.DatabaseRequired = strings.EqualFold(v, "true") || v == "1"
	}
	if v := strings.TrimSpace(vars["REDIS_URL"]); v != "" {
		cfg.RedisURL = v
	}
	if v := strings.TrimSpace(vars["REDIS_PASSWORD"]); v != "" {
		cfg.RedisPassword = v
	}
	if v := strings.TrimSpace(vars["REDIS_TLS_ENABLED"]); v != "" {
		cfg.RedisTLSEnabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := strings.TrimSpace(vars["REDIS_KEY_HMAC_SECRET"]); v != "" {
		cfg.RedisKeyHMACSecret = v
	}
}

// GetActiveProvider returns the currently configured AI provider
func (c *Config) GetActiveProvider() string {
	// Check if the configured provider has required credentials
	switch c.AIProvider {
	case "ollama_cloud":
		if c.OllamaCloudAPIKey != "" {
			return "ollama_cloud"
		}
		// Fallback to local if no cloud key
		return "ollama_local"
	case "anthropic":
		if c.AnthropicAPIKey != "" {
			return "anthropic"
		}
		return "ollama_local"
	case "groq":
		if c.GroqAPIKey != "" {
			return "groq"
		}
		return "ollama_local"
	case "ollama_local":
		return "ollama_local"
	default:
		return "ollama_local"
	}
}

// UseOllamaCloud returns true if Ollama Cloud should be used
func (c *Config) UseOllamaCloud() bool {
	return c.GetActiveProvider() == "ollama_cloud"
}

// UseAnthropic returns true if Anthropic/Claude should be used
func (c *Config) UseAnthropic() bool {
	return c.GetActiveProvider() == "anthropic"
}

// UseGroq returns true if Groq should be used
func (c *Config) UseGroq() bool {
	return c.GetActiveProvider() == "groq"
}

// UseOllamaLocal returns true if local Ollama should be used
func (c *Config) UseOllamaLocal() bool {
	return c.GetActiveProvider() == "ollama_local"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// LocalModelsAllowed returns true if local models can be used
// In production, this should typically be false
func (c *Config) LocalModelsAllowed() bool {
	// In production, respect the explicit flag
	if c.IsProduction() {
		return c.EnableLocalModels
	}
	// In development, always allow local models
	return true
}

// GetSearchProvider returns the active search provider
// Priority when "auto": Brave > Serper > Tavily > DuckDuckGo
func (c *Config) GetSearchProvider() string {
	if c.SearchProvider != "" && c.SearchProvider != "auto" {
		return c.SearchProvider
	}

	// Auto-select based on available API keys
	if c.BraveSearchAPIKey != "" {
		return "brave"
	}
	if c.SerperAPIKey != "" {
		return "serper"
	}
	if c.TavilyAPIKey != "" {
		return "tavily"
	}
	return "duckduckgo"
}

// HasBraveSearch returns true if Brave Search API is configured
func (c *Config) HasBraveSearch() bool {
	return c.BraveSearchAPIKey != ""
}

// HasSerper returns true if Serper API is configured
func (c *Config) HasSerper() bool {
	return c.SerperAPIKey != ""
}

// HasTavily returns true if Tavily API is configured
func (c *Config) HasTavily() bool {
	return c.TavilyAPIKey != ""
}

// Validate checks that the configuration is secure for the current environment
// SECURITY: This must be called on startup to prevent insecure production deployments
func (c *Config) Validate() error {
	var errs []string

	if c.IsProduction() {
		// CRITICAL: SECRET_KEY must be changed from default
		if c.SecretKey == "your-secret-key-change-this-in-production" {
			errs = append(errs, "SECRET_KEY must be changed from default value in production")
		}
		if len(c.SecretKey) < 32 {
			errs = append(errs, "SECRET_KEY must be at least 32 characters in production")
		}

		// CRITICAL: REDIS_KEY_HMAC_SECRET must be set for session security
		if c.RedisKeyHMACSecret == "" {
			errs = append(errs, "REDIS_KEY_HMAC_SECRET must be set in production (min 32 bytes)")
		}
		if len(c.RedisKeyHMACSecret) > 0 && len(c.RedisKeyHMACSecret) < 32 {
			errs = append(errs, "REDIS_KEY_HMAC_SECRET must be at least 32 characters")
		}

		// CRITICAL: Database URL must not be localhost in production
		if strings.Contains(c.DatabaseURL, "localhost") {
			errs = append(errs, "DATABASE_URL appears to be a development URL (contains 'localhost')")
		}

		// WARNING: Local models should typically be disabled in production
		if c.EnableLocalModels {
			// This is a warning, not an error - some deployments may need this
			slog.Warn("ENABLE_LOCAL_MODELS is true in production - ensure this is intentional")
		}
	}

	if len(errs) > 0 {
		return errors.New("configuration validation failed:\n  - " + strings.Join(errs, "\n  - "))
	}

	return nil
}

// MustValidate calls Validate and panics if validation fails
// Use this in main() to fail fast on misconfiguration
func (c *Config) MustValidate() {
	if err := c.Validate(); err != nil {
		panic(fmt.Sprintf("CRITICAL: %v", err))
	}
}
