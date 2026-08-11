import asyncio
from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from uuid import UUID

from src.config import settings
from src.adapters.out.persistence.database import init_db, close_db
from src.adapters.out.eventbus.redis_streams import RedisEventBus
from src.adapters.out.storage.s3_client import S3Storage
from src.adapters.out.persistence.repository import (
    PostgresBarrierRepository,
    PostgresBarrierPhotoRepository,
    PostgresBarrierConfirmationRepository,
    PostgresBarrierComplaintRepository,
)
from src.domain.services.barrier_service import BarrierService
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
from src.adapters.inbound.http.router import router as barriers_router, init_router


event_bus = RedisEventBus()
photo_storage = S3Storage()

barrier_repo = PostgresBarrierRepository()
photo_repo = PostgresBarrierPhotoRepository()
confirmation_repo = PostgresBarrierConfirmationRepository()
complaint_repo = PostgresBarrierComplaintRepository()

barrier_service = BarrierService(
    barrier_repo=barrier_repo,
    photo_repo=photo_repo,
    confirmation_repo=confirmation_repo,
    complaint_repo=complaint_repo,
    photo_storage=photo_storage,
    event_bus=event_bus,
)

create_barrier_uc = CreateBarrierUseCase(barrier_service=barrier_service)
get_barrier_uc = GetBarrierUseCase(
    barrier_service=barrier_service,
    photo_repo=photo_repo,
    confirmation_repo=confirmation_repo,
    complaint_repo=complaint_repo,
    photo_storage=photo_storage,
)
list_barriers_uc = ListBarriersUseCase(barrier_service=barrier_service)
upload_photo_uc = UploadPhotoUseCase(barrier_service=barrier_service)
confirm_barrier_uc = ConfirmBarrierUseCase(barrier_service=barrier_service)
complain_barrier_uc = ComplainBarrierUseCase(barrier_service=barrier_service)
approve_barrier_uc = ApproveBarrierUseCase(barrier_service=barrier_service)
reject_barrier_uc = RejectBarrierUseCase(barrier_service=barrier_service)
resolve_barrier_uc = ResolveBarrierUseCase(barrier_service=barrier_service)

init_router(
    create_uc=create_barrier_uc,
    get_uc=get_barrier_uc,
    list_uc=list_barriers_uc,
    upload_uc=upload_photo_uc,
    confirm_uc=confirm_barrier_uc,
    complain_uc=complain_barrier_uc,
    approve_uc=approve_barrier_uc,
    reject_uc=reject_barrier_uc,
    resolve_uc=resolve_barrier_uc,
)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    await init_db()
    await event_bus.connect()

    async def on_approved(event: dict) -> None:
        payload = event.get("payload", {})
        await barrier_service.handle_approved_event(
            UUID(event["aggregate_id"]), UUID(payload["moderator_id"])
        )

    async def on_rejected(event: dict) -> None:
        payload = event.get("payload", {})
        await barrier_service.handle_rejected_event(
            UUID(event["aggregate_id"]), UUID(payload["moderator_id"])
        )

    async def on_resolved(event: dict) -> None:
        await barrier_service.handle_resolved_event(UUID(event["aggregate_id"]))

    asyncio.create_task(event_bus.subscribe("barrier.approved", "barrier-group", "barrier-consumer-1", on_approved))
    asyncio.create_task(event_bus.subscribe("barrier.rejected", "barrier-group", "barrier-consumer-2", on_rejected))
    asyncio.create_task(event_bus.subscribe("barrier.resolved", "barrier-group", "barrier-consumer-3", on_resolved))

    yield
    await event_bus.disconnect()
    await close_db()


app = FastAPI(
    title="Barrier Service",
    description="CRUD barriers, photos, confirmations, complaints",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(barriers_router)


@app.get("/health")
async def health_check():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host=settings.HOST, port=settings.PORT)