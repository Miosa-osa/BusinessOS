import uuid
from datetime import datetime
from enum import Enum
from sqlalchemy import String, DateTime, Text, ForeignKey, Enum as SQLEnum
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import UUID, JSONB
from app.database import Base


class ContextType(str, Enum):
    PERSON = "person"
    BUSINESS = "business"
    PROJECT = "project"
    CUSTOM = "custom"


class Context(Base):
    __tablename__ = "contexts"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    # Better Auth manages users - no FK constraint since table is external
    user_id: Mapped[str] = mapped_column(String(255), index=True)
    name: Mapped[str] = mapped_column(String(255))
    type: Mapped[ContextType] = mapped_column(
        SQLEnum(ContextType), default=ContextType.CUSTOM
    )
    content: Mapped[str | None] = mapped_column(Text, nullable=True)
    structured_data: Mapped[dict | None] = mapped_column(JSONB, nullable=True)
    system_prompt_template: Mapped[str | None] = mapped_column(Text, nullable=True)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow, onupdate=datetime.utcnow
    )

    # Relationships (user relationship removed - Better Auth manages users)
    conversations: Mapped[list["Conversation"]] = relationship(
        "Conversation", back_populates="context"
    )
    nodes: Mapped[list["Node"]] = relationship(
        "Node", back_populates="context"
    )
