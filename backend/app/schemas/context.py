from pydantic import BaseModel
from datetime import datetime
from uuid import UUID
from app.models.context import ContextType


class ContextCreate(BaseModel):
    name: str
    type: ContextType = ContextType.CUSTOM
    content: str | None = None
    structured_data: dict | None = None
    system_prompt_template: str | None = None


class ContextUpdate(BaseModel):
    name: str | None = None
    type: ContextType | None = None
    content: str | None = None
    structured_data: dict | None = None
    system_prompt_template: str | None = None


class ContextResponse(BaseModel):
    id: UUID
    name: str
    type: ContextType
    content: str | None
    structured_data: dict | None
    system_prompt_template: str | None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True
