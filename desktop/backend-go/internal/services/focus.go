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
		SystemPromptPrefix:  `You are in Deep Research Mode with LIVE WEB SEARCH results provided below.

CRITICAL INSTRUCTIONS:
1. You MUST base your response primarily on the Search Results provided below
2. DO NOT make up or hallucinate information - only use data from the search results
3. Reference sources inline when making claims using [Source Name](URL) format
4. If the search results don't contain enough information, clearly state what is missing
5. ALWAYS end your response with a "## Sources" section listing ALL sources used

RESPONSE FORMAT:
- Start with a clear, comprehensive answer
- Use markdown formatting for readability
- End with:

## Sources
- [Source 1 Title](url1)
- [Source 2 Title](url2)
...

Your response should synthesize information from the search results to answer the user's question comprehensively.`,
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
			fmt.system_prompt_suffix,
			fc.auto_load_kb_categories
		FROM focus_mode_templates fmt
		LEFT JOIN focus_configurations fc ON fc.template_id = fmt.id AND fc.user_id = $1
		WHERE fmt.name = $2 AND fmt.is_active = true
	`, userID, focusMode)

	var settings FocusSettings
	var effectiveModel *string
	var thinkingStyle *string
	var systemPromptPrefix, systemPromptSuffix *string
	var autoLoadKBCategories []string

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
		&autoLoadKBCategories,
	)
	if err != nil {
		return nil, err
	}

	settings.EffectiveModel = effectiveModel
	settings.ThinkingStyle = thinkingStyle
	settings.AutoLoadKBCategories = autoLoadKBCategories
	if systemPromptPrefix != nil {
		settings.SystemPromptPrefix = *systemPromptPrefix
	}
	if systemPromptSuffix != nil {
		settings.SystemPromptSuffix = *systemPromptSuffix
	}

	// Merge with hardcoded defaults to ensure correct values
	// (DB may have NULL or outdated values)
	if defaults, ok := focusModeDefaults[focusMode]; ok {
		// Use hardcoded MaxTokens if DB returned default fallback (4096)
		if settings.MaxTokens == 4096 && defaults.MaxTokens > 4096 {
			settings.MaxTokens = defaults.MaxTokens
		}
		// Use hardcoded SystemPromptPrefix if DB returned empty
		if settings.SystemPromptPrefix == "" && defaults.SystemPromptPrefix != "" {
			settings.SystemPromptPrefix = defaults.SystemPromptPrefix
		}
		// Ensure AutoSearch and RequireSources from defaults
		if defaults.AutoSearch && !settings.AutoSearch {
			settings.AutoSearch = defaults.AutoSearch
		}
		if defaults.RequireSources && !settings.RequireSources {
			settings.RequireSources = defaults.RequireSources
		}
		if defaults.ThinkingEnabled && !settings.ThinkingEnabled {
			settings.ThinkingEnabled = defaults.ThinkingEnabled
		}
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

	// Build system prompt with output constraints
	basePrompt := s.buildSystemPrompt(settings)
	constraintInstructions := s.GetOutputConstraintsInstructions(focusCtx.OutputConstraints)
	if constraintInstructions != "" {
		focusCtx.SystemPrompt = basePrompt + "\n\n" + constraintInstructions
	} else {
		focusCtx.SystemPrompt = basePrompt
	}

	// Auto-load KB items - use intelligent loading if no specific categories configured
	if len(settings.AutoLoadKBCategories) > 0 {
		kbItems, err := s.loadKBItemsByCategories(ctx, userID, settings.AutoLoadKBCategories, settings.KBContextLimit)
		if err == nil && len(kbItems) > 0 {
			focusCtx.KBContext = kbItems
		}
	} else if userMessage != "" && settings.KBContextLimit > 0 {
		// Use intelligent auto-load based on query content and focus mode
		kbItems, err := s.AutoLoadKBContext(ctx, userID, focusMode, userMessage, settings.KBContextLimit)
		if err == nil && len(kbItems) > 0 {
			focusCtx.KBContext = kbItems
		}
	}

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

// loadKBItemsByCategories loads contexts (KB items) matching the specified categories
func (s *FocusService) loadKBItemsByCategories(ctx context.Context, userID string, categories []string, limit int) ([]KBContextItem, error) {
	if s.pool == nil || len(categories) == 0 {
		return nil, nil
	}

	if limit <= 0 {
		limit = 5 // Default limit
	}

	// Build query with category filter
	// Categories map to context types (PERSON, BUSINESS, PROJECT, CUSTOM, document, DOCUMENT)
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, type, content
		FROM contexts
		WHERE user_id = $1
		  AND is_archived = false
		  AND type = ANY($2::text[])
		ORDER BY updated_at DESC
		LIMIT $3
	`, userID, categories, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load KB items: %w", err)
	}
	defer rows.Close()

	var items []KBContextItem
	for rows.Next() {
		var item KBContextItem
		var contextType string
		var content *string

		err := rows.Scan(&item.ID, &item.Title, &contextType, &content)
		if err != nil {
			continue
		}

		item.Category = contextType
		if content != nil {
			item.Content = *content
		}

		items = append(items, item)
	}

	return items, nil
}

