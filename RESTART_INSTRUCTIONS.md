# 🔄 RESTART EVERYTHING - STEP BY STEP

**Follow these steps EXACTLY in order:**

---

## Step 1: Stop Frontend Dev Server

In the terminal running your frontend:
```bash
# Press Ctrl+C to stop the dev server
# Wait until you see the terminal prompt return
```

**Verify:** Terminal shows normal prompt (not "VITE" or "Local: http://...")

---

## Step 2: Stop Backend Server

In the terminal running your backend:
```bash
# Press Ctrl+C to stop the server
# Wait until you see the terminal prompt return
```

**Verify:** Terminal shows normal prompt (not server logs)

---

## Step 3: Clear Browser Cache COMPLETELY

### Chrome/Edge:
1. Press `Cmd+Shift+Delete` (Mac) or `Ctrl+Shift+Delete` (Windows)
2. In the popup:
   - Time range: **"All time"**
   - Check: **"Cached images and files"**
   - Check: **"Cookies and other site data"** (optional but recommended)
3. Click **"Clear data"**

### Safari:
1. Press `Cmd+Option+E` (Empty Caches)
2. Or: Safari menu → Settings → Privacy → Manage Website Data → Remove All

**Verify:** Popup closes, cache cleared message appears

---

## Step 4: Restart Backend

```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run ./cmd/server
```

**Wait for these logs:**
```
[Server] Starting BusinessOS Backend...
[Database] Connected to PostgreSQL
[Server] Server listening on :8001
```

**Verify:** No errors, server running on port 8001

---

## Step 5: Restart Frontend

```bash
cd /Users/rhl/Desktop/BusinessOS2/frontend
npm run dev
```

**Wait for these logs:**
```
VITE v5.x.x  ready in XXX ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
```

**Verify:** Vite dev server running on port 5173

---

## Step 6: Hard Refresh Browser

1. Navigate to http://localhost:5173
2. Press `Cmd+Shift+R` (Mac) or `Ctrl+Shift+F5` (Windows)
3. Watch browser console (F12)

**Expected logs:**
```
[LiveKit] 🆕 Instance 1 created
[LiveKit #1] State: disconnected -> connecting
[LiveKit #1] Getting token from backend...
[LiveKit #1] Token received for room: osa-voice-XXXXX
[LiveKit #1] Connected to room: osa-voice-XXXXX
```

---

## ✅ VERIFICATION CHECKLIST

After restart, verify:

### 1. Only ONE Instance
```
[LiveKit] 🆕 Instance 1 created    ← Should see THIS
```

**❌ BAD (if you see):**
```
[LiveKit] 🆕 Instance 2 created    ← Old instance still there
```

### 2. No Duplicate Logs
Speak to OSA, you should see:
```
[LiveKit #1] 🎙️  Audio collected: 45320 bytes in 2800ms
[LiveKit #1] 📤 Sending audio to backend: 45320 bytes
```

**Not:**
```
[LiveKit #1] 🎙️  Audio collected: 45320 bytes in 2800ms
[LiveKit #2] 🎙️  Audio collected: 45320 bytes in 2800ms  ← DUPLICATE
```

### 3. No Hallucination
Say something vague like "listeners" or "metrics".

**✅ GOOD:**
- "What do you mean by listeners?"
- "I don't have that info. What are you looking for?"

**❌ BAD:**
- "You've got 850 listeners..." (making up data)

### 4. Clean Console (No Spam)
You should NOT see:
- ❌ "Skipping recording cycle - audio is playing" (removed)
- ❌ "Skipping recording cycle - post-playback cooldown" (removed)
- ❌ "Discarding audio - playback occurred during recording" (removed)

You SHOULD see:
- ✅ "🎙️  Audio collected: X bytes in Xms"
- ✅ "⚡ PERFORMANCE: Network=Xms | Total=Xms"
- ✅ "Transcription result: {...}"

---

## 🐛 IF ISSUES PERSIST

### Still see duplicate instances?

**Nuclear option:**
```bash
# Stop everything
# Ctrl+C in both frontend and backend terminals

# Clear all Vite cache
rm -rf /Users/rhl/Desktop/BusinessOS2/frontend/node_modules/.vite
rm -rf /Users/rhl/Desktop/BusinessOS2/frontend/.svelte-kit

# Reinstall
cd /Users/rhl/Desktop/BusinessOS2/frontend
npm install

# Restart backend
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run ./cmd/server

# Restart frontend
cd /Users/rhl/Desktop/BusinessOS2/frontend
npm run dev

# Clear browser cache AGAIN (Cmd+Shift+Delete)
# Hard refresh (Cmd+Shift+R)
```

### OSA still hallucinating?

Check backend logs on startup:
```
[Prompts] Voice prompt loaded: 752 characters
```

If not there, backend didn't reload. Make sure:
1. You actually stopped the old backend (Ctrl+C)
2. You see NEW startup logs (not old ones scrolling)
3. Port 8001 isn't occupied by ghost process

**Check for ghost processes:**
```bash
lsof -i :8001
# If anything shows up, kill it:
kill -9 <PID>
```

---

**After restart is verified, come back and I'll implement proper LiveKit streaming.**
