from pydantic import BaseModel
from datetime import datetime
from uuid import UUID
from app.models.conversation import MessageRole


class MessageCreate(BaseModel):
    content: str


class MessageResponse(BaseModel):
    id: UUID
    role: MessageRole
    content: str
    created_at: datetime
    message_metadata: dict | None = None

    class Config:
        from_attributes = True


class ConversationCreate(BaseModel):
    title: str | None = "New Conversation"
    context_id: UUID | None = None


class ConversationResponse(BaseModel):
    id: UUID
    title: str
    context_id: UUID | None
    created_at: datetime
    updated_at: datetime
    messages: list[MessageResponse] = []

    class Config:
        from_attributes = True


class ConversationList(BaseModel):
    id: UUID
    title: str
    context_id: UUID | None
    created_at: datetime
    updated_at: datetime
    message_count: int = 0

    class Config:
        from_attributes = True


class ChatRequest(BaseModel):
    message: str
    conversation_id: UUID | None = None
    context_id: UUID | None = None
    model: str | None = None
    system_prompt_key: str | None = None  # e.g., "daily_planning", "code_review"
    node_context: str | None = None  # Active node context to inject into system prompt
