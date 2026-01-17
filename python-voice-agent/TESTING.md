# Testing Guide for OSA Voice Agent

## Phase 1 Testing: Basic Voice Functionality

### Prerequisites
1. Frontend running at http://localhost:5173
2. Python agent ready to start

### Test Steps

#### 1. Start the Python Voice Agent

```bash
cd python-voice-agent
source venv/bin/activate
source setup_env.sh
python agent.py dev
```

**Expected output:**
```
INFO:agent:Prewarming models...
INFO:agent:VAD model loaded
[LiveKit] Worker started, waiting for jobs...
```

The agent will wait for LiveKit to assign it a room.

#### 2. Open BusinessOS Frontend

Navigate to: http://localhost:5173/window

#### 3. Activate Voice

Click the voice orb/button to activate voice input.

**What should happen:**
- Frontend connects to LiveKit cloud via WebRTC
- LiveKit assigns the connection to our Python agent
- Python agent logs: "Starting voice agent for room: [room-id]"
- Python agent logs: "Voice agent ready and running"
- You should hear: "Hello! I'm OSA. How can I help you today?"

#### 4. Test Voice Interaction

**Test 1: Simple question**
- **Speak:** "Hello OSA"
- **Expected:** Response within <1 second
- **What to verify:**
  - Audio is clear
  - Response is fast (<1s from end of speech)
  - No "slow as fuck" feeling

**Test 2: Interruption**
- **Speak:** "Tell me a long story"
- **While OSA is speaking, interrupt:** "Stop!"
- **Expected:** OSA stops immediately and listens

**Test 3: Complex question**
- **Speak:** "What is BusinessOS?"
- **Expected:** Contextual response about the platform

### Success Criteria

✅ **PASS** if:
- Agent connects when frontend activates voice
- Response time is <1 second consistently
- Audio quality is clear (ElevenLabs TTS)
- Interruptions work (can cut off OSA mid-sentence)
- No playback issues or audio glitches

❌ **FAIL** if:
- Response time >2 seconds
- Audio doesn't play
- Interruptions don't work
- Connection errors in console

### Troubleshooting

**Issue: Agent doesn't connect**
- Check: Is frontend using same LiveKit URL?
- Check: Are LIVEKIT_URL/API_KEY/SECRET correct?
- Check: Frontend console for WebRTC errors

**Issue: No audio output**
- Check: Browser permissions for microphone
- Check: ELEVENLABS_API_KEY is valid
- Check: Browser console for audio errors
- Check: Volume is not muted

**Issue: Slow responses (>2s)**
- Check: Network latency to LiveKit cloud
- Check: GROQ_API_KEY quota remaining
- Check: ElevenLabs API usage limits

**Issue: "ModuleNotFoundError"**
- Run: `pip install -r requirements.txt` again
- Verify: `source venv/bin/activate` was run

### Logs to Check

**Python agent logs:**
```
INFO:agent:Starting voice agent for room: abc123
INFO:agent:Voice agent ready and running
```

**Frontend console (browser):**
```
[LiveKit] Connected to room
[LiveKit] Subscribed to audio track
```

### Performance Benchmarks

Target metrics (from LiveKit docs):
- **STT latency:** <200ms (Groq Whisper)
- **LLM latency:** <300ms (Groq Llama 3.1 8B)
- **TTS latency:** <400ms (ElevenLabs Turbo)
- **Total target:** <1000ms (1 second)

### Next Steps After Passing

Once Phase 1 tests pass:
1. ✅ Phase 1 complete
2. → Move to Phase 2: Add Go backend context integration
3. → Move to Phase 3: Production deployment

---

## Manual Test Checklist

Use this when testing:

- [ ] Python agent starts without errors
- [ ] Frontend connects to LiveKit
- [ ] Agent logs "Starting voice agent for room"
- [ ] Greeting plays: "Hello! I'm OSA..."
- [ ] User speech is transcribed correctly
- [ ] Response time <1 second
- [ ] Audio quality is clear
- [ ] Interruptions work
- [ ] No browser console errors
- [ ] No Python exceptions

**Test performed by:** _______________
**Date:** _______________
**Result:** PASS / FAIL
**Notes:** _______________
