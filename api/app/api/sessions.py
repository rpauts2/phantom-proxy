from fastapi import APIRouter

router = APIRouter()


@router.get("")
async def list_sessions(limit: int = 50, offset: int = 0):
    return {"sessions": [], "total": 0}


@router.get("/{id}")
async def get_session(id: str):
    return {"session": None}
