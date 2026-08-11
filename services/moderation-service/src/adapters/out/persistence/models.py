from datetime import datetime
from uuid import UUID, uuid4

from sqlalchemy import Column, String, Text, DateTime, ForeignKey, Enum as SQLEnum, Index, func
from sqlalchemy.dialects.postgresql import UUID as PGUUID
from sqlalchemy.orm import relationship, declarative_base

Base = declarative_base()


class ModerationRequestModel(Base):
    __tablename__ = "moderation_queue"

    id = Column(PGUUID(as_uuid=True), primary_key=True, default=uuid4)
    barrier_id = Column(PGUUID(as_uuid=True), nullable=False, unique=True)
    reporter_id = Column(PGUUID(as_uuid=True), nullable=False)
    status = Column(SQLEnum("pending", "approved", "rejected", name="moderation_status_enum"), nullable=False, default="pending")
    moderator_id = Column(PGUUID(as_uuid=True), nullable=True)
    moderator_comment = Column(Text, default="")
    created_at = Column(DateTime(timezone=True), nullable=False, default=func.now())
    updated_at = Column(DateTime(timezone=True), nullable=False, default=func.now(), onupdate=func.now())
    reviewed_at = Column(DateTime(timezone=True), nullable=True)

    __table_args__ = (
        Index("ix_moderation_queue_status", "status"),
        Index("ix_moderation_queue_reporter", "reporter_id"),
    )