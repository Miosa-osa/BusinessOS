package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AppProfilerService analyzes and profiles application codebases
type AppProfilerService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// ApplicationProfile represents a comprehensive profile of an application
type ApplicationProfile struct {
	ID                string                 `json:"id"`
	UserID            string                 `json:"user_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	RootPath          string                 `json:"root_path"`
	AutoSyncEnabled   bool                   `json:"auto_sync_enabled"`
	LastSyncedAt      *time.Time             `json:"last_synced_at,omitempty"`
	SyncSource        string                 `json:"sync_source,omitempty"`
	SyncBranch        string                 `json:"sync_branch,omitempty"`
	SyncCommit        string                 `json:"sync_commit,omitempty"`
	AppType           AppType                `json:"app_type"`
	Version           string                 `json:"version,omitempty"`
	TechStack         TechStack              `json:"tech_stack"`
	Languages         []LanguageInfo         `json:"languages"`
	Frameworks        []string               `json:"frameworks"`
	StructureTree     *DirectoryTree         `json:"structure_tree"`
	Components        []ComponentInfo        `json:"components"`
	TotalComponents   int                    `json:"total_components"`
	Modules           []ModuleInfo           `json:"modules"`
	TotalModules      int                    `json:"total_modules"`
	APIEndpoints      []APIEndpointInfo      `json:"api_endpoints"`
	TotalEndpoints    int                    `json:"total_endpoints"`
	DatabaseSchema    *DatabaseSchemaInfo    `json:"database_schema,omitempty"`
	Conventions       CodeConventions        `json:"conventions"`
	IntegrationPoints []IntegrationPoint     `json:"integration_points"`
	ReadmeSummary     string                 `json:"readme_summary,omitempty"`
	LinesOfCode       int                    `json:"lines_of_code"`
	FileCount         int                    `json:"file_count"`
	LastAnalyzedAt    time.Time              `json:"last_analyzed_at"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// AppType represents the type of application
type AppType string

const (
	AppTypeWeb        AppType = "web"
	AppTypeAPI        AppType = "api"
	AppTypeDesktop    AppType = "desktop"
	AppTypeMobile     AppType = "mobile"
	AppTypeCLI        AppType = "cli"
	AppTypeLibrary    AppType = "library"
	AppTypeMicroservice AppType = "microservice"
	AppTypeMonolith   AppType = "monolith"
	AppTypeFullStack  AppType = "fullstack"
)

// TechStack represents the technology stack
type TechStack struct {
	Frontend  []string `json:"frontend"`
	Backend   []string `json:"backend"`
	Database  []string `json:"database"`
	DevOps    []string `json:"devops"`
	Testing   []string `json:"testing"`
	BuildTool []string `json:"build_tool"`
}

// LanguageInfo represents programming language information
type LanguageInfo struct {
	Name       string  `json:"name"`
	Files      int     `json:"files"`
	Lines      int     `json:"lines"`
	Percentage float64 `json:"percentage"`
}

// DirectoryTree represents the project structure
type DirectoryTree struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	Type     string           `json:"type"` // file, directory
	Children []*DirectoryTree `json:"children,omitempty"`
	FileType string           `json:"file_type,omitempty"`
	Size     int64            `json:"size,omitempty"`
}

// ComponentInfo represents a UI component
type ComponentInfo struct {
	Name        string   `json:"name"`
	FilePath    string   `json:"file_path"`
	Type        string   `json:"type"` // page, component, layout, widget
	Description string   `json:"description,omitempty"`
	Props       []string `json:"props,omitempty"`
	Events      []string `json:"events,omitempty"`
	UsedIn      []string `json:"used_in,omitempty"`
	Lines       int      `json:"lines"`
}

// ModuleInfo represents a code module
type ModuleInfo struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Type        string   `json:"type"` // handler, service, repository, utility
	Description string   `json:"description,omitempty"`
	Exports     []string `json:"exports,omitempty"`
	Imports     []string `json:"imports,omitempty"`
	Lines       int      `json:"lines"`
}

// APIEndpointInfo represents an API endpoint
type APIEndpointInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Handler     string   `json:"handler"`
	HandlerPath string   `json:"handler_path"`
	Description string   `json:"description,omitempty"`
	AuthRequired bool    `json:"auth_required"`
	Tags        []string `json:"tags,omitempty"`
}

// DatabaseSchemaInfo represents database schema information
type DatabaseSchemaInfo struct {
	Tables      []TableInfo      `json:"tables"`
	TotalTables int              `json:"total_tables"`
	Migrations  []MigrationInfo  `json:"migrations,omitempty"`
}

// TableInfo represents database table information
type TableInfo struct {
	Name    string       `json:"name"`
	Columns []ColumnInfo `json:"columns"`
}

// ColumnInfo represents a database column
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Primary  bool   `json:"primary"`
}

// MigrationInfo represents a database migration
type MigrationInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Version   int    `json:"version"`
	AppliedAt string `json:"applied_at,omitempty"`
}

// CodeConventions represents coding conventions detected
type CodeConventions struct {
	NamingStyle     string   `json:"naming_style"` // camelCase, snake_case, PascalCase
	IndentStyle     string   `json:"indent_style"` // tabs, spaces
	IndentSize      int      `json:"indent_size"`
	QuoteStyle      string   `json:"quote_style"` // single, double
	Semicolons      bool     `json:"semicolons"`
	TrailingCommas  bool     `json:"trailing_commas"`
	FileNaming      string   `json:"file_naming"`
	DirectoryNaming string   `json:"directory_naming"`
	CommonPatterns  []string `json:"common_patterns"`
}

