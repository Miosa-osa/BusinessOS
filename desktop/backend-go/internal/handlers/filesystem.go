package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rhl/businessos-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

const (
	// containerWorkspaceRoot is the root directory in containers
	containerWorkspaceRoot = "/workspace"
)

// FileItem represents a file or directory
type FileItem struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"` // "file" or "folder"
	Path      string     `json:"path"`
	Size      int64      `json:"size,omitempty"`
	Modified  *time.Time `json:"modified,omitempty"`
	Extension string     `json:"extension,omitempty"`
	IsHidden  bool       `json:"isHidden"`
}

// ListDirectoryRequest represents the request for listing a directory
type ListDirectoryRequest struct {
	Path       string `json:"path" form:"path"`
	ShowHidden bool   `json:"showHidden" form:"showHidden"`
	Mode       string `json:"mode" form:"mode"` // "docker", "local", or "" (use default)
}

// ListDirectoryResponse represents the response for listing a directory
type ListDirectoryResponse struct {
	Path      string     `json:"path"`
	Items     []FileItem `json:"items"`
	ParentDir string     `json:"parentDir,omitempty"`
}

// FileContentResponse represents the response for reading a file
type FileContentResponse struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

// ListDirectory lists the contents of a directory
func (h *Handlers) ListDirectory(c *gin.Context) {
	var req ListDirectoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Determine filesystem mode based on request or global setting
	// "local" or "glimpse" → force local filesystem (Mac)
	// "docker" → use container if available
	// "" (empty) → use global default

	useLocal := false
	if req.Mode == "local" || req.Mode == "glimpse" {
		log.Printf("[Filesystem] 🚫 Mode=%s requested - FORCING local filesystem", req.Mode)
		useLocal = true
	} else if req.Mode == "docker" {
		if h.containerMgr != nil {
			log.Printf("[Filesystem] 🐳 Mode=docker requested - using container")
			useLocal = false
		} else {
			log.Printf("[Filesystem] ⚠️  Mode=docker requested but no containerMgr - using local")
			useLocal = true
		}
	} else {
		// Default behavior: use container if available
		useLocal = (h.containerMgr == nil)
		log.Printf("[Filesystem] ⚙️  Mode=DEFAULT - containerMgr=%v, useLocal=%v", h.containerMgr != nil, useLocal)
	}

	if useLocal {
		h.listDirectoryLocal(c, req)
	} else {
		h.listDirectoryContainer(c, req)
	}
}

// listDirectoryContainer lists directory contents from user's container workspace
func (h *Handlers) listDirectoryContainer(c *gin.Context, req ListDirectoryRequest) {
	// Get user from context
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := user.ID

	// Get or create filesystem container for this user
	containerID, err := h.containerMgr.GetOrCreateFilesystemContainer(userIDStr)
	if err != nil {
		log.Printf("[Filesystem] Failed to get container for user %s: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access filesystem"})
		return
	}

	// Normalize path - default to workspace root
	dirPath := req.Path
	if dirPath == "" || dirPath == "~" || dirPath == "/" {
		dirPath = containerWorkspaceRoot
	}

	// Expand ~ to workspace root
	if strings.HasPrefix(dirPath, "~/") {
		dirPath = filepath.Join(containerWorkspaceRoot, dirPath[2:])
	}

	// Ensure path is under workspace
	if !strings.HasPrefix(dirPath, containerWorkspaceRoot) {
		dirPath = filepath.Join(containerWorkspaceRoot, dirPath)
	}

	// List directory in container
	files, err := h.containerMgr.ListDirectoryInContainer(containerID, dirPath)
	if err != nil {
		log.Printf("[Filesystem] Failed to list directory %s: %v", dirPath, err)
		if strings.Contains(err.Error(), "not a directory") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Path is not a directory"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	// Convert container.FileInfo to FileItem
	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		name := f.Name
		isHidden := strings.HasPrefix(name, ".")

		// Skip hidden files unless requested
		if isHidden && !req.ShowHidden {
			continue
		}

		// Skip the base directory entry itself
		if name == "" || name == "." || name == filepath.Base(dirPath) {
			continue
		}

		item := FileItem{
			ID:       generateFileID(f.Path),
			Name:     name,
			Path:     f.Path,
			IsHidden: isHidden,
			Modified: &f.ModTime,
		}

		if f.IsDir {
			item.Type = "folder"
		} else {
			item.Type = "file"
			item.Size = f.Size
			item.Extension = strings.TrimPrefix(filepath.Ext(name), ".")
		}

		items = append(items, item)
	}

	// Sort: folders first, then alphabetically (case-insensitive)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "folder"
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	// Get parent directory
	parentDir := filepath.Dir(dirPath)
	if parentDir == dirPath || parentDir == containerWorkspaceRoot {
		parentDir = containerWorkspaceRoot // Can't go above workspace
	}

	c.JSON(http.StatusOK, ListDirectoryResponse{
		Path:      dirPath,
		Items:     items,
		ParentDir: parentDir,
	})
}

