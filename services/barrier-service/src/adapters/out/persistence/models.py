from datetime import datetime
from uuid import UUID, uuid4

from sqlalchemy import (
    Column, String, Text, Integer, Boolean, DateTime, ForeignKey, Enum as SQLEnum, Index, func
)
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import relationship, declarative_base

Base = declarative_base()


class BarrierModel(Base):
    __tablename__ = "barriers"

    id = Column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    type = Column(SQLEnum("high_curb", "broken_elevator", "closed_sidewalk", "stairs", "pothole", "uneven_surface", "parked_car", name="barrier_type_enum"), nullable=False)
    latitude = Column(String, nullable=False)
    longitude = Column(String, nullable=False)
    description = Column(Text, default="")
    severity = Column(Integer, nullable=False, default=2)
    status = Column(SQLEnum("pending", "approved", "rejected", "resolved", name="barrier_status_enum"), nullable=False, default="pending")
    reporter_id = Column(PGUUID(as_uuid=True), nullable=True)
    moderator_id = Column(PGUUID(as_uuid=True), nullable=True)
    created_at = Column(DateTime(timezone=True), nullable=False, default=func.now())
    updated_at = Column(DateTime(timezone=True), nullable=False, default=func.now(), onupdate=func.now())
    approved_at = Column(DateTime(timezone=True), nullable=True)
    resolved_at = Column(DateTime(timezone=True), nullable=True)

    photos = relationship("BarrierPhotoModel", back_populates="barrier", cascade="all, delete-orphan")
    confirmations = relationship("BarrierConfirmationModel", back_populates="barrier", cascade="all, delete-orphan")
    complaints = relationship("BarrierComplaintModel", back_populates="barrier", cascade="all, delete-orphan")

    __table_args__ = (
        Index("ix_barriers_status_type", "status", "type"),
        Index("ix_barriers_coordinates", "latitude", "longitude"),
        Index("ix_barriers_reporter", "reporter_id"),
    )


class BarrierPhotoModel(Base):
    __tablename__ = "barrier_photos"

    id = Column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    barrier_id = Column(PGUUID(as_uuid=True), ForeignKey("barriers.id", ondelete="CASCADE"), nullable=False)
    s3_key = Column(String, nullable=False)
    original_filename = Column(String, nullable=False)
    content_type = Column(String, nullable=False)
    size_bytes = Column(Integer, nullable=False)
    uploaded_by = Column(PGUUID(as_uuid=True), nullable=False)
    created_at = Column(DateTime(timezone=True), nullable=False, default=func.now())

    barrier = relationship("BarrierModel", back_populates="photos")


class BarrierConfirmationModel(Base):
    __tablename__ = "barrier_confirmations"

    id = Column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    barrier_id = Column(PGUUID(as_uuid=True), ForeignKey("barriers.id", ondelete="CASCADE"), nullable=False)
    user_id = Column(PGUUID(as_uuid=True), nullable=False)
    created_at = Column(DateTime(timezone=True), nullable=False, default=func.now())

    barrier = relationship("BarrierModel", back_populates="confirmations")

    __table_args__ = (
        Index("ix_barrier_confirmations_barrier_user", "barrier_id", "user_id", unique=True),
    )


class BarrierComplaintModel(Base):
    __tablename__ = "barrier_complaints"

    id = Column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    barrier_id = Column(PGUUID(as_uuid=True), ForeignKey("barriers.id", ondelete="CASCADE"), nullable=False)
    user_id = Column(PGUUID(as_uuid=True), nullable=False)
    reason = Column(Text, nullable=False)
    created_at = Column(DateTime(timezone=True), nullable=False, default=func.now())

    barrier = relationship("BarrierModel", back_populates="complaints")

    __table_args__ = (
        Index("ix_barrier_complaints_barrier_user", "barrier_id", "user_id", unique=True),
    )