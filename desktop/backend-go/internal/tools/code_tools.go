package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CodeToolRegistry manages code-related tools for the Coding Agent
type CodeToolRegistry struct {
	workspaceRoot string
	backupMgr     *BackupManager
	allowedCmds   map[string]bool
}

// NewCodeToolRegistry creates a new code tool registry
func NewCodeToolRegistry(workspaceRoot string) *CodeToolRegistry {
	return &CodeToolRegistry{
		workspaceRoot: workspaceRoot,
		backupMgr:     NewBackupManager(workspaceRoot),
		allowedCmds: map[string]bool{
			"go":     true,
			"npm":    true,
			"node":   true,
			"git":    true,
			"cat":    true,
			"ls":     true,
			"find":   true,
			"grep":   true,
			"head":   true,
			"tail":   true,
			"wc":     true,
			"diff":   true,
			"make":   true,
			"cargo":  true,
			"python": true,
			"pip":    true,
			"yarn":   true,
			"pnpm":   true,
			"tsc":    true,
			"eslint": true,
		},
	}
}

// ========================================
// READ FILE TOOL
// ========================================

type ReadFileInput struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
}

type ReadFileOutput struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	TotalLines int    `json:"total_lines"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
}

func (r *CodeToolRegistry) ReadFile(ctx context.Context, input json.RawMessage) (string, error) {
	var params ReadFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	absPath, err := r.resolveAndValidatePath(params.Path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", params.Path)
		}
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	// Handle line range
	startLine := 1
	endLine := totalLines

	if params.StartLine > 0 {
		startLine = params.StartLine
	}
	if params.EndLine > 0 {
		endLine = params.EndLine
	}

	// Clamp values
	if startLine < 1 {
		startLine = 1
	}
	if endLine > totalLines {
		endLine = totalLines
	}
	if startLine > endLine {
		startLine = endLine
	}

	// Build output with line numbers
	var sb strings.Builder
	for i := startLine - 1; i < endLine && i < len(lines); i++ {
		sb.WriteString(fmt.Sprintf("%4d | %s\n", i+1, lines[i]))
	}

	output := ReadFileOutput{
		Path:       params.Path,
		Content:    sb.String(),
		TotalLines: totalLines,
		StartLine:  startLine,
		EndLine:    endLine,
	}

	result, _ := json.Marshal(output)
	return string(result), nil
}

// ========================================
// WRITE FILE TOOL
// ========================================

type WriteFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WriteFileOutput struct {
	Path         string `json:"path"`
	BackupPath   string `json:"backup_path,omitempty"`
	BytesWritten int    `json:"bytes_written"`
	LinesWritten int    `json:"lines_written"`
}

func (r *CodeToolRegistry) WriteFile(ctx context.Context, input json.RawMessage) (string, error) {
	var params WriteFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	absPath, err := r.resolveAndValidatePath(params.Path)
	if err != nil {
		return "", err
	}

	// Create backup if file exists
	var backupPath string
	if _, err := os.Stat(absPath); err == nil {
		backup, backupErr := r.backupMgr.CreateBackup(params.Path)
		if backupErr == nil {
			backupPath = backup.BackupPath
		}
	}

	// Create directory if needed
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(absPath, []byte(params.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	lines := strings.Count(params.Content, "\n") + 1

	output := WriteFileOutput{
		Path:         params.Path,
		BackupPath:   backupPath,
		BytesWritten: len(params.Content),
		LinesWritten: lines,
	}

	result, _ := json.Marshal(output)
	return string(result), nil
}

// ========================================
// EDIT FILE TOOL
// ========================================

type EditFileInput struct {
	Path       string `json:"path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	Occurrence int    `json:"occurrence,omitempty"` // 0 = all, 1 = first (default), 2 = second, etc.
}

type EditFileOutput struct {
	Path         string `json:"path"`
	BackupPath   string `json:"backup_path,omitempty"`
	Replacements int    `json:"replacements"`
	Diff         string `json:"diff"`
}

