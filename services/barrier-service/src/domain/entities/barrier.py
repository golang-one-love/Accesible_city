from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from uuid import UUID, uuid4

from ..value_objects.coordinates import Coordinates


class BarrierType(str, Enum):
    HIGH_CURB = "high_curb"
    BROKEN_ELEVATOR = "broken_elevator"
    CLOSED_SIDEWALK = "closed_sidewalk"
    STAIRS = "stairs"
    POTHOLE = "pothole"
    UNEVEN_SURFACE = "uneven_surface"
    PARKED_CAR = "parked_car"


class BarrierStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    RESOLVED = "resolved"


class Severity(int, Enum):
    LOW = 1
    MEDIUM = 2
    HIGH = 3
    CRITICAL = 4
    BLOCKING = 5


@dataclass
class Barrier:
    coordinates: Coordinates
    id: UUID = field(default_factory=uuid4)
    type: BarrierType = BarrierType.HIGH_CURB
    description: str = ""
    severity: Severity = Severity.MEDIUM
    status: BarrierStatus = BarrierStatus.PENDING
    reporter_id: UUID | None = None
    moderator_id: UUID | None = None
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)
    approved_at: datetime | None = None
    resolved_at: datetime | None = None

    def approve(self, moderator_id: UUID) -> None:
        if self.status != BarrierStatus.PENDING:
            raise ValueError("Can only approve pending barriers")
        self.status = BarrierStatus.APPROVED
        self.moderator_id = moderator_id
        self.approved_at = datetime.utcnow()
        self.updated_at = datetime.utcnow()

    def reject(self, moderator_id: UUID) -> None:
        if self.status != BarrierStatus.PENDING:
            raise ValueError("Can only reject pending barriers")
        self.status = BarrierStatus.REJECTED
        self.moderator_id = moderator_id
        self.updated_at = datetime.utcnow()

    def resolve(self) -> None:
        if self.status != BarrierStatus.APPROVED:
            raise ValueError("Can only resolve approved barriers")
        self.status = BarrierStatus.RESOLVED
        self.resolved_at = datetime.utcnow()
        self.updated_at = datetime.utcnow()

    def add_confirmation(self) -> None:
        self.updated_at = datetime.utcnow()

    def add_complaint(self) -> None:
        self.updated_at = datetime.utcnow()


@dataclass
class BarrierPhoto:
    id: UUID = field(default_factory=uuid4)
    barrier_id: UUID | None = None
    s3_key: str = ""
    original_filename: str = ""
    content_type: str = ""
    size_bytes: int = 0
    uploaded_by: UUID | None = None
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class BarrierConfirmation:
    id: UUID = field(default_factory=uuid4)
    barrier_id: UUID | None = None
    user_id: UUID | None = None
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class BarrierComplaint:
    id: UUID = field(default_factory=uuid4)
    barrier_id: UUID | None = None
    user_id: UUID | None = None
    reason: str = ""
    created_at: datetime = field(default_factory=datetime.utcnow)