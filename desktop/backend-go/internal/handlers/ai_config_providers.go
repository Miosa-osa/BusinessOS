package handlers

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// GetLLMProviders returns available LLM providers and their configuration status
func (h *AIConfigHandler) GetLLMProviders(c *gin.Context) {
	providers := []LLMProvider{
		{
			ID:          "ollama_cloud",
			Name:        "Ollama Cloud",
			Type:        "cloud",
			Description: "Run Llama and other models via Ollama's cloud API",
			Configured:  h.cfg.OllamaCloudAPIKey != "",
		},
		{
			ID:          "ollama_local",
			Name:        "Ollama (Local)",
			Type:        "local",
			Description: "Run open-source models locally on your machine",
			Configured:  h.isOllamaAvailable(),
			BaseURL:     h.cfg.OllamaLocalURL,
		},
		{
			ID:          "groq",
			Name:        "Groq",
			Type:        "cloud",
			Description: "Ultra-fast inference with Groq's LPU hardware",
			Configured:  h.cfg.GroqAPIKey != "",
		},
		{
			ID:          "anthropic",
			Name:        "Anthropic Claude",
			Type:        "cloud",
			Description: "Claude AI models from Anthropic",
			Configured:  h.cfg.AnthropicAPIKey != "",
		},
	}

	// Get user's default model from their settings (if authenticated)
	defaultModel := h.cfg.DefaultModel
	user := middleware.GetCurrentUser(c)
	if user != nil {
		queries := sqlc.New(h.pool)
		settings, err := queries.GetUserSettings(c.Request.Context(), user.ID)
		if err == nil && settings.DefaultModel != nil && *settings.DefaultModel != "" {
			defaultModel = *settings.DefaultModel
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"providers":       providers,
		"active_provider": h.cfg.GetActiveProvider(),
		"default_model":   defaultModel,
	})
}

// SaveAPIKey saves an API key to the .env file
func (h *AIConfigHandler) SaveAPIKey(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
		APIKey   string `json:"api_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	// Map provider to env key
	envKeys := map[string]string{
		"ollama_cloud": "OLLAMA_CLOUD_API_KEY",
		"groq":         "GROQ_API_KEY",
		"anthropic":    "ANTHROPIC_API_KEY",
	}

	envKey, ok := envKeys[req.Provider]
	if !ok {
		utils.RespondBadRequest(c, slog.Default(), "Invalid provider")
		return
	}

	// Read current .env file
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "read .env file", err)
		return
	}

	// Update or add the key
	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, envKey+"=") {
			lines[i] = envKey + "=" + req.APIKey
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, envKey+"="+req.APIKey)
	}

	// Write back
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(envPath, []byte(newContent), 0644); err != nil {
		utils.RespondInternalError(c, slog.Default(), "write .env file", err)
		return
	}

	// Also update the in-memory config so it takes effect immediately
	switch req.Provider {
	case "ollama_cloud":
		h.cfg.OllamaCloudAPIKey = req.APIKey
	case "groq":
		h.cfg.GroqAPIKey = req.APIKey
	case "anthropic":
		h.cfg.AnthropicAPIKey = req.APIKey
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "API key saved",
		"provider":   req.Provider,
		"configured": true,
	})
}

// UpdateAIProvider updates the active AI provider
func (h *AIConfigHandler) UpdateAIProvider(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	validProviders := []string{"ollama_local", "ollama_cloud", "groq", "anthropic"}
	isValid := false
	for _, p := range validProviders {
		if p == req.Provider {
			isValid = true
			break
		}
	}
	if !isValid {
		utils.RespondBadRequest(c, slog.Default(), "Invalid provider")
		return
	}

	// Update .env file
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "read .env file", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "AI_PROVIDER=") {
			lines[i] = "AI_PROVIDER=" + req.Provider
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, "AI_PROVIDER="+req.Provider)
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(envPath, []byte(newContent), 0644); err != nil {
		utils.RespondInternalError(c, slog.Default(), "write .env file", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider updated",
		"provider": req.Provider,
		"note":     "Restart the backend for changes to take effect",
	})
}
