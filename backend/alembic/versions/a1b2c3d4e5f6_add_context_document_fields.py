"""Add context document editor fields

Revision ID: a1b2c3d4e5f6
Revises: 3077021564fc
Create Date: 2025-12-12 20:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision: str = 'a1b2c3d4e5f6'
down_revision: Union[str, Sequence[str], None] = '3077021564fc'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Add document editor fields to contexts table."""
    # Add new columns for document editor functionality
    op.add_column('contexts', sa.Column('blocks', postgresql.JSONB(astext_type=sa.Text()), nullable=True, server_default='[]'))
    op.add_column('contexts', sa.Column('cover_image', sa.String(length=500), nullable=True))
    op.add_column('contexts', sa.Column('icon', sa.String(length=50), nullable=True))
    op.add_column('contexts', sa.Column('parent_id', sa.UUID(), nullable=True))
    op.add_column('contexts', sa.Column('is_template', sa.Boolean(), nullable=False, server_default='false'))
    op.add_column('contexts', sa.Column('is_archived', sa.Boolean(), nullable=False, server_default='false'))
    op.add_column('contexts', sa.Column('last_edited_at', sa.DateTime(timezone=True), nullable=True))
    op.add_column('contexts', sa.Column('word_count', sa.Integer(), nullable=False, server_default='0'))
    op.add_column('contexts', sa.Column('is_public', sa.Boolean(), nullable=False, server_default='false'))
    op.add_column('contexts', sa.Column('share_id', sa.String(length=32), nullable=True))

    # Create indexes
    op.create_index('ix_contexts_parent_id', 'contexts', ['parent_id'], unique=False)
    op.create_index('ix_contexts_is_archived', 'contexts', ['is_archived'], unique=False)
    op.create_index('ix_contexts_share_id', 'contexts', ['share_id'], unique=True)

    # Add foreign key for parent_id (self-referential)
    op.create_foreign_key(
        'fk_contexts_parent_id',
        'contexts', 'contexts',
        ['parent_id'], ['id'],
        ondelete='SET NULL'
    )

    # Add 'document' to the ContextType enum
    # PostgreSQL enum modification
    op.execute("ALTER TYPE contexttype ADD VALUE IF NOT EXISTS 'document'")


def downgrade() -> None:
    """Remove document editor fields from contexts table."""
    # Drop foreign key
    op.drop_constraint('fk_contexts_parent_id', 'contexts', type_='foreignkey')

    # Drop indexes
    op.drop_index('ix_contexts_share_id', table_name='contexts')
    op.drop_index('ix_contexts_is_archived', table_name='contexts')
    op.drop_index('ix_contexts_parent_id', table_name='contexts')

    # Drop columns
    op.drop_column('contexts', 'share_id')
    op.drop_column('contexts', 'is_public')
    op.drop_column('contexts', 'word_count')
    op.drop_column('contexts', 'last_edited_at')
    op.drop_column('contexts', 'is_archived')
    op.drop_column('contexts', 'is_template')
    op.drop_column('contexts', 'parent_id')
    op.drop_column('contexts', 'icon')
    op.drop_column('contexts', 'cover_image')
    op.drop_column('contexts', 'blocks')

    # Note: Cannot easily remove enum value in PostgreSQL
    # The 'document' enum value will remain
