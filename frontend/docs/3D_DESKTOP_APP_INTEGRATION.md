# 3D Desktop App Integration

## Overview

This document covers the findings and implementation details for integrating external applications into the BusinessOS 3D Desktop environment.

**Last Updated:** January 2026
**Status:** Web Apps - Production Ready | Native Apps - Experimental (Deprioritized)

---

## Table of Contents

1. [Architecture Summary](#architecture-summary)
2. [Web App Integration (Recommended)](#web-app-integration-recommended)
3. [Native App Capture (Experimental)](#native-app-capture-experimental)
4. [Technical Limitations](#technical-limitations)
5. [Implementation Guide](#implementation-guide)
6. [Future Possibilities](#future-possibilities)

---

## Architecture Summary

The 3D Desktop supports two types of external app integration:

| Type | Status | User Experience | Complexity |
|------|--------|-----------------|------------|
| **Web Apps (iframe)** | Production Ready | Seamless, embedded | Low |
| **Native Apps (screen capture)** | Experimental | Limited, problematic | High |

**Recommendation:** Use web apps for all external integrations. Native app capture has fundamental macOS limitations that make it unsuitable for the "embedded app" experience users expect.

---

## Web App Integration (Recommended)

### How It Works

1. User adds an app via the App Registry modal
2. App metadata stored in `user_external_apps` table
3. App appears as an icon on the 3D desktop
4. Clicking opens an iframe window with the app's URL
5. Full interaction within BusinessOS - no focus issues

### Supported Apps

Any web application that allows iframe embedding:

| App | URL | Works in iframe |
|-----|-----|-----------------|
| Notion | notion.so | Yes |
| Linear | linear.app | Yes |
| ClickUp | app.clickup.com | Yes |
| Figma | figma.com | Yes |
| Slack | app.slack.com | Partial (some features) |
| Google Docs | docs.google.com | Yes |

### Adding a Web App

**Via Popular Apps (Quick Add):**
```
1. Open App Registry modal (gear icon)
2. Click on a popular app tile
3. App is instantly added with correct logo
```

**Via Custom URL:**
```
1. Open App Registry modal
2. Switch to "Custom" tab
3. Enter: Name, URL, select icon/color
4. Click "Add App"
```

### Database Schema

```sql
CREATE TABLE user_external_apps (
    id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    icon VARCHAR(100) DEFAULT 'app-window',
    color VARCHAR(7) DEFAULT '#6366F1',
    logo_url TEXT,                          -- Fetched from Google Favicon API
    category VARCHAR(100),
    description TEXT,
    position_x INTEGER DEFAULT 0,
    position_y INTEGER DEFAULT 0,
    position_z INTEGER DEFAULT 0,
    iframe_config JSONB,
    is_active BOOLEAN DEFAULT true,
    open_on_startup BOOLEAN DEFAULT false,
    app_type VARCHAR(50) DEFAULT 'web',     -- 'web' or 'native'
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    last_opened_at TIMESTAMPTZ
);
```

### Frontend Components

```
frontend/src/lib/
├── components/desktop/
│   ├── AppRegistryModal.svelte    # Add/manage apps UI
│   ├── DesktopIcon.svelte         # Renders app icons (supports logo_url)
│   └── Desktop3DWindow.svelte     # Renders iframe for web apps
├── stores/
│   ├── userAppsStore.ts           # User apps state management
│   └── windowStore.ts             # Window/icon registration
```

### Backend Endpoints

```
GET    /api/user-apps?workspace_id={uuid}   # List user's apps
POST   /api/user-apps                        # Create new app
GET    /api/user-apps/:id                    # Get single app
PUT    /api/user-apps/:id                    # Update app
DELETE /api/user-apps/:id                    # Delete app
POST   /api/user-apps/:id/open               # Record app opened
```

---

## Native App Capture (Experimental)

### What We Built

A system to capture and display native macOS applications inside BusinessOS windows:

1. **Screen Capture** - Uses macOS `CGWindowListCreateImage` API
2. **WebSocket Streaming** - Sends JPEG frames at 30fps
3. **Input Forwarding** - Sends clicks/keyboard to the real app

### Why It Doesn't Work Well

After extensive testing, we discovered fundamental limitations:

#### Problem 1: Mouse Hijacking

When forwarding mouse movements to the captured app, the system cursor physically moves to that app's window position, taking control away from the user.

**Original Code (problematic):**
```go
case "mousemove":
    InjectMouseMove(bounds, input.X, input.Y)  // Moves REAL cursor!
```

**Fix Applied:**
```go
case "mousemove":
    // Intentionally NOT forwarding - would hijack system cursor
    return
```

#### Problem 2: Window Must Be Visible

macOS screen capture APIs require the window to be **visible on screen**:

```c
CFArrayRef windowList = CGWindowListCopyWindowInfo(
    kCGWindowListOptionOnScreenOnly,  // <-- Only visible windows!
    kCGNullWindowID
);
```

- Minimized windows cannot be captured (no content exists)
- Windows must be on the current Space/Desktop
- Off-screen windows may not render content

#### Problem 3: Focus Stealing

When user clicks inside the capture canvas:
1. Click is forwarded to real app coordinates
2. macOS brings that app to the foreground
3. User loses focus on BusinessOS
4. Defeats the purpose of "embedded" experience

```go
case "click":
    BringWindowToFront(s.windowID)  // <-- App takes over!
    InjectMouseClick(bounds, input.X, input.Y, input.Button, true)
```

#### Problem 4: Permission Complexity

Requires macOS Screen Recording permission:
- User must manually grant in System Preferences
- Permission prompt can be confusing
- Some enterprise Macs have this locked down

### Files Involved

```
desktop/backend-go/internal/
├── windowcapture/
│   ├── capture_darwin.go    # CGWindowListCreateImage capture
│   ├── input_darwin.go      # CGEventPost input injection
│   ├── stream.go            # WebSocket frame streaming
│   └── *_other.go           # Stubs for non-macOS
├── handlers/
│   └── window_capture.go    # WebSocket endpoint

frontend/src/lib/components/desktop/
└── NativeAppCapture.svelte  # Capture viewer component
```

---

## Technical Limitations

### macOS Screen Capture Constraints

| Constraint | Impact |
|------------|--------|
| `kCGWindowListOptionOnScreenOnly` | Can only capture visible windows |
| `CGWindowListCreateImage` | Requires window to have rendered content |
| `CGEventPost` | Input injection moves real system cursor |
| Minimized windows | Have no capturable content |
| Different Spaces | Windows on other desktops not accessible |

### What Would Be Needed for True "Embedded" Native Apps

1. **Virtual Display Driver** - Create a fake display for apps to render to
2. **Headless Window Server** - Run apps without visible windows
3. **Custom Input Routing** - Intercept input without affecting system cursor
4. **App Sandboxing** - Isolate app rendering from main display

This would essentially require building a VM or container system - far beyond the scope of simple screen capture.

---

## Implementation Guide

### Adding a New Popular App

1. Edit `AppRegistryModal.svelte`:

```typescript
const popularApps = [
    // Add new app here
    {
        name: 'Asana',
        url: 'https://app.asana.com',
        icon: 'check-circle',
        color: '#F06A6A',
        logo: 'https://www.google.com/s2/favicons?domain=asana.com&sz=128',
        category: 'project-management',
        description: 'Work management platform'
    },
    // ... existing apps
];
```

2. The logo URL uses Google's Favicon API:
```
https://www.google.com/s2/favicons?domain={domain}&sz=128
```

### Handling iframe Restrictions

Some apps block iframe embedding via `X-Frame-Options` or CSP headers. Options:

1. **Proxy through backend** - Strip blocking headers (complex, legal concerns)
2. **Use app's embed URL** - Some apps have special embed endpoints
3. **Mark as "open in new tab"** - Fallback for incompatible apps

### Checking if App Allows iframe

```javascript
// In browser console, try:
fetch('https://app.example.com', { mode: 'cors' })
    .then(r => console.log(r.headers.get('x-frame-options')))
```

If response is `DENY` or `SAMEORIGIN`, the app won't work in iframe.

---

## Future Possibilities

### Short Term (Web Apps)

- [ ] Add more popular apps to quick-add list
- [ ] Auto-detect favicon for custom URLs
- [ ] "Open in new tab" fallback for blocked iframes
- [ ] App categories and search

### Medium Term (Enhanced Integration)

- [ ] OAuth integration for apps (show user's actual data)
- [ ] App-specific API integrations (Linear issues, Notion pages)
- [ ] Notification badges on app icons

### Long Term (Native Apps - If Revisited)

- [ ] Investigate macOS ScreenCaptureKit (12.3+) for better APIs
- [ ] Virtual display approach using CGVirtualDisplay
- [ ] Electron-style app wrapping (run web version as native)
- [ ] Linux/Windows support via different capture methods

---

## Decision Record

**Date:** January 2026
**Decision:** Deprioritize native app capture, focus on web apps
**Rationale:**

1. Web apps provide seamless embedded experience
2. Native capture has fundamental macOS limitations
3. Development effort better spent on web app features
4. Most popular productivity tools have web versions

**Consequences:**

- Users cannot embed truly native macOS apps (Finder, Preview, etc.)
- Apps without web versions are not supported
- Simpler codebase, fewer permission issues
- Better cross-platform potential (web apps work everywhere)

---

## Related Documentation

- [3D Desktop Architecture](./3D_DESKTOP_PHASE_STATUS.md)
- [Gesture System](./3D_DESKTOP_GESTURE_SYSTEM.md)
- [Backend CLAUDE.md](/desktop/backend-go/CLAUDE.md)
