package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/prompts"
)

// LLMProvider represents an LLM provider configuration
type LLMProvider struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // "local" or "cloud"
	Description string `json:"description"`
	Configured  bool   `json:"configured"`
	BaseURL     string `json:"base_url,omitempty"`
}

// LLMModel represents an available model
type LLMModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"` // "ollama", "anthropic", etc.
	Description string `json:"description,omitempty"`
	Size        string `json:"size,omitempty"`
	Family      string `json:"family,omitempty"`
}

// OllamaTagsResponse represents the response from Ollama's /api/tags endpoint
type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

// OllamaModel represents a model in Ollama's response
type OllamaModel struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
	Details    struct {
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}

// GetLLMProviders returns available LLM providers and their configuration status
func (h *Handlers) GetLLMProviders(c *gin.Context) {
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
		"providers":        providers,
		"active_provider":  h.cfg.GetActiveProvider(),
		"default_model":    defaultModel,
	})
}

// GetLocalModels returns models available from local Ollama instance
func (h *Handlers) GetLocalModels(c *gin.Context) {
	models, err := h.fetchOllamaModels()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Ollama not available",
			"message": "Make sure Ollama is running locally. Visit https://ollama.ai to download.",
			"models":  []LLMModel{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models":   models,
		"provider": "ollama",
		"base_url": h.cfg.OllamaLocalURL,
	})
}

// GetAllModels returns all available models from all configured providers
func (h *Handlers) GetAllModels(c *gin.Context) {
	var allModels []LLMModel

	// Fetch local Ollama models
	ollamaModels, err := h.fetchOllamaModels()
	if err == nil {
		allModels = append(allModels, ollamaModels...)
	}

	// Add Ollama Cloud models (always available if key is set)
	if h.cfg.OllamaCloudAPIKey != "" {
		ollamaCloudModels := []LLMModel{
			// Qwen 3 Coder models (cloud variants - best for coding/agentic tasks)
			{
				ID:          "qwen3-coder:480b-cloud",
				Name:        "Qwen3 Coder 480B",
				Provider:    "ollama_cloud",
				Description: "480B cloud model - best quality for coding",
				Family:      "qwen",
			},
			{
				ID:          "qwen3-coder:30b",
				Name:        "Qwen3 Coder 30B",
				Provider:    "ollama_cloud",
				Description: "30B coding model",
				Family:      "qwen",
			},
			// Qwen 3 standard models
			{
				ID:          "qwen3:4b",
				Name:        "Qwen 3 4B",
				Provider:    "ollama_cloud",
				Description: "Fast, efficient 4B model",
				Family:      "qwen",
			},
			{
				ID:          "qwen3:8b",
				Name:        "Qwen 3 8B",
				Provider:    "ollama_cloud",
				Description: "Balanced 8B model",
				Family:      "qwen",
			},
			{
				ID:          "qwen3:14b",
				Name:        "Qwen 3 14B",
				Provider:    "ollama_cloud",
				Description: "Capable 14B model",
				Family:      "qwen",
			},
			{
				ID:          "qwen3:30b",
				Name:        "Qwen 3 30B",
				Provider:    "ollama_cloud",
				Description: "Large 30B model",
				Family:      "qwen",
			},
			{
				ID:          "qwen3:32b",
				Name:        "Qwen 3 32B",
				Provider:    "ollama_cloud",
				Description: "Large 32B model",
				Family:      "qwen",
			},
			// Llama models
			{
				ID:          "llama3.3:70b",
				Name:        "Llama 3.3 70B",
				Provider:    "ollama_cloud",
				Description: "Latest Llama model from Meta",
				Family:      "llama",
			},
			{
				ID:          "llama3.2",
				Name:        "Llama 3.2",
				Provider:    "ollama_cloud",
				Description: "Fast Llama model",
				Family:      "llama",
			},
			// DeepSeek reasoning models
			{
				ID:          "deepseek-r1:671b",
				Name:        "DeepSeek R1 671B",
				Provider:    "ollama_cloud",
				Description: "Full reasoning model - cloud",
				Family:      "deepseek",
			},
			{
				ID:          "deepseek-r1:70b",
				Name:        "DeepSeek R1 70B",
				Provider:    "ollama_cloud",
				Description: "Reasoning model",
				Family:      "deepseek",
			},
			{
				ID:          "deepseek-r1:32b",
				Name:        "DeepSeek R1 32B",
				Provider:    "ollama_cloud",
				Description: "Compact reasoning model",
				Family:      "deepseek",
			},
			// Mistral models
			{
				ID:          "mistral",
				Name:        "Mistral",
				Provider:    "ollama_cloud",
				Description: "Mistral AI's flagship model",
				Family:      "mistral",
			},
		}
		allModels = append(allModels, ollamaCloudModels...)
	}

	// Add Groq models if configured
	if h.cfg.GroqAPIKey != "" {
		groqModels := []LLMModel{
			{
				ID:          "llama-3.3-70b-versatile",
				Name:        "Llama 3.3 70B Versatile",
				Provider:    "groq",
				Description: "Fast 70B model for general tasks",
				Family:      "llama",
			},
			{
				ID:          "llama-3.1-8b-instant",
				Name:        "Llama 3.1 8B Instant",
				Provider:    "groq",
				Description: "Ultra-fast 8B model",
				Family:      "llama",
			},
			{
				ID:          "mixtral-8x7b-32768",
				Name:        "Mixtral 8x7B",
				Provider:    "groq",
				Description: "Mixtral MoE model with 32k context",
				Family:      "mixtral",
			},
			{
				ID:          "gemma2-9b-it",
				Name:        "Gemma 2 9B",
				Provider:    "groq",
				Description: "Google's Gemma 2 model",
				Family:      "gemma",
			},
		}
		allModels = append(allModels, groqModels...)
	}

	// Add Anthropic models if configured
	if h.cfg.AnthropicAPIKey != "" {
		anthropicModels := []LLMModel{
			{
				ID:          "claude-sonnet-4-20250514",
				Name:        "Claude Sonnet 4",
				Provider:    "anthropic",
				Description: "Fast, intelligent model for everyday tasks",
				Family:      "claude-4",
			},
			{
				ID:          "claude-opus-4-20250514",
				Name:        "Claude Opus 4",
				Provider:    "anthropic",
				Description: "Most capable model for complex tasks",
				Family:      "claude-4",
			},
			{
				ID:          "claude-3-5-sonnet-20241022",
				Name:        "Claude 3.5 Sonnet",
				Provider:    "anthropic",
				Description: "Previous generation, still highly capable",
				Family:      "claude-3.5",
			},
		}
		allModels = append(allModels, anthropicModels...)
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
		"models":          allModels,
		"active_provider": h.cfg.GetActiveProvider(),
		"default_model":   defaultModel,
	})
}

