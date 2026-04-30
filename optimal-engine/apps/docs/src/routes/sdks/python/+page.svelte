<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Python SDK — Optimal Engine</title>
</svelte:head>

<h1>Python SDK</h1>
<p>
	Use Optimal Engine from Python via the REST API. Works with any framework — bare
	<code>requests</code>, OpenAI Agents SDK, LangChain, or LlamaIndex.
</p>

<div class="callout">
	<strong>No package published yet.</strong> Copy the client class below. The interface is stable.
</div>

<h2>Client Class</h2>

<CodeBlock code={`# optimal_engine.py
from __future__ import annotations
import os, requests
from dataclasses import dataclass
from typing import Literal, Optional

ENGINE = os.getenv("OPTIMAL_ENGINE_URL", "http://localhost:4200")

@dataclass
class RagResponse:
    answer: str
    sources: list[dict]
    wiki_hit: bool
    citations: list[str]

def ask(
    query: str,
    workspace: str = "default",
    audience: str = "default",
    bandwidth: Literal["l0", "medium", "full"] = "medium",
    format: Literal["markdown", "claude", "openai", "json"] = "json",
) -> RagResponse:
    """Wiki-first open question. Returns answer with citations."""
    r = requests.post(
        f"{ENGINE}/api/rag",
        json={"query": query, "workspace": workspace, "audience": audience,
              "bandwidth": bandwidth, "format": format},
        timeout=30,
    )
    r.raise_for_status()
    data = r.json()
    return RagResponse(
        answer=data["answer"],
        sources=data.get("sources", []),
        wiki_hit=data.get("wiki_hit", False),
        citations=data.get("citations", []),
    )

def store_memory(
    content: str,
    workspace: str = "default",
    is_static: bool = False,
    audience: Optional[str] = None,
    citation_uri: Optional[str] = None,
) -> dict:
    """Persist a memory entry. Returns the created memory struct."""
    r = requests.post(
        f"{ENGINE}/api/memory",
        json={"content": content, "workspace": workspace,
              "is_static": is_static, "audience": audience,
              "citation_uri": citation_uri},
        timeout=10,
    )
    r.raise_for_status()
    return r.json()

def recall_who(topic: str, workspace: str = "default", role: str = "owner") -> dict:
    r = requests.get(f"{ENGINE}/api/recall/who",
                     params={"topic": topic, "workspace": workspace, "role": role})
    r.raise_for_status()
    return r.json()

def recall_actions(actor: str, workspace: str = "default", since: Optional[str] = None) -> dict:
    params = {"actor": actor, "workspace": workspace}
    if since:
        params["since"] = since
    r = requests.get(f"{ENGINE}/api/recall/actions", params=params)
    r.raise_for_status()
    return r.json()`} lang="python" filename="optimal_engine.py" />

<h2>OpenAI Agents SDK</h2>

<CodeBlock code={`from agents import Agent, Runner, function_tool
from optimal_engine import ask, store_memory

@function_tool
def recall(query: str, workspace: str = "default") -> str:
    """Retrieve organizational memory for a query."""
    response = ask(query, workspace=workspace, format="markdown", bandwidth="medium")
    return response.answer

@function_tool
def remember(observation: str, workspace: str = "default") -> str:
    """Persist a new observation to organizational memory."""
    result = store_memory(observation, workspace=workspace)
    return f"Stored as {result['id']}"

agent = Agent(
    name="Company Brain",
    instructions="""You are a company assistant with access to organizational memory.
Always use recall() before answering questions about company decisions, people, or projects.
Use remember() to persist important observations from this conversation.""",
    tools=[recall, remember],
)

result = Runner.run_sync(agent, "What did Alice decide about Q4 pricing?")
print(result.final_output)`} lang="python" />

<h2>LangChain Tool</h2>

<CodeBlock code={`from langchain.tools import tool
from langchain_anthropic import ChatAnthropic
from langchain.agents import AgentExecutor, create_tool_calling_agent
from langchain_core.prompts import ChatPromptTemplate
from optimal_engine import ask, store_memory

@tool
def recall_memory(query: str) -> str:
    """Retrieve organizational memory. Use before answering any factual question."""
    response = ask(query, workspace="default", format="markdown")
    return response.answer

@tool
def persist_memory(observation: str) -> str:
    """Persist an important observation to organizational memory."""
    result = store_memory(observation)
    return f"Stored: {result['id']}"

tools = [recall_memory, persist_memory]
llm = ChatAnthropic(model="claude-opus-4-5")

prompt = ChatPromptTemplate.from_messages([
    ("system", "You are a company assistant. Always recall memory before answering."),
    ("human", "{input}"),
    ("placeholder", "{agent_scratchpad}"),
])

agent = create_tool_calling_agent(llm, tools, prompt)
executor = AgentExecutor(agent=agent, tools=tools, verbose=True)

result = executor.invoke({"input": "Who owns the pricing negotiation?"})
print(result["output"])`} lang="python" />

<h2>Async Client</h2>

<CodeBlock code={`import asyncio, os
import httpx

ENGINE = os.getenv("OPTIMAL_ENGINE_URL", "http://localhost:4200")

async def ask_async(query: str, workspace: str = "default") -> dict:
    async with httpx.AsyncClient() as client:
        r = await client.post(
            f"{ENGINE}/api/rag",
            json={"query": query, "workspace": workspace, "format": "json"},
            timeout=30,
        )
        r.raise_for_status()
        return r.json()

# Usage
async def main():
    result = await ask_async("what is the current pricing strategy?", "sales")
    print(result["answer"])

asyncio.run(main())`} lang="python" />

<h2>Environment Variables</h2>

<table>
	<thead>
		<tr><th>Variable</th><th>Default</th><th>Description</th></tr>
	</thead>
	<tbody>
		<tr>
			<td><code>OPTIMAL_ENGINE_URL</code></td>
			<td><code>http://localhost:4200</code></td>
			<td>Engine base URL</td>
		</tr>
		<tr>
			<td><code>OPTIMAL_WORKSPACE</code></td>
			<td><code>default</code></td>
			<td>Default workspace slug</td>
		</tr>
	</tbody>
</table>

<h2>See Also</h2>
<ul>
	<li><a href="/sdks/typescript">TypeScript SDK</a> — Vercel AI SDK adapter</li>
	<li><a href="/sdks/mcp">MCP Server</a> — Claude Desktop / Cursor integration</li>
	<li><a href="/api/retrieval">Retrieval API</a> — underlying endpoints</li>
</ul>
