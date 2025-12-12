import secrets
from datetime import datetime
from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select, or_
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.models.context import Context, ContextType
from app.models.project import Project, ProjectNote
from app.models.node import Node
from app.models.artifact import Artifact
from app.models.task import Task
from app.schemas.context import (
    ContextCreate, ContextUpdate, ContextResponse,
    ContextListItem, BlocksUpdate, ShareResponse,
    AggregateContextRequest, AggregateContextResponse, AggregatedContextItem
)
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/contexts", tags=["contexts"])


@router.get("", response_model=list[ContextListItem])
async def list_contexts(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    type_filter: str | None = None,
    include_archived: bool = False,
    templates_only: bool = False,
    parent_id: UUID | None = None,
    search: str | None = None,
    skip: int = 0,
    limit: int = 50,
):
    """List all contexts for the current user with filtering options."""
    query = select(Context).where(Context.user_id == current_user.id)

    if type_filter:
        query = query.where(Context.type == type_filter)

    if not include_archived:
        query = query.where(Context.is_archived == False)

    if templates_only:
        query = query.where(Context.is_template == True)

    if parent_id:
        query = query.where(Context.parent_id == parent_id)

    if search:
        query = query.where(
            or_(
                Context.name.ilike(f"%{search}%"),
                Context.content.ilike(f"%{search}%")
            )
        )

    result = await db.execute(
        query.order_by(Context.updated_at.desc()).offset(skip).limit(limit)
    )
    return result.scalars().all()


