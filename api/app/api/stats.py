from fastapi import APIRouter

router = APIRouter()


@router.get("/stats")
async def get_stats():
    return {
        "total_sessions": 0,
        "total_credentials": 0,
        "active_phishlets": 0,
    }
