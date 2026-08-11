from dataclasses import dataclass
from uuid import UUID

from src.application.dto.schemas import (
    ModerationActionRequest,
    ModerationActionResponse,
    ModerationQueueResponse,
    ModerationRequestResponse,
)
from src.domain.entities.moderation import ModerationRequest, ModerationStatus
from src.domain.services.moderation_service import ModerationService


class ModerationNotFoundError(Exception):
    pass


class ModerationPermissionError(Exception):
    pass


@dataclass
class GetQueueUseCase:
    moderation_service: ModerationService

    async def execute(
        self,
        status: ModerationStatus | None = None,
        limit: int = 50,
        offset: int = 0,
    ) -> ModerationQueueResponse:
        requests = await self.moderation_service.get_queue(status=status, limit=limit, offset=offset)
        total = await self.moderation_service.moderation_repo.count(status=status)

        return ModerationQueueResponse(
            requests=[self._to_response(r) for r in requests],
            total=total,
            limit=limit,
            offset=offset,
        )

    def _to_response(self, request: ModerationRequest) -> ModerationRequestResponse:
        return ModerationRequestResponse(
            id=request.id,
            barrier_id=request.barrier_id,
            reporter_id=request.reporter_id,
            status=request.status,
            moderator_id=request.moderator_id,
            moderator_comment=request.moderator_comment,
            created_at=request.created_at,
            updated_at=request.updated_at,
            reviewed_at=request.reviewed_at,
        )


@dataclass
class GetRequestUseCase:
    moderation_service: ModerationService

    async def execute(self, request_id: UUID) -> ModerationRequestResponse:
        request = await self.moderation_service.get_request(request_id)
        if not request:
            raise ModerationNotFoundError("Moderation request not found")
        return ModerationRequestResponse(
            id=request.id,
            barrier_id=request.barrier_id,
            reporter_id=request.reporter_id,
            status=request.status,
            moderator_id=request.moderator_id,
            moderator_comment=request.moderator_comment,
            created_at=request.created_at,
            updated_at=request.updated_at,
            reviewed_at=request.reviewed_at,
        )


@dataclass
class ApproveRequestUseCase:
    moderation_service: ModerationService

    async def execute(self, request_id: UUID, moderator_id: UUID, request: ModerationActionRequest) -> ModerationActionResponse:
        try:
            updated = await self.moderation_service.approve_request(request_id, moderator_id, request.comment)
        except ValueError as e:
            raise ModerationPermissionError(str(e))
        return ModerationActionResponse(
            id=updated.id,
            barrier_id=updated.barrier_id,
            status=updated.status,
            moderator_id=updated.moderator_id,
            reviewed_at=updated.reviewed_at,
        )


@dataclass
class RejectRequestUseCase:
    moderation_service: ModerationService

    async def execute(self, request_id: UUID, moderator_id: UUID, request: ModerationActionRequest) -> ModerationActionResponse:
        try:
            updated = await self.moderation_service.reject_request(request_id, moderator_id, request.comment)
        except ValueError as e:
            raise ModerationPermissionError(str(e))
        return ModerationActionResponse(
            id=updated.id,
            barrier_id=updated.barrier_id,
            status=updated.status,
            moderator_id=updated.moderator_id,
            reviewed_at=updated.reviewed_at,
        )