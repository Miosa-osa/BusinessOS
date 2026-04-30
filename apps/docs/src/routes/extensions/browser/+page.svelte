<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Browser Extension — Optimal Engine</title>
</svelte:head>

<h1>Browser Extension</h1>
<p>
	Clip any web page directly into your Optimal Engine workspace. The extension captures the page
	content, strips boilerplate, and sends it to the engine via the HTTP API for full pipeline
	processing.
</p>

<div class="callout">
	<strong>Status: Planned.</strong> The browser extension is on the roadmap. The HTTP API it will
	use is stable today — you can build your own clip-to-engine flow with the endpoints below.
</div>

<h2>Manual Web Clipping — Today</h2>

<p>
	Until the extension ships, use this bookmarklet or script to capture the current page and ingest
	it:
</p>

<CodeBlock code={`# Clip a URL via HTTP API (uses engine's HTML parser backend)
curl -X POST http://localhost:4200/api/rag \\
  -H 'Content-Type: application/json' \\
  -d '{
    "content": "$(curl -s https://example.com/article)",
    "workspace": "default",
    "genre": "article",
    "mode": "linguistic"
  }'`} lang="bash" />

<h2>Planned Extension Features</h2>

<ul>
	<li>One-click clip of full page content into a workspace</li>
	<li>Highlight-to-memory — select text, right-click, "Remember in Optimal"</li>
	<li>Workspace selector in the popup</li>
	<li>Instant "ask this page" via the RAG endpoint</li>
	<li>Citation back to the source URL preserved as <code>citation_uri</code></li>
</ul>

<h2>Build Your Own</h2>

<p>
	The content script pattern is straightforward. Any extension that can read page content can
	integrate with Optimal Engine:
</p>

<CodeBlock code={`// content-script.js (Chrome/Firefox/Safari WebExtension)
async function clipCurrentPage() {
  const content = document.body.innerText;
  const title = document.title;
  const url = window.location.href;

  const response = await fetch('http://localhost:4200/api/memory', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      content: \`# \${title}\\n\\n\${content}\`,
      workspace: 'default',
      citation_uri: url,
      metadata: { source_url: url, clipped_at: new Date().toISOString() },
    }),
  });

  const result = await response.json();
  return result.id; // mem:...
}`} lang="javascript" />

<h2>See Also</h2>
<ul>
	<li><a href="/api/memory">Memory API</a> — the endpoint the extension writes to</li>
	<li><a href="/extensions/raycast">Raycast Extension</a> — desktop quick-clip</li>
	<li><a href="/sdks/typescript">TypeScript SDK</a> — typed client for custom integrations</li>
</ul>
