from typing import Annotated
from uuid import UUID
from datetime import datetime, date, timedelta

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select, func, and_
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.models.task import Task, FocusItem, TaskStatus, TaskPriority as TaskPriorityModel
from app.models.project import Project, ProjectStatus
from app.models.conversation import Conversation, Message
from app.models.team_member import TeamMember, TeamMemberActivity
from app.schemas.dashboard import (
    TaskCreate,
    TaskUpdate,
    TaskResponse,
    FocusItemCreate,
    FocusItemUpdate,
    FocusItemResponse,
    DashboardSummary,
    DashboardTask,
    DashboardProject,
    DashboardActivity,
    ActivityType,
    TaskPriority,
)
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/dashboard", tags=["dashboard"])


# ============== Focus Items ==============

@router.get("/focus", response_model=list[FocusItemResponse])
async def list_focus_items(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    date_filter: date | None = None,
):
    """Get focus items for today (or specified date)"""
    target_date = date_filter or date.today()
    start_of_day = datetime.combine(target_date, datetime.min.time())
    end_of_day = datetime.combine(target_date, datetime.max.time())

    result = await db.execute(
        select(FocusItem)
        .where(
            FocusItem.user_id == current_user.id,
            FocusItem.focus_date >= start_of_day,
            FocusItem.focus_date <= end_of_day,
        )
        .order_by(FocusItem.created_at)
    )
    return result.scalars().all()


@router.post("/focus", response_model=FocusItemResponse)
async def create_focus_item(
    data: FocusItemCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new focus item for today"""
    item = FocusItem(
        user_id=current_user.id,
        text=data.text,
        focus_date=datetime.utcnow(),
    )
    db.add(item)
    await db.commit()
    await db.refresh(item)
    return item


@router.put("/focus/{item_id}", response_model=FocusItemResponse)
async def update_focus_item(
    item_id: UUID,
    data: FocusItemUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a focus item"""
    result = await db.execute(
        select(FocusItem).where(
            FocusItem.id == item_id,
            FocusItem.user_id == current_user.id,
        )
    )
    item = result.scalar_one_or_none()

    if not item:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Focus item not found",
        )

    if data.text is not None:
        item.text = data.text
    if data.completed is not None:
        item.completed = data.completed

    await db.commit()
    await db.refresh(item)
    return item


