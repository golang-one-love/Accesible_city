"""create moderation queue table

Revision ID: 001
Revises: 
Create Date: 2026-08-10

"""
import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects.postgresql import UUID

revision = '001'
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.execute("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

    moderation_status_enum = sa.Enum(
        'pending', 'approved', 'rejected',
        name='moderation_status_enum'
    )
    moderation_status_enum.create(op.get_bind(), checkfirst=True)

    op.create_table(
        'moderation_queue',
        sa.Column('id', UUID(as_uuid=True), primary_key=True, server_default=sa.text('gen_random_uuid()')),
        sa.Column('barrier_id', UUID(as_uuid=True), nullable=False, unique=True),
        sa.Column('reporter_id', UUID(as_uuid=True), nullable=False),
        sa.Column('status', moderation_status_enum, nullable=False, server_default='pending'),
        sa.Column('moderator_id', UUID(as_uuid=True), nullable=True),
        sa.Column('moderator_comment', sa.Text, server_default=''),
        sa.Column('created_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        sa.Column('updated_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        sa.Column('reviewed_at', sa.DateTime(timezone=True), nullable=True),
    )
    op.create_index('ix_moderation_queue_status', 'moderation_queue', ['status'])
    op.create_index('ix_moderation_queue_reporter', 'moderation_queue', ['reporter_id'])


def downgrade() -> None:
    op.drop_table('moderation_queue')
    sa.Enum(name='moderation_status_enum').drop(op.get_bind(), checkfirst=True)