<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Raycast Extension — Optimal Engine</title>
</svelte:head>

<h1>Raycast Extension</h1>
<p>
	Query and store organizational memory from Raycast. Ask questions, save observations, and recall
	who owns something — all without leaving your keyboard.
</p>

<div class="callout">
	<strong>Status: Planned.</strong> The Raycast extension is on the roadmap. The API it will use is
	stable today.
</div>

<h2>Planned Commands</h2>

<table>
	<thead>
		<tr><th>Command</th><th>Description</th><th>Shortcut</th></tr>
	</thead>
	<tbody>
		<tr>
			<td>Ask</td>
			<td>Open question against organizational memory</td>
			<td><code>⌘ Space → "ask"</code></td>
		</tr>
		<tr>
			<td>Remember</td>
			<td>Store the selected text or typed note as a memory</td>
			<td><code>⌘ Space → "remember"</code></td>
		</tr>
		<tr>
			<td>Who owns</td>
			<td>Typed ownership lookup</td>
			<td><code>⌘ Space → "who owns"</code></td>
		</tr>
		<tr>
			<td>Search</td>
			<td>Semantic document search across workspace</td>
			<td><code>⌘ Space → "search engine"</code></td>
		</tr>
		<tr>
			<td>Wiki</td>
			<td>Browse and read curated wiki pages</td>
			<td><code>⌘ Space → "wiki"</code></td>
		</tr>
	</tbody>
</table>

<h2>Build Your Own Raycast Extension</h2>

<p>
	Raycast extensions are React + TypeScript. You can build a minimal "Ask" command in under 50
	lines:
</p>

<CodeBlock code={`// src/ask.tsx
import { Action, ActionPanel, Detail, Form, useNavigation } from "@raycast/api";
import { useState } from "react";

const ENGINE = "http://localhost:4200";

export default function AskCommand() {
  const { push } = useNavigation();
  const [query, setQuery] = useState("");

  async function onSubmit() {
    const res = await fetch(\`\${ENGINE}/api/rag\`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query, workspace: "default", format: "markdown" }),
    });
    const data = await res.json();
    push(<Detail markdown={data.answer} />);
  }

  return (
    <Form actions={
      <ActionPanel>
        <Action.SubmitForm title="Ask" onSubmit={onSubmit} />
      </ActionPanel>
    }>
      <Form.TextField id="query" title="Question" value={query} onChange={setQuery} />
    </Form>
  );
}`} lang="typescript" filename="src/ask.tsx" />

<h2>See Also</h2>
<ul>
	<li><a href="/extensions/browser">Browser Extension</a> — web clipping</li>
	<li><a href="/sdks/typescript">TypeScript SDK</a> — typed client</li>
	<li><a href="/api/retrieval">Retrieval API</a> — underlying endpoints</li>
</ul>
