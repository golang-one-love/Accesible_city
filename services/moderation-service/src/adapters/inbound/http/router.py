from fastapi import APIRouter, Depends, HTTPException, status
from typing import Optional
from uuid import UUID

from src.application.use_cases.moderation_use_cases import (
    GetQueueUseCase,
    GetRequestUseCase,
    ApproveRequestUseCase,
    RejectRequestUseCase,
)
from src.application.dto.schemas import (
    ModerationQueueResponse,
    ModerationRequestResponse,
    ModerationActionRequest,
    ModerationActionResponse,
    ErrorResponse,
    ModerationStatusStr,
)
from src.domain.entities.moderation import ModerationStatus
from src.adapters.inbound.http.security import get_current_user_id


router = APIRouter(prefix="/api/v1/moderation", tags=["moderation"])

get_queue_uc: GetQueueUseCase = None
get_request_uc: GetRequestUseCase = None
approve_request_uc: ApproveRequestUseCase = None
reject_request_uc: RejectRequestUseCase = None


def init_router(
    get_queue: GetQueueUseCase,
    get_request: GetRequestUseCase,
    approve_request: ApproveRequestUseCase,
    reject_request: RejectRequestUseCase,
):
    global get_queue_uc, get_request_uc, approve_request_uc, reject_request_uc
    get_queue_uc = get_queue
    get_request_uc = get_request
    approve_request_uc = approve_request
    reject_request_uc = reject_request


@router.get(
    "/queue",
    response_model=ModerationQueueResponse,
    responses={401: {"model": ErrorResponse}},
)
async def get_queue(
    status: Optional[ModerationStatusStr] = None,
    limit: int = 50,
    offset: int = 0,
    moderator_id: UUID = Depends(get_current_user_id),
):
    mod_status = ModerationStatus(status.value) if status else None
    try:
        return await get_queue_uc.execute(status=mod_status, limit=limit, offset=offset)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get(
    "/queue/{request_id}",
    response_model=ModerationRequestResponse,
    responses={401: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def get_request(request_id: UUID, moderator_id: UUID = Depends(get_current_user_id)):
    try:
        return await get_request_uc.execute(request_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))


@router.post(
    "/queue/{request_id}/approve",
    response_model=ModerationActionResponse,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}, 403: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def approve_request(
    request_id: UUID,
    request: ModerationActionRequest,
    moderator_id: UUID = Depends(get_current_user_id),
):
    try:
        return await approve_request_uc.execute(request_id, moderator_id, request)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e))


@router.post(
    "/queue/{request_id}/reject",
    response_model=ModerationActionResponse,
    responses={400: {"model": ErrorResponse}, 401: {"model": ErrorResponse}, 403: {"model": ErrorResponse}, 404: {"model": ErrorResponse}},
)
async def reject_request(
    request_id: UUID,
    request: ModerationActionRequest,
    moderator_id: UUID = Depends(get_current_user_id),
):
    try:
        return await reject_request_uc.execute(request_id, moderator_id, request)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail=str(e))