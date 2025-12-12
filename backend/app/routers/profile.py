import os
import uuid
from typing import Annotated
from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, status, UploadFile, File
from fastapi.responses import FileResponse
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.models.user_settings import UserSettings
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/profile", tags=["profile"])

# Create uploads directory if it doesn't exist
UPLOAD_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), "uploads", "backgrounds")
os.makedirs(UPLOAD_DIR, exist_ok=True)

ALLOWED_EXTENSIONS = {".jpg", ".jpeg", ".png", ".gif", ".webp"}
MAX_FILE_SIZE = 5 * 1024 * 1024  # 5MB


@router.post("/background")
async def upload_background(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    file: UploadFile = File(...),
):
    """Upload a custom desktop background image"""

    # Check file extension
    _, ext = os.path.splitext(file.filename or "")
    ext = ext.lower()
    if ext not in ALLOWED_EXTENSIONS:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"File type not allowed. Allowed types: {', '.join(ALLOWED_EXTENSIONS)}"
        )

    # Check file size
    contents = await file.read()
    if len(contents) > MAX_FILE_SIZE:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"File too large. Maximum size: {MAX_FILE_SIZE // (1024 * 1024)}MB"
        )

    # Generate unique filename
    filename = f"{current_user.id}_{uuid.uuid4().hex}{ext}"
    filepath = os.path.join(UPLOAD_DIR, filename)

    # Delete old background if exists
    result = await db.execute(
        select(UserSettings).where(UserSettings.user_id == current_user.id)
    )
    user_settings = result.scalar_one_or_none()

    if user_settings and user_settings.custom_settings:
        old_background = user_settings.custom_settings.get("desktop_background")
        if old_background and old_background.startswith("/api/profile/background/"):
            old_filename = old_background.split("/")[-1]
            old_filepath = os.path.join(UPLOAD_DIR, old_filename)
            if os.path.exists(old_filepath):
                try:
                    os.remove(old_filepath)
                except:
                    pass  # Ignore deletion errors

    # Save new file
    with open(filepath, "wb") as f:
        f.write(contents)

    # Update user settings with new background URL
    background_url = f"/api/profile/background/{filename}"

    if not user_settings:
        user_settings = UserSettings(
            user_id=current_user.id,
            custom_settings={"desktop_background": background_url}
        )
        db.add(user_settings)
    else:
        custom_settings = user_settings.custom_settings or {}
        custom_settings["desktop_background"] = background_url
        user_settings.custom_settings = custom_settings
        user_settings.updated_at = datetime.utcnow()

    await db.commit()

    return {"url": background_url, "filename": filename}


@router.get("/background/{filename}")
async def get_background(
    filename: str,
):
    """Serve a background image"""
    filepath = os.path.join(UPLOAD_DIR, filename)

    if not os.path.exists(filepath):
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Background image not found"
        )

    # Determine content type
    _, ext = os.path.splitext(filename)
    ext = ext.lower()
    content_types = {
        ".jpg": "image/jpeg",
        ".jpeg": "image/jpeg",
        ".png": "image/png",
        ".gif": "image/gif",
        ".webp": "image/webp",
    }
    content_type = content_types.get(ext, "application/octet-stream")

    return FileResponse(filepath, media_type=content_type)


@router.delete("/background")
async def delete_background(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete the current custom background"""
    result = await db.execute(
        select(UserSettings).where(UserSettings.user_id == current_user.id)
    )
    user_settings = result.scalar_one_or_none()

    if user_settings and user_settings.custom_settings:
        old_background = user_settings.custom_settings.get("desktop_background")
        if old_background and old_background.startswith("/api/profile/background/"):
            old_filename = old_background.split("/")[-1]
            old_filepath = os.path.join(UPLOAD_DIR, old_filename)
            if os.path.exists(old_filepath):
                try:
                    os.remove(old_filepath)
                except:
                    pass

        custom_settings = user_settings.custom_settings
        custom_settings.pop("desktop_background", None)
        user_settings.custom_settings = custom_settings
        user_settings.updated_at = datetime.utcnow()
        await db.commit()

    return {"message": "Background deleted"}
