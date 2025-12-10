from typing import Annotated
from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.services.mcp_server import get_mcp_manager
from app.utils.auth import CurrentUser

router = APIRouter(prefix="/api/mcp", tags=["mcp"])


class ToolExecuteRequest(BaseModel):
    tool_name: str
    arguments: dict = {}


class ToolResponse(BaseModel):
    success: bool
    result: str | None = None
    error: str | None = None


@router.get("/tools")
async def list_tools(current_user: CurrentUser):
    """List all available MCP tools."""
    manager = get_mcp_manager()
    return {"tools": manager.get_all_tools()}


@router.post("/execute", response_model=ToolResponse)
async def execute_tool(
    request: ToolExecuteRequest,
    current_user: CurrentUser,
    db: Annotated[AsyncSession, Depends(get_db)],
):
    """Execute an MCP tool."""
    manager = get_mcp_manager()

    try:
        result = await manager.execute_tool(
            tool_name=request.tool_name,
            arguments=request.arguments,
            db_session=db,
            user_id=current_user.id,
        )
        return ToolResponse(success=True, result=result)
    except Exception as e:
        return ToolResponse(success=False, error=str(e))


@router.get("/health")
async def mcp_health():
    """Check MCP service health."""
    manager = get_mcp_manager()
    return {
        "status": "healthy",
        "builtin_tools": len(manager.get_all_tools()),
        "custom_tools": len(manager.custom_tools),
    }
