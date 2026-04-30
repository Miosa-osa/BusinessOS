# Optimal Engine — Docs

SvelteKit static docs site for [Optimal Engine](https://github.com/Miosa-osa/OptimalEngine).

## Run locally

```bash
cd apps/docs
npm install
npm run dev        # dev server at http://localhost:1422
```

## Build (static)

```bash
npm run build      # produces build/
```

The output is a fully static site — every route is pre-rendered at build time. Serve with any static host (Netlify, Vercel, Cloudflare Pages, nginx, S3).

## Type check

```bash
npm run check
```

## Deploy

```bash
# Netlify
netlify deploy --dir build --prod

# Cloudflare Pages
wrangler pages publish build

# nginx — point root at build/
```

## Structure

```
src/
├── routes/
│   ├── +layout.svelte           Glass header + sidebar shell
│   ├── +page.svelte             Landing page
│   ├── quickstart/
│   ├── concepts/
│   │   ├── three-tiers/
│   │   ├── nine-stages/
│   │   ├── signal-theory/
│   │   ├── workspaces/
│   │   ├── memory-primitive/
│   │   └── proactive-surfacing/
│   ├── api/
│   │   ├── retrieval/
│   │   ├── memory/
│   │   ├── recall/
│   │   ├── workspaces/
│   │   ├── wiki/
│   │   └── surfacing/
│   ├── sdks/
│   │   ├── typescript/
│   │   ├── python/
│   │   └── mcp/
│   ├── extensions/
│   │   ├── browser/
│   │   └── raycast/
│   └── self-host/
└── lib/
    ├── components/
    │   ├── Sidebar.svelte
    │   ├── CodeBlock.svelte
    │   ├── ApiSpec.svelte
    │   └── EngineThemeToggle.svelte
    └── data/
        ├── nav.ts               Sidebar navigation tree
        └── endpoints.ts         API endpoints as typed data
```

## Tech

- SvelteKit 2 + Svelte 5 (runes)
- adapter-static, prerender: true
- No external UI dependencies — all styles are hand-written CSS vars from the Foundation token system
- Port 1422 (dev + preview)
