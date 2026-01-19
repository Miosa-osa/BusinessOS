# P1: Integrations (Slack, Notion, Provider Catalog)

> **Priority:** P1 - High Value
> **Backend Status:** Complete (10+ endpoints per provider)
> **Frontend Status:** Partial (Google only)
> **Estimated Effort:** 2 sprints

---

## Overview

BusinessOS has a robust integrations backend supporting Google, Slack, and Notion. Currently only Google Calendar is exposed in the frontend. This doc covers implementing Slack and Notion integrations plus a unified provider catalog UI.

---

## Current State

| Provider | Backend | Frontend |
|----------|---------|----------|
| Google Calendar | ✅ Complete | ✅ Working |
| Google Drive | ✅ Complete | ❌ Not Started |
| Slack | ✅ Complete | ❌ Not Started |
| Notion | ✅ Complete | ❌ Not Started |
| Provider Catalog | ✅ Complete | ❌ Not Started |

---

## Backend API Endpoints

### Provider Catalog
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/integrations/providers` | List all integration providers |
| GET | `/api/integrations/providers/:id` | Get provider details |
| GET | `/api/integrations/status` | Get all integrations status |
| GET | `/api/integrations/connected` | List user's connected integrations |
| GET | `/api/integrations/:id` | Get specific integration |
| PATCH | `/api/integrations/:id/settings` | Update integration settings |
| DELETE | `/api/integrations/:id` | Disconnect integration |
| POST | `/api/integrations/:id/sync` | Trigger sync |

### Slack Integration
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/integrations/slack/auth` | Get Slack OAuth URL |
| GET | `/api/integrations/slack/callback` | OAuth callback |
| GET | `/api/integrations/slack/status` | Get connection status |
| DELETE | `/api/integrations/slack` | Disconnect Slack |
| GET | `/api/integrations/slack/channels` | List Slack channels |
| GET | `/api/integrations/slack/channels/:id/messages` | Get channel messages |
| POST | `/api/integrations/slack/channels/:id/messages` | Send message |
| GET | `/api/integrations/slack/users` | List Slack users |

### Notion Integration
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/integrations/notion/auth` | Get Notion OAuth URL |
| GET | `/api/integrations/notion/callback` | OAuth callback |
| GET | `/api/integrations/notion/status` | Get connection status |
| DELETE | `/api/integrations/notion` | Disconnect Notion |
| GET | `/api/integrations/notion/databases` | List Notion databases |
| GET | `/api/integrations/notion/databases/:id` | Get database details |
| GET | `/api/integrations/notion/pages` | List Notion pages |
| GET | `/api/integrations/notion/pages/:id` | Get page content |
| POST | `/api/integrations/notion/pages` | Create page |
| PATCH | `/api/integrations/notion/pages/:id` | Update page |
| POST | `/api/integrations/notion/sync` | Sync Notion data |

### Google Drive (Additional to Calendar)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/integrations/google/drive/files` | List Drive files |
| GET | `/api/integrations/google/drive/files/:id` | Get file content |
| POST | `/api/integrations/google/drive/upload` | Upload file |
| GET | `/api/integrations/google/drive/folders` | List folders |

---

## Data Models

### Integration Provider
```typescript
interface IntegrationProvider {
  id: string;
  name: string;           // e.g., "slack"
  display_name: string;   // e.g., "Slack"
  description: string;
  icon_url: string;
  category: ProviderCategory;
  features: string[];
  status: 'available' | 'coming_soon' | 'beta';
  auth_type: 'oauth' | 'api_key' | 'webhook';
}

type ProviderCategory =
  | 'communication'   // Slack, Discord
  | 'productivity'    // Notion, Asana
  | 'storage'         // Drive, Dropbox
  | 'calendar'        // Google Calendar
  | 'crm'             // Salesforce, HubSpot
  | 'developer';      // GitHub, GitLab
```

