"""add_document_properties_fields

Revision ID: d48b1145c63f
Revises: e2eb93e4ecbe
Create Date: 2025-12-12 23:11:51.188500

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision: str = 'd48b1145c63f'
down_revision: Union[str, Sequence[str], None] = 'e2eb93e4ecbe'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Add property_schema and properties columns to contexts table."""
    op.add_column('contexts', sa.Column('property_schema', postgresql.JSONB(astext_type=sa.Text()), nullable=True))
    op.add_column('contexts', sa.Column('properties', postgresql.JSONB(astext_type=sa.Text()), nullable=True))


def downgrade() -> None:
    """Remove property_schema and properties columns from contexts table."""
    op.drop_column('contexts', 'properties')
    op.drop_column('contexts', 'property_schema')
