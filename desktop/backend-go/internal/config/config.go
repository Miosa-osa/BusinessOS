package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// Environment (development, production)
	Environment string `mapstructure:"ENVIRONMENT"`

	// Database
	DatabaseURL string `mapstructure:"DATABASE_URL"`

	// Server
	ServerPort string `mapstructure:"SERVER_PORT"`

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
	RedisURL string `mapstructure:"REDIS_URL"`

	// Supermemory
	SupermemoryAPIKey string `mapstructure:"SUPERMEMORY_API_KEY"`

	// Google OAuth
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURI  string `mapstructure:"GOOGLE_REDIRECT_URI"`

	// CORS
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`

	// Feature Flags
	EnableLocalModels bool `mapstructure:"ENABLE_LOCAL_MODELS"`
}

var AppConfig *Config

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Set defaults
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SERVER_PORT", "8000")
	viper.SetDefault("DATABASE_URL", "postgres://postgres:password@localhost:5432/business_os")
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

	// Other services
	viper.SetDefault("REDIS_URL", "redis://localhost:6379/0")
	viper.SetDefault("SUPERMEMORY_API_KEY", "")
	viper.SetDefault("GOOGLE_CLIENT_ID", "")
	viper.SetDefault("GOOGLE_CLIENT_SECRET", "")
	viper.SetDefault("GOOGLE_REDIRECT_URI", "http://localhost:8000/api/integrations/google/callback")
	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174,http://localhost:3000,app://localhost")

	// Read from environment variables first (takes priority in production)
	viper.AutomaticEnv()

	// Try to read config file (optional - for local development)
	// Ignore all errors - we can run without a config file in production
	_ = viper.ReadInConfig()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	// Parse ALLOWED_ORIGINS from comma-separated string
	originsStr := viper.GetString("ALLOWED_ORIGINS")
	if originsStr != "" {
		config.AllowedOrigins = strings.Split(originsStr, ",")
		for i := range config.AllowedOrigins {
			config.AllowedOrigins[i] = strings.TrimSpace(config.AllowedOrigins[i])
		}
	}

	// Convert DATABASE_URL from asyncpg format to pgx format if needed
	// Python uses: postgresql+asyncpg://...
	// Go pgx uses: postgres://... or postgresql://...
	config.DatabaseURL = strings.Replace(config.DatabaseURL, "postgresql+asyncpg://", "postgres://", 1)

	AppConfig = config
	return config, nil
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
