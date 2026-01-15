package system

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rhl/businessos-backend/internal/logging"
)

// NativeApp represents a macOS application
type NativeApp struct {
	Name        string `json:"name"`
	BundleID    string `json:"bundle_id"`
	Path        string `json:"path"`
	IconPath    string `json:"icon_path"`
	Version     string `json:"version"`
	IsInstalled bool   `json:"is_installed"`
	IsRunning   bool   `json:"is_running"`
}

// InfoPlist represents the structure of an app's Info.plist
type InfoPlist struct {
	CFBundleName         string `json:"CFBundleName"`
	CFBundleDisplayName  string `json:"CFBundleDisplayName"`
	CFBundleIdentifier   string `json:"CFBundleIdentifier"`
	CFBundleVersion      string `json:"CFBundleVersion"`
	CFBundleIconFile     string `json:"CFBundleIconFile"`
	CFBundleShortVersion string `json:"CFBundleShortVersionString"`
}

// ListInstalledApps scans common macOS application directories
func ListInstalledApps() ([]NativeApp, error) {
	apps := []NativeApp{}
	seen := make(map[string]bool) // Deduplicate by bundle ID

	// Search standard macOS app locations
	searchPaths := []string{
		"/Applications",
		"/System/Applications",
		filepath.Join(os.Getenv("HOME"), "Applications"),
	}

	for _, searchPath := range searchPaths {
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue // Skip if directory doesn't exist
		}

		entries, err := os.ReadDir(searchPath)
		if err != nil {
			logging.Warn("[macOS] Failed to read directory %s: %v", searchPath, err)
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
				continue
			}

			appPath := filepath.Join(searchPath, entry.Name())
			app, err := GetAppMetadata(appPath)
			if err != nil {
				logging.Warn("[macOS] Failed to get metadata for %s: %v", appPath, err)
				continue
			}

			// Deduplicate by bundle ID
			if seen[app.BundleID] {
				continue
			}
			seen[app.BundleID] = true

			apps = append(apps, *app)
		}
	}

	logging.Info("[macOS] Found %d installed applications", len(apps))
	return apps, nil
}

// GetAppMetadata reads an app's Info.plist and extracts metadata
func GetAppMetadata(appPath string) (*NativeApp, error) {
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")

	// Check if Info.plist exists
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Info.plist not found at %s", plistPath)
	}

	// Convert plist to JSON using plutil
	cmd := exec.Command("plutil", "-convert", "json", "-o", "-", plistPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to convert plist: %w", err)
	}

	var plist InfoPlist
	if err := json.Unmarshal(output, &plist); err != nil {
		return nil, fmt.Errorf("failed to parse plist JSON: %w", err)
	}

	// Use display name if available, fallback to bundle name
	displayName := plist.CFBundleDisplayName
	if displayName == "" {
		displayName = plist.CFBundleName
	}
	if displayName == "" {
		// Fallback to app folder name without .app extension
		displayName = strings.TrimSuffix(filepath.Base(appPath), ".app")
	}

	// Find icon file
	iconPath := ""
	if plist.CFBundleIconFile != "" {
		iconFile := plist.CFBundleIconFile
		if !strings.HasSuffix(iconFile, ".icns") {
			iconFile += ".icns"
		}
		iconPath = filepath.Join(appPath, "Contents", "Resources", iconFile)

		// Verify icon exists
		if _, err := os.Stat(iconPath); os.IsNotExist(err) {
			iconPath = "" // Icon file doesn't exist
		}
	}

	// Version string
	version := plist.CFBundleShortVersion
	if version == "" {
		version = plist.CFBundleVersion
	}

	return &NativeApp{
		Name:        displayName,
		BundleID:    plist.CFBundleIdentifier,
		Path:        appPath,
		IconPath:    iconPath,
		Version:     version,
		IsInstalled: true,
		IsRunning:   false, // Will be populated by GetRunningApps
	}, nil
}

// GetRunningApps returns list of currently running applications
func GetRunningApps() (map[string]bool, error) {
	// Use AppleScript to get running app bundle IDs
	script := `
		tell application "System Events"
			set appList to ""
			repeat with proc in application processes
				try
					set bundleID to bundle identifier of proc
					set appList to appList & bundleID & "\n"
				end try
			end repeat
			return appList
		end tell
	`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get running apps: %w", err)
	}

	running := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		bundleID := strings.TrimSpace(line)
		if bundleID != "" {
			running[bundleID] = true
		}
	}

	logging.Info("[macOS] Found %d running applications", len(running))
	return running, nil
}

// FindPopularApps returns commonly used apps with enhanced metadata
func FindPopularApps() []NativeApp {
	popularBundleIDs := map[string]string{
		"notion.id":                      "Notion",
		"com.tinyspeck.slackmacgap":      "Slack",
		"com.linear":                     "Linear",
		"com.anthropic.claudefordesktop": "Claude",
		"com.hnc.Discord":                "Discord",
		"com.todesktop.230313mzl4w4u92":  "Cursor",
		"com.microsoft.VSCode":           "Visual Studio Code",
		"com.figma.Desktop":              "Figma",
		"com.github.GitHubClient":        "GitHub Desktop",
		"com.clickup.desktop-app":        "ClickUp",
		"com.airtable.airtable-desktop":  "Airtable",
		"com.canva.CanvaDesktop":         "Canva",
		"com.loom.desktop":               "Loom",
		"com.realtimeboard.miro":         "Miro",
		"com.asana.macOS":                "Asana",
	}

	allApps, err := ListInstalledApps()
	if err != nil {
		logging.Error("[macOS] Failed to list apps: %v", err)
		return []NativeApp{}
	}

	// Get running apps
	runningApps, err := GetRunningApps()
	if err != nil {
		logging.Warn("[macOS] Failed to get running apps: %v", err)
		runningApps = make(map[string]bool)
	}

	// Filter to popular apps only
	popular := []NativeApp{}
	for _, app := range allApps {
		if expectedName, isPopular := popularBundleIDs[app.BundleID]; isPopular {
			app.IsRunning = runningApps[app.BundleID]

			// Override name with expected name if needed
			if app.Name == "" {
				app.Name = expectedName
			}

			popular = append(popular, app)
		}
	}

	logging.Info("[macOS] Found %d popular apps installed", len(popular))
	return popular
}

// LaunchApp opens a macOS application by bundle ID
func LaunchApp(bundleID string) error {
	cmd := exec.Command("open", "-b", bundleID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch app %s: %w", bundleID, err)
	}

	logging.Info("[macOS] Launched app: %s", bundleID)
	return nil
}

// GetAppIcon extracts the icon from a .app bundle and converts to PNG
// Returns base64-encoded PNG data
func GetAppIcon(appPath string, size int) (string, error) {
	app, err := GetAppMetadata(appPath)
	if err != nil {
		return "", err
	}

	if app.IconPath == "" {
		return "", fmt.Errorf("no icon found for app")
	}

	// For now, just return the icon path - we'll serve it via HTTP
	// TODO: Convert .icns to PNG if needed for web display
	return app.IconPath, nil
}
