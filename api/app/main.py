"""
PHANTOM-PROXY v14.0 - FastAPI Backend
Replaces Python scripts with proper API
"""
from contextlib import asynccontextmanager
from datetime import datetime, timedelta
from typing import Optional, List
import os

from fastapi import FastAPI, Depends, HTTPException, status, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, EmailStr, Field
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import NullPool
import redis.asyncio as redis
from celery import Celery

# ============================================================================
# CONFIG
# ============================================================================

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgresql+asyncpg://phantom:phantom@localhost:5432/phantom"
)
REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379/0")
CELERY_BROKER = os.getenv("CELERY_BROKER", "redis://localhost:6379/1")

# ============================================================================
# MODELS
# ============================================================================

class TenantCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    slug: str = Field(..., min_length=1, max_length=100)
    domain: Optional[str] = None
    plan: str = "free"

class TenantResponse(TenantCreate):
    id: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

class CampaignCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=255)
    description: Optional[str] = None
    phishlet_id: str
    target_count: int = Field(default=0, ge=0)

class CampaignResponse(CampaignCreate):
    id: str
    status: str
    sent_count: int = 0
    open_count: int = 0
    click_count: int = 0
    cred_count: int = 0
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

class SessionResponse(BaseModel):
    id: str
    target_email: Optional[str]
    ip_address: Optional[str]
    mfa_bypassed: bool
    mfa_type: Optional[str]
    status: str
    created_at: datetime

    class Config:
        from_attributes = True

class PhishletCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    domain: str
    provider: str
    html_template: Optional[str] = None

class PhishletResponse(PhishletCreate):
    id: str
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True

class LoginRequest(BaseModel):
    email: EmailStr
    password: str

class LoginResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    expires_in: int = 3600

# ============================================================================
# DATABASE
# ============================================================================

engine = create_async_engine(
    DATABASE_URL,
    poolclass=NullPool,
    echo=os.getenv("SQL_ECHO", "false").lower() == "true",
)

async_session = sessionmaker(
    engine, class_=AsyncSession, expire_on_commit=False
)

async def get_db():
    async with async_session() as session:
        yield session

# ============================================================================
# REDIS
# ============================================================================

redis_client: Optional[redis.Redis] = None

async def get_redis() -> redis.Redis:
    if redis_client is None:
        raise HTTPException(status_code=503, detail="Redis not available")
    return redis_client

# ============================================================================
# CELERY
# ============================================================================

celery_app = Celery(
    "phantom",
    broker=CELERY_BROKER,
    backend=CELERY_BROKER
)

celery_app.conf.update(
    task_serializer="json",
    accept_content=["json"],
    result_serializer="json",
    timezone="UTC",
    enable_utc=True,
    task_track_started=True,
    task_time_limit=30 * 60,
    worker_prefetch_multiplier=4,
)

# ============================================================================
# AUTH
# ============================================================================

security = HTTPBearer()

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security)
):
    # In production, verify JWT token
    # For now, return mock user
    return {
        "id": "00000000-0000-0000-0000-000000000002",
        "tenant_id": "00000000-0000-0000-0000-000000000001",
        "email": "admin@phantom.local",
        "role": "admin"
    }

# ============================================================================
# LIFESPAN
# ============================================================================

@asynccontextmanager
async def lifespan(app: FastAPI):
    global redis_client
    
    # Startup
    redis_client = redis.from_url(REDIS_URL, decode_responses=True)
    
    # Initialize Redis
    await redis_client.ping()
    
    # Cache tenant info
    await redis_client.hset(
        "tenant:default",
        mapping={
            "id": "00000000-0000-0000-0000-000000000001",
            "name": "Default Organization",
            "plan": "enterprise"
        }
    )
    
    yield
    
    # Shutdown
    if redis_client:
        await redis_client.close()

# ============================================================================
# APP
# ============================================================================

app = FastAPI(
    title="PHANTOM-PROXY API",
    description="Enterprise Phishing Platform API",
    version="14.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan
)

# Middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
app.add_middleware(GZipMiddleware, minimum_size=1000)

# ============================================================================
# HEALTH
# ============================================================================

@app.get("/health")
async def health_check():
    redis_ok = False
    try:
        if redis_client:
            await redis_client.ping()
            redis_ok = True
    except:
        pass
    
    return {
        "status": "ok" if redis_ok else "degraded",
        "version": "14.0.0",
        "services": {
            "redis": "ok" if redis_ok else "unavailable",
            "database": "ok"
        },
        "timestamp": datetime.utcnow().isoformat()
    }

@app.get("/health/ready")
async def readiness_check():
    checks = {}
    
    # Redis
    try:
        await redis_client.ping()
        checks["redis"] = "ok"
    except:
        checks["redis"] = "unavailable"
    
    # Database
    try:
        async with async_session() as session:
            await session.execute("SELECT 1")
        checks["database"] = "ok"
    except:
        checks["database"] = "unavailable"
    
    all_ok = all(v == "ok" for v in checks.values())
    
    if not all_ok:
        raise HTTPException(status_code=503, detail=checks)
    
    return {"status": "ready", "checks": checks}