// AutoLoadKBContext intelligently loads KB items based on focus mode and query content
func (s *FocusService) AutoLoadKBContext(ctx context.Context, userID string, focusMode string, userQuery string, limit int) ([]KBContextItem, error) {
	if s.pool == nil {
		return nil, nil
	}

	if limit <= 0 {
		limit = 5
	}

	// Determine which context types to prioritize based on focus mode
	priorityTypes := s.getContextTypesForFocusMode(focusMode)

	// Extract keywords from query for relevance matching
	keywords := extractKeywords(userQuery)

	// Build query with smart relevance scoring
	query := `
		SELECT id, name, type, content,
		       (CASE WHEN type = ANY($3::text[]) THEN 2 ELSE 1 END) as type_score,
		       (CASE
		         WHEN name ILIKE ANY($4::text[]) THEN 3
		         WHEN content ILIKE ANY($4::text[]) THEN 2
		         ELSE 1
		       END) as relevance_score
		FROM contexts
		WHERE user_id = $1
		  AND is_archived = false
		  AND (name ILIKE ANY($4::text[]) OR content ILIKE ANY($4::text[]) OR type = ANY($3::text[]))
		ORDER BY (type_score * relevance_score) DESC, updated_at DESC
		LIMIT $2
	`

	// Build keyword patterns for ILIKE
	keywordPatterns := make([]string, len(keywords))
	for i, kw := range keywords {
		keywordPatterns[i] = "%" + kw + "%"
	}

	// If no keywords, fall back to category-based loading
	if len(keywordPatterns) == 0 {
		return s.loadKBItemsByCategories(ctx, userID, priorityTypes, limit)
	}

	rows, err := s.pool.Query(ctx, query, userID, limit, priorityTypes, keywordPatterns)
	if err != nil {
		// Fallback to simple category load
		return s.loadKBItemsByCategories(ctx, userID, priorityTypes, limit)
	}
	defer rows.Close()

	var items []KBContextItem
	for rows.Next() {
		var item KBContextItem
		var contextType string
		var content *string
		var typeScore, relevanceScore int

		err := rows.Scan(&item.ID, &item.Title, &contextType, &content, &typeScore, &relevanceScore)
		if err != nil {
			continue
		}

		item.Category = contextType
		if content != nil {
			item.Content = *content
		}

		items = append(items, item)
	}

	return items, nil
}

// getContextTypesForFocusMode returns priority context types based on focus mode
func (s *FocusService) getContextTypesForFocusMode(focusMode string) []string {
	switch focusMode {
	case "code", "build":
		return []string{"PROJECT", "DOCUMENT", "CUSTOM"}
	case "write":
		return []string{"DOCUMENT", "BUSINESS", "CUSTOM"}
	case "analyze":
		return []string{"BUSINESS", "PROJECT", "DOCUMENT"}
	case "plan", "planning":
		return []string{"PROJECT", "BUSINESS", "DOCUMENT"}
	case "research", "deep":
		return []string{"DOCUMENT", "CUSTOM", "BUSINESS"}
	case "creative":
		return []string{"CUSTOM", "DOCUMENT", "PERSON"}
	default:
		return []string{"DOCUMENT", "PROJECT", "BUSINESS", "CUSTOM"}
	}
}

