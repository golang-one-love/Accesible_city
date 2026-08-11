from uuid import uuid4

import pytest
from src.domain.entities.barrier import Barrier, BarrierStatus, BarrierType, Severity
from src.domain.value_objects.coordinates import Coordinates


@pytest.fixture
def barrier() -> Barrier:
    return Barrier(
        type=BarrierType.HIGH_CURB,
        coordinates=Coordinates(latitude=55.75, longitude=37.61),
        description="High curb at the entrance",
        severity=Severity.MEDIUM,
        reporter_id=uuid4(),
    )


def test_create_barrier_defaults_to_pending(barrier: Barrier) -> None:
    assert barrier.status == BarrierStatus.PENDING
    assert barrier.moderator_id is None
    assert barrier.approved_at is None


def test_approve_transitions_to_approved(barrier: Barrier) -> None:
    moderator_id = uuid4()
    barrier.approve(moderator_id)
    assert barrier.status == BarrierStatus.APPROVED
    assert barrier.moderator_id == moderator_id
    assert barrier.approved_at is not None
    assert barrier.updated_at is not None


def test_cannot_approve_non_pending_barrier(barrier: Barrier) -> None:
    barrier.approve(uuid4())
    with pytest.raises(ValueError):
        barrier.approve(uuid4())


def test_reject_transitions_to_rejected(barrier: Barrier) -> None:
    barrier.reject(uuid4())
    assert barrier.status == BarrierStatus.REJECTED


def test_cannot_reject_after_approval(barrier: Barrier) -> None:
    barrier.approve(uuid4())
    with pytest.raises(ValueError):
        barrier.reject(uuid4())


def test_resolve_requires_approved(barrier: Barrier) -> None:
    with pytest.raises(ValueError):
        barrier.resolve()


def test_full_lifecycle(barrier: Barrier) -> None:
    barrier.approve(uuid4())
    barrier.resolve()
    assert barrier.status == BarrierStatus.RESOLVED
    assert barrier.resolved_at is not None
