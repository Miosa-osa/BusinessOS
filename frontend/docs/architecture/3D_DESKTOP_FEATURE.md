# 3D Desktop Feature

## Overview
An experimental 3D spatial desktop where application modules are arranged in a sphere (orb) formation. Users can rotate the view, focus on windows, and navigate between modules in an immersive 3D environment.

## Design Principles

### Spatial Computing
The 3D desktop introduces spatial computing concepts to BusinessOS:
- **Depth as hierarchy** - Focused content comes forward, context fades back
- **Physical metaphors** - Windows exist as objects you can rotate around
- **Peripheral awareness** - Side previews let you see what's adjacent

### Non-Destructive Exploration
- Dragging rotates the view, never moves windows
- Click detection distinguishes between drag (rotate) and click (focus)
- Camera returns to sensible positions but allows free exploration

### Progressive Disclosure
- Orb view shows all modules at a glance
- Focus mode brings one module to full interactivity
- Grid view provides flat overview when spatial isn't needed

## What Was Added

### New Components
```
src/lib/components/desktop3d/
  Desktop3D.svelte          # Main container, keyboard handling
  Desktop3DScene.svelte     # Threlte scene setup, hover management
  Desktop3DWindow.svelte    # 3D window with click/drag detection
  Desktop3DControls.svelte  # UI overlay controls
  Desktop3DDock.svelte      # Bottom module dock
```

### New Store
```
src/lib/stores/desktop3dStore.ts
```
Manages:
- View mode (orb | grid | focused)
- Window positions and states
- Focus/unfocus transitions
- Auto-rotation toggle

### New Routes
```
src/routes/(app)/terminal/+page.svelte  # Terminal module page
```

### Modified Files
- `window/+page.svelte` - Added Knowledge iframe (was missing)
- `knowledge/+page.svelte` - Auto-switch from graph to list view on page select

## How It Works

### Ring-Based Sphere Layout
Windows are arranged in rings around a sphere:
- **Top ring**: 3 windows at y=0.7
- **Middle ring**: 7 windows at y=0 (equator)
- **Bottom ring**: 3 windows at y=-0.7

Each ring is offset by 60 degrees to prevent vertical alignment.

### Click vs Drag Detection
The system distinguishes clicks from drags using:
1. **Distance threshold** (15px) - Movement beyond this = drag
2. **Time threshold** (500ms) - Held too long = not a click
3. **isDragging state** - Tracks if user started dragging

### 3D Click Meshes
Each window has an invisible sphere mesh for 3D raycasting:
- Sphere geometry (radius 50) captures clicks from any angle
- Works regardless of window rotation
- Separate from visual HTML content

### HTML in 3D Space
Uses Threlte's `<HTML>` component to render DOM in 3D:
- `transform` prop enables 3D positioning
- `pointerEvents` toggled based on focus state
- Iframes load each module with `?embed=true`

### Hover Management
Scene-level state ensures only one window highlights at a time:
```typescript
let hoveredWindowId: string | null = $state(null);
```
Each window reports hover in/out, scene updates single source of truth.

## Core Modules (13)
1. Dashboard - Main overview
2. Chat - AI conversation
3. Tasks - Task management
4. Projects - Project tracking
5. Team - Team management
6. Clients - Client database
7. Calendar - Scheduling
8. Knowledge - Knowledge graph + docs
9. Nodes - Node management
10. Daily - Daily log
11. Terminal - Command line
12. Settings - Configuration
13. Help - Documentation

## Keyboard Shortcuts
| Key | Action |
|-----|--------|
| Space | Toggle orb/grid view |
| Escape | Unfocus or exit 3D mode |
| Arrow Left/Right | Navigate between focused windows |
| +/- | Resize focused window |
| 1-9 | Focus window by index |

## Technical Stack
- **Threlte** - Three.js integration for Svelte 5
- **@threlte/extras** - HTML component, OrbitControls
- **svelte/motion** - Spring animations
- **Three.js** - 3D rendering engine

## Configuration (desktop3dStore.ts)
```typescript
sphereRadius: 65        // Orb size
gridColumns: 4          // Grid layout columns
gridSpacing: 130        // Grid cell spacing
autoRotate: true        // Default rotation state
```

## Window Sizing
- Default: 1300x900px
- Min: 800x500px
- Max: 1600x1100px
- Resize via +/- buttons or keyboard

## Performance Considerations
- All 13 iframes load simultaneously (lazy loading possible future enhancement)
- Spring animations are GPU-accelerated
- Invisible meshes use minimal geometry (12-segment spheres)
- HTML component uses CSS3DRenderer for efficient DOM-in-3D

## Future Enhancements
- Lazy loading iframes (only load when near camera)
- Custom window arrangements (drag to reposition)
- Save/restore 3D desktop layouts
- VR/AR support via WebXR
