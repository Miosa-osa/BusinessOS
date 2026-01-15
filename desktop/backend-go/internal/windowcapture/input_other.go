//go:build !darwin
// +build !darwin

package windowcapture

// InputEvent represents an input event from the frontend
type InputEvent struct {
	Type      string `json:"type"`      // "mousemove", "mousedown", "mouseup", "click", "dblclick", "scroll", "keydown", "keyup", "char"
	X         int    `json:"x"`         // Mouse X position (relative to window)
	Y         int    `json:"y"`         // Mouse Y position (relative to window)
	Button    int    `json:"button"`    // 0=left, 1=right, 2=middle
	DeltaX    int    `json:"deltaX"`    // Scroll delta X
	DeltaY    int    `json:"deltaY"`    // Scroll delta Y
	KeyCode   int    `json:"keyCode"`   // Key code for keyboard events
	Char      string `json:"char"`      // Character for text input
	Modifiers int    `json:"modifiers"` // Bit flags: 1=shift, 2=ctrl, 4=alt, 8=cmd
}

// WindowBounds stores window position for coordinate translation
type WindowBounds struct {
	X      int
	Y      int
	Width  int
	Height int
}

// InjectMouseMove is a no-op on non-macOS platforms
func InjectMouseMove(windowBounds WindowBounds, relX, relY int) {}

// InjectMouseClick is a no-op on non-macOS platforms
func InjectMouseClick(windowBounds WindowBounds, relX, relY int, button int, isDown bool) {}

// InjectDoubleClick is a no-op on non-macOS platforms
func InjectDoubleClick(windowBounds WindowBounds, relX, relY int) {}

// InjectScroll is a no-op on non-macOS platforms
func InjectScroll(windowBounds WindowBounds, relX, relY int, deltaX, deltaY int) {}

// InjectKeyEvent is a no-op on non-macOS platforms
func InjectKeyEvent(keyCode int, isDown bool, modifiers int) {}

// InjectCharacter is a no-op on non-macOS platforms
func InjectCharacter(char rune) {}

// BringWindowToFront is a no-op on non-macOS platforms
func BringWindowToFront(windowID int) {}
