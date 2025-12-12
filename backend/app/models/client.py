"""Client/CRM models for BusinessOS"""
from datetime import datetime, date
from decimal import Decimal
from enum import Enum
from typing import Optional
from uuid import uuid4

from sqlalchemy import (
    Column,
    String,
    Text,
    DateTime,
    ForeignKey,
    Index,
    Boolean,
    Integer,
    Date,
    Numeric,
    Enum as SQLEnum,
)
from sqlalchemy.dialects.postgresql import UUID, JSONB
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from app.database import Base


class ClientType(str, Enum):
    """Type of client - company or individual"""
    company = "company"
    individual = "individual"


class ClientStatus(str, Enum):
    """Client lifecycle status"""
    lead = "lead"
    prospect = "prospect"
    active = "active"
    inactive = "inactive"
    churned = "churned"


class InteractionType(str, Enum):
    """Type of client interaction"""
    call = "call"
    email = "email"
    meeting = "meeting"
    note = "note"


class DealStage(str, Enum):
    """Deal/opportunity pipeline stage"""
    qualification = "qualification"
    proposal = "proposal"
    negotiation = "negotiation"
    closed_won = "closed_won"
    closed_lost = "closed_lost"


class Client(Base):
    """Main client/customer entity"""
    __tablename__ = "clients"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    user_id = Column(String, nullable=False, index=True)  # Better Auth user ID

    # Basic Info
    name = Column(String(255), nullable=False)
    type = Column(SQLEnum(ClientType), nullable=False, default=ClientType.company)
    email = Column(String(255), nullable=True)
    phone = Column(String(50), nullable=True)
    website = Column(String(255), nullable=True)

    # Company-specific
    industry = Column(String(100), nullable=True)
    company_size = Column(String(50), nullable=True)  # 1-10, 11-50, 51-200, etc.

    # Address
    address = Column(String(255), nullable=True)
    city = Column(String(100), nullable=True)
    state = Column(String(100), nullable=True)
    zip_code = Column(String(20), nullable=True)
    country = Column(String(100), nullable=True)

    # CRM Fields
    status = Column(SQLEnum(ClientStatus), nullable=False, default=ClientStatus.lead)
    source = Column(String(100), nullable=True)  # referral, website, cold-call, etc.
    assigned_to = Column(String(255), nullable=True)  # team member name/id

    # Value
    lifetime_value = Column(Numeric(12, 2), nullable=True)

    # Metadata
    tags = Column(JSONB, nullable=True, default=list)
    custom_fields = Column(JSONB, nullable=True, default=dict)
    notes = Column(Text, nullable=True)

    # Timestamps
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())
    last_contacted_at = Column(DateTime(timezone=True), nullable=True)

    # Relationships
    contacts = relationship("ClientContact", back_populates="client", cascade="all, delete-orphan")
    interactions = relationship("ClientInteraction", back_populates="client", cascade="all, delete-orphan")
    deals = relationship("ClientDeal", back_populates="client", cascade="all, delete-orphan")
    contexts = relationship("Context", back_populates="client")

    __table_args__ = (
        Index("ix_clients_user_status", "user_id", "status"),
        Index("ix_clients_user_type", "user_id", "type"),
    )


class ClientContact(Base):
    """Contact person for a client (B2B - multiple contacts per company)"""
    __tablename__ = "client_contacts"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    client_id = Column(UUID(as_uuid=True), ForeignKey("clients.id", ondelete="CASCADE"), nullable=False)

    name = Column(String(255), nullable=False)
    email = Column(String(255), nullable=True)
    phone = Column(String(50), nullable=True)
    role = Column(String(100), nullable=True)  # CEO, CTO, Buyer, Decision Maker, etc.
    is_primary = Column(Boolean, default=False)
    notes = Column(Text, nullable=True)

    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    # Relationships
    client = relationship("Client", back_populates="contacts")
    interactions = relationship("ClientInteraction", back_populates="contact")

    __table_args__ = (
        Index("ix_client_contacts_client", "client_id"),
    )


class ClientInteraction(Base):
    """Interaction/activity log for a client (calls, emails, meetings)"""
    __tablename__ = "client_interactions"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    client_id = Column(UUID(as_uuid=True), ForeignKey("clients.id", ondelete="CASCADE"), nullable=False)
    contact_id = Column(UUID(as_uuid=True), ForeignKey("client_contacts.id", ondelete="SET NULL"), nullable=True)

    type = Column(SQLEnum(InteractionType), nullable=False)
    subject = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    outcome = Column(String(255), nullable=True)

    occurred_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    # Relationships
    client = relationship("Client", back_populates="interactions")
    contact = relationship("ClientContact", back_populates="interactions")

    __table_args__ = (
        Index("ix_client_interactions_client", "client_id"),
        Index("ix_client_interactions_occurred", "occurred_at"),
    )


class ClientDeal(Base):
    """Deal/opportunity for pipeline tracking"""
    __tablename__ = "client_deals"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    client_id = Column(UUID(as_uuid=True), ForeignKey("clients.id", ondelete="CASCADE"), nullable=False)

    name = Column(String(255), nullable=False)
    value = Column(Numeric(12, 2), nullable=False, default=0)
    stage = Column(SQLEnum(DealStage), nullable=False, default=DealStage.qualification)
    probability = Column(Integer, nullable=False, default=0)  # 0-100
    expected_close_date = Column(Date, nullable=True)

    notes = Column(Text, nullable=True)

    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())
    closed_at = Column(DateTime(timezone=True), nullable=True)

    # Relationships
    client = relationship("Client", back_populates="deals")

    __table_args__ = (
        Index("ix_client_deals_client", "client_id"),
        Index("ix_client_deals_stage", "stage"),
    )
