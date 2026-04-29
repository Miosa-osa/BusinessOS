# Optimal Engine — UI Architecture

> The engine is headless. This doc defines the desktop application that sits on
> top — the view layer a human (or an agent through a human proxy) uses to
> navigate the three tiers, audit citations, and supervise curation.

---

## 1. Information architecture

Seven primary views. One at a time in the main panel. Every view is a window
onto the same underlying store.

| View         | Purpose                                                                          | Data source (engine)                          |
|--------------|----------------------------------------------------------------------------------|-----------------------------------------------|
| **Brief**    | Read curated wiki pages. Default landing view. Hot citations + directives resolve inline. | `wiki_pages`, `citations`                     |
| **Source**   | Browse the raw signal files. Tier 1. Preview any MD/PDF/image/audio.             | filesystem + `signals`                        |
| **Probe**    | Ask a question. Hybrid retrieval. Results show chunk scale + source + score.     | `SearchEngine.search/2`                       |
| **Atlas**    | Visual map: knowledge graph with clusters, hubs, and missing edges highlighted.  | `edges`, `clusters`, `cluster_members`        |
| **Flow**    | Live view of the ingestion pipeline: signals moving through stages 1–9.          | `events` table + telemetry                    |
| **Audit**    | Citation integrity, contradictions, stale pages, orphaned chunks.                | `OptimalEngine.Wiki.Integrity` + health checks|
| **Queue**    | Items the curator flagged for human judgment (contradictions, low-confidence intents, ambiguous routing). | curator-emitted review items                  |

Secondary surfaces:

- **Ask bar** — persistent bottom bar. A single input that sends a query through the wiki-first → hybrid retrieval flow and renders results inline in the current view. Like Raycast, not like Slack.
- **Sweep panel** — right-side drawer. Launched from Ask bar or Probe. Runs a multi-query deep investigation: rewrites the question N ways, runs retrieval for each, collates, auto-ingests new findings, updates affected Brief pages through the curation loop.
- **Clip** — system tray icon + browser extension endpoint. Capture any webpage → auto-ingest into Source → auto-trigger relevant wiki curation.

---

## 2. Layout

```
┌───┬───────────────────┬────────────────────────────────┬────────────────────┐
│ R │  Tree             │  Main panel                    │  Right panel       │
│ a │                   │                                │  (conditional)     │
│ i │  File tree if     │  One of:                       │                    │
│ l │  view = Source    │   Brief   — rendered wiki      │  On Brief:         │
│   │                   │            page w/ directives  │    Source preview  │
│ 8 │  Page tree if     │   Source  — file preview       │    for any cite    │
│ 4 │  view = Brief     │   Probe   — results list       │                    │
│ 0 │                   │   Atlas   — graph canvas       │  On Probe:         │
│ p │  Cluster tree if  │   Flow   — pipeline tail       │    Chunk detail    │
│ x │  view = Atlas     │   Audit   — integrity report   │                    │
│   │                   │   Queue   — review items       │  On Sweep open:    │
│ w │                   │                                │    Sweep panel     │
│ i │  Flow tail if     │                                │                    │
│ d │  view = Flow      │                                │                    │
│ e │                   │                                │                    │
├───┴───────────────────┴────────────────────────────────┴────────────────────┤
│ Ask bar: [ > ask anything, or /command… ]                                    │
└──────────────────────────────────────────────────────────────────────────────┘
```

Rail on the left (48px) with icons for the 7 views + Sweep toggle + settings.
Tree panel resizable 180–400px. Right panel appears only when there's
something to show (a selected cite, a chunk detail, or Sweep open). Ask bar
is always present at the bottom.

**Visual language:**

- Monochrome base (no blue), Geist variable font, Foundation tokens.
- Color only for signal: S/N ratio heat (cold → warm), intent chips, cluster tints.
- Pill buttons, shadow focus. HSL CSS vars.
- Dark mode default; light mode via `--background`/`--foreground` swap.

---

## 3. Feature inventory — what's in v0.1

### In (v0.1)

