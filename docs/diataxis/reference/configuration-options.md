# Reference: Configuration Options

**Information-oriented.** All BusinessOS configuration via environment variables. Look up by variable name.

**Format:** Variable → Type → Default → Purpose

---

## Environment & Server

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `ENVIRONMENT` | string | `development` | `development`, `staging`, or `production` |
| `SERVER_PORT` | string | `8001` | HTTP server listen port |
| `BASE_URL` | string | `http://localhost:5173` | Frontend base URL (for redirects, CORS) |
| `DATABASE_REQUIRED` | bool | `false` | Fail startup if database unavailable |

---

## Database (PostgreSQL)

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `DATABASE_URL` | string | Required | PostgreSQL connection string: `postgres://user:pass@host:5432/dbname` |

**Format:** `postgres://[user[:password]@][host][:port][/dbname][?param=value]`

**Example:** `postgres://postgres:password@localhost:5432/businessos`

---

## Redis (Cache & Session Store)

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `REDIS_URL` | string | `redis://localhost:6379` | Redis connection string |
| `REDIS_PASSWORD` | string | (empty) | Redis password (if auth enabled) |
| `REDIS_TLS_ENABLED` | bool | `false` | Enable TLS for Redis (production recommended) |
| ⚠️ `REDIS_KEY_HMAC_SECRET` | string | (empty) | **CRITICAL:** 32-byte secret for Redis key derivation. Prevents token enumeration attacks. **Must be set in production.** Generate: `openssl rand -base64 32` |

**Redis URL Format:** `redis://[:password@]host[:port][/db]`

**Example:** `redis://:mypassword@localhost:6379/0`

---

## Authentication & Security

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `SECRET_KEY` | string | Required | JWT signing secret (min 32 bytes, random) |
| `ALGORITHM` | string | `HS256` | JWT algorithm: `HS256` (HMAC), `RS256` (RSA) |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | int | `60` | JWT expiration time in minutes |
| ⚠️ `TOKEN_ENCRYPTION_KEY` | string | (empty) | **CRITICAL:** 32-byte base64-encoded key for encrypting OAuth tokens in database. **Must be set in production.** Generate: `openssl rand -base64 32` |

**How to generate SECRET_KEY:**
```bash
openssl rand -base64 32
```

---

## Google OAuth

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `GOOGLE_CLIENT_ID` | string | (empty) | OAuth app client ID from Google Console |
| `GOOGLE_CLIENT_SECRET` | string | (empty) | OAuth app client secret from Google Console |
| `GOOGLE_REDIRECT_URI` | string | (empty) | Login redirect: `https://yourdomain.com/api/v1/auth/callback/google` |
| `GOOGLE_INTEGRATION_REDIRECT_URI` | string | (empty) | Integration redirect: `https://yourdomain.com/api/v1/integrations/google/callback` |

**Setup:** See Google Console → OAuth 2.0 → Create credentials

---

## Microsoft OAuth

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `MICROSOFT_CLIENT_ID` | string | (empty) | Azure AD app client ID |
| `MICROSOFT_CLIENT_SECRET` | string | (empty) | Azure AD app client secret |
| `MICROSOFT_REDIRECT_URI` | string | (empty) | Redirect: `https://yourdomain.com/api/v1/auth/callback/microsoft` |

---

## Slack OAuth

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `SLACK_CLIENT_ID` | string | (empty) | Slack app client ID |
| `SLACK_CLIENT_SECRET` | string | (empty) | Slack app client secret |
| `SLACK_REDIRECT_URI` | string | (empty) | Redirect: `https://yourdomain.com/api/v1/integrations/slack/callback` |

---

## Notion OAuth

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `NOTION_CLIENT_ID` | string | (empty) | Notion integration client ID |
| `NOTION_CLIENT_SECRET` | string | (empty) | Notion integration client secret |
| `NOTION_REDIRECT_URI` | string | (empty) | Redirect: `https://yourdomain.com/api/v1/integrations/notion/callback` |

---

## HubSpot OAuth

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `HUBSPOT_CLIENT_ID` | string | (empty) | HubSpot app client ID |
| `HUBSPOT_CLIENT_SECRET` | string | (empty) | HubSpot app client secret |
| `HUBSPOT_REDIRECT_URI` | string | (empty) | Redirect: `https://yourdomain.com/api/v1/integrations/hubspot/callback` |

---

## Linear OAuth

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `LINEAR_CLIENT_ID` | string | (empty) | Linear app client ID |
| `LINEAR_CLIENT_SECRET` | string | (empty) | Linear app client secret |
| `LINEAR_REDIRECT_URI` | string | (empty) | Redirect: `https://yourdomain.com/api/v1/integrations/linear/callback` |

---

## AI Provider Configuration

### Provider Selection

| Variable | Type | Default | Options |
|----------|------|---------|---------|
| `AI_PROVIDER` | string | `ollama_local` | `ollama_local`, `ollama_cloud`, `anthropic`, `groq`, `openai`, `xai` |