// PullModel triggers pulling a model from Ollama registry with streaming progress
func (h *Handlers) PullModel(c *gin.Context) {
	var req struct {
		Model string `json:"model" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Make request to Ollama to pull the model
	pullReq := map[string]interface{}{
		"name":   req.Model,
		"stream": true,
	}
	body, _ := json.Marshal(pullReq)

	resp, err := http.Post(h.cfg.OllamaLocalURL+"/api/pull", "application/json",
		io.NopCloser(strings.NewReader(string(body))))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Ollama not available",
			"message": "Make sure Ollama is running locally",
		})
		return
	}
	defer resp.Body.Close()

	// Set SSE headers for streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Stream the response from Ollama
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var progress map[string]interface{}
		if err := decoder.Decode(&progress); err != nil {
			if err == io.EOF {
				break
			}
			break
		}

		// Send progress event
		data, _ := json.Marshal(progress)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()

		// Check if done
		if status, ok := progress["status"].(string); ok && status == "success" {
			break
		}
	}

	// Send completion event
	fmt.Fprintf(c.Writer, "data: {\"status\":\"complete\",\"model\":\"%s\"}\n\n", req.Model)
	flusher.Flush()
}

// fetchOllamaModels fetches available models from local Ollama
func (h *Handlers) fetchOllamaModels() ([]LLMModel, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(h.cfg.OllamaLocalURL + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var tagsResp OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, err
	}

	models := make([]LLMModel, 0, len(tagsResp.Models))
	for _, m := range tagsResp.Models {
		size := formatSize(m.Size)
		models = append(models, LLMModel{
			ID:          m.Name,
			Name:        m.Name,
			Provider:    "ollama",
			Description: m.Details.ParameterSize + " parameters",
			Size:        size,
			Family:      m.Details.Family,
		})
	}

	return models, nil
}

// isOllamaAvailable checks if Ollama is running
func (h *Handlers) isOllamaAvailable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(h.cfg.OllamaLocalURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// SystemInfo represents system hardware information
type SystemInfo struct {
	TotalRAM         int64  `json:"total_ram_gb"`
	AvailableRAM     int64  `json:"available_ram_gb"`
	Platform         string `json:"platform"`
	HasGPU           bool   `json:"has_gpu"`
	GPUName          string `json:"gpu_name,omitempty"`
	RecommendedModels []RecommendedModel `json:"recommended_models"`
}

type RecommendedModel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RAMRequired string `json:"ram_required"`
	Speed       string `json:"speed"`
	Quality     string `json:"quality"`
}

// GetSystemInfo returns system hardware info and model recommendations
func (h *Handlers) GetSystemInfo(c *gin.Context) {
	info := SystemInfo{
		Platform: runtime.GOOS,
	}

	// Get actual memory info
	totalRAM, availRAM := getSystemMemory()
	info.TotalRAM = totalRAM
	info.AvailableRAM = availRAM

	// Check for GPU (basic check for macOS Metal support)
	if runtime.GOOS == "darwin" {
		info.HasGPU = true
		info.GPUName = "Apple Silicon / Metal"
	}

	// Model recommendations based on RAM
	if info.TotalRAM >= 32 {
		info.RecommendedModels = []RecommendedModel{
			{Name: "llama3.2:latest", Description: "Best balance of speed and quality", RAMRequired: "8GB", Speed: "Fast", Quality: "Excellent"},
			{Name: "qwen2.5:14b", Description: "Strong reasoning, larger context", RAMRequired: "12GB", Speed: "Medium", Quality: "Excellent"},
			{Name: "mistral:7b", Description: "Great for general tasks", RAMRequired: "6GB", Speed: "Fast", Quality: "Good"},
			{Name: "codellama:13b", Description: "Best for code tasks", RAMRequired: "10GB", Speed: "Medium", Quality: "Excellent"},
		}
	} else if info.TotalRAM >= 16 {
		info.RecommendedModels = []RecommendedModel{
			{Name: "llama3.2:3b", Description: "Fast and efficient", RAMRequired: "4GB", Speed: "Very Fast", Quality: "Good"},
			{Name: "llama3.2:latest", Description: "Best balance", RAMRequired: "8GB", Speed: "Fast", Quality: "Excellent"},
			{Name: "phi3:mini", Description: "Microsoft's efficient model", RAMRequired: "3GB", Speed: "Very Fast", Quality: "Good"},
			{Name: "mistral:7b", Description: "Great for general tasks", RAMRequired: "6GB", Speed: "Fast", Quality: "Good"},
		}
	} else {
		info.RecommendedModels = []RecommendedModel{
			{Name: "llama3.2:1b", Description: "Minimal resources", RAMRequired: "2GB", Speed: "Very Fast", Quality: "Basic"},
			{Name: "phi3:mini", Description: "Efficient small model", RAMRequired: "3GB", Speed: "Very Fast", Quality: "Good"},
			{Name: "tinyllama", Description: "Ultra lightweight", RAMRequired: "1GB", Speed: "Instant", Quality: "Basic"},
		}
	}

	c.JSON(http.StatusOK, info)
}

// SaveAPIKey saves an API key to the .env file
func (h *Handlers) SaveAPIKey(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
		APIKey   string `json:"api_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}

	// Read current .env file
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read .env file"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not write .env file"})
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
		"message":   "API key saved",
		"provider":  req.Provider,
		"configured": true,
	})
}

