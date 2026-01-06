# BusinessOS Integrations - Phased Development Plan

## Executive Summary

**Goal:** Build a robust integration system where:
1. Users connect external tools (OAuth, API keys)
2. Data is imported and synced from connected services
3. Sorx agents use credentials to execute skills against external APIs
4. Everything flows through a secure credential vault

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         INTEGRATIONS ARCHITECTURE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  USER LAYER                                                                 │
│  ═══════════                                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     BusinessOS Frontend                              │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │   │
│  │  │ Integrations│  │  Settings   │  │   Modules   │                 │   │
│  │  │    Page     │  │   Panel     │  │ (use data)  │                 │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  GATEWAY LAYER                                                              │
│  ═════════════                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     BusinessOS Backend (Go)                          │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │   │
│  │  │   OAuth     │  │ Integration │  │   Sorx      │                 │   │
│  │  │  Handlers   │  │  Handlers   │  │  Handlers   │                 │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘                 │   │
│  │         │                │                │                         │   │
│  │         ▼                ▼                ▼                         │   │
│  │  ┌─────────────────────────────────────────────────────────────┐   │   │
│  │  │              CREDENTIAL VAULT (Encrypted)                    │   │   │
│  │  │  • OAuth tokens (AES-256 encrypted)                         │   │   │
│  │  │  • API keys (AES-256 encrypted)                             │   │   │
│  │  │  • Refresh tokens with auto-rotation                        │   │   │
│  │  └─────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  EXECUTION LAYER                                                            │
│  ═══════════════                                                            │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        Sorx Engine                                   │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │   │
│  │  │   Skills    │  │  Workflows  │  │   Actions   │                 │   │
│  │  │  Registry   │  │   Engine    │  │  Executor   │                 │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘                 │   │
│  │         │                │                │                         │   │
│  │         └────────────────┴────────────────┘                         │   │
│  │                          │                                          │   │
│  │                    get_credential()                                 │   │
│  │                          │                                          │   │
│  │                          ▼                                          │   │
│  │  ┌─────────────────────────────────────────────────────────────┐   │   │
│  │  │           INTERFACE ADAPTERS (per connection type)           │   │   │
│  │  │  REST │ GraphQL │ Database │ Legacy │ Desktop │ Hardware    │   │   │
│  │  └─────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  EXTERNAL LAYER                                                             │
│  ══════════════                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Slack │ Google │ Notion │ HubSpot │ Linear │ GitHub │ ... 40+     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Database Model (Current State)

### Existing Tables (Migration 025)

```sql
-- 1. Provider catalog (system-defined)
integration_providers
├── id (VARCHAR PRIMARY KEY)        -- e.g., 'slack', 'google_calendar'
├── name                            -- Display name
├── description
├── category                        -- communication, crm, tasks, etc.
├── oauth_config (JSONB)           -- OAuth endpoints, scopes
├── modules[]                       -- Which modules can use this
├── skills[]                        -- Available Sorx skills
└── status                          -- available, coming_soon, beta

-- 2. User connections
user_integrations
├── id (UUID)
├── user_id → user(id)
├── provider_id → integration_providers(id)
├── status                          -- connected, disconnected, expired
├── access_token_encrypted (BYTEA)  -- ⚠️ NOT ACTUALLY ENCRYPTED YET
├── refresh_token_encrypted (BYTEA) -- ⚠️ NOT ACTUALLY ENCRYPTED YET
├── token_expires_at
├── scopes[]
├── external_account_id
├── external_account_name
├── external_workspace_id
├── external_workspace_name
├── settings (JSONB)
└── UNIQUE(user_id, provider_id)

-- 3. Per-module settings
module_integration_settings
├── user_id
├── module_id                       -- dashboard, tasks, projects, etc.
├── provider_id
├── enabled
├── sync_direction                  -- import, export, bidirectional
└── sync_frequency                  -- realtime, hourly, daily, manual

-- 4. Human-in-the-loop decisions
pending_decisions
├── execution_id
├── skill_id
├── question
├── options[]
├── status                          -- pending, decided, expired

-- 5. Sync tracking
integration_sync_log
├── user_integration_id
├── sync_type
├── direction
├── records_processed/created/updated/failed

-- 6. Skill execution tracking
skill_executions
├── skill_id
├── user_id
├── status                          -- pending, running, complete, failed
├── params, result, error
└── metrics
```