@router.delete("/focus/{item_id}")
async def delete_focus_item(
    item_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete a focus item"""
    result = await db.execute(
        select(FocusItem).where(
            FocusItem.id == item_id,
            FocusItem.user_id == current_user.id,
        )
    )
    item = result.scalar_one_or_none()

    if not item:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Focus item not found",
        )

    await db.delete(item)
    await db.commit()
    return {"message": "Focus item deleted"}


# ============== Tasks ==============

@router.get("/tasks", response_model=list[TaskResponse])
async def list_tasks(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    status_filter: str | None = None,
    priority_filter: str | None = None,
    project_id: UUID | None = None,
    skip: int = 0,
    limit: int = 50,
):
    """List all tasks for current user"""
    query = select(Task).where(Task.user_id == current_user.id)

    if status_filter:
        query = query.where(Task.status == status_filter)
    if priority_filter:
        query = query.where(Task.priority == priority_filter)
    if project_id:
        query = query.where(Task.project_id == project_id)

    result = await db.execute(
        query.order_by(Task.due_date.asc().nullsfirst(), Task.created_at.desc())
        .offset(skip)
        .limit(limit)
    )
    return result.scalars().all()


@router.post("/tasks", response_model=TaskResponse)
async def create_task(
    data: TaskCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new task"""
    task = Task(
        user_id=current_user.id,
        title=data.title,
        description=data.description,
        priority=TaskPriorityModel(data.priority.value),
        due_date=data.due_date,
        project_id=data.project_id,
        assignee_id=data.assignee_id,
    )
    db.add(task)
    await db.commit()
    await db.refresh(task)
    return task


@router.put("/tasks/{task_id}", response_model=TaskResponse)
async def update_task(
    task_id: UUID,
    data: TaskUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a task"""
    result = await db.execute(
        select(Task).where(
            Task.id == task_id,
            Task.user_id == current_user.id,
        )
    )
    task = result.scalar_one_or_none()

    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Task not found",
        )

    update_data = data.model_dump(exclude_unset=True)

    # Handle status change to done
    if "status" in update_data:
        if update_data["status"] == TaskStatus.DONE and task.status != TaskStatus.DONE:
            task.completed_at = datetime.utcnow()
        elif update_data["status"] != TaskStatus.DONE:
            task.completed_at = None
        update_data["status"] = TaskStatus(update_data["status"])

    if "priority" in update_data:
        update_data["priority"] = TaskPriorityModel(update_data["priority"])

    for field, value in update_data.items():
        setattr(task, field, value)

    await db.commit()
    await db.refresh(task)
    return task


@router.delete("/tasks/{task_id}")
async def delete_task(
    task_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete a task"""
    result = await db.execute(
        select(Task).where(
            Task.id == task_id,
            Task.user_id == current_user.id,
        )
    )
    task = result.scalar_one_or_none()

    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Task not found",
        )

    await db.delete(task)
    await db.commit()
    return {"message": "Task deleted"}


@router.post("/tasks/{task_id}/toggle", response_model=TaskResponse)
async def toggle_task(
    task_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Toggle task completion status"""
    result = await db.execute(
        select(Task).where(
            Task.id == task_id,
            Task.user_id == current_user.id,
        )
    )
    task = result.scalar_one_or_none()

    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Task not found",
        )

    if task.status == TaskStatus.DONE:
        task.status = TaskStatus.TODO
        task.completed_at = None
    else:
        task.status = TaskStatus.DONE
        task.completed_at = datetime.utcnow()

    await db.commit()
    await db.refresh(task)
    return task


# ============== Dashboard Summary ==============

@router.get("/summary", response_model=DashboardSummary)
async def get_dashboard_summary(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get aggregated dashboard data"""
    today = date.today()
    start_of_day = datetime.combine(today, datetime.min.time())
    end_of_day = datetime.combine(today, datetime.max.time())
    next_week = datetime.combine(today + timedelta(days=7), datetime.max.time())

    # Get focus items for today
    focus_result = await db.execute(
        select(FocusItem)
        .where(
            FocusItem.user_id == current_user.id,
            FocusItem.focus_date >= start_of_day,
            FocusItem.focus_date <= end_of_day,
        )
        .order_by(FocusItem.created_at)
    )
    focus_items = focus_result.scalars().all()

    # Get tasks due soon (incomplete only)
    tasks_result = await db.execute(
        select(Task)
        .where(
            Task.user_id == current_user.id,
            Task.status != TaskStatus.DONE,
            Task.status != TaskStatus.CANCELLED,
        )
        .options(selectinload(Task.project))
        .order_by(Task.due_date.asc().nullsfirst())
        .limit(20)
    )
    tasks_raw = tasks_result.scalars().all()

    # Transform tasks for dashboard
    dashboard_tasks = []
    for task in tasks_raw:
        dashboard_tasks.append(DashboardTask(
            id=str(task.id),
            title=task.title,
            project_name=task.project.name if task.project else None,
            due_date=task.due_date.isoformat() if task.due_date else None,
            priority=TaskPriority(task.priority.value),
            completed=task.status == TaskStatus.DONE,
        ))

    # Get active projects
    projects_result = await db.execute(
        select(Project)
        .where(
            Project.user_id == current_user.id,
            Project.status == ProjectStatus.ACTIVE,
        )
        .order_by(Project.updated_at.desc())
        .limit(10)
    )
    projects_raw = projects_result.scalars().all()

    # Get team member count per project (simplified - just return 1 for now)
    dashboard_projects = []
    for project in projects_raw:
        # Calculate health based on priority and metadata
        health = "healthy"
        if project.priority.value == "critical":
            health = "critical"
        elif project.priority.value == "high":
            health = "at_risk"

        # Progress from metadata or default
        progress = 0
        if project.project_metadata and "progress" in project.project_metadata:
            progress = project.project_metadata["progress"]

        # Due date from metadata
        due_date = None
        if project.project_metadata and "due_date" in project.project_metadata:
            due_date = project.project_metadata["due_date"]

        dashboard_projects.append(DashboardProject(
            id=str(project.id),
            name=project.name,
            client_name=project.client_name,
            project_type=project.project_type,
            due_date=due_date,
            progress=progress,
            health=health,
            team_count=1,  # Simplified for now
        ))

    # Get recent activities (from conversations and team activities)
    activities: list[DashboardActivity] = []

    # Recent conversations
    conv_result = await db.execute(
        select(Conversation)
        .where(Conversation.user_id == current_user.id)
        .order_by(Conversation.updated_at.desc())
        .limit(5)
    )
    conversations = conv_result.scalars().all()
    for conv in conversations:
        activities.append(DashboardActivity(
            id=str(conv.id),
            type=ActivityType.CONVERSATION,
            description=f"updated conversation \"{conv.title}\"",
            actor_name=None,
            target_id=str(conv.id),
            target_type="conversation",
            created_at=conv.updated_at.isoformat(),
        ))

    # Recent team activities
    team_result = await db.execute(
        select(TeamMemberActivity)
        .join(TeamMember)
        .where(TeamMember.user_id == current_user.id)
        .options(selectinload(TeamMemberActivity.member))
        .order_by(TeamMemberActivity.created_at.desc())
        .limit(5)
    )
    team_activities = team_result.scalars().all()
    for activity in team_activities:
        activities.append(DashboardActivity(
            id=str(activity.id),
            type=ActivityType.TEAM,
            description=activity.description,
            actor_name=activity.member.name if activity.member else None,
            target_id=str(activity.member_id),
            target_type="team_member",
            created_at=activity.created_at.isoformat(),
        ))

    # Sort activities by date
    activities.sort(key=lambda a: a.created_at, reverse=True)
    activities = activities[:10]

    return DashboardSummary(
        focus_items=focus_items,
        tasks=dashboard_tasks,
        projects=dashboard_projects,
        activities=activities,
    )
