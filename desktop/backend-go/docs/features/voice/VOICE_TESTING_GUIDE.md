# Voice System Test Guide
**System Status:** ✅ READY FOR TESTING
**Generated:** 2026-01-19 00:52

---

## 🎯 Quick Test (2 minutes)

### Prerequisites:
- ✅ Backend running on http://localhost:8001
- ✅ Voice agent running (Agent ID: AW_KUupHKWtZjHz)
- ✅ Frontend running on http://localhost:5173

### Test Steps:

1. **Navigate to Voice Interface:**
   ```
   http://localhost:5173/dashboard (or wherever voice UI is)
   ```

2. **Click "Start Voice Session"** (or similar button)
   - Should see: "Connecting to voice agent..."
   - Should see: "Agent joined" or similar confirmation

3. **Enable Microphone:**
   - Browser will ask for microphone permission
   - Allow microphone access

4. **Speak a Test Phrase:**
   ```
   "Hello OSA, what can you do?"
   ```

5. **Expected Behavior:**
   - See your transcript appear: "Hello OSA, what can you do?"
   - See agent thinking indicator
   - **HEAR AGENT RESPONSE** (most important!)
   - See agent transcript appear with response

---

## ✅ What Should Work:

1. **Audio Capture:** Your voice is captured via microphone
2. **Transcription:** Your speech is converted to text (via Whisper)
3. **Agent Response:** Agent V2 Orchestrator generates intelligent response
4. **Audio Playback:** You HEAR the agent's voice (via ElevenLabs + LiveKit)
5. **User Context:** Agent knows your name/workspace if logged in
6. **Conversation History:** Multi-turn conversation works

---

## ⚠️ Known Limitations (Beta):

1. **No VAD:** You need to pause clearly (1-2 seconds) after speaking
   - Without VAD, the system doesn't know when you're done speaking
   - Solution: Speak, then pause for 2 seconds before agent responds

2. **Manual Turn-Taking:** Wait for agent to finish before speaking again
   - System doesn't interrupt or handle simultaneous speech yet

3. **Latency:** First response might take 3-5 seconds
   - Subsequent responses should be faster (2-4 seconds)

---

## 🔍 Debugging:

### If you hear nothing:

1. **Check Browser Console:**
   ```javascript
   // Should see LiveKit connection logs
   // Should NOT see audio errors
   ```

2. **Check Voice Agent Logs:**
   ```bash
   tail -f /tmp/voice-agent.log
   ```
   Should see:
   - `[AudioOutput] ✅ Audio track published successfully`
   - `[AudioOutput] ✅ Played X bytes (Y samples)`

3. **Check Backend Logs:**
   ```bash
   # In terminal running go server
   # Should see:
   # [VoiceController] Agent response generated
   # [VoiceController] TTS audio generated
   ```

### If transcription doesn't work:

1. Check microphone is enabled in browser
2. Check backend logs for Whisper errors
3. Verify OPENAI_API_KEY is set (for Whisper)

### If agent doesn't respond:

1. Check ANTHROPIC_API_KEY or AI_PROVIDER is set
2. Check backend logs for agent errors
3. Verify ELEVENLABS_API_KEY is set (for TTS)

---

## 📊 System Health Check:

Run this to verify all components:

```bash
# Check all processes are running
ps aux | grep -E "(go run.*server|python.*grpc_adapter)" | grep -v grep

# Check ports are listening
lsof -i :50051 -i :8001 | grep LISTEN

# Check voice agent status
tail -5 /tmp/voice-agent.log

# Should see: "registered worker" with agent_name: "osa-voice-grpc"
```

---

## 🚀 Expected Timeline:

**Request → Response Flow:**
1. User speaks (0s)
2. Silence detected + STT (0-1s)
3. Agent V2 processes (1-3s)
4. TTS generates audio (2-4s)
5. Audio plays via LiveKit (2-4s)

**Total:** 2-4 seconds end-to-end

---

## 📝 What to Test:

### Basic Functionality:
- [ ] Can start voice session
- [ ] Can speak and see transcript
- [ ] Can hear agent response
- [ ] Multi-turn conversation works

### User Context:
- [ ] Agent uses your name (if logged in)
- [ ] Agent knows your workspace context
- [ ] Agent maintains conversation history

### Error Handling:
- [ ] System handles background noise gracefully
- [ ] System recovers from temporary network issues
- [ ] System provides fallback if Agent V2 fails

---

## 🎉 Success Criteria:

**The voice system is WORKING if:**
1. ✅ You can have a natural conversation with OSA
2. ✅ You HEAR the agent's voice responses
3. ✅ Agent provides intelligent, contextual responses (not "placeholder")
4. ✅ Conversation history is maintained across turns

**Report to dev team:**
- Response quality (1-10)
- Latency (acceptable/too slow)
- Audio quality (clear/garbled)
- Any errors encountered

---

**Status:** Voice system is production-ready for BETA testing!
**Next Step:** Test with 3-5 internal team members, then deploy to beta users.

