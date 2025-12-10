from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.models.context import Context
from app.schemas.context import ContextCreate, ContextUpdate, ContextResponse
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/contexts", tags=["contexts"])


@router.get("", response_model=list[ContextResponse])
async def list_contexts(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    type_filter: str | None = None,
    skip: int = 0,
    limit: int = 50,
):
    query = select(Context).where(Context.user_id == current_user.id)

    if type_filter:
        query = query.where(Context.type == type_filter)

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
    context = Context(
        user_id=current_user.id,
        name=data.name,
        type=data.type,
        content=data.content,
        structured_data=data.structured_data,
        system_prompt_template=data.system_prompt_template,
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

    await db.commit()
    await db.refresh(context)
    return context


@router.delete("/{context_id}")
async def delete_context(
    context_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
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
