from dataclasses import dataclass
from uuid import UUID

from src.application.dto.schemas import (
    BarrierListResponse,
    BarrierPhotoResponse,
    BarrierResponse,
    BarrierStatusStr,
    BarrierTypeStr,
    ComplainBarrierRequest,
    ComplainBarrierResponse,
    ConfirmBarrierResponse,
    CoordinatesDTO,
    CreateBarrierRequest,
    SeverityInt,
    UploadPhotoResponse,
)
from src.domain.entities.barrier import (
    Barrier,
    BarrierStatus,
    BarrierType,
    Severity,
)
from src.domain.repositories import (
    BarrierComplaintRepository,
    BarrierConfirmationRepository,
    BarrierPhotoRepository,
    PhotoStorage,
)
from src.domain.services.barrier_service import BarrierService
from src.domain.value_objects.coordinates import Coordinates


class BarrierNotFoundError(Exception):
    pass


class BarrierValidationError(Exception):
    pass


class BarrierPermissionError(Exception):
    pass


@dataclass
class CreateBarrierUseCase:
    barrier_service: BarrierService

    async def execute(self, request: CreateBarrierRequest, reporter_id: UUID) -> BarrierResponse:
        coordinates = Coordinates(latitude=request.coordinates.latitude, longitude=request.coordinates.longitude)

        barrier_type = BarrierType(request.type.value)
        severity = Severity(request.severity.value)

        barrier = await self.barrier_service.create_barrier(
            type=barrier_type,
            coordinates=coordinates,
            description=request.description,
            severity=severity,
            reporter_id=reporter_id,
        )

        return self._to_response(barrier)

    def _to_response(self, barrier: Barrier) -> BarrierResponse:
        return BarrierResponse(
            id=barrier.id,
            type=BarrierTypeStr(barrier.type.value),
            coordinates=CoordinatesDTO(latitude=barrier.coordinates.latitude, longitude=barrier.coordinates.longitude),
            description=barrier.description,
            severity=SeverityInt(barrier.severity.value),
            status=BarrierStatusStr(barrier.status.value),
            reporter_id=barrier.reporter_id,
            moderator_id=barrier.moderator_id,
            created_at=barrier.created_at,
            updated_at=barrier.updated_at,
            approved_at=barrier.approved_at,
            resolved_at=barrier.resolved_at,
        )


@dataclass
class GetBarrierUseCase:
    barrier_service: BarrierService
    photo_repo: BarrierPhotoRepository
    confirmation_repo: BarrierConfirmationRepository
    complaint_repo: BarrierComplaintRepository
    photo_storage: PhotoStorage

    async def execute(self, barrier_id: UUID) -> BarrierResponse:
        barrier = await self.barrier_service.get_barrier(barrier_id)
        if not barrier:
            raise BarrierNotFoundError("Barrier not found")

        photos = await self.photo_repo.get_by_barrier_id(barrier_id)
        confirmations_count = await self.confirmation_repo.count_by_barrier(barrier_id)
        complaints_count = await self.complaint_repo.count_by_barrier(barrier_id)

        photo_responses = []
        for photo in photos:
            await self.photo_storage.generate_presigned_url(photo.s3_key)
            photo_responses.append(BarrierPhotoResponse(
                id=photo.id,
                barrier_id=photo.barrier_id,
                s3_key=photo.s3_key,
                original_filename=photo.original_filename,
                content_type=photo.content_type,
                size_bytes=photo.size_bytes,
                uploaded_by=photo.uploaded_by,
                created_at=photo.created_at,
            ))

        return BarrierResponse(
            id=barrier.id,
            type=BarrierTypeStr(barrier.type.value),
            coordinates=CoordinatesDTO(latitude=barrier.coordinates.latitude, longitude=barrier.coordinates.longitude),
            description=barrier.description,
            severity=SeverityInt(barrier.severity.value),
            status=BarrierStatusStr(barrier.status.value),
            reporter_id=barrier.reporter_id,
            moderator_id=barrier.moderator_id,
            created_at=barrier.created_at,
            updated_at=barrier.updated_at,
            approved_at=barrier.approved_at,
            resolved_at=barrier.resolved_at,
            photos=photo_responses,
            confirmations_count=confirmations_count,
            complaints_count=complaints_count,
        )


@dataclass
class ListBarriersUseCase:
    barrier_service: BarrierService

    async def execute(
        self,
        status: BarrierStatusStr | None = None,
        type: BarrierTypeStr | None = None,
        severity_min: SeverityInt | None = None,
        severity_max: SeverityInt | None = None,
        bounds: tuple[CoordinatesDTO, CoordinatesDTO] | None = None,
        limit: int = 100,
        offset: int = 0,
    ) -> BarrierListResponse:
        barrier_status = BarrierStatus(status.value) if status else None
        barrier_type = BarrierType(type.value) if type else None
        sev_min = Severity(severity_min.value) if severity_min else None
        sev_max = Severity(severity_max.value) if severity_max else None

        bounds_tuple = None
        if bounds:
            bounds_tuple = (
                Coordinates(latitude=bounds[0].latitude, longitude=bounds[0].longitude),
                Coordinates(latitude=bounds[1].latitude, longitude=bounds[1].longitude),
            )

        barriers = await self.barrier_service.list_barriers(
            status=barrier_status,
            type=barrier_type,
            severity_min=sev_min,
            severity_max=sev_max,
            bounds=bounds_tuple,
            limit=limit,
            offset=offset,
        )

        total = await self.barrier_service.barrier_repo.count(status=barrier_status, type=barrier_type)

        return BarrierListResponse(
            barriers=[self._to_response(b) for b in barriers],
            total=total,
            limit=limit,
            offset=offset,
        )

    def _to_response(self, barrier: Barrier) -> BarrierResponse:
        return BarrierResponse(
            id=barrier.id,
            type=BarrierTypeStr(barrier.type.value),
            coordinates=CoordinatesDTO(latitude=barrier.coordinates.latitude, longitude=barrier.coordinates.longitude),
            description=barrier.description,
            severity=SeverityInt(barrier.severity.value),
            status=BarrierStatusStr(barrier.status.value),
            reporter_id=barrier.reporter_id,
            moderator_id=barrier.moderator_id,
            created_at=barrier.created_at,
            updated_at=barrier.updated_at,
            approved_at=barrier.approved_at,
            resolved_at=barrier.resolved_at,
        )


