# API Patterns Visual Guide

## Function Flow Diagram

```
USER CODE
    │
    ├──────────────────────────────────┐
    │                                  │
    ▼                                  ▼
API FUNCTION                    API FUNCTION
(Standard Request)              (Streaming Request)
    │                                  │
    ├─ Use request<T>()        └─ Use custom fetch()
    │                                  │
    ▼                                  ▼
REQUEST WRAPPER            FETCH DIRECTLY
(frontend/src/lib/api/base.ts)   │
    │                            │
    ├─ Set Content-Type         ├─ Manual headers
    ├─ Add credentials          ├─ Manual credentials
    ├─ Stringify body           ├─ Manual body stringify
    ├─ Fetch from API_BASE      ├─ Fetch from getApiBaseUrl()
    │                            │
    ▼                            ▼
HTTP REQUEST                 HTTP REQUEST
(GET/POST/PUT/DELETE)       (POST with stream)
    │                            │
    ├─ Success (2xx)            ├─ Success (2xx)
    │   └─ Parse JSON           │   └─ Return response.body
    │       └─ Return T         │       └─ Return ReadableStream
    │                            │
    └─ Error (4xx/5xx)          └─ Error (4xx/5xx)
        └─ Parse error          └─ Parse error
            └─ Throw Error          └─ Throw Error

CALLER CODE
    │
    ├─ Receives T               ├─ Receives ReadableStream
    ├─ Can access properties    ├─ Must read chunks
    └─ Fully type-safe          └─ Fully type-safe
```

## Function Naming Structure

```
        ┌────────────────────────────────────┐
        │   Function Name Structure          │
        ├────────────────────────────────────┤
        │   [VERB][RESOURCE]([params])       │
        └────────────────────────────────────┘
             │           │         │
             │           │         └─ Parameters (ID, data, options)
             │           │
             │           └─ Resource type
             │             (Agent, Model, Provider, etc)
             │
             └─ Action verb
               (get, create, save, update, delete, execute, etc)
```

### Verb Categories

```
READ VERBS:
┌────────┐
│ get    │ ──> Fetch item(s) from server
└────────┘

CREATE VERBS:
┌────────┐
│ create │ ──> Make new item
│ save   │ ──> Store configuration
│ import │ ──> Load from external
└────────┘

ACTION VERBS:
┌────────┐
│execute │ ──> Run/process
│warmup  │ ──> Initialize
│pull    │ ──> Download/fetch
│test    │ ──> Validate
│clone   │ ──> Duplicate
└────────┘

MODIFY VERBS:
┌────────┐
│ update │ ──> Change existing
└────────┘

DELETE VERBS:
┌────────┐
│ delete │ ──> Remove item(s)
└────────┘
```

## Request Pattern Decision Tree

```
                        START: Need API call
                                 │
                    ┌────────────┴────────────┐
                    │                         │
              Fetching data?            Modifying data?
                    │                         │
                    ▼                         ▼
              GET Request                POST/PUT/DELETE
                    │                         │
        ┌───────────┼───────────┐    ┌───────┴────────┐
        │           │           │    │                │
      Single     Multiple    Filters  Create    Update   Delete
       item      items       needed?   new      existing  item(s)
        │           │           │      │           │       │
        ▼           ▼           ▼      ▼           ▼       ▼
    /resource   /resources  ?param  POST body   PUT body  DELETE
    /id                      for     with data   with     (only ID
                           filtering  data       updates   needed)
                                      │           │
                                      ├─ New      └─ Omit: id,
                                      │   fields    user_id,
                                      │   only      timestamps
                                      │
                                      └─ Omit: id,
                                          user_id,
                                          timestamps
```

## Error Handling Flow Diagram

```
API CALL
    │
    ▼
Is it a STREAMING endpoint?
    │
    ├─ YES ──────────────────────┐
    │                            │
    │                            ▼
    │                      Custom fetch()
    │                            │
    │                            ▼
    │                      response.ok?
    │                            │
    │                      ┌─────┴─────┐
    │                      │           │
    │                    YES          NO
    │                      │           │
    │                      ▼           ▼
    │                 Return      Parse error
    │              response.body  response.json()
    │                      │           │
    │                      └─────┬─────┘
    │                            │
    │                            ▼
    │                      Throw Error
    │
    └─ NO ───────────────────────┐
                                 │
                                 ▼
                            request<T>()
                                 │
                                 ▼
                            response.ok?
                                 │
                         ┌───────┴───────┐
                         │               │
                       YES             NO
                         │               │
                         ▼               ▼
                    response.json() Parse error
                         │        response.json()
                         ▼        │
                    Return T      └─┬────────┐
                    (type-safe)     │        │
                                    ▼        ▼
                                Catch error
                                Fallback msg
                                    │
                                    ▼
                                Throw Error
                                (formatted)
```

