# Hand Gesture Control - Phase 2 Plan

**Status**: Planning Complete
**Implementation**: Future Phase
**Created**: January 14, 2026

---

## Overview

Integrate MediaPipe Hands to control 3D Desktop via webcam hand tracking. This will allow users to control the 3D environment with natural hand gestures, creating an immersive interaction experience.

---

## Gesture Vocabulary

| Gesture | Detection | Action | Priority |
|---------|-----------|--------|----------|
| **Pinch** (thumb + index) | Distance < 30px | Zoom in/out based on separation | HIGH |
| **Swipe Left** | Hand moves left >100px | Rotate orb left | HIGH |
| **Swipe Right** | Hand moves right >100px | Rotate orb right | HIGH |
| **Point** (index extended) | Index tip position | Focus window at cursor | MEDIUM |
| **Open Palm** (all fingers extended) | All fingertips visible | Pause gesture detection | LOW |
| **Fist** (all fingers closed) | No fingertips visible | Resume gesture detection | LOW |

---

## Implementation Steps

### 1. Install MediaPipe Hands

```bash
npm install @mediapipe/hands
```

### 2. Create Gesture Processor

Extract gestures from 21-point hand skeleton:

```typescript
interface HandLandmarks {
  landmarks: Array<{x: number, y: number, z: number}>;
  handedness: 'left' | 'right';
}

function detectGesture(landmarks: HandLandmarks): GestureType {
  // Calculate finger distances, positions, angles
  // Classify gesture type based on hand shape
  // Return gesture with confidence score
}
```

### 3. Hook into Existing Camera Stream

Use `desktop3dPermissions.cameraStream`:

```typescript
import { cameraStream } from '$lib/services/desktop3dPermissions';

// Initialize MediaPipe with camera stream
hands.initialize({
  inputStream: $cameraStream,
  onResults: handleHandResults
});
```

### 4. Map Gestures to Desktop Actions

```typescript
function handleGesture(gesture: GestureEvent) {
  switch (gesture.type) {
    case 'pinch':
      const separation = calculateFingerDistance(gesture.landmarks);
      const delta = mapToDelta(separation); // -1 to 1
      desktop3dStore.adjustSphereRadius(delta);
      break;

    case 'swipe_left':
      // Rotate camera left
      orbitControls.rotateLeft(0.1);
      break;

    case 'swipe_right':
      // Rotate camera right
      orbitControls.rotateRight(0.1);
      break;

    case 'point':
      const windowAtPosition = findWindowAtScreenPosition(
        gesture.landmarks.indexTip.x,
        gesture.landmarks.indexTip.y
      );
      if (windowAtPosition) {
        desktop3dStore.focusWindow(windowAtPosition.id);
      }
      break;
  }
}
```

### 5. Add Visual Feedback (Optional)

Show hand skeleton overlay:

```svelte
<canvas bind:this={handOverlay} class="hand-skeleton-overlay" />
```

---

## Technical Architecture

```
Camera Stream (from desktop3dPermissions)
    ↓
MediaPipe Hands (detects 21 landmarks)
    ↓
Gesture Processor (classifies gesture type)
    ↓
Gesture Event (type + confidence)
    ↓
Desktop3D Handler (maps to store actions)
    ↓
3D Desktop (responds to gesture)
```

---

## MediaPipe Hand Landmarks

MediaPipe Hands provides 21 hand landmarks:

```
 0: Wrist
 1-4: Thumb (CMC, MCP, IP, Tip)
 5-8: Index Finger (MCP, PIP, DIP, Tip)
 9-12: Middle Finger (MCP, PIP, DIP, Tip)
 13-16: Ring Finger (MCP, PIP, DIP, Tip)
 17-20: Pinky (MCP, PIP, DIP, Tip)
```

**Key Landmarks for Gestures**:
- Pinch: Distance between landmarks #4 (thumb tip) and #8 (index tip)
- Point: Angle of landmarks #5, #6, #7, #8 (index finger)
- Swipe: X-position delta of landmark #0 (wrist) over time

---

## Files to Modify (Phase 2)

### New Files
- ✅ `handGestureService.ts` - MediaPipe integration (STUB CREATED)

### Existing Files to Modify
1. **`Desktop3D.svelte`**
   - Import `handGestureService`
   - Add gesture event handlers
   - Initialize service on mount

2. **`desktop3dPermissions.ts`**
   - Already has camera access ✅
   - No changes needed

3. **`package.json`**
   - Add `@mediapipe/hands` dependency

### Example Integration in Desktop3D.svelte

```typescript
import { handGestureService } from '$lib/services/handGestureService';

onMount(() => {
  // ... existing voice initialization ...

  // Initialize hand gestures
  handGestureService.init((gesture) => {
    handleGesture(gesture);
  });
});

function handleGesture(gesture: GestureEvent) {
  console.log('[Gesture]', gesture.type, gesture.confidence);

  switch (gesture.type) {
    case 'pinch':
      // Zoom based on pinch separation
      break;
    case 'swipe_left':
    case 'swipe_right':
      // Rotate orb
      break;
    case 'point':
      // Focus window
      break;
  }
}
```

