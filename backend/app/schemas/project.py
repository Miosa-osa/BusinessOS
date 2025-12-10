from pydantic import BaseModel
from datetime import datetime
from uuid import UUID
from app.models.project import ProjectStatus, ProjectPriority


class ProjectCreate(BaseModel):
    name: str
    description: str | None = None
    status: ProjectStatus = ProjectStatus.ACTIVE
    priority: ProjectPriority = ProjectPriority.MEDIUM
    client_name: str | None = None
    project_type: str = "internal"
    project_metadata: dict | None = None


class ProjectUpdate(BaseModel):
    name: str | None = None
    description: str | None = None
    status: ProjectStatus | None = None
    priority: ProjectPriority | None = None
    client_name: str | None = None
    project_type: str | None = None
    project_metadata: dict | None = None


class ProjectNoteCreate(BaseModel):
    content: str


class ProjectNoteResponse(BaseModel):
    id: UUID
    content: str
    created_at: datetime

    class Config:
        from_attributes = True


class ProjectResponse(BaseModel):
    id: UUID
    name: str
    description: str | None
    status: ProjectStatus
    priority: ProjectPriority
    client_name: str | None
    project_type: str
    project_metadata: dict | None
    created_at: datetime
    updated_at: datetime
    notes: list[ProjectNoteResponse] = []

    class Config:
        from_attributes = True