## Type Safety Flow

```
VARIABLE DECLARATION
        │
        ▼
declare result: ? (unknown)
        │
┌───────┴──────────┐
│                  │
▼                  ▼
Use request<T>    Use custom fetch
        │                 │
        └────────┬────────┘
                 │
        ┌────────▼────────┐
        │                 │
        ▼                 ▼
    Specify T      Return type annotation
    (generic)      (ReadableStream or other)
        │                 │
        └────────┬────────┘
                 │
                 ▼
            result: T (fully typed)
                 │
                 ▼
            ACCESS WITH CONFIDENCE
            ✓ IDE autocomplete works
            ✓ Type checking enforced
            ✓ No runtime surprises
            ✓ Compiler catches mistakes
```

## Response Type Hierarchy

```
RESPONSE TYPES
├── SINGLE ITEM RESPONSE
│   └─ CustomAgent
│       ├─ id: string
│       ├─ name: string
│       ├─ system_prompt: string
│       └─ ... more fields
│
├── ARRAY WRAPPER RESPONSE
│   └─ { agents: CustomAgent[] }
│       ├─ agents
│       │   ├─ [0]: CustomAgent
│       │   ├─ [1]: CustomAgent
│       │   └─ ...
│
├── STATUS MESSAGE RESPONSE
│   └─ { message: string }
│       └─ message: "Operation successful"
│
├── COMPLEX RESPONSE
│   └─ { agents: CustomAgent[]; total: number; page: number }
│       ├─ agents: CustomAgent[]
│       ├─ total: number
│       └─ page: number
│
└── DISCRIMINATED UNION RESPONSE
    └─ Success | Failure
        ├─ Success
        │   └─ { success: true; result: string; error?: never }
        │
        └─ Failure
            └─ { success: false; result?: never; error: string }
```

## TypeScript Pattern Matching

```
PATTERN                        CODE
─────────────────────────────────────────────────────────

Single object response:
  Get one item               const agent = await getCustomAgent(id)
                            // agent: CustomAgent

Array of objects:
  Get many items            const { agents } = await getCustomAgents()
                            // agents: CustomAgent[]

Status message:
  Simple confirmation       const { message } = await deleteAgent(id)
                            // message: string

Optional fields:
  Some fields null/undefined interface Agent {
                              id: string;         // Required
                              avatar?: string;    // Optional
                            }

Partial updates:
  Only changed fields       Partial<Omit<Agent, 'id'>>
                            // Can update any field except id

Discriminated union:
  Multiple outcomes         if (result.success) {
  (type-safe)               result.result;  // Available
                            } else {
                              result.error;   // Available
                            }
```

## Endpoint Structure Map

```
API BASE: http://localhost:8001/api

/ai/
├── /providers                    GET list all providers
├── /provider                     PUT change active provider
├── /models
│   ├── /                         GET all models
│   ├── /local                    GET local models only
│   ├── /pull                     POST download model (stream)
│   └── /warmup                   POST initialize model
├── /system                       GET system info
├── /api-key                      POST save API key
├── /agents
│   ├── /                         GET all agents
│   └── /:id                      GET specific agent
├── /custom-agents
│   ├── /                         GET all (GET creates data, POST updates)
│   ├── /?include_inactive=true   GET with filters
│   ├── /                         POST create new agent
│   ├── /:id                      GET agent details
│   ├── /:id                      PUT update agent
│   ├── /:id                      DELETE delete agent
│   ├── /:id
│   │   ├── /execute             POST run agent
│   │   ├── /execute/stream       POST run with stream
│   │   ├── /clone               POST duplicate agent
│   │   ├── /stats               GET usage stats
│   │   └── /export              GET export config
│   ├── /test                     POST validate config
│   ├── /execute-batch            POST run multiple
│   ├── /delete-batch             POST delete multiple
│   └── /import                   POST import config

/mcp/
├── /tools                        GET available tools
└── /execute                      POST run a tool
```

## Request Construction Template

