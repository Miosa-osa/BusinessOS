package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type SkillsConfig struct {
	Version         string            `yaml:"version"`
	SkillsDirectory string            `yaml:"skills_directory"`
	Settings        SkillsSettings    `yaml:"settings"`
	Skills          []SkillDefinition `yaml:"skills"`
	SkillGroups     map[string][]string `yaml:"skill_groups"`
}

// SkillsSettings contains global settings for skills loading
type SkillsSettings struct {
	MaxSkillsInContext   int  `yaml:"max_skills_in_context"`
	DefaultTokenBudget   int  `yaml:"default_token_budget"`
	EnableTelemetry      bool `yaml:"enable_telemetry"`
	EnableSchemaValidation bool `yaml:"enable_schema_validation"`
}

// SkillDefinition is an entry in skills.yaml
type SkillDefinition struct {
	Name            string `yaml:"name"`
	Path            string `yaml:"path"`
	Enabled         bool   `yaml:"enabled"`
	Priority        int    `yaml:"priority"`
	AlwaysAvailable bool   `yaml:"always_available"`
}

// SkillMetadata is parsed from the YAML frontmatter of SKILL.md
// This is the "discovery" data - minimal info for the agent to decide
// whether to load the full skill.
type SkillMetadata struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Version     string   `yaml:"version"`
	Author      string   `yaml:"author"`
	ToolsUsed   []string `yaml:"tools_used"`
	DependsOn   []string `yaml:"depends_on"`
	
	// Runtime fields (not from frontmatter)
	Path        string   `yaml:"-"` // Filesystem path to skill folder
	Enabled     bool     `yaml:"-"` // From skills.yaml
	Priority    int      `yaml:"-"` // From skills.yaml
}

// SkillMetadataWrapper wraps the nested metadata structure in frontmatter
type SkillMetadataWrapper struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Metadata    struct {
		Version   string   `yaml:"version"`
		Author    string   `yaml:"author"`
		ToolsUsed []string `yaml:"tools_used"`
		DependsOn []string `yaml:"depends_on"`
	} `yaml:"metadata"`
}

// ============================================================================
// SKILLS LOADER
// ============================================================================

// SkillsLoader manages loading and caching of skill definitions.
//
// Thread Safety:
// - Uses RWMutex for concurrent access
// - Safe to call from multiple goroutines (e.g., multiple agent conversations)
//
// Caching:
// - Skill metadata is cached after first load
// - Call Reload() to refresh from disk
type SkillsLoader struct {
	// Configuration
	configPath string // Path to skills.yaml
	basePath   string // Base directory for skill folders
	
	// Cached data
	config     *SkillsConfig
	skills     map[string]*SkillMetadata // name -> metadata
	
	// Thread safety
	mu         sync.RWMutex
	loaded     bool
}

// NewSkillsLoader creates a new skills loader.
//
// configPath: Path to skills.yaml (e.g., "./skills/skills.yaml")
//
// The loader doesn't read files until LoadConfig() is called.
// This allows graceful handling if skills directory doesn't exist.
func NewSkillsLoader(configPath string) *SkillsLoader {
	return &SkillsLoader{
		configPath: configPath,
		basePath:   filepath.Dir(configPath), // skills.yaml is in skills/ folder
		skills:     make(map[string]*SkillMetadata),
	}
}

// LoadConfig reads skills.yaml and parses all skill metadata.
//
// This should be called once at server startup. It:
// 1. Reads skills.yaml to get list of enabled skills
// 2. For each enabled skill, reads SKILL.md and parses frontmatter
// 3. Caches all metadata in memory
//
// Returns error if skills.yaml doesn't exist or is invalid.
// Individual skill errors are logged but don't fail the whole load.
func (l *SkillsLoader) LoadConfig() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Read skills.yaml
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return fmt.Errorf("failed to read skills config: %w", err)
	}
	
	var config SkillsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse skills config: %w", err)
	}
	
	l.config = &config
	l.skills = make(map[string]*SkillMetadata)
	
	// Load each enabled skill
	for _, def := range config.Skills {
		if !def.Enabled {
			continue
		}
		
		// Build path to SKILL.md
		skillPath := filepath.Join(l.basePath, def.Name, "SKILL.md")
		
		// Parse the skill
		metadata, err := l.parseSkillFile(skillPath)
		if err != nil {
			// Log error but continue loading other skills
			fmt.Printf("[SkillsLoader] Warning: failed to load skill %s: %v\n", def.Name, err)
			continue
		}
		
		// Merge runtime fields from skills.yaml
		metadata.Path = filepath.Join(l.basePath, def.Name)
		metadata.Enabled = def.Enabled
		metadata.Priority = def.Priority
		
		l.skills[def.Name] = metadata
	}
	
	l.loaded = true
	fmt.Printf("[SkillsLoader] Loaded %d skills\n", len(l.skills))
	
	return nil
}

