from fastapi import APIRouter, Depends, HTTPException, status
from typing import Optional
from uuid import UUID

from src.application.use_cases.poi_use_cases import (
    CreatePOIUseCase,
    GetPOIUseCase,
    UpdatePOIUseCase,
    DeletePOIUseCase,
    ListPOIsUseCase,
    SearchNearbyUseCase,
)
from src.application.dto.schemas import (
    CreatePOIRequest,
    UpdatePOIRequest,
    POIResponse,
    POIListResponse,
    SearchNearbyRequest,
    ErrorResponse,
    POICategoryStr,
    AccessibilityFeatureStr,
)
from src.domain.entities.poi import POICategory, AccessibilityFeature
from src.adapters.inbound.http.security import get_current_user_id, get_current_user_id_optional


router = APIRouter(prefix="/api/v1/poi", tags=["poi"])

create_poi_uc: CreatePOIUseCase = None
get_poi_uc: GetPOIUseCase = None
update_poi_uc: UpdatePOIUseCase = None
delete_poi_uc: DeletePOIUseCase = None
list_pois_uc: ListPOIsUseCase = None
search_nearby_uc: SearchNearbyUseCase = None


def init_router(
    create_uc: CreatePOIUseCase,
    get_uc: GetPOIUseCase,
    update_uc: UpdatePOIUseCase,
    delete_uc: DeletePOIUseCase,
    list_uc: ListPOIsUseCase,
    search_uc: SearchNearbyUseCase,
):
    global create_poi_uc, get_poi_uc, update_poi_uc, delete_poi_uc, list_pois_uc, search_nearby_uc
    create_poi_uc = create_uc
    get_poi_uc = get_uc
    update_poi_uc = update_uc
    delete_poi_uc = delete_uc
    list_pois_uc = list_uc
    search_nearby_uc = search_uc


@router.post(
    "",
    response_model=POIResponse,
    status_code=status.HTTP_201_CREATED,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}},
)
async def create_poi(
    request: CreatePOIRequest,
    owner_id: UUID = Depends(get_current_user_id),
):
    try:
        return await create_poi_uc.execute(request, owner_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get(
    "/{poi_id}",
    response_model=POIResponse,
    responses={404: {"model": ErrorResponse}},
)
async def get_poi(poi_id: UUID):
    try:
        return await get_poi_uc.execute(poi_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))


@router.patch(
    "/{poi_id}",
    response_model=POIResponse,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}, 403: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def update_poi(
    poi_id: UUID,
    request: UpdatePOIRequest,
    owner_id: UUID = Depends(get_current_user_id),
):
    try:
        return await update_poi_uc.execute(poi_id, owner_id, request)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e))


@router.delete(
    "/{poi_id}",
    status_code=status.HTTP_204_NO_CONTENT,
    responses={401: {"model": ErrorResponse}, 403: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def delete_poi(poi_id: UUID, owner_id: UUID = Depends(get_current_user_id)):
    try:
        await delete_poi_uc.execute(poi_id, owner_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e))


@router.get(
    "",
    response_model=POIListResponse,
)
async def list_pois(
    category: Optional[POICategoryStr] = None,
    sw_lat: Optional[float] = None,
    sw_lon: Optional[float] = None,
    ne_lat: Optional[float] = None,
    ne_lon: Optional[float] = None,
    features: Optional[list[AccessibilityFeatureStr]] = None,
    owner_id: Optional[UUID] = Depends(get_current_user_id_optional),
    limit: int = 50,
    offset: int = 0,
):
    bounds = None
    if all(v is not None for v in [sw_lat, sw_lon, ne_lat, ne_lon]):
        bounds = ((sw_lat, sw_lon), (ne_lat, ne_lon))
    
    try:
        return await list_pois_uc.execute(
            category=category,
            bounds=bounds,
            has_features=features,
            owner_id=owner_id,
            limit=limit,
            offset=offset,
        )
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post(
    "/search/nearby",
    response_model=POIListResponse,
)
async def search_nearby(request: SearchNearbyRequest):
    try:
        return await search_nearby_uc.execute(request)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))