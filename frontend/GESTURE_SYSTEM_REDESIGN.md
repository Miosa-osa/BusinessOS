# 🎮 3D Desktop Gesture System - Complete Redesign

## Problem with Current System

**Current:** Every small hand movement triggers actions → laggy, spammy, unusable

**Solution:** **MODE-BASED GESTURES** - Lock into a mode, then track movement

---

## 🎯 New Gesture System Design

### Core Concept: State Machine

```
NEUTRAL (no gesture)
    ↓
DETECT gesture type → LOCK into MODE
    ↓
TRACK hand movement → Execute actions based on mode
    ↓
RELEASE gesture → Return to NEUTRAL
```

---

## 🖐️ MODE 1: FIST = Camera Rotation

**How to activate:**
1. Make a **FIST** (all fingers closed tight)
2. System **LOCKS** into "Rotation Mode"
3. Visual indicator shows: 🔴 **ROTATION MODE**

**While in Rotation Mode:**
- **Move hand LEFT** → Camera rotates LEFT
- **Move hand RIGHT** → Camera rotates RIGHT
- **Move hand UP** → Camera tilts UP
- **Move hand DOWN** → Camera tilts DOWN

**Movement tracking:**
- Track WRIST position (x, y coordinates)
- Calculate delta from start position
- `deltaX > 0.05` → rotate right
- `deltaX < -0.05` → rotate left
- `deltaY > 0.05` → tilt down
- `deltaY < -0.05` → tilt up

**How to exit:**
- **Open hand** (release fist)
- System returns to NEUTRAL

**Rules:**
- MUST hold fist for 300ms before mode activates (prevent accidental)
- Movement threshold: 0.05 units (clear, deliberate movements only)
- Rotation speed: `deltaX * 50` (smooth rotation)

---

## 🤚 MODE 2: OPEN PALM = Zoom Control

**How to activate:**
1. Hold **OPEN PALM** (all fingers spread wide)
2. System **LOCKS** into "Zoom Mode"
3. Visual indicator shows: 🔵 **ZOOM MODE**

**While in Zoom Mode:**
- **Move hand FORWARD** (toward camera) → Zoom IN
- **Move hand BACKWARD** (away from camera) → Zoom OUT

**Movement tracking:**
- Track WRIST Z-coordinate (depth)
- Calculate delta from start position
- `deltaZ < -0.05` → zooming in (hand closer)
- `deltaZ > 0.05` → zooming out (hand farther)

**How to exit:**
- **Close hand** (make fist or pinch)
- System returns to NEUTRAL

**Rules:**
- MUST hold palm for 300ms before mode activates
- Movement threshold: 0.05 units in Z direction
- Zoom speed: `deltaZ * 100` (smooth zoom)

---

## 🤏 MODE 3: PINCH = Sphere Size Control

**How to activate:**
1. **PINCH** thumb and index finger together
2. System **LOCKS** into "Size Mode"
3. Visual indicator shows: 🟢 **SIZE MODE**

**While in Size Mode:**
- **Spread fingers WIDER** → EXPAND sphere
- **Squeeze fingers CLOSER** → SHRINK sphere

**Movement tracking:**
- Track distance between thumb tip and index finger tip
- Calculate change from initial pinch distance
- `distance increasing` → expand sphere
- `distance decreasing` → shrink sphere

**How to exit:**
- **Release pinch** (fingers apart > threshold)
- System returns to NEUTRAL

**Rules:**
- Initial pinch distance = baseline
- `distance > baseline + 0.03` → expand
- `distance < baseline - 0.03` → shrink
- Sphere size change: `distanceDelta * 500`

---

## 🔄 State Machine Implementation

