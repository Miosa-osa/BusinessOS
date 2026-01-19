# OSA Voice Agent Architecture

## ✅ PROPERLY STRUCTURED - NOT JUST ONE FILE

The Python voice agent is now properly structured with hierarchical context management and tool calling.

---

## 📁 File Structure

```
python-voice-agent/
├── agent.py                # Main entrypoint (LiveKit worker)
├── config.py               # Configuration management
├── prompts.py              # OSA personality (Level 0)
├── context.py              # User context fetching (Level 1)
├── tools.py                # Tool definitions & handlers (Level 2-3)
├── requirements.txt        # Python dependencies
├── setup_env.sh            # Environment variable setup script
├── Dockerfile              # Cloud Run deployment
├── README.md               # Quick start guide
├── TESTING.md              # Testing checklist
├── DEPLOYMENT.md           # Deployment instructions
└── ARCHITECTURE.md         # This file
```

---

## 🏗️ Hierarchical Context Architecture

### Level 0: Identity (Always Loaded - ~200 tokens)
**File:** `prompts.py`

- OSA personality and voice style
- Core behavioral guidelines
- "Never say" rules
- Response patterns

### Level 1: User Context (Session Start - ~300 tokens)
**File:** `context.py`

Fetched from Go backend when session starts:
- User name, email
- Current workspace
- Current active node
- Recent nodes accessed
- Recent activity

**Endpoint:** `GET /api/voice/user-context/:user_id`

### Level 2-3: Node Context (On-Demand via Tools)
**File:** `tools.py`

Six tools available for LLM to call:

1. **get_node_context** - Full node details (identity, relationships, state, focus)
2. **get_node_children** - Child nodes list
3. **search_nodes** - Find nodes by query/type
4. **get_project_tasks** - Project task list
5. **get_recent_activity** - Activity log
6. **get_node_decisions** - Decision queue & history

Each tool calls Go backend endpoint to fetch real data.

---

## 🔄 Request Flow

### Session Start
```
1. Frontend requests LiveKit token from Go backend
   POST /api/livekit/token

2. Go backend generates token with user identity
   Returns: { token, room_name, identity, url }

3. Frontend connects to LiveKit room via WebRTC

4. LiveKit assigns Python voice agent to room

5. Python agent extracts user_id from participant identity

6. Python agent fetches user context (Level 1)
   GET /api/voice/user-context/:user_id

7. Python agent builds instructions (Level 0 + Level 1)

8. Voice agent starts with personalized greeting
```

### During Conversation
```
User: "What's the status of the HBAI project?"

1. STT (GROQ Whisper): Transcribes speech to text

2. LLM (GROQ Llama 3.1 8B):
   - Receives: Instructions + User query
   - Decides: Need node data → Tool call
   - Calls: get_node_context("HBAI Automation")

3. Tool Handler (tools.py):
   - Executes: GET /api/nodes/HBAI-Automation/context
   - Returns: Node data to LLM

4. LLM:
   - Analyzes node data
   - Generates: Brief response (5-15 words)
   - Returns: "HBAI is at 65%, yellow status. Blocked on client approval."

5. TTS (ElevenLabs): Converts text to speech

6. WebRTC: Streams audio back to frontend
```

---

## 🎯 Key Design Decisions

### Why NOT One File?
- **Separation of concerns**: Config, prompts, context, tools all separate
- **Maintainability**: Easy to update each layer independently
- **Testability**: Each module can be tested in isolation
- **Scalability**: Easy to add new tools without touching agent logic

### Why Hierarchical Context?
- **Token efficiency**: Only load what's needed (200-500 tokens vs 5000+)
- **Performance**: Faster LLM inference with smaller context
- **Accuracy**: LLM fetches real data instead of guessing
- **Flexibility**: Easy to expand with new node types and tools

### Why Tool Calling?
- **Real-time data**: Always fetch fresh data from DB
- **No hallucination**: LLM can't invent data
- **Scalable**: Add new tools without changing agent logic
- **User control**: User's workspace data, not cached/stale info

---

