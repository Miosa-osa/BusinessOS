"""
OSA Voice Agent - Minimal Implementation
Simple voice chat: STT → LLM → TTS
- Sends transcripts to frontend
- Auto-disconnects when user leaves (prevents duplicates)
"""

import os
import json
import time
import asyncio
from livekit.agents import JobContext, WorkerOptions, cli, voice
from livekit.plugins import groq, elevenlabs, silero
from livekit import rtc
from dotenv import load_dotenv

load_dotenv()

SIMPLE_PROMPT = """You are OSA, a helpful voice assistant.
Keep responses short (1-2 sentences). Be conversational and natural."""

def prewarm_process(proc):
    """Prewarm VAD model for faster startup."""
    proc.userdata["vad"] = silero.VAD.load()
    print("[Agent] VAD model preloaded")

async def entrypoint(ctx: JobContext):
    """Main entrypoint for minimal voice agent."""
    print(f"[Agent] Starting for room: {ctx.room.name}")

    # Connect and wait for user
    await ctx.connect()
    participant = await ctx.wait_for_participant()
    print(f"[Agent] User connected: {participant.name}")

    # Track if user disconnected
    user_disconnected = False

    # Listen for user disconnect - shut down agent when user leaves
    @ctx.room.on("participant_disconnected")
    def on_participant_disconnected(p: rtc.RemoteParticipant):
        nonlocal user_disconnected
        if p.identity == participant.identity:
            print(f"[Agent] User {p.name} disconnected - shutting down agent")
            user_disconnected = True
            # Disconnect agent from room (schedule async task)
            asyncio.create_task(ctx.room.disconnect())

    # Get prewarmed VAD
    vad = ctx.proc.userdata.get("vad") or silero.VAD.load()

    # Create simple session
    session = voice.AgentSession(
        vad=vad,
        stt=groq.STT(
            api_key=os.getenv("GROQ_API_KEY"),
            model="whisper-large-v3-turbo"
        ),
        llm=groq.LLM(
            api_key=os.getenv("GROQ_API_KEY"),
            model="llama-3.1-8b-instant"
        ),
        tts=elevenlabs.TTS(
            api_key=os.getenv("ELEVENLABS_API_KEY"),
            voice_id=os.getenv("ELEVENLABS_VOICE_ID", "KoVIHoyLDrQyd4pGalbs"),
            model=os.getenv("ELEVENLABS_MODEL", "eleven_flash_v2_5"),
        ),
        allow_interruptions=True,
    )

    # Helper to send transcripts to frontend
    async def publish_transcript(msg_type: str, text: str):
        """Send transcript to frontend via data channel for captions."""
        try:
            data = json.dumps({"type": msg_type, "text": text})
            await ctx.room.local_participant.publish_data(
                data.encode(),
                reliable=True,
            )
        except Exception as e:
            print(f"[Agent] Failed to send transcript: {e}")

    # Log and send transcripts
    # NOTE: Event handlers MUST be synchronous - use asyncio.create_task for async operations
    @session.on("user_input_transcribed")
    def on_user_speech(msg):
        if msg.is_final:  # Only log final transcripts
            timestamp = time.strftime('%H:%M:%S')
            print(f"\n{'='*80}")
            print(f"🎤 USER [{timestamp}]: {msg.transcript}")
            print(f"{'='*80}\n")

            # Send to frontend (schedule async task)
            asyncio.create_task(publish_transcript("user_transcript", msg.transcript))

    # DEBUG: Listen to ALL session events to find the right one
    @session.on("conversation_item_added")
    def on_conversation_item(item):
        # Print EVERYTHING about this item to debug
        print(f"\n[DEBUG] ===== conversation_item_added =====")
        print(f"[DEBUG] item type: {type(item)}")
        print(f"[DEBUG] item attributes: {dir(item)}")
        print(f"[DEBUG] item.__dict__: {getattr(item, '__dict__', 'NO __dict__')}")

        # Try to get role
        role = getattr(item, 'role', None)
        print(f"[DEBUG] item.role = {role}")

        # Try to get message if it exists
        message = getattr(item, 'message', None)
        if message:
            print(f"[DEBUG] item.message type: {type(message)}")
            print(f"[DEBUG] item.message attributes: {dir(message)}")
            msg_role = getattr(message, 'role', None)
            msg_content = getattr(message, 'content', None)
            print(f"[DEBUG] item.message.role = {msg_role}")
            print(f"[DEBUG] item.message.content = {msg_content}")

        # Try text_content
        text_content = getattr(item, 'text_content', None)
        print(f"[DEBUG] item.text_content = {text_content}")
        print(f"[DEBUG] =====================================\n")

        # Only process agent messages (ignore user messages - we handle those separately)
        if hasattr(item, 'role') and item.role == 'agent':
            timestamp = time.strftime('%H:%M:%S')
            text = item.text_content if hasattr(item, 'text_content') else str(item)

            print(f"\n{'='*80}")
            print(f"🤖 OSA [{timestamp}]: {text}")
            print(f"{'='*80}\n")

            # Send to frontend (schedule async task)
            asyncio.create_task(publish_transcript("agent_transcript", text))

    @session.on("speech_created")
    def on_speech_created(evt):
        print(f"[DEBUG] speech_created fired! Event attributes: {dir(evt)}")
        print(f"[DEBUG] speech_created event: {evt}")

    @session.on("agent_state_changed")
    def on_agent_state_changed(state):
        print(f"[DEBUG] agent_state_changed fired! State: {state}")

    # Start agent with simple prompt
    agent = voice.Agent(instructions=SIMPLE_PROMPT)
    await session.start(agent, room=ctx.room)
    print("[Agent] Voice session started - waiting for speech")
    print("[Agent] Will auto-shutdown when user disconnects")

if __name__ == "__main__":
    cli.run_app(WorkerOptions(
        entrypoint_fnc=entrypoint,
        prewarm_fnc=prewarm_process,
        num_idle_processes=0  # No pre-spawned processes in dev mode
    ))
