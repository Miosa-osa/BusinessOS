from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.models.node import Node, NodeHealth
from app.schemas.node import (
    NodeCreate,
    NodeUpdate,
    NodeResponse,
    NodeTreeResponse,
    NodeDetailResponse,
    NodeActivateResponse,
)
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/nodes", tags=["nodes"])


def build_node_tree(nodes: list[Node], parent_id: UUID | None = None) -> list[NodeTreeResponse]:
    """Build a hierarchical tree from flat list of nodes"""
    result = []
    for node in nodes:
        if node.parent_id == parent_id:
            children = build_node_tree(nodes, node.id)
            result.append(NodeTreeResponse(
                id=node.id,
                parent_id=node.parent_id,
                name=node.name,
                type=node.type,
                health=node.health,
                purpose=node.purpose,
                this_week_focus=node.this_week_focus,
                is_active=node.is_active,
                is_archived=node.is_archived,
                sort_order=node.sort_order,
                updated_at=node.updated_at,
                children=children,
                children_count=len(children),
            ))
    return sorted(result, key=lambda x: x.sort_order)


@router.get("", response_model=list[NodeResponse])
async def list_nodes(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    include_archived: bool = False,
):
    """List all nodes for the user (flat list)"""
    query = select(Node).where(Node.user_id == current_user.id)
    if not include_archived:
        query = query.where(Node.is_archived == False)
    query = query.order_by(Node.sort_order, Node.created_at)

    result = await db.execute(query)
    return result.scalars().all()


@router.get("/tree", response_model=list[NodeTreeResponse])
async def get_node_tree(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    include_archived: bool = False,
):
    """Get nodes as a hierarchical tree"""
    query = select(Node).where(Node.user_id == current_user.id)
    if not include_archived:
        query = query.where(Node.is_archived == False)

    result = await db.execute(query)
    nodes = result.scalars().all()

    return build_node_tree(list(nodes), None)


@router.get("/active", response_model=NodeResponse | None)
async def get_active_node(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get the currently active node"""
    result = await db.execute(
        select(Node).where(
            Node.user_id == current_user.id,
            Node.is_active == True,
        )
    )
    return result.scalar_one_or_none()


@router.post("", response_model=NodeResponse, status_code=status.HTTP_201_CREATED)
async def create_node(
    data: NodeCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new node"""
    # Verify parent exists if provided
    if data.parent_id:
        parent_result = await db.execute(
            select(Node).where(
                Node.id == data.parent_id,
                Node.user_id == current_user.id,
            )
        )
        if not parent_result.scalar_one_or_none():
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Parent node not found",
            )

    # Get max sort_order for siblings
    sort_query = select(func.max(Node.sort_order)).where(
        Node.user_id == current_user.id,
        Node.parent_id == data.parent_id,
    )
    max_order_result = await db.execute(sort_query)
    max_order = max_order_result.scalar() or 0

    node = Node(
        user_id=current_user.id,
        name=data.name,
        type=data.type,
        parent_id=data.parent_id,
        purpose=data.purpose,
        context_id=data.context_id,
        health=NodeHealth.NOT_STARTED,
        sort_order=max_order + 1,
    )
    db.add(node)
    await db.commit()
    await db.refresh(node)
    return node


@router.get("/{node_id}", response_model=NodeDetailResponse)
async def get_node(
    node_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get a single node with full details"""
    result = await db.execute(
        select(Node)
        .where(
            Node.id == node_id,
            Node.user_id == current_user.id,
        )
        .options(selectinload(Node.parent))
    )
    node = result.scalar_one_or_none()

    if not node:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Node not found",
        )

    # Count children
    children_count_result = await db.execute(
        select(func.count(Node.id)).where(
            Node.parent_id == node_id,
            Node.is_archived == False,
        )
    )
    children_count = children_count_result.scalar() or 0

    # Get parent name
    parent_name = node.parent.name if node.parent else None

    return NodeDetailResponse(
        id=node.id,
        user_id=node.user_id,
        parent_id=node.parent_id,
        context_id=node.context_id,
        name=node.name,
        type=node.type,
        health=node.health,
        purpose=node.purpose,
        current_status=node.current_status,
        this_week_focus=node.this_week_focus,
        decision_queue=node.decision_queue,
        delegation_ready=node.delegation_ready,
        is_active=node.is_active,
        is_archived=node.is_archived,
        sort_order=node.sort_order,
        created_at=node.created_at,
        updated_at=node.updated_at,
        parent_name=parent_name,
        children_count=children_count,
        # TODO: Count linked items when those relations are set up
        linked_projects_count=0,
        linked_conversations_count=0,
        linked_artifacts_count=0,
    )


@router.patch("/{node_id}", response_model=NodeResponse)
async def update_node(
    node_id: UUID,
    data: NodeUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a node"""
    result = await db.execute(
        select(Node).where(
            Node.id == node_id,
            Node.user_id == current_user.id,
        )
    )
    node = result.scalar_one_or_none()

    if not node:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Node not found",
        )

    # Verify new parent if being changed
    if data.parent_id is not None and data.parent_id != node.parent_id:
        if data.parent_id == node_id:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Node cannot be its own parent",
            )
        parent_result = await db.execute(
            select(Node).where(
                Node.id == data.parent_id,
                Node.user_id == current_user.id,
            )
        )
        if not parent_result.scalar_one_or_none():
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Parent node not found",
            )

    # Update fields
    update_data = data.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(node, key, value)

    await db.commit()
    await db.refresh(node)
    return node


