# Voice System Testing Guide

## 🎤 Voice System Overview

BusinessOS 3D Desktop includes a full voice interaction system with:
- **Active Listening**: Real-time speech-to-text using Whisper
- **Voice Commands**: Natural language command parsing
- **OSA Voice Responses**: Text-to-speech using ElevenLabs
- **Conversational Mode**: Back-and-forth conversation with OSA

---

## 🧪 Testing Procedure

### Step 1: Verify Backend Services

Open a terminal and check if backend is running:

```bash
cd /home/miosa/Desktop/BusinessOS/desktop/backend-go
go run cmd/server/main.go
```

**Look for these startup messages:**
```
✓ Whisper service initialized (local speech-to-text)
✓ ElevenLabs service initialized (OSA voice enabled)
```

**If you see these warnings:**
- ❌ `Whisper service not fully configured` → Whisper binary/model missing
- ❌ `ElevenLabs service not configured` → API key/voice ID not set in `.env`

### Step 2: Check Frontend Build

```bash
cd /home/miosa/Desktop/BusinessOS/frontend
npm run dev
```

Navigate to `http://localhost:5173` and enter 3D Desktop mode.

### Step 3: Test Microphone Permissions

1. Click the **circular microphone button** (bottom-right, above dock)
2. Browser should prompt for microphone access
3. Click **Allow**

**Expected Result:**
- Button turns BLUE
- Status text shows "Listening..."
- Pulse rings animate around button

### Step 4: Open Debug Panel

1. Click the **🐛 Debug** button (top-right, yellow)
2. Debug panel opens showing:
   - Listening: 🎤 YES
   - Speaking: ❌ NO
   - Transcript: (empty or showing text)
   - Live logs

### Step 5: Test Voice Input

**Open Browser Console** (F12 → Console tab) - THIS IS CRITICAL!

**Speak clearly into your microphone:**
- "Open chat"
- "Switch to grid view"
- "Hello OSA"

**Watch the console for these logs:**

✅ **WORKING PIPELINE:**
```
[ActiveListening] 📤 Sending audio chunk { size: 12345, type: "audio/webm;codecs=opus", sessionId: "..." }
[ActiveListening] 🌐 Calling /api/transcribe/realtime...
[ActiveListening] 📨 Response received: 200
[ActiveListening] 📝 Transcription result: { text: "open chat", language: "en", is_final: true }
[ActiveListening] ✅ Transcribed: { text: "open chat", language: "en", isFinal: true }
[Voice Debug] 📝 Transcript (final): "open chat"
[Voice Debug] 🎯 Command detected: focus_module
[Voice Debug] 🔊 OSA responding: "Opening chat"
```

❌ **BROKEN PIPELINE - Possible Issues:**

**Issue 1: No audio chunks being sent**
```
// Console is silent - no [ActiveListening] logs at all
```
**Diagnosis:** MediaRecorder not working
**Fix:** Check browser compatibility, verify microphone stream

**Issue 2: Empty transcriptions**
```
[ActiveListening] 📤 Sending audio chunk { size: 45, type: "audio/webm;codecs=opus", ... }
[ActiveListening] 📨 Response received: 200
[ActiveListening] ⚠️ Empty transcription (silence or unclear audio)
```
**Diagnosis:** Audio too quiet OR Whisper not transcribing
**Fixes:**
1. Check microphone volume (speak louder)
2. Check Whisper service on backend
3. Verify audio format conversion (WebM → WAV)

**Issue 3: Backend error**
```
[ActiveListening] 📨 Response received: 500
[ActiveListening] ❌ Transcription failed: 500 "Internal Server Error"
```
**Diagnosis:** Backend processing error
**Fix:** Check backend terminal logs for error details

**Issue 4: Authentication error**
```
[ActiveListening] 📨 Response received: 401
[ActiveListening] ❌ Transcription failed: 401 "Unauthorized"
```
**Diagnosis:** Session/auth issue
**Fix:** Clear cookies, log in again

---

## 🔍 Debugging Checklist

### Frontend Checks

- [ ] Microphone button is clickable (z-index: 600)
- [ ] Button turns blue when clicked
- [ ] Debug panel shows "Listening: 🎤 YES"
- [ ] Browser console shows `[ActiveListening]` logs
- [ ] Audio chunks have size > 100 bytes
- [ ] Fetch calls are reaching `/api/transcribe/realtime`
- [ ] Response status is 200 (not 401, 404, 500)
- [ ] Response contains `{ text: "...", language: "en", is_final: true }`

