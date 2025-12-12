"""Add context_id to artifacts table

Revision ID: add_context_to_artifacts
Revises:
Create Date: 2025-01-01

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision: str = 'add_context_to_artifacts'
down_revision: Union[str, None] = 'a1b2c3d4e5f6'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Add context_id column to artifacts table
    op.add_column('artifacts', sa.Column('context_id', postgresql.UUID(as_uuid=True), nullable=True))
    op.create_foreign_key(
        'fk_artifacts_context_id',
        'artifacts', 'contexts',
        ['context_id'], ['id'],
        ondelete='SET NULL'
    )
    op.create_index('ix_artifacts_context_id', 'artifacts', ['context_id'])


def downgrade() -> None:
    op.drop_index('ix_artifacts_context_id', table_name='artifacts')
    op.drop_constraint('fk_artifacts_context_id', 'artifacts', type_='foreignkey')
    op.drop_column('artifacts', 'context_id')