// UpdateAIProvider updates the active AI provider
func (h *Handlers) UpdateAIProvider(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}

	// Update .env file
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read .env file"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not write .env file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider updated",
		"provider": req.Provider,
		"note":     "Restart the backend for changes to take effect",
	})
}

// formatSize formats bytes to human readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return "< 1 KB"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(bytes) / float64(div)
	if exp < len(units) {
		return fmt.Sprintf("%.1f %s", value, units[exp])
	}
	return fmt.Sprintf("%.1f TB", value)
}

// getSystemMemory returns total and available RAM in GB
func getSystemMemory() (total int64, available int64) {
	switch runtime.GOOS {
	case "darwin":
		// Get total memory using sysctl on macOS
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			if bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); err == nil {
				total = bytes / (1024 * 1024 * 1024)
			}
		}

		// Get available memory using vm_stat
		out, err = exec.Command("vm_stat").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			var freePages, inactivePages int64
			pageSize := int64(4096) // Default page size

			for _, line := range lines {
				if strings.Contains(line, "page size") {
					parts := strings.Fields(line)
					for _, p := range parts {
						if val, err := strconv.ParseInt(p, 10, 64); err == nil && val > 0 {
							pageSize = val
							break
						}
					}
				}
				if strings.Contains(line, "Pages free:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						val := strings.TrimSuffix(parts[2], ".")
						freePages, _ = strconv.ParseInt(val, 10, 64)
					}
				}
				if strings.Contains(line, "Pages inactive:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						val := strings.TrimSuffix(parts[2], ".")
						inactivePages, _ = strconv.ParseInt(val, 10, 64)
					}
				}
			}
			available = (freePages + inactivePages) * pageSize / (1024 * 1024 * 1024)
		}

	case "linux":
		// Read /proc/meminfo on Linux
		content, err := os.ReadFile("/proc/meminfo")
		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "MemTotal:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						if kb, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
							total = kb / (1024 * 1024)
						}
					}
				}
				if strings.HasPrefix(line, "MemAvailable:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						if kb, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
							available = kb / (1024 * 1024)
						}
					}
				}
			}
		}
	}

	// Defaults if detection failed
	if total == 0 {
		total = 16
	}
	if available == 0 {
		available = total / 2
	}

	return total, available
}

