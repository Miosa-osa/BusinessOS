# 🎯 3D Desktop Gesture Control System - Production Specification

**Version**: 2.0
**Date**: January 14, 2026
**Status**: IMPLEMENTING

---

## 🎬 User Experience Goals

### Core Principle
**Gestures should work EXACTLY like using a mouse** - stable, responsive, and predictable.

### User Expectations
1. **Make a fist** → Hand says "DRAGGING" and STAYS that way
2. **Move your fist** → 3D view rotates smoothly (like dragging with mouse)
3. **Open your hand** → Stops dragging
4. **NO flickering** between gestures
5. **Hand NEVER gets lost** during gestures
6. **Instant visual feedback** - current gesture always visible

---

## 🤚 Gesture Catalog

### 1. FIST (Grab & Rotate)
**How to perform**: Close all fingers into a fist
**Action**: DRAG TO ROTATE
**Behavior**:
- **LOCK**: Once fist detected, STAY in fist mode
- **Movement**: Horizontal hand movement = horizontal camera rotation
- **Movement**: Vertical hand movement = vertical camera rotation
- **Speed**: 1:1 mapping (hand moves 10cm → camera rotates proportionally)
- **Exit**: Only exits when hand FULLY opens
- **Display**: "DRAGGING" (not "rotate_continuous")

**Technical**:
```typescript
Trigger: avgDistance(fingertips to wrist) < 0.18
Lock: Once triggered, don't check for other gestures until hand opens
Delta: Calculate X/Y movement delta from previous frame
Rotation: Apply delta directly to camera azimuth/polar angles
Cooldown: 16ms (60 FPS updates)
```

---

### 2. PINCH (Zoom)
**How to perform**: Touch thumb tip to index finger tip
**Action**: ZOOM IN/OUT
**Behavior**:
- **LOCK**: Once pinching, STAY in pinch mode
- **Movement**: Move hand forward = zoom in, backward = zoom out
- **Speed**: Sensitive enough for precise control
- **Exit**: Only exits when fingers separate (distance > threshold)
- **Display**: "ZOOMING"

**Technical**:
```typescript
Trigger: distance(thumb_tip, index_tip) < 0.10
Lock: Once triggered, block fist detection
Delta: Calculate Z-axis movement
Zoom: Adjust camera distance proportionally
Cooldown: 16ms (60 FPS updates)
```

---

### 3. OPEN PALM (Neutral/Idle)
**How to perform**: All fingers extended
**Action**: NONE (idle state)
**Behavior**:
- **No action triggered**
- **Resets all gesture locks**
- **Ready to detect next gesture**
- **Display**: "READY" or nothing

---

### 4. POINT (Optional - Future)
**How to perform**: Index finger extended, others closed
**Action**: HOVER/SELECT
**Behavior**: TBD

---

### 5. THUMBS UP (Optional - Future)
**How to perform**: Thumb up, fingers closed
**Action**: CONFIRM/UNFOCUS
**Behavior**: TBD

---

## 🔒 Gesture State Machine

```
┌─────────────────────────────────────────┐
│           IDLE (Open Palm)              │
│  - No gesture active                    │
│  - Checking for fist or pinch           │
└────────────┬────────────────────────────┘
             │
      ┌──────┴──────┐
      │             │
      ▼             ▼
┌──────────┐  ┌───────────┐
│   FIST   │  │   PINCH   │
│  LOCKED  │  │  LOCKED   │
└────┬─────┘  └─────┬─────┘
     │              │
     │  Open hand   │  Separate fingers
     └────────┬─────┘
              │
              ▼
         IDLE (return)
```

**Critical Rules**:
1. **ONCE LOCKED, STAY LOCKED** until exit condition met
2. **NO flickering** between gesture types
3. **Clear exit conditions** (must fully open hand or separate fingers)
4. **Priority order**: Check FIST first, then PINCH (never both)

---

## 📊 Hand Tracking Performance

### Target Metrics
- **FPS**: 25-30 FPS minimum
- **Latency**: < 50ms from hand movement to visual update
- **Hand Detection**: Never lose hand during active gesture
- **Jitter**: Smoothing applied to prevent shaky movements

### Camera Resolution Strategy
```
Resolution | FPS  | Quality | Use Case
-----------|------|---------|----------
640x480    | 15   | Best    | High-end desktop, accuracy critical
480x360    | 20   | Good    | Balanced (DEFAULT)
320x240    | 30   | OK      | Performance critical, older hardware
```

**Current Setting**: 320x240 for 25-30 FPS
**Upgrade Path**: Auto-detect device capability and adjust

---

## 🎨 Visual Feedback

### Gesture Status Display

**Current State** (always visible in debug view):
```
┌─────────────────────────────────────┐
│ Current Gesture: DRAGGING           │  ← User-friendly name
│ Confidence: ████████░░ 85%          │  ← How sure we are
│ Action: Rotating camera             │  ← What's happening
└─────────────────────────────────────┘
```

### Gesture Log

**OLD (Technical)**:
```
14:37:50.852  fist      → rotate_continuous
14:37:50.089  fist      → rotate_continuous
14:37:49.728  fist      → grab_mode
```