```typescript
enum GestureMode {
  NEUTRAL = 'neutral',      // No gesture active
  ROTATION = 'rotation',    // Fist held - tracking hand movement for rotation
  ZOOM = 'zoom',            // Palm held - tracking depth for zoom
  SIZE = 'size'             // Pinch held - tracking pinch distance for size
}

class ModeBasedGestureDetector {
  private currentMode: GestureMode = GestureMode.NEUTRAL;
  private modeStartTime: number = 0;
  private modeStartPosition: HandPosition | null = null;

  // Mode activation delay (prevent accidental triggers)
  private readonly MODE_ACTIVATION_DELAY_MS = 300;

  detect(hand: HandLandmarks): GestureCommand | null {
    const now = performance.now();

    // STEP 1: Detect current hand pose
    const pose = this.detectHandPose(hand);

    // STEP 2: State machine logic
    switch (this.currentMode) {
      case GestureMode.NEUTRAL:
        return this.handleNeutralState(pose, hand, now);

      case GestureMode.ROTATION:
        return this.handleRotationMode(pose, hand, now);

      case GestureMode.ZOOM:
        return this.handleZoomMode(pose, hand, now);

      case GestureMode.SIZE:
        return this.handleSizeMode(pose, hand, now);
    }
  }

  private handleNeutralState(pose, hand, now): GestureCommand | null {
    // Check if user is making a gesture to enter a mode

    if (pose === 'fist') {
      // Start tracking fist - enter rotation mode after delay
      if (!this.modeStartTime) {
        this.modeStartTime = now;
        this.modeStartPosition = hand.wrist;
      }

      // Has user held fist long enough?
      if (now - this.modeStartTime >= this.MODE_ACTIVATION_DELAY_MS) {
        this.currentMode = GestureMode.ROTATION;
        return { type: 'mode_enter', mode: 'rotation' };
      }
    }
    else if (pose === 'open_palm') {
      // Similar logic for zoom mode
      if (!this.modeStartTime) {
        this.modeStartTime = now;
        this.modeStartPosition = hand.wrist;
      }

      if (now - this.modeStartTime >= this.MODE_ACTIVATION_DELAY_MS) {
        this.currentMode = GestureMode.ZOOM;
        return { type: 'mode_enter', mode: 'zoom' };
      }
    }
    else if (pose === 'pinch') {
      // Similar logic for size mode
      if (!this.modeStartTime) {
        this.modeStartTime = now;
        this.modeStartPosition = { ...hand.wrist, pinchDistance: calculatePinchDistance(hand) };
      }

      if (now - this.modeStartTime >= this.MODE_ACTIVATION_DELAY_MS) {
        this.currentMode = GestureMode.SIZE;
        return { type: 'mode_enter', mode: 'size' };
      }
    }
    else {
      // Hand changed pose - reset
      this.modeStartTime = 0;
      this.modeStartPosition = null;
    }

    return null;
  }

  private handleRotationMode(pose, hand, now): GestureCommand | null {
    // Check if still holding fist
    if (pose !== 'fist') {
      // Exit rotation mode
      this.currentMode = GestureMode.NEUTRAL;
      this.modeStartTime = 0;
      this.modeStartPosition = null;
      return { type: 'mode_exit', mode: 'rotation' };
    }

    // Track hand movement for rotation
    const currentPos = hand.wrist;
    const deltaX = currentPos.x - this.modeStartPosition.x;
    const deltaY = currentPos.y - this.modeStartPosition.y;

    // Only trigger rotation if movement is significant
    if (Math.abs(deltaX) > 0.05 || Math.abs(deltaY) > 0.05) {
      return {
        type: 'rotate_camera',
        deltaX: deltaX * 50,  // Scale factor for smooth rotation
        deltaY: deltaY * 50
      };
    }

    return null;
  }

  private handleZoomMode(pose, hand, now): GestureCommand | null {
    // Check if still holding open palm
    if (pose !== 'open_palm') {
      // Exit zoom mode
      this.currentMode = GestureMode.NEUTRAL;
      this.modeStartTime = 0;
      this.modeStartPosition = null;
      return { type: 'mode_exit', mode: 'zoom' };
    }

    // Track depth (Z) movement for zoom
    const currentPos = hand.wrist;
    const deltaZ = currentPos.z - this.modeStartPosition.z;

    // Significant forward/backward movement?
    if (Math.abs(deltaZ) > 0.05) {
      return {
        type: 'zoom_camera',
        deltaZ: deltaZ * 100  // Scale factor
      };
    }

    return null;
  }

  private handleSizeMode(pose, hand, now): GestureCommand | null {
    // Check if still pinching
    if (pose !== 'pinch') {
      // Exit size mode
      this.currentMode = GestureMode.NEUTRAL;
      this.modeStartTime = 0;
      this.modeStartPosition = null;
      return { type: 'mode_exit', mode: 'size' };
    }

    // Track pinch distance change
    const currentPinchDistance = calculatePinchDistance(hand);
    const initialPinchDistance = this.modeStartPosition.pinchDistance;
    const deltaPinch = currentPinchDistance - initialPinchDistance;

    // Significant pinch change?
    if (Math.abs(deltaPinch) > 0.03) {
      return {
        type: 'resize_sphere',
        delta: deltaPinch * 500  // Scale factor
      };
    }

    return null;
  }
}
```

---

## 🎨 Visual Feedback

### Mode Indicator (always visible when in mode)

```
┌─────────────────────────────────┐
│ 🔴 ROTATION MODE                │
│ Move hand to rotate camera      │
│ Release fist to exit            │
└─────────────────────────────────┘
```

```
┌─────────────────────────────────┐
│ 🔵 ZOOM MODE                    │
│ Move forward/back to zoom       │
│ Close hand to exit              │
└─────────────────────────────────┘
```

