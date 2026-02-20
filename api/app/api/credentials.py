from fastapi import APIRouter

router = APIRouter()


@router.get("")
async def list_credentials(limit: int = 50, offset: int = 0):
    return {"credentials": [], "total": 0}