// listDirectoryLocal lists directory contents from local filesystem (fallback)
func (h *Handlers) listDirectoryLocal(c *gin.Context, req ListDirectoryRequest) {
	// Default to home directory if no path specified
	dirPath := req.Path
	if dirPath == "" || dirPath == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get home directory"})
			return
		}
		dirPath = homeDir
	}

	// Expand ~ to home directory
	if strings.HasPrefix(dirPath, "~/") {
		homeDir, _ := os.UserHomeDir()
		dirPath = filepath.Join(homeDir, dirPath[2:])
	}

	// Clean the path to prevent path traversal attacks
	dirPath = filepath.Clean(dirPath)

	// Check if path exists and is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access directory"})
		return
	}

	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is not a directory"})
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
		return
	}

	items := make([]FileItem, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		isHidden := strings.HasPrefix(name, ".")

		// Skip hidden files unless requested
		if isHidden && !req.ShowHidden {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't access
		}

		item := FileItem{
			ID:       generateFileID(filepath.Join(dirPath, name)),
			Name:     name,
			Path:     filepath.Join(dirPath, name),
			IsHidden: isHidden,
		}

		if entry.IsDir() {
			item.Type = "folder"
		} else {
			item.Type = "file"
			item.Size = info.Size()
			item.Extension = strings.TrimPrefix(filepath.Ext(name), ".")
		}

		modTime := info.ModTime()
		item.Modified = &modTime

		items = append(items, item)
	}

	// Sort: folders first, then alphabetically (case-insensitive)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "folder"
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	// Get parent directory
	parentDir := filepath.Dir(dirPath)
	if parentDir == dirPath {
		parentDir = "" // At root
	}

	c.JSON(http.StatusOK, ListDirectoryResponse{
		Path:      dirPath,
		Items:     items,
		ParentDir: parentDir,
	})
}

// ReadFile reads the content of a file
func (h *Handlers) ReadFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Check if container isolation is enabled
	if h.containerMgr != nil {
		h.readFileContainer(c, filePath)
		return
	}

	// Fallback to local filesystem
	h.readFileLocal(c, filePath)
}

// readFileContainer reads file content from user's container workspace
func (h *Handlers) readFileContainer(c *gin.Context, filePath string) {
	// Get user from context
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := user.ID

	// Get or create filesystem container
	containerID, err := h.containerMgr.GetOrCreateFilesystemContainer(userIDStr)
	if err != nil {
		log.Printf("[Filesystem] Failed to get container for user %s: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access filesystem"})
		return
	}

	// Normalize path
	if strings.HasPrefix(filePath, "~/") {
		filePath = filepath.Join(containerWorkspaceRoot, filePath[2:])
	}
	if !strings.HasPrefix(filePath, containerWorkspaceRoot) {
		filePath = filepath.Join(containerWorkspaceRoot, filePath)
	}

	// Read file from container
	content, err := h.containerMgr.ReadFileFromContainer(containerID, filePath)
	if err != nil {
		log.Printf("[Filesystem] Failed to read file %s: %v", filePath, err)
		if strings.Contains(err.Error(), "file too large") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File too large for preview"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Detect MIME type
	mimeType := http.DetectContentType(content)

	c.JSON(http.StatusOK, FileContentResponse{
		Path:     filePath,
		Name:     filepath.Base(filePath),
		Content:  string(content),
		Size:     int64(len(content)),
		MimeType: mimeType,
	})
}

