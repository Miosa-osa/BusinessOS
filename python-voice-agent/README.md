# OSA Voice Agent - LiveKit Native Implementation

This is the official LiveKit voice.Agent implementation for OSA, replacing the custom voice system.

**IMPORTANT:** We use **GROQ** (not OpenAI) for STT and LLM! The `livekit-plugins-openai` package provides compatibility wrappers (`.with_groq()`) that make Groq work with LiveKit's API.

## Quick Start

### 1. Activate Virtual Environment
```bash
cd python-voice-agent
source venv/bin/activate
```

### 2. Set Environment Variables
```bash
source setup_env.sh
```

### 3. Run Agent in Development Mode
```bash
python agent.py dev
```

The agent will connect to LiveKit cloud and wait for participants.

### 4. Test with Frontend
1. Open BusinessOS frontend: http://localhost:5173/window
2. Click the voice orb to activate voice
3. Speak: "Hello OSA"
4. Expected: Response in <1 second via WebRTC

## Architecture

```
Frontend (Svelte)
    ↓ WebRTC audio tracks
LiveKit Cloud (wss://macstudiosystems-yn61tekm.livekit.cloud)
    ↓ Routes audio
Python Voice Agent (this)
    ├─ VAD: Silero (voice activity detection)
    ├─ STT: GROQ Whisper large-v3 (transcription) ← NOT OpenAI!
    ├─ LLM: GROQ Llama 3.1 8B Instant (responses) ← NOT OpenAI!
    └─ TTS: ElevenLabs Turbo v2.5 (speech synthesis) ← NOT OpenAI!
    ↓ WebRTC audio back
Frontend plays audio
```

**Why does the code say `openai.STT.with_groq()`?**
LiveKit's plugin architecture uses OpenAI-compatible wrappers. The `.with_groq()` method swaps out the backend to use Groq's API instead of OpenAI's. We NEVER call OpenAI - all requests go to Groq!

## Configuration

All configuration is in `setup_env.sh`:
- `LIVEKIT_URL`: WebSocket URL for LiveKit cloud
- `LIVEKIT_API_KEY/SECRET`: LiveKit credentials
- `GROQ_API_KEY`: Groq API for STT + LLM
- `ELEVENLABS_API_KEY`: ElevenLabs for TTS
- `ELEVENLABS_VOICE_ID`: Voice ID for OSA (Osa voice)

## Files

- `agent.py`: Main voice agent implementation (~80 lines)
- `requirements.txt`: Python dependencies
- `setup_env.sh`: Environment variable setup
- `venv/`: Python virtual environment

## Deployment

See `Dockerfile` for production deployment to Cloud Run.

## Benefits vs Old System

| Metric | Old (Custom) | New (LiveKit Native) |
|--------|--------------|---------------------|
| Response Time | 2-3+ seconds | <1 second (target: 236ms) |
| Code Size | 2000+ lines | 80 lines |
| Streaming | Fake (chunks) | Real (WebRTC) |
| Interruptions | Not supported | Built-in |
| Maintenance | High (15+ files) | Low (1 file) |
| Reliability | Custom VAD | Battle-tested |

## Next Steps

1. ✅ Phase 1: Proof of Concept (this)
2. ⏳ Phase 2: Add Go backend context integration
3. ⏳ Phase 3: Production deployment

## Troubleshooting

**Agent doesn't start:**
- Check `source setup_env.sh` was run
- Verify all API keys are valid
- Check `python agent.py dev` output for errors

**No audio response:**
- Verify frontend is connected to same LiveKit room
- Check browser console for WebRTC errors
- Ensure microphone permissions granted

**Slow responses:**
- Check network latency to LiveKit cloud
- Verify Groq API key has quota remaining
- Check ElevenLabs API usage
