from abc import ABC, abstractmethod
from typing import Optional, List, Tuple
from uuid import UUID

from .entities.poi import PointOfInterest, POICategory, AccessibilityFeature


class POIRepository(ABC):
    @abstractmethod
    async def create(self, poi: PointOfInterest) -> PointOfInterest:
        pass

    @abstractmethod
    async def get_by_id(self, poi_id: UUID) -> Optional[PointOfInterest]:
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
        category: Optional[POICategory] = None,
        bounds: Optional[Tuple[Tuple[float, float], Tuple[float, float]]] = None,
        has_features: Optional[List[AccessibilityFeature]] = None,
        owner_id: Optional[UUID] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> List[PointOfInterest]:
        pass

    @abstractmethod
    async def count(
        self,
        category: Optional[POICategory] = None,
        bounds: Optional[Tuple[Tuple[float, float], Tuple[float, float]]] = None,
    ) -> int:
        pass

    @abstractmethod
    async def search_nearby(
        self,
        latitude: float,
        longitude: float,
        radius_meters: float,
        category: Optional[POICategory] = None,
        limit: int = 20,
    ) -> List[PointOfInterest]:
        pass