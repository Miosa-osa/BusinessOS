"""
Configuration Management for OSA Voice Agent

Loads environment variables and provides configuration access.
"""

import os


class Config:
    """Voice agent configuration loaded from environment variables."""

    def __init__(self):
        # LiveKit credentials
        self.livekit_url: str = os.getenv("LIVEKIT_URL", "")
        self.livekit_api_key: str = os.getenv("LIVEKIT_API_KEY", "")
        self.livekit_api_secret: str = os.getenv("LIVEKIT_API_SECRET", "")

        # GROQ API (STT + LLM)
        self.groq_api_key: str = os.getenv("GROQ_API_KEY", "")

        # ElevenLabs (TTS)
        self.elevenlabs_api_key: str = os.getenv("ELEVENLABS_API_KEY", "")
        self.elevenlabs_voice_id: str = os.getenv(
            "ELEVENLABS_VOICE_ID",
            "KoVIHoyLDrQyd4pGalbs"  # OSA's voice ID
        )

        # Go backend URL (for fetching context and executing tools)
        self.go_backend_url: str = os.getenv("GO_BACKEND_URL", "http://localhost:8001")

        # Voice agent settings
        self.stt_model: str = "whisper-large-v3"  # Groq Whisper
        self.llm_model: str = "llama-3.1-8b-instant"  # Groq Llama 3.1 8B
        self.tts_model: str = "eleven_turbo_v2_5"  # ElevenLabs Turbo v2.5

    def validate(self) -> bool:
        """Validate that all required configuration is present."""
        required = [
            ("LIVEKIT_URL", self.livekit_url),
            ("LIVEKIT_API_KEY", self.livekit_api_key),
            ("LIVEKIT_API_SECRET", self.livekit_api_secret),
            ("GROQ_API_KEY", self.groq_api_key),
            ("ELEVENLABS_API_KEY", self.elevenlabs_api_key),
        ]

        missing = [name for name, value in required if not value]

        if missing:
            raise ValueError(
                f"Missing required environment variables: {', '.join(missing)}"
            )

        return True

    def __repr__(self) -> str:
        """Safe repr that doesn't expose secrets."""
        return (
            f"Config("
            f"livekit_url={self.livekit_url}, "
            f"go_backend_url={self.go_backend_url}, "
            f"stt_model={self.stt_model}, "
            f"llm_model={self.llm_model}, "
            f"tts_model={self.tts_model}"
            f")"
        )


# Global config instance
config = Config()
