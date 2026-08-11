from dataclasses import dataclass, field
from datetime import datetime
from typing import Any
from uuid import UUID


@dataclass
class DomainEvent:
    event_type: str
    aggregate_id: UUID
    timestamp: datetime = field(default_factory=datetime.utcnow)
    payload: dict[str, Any] | None = None

    def to_dict(self) -> dict[str, Any]:
        return {
            "event_type": self.event_type,
            "aggregate_id": str(self.aggregate_id),
            "timestamp": self.timestamp.isoformat(),
            "payload": self.payload or {},
        }


@dataclass
class BarrierCreated(DomainEvent):
    def __init__(self, barrier_id: UUID, payload: dict[str, Any]):
        super().__init__(
            event_type="barrier.created",
            aggregate_id=barrier_id,
            payload=payload,
        )


@dataclass
class BarrierApproved(DomainEvent):
    def __init__(self, barrier_id: UUID, payload: dict[str, Any]):
        super().__init__(
            event_type="barrier.approved",
            aggregate_id=barrier_id,
            payload=payload,
        )


@dataclass
class BarrierRejected(DomainEvent):
    def __init__(self, barrier_id: UUID, payload: dict[str, Any]):
        super().__init__(
            event_type="barrier.rejected",
            aggregate_id=barrier_id,
            payload=payload,
        )