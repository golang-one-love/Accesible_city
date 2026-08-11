from pydantic import BaseModel, Field
from typing import Optional, List
from uuid import UUID
from datetime import datetime
from enum import Enum


class BarrierTypeStr(str, Enum):
    HIGH_CURB = "high_curb"
    BROKEN_ELEVATOR = "broken_elevator"
    CLOSED_SIDEWALK = "closed_sidewalk"
    STAIRS = "stairs"
    POTHOLE = "pothole"
    UNEVEN_SURFACE = "uneven_surface"
    PARKED_CAR = "parked_car"


class BarrierStatusStr(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    RESOLVED = "resolved"


class SeverityInt(int, Enum):
    LOW = 1
    MEDIUM = 2
    HIGH = 3
    CRITICAL = 4
    BLOCKING = 5


class CoordinatesDTO(BaseModel):
    latitude: float = Field(ge=-90, le=90)
    longitude: float = Field(ge=-180, le=180)


class CreateBarrierRequest(BaseModel):
    type: BarrierTypeStr
    coordinates: CoordinatesDTO
    description: str = Field(default="", max_length=2000)
    severity: SeverityInt = Field(default=SeverityInt.MEDIUM)


class BarrierPhotoResponse(BaseModel):
    id: UUID
    barrier_id: UUID
    s3_key: str
    original_filename: str
    content_type: str
    size_bytes: int
    uploaded_by: UUID
    created_at: datetime

    class Config:
        from_attributes = True


class BarrierResponse(BaseModel):
    id: UUID
    type: BarrierTypeStr
    coordinates: CoordinatesDTO
    description: str
    severity: SeverityInt
    status: BarrierStatusStr
    reporter_id: Optional[UUID]
    moderator_id: Optional[UUID]
    created_at: datetime
    updated_at: datetime
    approved_at: Optional[datetime]
    resolved_at: Optional[datetime]
    photos: List[BarrierPhotoResponse] = []
    confirmations_count: int = 0
    complaints_count: int = 0

    class Config:
        from_attributes = True


class BarrierListResponse(BaseModel):
    barriers: List[BarrierResponse]
    total: int
    limit: int
    offset: int


class UploadPhotoResponse(BaseModel):
    id: UUID
    barrier_id: UUID
    s3_key: str
    original_filename: str
    content_type: str
    size_bytes: int
    presigned_url: str


class ConfirmBarrierResponse(BaseModel):
    id: UUID
    barrier_id: UUID
    user_id: UUID
    created_at: datetime


class ComplainBarrierRequest(BaseModel):
    reason: str = Field(min_length=1, max_length=1000)


class ComplainBarrierResponse(BaseModel):
    id: UUID
    barrier_id: UUID
    user_id: UUID
    reason: str
    created_at: datetime


class ErrorResponse(BaseModel):
    error: str
    detail: Optional[str] = None