### Missing Tables (Need to Add)

```sql
-- 7. Credential Vault (encrypted token storage)
credential_vault
├── id (UUID)
├── user_id
├── provider_id
├── credential_type                 -- oauth, api_key, custom
├── encrypted_data (BYTEA)          -- AES-256-GCM encrypted
├── encryption_key_id               -- Reference to key in KMS
├── expires_at
├── last_rotated_at
└── metadata (JSONB)

-- 8. Webhook registrations
integration_webhooks
├── id (UUID)
├── user_integration_id
├── provider_id
├── webhook_url
├── webhook_secret_encrypted
├── events[]                        -- e.g., ['deal.closed', 'task.created']
├── status                          -- active, paused, failed
└── last_triggered_at

-- 9. Data sync mappings
data_sync_mappings
├── id (UUID)
├── user_id
├── source_provider                 -- e.g., 'hubspot'
├── source_entity                   -- e.g., 'contact'
├── target_module                   -- e.g., 'clients'
├── field_mappings (JSONB)         -- {source_field: target_field}
├── transform_rules (JSONB)
└── enabled
```

---

## Phase 0: Security Hardening (BLOCKING)

**Status:** MUST DO FIRST - No deployment without this

### 0.1 Token Encryption Implementation

Create `internal/security/encryption.go`:

```go
package security

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
)

type TokenEncryption struct {
    key []byte // 32 bytes for AES-256
}

func NewTokenEncryption(keyBase64 string) (*TokenEncryption, error) {
    key, err := base64.StdEncoding.DecodeString(keyBase64)
    if err != nil || len(key) != 32 {
        return nil, errors.New("invalid encryption key")
    }
    return &TokenEncryption{key: key}, nil
}

func (te *TokenEncryption) Encrypt(plaintext string) ([]byte, error) {
    block, _ := aes.NewCipher(te.key)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    return gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func (te *TokenEncryption) Decrypt(ciphertext []byte) (string, error) {
    block, _ := aes.NewCipher(te.key)
    gcm, _ := cipher.NewGCM(block)
    nonceSize := gcm.NonceSize()
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    return string(plaintext), err
}
```

### 0.2 Fix OAuth Cookie Security

```go
// ALL OAuth handlers - set Secure=true
c.SetCookie("oauth_state", state, 600, "/", "", true, true)
//                                              ^^^^ Secure=true
```

### 0.3 Production Secret Validation

```go
func (c *Config) Validate() error {
    if c.IsProduction() {
        if c.SecretKey == "your-secret-key-change-this-in-production" {
            return errors.New("SECRET_KEY must be changed in production")
        }
        if c.TokenEncryptionKey == "" {
            return errors.New("TOKEN_ENCRYPTION_KEY required")
        }
    }
    return nil
}
```

### 0.4 State Generation Fix

```go
func generateSecureState() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("crypto/rand failed: %w", err)
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

**Deliverables:**
- [ ] `internal/security/encryption.go`
- [ ] All OAuth handlers use Secure cookies
- [ ] Config validation for production
- [ ] No hardcoded secrets in code

---

## Phase 1: Credential Vault & Database

### 1.1 Add Missing Tables

```sql
-- Migration: 027_credential_vault.sql

CREATE TABLE credential_vault (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    provider_id VARCHAR(50) NOT NULL,
    credential_type VARCHAR(20) NOT NULL, -- oauth, api_key, custom
    encrypted_data BYTEA NOT NULL,
    encryption_version INT DEFAULT 1,
    expires_at TIMESTAMPTZ,
    last_rotated_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider_id)
);

