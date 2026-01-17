#!/bin/bash

# Navigate to voice agent directory
cd "$(dirname "$0")"

# Activate virtual environment
source venv/bin/activate

# Set environment variables
export LIVEKIT_URL="wss://macstudiosystems-yn61tekm.livekit.cloud"
export LIVEKIT_API_KEY="APIcFNUEtCEkZpa"
export LIVEKIT_API_SECRET="iBtjeSlz2ioQ8Ptd9SiOOW5B2ihO1Ff6gSjWtKanflxA"
export GROQ_API_KEY="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"
export ELEVENLABS_API_KEY="sk_4fd29ef975197a42a9d5d9b0b4ac809720e6a7c2ee8ef657"
export ELEVENLABS_VOICE_ID="KoVIHoyLDrQyd4pGalbs"
export GO_BACKEND_URL="http://localhost:8001"

echo "🎙️ Starting Voice Agent..."
echo "Connecting to LiveKit and Go Backend..."
echo ""

# Start the agent
python agent.py dev
