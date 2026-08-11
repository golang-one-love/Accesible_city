from dataclasses import dataclass
from typing import Optional, List, Tuple
from uuid import UUID

from src.domain.entities.poi import PointOfInterest, POICategory, AccessibilityProfile, AccessibilityFeature
from src.domain.services.poi_service import POIService
from src.application.dto.schemas import (
    CreatePOIRequest,
    UpdatePOIRequest,
    POIResponse,
    POIListResponse,
    SearchNearbyRequest,
    AccessibilityProfileDTO,
    POICategoryStr,
    AccessibilityFeatureStr,
)


class POINotFoundError(Exception):
    pass


class POIPermissionError(Exception):
    pass


def _to_accessibility_profile(dto: Optional[AccessibilityProfileDTO]) -> AccessibilityProfile:
    if not dto:
        return AccessibilityProfile()
    return AccessibilityProfile(
        features=[AccessibilityFeature(f.value) for f in dto.features],
        entrance_step_height_cm=dto.entrance_step_height_cm,
        door_width_cm=dto.door_width_cm,
        has_accessible_toilet=dto.has_accessible_toilet,
        notes=dto.notes,
    )


def _to_accessibility_dto(profile: AccessibilityProfile) -> AccessibilityProfileDTO:
    return AccessibilityProfileDTO(
        features=[AccessibilityFeatureStr(f.value) for f in profile.features],
        entrance_step_height_cm=profile.entrance_step_height_cm,
        door_width_cm=profile.door_width_cm,
        has_accessible_toilet=profile.has_accessible_toilet,
        notes=profile.notes,
    )


def _to_response(poi: PointOfInterest) -> POIResponse:
    return POIResponse(
        id=poi.id,
        name=poi.name,
        category=POICategoryStr(poi.category.value),
        latitude=poi.latitude,
        longitude=poi.longitude,
        address=poi.address,
        phone=poi.phone,
        website=poi.website,
        opening_hours=poi.opening_hours,
        accessibility=_to_accessibility_dto(poi.accessibility),
        owner_id=poi.owner_id,
        is_verified=poi.is_verified,
        created_at=poi.created_at,
        updated_at=poi.updated_at,
    )


@dataclass
class CreatePOIUseCase:
    poi_service: POIService

    async def execute(self, request: CreatePOIRequest, owner_id: UUID) -> POIResponse:
        accessibility = _to_accessibility_profile(request.accessibility)
        poi = await self.poi_service.create_poi(
            name=request.name,
            category=POICategory(request.category.value),
            latitude=request.latitude,
            longitude=request.longitude,
            address=request.address,
            owner_id=owner_id,
            phone=request.phone,
            website=request.website,
            opening_hours=request.opening_hours,
            accessibility=accessibility,
        )
        return _to_response(poi)


@dataclass
class GetPOIUseCase:
    poi_service: POIService

    async def execute(self, poi_id: UUID) -> POIResponse:
        poi = await self.poi_service.get_poi(poi_id)
        if not poi:
            raise POINotFoundError("POI not found")
        return _to_response(poi)


@dataclass
class UpdatePOIUseCase:
    poi_service: POIService

    async def execute(self, poi_id: UUID, owner_id: UUID, request: UpdatePOIRequest) -> POIResponse:
        accessibility = _to_accessibility_profile(request.accessibility) if request.accessibility else None
        try:
            poi = await self.poi_service.update_poi(
                poi_id=poi_id,
                owner_id=owner_id,
                name=request.name,
                category=POICategory(request.category.value) if request.category else None,
                address=request.address,
                phone=request.phone,
                website=request.website,
                opening_hours=request.opening_hours,
                accessibility=accessibility,
            )
        except ValueError as e:
            raise POIPermissionError(str(e))
        return _to_response(poi)


@dataclass
class DeletePOIUseCase:
    poi_service: POIService

    async def execute(self, poi_id: UUID, owner_id: UUID) -> bool:
        try:
            return await self.poi_service.delete_poi(poi_id, owner_id)
        except ValueError as e:
            raise POIPermissionError(str(e))


@dataclass
class ListPOIsUseCase:
    poi_service: POIService

    async def execute(
        self,
        category: Optional[POICategoryStr] = None,
        bounds: Optional[Tuple[Tuple[float, float], Tuple[float, float]]] = None,
        has_features: Optional[List[AccessibilityFeatureStr]] = None,
        owner_id: Optional[UUID] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> POIListResponse:
        features = [AccessibilityFeature(f.value) for f in has_features] if has_features else None
        pois = await self.poi_service.list_pois(
            category=POICategory(category.value) if category else None,
            bounds=bounds,
            has_features=features,
            owner_id=owner_id,
            limit=limit,
            offset=offset,
        )
        total = await self.poi_service.poi_repo.count(
            category=POICategory(category.value) if category else None,
            bounds=bounds,
        )
        return POIListResponse(
            pois=[_to_response(p) for p in pois],
            total=total,
            limit=limit,
            offset=offset,
        )


@dataclass
class SearchNearbyUseCase:
    poi_service: POIService

    async def execute(self, request: SearchNearbyRequest) -> POIListResponse:
        pois = await self.poi_service.search_nearby(
            latitude=request.latitude,
            longitude=request.longitude,
            radius_meters=request.radius_meters,
            category=POICategory(request.category.value) if request.category else None,
            limit=request.limit,
        )
        return POIListResponse(
            pois=[_to_response(p) for p in pois],
            total=len(pois),
            limit=request.limit,
            offset=0,
        )