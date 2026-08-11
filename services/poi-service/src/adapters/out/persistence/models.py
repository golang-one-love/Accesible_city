from datetime import datetime
from uuid import UUID, uuid4

from sqlalchemy import ARRAY, Boolean, DateTime, Float, Index, String, Text, func
from sqlalchemy import Enum as SQLEnum
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import Mapped, declarative_base, mapped_column

from src.domain.entities.poi import POICategory

Base = declarative_base()


class POIModel(Base):
    __tablename__ = "points_of_interest"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    name: Mapped[str] = mapped_column(String(200), nullable=False)
    category: Mapped[str] = mapped_column(
        SQLEnum(*[c.value for c in POICategory], name="poi_category_enum"), nullable=False
    )
    latitude: Mapped[float] = mapped_column(Float, nullable=False)
    longitude: Mapped[float] = mapped_column(Float, nullable=False)
    address: Mapped[str] = mapped_column(String(500), nullable=False)
    phone: Mapped[str] = mapped_column(String(50), default="")
    website: Mapped[str] = mapped_column(String(200), default="")
    opening_hours: Mapped[str] = mapped_column(String(200), default="")
    accessibility_features: Mapped[list[str]] = mapped_column(ARRAY(String), default=[])
    entrance_step_height_cm: Mapped[float | None] = mapped_column(Float, nullable=True)
    door_width_cm: Mapped[float | None] = mapped_column(Float, nullable=True)
    has_accessible_toilet: Mapped[bool] = mapped_column(Boolean, default=False)
    accessibility_notes: Mapped[str] = mapped_column(Text, default="")
    owner_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=True)
    is_verified: Mapped[bool] = mapped_column(Boolean, default=False)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, default=func.now(), onupdate=func.now()
    )

    __table_args__ = (
        Index("ix_poi_category", "category"),
        Index("ix_poi_owner", "owner_id"),
        Index("ix_poi_coordinates", "latitude", "longitude"),
    )
