from typing import Annotated
from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.models.user_settings import UserSettings
from app.schemas.user_settings import (
    UserSettingsUpdate,
    UserSettingsResponse,
    SystemInfoResponse,
    AvailableModel,
)
from app.utils.auth import CurrentUser
from app.config import get_settings

router = APIRouter(prefix="/api/settings", tags=["settings"])

settings = get_settings()


@router.get("", response_model=UserSettingsResponse)
async def get_user_settings(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get current user's settings (creates defaults if not exists)"""
    result = await db.execute(
        select(UserSettings).where(UserSettings.user_id == current_user.id)
    )
    user_settings = result.scalar_one_or_none()

    if not user_settings:
        # Create default settings for this user
        user_settings = UserSettings(
            user_id=current_user.id,
            default_model=settings.default_model,
        )
        db.add(user_settings)
        await db.commit()
        await db.refresh(user_settings)

    return user_settings


@router.put("", response_model=UserSettingsResponse)
async def update_user_settings(
    data: UserSettingsUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update current user's settings"""
    result = await db.execute(
        select(UserSettings).where(UserSettings.user_id == current_user.id)
    )
    user_settings = result.scalar_one_or_none()

    if not user_settings:
        # Create settings with provided data
        user_settings = UserSettings(
            user_id=current_user.id,
            default_model=data.default_model or settings.default_model,
            email_notifications=data.email_notifications if data.email_notifications is not None else True,
            daily_summary=data.daily_summary if data.daily_summary is not None else False,
            theme=data.theme or "light",
            sidebar_collapsed=data.sidebar_collapsed if data.sidebar_collapsed is not None else False,
            share_analytics=data.share_analytics if data.share_analytics is not None else True,
            custom_settings=data.custom_settings,
        )
        db.add(user_settings)
    else:
        # Update existing settings
        if data.default_model is not None:
            user_settings.default_model = data.default_model
        if data.email_notifications is not None:
            user_settings.email_notifications = data.email_notifications
        if data.daily_summary is not None:
            user_settings.daily_summary = data.daily_summary
        if data.theme is not None:
            user_settings.theme = data.theme
        if data.sidebar_collapsed is not None:
            user_settings.sidebar_collapsed = data.sidebar_collapsed
        if data.share_analytics is not None:
            user_settings.share_analytics = data.share_analytics
        if data.custom_settings is not None:
            user_settings.custom_settings = data.custom_settings
        user_settings.updated_at = datetime.utcnow()

    await db.commit()
    await db.refresh(user_settings)
    return user_settings


@router.get("/system", response_model=SystemInfoResponse)
async def get_system_info(
    current_user: CurrentUser,
):
    """Get system information including available models"""
    # Available models - in the future this could be fetched from Ollama API
    available_models = [
        AvailableModel(
            name="qwen2.5:480b",
            display_name="Qwen 2.5 480B",
            provider="ollama",
            description="High capability model for complex tasks",
        ),
        AvailableModel(
            name="qwen2.5:72b",
            display_name="Qwen 2.5 72B",
            provider="ollama",
            description="Balanced performance and speed",
        ),
        AvailableModel(
            name="qwen2.5:32b",
            display_name="Qwen 2.5 32B",
            provider="ollama",
            description="Fast model for simple tasks",
        ),
        AvailableModel(
            name="llama3.2:latest",
            display_name="Llama 3.2",
            provider="ollama",
            description="Meta's latest open model",
        ),
        AvailableModel(
            name="deepseek-r1:latest",
            display_name="DeepSeek R1",
            provider="ollama",
            description="Reasoning-focused model",
        ),
    ]

    return SystemInfoResponse(
        ollama_mode=settings.ollama_mode,
        available_models=available_models,
        default_model=settings.default_model,
    )
