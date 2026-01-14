# 🎯 3D Desktop Gesture Control - Final Specification

**Date**: January 14, 2026
**Status**: AUTHORITATIVE SPEC - This is how it MUST work

---

## Core Philosophy

Gestures should work **EXACTLY like using a mouse**:
- **SMOOTH** - No jumps, no resets, continuous motion
- **STABLE** - Once locked in a gesture, STAY there
- **PREDICTABLE** - Same movement always produces same result
- **RESPONSIVE** - Immediate feedback, no lag

---

## 🤜 FIST GESTURE - Drag to Rotate (Like Mouse Drag)

### How It Works
```
1. Make FIST → System says "DRAGGING" → LOCK into drag mode
2. Move fist LEFT → Camera rotates LEFT (orb spins right)
3. Move fist RIGHT → Camera rotates RIGHT (orb spins left)
4. Move fist UP → Camera rotates UP (view from above)
5. Move fist DOWN → Camera rotates DOWN (view from below)
6. Open hand FULLY → Exit drag mode → Back to idle
```

### Technical Requirements
- **Lock Condition**: All 4 fingers (index, middle, ring, pinky) close to wrist (< 0.18 distance)
- **Stay Locked**: Once locked, ignore pinch/other gestures until hand FULLY opens
- **Exit Condition**: All 5 fingertips (including thumb) far from wrist (> 0.30 distance)
- **Movement Mapping**:
  - Hand moves 1cm left → Camera azimuth angle decreases proportionally
  - Hand moves 1cm right → Camera azimuth angle increases proportionally
  - Hand moves 1cm up → Camera polar angle decreases (look from above)
  - Hand moves 1cm down → Camera polar angle increases (look from below)
- **Smoothness**: Apply delta cumulatively, never reset position
- **Sensitivity**: 8.0x multiplier on deltaPosition for responsive feel

### Edge Cases
- **Hand temporarily lost**: Stay in drag mode, resume when hand reappears
- **Partial fist**: If confidence drops but not fully open, keep drag mode active
- **Screen edge**: No special handling, infinite rotation

### Visual Feedback
- **Display**: "DRAGGING" or "🤜 DRAG MODE"
- **Cursor**: Show hand position on screen
- **Camera**: Rotates smoothly in real-time

---

## 🤏 PINCH GESTURE - Zoom In/Out (Like Mouse Wheel)

### How It Works
```
1. PINCH fingers (thumb + index touch) → System says "ZOOMING" → LOCK into zoom mode
2. Move pinched hand TOWARD camera (closer to face) → Modules get BIGGER (zoom IN)
3. Move pinched hand AWAY from camera (toward desk) → Modules get SMALLER (zoom OUT)
4. Separate fingers → Exit zoom mode → Back to idle
```

### Technical Requirements
- **Lock Condition**: Distance between thumb_tip and index_tip < 0.10
- **Stay Locked**: Once locked, ignore fist/other gestures until fingers separate
- **Exit Condition**: Distance between thumb_tip and index_tip > 0.15 (need to separate MORE to exit)
- **Movement Mapping**:
  - Hand moves 1cm TOWARD camera (Z decreases, more negative) → Camera distance DECREASES → ZOOM IN
  - Hand moves 1cm AWAY from camera (Z increases, less negative) → Camera distance INCREASES → ZOOM OUT
- **Z-axis direction**:
  - MediaPipe Z: Negative = closer to camera, Positive = farther
  - deltaPosition.z < 0 → Moving TOWARD camera → ZOOM IN
  - deltaPosition.z > 0 → Moving AWAY → ZOOM OUT
- **Smoothness**: Continuous zoom, not discrete steps
- **Sensitivity**: 150x multiplier on deltaPosition.z

### Edge Cases
- **Hand lost**: Stay in zoom mode, resume when hand reappears
- **Fingers wiggle**: If pinch distance flickers around threshold, stay locked
- **Zoom limits**: Clamp camera distance (min: 100, max: 800)

### Visual Feedback
- **Display**: "ZOOMING IN" or "ZOOMING OUT" based on Z delta direction
- **Cursor**: Show pinch point on screen
- **Camera**: Distance changes smoothly

---

## 👏 CLAP GESTURE - Spread All Modules

### How It Works
```
1. CLAP hands together (audio detection) → Trigger action
2. All modules open and spread out in grid layout
3. Can see everything at once
```

### Technical Requirements
- **Detection**: Use microphone to detect double clap (two peaks within 300ms)
- **Threshold**: Audio level > 180 (loud clap)
- **Cooldown**: 2000ms (prevent accidental double-trigger)
- **Action**:
  ```typescript
  desktop3dStore.spreadAllModules() → {
    // Open all 22 modules
    // Switch to grid view
    // Arrange in 6x4 grid
    // Disable auto-rotate
  }
  ```

### Edge Cases
- **Single clap**: Ignore (need double clap)
- **Background noise**: Threshold prevents false positives
- **Already in grid**: Do nothing or toggle back to orb

### Visual Feedback
- **Display**: "SPREAD ALL" or "📢 CLAP DETECTED"
- **Animation**: Modules fly out from center to grid positions
- **Duration**: 1 second smooth transition

---

## 👐 OPEN PALM - Idle/Reset

### How It Works
```
ALL fingers extended → Neutral state → Ready for next gesture
```

