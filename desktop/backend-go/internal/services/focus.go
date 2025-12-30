package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FocusSettings represents the effective settings for a focus mode
type FocusSettings struct {
	Name                  string
	DisplayName           string
	EffectiveModel        *string
	Temperature           float64
	MaxTokens             int
	OutputStyle           string // concise, balanced, detailed, structured
	ResponseFormat        string // markdown, plain, json, artifact
	MaxResponseLength     *int
	RequireSources        bool
	AutoSearch            bool
	SearchDepth           string // quick, standard, deep
	KBContextLimit        int
	IncludeHistoryCount   int
	ThinkingEnabled       bool
	ThinkingStyle         *string
	SystemPromptPrefix    string
	SystemPromptSuffix    string
	CustomSystemPrompt    string
	AutoLoadKBCategories  []string
}

// FocusContext represents the pre-flight context to inject
type FocusContext struct {
	SystemPrompt      string             // Combined system prompt
	KBContext         []KBContextItem    // Knowledge base items to include
	SearchContext     []SearchContextItem // Web search results to include
	ProjectContext    []ProjectContextItem // Project context to include
	OutputConstraints OutputConstraints  // Server-side output constraints
	LLMOptions        LLMOptions         // LLM configuration
}

// KBContextItem represents a knowledge base item to inject
type KBContextItem struct {
	ID       uuid.UUID
	Title    string
	Content  string
	Category string
}

// SearchContextItem represents a search result to inject
type SearchContextItem struct {
	Title   string
	URL     string
	Snippet string
	Source  string
}

// ProjectContextItem represents project context to inject
type ProjectContextItem struct {
	ID          uuid.UUID
	Name        string
	Description string
	Status      string
}

// OutputConstraints defines server-side constraints for focus modes
type OutputConstraints struct {
	MaxLength       *int   // Maximum response length in chars
	Style           string // concise, balanced, detailed, structured
	Format          string // markdown, plain, json, artifact
	RequireSources  bool   // Must include sources/citations
	RequireArtifact bool   // Should generate artifact for long content
}

// FocusService handles focus mode configuration and context injection
type FocusService struct {
	pool *pgxpool.Pool
}

// NewFocusService creates a new focus service
func NewFocusService(pool *pgxpool.Pool) *FocusService {
	return &FocusService{pool: pool}
}

