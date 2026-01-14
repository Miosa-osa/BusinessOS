# 🎤 Voice System - Complete Testing Checklist

## ✅ System Status

- **Frontend:** Running on http://localhost:5174
- **Backend:** Running on http://localhost:8001
- **Deepgram API Key:** Configured (50 characters)
- **Enhanced Logging:** Active (very detailed console output)

---

## 🧪 Test Procedure - Follow These Steps EXACTLY

### Step 1: Open the Application
1. Go to: **http://localhost:5174**
2. Click "Enter 3D Desktop" or navigate to the 3D view

### Step 2: Open Browser Console
1. Press **F12** (or Cmd+Option+I on Mac)
2. Click the **Console** tab
3. Click the **Clear console** button (🚫 icon)

### Step 3: Click the Cloud Button
**Location:** Bottom-right corner (fluffy white/gray cloud)

**Expected Visual:**
- Cloud turns **blue** (listening state)
- Status label changes to "Listening..."
- Particles start floating around the cloud
- Audio bars appear at the bottom

**Expected Console Output:**
```
[Desktop3D] Starting voice commands...
[Desktop3D] ✅ Microphone permission already granted
[ActiveListening] 🚀 startListening() called
[ActiveListening] 🎤 Getting microphone stream from permissions service...
[ActiveListening] ✅ Microphone stream retrieved: {id: "...", active: true, tracks: 1, audioTracks: 1}
[ActiveListening] ✅ Audio track valid: {label: "...", enabled: true, muted: false, readyState: "live"}
[ActiveListening] 🔑 Checking for Deepgram API key...
[ActiveListening] ✅ Deepgram API key found (length: 50)
[ActiveListening] 🔑 Initializing Deepgram client...
[ActiveListening] 🌐 Connecting to Deepgram WebSocket...
[ActiveListening] ✅ Deepgram WebSocket connected!
[ActiveListening] 🎤 Ready to receive audio - speak now!
[ActiveListening] 📹 Starting MediaRecorder with 250ms chunks...
[ActiveListening] 🎙️ MediaRecorder started - capturing audio now
[ActiveListening] ✅ Started listening with Deepgram
[Desktop3D] ✅ Voice commands started - speak now!
```

**❌ If You See This Instead:**
- `❌ No microphone stream available` → Microphone permission not granted
- `❌ VITE_DEEPGRAM_API_KEY not found` → .env file not loaded
- `⚠️ Deepgram WebSocket taking longer than expected` → Network issue
- Nothing after "Starting voice commands..." → Silent failure (see troubleshooting below)

### Step 4: Speak Into Microphone
**Say clearly:** "Hello OSA"

**Expected Console Output (every 250ms):**
```
[ActiveListening] 📤 Audio chunk ready: {size: 1234, type: "audio/webm;codecs=opus", connectionState: 1}
[ActiveListening] ✅ Sent to Deepgram
[ActiveListening] 📤 Audio chunk ready: {size: 1456, type: "audio/webm;codecs=opus", connectionState: 1}
[ActiveListening] ✅ Sent to Deepgram
[ActiveListening] 📤 Audio chunk ready: {size: 1389, type: "audio/webm;codecs=opus", connectionState: 1}
[ActiveListening] ✅ Sent to Deepgram
```

**Then (after you finish speaking):**
```
[ActiveListening] 📝 Transcription received: {transcript: "hello", isFinal: false, speechFinal: false, confidence: 0.95}
[ActiveListening] 📝 Transcription (interim): "hello"
[ActiveListening] 📝 Transcription received: {transcript: "hello osa", isFinal: true, speechFinal: true, confidence: 0.98}
[ActiveListening] 📝 Transcription (final): "hello osa"
[ActiveListening] ✅ Final transcript: hello osa
[Desktop3D] 📝 Transcript (final): "hello osa"
[Desktop3D] 🎯 Command detected: unknown
[Desktop3D] 💬 Conversational mode: "hello osa"
[Desktop3D] Sending message to OSA agent...
```

**Expected Visual:**
- Transcript appears in live captions (if visible)
- OSA responds with speech (cloud turns purple)

### Step 5: Test a Command
**Say clearly:** "Open chat"

**Expected Console Output:**
```
[ActiveListening] 📝 Transcription received: {transcript: "open chat", isFinal: true, speechFinal: true}
[ActiveListening] ✅ Final transcript: open chat
[Desktop3D] 📝 Transcript (final): "open chat"
[Desktop3D] 🎯 Command detected: focus_module
[Desktop3D] 🔊 OSA acknowledging: "On it sir"
[OSA Voice] Requesting TTS {text_length: 9}
[OSA Voice] Audio received {size_bytes: 12345}
[OSA Voice] ▶️ Playing audio
[Desktop3D] ✅ Command executed: focus_module
```

**Expected Visual:**
- Cloud turns purple (OSA speaking)
- You hear OSA say "On it sir" (or variation)
- Chat module opens and focuses
- Cloud returns to blue (listening)

