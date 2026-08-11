from datetime import datetime
from uuid import UUID, uuid4

from sqlalchemy import DateTime, Index, Text, func
from sqlalchemy import Enum as SQLEnum
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import Mapped, declarative_base, mapped_column

Base = declarative_base()


class ModerationRequestModel(Base):
    __tablename__ = "moderation_queue"

    id: Mapped[UUID] = mapped_column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    barrier_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=False, unique=True)
    reporter_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=False)
    status: Mapped[str] = mapped_column(
        SQLEnum("pending", "approved", "rejected", name="moderation_status_enum"),
        nullable=False,
        default="pending",
    )
    moderator_id: Mapped[UUID | None] = mapped_column(PGUUID(as_uuid=True), nullable=True)
    moderator_comment: Mapped[str] = mapped_column(Text, default="")
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, default=func.now(), onupdate=func.now()
    )
    reviewed_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    __table_args__ = (
        Index("ix_moderation_queue_status", "status"),
        Index("ix_moderation_queue_reporter", "reporter_id"),
    )
