<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Memory Primitive — Optimal Engine</title>
</svelte:head>

<h1>Memory Primitive</h1>
<p>
	A Memory is a first-class, versioned, relation-tracked entry. It is not a key-value store. Every
	memory carries provenance: who wrote it, what it cites, how it relates to other memories, and
	whether it has been superseded or forgotten.
</p>

<h2>When to Use Memory vs. Ingest vs. Wiki</h2>

<table>
	<thead>
		<tr><th>Situation</th><th>What to do</th></tr>
	</thead>
	<tbody>
		<tr><td>An agent observes a fact during a session and needs it to persist</td><td><code>POST /api/memory</code></td></tr>
		<tr><td>A document, transcript, or signal file exists on disk</td><td><code>mix optimal.ingest --file</code></td></tr>
		<tr><td>The engine should maintain a curated summary of a topic</td><td>Let Stage 9 (Curator) handle it automatically on ingest</td></tr>
		<tr><td>An agent wants to annotate an existing wiki page</td><td><code>POST /api/memory</code> with <code>citation_uri</code> pointing at the page</td></tr>
		<tr><td>A fact was true last week but has changed</td><td><code>POST /api/memory/:id/update</code> with the new content</td></tr>
		<tr><td>A fact is a specialization of an existing memory</td><td><code>POST /api/memory/:id/extend</code></td></tr>
		<tr><td>A conclusion drawn from a set of facts</td><td><code>POST /api/memory/:id/derive</code></td></tr>
	</tbody>
</table>

<p>
	<strong>Rule of thumb:</strong> signals are what happened (append-only source documents). Memories
	are what was observed or concluded (first-class facts). The wiki is what the LLM curated from
	both.
</p>

<h2>Five Relation Types</h2>

<p>
	Relations create a typed directed graph over memories. Every relation is directed: source →
	target.
</p>

<table>
	<thead>
		<tr><th>Relation</th><th>Semantics</th><th>Use when</th></tr>
	</thead>
	<tbody>
		<tr>
			<td><code>updates</code></td>
			<td>Source is a newer version of target</td>
			<td>A fact changed; target is now stale</td>
		</tr>
		<tr>
			<td><code>extends</code></td>
			<td>Source adds detail to target without superseding it</td>
			<td>An addendum, clarification, or specialization</td>
		</tr>
		<tr>
			<td><code>derives</code></td>
			<td>Source is a conclusion drawn from target</td>
			<td>A summary, inference, or analysis</td>
		</tr>
		<tr>
			<td><code>contradicts</code></td>
			<td>Source conflicts with target</td>
			<td>Explicitly flagging a known conflict</td>
		</tr>
		<tr>
			<td><code>cites</code></td>
			<td>Source references target as evidence</td>
			<td>Attribution without version or derivation semantics</td>
		</tr>
	</tbody>
</table>

<CodeBlock code={`ENGINE=http://localhost:4200

# Create the original memory
ORIGINAL=$(curl -s -X POST $ENGINE/api/memory \\
  -H 'Content-Type: application/json' \\
  -d '{"content":"Alice owns the pricing negotiation","workspace":"sales"}' | jq -r '.id')

# Update it (relation: updates — original is now stale)
curl -s -X POST "$ENGINE/api/memory/$ORIGINAL/update" \\
  -H 'Content-Type: application/json' \\
  -d '{"content":"Alice and Bob co-own the pricing negotiation as of 2026-04-28"}'

# Extend it (relation: extends — original still valid)
curl -s -X POST "$ENGINE/api/memory/$ORIGINAL/extend" \\
  -H 'Content-Type: application/json' \\
  -d '{"content":"Bob is the fallback contact when Alice is OOO"}'

# Derive from it (relation: derives)
curl -s -X POST "$ENGINE/api/memory/$ORIGINAL/derive" \\
  -H 'Content-Type: application/json' \\
  -d '{"content":"Pricing negotiations have dual ownership — escalate to both"}'`} lang="bash" />

<h2>Versioning Model</h2>

