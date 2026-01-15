//go:build darwin
// +build darwin

package windowcapture

/*
#cgo CFLAGS: -x objective-c -Wno-deprecated-declarations -mmacosx-version-min=14.0
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework AppKit -framework ImageIO -framework ScreenCaptureKit

#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>
#include <AppKit/AppKit.h>
#include <ImageIO/ImageIO.h>
#include <stdlib.h>
#include <string.h>

// Window info structure
typedef struct {
    int windowID;
    int ownerPID;
    char* ownerName;
    char* windowName;
    int x, y, width, height;
    int layer;
    int isOnscreen;
} WindowInfo;

// Get list of windows for a specific bundle ID
// Uses kCGWindowListOptionAll to find windows even if they're off-screen
WindowInfo* getWindowsForBundleID(const char* bundleID, int* count) {
    // Get list of ALL windows (including off-screen) - needed for off-screen capture approach
    CFArrayRef windowList = CGWindowListCopyWindowInfo(
        kCGWindowListOptionAll | kCGWindowListExcludeDesktopElements,
        kCGNullWindowID
    );

    if (!windowList) {
        *count = 0;
        return NULL;
    }

    CFIndex windowCount = CFArrayGetCount(windowList);
    WindowInfo* windows = (WindowInfo*)malloc(sizeof(WindowInfo) * windowCount);
    int matchCount = 0;

    for (CFIndex i = 0; i < windowCount; i++) {
        CFDictionaryRef window = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, i);

        // Get owner name
        CFStringRef ownerNameRef = (CFStringRef)CFDictionaryGetValue(window, kCGWindowOwnerName);
        if (!ownerNameRef) continue;

        // Get owner PID
        CFNumberRef ownerPIDRef = (CFNumberRef)CFDictionaryGetValue(window, kCGWindowOwnerPID);
        if (!ownerPIDRef) continue;

        int ownerPID;
        CFNumberGetValue(ownerPIDRef, kCFNumberIntType, &ownerPID);

        // Get the running application by PID to check bundle ID
        NSRunningApplication* app = [NSRunningApplication runningApplicationWithProcessIdentifier:ownerPID];
        if (!app || !app.bundleIdentifier) continue;

        const char* appBundleID = [app.bundleIdentifier UTF8String];
        if (strcmp(appBundleID, bundleID) != 0) continue;

        // Get window ID
        CFNumberRef windowIDRef = (CFNumberRef)CFDictionaryGetValue(window, kCGWindowNumber);
        if (!windowIDRef) continue;

        int windowID;
        CFNumberGetValue(windowIDRef, kCFNumberIntType, &windowID);

        // Get window bounds
        CFDictionaryRef boundsRef = (CFDictionaryRef)CFDictionaryGetValue(window, kCGWindowBounds);
        if (!boundsRef) continue;

        CGRect bounds;
        CGRectMakeWithDictionaryRepresentation(boundsRef, &bounds);

        // Get window layer
        CFNumberRef layerRef = (CFNumberRef)CFDictionaryGetValue(window, kCGWindowLayer);
        int layer = 0;
        if (layerRef) {
            CFNumberGetValue(layerRef, kCFNumberIntType, &layer);
        }

        // Get window name
        CFStringRef windowNameRef = (CFStringRef)CFDictionaryGetValue(window, kCGWindowName);
        const char* windowName = "";
        if (windowNameRef) {
            windowName = [(NSString*)windowNameRef UTF8String];
        }

        // Copy owner name
        char* ownerNameCopy = strdup([(NSString*)ownerNameRef UTF8String]);
        char* windowNameCopy = strdup(windowName);

        // Fill window info
        windows[matchCount].windowID = windowID;
        windows[matchCount].ownerPID = ownerPID;
        windows[matchCount].ownerName = ownerNameCopy;
        windows[matchCount].windowName = windowNameCopy;
        windows[matchCount].x = (int)bounds.origin.x;
        windows[matchCount].y = (int)bounds.origin.y;
        windows[matchCount].width = (int)bounds.size.width;
        windows[matchCount].height = (int)bounds.size.height;
        windows[matchCount].layer = layer;
        windows[matchCount].isOnscreen = 1;

        matchCount++;
    }

    CFRelease(windowList);
    *count = matchCount;

    if (matchCount == 0) {
        free(windows);
        return NULL;
    }

    return windows;
}

// Free window info array
void freeWindowInfoArray(WindowInfo* windows, int count) {
    if (!windows) return;
    for (int i = 0; i < count; i++) {
        if (windows[i].ownerName) free(windows[i].ownerName);
        if (windows[i].windowName) free(windows[i].windowName);
    }
    free(windows);
}

// Capture a specific window as JPEG
unsigned char* captureWindowAsJPEG(int windowID, int* dataSize, float quality) {
    // Create image from window
    CGImageRef image = CGWindowListCreateImage(
        CGRectNull,  // Capture the window's bounds
        kCGWindowListOptionIncludingWindow,
        (CGWindowID)windowID,
        kCGWindowImageBoundsIgnoreFraming | kCGWindowImageNominalResolution
    );

    if (!image) {
        *dataSize = 0;
        return NULL;
    }

    // Create mutable data to hold JPEG
    CFMutableDataRef jpegData = CFDataCreateMutable(NULL, 0);

    // Create image destination
    CGImageDestinationRef dest = CGImageDestinationCreateWithData(
        jpegData,
        kUTTypeJPEG,
        1,
        NULL
    );

    if (!dest) {
        CGImageRelease(image);
        CFRelease(jpegData);
        *dataSize = 0;
        return NULL;
    }

    // Set compression quality
    CFStringRef keys[] = { kCGImageDestinationLossyCompressionQuality };
    CFNumberRef values[] = { CFNumberCreate(NULL, kCFNumberFloatType, &quality) };
    CFDictionaryRef options = CFDictionaryCreate(
        NULL,
        (const void**)keys,
        (const void**)values,
        1,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks
    );

    CGImageDestinationAddImage(dest, image, options);
    CGImageDestinationFinalize(dest);

    // Copy data
    CFIndex length = CFDataGetLength(jpegData);
    unsigned char* result = (unsigned char*)malloc(length);
    memcpy(result, CFDataGetBytePtr(jpegData), length);
    *dataSize = (int)length;

    // Cleanup
    CFRelease(values[0]);
    CFRelease(options);
    CFRelease(dest);
    CFRelease(jpegData);
    CGImageRelease(image);

    return result;
}

// Free captured image data
void freeCapturedData(unsigned char* data) {
    if (data) free(data);
}

// Check if screen capture permission is granted
int hasScreenCapturePermission() {
    if (@available(macOS 10.15, *)) {
        // Try to capture a small area - this will trigger permission request if not granted
        CGImageRef testImage = CGWindowListCreateImage(
            CGRectMake(0, 0, 1, 1),
            kCGWindowListOptionOnScreenOnly,
            kCGNullWindowID,
            kCGWindowImageDefault
        );

        if (testImage) {
            size_t width = CGImageGetWidth(testImage);
            CGImageRelease(testImage);
            return width > 0 ? 1 : 0;
        }
        return 0;
    }
    return 1; // Pre-Catalina, no permission needed
}

// Request screen capture permission (triggers system dialog)
void requestScreenCapturePermission() {
    if (@available(macOS 10.15, *)) {
        // Trigger permission dialog by attempting capture
        CGImageRef testImage = CGWindowListCreateImage(
            CGRectMake(0, 0, 1, 1),
            kCGWindowListOptionOnScreenOnly,
            kCGNullWindowID,
            kCGWindowImageDefault
        );
        if (testImage) {
            CGImageRelease(testImage);
        }
    }
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

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

// GetWindowsForBundleID returns all windows for a specific app bundle ID
func GetWindowsForBundleID(bundleID string) ([]WindowInfo, error) {
	cBundleID := C.CString(bundleID)
	defer C.free(unsafe.Pointer(cBundleID))

	var count C.int
	cWindows := C.getWindowsForBundleID(cBundleID, &count)

	if cWindows == nil || count == 0 {
		return nil, fmt.Errorf("no windows found for bundle ID: %s", bundleID)
	}
	defer C.freeWindowInfoArray(cWindows, count)

	// Convert C array to Go slice
	windows := make([]WindowInfo, int(count))
	cWindowsSlice := (*[1 << 20]C.WindowInfo)(unsafe.Pointer(cWindows))[:count:count]

	for i, cw := range cWindowsSlice {
		windows[i] = WindowInfo{
			WindowID:   int(cw.windowID),
			OwnerPID:   int(cw.ownerPID),
			OwnerName:  C.GoString(cw.ownerName),
			WindowName: C.GoString(cw.windowName),
			X:          int(cw.x),
			Y:          int(cw.y),
			Width:      int(cw.width),
			Height:     int(cw.height),
			Layer:      int(cw.layer),
			IsOnscreen: cw.isOnscreen != 0,
		}
	}

	return windows, nil
}

// CaptureWindowAsJPEG captures a window and returns JPEG data
func CaptureWindowAsJPEG(windowID int, quality float32) ([]byte, error) {
	var dataSize C.int
	cData := C.captureWindowAsJPEG(C.int(windowID), &dataSize, C.float(quality))

	if cData == nil || dataSize == 0 {
		return nil, fmt.Errorf("failed to capture window %d", windowID)
	}
	defer C.freeCapturedData(cData)

	// Copy data to Go slice
	data := C.GoBytes(unsafe.Pointer(cData), dataSize)
	return data, nil
}

// HasScreenCapturePermission checks if screen capture permission is granted
func HasScreenCapturePermission() bool {
	return C.hasScreenCapturePermission() != 0
}

// RequestScreenCapturePermission triggers the macOS permission dialog
func RequestScreenCapturePermission() {
	C.requestScreenCapturePermission()
}