| Area                  | Feature                                                                              |
|----------------------|--------------------------------------------------------------------------------------|
| Layout               | Icon rail + resizable tree panel + main panel + conditional right panel.             |
| Desktop shell        | Tauri 2 (Rust core, small cross-platform binary).                                    |
| Editor               | Milkdown for markdown authoring (WYSIWYG + source toggle).                           |
| Graph rendering      | Sigma.js + graphology for Atlas view.                                                |
| Graph clustering     | Louvain community detection over entity/chunk graph (alongside HDBSCAN for chunks).  |
| Relevance scoring    | Multi-signal composite: direct link weight, source overlap, Adamic-Adar, type affinity — folded into `SearchEngine` graph boost. |
| Pipeline visibility  | **Flow** view — persistent ingest queue with crash recovery, cancel, retry.          |
| Web capture          | **Clip** browser extension: capture page → auto-ingest.                              |
| Human-in-loop        | **Queue** view — curator flags items for human judgment with predefined actions.     |
| Deep investigation   | **Sweep** drawer — multi-query rewrite, parallel retrieval, auto-ingest findings.    |
| Curator loop         | Two-stage: analyze + plan first, then generate wiki page with citations.             |
| Directive syntax     | `[[wikilink]]` and `{{directive: arg}}` parsed and resolved at read time.            |
| Content catalog      | `.wiki/index.md` as the LLM's navigation entry point.                                |
| Governance           | `.wiki/SCHEMA.md` — human-owned contract the curator reads on every run.             |
| Block addressability | **Chunk-addressable URIs** — every chunk retrievable via `optimal://path#chunk-abc`. |
| Back-references      | **Incoming** section in right panel — reverse edges from the active chunk/page.      |
| Scaffolds            | Per-genre page templates (decision-log, person, topic, project).                     |
| Lens                 | Alternate renderings of a result set (list / table / kanban-by-intent / timeline).   |
| OCR                  | Tesseract in Parser stage 8.                                                         |
| Inline extras        | Mermaid, KaTeX, ECharts via Milkdown plugins.                                        |
| Deployment           | Docker image + Tauri desktop builds for macOS, Linux, Windows.                       |

### Out (deferred to later phases)

| Deferred feature              | Why                                                     |
|------------------------------|---------------------------------------------------------|
| Inline PDF annotation         | Expensive UX; revisit post-Phase 10.                    |
| WebDAV / S3 sync              | Multi-machine use — not v0.1.                           |
| Plugin system                 | Lock the core before opening an extension surface.      |
| Mobile app                    | Desktop-first, web-second.                              |
| Multi-column document layout  | Partial overlap with Lens; not a priority.              |
| i18n                          | English only for v0.1.                                  |

### Explicitly not in (ever, or wrong fit)

| Not used              | Why                                                             |
|----------------------|-----------------------------------------------------------------|
| React                 | SvelteKit matches MIOSA/BusinessOS stack; smaller runtime.       |
| Off-the-shelf component framework (React-coupled) | bits-ui + melt-ui for Svelte instead. |
| Dedicated vector DB (e.g. external service) | SQLite blobs + in-process cosine. Revisit if we outgrow ~50K vectors. |
| OpenAI-only LLM endpoints | Ollama-first, with OpenAI as one adapter among several.     |
| Dedicated append-only log file | We use the SQLite `events` table, rendered in Flow.        |
| Electron              | Tauri is the smaller, faster alternative.                        |
| Flashcards            | Different product.                                               |

---

## 4. Tech stack (locked for v0.1)

| Layer              | Choice                                    | Rationale                                      |
|--------------------|-------------------------------------------|------------------------------------------------|
| Desktop shell      | Tauri 2                                   | Small cross-platform binary; Rust core.        |
| Frontend framework | SvelteKit                                 | Matches MIOSA/BusinessOS stack.                |
| Styling            | Tailwind CSS v4 + Foundation tokens       | Alice's existing design system.              |
| Component primitives | bits-ui + melt-ui                       | Svelte-native, headless.                        |
| Markdown editor    | Milkdown                                  | CRDT-ready, plugin-rich.                       |
| Graph rendering    | Sigma.js + graphology                     | Battle-tested force-directed + community layouts. |
| State              | Svelte stores + persistence via Tauri store plugin | Small, idiomatic.                               |
| Engine transport   | HTTP API on `localhost:4200`              | Tauri front ↔ Elixir engine over localhost.    |
| IPC fallback       | Tauri commands → shell `optimal` CLI      | For actions that require the full engine VM.   |

