# Voice System Startup Guide - No Duplicates

## 🚨 CRITICAL: Prevent Duplicate Agents

The voice system will create duplicate agents if you:
- Run `python3 agent.py dev` multiple times
- Don't kill old agents before starting new ones
- Start the agent in multiple terminals

**Solution**: Use the startup script which auto-kills duplicates.

---

## 🚀 Quick Start (Use This Every Time)

### Option 1: Automated Script (Recommended)
```bash
cd /Users/rhl/Desktop/BusinessOS2

# Start the voice agent (kills duplicates automatically)
./start-voice-agent.sh
```

This script:
- ✅ Kills any existing voice agents
- ✅ Starts ONE fresh agent
- ✅ Shows console logs in your terminal
- ✅ Shuts down when you disconnect

### Option 2: Manual Start
```bash
# 1. Kill any running agents
pkill -9 -f "agent.py dev"

# 2. Start ONE agent
cd /Users/rhl/Desktop/BusinessOS2/python-voice-agent
python3 agent.py dev
```

---

## 🔴 How to Turn Voice On/Off

### In the Frontend:
1. **Turn ON**: Click the voice orb (cloud icon)
   - Creates LiveKit room
   - Agent auto-joins
   - Start speaking

2. **Turn OFF**: Click the voice orb again
   - Disconnects from LiveKit
   - Agent auto-shuts down
   - Room cleaned up

### What Happens Behind the Scenes:

**When you click ON:**
```
Frontend → Creates LiveKit room → Agent joins automatically
```

**When you click OFF:**
```
Frontend → Disconnects from LiveKit → Agent detects disconnect → Agent shuts down
```

---

## ✅ Verify No Duplicates

Run this to check:
```bash
ps aux | grep "agent.py" | grep -v grep
```

**Good** (1 agent):
```
rhl  12345  0.0  0.1  Python agent.py dev
```

**Bad** (2+ agents):
```
rhl  12345  0.0  0.1  Python agent.py dev
rhl  12346  0.0  0.1  Python agent.py dev  ← DUPLICATE!
```

If you see duplicates:
```bash
pkill -9 -f "agent.py dev"
./start-voice-agent.sh
```

---

## 📊 What You'll See in Console

### When Agent Starts:
```
[Agent] VAD model preloaded
[Agent] Starting for room: osa-voice-abc123
[Agent] User connected: Roberto
[Agent] Voice session started - waiting for speech
[Agent] Will auto-shutdown when user disconnects
```

### During Conversation:
```
================================================================================
🎤 USER [11:30:15]: Hello
================================================================================

================================================================================
🤖 OSA [11:30:16]: Hey! How can I help?
================================================================================
```

### When You Disconnect:
```
[Agent] User Roberto disconnected - shutting down agent
```

---

## 🎯 Testing Checklist

1. ✅ Kill all existing agents
   ```bash
   pkill -9 -f "agent.py dev"
   ```

2. ✅ Start ONE agent
   ```bash
   ./start-voice-agent.sh
   ```

3. ✅ Verify only ONE running
   ```bash
   ps aux | grep "agent.py" | grep -v grep | wc -l
   # Should show: 1
   ```

4. ✅ Test in browser
   - Open http://localhost:5173
   - Click 3D Desktop
   - Click voice orb
   - Say "Hello"
   - See captions appear (blue for you, purple for OSA)
   - Hear response

5. ✅ Turn OFF voice
   - Click voice orb again
   - Agent should shutdown in console

6. ✅ Turn ON again
   - Click voice orb
   - New agent joins
   - Speak again

---

## 🐛 Troubleshooting

### Problem: "Two agents joined the room!"

**Cause**: Multiple agents running
**Fix**:
```bash
pkill -9 -f "agent.py dev"
./start-voice-agent.sh
```

### Problem: "Agent doesn't shut down when I disconnect"

**Cause**: Old agent version without disconnect handler
**Fix**: Restart agent with new code
```bash
pkill -9 -f "agent.py dev"
./start-voice-agent.sh
```

### Problem: "Can't see transcripts in frontend"

**Cause**: Agent not sending data channel messages
**Fix**: Make sure you're using the updated agent.py (lines 69-77, 88-90, 99-101 should have `publish_transcript` calls)

### Problem: "Multiple agents spawn when I refresh page"

**Cause**: Frontend creating new rooms but agents not shutting down
**Fix**: This should NOT happen with the new agent code. If it does:
```bash
pkill -9 -f "agent.py dev"
./start-voice-agent.sh
```

---

## 📝 File Versions (Ensure You Have These)

### agent.py (Updated)
- Line 5: "Auto-disconnects when user leaves (prevents duplicates)"
- Line 13: `from livekit import rtc`
- Lines 39-46: Disconnect handler
- Lines 69-77, 88-90, 99-101: Transcript publishing

### start-voice-agent.sh (New)
- Kills duplicates automatically
- Starts ONE agent
- Shows console logs

---

## 🎉 Success Criteria

- ✅ Only ONE agent in `ps aux` output
- ✅ Agent joins when you click orb
- ✅ You see transcripts (blue/purple captions)
- ✅ You hear voice responses
- ✅ Agent shuts down when you click orb off
- ✅ Agent joins again when you click orb on

**No duplicates. Clean startup. Auto-shutdown. Perfect.**