// readFileLocal reads file content from local filesystem (fallback)
func (h *Handlers) readFileLocal(c *gin.Context, filePath string) {
	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access file"})
		return
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is a directory, not a file"})
		return
	}

	// Limit file size for reading (10MB max for text preview)
	const maxReadSize = 10 * 1024 * 1024
	if info.Size() > maxReadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large for preview"})
		return
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Detect MIME type
	mimeType := http.DetectContentType(content)

	c.JSON(http.StatusOK, FileContentResponse{
		Path:     filePath,
		Name:     filepath.Base(filePath),
		Content:  string(content),
		Size:     info.Size(),
		MimeType: mimeType,
	})
}

// DownloadFile downloads a file
func (h *Handlers) DownloadFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access file"})
		return
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot download a directory"})
		return
	}

	// Serve the file
	c.File(filePath)
}

// GetFileInfo returns information about a file or directory
func (h *Handlers) GetFileInfo(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Path not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access path"})
		return
	}

	name := filepath.Base(filePath)
	modTime := info.ModTime()

	item := FileItem{
		ID:       generateFileID(filePath),
		Name:     name,
		Path:     filePath,
		IsHidden: strings.HasPrefix(name, "."),
		Modified: &modTime,
	}

	if info.IsDir() {
		item.Type = "folder"
		// Count items in directory
		entries, err := os.ReadDir(filePath)
		if err == nil {
			item.Size = int64(len(entries))
		}
	} else {
		item.Type = "file"
		item.Size = info.Size()
		item.Extension = strings.TrimPrefix(filepath.Ext(name), ".")
	}

	c.JSON(http.StatusOK, item)
}

// CreateDirectory creates a new directory
func (h *Handlers) CreateDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Sanitize directory name
	if strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid directory name"})
		return
	}

	// Check if container isolation is enabled
	if h.containerMgr != nil {
		h.createDirectoryContainer(c, req.Path, req.Name)
		return
	}

	// Fallback to local filesystem
	h.createDirectoryLocal(c, req.Path, req.Name)
}

