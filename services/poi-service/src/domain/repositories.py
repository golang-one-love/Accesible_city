import builtins
from abc import ABC, abstractmethod
from uuid import UUID

from .entities.poi import AccessibilityFeature, POICategory, PointOfInterest


class POIRepository(ABC):
    @abstractmethod
    async def create(self, poi: PointOfInterest) -> PointOfInterest:
        pass

    @abstractmethod
    async def get_by_id(self, poi_id: UUID) -> PointOfInterest | None:
        pass

    @abstractmethod
    async def update(self, poi: PointOfInterest) -> PointOfInterest:
        pass

    @abstractmethod
    async def delete(self, poi_id: UUID) -> bool:
        pass

    @abstractmethod
    async def list(
        self,
        category: POICategory | None = None,
        bounds: tuple[tuple[float, float], tuple[float, float]] | None = None,
        has_features: list[AccessibilityFeature] | None = None,
        owner_id: UUID | None = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list[PointOfInterest]:
        pass

    @abstractmethod
    async def count(
        self,
        category: POICategory | None = None,
        bounds: tuple[tuple[float, float], tuple[float, float]] | None = None,
    ) -> int:
        pass

    @abstractmethod
    async def search_nearby(
        self,
        latitude: float,
        longitude: float,
        radius_meters: float,
        category: POICategory | None = None,
        limit: int = 20,
    ) -> builtins.list[PointOfInterest]:
        pass