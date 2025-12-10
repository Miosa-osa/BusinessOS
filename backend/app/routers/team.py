from typing import Annotated
from uuid import UUID
from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.models.team_member import TeamMember, TeamMemberActivity, MemberStatus
from app.models.task import Task, TaskStatus
from app.schemas.team_member import (
    TeamMemberCreate,
    TeamMemberUpdate,
    TeamMemberResponse,
    TeamMemberDetailResponse,
    TeamMemberListResponse,
    TeamMemberActivityResponse,
)
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/team", tags=["team"])


@router.get("", response_model=list[TeamMemberListResponse])
async def list_team_members(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    status_filter: str | None = None,
    skip: int = 0,
    limit: int = 100,
):
    """List all team members for the current user's workspace"""
    query = select(TeamMember).where(TeamMember.user_id == current_user.id)

    if status_filter:
        query = query.where(TeamMember.status == status_filter)

    result = await db.execute(
        query.order_by(TeamMember.name.asc())
        .offset(skip)
        .limit(limit)
    )
    members = result.scalars().all()

    # Get member IDs
    member_ids = [m.id for m in members]

    # Count open tasks per member (todo or in_progress)
    open_task_counts = {}
    active_project_counts = {}

    if member_ids:
        # Count open tasks
        task_count_query = (
            select(Task.assignee_id, func.count(Task.id).label("task_count"))
            .where(
                Task.assignee_id.in_(member_ids),
                Task.status.in_([TaskStatus.TODO, TaskStatus.IN_PROGRESS]),
            )
            .group_by(Task.assignee_id)
        )
        task_result = await db.execute(task_count_query)
        for row in task_result:
            open_task_counts[row.assignee_id] = row.task_count

        # Count unique projects with active tasks
        project_count_query = (
            select(Task.assignee_id, func.count(func.distinct(Task.project_id)).label("project_count"))
            .where(
                Task.assignee_id.in_(member_ids),
                Task.status.in_([TaskStatus.TODO, TaskStatus.IN_PROGRESS]),
                Task.project_id.isnot(None),
            )
            .group_by(Task.assignee_id)
        )
        project_result = await db.execute(project_count_query)
        for row in project_result:
            active_project_counts[row.assignee_id] = row.project_count

    return [
        TeamMemberListResponse(
            id=m.id,
            name=m.name,
            email=m.email,
            role=m.role,
            avatar_url=m.avatar_url,
            status=m.status.value,
            capacity=m.capacity,
            manager_id=m.manager_id,
            active_projects=active_project_counts.get(m.id, 0),
            open_tasks=open_task_counts.get(m.id, 0),
            joined_at=m.joined_at,
        )
        for m in members
    ]


@router.post("", response_model=TeamMemberResponse)
async def create_team_member(
    data: TeamMemberCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Add a new team member to the workspace"""
    # Check if manager exists if provided
    if data.manager_id:
        manager_result = await db.execute(
            select(TeamMember).where(
                TeamMember.id == data.manager_id,
                TeamMember.user_id == current_user.id,
            )
        )
        if not manager_result.scalar_one_or_none():
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Manager not found",
            )

    member = TeamMember(
        user_id=current_user.id,
        name=data.name,
        email=data.email,
        role=data.role,
        avatar_url=data.avatar_url,
        manager_id=data.manager_id,
        skills=data.skills or [],
        hourly_rate=data.hourly_rate,
        status=MemberStatus.AVAILABLE,
        capacity=0,
    )
    db.add(member)

    # Add initial activity
    activity = TeamMemberActivity(
        member_id=member.id,
        activity_type="joined",
        description="Joined the team",
    )
    db.add(activity)

    await db.commit()
    await db.refresh(member)
    return member


@router.get("/{member_id}", response_model=TeamMemberDetailResponse)
async def get_team_member(
    member_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get detailed info about a specific team member"""
    result = await db.execute(
        select(TeamMember).where(
            TeamMember.id == member_id,
            TeamMember.user_id == current_user.id,
        )
    )
    member = result.scalar_one_or_none()

    if not member:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Team member not found",
        )

    # Get recent activities
    activities_result = await db.execute(
        select(TeamMemberActivity)
        .where(TeamMemberActivity.member_id == member_id)
        .order_by(TeamMemberActivity.created_at.desc())
        .limit(10)
    )
    activities = activities_result.scalars().all()

    # Count open tasks for this member
    task_count_result = await db.execute(
        select(func.count(Task.id))
        .where(
            Task.assignee_id == member_id,
            Task.status.in_([TaskStatus.TODO, TaskStatus.IN_PROGRESS]),
        )
    )
    open_tasks = task_count_result.scalar() or 0

    # Count unique projects with active tasks
    project_count_result = await db.execute(
        select(func.count(func.distinct(Task.project_id)))
        .where(
            Task.assignee_id == member_id,
            Task.status.in_([TaskStatus.TODO, TaskStatus.IN_PROGRESS]),
            Task.project_id.isnot(None),
        )
    )
    active_projects = project_count_result.scalar() or 0

    return TeamMemberDetailResponse(
        id=member.id,
        name=member.name,
        email=member.email,
        role=member.role,
        avatar_url=member.avatar_url,
        status=member.status.value,
        capacity=member.capacity,
        manager_id=member.manager_id,
        skills=member.skills or [],
        hourly_rate=float(member.hourly_rate) if member.hourly_rate else None,
        joined_at=member.joined_at,
        created_at=member.created_at,
        updated_at=member.updated_at,
        active_projects=active_projects,
        open_tasks=open_tasks,
        activities=[
            TeamMemberActivityResponse(
                id=a.id,
                activity_type=a.activity_type,
                description=a.description,
                created_at=a.created_at,
            )
            for a in activities
        ],
    )