```
┌───────────────────────────────────────────────────────┐
│ REQUEST CONSTRUCTION                                  │
├───────────────────────────────────────────────────────┤
│                                                       │
│  export async function [NAME](params) {              │
│    return request<ReturnType>(                       │
│      '/endpoint/path',  ◄──── Endpoint path          │
│      {                                               │
│        method: 'GET|POST|PUT|DELETE',  ◄─ HTTP verb │
│        body: {  ◄──────────────────── Request body   │
│          field1: value1,    (snake_case for backend) │
│          field2: value2                              │
│        }                                             │
│      }                                               │
│    );                                                │
│  }                                                   │
│                                                       │
│  ⚠️ method defaults to 'GET' if omitted              │
│  ⚠️ body auto-stringified to JSON                    │
│  ⚠️ Content-Type auto-added if body present          │
│  ⚠️ credentials: 'include' auto-added                │
│                                                       │
└───────────────────────────────────────────────────────┘
```

## Creating vs Updating Patterns

```
CREATE (POST)
    │
    ├─ Endpoint: /resource
    ├─ Method: POST
    ├─ Body: NEW DATA (all fields except auto-generated)
    │   Omit: id, user_id, created_at, updated_at
    │
    └─ Response: Full object with id + timestamps

UPDATE (PUT)
    │
    ├─ Endpoint: /resource/:id
    ├─ Method: PUT
    ├─ Body: PARTIAL DATA (only fields being changed)
    │   Partial<Omit<T, auto-generated>>
    │
    └─ Response: Updated full object
```

## Common Method Patterns at a Glance

```
ACTION              METHOD    ENDPOINT          BODY
────────────────────────────────────────────────────────
Get all items       GET       /resource         (none)
Get single item     GET       /resource/:id     (none)
Get filtered        GET       /resource?filter  (none)
Create new          POST      /resource         new data
Update item         PUT       /resource/:id     partial data
Delete item         DELETE    /resource/:id     (none)
Execute action      POST      /resource/action  input data
Stream operation    POST      /resource/stream  input data
Validate config     POST      /resource/test    config data
Batch operation     POST      /resource/batch   { ids, data }
```

## Error Message Format

```
┌─────────────────────────────────────────┐
│ ERROR MESSAGE FORMAT                    │
├─────────────────────────────────────────┤
│                                         │
│ "[ERROR_MESSAGE] (HTTP [STATUS_CODE])" │
│                                         │
│ Examples:                               │
│  • "Not found (HTTP 404)"               │
│  • "Unauthorized (HTTP 401)"            │
│  • "Server error (HTTP 500)"            │
│  • "Bad request (HTTP 400)"             │
│                                         │
│ Always includes status code             │
│ Always includes descriptive message     │
│                                         │
└─────────────────────────────────────────┘
```

## Decision Matrix: Which Pattern to Use?

```
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║  NEED                    ENDPOINT           METHOD    PATTERN  ║
║  ────────────────────────────────────────────────────────────  ║
║  Fetch data              /resource          GET       request  ║
║  Create new              /resource          POST      request  ║
║  Change existing         /resource/:id      PUT       request  ║
║  Remove item             /resource/:id      DELETE    request  ║
║  Run action              /resource/action   POST      request  ║
║  Long operation          /resource/stream   POST      fetch    ║
║  Multiple items          /resource/batch    POST      request  ║
║  Optional filters        ?param=value       -         append   ║
║  Safe multiple results   See union types    -         union    ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
```

## Quick Visual Checklist for New Endpoint

```
☐ Function Name:
  [ ] Starts with verb (get, create, update, delete, execute)
  [ ] Contains resource name (Agent, Model, etc)
  [ ] Follows camelCase

☐ Type Definition:
  [ ] Specify generic type <T>
  [ ] Import from types.ts if needed
  [ ] Use Omit<> for create operations
  [ ] Use Partial<> for updates

☐ Request:
  [ ] Correct HTTP method
  [ ] Correct endpoint path
  [ ] Body in snake_case for backend
  [ ] Include all required parameters

☐ Response:
  [ ] Generic type matches actual response
  [ ] Error handling in place
  [ ] TypeScript inference works

☐ Documentation:
  [ ] JSDoc with description
  [ ] Parameter types documented
  [ ] Return type documented
  [ ] Usage example provided

☐ Testing:
  [ ] Test happy path
  [ ] Test error case
  [ ] Verify type safety
  [ ] Check error message format
```