**NEW (User-Friendly)**:
```
14:37:50  🤜 DRAGGING  (rotating view)
14:37:49  🤜 GRAB START
14:37:48  🤏 ZOOMING IN
```

### Hand Detection Indicator

```
✅ Hands: 1  ← Green = good
⚠️ Hands: 0  ← Yellow = lost (warning)
❌ Tracking OFF ← Red = disabled
```

---

## 💻 Technical Implementation

### File Structure
```
src/lib/
├── services/
│   ├── handTrackingService.ts    - MediaPipe integration (30 FPS)
│   ├── gestureDetector.ts        - Gesture recognition with state machine
│   └── gestureStateManager.ts    - NEW: Manages gesture locking
├── components/desktop3d/
│   ├── GestureDebugView.svelte   - Camera + debug UI
│   ├── Desktop3D.svelte          - Main 3D desktop
│   └── Desktop3DScene.svelte     - 3D rendering (OrbitControls)
├── stores/
│   └── desktop3dStore.ts         - State management + camera control
└── types/
    └── gestures.ts               - Type definitions
```

### Key Components

#### 1. gestureDetector.ts
```typescript
class GestureDetector {
  private currentGesture: 'idle' | 'fist' | 'pinch' = 'idle';
  private gestureLocked: boolean = false;

  detect(hand: HandLandmarks): GestureState {
    // If locked in a gesture, only check for exit condition
    if (this.gestureLocked) {
      return this.checkExitCondition(hand);
    }

    // Check gestures in priority order
    // 1. Fist (highest priority)
    if (this.isFist(hand)) {
      this.lockGesture('fist');
      return { type: 'fist', action: 'drag', locked: true };
    }

    // 2. Pinch (second priority)
    if (this.isPinch(hand)) {
      this.lockGesture('pinch');
      return { type: 'pinch', action: 'zoom', locked: true };
    }

    // 3. Idle (default)
    return { type: 'idle', action: 'none', locked: false };
  }

  checkExitCondition(hand: HandLandmarks): GestureState {
    if (this.currentGesture === 'fist') {
      // Exit if hand opens (all fingers extended)
      if (this.isOpenPalm(hand)) {
        this.unlockGesture();
        return { type: 'idle', action: 'none', locked: false };
      }
      // Still in fist, calculate movement delta
      return this.calculateFistDrag(hand);
    }

    if (this.currentGesture === 'pinch') {
      // Exit if fingers separate
      if (this.fingersApart(hand)) {
        this.unlockGesture();
        return { type: 'idle', action: 'none', locked: false };
      }
      // Still pinching, calculate zoom delta
      return this.calculatePinchZoom(hand);
    }
  }
}
```

#### 2. desktop3dStore.ts
```typescript
export interface Desktop3DState {
  // ... existing fields ...

  // Camera rotation control (NEW)
  cameraRotationDelta: { x: number; y: number };
  gestureDragging: boolean;
}

// Make rotation actually work
adjustRotationSpeed: (deltaX: number, deltaY: number) => {
  update((state) => ({
    ...state,
    cameraRotationDelta: { x: deltaX, y: deltaY },
    gestureDragging: true
  }));

  // Auto-clear after 100ms to stop rotation if no new delta comes
  setTimeout(() => {
    update((state) => ({
      ...state,
      cameraRotationDelta: { x: 0, y: 0 },
      gestureDragging: false
    }));
  }, 100);
}
```

#### 3. Desktop3DScene.svelte
```svelte
<script>
  // Receive rotation delta from store
  let {
    cameraRotationDelta = { x: 0, y: 0 },
    gestureDragging = false
  }: Props = $props();

  // Apply rotation to OrbitControls
  $effect(() => {
    if (gestureDragging && orbitControlsRef) {
      const controls = orbitControlsRef;

      // Apply delta to camera rotation
      // X delta = horizontal rotation (azimuth)
      // Y delta = vertical rotation (polar)
      controls.azimuthalAngle += cameraRotationDelta.x * 0.5;
      controls.polarAngle += cameraRotationDelta.y * 0.5;

      // Clamp polar angle (don't flip upside down)
      controls.polarAngle = Math.max(0.1, Math.min(Math.PI - 0.1, controls.polarAngle));

      controls.update();
    }
  });
</script>
```

#### 4. Desktop3D.svelte (Gesture Handler)
```typescript
function handleGesture(gesture: GestureState) {
  if (!gestureControlEnabled) return;

  // Update visual display
  currentGestureDisplay = getDisplayName(gesture);

  // Handle actions
  switch (gesture.action) {
    case 'drag':
      if (gesture.deltaPosition) {
        const rotX = gesture.deltaPosition.x * 5.0;
        const rotY = gesture.deltaPosition.y * 5.0;
        desktop3dStore.adjustRotationSpeed(rotX, rotY);
      }
      break;

    case 'zoom':
      if (gesture.deltaPosition) {
        const zoomDelta = gesture.deltaPosition.z * 100;
        desktop3dStore.adjustCameraDistance(zoomDelta);
      }
      break;
  }
}

function getDisplayName(gesture: GestureState): string {
  switch (gesture.action) {
    case 'drag': return 'DRAGGING';
    case 'zoom': return 'ZOOMING';
    case 'none': return 'READY';
    default: return gesture.action.toUpperCase();
  }
}
```

