"""
Artifact Tools for Business OS Agents

These tools allow the AI to create, read, update, and list business artifacts
like documents, proposals, SOPs, frameworks, etc.
"""

from typing import Any
from uuid import UUID
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

# Tool definitions for Ollama function calling
TOOL_DEFINITIONS = [
    {
        "type": "function",
        "function": {
            "name": "create_artifact",
            "description": "Create a new business artifact (document, proposal, SOP, framework, etc.). Use this whenever the user asks you to create any kind of business document.",
            "parameters": {
                "type": "object",
                "properties": {
                    "title": {
                        "type": "string",
                        "description": "The title of the artifact"
                    },
                    "content": {
                        "type": "string",
                        "description": "The full content of the artifact in markdown format"
                    },
                    "artifact_type": {
                        "type": "string",
                        "enum": ["proposal", "sop", "framework", "agenda", "report", "plan", "other"],
                        "description": "The type of artifact being created"
                    },
                    "summary": {
                        "type": "string",
                        "description": "A brief summary of the artifact (1-2 sentences)"
                    }
                },
                "required": ["title", "content", "artifact_type"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "read_artifact",
            "description": "Read an existing artifact by its ID",
            "parameters": {
                "type": "object",
                "properties": {
                    "artifact_id": {
                        "type": "string",
                        "description": "The UUID of the artifact to read"
                    }
                },
                "required": ["artifact_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "update_artifact",
            "description": "Update an existing artifact's content",
            "parameters": {
                "type": "object",
                "properties": {
                    "artifact_id": {
                        "type": "string",
                        "description": "The UUID of the artifact to update"
                    },
                    "title": {
                        "type": "string",
                        "description": "New title (optional)"
                    },
                    "content": {
                        "type": "string",
                        "description": "New content (optional)"
                    }
                },
                "required": ["artifact_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "list_artifacts",
            "description": "List artifacts for the current conversation or user",
            "parameters": {
                "type": "object",
                "properties": {
                    "artifact_type": {
                        "type": "string",
                        "enum": ["proposal", "sop", "framework", "agenda", "report", "plan", "other"],
                        "description": "Filter by artifact type (optional)"
                    },
                    "limit": {
                        "type": "integer",
                        "description": "Maximum number of artifacts to return (default 10)"
                    }
                },
                "required": []
            }
        }
    }
]


async def create_artifact_tool(
    db: AsyncSession,
    user_id: UUID,
    conversation_id: UUID | None,
    title: str,
    content: str,
    artifact_type: str,
    summary: str | None = None,
) -> dict[str, Any]:
    """Create a new artifact in the database."""
    from app.models.artifact import Artifact

    artifact = Artifact(
        user_id=user_id,
        conversation_id=conversation_id,
        title=title,
        content=content,
        artifact_type=artifact_type,
        summary=summary or title,
    )
    db.add(artifact)
    await db.commit()
    await db.refresh(artifact)

    return {
        "id": str(artifact.id),
        "title": artifact.title,
        "artifact_type": artifact.artifact_type,
        "summary": artifact.summary,
        "created_at": artifact.created_at.isoformat(),
    }


async def read_artifact_tool(
    db: AsyncSession,
    user_id: UUID,
    artifact_id: str,
) -> dict[str, Any] | None:
    """Read an artifact by ID."""
    from app.models.artifact import Artifact

    result = await db.execute(
        select(Artifact).where(
            Artifact.id == UUID(artifact_id),
            Artifact.user_id == user_id,
        )
    )
    artifact = result.scalar_one_or_none()

    if not artifact:
        return None

    return {
        "id": str(artifact.id),
        "title": artifact.title,
        "content": artifact.content,
        "artifact_type": artifact.artifact_type,
        "summary": artifact.summary,
        "created_at": artifact.created_at.isoformat(),
        "updated_at": artifact.updated_at.isoformat(),
    }


async def update_artifact_tool(
    db: AsyncSession,
    user_id: UUID,
    artifact_id: str,
    title: str | None = None,
    content: str | None = None,
) -> dict[str, Any] | None:
    """Update an artifact's content."""
    from app.models.artifact import Artifact

    result = await db.execute(
        select(Artifact).where(
            Artifact.id == UUID(artifact_id),
            Artifact.user_id == user_id,
        )
    )
    artifact = result.scalar_one_or_none()

    if not artifact:
        return None

    if title:
        artifact.title = title
    if content:
        artifact.content = content

    await db.commit()
    await db.refresh(artifact)

    return {
        "id": str(artifact.id),
        "title": artifact.title,
        "artifact_type": artifact.artifact_type,
        "updated_at": artifact.updated_at.isoformat(),
    }


async def list_artifacts_tool(
    db: AsyncSession,
    user_id: UUID,
    artifact_type: str | None = None,
    conversation_id: UUID | None = None,
    limit: int = 10,
) -> list[dict[str, Any]]:
    """List artifacts for a user."""
    from app.models.artifact import Artifact

    query = select(Artifact).where(Artifact.user_id == user_id)

    if artifact_type:
        query = query.where(Artifact.artifact_type == artifact_type)
    if conversation_id:
        query = query.where(Artifact.conversation_id == conversation_id)

    query = query.order_by(Artifact.created_at.desc()).limit(limit)

    result = await db.execute(query)
    artifacts = result.scalars().all()

    return [
        {
            "id": str(a.id),
            "title": a.title,
            "artifact_type": a.artifact_type,
            "summary": a.summary,
            "created_at": a.created_at.isoformat(),
        }
        for a in artifacts
    ]