<p>
	Every <code>update</code> call creates a new memory entry linked to the original via the
	<code>updates</code> relation. The original is marked <code>is_latest: false</code>.
</p>

<table>
	<thead>
		<tr><th>Field</th><th>Meaning</th></tr>
	</thead>
	<tbody>
		<tr><td><code>version</code></td><td>Monotonically increasing integer (1, 2, 3, ...)</td></tr>
		<tr><td><code>parent_memory_id</code></td><td>ID of the immediately preceding version</td></tr>
		<tr><td><code>root_memory_id</code></td><td>ID of the first version in the chain</td></tr>
		<tr><td><code>is_latest</code></td><td><code>true</code> on the newest version only</td></tr>
	</tbody>
</table>

<CodeBlock code={`curl http://localhost:4200/api/memory/mem:abc123/versions | jq '.'
# {
#   "memory_id": "mem:abc123",
#   "root_id": "mem:root000",
#   "versions": [
#     { "id": "mem:root000", "version": 1, "content": "...", "is_latest": false },
#     { "id": "mem:abc123",  "version": 2, "content": "...", "is_latest": true }
#   ]
# }`} lang="bash" />

<h2>Forgetting</h2>

<p>
	Forgetting is soft by default. The memory is marked <code>is_forgotten: true</code> but the data
	is preserved for audit and GDPR erasure workflows.
</p>

<CodeBlock code={`# Soft forget immediately
curl -X POST http://localhost:4200/api/memory/mem:abc123/forget \\
  -H 'Content-Type: application/json' \\
  -d '{"reason":"decision reversed"}'

# Schedule future forgetting (GDPR Art. 17)
curl -X POST http://localhost:4200/api/memory/mem:abc123/forget \\
  -H 'Content-Type: application/json' \\
  -d '{"reason":"retention policy","forget_after":"2027-01-01T00:00:00Z"}'

# Hard delete (irreversible)
curl -X DELETE http://localhost:4200/api/memory/mem:abc123`} lang="bash" />

<h2>Static vs. Dynamic</h2>

<p>
	The <code>is_static</code> boolean controls where a memory appears in the
	<a href="/api/retrieval">Profile</a> response.
</p>

<table>
	<thead>
		<tr><th>Value</th><th>Profile tier</th><th>Meaning</th></tr>
	</thead>
	<tbody>
		<tr>
			<td><code>true</code></td>
			<td><code>static</code></td>
			<td>Rarely-changing facts (team structure, product definitions, key decisions)</td>
		</tr>
		<tr>
			<td><code>false</code> (default)</td>
			<td><code>dynamic</code></td>
			<td>Rolling, session-scoped, or ephemeral observations</td>
		</tr>
	</tbody>
</table>

<h2>Audience Scoping</h2>

<p>
	Every memory can be tagged with an <code>audience</code>. When retrieving via
	<code>/api/profile?audience=exec</code>, only memories whose audience matches <code>"exec"</code>,
	<code>"default"</code>, or is unset are returned.
</p>

<CodeBlock code={`# Create an exec-only memory
curl -X POST http://localhost:4200/api/memory \\
  -H 'Content-Type: application/json' \\
  -d '{"content":"Board approved M&A budget at $50M","workspace":"default","audience":"exec"}'`} lang="bash" />

<h2>Citation URI</h2>

<p>
	<code>citation_uri</code> links a memory to its source using the <code>optimal://</code> URI
	scheme.
</p>

<CodeBlock code={`# Memory citing a wiki page
curl -X POST http://localhost:4200/api/memory \\
  -H 'Content-Type: application/json' \\
  -d '{
    "content": "Pricing decision confirmed — see wiki",
    "workspace": "sales",
    "citation_uri": "optimal://wiki/healthtech-pricing-decision"
  }'`} lang="bash" />

<h2>See Also</h2>
<ul>
	<li><a href="/api/memory">Memory API</a> — full endpoint reference</li>
	<li><a href="/concepts/workspaces">Workspaces</a> — workspace scoping on memories</li>
	<li><a href="/api/retrieval">Profile endpoint</a> — how static/dynamic memories surface</li>
</ul>
