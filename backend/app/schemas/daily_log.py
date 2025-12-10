from datetime import date as DateType, datetime
from typing import Optional
from uuid import UUID
from pydantic import BaseModel


class DailyLogCreate(BaseModel):
    content: str
    energy_level: Optional[int] = None
    date: Optional[DateType] = None  # Defaults to today


class DailyLogUpdate(BaseModel):
    content: Optional[str] = None
    energy_level: Optional[int] = None
    extracted_actions: Optional[dict] = None


class DailyLogResponse(BaseModel):
    id: UUID
    date: DateType
    content: str
    energy_level: Optional[int]
    extracted_actions: Optional[dict]
    extracted_patterns: Optional[dict]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True
