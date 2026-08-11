from typing import Optional, List, Tuple
from uuid import UUID
from sqlalchemy import select, func, and_, or_
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from src.domain.entities.barrier import (
    Barrier, BarrierType, BarrierStatus, Severity,
    BarrierPhoto, BarrierConfirmation, BarrierComplaint
)
from src.domain.value_objects.coordinates import Coordinates
from src.domain.repositories import (
    BarrierRepository, BarrierPhotoRepository,
    BarrierConfirmationRepository, BarrierComplaintRepository
)
from src.adapters.out.persistence.models import (
    BarrierModel, BarrierPhotoModel,
    BarrierConfirmationModel, BarrierComplaintModel
)
from src.adapters.out.persistence.database import async_session_factory


class PostgresBarrierRepository(BarrierRepository):
    async def create(self, barrier: Barrier) -> Barrier:
        async with async_session_factory() as session:
            model = self._to_model(barrier)
            session.add(model)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def get_by_id(self, barrier_id: UUID) -> Optional[Barrier]:
        async with async_session_factory() as session:
            stmt = select(BarrierModel).where(BarrierModel.id == barrier_id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if model:
                return self._to_entity(model)
            return None

    async def get_by_ids(self, barrier_ids: List[UUID]) -> List[Barrier]:
        async with async_session_factory() as session:
            stmt = select(BarrierModel).where(BarrierModel.id.in_(barrier_ids))
            result = await session.execute(stmt)
            return [self._to_entity(m) for m in result.scalars().all()]

    async def update(self, barrier: Barrier) -> Barrier:
        async with async_session_factory() as session:
            stmt = select(BarrierModel).where(BarrierModel.id == barrier.id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if not model:
                return None
            self._update_model(model, barrier)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def list(
        self,
        status: Optional[BarrierStatus] = None,
        type: Optional[BarrierType] = None,
        severity_min: Optional[Severity] = None,
        severity_max: Optional[Severity] = None,
        bounds: Optional[Tuple[Coordinates, Coordinates]] = None,
        limit: int = 100,
        offset: int = 0,
    ) -> List[Barrier]:
        async with async_session_factory() as session:
            stmt = select(BarrierModel)

            conditions = []
            if status:
                conditions.append(BarrierModel.status == status.value)
            if type:
                conditions.append(BarrierModel.type == type.value)
            if severity_min:
                conditions.append(BarrierModel.severity >= severity_min.value)
            if severity_max:
                conditions.append(BarrierModel.severity <= severity_max.value)
            if bounds:
                sw, ne = bounds
                conditions.append(
                    and_(
                        BarrierModel.latitude >= str(sw.latitude),
                        BarrierModel.latitude <= str(ne.latitude),
                        BarrierModel.longitude >= str(sw.longitude),
                        BarrierModel.longitude <= str(ne.longitude),
                    )
                )

            if conditions:
                stmt = stmt.where(and_(*conditions))

            stmt = stmt.order_by(BarrierModel.created_at.desc()).limit(limit).offset(offset)
            result = await session.execute(stmt)
            return [self._to_entity(m) for m in result.scalars().all()]

    async def count(
        self,
        status: Optional[BarrierStatus] = None,
        type: Optional[BarrierType] = None,
    ) -> int:
        async with async_session_factory() as session:
            stmt = select(func.count(BarrierModel.id))
            conditions = []
            if status:
                conditions.append(BarrierModel.status == status.value)
            if type:
                conditions.append(BarrierModel.type == type.value)
            if conditions:
                stmt = stmt.where(and_(*conditions))
            result = await session.execute(stmt)
            return result.scalar() or 0

    def _to_model(self, barrier: Barrier) -> BarrierModel:
        return BarrierModel(
            id=barrier.id,
            type=barrier.type.value,
            latitude=str(barrier.coordinates.latitude),
            longitude=str(barrier.coordinates.longitude),
            description=barrier.description,
            severity=barrier.severity.value,
            status=barrier.status.value,
            reporter_id=barrier.reporter_id,
            moderator_id=barrier.moderator_id,
            created_at=barrier.created_at,
            updated_at=barrier.updated_at,
            approved_at=barrier.approved_at,
            resolved_at=barrier.resolved_at,
        )

    def _update_model(self, model: BarrierModel, barrier: Barrier) -> None:
        model.type = barrier.type.value
        model.latitude = str(barrier.coordinates.latitude)
        model.longitude = str(barrier.coordinates.longitude)
        model.description = barrier.description
        model.severity = barrier.severity.value
        model.status = barrier.status.value
        model.moderator_id = barrier.moderator_id
        model.updated_at = barrier.updated_at
        model.approved_at = barrier.approved_at
        model.resolved_at = barrier.resolved_at

    def _to_entity(self, model: BarrierModel) -> Barrier:
        return Barrier(
            id=model.id,
            type=BarrierType(model.type),
            coordinates=Coordinates(latitude=float(model.latitude), longitude=float(model.longitude)),
            description=model.description,
            severity=Severity(model.severity),
            status=BarrierStatus(model.status),
            reporter_id=model.reporter_id,
            moderator_id=model.moderator_id,
            created_at=model.created_at,
            updated_at=model.updated_at,
            approved_at=model.approved_at,
            resolved_at=model.resolved_at,
        )


class PostgresBarrierPhotoRepository(BarrierPhotoRepository):
    async def create(self, photo: BarrierPhoto) -> BarrierPhoto:
        async with async_session_factory() as session:
            model = BarrierPhotoModel(
                id=photo.id,
                barrier_id=photo.barrier_id,
                s3_key=photo.s3_key,
                original_filename=photo.original_filename,
                content_type=photo.content_type,
                size_bytes=photo.size_bytes,
                uploaded_by=photo.uploaded_by,
                created_at=photo.created_at,
            )
            session.add(model)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def get_by_barrier_id(self, barrier_id: UUID) -> List[BarrierPhoto]:
        async with async_session_factory() as session:
            stmt = select(BarrierPhotoModel).where(BarrierPhotoModel.barrier_id == barrier_id)
            result = await session.execute(stmt)
            return [self._to_entity(m) for m in result.scalars().all()]

    async def get_by_id(self, photo_id: UUID) -> Optional[BarrierPhoto]:
        async with async_session_factory() as session:
            stmt = select(BarrierPhotoModel).where(BarrierPhotoModel.id == photo_id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if model:
                return self._to_entity(model)
            return None

    async def delete(self, photo_id: UUID) -> bool:
        async with async_session_factory() as session:
            stmt = select(BarrierPhotoModel).where(BarrierPhotoModel.id == photo_id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if model:
                await session.delete(model)
                await session.commit()
                return True
            return False

    def _to_entity(self, model: BarrierPhotoModel) -> BarrierPhoto:
        return BarrierPhoto(
            id=model.id,
            barrier_id=model.barrier_id,
            s3_key=model.s3_key,
            original_filename=model.original_filename,
            content_type=model.content_type,
            size_bytes=model.size_bytes,
            uploaded_by=model.uploaded_by,
            created_at=model.created_at,
        )


class PostgresBarrierConfirmationRepository(BarrierConfirmationRepository):
    async def create(self, confirmation: BarrierConfirmation) -> BarrierConfirmation:
        async with async_session_factory() as session:
            model = BarrierConfirmationModel(
                id=confirmation.id,
                barrier_id=confirmation.barrier_id,
                user_id=confirmation.user_id,
                created_at=confirmation.created_at,
            )
            session.add(model)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def exists(self, barrier_id: UUID, user_id: UUID) -> bool:
        async with async_session_factory() as session:
            stmt = select(BarrierConfirmationModel).where(
                and_(
                    BarrierConfirmationModel.barrier_id == barrier_id,
                    BarrierConfirmationModel.user_id == user_id,
                )
            )
            result = await session.execute(stmt)
            return result.scalar_one_or_none() is not None

    async def count_by_barrier(self, barrier_id: UUID) -> int:
        async with async_session_factory() as session:
            stmt = select(func.count(BarrierConfirmationModel.id)).where(
                BarrierConfirmationModel.barrier_id == barrier_id
            )
            result = await session.execute(stmt)
            return result.scalar() or 0

    def _to_entity(self, model: BarrierConfirmationModel) -> BarrierConfirmation:
        return BarrierConfirmation(
            id=model.id,
            barrier_id=model.barrier_id,
            user_id=model.user_id,
            created_at=model.created_at,
        )


class PostgresBarrierComplaintRepository(BarrierComplaintRepository):
    async def create(self, complaint: BarrierComplaint) -> BarrierComplaint:
        async with async_session_factory() as session:
            model = BarrierComplaintModel(
                id=complaint.id,
                barrier_id=complaint.barrier_id,
                user_id=complaint.user_id,
                reason=complaint.reason,
                created_at=complaint.created_at,
            )
            session.add(model)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def exists(self, barrier_id: UUID, user_id: UUID) -> bool:
        async with async_session_factory() as session:
            stmt = select(BarrierComplaintModel).where(
                and_(
                    BarrierComplaintModel.barrier_id == barrier_id,
                    BarrierComplaintModel.user_id == user_id,
                )
            )
            result = await session.execute(stmt)
            return result.scalar_one_or_none() is not None

    async def count_by_barrier(self, barrier_id: UUID) -> int:
        async with async_session_factory() as session:
            stmt = select(func.count(BarrierComplaintModel.id)).where(
                BarrierComplaintModel.barrier_id == barrier_id
            )
            result = await session.execute(stmt)
            return result.scalar() or 0

    def _to_entity(self, model: BarrierComplaintModel) -> BarrierComplaint:
        return BarrierComplaint(
            id=model.id,
            barrier_id=model.barrier_id,
            user_id=model.user_id,
            reason=model.reason,
            created_at=model.created_at,
        )