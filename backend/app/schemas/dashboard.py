from pydantic import BaseModel
from datetime import datetime
from uuid import UUID
from enum import Enum


# Task schemas
class TaskPriority(str, Enum):
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


class TaskStatus(str, Enum):
    TODO = "todo"
    IN_PROGRESS = "in_progress"
    DONE = "done"
    CANCELLED = "cancelled"


class TaskCreate(BaseModel):
    title: str
    description: str | None = None
    priority: TaskPriority = TaskPriority.MEDIUM
    due_date: datetime | None = None
    project_id: UUID | None = None
    assignee_id: UUID | None = None


class TaskUpdate(BaseModel):
    title: str | None = None
    description: str | None = None
    status: TaskStatus | None = None
    priority: TaskPriority | None = None
    due_date: datetime | None = None
    project_id: UUID | None = None
    assignee_id: UUID | None = None


class TaskResponse(BaseModel):
    id: UUID
    title: str
    description: str | None
    status: TaskStatus
    priority: TaskPriority
    due_date: datetime | None
    completed_at: datetime | None
    project_id: UUID | None
    assignee_id: UUID | None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DashboardTask(BaseModel):
    id: str
    title: str
    project_name: str | None = None
    due_date: str | None = None
    priority: TaskPriority
    completed: bool


# Focus Item schemas
class FocusItemCreate(BaseModel):
    text: str


class FocusItemUpdate(BaseModel):
    text: str | None = None
    completed: bool | None = None


class FocusItemResponse(BaseModel):
    id: UUID
    text: str
    completed: bool
    focus_date: datetime
    created_at: datetime

    class Config:
        from_attributes = True


# Dashboard-specific schemas
class DashboardProject(BaseModel):
    id: str
    name: str
    client_name: str | None = None
    project_type: str
    due_date: str | None = None
    progress: int
    health: str  # healthy, at_risk, critical
    team_count: int


class ActivityType(str, Enum):
    TASK_COMPLETED = "task_completed"
    TASK_STARTED = "task_started"
    PROJECT_CREATED = "project_created"
    PROJECT_UPDATED = "project_updated"
    CONVERSATION = "conversation"
    TEAM = "team"
    ARTIFACT = "artifact"


class DashboardActivity(BaseModel):
    id: str
    type: ActivityType
    description: str
    actor_name: str | None = None
    actor_avatar: str | None = None
    target_id: str | None = None
    target_type: str | None = None
    created_at: str


class DashboardSummary(BaseModel):
    focus_items: list[FocusItemResponse]
    tasks: list[DashboardTask]
    projects: list[DashboardProject]
    activities: list[DashboardActivity]
    energy_level: int | None = None
