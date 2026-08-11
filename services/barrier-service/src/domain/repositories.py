from abc import ABC, abstractmethod
from uuid import UUID

from .entities.barrier import (
    Barrier,
    BarrierComplaint,
    BarrierConfirmation,
    BarrierPhoto,
    BarrierStatus,
    BarrierType,
    Severity,
)
from .value_objects.coordinates import Coordinates


class BarrierRepository(ABC):
    @abstractmethod
    async def create(self, barrier: Barrier) -> Barrier:
        pass

    @abstractmethod
    async def get_by_id(self, barrier_id: UUID) -> Barrier | None:
        pass

    @abstractmethod
    async def get_by_ids(self, barrier_ids: list[UUID]) -> list[Barrier]:
        pass

    @abstractmethod
    async def update(self, barrier: Barrier) -> Barrier:
        pass

    @abstractmethod
    async def list(
        self,
        status: BarrierStatus | None = None,
        type: BarrierType | None = None,
        severity_min: Severity | None = None,
        severity_max: Severity | None = None,
        bounds: tuple[Coordinates, Coordinates] | None = None,
        limit: int = 100,
        offset: int = 0,
    ) -> list[Barrier]:
        pass

    @abstractmethod
    async def count(
        self,
        status: BarrierStatus | None = None,
        type: BarrierType | None = None,
    ) -> int:
        pass


class BarrierPhotoRepository(ABC):
    @abstractmethod
    async def create(self, photo: BarrierPhoto) -> BarrierPhoto:
        pass

    @abstractmethod
    async def get_by_barrier_id(self, barrier_id: UUID) -> list[BarrierPhoto]:
        pass

    @abstractmethod
    async def get_by_id(self, photo_id: UUID) -> BarrierPhoto | None:
        pass

    @abstractmethod
    async def delete(self, photo_id: UUID) -> bool:
        pass


class BarrierConfirmationRepository(ABC):
    @abstractmethod
    async def create(self, confirmation: BarrierConfirmation) -> BarrierConfirmation:
        pass

    @abstractmethod
    async def exists(self, barrier_id: UUID, user_id: UUID) -> bool:
        pass

    @abstractmethod
    async def count_by_barrier(self, barrier_id: UUID) -> int:
        pass


class BarrierComplaintRepository(ABC):
    @abstractmethod
    async def create(self, complaint: BarrierComplaint) -> BarrierComplaint:
        pass

    @abstractmethod
    async def exists(self, barrier_id: UUID, user_id: UUID) -> bool:
        pass

    @abstractmethod
    async def count_by_barrier(self, barrier_id: UUID) -> int:
        pass


class PhotoStorage(ABC):
    @abstractmethod
    async def upload(self, key: str, data: bytes, content_type: str) -> str:
        pass

    @abstractmethod
    async def download(self, key: str) -> bytes:
        pass

    @abstractmethod
    async def delete(self, key: str) -> bool:
        pass

    @abstractmethod
    async def generate_presigned_url(self, key: str, expires_in: int = 3600) -> str:
        pass


class EventBus(ABC):
    @abstractmethod
    async def publish(self, stream: str, event: dict) -> None:
        pass

    @abstractmethod
    async def subscribe(self, stream: str, group: str, consumer: str, handler) -> None:
        pass