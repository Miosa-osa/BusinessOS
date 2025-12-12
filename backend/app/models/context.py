import uuid
from datetime import datetime
from enum import Enum
from sqlalchemy import String, DateTime, Text, ForeignKey, Enum as SQLEnum, Boolean, Integer
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import UUID, JSONB
from app.database import Base


class ContextType(str, Enum):
    PERSON = "person"
    BUSINESS = "business"
    PROJECT = "project"
    CUSTOM = "custom"
    DOCUMENT = "document"  # New type for Notion-like documents


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

    # Document editor fields (Notion-like)
    blocks: Mapped[list | None] = mapped_column(JSONB, nullable=True, default=list)
    cover_image: Mapped[str | None] = mapped_column(String(500), nullable=True)
    icon: Mapped[str | None] = mapped_column(String(50), nullable=True)
    parent_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("contexts.id", ondelete="SET NULL"), nullable=True, index=True
    )
    is_template: Mapped[bool] = mapped_column(Boolean, default=False)
    is_archived: Mapped[bool] = mapped_column(Boolean, default=False, index=True)
    last_edited_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
    word_count: Mapped[int] = mapped_column(Integer, default=0)

    # Sharing fields
    is_public: Mapped[bool] = mapped_column(Boolean, default=False)
    share_id: Mapped[str | None] = mapped_column(String(32), nullable=True, unique=True, index=True)

    # Document properties (Notion-like)
    # property_schema defines what properties this document type has
    # e.g. [{"name": "Status", "type": "select", "options": ["Todo", "In Progress", "Done"]}]
    property_schema: Mapped[list | None] = mapped_column(JSONB, nullable=True, default=list)
    # properties stores the actual property values
    # e.g. {"Status": "In Progress", "Priority": "High", "Related": ["uuid1", "uuid2"]}
    properties: Mapped[dict | None] = mapped_column(JSONB, nullable=True, default=dict)

    # Entity linking
    client_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("clients.id", ondelete="SET NULL"), nullable=True, index=True
    )

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
    artifacts: Mapped[list["Artifact"]] = relationship(
        "Artifact", back_populates="context"
    )
    # Self-referential relationship for nested pages
    children: Mapped[list["Context"]] = relationship(
        "Context", back_populates="parent", remote_side=[id]
    )
    parent: Mapped["Context | None"] = relationship(
        "Context", back_populates="children", remote_side=[parent_id]
    )
    # Client relationship
    client: Mapped["Client | None"] = relationship(
        "Client", back_populates="contexts"
    )
