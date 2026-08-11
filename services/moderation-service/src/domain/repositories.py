from abc import ABC, abstractmethod
from uuid import UUID

from .entities.moderation import ModerationRequest, ModerationStatus


class ModerationRepository(ABC):
    @abstractmethod
    async def create(self, request: ModerationRequest) -> ModerationRequest:
        pass

    @abstractmethod
    async def get_by_id(self, request_id: UUID) -> ModerationRequest | None:
        pass

    @abstractmethod
    async def get_by_barrier_id(self, barrier_id: UUID) -> ModerationRequest | None:
        pass

    @abstractmethod
    async def update(self, request: ModerationRequest) -> ModerationRequest:
        pass

    @abstractmethod
    async def list(
        self,
        status: ModerationStatus | None = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list[ModerationRequest]:
        pass

    @abstractmethod
    async def count(self, status: ModerationStatus | None = None) -> int:
        pass


class EventBus(ABC):
    @abstractmethod
    async def publish(self, stream: str, event: dict) -> None:
        pass

    @abstractmethod
    async def subscribe(self, stream: str, group: str, consumer: str, handler) -> None:
        pass