from fastapi import APIRouter, Depends, HTTPException, status, UploadFile, File, Form
from typing import Optional
from uuid import UUID

from src.application.use_cases.barrier_use_cases import (
    CreateBarrierUseCase,
    GetBarrierUseCase,
    ListBarriersUseCase,
    UploadPhotoUseCase,
    ConfirmBarrierUseCase,
    ComplainBarrierUseCase,
    ApproveBarrierUseCase,
    RejectBarrierUseCase,
    ResolveBarrierUseCase,
)
from src.application.dto.schemas import (
    CreateBarrierRequest,
    BarrierResponse,
    BarrierListResponse,
    UploadPhotoResponse,
    ConfirmBarrierResponse,
    ComplainBarrierRequest,
    ComplainBarrierResponse,
    ErrorResponse,
    BarrierTypeStr,
    BarrierStatusStr,
    SeverityInt,
)
from src.domain.entities.barrier import BarrierType, BarrierStatus, Severity
from src.domain.value_objects.coordinates import Coordinates
from src.adapters.inbound.http.security import get_current_user_id, get_current_user_id_optional


router = APIRouter(prefix="/api/v1/barriers", tags=["barriers"])

create_barrier_uc: CreateBarrierUseCase = None
get_barrier_uc: GetBarrierUseCase = None
list_barriers_uc: ListBarriersUseCase = None
upload_photo_uc: UploadPhotoUseCase = None
confirm_barrier_uc: ConfirmBarrierUseCase = None
complain_barrier_uc: ComplainBarrierUseCase = None
approve_barrier_uc: ApproveBarrierUseCase = None
reject_barrier_uc: RejectBarrierUseCase = None
resolve_barrier_uc: ResolveBarrierUseCase = None


def init_router(
    create_uc: CreateBarrierUseCase,
    get_uc: GetBarrierUseCase,
    list_uc: ListBarriersUseCase,
    upload_uc: UploadPhotoUseCase,
    confirm_uc: ConfirmBarrierUseCase,
    complain_uc: ComplainBarrierUseCase,
    approve_uc: ApproveBarrierUseCase,
    reject_uc: RejectBarrierUseCase,
    resolve_uc: ResolveBarrierUseCase,
):
    global create_barrier_uc, get_barrier_uc, list_barriers_uc, upload_photo_uc
    global confirm_barrier_uc, complain_barrier_uc, approve_barrier_uc, reject_barrier_uc, resolve_barrier_uc
    
    create_barrier_uc = create_uc
    get_barrier_uc = get_uc
    list_barriers_uc = list_uc
    upload_photo_uc = upload_uc
    confirm_barrier_uc = confirm_uc
    complain_barrier_uc = complain_uc
    approve_barrier_uc = approve_uc
    reject_barrier_uc = reject_uc
    resolve_barrier_uc = resolve_uc


@router.post(
    "",
    response_model=BarrierResponse,
    status_code=status.HTTP_201_CREATED,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}},
)
async def create_barrier(
    request: CreateBarrierRequest,
    user_id: UUID = Depends(get_current_user_id),
):
    try:
        return await create_barrier_uc.execute(request, user_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get(
    "/{barrier_id}",
    response_model=BarrierResponse,
    responses={404: {"model": ErrorResponse}},
)
async def get_barrier(barrier_id: UUID):
    try:
        return await get_barrier_uc.execute(barrier_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))


@router.get(
    "",
    response_model=BarrierListResponse,
)
async def list_barriers(
    status: Optional[BarrierStatusStr] = None,
    type: Optional[BarrierTypeStr] = None,
    severity_min: Optional[SeverityInt] = None,
    severity_max: Optional[SeverityInt] = None,
    sw_lat: Optional[float] = None,
    sw_lon: Optional[float] = None,
    ne_lat: Optional[float] = None,
    ne_lon: Optional[float] = None,
    limit: int = 100,
    offset: int = 0,
):
    from src.application.dto.schemas import CoordinatesDTO
    
    bounds = None
    if all(v is not None for v in [sw_lat, sw_lon, ne_lat, ne_lon]):
        bounds = (
            CoordinatesDTO(latitude=sw_lat, longitude=sw_lon),
            CoordinatesDTO(latitude=ne_lat, longitude=ne_lon),
        )
    
    try:
        return await list_barriers_uc.execute(
            status=status,
            type=type,
            severity_min=severity_min,
            severity_max=severity_max,
            bounds=bounds,
            limit=limit,
            offset=offset,
        )
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post(
    "/{barrier_id}/photos",
    response_model=UploadPhotoResponse,
    status_code=status.HTTP_201_CREATED,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def upload_photo(
    barrier_id: UUID,
    file: UploadFile = File(...),
    user_id: UUID = Depends(get_current_user_id),
):
    try:
        data = await file.read()
        return await upload_photo_uc.execute(barrier_id, user_id, file.filename, file.content_type, data)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post(
    "/{barrier_id}/confirm",
    response_model=ConfirmBarrierResponse,
    status_code=status.HTTP_201_CREATED,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def confirm_barrier(barrier_id: UUID, user_id: UUID = Depends(get_current_user_id)):
    try:
        return await confirm_barrier_uc.execute(barrier_id, user_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post(
    "/{barrier_id}/complaint",
    response_model=ComplainBarrierResponse,
    status_code=status.HTTP_201_CREATED,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def complain_barrier(
    barrier_id: UUID,
    request: ComplainBarrierRequest,
    user_id: UUID = Depends(get_current_user_id),
):
    try:
        return await complain_barrier_uc.execute(barrier_id, user_id, request)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post(
    "/{barrier_id}/approve",
    response_model=BarrierResponse,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}, 403: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def approve_barrier(barrier_id: UUID, user_id: UUID = Depends(get_current_user_id)):
    try:
        return await approve_barrier_uc.execute(barrier_id, user_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e))


@router.post(
    "/{barrier_id}/reject",
    response_model=BarrierResponse,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}, 403: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def reject_barrier(barrier_id: UUID, user_id: UUID = Depends(get_current_user_id)):
    try:
        return await reject_barrier_uc.execute(barrier_id, user_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e))


@router.post(
    "/{barrier_id}/resolve",
    response_model=BarrierResponse,
    responses={400: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def resolve_barrier(barrier_id: UUID):
    try:
        return await resolve_barrier_uc.execute(barrier_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))