// focusModeDefaults contains hardcoded defaults for focus modes
var focusModeDefaults = map[string]*FocusSettings{
	"quick": {
		Name:                "quick",
		DisplayName:         "Quick",
		Temperature:         0.5,
		MaxTokens:           2048,
		OutputStyle:         "concise",
		ResponseFormat:      "markdown",
		ThinkingEnabled:     false,
		SystemPromptPrefix:  "You are in Quick Mode. Provide brief, direct answers. Be concise and to the point. Avoid unnecessary elaboration.",
	},
	"deep": {
		Name:                "deep",
		DisplayName:         "Deep Research",
		Temperature:         0.7,
		MaxTokens:           8192,
		OutputStyle:         "detailed",
		ResponseFormat:      "markdown",
		AutoSearch:          true,
		SearchDepth:         "deep",
		RequireSources:      true,
		ThinkingEnabled:     true,
		SystemPromptPrefix:  "You are in Deep Research Mode. Conduct thorough research and provide comprehensive, well-sourced answers. Include citations where possible. Analyze multiple perspectives.",
	},
	"creative": {
		Name:                "creative",
		DisplayName:         "Creative",
		Temperature:         0.9,
		MaxTokens:           4096,
		OutputStyle:         "balanced",
		ResponseFormat:      "markdown",
		ThinkingEnabled:     true,
		SystemPromptPrefix:  "You are in Creative Mode. Think outside the box. Explore unconventional ideas and approaches. Be imaginative and innovative in your responses.",
	},
	"analyze": {
		Name:                "analyze",
		DisplayName:         "Analysis",
		Temperature:         0.6,
		MaxTokens:           6144,
		OutputStyle:         "structured",
		ResponseFormat:      "markdown",
		ThinkingEnabled:     true,
		SystemPromptPrefix:  "You are in Analysis Mode. Focus on data-driven insights. Structure your response with clear sections. Use quantitative reasoning where applicable.",
	},
	"write": {
		Name:                "write",
		DisplayName:         "Writing",
		Temperature:         0.7,
		MaxTokens:           8192,
		OutputStyle:         "detailed",
		ResponseFormat:      "artifact",
		ThinkingEnabled:     false,
		SystemPromptPrefix:  "You are in Writing Mode. Create well-structured, polished content. Focus on clarity, flow, and appropriate tone. Generate artifacts for longer documents.",
	},
	"plan": {
		Name:                "plan",
		DisplayName:         "Planning",
		Temperature:         0.6,
		MaxTokens:           6144,
		OutputStyle:         "structured",
		ResponseFormat:      "markdown",
		ThinkingEnabled:     true,
		SystemPromptPrefix:  "You are in Planning Mode. Create actionable plans with clear steps. Consider dependencies and timelines. Structure output as organized lists or project artifacts.",
	},
	"code": {
		Name:                "code",
		DisplayName:         "Coding",
		Temperature:         0.4,
		MaxTokens:           8192,
		OutputStyle:         "structured",
		ResponseFormat:      "artifact",
		ThinkingEnabled:     true,
		SystemPromptPrefix:  "You are in Coding Mode. Write clean, efficient code. Follow best practices. Include comments where helpful. Generate code artifacts for complete implementations.",
	},
	"research": {
		Name:                "research",
		DisplayName:         "Research",
		Temperature:         0.7,
		MaxTokens:           8192,
		OutputStyle:         "detailed",
		ResponseFormat:      "markdown",
		AutoSearch:          true,
		SearchDepth:         "deep",
		RequireSources:      true,
		ThinkingEnabled:     true,
		SystemPromptPrefix:  "You are in Research Mode. Investigate the topic thoroughly. Gather information from multiple sources. Provide well-cited, comprehensive answers.",
	},
	"build": {
		Name:                "build",
		DisplayName:         "Build",
		Temperature:         0.5,
		MaxTokens:           8192,
		OutputStyle:         "structured",
		ResponseFormat:      "artifact",
		ThinkingEnabled:     true,
		SystemPromptPrefix:  "You are in Build Mode. Focus on implementation and construction. Create concrete deliverables. Generate artifacts for documents, code, or plans.",
	},
}

// GetEffectiveSettings retrieves merged focus settings for a user and mode
func (s *FocusService) GetEffectiveSettings(ctx context.Context, userID string, focusMode string) (*FocusSettings, error) {
	// First try to get from database
	settings, err := s.getSettingsFromDB(ctx, userID, focusMode)
	if err == nil && settings != nil {
		return settings, nil
	}

	// Fall back to hardcoded defaults
	return s.getDefaultSettings(focusMode), nil
}

// getSettingsFromDB attempts to load settings from database
func (s *FocusService) getSettingsFromDB(ctx context.Context, userID string, focusMode string) (*FocusSettings, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("no database pool")
	}

	// Query for focus mode template and user override
	row := s.pool.QueryRow(ctx, `
		SELECT
			fmt.name,
			fmt.display_name,
			COALESCE(fc.preferred_model, fmt.default_model) as effective_model,
			COALESCE(fc.temperature, fmt.temperature, 0.7) as temperature,
			COALESCE(fc.max_tokens, fmt.max_tokens, 4096) as max_tokens,
			COALESCE(fc.output_style, fmt.output_style, 'balanced') as output_style,
			COALESCE(fc.response_format, fmt.response_format, 'markdown') as response_format,
			COALESCE(fc.require_sources, fmt.require_sources, false) as require_sources,
			COALESCE(fc.auto_search, fmt.auto_search, false) as auto_search,
			COALESCE(fc.search_depth, fmt.search_depth, 'quick') as search_depth,
			COALESCE(fc.kb_context_limit, fmt.kb_context_limit, 5) as kb_context_limit,
			COALESCE(fc.thinking_enabled, fmt.thinking_enabled, false) as thinking_enabled,
			COALESCE(fc.thinking_style, fmt.thinking_style) as thinking_style,
			COALESCE(fc.custom_system_prompt, '') as custom_system_prompt,
			fmt.system_prompt_prefix,
			fmt.system_prompt_suffix
		FROM focus_mode_templates fmt
		LEFT JOIN focus_configurations fc ON fc.template_id = fmt.id AND fc.user_id = $1
		WHERE fmt.name = $2 AND fmt.is_active = true
	`, userID, focusMode)

	var settings FocusSettings
	var effectiveModel *string
	var thinkingStyle *string
	var systemPromptPrefix, systemPromptSuffix *string

	err := row.Scan(
		&settings.Name,
		&settings.DisplayName,
		&effectiveModel,
		&settings.Temperature,
		&settings.MaxTokens,
		&settings.OutputStyle,
		&settings.ResponseFormat,
		&settings.RequireSources,
		&settings.AutoSearch,
		&settings.SearchDepth,
		&settings.KBContextLimit,
		&settings.ThinkingEnabled,
		&thinkingStyle,
		&settings.CustomSystemPrompt,
		&systemPromptPrefix,
		&systemPromptSuffix,
	)
	if err != nil {
		return nil, err
	}

	settings.EffectiveModel = effectiveModel
	settings.ThinkingStyle = thinkingStyle
	if systemPromptPrefix != nil {
		settings.SystemPromptPrefix = *systemPromptPrefix
	}
	if systemPromptSuffix != nil {
		settings.SystemPromptSuffix = *systemPromptSuffix
	}

	return &settings, nil
}