### Connected Integration
```typescript
interface ConnectedIntegration {
  id: string;
  provider_id: string;
  user_id: string;
  workspace_id: string;
  status: 'active' | 'error' | 'expired';
  settings: Record<string, any>;
  scopes: string[];
  last_sync_at?: string;
  error_message?: string;
  created_at: string;
}
```

---

## Frontend Implementation Tasks

### Phase 1: Integrations Hub

#### 1.1 Integrations Page
**File:** `src/routes/(app)/settings/integrations/+page.svelte`

- [ ] Provider catalog with cards
- [ ] Filter by category
- [ ] Search providers
- [ ] Show connected status badge
- [ ] "Coming Soon" badge for future integrations

#### 1.2 Provider Card Component
**File:** `src/lib/components/integrations/ProviderCard.svelte`

```svelte
<div class="provider-card">
  <img src={provider.icon_url} alt={provider.display_name} />
  <h3>{provider.display_name}</h3>
  <p>{provider.description}</p>
  <Badge>{provider.category}</Badge>

  {#if isConnected}
    <Badge variant="success">Connected</Badge>
    <Button variant="outline" on:click={() => manage(provider)}>Manage</Button>
  {:else if provider.status === 'coming_soon'}
    <Badge variant="outline">Coming Soon</Badge>
  {:else}
    <Button on:click={() => connect(provider)}>Connect</Button>
  {/if}
</div>
```

#### 1.3 Connected Integrations List
- [ ] Show all connected integrations
- [ ] Last sync time
- [ ] Sync now button
- [ ] Disconnect button

### Phase 2: Slack Integration

#### 2.1 Slack Connect Flow
**File:** `src/lib/components/integrations/SlackConnect.svelte`

- [ ] Connect button → OAuth flow
- [ ] Handle callback
- [ ] Show success/error state
- [ ] Request workspace access

#### 2.2 Slack Settings
**File:** `src/routes/(app)/settings/integrations/slack/+page.svelte`

- [ ] Connected workspace info
- [ ] Channel sync settings
- [ ] Notification preferences
- [ ] Disconnect button

#### 2.3 Slack Channel Browser
**File:** `src/lib/components/integrations/SlackChannels.svelte`

- [ ] List accessible channels
- [ ] Channel selection for notifications
- [ ] Preview recent messages

#### 2.4 Slack in Chat
- [ ] "Send to Slack" action on messages
- [ ] Import context from Slack threads
- [ ] Slack notifications for AI responses

### Phase 3: Notion Integration

#### 3.1 Notion Connect Flow
**File:** `src/lib/components/integrations/NotionConnect.svelte`

- [ ] Connect button → OAuth flow
- [ ] Page/database access selection
- [ ] Handle callback

#### 3.2 Notion Settings
**File:** `src/routes/(app)/settings/integrations/notion/+page.svelte`

- [ ] Connected workspace info
- [ ] Synced databases list
- [ ] Sync frequency settings
- [ ] Disconnect button

#### 3.3 Notion Database Browser
**File:** `src/lib/components/integrations/NotionDatabases.svelte`

- [ ] List Notion databases
- [ ] Preview database schema
- [ ] Select databases to sync

#### 3.4 Notion Page Browser
**File:** `src/lib/components/integrations/NotionPages.svelte`

- [ ] Browse Notion pages
- [ ] Search pages
- [ ] Import page as Context

#### 3.5 Notion Sync to Context
- [ ] Sync Notion page → BusinessOS Context
- [ ] Bidirectional sync option
- [ ] Conflict resolution

### Phase 4: Google Drive (Enhancement)

#### 4.1 Drive File Browser
**File:** `src/lib/components/integrations/DriveBrowser.svelte`

- [ ] Browse Drive folders
- [ ] File list with icons
- [ ] Search files
- [ ] Select files for context

#### 4.2 Drive Import
- [ ] Import Drive document as Context
- [ ] Support: Docs, Sheets, Slides
- [ ] Maintain formatting

### Phase 5: API Client

#### 5.1 Integrations API
**File:** `src/lib/api/integrations/integrations.ts`

