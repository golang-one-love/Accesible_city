from uuid import UUID

from ..entities.barrier import (
    Barrier,
    BarrierComplaint,
    BarrierConfirmation,
    BarrierPhoto,
    BarrierStatus,
    BarrierType,
    Severity,
)
from ..events import BarrierApproved, BarrierCreated, BarrierRejected, BarrierResolved
from ..repositories import (
    BarrierComplaintRepository,
    BarrierConfirmationRepository,
    BarrierPhotoRepository,
    BarrierRepository,
    EventBus,
    PhotoStorage,
)
from ..value_objects.coordinates import Coordinates


class BarrierService:
    def __init__(
        self,
        barrier_repo: BarrierRepository,
        photo_repo: BarrierPhotoRepository,
        confirmation_repo: BarrierConfirmationRepository,
        complaint_repo: BarrierComplaintRepository,
        photo_storage: PhotoStorage,
        event_bus: EventBus,
    ):
        self.barrier_repo = barrier_repo
        self.photo_repo = photo_repo
        self.confirmation_repo = confirmation_repo
        self.complaint_repo = complaint_repo
        self.photo_storage = photo_storage
        self.event_bus = event_bus

    async def create_barrier(
        self,
        type: BarrierType,
        coordinates: Coordinates,
        description: str,
        severity: Severity,
        reporter_id: UUID,
    ) -> Barrier:
        barrier = Barrier(
            type=type,
            coordinates=coordinates,
            description=description,
            severity=severity,
            reporter_id=reporter_id,
        )
        created = await self.barrier_repo.create(barrier)

        await self.event_bus.publish(
            "barrier.created",
            BarrierCreated(
                barrier_id=created.id,
                payload={
                    "type": created.type.value,
                    "coordinates": {"lat": created.coordinates.latitude, "lon": created.coordinates.longitude},
                    "description": created.description,
                    "severity": created.severity.value,
                    "reporter_id": str(created.reporter_id),
                },
            ).to_dict(),
        )

        return created

    async def get_barrier(self, barrier_id: UUID) -> Barrier | None:
        return await self.barrier_repo.get_by_id(barrier_id)

    async def list_barriers(
        self,
        status: BarrierStatus | None = None,
        type: BarrierType | None = None,
        severity_min: Severity | None = None,
        severity_max: Severity | None = None,
        bounds: tuple | None = None,
        limit: int = 100,
        offset: int = 0,
    ) -> list:
        return await self.barrier_repo.list(
            status=status,
            type=type,
            severity_min=severity_min,
            severity_max=severity_max,
            bounds=bounds,
            limit=limit,
            offset=offset,
        )

    async def upload_photo(
        self,
        barrier_id: UUID,
        user_id: UUID,
        filename: str,
        content_type: str,
        data: bytes,
    ) -> BarrierPhoto:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")

        s3_key = f"barriers/{barrier_id}/{filename}"
        await self.photo_storage.upload(s3_key, data, content_type)

        photo = BarrierPhoto(
            barrier_id=barrier_id,
            s3_key=s3_key,
            original_filename=filename,
            content_type=content_type,
            size_bytes=len(data),
            uploaded_by=user_id,
        )
        return await self.photo_repo.create(photo)

    async def confirm_barrier(self, barrier_id: UUID, user_id: UUID) -> BarrierConfirmation:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")

        if barrier.status not in (BarrierStatus.PENDING, BarrierStatus.APPROVED):
            raise ValueError("Can only confirm pending or approved barriers")

        exists = await self.confirmation_repo.exists(barrier_id, user_id)
        if exists:
            raise ValueError("Already confirmed by this user")

        confirmation = BarrierConfirmation(barrier_id=barrier_id, user_id=user_id)
        created = await self.confirmation_repo.create(confirmation)

        barrier.add_confirmation()
        await self.barrier_repo.update(barrier)

        return created

    async def complain_barrier(self, barrier_id: UUID, user_id: UUID, reason: str) -> BarrierComplaint:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")

        if barrier.status not in (BarrierStatus.PENDING, BarrierStatus.APPROVED):
            raise ValueError("Can only complain about pending or approved barriers")

        exists = await self.complaint_repo.exists(barrier_id, user_id)
        if exists:
            raise ValueError("Already complained by this user")

        complaint = BarrierComplaint(barrier_id=barrier_id, user_id=user_id, reason=reason)
        return await self.complaint_repo.create(complaint)

    async def approve_barrier(self, barrier_id: UUID, moderator_id: UUID) -> Barrier:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")

        if barrier.reporter_id == moderator_id:
            raise ValueError("Cannot moderate own barrier")

        barrier.approve(moderator_id)
        updated = await self.barrier_repo.update(barrier)

        await self.event_bus.publish(
            "barrier.approved",
            BarrierApproved(
                barrier_id=updated.id,
                payload={
                    "type": updated.type.value,
                    "coordinates": {"lat": updated.coordinates.latitude, "lon": updated.coordinates.longitude},
                    "severity": updated.severity.value,
                    "barrier_id": str(updated.id),
                    "reporter_id": str(updated.reporter_id),
                    "moderator_id": str(moderator_id),
                },
            ).to_dict(),
        )

        return updated

    async def reject_barrier(self, barrier_id: UUID, moderator_id: UUID) -> Barrier:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")

        if barrier.reporter_id == moderator_id:
            raise ValueError("Cannot moderate own barrier")

        barrier.reject(moderator_id)
        updated = await self.barrier_repo.update(barrier)

        await self.event_bus.publish(
            "barrier.rejected",
            BarrierRejected(
                barrier_id=updated.id,
                payload={
                    "type": updated.type.value,
                    "barrier_id": str(updated.id),
                    "reporter_id": str(updated.reporter_id),
                    "moderator_id": str(moderator_id),
                },
            ).to_dict(),
        )

        return updated

    async def resolve_barrier(self, barrier_id: UUID) -> Barrier:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")

        barrier.resolve()
        updated = await self.barrier_repo.update(barrier)

        await self.event_bus.publish(
            "barrier.resolved",
            BarrierResolved(
                barrier_id=updated.id,
                payload={
                    "type": updated.type.value,
                    "coordinates": {"lat": updated.coordinates.latitude, "lon": updated.coordinates.longitude},
                    "barrier_id": str(updated.id),
                    "reporter_id": str(updated.reporter_id),
                },
            ).to_dict(),
        )

        return updated

    async def handle_approved_event(self, barrier_id: UUID, moderator_id: UUID) -> Barrier:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")
        if barrier.status != BarrierStatus.PENDING:
            return barrier
        barrier.approve(moderator_id)
        return await self.barrier_repo.update(barrier)

    async def handle_rejected_event(self, barrier_id: UUID, moderator_id: UUID) -> Barrier:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")
        if barrier.status != BarrierStatus.PENDING:
            return barrier
        barrier.reject(moderator_id)
        return await self.barrier_repo.update(barrier)

    async def handle_resolved_event(self, barrier_id: UUID) -> Barrier:
        barrier = await self.barrier_repo.get_by_id(barrier_id)
        if not barrier:
            raise ValueError("Barrier not found")
        if barrier.status != BarrierStatus.APPROVED:
            return barrier
        barrier.resolve()
        return await self.barrier_repo.update(barrier)