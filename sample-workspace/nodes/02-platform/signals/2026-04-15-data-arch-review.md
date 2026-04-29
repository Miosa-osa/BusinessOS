---
title: Data-architecture review — time-series + geo modalities
genre: spec
mode: linguistic
node: 02-platform
authored_at: 2026-04-15T11:00:00Z
sn_ratio: 0.85
entities:
  - { name: "Carol", type: person }
  - { name: "Bob", type: person }
  - { name: "Ivan", type: person }
  - { name: "DataArchitecture", type: product }
  - { name: "time-series", type: concept }
  - { name: "geo", type: concept }
---

## Summary

Review of adding time-series + geo as first-class modalities in the
DataArchitecture registry. Ship target: 2 sprints.

## Why now

Healthtech partners are pushing telemetry (vitals, device streams) +
geo (clinic coverage areas). Current registry forces them into
`structured` which loses every temporal / spatial affordance.

## Field additions

- `time-series` — numeric sequence with timestamps; processor =
  `ts_feature_extractor` (classical stats) with optional anomaly
  detector.
- `geo` — lat/lon point OR polygon; processor = a GIS encoder
  emitting a bbox + semantic tags (city / region / country).

## Work split

| Who   | Task                                         |
|-------|----------------------------------------------|
| Carol | Processor bindings + alignment policy        |
| Bob   | SQL schema extensions (time-series tables)   |
| Ivan  | Integration tests + reality-check probes     |

## Risks

- Alignment across 768-d space is fine for text + image; time-series
  features sit outside it. Decision: keep a parallel `features` table
  keyed on `chunk_id` rather than forcing into the aligned space.
- Geo encoder vendor choice is open — shortlist: two OSS options. Ivan
  drafts the comparison next week.
