package handlers

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/logging"
	"github.com/rhl/businessos-backend/internal/system"
)

// SystemAppsHandler handles macOS system app detection
type SystemAppsHandler struct{}

// NewSystemAppsHandler creates a new system apps handler
func NewSystemAppsHandler() *SystemAppsHandler {
	return &SystemAppsHandler{}
}

// ListInstalledApps godoc
// @Summary List all installed macOS applications
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/system/installed-apps [get]
func (h *SystemAppsHandler) ListInstalledApps(c *gin.Context) {
	apps, err := system.ListInstalledApps()
	if err != nil {
		logging.Error("[SystemApps] Failed to list apps: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list installed apps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apps":  apps,
		"count": len(apps),
	})
}

// ListPopularApps godoc
// @Summary List popular/common installed apps (Notion, Slack, etc.)
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/system/popular-apps [get]
func (h *SystemAppsHandler) ListPopularApps(c *gin.Context) {
	apps := system.FindPopularApps()

	c.JSON(http.StatusOK, gin.H{
		"apps":  apps,
		"count": len(apps),
	})
}

// LaunchApp godoc
// @Summary Launch a macOS application by bundle ID
// @Tags System
// @Produce json
// @Param bundle_id query string true "Bundle ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/system/launch-app [post]
func (h *SystemAppsHandler) LaunchApp(c *gin.Context) {
	bundleID := c.Query("bundle_id")
	if bundleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bundle_id is required"})
		return
	}

	if err := system.LaunchApp(bundleID); err != nil {
		logging.Error("[SystemApps] Failed to launch app: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to launch app"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "App launched successfully",
	})
}

// GetAppIcon godoc
// @Summary Get application icon file
// @Tags System
// @Produce image/x-icns
// @Param path query string true "Icon file path"
// @Success 200 {file} binary
// @Router /api/system/app-icon [get]
func (h *SystemAppsHandler) GetAppIcon(c *gin.Context) {
	iconPath := c.Query("path")
	if iconPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	// Security: Validate path is within /Applications or /System/Applications
	cleanPath := filepath.Clean(iconPath)
	if !strings.HasPrefix(cleanPath, "/Applications/") && !strings.HasPrefix(cleanPath, "/System/Applications/") {
		logging.Warn("[SystemApps] Rejected icon path outside Applications: %s", iconPath)
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid icon path"})
		return
	}

	// Validate file exists and has .icns extension
	if !strings.HasSuffix(cleanPath, ".icns") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type, must be .icns"})
		return
	}

	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		logging.Warn("[SystemApps] Icon file not found: %s", cleanPath)
		c.JSON(http.StatusNotFound, gin.H{"error": "Icon file not found"})
		return
	}

	// Convert .icns to PNG using sips (macOS built-in tool)
	// sips can't write to stdout, so use a temp file
	tmpFile, err := ioutil.TempFile("", "icon-*.png")
	if err != nil {
		logging.Error("[SystemApps] Failed to create temp file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert icon"})
		return
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath) // Clean up temp file

	// Convert to PNG
	cmd := exec.Command("sips", "-s", "format", "png", "-Z", "128", cleanPath, "--out", tmpPath)
	if err := cmd.Run(); err != nil {
		logging.Error("[SystemApps] Failed to convert icon to PNG: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert icon"})
		return
	}

	// Read PNG data
	pngData, err := os.ReadFile(tmpPath)
	if err != nil {
		logging.Error("[SystemApps] Failed to read PNG: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read converted icon"})
		return
	}

	// Serve as PNG
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 1 day
	c.Data(http.StatusOK, "image/png", pngData)
}
