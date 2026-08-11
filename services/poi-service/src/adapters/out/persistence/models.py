from datetime import datetime
from uuid import UUID, uuid4

from sqlalchemy import Column, String, Text, Float, Boolean, DateTime, ForeignKey, Enum as SQLEnum, Index, func, ARRAY
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import declarative_base

Base = declarative_base()


class POIModel(Base):
    __tablename__ = "points_of_interest"

    id = Column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    name = Column(String(200), nullable=False)
    category = Column(SQLEnum(*[c.value for c in __import__('src.domain.entities.poi', fromlist=['POICategory']).POICategory], name="poi_category_enum"), nullable=False)
    latitude = Column(Float, nullable=False)
    longitude = Column(Float, nullable=False)
    address = Column(String(500), nullable=False)
    phone = Column(String(50), default="")
    website = Column(String(200), default="")
    opening_hours = Column(String(200), default="")
    accessibility_features = Column(ARRAY(String), default=[])
    entrance_step_height_cm = Column(Float, nullable=True)
    door_width_cm = Column(Float, nullable=True)
    has_accessible_toilet = Column(Boolean, default=False)
    accessibility_notes = Column(Text, default="")
    owner_id = Column(PGUUID(as_uuid=True), nullable=True)
    is_verified = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), nullable=False, default=func.now())
    updated_at = Column(DateTime(timezone=True), nullable=False, default=func.now(), onupdate=func.now())

    __table_args__ = (
        Index("ix_poi_category", "category"),
        Index("ix_poi_owner", "owner_id"),
        Index("ix_poi_coordinates", "latitude", "longitude"),
    )