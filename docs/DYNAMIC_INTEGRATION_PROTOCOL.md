# Dynamic Integration Protocol (DIP)

## OSA Flow Engine - Universal AI-Powered Integration Framework

---

## Executive Summary

The **Dynamic Integration Protocol (DIP)** is BusinessOS's next-generation approach to universal integrations that solves the fundamental limitations of MCP (Model Context Protocol):

| Problem with MCP | DIP Solution |
|------------------|--------------|
| Tool definitions consume context | Only send minimal API patterns when needed |
| Pre-built tools for every integration | LLM generates scripts on-demand for ANY API |
| Credentials often exposed to LLM | Credentials NEVER leave the user's machine |
| Limited to what developers build | Truly universal - any API, any workflow |
| Can't adapt to new APIs | LLM learns patterns and generates custom code |

**Core Concept:** Instead of pre-defining every tool, teach the LLM the patterns of how APIs work, let it generate Python scripts dynamically, and execute them locally on the user's machine where credentials are securely stored.

---

## Architecture Overview

```
                                  CLOUD (BusinessOS Backend - Go)
    ┌──────────────────────────────────────────────────────────────────────────┐
    │                                                                          │
    │   ┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐  │
    │   │  Orchestrator   │────▶│  LLM (Claude)    │────▶│  Script         │  │
    │   │  Service        │◀────│  Generation      │◀────│  Validator      │  │
    │   └────────┬────────┘     └──────────────────┘     └────────┬────────┘  │
    │            │                                                  │          │
    │            │  API Pattern Templates                          │          │
    │            │  (minimal context)                              │          │
    │            ▼                                                  ▼          │
    │   ┌──────────────────────────────────────────────────────────────────┐  │
    │   │                    Script Distribution Service                    │  │
    │   │               (sends validated scripts to user)                  │  │
    │   └───────────────────────────────┬──────────────────────────────────┘  │
    │                                   │                                      │
    └───────────────────────────────────┼──────────────────────────────────────┘
                                        │
                                        │ WebSocket / SSE
                                        ▼
    ┌──────────────────────────────────────────────────────────────────────────┐
    │                     USER'S MACHINE (Desktop App / MIOSA OS)              │
    │                                                                          │
    │   ┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐  │
    │   │  DIP Engine     │◀───▶│  Python Runtime  │◀───▶│  Credential     │  │
    │   │  (Sidecar)      │     │  (Sandboxed)     │     │  Vault          │  │
    │   └────────┬────────┘     └────────┬─────────┘     └─────────────────┘  │
    │            │                       │                                     │
    │            │                       │  Execute scripts                    │
    │            │                       │  with local credentials             │
    │            ▼                       ▼                                     │
    │   ┌──────────────────────────────────────────────────────────────────┐  │
    │   │                        Generated Scripts                          │  │
    │   │                   ~/businessos/flows/scripts/                     │  │
    │   └───────────────────────────────┬──────────────────────────────────┘  │
    │                                   │                                      │
    │                                   │  Results (structured JSON)           │
    │                                   ▼                                      │
    │   ┌──────────────────────────────────────────────────────────────────┐  │
    │   │              Response Bridge (back to orchestrator)               │  │
    │   └──────────────────────────────────────────────────────────────────┘  │
    │                                                                          │
    └──────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Credential Vault (Local)

Secure, encrypted storage on the user's machine. **NEVER** sent to the LLM.

```python
# Location: ~/businessos/vault/credentials.enc (encrypted)
# Or: MIOSA OS uses system keyring

credentials = {
    "google": {
        "client_id": "...",
        "client_secret": "...",
        "access_token": "...",
        "refresh_token": "...",
        "scopes": ["gmail.readonly", "drive.readonly"]
    },
    "slack": {
        "bot_token": "xoxb-...",
        "user_token": "xoxp-..."
    },
    "hubspot": {
        "api_key": "...",
        "access_token": "..."
    },
    "custom_api": {
        "base_url": "https://api.example.com",
        "api_key": "...",
        "headers": {"X-Custom-Header": "value"}
    }
}
```

**Security Features:**
- AES-256 encryption at rest
- Unlocked with user's master password or biometrics
- MIOSA OS integrates with system keyring
- Never transmitted to cloud - scripts run locally

---

### 2. API Pattern Templates

Minimal context patterns that teach the LLM how to structure scripts. **NOT full tool definitions** - just patterns.

```python
# patterns/rest_api.py - Pattern template for REST APIs

