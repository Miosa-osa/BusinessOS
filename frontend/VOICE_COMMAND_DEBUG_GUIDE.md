# 🔍 Voice Command Debug Guide

**Status**: All logging added - ready for testing
**Date**: January 14, 2026

---

## ✅ What Was Fixed

### 1. **Critical Bug Fix: Nested Store Call**
- **Location**: `desktop3dStore.ts` line 430 (old code)
- **Problem**: When a window was already open, `openWindow()` called `focusWindow()` from within an `update()`, causing race conditions
- **Fix**: Now manipulates state directly instead of calling nested store method

### 2. **Comprehensive Logging Added**
- `desktop3dStore.openWindow()` - Full execution path logging
- `desktop3dStore.focusWindow()` - Target window and state change logging
- `Desktop3D.svelte executeCommandAction()` - Voice command execution logging

---

## 🎯 Testing "Open Terminal" Command

### Step 1: Open Browser Console (F12)

### Step 2: Say "Open Terminal"

### Step 3: Expected Console Output

You should see this EXACT flow:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Voice] 🎤 HEARD: open terminal
[Voice] 🧠 PARSED: { "type": "focus_module", "module": "terminal" }
[Voice] 🔊 SPEAKING ACK: Opening terminal
[Voice] ⚙️ EXECUTING: focus_module

[Voice] 📱 focus_module command for module: "terminal"
[Voice] Window search result: { found: false, windowId: undefined, totalOpenWindows: 3 }
[Voice] → Opening NEW window for module: "terminal"

[Desktop3D Store] 🔵 openWindow() called for module: "terminal"
[Desktop3D Store] Existing window check: { found: false, isOpen: undefined, windowId: undefined }
[Desktop3D Store] ✨ Creating NEW window for "terminal" at position: [x, y, z]
[Desktop3D Store] New window created: { id: "window-terminal-1736863200000", module: "terminal", title: "Terminal" }
[Desktop3D Store] Recalculating positions after opening "terminal"
[Desktop3D Store] ✅ openWindow() complete for "terminal"

[Voice] ✅ focus_module execution complete
[Voice] ✅ SUCCESS: focus_module
```

### Step 4: Visual Confirmation

After the logs, you should see:
1. ✅ Terminal window appears in 3D space
2. ✅ Camera focuses on terminal (enters 'focused' viewMode)
3. ✅ Terminal is fully interactive

---

## 🔁 Testing "Open Terminal" AGAIN (Window Already Open)

### Say "Open Terminal" a second time

Expected output:

```
[Voice] 🎤 HEARD: open terminal
[Voice] 🧠 PARSED: { "type": "focus_module", "module": "terminal" }
[Voice] 🔊 SPEAKING ACK: Opening terminal
[Voice] ⚙️ EXECUTING: focus_module

[Voice] 📱 focus_module command for module: "terminal"
[Voice] Window search result: { found: true, windowId: "window-terminal-1736863200000", totalOpenWindows: 4 }
[Voice] → Focusing existing window (id: window-terminal-1736863200000)

[Desktop3D Store] 🎯 focusWindow() called for windowId: "window-terminal-1736863200000"
[Desktop3D Store] Target window: { found: true, module: "terminal", isOpen: true }
[Desktop3D Store] ✅ Focused window "window-terminal-1736863200000", switching to 'focused' viewMode

[Voice] ✅ focus_module execution complete
[Voice] ✅ SUCCESS: focus_module
```

Visual result: Camera focuses on existing terminal (no duplicate created)

---

## 🚨 Troubleshooting

### Problem: NO Logs Appear After "🎤 HEARD:"

**Diagnosis**: Parser is routing to AI conversation instead of command execution

**Check**:
```
[Voice] 💬 ROUTING TO AI: unknown
```

**Solution**:
1. Check `voiceCommands.ts` parser patterns
2. Verify module name matches exactly (e.g., "terminal" not "terminals")
3. Try exact phrase: "open terminal" (under 5 words, direct)

### Problem: Logs Show Execution But Window Doesn't Appear

**Diagnosis**: Store methods are executing but UI isn't updating

**Check**:
```
[Desktop3D Store] ✅ openWindow() complete for "terminal"
```

But no visual change?

**Possible Causes**:
1. Window exists but is off-screen (check position values)
2. Module not registered in MODULE_INFO
3. 3D renderer not picking up state change

**Solution**:
```bash
# Check MODULE_INFO in desktop3dStore.ts
grep -A 2 "terminal:" src/lib/stores/desktop3dStore.ts
```

### Problem: Error in Console

**Check Error Message**:

```
[Voice] ❌ FAILED: focus_module Error: ...
```

**Common Errors**:
1. `Cannot read properties of undefined (reading 'module')` → MODULE_INFO missing entry
2. `Cannot find module` → Module name typo
3. `focusWindow is not a function` → Store method issue

---

## 🧪 Test All Commands Systematically

### Camera Control
```
Say: "zoom in"
Expected: [Desktop3D Store] Adjusting sphere radius: 120 → 118

Say: "zoom out"
Expected: [Desktop3D Store] Adjusting sphere radius: 118 → 120

Say: "expand"
Expected: [Desktop3D Store] Adjusting sphere radius: 120 → 123

Say: "contract"
Expected: [Desktop3D Store] Adjusting sphere radius: 123 → 120
```

### Module Control
```
Say: "open chat"
Expected: [Desktop3D Store] openWindow() → chat window appears

Say: "open dashboard"
Expected: [Desktop3D Store] openWindow() → dashboard appears

Say: "close chat"
Expected: [Voice] ❌ close_module command for module: "chat"
Expected: [Voice] → Closing window (id: ...)
```

### Navigation
```
Say: "next window"
Expected: Focus shifts to next window in rotation

Say: "previous window"
Expected: Focus shifts to previous window

Say: "unfocus"
Expected: Returns to orb view, all windows visible
```

---

## 📊 Logging Key

| Emoji | Meaning | Location |
|-------|---------|----------|
| 🎤 | Voice heard | Desktop3D.svelte (handleTranscript) |
| 🧠 | Parsed command | voiceCommands.ts (parse) |
| 🔊 | Speaking acknowledgment | osaVoice.ts (speak) |
| ⚙️ | Executing command | Desktop3D.svelte (executeVoiceCommand) |
| 📱 | Module command | Desktop3D.svelte (executeCommandAction) |
| 🔵 | Opening window | desktop3dStore.ts (openWindow) |
| 🎯 | Focusing window | desktop3dStore.ts (focusWindow) |
| ✨ | Creating NEW window | desktop3dStore.ts (openWindow - new path) |
| ♻️ | Reopening existing | desktop3dStore.ts (openWindow - reopen path) |
| ✅ | Success | All locations |
| ❌ | Closing window | Desktop3D.svelte (close_module) |
| ⚠️ | Warning | Various locations |

---

## 🎯 Quick Test Script

Run these commands in order and verify each works:

1. **"open terminal"** → Terminal appears
2. **"open chat"** → Chat appears
3. **"zoom out"** → Camera moves back
4. **"expand"** → Orb grows (windows spread)
5. **"contract"** → Orb shrinks (windows closer)
6. **"next window"** → Focus shifts
7. **"close terminal"** → Terminal closes
8. **"unfocus"** → Back to orb view

---

## ✅ Success Criteria

All 8 commands should:
- ✅ Produce complete log flow (no missing steps)
- ✅ Execute successfully (✅ SUCCESS logged)
- ✅ Show visual result matching command
- ✅ Complete within 1-2 seconds

---

**Ready to test! Open F12 console and start speaking commands.** 🎤