// AgentInfo represents an AI agent with its prompt
type AgentInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
	Category    string `json:"category"`
}

// GetAgentPrompts returns all available agent prompts
func (h *Handlers) GetAgentPrompts(c *gin.Context) {
	agents := []AgentInfo{
		{
			ID:          "default",
			Name:        "Business OS Assistant",
			Description: "General business operations assistant for comprehensive guidance",
			Prompt:      prompts.DefaultPrompt,
			Category:    "general",
		},
		{
			ID:          "document",
			Name:        "Document Creator",
			Description: "Creates polished, professional business documents with real content",
			Prompt:      prompts.DocumentCreatorPrompt,
			Category:    "specialist",
		},
		{
			ID:          "analyst",
			Name:        "Business Analyst",
			Description: "Analyzes data, identifies insights, and provides strategic recommendations",
			Prompt:      prompts.AnalystPrompt,
			Category:    "specialist",
		},
		{
			ID:          "planner",
			Name:        "Strategic Planner",
			Description: "Helps with planning, prioritization, and strategic thinking",
			Prompt:      prompts.PlannerPrompt,
			Category:    "specialist",
		},
		{
			ID:          "orchestrator",
			Name:        "Orchestrator",
			Description: "Main coordinator that routes requests to specialized agents",
			Prompt:      prompts.OrchestratorPrompt,
			Category:    "system",
		},
		{
			ID:          "daily_planning",
			Name:        "Daily Planning Assistant",
			Description: "Executive assistant for daily productivity and prioritization",
			Prompt:      prompts.DailyPlanningPrompt,
			Category:    "specialist",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
	})
}

// GetAgentPrompt returns a specific agent's prompt
func (h *Handlers) GetAgentPrompt(c *gin.Context) {
	agentID := c.Param("id")
	prompt := prompts.GetPrompt(agentID)

	if prompt == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     agentID,
		"prompt": prompt,
	})
}

// WarmupModel sends a minimal request to load the model into memory
// This significantly reduces first-message latency for Ollama models
func (h *Handlers) WarmupModel(c *gin.Context) {
	var req struct {
		Model string `json:"model" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if this is a local Ollama model (skip warmup for cloud providers)
	provider := h.cfg.GetActiveProvider()
	if provider != "ollama_local" {
		// For cloud providers, just return success (no warmup needed)
		c.JSON(http.StatusOK, gin.H{
			"status":   "skipped",
			"model":    req.Model,
			"provider": provider,
			"message":  "Cloud providers don't require warmup",
		})
		return
	}

	// Send minimal request to Ollama to load model into memory
	warmupReq := map[string]interface{}{
		"model":    req.Model,
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
		"stream":   false,
		"options": map[string]interface{}{
			"num_predict": 1, // Generate minimal tokens
		},
	}
	body, _ := json.Marshal(warmupReq)

	client := &http.Client{Timeout: 120 * time.Second} // Allow time for model loading
	resp, err := client.Post(h.cfg.OllamaLocalURL+"/api/chat", "application/json",
		strings.NewReader(string(body)))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Ollama not available",
			"message": "Make sure Ollama is running locally",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Warmup failed",
			"message": string(respBody),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ready",
		"model":    req.Model,
		"provider": provider,
		"message":  "Model loaded into memory",
	})
}