---

## 5. View-by-view specs

### Brief

- Left tree: wiki page hierarchy (`.wiki/` directory + category folders).
- Main: rendered markdown with:
  - Directives resolved inline (hover cite → popover with source chunk).
  - Backlinks section ("Incoming" — auto-generated from `edges` pointing in).
  - Cluster banner — which cluster this page belongs to.
  - Frontmatter shown as a compact bar at top.
- Right panel: on cite click, shows the source chunk at its native scale with nav-up buttons to `:paragraph → :section → :document`.

### Source

- Left tree: filesystem mirror of `nodes/**/signals/`.
- Main: file preview.
  - MD/TXT → rendered.
  - PDF/DOCX → pdftotext output + original-download button.
  - Image → inline + OCR text below.
  - Audio → waveform + transcript.
- Right panel: chunk breakdown of the selected file with scale toggles.

### Probe

- Main: search input + results list.
- Each result row: title, snippet with match highlight, chunk scale badge, S/N ratio pill, intent chip, score, source URI.
- Right panel: on result click → full chunk detail + lineage (parent/child chunks).
- **Lens toggle**: list / table / kanban-by-intent / timeline-by-recency.

### Atlas

- Main: Sigma.js force-directed graph of entities + chunks + clusters.
- Controls: node type filter, degree min, cluster color-by, community detection toggle.
- Clicking a node: pins it + fades non-neighbors + reveals edge list.
- Overlays: hubs (size by degree), triangles (A→B, A→C, missing B→C), gaps (reflect output).

### Flow

- Main: streaming tail of the `events` table, one row per pipeline transition.
- Columns: timestamp, signal id, stage (1–9), duration, status.
- Click a row → full event payload + rewind to prior stage.

### Audit

- Main: health report sections — broken citations, stale pages, contradictions, orphan chunks, cluster hygiene.
- Each section has a "fix" action where trivial + a "flag for review" button otherwise.

### Queue

- Main: list of curator-flagged items awaiting human call: contradictions, low-confidence intents, ambiguous routing.
- Each item has: the question, the evidence, 2–4 predefined actions, "send to Sweep" for deeper investigation.

---

## 6. Navigation

- `⌘K` → Ask bar focus from anywhere.
- `⌘1–7` → jump to view.
- `⌘O` → open source file by path.
- `⌘B` → toggle backlinks panel.
- `⌘⇧S` → launch Sweep from current context.
- Global `Esc` → close right panel or Sweep drawer.

---

## 7. UI build plan (after engine Phase 1–8)

UI work starts once the engine pipeline is built. It layers cleanly over the
HTTP API:

| UI Phase | Scope                                                      |
|----------|------------------------------------------------------------|
| U1       | Tauri + SvelteKit scaffold. Rail + tree + main panel shell.|
| U2       | Brief view (read-only wiki pages + directives resolve).    |
| U3       | Source view (file preview with modality-aware renderers).  |
| U4       | Probe view (search + Lens toggle).                         |
| U5       | Atlas view (Sigma.js + Louvain + Adamic-Adar overlays).    |
| U6       | Flow + Audit views.                                        |
| U7       | Queue + Sweep + Clip extension.                            |
| U8       | Wiki editing (Milkdown) for human curator overrides.       |
| U9       | Packaging: Tauri bundle for macOS + Linux + Windows.       |

---

## 8. Invariants

1. **The UI is a view. The engine is the authority.** Everything the UI shows comes from the HTTP API. No UI-side derived state that isn't round-tripped through the engine.
2. **No UI-exclusive content.** A user edit in Brief persists to a `.wiki/` file on disk. Delete the app — the knowledge survives.
3. **Keyboard-first, mouse-sufficient.** Every action has a shortcut. Mouse works but isn't required.
4. **Responsive to engine state changes.** The engine emits `wiki.page.updated`, `store.chunk.indexed`, `cluster.changed` signals. The UI subscribes and re-renders live.