CREATE TABLE integration_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_integration_id UUID REFERENCES user_integrations(id) ON DELETE CASCADE,
    provider_id VARCHAR(50) NOT NULL,
    webhook_url TEXT NOT NULL,
    webhook_secret_encrypted BYTEA,
    events TEXT[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    last_triggered_at TIMESTAMPTZ,
    failure_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE data_sync_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    source_provider VARCHAR(50) NOT NULL,
    source_entity VARCHAR(100) NOT NULL,
    target_module VARCHAR(50) NOT NULL,
    field_mappings JSONB NOT NULL DEFAULT '{}',
    transform_rules JSONB DEFAULT '{}',
    enabled BOOLEAN DEFAULT true,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_credential_vault_user ON credential_vault(user_id);
CREATE INDEX idx_credential_vault_provider ON credential_vault(provider_id);
CREATE INDEX idx_webhooks_integration ON integration_webhooks(user_integration_id);
CREATE INDEX idx_sync_mappings_user ON data_sync_mappings(user_id);
```

### 1.2 Credential Vault Service

```go
// internal/services/credential_vault.go

type CredentialVault interface {
    Store(ctx context.Context, userID, providerID string, creds Credentials) error
    Retrieve(ctx context.Context, userID, providerID string) (*Credentials, error)
    Rotate(ctx context.Context, userID, providerID string) error
    Revoke(ctx context.Context, userID, providerID string) error
    GetDecryptedToken(ctx context.Context, userID, providerID string) (string, error)
}

type Credentials struct {
    Type         string    // oauth, api_key
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Scopes       []string
    Metadata     map[string]interface{}
}
```

**Deliverables:**
- [ ] Migration 027_credential_vault.sql
- [ ] `internal/services/credential_vault.go`
- [ ] SQLC queries for credential operations
- [ ] Auto-rotation for expiring tokens

---

## Phase 2: Unified OAuth Architecture

### 2.1 OAuth Provider Registry

```go
// internal/integrations/oauth/registry.go

type OAuthProvider interface {
    GetConfig() *oauth2.Config
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
    RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error)
    GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}

var providers = map[string]OAuthProvider{
    "google":          &GoogleProvider{},
    "google_calendar": &GoogleCalendarProvider{},
    "slack":           &SlackProvider{},
    "notion":          &NotionProvider{},
    "hubspot":         &HubSpotProvider{},
    "linear":          &LinearProvider{},
    // Add more as needed
}
```

### 2.2 Unified OAuth Handler

```go
// internal/handlers/oauth.go

func (h *OAuthHandler) InitiateOAuth(c *gin.Context) {
    providerID := c.Param("provider")
    provider, ok := oauth.GetProvider(providerID)
    if !ok {
        c.JSON(400, gin.H{"error": "unknown provider"})
        return
    }

    state, _ := security.GenerateSecureState()
    h.stateManager.Store(state, providerID, time.Minute*10)

    c.SetCookie("oauth_state", state, 600, "/", "", true, true)
    c.JSON(200, gin.H{"auth_url": provider.GetAuthURL(state)})
}

func (h *OAuthHandler) HandleCallback(c *gin.Context) {
    state := c.Query("state")
    code := c.Query("code")

    providerID, valid := h.stateManager.Validate(state)
    if !valid {
        c.JSON(400, gin.H{"error": "invalid state"})
        return
    }

    provider := oauth.GetProvider(providerID)
    token, _ := provider.ExchangeCode(c, code)

    // Store encrypted in vault
    h.vault.Store(c, userID, providerID, Credentials{
        Type:         "oauth",
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        ExpiresAt:    token.Expiry,
    })

    // Create user_integration record
    h.queries.CreateUserIntegration(c, ...)

    c.Redirect(302, "/integrations?connected="+providerID)
}
```

### 2.3 API Key Handler

```go
// For providers that use API keys instead of OAuth

