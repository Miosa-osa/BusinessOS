# Voice & LiveKit Code Audit Report

**Date:** January 17, 2026  
**Purpose:** Identify what voice/LiveKit code exists and what we don't need

---

## 📊 EXECUTIVE SUMMARY

### ✅ What We NEED (Current Voice System)

**Go Backend:**
- ✅ `internal/handlers/livekit.go` - LiveKit token generation & agent dispatch
- ✅ `internal/handlers/voice_agent.go` - User context endpoint for Python agent
- ✅ Routes in `handlers.go`:
  - `/api/livekit/token` - Generate LiveKit room tokens
  - `/api/livekit/rooms` - Room status info
  - `/api/voice/user-context/:user_id` - User context for agent

**Python Agent:**
- ✅ `python-voice-agent/agent.py` - Main voice agent with LiveKit
- ✅ `python-voice-agent/config.py` - Configuration management
- ✅ `python-voice-agent/context.py` - User context fetching
- ✅ `python-voice-agent/prompts.py` - OSA personality & instructions
- ✅ `python-voice-agent/tools.py` - 8 tools (data + UI control)

**Frontend:**
- ✅ `frontend/src/lib/services/simpleVoice.ts` - LiveKit connection
- ✅ `frontend/src/lib/components/desktop3d/VoiceOrbPanel.svelte` - Voice UI

---

### ❌ What We DON'T NEED (Separate Features)

**Voice Notes System** (KEEP - Different feature):
- `internal/handlers/voice_notes.go` - Voice memo recording/transcription
- `internal/services/whisper.go` - Local Whisper transcription
- `internal/services/elevenlabs.go` - TTS for non-agent features
- Routes: `/api/voice-notes/*`

**Note:** Voice notes is a SEPARATE feature from voice agent - users can record voice memos that get transcribed and stored. This is NOT related to the real-time voice conversation agent.

---

### 🗑️ What to DELETE (Already Deleted in Git)

**Deprecated Files (Pending Commit):**
- ❌ `desktop/backend-go/internal/handlers/osa_voice.go` - Old voice handler
- ❌ `desktop/backend-go/internal/handlers/realtime_transcription.go` - Old realtime
- ❌ `frontend/src/lib/components/desktop3d/VoiceControlPanel.svelte` - Old UI
- ❌ `frontend/src/lib/services/voiceTranscriptionService.ts` - Old service

**Status:** These files are deleted but not yet committed to git.

---

## 📂 DETAILED FILE BREAKDOWN

### Go Backend - LiveKit/Voice Agent System

#### File: `internal/handlers/livekit.go` (169 lines) ✅ KEEP
**Purpose:** LiveKit token generation & automatic agent dispatch

**What it does:**
- Generates JWT tokens for LiveKit rooms
- Creates rooms with 5min empty timeout
- Dispatches voice agent automatically via LiveKit API
- Handles `/api/livekit/token` and `/api/livekit/rooms`

**Dependencies:**
```go
import (
    "github.com/livekit/protocol/auth"
    "github.com/livekit/protocol/livekit"
    lksdk "github.com/livekit/server-sdk-go/v2"
)
```

**Key functions:**
1. `HandleLiveKitToken()` - Generates token + dispatches agent
2. `HandleLiveKitRooms()` - Returns room configuration status

**Used by:** Frontend `simpleVoice.ts` when user clicks cloud icon

---

#### File: `internal/handlers/voice_agent.go` (76 lines) ✅ KEEP
**Purpose:** Provides user context to Python voice agent

**What it does:**
- Endpoint: `/api/voice/user-context/:user_id`
- Fetches user info from database
- Returns name, email, workspace for agent personalization

**Called by:** `python-voice-agent/context.py` at session start

**Response format:**
```json
{
  "name": "Roberto",
  "email": "user@example.com",
  "workspace": "BusinessOS",
  "recent_activity": "Active today",
  "preferences": []
}
```

**Note:** No authentication required - internal service-to-service call

---

### Python Voice Agent Files

#### File: `python-voice-agent/agent.py` (173 lines) ✅ KEEP
**Purpose:** Main LiveKit voice agent entrypoint

**Architecture:**
```python
# Technology Stack
STT: Groq Whisper large-v3
LLM: Groq Llama 3.1 8B Instant  
TTS: ElevenLabs Turbo v2.5
VAD: Silero
Transport: LiveKit WebRTC
```

**Key components:**
- `prewarm_process()` - Preloads VAD model
- `entrypoint()` - Main agent logic
- `request_fnc()` - Accepts all job requests

**Flow:**
1. Wait for participant
2. Fetch user context from Go backend
3. Build instructions with OSA personality
4. Create agent with 8 tools
5. Start voice session (STT → LLM → TTS)
6. Greet user and listen

