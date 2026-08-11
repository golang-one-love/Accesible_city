import builtins
from math import asin, cos, radians, sin, sqrt
from uuid import UUID

from sqlalchemy import and_, func, select

from src.adapters.out.persistence.database import async_session_factory
from src.adapters.out.persistence.models import POIModel
from src.domain.entities.poi import (
    AccessibilityFeature,
    AccessibilityProfile,
    POICategory,
    PointOfInterest,
)
from src.domain.repositories import POIRepository


def _haversine_km(lat1: float, lon1: float, lat2: float, lon2: float) -> float:
    r_km = 6371.0
    dlat = radians(lat2 - lat1)
    dlon = radians(lon2 - lon1)
    a = sin(dlat / 2) ** 2 + cos(radians(lat1)) * cos(radians(lat2)) * sin(dlon / 2) ** 2
    return 2 * r_km * asin(sqrt(a))


class PostgresPOIRepository(POIRepository):
    async def create(self, poi: PointOfInterest) -> PointOfInterest:
        async with async_session_factory() as session:
            model = self._to_model(poi)
            session.add(model)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def get_by_id(self, poi_id: UUID) -> PointOfInterest | None:
        async with async_session_factory() as session:
            stmt = select(POIModel).where(POIModel.id == poi_id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if model:
                return self._to_entity(model)
            return None

    async def update(self, poi: PointOfInterest) -> PointOfInterest:
        async with async_session_factory() as session:
            stmt = select(POIModel).where(POIModel.id == poi.id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if not model:
                raise ValueError("POI not found")
            self._update_model(model, poi)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def delete(self, poi_id: UUID) -> bool:
        async with async_session_factory() as session:
            stmt = select(POIModel).where(POIModel.id == poi_id)
            result = await session.execute(stmt)
            model = result.scalar_one_or_none()
            if model:
                await session.delete(model)
                await session.commit()
                return True
            return False

    async def list(
        self,
        category: POICategory | None = None,
        bounds: tuple[tuple[float, float], tuple[float, float]] | None = None,
        has_features: list[AccessibilityFeature] | None = None,
        owner_id: UUID | None = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list[PointOfInterest]:
        async with async_session_factory() as session:
            stmt = select(POIModel)

            conditions = []
            if category:
                conditions.append(POIModel.category == category.value)
            if owner_id:
                conditions.append(POIModel.owner_id == owner_id)
            if bounds:
                sw, ne = bounds
                conditions.append(
                    and_(
                        POIModel.latitude >= sw[0],
                        POIModel.latitude <= ne[0],
                        POIModel.longitude >= sw[1],
                        POIModel.longitude <= ne[1],
                    )
                )
            if has_features:
                for feature in has_features:
                    conditions.append(POIModel.accessibility_features.contains([feature.value]))

            if conditions:
                stmt = stmt.where(and_(*conditions))

            stmt = stmt.order_by(POIModel.created_at.desc()).limit(limit).offset(offset)
            result = await session.execute(stmt)
            return [self._to_entity(m) for m in result.scalars().all()]

    async def count(
        self,
        category: POICategory | None = None,
        bounds: tuple[tuple[float, float], tuple[float, float]] | None = None,
    ) -> int:
        async with async_session_factory() as session:
            stmt = select(func.count(POIModel.id))
            conditions = []
            if category:
                conditions.append(POIModel.category == category.value)
            if bounds:
                sw, ne = bounds
                conditions.append(
                    and_(
                        POIModel.latitude >= sw[0],
                        POIModel.latitude <= ne[0],
                        POIModel.longitude >= sw[1],
                        POIModel.longitude <= ne[1],
                    )
                )
            if conditions:
                stmt = stmt.where(and_(*conditions))
            result = await session.execute(stmt)
            return result.scalar() or 0

    async def search_nearby(
        self,
        latitude: float,
        longitude: float,
        radius_meters: float,
        category: POICategory | None = None,
        limit: int = 20,
    ) -> builtins.list[PointOfInterest]:
        async with async_session_factory() as session:
            deg_per_meter = 1 / 111000
            delta = radius_meters * deg_per_meter
            
            stmt = select(POIModel).where(
                and_(
                    POIModel.latitude.between(latitude - delta, latitude + delta),
                    POIModel.longitude.between(longitude - delta, longitude + delta),
                )
            )
            if category:
                stmt = stmt.where(POIModel.category == category.value)
            
            stmt = stmt.limit(limit)
            result = await session.execute(stmt)
            
            pois = [self._to_entity(m) for m in result.scalars().all()]

            filtered = []
            for poi in pois:
                dist = _haversine_km(latitude, longitude, poi.latitude, poi.longitude) * 1000
                if dist <= radius_meters:
                    filtered.append(poi)

            return filtered[:limit]

    def _to_model(self, poi: PointOfInterest) -> POIModel:
        return POIModel(
            id=poi.id,
            name=poi.name,
            category=poi.category.value,
            latitude=poi.latitude,
            longitude=poi.longitude,
            address=poi.address,
            phone=poi.phone,
            website=poi.website,
            opening_hours=poi.opening_hours,
            accessibility_features=[f.value for f in poi.accessibility.features],
            entrance_step_height_cm=(
                float(poi.accessibility.entrance_step_height_cm)
                if poi.accessibility.entrance_step_height_cm is not None
                else None
            ),
            door_width_cm=(
                float(poi.accessibility.door_width_cm) if poi.accessibility.door_width_cm is not None else None
            ),
            has_accessible_toilet=poi.accessibility.has_accessible_toilet,
            accessibility_notes=poi.accessibility.notes,
            owner_id=poi.owner_id,
            is_verified=poi.is_verified,
            created_at=poi.created_at,
            updated_at=poi.updated_at,
        )

    def _update_model(self, model: POIModel, poi: PointOfInterest) -> None:
        model.name = poi.name
        model.category = poi.category.value
        model.latitude = poi.latitude
        model.longitude = poi.longitude
        model.address = poi.address
        model.phone = poi.phone
        model.website = poi.website
        model.opening_hours = poi.opening_hours
        model.accessibility_features = [f.value for f in poi.accessibility.features]
        model.entrance_step_height_cm = poi.accessibility.entrance_step_height_cm
        model.door_width_cm = poi.accessibility.door_width_cm
        model.has_accessible_toilet = poi.accessibility.has_accessible_toilet
        model.accessibility_notes = poi.accessibility.notes
        model.updated_at = poi.updated_at

    def _to_entity(self, model: POIModel) -> PointOfInterest:
        return PointOfInterest(
            id=model.id,
            name=model.name,
            category=POICategory(model.category),
            latitude=model.latitude,
            longitude=model.longitude,
            address=model.address,
            phone=model.phone,
            website=model.website,
            opening_hours=model.opening_hours,
            accessibility=AccessibilityProfile(
                features=[AccessibilityFeature(f) for f in model.accessibility_features],
                entrance_step_height_cm=(
                    int(model.entrance_step_height_cm) if model.entrance_step_height_cm is not None else None
                ),
                door_width_cm=int(model.door_width_cm) if model.door_width_cm is not None else None,
                has_accessible_toilet=model.has_accessible_toilet,
                notes=model.accessibility_notes,
            ),
            owner_id=model.owner_id,
            is_verified=model.is_verified,
            created_at=model.created_at,
            updated_at=model.updated_at,
        )