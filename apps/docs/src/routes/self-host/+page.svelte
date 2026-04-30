<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Self-host — Optimal Engine</title>
</svelte:head>

<h1>Self-host</h1>
<p>
	Optimal Engine is MIT licensed and designed to run on your infrastructure — from a laptop to a
	production server. No cloud dependency. No telemetry. All models run locally via Ollama.
</p>

<h2>Option 1 — From Source (Recommended for Development)</h2>

<CodeBlock code={`git clone https://github.com/Miosa-osa/OptimalEngine.git
cd OptimalEngine
make install    # mix deps.get + mix compile
make bootstrap  # migrate + ingest sample-workspace/
iex -S mix      # starts engine + HTTP API on :4200`} lang="bash" />

<h2>Option 2 — Docker Compose</h2>

<CodeBlock code={`version: "3.9"

services:
  optimal-engine:
    image: ghcr.io/miosa-osa/optimal-engine:latest
    ports:
      - "4200:4200"
    volumes:
      - ./workspaces:/data/workspaces
      - ./config:/data/config
    environment:
      - OPTIMAL_DATA_DIR=/data/workspaces
      - OPTIMAL_API_PORT=4200
      - OPTIMAL_API_ENABLED=true
      - OPTIMAL_OLLAMA_URL=http://ollama:11434
    depends_on:
      - ollama

  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama-data:/root/.ollama

volumes:
  ollama-data:`} lang="yaml" filename="docker-compose.yml" />

<CodeBlock code={`# Pull required models into Ollama before first start
docker compose run --rm ollama ollama pull nomic-embed-text
docker compose run --rm ollama ollama pull llama3.2

# Start
docker compose up -d

# Ingest a workspace
docker compose exec optimal-engine mix optimal.ingest_workspace /data/workspaces/default`} lang="bash" />

<h2>Environment Variables</h2>

<table>
	<thead>
		<tr><th>Variable</th><th>Default</th><th>Description</th></tr>
	</thead>
	<tbody>
		<tr>
			<td><code>OPTIMAL_DATA_DIR</code></td>
			<td><code>./workspaces</code></td>
			<td>Root directory for all workspace data</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_API_PORT</code></td>
			<td><code>4200</code></td>
			<td>HTTP API listen port</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_API_ENABLED</code></td>
			<td><code>false</code></td>
			<td>Must be <code>true</code> to enable the HTTP API</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_OLLAMA_URL</code></td>
			<td><code>http://localhost:11434</code></td>
			<td>Ollama server URL for embeddings and curation</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_OLLAMA_EMBED_MODEL</code></td>
			<td><code>nomic-embed-text</code></td>
			<td>Embedding model to use</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_OLLAMA_CURATE_MODEL</code></td>
			<td><code>llama3.2</code></td>
			<td>LLM for wiki curation (Stage 9)</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_WHISPER_URL</code></td>
			<td>unset</td>
			<td>whisper.cpp server URL for audio transcription. Optional.</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_LOG_LEVEL</code></td>
			<td><code>info</code></td>
			<td>Elixir Logger level: <code>debug | info | warning | error</code></td>
		</tr>
		<tr>
			<td><code>OPTIMAL_LOG_FORMAT</code></td>
			<td><code>text</code></td>
			<td><code>text | json</code> — use <code>json</code> for structured log ingestion</td>
		</tr>
	</tbody>
</table>

<h2>Production Config</h2>

<CodeBlock code={`# config/prod.exs
import Config

config :optimal_engine, :api,
  enabled: true,
  port: String.to_integer(System.get_env("OPTIMAL_API_PORT", "4200"))

config :optimal_engine, :ollama,
  url: System.get_env("OPTIMAL_OLLAMA_URL", "http://localhost:11434"),
  embed_model: System.get_env("OPTIMAL_OLLAMA_EMBED_MODEL", "nomic-embed-text"),
  curate_model: System.get_env("OPTIMAL_OLLAMA_CURATE_MODEL", "llama3.2")

config :optimal_engine, :store,
  data_dir: System.get_env("OPTIMAL_DATA_DIR", "./workspaces")

config :logger,
  level: String.to_existing_atom(System.get_env("OPTIMAL_LOG_LEVEL", "info"))`} lang="elixir" filename="config/prod.exs" />

<h2>OTP Release</h2>

<CodeBlock code={`# Build a standalone OTP release
mix release optimal

# Run it (no Elixir/Erlang required on the target machine via Burrito)
./_build/prod/rel/optimal/bin/optimal start`} lang="bash" />

<h2>Health Check</h2>

<CodeBlock code={`# Liveness + readiness
curl http://localhost:4200/api/status

# Full knowledge-base diagnostic
curl http://localhost:4200/api/health

# CLI sanity check
mix optimal.reality_check --hard`} lang="bash" />

<p>
	The <code>/api/status</code> endpoint checks: store connectivity, migrations, credentials, and
	embedder availability. Returns:
</p>

<CodeBlock code={`{
  "status": "ok",
  "ok": true,
  "checks": {
    "store": "ok",
    "migrations": "ok",
    "embedder": "ok"
  },
  "degraded": []
}`} lang="json" />

<h2>Backup and Restore</h2>

<CodeBlock code={`# Backup — copy Tier 1 files (Tier 2 is rebuildable)
rsync -av ./workspaces/ /backup/optimal-$(date +%Y%m%d)/

# Restore — copy back, then rebuild Tier 2
rsync -av /backup/optimal-20260428/ ./workspaces/
mix optimal.rebuild

# Verify after restore
mix optimal.reality_check --hard`} lang="bash" />

<div class="callout">
	<strong>Tier 2 is always rebuildable from Tier 1.</strong> You only need to back up the
	<code>workspaces/</code> directory (raw signal files, wiki pages). The
	<code>.optimal/index.db</code> can be reconstructed with <code>mix optimal.rebuild</code>.
</div>

<h2>Scaling Considerations</h2>

<ul>
	<li>
		<strong>Ollama on GPU</strong> — point <code>OPTIMAL_OLLAMA_URL</code> at a GPU-backed Ollama
		instance for faster embedding and curation. Embedding throughput is the primary bottleneck on
		large ingests.
	</li>
	<li>
		<strong>SQLite WAL mode</strong> — enabled by default. Supports concurrent reads with a single
		writer. Sufficient for most organizations up to ~100M chunks.
	</li>
	<li>
		<strong>Multiple engine instances</strong> — not supported in v0.1. A single engine process
		handles all workspaces.
	</li>
	<li>
		<strong>whisper.cpp</strong> — optional. Without it, audio/video files are skipped during
		ingest. Run a local whisper.cpp server and set <code>OPTIMAL_WHISPER_URL</code>.
	</li>
</ul>

<h2>See Also</h2>
<ul>
	<li><a href="/quickstart">Quickstart</a> — local development setup</li>
	<li><a href="/api">API Overview</a> — full endpoint reference</li>
	<li><a href="/concepts/three-tiers">Three Tiers</a> — what data lives where on disk</li>
</ul>
