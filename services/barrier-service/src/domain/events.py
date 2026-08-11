from dataclasses import dataclass, asdict
from datetime import datetime
from uuid import UUID
from typing import Dict, Any


@dataclass
class DomainEvent:
    event_type: str
    aggregate_id: UUID
    timestamp: datetime = datetime.utcnow()
    payload: Dict[str, Any] = None

    def to_dict(self) -> Dict[str, Any]:
        return {
            "event_type": self.event_type,
            "aggregate_id": str(self.aggregate_id),
            "timestamp": self.timestamp.isoformat(),
            "payload": self.payload or {},
        }


@dataclass
class BarrierCreated(DomainEvent):
    def __init__(self, barrier_id: UUID, payload: Dict[str, Any]):
        super().__init__(
            event_type="barrier.created",
            aggregate_id=barrier_id,
            payload=payload,
        )


@dataclass
class BarrierApproved(DomainEvent):
    def __init__(self, barrier_id: UUID, payload: Dict[str, Any]):
        super().__init__(
            event_type="barrier.approved",
            aggregate_id=barrier_id,
            payload=payload,
        )


@dataclass
class BarrierRejected(DomainEvent):
    def __init__(self, barrier_id: UUID, payload: Dict[str, Any]):
        super().__init__(
            event_type="barrier.rejected",
            aggregate_id=barrier_id,
            payload=payload,
        )


@dataclass
class BarrierResolved(DomainEvent):
    def __init__(self, barrier_id: UUID, payload: Dict[str, Any]):
        super().__init__(
            event_type="barrier.resolved",
            aggregate_id=barrier_id,
            payload=payload,
        )