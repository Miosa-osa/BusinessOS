from datetime import datetime
from uuid import UUID
from pydantic import BaseModel


class UserSettingsBase(BaseModel):
    default_model: str | None = None
    email_notifications: bool = True
    daily_summary: bool = False
    theme: str = "light"
    sidebar_collapsed: bool = False
    share_analytics: bool = True
    custom_settings: dict | None = None


class UserSettingsUpdate(BaseModel):
    default_model: str | None = None
    email_notifications: bool | None = None
    daily_summary: bool | None = None
    theme: str | None = None
    sidebar_collapsed: bool | None = None
    share_analytics: bool | None = None
    custom_settings: dict | None = None


class UserSettingsResponse(UserSettingsBase):
    id: UUID
    user_id: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AvailableModel(BaseModel):
    name: str
    display_name: str
    provider: str
    description: str | None = None


class SystemInfoResponse(BaseModel):
    ollama_mode: str
    available_models: list[AvailableModel]
    default_model: str
