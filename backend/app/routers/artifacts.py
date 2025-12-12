"""
Artifacts Router for Business OS

API endpoints for managing AI-generated artifacts:
- Proposals, SOPs, Frameworks, Agendas, Reports, Plans
"""

from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.models.artifact import Artifact, ArtifactType
from app.utils.auth import CurrentUser
from pydantic import BaseModel, Field


router = APIRouter(prefix="/api/artifacts", tags=["artifacts"])


# Pydantic schemas
class ArtifactCreate(BaseModel):
    title: str = Field(..., max_length=255)
    content: str
    type: str = Field(default="other")
    summary: str | None = Field(None, max_length=500)
    conversation_id: UUID | None = None
    project_id: UUID | None = None
    context_id: UUID | None = None


class ArtifactUpdate(BaseModel):
    title: str | None = Field(None, max_length=255)
    content: str | None = None
    summary: str | None = Field(None, max_length=500)
    project_id: UUID | None = None
    context_id: UUID | None = None


class ArtifactLinkUpdate(BaseModel):
    """Schema for linking artifact to a project or context"""
    project_id: UUID | None = None
    context_id: UUID | None = None


class ArtifactResponse(BaseModel):
    id: UUID
    title: str
    type: str
    content: str
    summary: str | None
    conversation_id: UUID | None
    project_id: UUID | None
    context_id: UUID | None
    context_name: str | None = None
    version: int
    created_at: str
    updated_at: str

    class Config:
        from_attributes = True


class ArtifactListItem(BaseModel):
    id: UUID
    title: str
    type: str
    summary: str | None
    conversation_id: UUID | None
    project_id: UUID | None
    context_id: UUID | None
    context_name: str | None = None
    created_at: str
    updated_at: str

    class Config:
        from_attributes = True


@router.get("", response_model=list[ArtifactListItem])
async def list_artifacts(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    type: str | None = None,
    conversation_id: UUID | None = None,
    project_id: UUID | None = None,
    context_id: UUID | None = None,
    unassigned_only: bool = False,
    skip: int = 0,
    limit: int = 50,
):
    """List all artifacts for the current user.

    Args:
        unassigned_only: If True, only return artifacts not linked to any project or context
    """
    from sqlalchemy.orm import selectinload

    query = select(Artifact).where(Artifact.user_id == current_user.id)

    if type:
        # Convert type string to enum (handle both uppercase and lowercase)
        try:
            artifact_type = ArtifactType(type.lower())
            query = query.where(Artifact.type == artifact_type)
        except ValueError:
            pass  # Invalid type, skip filter
    if conversation_id:
        query = query.where(Artifact.conversation_id == conversation_id)
    if project_id:
        query = query.where(Artifact.project_id == project_id)
    if context_id:
        query = query.where(Artifact.context_id == context_id)
    if unassigned_only:
        query = query.where(Artifact.project_id == None, Artifact.context_id == None)

    # Eager load context relationship for context_name
    query = query.options(selectinload(Artifact.context))
    query = query.order_by(Artifact.created_at.desc()).offset(skip).limit(limit)

    result = await db.execute(query)
    artifacts = result.scalars().all()

    return [
        ArtifactListItem(
            id=a.id,
            title=a.title,
            type=a.type.value if isinstance(a.type, ArtifactType) else a.type,
            summary=a.summary,
            conversation_id=a.conversation_id,
            project_id=a.project_id,
            context_id=a.context_id,
            context_name=a.context.name if a.context else None,
            created_at=a.created_at.isoformat(),
            updated_at=a.updated_at.isoformat(),
        )
        for a in artifacts
    ]


