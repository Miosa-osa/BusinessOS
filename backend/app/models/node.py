import uuid
from datetime import datetime
from enum import Enum
from sqlalchemy import String, DateTime, Text, ForeignKey, Boolean, Integer, Enum as SQLEnum
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.dialects.postgresql import UUID, JSONB
from app.database import Base


class NodeType(str, Enum):
    BUSINESS = "business"
    PROJECT = "project"
    LEARNING = "learning"
    OPERATIONAL = "operational"


class NodeHealth(str, Enum):
    HEALTHY = "healthy"
    NEEDS_ATTENTION = "needs_attention"
    CRITICAL = "critical"
    NOT_STARTED = "not_started"


class Node(Base):
    __tablename__ = "nodes"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    # Better Auth manages users - no FK constraint since table is external
    user_id: Mapped[str] = mapped_column(String(255), index=True)
    parent_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("nodes.id", ondelete="SET NULL"), nullable=True
    )
    context_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("contexts.id", ondelete="SET NULL"), nullable=True
    )
    name: Mapped[str] = mapped_column(String(255))
    type: Mapped[NodeType] = mapped_column(SQLEnum(NodeType))
    health: Mapped[NodeHealth] = mapped_column(
        SQLEnum(NodeHealth), default=NodeHealth.NOT_STARTED
    )
    purpose: Mapped[str | None] = mapped_column(Text, nullable=True)
    current_status: Mapped[str | None] = mapped_column(Text, nullable=True)
    this_week_focus: Mapped[list | None] = mapped_column(JSONB, nullable=True)  # Array of focus items
    decision_queue: Mapped[list | None] = mapped_column(JSONB, nullable=True)  # Array of decisions
    delegation_ready: Mapped[list | None] = mapped_column(JSONB, nullable=True)  # Array of items to delegate
    is_active: Mapped[bool] = mapped_column(Boolean, default=False)  # Currently active node
    is_archived: Mapped[bool] = mapped_column(Boolean, default=False)
    sort_order: Mapped[int] = mapped_column(Integer, default=0)  # For ordering siblings
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow, onupdate=datetime.utcnow
    )

    # Relationships (user relationship removed - Better Auth manages users)
    parent: Mapped["Node"] = relationship(
        "Node", remote_side=[id], back_populates="children"
    )
    children: Mapped[list["Node"]] = relationship(
        "Node", back_populates="parent"
    )
    context: Mapped["Context"] = relationship("Context", back_populates="nodes")
    metrics: Mapped[list["NodeMetric"]] = relationship(
        "NodeMetric", back_populates="node", cascade="all, delete-orphan"
    )


class NodeMetric(Base):
    __tablename__ = "node_metrics"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), primary_key=True, default=uuid.uuid4
    )
    node_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("nodes.id", ondelete="CASCADE")
    )
    metric_name: Mapped[str] = mapped_column(String(255))
    metric_value: Mapped[str] = mapped_column(String(500))
    recorded_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow
    )

    # Relationships
    node: Mapped["Node"] = relationship("Node", back_populates="metrics")
