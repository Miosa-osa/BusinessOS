"""Clients/CRM API routes"""
from typing import Annotated
from uuid import UUID
from datetime import datetime
from decimal import Decimal

from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy import select, func, case
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.models.client import (
    Client,
    ClientContact,
    ClientInteraction,
    ClientDeal,
    ClientStatus,
    ClientType,
    DealStage,
)
from app.schemas.client import (
    ClientCreate,
    ClientUpdate,
    ClientStatusUpdate,
    ClientResponse,
    ClientDetailResponse,
    ClientListResponse,
    ContactCreate,
    ContactUpdate,
    ContactResponse,
    InteractionCreate,
    InteractionResponse,
    DealCreate,
    DealUpdate,
    DealStageUpdate,
    DealResponse,
)
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/clients", tags=["clients"])
deals_router = APIRouter(prefix="/api/deals", tags=["deals"])


# ============ Client Endpoints ============

@router.get("", response_model=list[ClientListResponse])
async def list_clients(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    status_filter: ClientStatus | None = None,
    type_filter: ClientType | None = None,
    search: str | None = None,
    tags: list[str] = Query(default=[]),
    skip: int = 0,
    limit: int = 50,
):
    """List all clients with optional filters"""
    query = select(Client).where(Client.user_id == current_user.id)

    if status_filter:
        query = query.where(Client.status == status_filter)
    if type_filter:
        query = query.where(Client.type == type_filter)
    if search:
        search_term = f"%{search}%"
        query = query.where(
            (Client.name.ilike(search_term)) |
            (Client.email.ilike(search_term)) |
            (Client.phone.ilike(search_term))
        )
    if tags:
        # Filter clients that have any of the specified tags
        for tag in tags:
            query = query.where(Client.tags.contains([tag]))

    query = (
        query.order_by(Client.updated_at.desc())
        .offset(skip)
        .limit(limit)
        .options(
            selectinload(Client.contacts),
            selectinload(Client.interactions),
            selectinload(Client.deals),
        )
    )

    result = await db.execute(query)
    clients = result.scalars().all()

    # Build response with counts
    response = []
    for client in clients:
        active_deals_value = sum(
            d.value for d in client.deals
            if d.stage not in (DealStage.closed_won, DealStage.closed_lost)
        )
        response.append(ClientListResponse(
            id=client.id,
            name=client.name,
            type=client.type,
            email=client.email,
            phone=client.phone,
            status=client.status,
            source=client.source,
            assigned_to=client.assigned_to,
            lifetime_value=client.lifetime_value,
            tags=client.tags,
            created_at=client.created_at,
            last_contacted_at=client.last_contacted_at,
            contacts_count=len(client.contacts),
            interactions_count=len(client.interactions),
            deals_count=len(client.deals),
            active_deals_value=active_deals_value,
        ))
    return response


@router.post("", response_model=ClientResponse)
async def create_client(
    data: ClientCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new client"""
    client = Client(
        user_id=current_user.id,
        name=data.name,
        type=data.type,
        email=data.email,
        phone=data.phone,
        website=data.website,
        industry=data.industry,
        company_size=data.company_size,
        address=data.address,
        city=data.city,
        state=data.state,
        zip_code=data.zip_code,
        country=data.country,
        status=data.status,
        source=data.source,
        assigned_to=data.assigned_to,
        tags=data.tags,
        custom_fields=data.custom_fields,
        notes=data.notes,
    )
    db.add(client)
    await db.commit()
    await db.refresh(client)
    return client


@router.get("/{client_id}", response_model=ClientDetailResponse)
async def get_client(
    client_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Get client details with contacts, interactions, and deals"""
    result = await db.execute(
        select(Client)
        .where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
        .options(
            selectinload(Client.contacts),
            selectinload(Client.interactions),
            selectinload(Client.deals),
        )
    )
    client = result.scalar_one_or_none()

    if not client:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )
    return client


