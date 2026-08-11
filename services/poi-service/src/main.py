import asyncio
from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.config import settings
from src.adapters.out.persistence.database import init_db, close_db
from src.adapters.out.persistence.repository import PostgresPOIRepository
from src.domain.services.poi_service import POIService
from src.application.use_cases.poi_use_cases import (
    CreatePOIUseCase,
    GetPOIUseCase,
    UpdatePOIUseCase,
    DeletePOIUseCase,
    ListPOIsUseCase,
    SearchNearbyUseCase,
)
from src.adapters.inbound.http.router import router as poi_router, init_router


poi_repo = PostgresPOIRepository()
poi_service = POIService(poi_repo=poi_repo)

create_poi_uc = CreatePOIUseCase(poi_service=poi_service)
get_poi_uc = GetPOIUseCase(poi_service=poi_service)
update_poi_uc = UpdatePOIUseCase(poi_service=poi_service)
delete_poi_uc = DeletePOIUseCase(poi_service=poi_service)
list_pois_uc = ListPOIsUseCase(poi_service=poi_service)
search_nearby_uc = SearchNearbyUseCase(poi_service=poi_service)

init_router(
    create_uc=create_poi_uc,
    get_uc=get_poi_uc,
    update_uc=update_poi_uc,
    delete_uc=delete_poi_uc,
    list_uc=list_pois_uc,
    search_uc=search_nearby_uc,
)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    await init_db()
    yield
    await close_db()


app = FastAPI(
    title="POI Service",
    description="Points of Interest with accessibility information",
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

app.include_router(poi_router)


@app.get("/health")
async def health_check():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host=settings.HOST, port=settings.PORT)