// IntegrationPoint represents an external integration
type IntegrationPoint struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // api, database, service, webhook
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// ProfileOptions configures profiling behavior
type ProfileOptions struct {
	MaxDepth          int      `json:"max_depth"`
	IncludeHidden     bool     `json:"include_hidden"`
	ExcludePatterns   []string `json:"exclude_patterns"`
	AnalyzeComponents bool     `json:"analyze_components"`
	AnalyzeEndpoints  bool     `json:"analyze_endpoints"`
	AnalyzeDatabase   bool     `json:"analyze_database"`
	ExtractReadme     bool     `json:"extract_readme"`
}

// DefaultProfileOptions returns default profiling options
func DefaultProfileOptions() *ProfileOptions {
	return &ProfileOptions{
		MaxDepth:      10,
		IncludeHidden: false,
		ExcludePatterns: []string{
			"node_modules", "vendor", ".git", "__pycache__", ".next",
			"build", "dist", ".svelte-kit", "coverage", ".cache",
		},
		AnalyzeComponents: true,
		AnalyzeEndpoints:  true,
		AnalyzeDatabase:   true,
		ExtractReadme:     true,
	}
}

// NewAppProfilerService creates a new app profiler service
func NewAppProfilerService(pool *pgxpool.Pool) *AppProfilerService {
	return &AppProfilerService{
		pool:   pool,
		logger: slog.Default().With("service", "app_profiler"),
	}
}

// ProfileApplication analyzes and profiles an application codebase
func (s *AppProfilerService) ProfileApplication(ctx context.Context, userID, rootPath, name string, opts *ProfileOptions) (*ApplicationProfile, error) {
	if opts == nil {
		opts = DefaultProfileOptions()
	}

	// Verify path exists
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %s", rootPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", rootPath)
	}

	profile := &ApplicationProfile{
		ID:                uuid.New().String(),
		UserID:            userID,
		Name:              name,
		RootPath:          rootPath,
		Languages:         make([]LanguageInfo, 0),
		Frameworks:        make([]string, 0),
		Components:        make([]ComponentInfo, 0),
		Modules:           make([]ModuleInfo, 0),
		APIEndpoints:      make([]APIEndpointInfo, 0),
		IntegrationPoints: make([]IntegrationPoint, 0),
		Metadata:          make(map[string]interface{}),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	s.logger.Info("starting application profiling", "path", rootPath, "name", name)

	// Build directory tree
	profile.StructureTree = s.buildDirectoryTree(rootPath, opts, 0)

	// Analyze languages
	profile.Languages, profile.LinesOfCode, profile.FileCount = s.analyzeLanguages(rootPath, opts)

	// Detect app type and tech stack
	profile.AppType, profile.TechStack = s.detectTechStack(rootPath, profile.Languages)
	profile.Frameworks = s.detectFrameworks(rootPath)

	// Detect conventions
	profile.Conventions = s.detectConventions(rootPath, opts)

	// Analyze components
	if opts.AnalyzeComponents {
		profile.Components = s.analyzeComponents(rootPath, profile.TechStack, opts)
		profile.TotalComponents = len(profile.Components)
	}

	// Analyze modules
	profile.Modules = s.analyzeModules(rootPath, profile.Languages, opts)
	profile.TotalModules = len(profile.Modules)

	// Analyze API endpoints
	if opts.AnalyzeEndpoints {
		profile.APIEndpoints = s.analyzeEndpoints(rootPath, profile.TechStack, opts)
		profile.TotalEndpoints = len(profile.APIEndpoints)
	}

	// Analyze database schema
	if opts.AnalyzeDatabase {
		profile.DatabaseSchema = s.analyzeDatabaseSchema(rootPath)
	}

	// Extract README summary
	if opts.ExtractReadme {
		profile.ReadmeSummary = s.extractReadmeSummary(rootPath)
	}

	// Detect integrations
	profile.IntegrationPoints = s.detectIntegrations(rootPath)

	// Generate description
	profile.Description = s.generateDescription(profile)

	profile.LastAnalyzedAt = time.Now()

	// Save profile
	if err := s.saveProfile(ctx, profile); err != nil {
		s.logger.Warn("failed to save profile", "error", err)
	}

	return profile, nil
}

// buildDirectoryTree builds the directory structure tree
func (s *AppProfilerService) buildDirectoryTree(rootPath string, opts *ProfileOptions, depth int) *DirectoryTree {
	if depth > opts.MaxDepth {
		return nil
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return nil
	}

	name := filepath.Base(rootPath)

	// Check exclusions
	for _, pattern := range opts.ExcludePatterns {
		if name == pattern || strings.Contains(rootPath, pattern) {
			return nil
		}
	}

	// Skip hidden files/dirs
	if !opts.IncludeHidden && strings.HasPrefix(name, ".") && name != "." {
		return nil
	}

	tree := &DirectoryTree{
		Name: name,
		Path: rootPath,
		Size: info.Size(),
	}

	if info.IsDir() {
		tree.Type = "directory"
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return tree
		}

		tree.Children = make([]*DirectoryTree, 0)
		for _, entry := range entries {
			childPath := filepath.Join(rootPath, entry.Name())
			child := s.buildDirectoryTree(childPath, opts, depth+1)
			if child != nil {
				tree.Children = append(tree.Children, child)
			}
		}
	} else {
		tree.Type = "file"
		tree.FileType = strings.TrimPrefix(filepath.Ext(name), ".")
	}

	return tree
}