---

## Gesture Detection Algorithms

### Pinch Detection

```typescript
function detectPinch(landmarks: HandLandmarks): number | null {
  const thumbTip = landmarks[4];
  const indexTip = landmarks[8];

  const distance = Math.sqrt(
    Math.pow(thumbTip.x - indexTip.x, 2) +
    Math.pow(thumbTip.y - indexTip.y, 2)
  );

  if (distance < 0.05) { // Threshold in normalized coords
    return distance; // Return separation for zoom control
  }

  return null; // Not pinching
}
```

### Swipe Detection

```typescript
const SWIPE_THRESHOLD = 0.15; // Normalized X movement
const SWIPE_TIME_WINDOW = 500; // ms

function detectSwipe(landmarks: HandLandmarks): 'left' | 'right' | null {
  const wrist = landmarks[0];
  const currentX = wrist.x;
  const deltaX = currentX - previousX;
  const deltaTime = Date.now() - previousTime;

  if (deltaTime < SWIPE_TIME_WINDOW) {
    if (deltaX > SWIPE_THRESHOLD) return 'right';
    if (deltaX < -SWIPE_THRESHOLD) return 'left';
  }

  return null;
}
```

### Point Detection

```typescript
function detectPoint(landmarks: HandLandmarks): boolean {
  const indexMCP = landmarks[5];
  const indexTip = landmarks[8];
  const middleTip = landmarks[12];

  // Index extended, middle finger down
  const indexExtended = indexTip.y < indexMCP.y;
  const middleDown = middleTip.y > indexMCP.y;

  return indexExtended && middleDown;
}
```

---

## Performance Considerations

### Frame Rate
- MediaPipe Hands: ~30 FPS on average hardware
- Desktop 3D: 60 FPS target
- **Strategy**: Decouple gesture detection from render loop

### Debouncing
```typescript
const GESTURE_COOLDOWN = 200; // ms between gesture triggers
let lastGestureTime = 0;

function processGesture(gesture: GestureEvent) {
  const now = Date.now();
  if (now - lastGestureTime < GESTURE_COOLDOWN) return;

  lastGestureTime = now;
  handleGesture(gesture);
}
```

### Confidence Threshold
```typescript
const MIN_CONFIDENCE = 0.7; // Only process high-confidence detections

if (gesture.confidence >= MIN_CONFIDENCE) {
  processGesture(gesture);
}
```

---

## User Experience Considerations

### Gesture Tutorial
Show overlay on first use:
```
"Try these gestures:
- 🤏 Pinch to zoom
- 👈 Swipe to rotate
- 👆 Point to select"
```

### Visual Feedback
- Show hand outline when detected
- Highlight gesture when recognized
- Animate transitions smoothly

### Accessibility
- Keep voice commands as primary input
- Gestures are optional enhancement
- Provide keyboard shortcuts as fallback

---

## Testing Strategy

### Unit Tests
- Gesture detection algorithms
- Landmark distance calculations
- Swipe velocity calculations

### Integration Tests
- Camera stream initialization
- MediaPipe integration
- Desktop3D action mapping

### User Testing
- Gesture recognition accuracy
- Response time measurement
- User comfort and fatigue
- Accessibility evaluation

---

## Future Enhancements

### Multi-Hand Gestures
- Two-hand pinch for advanced zoom
- Hand clapping for reset
- Two-hand swipe for faster rotation

### Custom Gestures
- User-defined gesture mapping
- Gesture recording and playback
- Per-user gesture profiles

### Voice + Gesture Combo
- "Zoom" (voice) + Pinch amount (gesture)
- "Rotate" (voice) + Swipe direction (gesture)
- Multimodal interaction patterns

---

## Dependencies

```json
{
  "@mediapipe/hands": "^0.4.1646424915",
  "@mediapipe/camera_utils": "^0.3.1640029074",
  "@mediapipe/drawing_utils": "^0.3.1620248257"
}
```

---

## References

- [MediaPipe Hands Documentation](https://google.github.io/mediapipe/solutions/hands.html)
- [Hand Landmark Model](https://google.github.io/mediapipe/solutions/hands#hand-landmark-model)
- [Gesture Recognition Best Practices](https://developers.google.com/mediapipe/solutions/vision/gesture_recognizer)

---

## Summary

**Phase 1 (Complete)**: Voice system with command parsing and OSA speech
**Phase 2 (Planned)**: Hand gesture control using MediaPipe Hands
**Phase 3 (Future)**: Multimodal interaction (voice + gesture + keyboard)

The hand gesture system will complement the existing voice control, providing users with natural, intuitive ways to interact with the 3D Desktop.

---

**Ready for implementation when Phase 2 begins!** 🚀
