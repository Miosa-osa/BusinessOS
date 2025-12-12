"""add_client_id_to_contexts

Revision ID: e2eb93e4ecbe
Revises: add_context_to_artifacts
Create Date: 2025-12-12 22:17:43.620853

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

# revision identifiers, used by Alembic.
revision: str = 'e2eb93e4ecbe'
down_revision: Union[str, Sequence[str], None] = 'add_context_to_artifacts'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Upgrade schema."""
    # Add client_id column to contexts table
    op.add_column('contexts', sa.Column('client_id', sa.UUID(), nullable=True))
    op.create_index(op.f('ix_contexts_client_id'), 'contexts', ['client_id'], unique=False)
    op.create_foreign_key('fk_contexts_client_id', 'contexts', 'clients', ['client_id'], ['id'], ondelete='SET NULL')


def downgrade() -> None:
    """Downgrade schema."""
    op.drop_constraint('fk_contexts_client_id', 'contexts', type_='foreignkey')
    op.drop_index(op.f('ix_contexts_client_id'), table_name='contexts')
    op.drop_column('contexts', 'client_id')