// getDefaultSettings returns hardcoded defaults for a focus mode
func (s *FocusService) getDefaultSettings(focusMode string) *FocusSettings {
	if settings, ok := focusModeDefaults[focusMode]; ok {
		// Return a copy to prevent mutation
		copy := *settings
		return &copy
	}

	// Return general defaults for unknown modes
	return &FocusSettings{
		Name:           focusMode,
		DisplayName:    focusMode,
		Temperature:    0.7,
		MaxTokens:      4096,
		OutputStyle:    "balanced",
		ResponseFormat: "markdown",
	}
}

// BuildPreflightContext builds the complete context for a chat request
func (s *FocusService) BuildPreflightContext(
	ctx context.Context,
	userID string,
	focusMode string,
	userMessage string,
	contextIDs []uuid.UUID,
	projectID *uuid.UUID,
) (*FocusContext, error) {
	// Get effective settings
	settings, err := s.GetEffectiveSettings(ctx, userID, focusMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get focus settings: %w", err)
	}

	focusCtx := &FocusContext{
		OutputConstraints: s.buildOutputConstraints(settings),
		LLMOptions:        s.buildLLMOptions(settings),
	}

	// Build system prompt
	focusCtx.SystemPrompt = s.buildSystemPrompt(settings)

	// Perform web search if AutoSearch is enabled and user message is provided
	if settings.AutoSearch && userMessage != "" {
		searchResults, err := s.performWebSearch(ctx, userMessage, settings.SearchDepth)
		if err == nil && len(searchResults) > 0 {
			focusCtx.SearchContext = searchResults
		}
	}

	// Track usage (fire and forget with timeout)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.trackFocusModeUsage(ctx, userID, focusMode)
	}()

	return focusCtx, nil
}

// performWebSearch executes web search based on search depth
func (s *FocusService) performWebSearch(ctx context.Context, query string, searchDepth string) ([]SearchContextItem, error) {
	searchService := NewWebSearchService()

	// Determine max results based on search depth
	maxResults := 5
	switch searchDepth {
	case "quick":
		maxResults = 3
	case "standard":
		maxResults = 5
	case "deep":
		maxResults = 10
	}

	results, err := searchService.Search(ctx, query, maxResults)
	if err != nil {
		return nil, err
	}

	// Convert to SearchContextItem
	var contextItems []SearchContextItem
	for _, r := range results.Results {
		contextItems = append(contextItems, SearchContextItem{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Snippet,
			Source:  r.Source,
		})
	}

	return contextItems, nil
}

// buildOutputConstraints creates output constraints from settings
func (s *FocusService) buildOutputConstraints(settings *FocusSettings) OutputConstraints {
	constraints := OutputConstraints{
		Style:          settings.OutputStyle,
		Format:         settings.ResponseFormat,
		RequireSources: settings.RequireSources,
	}

	if settings.MaxResponseLength != nil {
		constraints.MaxLength = settings.MaxResponseLength
	}

	// Auto-artifact for write mode or long content
	if settings.ResponseFormat == "artifact" {
		constraints.RequireArtifact = true
	}

	return constraints
}