@router.post("/{node_id}/activate", response_model=NodeActivateResponse)
async def activate_node(
    node_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Activate a node (deactivates any other active node)"""
    result = await db.execute(
        select(Node).where(
            Node.id == node_id,
            Node.user_id == current_user.id,
        )
    )
    node = result.scalar_one_or_none()

    if not node:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Node not found",
        )

    # Find currently active node
    active_result = await db.execute(
        select(Node).where(
            Node.user_id == current_user.id,
            Node.is_active == True,
        )
    )
    previously_active = active_result.scalar_one_or_none()
    previous_active_id = previously_active.id if previously_active else None

    # Deactivate all nodes for this user
    all_nodes_result = await db.execute(
        select(Node).where(
            Node.user_id == current_user.id,
            Node.is_active == True,
        )
    )
    for n in all_nodes_result.scalars():
        n.is_active = False

    # Activate the target node
    node.is_active = True
    await db.commit()
    await db.refresh(node)

    # Build context prompt for chat
    context_prompt = f"""Current Active Node: {node.name}

Purpose: {node.purpose or 'Not defined'}

Current Status: {node.current_status or 'Not defined'}

This Week's Focus:
{chr(10).join(f'- {item}' for item in (node.this_week_focus or [])) or '- Not defined'}

Use this context to inform your responses."""

    return NodeActivateResponse(
        node=node,
        previous_active_id=previous_active_id,
        context_prompt=context_prompt,
    )


@router.post("/{node_id}/deactivate", response_model=NodeResponse)
async def deactivate_node(
    node_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Deactivate a node"""
    result = await db.execute(
        select(Node).where(
            Node.id == node_id,
            Node.user_id == current_user.id,
        )
    )
    node = result.scalar_one_or_none()

    if not node:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Node not found",
        )

    node.is_active = False
    await db.commit()
    await db.refresh(node)
    return node


@router.delete("/{node_id}")
async def delete_node(
    node_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete a node (and all its children due to CASCADE)"""
    result = await db.execute(
        select(Node).where(
            Node.id == node_id,
            Node.user_id == current_user.id,
        )
    )
    node = result.scalar_one_or_none()

    if not node:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Node not found",
        )

    await db.delete(node)
    await db.commit()
    return {"message": "Node deleted"}


@router.get("/{node_id}/children", response_model=list[NodeResponse])
async def get_node_children(
    node_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    include_archived: bool = False,
):
    """Get immediate children of a node"""
    # Verify node exists
    result = await db.execute(
        select(Node).where(
            Node.id == node_id,
            Node.user_id == current_user.id,
        )
    )
    if not result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Node not found",
        )

    # Get children
    query = select(Node).where(
        Node.parent_id == node_id,
        Node.user_id == current_user.id,
    )
    if not include_archived:
        query = query.where(Node.is_archived == False)
    query = query.order_by(Node.sort_order)

    children_result = await db.execute(query)
    return children_result.scalars().all()


@router.post("/{node_id}/reorder")
async def reorder_nodes(
    node_id: UUID,
    new_order: int,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Reorder a node among its siblings"""
    result = await db.execute(
        select(Node).where(
            Node.id == node_id,
            Node.user_id == current_user.id,
        )
    )
    node = result.scalar_one_or_none()

    if not node:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Node not found",
        )

    # Get all siblings
    siblings_result = await db.execute(
        select(Node).where(
            Node.parent_id == node.parent_id,
            Node.user_id == current_user.id,
            Node.is_archived == False,
        ).order_by(Node.sort_order)
    )
    siblings = list(siblings_result.scalars())

    # Remove node from current position
    siblings = [s for s in siblings if s.id != node_id]

    # Insert at new position
    new_order = max(0, min(new_order, len(siblings)))
    siblings.insert(new_order, node)

    # Update sort_order for all siblings
    for idx, sibling in enumerate(siblings):
        sibling.sort_order = idx

    await db.commit()
    return {"message": "Nodes reordered"}