### Backend Checks

- [ ] Server is running on port 8080
- [ ] Whisper service initialized successfully
- [ ] FFmpeg is installed (`which ffmpeg` returns path)
- [ ] Whisper model exists at expected path
- [ ] `/api/transcribe/realtime` endpoint registered
- [ ] Audio file conversion working (WebM → WAV)
- [ ] Temp files being created/cleaned up

### Audio Pipeline Checks

- [ ] MediaRecorder creating audio blobs every 2 seconds
- [ ] Audio blobs are WebM/Opus format
- [ ] FormData correctly sending audio file
- [ ] Backend receiving multipart/form-data
- [ ] FFmpeg converting to 16kHz mono WAV
- [ ] Whisper transcribing WAV file
- [ ] Response JSON returning to frontend

---

## 🐛 Common Issues and Fixes

### Issue: "MediaRecorder not supported"
**Symptoms:** Alert saying MediaRecorder not supported
**Fix:** Use Chrome/Edge/Firefox (Safari may have issues)

### Issue: "No microphone stream available"
**Symptoms:** Alert saying no microphone access
**Fixes:**
1. Grant microphone permission in browser settings
2. Check if another app is using microphone
3. Verify microphone is connected/working

### Issue: "Empty transcription" every time
**Symptoms:** Logs show empty results, never any text
**Possible Causes:**
1. **Audio too quiet:** Speak louder, closer to mic
2. **Wrong audio format:** Check FFmpeg conversion logs
3. **Whisper not working:** Verify Whisper service initialized
4. **Audio chunk too short:** Default is 2 seconds, may need adjustment

**Debug Steps:**
1. Check audio chunk size in console (should be > 1000 bytes)
2. Check backend logs for FFmpeg errors
3. Check backend logs for Whisper errors
4. Test Whisper manually with a WAV file

### Issue: Transcription works but commands don't execute
**Symptoms:** Text appears in debug panel but nothing happens
**Debug:**
1. Check what command was detected: `[Voice Debug] 🎯 Command detected: unknown`
2. If "unknown", it triggers conversational mode
3. Check if OSA agent responds

**Fixes:**
1. Use exact command phrases: "open chat", "switch to grid view"
2. Check `voiceCommands.ts` for recognized patterns
3. Add new patterns if needed

### Issue: OSA not speaking (no voice response)
**Symptoms:** Commands execute but no audio
**Debug Logs:**
```
[Voice Debug] 🔊 OSA responding: "Opening chat"
// But no audio plays
```

**Possible Causes:**
1. **ElevenLabs not configured:** Check `.env` for API key/voice ID
2. **Network issue:** Check browser network tab for `/api/osa/speak` calls
3. **Audio playback blocked:** Browser may block audio, check console for errors
4. **API quota exceeded:** Check ElevenLabs dashboard

**Fixes:**
1. Verify `.env` has correct credentials
2. Check network tab - should see POST to `/api/osa/speak` returning audio/mpeg
3. User must interact with page first (click something) before audio can play
4. Check ElevenLabs quota/billing

---

## 🎯 Test Cases

### Test Case 1: Simple Voice Command
1. Click microphone button
2. Say: "open chat"
3. **Expected:** Chat window opens, OSA says "Opening chat"

### Test Case 2: Layout Management
1. Say: "enter edit mode"
2. **Expected:** Edit mode toolbar appears, OSA confirms
3. Say: "exit edit mode"
4. **Expected:** Edit mode exits

### Test Case 3: View Switching
1. Say: "switch to grid view"
2. **Expected:** Windows spread to grid, OSA confirms
3. Say: "switch to orb view"
4. **Expected:** Windows collapse to orb

### Test Case 4: Conversational Mode
1. Say: "what can you help me with?"
2. **Expected:** OSA responds with capabilities description
3. Say: "tell me a joke"
4. **Expected:** OSA tells a joke (via chat agent)

### Test Case 5: Multiple Commands in Sequence
1. Say: "open dashboard"
2. Wait for OSA response
3. Say: "zoom in"
4. **Expected:** Both commands execute with voice confirmations

---

## 📊 Performance Metrics