---

## 🐛 Troubleshooting

### Problem 1: Nothing happens after clicking cloud
**Symptoms:** Console shows only "Starting voice commands..." then stops

**Possible Causes:**
1. Microphone stream not available
2. .env file not loaded (refresh browser with Cmd+Shift+R)
3. Deepgram SDK not loaded properly

**Solution:**
```bash
# Verify .env is correct
cat /home/miosa/Desktop/BusinessOS/frontend/.env | grep DEEPGRAM

# Hard refresh browser
# Press Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows/Linux)

# Check console for error messages
# Look for red ❌ logs
```

### Problem 2: Cloud turns blue but no audio chunks sent
**Symptoms:** Console shows "MediaRecorder started" but no "📤 Audio chunk ready" logs

**Possible Causes:**
1. MediaRecorder not capturing audio
2. Microphone muted in system settings
3. Browser blocked microphone access

**Solution:**
- Check system sound settings - microphone not muted
- Check browser permissions (click lock icon in address bar)
- Try different browser (Chrome/Firefox)

### Problem 3: Audio chunks sent but no transcription
**Symptoms:** Console shows "✅ Sent to Deepgram" but no "📝 Transcription received"

**Possible Causes:**
1. Deepgram API key invalid
2. Network blocking WebSocket connection
3. Audio format not supported

**Solution:**
```bash
# Test Deepgram API key
curl -X POST "https://api.deepgram.com/v1/listen?model=nova-3" \
  -H "Authorization: Token dc01d8ab4cf41137a492e1526f4923b370d5a5ed" \
  -H "Content-Type: audio/wav" \
  --data-binary "@/dev/null"

# If you get {"err_code":"Bad Request"...} = key is valid
# If you get authentication error = key is invalid
```

**Check Console for:**
- `❌ Deepgram error:` logs (shows exact error from Deepgram)
- `⚠️ Deepgram warning:` logs (shows warnings)

### Problem 4: Transcription received but no command execution
**Symptoms:** Console shows "📝 Transcription received" but no "🎯 Command detected"

**Possible Causes:**
1. Voice command parser not working
2. JavaScript error in Desktop3D.svelte

**Solution:**
- Check console for JavaScript errors (red text)
- Look for "🎯 Command detected:" log
- Check the command type that was detected

---

## 📋 Quick Diagnostic Commands

### Check if frontend loaded .env correctly:
Open browser console and type:
```javascript
// This won't work in production, but shows if env is available
console.log(import.meta.env.VITE_DEEPGRAM_API_KEY ? 'Key loaded' : 'Key NOT loaded');
```

### Check Deepgram SDK loaded:
```javascript
console.log(typeof window !== 'undefined' ? 'Window available' : 'No window');
```

### Force stop listening:
```javascript
// If stuck in listening state
activeListeningService.stopListening();
```

---

## 🎯 Expected Full Flow

### Perfect Working Flow:

1. **Click Cloud** → Blue, "Listening..."
2. **Console:** 15+ logs showing setup (see Step 3 above)
3. **Speak:** "Open chat"
4. **Console:** Audio chunks every 250ms
5. **Console:** Transcription received: "open chat"
6. **Console:** Command detected: focus_module
7. **Console:** OSA acknowledging: "On it sir"
8. **Audio:** Hear OSA say "On it sir"
9. **Visual:** Cloud turns purple (speaking)
10. **Action:** Chat module opens
11. **Visual:** Cloud returns to blue (listening)
12. **Result:** ✅ SUCCESS

---

## 📊 Key Console Log Checkpoints

| Checkpoint | Log Message | What It Means |
|------------|-------------|---------------|
| 1 | `🚀 startListening() called` | Function started |
| 2 | `✅ Microphone stream retrieved` | Mic access OK |
| 3 | `✅ Audio track valid` | Mic is working |
| 4 | `✅ Deepgram API key found` | Key loaded |
| 5 | `✅ Deepgram WebSocket connected` | Network OK |
| 6 | `🎙️ MediaRecorder started` | Recording started |
| 7 | `📤 Audio chunk ready` | Audio captured |
| 8 | `✅ Sent to Deepgram` | Audio transmitted |
| 9 | `📝 Transcription received` | Deepgram working |
| 10 | `🎯 Command detected` | Parser working |
| 11 | `🔊 OSA acknowledging` | TTS triggered |
| 12 | `✅ Command executed` | Action completed |

**All 12 checkpoints must pass for full success!**

---

## 🚀 Next Steps After Testing

Once you confirm it works:
1. Test interrupt feature (talk over OSA)
2. Test different commands (see VOICE_SYSTEM_FIXES.md)
3. Test AI conversations ("What can you help me with?")
4. Enjoy the cloud animations!

If it doesn't work:
1. Take screenshot of console
2. Copy all console logs
3. Report which checkpoint failed
4. I'll debug further

---

**Test now and report back!** 🎤☁️✨