"""
REST API Pattern for DIP

The LLM receives this pattern to understand how to generate scripts
for any REST API. This is much smaller than full MCP tool definitions.
"""

PATTERN = """
## REST API Integration Pattern

### Authentication Types:
1. API Key in header: headers["Authorization"] = f"Bearer {api_key}"
2. API Key in query: params["api_key"] = api_key
3. OAuth2 Bearer: headers["Authorization"] = f"Bearer {access_token}"
4. Basic Auth: auth=(username, password)

### Request Structure:
```python
import requests
from dip_engine import get_credential, return_result

def execute():
    # Get credentials from vault (NEVER hardcode)
    creds = get_credential("provider_name")

    # Make request
    response = requests.get(
        url=f"{creds['base_url']}/endpoint",
        headers={"Authorization": f"Bearer {creds['access_token']}"},
        params={"query": "value"}
    )

    # Return structured result
    return_result({
        "success": response.ok,
        "data": response.json(),
        "status_code": response.status_code
    })
```

### Error Handling:
- Check response.ok before processing
- Handle rate limits (429) with exponential backoff
- Refresh OAuth tokens if 401 returned

### Common Endpoints by Category:
- List: GET /resource
- Get: GET /resource/{id}
- Create: POST /resource
- Update: PUT /resource/{id} or PATCH /resource/{id}
- Delete: DELETE /resource/{id}
- Search: GET /resource?query=...
"""
```

```python
# patterns/oauth_flow.py - Pattern for OAuth APIs

PATTERN = """
## OAuth 2.0 Integration Pattern

### Token Refresh:
```python
def refresh_token(refresh_token, client_id, client_secret, token_url):
    response = requests.post(token_url, data={
        "grant_type": "refresh_token",
        "refresh_token": refresh_token,
        "client_id": client_id,
        "client_secret": client_secret
    })
    return response.json()
```

### With Token Refresh Wrapper:
```python
from dip_engine import get_credential, update_credential, return_result

def execute():
    creds = get_credential("provider_name")

    response = make_request(creds)

    if response.status_code == 401:
        # Token expired, refresh
        new_tokens = refresh_token(...)
        update_credential("provider_name", new_tokens)
        response = make_request(new_tokens)

    return_result({"data": response.json()})
```
"""
```

---

### 3. Provider-Specific Patterns

For known providers, we give the LLM specific API structure knowledge:

```python
# patterns/providers/google.py

GOOGLE_PATTERN = """
## Google API Pattern

### Base URLs:
- Gmail: https://gmail.googleapis.com/gmail/v1
- Drive: https://www.googleapis.com/drive/v3
- Calendar: https://www.googleapis.com/calendar/v3
- Sheets: https://sheets.googleapis.com/v4

### Common Headers:
Authorization: Bearer {access_token}

### Gmail Examples:
- List messages: GET /users/me/messages?q={query}
- Get message: GET /users/me/messages/{id}
- Get thread: GET /users/me/threads/{id}
- Labels: GET /users/me/labels

### Drive Examples:
- List files: GET /files?q={query}
- Get file: GET /files/{id}
- Get content: GET /files/{id}?alt=media
- Search: GET /files?q=name contains '{name}'

### Calendar Examples:
- List events: GET /calendars/{calendarId}/events
- Create event: POST /calendars/{calendarId}/events
- Get event: GET /calendars/{calendarId}/events/{eventId}
"""
```

```python
# patterns/providers/slack.py