@router.put("/{client_id}", response_model=ClientResponse)
async def update_client(
    client_id: UUID,
    data: ClientUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a client"""
    result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    client = result.scalar_one_or_none()

    if not client:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    update_data = data.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(client, field, value)

    await db.commit()
    await db.refresh(client)
    return client


@router.patch("/{client_id}/status", response_model=ClientResponse)
async def update_client_status(
    client_id: UUID,
    data: ClientStatusUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Quick status update for a client"""
    result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    client = result.scalar_one_or_none()

    if not client:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    client.status = data.status
    await db.commit()
    await db.refresh(client)
    return client


@router.delete("/{client_id}")
async def delete_client(
    client_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete a client and all related data"""
    result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    client = result.scalar_one_or_none()

    if not client:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    await db.delete(client)
    await db.commit()
    return {"message": "Client deleted"}


# ============ Contact Endpoints ============

@router.get("/{client_id}/contacts", response_model=list[ContactResponse])
async def list_contacts(
    client_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """List all contacts for a client"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    if not client_result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    result = await db.execute(
        select(ClientContact)
        .where(ClientContact.client_id == client_id)
        .order_by(ClientContact.is_primary.desc(), ClientContact.name)
    )
    return result.scalars().all()


@router.post("/{client_id}/contacts", response_model=ContactResponse)
async def create_contact(
    client_id: UUID,
    data: ContactCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Add a contact to a client"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    if not client_result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    # If this is primary, unset other primary contacts
    if data.is_primary:
        await db.execute(
            select(ClientContact)
            .where(
                ClientContact.client_id == client_id,
                ClientContact.is_primary == True,
            )
        )
        existing_primary = await db.execute(
            select(ClientContact).where(
                ClientContact.client_id == client_id,
                ClientContact.is_primary == True,
            )
        )
        for contact in existing_primary.scalars().all():
            contact.is_primary = False

    contact = ClientContact(
        client_id=client_id,
        name=data.name,
        email=data.email,
        phone=data.phone,
        role=data.role,
        is_primary=data.is_primary,
        notes=data.notes,
    )
    db.add(contact)
    await db.commit()
    await db.refresh(contact)
    return contact


@router.put("/{client_id}/contacts/{contact_id}", response_model=ContactResponse)
async def update_contact(
    client_id: UUID,
    contact_id: UUID,
    data: ContactUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a contact"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    if not client_result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    result = await db.execute(
        select(ClientContact).where(
            ClientContact.id == contact_id,
            ClientContact.client_id == client_id,
        )
    )
    contact = result.scalar_one_or_none()

    if not contact:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Contact not found",
        )

    # If setting as primary, unset other primary contacts
    if data.is_primary:
        existing_primary = await db.execute(
            select(ClientContact).where(
                ClientContact.client_id == client_id,
                ClientContact.is_primary == True,
                ClientContact.id != contact_id,
            )
        )
        for c in existing_primary.scalars().all():
            c.is_primary = False

    update_data = data.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(contact, field, value)

    await db.commit()
    await db.refresh(contact)
    return contact


@router.delete("/{client_id}/contacts/{contact_id}")
async def delete_contact(
    client_id: UUID,
    contact_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Delete a contact"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    if not client_result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    result = await db.execute(
        select(ClientContact).where(
            ClientContact.id == contact_id,
            ClientContact.client_id == client_id,
        )
    )
    contact = result.scalar_one_or_none()

    if not contact:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Contact not found",
        )

    await db.delete(contact)
    await db.commit()
    return {"message": "Contact deleted"}


# ============ Interaction Endpoints ============

@router.get("/{client_id}/interactions", response_model=list[InteractionResponse])
async def list_interactions(
    client_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    skip: int = 0,
    limit: int = 50,
):
    """List all interactions for a client (timeline)"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    if not client_result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    result = await db.execute(
        select(ClientInteraction)
        .where(ClientInteraction.client_id == client_id)
        .order_by(ClientInteraction.occurred_at.desc())
        .offset(skip)
        .limit(limit)
    )
    return result.scalars().all()


@router.post("/{client_id}/interactions", response_model=InteractionResponse)
async def create_interaction(
    client_id: UUID,
    data: InteractionCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Log a new interaction with a client"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    client = client_result.scalar_one_or_none()
    if not client:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    # Verify contact belongs to client if provided
    if data.contact_id:
        contact_result = await db.execute(
            select(ClientContact).where(
                ClientContact.id == data.contact_id,
                ClientContact.client_id == client_id,
            )
        )
        if not contact_result.scalar_one_or_none():
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Contact does not belong to this client",
            )

    interaction = ClientInteraction(
        client_id=client_id,
        contact_id=data.contact_id,
        type=data.type,
        subject=data.subject,
        description=data.description,
        outcome=data.outcome,
        occurred_at=data.occurred_at or datetime.utcnow(),
    )
    db.add(interaction)

    # Update last_contacted_at on client
    client.last_contacted_at = interaction.occurred_at

    await db.commit()
    await db.refresh(interaction)
    return interaction


# ============ Deal Endpoints (nested under clients) ============

@router.get("/{client_id}/deals", response_model=list[DealResponse])
async def list_client_deals(
    client_id: UUID,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """List all deals for a client"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    if not client_result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    result = await db.execute(
        select(ClientDeal)
        .where(ClientDeal.client_id == client_id)
        .order_by(ClientDeal.created_at.desc())
    )
    return result.scalars().all()


