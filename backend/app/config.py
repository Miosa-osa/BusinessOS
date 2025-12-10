from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    # Database
    database_url: str = "postgresql+asyncpg://postgres:password@localhost:5432/business_os"

    # JWT Auth
    secret_key: str = "your-secret-key-change-this-in-production"
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 1440  # 24 hours

    # Ollama Configuration
    ollama_mode: str = "local"  # "local" or "cloud"
    ollama_local_url: str = "http://localhost:11434"
    ollama_cloud_url: str = "https://api.ollama.com/v1"
    ollama_cloud_api_key: str = ""

    # Default Model
    default_model: str = "qwen3-coder:480b"

    # Redis
    redis_url: str = "redis://localhost:6379/0"

    # Supermemory
    supermemory_api_key: str = ""

    class Config:
        env_file = ".env"
        extra = "ignore"


@lru_cache()
def get_settings() -> Settings:
    return Settings()