## 🔧 Technology Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| **STT** | GROQ Whisper large-v3 | Fast transcription (~300ms) |
| **LLM** | GROQ Llama 3.1 8B Instant | Fast inference + tool calling support |
| **TTS** | ElevenLabs Turbo v2.5 | Low latency, natural voice |
| **VAD** | Silero | Accurate voice activity detection |
| **Transport** | LiveKit WebRTC | Real-time bidirectional audio streaming |
| **HTTP** | httpx (async) | Non-blocking calls to Go backend |

---

## 🚀 How to Test Locally

### 1. Set up environment
```bash
cd python-voice-agent
source venv/bin/activate
source setup_env.sh  # Sets all required env vars
```

### 2. Start Go backend (Terminal 1)
```bash
cd desktop/backend-go
go run ./cmd/server
# Should see: "Voice system: Python LiveKit agents..."
```

### 3. Start Python agent (Terminal 2)
```bash
cd python-voice-agent
python agent.py dev
# Should see: "[Agent] Prewarming models..."
```

### 4. Start frontend (Terminal 3)
```bash
cd frontend
npm run dev
# Open: http://localhost:5173/window
```

### 5. Test voice orb
- Click voice orb in frontend
- Speak: "Hello OSA"
- Expected: Personalized greeting within 1 second

---

## 📋 Go Backend Endpoints (NEW)

The Python agent expects these Go backend endpoints:

### User Context (EXISTS)
```
GET /api/voice/user-context/:user_id
Returns: { name, email, workspace, current_node, recent_nodes, recent_activity }
```

### Node Tools (NEED TO IMPLEMENT)
```
GET /api/nodes/:node_id/context
GET /api/nodes/:node_id/children
GET /api/nodes/search?q=query&type=TYPE
GET /api/projects/:project_id/tasks
GET /api/nodes/:node_id/activity?limit=5
GET /api/nodes/:node_id/decisions
```

**Status:** User context endpoint exists. Node tool endpoints need implementation.

---

## 🎨 Example Tool Call Flow

```python
# User asks: "What are the tasks for the HBAI project?"

# 1. LLM decides to call tool
tool_call = {
    "name": "get_project_tasks",
    "arguments": {"project_id": "HBAI-Automation"}
}

# 2. tools.py routes to handler
result = await handle_get_project_tasks({"project_id": "HBAI-Automation"})

# 3. Handler calls Go backend
GET /api/projects/HBAI-Automation/tasks

# 4. Go backend returns real data
{
    "tasks": [
        {"id": "T-1", "name": "Setup automation flow", "status": "Done"},
        {"id": "T-2", "name": "Test with client data", "status": "In Progress"},
        {"id": "T-3", "name": "Deploy to prod", "status": "Blocked"}
    ]
}

# 5. LLM receives data and responds
"You've got 3 tasks. One done, one in progress, one blocked on deploy."
```

---

## 🐛 Debugging

### Check imports work:
```bash
python -c "from agent import *; from config import config; print('✅ OK')"
```

### Check configuration:
```bash
python -c "from config import config; config.validate(); print(config)"
```

### Test user context fetch:
```bash
python -c "
import asyncio
from context import fetch_user_context
result = asyncio.run(fetch_user_context('test-user-id'))
print(result)
"
```

### Check tool definitions:
```bash
python -c "from tools import TOOL_DEFINITIONS; print(f'✅ {len(TOOL_DEFINITIONS)} tools loaded')"
```

---

## ✅ What's Complete

- ✅ Proper multi-file structure (not one file)
- ✅ Config management with validation
- ✅ OSA personality prompts (Level 0)
- ✅ User context fetching (Level 1)
- ✅ Tool definitions (6 tools for Level 2-3)
- ✅ Tool handlers with Go backend integration
- ✅ Hierarchical context architecture
- ✅ Environment setup script
- ✅ Go backend compiles
- ✅ Imports working

## ⏳ What's Pending

- ⏳ Go backend tool endpoints (nodes, projects, decisions)
- ⏳ LLM tool calling integration (LiveKit agents API)
- ⏳ Local testing with frontend
- ⏳ Deployment to Cloud Run
- ⏳ End-to-end verification

---

**Architecture designed:** January 17, 2026
**Based on:** CONTEXT_HIERARCHY_PLAN.md + voice.go prompts
**Status:** ✅ READY FOR BACKEND TOOL IMPLEMENTATION