// createDirectoryContainer creates a directory in user's container workspace
func (h *Handlers) createDirectoryContainer(c *gin.Context, basePath, name string) {
	// Get user from context
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := user.ID

	// Get or create filesystem container
	containerID, err := h.containerMgr.GetOrCreateFilesystemContainer(userIDStr)
	if err != nil {
		log.Printf("[Filesystem] Failed to get container for user %s: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access filesystem"})
		return
	}

	// Normalize path
	if strings.HasPrefix(basePath, "~/") {
		basePath = filepath.Join(containerWorkspaceRoot, basePath[2:])
	}
	if !strings.HasPrefix(basePath, containerWorkspaceRoot) {
		basePath = filepath.Join(containerWorkspaceRoot, basePath)
	}

	newDirPath := filepath.Join(basePath, name)

	// Create directory in container
	if err := h.containerMgr.CreateDirectoryInContainer(containerID, newDirPath); err != nil {
		log.Printf("[Filesystem] Failed to create directory %s: %v", newDirPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	now := time.Now()
	c.JSON(http.StatusCreated, FileItem{
		ID:       generateFileID(newDirPath),
		Name:     name,
		Type:     "folder",
		Path:     newDirPath,
		IsHidden: strings.HasPrefix(name, "."),
		Modified: &now,
	})
}

// createDirectoryLocal creates a directory on local filesystem (fallback)
func (h *Handlers) createDirectoryLocal(c *gin.Context, basePath, name string) {
	// Expand ~ to home directory
	if strings.HasPrefix(basePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		basePath = filepath.Join(homeDir, basePath[2:])
	}

	newDirPath := filepath.Join(basePath, name)

	// Check if already exists
	if _, err := os.Stat(newDirPath); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Directory already exists"})
		return
	}

	// Create directory
	if err := os.MkdirAll(newDirPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Return info about new directory
	info, _ := os.Stat(newDirPath)
	modTime := info.ModTime()

	c.JSON(http.StatusCreated, FileItem{
		ID:       generateFileID(newDirPath),
		Name:     name,
		Type:     "folder",
		Path:     newDirPath,
		IsHidden: strings.HasPrefix(name, "."),
		Modified: &modTime,
	})
}

// DeleteFile deletes a file or empty directory
func (h *Handlers) DeleteFileOrDir(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Check if container isolation is enabled
	if h.containerMgr != nil {
		h.deleteFileOrDirContainer(c, filePath)
		return
	}

	// Fallback to local filesystem
	h.deleteFileOrDirLocal(c, filePath)
}

// deleteFileOrDirContainer deletes a file/directory from user's container workspace
func (h *Handlers) deleteFileOrDirContainer(c *gin.Context, filePath string) {
	// Get user from context
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := user.ID

	// Get or create filesystem container
	containerID, err := h.containerMgr.GetOrCreateFilesystemContainer(userIDStr)
	if err != nil {
		log.Printf("[Filesystem] Failed to get container for user %s: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access filesystem"})
		return
	}

	// Normalize path
	if strings.HasPrefix(filePath, "~/") {
		filePath = filepath.Join(containerWorkspaceRoot, filePath[2:])
	}
	if !strings.HasPrefix(filePath, containerWorkspaceRoot) {
		filePath = filepath.Join(containerWorkspaceRoot, filePath)
	}

	// Safety: prevent deleting workspace root
	if filePath == containerWorkspaceRoot || filePath == containerWorkspaceRoot+"/" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete workspace root"})
		return
	}

	// Delete in container
	if err := h.containerMgr.DeletePathInContainer(containerID, filePath); err != nil {
		log.Printf("[Filesystem] Failed to delete %s: %v", filePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully", "path": filePath})
}

// deleteFileOrDirLocal deletes a file/directory from local filesystem (fallback)
func (h *Handlers) deleteFileOrDirLocal(c *gin.Context, filePath string) {
	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Safety: prevent deleting system directories
	homeDir, _ := os.UserHomeDir()
	dangerousPaths := []string{"/", "/bin", "/usr", "/etc", "/System", homeDir}
	for _, dangerous := range dangerousPaths {
		if filePath == dangerous {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete this path"})
			return
		}
	}

	// Check if path exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Path not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access path"})
		return
	}

	// For directories, only allow deleting if empty (use recursive flag for non-empty)
	if info.IsDir() {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
			return
		}
		if len(entries) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Directory is not empty. Use recursive delete."})
			return
		}
	}

	// Delete the file/directory
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully", "path": filePath})
}

// UploadFile handles file uploads
func (h *Handlers) UploadFile(c *gin.Context) {
	destPath := c.PostForm("path")
	if destPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Destination path is required"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}

	// Check if container isolation is enabled
	if h.containerMgr != nil {
		h.uploadFileContainer(c, destPath, header.Filename, content)
		return
	}

	// Fallback to local filesystem
	h.uploadFileLocal(c, destPath, header.Filename, content)
}

