"""
PhantomProxy API - Enterprise Killer v13.0
FastAPI + SQLModel + OpenTelemetry
"""
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.core.config import settings
from app.core.telemetry import setup_telemetry
from app.api import sessions, credentials, stats


@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_telemetry(settings.service_name)
    yield
    # shutdown


app = FastAPI(
    title="PhantomProxy API",
    version="13.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(sessions.router, prefix="/api/v1/sessions", tags=["sessions"])
app.include_router(credentials.router, prefix="/api/v1/credentials", tags=["credentials"])
app.include_router(stats.router, prefix="/api/v1", tags=["stats"])


@app.get("/health")
def health():
    return {"status": "ok", "service": "phantomproxy-api"}