SLACK_PATTERN = """
## Slack API Pattern

### Base URL: https://slack.com/api

### Authentication:
Authorization: Bearer {bot_token} or {user_token}

### Common Endpoints:
- List channels: POST /conversations.list
- Send message: POST /chat.postMessage {channel, text}
- Get history: POST /conversations.history {channel, limit}
- Search: POST /search.messages {query}
- List users: POST /users.list

### Important Notes:
- Most endpoints use POST with JSON body
- Pagination uses cursor-based pagination
- Rate limits: Tier 1 (1+ per minute) to Tier 4 (100+ per minute)
"""
```

---

### 4. DIP Engine (Python Sidecar)

The local Python service that runs on the user's machine:

```python
# dip_engine/engine.py

"""
DIP Engine - Dynamic Integration Protocol Executor

This runs as a sidecar service on the user's machine (desktop app or MIOSA OS).
It receives scripts from the cloud, executes them locally with credentials,
and returns structured results.
"""

import asyncio
import json
import os
import subprocess
import sys
import tempfile
import hashlib
from pathlib import Path
from typing import Any, Dict, Optional
import websockets
from cryptography.fernet import Fernet

class DIPEngine:
    def __init__(self, config_path: str = "~/.businessos/dip"):
        self.config_path = Path(config_path).expanduser()
        self.scripts_dir = self.config_path / "scripts"
        self.vault_path = self.config_path / "vault" / "credentials.enc"
        self.scripts_dir.mkdir(parents=True, exist_ok=True)

        # Load encryption key from secure storage
        self.cipher = self._load_cipher()

        # Script execution sandbox
        self.venv_path = self.config_path / "venv"
        self._ensure_venv()

    def _load_cipher(self) -> Fernet:
        """Load encryption key from system keyring or secure storage."""
        # In MIOSA OS, this uses the system keyring
        # In desktop app, uses encrypted file unlocked by master password
        key_path = self.config_path / "vault" / ".key"
        if key_path.exists():
            with open(key_path, "rb") as f:
                return Fernet(f.read())
        else:
            # First run - generate key
            key = Fernet.generate_key()
            key_path.parent.mkdir(parents=True, exist_ok=True)
            with open(key_path, "wb") as f:
                f.write(key)
            os.chmod(key_path, 0o600)
            return Fernet(key)

    def _ensure_venv(self):
        """Ensure Python virtual environment exists for sandboxed execution."""
        if not self.venv_path.exists():
            subprocess.run([sys.executable, "-m", "venv", str(self.venv_path)])
            # Install base dependencies
            pip = self.venv_path / "bin" / "pip"
            subprocess.run([str(pip), "install", "requests", "aiohttp", "google-auth"])

    # ─────────────────────────────────────────────────────────────────
    # CREDENTIAL MANAGEMENT
    # ─────────────────────────────────────────────────────────────────

    def get_credential(self, provider: str) -> Dict[str, Any]:
        """Get credentials for a provider from the encrypted vault."""
        if not self.vault_path.exists():
            raise ValueError(f"No credentials found for {provider}")

        with open(self.vault_path, "rb") as f:
            encrypted_data = f.read()

        decrypted = self.cipher.decrypt(encrypted_data)
        all_creds = json.loads(decrypted.decode())

        if provider not in all_creds:
            raise ValueError(f"No credentials found for {provider}")

        return all_creds[provider]

    def store_credential(self, provider: str, credentials: Dict[str, Any]):
        """Store credentials in the encrypted vault."""
        # Load existing
        all_creds = {}
        if self.vault_path.exists():
            with open(self.vault_path, "rb") as f:
                decrypted = self.cipher.decrypt(f.read())
                all_creds = json.loads(decrypted.decode())

        # Update
        all_creds[provider] = credentials

        # Save
        encrypted = self.cipher.encrypt(json.dumps(all_creds).encode())
        self.vault_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.vault_path, "wb") as f:
            f.write(encrypted)
        os.chmod(self.vault_path, 0o600)

    def update_credential(self, provider: str, updates: Dict[str, Any]):
        """Update specific fields in a provider's credentials."""
        creds = self.get_credential(provider)
        creds.update(updates)
        self.store_credential(provider, creds)

    # ─────────────────────────────────────────────────────────────────
    # SCRIPT EXECUTION
    # ─────────────────────────────────────────────────────────────────

    async def execute_script(
        self,
        script_content: str,
        script_id: str,
        timeout: int = 30
    ) -> Dict[str, Any]:
        """
        Execute a generated script in the sandboxed environment.

        The script has access to:
        - get_credential(provider) - Get credentials from vault
        - update_credential(provider, updates) - Update credentials
        - return_result(data) - Return structured data to orchestrator
        """

        # Create wrapper that provides the DIP API
        wrapper = f'''
import json
import sys
sys.path.insert(0, "{self.config_path}")
from dip_runtime import get_credential, update_credential, return_result

# ─── GENERATED SCRIPT ───
{script_content}
# ─── END GENERATED SCRIPT ───

if __name__ == "__main__":
    try:
        execute()
    except Exception as e:
        return_result({{"error": str(e), "success": False}})
'''

        # Save script
        script_path = self.scripts_dir / f"{script_id}.py"
        with open(script_path, "w") as f:
            f.write(wrapper)

        try:
            # Execute in venv with timeout
            python = self.venv_path / "bin" / "python"
            process = await asyncio.create_subprocess_exec(
                str(python),
                str(script_path),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                env={
                    **os.environ,
                    "DIP_SCRIPT_ID": script_id,
                    "DIP_CONFIG_PATH": str(self.config_path)
                }
            )

            stdout, stderr = await asyncio.wait_for(
                process.communicate(),
                timeout=timeout
            )

            if process.returncode == 0:
                # Parse result from stdout
                result = json.loads(stdout.decode())
                return {"success": True, "data": result}
            else:
                return {
                    "success": False,
                    "error": stderr.decode(),
                    "exit_code": process.returncode
                }

        except asyncio.TimeoutError:
            process.kill()
            return {"success": False, "error": "Script execution timed out"}

        except Exception as e:
            return {"success": False, "error": str(e)}

        finally:
            # Cleanup script file (optional - keep for debugging)
            # script_path.unlink(missing_ok=True)
            pass

    # ─────────────────────────────────────────────────────────────────
    # WEBSOCKET CONNECTION TO CLOUD
    # ─────────────────────────────────────────────────────────────────

    async def connect_to_orchestrator(self, url: str, user_token: str):
        """
        Maintain persistent connection to the cloud orchestrator.
        Receives scripts to execute and returns results.
        """
        async with websockets.connect(
            url,
            extra_headers={"Authorization": f"Bearer {user_token}"}
        ) as ws:
            print(f"Connected to orchestrator at {url}")

            async for message in ws:
                request = json.loads(message)

                if request["type"] == "execute_script":
                    result = await self.execute_script(
                        script_content=request["script"],
                        script_id=request["script_id"],
                        timeout=request.get("timeout", 30)
                    )

                    await ws.send(json.dumps({
                        "type": "script_result",
                        "script_id": request["script_id"],
                        "result": result
                    }))

                elif request["type"] == "store_credential":
                    self.store_credential(
                        request["provider"],
                        request["credentials"]
                    )
                    await ws.send(json.dumps({
                        "type": "credential_stored",
                        "provider": request["provider"]
                    }))

                elif request["type"] == "ping":
                    await ws.send(json.dumps({"type": "pong"}))

