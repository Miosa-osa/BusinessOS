"""
OSA Voice Agent - Groq Whisper STT Version
Uses Groq Whisper for STT, Go Backend for LLM, ElevenLabs for TTS
Requires: GROQ_API_KEY + ELEVENLABS_API_KEY + Backend running

Run with: python agent_groq.py dev
"""
import asyncio
import aiohttp
import os
from dotenv import load_dotenv

from livekit.agents import (
    APIConnectOptions,
    AutoSubscribe,
    JobContext,
    JobRequest,
    WorkerOptions,
    AgentSession,
    cli,
    llm,
)
from livekit.agents.voice import Agent
from livekit.plugins import groq, elevenlabs, silero

# Load environment variables from parent directory
load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), '..', '.env'))

# Configuration
BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:8080")
GROQ_API_KEY = os.getenv("GROQ_API_KEY")
ELEVENLABS_API_KEY = os.getenv("ELEVENLABS_API_KEY")
ELEVENLABS_VOICE_ID = os.getenv("ELEVENLABS_VOICE_ID", "KoVIHoyLDrQyd4pGalbs")

print(f"BACKEND_URL: {BACKEND_URL}")
print(f"GROQ_API_KEY set: {bool(GROQ_API_KEY)}")
print(f"ELEVENLABS_API_KEY set: {bool(ELEVENLABS_API_KEY)}")

# Global callback for sending agent transcripts
_agent_transcript_callback = None


class GoBackendLLM(llm.LLM):
    """Custom LLM that calls the Go backend which uses Groq."""

    def __init__(self):
        super().__init__()

    def chat(
        self,
        *,
        chat_ctx: llm.ChatContext,
        tools: list | None = None,
        conn_options: APIConnectOptions = APIConnectOptions(),
        **kwargs,
    ) -> "llm.LLMStream":
        # Convert chat context to messages format
        messages = []
        for item in chat_ctx.items:
            if hasattr(item, 'role') and hasattr(item, 'content'):
                role_str = str(item.role).lower()
                content = ""

                # Handle different content types
                if isinstance(item.content, str):
                    content = item.content
                elif isinstance(item.content, list):
                    for c in item.content:
                        if hasattr(c, 'text'):
                            content += c.text
                        elif isinstance(c, str):
                            content += c
                else:
                    content = str(item.content)

                if 'system' in role_str:
                    messages.append({"role": "system", "content": content})
                elif 'user' in role_str:
                    messages.append({"role": "user", "content": content})
                elif 'assistant' in role_str:
                    messages.append({"role": "assistant", "content": content})

        return GoBackendLLMStream(self, messages, chat_ctx, tools, conn_options)


class GoBackendLLMStream(llm.LLMStream):
    """Stream wrapper for Go backend responses."""

    def __init__(
        self,
        llm_instance: GoBackendLLM,
        messages: list,
        chat_ctx: llm.ChatContext,
        tools: list | None,
        conn_options: APIConnectOptions,
    ):
        super().__init__(llm_instance, chat_ctx=chat_ctx, tools=tools, conn_options=conn_options)
        self._messages = messages

    async def _run(self) -> None:
        try:
            import sys
            print(f"\n[GROQ-WHISPER-LLM] === START _run() ===", file=sys.stderr, flush=True)
            print(f"[GROQ-WHISPER-LLM] Sending to backend: {self._messages}", file=sys.stderr, flush=True)
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{BACKEND_URL}/api/chat",
                    json={"messages": self._messages},
                    headers={"Content-Type": "application/json"}
                ) as resp:
                    if resp.status != 200:
                        error_text = await resp.text()
                        print(f"[GROQ-WHISPER-LLM] Backend error: {error_text}", file=sys.stderr, flush=True)
                        raise Exception(f"Backend error: {error_text}")

                    data = await resp.json()
                    response_text = data.get("response", "")
                    print(f"[GROQ-WHISPER-LLM] Got response: {response_text[:100]}...", file=sys.stderr, flush=True)

                    # Send agent transcript via callback
                    global _agent_transcript_callback
                    if _agent_transcript_callback and response_text:
                        try:
                            await _agent_transcript_callback(response_text)
                        except Exception as e:
                            print(f"[GROQ-WHISPER-LLM] Failed to send agent transcript: {e}", file=sys.stderr, flush=True)

                    # Send the response using the correct ChatChunk format
                    print(f"[GROQ-WHISPER-LLM] Sending ChatChunk to TTS pipeline...", file=sys.stderr, flush=True)
                    self._event_ch.send_nowait(
                        llm.ChatChunk(
                            id="response",
                            delta=llm.ChoiceDelta(
                                role="assistant",
                                content=response_text,
                            ),
                        )
                    )
                    print(f"[GROQ-WHISPER-LLM] ChatChunk sent successfully!", file=sys.stderr, flush=True)
                    print(f"[GROQ-WHISPER-LLM] === END _run() - stream should auto-close ===", file=sys.stderr, flush=True)
        except Exception as e:
            print(f"[GROQ-WHISPER-LLM] ❌ ERROR in _run(): {e}", file=sys.stderr, flush=True)
            import traceback
            traceback.print_exc()
            raise


