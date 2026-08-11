from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from uuid import UUID, uuid4


class ModerationStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"


@dataclass
class ModerationRequest:
    id: UUID = field(default_factory=uuid4)
    barrier_id: UUID | None = None
    reporter_id: UUID | None = None
    status: ModerationStatus = ModerationStatus.PENDING
    moderator_id: UUID | None = None
    moderator_comment: str = ""
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)
    reviewed_at: datetime | None = None

    def approve(self, moderator_id: UUID, comment: str = "") -> None:
        if self.status != ModerationStatus.PENDING:
            raise ValueError("Can only approve pending requests")
        if self.reporter_id == moderator_id:
            raise ValueError("Cannot moderate own request")
        self.status = ModerationStatus.APPROVED
        self.moderator_id = moderator_id
        self.moderator_comment = comment
        self.reviewed_at = datetime.utcnow()
        self.updated_at = datetime.utcnow()

    def reject(self, moderator_id: UUID, comment: str = "") -> None:
        if self.status != ModerationStatus.PENDING:
            raise ValueError("Can only reject pending requests")
        if self.reporter_id == moderator_id:
            raise ValueError("Cannot moderate own request")
        self.status = ModerationStatus.REJECTED
        self.moderator_id = moderator_id
        self.moderator_comment = comment
        self.reviewed_at = datetime.utcnow()
        self.updated_at = datetime.utcnow()