// analyzeLanguages analyzes programming languages used
func (s *AppProfilerService) analyzeLanguages(rootPath string, opts *ProfileOptions) ([]LanguageInfo, int, int) {
	languageExtensions := map[string]string{
		".go":      "Go",
		".js":      "JavaScript",
		".ts":      "TypeScript",
		".jsx":     "JavaScript (React)",
		".tsx":     "TypeScript (React)",
		".svelte":  "Svelte",
		".vue":     "Vue",
		".py":      "Python",
		".rb":      "Ruby",
		".java":    "Java",
		".kt":      "Kotlin",
		".swift":   "Swift",
		".rs":      "Rust",
		".c":       "C",
		".cpp":     "C++",
		".cs":      "C#",
		".php":     "PHP",
		".sql":     "SQL",
		".html":    "HTML",
		".css":     "CSS",
		".scss":    "SCSS",
		".json":    "JSON",
		".yaml":    "YAML",
		".yml":     "YAML",
		".md":      "Markdown",
		".sh":      "Shell",
		".bat":     "Batch",
	}

	langStats := make(map[string]struct{ files, lines int })
	totalLines := 0
	totalFiles := 0

	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Check exclusions
		name := d.Name()
		for _, pattern := range opts.ExcludePatterns {
			if name == pattern {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(name))
		if lang, ok := languageExtensions[ext]; ok {
			lines := s.countLines(path)
			stats := langStats[lang]
			stats.files++
			stats.lines += lines
			langStats[lang] = stats
			totalLines += lines
			totalFiles++
		}

		return nil
	})

	// Convert to slice and calculate percentages
	languages := make([]LanguageInfo, 0, len(langStats))
	for lang, stats := range langStats {
		percentage := 0.0
		if totalLines > 0 {
			percentage = float64(stats.lines) / float64(totalLines) * 100
		}
		languages = append(languages, LanguageInfo{
			Name:       lang,
			Files:      stats.files,
			Lines:      stats.lines,
			Percentage: percentage,
		})
	}

	// Sort by lines (descending)
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].Lines > languages[j].Lines
	})

	return languages, totalLines, totalFiles
}

// countLines counts lines in a file
func (s *AppProfilerService) countLines(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return len(strings.Split(string(content), "\n"))
}

// detectTechStack detects the technology stack
func (s *AppProfilerService) detectTechStack(rootPath string, languages []LanguageInfo) (AppType, TechStack) {
	stack := TechStack{
		Frontend:  make([]string, 0),
		Backend:   make([]string, 0),
		Database:  make([]string, 0),
		DevOps:    make([]string, 0),
		Testing:   make([]string, 0),
		BuildTool: make([]string, 0),
	}

	appType := AppTypeWeb

	// Check for common config files
	checks := map[string]func(){
		"package.json": func() {
			content, _ := os.ReadFile(filepath.Join(rootPath, "package.json"))
			contentStr := string(content)
			if strings.Contains(contentStr, "react") {
				stack.Frontend = append(stack.Frontend, "React")
			}
			if strings.Contains(contentStr, "svelte") {
				stack.Frontend = append(stack.Frontend, "Svelte")
			}
			if strings.Contains(contentStr, "vue") {
				stack.Frontend = append(stack.Frontend, "Vue")
			}
			if strings.Contains(contentStr, "next") {
				stack.Frontend = append(stack.Frontend, "Next.js")
				appType = AppTypeFullStack
			}
			if strings.Contains(contentStr, "express") {
				stack.Backend = append(stack.Backend, "Express")
			}
			if strings.Contains(contentStr, "tailwind") {
				stack.Frontend = append(stack.Frontend, "Tailwind CSS")
			}
			if strings.Contains(contentStr, "jest") || strings.Contains(contentStr, "vitest") {
				stack.Testing = append(stack.Testing, "Jest/Vitest")
			}
		},
		"go.mod": func() {
			stack.Backend = append(stack.Backend, "Go")
			appType = AppTypeAPI
		},
		"requirements.txt": func() {
			stack.Backend = append(stack.Backend, "Python")
		},
		"Gemfile": func() {
			stack.Backend = append(stack.Backend, "Ruby")
		},
		"docker-compose.yml": func() {
			stack.DevOps = append(stack.DevOps, "Docker Compose")
		},
		"Dockerfile": func() {
			stack.DevOps = append(stack.DevOps, "Docker")
		},
		".github/workflows": func() {
			stack.DevOps = append(stack.DevOps, "GitHub Actions")
		},
		"prisma": func() {
			stack.Database = append(stack.Database, "Prisma")
		},
	}

	for file, check := range checks {
		if _, err := os.Stat(filepath.Join(rootPath, file)); err == nil {
			check()
		}
	}

	// Infer database from languages
	for _, lang := range languages {
		if lang.Name == "SQL" {
			stack.Database = append(stack.Database, "SQL")
		}
	}

	// Detect if it's a fullstack app
	if len(stack.Frontend) > 0 && len(stack.Backend) > 0 {
		appType = AppTypeFullStack
	}

	return appType, stack
}

// detectFrameworks detects frameworks used
func (s *AppProfilerService) detectFrameworks(rootPath string) []string {
	frameworks := make([]string, 0)

	// SvelteKit
	if _, err := os.Stat(filepath.Join(rootPath, "svelte.config.js")); err == nil {
		frameworks = append(frameworks, "SvelteKit")
	}

	// Next.js
	if _, err := os.Stat(filepath.Join(rootPath, "next.config.js")); err == nil {
		frameworks = append(frameworks, "Next.js")
	}

	// Go frameworks
	goMod := filepath.Join(rootPath, "go.mod")
	if content, err := os.ReadFile(goMod); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "gin-gonic/gin") {
			frameworks = append(frameworks, "Gin")
		}
		if strings.Contains(contentStr, "labstack/echo") {
			frameworks = append(frameworks, "Echo")
		}
		if strings.Contains(contentStr, "go-chi/chi") {
			frameworks = append(frameworks, "Chi")
		}
		if strings.Contains(contentStr, "gorilla/mux") {
			frameworks = append(frameworks, "Gorilla Mux")
		}
	}

	return frameworks
}