func (h *IntegrationHandler) ConnectWithAPIKey(c *gin.Context) {
    var req struct {
        ProviderID string `json:"provider_id"`
        APIKey     string `json:"api_key"`
    }
    c.BindJSON(&req)

    // Validate the API key works
    if err := h.validateAPIKey(req.ProviderID, req.APIKey); err != nil {
        c.JSON(400, gin.H{"error": "invalid API key"})
        return
    }

    // Store encrypted
    h.vault.Store(c, userID, req.ProviderID, Credentials{
        Type:        "api_key",
        AccessToken: req.APIKey,
    })

    c.JSON(200, gin.H{"success": true})
}
```

**Deliverables:**
- [ ] `internal/integrations/oauth/` package
- [ ] Provider implementations for: Google, Slack, Notion, HubSpot
- [ ] Unified `/api/oauth/:provider/init` and `/api/oauth/:provider/callback`
- [ ] API key validation and storage
- [ ] Frontend OAuth flow components

---

## Phase 3: Data Import & Sync

### 3.1 Import Engine

```go
// internal/services/import_engine.go

type ImportEngine interface {
    ImportFromProvider(ctx context.Context, userID, providerID string, options ImportOptions) (*ImportResult, error)
    GetImportStatus(ctx context.Context, importID string) (*ImportStatus, error)
}

type ImportOptions struct {
    Entities    []string          // contacts, tasks, messages, etc.
    Since       time.Time         // Only import after this date
    Mappings    DataSyncMapping   // Field mappings
    BatchSize   int
}

type ImportResult struct {
    ImportID        string
    RecordsImported int
    RecordsFailed   int
    Errors          []ImportError
}
```

### 3.2 Provider-Specific Importers

```go
// internal/integrations/importers/hubspot.go

type HubSpotImporter struct {
    vault  CredentialVault
    client *hubspot.Client
}

func (i *HubSpotImporter) ImportContacts(ctx context.Context, userID string) ([]Contact, error) {
    token, _ := i.vault.GetDecryptedToken(ctx, userID, "hubspot")

    contacts, _ := i.client.WithToken(token).Contacts.List()

    // Map to BusinessOS clients
    var results []Contact
    for _, c := range contacts {
        results = append(results, Contact{
            Name:  c.Properties.FirstName + " " + c.Properties.LastName,
            Email: c.Properties.Email,
            // ... more mappings
        })
    }
    return results, nil
}
```

### 3.3 Sync Scheduler

```go
// internal/services/sync_scheduler.go

type SyncScheduler struct {
    db     *sql.DB
    engine ImportEngine
}

func (s *SyncScheduler) ProcessPendingSyncs(ctx context.Context) {
    // Get all user_integrations with sync enabled
    integrations := s.queries.GetSyncableIntegrations(ctx)

    for _, integration := range integrations {
        switch integration.SyncFrequency {
        case "realtime":
            // Uses webhooks, handled separately
        case "hourly":
            if time.Since(integration.LastSyncedAt) > time.Hour {
                s.engine.ImportFromProvider(ctx, integration.UserID, integration.ProviderID, ...)
            }
        case "daily":
            // ...
        }
    }
}
```

**Deliverables:**
- [ ] Import engine with batch processing
- [ ] Importers for: HubSpot, Slack, Notion, Google Calendar
- [ ] Field mapping UI in frontend
- [ ] Sync scheduler (cron job)
- [ ] Webhook receivers for real-time sync

---

## Phase 4: Sorx Skill Integration

### 4.1 Credential Bridge

```go
// internal/sorx/credential_bridge.go

