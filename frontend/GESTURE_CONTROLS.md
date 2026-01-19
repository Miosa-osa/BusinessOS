# 🖐️ 3D Desktop Gesture Controls

## Quick Start Guide

### Performance Optimized Settings
- **MediaPipe Lite Model** - Fast, lightweight hand tracking
- **320x240 Resolution** - Optimized for speed
- **15 FPS Target** - Smooth without lag
- **Minimal Debug Overlay** - Only FPS text, no landmark drawing

---

## 🎮 Gesture Controls

### 1. 👊 FIST = Grab and Rotate Sphere

**How to use:**
1. Make a **FIST** (close all fingers)
2. **Move your fist** left/right/up/down
3. The sphere **rotates** with your hand movement
4. **Release fist** (open hand) to stop

**When it triggers:**
- Fist detected
- Hand moves at least **0.02 units**
- Cooldown: **150ms** between rotation updates

**What it does:**
- Continuous rotation based on hand movement
- Horizontal movement = rotate sphere
- Works while fist is held

---

### 2. 🤏 PINCH = Zoom In/Out

**How to use:**
1. **PINCH** (touch thumb and index finger together)
2. **Move hand forward** = Zoom IN (closer)
3. **Move hand backward** = Zoom OUT (farther)
4. **Move hand left/right** = Rotate sphere

**When it triggers:**
- Thumb and index finger closer than threshold
- Hand moves at least **0.03 units** (strict to avoid jitter)
- Cooldown: **250ms** between zoom actions

**What it does:**
- **Z-axis movement** (forward/back) = Zoom in/out
- **X-axis movement** (left/right) = Rotate left/right
- **Y-axis movement** (up/down) = Expand/contract

---

## 🔇 NEUTRAL STATE (Open Palm)

**What it does:** NOTHING

When your hand is **open** (palm showing), the system is in **neutral mode**.
- No gestures detected
- No actions triggered
- Just tracking your hand position

This prevents accidental triggers!

---

## ⚙️ Performance Optimizations

### What We Changed for Speed

1. **MediaPipe Lite Model** (`modelComplexity: 0`)
   - 3x faster than full model
   - Still accurate for basic gestures

2. **Lower Resolution** (320x240)
   - 4x fewer pixels to process
   - Cameras are usually 640x480, we halved it

3. **Frame Throttling** (15 FPS target)
   - Skips frames if processing too fast
   - ~67ms between frames

4. **Minimal Canvas Drawing**
   - ONLY draws FPS and hand count text
   - Skips ALL landmark/connection drawing
   - Canvas operations are VERY expensive!

5. **Strict Movement Thresholds**
   - FIST: 0.02 units minimum movement
   - PINCH: 0.03 units minimum movement
   - Prevents jitter/lag from triggering actions

6. **Gesture Cooldowns**
   - FIST rotation: 150ms cooldown
   - PINCH zoom: 250ms cooldown
   - Maximum 4-6 actions per second

7. **Reduced Logging**
   - Log FPS every 60 frames (once per second)
   - Reduces console spam

---

## 🐛 Troubleshooting

### "FPS is still low (5-10 FPS)"

**Possible causes:**
1. **MediaPipe WASM loading** - First load is slow, improves after
2. **Browser hardware acceleration** - Make sure it's enabled
3. **Too many browser tabs** - Close other tabs
4. **Computer performance** - MediaPipe needs decent CPU/GPU

**Try:**
- Refresh the page (WASM files get cached)
- Close other tabs/applications
- Check browser dev tools > Performance tab

### "Gestures not triggering"

**Check:**
1. Hand clearly visible in camera?
2. Good lighting?
3. Making distinct gestures (clear fist, clear pinch)?
4. Moving hand enough? (thresholds are strict)

### "Too sensitive / triggering randomly"

**Fix:**
- Increase movement thresholds in `gestureDetector.ts`
- Increase cooldown times
- Reduce smoothing factor

---

## 📊 Expected Performance

**Target FPS:** 15 FPS
**Realistic range:** 10-20 FPS

**With hand visible:**
- Lite model: 10-15 FPS ✅
- Full model: 3-8 FPS ❌ (we don't use this)

**Without hand:**
- Should be ~15 FPS (just camera feed)

---

## 🎯 Best Practices

1. **Clear, intentional movements** - Jittery tracking needs clear gestures
2. **One gesture at a time** - Don't combine fist + pinch
3. **Hold gesture steady** - Wait for cooldown before next action
4. **Good lighting** - Helps MediaPipe detect hands better
5. **Camera position** - Hands at comfortable distance from camera

---

## 🚀 Future Improvements

**If we need better performance:**
- Use Web Workers for MediaPipe processing
- Reduce camera resolution to 240x180
- Skip even more frames (10 FPS target)
- Disable canvas overlay entirely
- Use GPU acceleration if available

**If we need more gestures:**
- Two-hand gestures (pinch with both hands = special action)
- Audio gestures (clap detection - already exists!)
- Voice commands to enable/disable gestures

---

**Last updated:** January 14, 2026
**Version:** Phase 3 - Performance Optimized
