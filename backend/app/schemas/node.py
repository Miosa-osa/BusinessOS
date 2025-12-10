from pydantic import BaseModel
from datetime import datetime
from uuid import UUID
from typing import Optional
from app.models.node import NodeType, NodeHealth


# Decision item schema
class DecisionItem(BaseModel):
    id: str
    question: str
    added_at: str
    decided: bool = False
    decision: str | None = None


# Delegation item schema
class DelegationItem(BaseModel):
    id: str
    task: str
    assignee_id: str | None = None
    assignee_name: str | None = None
    status: str = "pending"  # pending, delegated, completed


class NodeCreate(BaseModel):
    name: str
    type: NodeType
    parent_id: UUID | None = None
    purpose: str | None = None
    context_id: UUID | None = None


class NodeUpdate(BaseModel):
    name: str | None = None
    type: NodeType | None = None
    parent_id: UUID | None = None
    health: NodeHealth | None = None
    purpose: str | None = None
    current_status: str | None = None
    this_week_focus: list[str] | None = None
    decision_queue: list[DecisionItem] | None = None
    delegation_ready: list[DelegationItem] | None = None
    is_active: bool | None = None
    is_archived: bool | None = None
    sort_order: int | None = None
    context_id: UUID | None = None


class NodeResponse(BaseModel):
    id: UUID
    user_id: str
    parent_id: UUID | None
    context_id: UUID | None
    name: str
    type: NodeType
    health: NodeHealth
    purpose: str | None
    current_status: str | None
    this_week_focus: list[str] | None
    decision_queue: list | None
    delegation_ready: list | None
    is_active: bool
    is_archived: bool
    sort_order: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class NodeTreeResponse(BaseModel):
    """Node with children for tree view"""
    id: UUID
    parent_id: UUID | None
    name: str
    type: NodeType
    health: NodeHealth
    purpose: str | None
    this_week_focus: list[str] | None
    is_active: bool
    is_archived: bool
    sort_order: int
    updated_at: datetime
    children: list["NodeTreeResponse"] = []
    children_count: int = 0

    class Config:
        from_attributes = True


class NodeDetailResponse(BaseModel):
    """Full node detail with all fields"""
    id: UUID
    user_id: str
    parent_id: UUID | None
    context_id: UUID | None
    name: str
    type: NodeType
    health: NodeHealth
    purpose: str | None
    current_status: str | None
    this_week_focus: list[str] | None
    decision_queue: list | None
    delegation_ready: list | None
    is_active: bool
    is_archived: bool
    sort_order: int
    created_at: datetime
    updated_at: datetime
    parent_name: str | None = None
    children_count: int = 0
    # Linked items counts
    linked_projects_count: int = 0
    linked_conversations_count: int = 0
    linked_artifacts_count: int = 0

    class Config:
        from_attributes = True


class NodeActivateResponse(BaseModel):
    """Response when activating a node"""
    node: NodeResponse
    previous_active_id: UUID | None = None
    context_prompt: str | None = None  # The prompt to inject into chat


# Rebuild models for forward references
NodeTreeResponse.model_rebuild()
