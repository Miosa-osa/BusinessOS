# Sorx 2.0 - System of Reasoning

## Universal Skill-Based Integration Framework

---

## What is Sorx 2.0?

**Sorx 2.0** (System of Reasoning) is a next-generation integration framework where AI agents **learn skills** to connect with any system - modern APIs, legacy systems, databases, desktop applications, hardware, or anything with an interface.

Unlike traditional integration platforms that provide pre-built tools, Sorx 2.0 agents **acquire skills through experience** and **improve over time** - just like a human learning their job.

### The USB-C Paradigm

Just as USB-C is a universal connector that works with any device, Sorx 2.0 is a universal interface that works with any system:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SORX 2.0                                       │
│                         Universal Connector                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Modern Systems          Legacy Systems          Physical World            │
│   ──────────────          ──────────────          ──────────────            │
│   REST APIs               SOAP/XML-RPC            IoT Devices               │
│   GraphQL                 FTP/SFTP                Hardware APIs             │
│   WebSocket               EDI                     Serial/USB                │
│   gRPC                    Mainframe               Bluetooth                 │
│                                                                             │
│   Databases               Desktop Apps            File Systems              │
│   ──────────────          ──────────────          ──────────────            │
│   PostgreSQL              AppleScript             Local Files               │
│   MySQL                   PowerShell              Cloud Storage             │
│   MongoDB                 Windows COM             Network Shares            │
│   Redis                   X11 Automation          FTP/SFTP                  │
│                                                                             │
│   Enterprise              Communication           Custom                    │
│   ──────────────          ──────────────          ──────────────            │
│   SAP                     Email (SMTP/IMAP)       Any Protocol              │
│   Oracle                  SMS Gateways            Any Interface             │
│   Salesforce              Voice APIs              Any System                │
│   AS/400                  Fax Services            Anything                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Concept: Skills

### What is a Skill?

A **Skill** is a learned capability that an agent acquires to perform a specific task on a specific system. Skills are:

1. **Acquired** - Generated the first time an agent needs to do something
2. **Saved** - Stored for reuse
3. **Evolved** - Improved based on feedback and experience
4. **Shared** - Can be shared across agents and organizations
5. **Role-Based** - Different agent roles have different skill sets

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SKILL LIFECYCLE                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   1. ACQUISITION                                                            │
│      User: "Send an email to John with the project update"                  │
│      Agent: "I don't have this skill yet. Let me learn it..."              │
│      [Agent generates script based on pattern templates]                    │
│      [Skill saved: "gmail_send_email"]                                      │
│                                                                             │
│   2. EXECUTION                                                              │
│      User: "Email the team about the meeting"                              │
│      Agent: "I have the gmail_send_email skill. Executing..."              │
│      [Reuses saved skill with new parameters]                              │
│                                                                             │
│   3. EVOLUTION                                                              │
│      User: "That email didn't include attachments properly"                │
│      Agent: "Let me improve the skill..."                                  │
│      [Updates skill to handle attachments better]                          │
│      [Skill version incremented]                                           │
│                                                                             │
│   4. MASTERY                                                                │
│      After 50 executions with 98% success rate:                            │
│      [Skill marked as "mastered"]                                          │
│      [Can be shared to skill library]                                      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Skill Structure

```python
class Skill:
    # Identity
    id: str                      # "gmail_send_email_v3"
    name: str                    # "Send Email via Gmail"
    description: str             # "Sends an email using Gmail API"

    # Classification
    provider: str                # "google"
    category: str                # "communication"
    role_affinity: list[str]     # ["sales", "support", "marketing"]

    # The actual capability
    script: str                  # Python code to execute
    pattern_used: str            # Which pattern template was used
    interface_type: str          # "rest_api", "database", "desktop_app", etc.

    # Requirements
    credentials_needed: list[str]  # ["google"]
    dependencies: list[str]        # ["requests", "google-auth"]

    # Learning data
    version: int                 # 3
    executions: int              # 47
    success_rate: float          # 0.98
    avg_execution_time: float    # 1.2 seconds
    last_improved: datetime
    improvement_notes: list[str]

    # Evolution history
    previous_versions: list[str]  # Links to older versions
    failure_patterns: list[str]   # What went wrong before
    optimization_history: list[str]  # How it improved

    # Sharing
    is_public: bool              # Available in skill library?
    usage_count: int             # Times used by others
    rating: float                # Community rating
```

---

## Architecture

### The Dynamic Integration Protocol (DIP)