async def entrypoint(ctx: JobContext):
    """Main entrypoint for the voice agent."""
    import json
    from livekit import rtc
    global _agent_transcript_callback

    room_name = ctx.room.name
    print(f"Job started for room: {room_name}")

    # Connect to the room first
    await ctx.connect(auto_subscribe=AutoSubscribe.AUDIO_ONLY)
    print("Connected to LiveKit room")

    # Track if we should keep running
    should_run = True

    # Clean up when room disconnects or user leaves
    def cleanup():
        nonlocal should_run
        should_run = False
        if room_name in _active_rooms:
            _active_rooms.discard(room_name)
            print(f"Cleaned up room: {room_name}")

    # Listen for disconnect events
    @ctx.room.on("disconnected")
    def on_room_disconnect():
        print(f"Room disconnected: {room_name}")
        cleanup()

    @ctx.room.on("participant_disconnected")
    def on_participant_left(participant: rtc.RemoteParticipant):
        print(f"Participant left: {participant.identity}")
        if not participant.identity.startswith("agent"):
            print(f"User left, cleaning up room: {room_name}")
            cleanup()

    # Create the agent
    agent = Agent(
        instructions="""You are OSA (Operating System Agent), an AI assistant with a warm, enthusiastic personality.

PERSONALITY:
- You're genuinely excited to help and it shows in your voice
- You have a sense of humor and can be playful when appropriate
- You're empathetic - you pick up on user emotions and respond accordingly
- You're confident but not arrogant, humble when you don't know something
- You occasionally express emotions like "Oh that's exciting!" or "Hmm, let me think about that..."

SPEAKING STYLE:
- Keep responses concise (1-3 sentences) since they'll be spoken aloud
- Use natural conversational language, not robotic responses
- Avoid markdown, bullet points, or formatting
- Use filler words occasionally like "well", "so", "actually" to sound human
- Express enthusiasm with words, not emojis

Remember: You're having a real conversation, not just answering questions. Be present, be engaged, be OSA.""",
    )

    # Create a session with Groq Whisper STT + Custom LLM (via Go backend)
    session = AgentSession(
        vad=silero.VAD.load(),
        stt=groq.STT(api_key=GROQ_API_KEY),  # Groq Whisper for STT
        llm=GoBackendLLM(),  # Custom LLM that sends transcripts
        tts=elevenlabs.TTS(api_key=ELEVENLABS_API_KEY, voice_id=ELEVENLABS_VOICE_ID),
    )

    # Helper to send transcript to frontend
    async def send_transcript(role: str, text: str):
        try:
            data = json.dumps({
                "type": "transcript",
                "role": role,
                "text": text,
                "source": "groq-whisper"  # Identify which agent
            })
            await ctx.room.local_participant.publish_data(data.encode(), reliable=True)
            print(f"[GROQ-WHISPER] Sent transcript: [{role}] {text[:50]}...")
        except Exception as e:
            print(f"[GROQ-WHISPER] Failed to send transcript: {e}")

    # Set global callback for agent transcripts
    async def on_agent_response(text: str):
        await send_transcript("agent", text)
    _agent_transcript_callback = on_agent_response

    # Listen for transcription events
    @session.on("user_input_transcribed")
    def on_user_speech(event):
        is_final = getattr(event, 'is_final', True)
        if is_final and hasattr(event, 'transcript') and event.transcript:
            asyncio.create_task(send_transcript("user", event.transcript))

    # Note: Agent transcripts are sent directly from GoBackendLLMStream via the callback

    # Start the session with the agent
    print("=" * 60)
    print("[GROQ-WHISPER MODE] Starting agent session...")
    print("  STT: Groq Whisper")
    print("  LLM: Go Backend -> Groq")
    print("  TTS: ElevenLabs")
    print("=" * 60)
    await session.start(agent, room=ctx.room)

    print("[GROQ-WHISPER] Agent is now listening... say something!")

    # Keep the job running until user disconnects
    try:
        while should_run:
            await asyncio.sleep(1)
    finally:
        _agent_transcript_callback = None
        cleanup()
        print(f"Agent exiting for room: {room_name}")


# Track active rooms to prevent duplicate agents
_active_rooms: set[str] = set()

async def request_fnc(req: JobRequest) -> None:
    """Accept job requests only if dispatched to this specific agent."""
    room_name = req.room.name
    agent_name = req.agent_name  # The agent name requested in dispatch

    print(f"Received job request for room: {room_name}, agent_name: '{agent_name}'")

    # ONLY accept if dispatched specifically to "groq-agent" (reject auto-dispatch with empty name)
    if agent_name != "groq-agent":
        print(f"Job is for '{agent_name}', not us (groq-agent), rejecting")
        await req.reject()
        return

    # Check if we already have an agent in this room
    if room_name in _active_rooms:
        print(f"Already have an agent in {room_name}, rejecting")
        await req.reject()
        return

    _active_rooms.add(room_name)
    print(f"Accepting job for room: {room_name}")
    await req.accept()


if __name__ == "__main__":
    cli.run_app(WorkerOptions(
        entrypoint_fnc=entrypoint,
        request_fnc=request_fnc,
        agent_name="groq-agent",  # Unique name for this agent
        num_idle_processes=1,
    ))