---

#### File: `python-voice-agent/config.py` ✅ KEEP
**Purpose:** Environment variable configuration

**Loaded vars:**
- LiveKit credentials (URL, API key, secret)
- Groq API key
- ElevenLabs API key & voice ID
- Go backend URL
- Model names

---

#### File: `python-voice-agent/context.py` ✅ KEEP
**Purpose:** Fetch user context from Go backend

**Functions:**
- `extract_user_id_from_identity()` - Parse user ID from LiveKit identity
- `fetch_user_context()` - HTTP call to `/api/voice/user-context/:user_id`

---

#### File: `python-voice-agent/prompts.py` ✅ KEEP
**Purpose:** OSA personality and instruction building

**Contains:**
- Level 0: Core OSA identity (personality, tone, capabilities)
- Level 1: User context integration
- Greeting generation

---

#### File: `python-voice-agent/tools.py` (12,412 bytes) ✅ KEEP
**Purpose:** 8 tools for data retrieval and UI control

**Tools:**
1. `get_node_context` - Full node details
2. `get_node_children` - Sub-nodes
3. `search_nodes` - Search by name/type
4. `get_project_tasks` - Project tasks
5. `get_recent_activity` - Recent updates
6. `get_node_decisions` - Pending decisions
7. `activate_node` - **UI control - switch to node**
8. `list_all_nodes` - List all nodes

**Each tool:**
- Async function using `httpx.AsyncClient`
- Calls Go backend REST API
- Returns JSON response

---

### Frontend Files

#### File: `frontend/src/lib/services/simpleVoice.ts` ✅ KEEP
**Purpose:** LiveKit connection management

**What it does:**
- Requests token from backend
- Connects to LiveKit room
- Manages microphone
- Handles connection states

**Key methods:**
- `connect()` - Request token, join room
- `disconnect()` - Leave room
- `toggleMicrophone()` - Enable/disable mic

---

#### File: `frontend/src/lib/components/desktop3d/VoiceOrbPanel.svelte` ✅ KEEP
**Purpose:** Cloud icon UI component

**Features:**
- Blue glow when listening
- Purple glow when agent speaking
- Click to connect/disconnect
- Used in both Window and 3D Desktop

---

## 🏗️ VOICE NOTES SYSTEM (SEPARATE FEATURE)

### ❓ Keep or Remove?

**These files are for a DIFFERENT feature (voice memos):**

#### File: `internal/handlers/voice_notes.go` (14,019 bytes)
**Purpose:** Record and transcribe voice memos (not real-time conversation)

**Features:**
- Upload audio files
- Transcribe with local Whisper
- Store transcriptions in database
- Generate embeddings for search

**Routes:**
- `POST /api/voice-notes` - Upload voice note
- `GET /api/voice-notes` - List voice notes
- `GET /api/voice-notes/:id` - Get specific note
- `DELETE /api/voice-notes/:id` - Delete note
- `POST /api/voice-notes/:id/retranscribe` - Re-transcribe

**Database table:** `voice_notes`

---

#### File: `internal/services/whisper.go` (5,997 bytes)
**Purpose:** Local Whisper STT for voice notes

**Used by:** `voice_notes.go` handler

**Note:** This is NOT used by the voice agent (agent uses Groq Whisper in Python)

---

#### File: `internal/services/elevenlabs.go` (9,530 bytes)
**Purpose:** TTS service for non-agent features

**Used by:** Potentially other features, NOT the voice agent

**Note:** Voice agent uses ElevenLabs directly in Python

---

### Recommendation:
**KEEP voice notes system** - It's a separate feature from the voice agent. Users can:
- Voice Agent: Real-time conversation with OSA
- Voice Notes: Record voice memos that get transcribed and saved

These are TWO different features that happen to both use "voice".

---

## 🧹 CLEANUP RECOMMENDATIONS

### 1. Commit Deletions
```bash
git add desktop/backend-go/internal/handlers/osa_voice.go
git add desktop/backend-go/internal/handlers/realtime_transcription.go
git add frontend/src/lib/components/desktop3d/VoiceControlPanel.svelte
git add frontend/src/lib/services/voiceTranscriptionService.ts
git commit -m "chore: remove deprecated voice handlers (replaced by LiveKit agent)"
```

### 2. Check for Unused Imports
Search for imports of deleted files:
```bash
grep -r "osa_voice" desktop/backend-go/
grep -r "realtime_transcription" desktop/backend-go/
grep -r "VoiceControlPanel" frontend/
grep -r "voiceTranscriptionService" frontend/
```

### 3. Update Comments
Check handlers.go for outdated comments referencing old system:
- Line 896: "NOTE: These routes are deprecated..."
- Line 904: Commented out old voice routes