// This is what skills call to get credentials
func (e *Engine) GetCredential(ctx context.Context, providerID string) (*Credential, error) {
    userID := ctx.Value("user_id").(string)

    creds, err := e.vault.Retrieve(ctx, userID, providerID)
    if err != nil {
        return nil, fmt.Errorf("no %s credentials found", providerID)
    }

    // Auto-refresh if expired
    if creds.ExpiresAt.Before(time.Now()) {
        creds, err = e.refreshCredentials(ctx, userID, providerID, creds)
    }

    return &Credential{
        AccessToken: creds.AccessToken,
        Type:        creds.Type,
    }, nil
}
```

### 4.2 Real Action Handlers

Replace mocked actions with real implementations:

```go
// internal/sorx/actions/hubspot.go

func (h *HubSpotActions) GetContact(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    creds, _ := h.engine.GetCredential(ctx, "hubspot")

    client := hubspot.NewClient(creds.AccessToken)
    contactID := params["contact_id"].(string)

    contact, err := client.Contacts.Get(contactID)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "id":    contact.ID,
        "name":  contact.Properties.FirstName + " " + contact.Properties.LastName,
        "email": contact.Properties.Email,
        "phone": contact.Properties.Phone,
    }, nil
}

func (h *HubSpotActions) CreateDeal(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    creds, _ := h.engine.GetCredential(ctx, "hubspot")

    client := hubspot.NewClient(creds.AccessToken)

    deal, err := client.Deals.Create(hubspot.DealCreateRequest{
        Properties: hubspot.DealProperties{
            DealName:  params["name"].(string),
            Amount:    params["amount"].(float64),
            Stage:     params["stage"].(string),
        },
    })

    return deal, err
}
```

### 4.3 Skill-to-Credential Mapping

```go
// Map which credentials each skill needs
var skillCredentials = map[string][]string{
    "hubspot.qualify_lead":        {"hubspot"},
    "hubspot.deal_won_onboarding": {"hubspot", "slack", "google_calendar"},
    "slack.send_message":          {"slack"},
    "notion.create_page":          {"notion"},
    "gmail.send_email":            {"google"},
    // ...
}

func (e *Engine) ValidateSkillCredentials(ctx context.Context, skillID string) error {
    required := skillCredentials[skillID]
    userID := ctx.Value("user_id").(string)

    for _, provider := range required {
        if _, err := e.vault.Retrieve(ctx, userID, provider); err != nil {
            return fmt.Errorf("skill %s requires %s connection", skillID, provider)
        }
    }
    return nil
}
```

**Deliverables:**
- [ ] Credential bridge in Sorx engine
- [ ] Real action implementations for top 5 providers
- [ ] Skill-to-credential validation
- [ ] Error handling for expired/revoked tokens

---

## Phase 5: Frontend Integration UI

### 5.1 Integration Management Page

```svelte
<!-- src/routes/(app)/integrations/+page.svelte -->

<script lang="ts">
  import { integrationApi } from '$lib/api/integrations';
  import IntegrationCard from '$lib/components/integrations/IntegrationCard.svelte';

  let providers = $state<IntegrationProviderInfo[]>([]);
  let connections = $state<UserIntegration[]>([]);

  // Group by category
  $effect(() => {
    const grouped = {
      communication: providers.filter(p => p.category === 'communication'),
      crm: providers.filter(p => p.category === 'crm'),
      // ...
    };
  });

  async function connect(providerId: string) {
    const { auth_url } = await integrationApi.initiateOAuth(providerId);
    window.location.href = auth_url;
  }

  async function disconnect(providerId: string) {
    await integrationApi.disconnect(providerId);
    connections = connections.filter(c => c.provider_id !== providerId);
  }
</script>

