# 🔍 Voice System Audit - Debug Guide

**Created**: January 14, 2026
**Status**: Ready for testing

---

## ✅ What Was Just Fixed

### 1. **Zoom Direction** - FIXED ✅
- **Before**: "zoom out" zoomed IN (backwards!)
- **After**: "zoom out" zooms OUT correctly
- **Amount**: Increased from 5 units to 20 units (4x stronger)

### 2. **Comprehensive Logging** - ADDED ✅
All voice commands now show detailed debug info in console.

---

## 🎤 Testing Right Now

### Open Browser Console (F12)

You'll see this when you speak:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Voice] 🎤 HEARD: "zoom out"
[Parser] Original: zoom out
[Parser] Normalized: zoom out
[Parser] ✅ Matched: VIEW
[Voice] 🧠 PARSED: {
  "type": "zoom_out"
}
[Voice] 🔊 SPEAKING ACK: "Zooming out"
[Voice] ⚙️ EXECUTING: zoom_out
[Desktop3D Store] Adjusting sphere radius: 120 → 100
[Voice] ✅ SUCCESS: zoom_out
```

---

## 🐛 Common Issues & Solutions

### Issue 1: "Switch from Terminal to Chat" Doesn't Work

**What You Said**: "switch from terminal to chat"
**Word Count**: 5 words
**What Happens**: Might route to AI conversation

**Console Shows**:
```
[Parser] ❌ No match - Word count: 5, Question: false, Conversational: false
[Parser] → Routing to CONVERSATION
```

**Solution**:
- Say: **"open chat"** (2 words) ✅
- Or: **"close terminal"** then **"open chat"**
- Or: **"next window"** to cycle through

**Why**: Commands over 5 words → conversation mode

---

### Issue 2: Commands Going to AI

**Example**: "can you zoom out for me"
**Problem**: Detected as conversational due to "can you"

**Console Shows**:
```
[Parser] ❌ No match - Conversational: true
[Parser] → Routing to CONVERSATION
[Voice] 💬 ROUTING TO AI: unknown
```

**Solution**: Use direct commands without filler words:
- ❌ "can you zoom out"
- ✅ "zoom out"

---

### Issue 3: Deepgram Mishearing

**Example**: You say "OSA" but it hears "Elsa"

**Console Shows**:
```
[Voice] 🎤 HEARD: "Elsa zoom out"
```

**Fix**: Keyword boosting already enabled:
```typescript
keywords: ['OSA:2', 'BusinessOS:1.5']
```

**If still happening**: Report the exact phrase

---

## 📊 Test Each Command Category

### 1. Module Commands
```bash
Say: "open terminal"
Expected Log:
[Parser] ✅ Matched: MODULE
[Voice] ✅ SUCCESS: focus_module

Say: "close terminal"
Expected Log:
[Parser] ✅ Matched: MODULE
[Voice] ✅ SUCCESS: close_module
```

### 2. Zoom Commands
```bash
Say: "zoom out"
Expected Log:
[Parser] ✅ Matched: VIEW
[Desktop3D Store] Adjusting sphere radius: 120 → 100
[Voice] ✅ SUCCESS: zoom_out

Say: "zoom in"
Expected Log:
[Desktop3D Store] Adjusting sphere radius: 100 → 120
[Voice] ✅ SUCCESS: zoom_in
```

### 3. Navigation Commands
```bash
Say: "next window"
Expected Log:
[Parser] ✅ Matched: NAVIGATION
[Voice] ✅ SUCCESS: next_window

Say: "previous window"
Expected Log:
[Parser] ✅ Matched: NAVIGATION
[Voice] ✅ SUCCESS: previous_window
```

### 4. View Commands
```bash
Say: "unfocus"
Expected Log:
[Parser] ✅ Matched: VIEW
[Voice] ✅ SUCCESS: unfocus

