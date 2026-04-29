# Wiki governance schema

Rules the curator (`OptimalEngine.Wiki.Curator`) honors when appending
new signals or rewriting existing wiki pages.

## Required frontmatter keys

Every wiki page must carry these in its YAML frontmatter:

- `slug` — filename stem, URL-safe, lowercase, kebab-case.
- `title` — human display title.
- `audience` — one of `default | sales | legal | exec | engineering`.

## Required sections

Every wiki page must have:

- `## Summary` — 1–3 sentences, L0-grade.
- `## Related` — other pages + key signals, each cited.

Pages in the `engineering` audience must additionally carry
`## Architecture` and `## Acceptance Criteria`.

## Size ceiling

- `max_bytes: 8192` — soft cap. Past this the curator must either
  spawn child pages under `<slug>/` or split into a hub/spoke.

## Citation density

- Every factual paragraph must include at least one `{{cite: uri}}`
  directive pointing back to a Tier-1 chunk.
- Paragraphs over 200 characters without a citation are flagged by
  `mix optimal.wiki verify`.

## Directive whitelist

The executable-directive lexer accepts only these verbs:

```
cite · include · expand · search · table · trace · recent · [[wikilink]]
```

Anything else renders verbatim with a warning. Never add new verbs
without a migration in `OptimalEngine.Wiki.Directives.@whitelist`.

## Version policy

- Version 1 is the first curated shape.
- Bump version when the `## Summary` changes materially or a new
  section is added.
- Frontmatter `last_curated` and `curated_by` are set by the curator;
  don't edit by hand.