# ============================================================================
# AUTH
# ============================================================================

@app.post("/api/v1/auth/login", response_model=LoginResponse)
async def login(request: LoginRequest):
    # In production, verify against database
    # For demo, accept any login
    
    # Create mock token
    token = f"phantom_{request.email}_{datetime.utcnow().timestamp()}"
    
    # Cache token
    await redis_client.setex(
        f"token:{token}",
        3600,
        f"{request.email}"
    )
    
    return LoginResponse(
        access_token=token,
        expires_in=3600
    )

@app.post("/api/v1/auth/logout")
async def logout(credentials: HTTPAuthorizationCredentials = Depends(security)):
    token = credentials.credentials
    await redis_client.delete(f"token:{token}")
    return {"status": "logged_out"}

# ============================================================================
# TENANTS
# ============================================================================

@app.get("/api/v1/tenants", response_model=List[TenantResponse])
async def list_tenants(
    current_user: dict = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    # In production, query from database
    return [
        TenantResponse(
            id="00000000-0000-0000-0000-000000000001",
            name="Default Organization",
            slug="default",
            domain="phantom.local",
            plan="enterprise",
            created_at=datetime.utcnow(),
            updated_at=datetime.utcnow()
        )
    ]

@app.post("/api/v1/tenants", response_model=TenantResponse)
async def create_tenant(
    tenant: TenantCreate,
    current_user: dict = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    # In production, insert into database
    return TenantResponse(
        id="new-tenant-id",
        **tenant.model_dump(),
        created_at=datetime.utcnow(),
        updated_at=datetime.utcnow()
    )

# ============================================================================
# CAMPAIGNS
# ============================================================================

@app.get("/api/v1/campaigns", response_model=List[CampaignResponse])
async def list_campaigns(
    current_user: dict = Depends(get_current_user),
    status: Optional[str] = None,
    db: AsyncSession = Depends(get_db)
):
    # In production, query from database
    return [
        CampaignResponse(
            id="campaign-1",
            name="Q1 2024 Test",
            description="Quarterly security test",
            phishlet_id="phishlet-1",
            target_count=100,
            status="running",
            sent_count=75,
            open_count=45,
            click_count=30,
            cred_count=12,
            created_at=datetime.utcnow() - timedelta(days=7),
            updated_at=datetime.utcnow()
        )
    ]

@app.post("/api/v1/campaigns", response_model=CampaignResponse)
async def create_campaign(
    campaign: CampaignCreate,
    current_user: dict = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    # Queue campaign creation task
    celery_app.send_task(
        "phantom.tasks.create_campaign",
        args=[campaign.model_dump(), current_user["tenant_id"]]
    )
    
    return CampaignResponse(
        id="new-campaign-id",
        **campaign.model_dump(),
        status="draft",
        created_at=datetime.utcnow(),
        updated_at=datetime.utcnow()
    )

@app.get("/api/v1/campaigns/{campaign_id}", response_model=CampaignResponse)
async def get_campaign(
    campaign_id: str,
    current_user: dict = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    # Get campaign stats from Redis cache
    stats = await redis_client.hgetall(f"campaign:{campaign_id}:stats")
    
    return CampaignResponse(
        id=campaign_id,
        name="Campaign",
        description="",
        phishlet_id="phishlet-1",
        target_count=int(stats.get("target_count", 0)),
        status=stats.get("status", "running"),
        sent_count=int(stats.get("sent_count", 0)),
        open_count=int(stats.get("open_count", 0)),
        click_count=int(stats.get("click_count", 0)),
        cred_count=int(stats.get("cred_count", 0)),
        created_at=datetime.utcnow() - timedelta(days=1),
        updated_at=datetime.utcnow()
    )

@app.post("/api/v1/campaigns/{campaign_id}/start")
async def start_campaign(
    campaign_id: str,
    current_user: dict = Depends(get_current_user)
):
    # Queue campaign start task
    celery_app.send_task(
        "phantom.tasks.start_campaign",
        args=[campaign_id, current_user["tenant_id"]]
    )
    
    await redis_client.hset(
        f"campaign:{campaign_id}:stats",
        "status",
        "running"
    )
    
    return {"status": "started", "campaign_id": campaign_id}

@app.post("/api/v1/campaigns/{campaign_id}/pause")
async def pause_campaign(
    campaign_id: str,
    current_user: dict = Depends(get_current_user)
):
    await redis_client.hset(
        f"campaign:{campaign_id}:stats",
        "status",
        "paused"
    )
    
    return {"status": "paused", "campaign_id": campaign_id}

# ============================================================================
# SESSIONS
# ============================================================================

@app.get("/api/v1/sessions", response_model=List[SessionResponse])
async def list_sessions(
    current_user: dict = Depends(get_current_user),
    campaign_id: Optional[str] = None,
    limit: int = 100,
    db: AsyncSession = Depends(get_db)
):
    # Get from Redis cache
    key = f"tenant:{current_user['tenant_id']}:sessions"
    if campaign_id:
        key = f"campaign:{campaign_id}:sessions"
    
    session_ids = await redis_client.lrange(key, 0, limit - 1)
    
    sessions = []
    for sid in session_ids:
        session_data = await redis_client.hgetall(f"session:{sid}")
        if session_data:
            sessions.append(SessionResponse(
                id=sid,
                target_email=session_data.get("email"),
                ip_address=session_data.get("ip"),
                mfa_bypassed=session_data.get("mfa_bypassed") == "true",
                mfa_type=session_data.get("mfa_type"),
                status=session_data.get("status", "active"),
                created_at=datetime.fromisoformat(
                    session_data.get("created_at", datetime.utcnow().isoformat())
                )
            ))
    
    return sessions

# ============================================================================
# PHSIHLETS
# ============================================================================

@app.get("/api/v1/phishlets", response_model=List[PhishletResponse])
async def list_phishlets(
    current_user: dict = Depends(get_current_user),
    provider: Optional[str] = None,
    db: AsyncSession = Depends(get_db)
):
    return [
        PhishletResponse(
            id="phishlet-microsoft365",
            name="Microsoft 365",
            domain="login.microsoftonline.com",
            provider="microsoft365",
            is_active=True,
            created_at=datetime.utcnow() - timedelta(days=30)
        ),
        PhishletResponse(
            id="phishlet-google",
            name="Google Workspace",
            domain="accounts.google.com",
            provider="google",
            is_active=True,
            created_at=datetime.utcnow() - timedelta(days=30)
        ),
        PhishletResponse(
            id="phishlet-okta",
            name="Okta",
            domain="login.okta.com",
            provider="okta",
            is_active=True,
            created_at=datetime.utcnow() - timedelta(days=30)
        )
    ]

# ============================================================================
# STATS
# ============================================================================

@app.get("/api/v1/stats/dashboard")
async def get_dashboard_stats(
    current_user: dict = Depends(get_current_user)
):
    tenant_id = current_user["tenant_id"]
    
    # Get cached stats
    stats = await redis_client.hgetall(f"tenant:{tenant_id}:stats")
    
    return {
        "total_campaigns": int(stats.get("campaigns", 0)),
        "active_campaigns": int(stats.get("active", 0)),
        "total_captures": int(stats.get("captures", 0)),
        "mfa_bypassed": int(stats.get("mfa_bypassed", 0)),
        "click_rate": float(stats.get("click_rate", 0)),
        "recent_activity": [
            {"type": "session", "email": "user@company.com", "time": "2 min ago"},
            {"type": "campaign", "name": "Q1 Test", "time": "1 hour ago"}
        ]
    }

@app.get("/api/v1/stats/analytics")
async def get_analytics(
    current_user: dict = Depends(get_current_user),
    period: str = "7d"
):
    # Return mock analytics data
    return {
        "period": period,
        "campaigns": [
            {"name": "Q1 Test", "sent": 100, "opened": 45, "clicked": 30, "captured": 12},
            {"name": "February Test", "sent": 150, "opened": 60, "clicked": 40, "captured": 18}
        ],
        "top_targets": [
            {"email": "ceo@company.com", "clicks": 5, "captured": true},
            {"email": "cfo@company.com", "clicks": 3, "captured": true},
            {"email": "it@company.com", "clicks": 2, "captured": false}
        ],
        "geography": [
            {"country": "US", "count": 45},
            {"country": "UK", "count": 30},
            {"country": "DE", "count": 15}
        ]
    }

# ============================================================================
# WEBSOCKET
# ============================================================================

from fastapi import WebSocket, WebSocketDisconnect

class ConnectionManager:
    def __init__(self):
        self.active_connections: dict[str, WebSocket] = {}
    
    async def connect(self, websocket: WebSocket, client_id: str):
        await websocket.accept()
        self.active_connections[client_id] = websocket
    
    def disconnect(self, client_id: str):
        if client_id in self.active_connections:
            del self.active_connections[client_id]
    
    async def send_personal_message(self, message: dict, client_id: str):
        if client_id in self.active_connections:
            await self.active_connections[client_id].send_json(message)
    
    async def broadcast(self, message: dict):
        for connection in self.active_connections.values():
            await connection.send_json(message)

manager = ConnectionManager()

@app.websocket("/ws/{client_id}")
async def websocket_endpoint(websocket: WebSocket, client_id: str):
    await manager.connect(websocket, client_id)
    try:
        while True:
            data = await websocket.receive_text()
            # Handle incoming messages
            await manager.send_personal_message({"status": "ok"}, client_id)
    except WebSocketDisconnect:
        manager.disconnect(client_id)

# ============================================================================
# RUN
# ============================================================================

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8000,
        reload=os.getenv("RELOAD", "false").lower() == "true",
        workers=int(os.getenv("WORKERS", "4"))
    )