@router.post("", response_model=ArtifactResponse, status_code=status.HTTP_201_CREATED)
async def create_artifact(
    data: ArtifactCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new artifact."""
    # Validate artifact type
    try:
        artifact_type = ArtifactType(data.type)
    except ValueError:
        artifact_type = ArtifactType.OTHER

    artifact = Artifact(
        user_id=current_user.id,
        title=data.title,
        content=data.content,
        type=artifact_type,
        summary=data.summary,
        conversation_id=data.conversation_id,
        project_id=data.project_id,
        context_id=data.context_id,
    )
    db.add(artifact)
    await db.commit()
    await db.refresh(artifact)

    return ArtifactResponse(
        id=artifact.id,
        title=artifact.title,
        type=artifact.type.value if isinstance(artifact.type, ArtifactType) else artifact.type,
        content=artifact.content,
        summary=artifact.summary,
        conversation_id=artifact.conversation_id,
        project_id=artifact.project_id,
        context_id=artifact.context_id,
        context_name=None,  # Not loaded via this endpoint
        version=artifact.version,
        created_at=artifact.created_at.isoformat(),
        updated_at=artifact.updated_at.isoformat(),
    )


@router.get("/{artifact_id}", response_model=ArtifactResponse)
async def get_artifact(
    artifact_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get a specific artifact by ID."""
    from sqlalchemy.orm import selectinload

    result = await db.execute(
        select(Artifact)
        .where(
            Artifact.id == artifact_id,
            Artifact.user_id == current_user.id,
        )
        .options(selectinload(Artifact.context))
    )
    artifact = result.scalar_one_or_none()

    if not artifact:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Artifact not found",
        )

    return ArtifactResponse(
        id=artifact.id,
        title=artifact.title,
        type=artifact.type.value if isinstance(artifact.type, ArtifactType) else artifact.type,
        content=artifact.content,
        summary=artifact.summary,
        conversation_id=artifact.conversation_id,
        project_id=artifact.project_id,
        context_id=artifact.context_id,
        context_name=artifact.context.name if artifact.context else None,
        version=artifact.version,
        created_at=artifact.created_at.isoformat(),
        updated_at=artifact.updated_at.isoformat(),
    )


@router.patch("/{artifact_id}", response_model=ArtifactResponse)
async def update_artifact(
    artifact_id: UUID,
    data: ArtifactUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update an artifact."""
    result = await db.execute(
        select(Artifact).where(
            Artifact.id == artifact_id,
            Artifact.user_id == current_user.id,
        )
    )
    artifact = result.scalar_one_or_none()

    if not artifact:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Artifact not found",
        )

    if data.title is not None:
        artifact.title = data.title
    if data.content is not None:
        artifact.content = data.content
        artifact.version += 1
    if data.summary is not None:
        artifact.summary = data.summary
    if data.project_id is not None:
        artifact.project_id = data.project_id
    if data.context_id is not None:
        artifact.context_id = data.context_id

    await db.commit()
    await db.refresh(artifact)

    return ArtifactResponse(
        id=artifact.id,
        title=artifact.title,
        type=artifact.type.value if isinstance(artifact.type, ArtifactType) else artifact.type,
        content=artifact.content,
        summary=artifact.summary,
        conversation_id=artifact.conversation_id,
        project_id=artifact.project_id,
        context_id=artifact.context_id,
        version=artifact.version,
        created_at=artifact.created_at.isoformat(),
        updated_at=artifact.updated_at.isoformat(),
    )


@router.patch("/{artifact_id}/link", response_model=ArtifactResponse)
async def link_artifact(
    artifact_id: UUID,
    data: ArtifactLinkUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Link an artifact to a project and/or context."""
    result = await db.execute(
        select(Artifact).where(
            Artifact.id == artifact_id,
            Artifact.user_id == current_user.id,
        )
    )
    artifact = result.scalar_one_or_none()

    if not artifact:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Artifact not found",
        )

    if data.project_id is not None:
        artifact.project_id = data.project_id
    if data.context_id is not None:
        artifact.context_id = data.context_id

    await db.commit()
    await db.refresh(artifact)

    return ArtifactResponse(
        id=artifact.id,
        title=artifact.title,
        type=artifact.type.value if isinstance(artifact.type, ArtifactType) else artifact.type,
        content=artifact.content,
        summary=artifact.summary,
        conversation_id=artifact.conversation_id,
        project_id=artifact.project_id,
        context_id=artifact.context_id,
        version=artifact.version,
        created_at=artifact.created_at.isoformat(),
        updated_at=artifact.updated_at.isoformat(),
    )


@router.delete("/{artifact_id}")
async def delete_artifact(
    artifact_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete an artifact."""
    result = await db.execute(
        select(Artifact).where(
            Artifact.id == artifact_id,
            Artifact.user_id == current_user.id,
        )
    )
    artifact = result.scalar_one_or_none()

    if not artifact:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Artifact not found",
        )

    await db.delete(artifact)
    await db.commit()

    return {"message": "Artifact deleted"}
