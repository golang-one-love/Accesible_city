from uuid import uuid4

import pytest
from src.domain.entities.moderation import ModerationRequest, ModerationStatus


@pytest.fixture
def request_fixture() -> ModerationRequest:
    return ModerationRequest(barrier_id=uuid4(), reporter_id=uuid4())


def test_request_defaults_to_pending(request_fixture: ModerationRequest) -> None:
    assert request_fixture.status == ModerationStatus.PENDING
    assert request_fixture.moderator_id is None
    assert request_fixture.reviewed_at is None


def test_approve_sets_moderator_and_review_time(request_fixture: ModerationRequest) -> None:
    moderator_id = uuid4()
    request_fixture.approve(moderator_id, comment="Looks ok")
    assert request_fixture.status == ModerationStatus.APPROVED
    assert request_fixture.moderator_id == moderator_id
    assert request_fixture.moderator_comment == "Looks ok"
    assert request_fixture.reviewed_at is not None


def test_cannot_approve_own_request(request_fixture: ModerationRequest) -> None:
    with pytest.raises(ValueError):
        request_fixture.approve(request_fixture.reporter_id)


def test_cannot_reject_own_request(request_fixture: ModerationRequest) -> None:
    with pytest.raises(ValueError):
        request_fixture.reject(request_fixture.reporter_id)


def test_cannot_approve_twice(request_fixture: ModerationRequest) -> None:
    request_fixture.approve(uuid4())
    with pytest.raises(ValueError):
        request_fixture.approve(uuid4())


def test_reject_transitions_to_rejected(request_fixture: ModerationRequest) -> None:
    request_fixture.reject(uuid4(), comment="Duplicate")
    assert request_fixture.status == ModerationStatus.REJECTED
    assert request_fixture.reviewed_at is not None
