//go:build darwin
// +build darwin

package windowcapture

/*
#cgo CFLAGS: -x objective-c -Wno-deprecated-declarations -mmacosx-version-min=14.0
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework AppKit

#include <CoreGraphics/CoreGraphics.h>
#include <AppKit/AppKit.h>

// Move mouse to position
void moveMouse(int x, int y) {
    CGEventRef move = CGEventCreateMouseEvent(
        NULL,
        kCGEventMouseMoved,
        CGPointMake(x, y),
        kCGMouseButtonLeft
    );
    CGEventPost(kCGHIDEventTap, move);
    CFRelease(move);
}

// Click at position
void clickMouse(int x, int y, int button, int isDown) {
    CGEventType eventType;
    CGMouseButton mouseButton;

    if (button == 0) { // Left button
        mouseButton = kCGMouseButtonLeft;
        eventType = isDown ? kCGEventLeftMouseDown : kCGEventLeftMouseUp;
    } else if (button == 1) { // Right button
        mouseButton = kCGMouseButtonRight;
        eventType = isDown ? kCGEventRightMouseDown : kCGEventRightMouseUp;
    } else { // Middle button
        mouseButton = kCGMouseButtonCenter;
        eventType = isDown ? kCGEventOtherMouseDown : kCGEventOtherMouseUp;
    }

    CGEventRef click = CGEventCreateMouseEvent(
        NULL,
        eventType,
        CGPointMake(x, y),
        mouseButton
    );
    CGEventPost(kCGHIDEventTap, click);
    CFRelease(click);
}

// Double click at position
void doubleClickMouse(int x, int y) {
    CGEventRef click = CGEventCreateMouseEvent(
        NULL,
        kCGEventLeftMouseDown,
        CGPointMake(x, y),
        kCGMouseButtonLeft
    );
    CGEventSetIntegerValueField(click, kCGMouseEventClickState, 2);
    CGEventPost(kCGHIDEventTap, click);

    CGEventSetType(click, kCGEventLeftMouseUp);
    CGEventPost(kCGHIDEventTap, click);
    CFRelease(click);
}

// Scroll at position
void scrollMouse(int x, int y, int deltaX, int deltaY) {
    // First move to position
    CGEventRef move = CGEventCreateMouseEvent(
        NULL,
        kCGEventMouseMoved,
        CGPointMake(x, y),
        kCGMouseButtonLeft
    );
    CGEventPost(kCGHIDEventTap, move);
    CFRelease(move);

    // Then scroll
    CGEventRef scroll = CGEventCreateScrollWheelEvent(
        NULL,
        kCGScrollEventUnitPixel,
        2,
        deltaY,
        deltaX
    );
    CGEventPost(kCGHIDEventTap, scroll);
    CFRelease(scroll);
}

// Send key event
void sendKey(int keyCode, int isDown, int modifiers) {
    CGEventRef key = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)keyCode, isDown);

    // Apply modifiers
    CGEventFlags flags = 0;
    if (modifiers & 1) flags |= kCGEventFlagMaskShift;
    if (modifiers & 2) flags |= kCGEventFlagMaskControl;
    if (modifiers & 4) flags |= kCGEventFlagMaskAlternate;
    if (modifiers & 8) flags |= kCGEventFlagMaskCommand;

    CGEventSetFlags(key, flags);
    CGEventPost(kCGHIDEventTap, key);
    CFRelease(key);
}

// Type a character (for text input)
void typeCharacter(UniChar character) {
    CGEventRef keyDown = CGEventCreateKeyboardEvent(NULL, 0, true);
    CGEventRef keyUp = CGEventCreateKeyboardEvent(NULL, 0, false);

    CGEventKeyboardSetUnicodeString(keyDown, 1, &character);
    CGEventKeyboardSetUnicodeString(keyUp, 1, &character);

    CGEventPost(kCGHIDEventTap, keyDown);
    CGEventPost(kCGHIDEventTap, keyUp);

    CFRelease(keyDown);
    CFRelease(keyUp);
}

// Move window to specific position (used for off-screen capture)
void moveWindowToPosition(int windowID, int x, int y) {
    // Get window owner PID
    CFArrayRef windowList = CGWindowListCopyWindowInfo(
        kCGWindowListOptionAll,
        (CGWindowID)windowID
    );

    if (!windowList || CFArrayGetCount(windowList) == 0) {
        if (windowList) CFRelease(windowList);
        return;
    }

    CFDictionaryRef window = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, 0);
    CFNumberRef ownerPIDRef = (CFNumberRef)CFDictionaryGetValue(window, kCGWindowOwnerPID);

    if (ownerPIDRef) {
        int ownerPID;
        CFNumberGetValue(ownerPIDRef, kCFNumberIntType, &ownerPID);

        // Use AppleScript to move the window (most reliable cross-app method)
        NSString* script = [NSString stringWithFormat:
            @"tell application \"System Events\"\n"
            @"  set targetProcess to first process whose unix id is %d\n"
            @"  tell targetProcess\n"
            @"    set position of window 1 to {%d, %d}\n"
            @"  end tell\n"
            @"end tell", ownerPID, x, y];

        NSAppleScript* appleScript = [[NSAppleScript alloc] initWithSource:script];
        NSDictionary* errorDict = nil;
        [appleScript executeAndReturnError:&errorDict];
    }

    CFRelease(windowList);
}

// Get window's current position
void getWindowPosition(int windowID, int* outX, int* outY, int* outWidth, int* outHeight) {
    CFArrayRef windowList = CGWindowListCopyWindowInfo(
        kCGWindowListOptionAll,
        kCGNullWindowID
    );

    if (!windowList) {
        *outX = 0; *outY = 0; *outWidth = 0; *outHeight = 0;
        return;
    }

    CFIndex count = CFArrayGetCount(windowList);
    for (CFIndex i = 0; i < count; i++) {
        CFDictionaryRef window = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, i);
        CFNumberRef windowIDRef = (CFNumberRef)CFDictionaryGetValue(window, kCGWindowNumber);

        if (windowIDRef) {
            int wID;
            CFNumberGetValue(windowIDRef, kCFNumberIntType, &wID);

            if (wID == windowID) {
                CFDictionaryRef boundsRef = (CFDictionaryRef)CFDictionaryGetValue(window, kCGWindowBounds);
                if (boundsRef) {
                    CGRect bounds;
                    CGRectMakeWithDictionaryRepresentation(boundsRef, &bounds);
                    *outX = (int)bounds.origin.x;
                    *outY = (int)bounds.origin.y;
                    *outWidth = (int)bounds.size.width;
                    *outHeight = (int)bounds.size.height;
                }
                break;
            }
        }
    }

    CFRelease(windowList);
}

// Bring window to front
void bringWindowToFront(int windowID) {
    // Get window owner PID
    CFArrayRef windowList = CGWindowListCopyWindowInfo(
        kCGWindowListOptionIncludingWindow,
        (CGWindowID)windowID
    );

    if (!windowList || CFArrayGetCount(windowList) == 0) {
        if (windowList) CFRelease(windowList);
        return;
    }

    CFDictionaryRef window = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, 0);
    CFNumberRef ownerPIDRef = (CFNumberRef)CFDictionaryGetValue(window, kCGWindowOwnerPID);

    if (ownerPIDRef) {
        int ownerPID;
        CFNumberGetValue(ownerPIDRef, kCFNumberIntType, &ownerPID);

        NSRunningApplication* app = [NSRunningApplication runningApplicationWithProcessIdentifier:ownerPID];
        if (app) {
            [app activateWithOptions:NSApplicationActivateIgnoringOtherApps];
        }
    }

    CFRelease(windowList);
}
*/
import "C"

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

