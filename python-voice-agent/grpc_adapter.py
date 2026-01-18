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
import io
import json
import os
import subprocess
from typing import AsyncIterator

import grpc
import numpy as np
from dotenv import load_dotenv
from livekit import rtc
from livekit.agents import JobContext, WorkerOptions, cli

# Import generated gRPC stubs
from voice.v1 import voice_pb2, voice_pb2_grpc

load_dotenv()

# Configuration
GRPC_SERVER = os.getenv("GRPC_VOICE_SERVER", "localhost:50051")
LIVEKIT_SAMPLE_RATE = 48000  # LiveKit standard
FRAME_DURATION_MS = 20  # WebRTC standard


class AudioOutputManager:
    """Manages audio output to LiveKit room via published track."""

    def __init__(self, room: rtc.Room):
        self.room = room
        self.source = rtc.AudioSource(LIVEKIT_SAMPLE_RATE, 1)  # 48kHz, mono
        self.track = rtc.LocalAudioTrack.create_audio_track("agent-voice", self.source)
        self.publication = None
        self._is_published = False

    async def initialize(self):
        """Publish the audio track to the room."""
        try:
            options = rtc.TrackPublishOptions()
            options.source = rtc.TrackSource.SOURCE_MICROPHONE

            self.publication = await self.room.local_participant.publish_track(
                self.track,
                options
            )
            self._is_published = True
            print("[AudioOutput] ✅ Audio track published successfully")
        except Exception as e:
            print(f"[AudioOutput] ❌ Failed to publish track: {e}")
            raise

    async def play_audio_chunk(self, audio_bytes: bytes):
        """
        Convert MP3 audio to PCM and play via LiveKit using ffmpeg.

        Args:
            audio_bytes: Raw MP3 audio data from TTS
        """
        if not self._is_published:
            print("[AudioOutput] ⚠️ Track not published, skipping audio")
            return

        try:
            # Use ffmpeg to convert MP3 → PCM (48kHz, mono, s16le)
            process = subprocess.Popen(
                [
                    'ffmpeg',
                    '-i', 'pipe:0',  # Input from stdin
                    '-f', 's16le',   # Output format: signed 16-bit little-endian
                    '-acodec', 'pcm_s16le',  # PCM codec
                    '-ar', str(LIVEKIT_SAMPLE_RATE),  # Sample rate: 48kHz
                    '-ac', '1',      # Channels: mono
                    'pipe:1'         # Output to stdout
                ],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )

            # Send MP3 data and get PCM output
            pcm_bytes, stderr = process.communicate(input=audio_bytes)

            if process.returncode != 0:
                print(f"[AudioOutput] ❌ ffmpeg error: {stderr.decode()}")
                return

            # Convert bytes to int16 numpy array
            pcm_data = np.frombuffer(pcm_bytes, dtype=np.int16)

            # Calculate samples per frame (20ms at 48kHz = 960 samples)
            samples_per_frame = int(LIVEKIT_SAMPLE_RATE * FRAME_DURATION_MS / 1000)

            # Send audio in 20ms chunks
            for i in range(0, len(pcm_data), samples_per_frame):
                chunk = pcm_data[i:i + samples_per_frame]

                # Pad last frame if needed
                if len(chunk) < samples_per_frame:
                    chunk = np.pad(chunk, (0, samples_per_frame - len(chunk)), 'constant')

                # Create AudioFrame and capture
                frame = rtc.AudioFrame(
                    data=chunk.tobytes(),
                    sample_rate=LIVEKIT_SAMPLE_RATE,
                    num_channels=1,
                    samples_per_channel=samples_per_frame
                )

                await self.source.capture_frame(frame)

                # Small delay to maintain real-time playback
                await asyncio.sleep(FRAME_DURATION_MS / 1000)

            print(f"[AudioOutput] ✅ Played {len(audio_bytes)} bytes ({len(pcm_data)} samples)")

        except Exception as e:
            print(f"[AudioOutput] ❌ Playback error: {e}")
            import traceback
            traceback.print_exc()

    async def cleanup(self):
        """Cleanup resources."""
        if self.publication:
            try:
                await self.room.local_participant.unpublish_track(self.track.sid)
                print("[AudioOutput] 🧹 Track unpublished")
            except Exception as e:
                print(f"[AudioOutput] ⚠️ Cleanup error: {e}")