// parseSkillFile reads a SKILL.md file and extracts the YAML frontmatter.
//
// SKILL.md format:
//   ---
//   name: dashboard-management
//   description: Create and configure custom dashboards...
//   metadata:
//     version: "1.0.0"
//     tools_used:
//       - configure_dashboard
//   ---
//   
//   # Dashboard Management
//   (rest of markdown content)
//
// We only parse the frontmatter here (between --- markers).
// The full content is loaded on-demand via GetSkillContent().
func (l *SkillsLoader) parseSkillFile(path string) (*SkillMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}
	
	// Extract frontmatter (between --- markers)
	frontmatter, _, err := extractFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}
	
	// Parse YAML frontmatter
	var wrapper SkillMetadataWrapper
	if err := yaml.Unmarshal([]byte(frontmatter), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter YAML: %w", err)
	}
	
	// Build metadata from wrapper
	metadata := &SkillMetadata{
		Name:        wrapper.Name,
		Description: wrapper.Description,
		Version:     wrapper.Metadata.Version,
		Author:      wrapper.Metadata.Author,
		ToolsUsed:   wrapper.Metadata.ToolsUsed,
		DependsOn:   wrapper.Metadata.DependsOn,
	}
	
	return metadata, nil
}

// extractFrontmatter splits a markdown file into frontmatter and content.
//
// Input:
//   ---
//   name: foo
//   ---
//   # Content here
//
// Output:
//   frontmatter: "name: foo"
//   content: "# Content here"
func extractFrontmatter(content string) (frontmatter, body string, err error) {
	// Normalize line endings to Unix-style (\n) for consistent parsing
	// This handles files saved with Windows (\r\n) or old Mac (\r) line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	
	// Regex to match frontmatter block
	// ^---\n captures opening delimiter
	// ([\s\S]*?) captures frontmatter content (non-greedy)
	// \n---\n captures closing delimiter
	re := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)
	
	matches := re.FindStringSubmatch(content)
	if len(matches) != 3 {
		return "", content, fmt.Errorf("no valid frontmatter found")
	}
	
	return matches[1], matches[2], nil
}

// ============================================================================
// SKILL ACCESS METHODS
// ============================================================================

// GetEnabledSkills returns metadata for all enabled skills.
//
// Use this to list skills in admin UI or debug endpoints.
// Sorted by priority (highest first).
func (l *SkillsLoader) GetEnabledSkills() []*SkillMetadata {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	if !l.loaded {
		return nil
	}
	
	// Collect and sort by priority
	skills := make([]*SkillMetadata, 0, len(l.skills))
	for _, skill := range l.skills {
		skills = append(skills, skill)
	}
	
	// Sort by priority descending
	for i := 0; i < len(skills)-1; i++ {
		for j := i + 1; j < len(skills); j++ {
			if skills[j].Priority > skills[i].Priority {
				skills[i], skills[j] = skills[j], skills[i]
			}
		}
	}
	
	return skills
}

// GetSkillMetadata returns metadata for a specific skill.
//
// Returns nil if skill doesn't exist or isn't enabled.
func (l *SkillsLoader) GetSkillMetadata(name string) *SkillMetadata {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return l.skills[name]
}

// GetSkillContent returns the full SKILL.md content for a skill.
//
// This is called when the agent activates a skill and needs the full
// instructions. The content includes the frontmatter and all markdown.
//
// Returns error if skill doesn't exist or file can't be read.
func (l *SkillsLoader) GetSkillContent(name string) (string, error) {
	l.mu.RLock()
	skill := l.skills[name]
	l.mu.RUnlock()
	
	if skill == nil {
		return "", fmt.Errorf("skill not found: %s", name)
	}
	
	// Read the full file
	path := filepath.Join(skill.Path, "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read skill content: %w", err)
	}
	
	return string(data), nil
}

