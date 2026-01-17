"""
OSA Voice Agent - LiveKit Native Implementation with Hierarchical Context

Architecture:
- Level 0 (Identity): OSA personality (always loaded)
- Level 1 (User Context): Fetched at session start from Go backend
- Level 2-3 (Node/Data Context): Fetched on-demand via LLM tool calls

Technology Stack:
- STT: GROQ Whisper large-v3
- LLM: GROQ Llama 3.1 8B Instant
- TTS: ElevenLabs Turbo v2.5
- VAD: Silero
- Transport: LiveKit WebRTC
"""

import os
import json
import logging
from livekit.agents import (
    AutoSubscribe,
    JobContext,
    WorkerOptions,
    cli,
    voice,
)
from livekit import rtc
# Use dedicated Groq plugin for STT/LLM
from livekit.plugins import groq, elevenlabs, silero
from dotenv import load_dotenv

# Import our modules
from config import config
from context import fetch_user_context, extract_user_id_from_identity
from prompts import build_instructions, get_greeting
from tools import TOOL_DEFINITIONS, execute_tool

# Load environment variables
load_dotenv()

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def prewarm_process(proc):
    """Prewarm models for faster startup."""
    logger.info("[Agent] Prewarming models...")
    # Prewarm VAD model
    proc.userdata["vad"] = silero.VAD.load()
    logger.info("[Agent] VAD model loaded")


async def entrypoint(ctx: JobContext):
    """Main entrypoint for voice agent with hierarchical context."""
    logger.info(f"[Agent] Starting voice agent for room: {ctx.room.name}")

    # Validate configuration
    try:
        config.validate()
    except ValueError as e:
        logger.error(f"[Agent] Configuration error: {e}")
        return

    # Wait for participant to connect first
    await ctx.connect(auto_subscribe=AutoSubscribe.AUDIO_ONLY)

    # Get participant info
    participant = await ctx.wait_for_participant()
    user_id = extract_user_id_from_identity(participant.identity)
    logger.info(f"[Agent] Participant connected: {participant.identity} (user: {user_id})")

    # LEVEL 1: Fetch user context from Go backend
    user_context = await fetch_user_context(user_id)

    # LEVEL 0 + LEVEL 1: Build comprehensive instructions
    instructions = build_instructions(user_context)
    logger.info(f"[Agent] Instructions prepared for user: {user_context.get('name', 'Unknown')}")

    # Get prewarmed VAD or load it
    vad_instance = ctx.proc.userdata.get("vad") or silero.VAD.load()

    # Create the agent with instructions and tools
    agent = voice.Agent(
        instructions=instructions,
        # Register all tools for LLM function calling
        tools=TOOL_DEFINITIONS,
    )
    logger.info(f"[Agent] Agent configured with {len(TOOL_DEFINITIONS)} tools")

    # Create the agent session with STT, LLM, TTS, VAD
    session = voice.AgentSession(
        vad=vad_instance,
        # STT: GROQ Whisper
        stt=groq.STT(
            api_key=config.groq_api_key,
            model=config.stt_model,
        ),
        # LLM: GROQ Llama with tool calling
        llm=groq.LLM(
            api_key=config.groq_api_key,
            model=config.llm_model,
        ),
        # TTS: ElevenLabs
        tts=elevenlabs.TTS(
            api_key=config.elevenlabs_api_key,
            voice_id=config.elevenlabs_voice_id,
            model=config.tts_model,
        ),
        allow_interruptions=True,
    )

    # Helper to publish transcripts to frontend via data channel
    async def publish_transcript(type: str, text: str):
        """Publish transcript to frontend via data channel."""
        try:
            data = json.dumps({"type": type, "text": text})
            await ctx.room.local_participant.publish_data(
                data.encode(),
                reliable=True,
            )
            logger.info(f"[Agent] Published {type}: {text[:50]}...")
        except Exception as e:
            logger.error(f"[Agent] Failed to publish transcript: {e}")

    # Register event handlers for transcripts
    @session.on("user_speech_committed")
    def on_user_speech(msg):
        """Called when user finishes speaking."""
        logger.info(f"[Agent] User said: {msg.content}")
        # Publish to frontend
        import asyncio
        asyncio.create_task(publish_transcript("user_transcript", msg.content))

    @session.on("agent_speech_committed")
    def on_agent_speech(msg):
        """Called when agent finishes speaking."""
        logger.info(f"[Agent] Agent said: {msg.content}")
        # Publish to frontend
        import asyncio
        asyncio.create_task(publish_transcript("agent_transcript", msg.content))

    # Start the agent session with the agent and room
    await session.start(agent, room=ctx.room)

    # Greet the user naturally
    greeting = get_greeting(user_context.get("name"))
    await session.say(greeting)
    logger.info(f"[Agent] Greeting sent: {greeting}")

    logger.info("[Agent] Voice agent ready and running")


async def request_fnc(ctx):
    """Accept all job requests - agent will join any room"""
    logger.info(f"[Agent] Job request received for room: {ctx.room.name}")
    # Accept all jobs
    await ctx.accept()


if __name__ == "__main__":
    # Run the agent with explicit configuration
    logger.info("[Agent] Starting OSA voice agent worker...")

    cli.run_app(
        WorkerOptions(
            entrypoint_fnc=entrypoint,
            prewarm_fnc=prewarm_process,
            request_fnc=request_fnc,  # Accept all job requests
            agent_name="osa-voice-agent",  # Explicit agent name for dispatch
        ),
    )
