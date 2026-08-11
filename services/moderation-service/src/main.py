import asyncio
from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.config import settings
from src.adapters.out.persistence.database import init_db, close_db
from src.adapters.out.eventbus.redis_streams import RedisEventBus
from src.adapters.out.persistence.repository import PostgresModerationRepository
from src.domain.services.moderation_service import ModerationService
from src.application.use_cases.moderation_use_cases import (
    GetQueueUseCase,
    GetRequestUseCase,
    ApproveRequestUseCase,
    RejectRequestUseCase,
)
from src.adapters.inbound.http.router import router as moderation_router, init_router
from src.domain.events import BarrierCreated
from uuid import UUID


event_bus = RedisEventBus()
moderation_repo = PostgresModerationRepository()
moderation_service = ModerationService(
    moderation_repo=moderation_repo,
    event_bus=event_bus,
    barrier_service_url=settings.BARRIER_SERVICE_URL,
)

get_queue_uc = GetQueueUseCase(moderation_service=moderation_service)
get_request_uc = GetRequestUseCase(moderation_service=moderation_service)
approve_request_uc = ApproveRequestUseCase(moderation_service=moderation_service)
reject_request_uc = RejectRequestUseCase(moderation_service=moderation_service)

init_router(
    get_queue=get_queue_uc,
    get_request=get_request_uc,
    approve_request=approve_request_uc,
    reject_request=reject_request_uc,
)


async def barrier_created_handler(event_data: dict) -> None:
    payload = event_data.get("payload", {})
    event = BarrierCreated(
        barrier_id=UUID(event_data.get("aggregate_id", "")),
        payload=payload,
    )
    await moderation_service.handle_barrier_created(event)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    await init_db()
    await event_bus.connect()
    asyncio.create_task(event_bus.subscribe(
        settings.STREAM_BARRIER_CREATED,
        settings.CONSUMER_GROUP_MODERATION,
        "moderation-consumer-1",
        barrier_created_handler,
    ))
    yield
    await event_bus.disconnect()
    await close_db()


app = FastAPI(
    title="Moderation Service",
    description="Moderation queue for barrier reports",
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

app.include_router(moderation_router)


@app.get("/health")
async def health_check():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host=settings.HOST, port=settings.PORT)