// detectConventions detects coding conventions
func (s *AppProfilerService) detectConventions(rootPath string, opts *ProfileOptions) CodeConventions {
	conv := CodeConventions{
		CommonPatterns: make([]string, 0),
	}

	// Sample a few files to detect conventions
	var sampleContent string
	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if ext == ".go" || ext == ".ts" || ext == ".js" || ext == ".svelte" {
			content, _ := os.ReadFile(path)
			if len(sampleContent) < 10000 {
				sampleContent += string(content)
			}
		}
		return nil
	})

	// Detect naming style
	if regexp.MustCompile(`[a-z]+_[a-z]+`).MatchString(sampleContent) {
		conv.NamingStyle = "snake_case"
	} else if regexp.MustCompile(`[a-z]+[A-Z][a-z]+`).MatchString(sampleContent) {
		conv.NamingStyle = "camelCase"
	}

	// Detect indent style
	if strings.Contains(sampleContent, "\t") {
		conv.IndentStyle = "tabs"
	} else {
		conv.IndentStyle = "spaces"
		// Detect indent size
		if regexp.MustCompile(`\n {4}[^ ]`).MatchString(sampleContent) {
			conv.IndentSize = 4
		} else if regexp.MustCompile(`\n {2}[^ ]`).MatchString(sampleContent) {
			conv.IndentSize = 2
		}
	}

	// Detect quote style
	singleQuotes := strings.Count(sampleContent, "'")
	doubleQuotes := strings.Count(sampleContent, "\"")
	if singleQuotes > doubleQuotes {
		conv.QuoteStyle = "single"
	} else {
		conv.QuoteStyle = "double"
	}

	// Detect semicolons
	conv.Semicolons = strings.Contains(sampleContent, ";")

	// Detect common patterns
	if strings.Contains(sampleContent, "interface") {
		conv.CommonPatterns = append(conv.CommonPatterns, "interfaces")
	}
	if strings.Contains(sampleContent, "async") {
		conv.CommonPatterns = append(conv.CommonPatterns, "async/await")
	}
	if strings.Contains(sampleContent, "struct") {
		conv.CommonPatterns = append(conv.CommonPatterns, "structs")
	}

	return conv
}

// analyzeComponents analyzes UI components
func (s *AppProfilerService) analyzeComponents(rootPath string, stack TechStack, opts *ProfileOptions) []ComponentInfo {
	components := make([]ComponentInfo, 0)

	// Determine component patterns based on stack
	var patterns []string
	var componentRegex *regexp.Regexp

	if containsAny(stack.Frontend, "Svelte", "SvelteKit") {
		patterns = []string{"*.svelte"}
		componentRegex = regexp.MustCompile(`export\s+let\s+(\w+)`)
	} else if containsAny(stack.Frontend, "React", "Next.js") {
		patterns = []string{"*.tsx", "*.jsx"}
		componentRegex = regexp.MustCompile(`(?:interface|type)\s+\w*Props`)
	} else if containsAny(stack.Frontend, "Vue") {
		patterns = []string{"*.vue"}
	}

	for _, pattern := range patterns {
		filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			// Check exclusions
			for _, exclude := range opts.ExcludePatterns {
				if strings.Contains(path, exclude) {
					return nil
				}
			}

			matched, _ := filepath.Match(pattern, d.Name())
			if !matched {
				return nil
			}

			content, _ := os.ReadFile(path)
			contentStr := string(content)

			comp := ComponentInfo{
				Name:     strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())),
				FilePath: path,
				Lines:    len(strings.Split(contentStr, "\n")),
				Props:    make([]string, 0),
				Events:   make([]string, 0),
				UsedIn:   make([]string, 0),
			}

			// Determine component type
			relPath, _ := filepath.Rel(rootPath, path)
			if strings.Contains(relPath, "page") || strings.Contains(relPath, "routes") {
				comp.Type = "page"
			} else if strings.Contains(relPath, "layout") {
				comp.Type = "layout"
			} else {
				comp.Type = "component"
			}

			// Extract props
			if componentRegex != nil {
				matches := componentRegex.FindAllStringSubmatch(contentStr, -1)
				for _, match := range matches {
					if len(match) > 1 {
						comp.Props = append(comp.Props, match[1])
					}
				}
			}

			components = append(components, comp)
			return nil
		})
	}

	return components
}