func (r *CodeToolRegistry) EditFile(ctx context.Context, input json.RawMessage) (string, error) {
	var params EditFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Path == "" {
		return "", fmt.Errorf("path is required")
	}
	if params.OldString == "" {
		return "", fmt.Errorf("old_string is required")
	}

	absPath, err := r.resolveAndValidatePath(params.Path)
	if err != nil {
		return "", err
	}

	// Read current content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	oldContent := string(content)

	// Check if old_string exists
	if !strings.Contains(oldContent, params.OldString) {
		return "", fmt.Errorf("old_string not found in file")
	}

	// Create backup before editing
	var backupPath string
	backup, backupErr := r.backupMgr.CreateBackup(params.Path)
	if backupErr == nil {
		backupPath = backup.BackupPath
	}

	// Perform replacement
	var newContent string
	var replacements int

	if params.Occurrence == 0 {
		// Replace all occurrences
		replacements = strings.Count(oldContent, params.OldString)
		newContent = strings.ReplaceAll(oldContent, params.OldString, params.NewString)
	} else {
		// Replace specific occurrence
		occurrence := params.Occurrence
		if occurrence < 1 {
			occurrence = 1
		}

		parts := strings.SplitN(oldContent, params.OldString, occurrence+1)
		if len(parts) <= occurrence {
			return "", fmt.Errorf("occurrence %d not found (only %d occurrences exist)", occurrence, len(parts)-1)
		}

		// Rebuild with replacement at specific occurrence
		newContent = strings.Join(parts[:occurrence], params.OldString) + params.NewString + parts[occurrence]
		replacements = 1
	}

	// Write new content
	if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Generate simple diff
	diff := generateSimpleDiff(params.OldString, params.NewString)

	output := EditFileOutput{
		Path:         params.Path,
		BackupPath:   backupPath,
		Replacements: replacements,
		Diff:         diff,
	}

	result, _ := json.Marshal(output)
	return string(result), nil
}

// ========================================
// SEARCH CODE TOOL
// ========================================

type SearchCodeInput struct {
	Query       string `json:"query"`
	Path        string `json:"path,omitempty"`
	FilePattern string `json:"file_pattern,omitempty"`
	MaxResults  int    `json:"max_results,omitempty"`
	Regex       bool   `json:"regex,omitempty"`
}

type SearchMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

type SearchCodeOutput struct {
	Query   string        `json:"query"`
	Matches []SearchMatch `json:"matches"`
	Total   int           `json:"total"`
}