```
┌─────────────────────────────────┐
│ 🟢 SIZE MODE                    │
│ Spread/squeeze to resize sphere │
│ Release pinch to exit           │
└─────────────────────────────────┘
```

### Hand Position Indicator

Show a dot on screen representing hand position:
- In ROTATION mode: Shows X/Y position
- In ZOOM mode: Shows depth (size changes as you move forward/back)
- In SIZE mode: Shows pinch distance (visual squeeze indicator)

---

## 📊 Gesture Detection Thresholds

### Hand Pose Detection

```typescript
interface PoseThresholds {
  fist: {
    // All fingertips must be within 0.2 units of wrist
    maxDistanceFromWrist: 0.2
  },

  open_palm: {
    // All fingertips must be > 0.25 units from wrist
    minDistanceFromWrist: 0.25,
    // Fingers must be spread (not touching each other)
    minFingerSpacing: 0.08
  },

  pinch: {
    // Thumb and index finger must be < 0.15 units apart
    maxPinchDistance: 0.15
  }
}
```

### Movement Thresholds

```typescript
interface MovementThresholds {
  rotation: {
    minDeltaX: 0.05,  // Must move 5% of screen width
    minDeltaY: 0.05   // Must move 5% of screen height
  },

  zoom: {
    minDeltaZ: 0.05   // Must move 5% in depth
  },

  size: {
    minPinchChange: 0.03  // Pinch distance must change by 3%
  }
}
```

### Timing

```typescript
const GESTURE_TIMINGS = {
  MODE_ACTIVATION_DELAY: 300,   // Hold gesture 300ms before mode activates
  CONTINUOUS_UPDATE_RATE: 100,  // Update camera every 100ms while in mode
  MODE_EXIT_DELAY: 100          // Grace period before exiting mode
};
```

---

## 🧪 Testing Plan

### Test Case 1: Fist Rotation
1. Make fist → wait 300ms → see "🔴 ROTATION MODE"
2. Move hand left → camera rotates left ✓
3. Move hand right → camera rotates right ✓
4. Move hand up → camera tilts up ✓
5. Move hand down → camera tilts down ✓
6. Open hand → mode exits → back to neutral ✓

### Test Case 2: Palm Zoom
1. Open palm → wait 300ms → see "🔵 ZOOM MODE"
2. Move hand toward camera → zoom in ✓
3. Move hand away → zoom out ✓
4. Close hand → mode exits ✓

### Test Case 3: Pinch Size
1. Pinch fingers → wait 300ms → see "🟢 SIZE MODE"
2. Spread fingers wider → sphere expands ✓
3. Squeeze fingers closer → sphere shrinks ✓
4. Release pinch → mode exits ✓

### Test Case 4: Accidental Gestures
1. Make fist for 200ms → release → NO mode activated ✓
2. Make fist for 100ms → palm → fist → NO mode activated ✓
3. Only deliberate, held gestures activate modes ✓

---

## 🎯 Benefits of This System

### 1. **No False Triggers**
- Must hold gesture for 300ms
- Clear intention required

### 2. **Predictable Behavior**
- Each mode does ONE thing
- User knows exactly what will happen

### 3. **Visual Feedback**
- Mode indicator shows current state
- User always knows what mode they're in

### 4. **Clean State Machine**
- Clear enter/exit conditions
- No ambiguity

### 5. **Smooth Performance**
- Only track movement while in mode
- Continuous updates at 100ms intervals
- No gesture spam

---

## 📋 Implementation Checklist

- [ ] Create `ModeBasedGestureDetector` class
- [ ] Implement hand pose detection (fist, palm, pinch)
- [ ] Implement state machine (NEUTRAL → MODE → NEUTRAL)
- [ ] Add 300ms activation delay
- [ ] Add movement tracking for each mode
- [ ] Create mode indicator UI component
- [ ] Add hand position visualization
- [ ] Test all three modes
- [ ] Add smooth camera rotation
- [ ] Add smooth zoom
- [ ] Add smooth sphere resizing

---

## 🚀 Next Steps

1. **Implement this system** (replaces current gesture detector)
2. **Test with MediaPipe** (will still be 5-10 FPS)
3. **If still too slow**, make gesture tracking **optional** (off by default)

**Key insight:** This system is MUCH better than current one, but **MediaPipe speed is still the bottleneck**.

Even with perfect gesture logic, **5-10 FPS tracking will feel laggy**.

**Recommendation:** Implement this system + make gestures optional (disabled by default).

---

**Date:** January 14, 2026
**Status:** Design Complete - Ready for Implementation
**Estimated Time:** 4-6 hours to implement
