#!/bin/bash
# Setup environment variables for OSA Voice Agent

# LiveKit Configuration
export LIVEKIT_URL="wss://macstudiosystems-yn61tekm.livekit.cloud"
export LIVEKIT_API_KEY="APIcFNUEtCEkZpa"
export LIVEKIT_API_SECRET="iBtjeSlz2ioQ8Ptd9SiOOW5B2ihO1Ff6gSjWtKanflxA"

# Groq API (STT + LLM)
export GROQ_API_KEY="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"

# ElevenLabs API (TTS)
export ELEVENLABS_API_KEY="sk_4fd29ef975197a42a9d5d9b0b4ac809720e6a7c2ee8ef657"
export ELEVENLABS_VOICE_ID="KoVIHoyLDrQyd4pGalbs"

# Go Backend URL (for user context & tools)
export GO_BACKEND_URL="http://localhost:8001"

echo "✅ Environment variables set successfully!"
echo ""
echo "To start the agent:"
echo "  python agent.py dev"
