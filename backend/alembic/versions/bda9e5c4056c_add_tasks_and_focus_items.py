"""add_tasks_and_focus_items

Revision ID: bda9e5c4056c
Revises: 60c82a6392b8
Create Date: 2025-12-10 19:00:46.459437

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'bda9e5c4056c'
down_revision: Union[str, Sequence[str], None] = '60c82a6392b8'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Upgrade schema."""
    from sqlalchemy.dialects import postgresql

    # Create task priority enum
    taskpriority = postgresql.ENUM('critical', 'high', 'medium', 'low', name='taskpriority', create_type=False)
    taskpriority.create(op.get_bind(), checkfirst=True)

    # Create task status enum
    taskstatus = postgresql.ENUM('todo', 'in_progress', 'done', 'cancelled', name='taskstatus', create_type=False)
    taskstatus.create(op.get_bind(), checkfirst=True)

    # Create tasks table
    op.create_table('tasks',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.String(length=255), nullable=False),
        sa.Column('title', sa.String(length=500), nullable=False),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('status', postgresql.ENUM('todo', 'in_progress', 'done', 'cancelled', name='taskstatus', create_type=False), nullable=False, server_default='todo'),
        sa.Column('priority', postgresql.ENUM('critical', 'high', 'medium', 'low', name='taskpriority', create_type=False), nullable=False, server_default='medium'),
        sa.Column('due_date', sa.DateTime(), nullable=True),
        sa.Column('completed_at', sa.DateTime(), nullable=True),
        sa.Column('project_id', sa.UUID(), nullable=True),
        sa.Column('assignee_id', sa.UUID(), nullable=True),
        sa.Column('created_at', sa.DateTime(), nullable=False, server_default=sa.text('NOW()')),
        sa.Column('updated_at', sa.DateTime(), nullable=False, server_default=sa.text('NOW()')),
        sa.ForeignKeyConstraint(['project_id'], ['projects.id'], ondelete='SET NULL'),
        sa.ForeignKeyConstraint(['assignee_id'], ['team_members.id'], ondelete='SET NULL'),
        sa.PrimaryKeyConstraint('id')
    )
    op.create_index(op.f('ix_tasks_user_id'), 'tasks', ['user_id'], unique=False)

    # Create focus_items table
    op.create_table('focus_items',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.String(length=255), nullable=False),
        sa.Column('text', sa.String(length=500), nullable=False),
        sa.Column('completed', sa.Boolean(), nullable=False, server_default='false'),
        sa.Column('focus_date', sa.DateTime(), nullable=False, server_default=sa.text('NOW()')),
        sa.Column('created_at', sa.DateTime(), nullable=False, server_default=sa.text('NOW()')),
        sa.Column('updated_at', sa.DateTime(), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id')
    )
    op.create_index(op.f('ix_focus_items_user_id'), 'focus_items', ['user_id'], unique=False)


def downgrade() -> None:
    """Downgrade schema."""
    op.drop_index(op.f('ix_focus_items_user_id'), table_name='focus_items')
    op.drop_table('focus_items')
    op.drop_index(op.f('ix_tasks_user_id'), table_name='tasks')
    op.drop_table('tasks')
    op.execute("DROP TYPE IF EXISTS taskstatus")
    op.execute("DROP TYPE IF EXISTS taskpriority")