// buildLLMOptions creates LLM options from focus settings
func (s *FocusService) buildLLMOptions(settings *FocusSettings) LLMOptions {
	opts := DefaultLLMOptions()
	opts.Temperature = settings.Temperature
	opts.MaxTokens = settings.MaxTokens
	opts.ThinkingEnabled = settings.ThinkingEnabled

	return opts
}

// buildSystemPrompt builds the combined system prompt
func (s *FocusService) buildSystemPrompt(settings *FocusSettings) string {
	var parts []string

	// Add prefix from template
	if settings.SystemPromptPrefix != "" {
		parts = append(parts, settings.SystemPromptPrefix)
	}

	// Add custom user prompt if set
	if settings.CustomSystemPrompt != "" {
		parts = append(parts, settings.CustomSystemPrompt)
	}

	// Add output style instructions
	styleInstructions := s.getOutputStyleInstructions(settings.OutputStyle)
	if styleInstructions != "" {
		parts = append(parts, styleInstructions)
	}

	// Add source requirements
	if settings.RequireSources {
		parts = append(parts, "IMPORTANT: Include sources and citations for your claims. Format sources as [Source Title](URL) where available.")
	}

	// Add suffix from template
	if settings.SystemPromptSuffix != "" {
		parts = append(parts, settings.SystemPromptSuffix)
	}

	return strings.Join(parts, "\n\n")
}

// getOutputStyleInstructions returns instructions for output style
func (s *FocusService) getOutputStyleInstructions(style string) string {
	switch style {
	case "concise":
		return `## Output Style: Concise
- Keep responses brief and to the point
- Use bullet points where appropriate
- Avoid unnecessary elaboration
- Focus on actionable information
- Target 2-4 paragraphs maximum for most responses`

	case "detailed":
		return `## Output Style: Detailed
- Provide comprehensive, thorough responses
- Include relevant context and background
- Explain reasoning and methodology
- Cover edge cases and considerations
- Use examples to illustrate points`

	case "structured":
		return `## Output Style: Structured
- Organize response with clear sections and headers
- Use numbered lists for sequential steps
- Use bullet points for related items
- Include summary at the beginning or end
- Format data in tables where appropriate`

	case "balanced":
		fallthrough
	default:
		return "" // No additional instructions for balanced
	}
}

// trackFocusModeUsage increments usage counter
func (s *FocusService) trackFocusModeUsage(ctx context.Context, userID string, focusMode string) {
	if s.pool == nil {
		return
	}

	_, _ = s.pool.Exec(ctx, `
		UPDATE focus_configurations SET
			use_count = use_count + 1,
			last_used_at = NOW()
		WHERE user_id = $1 AND template_id = (
			SELECT id FROM focus_mode_templates WHERE name = $2
		)
	`, userID, focusMode)
}

// FormatContextForPrompt formats the focus context for injection into the prompt
func (s *FocusService) FormatContextForPrompt(focusCtx *FocusContext) string {
	var parts []string

	// Add KB context
	if len(focusCtx.KBContext) > 0 {
		parts = append(parts, "## Relevant Knowledge Base Context:")
		for _, item := range focusCtx.KBContext {
			parts = append(parts, fmt.Sprintf("### %s\n%s", item.Title, truncateContent(item.Content, 2000)))
		}
	}

	// Add search context
	if len(focusCtx.SearchContext) > 0 {
		parts = append(parts, "## Search Results:")
		for i, item := range focusCtx.SearchContext {
			parts = append(parts, fmt.Sprintf("%d. **%s**\n   URL: %s\n   %s", i+1, item.Title, item.URL, item.Snippet))
		}
	}

	// Add project context
	if len(focusCtx.ProjectContext) > 0 {
		parts = append(parts, "## Project Context:")
		for _, item := range focusCtx.ProjectContext {
			parts = append(parts, fmt.Sprintf("**Project: %s** (%s)\n%s", item.Name, item.Status, item.Description))
		}
	}

	return strings.Join(parts, "\n\n")
}

// Helper function
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
