"""update_nodes_add_health_and_fields

Revision ID: 089de6f1f16d
Revises: 9a2a45ab63ac
Create Date: 2025-12-10 23:15:09.919123

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision: str = '089de6f1f16d'
down_revision: Union[str, Sequence[str], None] = '9a2a45ab63ac'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Upgrade schema."""
    # Create the enum type first
    nodehealth = postgresql.ENUM('HEALTHY', 'NEEDS_ATTENTION', 'CRITICAL', 'NOT_STARTED', name='nodehealth', create_type=False)
    nodehealth.create(op.get_bind(), checkfirst=True)

    # Add new columns to nodes table
    op.add_column('nodes', sa.Column('health', sa.Enum('HEALTHY', 'NEEDS_ATTENTION', 'CRITICAL', 'NOT_STARTED', name='nodehealth'), nullable=True))
    op.add_column('nodes', sa.Column('is_archived', sa.Boolean(), nullable=True, server_default='false'))
    op.add_column('nodes', sa.Column('sort_order', sa.Integer(), nullable=True, server_default='0'))

    # Set defaults for existing rows
    op.execute("UPDATE nodes SET health = 'NOT_STARTED' WHERE health IS NULL")
    op.execute("UPDATE nodes SET is_archived = false WHERE is_archived IS NULL")
    op.execute("UPDATE nodes SET sort_order = 0 WHERE sort_order IS NULL")

    # Make columns non-nullable after setting defaults
    op.alter_column('nodes', 'health', nullable=False)
    op.alter_column('nodes', 'is_archived', nullable=False)
    op.alter_column('nodes', 'sort_order', nullable=False)

    # Change this_week_focus from TEXT to JSONB
    op.alter_column('nodes', 'this_week_focus',
               existing_type=sa.TEXT(),
               type_=postgresql.JSONB(astext_type=sa.Text()),
               existing_nullable=True,
               postgresql_using='this_week_focus::jsonb')


def downgrade() -> None:
    """Downgrade schema."""
    op.alter_column('nodes', 'this_week_focus',
               existing_type=postgresql.JSONB(astext_type=sa.Text()),
               type_=sa.TEXT(),
               existing_nullable=True)
    op.drop_column('nodes', 'sort_order')
    op.drop_column('nodes', 'is_archived')
    op.drop_column('nodes', 'health')

    # Drop the enum type
    nodehealth = postgresql.ENUM('HEALTHY', 'NEEDS_ATTENTION', 'CRITICAL', 'NOT_STARTED', name='nodehealth')
    nodehealth.drop(op.get_bind(), checkfirst=True)