---

## 🔧 Configuration

### Gesture Thresholds
```typescript
export const PRODUCTION_GESTURE_CONFIG = {
  // Hand tracking
  maxHands: 1,                    // Track 1 hand only
  modelComplexity: 0,             // LITE model for performance
  minDetectionConfidence: 0.6,     // Don't lose hands easily
  minTrackingConfidence: 0.6,      // Don't lose hands easily

  // Gesture detection
  fistThreshold: 0.18,            // Stricter fist detection
  pinchThreshold: 0.10,           // Stricter pinch detection

  // Movement sensitivity
  minimumMovementDelta: 0.01,     // Minimum movement to register
  smoothingFactor: 0.7,           // Smoothing (0-1, higher = smoother)

  // Update rates
  gestureUpdateIntervalMs: 16,    // 60 FPS gesture detection
  cameraUpdateIntervalMs: 16,     // 60 FPS camera updates

  // Locking
  gestureLockEnabled: true,       // CRITICAL: Enable gesture locking
  exitThreshold: 0.25,            // How "open" hand must be to exit fist
};
```

### Camera Resolution
```typescript
// Auto-detect best resolution for device
export function getOptimalResolution(): { width: number; height: number } {
  // Test with highest resolution first
  const testSizes = [
    { width: 640, height: 480, targetFPS: 20 },
    { width: 480, height: 360, targetFPS: 25 },
    { width: 320, height: 240, targetFPS: 30 },
  ];

  // TODO: Run performance test and pick best
  // For now, use middle option
  return { width: 480, height: 360 };
}
```

---

## ✅ Success Criteria

### Functional Requirements
- [ ] Fist gesture LOCKS and stays locked until hand opens
- [ ] Moving fist rotates 3D view smoothly (like mouse drag)
- [ ] Pinch gesture LOCKS and stays locked until fingers separate
- [ ] Moving pinch zooms camera in/out
- [ ] Hand tracking never loses hand during active gesture
- [ ] FPS stays above 25 FPS
- [ ] Gesture display shows user-friendly names (DRAGGING not rotate_continuous)
- [ ] No flickering between gesture types

### Performance Requirements
- [ ] 25+ FPS hand tracking
- [ ] < 50ms latency from gesture to visual update
- [ ] < 5% hand detection loss rate during gesture
- [ ] Smooth camera movement (no jitter)

### UX Requirements
- [ ] Clear visual feedback for current gesture
- [ ] Intuitive gesture feel (like using mouse)
- [ ] Predictable behavior (no surprises)
- [ ] Responsive (immediate feedback)

---

## 🚀 Implementation Plan

### Phase 1: Core Fixes (NOW)
1. ✅ Fix `adjustRotationSpeed` to actually rotate camera
2. ✅ Add gesture locking to prevent flickering
3. ✅ Update gesture detector thresholds
4. ✅ Change display names to user-friendly labels
5. ✅ Improve hand tracking confidence

### Phase 2: Polish
1. Add gesture state machine
2. Smooth camera movements
3. Auto-detect optimal resolution
4. Add gesture confidence meter

### Phase 3: Advanced Features
1. Two-hand gestures (expand/contract)
2. Gesture combinations
3. Haptic feedback (if supported)
4. Voice + gesture combos

---

## 📝 Testing Protocol

### Manual Testing
1. **Fist Lock Test**
   - Make fist → Should say "DRAGGING" or "GRABBED"
   - Move fist around → View should rotate smoothly
   - Gesture display should NOT flicker
   - Only exits when hand fully opens

2. **Pinch Lock Test**
   - Pinch fingers → Should say "ZOOMING"
   - Move hand forward/back → Camera should zoom
   - Should NOT switch to fist during pinch
   - Only exits when fingers separate

3. **Performance Test**
   - Check FPS counter → Should be 25-30 FPS
   - Hand tracking should NEVER show "Hands: 0" during gesture
   - Camera movement should be smooth (no stuttering)

4. **Transition Test**
   - Idle → Fist → Idle: Clean transitions
   - Idle → Pinch → Idle: Clean transitions
   - No accidental gesture triggers

### Automated Testing (Future)
- Unit tests for gesture detector
- Integration tests for camera control
- Performance benchmarks

---

## 🐛 Known Issues (To Fix)

### Critical
- [x] `adjustRotationSpeed` does nothing (stub function)
- [x] Gesture flickering between fist and pinch
- [x] Hand tracking loses hands (shows "Hands: 0")
- [x] Display shows technical names (rotate_continuous) not user names

### High Priority
- [ ] Camera rotation not applied to OrbitControls
- [ ] Gesture locking not implemented
- [ ] Movement deltas too small to register

### Medium Priority
- [ ] FPS only 12.3 (should be 25+)
- [ ] Resolution could be optimized
- [ ] Smoothing could be better

---

**END OF SPECIFICATION**