DIP is the underlying protocol that enables Sorx 2.0 to connect to any system:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SORX 2.0 ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   CLOUD LAYER (BusinessOS Backend)                                          │
│   ════════════════════════════════                                          │
│                                                                             │
│   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   │
│   │   Agent     │   │   Skill     │   │  Pattern    │   │  Skill      │   │
│   │   Router    │──▶│  Generator  │◀──│  Library    │   │  Library    │   │
│   └──────┬──────┘   └──────┬──────┘   └─────────────┘   └──────┬──────┘   │
│          │                 │                                    │          │
│          │    ┌────────────┴────────────┐                      │          │
│          │    │                         │                      │          │
│          ▼    ▼                         ▼                      ▼          │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                      Skill Orchestrator                              │ │
│   │  - Routes requests to appropriate skills                            │ │
│   │  - Manages skill acquisition, execution, evolution                  │ │
│   │  - Tracks skill performance and learning                            │ │
│   └─────────────────────────────────────┬───────────────────────────────┘ │
│                                         │                                  │
│                                         │ WebSocket / SSE                  │
│                                         ▼                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   LOCAL LAYER (User's Machine - Desktop App / MIOSA OS)                     │
│   ══════════════════════════════════════════════════════                    │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                        Sorx Engine                                    │ │
│   │                                                                       │ │
│   │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │ │
│   │  │  Credential   │  │    Skill      │  │   Execution   │           │ │
│   │  │    Vault      │  │    Cache      │  │   Sandbox     │           │ │
│   │  └───────────────┘  └───────────────┘  └───────────────┘           │ │
│   │                                                                       │ │
│   │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │ │
│   │  │   Interface   │  │   Protocol    │  │   Result      │           │ │
│   │  │   Adapters    │  │   Handlers    │  │   Processor   │           │ │
│   │  └───────────────┘  └───────────────┘  └───────────────┘           │ │
│   │                                                                       │ │
│   └─────────────────────────────────────────────────────────────────────┘ │
│                                         │                                  │
│                                         ▼                                  │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                     TARGET SYSTEMS                                    │ │
│   │  APIs │ Databases │ Desktop Apps │ Files │ Hardware │ Legacy        │ │
│   └─────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Interface Adapters

The key to universal connectivity is **Interface Adapters** - specialized handlers for different connection types:

### 1. REST API Adapter

```python
# adapters/rest_api.py

class RESTAdapter:
    """Handles REST API connections - the most common interface type."""

    capabilities = [
        "GET", "POST", "PUT", "PATCH", "DELETE",
        "OAuth2", "API Key", "Bearer Token", "Basic Auth",
        "JSON", "XML", "Form Data", "Multipart"
    ]

    def generate_skill(self, task: str, api_docs: str) -> Skill:
        """Generate a skill for a REST API endpoint."""
        # LLM generates Python requests code
        pass

    def execute(self, skill: Skill, params: dict) -> Result:
        """Execute a REST API skill."""
        pass
```

### 2. Database Adapter

```python
# adapters/database.py

class DatabaseAdapter:
    """Handles direct database connections."""

    capabilities = [
        "PostgreSQL", "MySQL", "SQLite", "MSSQL",
        "MongoDB", "Redis", "Elasticsearch",
        "Read", "Write", "Transactions"
    ]

    def generate_skill(self, task: str, schema: str) -> Skill:
        """Generate a skill for database operations."""
        # LLM generates SQL or query code
        pass
```

### 3. Legacy System Adapter

```python
# adapters/legacy.py

class LegacyAdapter:
    """Handles legacy system connections."""

    capabilities = [
        "SOAP", "XML-RPC", "EDI", "AS/400",
        "FTP", "SFTP", "Telnet", "SSH",
        "Mainframe", "COBOL Interfaces"
    ]

    def generate_skill(self, task: str, wsdl: str = None) -> Skill:
        """Generate a skill for legacy system operations."""
        # LLM generates SOAP calls, EDI messages, etc.
        pass
```

### 4. Desktop Automation Adapter

```python
# adapters/desktop.py

class DesktopAdapter:
    """Handles desktop application automation."""

    capabilities = {
        "macos": ["AppleScript", "JXA", "Automator"],
        "windows": ["PowerShell", "COM", "UI Automation"],
        "linux": ["X11", "DBus", "xdotool"]
    }

    def generate_skill(self, task: str, app: str) -> Skill:
        """Generate a skill to automate a desktop application."""
        # LLM generates AppleScript, PowerShell, etc.
        pass
```

### 5. File System Adapter

```python
# adapters/filesystem.py

class FileSystemAdapter:
    """Handles file system operations."""

    capabilities = [
        "Local Files", "Network Shares", "Cloud Storage",
        "FTP", "SFTP", "S3", "GCS", "Azure Blob",
        "Read", "Write", "Watch", "Transform"
    ]

    def generate_skill(self, task: str) -> Skill:
        """Generate a skill for file operations."""
        pass
```

### 6. Hardware/IoT Adapter

```python
# adapters/hardware.py

class HardwareAdapter:
    """Handles hardware and IoT connections."""

    capabilities = [
        "Serial", "USB", "Bluetooth", "Zigbee",
        "MQTT", "CoAP", "Modbus", "OPC-UA",
        "GPIO", "I2C", "SPI"
    ]

    def generate_skill(self, task: str, device_docs: str) -> Skill:
        """Generate a skill to interact with hardware."""
        pass
```

### 7. Communication Adapter

```python
# adapters/communication.py

class CommunicationAdapter:
    """Handles communication protocols."""

    capabilities = [
        "SMTP", "IMAP", "POP3",      # Email
        "SMS Gateways", "Twilio",     # SMS
        "SIP", "WebRTC",              # Voice
        "Fax APIs"                    # Fax
    ]

    def generate_skill(self, task: str) -> Skill:
        """Generate a skill for communication."""
        pass
```

---

## Skill Learning System

### Acquisition Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SKILL ACQUISITION FLOW                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   User Request: "Create a task in ClickUp when a deal closes in HubSpot"   │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 1. SKILL CHECK                                                       │  │
│   │    Agent checks skill library:                                       │  │
│   │    - hubspot_watch_deal_close: NOT FOUND                            │  │
│   │    - clickup_create_task: FOUND (v2, 94% success)                   │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 2. SKILL ACQUISITION                                                 │  │
│   │    Need to learn: hubspot_watch_deal_close                          │  │
│   │                                                                       │  │
│   │    a) Load HubSpot API pattern template                             │  │
│   │    b) Identify interface type: REST API + Webhooks                  │  │
│   │    c) Generate skill script:                                         │  │
│   │       - Subscribe to deal.propertyChange webhook                    │  │
│   │       - Filter for dealstage = "closedwon"                         │  │
│   │       - Extract deal data for downstream use                        │  │
│   │    d) Validate script (security, efficiency)                        │  │
│   │    e) Save skill to library                                         │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 3. WORKFLOW COMPOSITION                                              │  │
│   │    Combine skills into workflow:                                     │  │
│   │                                                                       │  │
│   │    [hubspot_watch_deal_close] ──▶ [clickup_create_task]            │  │
│   │                                                                       │  │
│   │    Workflow saved as compound skill:                                │  │
│   │    "hubspot_deal_to_clickup_task"                                   │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 4. EXECUTION                                                         │  │
│   │    - Deploy webhook listener                                        │  │
│   │    - Monitor for deal closes                                        │  │
│   │    - Create ClickUp tasks automatically                             │  │
│   │    - Track success/failure                                          │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Evolution Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SKILL EVOLUTION FLOW                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Feedback: "The ClickUp task didn't include the deal amount"              │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 1. ANALYZE FAILURE                                                   │  │
│   │    - Load skill: clickup_create_task v2                             │  │
│   │    - Identify gap: Missing field mapping for deal.amount            │  │
│   │    - Root cause: Skill wasn't configured to pass financial data     │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 2. IMPROVE SKILL                                                     │  │
│   │    Agent generates improved version:                                │  │
│   │                                                                       │  │
│   │    # v2: Original                                                    │  │
│   │    task_data = {                                                     │  │
│   │        "name": deal["dealname"],                                    │  │
│   │        "description": deal["description"]                           │  │
│   │    }                                                                 │  │
│   │                                                                       │  │
│   │    # v3: Improved                                                    │  │
│   │    task_data = {                                                     │  │
│   │        "name": deal["dealname"],                                    │  │
│   │        "description": deal["description"],                          │  │
│   │        "custom_fields": [                                           │  │
│   │            {"name": "Deal Amount", "value": deal["amount"]},       │  │
│   │            {"name": "Close Date", "value": deal["closedate"]}      │  │
│   │        ]                                                             │  │
│   │    }                                                                 │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 3. VERSION & SAVE                                                    │  │
│   │    - Increment version: v2 → v3                                     │  │
│   │    - Save improvement notes                                         │  │
│   │    - Archive v2 (don't delete - may need rollback)                  │  │
│   │    - Update success tracking                                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 4. LEARN & GENERALIZE                                                │  │
│   │    Agent notes pattern for future:                                  │  │
│   │    "When creating tasks from deals, always include financial data"  │  │
│   │                                                                       │  │
│   │    This learning is applied to:                                     │  │
│   │    - Similar skills (asana_create_task, linear_create_issue)        │  │
│   │    - Future skill generation                                         │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Pattern Library

Patterns are minimal templates that teach agents HOW to connect to different interface types. They're NOT full implementations - just enough for the LLM to generate correct skills.

### Pattern Categories

```
patterns/
├── interfaces/                      # Interface type patterns
│   ├── rest_api.py                  # REST API pattern
│   ├── graphql.py                   # GraphQL pattern
│   ├── soap.py                      # SOAP/XML-RPC pattern
│   ├── grpc.py                      # gRPC pattern
│   ├── websocket.py                 # WebSocket pattern
│   ├── database.py                  # Database pattern
│   ├── file_system.py               # File operations pattern
│   ├── desktop_macos.py             # macOS automation pattern
│   ├── desktop_windows.py           # Windows automation pattern
│   ├── desktop_linux.py             # Linux automation pattern
│   └── hardware.py                  # Hardware/IoT pattern
│
├── auth/                            # Authentication patterns
│   ├── oauth2.py                    # OAuth 2.0 flows
│   ├── api_key.py                   # API key auth
│   ├── basic_auth.py                # Basic auth
│   ├── jwt.py                       # JWT tokens
│   ├── saml.py                      # SAML (enterprise SSO)
│   └── certificate.py               # Certificate-based auth
│
├── providers/                       # Provider-specific patterns
│   ├── google/
│   │   ├── gmail.py
│   │   ├── drive.py
│   │   ├── calendar.py
│   │   └── sheets.py
│   ├── microsoft/
│   │   ├── outlook.py
│   │   ├── teams.py
│   │   └── onedrive.py
│   ├── salesforce/
│   │   └── salesforce.py
│   ├── hubspot/
│   │   └── hubspot.py
│   └── ... (all planned integrations)
│
└── enterprise/                      # Enterprise system patterns
    ├── sap.py                       # SAP integration
    ├── oracle.py                    # Oracle integration
    ├── as400.py                     # AS/400 mainframe
    ├── edi.py                       # EDI messaging
    └── ldap.py                      # LDAP/Active Directory
```

### Pattern Example: REST API

```python
# patterns/interfaces/rest_api.py

REST_API_PATTERN = """
## REST API Skill Pattern

### Structure:
```python
import requests
from sorx import get_credential, return_result, log

def execute(params: dict):
    '''
    [SKILL_DESCRIPTION]

    Args:
        params: {
            [PARAM_DEFINITIONS]
        }

    Returns:
        Structured result via return_result()
    '''

    # 1. Get credentials from local vault
    creds = get_credential("[PROVIDER]")

    # 2. Build request
    url = f"{creds['base_url']}/[ENDPOINT]"
    headers = {
        "Authorization": f"Bearer {creds['access_token']}",
        "Content-Type": "application/json"
    }

    # 3. Make request
    response = requests.[METHOD](
        url,
        headers=headers,
        json=params.get("body"),      # For POST/PUT
        params=params.get("query")     # For GET
    )

    # 4. Handle token expiration
    if response.status_code == 401:
        # Trigger token refresh flow
        return_result({"error": "token_expired", "refresh": True})
        return

    # 5. Handle rate limiting
    if response.status_code == 429:
        retry_after = response.headers.get("Retry-After", 60)
        return_result({"error": "rate_limited", "retry_after": retry_after})
        return

    # 6. Return structured result
    if response.ok:
        return_result({
            "success": True,
            "data": response.json()
        })
    else:
        return_result({
            "success": False,
            "error": response.text,
            "status_code": response.status_code
        })
```

### Common Patterns:
- List: GET /resources → returns array
- Get: GET /resources/{id} → returns object
- Create: POST /resources → returns created object
- Update: PUT/PATCH /resources/{id} → returns updated object
- Delete: DELETE /resources/{id} → returns success/empty
- Search: GET /resources?query=... → returns filtered array

### Pagination:
- Cursor-based: ?cursor=abc123
- Offset-based: ?offset=100&limit=50
- Page-based: ?page=2&per_page=50

### Error Handling:
- 400: Bad request (log params, fix request)
- 401: Unauthorized (refresh token)
- 403: Forbidden (check permissions)
- 404: Not found (verify resource exists)
- 429: Rate limited (backoff and retry)
- 500: Server error (retry with backoff)
"""
```

### Pattern Example: Legacy SOAP

```python
# patterns/interfaces/soap.py

SOAP_PATTERN = """
## SOAP/XML-RPC Skill Pattern

For connecting to legacy enterprise systems using SOAP.

### Structure:
```python
from zeep import Client
from zeep.wsse.username import UsernameToken
from sorx import get_credential, return_result

def execute(params: dict):
    '''
    [SKILL_DESCRIPTION]
    '''

    # 1. Get credentials
    creds = get_credential("[PROVIDER]")

    # 2. Create SOAP client with WSDL
    wsse = UsernameToken(creds['username'], creds['password'])
    client = Client(
        wsdl=creds['wsdl_url'],
        wsse=wsse
    )

    # 3. Call SOAP method
    try:
        result = client.service.[METHOD_NAME](
            [PARAMETERS]
        )

        return_result({
            "success": True,
            "data": serialize_zeep_result(result)
        })

    except Exception as e:
        return_result({
            "success": False,
            "error": str(e)
        })
```

### Notes:
- Always use zeep library for SOAP
- Get WSDL URL from provider
- Handle complex types properly
- Serialize results to JSON
"""
```

### Pattern Example: Desktop Automation (macOS)

```python
# patterns/interfaces/desktop_macos.py

MACOS_PATTERN = """
## macOS Desktop Automation Pattern

For automating macOS applications using AppleScript/JXA.

### Structure:
```python
import subprocess
from sorx import return_result

def execute(params: dict):
    '''
    [SKILL_DESCRIPTION]
    '''

    # AppleScript to execute
    script = '''
    tell application "[APP_NAME]"
        [COMMANDS]
    end tell
    '''

    # Execute AppleScript
    result = subprocess.run(
        ['osascript', '-e', script],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        return_result({
            "success": True,
            "output": result.stdout
        })
    else:
        return_result({
            "success": False,
            "error": result.stderr
        })
```

### Common Apps:
- Finder: File operations
- Mail: Email (alternative to API)
- Calendar: Local calendar
- Notes: Local notes
- Numbers/Pages/Keynote: Office suite
- Any app with AppleScript support

### JXA Alternative:
```javascript
// For JavaScript-based automation
const app = Application('[APP_NAME]');
app.includeStandardAdditions = true;
// ... commands
```
"""
```

---

## Role-Based Skills

Different agent roles have affinities for different skill categories:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          ROLE-BASED SKILL MATRIX                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   SALES AGENT                                                               │
│   ───────────                                                               │
│   Primary Skills:                                                           │
│   - CRM operations (HubSpot, Salesforce, Pipedrive)                        │
│   - Email outreach (Gmail, Outlook)                                        │
│   - Calendar management                                                     │
│   - Document generation (proposals, quotes)                                │
│                                                                             │
│   SUPPORT AGENT                                                             │
│   ─────────────                                                             │
│   Primary Skills:                                                           │
│   - Ticketing systems (Zendesk, Intercom, Freshdesk)                       │
│   - Knowledge base operations                                               │
│   - Customer lookup (CRM)                                                   │
│   - Communication (email, chat, SMS)                                       │
│                                                                             │
│   MARKETING AGENT                                                           │
│   ───────────────                                                           │
│   Primary Skills:                                                           │
│   - Email campaigns (Mailchimp, Klaviyo)                                   │
│   - Social media (Buffer, Hootsuite)                                       │
│   - Analytics (Google Analytics, Mixpanel)                                 │
│   - Content management (WordPress, Webflow)                                │
│                                                                             │
│   OPERATIONS AGENT                                                          │
│   ────────────────                                                          │
│   Primary Skills:                                                           │
│   - Task management (ClickUp, Asana, Monday)                               │
│   - Documentation (Notion, Confluence)                                      │
│   - File management (Drive, Dropbox)                                       │
│   - Team communication (Slack, Teams)                                      │
│                                                                             │
│   FINANCE AGENT                                                             │
│   ─────────────                                                             │
│   Primary Skills:                                                           │
│   - Accounting (QuickBooks, Xero)                                          │
│   - Payments (Stripe, PayPal)                                              │
│   - Invoicing                                                               │
│   - Reporting                                                               │
│                                                                             │
│   DEVELOPER AGENT                                                           │
│   ───────────────                                                           │
│   Primary Skills:                                                           │
│   - Code repositories (GitHub, GitLab)                                     │
│   - Issue tracking (Jira, Linear)                                          │
│   - CI/CD pipelines                                                         │
│   - Monitoring (Datadog, Sentry)                                           │
│                                                                             │
│   IT ADMIN AGENT                                                            │
│   ──────────────                                                            │
│   Primary Skills:                                                           │
│   - Directory services (LDAP, Azure AD)                                    │
│   - Device management                                                       │
│   - Security tools                                                          │
│   - Legacy system access                                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## BusinessOS Skill Catalog

Concrete skill examples mapped to BusinessOS integrations. These skills grow organically as agents learn to perform business workflows.

### Skill File Structure

```
~/.businessos/skills/
├── crm/                                    # CRM & Sales
│   ├── hubspot-deal-pipeline.md            ~3.2k tokens
│   ├── hubspot-contact-enrichment.md       ~2.8k tokens
│   ├── hubspot-company-lookup.md           ~1.9k tokens
│   ├── lead-qualification-workflow.md      ~4.1k tokens
│   ├── client-health-score.md              ~3.5k tokens
│   └── sales-handoff-automation.md         ~2.7k tokens
│
├── communication/                          # Email & Messaging
│   ├── gmail-client-outreach.md            ~2.4k tokens
│   ├── gmail-followup-sequence.md          ~3.8k tokens
│   ├── gmail-meeting-request.md            ~1.7k tokens
│   ├── slack-team-notification.md          ~1.5k tokens
│   ├── slack-client-channel-update.md      ~2.1k tokens
│   ├── slack-standup-collection.md         ~2.9k tokens
│   └── multi-channel-announcement.md       ~3.3k tokens
│
├── tasks/                                  # Task & Project Management
│   ├── clickup-task-from-email.md          ~2.6k tokens
│   ├── clickup-project-setup.md            ~4.2k tokens
│   ├── clickup-sprint-planning.md          ~3.7k tokens
│   ├── asana-milestone-tracking.md         ~2.8k tokens
│   ├── task-delegation-workflow.md         ~3.1k tokens
│   └── weekly-task-summary.md              ~2.3k tokens
│
├── documents/                              # Documents & Knowledge
│   ├── notion-meeting-notes.md             ~2.5k tokens
│   ├── notion-project-wiki.md              ~3.4k tokens
│   ├── notion-client-database.md           ~2.9k tokens
│   ├── drive-proposal-generation.md        ~4.5k tokens
│   ├── drive-contract-organization.md      ~2.2k tokens
│   └── knowledge-base-update.md            ~2.7k tokens
│
├── meetings/                               # Calendar & Meetings
│   ├── calendar-smart-scheduling.md        ~3.1k tokens
│   ├── calendar-availability-check.md      ~1.8k tokens
│   ├── zoom-meeting-setup.md               ~2.3k tokens
│   ├── meeting-prep-workflow.md            ~3.6k tokens
│   ├── meeting-followup-automation.md      ~4.0k tokens
│   └── recurring-meeting-management.md     ~2.4k tokens
│
├── finance/                                # Finance & Billing
│   ├── stripe-invoice-generation.md        ~2.8k tokens
│   ├── stripe-payment-tracking.md          ~2.1k tokens
│   ├── overdue-payment-reminder.md         ~3.2k tokens
│   ├── revenue-report-generation.md        ~3.9k tokens
│   └── expense-categorization.md           ~2.5k tokens
│
├── workflows/                              # Compound Workflows
│   ├── new-client-onboarding.md            ~5.8k tokens
│   ├── deal-won-automation.md              ~4.7k tokens
│   ├── project-kickoff-sequence.md         ~5.2k tokens
│   ├── weekly-report-distribution.md       ~3.8k tokens
│   ├── client-renewal-workflow.md          ~4.4k tokens
│   └── escalation-handling.md              ~3.6k tokens
│
└── meta/                                   # Skills about Skills
    ├── skill-creator.md                    ~2.1k tokens
    ├── workflow-analyzer.md                ~2.8k tokens
    └── skill-optimizer.md                  ~2.4k tokens
```

---

### Skill Examples: CRM (HubSpot)

#### hubspot-deal-pipeline.md

```markdown
# HubSpot Deal Pipeline Management

## Skill Metadata
- **ID**: hubspot-deal-pipeline-v4
- **Category**: crm
- **Provider**: hubspot
- **Role Affinity**: sales, account-management
- **Credentials**: hubspot
- **Version**: 4
- **Executions**: 847
- **Success Rate**: 98.7%

## Description
Manages the full deal lifecycle in HubSpot - creating deals, updating stages,
tracking activities, and triggering downstream workflows.

## Capabilities
- Create new deals with proper associations
- Move deals through pipeline stages
- Add notes and activities to deals
- Track deal value and probability
- Trigger notifications on stage changes

## Script

```python
import requests
from sorx import get_credential, return_result, trigger_skill

def execute(params: dict):
    """
    HubSpot Deal Pipeline Management

    Actions:
        - create: Create new deal
        - update_stage: Move deal to new stage
        - add_activity: Add note/call/email to deal
        - get_pipeline: Get pipeline summary
    """
    creds = get_credential("hubspot")
    base_url = "https://api.hubapi.com"
    headers = {
        "Authorization": f"Bearer {creds['access_token']}",
        "Content-Type": "application/json"
    }

    action = params.get("action")

    if action == "create":
        deal_data = {
            "properties": {
                "dealname": params["name"],
                "amount": params.get("amount", 0),
                "dealstage": params.get("stage", "appointmentscheduled"),
                "pipeline": params.get("pipeline", "default"),
                "closedate": params.get("close_date"),
                "hubspot_owner_id": params.get("owner_id")
            }
        }

        # Associate with contact and company if provided
        if params.get("contact_id") or params.get("company_id"):
            deal_data["associations"] = []
            if params.get("contact_id"):
                deal_data["associations"].append({
                    "to": {"id": params["contact_id"]},
                    "types": [{"associationCategory": "HUBSPOT_DEFINED", "associationTypeId": 3}]
                })
            if params.get("company_id"):
                deal_data["associations"].append({
                    "to": {"id": params["company_id"]},
                    "types": [{"associationCategory": "HUBSPOT_DEFINED", "associationTypeId": 5}]
                })

        response = requests.post(
            f"{base_url}/crm/v3/objects/deals",
            headers=headers,
            json=deal_data
        )

        if response.ok:
            deal = response.json()
            # Trigger notification skill
            trigger_skill("slack-team-notification", {
                "channel": "#sales",
                "message": f"New deal created: {params['name']} (${params.get('amount', 0):,})"
            })
            return_result({"success": True, "deal_id": deal["id"], "deal": deal})
        else:
            return_result({"success": False, "error": response.text})

    elif action == "update_stage":
        deal_id = params["deal_id"]
        new_stage = params["stage"]

        # Get current deal first
        current = requests.get(
            f"{base_url}/crm/v3/objects/deals/{deal_id}",
            headers=headers
        ).json()

        old_stage = current["properties"].get("dealstage")

        response = requests.patch(
            f"{base_url}/crm/v3/objects/deals/{deal_id}",
            headers=headers,
            json={"properties": {"dealstage": new_stage}}
        )

        if response.ok:
            # Check if deal was won
            if new_stage == "closedwon":
                trigger_skill("deal-won-automation", {
                    "deal_id": deal_id,
                    "deal_name": current["properties"]["dealname"],
                    "amount": current["properties"].get("amount")
                })

            return_result({
                "success": True,
                "deal_id": deal_id,
                "old_stage": old_stage,
                "new_stage": new_stage
            })
        else:
            return_result({"success": False, "error": response.text})

    elif action == "add_activity":
        # Add engagement (note, call, email, meeting)
        engagement_data = {
            "engagement": {
                "type": params.get("type", "NOTE"),
                "timestamp": params.get("timestamp", int(time.time() * 1000))
            },
            "associations": {
                "dealIds": [params["deal_id"]]
            },
            "metadata": {
                "body": params["content"]
            }
        }

        response = requests.post(
            f"{base_url}/engagements/v1/engagements",
            headers=headers,
            json=engagement_data
        )

        return_result({"success": response.ok, "engagement": response.json() if response.ok else None})

    elif action == "get_pipeline":
        # Get all deals in pipeline with summary
        response = requests.post(
            f"{base_url}/crm/v3/objects/deals/search",
            headers=headers,
            json={
                "filterGroups": [{
                    "filters": [{
                        "propertyName": "pipeline",
                        "operator": "EQ",
                        "value": params.get("pipeline", "default")
                    }]
                }],
                "properties": ["dealname", "amount", "dealstage", "closedate"],
                "limit": 100
            }
        )

        if response.ok:
            deals = response.json()["results"]
            # Group by stage
            by_stage = {}
            total_value = 0
            for deal in deals:
                stage = deal["properties"]["dealstage"]
                if stage not in by_stage:
                    by_stage[stage] = {"count": 0, "value": 0}
                by_stage[stage]["count"] += 1
                amount = float(deal["properties"].get("amount") or 0)
                by_stage[stage]["value"] += amount
                total_value += amount

            return_result({
                "success": True,
                "total_deals": len(deals),
                "total_value": total_value,
                "by_stage": by_stage,
                "deals": deals
            })
        else:
            return_result({"success": False, "error": response.text})
```

## Evolution History
- **v1**: Basic deal creation
- **v2**: Added stage updates and associations
- **v3**: Added activity tracking, fixed amount formatting
- **v4**: Added pipeline summary, integrated with notification skills

## Learned Patterns
- Always associate deals with contacts AND companies when available
- Include dollar amounts in notifications for context
- Trigger downstream automations on stage changes (especially closedwon)
```

---

### Skill Examples: Communication (Gmail + Slack)

#### gmail-followup-sequence.md

```markdown
# Gmail Follow-up Sequence

## Skill Metadata
- **ID**: gmail-followup-sequence-v3
- **Category**: communication
- **Provider**: google
- **Role Affinity**: sales, account-management, support
- **Credentials**: google
- **Version**: 3
- **Executions**: 1,247
- **Success Rate**: 97.2%

## Description
Manages intelligent email follow-up sequences. Tracks sent emails, schedules
follow-ups, personalizes based on context, and stops when recipient responds.

## Capabilities
- Create follow-up sequences (3-5 touches)
- Personalize emails based on recipient data
- Track opens and responses
- Auto-stop when recipient replies
- Variable timing between touches

## Script

```python
import requests
import base64
from email.mime.text import MIMEText
from datetime import datetime, timedelta
from sorx import get_credential, return_result, schedule_skill, get_context

def execute(params: dict):
    """
    Gmail Follow-up Sequence Management

    Actions:
        - start_sequence: Begin a new follow-up sequence
        - send_followup: Send the next follow-up in sequence
        - check_response: Check if recipient responded
        - stop_sequence: End sequence early
    """
    creds = get_credential("google")
    headers = {"Authorization": f"Bearer {creds['access_token']}"}

    action = params.get("action")

    if action == "start_sequence":
        recipient = params["to"]
        subject = params["subject"]
        sequence_id = f"seq_{recipient}_{int(time.time())}"

        # Get recipient context from CRM if available
        context = get_context("client", recipient)

        # Personalize first email
        body = personalize_email(
            template=params["template"],
            recipient_name=context.get("name", "there"),
            company=context.get("company"),
            custom_fields=params.get("custom_fields", {})
        )

        # Send first email
        message = create_email(recipient, subject, body)
        sent = send_email(headers, message)

        if sent["success"]:
            # Schedule follow-ups
            schedule_followups(sequence_id, params)

            return_result({
                "success": True,
                "sequence_id": sequence_id,
                "first_email_id": sent["message_id"],
                "followups_scheduled": len(params.get("followup_templates", []))
            })
        else:
            return_result({"success": False, "error": sent["error"]})

    elif action == "send_followup":
        sequence_id = params["sequence_id"]
        followup_number = params["followup_number"]

        # Check if recipient already responded
        if check_for_response(headers, params["thread_id"]):
            # Stop sequence - they replied!
            return_result({
                "success": True,
                "action": "sequence_stopped",
                "reason": "recipient_replied"
            })

        # Get the right template
        template = params["templates"][followup_number - 1]

        # Personalize
        context = get_context("client", params["to"])
        body = personalize_email(
            template=template,
            recipient_name=context.get("name", "there"),
            company=context.get("company"),
            followup_number=followup_number
        )

        # Send as reply to thread
        message = create_reply(params["thread_id"], body)
        sent = send_email(headers, message)

        if sent["success"]:
            # Schedule next if more remain
            if followup_number < len(params["templates"]):
                schedule_skill("gmail-followup-sequence", {
                    "action": "send_followup",
                    "sequence_id": sequence_id,
                    "followup_number": followup_number + 1,
                    "thread_id": params["thread_id"],
                    "to": params["to"],
                    "templates": params["templates"]
                }, delay_days=params.get("days_between", 3))

            return_result({
                "success": True,
                "followup_sent": followup_number,
                "remaining": len(params["templates"]) - followup_number
            })
        else:
            return_result({"success": False, "error": sent["error"]})

    elif action == "check_response":
        thread_id = params["thread_id"]
        response = requests.get(
            f"https://gmail.googleapis.com/gmail/v1/users/me/threads/{thread_id}",
            headers=headers
        )

        if response.ok:
            thread = response.json()
            messages = thread.get("messages", [])

            # Check if any message is FROM the recipient (not us)
            our_email = creds.get("email")
            for msg in messages:
                headers_list = msg.get("payload", {}).get("headers", [])
                from_header = next((h["value"] for h in headers_list if h["name"] == "From"), "")
                if our_email not in from_header:
                    return_result({
                        "success": True,
                        "has_response": True,
                        "response_snippet": msg.get("snippet")
                    })

            return_result({"success": True, "has_response": False})
        else:
            return_result({"success": False, "error": response.text})


def personalize_email(template, recipient_name, company=None, **kwargs):
    """Replace placeholders with actual values."""
    result = template
    result = result.replace("{{name}}", recipient_name)
    result = result.replace("{{company}}", company or "your company")
    result = result.replace("{{followup_number}}", str(kwargs.get("followup_number", 1)))

    for key, value in kwargs.get("custom_fields", {}).items():
        result = result.replace(f"{{{{{key}}}}}", str(value))

    return result


def create_email(to, subject, body):
    """Create a MIME email message."""
    message = MIMEText(body)
    message["to"] = to
    message["subject"] = subject
    return {"raw": base64.urlsafe_b64encode(message.as_bytes()).decode()}


def send_email(headers, message):
    """Send email via Gmail API."""
    response = requests.post(
        "https://gmail.googleapis.com/gmail/v1/users/me/messages/send",
        headers={**headers, "Content-Type": "application/json"},
        json=message
    )
    if response.ok:
        return {"success": True, "message_id": response.json()["id"]}
    return {"success": False, "error": response.text}
```

## Sequence Templates

### Sales Follow-up (3 touches)
```
Touch 1: Initial outreach with value prop
Touch 2 (+3 days): "Wanted to make sure you saw my email..."
Touch 3 (+5 days): "One last follow-up..." with different angle
```

### Meeting Request (2 touches)
```
Touch 1: Meeting request with proposed times
Touch 2 (+2 days): "Still hoping to connect..."
```

## Evolution History
- **v1**: Basic single follow-up email
- **v2**: Added sequence support with scheduling
- **v3**: Added response detection, personalization, CRM integration

## Learned Patterns
- Always check for response before sending follow-up
- 3-day spacing works well for sales, 2-day for urgent
- Include "Re:" in subject to improve open rates
- Stop immediately when recipient replies (don't annoy)
```

---

#### slack-team-notification.md

```markdown
# Slack Team Notification

## Skill Metadata
- **ID**: slack-team-notification-v5
- **Category**: communication
- **Provider**: slack
- **Role Affinity**: all
- **Credentials**: slack
- **Version**: 5
- **Executions**: 3,892
- **Success Rate**: 99.4%

## Description
Sends formatted notifications to Slack channels with context-aware formatting,
thread support, and interactive elements.

## Script

```python
import requests
from sorx import get_credential, return_result

def execute(params: dict):
    """
    Slack Team Notification

    Supports:
        - Simple messages
        - Rich formatted blocks
        - Thread replies
        - Mentions (@user, @channel)
        - Attachments with color coding
    """
    creds = get_credential("slack")
    headers = {
        "Authorization": f"Bearer {creds['bot_token']}",
        "Content-Type": "application/json"
    }

    channel = params["channel"]
    message = params.get("message", "")
    notification_type = params.get("type", "info")

    # Build message payload
    payload = {"channel": channel}

    # If simple message, just send text
    if params.get("simple"):
        payload["text"] = message
    else:
        # Build rich blocks
        blocks = []

        # Add header if provided
        if params.get("title"):
            blocks.append({
                "type": "header",
                "text": {"type": "plain_text", "text": params["title"]}
            })

        # Add main message
        if message:
            blocks.append({
                "type": "section",
                "text": {"type": "mrkdwn", "text": message}
            })

        # Add fields if provided (key-value pairs)
        if params.get("fields"):
            fields_block = {
                "type": "section",
                "fields": [
                    {"type": "mrkdwn", "text": f"*{k}:*\n{v}"}
                    for k, v in params["fields"].items()
                ]
            }
            blocks.append(fields_block)

        # Add action buttons if provided
        if params.get("actions"):
            actions_block = {
                "type": "actions",
                "elements": [
                    {
                        "type": "button",
                        "text": {"type": "plain_text", "text": action["text"]},
                        "url": action.get("url"),
                        "action_id": action.get("action_id", f"action_{i}")
                    }
                    for i, action in enumerate(params["actions"])
                ]
            }
            blocks.append(actions_block)

        # Add context footer
        if params.get("footer"):
            blocks.append({
                "type": "context",
                "elements": [{"type": "mrkdwn", "text": params["footer"]}]
            })

        payload["blocks"] = blocks

        # Add color-coded attachment based on type
        colors = {
            "success": "#36a64f",
            "warning": "#ffcc00",
            "error": "#ff0000",
            "info": "#0066ff"
        }
        if notification_type in colors:
            payload["attachments"] = [{
                "color": colors[notification_type],
                "blocks": blocks
            }]
            payload.pop("blocks")  # Move blocks into attachment

    # Thread reply
    if params.get("thread_ts"):
        payload["thread_ts"] = params["thread_ts"]

    # Send
    response = requests.post(
        "https://slack.com/api/chat.postMessage",
        headers=headers,
        json=payload
    )

    if response.ok and response.json().get("ok"):
        result = response.json()
        return_result({
            "success": True,
            "ts": result["ts"],
            "channel": result["channel"],
            "thread_ts": result.get("message", {}).get("thread_ts")
        })
    else:
        return_result({
            "success": False,
            "error": response.json().get("error", response.text)
        })
```

## Usage Examples

### Deal Won Notification
```python
trigger_skill("slack-team-notification", {
    "channel": "#sales-wins",
    "type": "success",
    "title": "Deal Won!",
    "message": "Congratulations! We just closed *Acme Corp*",
    "fields": {
        "Deal Value": "$45,000",
        "Sales Rep": "@john",
        "Close Date": "Jan 4, 2026"
    },
    "actions": [
        {"text": "View in HubSpot", "url": "https://app.hubspot.com/deals/123"}
    ],
    "footer": "Via BusinessOS Automation"
})
```

### Error Alert
```python
trigger_skill("slack-team-notification", {
    "channel": "#alerts",
    "type": "error",
    "title": "Integration Error",
    "message": "Failed to sync contacts from HubSpot",
    "fields": {
        "Error": "Rate limit exceeded",
        "Retry In": "15 minutes"
    }
})
```

## Evolution History
- **v1**: Simple text messages
- **v2**: Added blocks formatting
- **v3**: Added attachments with colors
- **v4**: Added thread support, actions
- **v5**: Added fields layout, footer, type-based coloring
```

---

### Skill Examples: Tasks (ClickUp)

#### clickup-task-from-email.md

```markdown
# ClickUp Task from Email

## Skill Metadata
- **ID**: clickup-task-from-email-v4
- **Category**: tasks
- **Provider**: clickup
- **Role Affinity**: operations, project-management, support
- **Credentials**: clickup
- **Version**: 4
- **Executions**: 562
- **Success Rate**: 96.8%

## Description
Intelligently converts emails into ClickUp tasks with proper categorization,
priority detection, assignee matching, and deadline extraction.

## Script

```python
import requests
import re
from datetime import datetime, timedelta
from sorx import get_credential, return_result, call_llm

def execute(params: dict):
    """
    Create ClickUp task from email content.

    Automatically extracts:
        - Task name from subject
        - Description from body
        - Priority from urgency keywords
        - Due date from mentioned dates
        - Assignee from @mentions or context
    """
    creds = get_credential("clickup")
    headers = {"Authorization": creds["api_key"]}

    email = params["email"]
    list_id = params.get("list_id") or creds.get("default_list_id")

    # Extract task details using LLM
    extraction = call_llm(
        prompt=f"""Extract task details from this email:

Subject: {email['subject']}
From: {email['from']}
Body: {email['body']}

Return JSON with:
- task_name: Clear, actionable task title
- description: Key details and context
- priority: 1 (urgent), 2 (high), 3 (normal), 4 (low)
- due_date: ISO date if mentioned, null otherwise
- tags: relevant tags (max 3)
""",
        response_format="json"
    )

    # Build task payload
    task_data = {
        "name": extraction["task_name"],
        "description": f"""**From Email**
From: {email['from']}
Subject: {email['subject']}
Date: {email['date']}

---

{extraction['description']}

---
*Original email:*
{email['body'][:2000]}
""",
        "priority": extraction["priority"],
        "tags": extraction.get("tags", []),
        "custom_fields": []
    }

    # Add due date if extracted
    if extraction.get("due_date"):
        task_data["due_date"] = int(datetime.fromisoformat(
            extraction["due_date"]
        ).timestamp() * 1000)

    # Add source tracking
    task_data["custom_fields"].append({
        "id": creds.get("source_field_id"),
        "value": "email"
    })

    # Link to original email if we have message ID
    if email.get("message_id"):
        task_data["custom_fields"].append({
            "id": creds.get("email_link_field_id"),
            "value": f"gmail://message/{email['message_id']}"
        })

    # Try to match assignee
    assignee = match_assignee(email, creds.get("team_members", []))
    if assignee:
        task_data["assignees"] = [assignee["id"]]

    # Create task
    response = requests.post(
        f"https://api.clickup.com/api/v2/list/{list_id}/task",
        headers=headers,
        json=task_data
    )

    if response.ok:
        task = response.json()

        # Notify assignee
        if assignee:
            trigger_skill("slack-team-notification", {
                "channel": f"@{assignee['username']}",
                "type": "info",
                "message": f"New task assigned from email: *{extraction['task_name']}*",
                "actions": [{"text": "View Task", "url": task["url"]}]
            })

        return_result({
            "success": True,
            "task_id": task["id"],
            "task_url": task["url"],
            "task_name": extraction["task_name"],
            "priority": extraction["priority"],
            "assignee": assignee["name"] if assignee else None
        })
    else:
        return_result({"success": False, "error": response.text})


def match_assignee(email, team_members):
    """Match email sender or @mentions to team member."""
    from_email = email["from"].lower()

    # Check if sender is a team member
    for member in team_members:
        if member["email"].lower() in from_email:
            return member

    # Check for @mentions in body
    mentions = re.findall(r'@(\w+)', email["body"])
    for mention in mentions:
        for member in team_members:
            if mention.lower() in member["username"].lower():
                return member

    return None
```

## Example Transformation

**Input Email:**
```
From: john@client.com
Subject: URGENT: Need proposal updates by Friday
Body: Hi team, we need the updated proposal with the new pricing
by end of day Friday. @sarah can you handle this?
```

**Output Task:**
```
Name: Update proposal with new pricing for John
Priority: 1 (Urgent)
Due: Friday EOD
Assignee: Sarah
Tags: [proposal, client-request]
Description: Client needs updated proposal with new pricing...
```

## Evolution History
- **v1**: Basic email-to-task with manual fields
- **v2**: Added LLM extraction for smart parsing
- **v3**: Added assignee matching, priority detection
- **v4**: Added email linking, source tracking, notifications

## Learned Patterns
- Include original email snippet in description for context
- "URGENT" in subject = priority 1
- Dates like "by Friday" should be parsed to actual dates
- Always link back to source email for reference
```

---

### Skill Examples: Compound Workflows

#### deal-won-automation.md

```markdown
# Deal Won Automation

## Skill Metadata
- **ID**: deal-won-automation-v3
- **Category**: workflows
- **Provider**: multi (hubspot, clickup, slack, gmail, notion)
- **Role Affinity**: sales, operations, account-management
- **Credentials**: hubspot, clickup, slack, google, notion
- **Version**: 3
- **Executions**: 234
- **Success Rate**: 98.3%

## Description
Comprehensive automation triggered when a deal is marked as won in HubSpot.
Orchestrates multiple downstream actions across systems.

## Workflow Diagram

```
Deal Won in HubSpot
        │
        ├──► Slack: Celebrate in #sales-wins
        │
        ├──► ClickUp: Create onboarding project
        │       └──► With checklist tasks
        │
        ├──► Gmail: Send welcome email to client
        │
        ├──► Notion: Create client wiki page
        │
        ├──► HubSpot: Update deal properties
        │       ├──► Set closed date
        │       └──► Move contact to "Customer" lifecycle
        │
        └──► Calendar: Schedule kickoff meeting
```

## Script

```python
import asyncio
from sorx import get_credential, return_result, trigger_skill, call_llm
from datetime import datetime, timedelta

def execute(params: dict):
    """
    Deal Won Automation - Full post-sale workflow

    Triggered when: HubSpot deal stage = closedwon
    """
    deal_id = params["deal_id"]
    deal_name = params["deal_name"]
    amount = params.get("amount", 0)

    # Get full deal details from HubSpot
    deal = get_deal_details(deal_id)
    contact = get_associated_contact(deal_id)
    company = get_associated_company(deal_id)

    results = {
        "deal_id": deal_id,
        "deal_name": deal_name,
        "actions_completed": []
    }

    # ═══════════════════════════════════════════════════════════════
    # 1. CELEBRATE - Notify the team
    # ═══════════════════════════════════════════════════════════════

    slack_result = trigger_skill("slack-team-notification", {
        "channel": "#sales-wins",
        "type": "success",
        "title": f"Deal Won: {deal_name}",
        "message": f"*{deal.get('owner_name', 'Team')}* just closed a deal!",
        "fields": {
            "Client": company.get("name", "Unknown"),
            "Value": f"${amount:,.2f}",
            "Sales Cycle": f"{deal.get('days_to_close', '?')} days"
        },
        "actions": [
            {"text": "View Deal", "url": f"https://app.hubspot.com/deals/{deal_id}"}
        ]
    })
    results["actions_completed"].append({"action": "slack_notification", "success": slack_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # 2. CREATE PROJECT - Set up onboarding in ClickUp
    # ═══════════════════════════════════════════════════════════════

    project_result = trigger_skill("clickup-project-setup", {
        "template": "client-onboarding",
        "name": f"Onboarding: {company.get('name', deal_name)}",
        "custom_fields": {
            "client_name": company.get("name"),
            "deal_value": amount,
            "hubspot_deal_id": deal_id,
            "primary_contact": contact.get("email")
        },
        "tasks": [
            {"name": "Send welcome packet", "assignee": "onboarding-team", "due_days": 1},
            {"name": "Schedule kickoff call", "assignee": "account-manager", "due_days": 2},
            {"name": "Set up client workspace", "assignee": "ops-team", "due_days": 3},
            {"name": "Create project plan", "assignee": "project-manager", "due_days": 5},
            {"name": "Send first invoice", "assignee": "finance", "due_days": 7}
        ]
    })
    results["actions_completed"].append({"action": "clickup_project", "success": project_result["success"]})
    results["project_url"] = project_result.get("project_url")

    # ═══════════════════════════════════════════════════════════════
    # 3. WELCOME EMAIL - Send to client
    # ═══════════════════════════════════════════════════════════════

    if contact.get("email"):
        # Generate personalized welcome email
        email_content = call_llm(
            prompt=f"""Write a warm, professional welcome email for a new client.

Client: {contact.get('firstname', 'there')} at {company.get('name')}
Deal: {deal_name}
Our company: [Your Company]

Include:
- Gratitude for choosing us
- What happens next (onboarding process)
- Who their point of contact will be
- How to reach us

Keep it concise and friendly.
"""
        )

        email_result = trigger_skill("gmail-client-outreach", {
            "to": contact["email"],
            "subject": f"Welcome to [Your Company], {contact.get('firstname', '')}!",
            "body": email_content,
            "track": True
        })
        results["actions_completed"].append({"action": "welcome_email", "success": email_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # 4. DOCUMENTATION - Create client wiki in Notion
    # ═══════════════════════════════════════════════════════════════

    notion_result = trigger_skill("notion-client-database", {
        "action": "create_page",
        "database": "Clients",
        "properties": {
            "Name": company.get("name", deal_name),
            "Status": "Onboarding",
            "Deal Value": amount,
            "Primary Contact": contact.get("email"),
            "Start Date": datetime.now().isoformat(),
            "HubSpot ID": deal_id
        },
        "content": f"""
# {company.get('name', deal_name)}

## Overview
- **Deal Closed**: {datetime.now().strftime('%B %d, %Y')}
- **Value**: ${amount:,.2f}
- **Primary Contact**: {contact.get('firstname', '')} {contact.get('lastname', '')}

## Onboarding Checklist
- [ ] Welcome email sent
- [ ] Kickoff call scheduled
- [ ] Workspace set up
- [ ] Project plan created
- [ ] First invoice sent

## Notes
*Add meeting notes, decisions, and important context here*

## Links
- [HubSpot Deal](https://app.hubspot.com/deals/{deal_id})
- [ClickUp Project]({project_result.get('project_url', '#')})
"""
    })
    results["actions_completed"].append({"action": "notion_page", "success": notion_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # 5. UPDATE CRM - Set lifecycle stage
    # ═══════════════════════════════════════════════════════════════

    if contact.get("id"):
        hubspot_result = trigger_skill("hubspot-contact-enrichment", {
            "action": "update",
            "contact_id": contact["id"],
            "properties": {
                "lifecyclestage": "customer",
                "hs_lead_status": "CONNECTED",
                "became_customer_date": datetime.now().isoformat()
            }
        })
        results["actions_completed"].append({"action": "hubspot_update", "success": hubspot_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # 6. SCHEDULE KICKOFF - Book meeting
    # ═══════════════════════════════════════════════════════════════

    if contact.get("email"):
        calendar_result = trigger_skill("calendar-smart-scheduling", {
            "action": "propose_meeting",
            "attendees": [contact["email"], deal.get("owner_email")],
            "title": f"Kickoff Call: {company.get('name', deal_name)}",
            "duration": 45,
            "preferred_days": 3,  # Within next 3 business days
            "description": f"Kickoff call to begin onboarding for {company.get('name')}."
        })
        results["actions_completed"].append({"action": "kickoff_scheduled", "success": calendar_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # SUMMARY
    # ═══════════════════════════════════════════════════════════════

    successful = sum(1 for a in results["actions_completed"] if a["success"])
    total = len(results["actions_completed"])

    results["summary"] = f"{successful}/{total} actions completed successfully"
    results["success"] = successful == total

    return_result(results)


def get_deal_details(deal_id):
    """Fetch deal from HubSpot."""
    creds = get_credential("hubspot")
    # ... API call
    pass

def get_associated_contact(deal_id):
    """Get primary contact for deal."""
    # ... API call
    pass

def get_associated_company(deal_id):
    """Get company for deal."""
    # ... API call
    pass
```

## Trigger Configuration

```yaml
trigger:
  type: webhook
  source: hubspot
  event: deal.propertyChange
  filter:
    property: dealstage
    value: closedwon
```

## Evolution History
- **v1**: Slack notification only
- **v2**: Added ClickUp project creation
- **v3**: Full workflow - email, Notion, calendar, CRM updates

## Learned Patterns
- Run independent tasks in parallel for speed
- Always include links between systems (HubSpot ↔ ClickUp ↔ Notion)
- Personalize communications using client data
- Track which actions succeeded for debugging
```

---

### Skill Examples: Meta Skills

#### skill-creator.md

```markdown
# Skill Creator

## Skill Metadata
- **ID**: skill-creator-v2
- **Category**: meta
- **Role Affinity**: all
- **Version**: 2
- **Executions**: 89
- **Success Rate**: 94.4%

## Description
Meta-skill that creates new skills. When the agent encounters a task it doesn't
have a skill for, this skill generates a new one.

## Script

```python
from sorx import return_result, call_llm, save_skill, get_pattern, validate_skill

def execute(params: dict):
    """
    Create a new skill from a task description.

    Steps:
    1. Analyze the task requirements
    2. Identify the interface type and provider
    3. Load the appropriate pattern template
    4. Generate the skill script
    5. Validate and save
    """
    task = params["task"]
    context = params.get("context", {})

    # Step 1: Analyze task
    analysis = call_llm(
        prompt=f"""Analyze this task and determine what's needed:

Task: {task}
Context: {context}

Return JSON with:
- skill_name: snake_case name for the skill
- category: crm|communication|tasks|documents|meetings|finance|workflows
- provider: The primary service (hubspot, gmail, clickup, etc.)
- interface_type: rest_api|graphql|database|desktop|file_system
- description: What this skill does
- required_credentials: List of credential keys needed
- inputs: List of expected input parameters
- outputs: What the skill returns
"""
    )

    # Step 2: Load pattern template
    pattern = get_pattern(
        interface_type=analysis["interface_type"],
        provider=analysis["provider"]
    )

    # Step 3: Generate skill script
    script = call_llm(
        prompt=f"""Generate a Sorx skill script.

Skill: {analysis['skill_name']}
Description: {analysis['description']}
Provider: {analysis['provider']}
Inputs: {analysis['inputs']}
Outputs: {analysis['outputs']}

Use this pattern as a template:
{pattern}

Requirements:
- Use get_credential() for auth, never hardcode
- Use return_result() for all outputs
- Include proper error handling
- Add docstring explaining usage
- Handle common edge cases

Return only the Python code.
"""
    )

    # Step 4: Validate
    validation = validate_skill(script, analysis)

    if not validation["valid"]:
        # Try to fix issues
        script = call_llm(
            prompt=f"""Fix these issues in the skill:

Script:
{script}

Issues:
{validation['issues']}

Return the corrected code only.
"""
        )
        validation = validate_skill(script, analysis)

    if validation["valid"]:
        # Step 5: Save skill
        skill = save_skill(
            name=analysis["skill_name"],
            category=analysis["category"],
            provider=analysis["provider"],
            description=analysis["description"],
            script=script,
            credentials=analysis["required_credentials"],
            version=1
        )

        return_result({
            "success": True,
            "skill_id": skill["id"],
            "skill_name": analysis["skill_name"],
            "message": f"Created new skill: {analysis['skill_name']}"
        })
    else:
        return_result({
            "success": False,
            "error": "Could not generate valid skill",
            "issues": validation["issues"]
        })
```

## Example Usage

**User Request:**
"I need to check our Stripe balance and send a Slack message if it's below $10,000"

**Skill Creator Output:**
```python
# Generated skill: stripe-balance-alert-v1

from sorx import get_credential, return_result, trigger_skill

def execute(params: dict):
    """
    Check Stripe balance and alert if below threshold.

    Params:
        threshold: Minimum balance (default: 10000)
        channel: Slack channel for alerts (default: #finance)
    """
    creds = get_credential("stripe")
    threshold = params.get("threshold", 10000)

    response = requests.get(
        "https://api.stripe.com/v1/balance",
        headers={"Authorization": f"Bearer {creds['secret_key']}"}
    )

    if response.ok:
        balance = response.json()
        available = sum(b["amount"] for b in balance["available"]) / 100

        if available < threshold:
            trigger_skill("slack-team-notification", {
                "channel": params.get("channel", "#finance"),
                "type": "warning",
                "title": "Low Stripe Balance Alert",
                "message": f"Current balance: ${available:,.2f}",
                "fields": {"Threshold": f"${threshold:,}"}
            })

        return_result({
            "success": True,
            "balance": available,
            "alert_sent": available < threshold
        })
    else:
        return_result({"success": False, "error": response.text})
```

## Evolution History
- **v1**: Basic skill generation
- **v2**: Added validation, auto-fix, pattern loading
```

---

## Skill Library

A shared repository of learned skills that can be discovered and reused:

### Skill Discovery

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            SKILL LIBRARY                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Search: [hubspot deal notification____________] [Search]                  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ hubspot_deal_stage_notification                                      │  │
│   │ ─────────────────────────────────                                    │  │
│   │ Sends Slack notification when HubSpot deal changes stage            │  │
│   │                                                                       │  │
│   │ Version: 5  │  Executions: 12,847  │  Success: 99.2%                │  │
│   │ Author: BusinessOS Team  │  Rating: 4.8/5                           │  │
│   │                                                                       │  │
│   │ Requires: hubspot, slack                                             │  │
│   │ [Install] [View Code] [Fork]                                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ hubspot_deal_to_clickup_task                                         │  │
│   │ ────────────────────────────                                         │  │
│   │ Creates ClickUp task when HubSpot deal closes                       │  │
│   │                                                                       │  │
│   │ Version: 3  │  Executions: 3,241  │  Success: 97.8%                 │  │
│   │ Author: Community  │  Rating: 4.5/5                                 │  │
│   │                                                                       │  │
│   │ Requires: hubspot, clickup                                           │  │
│   │ [Install] [View Code] [Fork]                                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ hubspot_closed_deal_summary_email                                    │  │
│   │ ───────────────────────────────                                      │  │
│   │ Sends daily summary email of closed deals                           │  │
│   │                                                                       │  │
│   │ Version: 2  │  Executions: 892  │  Success: 98.1%                   │  │
│   │ Author: Community  │  Rating: 4.2/5                                 │  │
│   │                                                                       │  │
│   │ Requires: hubspot, gmail                                             │  │
│   │ [Install] [View Code] [Fork]                                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   Categories: [All] [CRM] [Communication] [Tasks] [Files] [Custom]        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Skill Publishing

When a skill reaches "mastery" (high execution count + success rate), users can publish to the library:

```python
class SkillPublisher:
    def can_publish(self, skill: Skill) -> bool:
        """Check if skill meets publishing criteria."""
        return (
            skill.executions >= 50 and
            skill.success_rate >= 0.95 and
            skill.version >= 2  # Has been improved at least once
        )

    def publish(self, skill: Skill, user: User) -> PublishedSkill:
        """Publish skill to community library."""
        # Security review
        self.security_review(skill)

        # Remove user-specific data
        sanitized = self.sanitize_skill(skill)

        # Publish
        return PublishedSkill(
            skill=sanitized,
            author=user,
            published_at=now()
        )
```

---

## Comparison: Sorx 2.0 vs Composio vs Traditional

| Feature | Composio | Traditional MCP | Sorx 2.0 |
|---------|----------|-----------------|----------|
| **Tool Count** | 250+ pre-built | Custom built | Unlimited (generated) |
| **New Integrations** | Wait for dev team | Build yourself | Agent learns on-demand |
| **Legacy Systems** | Limited | Build yourself | Universal adapters |
| **Credential Security** | Cloud-stored | Context-exposed | Local vault only |
| **Learning** | None | None | Skills improve over time |
| **Customization** | Limited config | Full code | Auto-customized |
| **Context Usage** | Tool definitions | Tool definitions | Minimal patterns |
| **Enterprise Systems** | Some | Manual | Full support |
| **Desktop Apps** | No | No | Yes (OS automation) |
| **Hardware/IoT** | No | No | Yes |
| **Offline Capable** | No | No | Yes |
| **Skill Sharing** | No | No | Community library |
| **Small Business** | Works | Complex | Efficient |
| **Enterprise** | Works | Complex | Carrier-grade |

---

## Enterprise Features

### Carrier-Grade Reliability

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      ENTERPRISE RELIABILITY FEATURES                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   HIGH AVAILABILITY                                                         │
│   - Multi-region Sorx Engine deployment                                    │
│   - Automatic failover                                                      │
│   - Zero-downtime updates                                                   │
│                                                                             │
│   SECURITY                                                                  │
│   - SOC 2 Type II compliant                                                │
│   - GDPR compliant                                                          │
│   - End-to-end encryption                                                   │
│   - Credential vault with HSM backing                                      │
│   - Audit logging for compliance                                           │
│                                                                             │
│   SCALABILITY                                                               │
│   - Horizontal scaling of skill execution                                  │
│   - Rate limit handling                                                     │
│   - Queue-based execution for high volume                                  │
│                                                                             │
│   MONITORING                                                                │
│   - Real-time skill performance metrics                                    │
│   - Alerting on failures                                                   │
│   - Detailed execution logs                                                │
│                                                                             │
│   GOVERNANCE                                                                │
│   - Skill approval workflows                                               │
│   - Access control by role                                                 │
│   - Data classification enforcement                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Small Business Efficiency

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      SMALL BUSINESS EFFICIENCY                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ZERO CONFIGURATION                                                        │
│   - Skills generated automatically                                         │
│   - No developer required                                                   │
│   - Natural language requests                                              │
│                                                                             │
│   COST EFFICIENT                                                            │
│   - Only pay for what you use                                              │
│   - No pre-built tool licensing                                            │
│   - Skills reused across requests                                          │
│                                                                             │
│   FAST TIME TO VALUE                                                        │
│   - First skill in minutes                                                 │
│   - No integration project                                                 │
│   - Immediate automation                                                   │
│                                                                             │
│   SIMPLE MANAGEMENT                                                         │
│   - Single dashboard                                                       │
│   - Clear skill inventory                                                  │
│   - Easy credential management                                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Implementation Priority

### Phase 1: Core Engine (CRITICAL)

1. Sorx Engine (Python sidecar)
2. Credential Vault
3. REST API Adapter
4. Basic skill generation
5. Skill storage and retrieval

### Phase 2: Provider Patterns (HIGH)

Priority based on planned integrations:
1. Google (Gmail, Drive, Calendar)
2. Slack
3. Notion
4. HubSpot
5. ClickUp, Asana

### Phase 3: Skill Learning (HIGH)

1. Skill versioning
2. Evolution based on feedback
3. Success tracking
4. Performance metrics

### Phase 4: Universal Adapters (MEDIUM)

1. Database adapter
2. Legacy system adapter (SOAP)
3. Desktop automation adapter
4. File system adapter

### Phase 5: Skill Library (MEDIUM)

1. Skill publishing
2. Discovery/search
3. Installation
4. Ratings/reviews

### Phase 6: Enterprise Features (MEDIUM)

1. SOC 2 compliance
2. Audit logging
3. Access control
4. Governance workflows

---

## Summary

**Sorx 2.0** is not just an integration platform - it's a **skill-based learning system** where AI agents:

1. **Learn** new skills on-demand by generating code from patterns
2. **Execute** skills locally with secure credential access
3. **Improve** skills over time based on feedback and experience
4. **Share** mastered skills with other agents and organizations
5. **Connect** to ANYTHING - APIs, databases, legacy systems, desktop apps, hardware

This is the **USB-C of AI integrations** - one universal interface that works with any system, learns from experience, and gets better over time.

---

## Related Documents

- `INTEGRATION_IMPLEMENTATION_PLAN.md` - Traditional integrations for data sync
- `INTEGRATION_INFRASTRUCTURE.md` - Backend architecture patterns
- `INTEGRATIONS_MASTER_LIST.md` - Complete provider inventory

---

**Sorx 2.0: System of Reasoning - Where AI Agents Learn to Connect**