async def entrypoint(ctx: JobContext):
    """Minimal LiveKit entrypoint - audio bridge to Go gRPC."""

    print(f"[Adapter] 🚀 ENTRYPOINT CALLED - Starting for room: {ctx.room.name}", flush=True)

    audio_output = None
    channel = None
    session_id = ctx.room.name

    try:
        # Connect to room and wait for user
        print(f"[Adapter] Connecting to room...", flush=True)
        await ctx.connect()
        print(f"[Adapter] Connected! Waiting for participant...", flush=True)
        participant = await ctx.wait_for_participant()
        print(f"[Adapter] User connected: {participant.name}", flush=True)

        # Session identifiers
        session_id = ctx.room.name
        user_id = participant.identity

        # Initialize audio output manager
        print(f"[Adapter] Initializing audio output...", flush=True)
        audio_output = AudioOutputManager(ctx.room)
        await audio_output.initialize()

        # Connect to Go gRPC server
        print(f"[Adapter] Connecting to gRPC server: {GRPC_SERVER}", flush=True)
        channel = grpc.aio.insecure_channel(GRPC_SERVER)
        stub = voice_pb2_grpc.VoiceServiceStub(channel)
        print(f"[Adapter] ✅ Connected to gRPC server: {GRPC_SERVER}", flush=True)

        # Audio queue for user speech
        audio_queue: asyncio.Queue[bytes] = asyncio.Queue()
        sequence = 0

        # Subscribe to user audio track
        @ctx.room.on("track_subscribed")
        def on_track_subscribed(track: rtc.Track, *_):
            if track.kind == rtc.TrackKind.KIND_AUDIO:
                print(f"[Adapter] 🎤 User audio track subscribed", flush=True)
                audio_stream = rtc.AudioStream(track)
                asyncio.create_task(capture_audio(audio_stream, audio_queue))

        # Start bidirectional gRPC stream
        print(f"[Adapter] Starting gRPC bidirectional stream...", flush=True)
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
                    print(f"[Adapter] ❌ Send error: {e}", flush=True)
                    break

        # Task 2: Receive responses from Go
        async def receive_responses():
            async for response in grpc_stream:
                try:
                    # Handle different response types
                    if response.type == voice_pb2.ResponseType.TRANSCRIPT_USER:
                        print(f"[Adapter] 🎤 User: {response.text}", flush=True)
                        await send_to_frontend("user_transcript", response.text)

                    elif response.type == voice_pb2.ResponseType.TRANSCRIPT_AGENT:
                        print(f"[Adapter] 🤖 OSA: {response.text}", flush=True)
                        await send_to_frontend("agent_transcript", response.text)

                    elif response.type == voice_pb2.ResponseType.AUDIO:
                        # Play audio from Go TTS via LiveKit
                        print(f"[Adapter] 🔊 Playing audio: {len(response.audio_data)} bytes", flush=True)
                        await audio_output.play_audio_chunk(response.audio_data)

                    elif response.type == voice_pb2.ResponseType.STATE_UPDATE:
                        print(f"[Adapter] 🔄 State: {response.state}", flush=True)

                    elif response.type == voice_pb2.ResponseType.DONE:
                        print(f"[Adapter] ✅ Agent done speaking", flush=True)

                    elif response.type == voice_pb2.ResponseType.ERROR:
                        print(f"[Adapter] ❌ Error: {response.error}", flush=True)

                except Exception as e:
                    print(f"[Adapter] ❌ Receive error: {e}", flush=True)
                    import traceback
                    traceback.print_exc()
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
                print(f"[Adapter] ❌ Data send failed: {e}", flush=True)

        # Run both tasks concurrently
        print(f"[Adapter] Starting send/receive tasks...", flush=True)
        await asyncio.gather(send_frames(), receive_responses())

    except Exception as e:
        print(f"[Adapter] ❌ Fatal error: {e}", flush=True)
        import traceback
        traceback.print_exc()
    finally:
        print(f"[Adapter] Cleaning up...", flush=True)
        try:
            if audio_output:
                await audio_output.cleanup()
            if channel:
                await channel.close()
        except Exception as cleanup_err:
            print(f"[Adapter] Cleanup error: {cleanup_err}", flush=True)
        print(f"[Adapter] Session ended: {session_id}", flush=True)


async def capture_audio(stream: rtc.AudioStream, queue: asyncio.Queue):
    """Capture audio frames from LiveKit and queue for gRPC."""
    async for event in stream:
        # AudioStream yields AudioFrameEvent objects, not AudioFrame directly
        audio_bytes = bytes(event.frame.data)
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
