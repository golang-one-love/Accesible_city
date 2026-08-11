from datetime import datetime
from uuid import UUID, uuid4

from sqlalchemy import DateTime, ForeignKey, Index, Integer, String, Text, func
from sqlalchemy import Enum as SQLEnum
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import Mapped, declarative_base, mapped_column, relationship

Base = declarative_base()


class BarrierModel(Base):
    __tablename__ = "barriers"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    type: Mapped[str] = mapped_column(
        SQLEnum(
            "high_curb",
            "broken_elevator",
            "closed_sidewalk",
            "stairs",
            "pothole",
            "uneven_surface",
            "parked_car",
            name="barrier_type_enum",
        ),
        nullable=False,
    )
    latitude: Mapped[str] = mapped_column(String, nullable=False)
    longitude: Mapped[str] = mapped_column(String, nullable=False)
    description: Mapped[str] = mapped_column(Text, default="")
    severity: Mapped[int] = mapped_column(Integer, nullable=False, default=2)
    status: Mapped[str] = mapped_column(
        SQLEnum("pending", "approved", "rejected", "resolved", name="barrier_status_enum"),
        nullable=False,
        default="pending",
    )
    reporter_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=True)
    moderator_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, default=func.now(), onupdate=func.now()
    )
    approved_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    resolved_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    photos: Mapped[list["BarrierPhotoModel"]] = relationship(
        "BarrierPhotoModel", back_populates="barrier", cascade="all, delete-orphan"
    )
    confirmations: Mapped[list["BarrierConfirmationModel"]] = relationship(
        "BarrierConfirmationModel", back_populates="barrier", cascade="all, delete-orphan"
    )
    complaints: Mapped[list["BarrierComplaintModel"]] = relationship(
        "BarrierComplaintModel", back_populates="barrier", cascade="all, delete-orphan"
    )

    __table_args__ = (
        Index("ix_barriers_status_type", "status", "type"),
        Index("ix_barriers_coordinates", "latitude", "longitude"),
        Index("ix_barriers_reporter", "reporter_id"),
    )


class BarrierPhotoModel(Base):
    __tablename__ = "barrier_photos"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    barrier_id: Mapped[UUID | None] = mapped_column(
        PGUUID(as_uuid=True), ForeignKey("barriers.id", ondelete="CASCADE"), nullable=False
    )
    s3_key: Mapped[str] = mapped_column(String, nullable=False)
    original_filename: Mapped[str] = mapped_column(String, nullable=False)
    content_type: Mapped[str] = mapped_column(String, nullable=False)
    size_bytes: Mapped[int] = mapped_column(Integer, nullable=False)
    uploaded_by: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=func.now())

    barrier: Mapped["BarrierModel"] = relationship("BarrierModel", back_populates="photos")


class BarrierConfirmationModel(Base):
    __tablename__ = "barrier_confirmations"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    barrier_id: Mapped[UUID | None] = mapped_column(
        PGUUID(as_uuid=True), ForeignKey("barriers.id", ondelete="CASCADE"), nullable=False
    )
    user_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=func.now())

    barrier: Mapped["BarrierModel"] = relationship("BarrierModel", back_populates="confirmations")

    __table_args__ = (
        Index("ix_barrier_confirmations_barrier_user", "barrier_id", "user_id", unique=True),
    )


class BarrierComplaintModel(Base):
    __tablename__ = "barrier_complaints"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    barrier_id: Mapped[UUID | None] = mapped_column(
        PGUUID(as_uuid=True), ForeignKey("barriers.id", ondelete="CASCADE"), nullable=False
    )
    user_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=False)
    reason: Mapped[str] = mapped_column(Text, nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=func.now())

    barrier: Mapped["BarrierModel"] = relationship("BarrierModel", back_populates="complaints")

    __table_args__ = (
        Index("ix_barrier_complaints_barrier_user", "barrier_id", "user_id", unique=True),
    )