// analyzeModules analyzes code modules
func (s *AppProfilerService) analyzeModules(rootPath string, languages []LanguageInfo, opts *ProfileOptions) []ModuleInfo {
	modules := make([]ModuleInfo, 0)

	// Determine primary language
	primaryLang := ""
	if len(languages) > 0 {
		primaryLang = languages[0].Name
	}

	var patterns []string
	switch primaryLang {
	case "Go":
		patterns = []string{"*.go"}
	case "TypeScript", "JavaScript":
		patterns = []string{"*.ts", "*.js"}
	case "Python":
		patterns = []string{"*.py"}
	}

	for _, pattern := range patterns {
		filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			// Check exclusions
			for _, exclude := range opts.ExcludePatterns {
				if strings.Contains(path, exclude) {
					return nil
				}
			}

			matched, _ := filepath.Match(pattern, d.Name())
			if !matched {
				return nil
			}

			content, _ := os.ReadFile(path)
			contentStr := string(content)

			mod := ModuleInfo{
				Name:    strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())),
				Path:    path,
				Lines:   len(strings.Split(contentStr, "\n")),
				Exports: make([]string, 0),
				Imports: make([]string, 0),
			}

			// Determine module type
			relPath, _ := filepath.Rel(rootPath, path)
			relPathLower := strings.ToLower(relPath)
			switch {
			case strings.Contains(relPathLower, "handler"):
				mod.Type = "handler"
			case strings.Contains(relPathLower, "service"):
				mod.Type = "service"
			case strings.Contains(relPathLower, "repository") || strings.Contains(relPathLower, "repo"):
				mod.Type = "repository"
			case strings.Contains(relPathLower, "util") || strings.Contains(relPathLower, "helper"):
				mod.Type = "utility"
			case strings.Contains(relPathLower, "model") || strings.Contains(relPathLower, "entity"):
				mod.Type = "model"
			case strings.Contains(relPathLower, "middleware"):
				mod.Type = "middleware"
			default:
				mod.Type = "module"
			}

			// Extract exports (simplified)
			if primaryLang == "Go" {
				exportRegex := regexp.MustCompile(`func\s+([A-Z]\w+)`)
				matches := exportRegex.FindAllStringSubmatch(contentStr, -1)
				for _, match := range matches {
					if len(match) > 1 {
						mod.Exports = append(mod.Exports, match[1])
					}
				}
			}

			modules = append(modules, mod)
			return nil
		})
	}

	return modules
}

// analyzeEndpoints analyzes API endpoints
func (s *AppProfilerService) analyzeEndpoints(rootPath string, stack TechStack, opts *ProfileOptions) []APIEndpointInfo {
	endpoints := make([]APIEndpointInfo, 0)

	// Go patterns
	goEndpointPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:Get|Post|Put|Delete|Patch|Handle)\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`r\.(?:Get|Post|Put|Delete|Patch)\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`\.(?:GET|POST|PUT|DELETE|PATCH)\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`HandleFunc\s*\(\s*"([^"]+)"`),
	}

	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Check exclusions
		for _, exclude := range opts.ExcludePatterns {
			if strings.Contains(path, exclude) {
				return nil
			}
		}

		ext := filepath.Ext(d.Name())
		if ext != ".go" && ext != ".ts" && ext != ".js" {
			return nil
		}

		content, _ := os.ReadFile(path)
		contentStr := string(content)

		for _, pattern := range goEndpointPatterns {
			matches := pattern.FindAllStringSubmatch(contentStr, -1)
			for _, match := range matches {
				if len(match) > 1 {
					endpointPath := match[1]
					method := "GET"

					// Infer method from pattern
					fullMatch := match[0]
					if strings.Contains(strings.ToLower(fullMatch), "post") {
						method = "POST"
					} else if strings.Contains(strings.ToLower(fullMatch), "put") {
						method = "PUT"
					} else if strings.Contains(strings.ToLower(fullMatch), "delete") {
						method = "DELETE"
					} else if strings.Contains(strings.ToLower(fullMatch), "patch") {
						method = "PATCH"
					}

					endpoints = append(endpoints, APIEndpointInfo{
						Method:      method,
						Path:        endpointPath,
						HandlerPath: path,
					})
				}
			}
		}

		return nil
	})

	return endpoints
}

// analyzeDatabaseSchema analyzes database schema from migrations
func (s *AppProfilerService) analyzeDatabaseSchema(rootPath string) *DatabaseSchemaInfo {
	schema := &DatabaseSchemaInfo{
		Tables:     make([]TableInfo, 0),
		Migrations: make([]MigrationInfo, 0),
	}

	// Look for migrations directory
	migrationPaths := []string{
		filepath.Join(rootPath, "migrations"),
		filepath.Join(rootPath, "db", "migrations"),
		filepath.Join(rootPath, "internal", "database", "migrations"),
	}

	for _, migPath := range migrationPaths {
		if _, err := os.Stat(migPath); err == nil {
			files, _ := os.ReadDir(migPath)
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".sql") {
					schema.Migrations = append(schema.Migrations, MigrationInfo{
						Name: f.Name(),
						Path: filepath.Join(migPath, f.Name()),
					})

					// Parse SQL for tables
					content, _ := os.ReadFile(filepath.Join(migPath, f.Name()))
					tables := s.parseTablesFromSQL(string(content))
					schema.Tables = append(schema.Tables, tables...)
				}
			}
			break
		}
	}

	schema.TotalTables = len(schema.Tables)
	return schema
}

// parseTablesFromSQL extracts table information from SQL
func (s *AppProfilerService) parseTablesFromSQL(sql string) []TableInfo {
	tables := make([]TableInfo, 0)

	tableRegex := regexp.MustCompile(`CREATE TABLE\s+(?:IF NOT EXISTS\s+)?(\w+)`)
	matches := tableRegex.FindAllStringSubmatch(sql, -1)

	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, TableInfo{
				Name:    match[1],
				Columns: make([]ColumnInfo, 0),
			})
		}
	}

	return tables
}

// extractReadmeSummary extracts summary from README
func (s *AppProfilerService) extractReadmeSummary(rootPath string) string {
	readmeFiles := []string{"README.md", "readme.md", "README", "README.txt"}

	for _, readme := range readmeFiles {
		path := filepath.Join(rootPath, readme)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// Extract first paragraph or section
		lines := strings.Split(string(content), "\n")
		var summary strings.Builder
		inContent := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Skip title
			if strings.HasPrefix(trimmed, "#") && !inContent {
				continue
			}

			// Skip empty lines at start
			if trimmed == "" && !inContent {
				continue
			}

			inContent = true

			// Stop at next heading or after enough content
			if strings.HasPrefix(trimmed, "#") || summary.Len() > 500 {
				break
			}

			summary.WriteString(trimmed)
			summary.WriteString(" ")
		}

		return strings.TrimSpace(summary.String())
	}

	return ""
}

