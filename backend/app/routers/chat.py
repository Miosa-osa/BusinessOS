from typing import Annotated
from uuid import UUID
import re
import json

from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.responses import StreamingResponse
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.models.conversation import Conversation, Message, MessageRole
from app.models.context import Context
from app.models.artifact import Artifact, ArtifactType
from app.schemas.conversation import (
    ConversationCreate,
    ConversationResponse,
    ConversationList,
    ChatRequest,
)
from app.services.ollama import get_ollama_service, SYSTEM_PROMPTS
from app.utils.auth import CurrentUser


# Artifact creation instruction to append to system prompts
ARTIFACT_INSTRUCTION = """

## Creating Artifacts
When the user asks you to create a document, proposal, SOP, framework, agenda, report, plan, or any business artifact:

**IMPORTANT - Follow this exact pattern:**

1. **FIRST**: Briefly explain what you're about to create (1-2 sentences)
   Example: "I'll create a comprehensive platform development framework for MIOSA that covers strategy, architecture, and implementation phases."

2. **THEN**: Output the artifact in this format:

```artifact
{
  "type": "proposal|sop|framework|agenda|report|plan|other",
  "title": "Title of the artifact",
  "summary": "Brief 1-2 sentence summary",
  "content": "The full content of the document in markdown format..."
}
```

3. **AFTER**: Continue the conversation naturally. Ask for feedback or offer to make changes.
   Example: "I've created the framework above. Would you like me to expand any section, adjust the timeline, or add more detail to a specific area?"

The artifact block will be automatically saved and displayed in the Artifacts panel.

Types:
- proposal: Business proposals, pitches, partnership proposals
- sop: Standard Operating Procedures
- framework: Strategic frameworks, decision frameworks, playbooks
- agenda: Meeting agendas, briefs
- report: Reports, executive summaries, analysis documents
- plan: Project plans, roadmaps, implementation plans
- other: Any other document type

**Key Rules:**
- NEVER start your response directly with the artifact block
- ALWAYS introduce what you're creating first
- ALWAYS follow up after the artifact asking for feedback
- Keep the introduction brief (1-2 sentences max)
"""


def parse_artifacts_from_response(response: str) -> tuple[str, list[dict]]:
    """
    Parse artifact blocks from the model's response.
    Returns: (cleaned_response, list_of_artifacts)
    """
    artifacts = []
    cleaned_response = response

    # Pattern to match artifact blocks
    pattern = r'```artifact\s*\n([\s\S]*?)\n```'
    matches = re.finditer(pattern, response)

    for match in matches:
        try:
            artifact_json = match.group(1).strip()
            artifact_data = json.loads(artifact_json)

            # Validate required fields
            if all(k in artifact_data for k in ['type', 'title', 'content']):
                artifacts.append(artifact_data)
                # Remove the artifact block from response but keep a reference
                cleaned_response = cleaned_response.replace(
                    match.group(0),
                    f"\n\n**[Artifact Created: {artifact_data['title']}]**\n"
                )
        except json.JSONDecodeError:
            # If JSON parsing fails, leave the block as-is
            continue

    return cleaned_response, artifacts


async def save_artifacts(
    db: AsyncSession,
    user_id: str,
    conversation_id: UUID,
    artifacts: list[dict],
) -> list[Artifact]:
    """Save parsed artifacts to the database."""
    saved = []
    for artifact_data in artifacts:
        # Map type string to enum
        type_str = artifact_data.get('type', 'other').lower()
        try:
            artifact_type = ArtifactType(type_str)
        except ValueError:
            artifact_type = ArtifactType.OTHER

        artifact = Artifact(
            user_id=user_id,
            conversation_id=conversation_id,
            title=artifact_data['title'],
            content=artifact_data['content'],
            type=artifact_type,
            summary=artifact_data.get('summary'),
        )
        db.add(artifact)
        saved.append(artifact)

    if saved:
        await db.commit()

    return saved

router = APIRouter(prefix="/api/chat", tags=["chat"])


@router.get("/conversations", response_model=list[ConversationList])
async def list_conversations(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    skip: int = 0,
    limit: int = 50,
):
    result = await db.execute(
        select(Conversation)
        .where(Conversation.user_id == current_user.id)
        .order_by(Conversation.updated_at.desc())
        .offset(skip)
        .limit(limit)
        .options(selectinload(Conversation.messages))
    )
    conversations = result.scalars().all()

    return [
        ConversationList(
            id=conv.id,
            title=conv.title,
            context_id=conv.context_id,
            created_at=conv.created_at,
            updated_at=conv.updated_at,
            message_count=len(conv.messages),
        )
        for conv in conversations
    ]


