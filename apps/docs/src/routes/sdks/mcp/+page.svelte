<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>MCP Server — Optimal Engine</title>
</svelte:head>

<h1>MCP Server</h1>
<p>
	The Optimal Engine MCP server exposes organizational memory as tools available to any MCP-compatible client: Claude Desktop, Cursor, Windsurf, and any agent runtime that supports the Model Context Protocol.
</p>

<h2>Install</h2>

<p>Add to your MCP client config. The server is distributed as an npm package:</p>

<CodeBlock code={`{
  "mcpServers": {
    "optimal-engine": {
      "command": "npx",
      "args": ["-y", "@optimal-engine/mcp-server"],
      "env": {
        "OPTIMAL_ENGINE_URL": "http://localhost:4200",
        "OPTIMAL_WORKSPACE": "default"
      }
    }
  }
}`} lang="json" filename="claude_desktop_config.json" />

<h2>Claude Desktop Config Location</h2>

<table>
	<thead>
		<tr><th>Platform</th><th>Path</th></tr>
	</thead>
	<tbody>
		<tr>
			<td>macOS</td>
			<td><code>~/Library/Application Support/Claude/claude_desktop_config.json</code></td>
		</tr>
		<tr>
			<td>Windows</td>
			<td><code>%APPDATA%\Claude\claude_desktop_config.json</code></td>
		</tr>
		<tr>
			<td>Linux</td>
			<td><code>~/.config/claude/claude_desktop_config.json</code></td>
		</tr>
	</tbody>
</table>

<h2>Available Tools</h2>

<p>Once connected, the agent runtime can call these tools:</p>

<table>
	<thead>
		<tr><th>Tool</th><th>Description</th><th>Maps to</th></tr>
	</thead>
	<tbody>
		<tr>
			<td><code>recall</code></td>
			<td>Ask an open question about organizational memory</td>
			<td><code>POST /api/rag</code></td>
		</tr>
		<tr>
			<td><code>remember</code></td>
			<td>Persist a new observation to memory</td>
			<td><code>POST /api/memory</code></td>
		</tr>
		<tr>
			<td><code>search</code></td>
			<td>Find documents by keyword or semantic similarity</td>
			<td><code>GET /api/search</code></td>
		</tr>
		<tr>
			<td><code>recall_who</code></td>
			<td>Ownership / contact lookup</td>
			<td><code>GET /api/recall/who</code></td>
		</tr>
		<tr>
			<td><code>recall_actions</code></td>
			<td>Past decisions and commitments</td>
			<td><code>GET /api/recall/actions</code></td>
		</tr>
		<tr>
			<td><code>recall_when</code></td>
			<td>Schedule and temporal lookup</td>
			<td><code>GET /api/recall/when</code></td>
		</tr>
		<tr>
			<td><code>recall_where</code></td>
			<td>Object-location lookup</td>
			<td><code>GET /api/recall/where</code></td>
		</tr>
		<tr>
			<td><code>list_wiki</code></td>
			<td>List curated wiki pages</td>
			<td><code>GET /api/wiki</code></td>
		</tr>
		<tr>
			<td><code>read_wiki</code></td>
			<td>Read a specific wiki page</td>
			<td><code>GET /api/wiki/:slug</code></td>
		</tr>
	</tbody>
</table>

<h2>Cursor Config</h2>

<CodeBlock code={`# .cursor/mcp.json (project root) or ~/.cursor/mcp.json (global)
{
  "mcpServers": {
    "optimal-engine": {
      "command": "npx",
      "args": ["-y", "@optimal-engine/mcp-server"],
      "env": {
        "OPTIMAL_ENGINE_URL": "http://localhost:4200",
        "OPTIMAL_WORKSPACE": "engineering"
      }
    }
  }
}`} lang="json" filename=".cursor/mcp.json" />

<h2>Multi-Workspace Setup</h2>

<p>
	Run multiple MCP server instances — one per workspace — and alias them in your config:
</p>

<CodeBlock code={`{
  "mcpServers": {
    "optimal-sales": {
      "command": "npx",
      "args": ["-y", "@optimal-engine/mcp-server"],
      "env": {
        "OPTIMAL_ENGINE_URL": "http://localhost:4200",
        "OPTIMAL_WORKSPACE": "sales"
      }
    },
    "optimal-engineering": {
      "command": "npx",
      "args": ["-y", "@optimal-engine/mcp-server"],
      "env": {
        "OPTIMAL_ENGINE_URL": "http://localhost:4200",
        "OPTIMAL_WORKSPACE": "engineering"
      }
    }
  }
}`} lang="json" />

<h2>Environment Variables</h2>

<table>
	<thead>
		<tr><th>Variable</th><th>Default</th><th>Description</th></tr>
	</thead>
	<tbody>
		<tr>
			<td><code>OPTIMAL_ENGINE_URL</code></td>
			<td><code>http://localhost:4200</code></td>
			<td>Engine base URL. Must be reachable from the MCP server process.</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_WORKSPACE</code></td>
			<td><code>default</code></td>
			<td>Default workspace for all tool calls. Can be overridden per-call.</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_AUDIENCE</code></td>
			<td><code>default</code></td>
			<td>Default audience tag for retrieval calls.</td>
		</tr>
	</tbody>
</table>

<h2>See Also</h2>
<ul>
	<li><a href="/sdks/typescript">TypeScript SDK</a> — for direct HTTP integration</li>
	<li><a href="/sdks/python">Python SDK</a> — for Python agent frameworks</li>
	<li><a href="/self-host">Self-host</a> — running the engine server the MCP client connects to</li>
</ul>
