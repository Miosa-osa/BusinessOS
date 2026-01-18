"""
OSA Voice Agent - gRPC Thin Adapter
====================================

Minimal (~150 lines) adapter bridging LiveKit audio to Go Voice Controller via gRPC.

Architecture:
    LiveKit Room ← Audio I/O → Python Adapter ← gRPC Stream → Go Voice Controller
                                                                    ↓
                                                              Agent V2 + Memory + RAG

ALL intelligence runs in Go. Python is just audio I/O.
"""

import asyncio
import json
import os
from typing import AsyncIterator

import grpc
from dotenv import load_dotenv
from livekit import rtc
from livekit.agents import JobContext, WorkerOptions, cli

# Import generated gRPC stubs
from voice.v1 import voice_pb2, voice_pb2_grpc

load_dotenv()

# Configuration
GRPC_SERVER = os.getenv("GRPC_VOICE_SERVER", "localhost:50051")


async def entrypoint(ctx: JobContext):
    """Minimal LiveKit entrypoint - audio bridge to Go gRPC."""

    print(f"[Adapter] Starting for room: {ctx.room.name}")

    # Connect to room and wait for user
    await ctx.connect()
    participant = await ctx.wait_for_participant()
    print(f"[Adapter] User connected: {participant.name}")

    # Session identifiers
    session_id = ctx.room.name
    user_id = participant.identity

    # Connect to Go gRPC server
    channel = grpc.aio.insecure_channel(GRPC_SERVER)
    stub = voice_pb2_grpc.VoiceServiceStub(channel)
    print(f"[Adapter] Connected to gRPC server: {GRPC_SERVER}")

    # Audio queue for user speech
    audio_queue: asyncio.Queue[bytes] = asyncio.Queue()
    sequence = 0

    # Subscribe to user audio track
    @ctx.room.on("track_subscribed")
    def on_track_subscribed(track: rtc.Track, *_):
        if track.kind == rtc.TrackKind.KIND_AUDIO:
            print(f"[Adapter] 🎤 User audio track subscribed")
            audio_stream = rtc.AudioStream(track)
            asyncio.create_task(capture_audio(audio_stream, audio_queue))

    # Start bidirectional gRPC stream
    grpc_stream = stub.ProcessVoice()

    # Task 1: Send audio frames to Go
    async def send_frames():
        nonlocal sequence
        while True:
            try:
                audio_bytes = await audio_queue.get()
                if audio_bytes is None:
                    break

                # Send audio frame to Go
                await grpc_stream.write(voice_pb2.AudioFrame(
                    session_id=session_id,
                    user_id=user_id,
                    audio_data=audio_bytes,
                    sequence=sequence,
                    direction="user",
                    sample_rate=24000,
                    is_final=False,  # TODO: VAD detection
                ))
                sequence += 1
            except Exception as e:
                print(f"[Adapter] ❌ Send error: {e}")
                break

    # Task 2: Receive responses from Go
    async def receive_responses():
        async for response in grpc_stream:
            try:
                # Handle different response types
                if response.type == voice_pb2.ResponseType.TRANSCRIPT_USER:
                    print(f"[Adapter] 🎤 User: {response.text}")
                    await send_to_frontend("user_transcript", response.text)

                elif response.type == voice_pb2.ResponseType.TRANSCRIPT_AGENT:
                    print(f"[Adapter] 🤖 OSA: {response.text}")
                    await send_to_frontend("agent_transcript", response.text)

                elif response.type == voice_pb2.ResponseType.AUDIO:
                    # Play audio from Go TTS
                    await play_audio(response.audio_data)

                elif response.type == voice_pb2.ResponseType.STATE_UPDATE:
                    print(f"[Adapter] 🔄 State: {response.state}")

                elif response.type == voice_pb2.ResponseType.DONE:
                    print(f"[Adapter] ✅ Agent done speaking")

                elif response.type == voice_pb2.ResponseType.ERROR:
                    print(f"[Adapter] ❌ Error: {response.error}")

            except Exception as e:
                print(f"[Adapter] ❌ Receive error: {e}")
                break

    async def send_to_frontend(msg_type: str, text: str):
        """Send transcript to frontend via data channel."""
        try:
            data = json.dumps({"type": msg_type, "text": text})
            await ctx.room.local_participant.publish_data(
                data.encode(),
                reliable=True,
            )
        except Exception as e:
            print(f"[Adapter] ❌ Data send failed: {e}")

    async def play_audio(audio_bytes: bytes):
        """Play audio response via LiveKit (placeholder)."""
        # TODO: Implement audio playback via LiveKit track
        print(f"[Adapter] 🔊 Would play {len(audio_bytes)} bytes")

    # Run both tasks concurrently
    try:
        await asyncio.gather(send_frames(), receive_responses())
    except Exception as e:
        print(f"[Adapter] ❌ Stream error: {e}")
    finally:
        await channel.close()
        print(f"[Adapter] Session ended: {session_id}")


async def capture_audio(stream: rtc.AudioStream, queue: asyncio.Queue):
    """Capture audio frames from LiveKit and queue for gRPC."""
    async for frame in stream:
        audio_bytes = bytes(frame.data)
        await queue.put(audio_bytes)


if __name__ == "__main__":
    print("=" * 80)
    print("🎤 OSA Voice Agent - gRPC Thin Adapter")
    print(f"gRPC Server: {GRPC_SERVER}")
    print("Connecting to Go Voice Controller...")
    print("=" * 80)

    cli.run_app(WorkerOptions(
        entrypoint_fnc=entrypoint,
        agent_name="osa-voice-grpc",
    ))