// GetSkillReference returns a reference file from a skill's references/ folder.
//
// Example: GetSkillReference("dashboard-management", "WIDGETS.md")
// Returns contents of: skills/dashboard-management/references/WIDGETS.md
//
// This is for on-demand loading of detailed documentation.
func (l *SkillsLoader) GetSkillReference(skillName, refName string) (string, error) {
	l.mu.RLock()
	skill := l.skills[skillName]
	l.mu.RUnlock()
	
	if skill == nil {
		return "", fmt.Errorf("skill not found: %s", skillName)
	}
	
	// Security: Prevent path traversal attacks
	// Only allow simple filenames, no slashes or dots
	if strings.Contains(refName, "/") || strings.Contains(refName, "\\") || strings.HasPrefix(refName, ".") {
		return "", fmt.Errorf("invalid reference name: %s", refName)
	}
	
	// Build path to reference file
	path := filepath.Join(skill.Path, "references", refName)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read reference: %w", err)
	}
	
	return string(data), nil
}

// GetSkillSchema returns the JSON schema for a skill's tool input.
//
// Example: GetSkillSchema("dashboard-management", "input.schema.json")
// Returns contents of: skills/dashboard-management/schemas/input.schema.json
func (l *SkillsLoader) GetSkillSchema(skillName, schemaName string) (string, error) {
	l.mu.RLock()
	skill := l.skills[skillName]
	l.mu.RUnlock()
	
	if skill == nil {
		return "", fmt.Errorf("skill not found: %s", skillName)
	}
	
	// Security: Prevent path traversal
	if strings.Contains(schemaName, "/") || strings.Contains(schemaName, "\\") || strings.HasPrefix(schemaName, ".") {
		return "", fmt.Errorf("invalid schema name: %s", schemaName)
	}
	
	path := filepath.Join(skill.Path, "schemas", schemaName)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read schema: %w", err)
	}
	
	return string(data), nil
}

// ListSkillReferences returns available reference files for a skill.
func (l *SkillsLoader) ListSkillReferences(skillName string) ([]string, error) {
	l.mu.RLock()
	skill := l.skills[skillName]
	l.mu.RUnlock()
	
	if skill == nil {
		return nil, fmt.Errorf("skill not found: %s", skillName)
	}
	
	refsPath := filepath.Join(skill.Path, "references")
	entries, err := os.ReadDir(refsPath)
	if err != nil {
		// No references folder is OK, return empty list
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read references: %w", err)
	}
	
	refs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			refs = append(refs, entry.Name())
		}
	}
	
	return refs, nil
}

// ============================================================================
// AGENT INTEGRATION
// ============================================================================

// GetSkillsPromptXML generates the <available_skills> XML block for injection
// into the agent's system prompt.
//
// This is the "discovery" layer - tells the agent what skills exist and what
// they're for, using only ~50 tokens per skill.
//
// Output format:
//   <available_skills>
//     <skill name="dashboard-management">
//       Create and configure custom dashboards with widgets. Add task summaries,
//       burndown charts, project progress, upcoming deadlines, and metric cards.
//     </skill>
//     <skill name="task-management">
//       Create, update, complete, and delete tasks. Filter by status, project,
//       priority, or due date. Bulk operations for efficiency.
//     </skill>
//   </available_skills>
//
// The agent uses this to decide which skill to activate based on user request.
func (l *SkillsLoader) GetSkillsPromptXML() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	if !l.loaded || len(l.skills) == 0 {
		return ""
	}
	
	var sb strings.Builder
	sb.WriteString("<available_skills>\n")
	
	// Collect and sort skills by priority (inline to avoid deadlock)
	skills := make([]*SkillMetadata, 0, len(l.skills))
	for _, skill := range l.skills {
		skills = append(skills, skill)
	}
	for i := 0; i < len(skills)-1; i++ {
		for j := i + 1; j < len(skills); j++ {
			if skills[j].Priority > skills[i].Priority {
				skills[i], skills[j] = skills[j], skills[i]
			}
		}
	}
	
	for _, skill := range skills {
		// Truncate description if too long (aim for ~50 tokens per skill)
		desc := skill.Description
		if len(desc) > 300 {
			desc = desc[:297] + "..."
		}
		
		// Clean up description (remove newlines, extra spaces)
		desc = strings.ReplaceAll(desc, "\n", " ")
		desc = strings.Join(strings.Fields(desc), " ")
		
		sb.WriteString(fmt.Sprintf("  <skill name=\"%s\">\n", skill.Name))
		sb.WriteString(fmt.Sprintf("    %s\n", desc))
		sb.WriteString("  </skill>\n")
	}
	
	sb.WriteString("</available_skills>")
	
	return sb.String()
}

