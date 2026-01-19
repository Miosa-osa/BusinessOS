# BusinessOS Voice Agent (Groq Whisper + ElevenLabs)

Python voice agent for BusinessOS using LiveKit WebRTC.

## Stack

- **STT**: Groq Whisper (fast, accurate speech-to-text)
- **LLM**: BusinessOS Go Backend → Groq (llama-3.3-70b-versatile)
- **TTS**: ElevenLabs (high-quality voice synthesis)
- **Transport**: LiveKit WebRTC

## Setup

### 1. Install Dependencies

```bash
cd voice-agent
pip install -r requirements.txt
```

### 2. Configure Environment

Copy `.env.example` to `.env` and fill in your API keys:

```bash
cp .env.example .env
```

Required keys:
- `LIVEKIT_API_KEY` - From LiveKit Cloud dashboard
- `LIVEKIT_API_SECRET` - From LiveKit Cloud dashboard
- `LIVEKIT_URL` - Your LiveKit server URL (wss://your-project.livekit.cloud)
- `GROQ_API_KEY` - From https://console.groq.com/keys
- `ELEVENLABS_API_KEY` - From https://elevenlabs.io/app/settings/api-keys
- `ELEVENLABS_VOICE_ID` - Voice ID from ElevenLabs (default: 21m00Tcm4TlvDq8ikWAM)
- `BACKEND_URL` - BusinessOS backend URL (http://localhost:8080 for local)

### 3. Run the Agent

```bash
python agent_groq.py dev
```

The agent will:
1. Connect to LiveKit Cloud
2. Wait for voice sessions from BusinessOS frontend
3. Process audio with Groq Whisper
4. Send transcripts to Go backend for LLM processing
5. Synthesize responses with ElevenLabs
6. Stream audio back to user

## Architecture

```
User → Frontend → LiveKit → Python Agent → Go Backend → Groq LLM
  ↑                              ↓
  └──────── ElevenLabs TTS ──────┘
```

## Files

- `agent_groq.py` - Main agent (Groq Whisper STT)
- `agent.py` - Alternative agent (Deepgram STT)
- `requirements.txt` - Python dependencies
- `Dockerfile` - Container build (for production deployment)

## Deployment

For production deployment to Cloud Run or similar:

```bash
docker build -t businessos-voice-agent .
docker push gcr.io/YOUR_PROJECT/businessos-voice-agent
gcloud run deploy businessos-voice-agent \
  --image gcr.io/YOUR_PROJECT/businessos-voice-agent \
  --platform managed \
  --region us-central1 \
  --set-env-vars "LIVEKIT_URL=...,GROQ_API_KEY=...,BACKEND_URL=..."
```

## Troubleshooting

**Agent not connecting:**
- Check LIVEKIT_URL, API_KEY, API_SECRET are correct
- Verify LiveKit Cloud dashboard shows the agent as connected

**No transcription:**
- Verify GROQ_API_KEY is valid
- Check agent logs for Groq API errors

**No speech output:**
- Verify ELEVENLABS_API_KEY is valid
- Check ELEVENLABS_VOICE_ID exists in your account

**Backend connection issues:**
- Ensure Go backend is running at BACKEND_URL
- Check `/api/chat` endpoint is accessible
- Verify CORS allows agent requests