// uploadFileContainer uploads a file to user's container workspace
func (h *Handlers) uploadFileContainer(c *gin.Context, destPath, filename string, content []byte) {
	// Get user from context
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := user.ID

	// Get or create filesystem container
	containerID, err := h.containerMgr.GetOrCreateFilesystemContainer(userIDStr)
	if err != nil {
		log.Printf("[Filesystem] Failed to get container for user %s: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access filesystem"})
		return
	}

	// Normalize path
	if strings.HasPrefix(destPath, "~/") {
		destPath = filepath.Join(containerWorkspaceRoot, destPath[2:])
	}
	if !strings.HasPrefix(destPath, containerWorkspaceRoot) {
		destPath = filepath.Join(containerWorkspaceRoot, destPath)
	}

	fullPath := filepath.Join(destPath, filename)

	// Write file to container
	if err := h.containerMgr.WriteFileToContainer(containerID, fullPath, content, 0644); err != nil {
		log.Printf("[Filesystem] Failed to upload file %s: %v", fullPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	now := time.Now()
	c.JSON(http.StatusCreated, FileItem{
		ID:        generateFileID(fullPath),
		Name:      filename,
		Type:      "file",
		Path:      fullPath,
		Size:      int64(len(content)),
		Modified:  &now,
		Extension: strings.TrimPrefix(filepath.Ext(filename), "."),
		IsHidden:  strings.HasPrefix(filename, "."),
	})
}

// uploadFileLocal uploads a file to local filesystem (fallback)
func (h *Handlers) uploadFileLocal(c *gin.Context, destPath, filename string, content []byte) {
	// Expand ~ to home directory
	if strings.HasPrefix(destPath, "~/") {
		homeDir, _ := os.UserHomeDir()
		destPath = filepath.Join(homeDir, destPath[2:])
	}

	// Create full file path
	fullPath := filepath.Join(destPath, filename)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer dst.Close()

	// Write content
	if _, err := dst.Write(content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Get file info
	info, _ := os.Stat(fullPath)
	modTime := info.ModTime()

	c.JSON(http.StatusCreated, FileItem{
		ID:        generateFileID(fullPath),
		Name:      filename,
		Type:      "file",
		Path:      fullPath,
		Size:      info.Size(),
		Modified:  &modTime,
		Extension: strings.TrimPrefix(filepath.Ext(filename), "."),
		IsHidden:  strings.HasPrefix(filename, "."),
	})
}

// GetQuickAccessPaths returns commonly used paths
func (h *Handlers) GetQuickAccessPaths(c *gin.Context) {
	// Check if container isolation is enabled
	if h.containerMgr != nil {
		// Return container workspace paths
		paths := []struct {
			Name string `json:"name"`
			Path string `json:"path"`
			Icon string `json:"icon"`
		}{
			{Name: "Workspace", Path: containerWorkspaceRoot, Icon: "home"},
			{Name: "Documents", Path: filepath.Join(containerWorkspaceRoot, "documents"), Icon: "document"},
			{Name: "Projects", Path: filepath.Join(containerWorkspaceRoot, "projects"), Icon: "folder"},
			{Name: "Downloads", Path: filepath.Join(containerWorkspaceRoot, "downloads"), Icon: "download"},
		}
		c.JSON(http.StatusOK, gin.H{"paths": paths})
		return
	}

	// Fallback to local filesystem paths
	homeDir, _ := os.UserHomeDir()

	paths := []struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Icon string `json:"icon"`
	}{
		{Name: "Home", Path: homeDir, Icon: "home"},
		{Name: "Desktop", Path: filepath.Join(homeDir, "Desktop"), Icon: "desktop"},
		{Name: "Documents", Path: filepath.Join(homeDir, "Documents"), Icon: "document"},
		{Name: "Downloads", Path: filepath.Join(homeDir, "Downloads"), Icon: "download"},
		{Name: "Pictures", Path: filepath.Join(homeDir, "Pictures"), Icon: "image"},
		{Name: "Music", Path: filepath.Join(homeDir, "Music"), Icon: "music"},
		{Name: "Videos", Path: filepath.Join(homeDir, "Movies"), Icon: "video"},
	}

	// Filter to only paths that exist
	validPaths := make([]struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Icon string `json:"icon"`
	}, 0)

	for _, p := range paths {
		if _, err := os.Stat(p.Path); err == nil {
			validPaths = append(validPaths, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{"paths": validPaths})
}

// generateFileID creates a unique ID for a file path
func generateFileID(path string) string {
	// Use base64 of path for a stable ID
	return fmt.Sprintf("file_%x", path)
}
