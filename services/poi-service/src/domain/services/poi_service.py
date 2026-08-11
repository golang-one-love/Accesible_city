from uuid import UUID

from ..entities.poi import (
    AccessibilityFeature,
    AccessibilityProfile,
    POICategory,
    PointOfInterest,
)
from ..repositories import POIRepository


class POIService:
    def __init__(self, poi_repo: POIRepository):
        self.poi_repo = poi_repo

    async def create_poi(
        self,
        name: str,
        category: POICategory,
        latitude: float,
        longitude: float,
        address: str,
        owner_id: UUID,
        phone: str = "",
        website: str = "",
        opening_hours: str = "",
        accessibility: AccessibilityProfile | None = None,
    ) -> PointOfInterest:
        poi = PointOfInterest(
            name=name,
            category=category,
            latitude=latitude,
            longitude=longitude,
            address=address,
            phone=phone,
            website=website,
            opening_hours=opening_hours,
            accessibility=accessibility or AccessibilityProfile(),
            owner_id=owner_id,
        )
        return await self.poi_repo.create(poi)

    async def get_poi(self, poi_id: UUID) -> PointOfInterest | None:
        return await self.poi_repo.get_by_id(poi_id)

    async def update_poi(
        self,
        poi_id: UUID,
        owner_id: UUID,
        name: str | None = None,
        category: POICategory | None = None,
        address: str | None = None,
        phone: str | None = None,
        website: str | None = None,
        opening_hours: str | None = None,
        accessibility: AccessibilityProfile | None = None,
    ) -> PointOfInterest:
        poi = await self.poi_repo.get_by_id(poi_id)
        if not poi:
            raise ValueError("POI not found")
        if poi.owner_id != owner_id:
            raise ValueError("Not authorized to update this POI")

        if name is not None:
            poi.name = name
        if category is not None:
            poi.category = category
        if address is not None:
            poi.address = address
        if phone is not None:
            poi.phone = phone
        if website is not None:
            poi.website = website
        if opening_hours is not None:
            poi.opening_hours = opening_hours
        if accessibility is not None:
            poi.update_accessibility(accessibility)

        poi.updated_at = __import__('datetime').datetime.utcnow()
        return await self.poi_repo.update(poi)

    async def delete_poi(self, poi_id: UUID, owner_id: UUID) -> bool:
        poi = await self.poi_repo.get_by_id(poi_id)
        if not poi:
            raise ValueError("POI not found")
        if poi.owner_id != owner_id:
            raise ValueError("Not authorized to delete this POI")
        return await self.poi_repo.delete(poi_id)

    async def list_pois(
        self,
        category: POICategory | None = None,
        bounds: tuple[tuple[float, float], tuple[float, float]] | None = None,
        has_features: list[AccessibilityFeature] | None = None,
        owner_id: UUID | None = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list[PointOfInterest]:
        return await self.poi_repo.list(
            category=category,
            bounds=bounds,
            has_features=has_features,
            owner_id=owner_id,
            limit=limit,
            offset=offset,
        )

    async def search_nearby(
        self,
        latitude: float,
        longitude: float,
        radius_meters: float,
        category: POICategory | None = None,
        limit: int = 20,
    ) -> list[PointOfInterest]:
        return await self.poi_repo.search_nearby(
            latitude=latitude,
            longitude=longitude,
            radius_meters=radius_meters,
            category=category,
            limit=limit,
        )