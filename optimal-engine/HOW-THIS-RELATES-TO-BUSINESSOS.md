# OptimalEngine inside BusinessOS

This directory holds the **canonical Elixir reference implementation** of
OptimalEngine, vendored into the BusinessOS-5 monorepo via `git subtree`.

## Why both languages live here

```
businessos-5/
├── optimal-engine/                 ← Elixir source (this directory)
│                                     The reference. Read here when in doubt
│                                     about what an algorithm is supposed to do.
│
└── desktop/backend-go/             ← BusinessOS Go backend
    ├── go.mod                      require github.com/Miosa-osa/OptimalEngine-go vX.Y.Z
    │                                 replace ... => ../../../OptimalEngine-go
    ├── internal/connectorbridges/  BusinessOS-specific glue between
    │                                 engine adapters and BO OAuth providers
    └── internal/services/*_processor.go
                                    BusinessOS-specific architecture.Processor
                                    impls (text/code/image/audio/ts) on top
                                    of EmbeddingService, WhisperService, etc.
```

## What runs in production

The Go server in `desktop/backend-go/` runs OptimalEngine via
[`Miosa-osa/OptimalEngine-go`](https://github.com/Miosa-osa/OptimalEngine-go)
— a separate repo cloned as a sibling under `~/code/`. **The Elixir source
in this directory is not deployed.** It's here so:

1. The Go port has a canonical reference checked into the same repo. Anyone
   reviewing a Go change can grep the Elixir equivalent in ten seconds.
2. The BusinessOS repo can be cloned standalone and contain everything —
   the engine's full design alongside its consumer.
3. Documentation, mix tasks, and architectural notes from the Elixir side
   stay accessible without an extra clone.

## Updating this snapshot

The subtree is a snapshot of `Miosa-osa/OptimalEngine` at import time. To
pull in upstream changes:

```bash
cd ~/code/businessos-5
git subtree pull --prefix=optimal-engine optimal-engine-elixir main --squash
```

(`optimal-engine-elixir` is the remote configured during the initial
import; if it's missing locally:
`git remote add optimal-engine-elixir https://github.com/Miosa-osa/OptimalEngine.git`.)

## Three places, one engine

| Location | Language | Role |
|---|---|---|
| `Miosa-osa/OptimalEngine` (GitHub) | Elixir | Canonical OSS reference |
| `optimal-engine/` (this dir) | Elixir | Vendored snapshot for BusinessOS contributors |
| `Miosa-osa/OptimalEngine-go` (GitHub) | Go | Production runtime; consumed via go.mod |

Algorithm parity is enforced by manual review against this snapshot. The Go
port's commit messages cite the Elixir filenames they ported from — see the
`Phase X` commits in `Miosa-osa/OptimalEngine-go` history.

## License

Apache 2.0 — same as the upstream Elixir source.