# ─────────────────────────────────────────────────────────────────
# DIP RUNTIME (injected into scripts)
# ─────────────────────────────────────────────────────────────────

# dip_runtime.py - This file is in the config path

'''
# dip_runtime.py
import json
import os
from pathlib import Path
from cryptography.fernet import Fernet

_config_path = Path(os.environ.get("DIP_CONFIG_PATH", "~/.businessos/dip")).expanduser()
_result = None

def _get_cipher():
    key_path = _config_path / "vault" / ".key"
    with open(key_path, "rb") as f:
        return Fernet(f.read())

def get_credential(provider: str) -> dict:
    """Get credentials for a provider from the local vault."""
    vault_path = _config_path / "vault" / "credentials.enc"
    cipher = _get_cipher()

    with open(vault_path, "rb") as f:
        decrypted = cipher.decrypt(f.read())

    all_creds = json.loads(decrypted.decode())
    if provider not in all_creds:
        raise ValueError(f"No credentials for {provider}. Please connect first.")

    return all_creds[provider]

def update_credential(provider: str, updates: dict):
    """Update credentials in the local vault (e.g., refresh tokens)."""
    vault_path = _config_path / "vault" / "credentials.enc"
    cipher = _get_cipher()

    with open(vault_path, "rb") as f:
        decrypted = cipher.decrypt(f.read())

    all_creds = json.loads(decrypted.decode())
    if provider not in all_creds:
        all_creds[provider] = {}
    all_creds[provider].update(updates)

    encrypted = cipher.encrypt(json.dumps(all_creds).encode())
    with open(vault_path, "wb") as f:
        f.write(encrypted)

def return_result(data: dict):
    """Return structured data to the orchestrator."""
    print(json.dumps(data))
'''
```

---

### 5. Script Generation Flow

How the LLM generates scripts when a user makes a request:

```
User: "Check my Gmail for emails from John about the project proposal"

