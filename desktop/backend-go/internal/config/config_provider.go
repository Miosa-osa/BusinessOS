package config

// GetActiveProvider returns the currently configured AI provider.
// Falls back to "ollama_local" when the configured provider lacks credentials.
func (c *Config) GetActiveProvider() string {
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

// GetModelForProvider delegates to GetActiveModel for backwards compatibility.
func (c *Config) GetModelForProvider() string {
	return c.GetActiveModel()
}

// GetActiveModel returns the appropriate model name based on the active AI provider.
func (c *Config) GetActiveModel() string {
	switch c.GetActiveProvider() {
	case "ollama_cloud":
		if c.OllamaCloudModel != "" {
			return c.OllamaCloudModel
		}
		return "llama3.2"
	case "groq":
		if c.GroqModel != "" {
			return c.GroqModel
		}
		return "openai/gpt-oss-20b"
	case "anthropic":
		if c.AnthropicModel != "" {
			return c.AnthropicModel
		}
		return "claude-sonnet-4-20250514"
	case "ollama_local":
		fallthrough
	default:
		if c.DefaultModel != "" {
			return c.DefaultModel
		}
		return "llama3.2:latest"
	}
}

// UseOllamaCloud returns true if Ollama Cloud should be used.
func (c *Config) UseOllamaCloud() bool {
	return c.GetActiveProvider() == "ollama_cloud"
}

// UseAnthropic returns true if Anthropic/Claude should be used.
func (c *Config) UseAnthropic() bool {
	return c.GetActiveProvider() == "anthropic"
}

// UseGroq returns true if Groq should be used.
func (c *Config) UseGroq() bool {
	return c.GetActiveProvider() == "groq"
}

// UseOllamaLocal returns true if local Ollama should be used.
func (c *Config) UseOllamaLocal() bool {
	return c.GetActiveProvider() == "ollama_local"
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// LocalModelsAllowed returns true if local models can be used.
// In production, this respects the explicit EnableLocalModels flag.
// In development, local models are always allowed.
func (c *Config) LocalModelsAllowed() bool {
	if c.IsProduction() {
		return c.EnableLocalModels
	}
	return true
}
