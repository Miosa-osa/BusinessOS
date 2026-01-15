//go:build !darwin
// +build !darwin

package windowcapture

import "errors"

// WindowInfo represents information about a captured window
type WindowInfo struct {
	WindowID   int    `json:"window_id"`
	OwnerPID   int    `json:"owner_pid"`
	OwnerName  string `json:"owner_name"`
	WindowName string `json:"window_name"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Layer      int    `json:"layer"`
	IsOnscreen bool   `json:"is_onscreen"`
}

var errNotSupported = errors.New("window capture is only supported on macOS")

// GetWindowsForBundleID returns all windows for a specific app bundle ID
func GetWindowsForBundleID(bundleID string) ([]WindowInfo, error) {
	return nil, errNotSupported
}

// CaptureWindowAsJPEG captures a window and returns JPEG data
func CaptureWindowAsJPEG(windowID int, quality float32) ([]byte, error) {
	return nil, errNotSupported
}

// HasScreenCapturePermission checks if screen capture permission is granted
func HasScreenCapturePermission() bool {
	return false
}

// RequestScreenCapturePermission triggers the macOS permission dialog
func RequestScreenCapturePermission() {
	// No-op on non-macOS platforms
}