// extractKeywords extracts important keywords from a query for matching
func extractKeywords(query string) []string {
	// Remove common stop words and extract meaningful terms
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "is": true, "are": true, "was": true,
		"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "can": true, "may": true, "might": true, "must": true,
		"i": true, "me": true, "my": true, "we": true, "our": true, "you": true, "your": true,
		"he": true, "she": true, "it": true, "they": true, "them": true, "their": true,
		"this": true, "that": true, "these": true, "those": true,
		"what": true, "which": true, "who": true, "whom": true, "where": true, "when": true,
		"why": true, "how": true, "with": true, "about": true, "for": true, "from": true,
		"of": true, "on": true, "in": true, "to": true, "at": true, "by": true, "as": true,
		"and": true, "or": true, "but": true, "if": true, "then": true, "so": true,
		"please": true, "help": true, "want": true, "need": true, "like": true,
		"tell": true, "explain": true, "show": true, "give": true, "make": true,
	}

	words := strings.Fields(strings.ToLower(query))
	var keywords []string
	seen := make(map[string]bool)

	for _, word := range words {
		// Clean punctuation
		word = strings.Trim(word, ".,!?;:'\"()[]{}")

		// Skip short words, stop words, and duplicates
		if len(word) < 3 || stopWords[word] || seen[word] {
			continue
		}

		seen[word] = true
		keywords = append(keywords, word)
	}

	// Limit to most relevant keywords
	if len(keywords) > 5 {
		keywords = keywords[:5]
	}

	return keywords
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
	} else {
		// Set default max lengths per mode/style
		constraints.MaxLength = s.getDefaultMaxLength(settings.Name, settings.OutputStyle)
	}

	// Auto-artifact for write mode or long content
	if settings.ResponseFormat == "artifact" {
		constraints.RequireArtifact = true
	}

	return constraints
}

// getDefaultMaxLength returns default max length based on mode and style
func (s *FocusService) getDefaultMaxLength(modeName string, style string) *int {
	// Mode-specific defaults (in characters)
	modeDefaults := map[string]int{
		"quick":    2000,  // ~500 words - concise
		"creative": 8000,  // ~2000 words - flexible
		"analyze":  12000, // ~3000 words - detailed analysis
		"write":    20000, // ~5000 words - full documents
		"plan":     10000, // ~2500 words - structured plans
		"code":     16000, // ~4000 words - code + explanations
		"deep":     16000, // ~4000 words - research
		"research": 16000, // ~4000 words - research
		"build":    12000, // ~3000 words - implementation
	}

	// Style overrides
	styleDefaults := map[string]int{
		"concise":    2000,  // Short responses
		"balanced":   6000,  // Medium responses
		"detailed":   12000, // Long responses
		"structured": 10000, // Organized responses
	}

	// Prefer mode-specific default
	if maxLen, ok := modeDefaults[modeName]; ok {
		return &maxLen
	}

	// Fall back to style default
	if maxLen, ok := styleDefaults[style]; ok {
		return &maxLen
	}

	// General default
	defaultLen := 6000
	return &defaultLen
}

// GetOutputConstraintsInstructions returns system prompt instructions for output constraints
func (s *FocusService) GetOutputConstraintsInstructions(constraints OutputConstraints) string {
	var instructions []string

	// Max length guidance
	if constraints.MaxLength != nil {
		maxWords := *constraints.MaxLength / 4 // ~4 chars per word
		switch {
		case maxWords <= 500:
			instructions = append(instructions, "Keep your response brief and focused. Target 2-4 paragraphs maximum.")
		case maxWords <= 1500:
			instructions = append(instructions, "Provide a moderate-length response. Be thorough but avoid unnecessary repetition.")
		case maxWords <= 3000:
			instructions = append(instructions, "You may provide a detailed response. Include relevant context and explanations.")
		default:
			instructions = append(instructions, "You may provide a comprehensive, in-depth response as needed.")
		}
	}

	// Format requirements
	if constraints.RequireArtifact {
		instructions = append(instructions, "For substantial content (documents, code, plans), generate an artifact that can be saved separately.")
	}

	// Source requirements
	if constraints.RequireSources {
		instructions = append(instructions, "CRITICAL: Include sources and citations. End with a '## Sources' section listing all references.")
	}

	if len(instructions) == 0 {
		return ""
	}

	return "## Output Requirements\n" + strings.Join(instructions, "\n")
}

// buildLLMOptions creates LLM options from focus settings
func (s *FocusService) buildLLMOptions(settings *FocusSettings) LLMOptions {
	opts := DefaultLLMOptions()
	opts.Temperature = settings.Temperature
	opts.MaxTokens = settings.MaxTokens
	opts.ThinkingEnabled = settings.ThinkingEnabled
	opts.Model = settings.EffectiveModel // Model override from focus mode

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

	// Add search context with explicit instructions
	if len(focusCtx.SearchContext) > 0 {
		parts = append(parts, `## WEB SEARCH RESULTS
The following are real-time search results. You MUST use these to answer the question:`)
		for i, item := range focusCtx.SearchContext {
			parts = append(parts, fmt.Sprintf("\n### Source %d: %s\n- **URL:** %s\n- **Summary:** %s", i+1, item.Title, item.URL, item.Snippet))
		}
		parts = append(parts, "\n---\nIMPORTANT: Base your answer on the sources above. Do not hallucinate information.")
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
