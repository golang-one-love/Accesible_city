"""create poi table

Revision ID: 001
Revises: 
Create Date: 2026-08-10

"""
import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects.postgresql import ARRAY, UUID

revision = '001'
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.execute("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
    op.execute("CREATE EXTENSION IF NOT EXISTS postgis")

    poi_category_enum = sa.Enum(
        'restaurant', 'cafe', 'shop', 'pharmacy', 'hospital', 'clinic',
        'bank', 'post_office', 'government', 'park', 'museum', 'theater',
        'library', 'school', 'university', 'hotel', 'transport', 'other',
        name='poi_category_enum'
    )
    poi_category_enum.create(op.get_bind(), checkfirst=True)

    op.create_table(
        'points_of_interest',
        sa.Column('id', UUID(as_uuid=True), primary_key=True, server_default=sa.text('gen_random_uuid()')),
        sa.Column('name', sa.String(200), nullable=False),
        sa.Column('category', poi_category_enum, nullable=False),
        sa.Column('latitude', sa.Float, nullable=False),
        sa.Column('longitude', sa.Float, nullable=False),
        sa.Column('address', sa.String(500), nullable=False),
        sa.Column('phone', sa.String(50), server_default=''),
        sa.Column('website', sa.String(200), server_default=''),
        sa.Column('opening_hours', sa.String(200), server_default=''),
        sa.Column('accessibility_features', ARRAY(sa.String), server_default='{}'),
        sa.Column('entrance_step_height_cm', sa.Float, nullable=True),
        sa.Column('door_width_cm', sa.Float, nullable=True),
        sa.Column('has_accessible_toilet', sa.Boolean, server_default='false'),
        sa.Column('accessibility_notes', sa.Text, server_default=''),
        sa.Column('owner_id', UUID(as_uuid=True), nullable=True),
        sa.Column('is_verified', sa.Boolean, server_default='false'),
        sa.Column('created_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
        sa.Column('updated_at', sa.DateTime(timezone=True), nullable=False, server_default=sa.func.now()),
    )
    op.create_index('ix_poi_category', 'points_of_interest', ['category'])
    op.create_index('ix_poi_owner', 'points_of_interest', ['owner_id'])
    op.create_index('ix_poi_coordinates', 'points_of_interest', ['latitude', 'longitude'])


def downgrade() -> None:
    op.drop_table('points_of_interest')
    sa.Enum(name='poi_category_enum').drop(op.get_bind(), checkfirst=True)