from pydantic import BaseModel, Field
from typing import Optional, List
from uuid import UUID
from datetime import datetime
from enum import Enum


class ModerationStatusStr(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"


class ModerationRequestResponse(BaseModel):
    id: UUID
    barrier_id: UUID
    reporter_id: UUID
    status: ModerationStatusStr
    moderator_id: Optional[UUID]
    moderator_comment: str
    created_at: datetime
    updated_at: datetime
    reviewed_at: Optional[datetime]

    class Config:
        from_attributes = True


class ModerationQueueResponse(BaseModel):
    requests: List[ModerationRequestResponse]
    total: int
    limit: int
    offset: int


class ModerationActionRequest(BaseModel):
    comment: str = Field(default="", max_length=1000)


class ModerationActionResponse(BaseModel):
    id: UUID
    barrier_id: UUID
    status: ModerationStatusStr
    moderator_id: UUID
    reviewed_at: datetime


class ErrorResponse(BaseModel):
    error: str
    detail: Optional[str] = None