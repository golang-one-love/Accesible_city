from typing import cast
from uuid import UUID

from fastapi import APIRouter, Depends, File, HTTPException, Query, UploadFile, status

from src.adapters.inbound.http.security import get_current_user_id
from src.application.dto.schemas import (
    BarrierListResponse,
    BarrierResponse,
    BarrierStatusStr,
    BarrierTypeStr,
    ComplainBarrierRequest,
    ComplainBarrierResponse,
    ConfirmBarrierResponse,
    CreateBarrierRequest,
    ErrorResponse,
    SeverityInt,
    UploadPhotoResponse,
)
from src.application.use_cases.barrier_use_cases import (
    ApproveBarrierUseCase,
    BarrierNotFoundError,
    ComplainBarrierUseCase,
    ConfirmBarrierUseCase,
    CreateBarrierUseCase,
    GetBarrierUseCase,
    ListBarriersUseCase,
    RejectBarrierUseCase,
    ResolveBarrierUseCase,
    UploadPhotoUseCase,
)

router = APIRouter(prefix="/api/v1/barriers", tags=["barriers"])

create_barrier_uc = cast(CreateBarrierUseCase, None)
get_barrier_uc = cast(GetBarrierUseCase, None)
list_barriers_uc = cast(ListBarriersUseCase, None)
upload_photo_uc = cast(UploadPhotoUseCase, None)
confirm_barrier_uc = cast(ConfirmBarrierUseCase, None)
complain_barrier_uc = cast(ComplainBarrierUseCase, None)
approve_barrier_uc = cast(ApproveBarrierUseCase, None)
reject_barrier_uc = cast(RejectBarrierUseCase, None)
resolve_barrier_uc = cast(ResolveBarrierUseCase, None)


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
    except BarrierNotFoundError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))


@router.get(
    "",
    response_model=BarrierListResponse,
)
async def list_barriers(
    barrier_status: BarrierStatusStr | None = Query(default=None, alias="status"),
    type: BarrierTypeStr | None = None,
    severity_min: SeverityInt | None = None,
    severity_max: SeverityInt | None = None,
    sw_lat: float | None = None,
    sw_lon: float | None = None,
    ne_lat: float | None = None,
    ne_lon: float | None = None,
    limit: int = 100,
    offset: int = 0,
):
    from src.application.dto.schemas import CoordinatesDTO
    
    bounds = None
    if sw_lat is not None and sw_lon is not None and ne_lat is not None and ne_lon is not None:
        bounds = (
            CoordinatesDTO(latitude=sw_lat, longitude=sw_lon),
            CoordinatesDTO(latitude=ne_lat, longitude=ne_lon),
        )
    
    try:
        return await list_barriers_uc.execute(
            status=barrier_status,
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
        return await upload_photo_uc.execute(
            barrier_id,
            user_id,
            file.filename or "upload",
            file.content_type or "application/octet-stream",
            data,
        )
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