### Technical Requirements
- **Condition**: All 5 fingertips far from wrist (> 0.30 distance)
- **Purpose**: Exit condition for fist and emergency reset
- **No Action**: Does nothing, just resets gesture system

---

## 🎯 Gesture State Machine (CRITICAL)

```
        IDLE (Open Palm)
             │
      ┌──────┴──────┐
      │             │
      ▼             ▼
   FIST          PINCH
  (LOCKED)      (LOCKED)
      │             │
      │  Open hand  │  Separate fingers
      └──────┬──────┘
             │
             ▼
          IDLE
```

### Rules (NON-NEGOTIABLE)
1. **Once LOCKED, STAY LOCKED** - Don't check for other gestures
2. **Only ONE gesture active at a time** - Never fist AND pinch simultaneously
3. **Clear exit conditions** - Must FULLY open hand or separate fingers to exit
4. **Priority when idle**: Check FIST first, then PINCH
5. **Confidence threshold**: If hand lost briefly (< 200ms), stay in current mode

---

## 📊 Configuration Values (Tuned for Production)

```typescript
export const PRODUCTION_GESTURE_CONFIG = {
  // Hand Tracking
  maxHands: 1,
  modelComplexity: 0,  // LITE model
  minDetectionConfidence: 0.7,  // Don't lose hands easily
  minTrackingConfidence: 0.7,   // Keep tracking stable
  cameraResolution: { width: 320, height: 240 },  // 25-30 FPS

  // Gesture Thresholds
  fistThreshold: 0.18,     // Fingers to wrist distance for fist
  fistExitThreshold: 0.30,  // Must open THIS much to exit fist
  pinchThreshold: 0.10,    // Thumb-index distance for pinch
  pinchExitThreshold: 0.15, // Must separate THIS much to exit pinch

  // Movement Detection
  minimumMovementDelta: 0.01,  // Minimum hand movement to register
  smoothingFactor: 0.75,       // Position smoothing (0-1)

  // Update Rates
  gestureUpdateInterval: 16,   // 60 FPS (ms)
  cameraUpdateInterval: 16,    // 60 FPS (ms)

  // Sensitivity
  fistRotationSensitivity: 8.0,   // Multiplier for fist drag
  pinchZoomSensitivity: 150.0,    // Multiplier for pinch zoom

  // Clap Detection
  clapThreshold: 180,           // Audio level for clap
  doubleClap Timeout: 300,       // Max time between claps (ms)
  clapCooldown: 2000,           // Prevent rapid re-trigger (ms)
};
```

---

## 🐛 Edge Cases & Error Handling

### Hand Tracking Failures
| Scenario | Behavior |
|----------|----------|
| Hand temporarily lost (< 200ms) | Stay in current gesture mode |
| Hand lost > 200ms | Exit gesture mode, return to idle |
| Camera blocked | Show warning, disable gestures |
| Multiple hands detected | Use only first hand |
| No hands detected in idle | No action, keep waiting |

### Gesture Ambiguity
| Scenario | Resolution |
|----------|-----------|
| Fist looks like pinch | Priority: Check fist first |
| Pinch while in fist mode | Ignore pinch, stay in fist |
| Fist while in pinch mode | Ignore fist, stay in pinch |
| Hand between gestures | Stay in current mode (hysteresis) |

### Performance Issues
| Scenario | Behavior |
|----------|----------|
| FPS drops < 15 | Lower camera resolution automatically |
| FPS drops < 10 | Show performance warning |
| Browser tab hidden | Pause gesture detection |
| Low battery (if detectable) | Reduce update rate |

---

## ✅ Acceptance Criteria

### Functional
- [ ] Fist drag rotates camera smoothly (no jumps/resets)
- [ ] Pinch zoom: toward camera = bigger, away = smaller
- [ ] Clap spreads all modules in grid
- [ ] Gestures lock and don't flicker
- [ ] Open hand exits any gesture

### Performance
- [ ] 25+ FPS hand tracking
- [ ] < 50ms latency gesture → camera movement
- [ ] < 5% hand detection loss rate

### UX
- [ ] Display shows current gesture clearly
- [ ] Hand landmarks visible on camera feed
- [ ] Smooth, mouse-like control feel
- [ ] No unexpected behavior

---

## 🚀 Implementation Checklist

### Phase 1: Fix Existing Gestures
- [ ] Fix fist drag (cumulative, smooth rotation)
- [ ] Fix pinch zoom direction (toward = IN, away = OUT)
- [ ] Test gesture locking stability

### Phase 2: Add Clap Gesture
- [ ] Integrate audio gesture detector
- [ ] Add clap event handling
- [ ] Implement spreadAllModules() action
- [ ] Test double clap detection

### Phase 3: UI Improvements
- [ ] Remove duplicate FPS counter
- [ ] Add hand landmark visualization (21 points + connections)
- [ ] Show clear gesture names
- [ ] Add gesture confidence meter

### Phase 4: Polish
- [ ] Performance auto-tuning
- [ ] Error recovery
- [ ] Edge case handling
- [ ] User testing & iteration

---

**This is the FINAL spec. Implement EXACTLY as written.**
**No deviations. No "improvements". Just make it work as specified.**

---

**Date**: January 14, 2026
**Version**: 3.0 FINAL
**Status**: AUTHORITATIVE
