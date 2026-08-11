from uuid import UUID

from sqlalchemy import func, select

from src.adapters.out.persistence.database import async_session_factory
from src.adapters.out.persistence.models import ModerationRequestModel
from src.domain.entities.moderation import ModerationRequest, ModerationStatus
from src.domain.repositories import ModerationRepository


class PostgresModerationRepository(ModerationRepository):
    async def create(self, request: ModerationRequest) -> ModerationRequest:
        async with async_session_factory() as session:
            model = self._to_model(request)
            session.add(model)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def get_by_id(self, request_id: UUID) -> ModerationRequest | None:
        async with async_session_factory() as session:
            stmt = select(ModerationRequestModel).where(ModerationRequestModel.id == request_id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if model:
                return self._to_entity(model)
            return None

    async def get_by_barrier_id(self, barrier_id: UUID) -> ModerationRequest | None:
        async with async_session_factory() as session:
            stmt = select(ModerationRequestModel).where(ModerationRequestModel.barrier_id == barrier_id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if model:
                return self._to_entity(model)
            return None

    async def update(self, request: ModerationRequest) -> ModerationRequest:
        async with async_session_factory() as session:
            stmt = select(ModerationRequestModel).where(ModerationRequestModel.id == request.id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if not model:
                raise ValueError("Moderation request not found")
            self._update_model(model, request)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def list(
        self,
        status: ModerationStatus | None = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list[ModerationRequest]:
        async with async_session_factory() as session:
            stmt = select(ModerationRequestModel)
            if status:
                stmt = stmt.where(ModerationRequestModel.status == status.value)
            stmt = stmt.order_by(ModerationRequestModel.created_at.desc()).limit(limit).offset(offset)
            result = await session.execute(stmt)
            return [self._to_entity(m) for m in result.scalars().all()]

    async def count(self, status: ModerationStatus | None = None) -> int:
        async with async_session_factory() as session:
            stmt = select(func.count(ModerationRequestModel.id))
            if status:
                stmt = stmt.where(ModerationRequestModel.status == status.value)
            result = await session.execute(stmt)
            return result.scalar() or 0

    def _to_model(self, request: ModerationRequest) -> ModerationRequestModel:
        return ModerationRequestModel(
            id=request.id,
            barrier_id=request.barrier_id,
            reporter_id=request.reporter_id,
            status=request.status.value,
            moderator_id=request.moderator_id,
            moderator_comment=request.moderator_comment,
            created_at=request.created_at,
            updated_at=request.updated_at,
            reviewed_at=request.reviewed_at,
        )

    def _update_model(self, model: ModerationRequestModel, request: ModerationRequest) -> None:
        model.status = request.status.value
        model.moderator_id = request.moderator_id
        model.moderator_comment = request.moderator_comment
        model.updated_at = request.updated_at
        model.reviewed_at = request.reviewed_at

    def _to_entity(self, model: ModerationRequestModel) -> ModerationRequest:
        return ModerationRequest(
            id=model.id,
            barrier_id=model.barrier_id,
            reporter_id=model.reporter_id,
            status=ModerationStatus(model.status),
            moderator_id=model.moderator_id,
            moderator_comment=model.moderator_comment,
            created_at=model.created_at,
            updated_at=model.updated_at,
            reviewed_at=model.reviewed_at,
        )