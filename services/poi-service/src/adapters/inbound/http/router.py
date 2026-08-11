from typing import cast
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status

from src.adapters.inbound.http.security import (
    get_current_user_id,
    get_current_user_id_optional,
)
from src.application.dto.schemas import (
    AccessibilityFeatureStr,
    CreatePOIRequest,
    ErrorResponse,
    POICategoryStr,
    POIListResponse,
    POIResponse,
    SearchNearbyRequest,
    UpdatePOIRequest,
)
from src.application.use_cases.poi_use_cases import (
    CreatePOIUseCase,
    DeletePOIUseCase,
    GetPOIUseCase,
    ListPOIsUseCase,
    SearchNearbyUseCase,
    UpdatePOIUseCase,
)

router = APIRouter(prefix="/api/v1/poi", tags=["poi"])

create_poi_uc = cast(CreatePOIUseCase, None)
get_poi_uc = cast(GetPOIUseCase, None)
update_poi_uc = cast(UpdatePOIUseCase, None)
delete_poi_uc = cast(DeletePOIUseCase, None)
list_pois_uc = cast(ListPOIsUseCase, None)
search_nearby_uc = cast(SearchNearbyUseCase, None)


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
    category: POICategoryStr | None = None,
    sw_lat: float | None = None,
    sw_lon: float | None = None,
    ne_lat: float | None = None,
    ne_lon: float | None = None,
    features: list[AccessibilityFeatureStr] | None = None,
    owner_id: UUID | None = Depends(get_current_user_id_optional),
    limit: int = 50,
    offset: int = 0,
):
    bounds = None
    if sw_lat is not None and sw_lon is not None and ne_lat is not None and ne_lon is not None:
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