<div class="integrations-page">
  {#each Object.entries(grouped) as [category, categoryProviders]}
    <section class="category-section">
      <h2>{category}</h2>
      <div class="provider-grid">
        {#each categoryProviders as provider}
          <IntegrationCard
            {provider}
            connection={connections.find(c => c.provider_id === provider.id)}
            onConnect={() => connect(provider.id)}
            onDisconnect={() => disconnect(provider.id)}
          />
        {/each}
      </div>
    </section>
  {/each}
</div>
```

### 5.2 Module Integration Panel

```svelte
<!-- Show available integrations per module -->
<script lang="ts">
  export let module: string; // 'tasks', 'clients', etc.

  let moduleIntegrations = $state<ModuleIntegrations>();

  onMount(async () => {
    moduleIntegrations = await integrationApi.getModuleIntegrations(module);
  });
</script>

<aside class="integration-panel">
  <h3>Connected Tools</h3>
  {#each moduleIntegrations?.connected_integrations ?? [] as integration}
    <div class="connected-item">
      <img src={integration.icon_url} alt={integration.provider_name} />
      <span>{integration.provider_name}</span>
      <button onclick={() => sync(integration.id)}>Sync</button>
    </div>
  {/each}

  <h4>Available</h4>
  {#each moduleIntegrations?.available_providers ?? [] as provider}
    <button onclick={() => connect(provider.id)}>
      Connect {provider.name}
    </button>
  {/each}
</aside>
```

**Deliverables:**
- [ ] `/integrations` page with category grouping
- [ ] IntegrationCard component
- [ ] Module-specific integration panels
- [ ] OAuth callback handling page
- [ ] Connection status indicators

---

## Implementation Priority

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         IMPLEMENTATION PRIORITY                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  WEEK 1: Phase 0 (Security) + Phase 1 (Database)                           │
│  ═══════════════════════════════════════════════                           │
│  □ Token encryption implementation                                          │
│  □ OAuth cookie security fixes                                              │
│  □ Production secret validation                                             │
│  □ Credential vault migration                                               │
│  □ Vault service implementation                                             │
│                                                                             │
│  WEEK 2: Phase 2 (OAuth Architecture)                                       │
│  ═════════════════════════════════════                                      │
│  □ Unified OAuth handler                                                    │
│  □ Google OAuth (calendar, drive, gmail)                                    │
│  □ Slack OAuth                                                              │
│  □ Notion OAuth                                                             │
│  □ API key handler pattern                                                  │
│                                                                             │
│  WEEK 3: Phase 3 (Data Import)                                              │
│  ══════════════════════════════                                             │
│  □ Import engine                                                            │
│  □ HubSpot importer (contacts, deals)                                       │
│  □ Slack importer (channels, messages)                                      │
│  □ Sync scheduler                                                           │
│                                                                             │
│  WEEK 4: Phase 4 (Sorx Integration)                                         │
│  ═══════════════════════════════════                                        │
│  □ Credential bridge                                                        │
│  □ Real HubSpot actions                                                     │
│  □ Real Slack actions                                                       │
│  □ Skill validation                                                         │
│                                                                             │
│  WEEK 5: Phase 5 (Frontend)                                                 │
│  ═════════════════════════════                                              │
│  □ Integrations page                                                        │
│  □ Module panels                                                            │
│  □ OAuth callback handling                                                  │
│  □ Status indicators                                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Provider Implementation Checklist

| Provider | OAuth | Import | Actions | Skills | Priority |
|----------|-------|--------|---------|--------|----------|
| Google Calendar | Working | - | Mocked | 3 | P0 |
| Google Drive | - | - | - | - | P1 |
| Gmail | - | - | Mocked | 2 | P1 |
| Slack | Working | - | Mocked | 3 | P0 |
| Notion | Working | - | Mocked | 3 | P0 |
| HubSpot | - | - | Mocked | 3 | P0 |
| Linear | - | - | - | 2 | P1 |
| ClickUp | - | - | Mocked | 3 | P1 |
| GitHub | - | - | - | 2 | P2 |
| Stripe | - | - | - | 2 | P2 |

---

## Success Metrics

1. **Security:** 0 plaintext tokens in database
2. **Coverage:** 10+ providers fully implemented
3. **Reliability:** 99% success rate on OAuth flows
4. **Performance:** < 500ms for credential retrieval
5. **Usage:** Skills executing real API calls with valid credentials

---

**Document Version:** 1.0
**Created:** 2026-01-05
**Author:** @architect + @security-auditor + @database-specialist