```typescript
// Catalog
export async function getProviders(): Promise<IntegrationProvider[]>
export async function getProvider(id: string): Promise<IntegrationProvider>
export async function getIntegrationsStatus(): Promise<IntegrationStatus[]>
export async function getConnectedIntegrations(): Promise<ConnectedIntegration[]>
export async function disconnectIntegration(id: string): Promise<void>
export async function syncIntegration(id: string): Promise<void>
export async function updateIntegrationSettings(id: string, settings: any): Promise<void>

// Slack
export async function initiateSlackAuth(): Promise<{ url: string }>
export async function getSlackStatus(): Promise<SlackStatus>
export async function disconnectSlack(): Promise<void>
export async function getSlackChannels(): Promise<SlackChannel[]>
export async function getSlackMessages(channelId: string): Promise<SlackMessage[]>
export async function sendSlackMessage(channelId: string, text: string): Promise<void>

// Notion
export async function initiateNotionAuth(): Promise<{ url: string }>
export async function getNotionStatus(): Promise<NotionStatus>
export async function disconnectNotion(): Promise<void>
export async function getNotionDatabases(): Promise<NotionDatabase[]>
export async function getNotionPages(): Promise<NotionPage[]>
export async function getNotionPage(id: string): Promise<NotionPageContent>
export async function createNotionPage(data: CreatePageInput): Promise<NotionPage>
export async function syncNotion(): Promise<void>

// Google Drive
export async function getDriveFiles(folderId?: string): Promise<DriveFile[]>
export async function getDriveFile(id: string): Promise<DriveFileContent>
export async function uploadToDrive(file: File, folderId?: string): Promise<DriveFile>
```

#### 5.2 Integrations Store
**File:** `src/lib/stores/integrations.ts`

```typescript
interface IntegrationsStore {
  providers: IntegrationProvider[];
  connected: ConnectedIntegration[];
  isLoading: boolean;

  // Slack
  slackChannels: SlackChannel[];
  slackStatus: SlackStatus | null;

  // Notion
  notionDatabases: NotionDatabase[];
  notionStatus: NotionStatus | null;

  // Actions
  loadProviders(): Promise<void>;
  loadConnected(): Promise<void>;
  connectProvider(providerId: string): Promise<void>;
  disconnectProvider(providerId: string): Promise<void>;
  syncProvider(providerId: string): Promise<void>;
}
```

---

## UI/UX Requirements

### OAuth Flow
- Open OAuth in popup window
- Handle success/error callbacks
- Show loading state during OAuth
- Clear error messages

### Provider Cards
- Consistent design across providers
- Clear connected/disconnected state
- Easy access to settings

### Sync Status
- Visual indicator of last sync
- "Syncing..." state
- Error state with retry option

---

## Security Considerations

- OAuth tokens stored securely in backend
- Minimal scope requests
- Token refresh handling
- Secure callback URLs

---

## Testing Requirements

- [ ] Unit tests for integrations store
- [ ] Component tests for ProviderCard
- [ ] E2E: Connect Slack flow (mock OAuth)
- [ ] E2E: Connect Notion flow (mock OAuth)
- [ ] E2E: Browse/import from Notion

---

## Linear Issues to Create

1. **[INT-001]** Create Integrations Hub page
2. **[INT-002]** Build ProviderCard component
3. **[INT-003]** Implement Slack OAuth flow
4. **[INT-004]** Build Slack settings page
5. **[INT-005]** Create Slack channel browser
6. **[INT-006]** Implement Notion OAuth flow
7. **[INT-007]** Build Notion settings page
8. **[INT-008]** Create Notion database/page browser
9. **[INT-009]** Implement Notion → Context sync
10. **[INT-010]** Add Google Drive file browser
11. **[INT-011]** API client and store implementation
12. **[INT-012]** E2E tests

---

## Dependencies

- OAuth redirect URLs must be configured in each provider's dev console

## Blockers

- None identified

---

## Notes

- Consider webhooks for real-time sync (Slack events, Notion changes)
- Rate limiting awareness for API calls
- Future: More providers (Asana, Linear, GitHub, etc.)