// InjectMouseMove moves the mouse to the specified position
func InjectMouseMove(windowBounds WindowBounds, relX, relY int) {
	// Translate relative coordinates to screen coordinates
	screenX := windowBounds.X + relX
	screenY := windowBounds.Y + relY
	C.moveMouse(C.int(screenX), C.int(screenY))
}

// InjectMouseClick sends a mouse click at the specified position
func InjectMouseClick(windowBounds WindowBounds, relX, relY int, button int, isDown bool) {
	screenX := windowBounds.X + relX
	screenY := windowBounds.Y + relY
	down := 0
	if isDown {
		down = 1
	}
	C.clickMouse(C.int(screenX), C.int(screenY), C.int(button), C.int(down))
}

// InjectDoubleClick sends a double click at the specified position
func InjectDoubleClick(windowBounds WindowBounds, relX, relY int) {
	screenX := windowBounds.X + relX
	screenY := windowBounds.Y + relY
	C.doubleClickMouse(C.int(screenX), C.int(screenY))
}

// InjectScroll sends a scroll event at the specified position
func InjectScroll(windowBounds WindowBounds, relX, relY int, deltaX, deltaY int) {
	screenX := windowBounds.X + relX
	screenY := windowBounds.Y + relY
	C.scrollMouse(C.int(screenX), C.int(screenY), C.int(deltaX), C.int(deltaY))
}

// InjectKeyEvent sends a key event
func InjectKeyEvent(keyCode int, isDown bool, modifiers int) {
	down := 0
	if isDown {
		down = 1
	}
	C.sendKey(C.int(keyCode), C.int(down), C.int(modifiers))
}

// InjectCharacter types a single character
func InjectCharacter(char rune) {
	C.typeCharacter(C.UniChar(char))
}

// BringWindowToFront activates the window's application
func BringWindowToFront(windowID int) {
	C.bringWindowToFront(C.int(windowID))
}
