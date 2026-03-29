/**
 * Desktop 3D Camera Store
 * Manages camera distance and gesture-based rotation state.
 * Completely independent — zero dependencies on other desktop3d state.
 *
 * Uses Svelte 5 runes ($state) for fine-grained reactivity.
 */

const DEFAULT_CAMERA_DISTANCE = 400;

function createCameraStore() {
  let cameraDistance = $state(DEFAULT_CAMERA_DISTANCE);
  let cameraRotationDelta = $state<{ x: number; y: number }>({ x: 0, y: 0 });
  let gestureDragging = $state(false);

  return {
    get cameraDistance() {
      return cameraDistance;
    },
    get cameraRotationDelta() {
      return cameraRotationDelta;
    },
    get gestureDragging() {
      return gestureDragging;
    },

    // Adjust camera distance (TRUE zoom — moves camera closer/farther)
    // Range: 200 (very close) to 800 (very far)
    adjustCameraDistance(delta: number) {
      const newDistance = Math.max(
        200,
        Math.min(800, cameraDistance + delta * 20),
      );

      if (newDistance === cameraDistance) {
        console.log("[Camera Store] Camera distance at limit:", newDistance);
        return;
      }

      // Log only significant changes (> 5 units)
      if (Math.abs(delta) > 5) {
        console.log(
          `[Camera Store] Camera distance: ${cameraDistance.toFixed(1)} → ${newDistance.toFixed(1)}`,
        );
      }

      cameraDistance = newDistance;
    },

    // Reset camera distance to default (400)
    resetCameraDistance() {
      console.log("[Camera Store] Resetting camera distance to default (400)");
      cameraDistance = DEFAULT_CAMERA_DISTANCE;
    },

    // Update gesture-based rotation delta.
    // Send { x: 0, y: 0 } when the hand stops moving to clear the delta.
    adjustRotationSpeed(deltaX: number, deltaY: number = 0) {
      cameraRotationDelta = { x: deltaX, y: deltaY };
      gestureDragging = deltaX !== 0 || deltaY !== 0;
    },

    // Reset to initial camera state
    reset() {
      cameraDistance = DEFAULT_CAMERA_DISTANCE;
      cameraRotationDelta = { x: 0, y: 0 };
      gestureDragging = false;
    },
  };
}

export const cameraStore = createCameraStore();
