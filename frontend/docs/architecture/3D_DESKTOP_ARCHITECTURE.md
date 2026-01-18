# 3D Desktop Architecture

## Overview

The 3D Desktop is an experimental spatial interface that renders BusinessOS modules as floating windows orbiting around a central space. Users can navigate between windows, focus on specific modules, and switch between different view modes.

## File Structure

```
frontend/src/
├── lib/
│   ├── stores/
│   │   └── desktop3dStore.ts          # State management for 3D desktop
│   │
│   └── components/
│       └── desktop3d/
│           ├── Desktop3D.svelte        # Main container component
│           ├── Desktop3DScene.svelte   # Threlte 3D scene (camera, lighting)
│           ├── Desktop3DWindow.svelte  # Individual 3D window with iframe
│           ├── Desktop3DControls.svelte # UI overlay (exit, rotate, view toggle)
│           └── Desktop3DDock.svelte    # Bottom dock for module selection
```

## Component Hierarchy

```
Desktop3D.svelte
├── MenuBar (shared with regular desktop)
├── Canvas (Threlte)
│   └── Desktop3DScene.svelte
│       ├── Camera + OrbitControls
│       ├── Lighting (ambient + directional)
│       ├── Central Orb (visual anchor)
│       └── Desktop3DWindow.svelte (for each open module)
│           └── <HTML> component
│               └── iframe (module content)
├── Desktop3DControls.svelte (overlay)
└── Desktop3DDock.svelte (bottom navigation)
```

## State Management

### desktop3dStore.ts

The store manages all 3D desktop state using Svelte's writable stores.

#### State Interface

```typescript
interface Desktop3DState {
  viewMode: 'orb' | 'grid' | 'focused';  // Current view layout
  windows: Window3DState[];               // All window states
  focusedWindowId: string | null;         // Currently focused window
  sphereRadius: number;                   // Orb layout radius (95)
  gridColumns: number;                    // Grid layout columns (4)
  gridSpacing: number;                    // Grid spacing (130)
  autoRotate: boolean;                    // Camera auto-rotation
  animating: boolean;                     // Animation in progress
}

interface Window3DState {
  id: string;                             // Unique window ID
  module: ModuleId;                       // Module type
  title: string;                          // Display title
  position: [x, y, z];                    // Current 3D position
  targetPosition: [x, y, z];              // Animation target
  rotation: [x, y, z];                    // Rotation angles
  scale: number;                          // Current scale
  targetScale: number;                    // Animation target scale
  opacity: number;                        // Current opacity
  targetOpacity: number;                  // Animation target
  isCore: boolean;                        // Core module (can't close)
  isOpen: boolean;                        // Window open state
  isFocused: boolean;                     // Focus state
  lastFocused: number;                    // Last focus timestamp
  color: string;                          // Module color
  width: number;                          // Window width (px)
  height: number;                         // Window height (px)
}
```

#### Core Modules

These modules are always visible and cannot be closed:

```typescript
const CORE_MODULES = [
  'dashboard', 'chat', 'tasks', 'projects', 'team',
  'clients', 'calendar', 'knowledge', 'nodes', 'daily',
  'terminal', 'settings', 'help'
];
```

#### Store Actions

| Action | Description |
|--------|-------------|
| `initialize()` | Create windows for all core modules |
| `recalculatePositions()` | Update window positions based on view mode |
| `setViewMode(mode)` | Switch between orb/grid/focused |
| `toggleViewMode()` | Toggle orb <-> grid |
| `focusWindow(id)` | Focus on a specific window |
| `unfocusWindow()` | Return to orb view |
| `openWindow(module)` | Open a new module window |
| `closeWindow(id)` | Close window (non-core only) |
| `toggleAutoRotate()` | Toggle camera auto-rotation |
| `focusNext()` / `focusPrevious()` | Navigate between windows |
| `resizeFocusedWindow(w, h)` | Resize focused window |

## Layout Algorithms

### Orb Layout (Ring-Based Sphere)

Windows are arranged in 3 horizontal rings around a sphere:

```
Ring Layout Distribution:
- 1-3 windows:  1 ring (middle)
- 4-6 windows:  2 rings (top, bottom)
- 7-9 windows:  3 rings (top, middle, bottom)
- 10+ windows:  3 rings (3 top, N middle, 3 bottom)

Ring Heights:
- Top ring:    y = +0.6 * radius
- Middle ring: y = 0
- Bottom ring: y = -0.6 * radius

Ring Offsets:
- Each ring is offset by 60 degrees to prevent vertical alignment
```

Windows face INWARD toward the center and TILT based on vertical position:
- Top windows (positive Y) tilt DOWNWARD (~45 degrees)
- Bottom windows (negative Y) tilt UPWARD (~45 degrees)

### Grid Layout (Flat Spread)

Windows arranged in a flat grid:
- 4 columns
- 130 unit spacing
- Centered at z=0 plane

## Components

### Desktop3D.svelte

Main container that orchestrates all 3D desktop functionality.

**Responsibilities:**
- Initialize store on mount
- Handle keyboard shortcuts
- Coordinate between scene, controls, and dock
- Render focused window title bar

