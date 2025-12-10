import uuid
from datetime import datetime
from sqlalchemy import String, DateTime, Boolean
from sqlalchemy.orm import Mapped, mapped_column
from sqlalchemy.dialects.postgresql import UUID, JSONB
from app.database import Base


class UserSettings(Base):
    __tablename__ = "user_settings"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    # Better Auth manages users - no FK constraint since table is external
    user_id: Mapped[str] = mapped_column(String(255), unique=True, index=True)

    # AI Settings
    default_model: Mapped[str | None] = mapped_column(String(100), nullable=True)

    # Notification Settings
    email_notifications: Mapped[bool] = mapped_column(Boolean, default=True)
    daily_summary: Mapped[bool] = mapped_column(Boolean, default=False)

    # Appearance
    theme: Mapped[str] = mapped_column(String(20), default="light")
    sidebar_collapsed: Mapped[bool] = mapped_column(Boolean, default=False)

    # Privacy
    share_analytics: Mapped[bool] = mapped_column(Boolean, default=True)

    # Additional settings stored as JSON
    custom_settings: Mapped[dict | None] = mapped_column(JSONB, nullable=True)

    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow, onupdate=datetime.utcnow
    )