┌────────────────────────────────────────────────────────────────────┐
│ 1. ORCHESTRATOR receives request                                   │
└────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────────┐
│ 2. ORCHESTRATOR checks: Does user have Gmail credentials stored?   │
│    - Query DIP Engine: "What providers are connected?"             │
│    - Response: ["google", "slack", "hubspot"]                      │
│    - Google connected = Gmail available                            │
└────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────────┐
│ 3. ORCHESTRATOR sends to LLM with MINIMAL context:                 │
│                                                                    │
│    System: "Generate a DIP script to search Gmail.                 │
│             Use pattern: {gmail_pattern}                           │
│             User has 'google' credentials configured.              │
│             Script must use get_credential('google') for auth."    │
│                                                                    │
│    User: "Find emails from John about project proposal"            │
└────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────────┐
│ 4. LLM generates script:                                          │
│                                                                    │
│    ```python                                                       │
│    import requests                                                 │
│    from dip_engine import get_credential, return_result           │
│                                                                    │
│    def execute():                                                  │
│        creds = get_credential("google")                           │
│        headers = {"Authorization": f"Bearer {creds['access_token']}"}│
│                                                                    │
│        # Search Gmail                                              │
│        query = "from:john subject:project proposal"                │
│        response = requests.get(                                    │
│            "https://gmail.googleapis.com/gmail/v1/users/me/messages",│
│            headers=headers,                                        │
│            params={"q": query, "maxResults": 10}                  │
│        )                                                           │
│                                                                    │
│        if response.status_code == 401:                            │
│            # Token expired - would trigger refresh                 │
│            return_result({"error": "token_expired", "refresh": True})│
│            return                                                  │
│                                                                    │
│        messages = response.json().get("messages", [])             │
│                                                                    │
│        # Get details for each message                              │
│        results = []                                                │
│        for msg in messages[:5]:                                    │
│            detail = requests.get(                                  │
│                f"https://gmail.googleapis.com/gmail/v1/users/me/messages/{msg['id']}",│
│                headers=headers,                                    │
│                params={"format": "metadata"}                       │
│            ).json()                                                │
│                                                                    │
│            # Extract relevant headers                              │
│            headers_dict = {h["name"]: h["value"] for h in detail.get("payload", {}).get("headers", [])}│
│            results.append({                                        │
│                "id": msg["id"],                                    │
│                "subject": headers_dict.get("Subject", ""),        │
│                "from": headers_dict.get("From", ""),              │
│                "date": headers_dict.get("Date", ""),              │
│                "snippet": detail.get("snippet", "")               │
│            })                                                      │
│                                                                    │
│        return_result({                                             │
│            "success": True,                                        │
│            "count": len(results),                                  │
│            "emails": results                                       │
│        })                                                          │
│    ```                                                             │
└────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────────┐
│ 5. VALIDATOR checks script:                                        │
│    - No hardcoded credentials                                      │
│    - Uses get_credential() properly                                │
│    - Has proper error handling                                     │
│    - Returns structured data via return_result()                   │
│    - No dangerous operations (file system access, etc.)            │
└────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────────┐
│ 6. SCRIPT sent to user's DIP Engine via WebSocket                  │
│    - Script executes LOCALLY on user's machine                     │
│    - Credentials fetched from LOCAL vault                          │
│    - API calls made from user's IP                                 │
│    - Results returned to orchestrator                              │
└────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────────┐
│ 7. RESULTS returned to user:                                       │
│                                                                    │
│    "I found 3 emails from John about the project proposal:         │
│     1. 'Re: Project Proposal Draft' - Dec 28                       │
│     2. 'Project Proposal Updates' - Dec 25                         │
│     3. 'Initial Project Proposal' - Dec 20                         │
│                                                                    │
│     Would you like me to read any of these?"                       │
└────────────────────────────────────────────────────────────────────┘
```

---

### 6. Agentic Workflows

Scripts can trigger other agents and wait for responses:

```python
# Example: Multi-step agentic workflow
# User: "Find client emails, summarize them, and create tasks in ClickUp"