@dataclass
class UploadPhotoUseCase:
    barrier_service: BarrierService

    async def execute(
        self,
        barrier_id: UUID,
        user_id: UUID,
        filename: str,
        content_type: str,
        data: bytes,
    ) -> UploadPhotoResponse:
        photo = await self.barrier_service.upload_photo(barrier_id, user_id, filename, content_type, data)
        presigned_url = await self.barrier_service.photo_storage.generate_presigned_url(photo.s3_key)

        return UploadPhotoResponse(
            id=photo.id,
            barrier_id=photo.barrier_id,
            s3_key=photo.s3_key,
            original_filename=photo.original_filename,
            content_type=photo.content_type,
            size_bytes=photo.size_bytes,
            presigned_url=presigned_url,
        )


@dataclass
class ConfirmBarrierUseCase:
    barrier_service: BarrierService

    async def execute(self, barrier_id: UUID, user_id: UUID) -> ConfirmBarrierResponse:
        confirmation = await self.barrier_service.confirm_barrier(barrier_id, user_id)
        return ConfirmBarrierResponse(
            id=confirmation.id,
            barrier_id=confirmation.barrier_id,
            user_id=confirmation.user_id,
            created_at=confirmation.created_at,
        )


@dataclass
class ComplainBarrierUseCase:
    barrier_service: BarrierService

    async def execute(self, barrier_id: UUID, user_id: UUID, request: ComplainBarrierRequest) -> ComplainBarrierResponse:
        complaint = await self.barrier_service.complain_barrier(barrier_id, user_id, request.reason)
        return ComplainBarrierResponse(
            id=complaint.id,
            barrier_id=complaint.barrier_id,
            user_id=complaint.user_id,
            reason=complaint.reason,
            created_at=complaint.created_at,
        )


@dataclass
class ApproveBarrierUseCase:
    barrier_service: BarrierService

    async def execute(self, barrier_id: UUID, moderator_id: UUID) -> BarrierResponse:
        barrier = await self.barrier_service.approve_barrier(barrier_id, moderator_id)
        return BarrierResponse(
            id=barrier.id,
            type=BarrierTypeStr(barrier.type.value),
            coordinates=CoordinatesDTO(latitude=barrier.coordinates.latitude, longitude=barrier.coordinates.longitude),
            description=barrier.description,
            severity=SeverityInt(barrier.severity.value),
            status=BarrierStatusStr(barrier.status.value),
            reporter_id=barrier.reporter_id,
            moderator_id=barrier.moderator_id,
            created_at=barrier.created_at,
            updated_at=barrier.updated_at,
            approved_at=barrier.approved_at,
            resolved_at=barrier.resolved_at,
        )


@dataclass
class RejectBarrierUseCase:
    barrier_service: BarrierService

    async def execute(self, barrier_id: UUID, moderator_id: UUID) -> BarrierResponse:
        barrier = await self.barrier_service.reject_barrier(barrier_id, moderator_id)
        return BarrierResponse(
            id=barrier.id,
            type=BarrierTypeStr(barrier.type.value),
            coordinates=CoordinatesDTO(latitude=barrier.coordinates.latitude, longitude=barrier.coordinates.longitude),
            description=barrier.description,
            severity=SeverityInt(barrier.severity.value),
            status=BarrierStatusStr(barrier.status.value),
            reporter_id=barrier.reporter_id,
            moderator_id=barrier.moderator_id,
            created_at=barrier.created_at,
            updated_at=barrier.updated_at,
            approved_at=barrier.approved_at,
            resolved_at=barrier.resolved_at,
        )


@dataclass
class ResolveBarrierUseCase:
    barrier_service: BarrierService

    async def execute(self, barrier_id: UUID) -> BarrierResponse:
        barrier = await self.barrier_service.resolve_barrier(barrier_id)
        return BarrierResponse(
            id=barrier.id,
            type=BarrierTypeStr(barrier.type.value),
            coordinates=CoordinatesDTO(latitude=barrier.coordinates.latitude, longitude=barrier.coordinates.longitude),
            description=barrier.description,
            severity=SeverityInt(barrier.severity.value),
            status=BarrierStatusStr(barrier.status.value),
            reporter_id=barrier.reporter_id,
            moderator_id=barrier.moderator_id,
            created_at=barrier.created_at,
            updated_at=barrier.updated_at,
            approved_at=barrier.approved_at,
            resolved_at=barrier.resolved_at,
        )