**Action:** Remove commented-out code and outdated comments.

---

## 📋 FINAL INVENTORY

### ✅ ESSENTIAL - Voice Agent System (KEEP ALL)

**Go Backend (3 files):**
1. `internal/handlers/livekit.go` - Token + dispatch
2. `internal/handlers/voice_agent.go` - User context
3. Routes in `handlers.go` (lines 906-918)

**Python Agent (5 files):**
1. `agent.py` - Main entrypoint
2. `config.py` - Configuration
3. `context.py` - User context fetching
4. `prompts.py` - OSA personality
5. `tools.py` - 8 tools

**Frontend (2 files):**
1. `services/simpleVoice.ts` - LiveKit connection
2. `components/desktop3d/VoiceOrbPanel.svelte` - UI

**Total: 10 essential files**

---

### 📝 SEPARATE FEATURE - Voice Notes (KEEP)

**Go Backend (4 files):**
1. `internal/handlers/voice_notes.go` - Voice memo handler
2. `internal/services/whisper.go` - Local transcription
3. `internal/services/elevenlabs.go` - TTS service
4. `internal/handlers/transcription.go` - Transcription utilities

**Purpose:** Record and transcribe voice memos (different from agent)

---

### 🗑️ DEPRECATED - Delete These (COMMIT REMOVAL)

1. ❌ `internal/handlers/osa_voice.go` 
2. ❌ `internal/handlers/realtime_transcription.go`
3. ❌ `frontend/src/lib/components/desktop3d/VoiceControlPanel.svelte`
4. ❌ `frontend/src/lib/services/voiceTranscriptionService.ts`

**Action:** Commit these deletions to git

---

## 🎯 SUMMARY

### What We Have:
- ✅ **Voice Agent System** (10 files) - Real-time conversation with OSA
- ✅ **Voice Notes System** (4 files) - Record and transcribe voice memos
- ❌ **Deprecated Old Code** (4 files) - Pending deletion

### What We Need:
- ✅ Keep Voice Agent System (currently working)
- ✅ Keep Voice Notes System (separate feature)
- ❌ Delete old deprecated handlers

### Next Steps:
1. Commit deletion of 4 deprecated files
2. Remove commented-out code from handlers.go
3. Update any outdated comments
4. Verify no imports of deleted files

---

**Audit Completed:** January 17, 2026  
**Status:** All essential code identified ✅

---

## 📊 ARCHITECTURE DIAGRAM

### Current Voice Agent System (✅ KEEP)

```
┌─────────────────────────────────────────────────────────────────┐
│                        USER INTERACTION                          │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                    [Clicks Cloud Icon]
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ FRONTEND: VoiceOrbPanel.svelte + simpleVoice.ts                 │
│ - Requests token from backend                                   │
│ - Connects to LiveKit Cloud                                     │
│ - Manages microphone                                            │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                    [POST /api/livekit/token]
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ GO BACKEND: livekit.go                                          │
│ 1. Generate JWT token                                           │
│ 2. Create room (osa-voice-xxxxx)                                │
│ 3. Dispatch agent via LiveKit API ─────────────┐                │
└───────────────────────────┬────────────────────┘                │
                            │                                     │
                    [Returns token]                               │
                            │                                     │
                            ▼                                     │
┌─────────────────────────────────────────────────────────────────┐
│ LIVEKIT CLOUD                                                   │
│ - wss://macstudiosystems-yn61tekm.livekit.cloud                 │
│ - WebRTC transport                                              │
└───────────────────────────┬─────────────────────────────────────┘
                            │                                     │
                     [Agent Dispatch] ◄──────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ PYTHON VOICE AGENT: agent.py                                    │
│ 1. Receives job request                                         │
│ 2. Fetches user context ────────┐                               │
│ 3. Loads OSA personality        │                               │
│ 4. Starts voice session         │                               │
│    - STT: Groq Whisper          │                               │
│    - LLM: Groq Llama 3.1        │                               │
│    - TTS: ElevenLabs            │                               │
│    - VAD: Silero                │                               │
│ 5. Listens and responds         │                               │
└─────────────────────────────────┼───────────────────────────────┘
                                  │
                    [GET /api/voice/user-context/:user_id]
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│ GO BACKEND: voice_agent.go                                      │
│ - Fetches user from database                                    │
│ - Returns: name, email, workspace                               │
└─────────────────────────────────────────────────────────────────┘
```

---

### Voice Notes System (✅ KEEP - Different Feature)

