from typing import Annotated
from uuid import UUID
from datetime import date, datetime

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.models.daily_log import DailyLog
from app.schemas.daily_log import DailyLogCreate, DailyLogUpdate, DailyLogResponse
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/daily", tags=["daily-logs"])


@router.get("/logs", response_model=list[DailyLogResponse])
async def list_daily_logs(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    skip: int = 0,
    limit: int = 30,
):
    """Get daily logs for the current user, most recent first"""
    result = await db.execute(
        select(DailyLog)
        .where(DailyLog.user_id == current_user.id)
        .order_by(DailyLog.date.desc())
        .offset(skip)
        .limit(limit)
    )
    return result.scalars().all()


@router.get("/logs/today", response_model=DailyLogResponse | None)
async def get_today_log(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get today's daily log if it exists"""
    today = date.today()
    result = await db.execute(
        select(DailyLog).where(
            DailyLog.user_id == current_user.id,
            DailyLog.date == today,
        )
    )
    return result.scalar_one_or_none()


@router.get("/logs/{log_date}", response_model=DailyLogResponse | None)
async def get_log_by_date(
    log_date: date,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get daily log for a specific date"""
    result = await db.execute(
        select(DailyLog).where(
            DailyLog.user_id == current_user.id,
            DailyLog.date == log_date,
        )
    )
    return result.scalar_one_or_none()


@router.post("/logs", response_model=DailyLogResponse)
async def create_or_update_daily_log(
    data: DailyLogCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create or update a daily log (upsert based on date)"""
    target_date = data.date or date.today()

    # Check if log already exists for this date
    result = await db.execute(
        select(DailyLog).where(
            DailyLog.user_id == current_user.id,
            DailyLog.date == target_date,
        )
    )
    existing_log = result.scalar_one_or_none()

    if existing_log:
        # Update existing log
        existing_log.content = data.content
        if data.energy_level is not None:
            existing_log.energy_level = data.energy_level
        existing_log.updated_at = datetime.utcnow()
        await db.commit()
        await db.refresh(existing_log)
        return existing_log
    else:
        # Create new log
        log = DailyLog(
            user_id=current_user.id,
            date=target_date,
            content=data.content,
            energy_level=data.energy_level,
        )
        db.add(log)
        await db.commit()
        await db.refresh(log)
        return log


@router.put("/logs/{log_id}", response_model=DailyLogResponse)
async def update_daily_log(
    log_id: UUID,
    data: DailyLogUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a specific daily log"""
    result = await db.execute(
        select(DailyLog).where(
            DailyLog.id == log_id,
            DailyLog.user_id == current_user.id,
        )
    )
    log = result.scalar_one_or_none()

    if not log:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Daily log not found",
        )

    if data.content is not None:
        log.content = data.content
    if data.energy_level is not None:
        log.energy_level = data.energy_level
    if data.extracted_actions is not None:
        log.extracted_actions = data.extracted_actions

    log.updated_at = datetime.utcnow()
    await db.commit()
    await db.refresh(log)
    return log


@router.delete("/logs/{log_id}")
async def delete_daily_log(
    log_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete a daily log"""
    result = await db.execute(
        select(DailyLog).where(
            DailyLog.id == log_id,
            DailyLog.user_id == current_user.id,
        )
    )
    log = result.scalar_one_or_none()

    if not log:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Daily log not found",
        )

    await db.delete(log)
    await db.commit()
    return {"message": "Daily log deleted"}
