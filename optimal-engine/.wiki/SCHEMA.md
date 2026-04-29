---
name: Optimal Engine Wiki Schema
version: 1
type: governance
---

# Wiki Schema

Authoritative rules for the Tier-3 wiki — enforced by
`OptimalEngine.Wiki.Integrity.against_schema/2` when serialized as the
map below, and read by `OptimalEngine.Wiki.Curator` when constructing
the curator's system prompt.

A page that violates the "error" rules cannot be committed. A page that
violates the "warning" rules still commits but surfaces in
`mix optimal.wiki verify`.

## Required sections (errors)

Every curated wiki page MUST contain these level-2 headings:

- `## Summary` — 2–4 paragraph overview with citations.
- `## Open threads` — active topics, decisions pending, outstanding asks.
- `## Related` — cross-links to other wiki pages via `[[slug]]`.

## Required frontmatter (errors)

Every page MUST carry this frontmatter:

- `slug` — kebab-case, unique within `(tenant_id, audience)`.
- `audience` — one of: `default`, `sales`, `engineering`, `exec-brief`, or a
  tenant-defined custom audience.
- `version` — monotonic integer, bumped by the curator on every write.
- `last_curated` — ISO-8601 timestamp of the most recent curator run.

## Citation rules (errors)

- Every factual claim in `## Summary` and `## Decisions` (if present) MUST
  carry at least one `{{cite: optimal://…}}` directive.
- Every `{{cite: uri}}` MUST resolve to a real chunk in the Store. Broken
  citations block the commit.
- Citations never point upward across tiers: a Tier-3 citation resolves to
  Tier-2 chunks or Tier-1 raw signal files, never to other wiki pages.
  Cross-page links use `[[slug]]`, not `{{cite}}`.

## Size ceilings (warnings)

- `max_bytes: 50_000` — pages above this threshold should spawn child
  pages; leave a `## Related` pointer.

## Directive whitelist (errors)

Only these verbs are allowed inside `{{verb: argument [key=value…]}}`:

- `cite` — URI pointer (rendered as footnote)
- `include` — inline the referenced content, optionally at a tier
- `expand` — sub-query the wiki for a slug
- `search` — invoke hybrid retrieval
- `table` — fetch a structured row from a CSV/sheet
- `trace` — walk the knowledge graph from an entity
- `recent` — inject recent signals from a node

Anything else triggers a `:invalid_verb` error.

## Machine-readable subset

The integrity checker consumes this schema as a map with the fields:

```elixir
%{
  "required_sections" => ["Summary", "Open threads", "Related"],
  "required_frontmatter" => ["slug", "audience", "version", "last_curated"],
  "max_bytes" => 50_000
}
```

Add rules here; don't hardcode them in the checker. Schema evolution is
a governance action, not a code change.
