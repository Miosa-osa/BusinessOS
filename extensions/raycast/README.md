# Optimal Engine — Raycast Extension

Search, add, and query your [Optimal Engine](https://github.com/miosa/OptimalEngine) second brain directly from Raycast.

## Commands

| Command | Description |
|---------|-------------|
| **Search Memory** | Instant fuzzy search across your active workspace |
| **Add Memory** | Quick-add a fact, decision, or observation |
| **Ask Engine** | Natural-language Q&A answered from your workspace memories |

## Setup

### 1. Install the Raycast app

Download from [raycast.com](https://raycast.com) and install.

### 2. Install this extension (development mode)

```bash
cd extensions/raycast
npm install
npm run dev
```

Raycast will open and import the extension automatically.

### 3. Configure preferences

Open Raycast → Extension Preferences → Optimal Engine and set:

| Preference | Default | Description |
|-----------|---------|-------------|
| Engine URL | `http://localhost:4200` | Base URL of your running Optimal Engine instance |
| Default Workspace | `default` | Workspace slug to search / write to |
| API Key | _(empty)_ | Bearer token — only required when `auth_required=true` |

## Usage

### Search Memory

Type any term. Results stream in with a 200 ms debounce. The detail panel on the right shows the full content, audience, relevance score, and metadata.

Actions available on each result:
- **Copy Content** — copy the full memory text to clipboard
- **Copy Slug** — copy the memory slug
- **Open Source URL** — open `citation_uri` in the browser (when present)

### Add Memory

Fill in the form:
- **Content** (required) — the fact, decision, or observation
- **Audience** — `general | technical | executive | internal`
- **Static** — pin the memory so it is never evicted by the engine
- **Source URL** — optional provenance link

On success a toast displays the new memory ID and the form closes automatically.

### Ask Engine

Type a natural-language question and press **Enter**. The engine answers using RAG over your workspace memories. The detail view renders the answer as Markdown with sources listed below.

Actions available on each answer:
- **Copy Answer** — copy the full response body
- **Copy Sources** — copy the source slugs / URLs
- **Ask Follow-up** — pre-fills the search bar with the previous query
- **Open: \<slug\>** — opens the citation URL for any source that has one

## Development

```bash
# Type-check only (no Raycast CLI required)
npx tsc --noEmit

# Full build (requires Raycast app installed)
npm run build

# Live-reload dev mode
npm run dev
```

## File Structure

```
src/
├── search-memory.tsx       Search command
├── add-memory.tsx          Add command
├── ask-engine.tsx          Ask command
└── lib/
    ├── client.ts           Engine HTTP wrapper (native fetch, no deps)
    ├── preferences.ts      Typed preference accessor
    └── types.ts            All shared TypeScript types
└── hooks/
    └── use-workspaces.ts   Cached workspace list hook
```

## Metadata / Screenshots

Place Raycast Store screenshots (1280×800 PNG) in `metadata/`. See the
[Raycast developer docs](https://developers.raycast.com/basics/prepare-an-extension-for-store#screenshots)
for naming conventions.
