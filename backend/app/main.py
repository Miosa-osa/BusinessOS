from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

# Auth is now handled by Better Auth on the frontend
from app.routers import chat_router, projects_router, contexts_router, mcp_router, team_router, dashboard_router, daily_logs_router, settings_router, artifacts_router, nodes_router
from app.config import get_settings

settings = get_settings()

app = FastAPI(
    title="Business OS",
    description="Your internal command center",
    version="1.0.0",
)

# CORS middleware - allow credentials for Better Auth cookies
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173", "http://localhost:5174", "http://localhost:3000"],  # SvelteKit dev
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
    expose_headers=["X-Conversation-Id"],
)

# Include routers (auth is handled by Better Auth on frontend)
app.include_router(chat_router)
app.include_router(projects_router)
app.include_router(contexts_router)
app.include_router(mcp_router)
app.include_router(team_router)
app.include_router(dashboard_router)
app.include_router(daily_logs_router)
app.include_router(settings_router)
app.include_router(artifacts_router)
app.include_router(nodes_router)


@app.get("/")
async def root():
    return {"message": "Business OS API", "version": "1.0.0"}


@app.get("/health")
async def health_check():
    return {"status": "healthy"}