func (r *CodeToolRegistry) SearchCode(ctx context.Context, input json.RawMessage) (string, error) {
	var params SearchCodeInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	searchPath := r.workspaceRoot
	if params.Path != "" {
		absPath, err := r.resolveAndValidatePath(params.Path)
		if err != nil {
			return "", err
		}
		searchPath = absPath
	}

	maxResults := 50
	if params.MaxResults > 0 {
		maxResults = params.MaxResults
	}

	var matches []SearchMatch
	var searchRegex *regexp.Regexp

	if params.Regex {
		var err error
		searchRegex, err = regexp.Compile(params.Query)
		if err != nil {
			return "", fmt.Errorf("invalid regex: %w", err)
		}
	}

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and hidden files
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip backup directory
		if strings.Contains(path, BackupDirName) {
			return nil
		}

		// Apply file pattern filter
		if params.FilePattern != "" {
			matched, _ := filepath.Match(params.FilePattern, info.Name())
			if !matched {
				return nil
			}
		}

		// Skip binary files (simple heuristic)
		if isBinaryFile(info.Name()) {
			return nil
		}

		// Search in file
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		relPath, _ := filepath.Rel(r.workspaceRoot, path)
		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			var found bool
			if params.Regex && searchRegex != nil {
				found = searchRegex.MatchString(line)
			} else {
				found = strings.Contains(line, params.Query)
			}

			if found {
				matches = append(matches, SearchMatch{
					File:    relPath,
					Line:    lineNum,
					Content: truncateString(line, 200),
				})

				if len(matches) >= maxResults {
					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("search failed: %w", err)
	}

	output := SearchCodeOutput{
		Query:   params.Query,
		Matches: matches,
		Total:   len(matches),
	}

	result, _ := json.Marshal(output)
	return string(result), nil
}

// ========================================
// LIST FILES TOOL
// ========================================

type ListFilesInput struct {
	Path      string `json:"path,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	Recursive bool   `json:"recursive,omitempty"`
	MaxDepth  int    `json:"max_depth,omitempty"`
}

type FileEntry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsDir    bool      `json:"is_dir"`
	Size     int64     `json:"size,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
}

type ListFilesOutput struct {
	Path  string      `json:"path"`
	Files []FileEntry `json:"files"`
	Total int         `json:"total"`
}

func (r *CodeToolRegistry) ListFiles(ctx context.Context, input json.RawMessage) (string, error) {
	var params ListFilesInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	searchPath := r.workspaceRoot
	if params.Path != "" {
		absPath, err := r.resolveAndValidatePath(params.Path)
		if err != nil {
			return "", err
		}
		searchPath = absPath
	}

	maxDepth := 1
	if params.Recursive {
		maxDepth = 10
	}
	if params.MaxDepth > 0 {
		maxDepth = params.MaxDepth
	}

	var files []FileEntry
	baseDepth := strings.Count(searchPath, string(filepath.Separator))

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Calculate depth
		currentDepth := strings.Count(path, string(filepath.Separator)) - baseDepth

		// Skip if too deep
		if currentDepth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip root
		if path == searchPath {
			return nil
		}

		// Skip hidden files/dirs
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply pattern filter
		if params.Pattern != "" {
			matched, _ := filepath.Match(params.Pattern, info.Name())
			if !matched && !info.IsDir() {
				return nil
			}
		}

		relPath, _ := filepath.Rel(r.workspaceRoot, path)

		entry := FileEntry{
			Name:     info.Name(),
			Path:     relPath,
			IsDir:    info.IsDir(),
			Modified: info.ModTime(),
		}

		if !info.IsDir() {
			entry.Size = info.Size()
		}

		files = append(files, entry)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	relPath, _ := filepath.Rel(r.workspaceRoot, searchPath)
	if relPath == "." {
		relPath = "/"
	}

	output := ListFilesOutput{
		Path:  relPath,
		Files: files,
		Total: len(files),
	}

	result, _ := json.Marshal(output)
	return string(result), nil
}

// ========================================
// RUN COMMAND TOOL
// ========================================

type RunCommandInput struct {
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir,omitempty"`
	Timeout    int    `json:"timeout,omitempty"` // seconds
}

type RunCommandOutput struct {
	Command  string `json:"command"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Duration string `json:"duration"`
}

func (r *CodeToolRegistry) RunCommand(ctx context.Context, input json.RawMessage) (string, error) {
	var params RunCommandInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.Command == "" {
		return "", fmt.Errorf("command is required")
	}

	// Parse command
	parts := strings.Fields(params.Command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Check if command is allowed
	cmdName := parts[0]
	if !r.allowedCmds[cmdName] {
		return "", fmt.Errorf("command not allowed: %s (allowed: go, npm, git, etc.)", cmdName)
	}

	// Set working directory
	workDir := r.workspaceRoot
	if params.WorkingDir != "" {
		absPath, err := r.resolveAndValidatePath(params.WorkingDir)
		if err != nil {
			return "", err
		}
		workDir = absPath
	}

	// Set timeout
	timeout := 30 * time.Second
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Second
	}
	if timeout > 5*time.Minute {
		timeout = 5 * time.Minute // Max 5 minutes
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	cmd.Dir = workDir

	// Capture output
	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if cmdCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", timeout)
		}
	}

	result := RunCommandOutput{
		Command:  params.Command,
		Output:   truncateString(string(output), 10000),
		ExitCode: exitCode,
		Duration: duration.String(),
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// ========================================
// HELPER FUNCTIONS
// ========================================

func (r *CodeToolRegistry) resolveAndValidatePath(path string) (string, error) {
	var absPath string
	if filepath.IsAbs(path) {
		absPath = filepath.Clean(path)
	} else {
		absPath = filepath.Clean(filepath.Join(r.workspaceRoot, path))
	}

	// Validate path is under workspace
	if !strings.HasPrefix(absPath, r.workspaceRoot) {
		return "", fmt.Errorf("path must be under workspace: %s", path)
	}

	// Check for path traversal
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", path)
	}

	return absPath, nil
}

func generateSimpleDiff(old, new string) string {
	var sb strings.Builder
	sb.WriteString("--- old\n+++ new\n")

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	for _, line := range oldLines {
		sb.WriteString("- " + line + "\n")
	}
	for _, line := range newLines {
		sb.WriteString("+ " + line + "\n")
	}

	return sb.String()
}

func isBinaryFile(name string) bool {
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
		".pdf": true, ".zip": true, ".tar": true, ".gz": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".db": true, ".sqlite": true,
	}
	ext := strings.ToLower(filepath.Ext(name))
	return binaryExts[ext]
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ========================================
// TOOL INTERFACE IMPLEMENTATIONS
// ========================================

// ReadFileTool implements AgentTool interface
type ReadFileTool struct {
	registry *CodeToolRegistry
}

func NewReadFileTool(workspaceRoot string) *ReadFileTool {
	return &ReadFileTool{registry: NewCodeToolRegistry(workspaceRoot)}
}

func (t *ReadFileTool) Name() string        { return "read_file" }
func (t *ReadFileTool) Description() string { return "Read file content with line numbers" }
func (t *ReadFileTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path":       map[string]interface{}{"type": "string", "description": "File path relative to workspace"},
			"start_line": map[string]interface{}{"type": "integer", "description": "Start line (1-indexed)"},
			"end_line":   map[string]interface{}{"type": "integer", "description": "End line (1-indexed)"},
		},
		"required": []string{"path"},
	}
}
func (t *ReadFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	return t.registry.ReadFile(ctx, input)
}

// WriteFileTool implements AgentTool interface
type WriteFileTool struct {
	registry *CodeToolRegistry
}

func NewWriteFileTool(workspaceRoot string) *WriteFileTool {
	return &WriteFileTool{registry: NewCodeToolRegistry(workspaceRoot)}
}

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Write content to file (creates backup automatically)"
}
func (t *WriteFileTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path":    map[string]interface{}{"type": "string", "description": "File path"},
			"content": map[string]interface{}{"type": "string", "description": "Full file content"},
		},
		"required": []string{"path", "content"},
	}
}
func (t *WriteFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	return t.registry.WriteFile(ctx, input)
}

// EditFileTool implements AgentTool interface
type EditFileTool struct {
	registry *CodeToolRegistry
}

func NewEditFileTool(workspaceRoot string) *EditFileTool {
	return &EditFileTool{registry: NewCodeToolRegistry(workspaceRoot)}
}

func (t *EditFileTool) Name() string        { return "edit_file" }
func (t *EditFileTool) Description() string { return "Make surgical edit to file (find and replace)" }
func (t *EditFileTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path":       map[string]interface{}{"type": "string", "description": "File path"},
			"old_string": map[string]interface{}{"type": "string", "description": "Exact text to find"},
			"new_string": map[string]interface{}{"type": "string", "description": "Replacement text"},
			"occurrence": map[string]interface{}{"type": "integer", "description": "Which occurrence (0=all, 1=first)"},
		},
		"required": []string{"path", "old_string", "new_string"},
	}
}
func (t *EditFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	return t.registry.EditFile(ctx, input)
}

// SearchCodeTool implements AgentTool interface
type SearchCodeTool struct {
	registry *CodeToolRegistry
}

func NewSearchCodeTool(workspaceRoot string) *SearchCodeTool {
	return &SearchCodeTool{registry: NewCodeToolRegistry(workspaceRoot)}
}

func (t *SearchCodeTool) Name() string        { return "search_code" }
func (t *SearchCodeTool) Description() string { return "Search for pattern in codebase" }
func (t *SearchCodeTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query":        map[string]interface{}{"type": "string", "description": "Search pattern"},
			"path":         map[string]interface{}{"type": "string", "description": "Directory to search"},
			"file_pattern": map[string]interface{}{"type": "string", "description": "Glob pattern (e.g., *.go)"},
			"max_results":  map[string]interface{}{"type": "integer", "description": "Max results"},
			"regex":        map[string]interface{}{"type": "boolean", "description": "Use regex"},
		},
		"required": []string{"query"},
	}
}
func (t *SearchCodeTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	return t.registry.SearchCode(ctx, input)
}

// ListFilesTool implements AgentTool interface
type ListFilesTool struct {
	registry *CodeToolRegistry
}

func NewListFilesTool(workspaceRoot string) *ListFilesTool {
	return &ListFilesTool{registry: NewCodeToolRegistry(workspaceRoot)}
}

func (t *ListFilesTool) Name() string        { return "list_files" }
func (t *ListFilesTool) Description() string { return "List files in directory" }
func (t *ListFilesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path":      map[string]interface{}{"type": "string", "description": "Directory path"},
			"pattern":   map[string]interface{}{"type": "string", "description": "Glob pattern filter"},
			"recursive": map[string]interface{}{"type": "boolean", "description": "Include subdirectories"},
			"max_depth": map[string]interface{}{"type": "integer", "description": "Max recursion depth"},
		},
	}
}
func (t *ListFilesTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	return t.registry.ListFiles(ctx, input)
}

// RunCommandTool implements AgentTool interface
type RunCommandTool struct {
	registry *CodeToolRegistry
}

func NewRunCommandTool(workspaceRoot string) *RunCommandTool {
	return &RunCommandTool{registry: NewCodeToolRegistry(workspaceRoot)}
}

func (t *RunCommandTool) Name() string { return "run_command" }
func (t *RunCommandTool) Description() string {
	return "Execute shell command in workspace (sandboxed)"
}
func (t *RunCommandTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command":     map[string]interface{}{"type": "string", "description": "Command to execute"},
			"working_dir": map[string]interface{}{"type": "string", "description": "Working directory"},
			"timeout":     map[string]interface{}{"type": "integer", "description": "Timeout in seconds"},
		},
		"required": []string{"command"},
	}
}
func (t *RunCommandTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	return t.registry.RunCommand(ctx, input)
}