```
┌─────────────────────────────────────────────────────────────────┐
│                      USER RECORDS VOICE MEMO                     │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                    [Uploads audio file]
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ GO BACKEND: voice_notes.go                                      │
│ 1. Receives audio file                                          │
│ 2. Saves to storage                                             │
│ 3. Calls WhisperService ────────┐                               │
│ 4. Stores transcript in DB      │                               │
│ 5. Generates embeddings         │                               │
└─────────────────────────────────┼───────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│ LOCAL WHISPER SERVICE: whisper.go                               │
│ - Runs Whisper locally                                          │
│ - Returns transcript                                            │
└─────────────────────────────────────────────────────────────────┘
```

**Note:** Voice notes and voice agent are SEPARATE features:
- **Voice Agent:** Real-time conversation
- **Voice Notes:** Record memos for later transcription

---

### Deprecated Old System (🗑️ DELETE)

```
┌─────────────────────────────────────────────────────────────────┐
│ OLD SYSTEM (DELETE THESE FILES)                                 │
├─────────────────────────────────────────────────────────────────┤
│ ❌ osa_voice.go - Old voice handler (replaced by livekit.go)    │
│ ❌ realtime_transcription.go - Old realtime (use LiveKit now)   │
│ ❌ VoiceControlPanel.svelte - Old UI (use VoiceOrbPanel.svelte) │
│ ❌ voiceTranscriptionService.ts - Old service (use simpleVoice) │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔍 DEPENDENCY ANALYSIS

### Voice Agent Dependencies

**Go Backend:**
```
github.com/livekit/protocol/auth        ← Token generation
github.com/livekit/protocol/livekit     ← Room creation
github.com/livekit/server-sdk-go/v2     ← Agent dispatch
```

**Python Agent:**
```
livekit-agents         ← LiveKit integration
livekit-plugins-groq   ← STT + LLM
livekit-plugins-elevenlabs  ← TTS
livekit-plugins-silero ← VAD
httpx                  ← HTTP client for tools
```

**Frontend:**
```
livekit-client         ← LiveKit connection
```

---

### Voice Notes Dependencies

**Go Backend:**
```
github.com/ggerganov/whisper.cpp/bindings/go  ← Local Whisper
(Custom ElevenLabs HTTP client)                ← TTS
```

**Note:** Voice notes uses LOCAL Whisper, not cloud API

---

## 📈 FILE SIZE BREAKDOWN

### Voice Agent System
```
Go Backend:
  livekit.go         169 lines   4,570 bytes   ✅
  voice_agent.go      76 lines   2,049 bytes   ✅
  
Python Agent:
  agent.py           173 lines   5,569 bytes   ✅
  config.py           ~80 lines   2,291 bytes   ✅
  context.py          ~90 lines   2,919 bytes   ✅
  prompts.py         ~350 lines  16,078 bytes   ✅
  tools.py           ~320 lines  12,412 bytes   ✅

Frontend:
  simpleVoice.ts     ~150 lines  ~5,000 bytes   ✅
  VoiceOrbPanel.svelte ~200 lines ~7,000 bytes  ✅

TOTAL: ~1,600 lines, ~58KB
```

### Voice Notes System
```
Go Backend:
  voice_notes.go     ~400 lines  14,019 bytes   📝
  whisper.go         ~180 lines   5,997 bytes   📝
  elevenlabs.go      ~270 lines   9,530 bytes   📝

TOTAL: ~850 lines, ~30KB
```

### Deprecated (To Delete)
```
  osa_voice.go       ???         ???            ❌
  realtime_trans...  ???         ???            ❌
  VoiceControlP...   ???         ???            ❌
  voiceTranscri...   ???         ???            ❌
```

---

## ✅ FINAL RECOMMENDATIONS

### 1. Commit Deletions NOW
```bash
# These files are already deleted, just commit
git add -u
git commit -m "chore: remove deprecated voice handlers

- Removed osa_voice.go (replaced by livekit.go)
- Removed realtime_transcription.go (use LiveKit agent)
- Removed VoiceControlPanel.svelte (use VoiceOrbPanel.svelte)
- Removed voiceTranscriptionService.ts (use simpleVoice.ts)

Voice system now uses:
- Go: livekit.go + voice_agent.go
- Python: LiveKit agents framework
- Frontend: simpleVoice.ts + VoiceOrbPanel.svelte"
```

### 2. Clean Up handlers.go
Remove commented-out code around lines 896-904:
```go
// ⚡ Removed deprecated realtime endpoint - LiveKit handles all voice now
// DEPRECATED routes (lines 896-904) - DELETE THESE COMMENTS
```

### 3. No Other Changes Needed
**Everything else is essential and should be kept.**

---

**Audit Complete** ✅  
**Files to Delete:** 4  
**Files to Keep:** 17 (10 voice agent + 4 voice notes + 3 routes)

