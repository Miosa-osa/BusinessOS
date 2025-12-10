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


class ArtifactUpdate(BaseModel):
    title: str | None = Field(None, max_length=255)
    content: str | None = None
    summary: str | None = Field(None, max_length=500)


class ArtifactResponse(BaseModel):
    id: UUID
    title: str
    type: str
    content: str
    summary: str | None
    conversation_id: UUID | None
    project_id: UUID | None
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
    skip: int = 0,
    limit: int = 50,
):
    """List all artifacts for the current user."""
    query = select(Artifact).where(Artifact.user_id == current_user.id)

    if type:
        query = query.where(Artifact.type == type)
    if conversation_id:
        query = query.where(Artifact.conversation_id == conversation_id)
    if project_id:
        query = query.where(Artifact.project_id == project_id)

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

    return ArtifactResponse(
        id=artifact.id,
        title=artifact.title,
        type=artifact.type.value if isinstance(artifact.type, ArtifactType) else artifact.type,
        content=artifact.content,
        summary=artifact.summary,
        conversation_id=artifact.conversation_id,
        project_id=artifact.project_id,
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
