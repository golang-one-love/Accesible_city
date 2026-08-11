from uuid import UUID

from ..entities.moderation import ModerationRequest, ModerationStatus
from ..events import BarrierApproved, BarrierCreated, BarrierRejected
from ..repositories import EventBus, ModerationRepository


class ModerationService:
    def __init__(
        self,
        moderation_repo: ModerationRepository,
        event_bus: EventBus,
        barrier_service_url: str,
    ):
        self.moderation_repo = moderation_repo
        self.event_bus = event_bus
        self.barrier_service_url = barrier_service_url

    async def handle_barrier_created(self, event: BarrierCreated) -> ModerationRequest:
        existing = await self.moderation_repo.get_by_barrier_id(event.aggregate_id)
        if existing:
            return existing

        request = ModerationRequest(
            barrier_id=event.aggregate_id,
            reporter_id=UUID(event.payload.get("reporter_id", "") if event.payload else ""),
        )
        return await self.moderation_repo.create(request)

    async def approve_request(self, request_id: UUID, moderator_id: UUID, comment: str = "") -> ModerationRequest:
        request = await self.moderation_repo.get_by_id(request_id)
        if not request:
            raise ValueError("Moderation request not found")

        if request.reporter_id == moderator_id:
            raise ValueError("Cannot moderate own request")

        request.approve(moderator_id, comment)
        updated = await self.moderation_repo.update(request)

        if request.barrier_id is None:
            raise ValueError("Moderation request has no barrier_id")

        await self.event_bus.publish(
            "barrier.approved",
            BarrierApproved(
                barrier_id=request.barrier_id,
                payload={
                    "type": "approved",
                    "barrier_id": str(request.barrier_id),
                    "reporter_id": str(request.reporter_id),
                    "moderator_id": str(moderator_id),
                    "comment": comment,
                },
            ).to_dict(),
        )

        return updated

    async def reject_request(self, request_id: UUID, moderator_id: UUID, comment: str = "") -> ModerationRequest:
        request = await self.moderation_repo.get_by_id(request_id)
        if not request:
            raise ValueError("Moderation request not found")

        if request.reporter_id == moderator_id:
            raise ValueError("Cannot moderate own request")

        request.reject(moderator_id, comment)
        updated = await self.moderation_repo.update(request)

        if request.barrier_id is None:
            raise ValueError("Moderation request has no barrier_id")

        await self.event_bus.publish(
            "barrier.rejected",
            BarrierRejected(
                barrier_id=request.barrier_id,
                payload={
                    "type": "rejected",
                    "barrier_id": str(request.barrier_id),
                    "reporter_id": str(request.reporter_id),
                    "moderator_id": str(moderator_id),
                    "comment": comment,
                },
            ).to_dict(),
        )

        return updated

    async def get_queue(
        self,
        status: ModerationStatus | None = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list:
        return await self.moderation_repo.list(status=status, limit=limit, offset=offset)

    async def get_request(self, request_id: UUID) -> ModerationRequest | None:
        return await self.moderation_repo.get_by_id(request_id)