// GetSkillsPromptInstructions returns instructions for how the agent should
// use the available skills. This is injected alongside the skills XML.
func (l *SkillsLoader) GetSkillsPromptInstructions() string {
	return `When a user request matches a skill description, you should:
1. Identify the most relevant skill based on keywords and intent
2. If you need the full instructions, request to load the skill
3. Follow the skill's "Request → Tool Mapping" to call the appropriate tool
4. Use the skill's examples as templates for your responses

If a request could match multiple skills, prefer the one with higher specificity.
For ambiguous requests, ask the user for clarification.`
}

// ============================================================================
// UTILITY METHODS
// ============================================================================

// Reload clears the cache and reloads all skills from disk.
//
// Use this if skills are updated while server is running.
// Returns error if reload fails (but old data is preserved).
func (l *SkillsLoader) Reload() error {
	// Store old data in case reload fails
	l.mu.Lock()
	oldConfig := l.config
	oldSkills := l.skills
	l.loaded = false
	l.mu.Unlock()
	
	// Attempt reload
	err := l.LoadConfig()
	if err != nil {
		// Restore old data
		l.mu.Lock()
		l.config = oldConfig
		l.skills = oldSkills
		l.loaded = true
		l.mu.Unlock()
		return err
	}
	
	return nil
}

// IsLoaded returns whether skills have been loaded.
func (l *SkillsLoader) IsLoaded() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loaded
}

// GetSettings returns the skills configuration settings.
func (l *SkillsLoader) GetSettings() *SkillsSettings {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	if l.config == nil {
		return nil
	}
	return &l.config.Settings
}

// ValidateSkill checks if a skill has all required files and valid structure.
//
// Checks:
// - SKILL.md exists and has valid frontmatter
// - All referenced tools exist (TODO: integrate with tool registry)
// - References folder structure is valid
//
// Returns list of warnings/errors.
func (l *SkillsLoader) ValidateSkill(name string) []string {
	l.mu.RLock()
	skill := l.skills[name]
	l.mu.RUnlock()
	
	if skill == nil {
		return []string{fmt.Sprintf("skill not found: %s", name)}
	}
	
	var issues []string
	
	// Check SKILL.md exists
	skillPath := filepath.Join(skill.Path, "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		issues = append(issues, "SKILL.md not found")
	}
	
	// Check name matches folder
	folderName := filepath.Base(skill.Path)
	if skill.Name != folderName {
		issues = append(issues, fmt.Sprintf("skill name '%s' doesn't match folder '%s'", skill.Name, folderName))
	}
	
	// Check description exists
	if skill.Description == "" {
		issues = append(issues, "description is empty")
	}
	
	// Check tools_used is specified
	if len(skill.ToolsUsed) == 0 {
		issues = append(issues, "no tools_used specified")
	}
	
	return issues
}

// ============================================================================
// SKILL GROUP METHODS
// ============================================================================

// GetSkillGroup returns skill names in a named group.
//
// Groups are defined in skills.yaml under skill_groups.
// Example: GetSkillGroup("productivity") returns ["task-management", "project-management", "dashboard-management"]
func (l *SkillsLoader) GetSkillGroup(groupName string) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	if l.config == nil || l.config.SkillGroups == nil {
		return nil
	}
	
	return l.config.SkillGroups[groupName]
}

// GetSkillsByTool returns skills that use a specific tool.
//
// Example: GetSkillsByTool("configure_dashboard") returns ["dashboard-management"]
// Useful for reverse lookup when agent wants to know which skill explains a tool.
func (l *SkillsLoader) GetSkillsByTool(toolName string) []*SkillMetadata {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	var matches []*SkillMetadata
	for _, skill := range l.skills {
		for _, tool := range skill.ToolsUsed {
			if tool == toolName {
				matches = append(matches, skill)
				break
			}
		}
	}
	
	return matches
}