**Keyboard Shortcuts:**
| Key | Action |
|-----|--------|
| `Escape` | Unfocus window or exit 3D mode |
| `Space` | Toggle orb/grid view (when not focused) |
| `Arrow Left/Right` | Navigate windows (when focused) |
| `+` / `-` | Resize focused window |
| `1-9` | Focus window by index |

### Desktop3DScene.svelte

Threlte scene setup with camera, lighting, and window rendering.

**Camera Configuration:**
```typescript
OrbitControls:
  minDistance: 30
  maxDistance: 200
  maxPolarAngle: Math.PI * 0.85  // Prevent going below
  minPolarAngle: 0.1
  enableDamping: true
  dampingFactor: 0.05
  autoRotate: conditional
  autoRotateSpeed: 0.5
```

**Lighting:**
- Ambient light (intensity: 0.8)
- Directional light from above-right

**Central Orb:**
- Decorative sphere at origin
- Slight transparency (0.8 opacity)
- Visible reference point

### Desktop3DWindow.svelte

Individual 3D window using Threlte's `<HTML>` component.

**Rotation System:**
Uses nested T.Group elements for correct rotation order:
1. Outer group: Position animation
2. Middle group: Y rotation (facing direction)
3. Inner group: X rotation (tilt)

```svelte
<T.Group position={$animatedPosition}>
  <T.Group rotation={[0, $animatedYRotation, 0]}>
    <T.Group rotation={[$animatedXRotation, 0, 0]}>
      <HTML>
        <!-- Window content -->
      </HTML>
    </T.Group>
  </T.Group>
</T.Group>
```

**Spring Animations:**
- Position: stiffness 0.08, damping 0.7
- Y Rotation: stiffness 0.08, damping 0.7
- X Rotation (tilt): stiffness 0.08, damping 0.7
- Scale: stiffness 0.15, damping 0.8
- Opacity: stiffness 0.15, damping 0.8

**Focus Behavior:**
- Focused window: Full opacity, 1.5x scale
- Other windows: 30% opacity, 0.8x scale

### Desktop3DControls.svelte

Overlay UI controls positioned over the 3D canvas.

**Positions:**
- Top left: Exit button
- Top right: Auto-rotate toggle, View mode toggle
- Bottom center: Instructions

**Instructions:**
- Unfocused: Drag, Scroll, Click, Space, Esc
- Focused: Click outside or Esc to unfocus

### Desktop3DDock.svelte

Bottom navigation dock mirroring the regular desktop dock.

**Features:**
- Shows all open modules as icons
- Hover tooltips with module names
- Respects desktop icon style settings
- Highlights focused module
- Click to focus/open module

## View Modes

### Orb Mode (Default)

```
          [Knowledge]
              |
    [Chat]---[ORB]---[Tasks]
              |
          [Projects]
```

- Windows orbit around central sphere
- Camera can rotate freely
- Windows face inward with vertical tilt
- Auto-rotate optional

### Grid Mode

```
[Dashboard] [Chat]  [Tasks]   [Projects]
[Team]      [Clients] [Calendar] [Knowledge]
[Nodes]     [Daily] [Terminal] [Settings]
```

- Windows laid out in flat 4-column grid
- Camera positioned front-center
- No auto-rotation
- Easier overview of all windows

### Focused Mode

```
                    ┌─────────────────────┐
                    │                     │
    [faded]         │   FOCUSED WINDOW    │        [faded]
                    │                     │
                    └─────────────────────┘
```

- Selected window comes forward (1.5x scale)
- Other windows fade (30% opacity, 0.8x scale)
- Navigation arrows appear on sides
- Title bar with resize controls at top

## Interactions

| Action | Result |
|--------|--------|
| Click window | Focus (animate forward) |
| Click outside focused | Unfocus (return to orb) |
| Drag (unfocused) | Orbit camera around orb |
| Scroll | Zoom in/out |
| Space key | Toggle orb/grid view |
| Escape | Unfocus or exit 3D mode |
| Dock click | Focus that module |
| Arrow keys | Navigate between windows |

## Dark Mode Support

All components support dark mode via `:global(.dark)` CSS selectors:

- Background gradient adapts (light gray -> dark gray)
- Control buttons use dark backgrounds
- Instructions panel uses dark theme
- Dock uses dark backgrounds with light text
- All hover states adjusted for dark mode

## Performance Considerations

- **Lazy Loading**: Iframes load module content on-demand
- **Max Windows**: 12-15 recommended for smooth performance
- **Animation Springs**: Balanced for smoothness vs. responsiveness
- **Pointer Events**: Disabled on unfocused windows to reduce overhead

## Entry Points

The 3D Desktop is accessed via:
1. MenuBar toggle in regular desktop
2. Settings -> Desktop -> Enable 3D Mode
3. Keyboard shortcut (configurable)

## Dependencies

- **@threlte/core**: Svelte wrapper for Three.js
- **@threlte/extras**: HTML component for DOM in 3D
- **three**: 3D rendering engine
- **svelte/motion**: Spring animations
