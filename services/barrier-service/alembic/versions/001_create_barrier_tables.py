"""create barrier tables

Revision ID: 001
Revises: 
Create Date: 2026-08-10

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects.postgresql import UUID


revision = '001'
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.execute("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
    op.execute("CREATE EXTENSION IF NOT EXISTS postgis")

    barrier_type_enum = sa.Enum(
        'high_curb', 'broken_elevator', 'closed_sidewalk', 'stairs',
        'pothole', 'uneven_surface', 'parked_car',
        name='barrier_type_enum'
    )
    barrier_status_enum = sa.Enum(
        'pending', 'approved', 'rejected', 'resolved',
        name='barrier_status_enum'
    )

    barrier_type_enum.create(op.get_bind(), checkfirst=True)
    barrier_status_enum.create(op.get_bind(), checkfirst=True)

    op.create_table(
        'barriers',
        sa.Column('id', UUID(as_uuid=True), primary_key=True, server_default=sa.text('gen_random_uuid()')),
        sa.Column('type', barrier_type_enum, nullable=False),
        sa.Column('latitude', sa.String, nullable=False),
        sa.Column('longitude', sa.String, nullable=False),
        sa.Column('description', sa.Text, server_default=''),
        sa.Column('severity', sa.Integer, nullable=False, server_default='2'),
        sa.Column('status', barrier_status_enum, nullable=False, server_default='pending'),
        sa.Column('reporter_id', UUID(as_uuid=True), nullable=True),
        sa.Column('moderator_id', UUID(as_uuid=True), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        sa.Column('updated_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        sa.Column('approved_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('resolved_at', sa.DateTime(timezone=True), nullable=True),
    )
    op.create_index('ix_barriers_status_type', 'barriers', ['status', 'type'])
    op.create_index('ix_barriers_coordinates', 'barriers', ['latitude', 'longitude'])
    op.create_index('ix_barriers_reporter', 'barriers', ['reporter_id'])

    op.create_table(
        'barrier_photos',
        sa.Column('id', UUID(as_uuid=True), primary_key=True, server_default=sa.text('gen_random_uuid()')),
        sa.Column('barrier_id', UUID(as_uuid=True), sa.ForeignKey('barriers.id', ondelete='CASCADE'), nullable=False),
        sa.Column('s3_key', sa.String, nullable=False),
        sa.Column('original_filename', sa.String, nullable=False),
        sa.Column('content_type', sa.String, nullable=False),
        sa.Column('size_bytes', sa.Integer, nullable=False),
        sa.Column('uploaded_by', UUID(as_uuid=True), nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
    )

    op.create_table(
        'barrier_confirmations',
        sa.Column('id', UUID(as_uuid=True), primary_key=True, server_default=sa.text('gen_random_uuid()')),
        sa.Column('barrier_id', UUID(as_uuid=True), sa.ForeignKey('barriers.id', ondelete='CASCADE'), nullable=False),
        sa.Column('user_id', UUID(as_uuid=True), nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
    )
    op.create_index('ix_barrier_confirmations_barrier_user', 'barrier_confirmations', ['barrier_id', 'user_id'], unique=True)

    op.create_table(
        'barrier_complaints',
        sa.Column('id', UUID(as_uuid=True), primary_key=True, server_default=sa.text('gen_random_uuid()')),
        sa.Column('barrier_id', UUID(as_uuid=True), sa.ForeignKey('barriers.id', ondelete='CASCADE'), nullable=False),
        sa.Column('user_id', UUID(as_uuid=True), nullable=False),
        sa.Column('reason', sa.Text, nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
    )
    op.create_index('ix_barrier_complaints_barrier_user', 'barrier_complaints', ['barrier_id', 'user_id'], unique=True)


def downgrade() -> None:
    op.drop_table('barrier_complaints')
    op.drop_table('barrier_confirmations')
    op.drop_table('barrier_photos')
    op.drop_table('barriers')

    sa.Enum(name='barrier_status_enum').drop(op.get_bind(), checkfirst=True)
    sa.Enum(name='barrier_type_enum').drop(op.get_bind(), checkfirst=True)