def execute():
    # Step 1: Get emails (this script)
    creds = get_credential("google")
    emails = fetch_emails_about_clients(creds)

    # Step 2: Request AI summarization (calls back to orchestrator)
    from dip_engine import request_agent

    summary = request_agent(
        agent="summarizer",
        task="Summarize these emails and extract action items",
        data={"emails": emails}
    )

    # Step 3: Create tasks in ClickUp
    clickup_creds = get_credential("clickup")

    for action_item in summary["action_items"]:
        create_clickup_task(clickup_creds, action_item)

    return_result({
        "emails_processed": len(emails),
        "summary": summary["summary"],
        "tasks_created": len(summary["action_items"])
    })
```

---

### 7. Security Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                        SECURITY LAYERS                             │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  LAYER 1: Credential Isolation                                     │
│  ─────────────────────────────                                     │
│  - Credentials NEVER leave user's machine                          │
│  - AES-256 encryption at rest                                      │
│  - Decryption only at execution time                              │
│  - LLM never sees credentials                                      │
│                                                                    │
│  LAYER 2: Script Validation                                        │
│  ───────────────────────────                                       │
│  - AST analysis before execution                                   │
│  - No file system access outside sandbox                           │
│  - No network calls except approved APIs                           │
│  - No shell/subprocess execution                                   │
│  - No imports except whitelisted modules                          │
│                                                                    │
│  LAYER 3: Sandboxed Execution                                      │
│  ──────────────────────────────                                    │
│  - Scripts run in isolated venv                                    │
│  - Resource limits (CPU, memory, time)                            │
│  - No access to other user processes                              │
│  - Containerized in MIOSA OS                                       │
│                                                                    │
│  LAYER 4: Result Sanitization                                      │
│  ─────────────────────────────                                     │
│  - Results stripped of sensitive data before LLM                  │
│  - PII detection and redaction                                    │
│  - Size limits on returned data                                   │
│                                                                    │
│  LAYER 5: Audit Logging                                            │
│  ───────────────────────────                                       │
│  - All script executions logged                                   │
│  - All API calls logged                                           │
│  - Anomaly detection                                              │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## Comparison: DIP vs MCP vs Traditional

| Feature | MCP | Traditional Integration | DIP |
|---------|-----|------------------------|-----|
| Context Usage | HIGH (tool definitions) | N/A | LOW (patterns only) |
| Flexibility | Limited to pre-built tools | Fixed endpoints | Unlimited - any API |
| Credential Security | Often in context | Backend only | Local vault only |
| New API Support | Requires dev work | Requires dev work | LLM generates on-demand |
| Agentic Workflows | Limited | None | Full support |
| Offline Capable | No | Partial | Yes (local execution) |
| User Data Location | Cloud | Cloud | User's machine |
| Real-time Adaptation | No | No | Yes |

---

## Implementation File Structure

```
backend-go/
├── internal/
│   └── dip/
│       ├── types.go                    # Core DIP types
│       ├── orchestrator.go             # Script generation orchestration
│       ├── validator.go                # Script validation (AST analysis)
│       ├── distributor.go              # Script distribution to clients
│       ├── patterns/
│       │   ├── loader.go               # Pattern loading
│       │   ├── rest.go                 # REST API pattern
│       │   ├── oauth.go                # OAuth pattern
│       │   └── providers/
│       │       ├── google.go           # Google-specific patterns
│       │       ├── slack.go            # Slack-specific patterns
│       │       ├── hubspot.go          # HubSpot-specific patterns
│       │       └── ...
│       └── websocket/
│           ├── handler.go              # WebSocket connection handler
│           └── messages.go             # Message types

