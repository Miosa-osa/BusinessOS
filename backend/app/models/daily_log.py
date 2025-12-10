import uuid
from datetime import datetime, date
from sqlalchemy import String, DateTime, Text, ForeignKey, Date, Integer
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import UUID, JSONB
from app.database import Base


class DailyLog(Base):
    __tablename__ = "daily_logs"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    # Better Auth manages users - no FK constraint since table is external
    user_id: Mapped[str] = mapped_column(String(255), index=True)
    date: Mapped[date] = mapped_column(Date, index=True)
    content: Mapped[str] = mapped_column(Text)
    transcription_source: Mapped[str | None] = mapped_column(String(50), nullable=True)
    extracted_actions: Mapped[dict | None] = mapped_column(JSONB, nullable=True)
    extracted_patterns: Mapped[dict | None] = mapped_column(JSONB, nullable=True)
    energy_level: Mapped[int | None] = mapped_column(Integer, nullable=True)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow, onupdate=datetime.utcnow
    )

    # Relationships (user relationship removed - Better Auth manages users)
