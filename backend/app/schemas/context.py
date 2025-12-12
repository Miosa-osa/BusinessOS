from pydantic import BaseModel, Field
from datetime import datetime
from uuid import UUID
from typing import Any
from app.models.context import ContextType


# Block schema for document editor
class Block(BaseModel):
    id: str
    type: str  # paragraph, heading1, heading2, heading3, bulletList, numberedList, todo, quote, code, divider, image, callout, table, embed, artifact
    content: str | None = None
    properties: dict[str, Any] | None = None  # checked, language, artifactId, etc.
    children: list["Block"] | None = None


class PropertySchema(BaseModel):
    """Schema for defining document properties (Notion-like)"""
    name: str
    type: str  # text, select, multi_select, date, person, relation, number, checkbox, url, email
    options: list[str] | None = None  # For select/multi_select
    relation_type: str | None = None  # For relation: 'context', 'project', 'client'


class ContextCreate(BaseModel):
    name: str
    type: ContextType = ContextType.CUSTOM
    content: str | None = None
    structured_data: dict | None = None
    system_prompt_template: str | None = None
    # Document editor fields
    blocks: list[dict] | None = None
    cover_image: str | None = None
    icon: str | None = None
    parent_id: UUID | None = None
    is_template: bool = False
    # Document properties (Notion-like)
    property_schema: list[PropertySchema] | None = None
    properties: dict | None = None
    # Entity linking
    client_id: UUID | None = None


class ContextUpdate(BaseModel):
    name: str | None = None
    type: ContextType | None = None
    content: str | None = None
    structured_data: dict | None = None
    system_prompt_template: str | None = None
    # Document editor fields
    blocks: list[dict] | None = None
    cover_image: str | None = None
    icon: str | None = None
    parent_id: UUID | None = None
    is_template: bool | None = None
    is_archived: bool | None = None
    is_public: bool | None = None
    # Document properties (Notion-like)
    property_schema: list[dict] | None = None
    properties: dict | None = None
    # Entity linking
    client_id: UUID | None = None


class BlocksUpdate(BaseModel):
    """Optimized schema for frequent block updates"""
    blocks: list[dict]
    word_count: int | None = None


class ContextResponse(BaseModel):
    id: UUID
    name: str
    type: ContextType
    content: str | None
    structured_data: dict | None
    system_prompt_template: str | None
    # Document editor fields
    blocks: list[dict] | None
    cover_image: str | None
    icon: str | None
    parent_id: UUID | None
    is_template: bool
    is_archived: bool
    last_edited_at: datetime | None
    word_count: int
    is_public: bool
    share_id: str | None
    # Document properties (Notion-like)
    property_schema: list[dict] | None
    properties: dict | None
    # Entity linking
    client_id: UUID | None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ContextListItem(BaseModel):
    """Lightweight response for list views"""
    id: UUID
    name: str
    type: ContextType
    icon: str | None
    cover_image: str | None
    parent_id: UUID | None
    is_template: bool
    is_archived: bool
    word_count: int
    # Document properties (Notion-like)
    property_schema: list[dict] | None
    properties: dict | None
    # Entity linking
    client_id: UUID | None
    updated_at: datetime

    class Config:
        from_attributes = True


class ShareResponse(BaseModel):
    """Response for share operations"""
    share_id: str
    is_public: bool
    share_url: str


class AggregateContextRequest(BaseModel):
    """Request for aggregating context from multiple sources for AI"""
    context_ids: list[UUID] | None = None
    project_ids: list[UUID] | None = None
    node_ids: list[UUID] | None = None
    include_children: bool = True  # Include child documents
    include_artifacts: bool = True  # Include linked artifacts
    include_tasks: bool = True  # Include project tasks
    max_depth: int = 2  # Max depth for nested documents


class AggregatedContextItem(BaseModel):
    """Individual item in aggregated context"""
    source_type: str  # 'context', 'project', 'node', 'artifact', 'task'
    source_id: UUID
    source_name: str
    content: str
    metadata: dict | None = None


class AggregateContextResponse(BaseModel):
    """Response containing aggregated context for AI"""
    items: list[AggregatedContextItem]
    total_items: int
    total_characters: int
    formatted_context: str  # Pre-formatted text ready for AI