// detectIntegrations detects external integrations
func (s *AppProfilerService) detectIntegrations(rootPath string) []IntegrationPoint {
	integrations := make([]IntegrationPoint, 0)

	// Check for common integration patterns
	envFile := filepath.Join(rootPath, ".env.example")
	content, err := os.ReadFile(envFile)
	if err != nil {
		envFile = filepath.Join(rootPath, ".env.sample")
		content, err = os.ReadFile(envFile)
	}

	if err == nil {
		contentStr := string(content)

		// Database URLs
		if strings.Contains(contentStr, "DATABASE_URL") {
			integrations = append(integrations, IntegrationPoint{
				Name: "Database",
				Type: "database",
			})
		}

		// Redis
		if strings.Contains(contentStr, "REDIS") {
			integrations = append(integrations, IntegrationPoint{
				Name: "Redis",
				Type: "database",
			})
		}

		// OpenAI/Anthropic
		if strings.Contains(contentStr, "OPENAI") || strings.Contains(contentStr, "ANTHROPIC") {
			integrations = append(integrations, IntegrationPoint{
				Name: "AI Provider",
				Type: "api",
			})
		}

		// Stripe
		if strings.Contains(contentStr, "STRIPE") {
			integrations = append(integrations, IntegrationPoint{
				Name: "Stripe",
				Type: "api",
			})
		}
	}

	return integrations
}

// generateDescription generates a description for the profile
func (s *AppProfilerService) generateDescription(profile *ApplicationProfile) string {
	var desc strings.Builder

	desc.WriteString(fmt.Sprintf("A %s application", profile.AppType))

	if len(profile.Languages) > 0 {
		langs := make([]string, 0)
		for i, l := range profile.Languages {
			if i >= 3 {
				break
			}
			langs = append(langs, l.Name)
		}
		desc.WriteString(fmt.Sprintf(" built with %s", strings.Join(langs, ", ")))
	}

	if len(profile.Frameworks) > 0 {
		desc.WriteString(fmt.Sprintf(" using %s", strings.Join(profile.Frameworks, ", ")))
	}

	desc.WriteString(fmt.Sprintf(". Contains %d files with %d lines of code.", profile.FileCount, profile.LinesOfCode))

	return desc.String()
}

