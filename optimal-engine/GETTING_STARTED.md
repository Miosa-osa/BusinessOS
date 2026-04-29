# Getting Started

Zero-to-running in about 5 minutes.

## 1. Prerequisites

| What         | Version             | Check                         |
|--------------|---------------------|-------------------------------|
| Elixir       | `~> 1.17`           | `elixir --version`            |
| Erlang / OTP | `26+`               | `erl -version`                |
| C toolchain  | (for the SQLite NIF)| `cc --version`                |
| Node         | `20+` (desktop UI)  | `node --version`              |
| Rust stable  | (Tauri bundle only) | `rustc --version`             |

On macOS:

```bash
brew install elixir node rust
```

On Debian / Ubuntu:

```bash
sudo apt install elixir build-essential nodejs
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

Optional — enrichens parser coverage (graceful no-op when absent):

```bash
brew install pdftotext tesseract ffmpeg
```

## 2. Clone + bootstrap

```bash
git clone https://github.com/Miosa-osa/OptimalEngine.git
cd OptimalEngine
make install       # deps + compile
make bootstrap     # migrate + ingest sample-workspace/
```

Or, without make:

```bash
mix deps.get
mix compile
mix optimal.bootstrap
```

`bootstrap` is idempotent — run it again to re-seed after pulling.

## 3. Use the CLI

```bash
mix optimal.rag "healthtech pricing decision" --trace
mix optimal.search "platform"
mix optimal.wiki list
mix optimal.graph hubs
mix optimal.architectures
```

## 4. Launch the desktop UI

Enable the HTTP API once in `config/dev.exs`:

```elixir
config :optimal_engine, :api, enabled: true, port: 4200
```

Then in two terminals:

```bash
# terminal A — engine + API
iex -S mix

# terminal B — desktop
cd desktop
npm install
npm run dev           # browser preview at http://localhost:1420
# or
npm run tauri:dev     # native window
```

The desktop has seven routes: **Ask**, **Workspace**, **Graph**, **Wiki**, **Architectures**, **Activity**, **Status**. Light and dark themes are selectable in the header.

## 5. Make it yours

The `sample-workspace/` directory is your on-disk reference. Copy its shape to a fresh location and start replacing fixtures with real signals:

```bash
mix optimal.init ~/my-engine
# edit files under ~/my-engine/nodes/*/signals/*.md
mix optimal.ingest_workspace ~/my-engine
```

Every signal file carries YAML frontmatter + a markdown body. Frontmatter field reference lives in [`sample-workspace/README.md`](sample-workspace/README.md).

## 6. Sanity-check at any time

```bash
mix optimal.reality_check --hard
```

Runs 50+ probes across every storage table, every pipeline stage, every retrieval path, every compliance workflow. Prints OK/WARN/FAIL + elapsed ms. Target: all green, total wall-clock under 1 second.

## Troubleshooting

**`mix deps.get` fails with compilation errors on `exqlite`** — install the C toolchain for your platform (Xcode Command-Line Tools on macOS, `build-essential` on Debian).

**`mix optimal.rag` takes several seconds** — you have Ollama running but without `nomic-embed-text` pulled. Either `ollama pull nomic-embed-text`, or ignore — the engine detects the gap and falls through to BM25-only.

**Desktop boots but `Status` reads "down"** — the engine isn't running on `127.0.0.1:4200`. Check `iex -S mix` is up and `config/dev.exs` has the API block enabled.

**`npm run dev` fails with `pixi.js` / `three` resolution errors** — the `node_modules` cache is stale. Delete and reinstall: `rm -rf node_modules && npm install`.

## What next

- Read [`docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md) for the 9-stage pipeline.
- Read [`docs/architecture/DATA_ARCHITECTURE.md`](docs/architecture/DATA_ARCHITECTURE.md) for the universal data-point layer.
- Read [`docs/concepts/signal-theory.md`](docs/concepts/signal-theory.md) for `S=(M,G,T,F,W)`.