@router.put("/{member_id}", response_model=TeamMemberResponse)
async def update_team_member(
    member_id: UUID,
    data: TeamMemberUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a team member's information"""
    result = await db.execute(
        select(TeamMember).where(
            TeamMember.id == member_id,
            TeamMember.user_id == current_user.id,
        )
    )
    member = result.scalar_one_or_none()

    if not member:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Team member not found",
        )

    # Check if new manager exists if provided
    if data.manager_id:
        manager_result = await db.execute(
            select(TeamMember).where(
                TeamMember.id == data.manager_id,
                TeamMember.user_id == current_user.id,
            )
        )
        if not manager_result.scalar_one_or_none():
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Manager not found",
            )

    update_data = data.model_dump(exclude_unset=True)

    # Handle status enum conversion
    if "status" in update_data:
        update_data["status"] = MemberStatus(update_data["status"])

    for field, value in update_data.items():
        setattr(member, field, value)

    await db.commit()
    await db.refresh(member)
    return member


@router.delete("/{member_id}")
async def delete_team_member(
    member_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Remove a team member from the workspace"""
    result = await db.execute(
        select(TeamMember).where(
            TeamMember.id == member_id,
            TeamMember.user_id == current_user.id,
        )
    )
    member = result.scalar_one_or_none()

    if not member:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Team member not found",
        )

    # Update reports to have no manager
    reports_result = await db.execute(
        select(TeamMember).where(TeamMember.manager_id == member_id)
    )
    for report in reports_result.scalars():
        report.manager_id = None

    await db.delete(member)
    await db.commit()
    return {"message": "Team member removed"}


@router.post("/{member_id}/activity", response_model=TeamMemberActivityResponse)
async def add_member_activity(
    member_id: UUID,
    activity_type: str,
    description: str,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Add an activity entry for a team member"""
    # Verify member exists
    result = await db.execute(
        select(TeamMember).where(
            TeamMember.id == member_id,
            TeamMember.user_id == current_user.id,
        )
    )
    if not result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Team member not found",
        )

    activity = TeamMemberActivity(
        member_id=member_id,
        activity_type=activity_type,
        description=description,
    )
    db.add(activity)
    await db.commit()
    await db.refresh(activity)
    return activity


@router.patch("/{member_id}/status", response_model=TeamMemberResponse)
async def update_member_status(
    member_id: UUID,
    new_status: str,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Quick update for member status"""
    result = await db.execute(
        select(TeamMember).where(
            TeamMember.id == member_id,
            TeamMember.user_id == current_user.id,
        )
    )
    member = result.scalar_one_or_none()

    if not member:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Team member not found",
        )

    try:
        member.status = MemberStatus(new_status)
    except ValueError:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Invalid status. Must be one of: {[s.value for s in MemberStatus]}",
        )

    await db.commit()
    await db.refresh(member)
    return member


@router.patch("/{member_id}/capacity", response_model=TeamMemberResponse)
async def update_member_capacity(
    member_id: UUID,
    capacity: int,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Quick update for member capacity"""
    if not 0 <= capacity <= 100:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Capacity must be between 0 and 100",
        )

    result = await db.execute(
        select(TeamMember).where(
            TeamMember.id == member_id,
            TeamMember.user_id == current_user.id,
        )
    )
    member = result.scalar_one_or_none()

    if not member:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Team member not found",
        )

    member.capacity = capacity

    # Auto-update status based on capacity
    if capacity < 70:
        member.status = MemberStatus.AVAILABLE
    elif capacity < 90:
        member.status = MemberStatus.BUSY
    else:
        member.status = MemberStatus.OVERLOADED

    await db.commit()
    await db.refresh(member)
    return member