// saveProfile saves the application profile to the database
func (s *AppProfilerService) saveProfile(ctx context.Context, profile *ApplicationProfile) error {
	techStackJSON, _ := json.Marshal(profile.TechStack)
	languagesJSON, _ := json.Marshal(profile.Languages)
	structureJSON, _ := json.Marshal(profile.StructureTree)
	componentsJSON, _ := json.Marshal(profile.Components)
	modulesJSON, _ := json.Marshal(profile.Modules)
	endpointsJSON, _ := json.Marshal(profile.APIEndpoints)
	dbSchemaJSON, _ := json.Marshal(profile.DatabaseSchema)
	conventionsJSON, _ := json.Marshal(profile.Conventions)
	integrationsJSON, _ := json.Marshal(profile.IntegrationPoints)
	metadataJSON, _ := json.Marshal(profile.Metadata)
	frameworksJSON, _ := json.Marshal(profile.Frameworks)

	// Try upsert first. Some environments may not have the unique constraint yet;
	// if ON CONFLICT fails, fall back to update-or-insert.
	_, err := s.pool.Exec(ctx,
		`INSERT INTO application_profiles
		 (id, user_id, name, description, root_path, app_type, tech_stack, languages, frameworks,
		  structure_tree, components, total_components, modules, total_modules, api_endpoints,
		  total_endpoints, database_schema, conventions, integration_points, readme_summary,
		  lines_of_code, file_count, last_analyzed_at, metadata, created_at, updated_at,
		  auto_sync_enabled, last_synced_at, sync_source, sync_branch, sync_commit)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31)
		 ON CONFLICT (user_id, name) DO UPDATE SET
		    description = EXCLUDED.description,
		    root_path = EXCLUDED.root_path,
		    app_type = EXCLUDED.app_type,
		    tech_stack = EXCLUDED.tech_stack,
		    languages = EXCLUDED.languages,
		    frameworks = EXCLUDED.frameworks,
		    structure_tree = EXCLUDED.structure_tree,
		    components = EXCLUDED.components,
		    total_components = EXCLUDED.total_components,
		    modules = EXCLUDED.modules,
		    total_modules = EXCLUDED.total_modules,
		    api_endpoints = EXCLUDED.api_endpoints,
		    total_endpoints = EXCLUDED.total_endpoints,
		    database_schema = EXCLUDED.database_schema,
		    conventions = EXCLUDED.conventions,
		    integration_points = EXCLUDED.integration_points,
		    readme_summary = EXCLUDED.readme_summary,
		    lines_of_code = EXCLUDED.lines_of_code,
		    file_count = EXCLUDED.file_count,
		    last_analyzed_at = EXCLUDED.last_analyzed_at,
		    metadata = EXCLUDED.metadata,
		    auto_sync_enabled = EXCLUDED.auto_sync_enabled,
		    last_synced_at = EXCLUDED.last_synced_at,
		    sync_source = EXCLUDED.sync_source,
		    sync_branch = EXCLUDED.sync_branch,
		    sync_commit = EXCLUDED.sync_commit,
		    updated_at = EXCLUDED.updated_at`,
		profile.ID, profile.UserID, profile.Name, profile.Description, profile.RootPath,
		string(profile.AppType), techStackJSON, languagesJSON, frameworksJSON, structureJSON,
		componentsJSON, profile.TotalComponents, modulesJSON, profile.TotalModules,
		endpointsJSON, profile.TotalEndpoints, dbSchemaJSON, conventionsJSON,
		integrationsJSON, profile.ReadmeSummary, profile.LinesOfCode, profile.FileCount,
		profile.LastAnalyzedAt, metadataJSON, profile.CreatedAt, profile.UpdatedAt,
		profile.AutoSyncEnabled, profile.LastSyncedAt, profile.SyncSource, profile.SyncBranch, profile.SyncCommit,
	)
	if err == nil {
		return nil
	}

	// Fallback path: update existing row first.
	cmdTag, updErr := s.pool.Exec(ctx,
		`UPDATE application_profiles
		 SET description=$3, root_path=$4, app_type=$5, tech_stack=$6, languages=$7, frameworks=$8,
		     structure_tree=$9, components=$10, total_components=$11, modules=$12, total_modules=$13,
		     api_endpoints=$14, total_endpoints=$15, database_schema=$16, conventions=$17,
		     integration_points=$18, readme_summary=$19, lines_of_code=$20, file_count=$21,
		     last_analyzed_at=$22, metadata=$23, auto_sync_enabled=$24, last_synced_at=$25,
		     sync_source=$26, sync_branch=$27, sync_commit=$28, updated_at=$29
		 WHERE user_id=$1 AND name=$2`,
		profile.UserID, profile.Name,
		profile.Description, profile.RootPath, string(profile.AppType), techStackJSON, languagesJSON, frameworksJSON,
		structureJSON, componentsJSON, profile.TotalComponents, modulesJSON, profile.TotalModules,
		endpointsJSON, profile.TotalEndpoints, dbSchemaJSON, conventionsJSON, integrationsJSON,
		profile.ReadmeSummary, profile.LinesOfCode, profile.FileCount, profile.LastAnalyzedAt,
		metadataJSON, profile.AutoSyncEnabled, profile.LastSyncedAt, profile.SyncSource, profile.SyncBranch, profile.SyncCommit,
		profile.UpdatedAt,
	)
	if updErr == nil && cmdTag.RowsAffected() > 0 {
		return nil
	}

	// If update didn't touch anything, insert without ON CONFLICT.
	_, insErr := s.pool.Exec(ctx,
		`INSERT INTO application_profiles
		 (id, user_id, name, description, root_path, app_type, tech_stack, languages, frameworks,
		  structure_tree, components, total_components, modules, total_modules, api_endpoints,
		  total_endpoints, database_schema, conventions, integration_points, readme_summary,
		  lines_of_code, file_count, last_analyzed_at, metadata, created_at, updated_at,
		  auto_sync_enabled, last_synced_at, sync_source, sync_branch, sync_commit)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31)`,
		profile.ID, profile.UserID, profile.Name, profile.Description, profile.RootPath,
		string(profile.AppType), techStackJSON, languagesJSON, frameworksJSON, structureJSON,
		componentsJSON, profile.TotalComponents, modulesJSON, profile.TotalModules,
		endpointsJSON, profile.TotalEndpoints, dbSchemaJSON, conventionsJSON,
		integrationsJSON, profile.ReadmeSummary, profile.LinesOfCode, profile.FileCount,
		profile.LastAnalyzedAt, metadataJSON, profile.CreatedAt, profile.UpdatedAt,
		profile.AutoSyncEnabled, profile.LastSyncedAt, profile.SyncSource, profile.SyncBranch, profile.SyncCommit,
	)
	if insErr != nil {
		// Return original upsert error as it's usually the most informative.
		return err
	}
	return nil
}

type AutoSyncProfileTarget struct {
	UserID      string
	Name        string
	RootPath    string
	SyncCommit  string
	SyncSource  string
	SyncBranch  string
}

// ListAutoSyncTargets returns lightweight sync targets for the background job.
func (s *AppProfilerService) ListAutoSyncTargets(ctx context.Context, limit int) ([]AutoSyncProfileTarget, error) {
	if limit <= 0 {
		limit = 5
	}

	rows, err := s.pool.Query(ctx, `
		SELECT user_id, name, root_path, COALESCE(sync_commit, ''), COALESCE(sync_source, ''), COALESCE(sync_branch, '')
		FROM application_profiles
		WHERE auto_sync_enabled = true
		  AND root_path IS NOT NULL
		  AND root_path <> ''
		ORDER BY COALESCE(last_synced_at, last_analyzed_at, updated_at) ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []AutoSyncProfileTarget
	for rows.Next() {
		var t AutoSyncProfileTarget
		if err := rows.Scan(&t.UserID, &t.Name, &t.RootPath, &t.SyncCommit, &t.SyncSource, &t.SyncBranch); err != nil {
			continue
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (s *AppProfilerService) UpdateSyncInfo(ctx context.Context, userID, name string, enabled bool, source, branch, commit string, syncedAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE application_profiles
		SET auto_sync_enabled = $3,
		    sync_source = $4,
		    sync_branch = $5,
		    sync_commit = $6,
		    last_synced_at = $7,
		    updated_at = NOW()
		WHERE user_id = $1 AND name = $2
	`, userID, name, enabled, source, branch, commit, syncedAt)
	return err
}