@router.post("", response_model=ContextResponse)
async def create_context(
    data: ContextCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new context/document."""
    context = Context(
        user_id=current_user.id,
        name=data.name,
        type=data.type,
        content=data.content,
        structured_data=data.structured_data,
        system_prompt_template=data.system_prompt_template,
        blocks=data.blocks or [],
        cover_image=data.cover_image,
        icon=data.icon,
        parent_id=data.parent_id,
        is_template=data.is_template,
        last_edited_at=datetime.utcnow(),
    )
    db.add(context)
    await db.commit()
    await db.refresh(context)
    return context


@router.get("/{context_id}", response_model=ContextResponse)
async def get_context(
    context_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get a specific context by ID."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )
    return context


@router.put("/{context_id}", response_model=ContextResponse)
async def update_context(
    context_id: UUID,
    data: ContextUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a context."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )

    update_data = data.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(context, field, value)

    context.last_edited_at = datetime.utcnow()
    await db.commit()
    await db.refresh(context)
    return context


@router.patch("/{context_id}/blocks", response_model=ContextResponse)
async def update_blocks(
    context_id: UUID,
    data: BlocksUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Optimized endpoint for frequent block updates (auto-save)."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )

    context.blocks = data.blocks
    if data.word_count is not None:
        context.word_count = data.word_count
    context.last_edited_at = datetime.utcnow()

    await db.commit()
    await db.refresh(context)
    return context


@router.post("/{context_id}/share", response_model=ShareResponse)
async def enable_sharing(
    context_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Enable public sharing for a context and generate share link."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )

    # Generate share_id if not exists
    if not context.share_id:
        context.share_id = secrets.token_urlsafe(16)

    context.is_public = True
    await db.commit()
    await db.refresh(context)

    return ShareResponse(
        share_id=context.share_id,
        is_public=context.is_public,
        share_url=f"/public/doc/{context.share_id}"
    )


@router.delete("/{context_id}/share")
async def disable_sharing(
    context_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Disable public sharing for a context."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )

    context.is_public = False
    await db.commit()

    return {"message": "Sharing disabled"}


@router.get("/public/{share_id}", response_model=ContextResponse)
async def get_public_context(
    share_id: str,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get a publicly shared context (no auth required)."""
    result = await db.execute(
        select(Context).where(
            Context.share_id == share_id,
            Context.is_public == True,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Document not found or not public",
        )
    return context


@router.post("/{context_id}/duplicate", response_model=ContextResponse)
async def duplicate_context(
    context_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Duplicate a context/document."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    original = result.scalar_one_or_none()

    if not original:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )

    duplicate = Context(
        user_id=current_user.id,
        name=f"{original.name} (Copy)",
        type=original.type,
        content=original.content,
        structured_data=original.structured_data,
        system_prompt_template=original.system_prompt_template,
        blocks=original.blocks,
        cover_image=original.cover_image,
        icon=original.icon,
        parent_id=original.parent_id,
        is_template=False,  # Duplicates are never templates
        last_edited_at=datetime.utcnow(),
    )
    db.add(duplicate)
    await db.commit()
    await db.refresh(duplicate)
    return duplicate


@router.patch("/{context_id}/archive")
async def archive_context(
    context_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Archive a context."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )

    context.is_archived = True
    await db.commit()

    return {"message": "Context archived"}


@router.patch("/{context_id}/unarchive")
async def unarchive_context(
    context_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Unarchive a context."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )

    context.is_archived = False
    await db.commit()

    return {"message": "Context unarchived"}


@router.delete("/{context_id}")
async def delete_context(
    context_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete a context."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == current_user.id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Context not found",
        )

    await db.delete(context)
    await db.commit()
    return {"message": "Context deleted"}


def extract_text_from_blocks(blocks: list | None) -> str:
    """Extract plain text from document blocks recursively."""
    if not blocks:
        return ""

    text_parts = []
    for block in blocks:
        block_type = block.get("type", "")
        content = block.get("content", "")

        # Add content based on block type
        if block_type in ["paragraph", "quote", "callout"]:
            if content:
                text_parts.append(content)
        elif block_type.startswith("heading"):
            if content:
                text_parts.append(f"\n## {content}\n")
        elif block_type in ["bulletList", "numberedList"]:
            if content:
                text_parts.append(f"• {content}")
        elif block_type == "todo":
            checked = block.get("properties", {}).get("checked", False)
            prefix = "[x]" if checked else "[ ]"
            if content:
                text_parts.append(f"{prefix} {content}")
        elif block_type == "code":
            language = block.get("properties", {}).get("language", "")
            if content:
                text_parts.append(f"```{language}\n{content}\n```")

        # Process children recursively
        children = block.get("children")
        if children:
            text_parts.append(extract_text_from_blocks(children))

    return "\n".join(filter(None, text_parts))


async def get_context_with_children(
    db: AsyncSession,
    context_id: UUID,
    user_id: str,
    depth: int = 0,
    max_depth: int = 2
) -> list[Context]:
    """Recursively get context and its children up to max_depth."""
    result = await db.execute(
        select(Context).where(
            Context.id == context_id,
            Context.user_id == user_id,
        )
    )
    context = result.scalar_one_or_none()

    if not context:
        return []

    contexts = [context]

    if depth < max_depth:
        # Get child documents
        children_result = await db.execute(
            select(Context).where(
                Context.parent_id == context_id,
                Context.user_id == user_id,
                Context.is_archived == False,
            )
        )
        children = children_result.scalars().all()

        for child in children:
            contexts.extend(
                await get_context_with_children(db, child.id, user_id, depth + 1, max_depth)
            )

    return contexts


@router.post("/aggregate", response_model=AggregateContextResponse)
async def aggregate_context(
    data: AggregateContextRequest,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Aggregate context from profiles, projects, and nodes for AI consumption.

    This endpoint collects and formats content from multiple sources:
    - Contexts/Documents: Name, content, blocks (extracted text), and optionally child documents
    - Projects: Name, description, notes, and optionally tasks
    - Nodes: Name, purpose, status, weekly focus, and linked context
    - Artifacts: Linked to any of the above sources

    Returns a pre-formatted context string suitable for AI system prompts.
    """
    items: list[AggregatedContextItem] = []

    # Process contexts/documents
    if data.context_ids:
        for context_id in data.context_ids:
            if data.include_children:
                contexts = await get_context_with_children(
                    db, context_id, current_user.id, max_depth=data.max_depth
                )
            else:
                result = await db.execute(
                    select(Context).where(
                        Context.id == context_id,
                        Context.user_id == current_user.id,
                    )
                )
                context = result.scalar_one_or_none()
                contexts = [context] if context else []

            for context in contexts:
                # Build content from context
                content_parts = []

                if context.content:
                    content_parts.append(context.content)

                if context.blocks:
                    block_text = extract_text_from_blocks(context.blocks)
                    if block_text:
                        content_parts.append(block_text)

                if context.system_prompt_template:
                    content_parts.append(f"System Template: {context.system_prompt_template}")

                if context.structured_data:
                    content_parts.append(f"Data: {context.structured_data}")

                if content_parts:
                    items.append(AggregatedContextItem(
                        source_type="context",
                        source_id=context.id,
                        source_name=context.name,
                        content="\n\n".join(content_parts),
                        metadata={
                            "type": context.type.value,
                            "icon": context.icon,
                            "word_count": context.word_count
                        }
                    ))

                # Get linked artifacts
                if data.include_artifacts:
                    artifacts_result = await db.execute(
                        select(Artifact).where(
                            Artifact.context_id == context.id,
                            Artifact.user_id == current_user.id,
                        )
                    )
                    artifacts = artifacts_result.scalars().all()

                    for artifact in artifacts:
                        items.append(AggregatedContextItem(
                            source_type="artifact",
                            source_id=artifact.id,
                            source_name=artifact.title,
                            content=artifact.content,
                            metadata={
                                "type": artifact.type.value,
                                "language": artifact.language,
                                "summary": artifact.summary,
                                "parent_context": str(context.id)
                            }
                        ))

    # Process projects
    if data.project_ids:
        for project_id in data.project_ids:
            result = await db.execute(
                select(Project).where(
                    Project.id == project_id,
                    Project.user_id == current_user.id,
                )
            )
            project = result.scalar_one_or_none()

            if project:
                # Build project content
                content_parts = []

                if project.description:
                    content_parts.append(f"Description: {project.description}")

                content_parts.append(f"Status: {project.status.value}")
                content_parts.append(f"Priority: {project.priority.value}")

                if project.client_name:
                    content_parts.append(f"Client: {project.client_name}")

                if project.project_metadata:
                    content_parts.append(f"Metadata: {project.project_metadata}")

                items.append(AggregatedContextItem(
                    source_type="project",
                    source_id=project.id,
                    source_name=project.name,
                    content="\n".join(content_parts),
                    metadata={
                        "status": project.status.value,
                        "priority": project.priority.value,
                        "type": project.project_type
                    }
                ))

                # Get project notes
                notes_result = await db.execute(
                    select(ProjectNote).where(ProjectNote.project_id == project.id)
                )
                notes = notes_result.scalars().all()

                for note in notes:
                    items.append(AggregatedContextItem(
                        source_type="project_note",
                        source_id=note.id,
                        source_name=f"Note for {project.name}",
                        content=note.content,
                        metadata={"project_id": str(project.id)}
                    ))

                # Get project tasks
                if data.include_tasks:
                    tasks_result = await db.execute(
                        select(Task).where(
                            Task.project_id == project.id,
                            Task.user_id == current_user.id,
                        )
                    )
                    tasks = tasks_result.scalars().all()

                    for task in tasks:
                        task_content = f"[{task.status.value}] {task.title}"
                        if task.description:
                            task_content += f"\n{task.description}"
                        if task.due_date:
                            task_content += f"\nDue: {task.due_date.strftime('%Y-%m-%d')}"

                        items.append(AggregatedContextItem(
                            source_type="task",
                            source_id=task.id,
                            source_name=task.title,
                            content=task_content,
                            metadata={
                                "status": task.status.value,
                                "priority": task.priority.value,
                                "project_id": str(project.id)
                            }
                        ))

                # Get project artifacts
                if data.include_artifacts:
                    artifacts_result = await db.execute(
                        select(Artifact).where(
                            Artifact.project_id == project.id,
                            Artifact.user_id == current_user.id,
                        )
                    )
                    artifacts = artifacts_result.scalars().all()

                    for artifact in artifacts:
                        items.append(AggregatedContextItem(
                            source_type="artifact",
                            source_id=artifact.id,
                            source_name=artifact.title,
                            content=artifact.content,
                            metadata={
                                "type": artifact.type.value,
                                "language": artifact.language,
                                "summary": artifact.summary,
                                "parent_project": str(project.id)
                            }
                        ))

    # Process nodes
    if data.node_ids:
        for node_id in data.node_ids:
            result = await db.execute(
                select(Node).where(
                    Node.id == node_id,
                    Node.user_id == current_user.id,
                )
            )
            node = result.scalar_one_or_none()

            if node:
                # Build node content
                content_parts = []

                if node.purpose:
                    content_parts.append(f"Purpose: {node.purpose}")

                if node.current_status:
                    content_parts.append(f"Current Status: {node.current_status}")

                if node.this_week_focus:
                    focus_items = ", ".join(str(item) for item in node.this_week_focus)
                    content_parts.append(f"This Week's Focus: {focus_items}")

                if node.decision_queue:
                    decisions = ", ".join(str(item) for item in node.decision_queue)
                    content_parts.append(f"Decisions Pending: {decisions}")

                if node.delegation_ready:
                    delegations = ", ".join(str(item) for item in node.delegation_ready)
                    content_parts.append(f"Ready to Delegate: {delegations}")

                items.append(AggregatedContextItem(
                    source_type="node",
                    source_id=node.id,
                    source_name=node.name,
                    content="\n".join(content_parts) if content_parts else f"Node: {node.name}",
                    metadata={
                        "type": node.type.value,
                        "health": node.health.value,
                        "is_active": node.is_active
                    }
                ))

                # Get node's linked context
                if node.context_id and data.include_children:
                    context_result = await db.execute(
                        select(Context).where(
                            Context.id == node.context_id,
                            Context.user_id == current_user.id,
                        )
                    )
                    context = context_result.scalar_one_or_none()

                    if context:
                        content_parts = []
                        if context.content:
                            content_parts.append(context.content)
                        if context.blocks:
                            block_text = extract_text_from_blocks(context.blocks)
                            if block_text:
                                content_parts.append(block_text)

                        if content_parts:
                            items.append(AggregatedContextItem(
                                source_type="context",
                                source_id=context.id,
                                source_name=f"{node.name} - {context.name}",
                                content="\n\n".join(content_parts),
                                metadata={
                                    "type": context.type.value,
                                    "parent_node": str(node.id)
                                }
                            ))

    # Calculate totals and format output
    total_chars = sum(len(item.content) for item in items)

    # Build formatted context string
    formatted_parts = []

    # Group by source type
    contexts = [i for i in items if i.source_type == "context"]
    projects = [i for i in items if i.source_type == "project"]
    nodes = [i for i in items if i.source_type == "node"]
    tasks = [i for i in items if i.source_type == "task"]
    artifacts = [i for i in items if i.source_type == "artifact"]
    notes = [i for i in items if i.source_type == "project_note"]

    if contexts:
        formatted_parts.append("=== DOCUMENTS & PROFILES ===")
        for ctx in contexts:
            formatted_parts.append(f"\n--- {ctx.source_name} ---")
            formatted_parts.append(ctx.content)

    if projects:
        formatted_parts.append("\n\n=== PROJECTS ===")
        for proj in projects:
            formatted_parts.append(f"\n--- {proj.source_name} ---")
            formatted_parts.append(proj.content)

    if nodes:
        formatted_parts.append("\n\n=== BUSINESS NODES ===")
        for node in nodes:
            formatted_parts.append(f"\n--- {node.source_name} ---")
            formatted_parts.append(node.content)

    if tasks:
        formatted_parts.append("\n\n=== TASKS ===")
        for task in tasks:
            formatted_parts.append(f"• {task.content}")

    if notes:
        formatted_parts.append("\n\n=== PROJECT NOTES ===")
        for note in notes:
            formatted_parts.append(f"\n--- {note.source_name} ---")
            formatted_parts.append(note.content)

    if artifacts:
        formatted_parts.append("\n\n=== ARTIFACTS & CODE ===")
        for artifact in artifacts:
            lang = artifact.metadata.get("language", "") if artifact.metadata else ""
            formatted_parts.append(f"\n--- {artifact.source_name} ({lang}) ---")
            formatted_parts.append(artifact.content)

    formatted_context = "\n".join(formatted_parts)

    return AggregateContextResponse(
        items=items,
        total_items=len(items),
        total_characters=total_chars,
        formatted_context=formatted_context
    )