### Ollama Local

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `OLLAMA_LOCAL_URL` | string | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_MODE` | string | `local` | Legacy config (use `AI_PROVIDER` instead) |
| `DEFAULT_MODEL` | string | `mistral` | Default local model name |

**Common Models:** `mistral`, `llama2`, `neural-chat`, `orca-mini`

### Ollama Cloud (api.ollama.com)

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `OLLAMA_CLOUD_API_KEY` | string | (empty) | Ollama Cloud API key |
| `OLLAMA_CLOUD_MODEL` | string | `mistral` | Model name on Ollama Cloud |

### Anthropic (Claude)

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `ANTHROPIC_API_KEY` | string | (empty) | Claude API key from Anthropic Console |
| `ANTHROPIC_MODEL` | string | `claude-3-sonnet-20240229` | Model: `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku` |
| `ANTHROPIC_BASE_URL` | string | `https://api.anthropic.com/v1` | Base URL (override for proxies) |

### Groq

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `GROQ_API_KEY` | string | (empty) | Groq API key |
| `GROQ_MODEL` | string | `openai/gpt-oss-20b` | Model name |

### OpenAI

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `OPENAI_API_KEY` | string | (empty) | OpenAI API key |
| `OPENAI_MODEL` | string | `gpt-4-turbo` | Model: `gpt-4-turbo`, `gpt-4`, `gpt-3.5-turbo` |

### xAI (Grok)

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `XAI_API_KEY` | string | (empty) | xAI API key for Grok |
| `XAI_MODEL` | string | `grok-beta` | Model name |
| `XAI_BASE_URL` | string | `https://api.x.ai` | Base URL |

---

## Supermemory Integration

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `SUPERMEMORY_API_KEY` | string | (empty) | Supermemory API key for long-term knowledge storage |

---

## Process Mining (pm4py-rust Integration)

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `PM4PY_RUST_URL` | string | `http://localhost:8090` | pm4py-rust gateway URL |

---

## Logging

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `LOG_LEVEL` | string | `info` | Log level: `debug`, `info`, `warn`, `error` |

**Implementation:** Uses `slog` (Go structured logging)

---

## CORS & Security

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `COOKIE_DOMAIN` | string | (development only) | CORS cookie domain in dev mode |
| `COOKIE_SECURE` | bool | `true` (prod), `false` (dev) | HTTPS-only cookies |

**CSRF Protection:** Built-in via middleware, exceptions for webhooks.

---

## SOC2 Compliance ⚠️

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `COMPLIANCE_FRAMEWORK` | string | `SOC2` | Active framework: `SOC2`, `HIPAA`, `GDPR`, `SOX`, `CUSTOM` |
| `AUDIT_LOG_RETENTION_DAYS` | int | `365` | Audit log retention period |
| `ENCRYPTION_AT_REST_ENABLED` | bool | `true` (prod) | Enable database encryption at rest |

---

## Feature Flags

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `ENABLE_A2A` | bool | `true` | Enable agent-to-agent communication |
| `ENABLE_MCP` | bool | `true` | Enable Model Context Protocol support |
| `ENABLE_REAL_TIME_SYNC` | bool | `true` | Enable real-time data synchronization |

---

## Performance & Limits

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `MAX_CONCURRENT_PM4PY_REQUESTS` | int | `10` | Semaphore limit for pm4py-rust requests |
| `PM4PY_CONNECTION_POOL_SIZE` | int | `100` | HTTP connection pool size |
| `MAX_SSE_SUBSCRIBERS` | int | `1000` | Max concurrent SSE stream subscribers |
| `REQUEST_TIMEOUT_SECONDS` | int | `30` | HTTP request timeout |
| `DATABASE_MAX_OPEN_CONNS` | int | `25` | PostgreSQL connection pool size |
| `DATABASE_MAX_IDLE_CONNS` | int | `5` | PostgreSQL idle connections kept alive |

---

## Example .env File

```bash
# Environment
ENVIRONMENT=production
SERVER_PORT=8001
BASE_URL=https://app.yourdomain.com
DATABASE_REQUIRED=true

# Database
DATABASE_URL=postgres://user:password@db.example.com:5432/businessos

# Redis
REDIS_URL=redis://cache.example.com:6379/0
REDIS_PASSWORD=your_redis_password
REDIS_TLS_ENABLED=true
REDIS_KEY_HMAC_SECRET=your_base64_encoded_32_byte_secret

# Security
SECRET_KEY=your_base64_encoded_32_byte_secret
ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=60
TOKEN_ENCRYPTION_KEY=your_base64_encoded_32_byte_encryption_key

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=https://app.yourdomain.com/api/v1/auth/callback/google

# AI Provider
AI_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-3-sonnet-20240229

# Process Mining
PM4PY_RUST_URL=http://pm4py-rust:8090

# Logging
LOG_LEVEL=info

# Compliance
COMPLIANCE_FRAMEWORK=SOC2
AUDIT_LOG_RETENTION_DAYS=365
ENCRYPTION_AT_REST_ENABLED=true

# Performance
MAX_CONCURRENT_PM4PY_REQUESTS=10
DATABASE_MAX_OPEN_CONNS=25
```

---

## Loading Configuration

1. **Default values:** Hardcoded in code
2. **.env file:** Loaded from project root if exists (dev only)
3. **Environment variables:** Override .env values
4. **Validation:** Config validated at startup (`Validate()` method)

**Priority:** Environment Variables > .env > Defaults

---

## See Also

- How-to: Set up OAuth
- How-to: Configure AI provider
- How-to: Enable SOC2 compliance
- Error codes & troubleshooting