func (s *AppProfilerService) detectGitInfo(ctx context.Context, rootPath string) (branch string, commit string, ok bool) {
	gitDir := filepath.Join(rootPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return "", "", false
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	commitOut, err := exec.CommandContext(cmdCtx, "git", "-C", rootPath, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", "", false
	}
	commit = strings.TrimSpace(string(commitOut))

	branchOut, err := exec.CommandContext(cmdCtx, "git", "-C", rootPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}
	return branch, commit, true
}

func (s *AppProfilerService) detectFilesystemFingerprint(rootPath string, opts *ProfileOptions) (string, error) {
	if opts == nil {
		opts = DefaultProfileOptions()
	}
	latest := time.Time{}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip excluded directories/files
		rel, relErr := filepath.Rel(rootPath, path)
		if relErr == nil {
			rel = filepath.ToSlash(rel)
			for _, ex := range opts.ExcludePatterns {
				ex = strings.TrimSpace(ex)
				if ex == "" {
					continue
				}
				if strings.Contains(rel, ex) {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		mt := info.ModTime()
		if mt.After(latest) {
			latest = mt
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if latest.IsZero() {
		return "fs:0", nil
	}
	return fmt.Sprintf("fs:%d", latest.UTC().UnixNano()), nil
}

// SyncAutoProfiles runs one batch of auto-sync work.
// It refreshes profiles whose git HEAD (preferred) or filesystem fingerprint changed.
func (s *AppProfilerService) SyncAutoProfiles(ctx context.Context, limit int) (checked int, refreshed int, err error) {
	targets, err := s.ListAutoSyncTargets(ctx, limit)
	if err != nil {
		return 0, 0, err
	}

	for _, t := range targets {
		checked++
		// Detect change fingerprint
		branch, commit, ok := s.detectGitInfo(ctx, t.RootPath)
		source := t.RootPath
		fingerprint := ""
		if ok {
			fingerprint = commit
		} else {
			fp, fpErr := s.detectFilesystemFingerprint(t.RootPath, nil)
			if fpErr != nil {
				s.logger.Warn("auto-sync fingerprint failed", "user", t.UserID, "name", t.Name, "error", fpErr)
				continue
			}
			fingerprint = fp
		}

		if fingerprint != "" && fingerprint == t.SyncCommit {
			continue
		}

		// Refresh profile
		_, profErr := s.ProfileApplication(ctx, t.UserID, t.RootPath, t.Name, nil)
		if profErr != nil {
			s.logger.Warn("auto-sync profile refresh failed", "user", t.UserID, "name", t.Name, "error", profErr)
			continue
		}
		syncedAt := time.Now().UTC()
		if updErr := s.UpdateSyncInfo(ctx, t.UserID, t.Name, true, source, branch, fingerprint, syncedAt); updErr != nil {
			s.logger.Warn("auto-sync update sync info failed", "user", t.UserID, "name", t.Name, "error", updErr)
			continue
		}
		refreshed++
	}

	return checked, refreshed, nil
}

// GetProfile retrieves an application profile
func (s *AppProfilerService) GetProfile(ctx context.Context, userID, name string) (*ApplicationProfile, error) {
	var profile ApplicationProfile
	var techStackJSON, languagesJSON, structureJSON, componentsJSON []byte
	var modulesJSON, endpointsJSON, dbSchemaJSON, conventionsJSON []byte
	var integrationsJSON, metadataJSON, frameworksJSON []byte

	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, name, description, root_path, app_type, tech_stack, languages, frameworks,
		        structure_tree, components, total_components, modules, total_modules, api_endpoints,
		        total_endpoints, database_schema, conventions, integration_points, readme_summary,
		        lines_of_code, file_count, last_analyzed_at, metadata, created_at, updated_at
		 FROM application_profiles WHERE user_id = $1 AND name = $2`,
		userID, name).Scan(
		&profile.ID, &profile.UserID, &profile.Name, &profile.Description, &profile.RootPath,
		&profile.AppType, &techStackJSON, &languagesJSON, &frameworksJSON, &structureJSON,
		&componentsJSON, &profile.TotalComponents, &modulesJSON, &profile.TotalModules,
		&endpointsJSON, &profile.TotalEndpoints, &dbSchemaJSON, &conventionsJSON,
		&integrationsJSON, &profile.ReadmeSummary, &profile.LinesOfCode, &profile.FileCount,
		&profile.LastAnalyzedAt, &metadataJSON, &profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(techStackJSON, &profile.TechStack)
	json.Unmarshal(languagesJSON, &profile.Languages)
	json.Unmarshal(frameworksJSON, &profile.Frameworks)
	json.Unmarshal(structureJSON, &profile.StructureTree)
	json.Unmarshal(componentsJSON, &profile.Components)
	json.Unmarshal(modulesJSON, &profile.Modules)
	json.Unmarshal(endpointsJSON, &profile.APIEndpoints)
	json.Unmarshal(dbSchemaJSON, &profile.DatabaseSchema)
	json.Unmarshal(conventionsJSON, &profile.Conventions)
	json.Unmarshal(integrationsJSON, &profile.IntegrationPoints)
	json.Unmarshal(metadataJSON, &profile.Metadata)

	return &profile, nil
}

// ListProfiles lists all profiles for a user
func (s *AppProfilerService) ListProfiles(ctx context.Context, userID string) ([]ApplicationProfile, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name, description, app_type, lines_of_code, file_count, last_analyzed_at, created_at
		 FROM application_profiles WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]ApplicationProfile, 0)
	for rows.Next() {
		var p ApplicationProfile
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.AppType,
			&p.LinesOfCode, &p.FileCount, &p.LastAnalyzedAt, &p.CreatedAt)
		if err != nil {
			continue
		}
		profiles = append(profiles, p)
	}

	return profiles, nil
}

// Helper function
func containsAny(slice []string, items ...string) bool {
	for _, s := range slice {
		for _, item := range items {
			if strings.Contains(strings.ToLower(s), strings.ToLower(item)) {
				return true
			}
		}
	}
	return false
}
