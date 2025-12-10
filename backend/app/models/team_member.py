import uuid
from datetime import datetime
from enum import Enum
from sqlalchemy import String, DateTime, Text, ForeignKey, Enum as SQLEnum, Integer, Numeric
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import UUID, JSONB, ARRAY
from app.database import Base


class MemberStatus(str, Enum):
    AVAILABLE = "available"
    BUSY = "busy"
    OVERLOADED = "overloaded"
    OOO = "ooo"  # Out of Office


class TeamMember(Base):
    __tablename__ = "team_members"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    # Better Auth manages users - this is the workspace owner
    user_id: Mapped[str] = mapped_column(String(255), index=True)

    # Member info
    name: Mapped[str] = mapped_column(String(255))
    email: Mapped[str] = mapped_column(String(255))
    role: Mapped[str] = mapped_column(String(255))
    avatar_url: Mapped[str | None] = mapped_column(Text, nullable=True)

    # Status and capacity
    status: Mapped[MemberStatus] = mapped_column(
        SQLEnum(MemberStatus), default=MemberStatus.AVAILABLE
    )
    capacity: Mapped[int] = mapped_column(Integer, default=0)  # 0-100 percentage

    # Org structure
    manager_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("team_members.id", ondelete="SET NULL"), nullable=True
    )

    # Additional info
    skills: Mapped[list | None] = mapped_column(JSONB, nullable=True)  # Array of skill strings
    hourly_rate: Mapped[float | None] = mapped_column(Numeric(10, 2), nullable=True)

    # Timestamps
    joined_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow, onupdate=datetime.utcnow
    )

    # Self-referential relationship for manager/reports
    manager: Mapped["TeamMember | None"] = relationship(
        "TeamMember",
        remote_side=[id],
        back_populates="reports",
        foreign_keys=[manager_id]
    )
    reports: Mapped[list["TeamMember"]] = relationship(
        "TeamMember",
        back_populates="manager",
        foreign_keys=[manager_id]
    )


class TeamMemberActivity(Base):
    __tablename__ = "team_member_activities"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    member_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("team_members.id", ondelete="CASCADE")
    )
    activity_type: Mapped[str] = mapped_column(String(100))
    description: Mapped[str] = mapped_column(Text)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )

    # Relationships
    member: Mapped["TeamMember"] = relationship("TeamMember")
