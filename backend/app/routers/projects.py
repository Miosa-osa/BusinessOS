from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.models.project import Project, ProjectNote
from app.schemas.project import (
    ProjectCreate,
    ProjectUpdate,
    ProjectResponse,
    ProjectNoteCreate,
    ProjectNoteResponse,
)
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/projects", tags=["projects"])


@router.get("", response_model=list[ProjectResponse])
async def list_projects(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    status_filter: str | None = None,
    skip: int = 0,
    limit: int = 50,
):
    query = select(Project).where(Project.user_id == current_user.id)

    if status_filter:
        query = query.where(Project.status == status_filter)

    result = await db.execute(
        query.order_by(Project.updated_at.desc())
        .offset(skip)
        .limit(limit)
        .options(selectinload(Project.notes))
    )
    return result.scalars().all()


@router.post("", response_model=ProjectResponse)
async def create_project(
    data: ProjectCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    project = Project(
        user_id=current_user.id,
        name=data.name,
        description=data.description,
        status=data.status,
        priority=data.priority,
        client_name=data.client_name,
        project_type=data.project_type,
        project_metadata=data.project_metadata,
    )
    db.add(project)
    await db.commit()

    # Re-fetch with notes eagerly loaded for response serialization
    result = await db.execute(
        select(Project)
        .where(Project.id == project.id)
        .options(selectinload(Project.notes))
    )
    return result.scalar_one()


@router.get("/{project_id}", response_model=ProjectResponse)
async def get_project(
    project_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    result = await db.execute(
        select(Project)
        .where(
            Project.id == project_id,
            Project.user_id == current_user.id,
        )
        .options(selectinload(Project.notes))
    )
    project = result.scalar_one_or_none()

    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Project not found",
        )
    return project


@router.put("/{project_id}", response_model=ProjectResponse)
async def update_project(
    project_id: UUID,
    data: ProjectUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    result = await db.execute(
        select(Project).where(
            Project.id == project_id,
            Project.user_id == current_user.id,
        )
    )
    project = result.scalar_one_or_none()

    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Project not found",
        )

    update_data = data.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(project, field, value)

    await db.commit()

    # Re-fetch with notes eagerly loaded for response serialization
    result = await db.execute(
        select(Project)
        .where(Project.id == project_id)
        .options(selectinload(Project.notes))
    )
    return result.scalar_one()


@router.delete("/{project_id}")
async def delete_project(
    project_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    result = await db.execute(
        select(Project).where(
            Project.id == project_id,
            Project.user_id == current_user.id,
        )
    )
    project = result.scalar_one_or_none()

    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Project not found",
        )

    await db.delete(project)
    await db.commit()
    return {"message": "Project deleted"}


@router.post("/{project_id}/notes", response_model=ProjectNoteResponse)
async def add_project_note(
    project_id: UUID,
    data: ProjectNoteCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    # Verify project ownership
    result = await db.execute(
        select(Project).where(
            Project.id == project_id,
            Project.user_id == current_user.id,
        )
    )
    if not result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Project not found",
        )

    note = ProjectNote(
        project_id=project_id,
        content=data.content,
    )
    db.add(note)
    await db.commit()
    await db.refresh(note)
    return note
