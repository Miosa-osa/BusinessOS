<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Quickstart — Optimal Engine</title>
</svelte:head>

<h1>Quickstart</h1>
<p>Zero to running in about 5 minutes.</p>

<h2>Prerequisites</h2>

<table>
	<thead>
		<tr>
			<th>What</th>
			<th>Version</th>
			<th>Check</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Elixir</td>
			<td><code>~&gt; 1.17</code></td>
			<td><code>elixir --version</code></td>
		</tr>
		<tr>
			<td>Erlang / OTP</td>
			<td><code>26+</code></td>
			<td><code>erl -version</code></td>
		</tr>
		<tr>
			<td>C toolchain</td>
			<td>any</td>
			<td><code>cc --version</code></td>
		</tr>
		<tr>
			<td>Node</td>
			<td><code>20+</code></td>
			<td><code>node --version</code> (desktop UI only)</td>
		</tr>
	</tbody>
</table>

<p>On macOS:</p>
<CodeBlock code={`brew install elixir node`} lang="bash" />

<p>On Debian / Ubuntu:</p>
<CodeBlock code={`sudo apt install elixir build-essential nodejs`} lang="bash" />

<p>Optional — enriches parser coverage (engine degrades gracefully when absent):</p>
<CodeBlock code={`brew install pdftotext tesseract ffmpeg`} lang="bash" />

<h2>Clone and bootstrap</h2>

<CodeBlock code={`git clone https://github.com/Miosa-osa/OptimalEngine.git
cd OptimalEngine
make install       # deps + compile
make bootstrap     # migrate + ingest sample-workspace/`} lang="bash" />

<p>
	<code>make bootstrap</code> is idempotent — run it again after pulling to re-seed. Without make:
</p>

<CodeBlock code={`mix deps.get
mix compile
mix optimal.bootstrap`} lang="bash" />

<h2>Use the CLI</h2>

<CodeBlock code={`mix optimal.rag "healthtech pricing decision" --trace
mix optimal.search "platform"
mix optimal.wiki list
mix optimal.graph hubs
mix optimal.architectures`} lang="bash" />

<h2>Start the HTTP API</h2>

<p>Enable the API once in <code>config/dev.exs</code>:</p>

<CodeBlock code={`config :optimal_engine, :api, enabled: true, port: 4200`} lang="elixir" filename="config/dev.exs" />

<p>Then start the engine:</p>

<CodeBlock code={`iex -S mix`} lang="bash" />

<p>The API is now live at <code>http://localhost:4200</code>. Test it:</p>

<CodeBlock code={`curl -X POST http://localhost:4200/api/rag \\
  -H 'Content-Type: application/json' \\
  -d '{"query":"healthtech pricing decision","workspace":"default","format":"markdown"}'`} lang="bash" />

<h2>Launch the desktop UI</h2>

<p>In a second terminal:</p>

<CodeBlock code={`cd desktop
npm install
npm run dev           # browser preview at http://localhost:1420`} lang="bash" />

<p>
	The desktop has seven routes: <strong>Ask</strong>, <strong>Workspace</strong>,
	<strong>Graph</strong>, <strong>Wiki</strong>, <strong>Architectures</strong>,
	<strong>Activity</strong>, <strong>Status</strong>.
</p>

<h2>Make it yours</h2>

<p>
	The <code>sample-workspace/</code> directory is the on-disk reference. Copy its shape to a new
	location and replace the fixtures with real signals:
</p>

<CodeBlock code={`mix optimal.init ~/my-engine
# edit files under ~/my-engine/nodes/*/signals/*.md
mix optimal.ingest_workspace ~/my-engine`} lang="bash" />

<p>
	Every signal file carries YAML frontmatter + a markdown body. See
	<a href="/concepts/workspaces">Workspaces</a> for the full frontmatter field reference.
</p>

<h2>Sanity-check</h2>

<CodeBlock code={`mix optimal.reality_check --hard`} lang="bash" />

<p>
	Runs 50+ probes across every storage table, pipeline stage, retrieval path, and compliance
	workflow. Prints OK / WARN / FAIL + elapsed ms. Target: all green, under 1 second total.
</p>

<h2>Troubleshooting</h2>

<div class="callout">
	<strong>mix deps.get fails with exqlite compilation errors</strong>
	Install the C toolchain for your platform: Xcode Command-Line Tools on macOS
	(<code>xcode-select --install</code>), <code>build-essential</code> on Debian.
</div>

<div class="callout">
	<strong>mix optimal.rag takes several seconds</strong>
	Ollama is running but <code>nomic-embed-text</code> is not pulled. Either run
	<code>ollama pull nomic-embed-text</code>, or ignore — the engine detects the gap and falls
	through to BM25-only retrieval.
</div>

<div class="callout">
	<strong>Desktop boots but Status reads "down"</strong>
	The engine is not running on <code>127.0.0.1:4200</code>. Confirm <code>iex -S mix</code> is up
	and <code>config/dev.exs</code> has the API block enabled.
</div>

<h2>What next</h2>

<ul>
	<li><a href="/concepts/three-tiers">Three Tiers</a> — the storage model</li>
	<li><a href="/concepts/nine-stages">Nine Stages</a> — the ingestion pipeline</li>
	<li><a href="/concepts/signal-theory">Signal Theory</a> — S=(M,G,T,F,W)</li>
	<li><a href="/api">API Reference</a> — every endpoint</li>
	<li><a href="/sdks/typescript">TypeScript SDK</a> — drop-in integration</li>
</ul>