@router.post("/conversations", response_model=ConversationResponse)
async def create_conversation(
    data: ConversationCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    conversation = Conversation(
        user_id=current_user.id,
        title=data.title,
        context_id=data.context_id,
    )
    db.add(conversation)
    await db.commit()

    # Re-fetch with messages eagerly loaded for response serialization
    result = await db.execute(
        select(Conversation)
        .where(Conversation.id == conversation.id)
        .options(selectinload(Conversation.messages))
    )
    return result.scalar_one()


@router.get("/conversations/{conversation_id}", response_model=ConversationResponse)
async def get_conversation(
    conversation_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    result = await db.execute(
        select(Conversation)
        .where(
            Conversation.id == conversation_id,
            Conversation.user_id == current_user.id,
        )
        .options(selectinload(Conversation.messages))
    )
    conversation = result.scalar_one_or_none()

    if not conversation:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Conversation not found",
        )
    return conversation


@router.delete("/conversations/{conversation_id}")
async def delete_conversation(
    conversation_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    result = await db.execute(
        select(Conversation).where(
            Conversation.id == conversation_id,
            Conversation.user_id == current_user.id,
        )
    )
    conversation = result.scalar_one_or_none()

    if not conversation:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Conversation not found",
        )

    await db.delete(conversation)
    await db.commit()
    return {"message": "Conversation deleted"}


@router.post("/message")
async def send_message(
    data: ChatRequest,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Send a message and get a streaming response."""

    # Get or create conversation
    if data.conversation_id:
        result = await db.execute(
            select(Conversation)
            .where(
                Conversation.id == data.conversation_id,
                Conversation.user_id == current_user.id,
            )
            .options(selectinload(Conversation.messages))
        )
        conversation = result.scalar_one_or_none()
        if not conversation:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Conversation not found",
            )
    else:
        # Create new conversation
        conversation = Conversation(
            user_id=current_user.id,
            title=data.message[:50] + "..." if len(data.message) > 50 else data.message,
            context_id=data.context_id,
        )
        db.add(conversation)
        await db.commit()
        # Reload the conversation with messages relationship
        result = await db.execute(
            select(Conversation)
            .where(Conversation.id == conversation.id)
            .options(selectinload(Conversation.messages))
        )
        conversation = result.scalar_one()

    # Build system prompt
    system_prompt = SYSTEM_PROMPTS.get(data.system_prompt_key or "default", SYSTEM_PROMPTS["default"])

    # If context_id is set, get context and use its prompt template
    if conversation.context_id or data.context_id:
        context_id = data.context_id or conversation.context_id
        result = await db.execute(
            select(Context).where(
                Context.id == context_id,
                Context.user_id == current_user.id,
            )
        )
        context = result.scalar_one_or_none()
        if context and context.system_prompt_template:
            system_prompt = context.system_prompt_template
        elif context and context.content:
            system_prompt = f"{system_prompt}\n\nContext information:\n{context.content}"

    # Add artifact creation instruction to system prompt
    system_prompt = system_prompt + ARTIFACT_INSTRUCTION

    # Add node context if provided
    if data.node_context:
        system_prompt = system_prompt + f"\n\n## Active Node Context\n{data.node_context}"

    # Add user message to conversation
    user_message = Message(
        conversation_id=conversation.id,
        role=MessageRole.USER,
        content=data.message,
    )
    db.add(user_message)
    await db.commit()

    # Build message history for the model
    messages = [
        {"role": msg.role.value, "content": msg.content}
        for msg in conversation.messages
    ]
    messages.append({"role": "user", "content": data.message})

    # Get Ollama service (auto-detects cloud vs local based on model name)
    ollama = get_ollama_service(model=data.model)

    async def generate():
        full_response = ""
        async for chunk in ollama.chat(messages, model=data.model, system_prompt=system_prompt):
            full_response += chunk
            yield chunk

        # Parse any artifacts from the response
        cleaned_response, artifacts = parse_artifacts_from_response(full_response)

        # Save artifacts if any were created
        if artifacts:
            await save_artifacts(
                db=db,
                user_id=current_user.id,
                conversation_id=conversation.id,
                artifacts=artifacts,
            )

        # Save assistant message after streaming completes
        assistant_message = Message(
            conversation_id=conversation.id,
            role=MessageRole.ASSISTANT,
            content=full_response,  # Keep original response with artifact blocks
            message_metadata={
                "model": data.model or ollama.model,
                "artifacts_created": len(artifacts),
            },
        )
        db.add(assistant_message)
        await db.commit()

    return StreamingResponse(
        generate(),
        media_type="text/plain",
        headers={
            "X-Conversation-Id": str(conversation.id),
            "Cache-Control": "no-cache",
        },
    )


@router.get("/search")
async def search_conversations(
    q: str,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    limit: int = 20,
):
    """Search across all conversations and messages."""
    # Simple search using ILIKE for now
    # TODO: Implement full-text search with PostgreSQL
    result = await db.execute(
        select(Message)
        .join(Conversation)
        .where(
            Conversation.user_id == current_user.id,
            Message.content.ilike(f"%{q}%"),
        )
        .limit(limit)
    )
    messages = result.scalars().all()

    return [
        {
            "message_id": msg.id,
            "conversation_id": msg.conversation_id,
            "content": msg.content[:200] + "..." if len(msg.content) > 200 else msg.content,
            "role": msg.role.value,
            "created_at": msg.created_at,
        }
        for msg in messages
    ]
