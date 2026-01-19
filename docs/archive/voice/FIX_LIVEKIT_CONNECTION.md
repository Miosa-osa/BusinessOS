# Fix LiveKit Connection Error

**Error You're Seeing**:
```
WebSocket connection to 'wss://macstudiosystems-yn61tekm.livekit.cloud/...' failed
ConnectionError: Client initiated disconnect
```

---

## Problem

Frontend is trying to connect to LiveKit Cloud, but the Pure Go voice agent isn't there to receive the connection.

---

## Solution: Check Backend is Running

### Step 1: Is Backend Running?

```bash
# Check if backend is running
ps aux | grep "cmd/server"
```

If NO output → Backend is NOT running

**Start it**:
```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run ./cmd/server
```

### Step 2: Watch Backend Logs

When you start the backend, you MUST see these lines:

```
✅ Pure Go Voice Agent started successfully
🎙️ Pure Go LiveKit Voice Agent starting
   livekit_url: wss://macstudiosystems-yn61tekm.livekit.cloud
   architecture: Direct WebRTC (no Python/gRPC)
```

**If you DON'T see these lines** → Voice agent didn't start

---

## Check Environment Variables

The backend needs these to connect to LiveKit:

```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
cat .env | grep LIVEKIT
```

**Should show**:
```
LIVEKIT_URL=wss://macstudiosystems-yn61tekm.livekit.cloud
LIVEKIT_API_KEY=APIc...
LIVEKIT_API_SECRET=...
```

**If missing** → Add them to `.env` file

---

## Test Backend Separately

### 1. Start Backend with Logs

```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run ./cmd/server 2>&1 | grep -i "voice\|livekit"
```

You should see:
```
gRPC Voice Server starting on port 50051
✅ Pure Go Voice Agent started successfully
🎙️ Pure Go LiveKit Voice Agent starting
```

### 2. Check Voice Health

```bash
curl http://localhost:8080/api/voice/health | jq
```

**Expected**:
```json
{
  "status": "healthy",
  "mode": "pure-go",
  "active_sessions": 0,
  "livekit_url": "wss://macstudiosystems-yn61tekm.livekit.cloud"
}
```

**If you get 404** → Voice endpoint not registered

---

## Common Issues

### Issue 1: VOICE_AGENT_MODE not set

**Symptom**: Backend starts but no voice logs

**Check**:
```bash
grep VOICE_AGENT_MODE .env
```

**Fix**: Add to `.env`:
```
VOICE_AGENT_MODE=pure-go
```

Restart backend.

---

### Issue 2: LiveKit credentials wrong

**Symptom**: Backend logs show "Failed to connect to LiveKit"

**Fix**: Check your LiveKit credentials at https://cloud.livekit.io

```bash
# .env should have:
LIVEKIT_URL=wss://your-project.livekit.cloud
LIVEKIT_API_KEY=API...
LIVEKIT_API_SECRET=...
```

---

### Issue 3: Backend crashes on startup

**Symptom**: `go run ./cmd/server` exits immediately

**Check**: Look for error in logs:
```bash
go run ./cmd/server 2>&1 | head -50
```

Common errors:
- "Database connection failed" → Fix `DATABASE_URL`
- "Whisper not found" → Run `brew install whisper-cpp`
- "Port already in use" → Kill existing server: `pkill -f "cmd/server"`

---

## Full Debug Procedure

Run this to see EVERYTHING:

```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go

# Set debug logging
export LOG_LEVEL=debug

# Run backend and filter voice logs
go run ./cmd/server 2>&1 | grep -E "voice|livekit|Voice|LiveKit" &

# In another terminal, test health
sleep 5
curl http://localhost:8080/api/voice/health
```

---

## Expected Flow (When Working)

1. **Backend starts**:
   ```
   gRPC Voice Server starting on port 50051
   Pure Go Voice Agent started successfully
   ```

2. **Frontend clicks mic**:
   ```
   [Frontend] Connecting to LiveKit room: osa-voice-...
   [Backend] New participant joined: user-...
   [Backend] Audio track subscribed
   ```

3. **User speaks**:
   ```
   [Backend] Processing audio frame
   [Backend] VAD: Speech detected
   [Backend] STT: "Hello"
   ```

---

## Quick Fix Script

Save as `fix-voice-connection.sh`:

```bash
#!/bin/bash
echo "🔧 Fixing Voice Connection..."

# 1. Check if backend running
if pgrep -f "cmd/server" > /dev/null; then
    echo "✅ Backend is running"
else
    echo "❌ Backend NOT running"
    echo "   Starting it now..."
    cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
    go run ./cmd/server &
    sleep 5
fi

# 2. Check voice health
echo ""
echo "🏥 Checking voice health..."
HEALTH=$(curl -s http://localhost:8080/api/voice/health)
if echo "$HEALTH" | grep -q "healthy"; then
    echo "✅ Voice system healthy"
    echo "$HEALTH" | jq
else
    echo "❌ Voice system NOT healthy"
    echo "Response: $HEALTH"
fi

# 3. Check LiveKit vars
echo ""
echo "🔑 Checking LiveKit credentials..."
if [ -f "/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/.env" ]; then
    cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
    [ -n "$(grep LIVEKIT_URL .env)" ] && echo "  ✅ LIVEKIT_URL set" || echo "  ❌ LIVEKIT_URL missing"
    [ -n "$(grep LIVEKIT_API_KEY .env)" ] && echo "  ✅ LIVEKIT_API_KEY set" || echo "  ❌ LIVEKIT_API_KEY missing"
    [ -n "$(grep LIVEKIT_API_SECRET .env)" ] && echo "  ✅ LIVEKIT_API_SECRET set" || echo "  ❌ LIVEKIT_API_SECRET missing"
else
    echo "  ❌ .env file not found"
fi

echo ""
echo "If all ✅, try clicking mic button again!"
```

Run it:
```bash
chmod +x fix-voice-connection.sh
./fix-voice-connection.sh
```

---

## After Fixing

1. Restart backend
2. Refresh frontend browser page (F5)
3. Click mic button
4. Should connect without error

---

**TL;DR**:
1. Make sure backend is running: `go run ./cmd/server`
2. Check you see "Pure Go Voice Agent started" in logs
3. Verify LiveKit credentials in `.env`
4. Try mic button again
