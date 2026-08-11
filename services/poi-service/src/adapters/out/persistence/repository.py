from typing import Optional, List, Tuple
from uuid import UUID
from sqlalchemy import select, func, and_, or_
from sqlalchemy.ext.asyncio import AsyncSession

from src.domain.entities.poi import PointOfInterest, POICategory, AccessibilityFeature, AccessibilityProfile
from src.domain.repositories import POIRepository
from src.adapters.out.persistence.models import POIModel
from src.adapters.out.persistence.database import async_session_factory


class PostgresPOIRepository(POIRepository):
    async def create(self, poi: PointOfInterest) -> PointOfInterest:
        async with async_session_factory() as session:
            model = self._to_model(poi)
            session.add(model)
            await session.commit()
            await session.refresh(model)
            return self._to_entity(model)

    async def get_by_id(self, poi_id: UUID) -> Optional[PointOfInterest]:
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
                return None
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
        category: Optional[POICategory] = None,
        bounds: Optional[Tuple[Tuple[float, float], Tuple[float, float]]] = None,
        has_features: Optional[List[AccessibilityFeature]] = None,
        owner_id: Optional[UUID] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> List[PointOfInterest]:
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
        category: Optional[POICategory] = None,
        bounds: Optional[Tuple[Tuple[float, float], Tuple[float, float]]] = None,
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
        category: Optional[POICategory] = None,
        limit: int = 20,
    ) -> List[PointOfInterest]:
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
                dist = poi.accessibility.distance_to(latitude, longitude) if hasattr(poi.accessibility, 'distance_to') else 0
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
            entrance_step_height_cm=poi.accessibility.entrance_step_height_cm,
            door_width_cm=poi.accessibility.door_width_cm,
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
                entrance_step_height_cm=model.entrance_step_height_cm,
                door_width_cm=model.door_width_cm,
                has_accessible_toilet=model.has_accessible_toilet,
                notes=model.accessibility_notes,
            ),
            owner_id=model.owner_id,
            is_verified=model.is_verified,
            created_at=model.created_at,
            updated_at=model.updated_at,
        )