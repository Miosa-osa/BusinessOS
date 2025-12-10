# Auth is now handled by Better Auth on the frontend
from app.routers.chat import router as chat_router
from app.routers.projects import router as projects_router
from app.routers.contexts import router as contexts_router
from app.routers.mcp import router as mcp_router
from app.routers.team import router as team_router
from app.routers.dashboard import router as dashboard_router
from app.routers.daily_logs import router as daily_logs_router
from app.routers.settings import router as settings_router
from app.routers.artifacts import router as artifacts_router
from app.routers.nodes import router as nodes_router

__all__ = ["chat_router", "projects_router", "contexts_router", "mcp_router", "team_router", "dashboard_router", "daily_logs_router", "settings_router", "artifacts_router", "nodes_router"]