@router.post("/{client_id}/deals", response_model=DealResponse)
async def create_deal(
    client_id: UUID,
    data: DealCreate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Create a new deal for a client"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    if not client_result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    deal = ClientDeal(
        client_id=client_id,
        name=data.name,
        value=data.value,
        stage=data.stage,
        probability=data.probability,
        expected_close_date=data.expected_close_date,
        notes=data.notes,
    )
    db.add(deal)
    await db.commit()
    await db.refresh(deal)
    return deal


@router.put("/{client_id}/deals/{deal_id}", response_model=DealResponse)
async def update_deal(
    client_id: UUID,
    deal_id: UUID,
    data: DealUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Update a deal"""
    # Verify client ownership
    client_result = await db.execute(
        select(Client).where(
            Client.id == client_id,
            Client.user_id == current_user.id,
        )
    )
    if not client_result.scalar_one_or_none():
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Client not found",
        )

    result = await db.execute(
        select(ClientDeal).where(
            ClientDeal.id == deal_id,
            ClientDeal.client_id == client_id,
        )
    )
    deal = result.scalar_one_or_none()

    if not deal:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Deal not found",
        )

    update_data = data.model_dump(exclude_unset=True)

    # Handle stage change to closed
    if "stage" in update_data:
        new_stage = update_data["stage"]
        if new_stage in (DealStage.closed_won, DealStage.closed_lost) and deal.closed_at is None:
            deal.closed_at = datetime.utcnow()
        elif new_stage not in (DealStage.closed_won, DealStage.closed_lost):
            deal.closed_at = None

    for field, value in update_data.items():
        setattr(deal, field, value)

    await db.commit()
    await db.refresh(deal)
    return deal


# ============ Deal Endpoints (standalone for pipeline view) ============

@deals_router.get("", response_model=list[DealResponse])
async def list_all_deals(
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
    stage_filter: DealStage | None = None,
    skip: int = 0,
    limit: int = 100,
):
    """List all deals across all clients (for pipeline view)"""
    # Get all client IDs owned by user
    client_query = select(Client.id).where(Client.user_id == current_user.id)
    client_result = await db.execute(client_query)
    client_ids = [row[0] for row in client_result.all()]

    if not client_ids:
        return []

    query = select(ClientDeal).where(ClientDeal.client_id.in_(client_ids))

    if stage_filter:
        query = query.where(ClientDeal.stage == stage_filter)

    query = query.order_by(ClientDeal.created_at.desc()).offset(skip).limit(limit)
    result = await db.execute(query)
    return result.scalars().all()


@deals_router.patch("/{deal_id}/stage", response_model=DealResponse)
async def update_deal_stage(
    deal_id: UUID,
    data: DealStageUpdate,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Quick stage update for a deal (drag-and-drop in pipeline)"""
    # Get deal with client check
    result = await db.execute(
        select(ClientDeal)
        .join(Client)
        .where(
            ClientDeal.id == deal_id,
            Client.user_id == current_user.id,
        )
    )
    deal = result.scalar_one_or_none()

    if not deal:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Deal not found",
        )

    # Handle stage change to closed
    if data.stage in (DealStage.closed_won, DealStage.closed_lost) and deal.closed_at is None:
        deal.closed_at = datetime.utcnow()
    elif data.stage not in (DealStage.closed_won, DealStage.closed_lost):
        deal.closed_at = None

    deal.stage = data.stage
    await db.commit()
    await db.refresh(deal)
    return deal