### Expected Latency:
- **Audio chunk creation:** Every 2 seconds
- **Transcription time:** 500-1500ms (depends on audio length)
- **Command execution:** < 100ms
- **OSA voice response:** 1-2 seconds (TTS generation + playback)
- **Total loop:** 3-5 seconds from speech to OSA response

### Audio Chunk Sizes:
- **Silence/quiet:** 50-200 bytes
- **Normal speech:** 500-2000 bytes
- **Loud/active speech:** 2000-5000 bytes

If chunks are consistently < 100 bytes, microphone may not be capturing audio.

---

## 🔧 Manual Testing Tools

### Test Whisper Backend Directly

Create a test WAV file and transcribe:

```bash
# Record a test audio file
ffmpeg -f avfoundation -i ":0" -t 3 -ar 16000 -ac 1 test.wav

# Test Whisper transcription via API
curl -X POST http://localhost:8080/api/transcribe/realtime \
  -F "audio=@test.wav" \
  -H "Cookie: session_token=YOUR_SESSION_TOKEN"
```

### Test ElevenLabs TTS Directly

```bash
curl -X POST http://localhost:8080/api/osa/speak \
  -H "Content-Type: application/json" \
  -H "Cookie: session_token=YOUR_SESSION_TOKEN" \
  -d '{"text":"Hello, this is a test"}' \
  --output test_tts.mp3

# Play the audio
open test_tts.mp3  # macOS
```

### Check FFmpeg Conversion

```bash
# Check if FFmpeg can convert WebM to WAV
ffmpeg -i test.webm -ar 16000 -ac 1 test_converted.wav

# Verify WAV format
file test_converted.wav
# Should show: RIFF (little-endian) data, WAVE audio, Microsoft PCM, 16 bit, mono 16000 Hz
```

---

## 📝 Reporting Issues

When reporting voice system issues, include:

1. **Console logs** (all `[ActiveListening]` and `[Voice Debug]` messages)
2. **Debug panel screenshot** showing status
3. **Browser and version** (Chrome 120, Firefox 115, etc.)
4. **Microphone type** (built-in, USB, Bluetooth)
5. **Backend logs** (from Go server terminal)
6. **What you said** vs **what was transcribed** (if anything)
7. **Network tab** showing `/api/transcribe/realtime` requests

### Example Good Bug Report:

```
**Issue:** Voice transcription not working

**Environment:**
- Browser: Chrome 120.0.6099.129
- OS: macOS 14.2
- Microphone: Built-in MacBook microphone

**Console Logs:**
[ActiveListening] 📤 Sending audio chunk { size: 1234, type: "audio/webm;codecs=opus", sessionId: "abc123" }
[ActiveListening] 🌐 Calling /api/transcribe/realtime...
[ActiveListening] 📨 Response received: 500
[ActiveListening] ❌ Transcription failed: 500 "FFmpeg conversion error"

**Backend Logs:**
ERROR: Failed to convert WebM to WAV: exec: "ffmpeg": executable file not found in $PATH

**What I Did:**
1. Clicked microphone button
2. Said "open chat"
3. Saw error in console

**Expected:** Transcription should work
**Actual:** 500 error, FFmpeg not found
```

---

## ✅ Success Criteria

Voice system is working correctly when:

- [ ] Microphone button activates (turns blue)
- [ ] Debug panel shows "Listening: 🎤 YES"
- [ ] Console shows audio chunks being sent (size > 500 bytes)
- [ ] Console shows successful transcriptions
- [ ] Spoken commands execute correctly
- [ ] OSA voice responses play
- [ ] Conversational mode works for unknown phrases
- [ ] No console errors
- [ ] Latency < 5 seconds for full loop

---

## 🚀 Next Steps After Testing

Once voice system is confirmed working:

1. **Fine-tune voice command patterns** - Add more natural language variations
2. **Improve error handling** - Better user feedback for failures
3. **Add voice command discovery** - Help users learn available commands
4. **Optimize latency** - Reduce chunk intervals, faster transcription
5. **Add language support** - Multi-language voice commands
6. **Implement push-to-talk** - Hold button instead of toggle
7. **Add voice profiles** - User-specific wake words/preferences

---

**Last Updated:** 2026-01-14
**Voice System Version:** 1.0.0
**Backend:** Go 1.24.1 + Whisper + ElevenLabs
**Frontend:** SvelteKit + MediaRecorder API
