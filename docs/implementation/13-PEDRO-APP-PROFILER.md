# Pedro Tasks: App Profiler

> **Priority:** P2 - Nice to Have (Pedro's Work)
> **Backend Status:** Complete (9 endpoints)
> **Frontend Status:** Not Started
> **Owner:** Pedro
> **Estimated Effort:** 1 sprint

---

## Overview

App Profiler analyzes codebases to extract structure, components, API endpoints, and tech stack. This enables context-aware AI assistance for development tasks.

---

## Backend API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/app-profiles` | Profile a new application |
| GET | `/api/app-profiles` | List all profiles |
| GET | `/api/app-profiles/:name` | Get profile summary |
| POST | `/api/app-profiles/:name/refresh` | Re-profile application |
| GET | `/api/app-profiles/:name/components` | List components/modules |
| GET | `/api/app-profiles/:name/endpoints` | List API endpoints |
| GET | `/api/app-profiles/:name/structure` | Get directory structure |
| GET | `/api/app-profiles/:name/modules` | Get module information |
| GET | `/api/app-profiles/:name/tech-stack` | Get tech stack details |

---

## Data Models

```typescript
interface AppProfile {
  id: string;
  name: string;
  root_path: string;
  description?: string;

  // Analysis Results
  tech_stack: TechStack;
  components: Component[];
  endpoints: APIEndpoint[];
  modules: Module[];
  directory_structure: DirectoryNode;

  // Metadata
  file_count: number;
  line_count: number;
  last_analyzed: string;
  created_at: string;
}

interface TechStack {
  languages: Language[];
  frameworks: Framework[];
  libraries: Library[];
  conventions: string[];
}

interface Component {
  name: string;
  path: string;
  type: 'component' | 'page' | 'layout' | 'util' | 'hook' | 'service';
  description?: string;
  dependencies: string[];
}

interface APIEndpoint {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  path: string;
  handler: string;
  description?: string;
  parameters?: Parameter[];
}

interface Module {
  name: string;
  path: string;
  exports: string[];
  imports: string[];
  description?: string;
}

interface DirectoryNode {
  name: string;
  type: 'file' | 'directory';
  path: string;
  children?: DirectoryNode[];
}
```

---

## Frontend Implementation Tasks

### Phase 1: Profile Management

#### 1.1 App Profiles Page
**File:** `src/routes/(app)/settings/app-profiles/+page.svelte`

- [ ] List of profiled applications
- [ ] "Profile New App" button
- [ ] Last analyzed timestamp
- [ ] Refresh profile button
- [ ] Delete profile

#### 1.2 Create Profile Modal
**File:** `src/lib/components/profiler/CreateProfileModal.svelte`

- [ ] Application name input
- [ ] Root path input (with file picker if possible)
- [ ] Description textarea
- [ ] Analyze button with progress

### Phase 2: Profile Viewer

#### 2.1 Profile Detail Page
**File:** `src/routes/(app)/settings/app-profiles/[name]/+page.svelte`

- [ ] Overview tab (summary, stats)
- [ ] Tech Stack tab
- [ ] Components tab
- [ ] Endpoints tab
- [ ] Structure tab
- [ ] Modules tab

#### 2.2 Tech Stack View
**File:** `src/lib/components/profiler/TechStackView.svelte`

- [ ] Languages with icons
- [ ] Frameworks with versions
- [ ] Libraries list
- [ ] Detected conventions

#### 2.3 Components Browser
**File:** `src/lib/components/profiler/ComponentsBrowser.svelte`

- [ ] Tree or list view of components
- [ ] Filter by type
- [ ] Search components
- [ ] Click to see details

#### 2.4 Endpoints Documentation
**File:** `src/lib/components/profiler/EndpointsView.svelte`

- [ ] API endpoints list
- [ ] Group by path prefix
- [ ] Method badges (GET, POST, etc.)
- [ ] Expand for parameters/details

#### 2.5 Directory Tree
**File:** `src/lib/components/profiler/DirectoryTree.svelte`

- [ ] Collapsible tree view
- [ ] File/folder icons
- [ ] Click to navigate
- [ ] Search within tree

### Phase 3: Chat Integration

#### 3.1 App Context in Chat
- [ ] "Add app profile context" button
- [ ] AI can reference app structure
- [ ] Smart code suggestions based on profile

### Phase 4: API Client

#### 4.1 App Profiler API
**File:** `src/lib/api/profiler/profiler.ts`

```typescript
export async function createProfile(data: CreateProfileInput): Promise<AppProfile>
export async function getProfiles(): Promise<AppProfile[]>
export async function getProfile(name: string): Promise<AppProfile>
export async function refreshProfile(name: string): Promise<AppProfile>
export async function deleteProfile(name: string): Promise<void>
export async function getComponents(name: string): Promise<Component[]>
export async function getEndpoints(name: string): Promise<APIEndpoint[]>
export async function getStructure(name: string): Promise<DirectoryNode>
export async function getModules(name: string): Promise<Module[]>
export async function getTechStack(name: string): Promise<TechStack>
```

---

## UI/UX Requirements

### Profile Creation
- Clear path input
- Progress indicator during analysis
- Error handling for invalid paths

### Profile Viewer
- Tab-based navigation
- Search within each section
- Copy-friendly text

---

## Testing Requirements

- [ ] Unit tests for profiler store
- [ ] Component tests for tree view
- [ ] E2E: Create profile flow
- [ ] E2E: Browse profile data

---

## Linear Issues to Create

1. **[PROF-001]** Create App Profiles page
2. **[PROF-002]** Build create profile modal
3. **[PROF-003]** Implement profile detail page
4. **[PROF-004]** Add Tech Stack view
5. **[PROF-005]** Build Components browser
6. **[PROF-006]** Create Endpoints documentation view
7. **[PROF-007]** Add Directory tree component
8. **[PROF-008]** Integrate with chat context
9. **[PROF-009]** API client implementation

---

## Notes

- This is powerful for developer onboarding
- Could auto-profile on project creation
- Consider incremental updates vs full re-analysis