Say: "switch to grid"
Expected Log:
[Parser] ✅ Matched: VIEW
[Voice] ✅ SUCCESS: switch_view
```

---

## 🔍 Understanding the Console Flow

### Step 1: What You Said
```
[Voice] 🎤 HEARD: "your words here"
```
- This is what Deepgram heard
- Check if transcription is correct

### Step 2: Normalization
```
[Parser] Original: your words here
[Parser] Normalized: normalized version
```
- Filler words removed
- Lowercased
- Cleaned up

### Step 3: Pattern Matching
```
[Parser] ✅ Matched: CATEGORY
```
- Shows which command category matched
- Categories: LAYOUT, MODULE, RESIZE, VIEW, NAVIGATION

### Step 4: Parsed Command
```
[Voice] 🧠 PARSED: {
  "type": "command_type"
}
```
- The final command structure
- Shows command type and parameters

### Step 5: Execution
```
[Voice] 🔊 SPEAKING ACK: "acknowledgment"
[Voice] ⚙️ EXECUTING: command_type
[Desktop3D Store] <action details>
[Voice] ✅ SUCCESS: command_type
```
- Quick acknowledgment spoken
- Command executes
- Store logs the action
- Success confirmed

### If It Fails:
```
[Voice] ❌ FAILED: command_type
Error: <error details>
```

---

## 🎯 Best Practices for Voice Commands

### Keep It Short (Under 5 Words)
- ✅ "zoom out"
- ✅ "open chat"
- ✅ "next window"
- ❌ "can you please zoom out for me"
- ❌ "switch me from terminal to chat"

### Be Direct
- ✅ "make wider"
- ❌ "can you make this wider"
- ✅ "unfocus"
- ❌ "go back to the orb view"

### Use Exact Phrases
Check the command list for exact patterns:
- "zoom out" ✅
- "zoom me out" ❌ (not in pattern list)

---

## 📋 All Working Commands

### Navigation
- "next" / "next window"
- "previous" / "previous window"

### Modules
- "open [module]" - terminal, chat, tasks, etc.
- "close [module]"
- "focus [module]"

### Camera
- "zoom in"
- "zoom out"
- "reset zoom"
- "toggle auto rotate"

### Window
- "unfocus" / "back to orb"
- "make wider" / "make narrower"
- "make taller" / "make shorter"

### View
- "switch to grid"
- "switch to orb"

### Layout
- "manage layouts"
- "enter edit mode"
- "exit edit mode"

---

## 🚨 If Nothing Works

### Check These in Console:

1. **Is microphone active?**
   ```
   [Voice] Starting...
   [Voice] Connected
   [Voice] Ready
   ```

2. **Is Deepgram transcribing?**
   ```
   [Voice] 🎤 HEARD: <something>
   ```
   - If you don't see this, microphone not working

3. **Is parser running?**
   ```
   [Parser] Original: <text>
   ```
   - If you don't see this, parser not initialized

4. **Are commands executing?**
   ```
   [Voice] ⚙️ EXECUTING: <type>
   ```
   - If you don't see this, execution blocked

5. **Any errors?**
   ```
   [Voice] ❌ FAILED: <type>
   Error: <details>
   ```
   - Shows why it failed

---

## 📈 What Good Logs Look Like

### Successful Command:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Voice] 🎤 HEARD: zoom out
[Parser] Original: zoom out
[Parser] Normalized: zoom out
[Parser] ✅ Matched: VIEW
[Voice] 🧠 PARSED: {
  "type": "zoom_out"
}
[Voice] 🔊 SPEAKING ACK: Pulling back
[Voice] ⚙️ EXECUTING: zoom_out
[Desktop3D Store] Adjusting sphere radius: 120 → 100
[Voice] ✅ SUCCESS: zoom_out
```

### Conversational (Routed to AI):
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Voice] 🎤 HEARD: can you tell me about the desktop
[Parser] Original: can you tell me about the desktop
[Parser] Normalized: tell about desktop
[Parser] ❌ No match - Word count: 7, Conversational: true
[Parser] → Routing to CONVERSATION
[Voice] 🧠 PARSED: {
  "type": "unknown",
  "text": "can you tell me about the desktop"
}
[Voice] 💬 ROUTING TO AI: unknown
```

---

## ✅ Testing Checklist

Test each category and check console:

- [ ] Open/close module commands
- [ ] Zoom in/out (verify direction is correct!)
- [ ] Next/previous window navigation
- [ ] Unfocus command
- [ ] Window resize (wider/taller)
- [ ] View switching (grid/orb)
- [ ] Layout manager
- [ ] Conversational phrases (should go to AI)

---

**All logging is now ACTIVE. Open console (F12) and start speaking!** 🎤