dip-engine/                             # Python package for local execution
├── pyproject.toml
├── dip_engine/
│   ├── __init__.py
│   ├── engine.py                       # Main DIP engine
│   ├── vault.py                        # Credential vault
│   ├── runtime.py                      # Script runtime
│   ├── sandbox.py                      # Execution sandbox
│   └── agents.py                       # Agent communication
└── tests/
    └── ...

desktop-app/                            # Electron/Tauri wrapper
├── src/
│   ├── dip/
│   │   ├── manager.ts                  # DIP engine management
│   │   ├── connection.ts               # WebSocket to cloud
│   │   └── credentials.ts              # Credential UI
│   └── ...

miosa-os/                               # Linux distro integration
├── dip/
│   ├── systemd/
│   │   └── dip-engine.service          # Systemd service
│   ├── keyring/
│   │   └── integration.py              # System keyring integration
│   └── ...
```

---

## Implementation Priority

### Phase 1: Core DIP Engine (CRITICAL)

1. `dip_engine/engine.py` - Local Python executor
2. `dip_engine/vault.py` - Encrypted credential vault
3. `dip_engine/runtime.py` - Script runtime environment
4. `backend-go/internal/dip/types.go` - Core types
5. `backend-go/internal/dip/validator.go` - Script validation

### Phase 2: Pattern System (HIGH)

1. REST API pattern
2. OAuth pattern
3. Google provider pattern
4. Slack provider pattern
5. HubSpot provider pattern

### Phase 3: Distribution (HIGH)

1. WebSocket handler
2. Script distribution
3. Result collection
4. Error handling

### Phase 4: Desktop App Integration (MEDIUM)

1. Electron/Tauri wrapper
2. DIP engine management
3. Credential UI
4. Connection management

### Phase 5: MIOSA OS Integration (MEDIUM)

1. Systemd service
2. System keyring integration
3. Containerization
4. Auto-start

### Phase 6: Advanced Features (LOW)

1. Agentic workflows
2. Script caching
3. Offline mode
4. Analytics/logging

---

## Success Criteria

### Phase 1 Complete When:
- [ ] DIP Engine runs locally and executes scripts
- [ ] Credentials stored encrypted
- [ ] Basic Gmail search works

### Phase 2 Complete When:
- [ ] LLM generates correct scripts from patterns
- [ ] 5+ providers supported
- [ ] Scripts pass validation

### Phase 3 Complete When:
- [ ] WebSocket connection stable
- [ ] Scripts execute on user's machine
- [ ] Results returned to orchestrator

### Phase 4 Complete When:
- [ ] Desktop app bundles DIP Engine
- [ ] Users can connect integrations via UI
- [ ] Scripts execute from desktop app

### Phase 5 Complete When:
- [ ] DIP Engine runs as system service
- [ ] Uses system keyring
- [ ] Auto-connects on boot

---

## Summary

The **Dynamic Integration Protocol (DIP)** represents a paradigm shift in how AI agents interact with external services:

1. **No Context Bloat** - Only minimal patterns sent to LLM, not full tool definitions
2. **Universal** - Can integrate with ANY API, not just pre-built tools
3. **Secure** - Credentials never leave user's machine
4. **Agentic** - Scripts can call other agents and orchestrate workflows
5. **Privacy-Preserving** - User data stays local, only structured results shared

This is the foundation for true AI-powered automation that respects user privacy while enabling unlimited integration possibilities.
