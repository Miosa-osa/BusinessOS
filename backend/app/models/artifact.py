import uuid
from datetime import datetime
from enum import Enum
from sqlalchemy import String, DateTime, Text, ForeignKey, Integer, Enum as SQLEnum
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import UUID
from app.database import Base


class ArtifactType(str, Enum):
    # Business document types
    PROPOSAL = "proposal"
    SOP = "sop"
    FRAMEWORK = "framework"
    AGENDA = "agenda"
    REPORT = "report"
    PLAN = "plan"
    # Legacy code types
    CODE = "code"
    DOCUMENT = "document"
    MARKDOWN = "markdown"
    REACT = "react"
    HTML = "html"
    SVG = "svg"
    # Catch-all
    OTHER = "other"


class Artifact(Base):
    __tablename__ = "artifacts"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    # Better Auth manages users - no FK constraint since table is external
    user_id: Mapped[str] = mapped_column(String(255), index=True)
    conversation_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("conversations.id", ondelete="SET NULL"), nullable=True
    )
    message_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("messages.id", ondelete="SET NULL"), nullable=True
    )
    project_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("projects.id", ondelete="SET NULL"), nullable=True
    )
    title: Mapped[str] = mapped_column(String(255))
    type: Mapped[ArtifactType] = mapped_column(SQLEnum(ArtifactType))
    language: Mapped[str | None] = mapped_column(String(50), nullable=True)
    content: Mapped[str] = mapped_column(Text)
    summary: Mapped[str | None] = mapped_column(String(500), nullable=True)
    version: Mapped[int] = mapped_column(Integer, default=1)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow, onupdate=datetime.utcnow
    )

    # Relationships (user relationship removed - Better Auth manages users)
    conversation: Mapped["Conversation"] = relationship(
        "Conversation", back_populates="artifacts"
    )
    message: Mapped["Message"] = relationship("Message", back_populates="artifacts")
    project: Mapped["Project"] = relationship("Project", back_populates="artifacts")
    versions: Mapped[list["ArtifactVersion"]] = relationship(
        "ArtifactVersion", back_populates="artifact", cascade="all, delete-orphan"
    )


class ArtifactVersion(Base):
    __tablename__ = "artifact_versions"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    artifact_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("artifacts.id", ondelete="CASCADE")
    )
    version: Mapped[int] = mapped_column(Integer)
    content: Mapped[str] = mapped_column(Text)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )

    # Relationships
    artifact: Mapped["Artifact"] = relationship("Artifact", back_populates="versions")
