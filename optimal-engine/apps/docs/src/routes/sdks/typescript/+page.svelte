<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>TypeScript SDK — Optimal Engine</title>
</svelte:head>

<h1>TypeScript SDK</h1>
<p>
	Typed HTTP client for Optimal Engine. Works in Node.js, Deno, Bun, and any browser environment.
	The SDK wraps the REST API with full TypeScript types and convenience methods.
</p>

<div class="callout">
	<strong>No package published yet.</strong> The engine is in active development. Use the raw HTTP
	examples below until the npm package ships. The interface shown here is stable.
</div>

<h2>Raw HTTP — Typed Wrapper</h2>

<p>Copy this into <code>lib/optimal-engine.ts</code> and use it directly:</p>

<CodeBlock code={`// lib/optimal-engine.ts
const ENGINE = process.env.OPTIMAL_ENGINE_URL ?? 'http://localhost:4200';

export interface RagResponse {
  answer: string;
  sources: Array<{ slug: string; score: number; snippet: string }>;
  wiki_hit: boolean;
  citations: string[];
}

export interface Memory {
  id: string;
  content: string;
  workspace_id: string;
  is_static: boolean;
  audience: string | null;
  version: number;
  parent_memory_id: string | null;
  root_memory_id: string | null;
  is_latest: boolean;
  is_forgotten: boolean;
  citation_uri: string | null;
  inserted_at: string;
}

export async function ask(
  query: string,
  options: {
    workspace?: string;
    audience?: string;
    bandwidth?: 'l0' | 'medium' | 'full';
    format?: 'markdown' | 'claude' | 'openai' | 'json';
  } = {}
): Promise<RagResponse> {
  const res = await fetch(\`\${ENGINE}/api/rag\`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query,
      workspace: options.workspace ?? 'default',
      audience: options.audience ?? 'default',
      bandwidth: options.bandwidth ?? 'medium',
      format: options.format ?? 'json',
    }),
  });
  if (!res.ok) throw new Error(\`Optimal Engine error: \${res.status}\`);
  return res.json() as Promise<RagResponse>;
}

export async function storeMemory(
  content: string,
  options: {
    workspace?: string;
    isStatic?: boolean;
    audience?: string;
    citationUri?: string;
  } = {}
): Promise<Memory> {
  const res = await fetch(\`\${ENGINE}/api/memory\`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      content,
      workspace: options.workspace ?? 'default',
      is_static: options.isStatic ?? false,
      audience: options.audience,
      citation_uri: options.citationUri,
    }),
  });
  if (!res.ok) throw new Error(\`Memory store error: \${res.status}\`);
  return res.json() as Promise<Memory>;
}`} lang="typescript" filename="lib/optimal-engine.ts" />

<h2>Vercel AI SDK Adapter</h2>

<p>Use organizational memory as system context in a Vercel AI SDK route:</p>

<CodeBlock code={`// app/api/chat/route.ts
import { streamText } from 'ai';
import { anthropic } from '@ai-sdk/anthropic';
import { ask } from '@/lib/optimal-engine';

export async function POST(req: Request) {
  const { messages, workspace = 'default' } = await req.json();
  const lastMessage = messages[messages.length - 1].content as string;

  // Get organizational context — wiki-first, falls back to hybrid retrieval
  const memory = await ask(lastMessage, {
    workspace,
    format: 'claude',       // Returns system-prompt-ready string
    bandwidth: 'medium',    // ~2K tokens
  });

  return streamText({
    model: anthropic('claude-opus-4-5'),
    system: \`You have access to organizational memory:\\n\\n\${memory.answer}\`,
    messages,
  }).toDataStreamResponse();
}`} lang="typescript" filename="app/api/chat/route.ts" />

<h2>Next.js — Memory-Backed Agent</h2>

<CodeBlock code={`// app/api/agent/route.ts
import { ask, storeMemory } from '@/lib/optimal-engine';

export async function POST(req: Request) {
  const { query, workspace } = await req.json();

  // Retrieve context
  const context = await ask(query, { workspace, format: 'json' });

  // ... run agent logic ...

  // Persist new observations
  await storeMemory('Alice confirmed $2K/seat pricing on 2026-04-28', {
    workspace,
    isStatic: false,
  });

  return Response.json({ answer: context.answer, wiki_hit: context.wiki_hit });
}`} lang="typescript" filename="app/api/agent/route.ts" />

<h2>Recall Helpers</h2>

<CodeBlock code={`// lib/recall.ts
const ENGINE = process.env.OPTIMAL_ENGINE_URL ?? 'http://localhost:4200';

export async function recallWho(topic: string, workspace: string) {
  const res = await fetch(
    \`\${ENGINE}/api/recall/who?topic=\${encodeURIComponent(topic)}&workspace=\${workspace}\`
  );
  return res.json();
}

export async function recallActions(actor: string, workspace: string, since?: string) {
  const params = new URLSearchParams({ actor, workspace });
  if (since) params.set('since', since);
  const res = await fetch(\`\${ENGINE}/api/recall/actions?\${params}\`);
  return res.json();
}

// Usage:
const owner = await recallWho('pricing negotiation', 'sales');
const decisions = await recallActions('Alice', 'sales', '2026-01-01');`} lang="typescript" filename="lib/recall.ts" />

<h2>Environment Variables</h2>

<table>
	<thead>
		<tr><th>Variable</th><th>Default</th><th>Description</th></tr>
	</thead>
	<tbody>
		<tr>
			<td><code>OPTIMAL_ENGINE_URL</code></td>
			<td><code>http://localhost:4200</code></td>
			<td>Engine base URL. Set to your self-hosted address in production.</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_WORKSPACE</code></td>
			<td><code>default</code></td>
			<td>Default workspace slug. Override per-call with the workspace option.</td>
		</tr>
	</tbody>
</table>

<h2>See Also</h2>
<ul>
	<li><a href="/sdks/python">Python SDK</a> — OpenAI Agents + LangChain adapters</li>
	<li><a href="/sdks/mcp">MCP Server</a> — Claude Desktop / Cursor / Windsurf integration</li>
	<li><a href="/api/retrieval">Retrieval API</a> — underlying endpoints</li>
</ul>
