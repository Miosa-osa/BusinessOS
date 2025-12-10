"""add_artifact_summary_column

Revision ID: 9a2a45ab63ac
Revises: 66792c95f408
Create Date: 2025-12-10 23:08:08.691574

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '9a2a45ab63ac'
down_revision: Union[str, Sequence[str], None] = '66792c95f408'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Upgrade schema."""
    # Add summary column to artifacts table
    op.add_column('artifacts', sa.Column('summary', sa.String(length=500), nullable=True))


def downgrade() -> None:
    """Downgrade schema."""
    op.drop_column('artifacts', 'summary')
