from datetime import datetime
from uuid import UUID
from pydantic import BaseModel, EmailStr
from typing import Optional


class TeamMemberCreate(BaseModel):
    name: str
    email: EmailStr
    role: str
    avatar_url: Optional[str] = None
    manager_id: Optional[UUID] = None
    skills: Optional[list[str]] = None
    hourly_rate: Optional[float] = None


class TeamMemberUpdate(BaseModel):
    name: Optional[str] = None
    email: Optional[EmailStr] = None
    role: Optional[str] = None
    avatar_url: Optional[str] = None
    status: Optional[str] = None
    capacity: Optional[int] = None
    manager_id: Optional[UUID] = None
    skills: Optional[list[str]] = None
    hourly_rate: Optional[float] = None


class TeamMemberActivityResponse(BaseModel):
    id: UUID
    activity_type: str
    description: str
    created_at: datetime

    class Config:
        from_attributes = True


class TeamMemberResponse(BaseModel):
    id: UUID
    name: str
    email: str
    role: str
    avatar_url: Optional[str] = None
    status: str
    capacity: int
    manager_id: Optional[UUID] = None
    skills: Optional[list[str]] = None
    hourly_rate: Optional[float] = None
    joined_at: datetime
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class TeamMemberDetailResponse(TeamMemberResponse):
    """Extended response with related data"""
    active_projects: int = 0
    open_tasks: int = 0
    activities: list[TeamMemberActivityResponse] = []


class TeamMemberListResponse(BaseModel):
    id: UUID
    name: str
    email: str
    role: str
    avatar_url: Optional[str] = None
    status: str
    capacity: int
    manager_id: Optional[UUID] = None
    active_projects: int = 0
    open_tasks: int = 0
    joined_at: datetime

    class Config:
        from_attributes = True
