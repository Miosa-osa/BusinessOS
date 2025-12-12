"""Pydantic schemas for Clients/CRM API"""
from pydantic import BaseModel, Field
from datetime import datetime, date
from decimal import Decimal
from uuid import UUID
from app.models.client import ClientType, ClientStatus, InteractionType, DealStage


# ============ Contact Schemas ============

class ContactCreate(BaseModel):
    name: str
    email: str | None = None
    phone: str | None = None
    role: str | None = None
    is_primary: bool = False
    notes: str | None = None


class ContactUpdate(BaseModel):
    name: str | None = None
    email: str | None = None
    phone: str | None = None
    role: str | None = None
    is_primary: bool | None = None
    notes: str | None = None


class ContactResponse(BaseModel):
    id: UUID
    client_id: UUID
    name: str
    email: str | None
    phone: str | None
    role: str | None
    is_primary: bool
    notes: str | None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


# ============ Interaction Schemas ============

class InteractionCreate(BaseModel):
    type: InteractionType
    subject: str
    description: str | None = None
    outcome: str | None = None
    contact_id: UUID | None = None
    occurred_at: datetime | None = None  # Defaults to now on server


class InteractionResponse(BaseModel):
    id: UUID
    client_id: UUID
    contact_id: UUID | None
    type: InteractionType
    subject: str
    description: str | None
    outcome: str | None
    occurred_at: datetime
    created_at: datetime

    class Config:
        from_attributes = True


# ============ Deal Schemas ============

class DealCreate(BaseModel):
    name: str
    value: Decimal = Decimal("0")
    stage: DealStage = DealStage.qualification
    probability: int = Field(default=0, ge=0, le=100)
    expected_close_date: date | None = None
    notes: str | None = None


class DealUpdate(BaseModel):
    name: str | None = None
    value: Decimal | None = None
    stage: DealStage | None = None
    probability: int | None = Field(default=None, ge=0, le=100)
    expected_close_date: date | None = None
    notes: str | None = None


class DealStageUpdate(BaseModel):
    stage: DealStage


class DealResponse(BaseModel):
    id: UUID
    client_id: UUID
    name: str
    value: Decimal
    stage: DealStage
    probability: int
    expected_close_date: date | None
    notes: str | None
    created_at: datetime
    updated_at: datetime
    closed_at: datetime | None

    class Config:
        from_attributes = True


# ============ Client Schemas ============

class ClientCreate(BaseModel):
    name: str
    type: ClientType = ClientType.company
    email: str | None = None
    phone: str | None = None
    website: str | None = None

    # Company-specific
    industry: str | None = None
    company_size: str | None = None

    # Address
    address: str | None = None
    city: str | None = None
    state: str | None = None
    zip_code: str | None = None
    country: str | None = None

    # CRM Fields
    status: ClientStatus = ClientStatus.lead
    source: str | None = None
    assigned_to: str | None = None

    # Metadata
    tags: list[str] = []
    custom_fields: dict = {}
    notes: str | None = None


class ClientUpdate(BaseModel):
    name: str | None = None
    type: ClientType | None = None
    email: str | None = None
    phone: str | None = None
    website: str | None = None

    # Company-specific
    industry: str | None = None
    company_size: str | None = None

    # Address
    address: str | None = None
    city: str | None = None
    state: str | None = None
    zip_code: str | None = None
    country: str | None = None

    # CRM Fields
    status: ClientStatus | None = None
    source: str | None = None
    assigned_to: str | None = None
    lifetime_value: Decimal | None = None

    # Metadata
    tags: list[str] | None = None
    custom_fields: dict | None = None
    notes: str | None = None


class ClientStatusUpdate(BaseModel):
    status: ClientStatus


class ClientResponse(BaseModel):
    id: UUID
    user_id: str
    name: str
    type: ClientType
    email: str | None
    phone: str | None
    website: str | None

    industry: str | None
    company_size: str | None

    address: str | None
    city: str | None
    state: str | None
    zip_code: str | None
    country: str | None

    status: ClientStatus
    source: str | None
    assigned_to: str | None
    lifetime_value: Decimal | None

    tags: list[str] | None
    custom_fields: dict | None
    notes: str | None

    created_at: datetime
    updated_at: datetime
    last_contacted_at: datetime | None

    class Config:
        from_attributes = True


class ClientDetailResponse(ClientResponse):
    """Extended client response with related entities"""
    contacts: list[ContactResponse] = []
    interactions: list[InteractionResponse] = []
    deals: list[DealResponse] = []

    class Config:
        from_attributes = True


class ClientListResponse(BaseModel):
    """Response for client list with counts"""
    id: UUID
    name: str
    type: ClientType
    email: str | None
    phone: str | None
    status: ClientStatus
    source: str | None
    assigned_to: str | None
    lifetime_value: Decimal | None
    tags: list[str] | None
    created_at: datetime
    last_contacted_at: datetime | None

    # Counts for display
    contacts_count: int = 0
    interactions_count: int = 0
    deals_count: int = 0
    active_deals_value: Decimal = Decimal("0")

    class Config:
        from_attributes = True
